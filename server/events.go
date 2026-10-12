// Events service: minimal pull-point subscriptions (#83).
//
// With SupportEvents=true the server advertises the events XAddr in
// GetCapabilities and GetServices (#46); this file makes that promise
// true. Service-level actions (GetServiceCapabilities, GetEventProperties,
// CreatePullPointSubscription) serve on events_service; the subscription-
// scoped actions (PullMessages, Renew, Unsubscribe) serve on the
// per-subscription address returned in SubscriptionReference —
// events_service/sub/<opaque id>.
//
// Notification payloads use the canonical WS-BaseNotification + ONVIF
// double-layer shape (outer wsnt:Message wrapping an inner tt:Message —
// see the spec sample in issue #82) so spec-faithful clients parse
// Source/Data SimpleItems.

package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Events service namespaces (referenced by the struct tags below):
// tev  http://www.onvif.org/ver10/events/wsdl
// wsnt http://docs.oasis-open.org/wsn/b-2
// wsa  http://www.w3.org/2005/08/addressing
// wstop http://docs.oasis-open.org/wsn/t-1
// tt   http://www.onvif.org/ver10/schema

// Pull-point tuning. MaxPullPoints bounds concurrent subscriptions (the
// advertised cap); termination times are clamped to the ONVIF-typical
// 24h; long polls are clamped server-side so a hostile Timeout cannot pin
// goroutines; queues are lossy at the head (newest wins) like real
// device notification buffers.
const (
	defaultMaxPullPoints        = 10
	defaultPullPointTermination = time.Hour
	maxPullPointTermination     = 24 * time.Hour
	maxPullWait                 = time.Minute
	maxEventQueue               = 100
)

// Events service SOAP message types. Namespaces follow the WSDL facts:
// tev wrappers are ver10/events/wsdl; WS-BaseNotification children
// (CurrentTime, TerminationTime, NotificationMessage, Topic, Message) are
// wsn b-2; endpoint references are wsa; the inner ONVIF message payload
// is ver10/schema.

// GetEventServiceCapabilitiesResponse represents the
// GetServiceCapabilities response of the events service.
type GetEventServiceCapabilitiesResponse struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl GetServiceCapabilitiesResponse"`
	Capabilities struct {
		WSPullPointSupport bool `xml:"WSPullPointSupport,attr"`
		MaxPullPoints      int  `xml:"MaxPullPoints,attr,omitempty"`
	} `xml:"http://www.onvif.org/ver10/events/wsdl Capabilities"`
}

// GetEventPropertiesResponse represents the GetEventProperties response:
// a fixed, empty topic set and the two mandatory ONVIF topic-expression
// dialects. Message-content filtering is not applied, so the spec-blessed
// single empty MessageContentFilterDialect is returned; the mandatory
// TopicNamespaceLocation/MessageContentSchemaLocation point at the ONVIF
// topic namespace and schema files.
type GetEventPropertiesResponse struct {
	XMLName                xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl GetEventPropertiesResponse"`
	TopicNamespaceLocation []string `xml:"http://www.onvif.org/ver10/events/wsdl TopicNamespaceLocation"`
	FixedTopicSet          bool     `xml:"http://docs.oasis-open.org/wsn/b-2 FixedTopicSet"`
	TopicSet               struct {
		XMLName xml.Name `xml:"http://docs.oasis-open.org/wsn/t-1 TopicSet"`
	} `xml:"http://docs.oasis-open.org/wsn/t-1 TopicSet"`
	TopicExpressionDialect       []string `xml:"http://docs.oasis-open.org/wsn/b-2 TopicExpressionDialect"`
	MessageContentFilterDialect  []string `xml:"http://docs.oasis-open.org/wsn/b-2 MessageContentFilterDialect"`
	MessageContentSchemaLocation []string `xml:"http://www.onvif.org/ver10/events/wsdl MessageContentSchemaLocation"`
}

// The mandatory ONVIF topic expression dialects (event.wsdl annotations)
// and the canonical locations advertised in GetEventProperties.
const (
	topicNamespaceLocation = "http://www.onvif.org/ver10/tev/topicns.xml"
	dialectConcrete        = "http://docs.oasis-open.org/wsn/t-1/TopicExpression/Concrete"
	dialectConcreteSet     = "http://www.onvif.org/ver10/tev/topicExpression/ConcreteSet"
	messageSchemaLocation  = "http://www.onvif.org/ver10/schema/onvif.xsd"
)

// endpointReference is a WS-Addressing endpoint reference (the
// SubscriptionReference / ProducerReference wire shape).
type endpointReference struct {
	Address string `xml:"http://www.w3.org/2005/08/addressing Address"`
}

// CreatePullPointSubscriptionResponse represents the
// CreatePullPointSubscription response.
type CreatePullPointSubscriptionResponse struct {
	XMLName               xml.Name          `xml:"http://www.onvif.org/ver10/events/wsdl CreatePullPointSubscriptionResponse"`
	SubscriptionReference endpointReference `xml:"http://www.onvif.org/ver10/events/wsdl SubscriptionReference"`
	CurrentTime           string            `xml:"http://docs.oasis-open.org/wsn/b-2 CurrentTime"`
	TerminationTime       string            `xml:"http://docs.oasis-open.org/wsn/b-2 TerminationTime"`
}

// eventTopic is the wsnt:Topic element (chardata expression, optional
// dialect attribute).
type eventTopic struct {
	Dialect string `xml:"Dialect,attr,omitempty"`
	Value   string `xml:",chardata"`
}

// eventSimpleItem is one tt:SimpleItem name/value pair.
type eventSimpleItem struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:"Value,attr"`
}

// eventItemGroup is one tt:Message SimpleItem group (Source, Key, or
// Data). The wrapper struct exists because encoding/xml path tags
// ("a>b") cannot carry a namespace on the nested segment.
type eventItemGroup struct {
	SimpleItems []eventSimpleItem `xml:"http://www.onvif.org/ver10/schema SimpleItem"`
}

// eventMessage is the inner ONVIF tt:Message: the property operation,
// timestamp, and Source/Key/Data SimpleItem groups.
type eventMessage struct {
	PropertyOperation string         `xml:"PropertyOperation,attr"`
	UtcTime           string         `xml:"UtcTime,attr"`
	Source            eventItemGroup `xml:"http://www.onvif.org/ver10/schema Source"`
	Key               eventItemGroup `xml:"http://www.onvif.org/ver10/schema Key"`
	Data              eventItemGroup `xml:"http://www.onvif.org/ver10/schema Data"`
}

// notificationPayload is the outer wsnt:Message wrapper (an opaque xs:any
// holder) around the inner tt:Message.
type notificationPayload struct {
	Inner eventMessage `xml:"http://www.onvif.org/ver10/schema Message"`
}

// notificationMessage is one wsnt:NotificationMessage.
type notificationMessage struct {
	Topic             eventTopic          `xml:"http://docs.oasis-open.org/wsn/b-2 Topic"`
	ProducerReference *endpointReference  `xml:"http://docs.oasis-open.org/wsn/b-2 ProducerReference,omitempty"`
	Message           notificationPayload `xml:"http://docs.oasis-open.org/wsn/b-2 Message"`
}

// PullMessagesResponse represents the PullMessages response.
type PullMessagesResponse struct {
	XMLName              xml.Name              `xml:"http://www.onvif.org/ver10/events/wsdl PullMessagesResponse"`
	CurrentTime          string                `xml:"http://www.onvif.org/ver10/events/wsdl CurrentTime"`
	TerminationTime      string                `xml:"http://www.onvif.org/ver10/events/wsdl TerminationTime"`
	NotificationMessages []notificationMessage `xml:"http://docs.oasis-open.org/wsn/b-2 NotificationMessage,omitempty"`
}

// RenewResponse represents the wsnt:Renew response.
type RenewResponse struct {
	XMLName         xml.Name `xml:"http://docs.oasis-open.org/wsn/b-2 RenewResponse"`
	CurrentTime     string   `xml:"http://docs.oasis-open.org/wsn/b-2 CurrentTime"`
	TerminationTime string   `xml:"http://docs.oasis-open.org/wsn/b-2 TerminationTime"`
}

// UnsubscribeResponse represents the wsnt:Unsubscribe response.
type UnsubscribeResponse struct {
	XMLName xml.Name `xml:"http://docs.oasis-open.org/wsn/b-2 UnsubscribeResponse"`
}

// Events service request types

type createPullPointSubscriptionRequest struct {
	InitialTerminationTime string `xml:"InitialTerminationTime"`
	Filter                 *struct {
		TopicExpression *struct {
			Dialect string `xml:"Dialect,attr"`
			Value   string `xml:",chardata"`
		} `xml:"TopicExpression"`
	} `xml:"Filter"`
}

type pullMessagesRequest struct {
	Timeout      string `xml:"Timeout"`
	MessageLimit int    `xml:"MessageLimit"`
}

type renewRequest struct {
	TerminationTime string `xml:"TerminationTime"`
}

// queuedNotification is an event stamped and buffered for one subscriber.
type queuedNotification struct {
	topic   string
	message eventMessage
}

// pullPoint is one live subscription: a pull point, or — with consumer
// set — a basic-notification subscription drained by a push sender.
type pullPoint struct {
	id           string
	termination  time.Time
	queue        []queuedNotification
	notify       chan struct{} // cap-1 wakeup for long-polling PullMessages and the push sender
	filter       *topicFilter  // nil → every topic is delivered
	consumer     *pushTarget   // nil → pull point; non-nil → push subscription
	pushFailures int           // consecutive Notify delivery failures (push subscriptions)
}

// Event is one property-event notification handed to the server by the
// host (the PublishEvent seam): an AI motion alarm, a tamper switch, a
// signal-loss detector, … UtcTime and the default PropertyOperation are
// stamped at publish time.
type Event struct {
	// Topic is the topic expression, e.g. "tns1:VideoSource/MotionAlarm".
	Topic string
	// PropertyOperation defaults to "Changed" when empty.
	PropertyOperation string
	// Source, Key, Data are the ONVIF SimpleItem groups of tt:Message.
	Source []SimpleItem
	Key    []SimpleItem
	Data   []SimpleItem
}

// PublishEvent fans an event out to every live pull-point subscription
// (#83 host seam). With no subscribers — SupportEvents=false, or nobody
// pulled a SubscriptionReference yet — it is a safe no-op. Queues are
// lossy at the head: a slow subscriber beyond maxEventQueue pending
// messages loses the oldest first, like a real device notification buffer.
func (s *Server) PublishEvent(ev Event) {
	operation := ev.PropertyOperation
	if operation == "" {
		operation = "Changed"
	}

	qn := queuedNotification{
		topic: ev.Topic,
		message: eventMessage{
			PropertyOperation: operation,
			UtcTime:           time.Now().UTC().Format(time.RFC3339),
		},
	}
	qn.message.Source = simpleItemsToWire(ev.Source)
	qn.message.Key = simpleItemsToWire(ev.Key)
	qn.message.Data = simpleItemsToWire(ev.Data)

	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	s.pruneExpiredPullPointsLocked(time.Now())

	for _, pp := range s.pullPoints {
		if pp.filter != nil && !pp.filter.matches(qn.topic) {
			continue
		}

		if len(pp.queue) >= maxEventQueue {
			pp.queue = pp.queue[1:]
		}

		pp.queue = append(pp.queue, qn)

		select {
		case pp.notify <- struct{}{}:
		default:
		}
	}
}

// registerEventsService registers the events service handlers: the
// service endpoint plus the per-subscription subtree the
// SubscriptionReference addresses point at.
func (s *Server) registerEventsService(mux *http.ServeMux) {
	service := s.newSOAPHandler()
	service.RegisterContextHandler("GetServiceCapabilities", s.HandleGetEventServiceCapabilities)
	service.RegisterContextHandler("GetEventProperties", s.HandleGetEventProperties)
	service.RegisterContextHandler("CreatePullPointSubscription", s.HandleCreatePullPointSubscription)
	// Basic notification interface (wsnt:Subscribe + push Notify).
	service.RegisterContextHandler("Subscribe", s.HandleSubscribe)

	subscription := s.newSOAPHandler()
	subscription.RegisterContextHandler("PullMessages", s.HandlePullMessages)
	subscription.RegisterContextHandler("Renew", s.HandleRenew)
	subscription.RegisterContextHandler("Unsubscribe", s.HandleUnsubscribe)

	mux.Handle(s.config.BasePath+"/events_service", service)
	mux.Handle(s.config.BasePath+"/events_service/sub/", subscription)
}

// Events service handlers

// HandleGetEventServiceCapabilities handles GetServiceCapabilities on the
// events service.
func (s *Server) HandleGetEventServiceCapabilities(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	resp := &GetEventServiceCapabilitiesResponse{}
	resp.Capabilities.WSPullPointSupport = true
	resp.Capabilities.MaxPullPoints = defaultMaxPullPoints

	return resp, nil
}

// HandleGetEventProperties handles GetEventProperties: fixed empty topic
// set, the two mandatory topic-expression dialects (honored — see the
// pull-point topic filter), and the empty message-content filter dialect
// (content filters are not applied).
func (s *Server) HandleGetEventProperties(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	return &GetEventPropertiesResponse{
		TopicNamespaceLocation:       []string{topicNamespaceLocation},
		FixedTopicSet:                true,
		TopicExpressionDialect:       []string{dialectConcrete, dialectConcreteSet},
		MessageContentFilterDialect:  []string{""},
		MessageContentSchemaLocation: []string{messageSchemaLocation},
	}, nil
}

// HandleCreatePullPointSubscription creates a pull point and answers its
// SubscriptionReference address plus the granted termination window.
func (s *Server) HandleCreatePullPointSubscription(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req createPullPointSubscriptionRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	filter, err := parseTopicFilter(req.Filter)
	if err != nil {
		return nil, err
	}

	termination := defaultPullPointTermination
	if req.InitialTerminationTime != "" {
		parsed, err := parseISO8601Duration(req.InitialTerminationTime)
		if err != nil || parsed <= 0 {
			return nil, &soap.SenderFaultError{
				Reason: "Invalid InitialTerminationTime",
				Detail: fmt.Sprintf("got %q, want a positive ISO 8601 duration", req.InitialTerminationTime),
			}
		}

		termination = min(parsed, maxPullPointTermination)
	}

	id, err := randomSubscriptionID()
	if err != nil {
		return nil, fmt.Errorf("generate subscription id: %w", err)
	}

	now := time.Now()

	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	s.pruneExpiredPullPointsLocked(now)

	if len(s.pullPoints) >= defaultMaxPullPoints {
		return nil, &soap.SenderFaultError{
			Reason: "Too many active pull point subscriptions",
			Detail: fmt.Sprintf("device supports at most %d concurrent pull points", defaultMaxPullPoints),
		}
	}

	s.pullPoints[id] = &pullPoint{
		id:          id,
		termination: now.Add(termination),
		notify:      make(chan struct{}, 1),
		filter:      filter,
	}

	return &CreatePullPointSubscriptionResponse{
		SubscriptionReference: endpointReference{
			Address: fmt.Sprintf("http://%s:%d%s/events_service/sub/%s",
				s.advertiseHost(rc), s.config.Port, s.config.BasePath, id),
		},
		CurrentTime:     now.UTC().Format(time.RFC3339),
		TerminationTime: now.Add(termination).UTC().Format(time.RFC3339),
	}, nil
}

// HandlePullMessages long-polls the subscription addressed by the request
// URL: it answers as soon as a notification is queued, when the requested
// Timeout (clamped to maxPullWait) elapses, or when the client goes away.
func (s *Server) HandlePullMessages(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req pullMessagesRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.MessageLimit <= 0 {
		return nil, &soap.SenderFaultError{
			Reason: "Invalid MessageLimit",
			Detail: fmt.Sprintf("got %d, want a positive message limit", req.MessageLimit),
		}
	}

	// A basic-notification subscription is drained by its push sender;
	// PullMessages is not its transport.
	if pp := s.pullPointForRequest(rc); pp != nil && pp.consumer != nil {
		return nil, &soap.SenderFaultError{
			Reason: "Not a pull point",
			Detail: "this subscription delivers via Notify push; PullMessages applies to pull points",
		}
	}

	// PT0S is legal: an immediate, non-blocking poll. Only unparsable
	// timeouts fault.
	wait, err := parseISO8601Duration(req.Timeout)
	if err != nil {
		return nil, &soap.SenderFaultError{
			Reason: "Invalid Timeout",
			Detail: fmt.Sprintf("got %q, want an ISO 8601 duration", req.Timeout),
		}
	}

	wait = min(wait, maxPullWait)
	deadline := time.Now().Add(wait)

	for {
		now := time.Now()

		msgs, pp, err := s.drainSubscription(rc, req.MessageLimit)
		if err != nil {
			return nil, err
		}

		// A cancelled request context ends the long poll with an empty
		// answer — the client is gone, so there is nobody to fault to.
		if len(msgs) > 0 || !now.Before(deadline) || rc.Context().Err() != nil {
			return &PullMessagesResponse{ //nolint:nilerr // context cancellation exits the long poll, it is not a swallowed handler error
				CurrentTime:          now.UTC().Format(time.RFC3339),
				TerminationTime:      pp.termination.UTC().Format(time.RFC3339),
				NotificationMessages: msgs,
			}, nil
		}

		timer := time.NewTimer(time.Until(deadline))

		select {
		case <-pp.notify:
		case <-timer.C:
		case <-rc.Context().Done():
		}

		timer.Stop()
	}
}

// HandleRenew extends the addressed subscription's termination time.
func (s *Server) HandleRenew(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req renewRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	duration, err := parseISO8601Duration(req.TerminationTime)
	if err != nil || duration <= 0 {
		return nil, &soap.SenderFaultError{
			Reason: "Invalid TerminationTime",
			Detail: fmt.Sprintf("got %q, want a positive ISO 8601 duration", req.TerminationTime),
		}
	}

	now := time.Now()

	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	pp, fault := s.livePullPointLocked(rc)
	if fault != nil {
		return nil, fault
	}

	pp.termination = now.Add(min(duration, maxPullPointTermination))

	return &RenewResponse{
		CurrentTime:     now.UTC().Format(time.RFC3339),
		TerminationTime: pp.termination.UTC().Format(time.RFC3339),
	}, nil
}

// HandleUnsubscribe removes the addressed subscription.
func (s *Server) HandleUnsubscribe(rc *soap.RequestContext, _ []byte) (interface{}, error) {
	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	pp, fault := s.livePullPointLocked(rc)
	if fault != nil {
		return nil, fault
	}

	delete(s.pullPoints, pp.id)
	// Wake a push sender so it observes the removal promptly.
	select {
	case pp.notify <- struct{}{}:
	default:
	}

	return &UnsubscribeResponse{}, nil
}

// pullPointForRequest resolves the subscription addressed by the request
// URL without pruning; nil when unknown or expired.
func (s *Server) pullPointForRequest(rc *soap.RequestContext) *pullPoint {
	id, err := subscriptionIDFromRequest(s.config.BasePath, rc)
	if err != nil {
		return nil
	}

	return s.pullPoints[id]
}

// livePullPointLocked resolves the subscription addressed by the request
// URL. Callers hold eventsMu; expired pull points are pruned and reported
// like unknown ones (both are gone from the client's perspective).
func (s *Server) livePullPointLocked(rc *soap.RequestContext) (*pullPoint, error) {
	id, err := subscriptionIDFromRequest(s.config.BasePath, rc)
	if err != nil {
		return nil, err
	}

	pp, ok := s.pullPoints[id]
	if !ok {
		return nil, unknownSubscriptionFault(id)
	}

	if time.Now().After(pp.termination) {
		delete(s.pullPoints, id)

		return nil, unknownSubscriptionFault(id)
	}

	return pp, nil
}

// drainSubscription pops up to limit queued notifications from the
// addressed subscription (long-poll callers re-invoke it after a wakeup).
func (s *Server) drainSubscription(
	rc *soap.RequestContext, limit int,
) ([]notificationMessage, *pullPoint, error) {
	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	pp, fault := s.livePullPointLocked(rc)
	if fault != nil {
		return nil, nil, fault
	}

	count := min(limit, len(pp.queue))
	drained := make([]notificationMessage, count)
	for i := range count {
		drained[i] = s.notificationMessageFor(pp.queue[i])
	}

	pp.queue = pp.queue[count:]

	return drained, pp, nil
}

// notificationMessageFor wraps a queued event in the wire shape, stamping
// the producer as the advertised device service endpoint.
func (s *Server) notificationMessageFor(qn queuedNotification) notificationMessage {
	return notificationMessage{
		Topic: eventTopic{Value: qn.topic},
		ProducerReference: &endpointReference{
			Address: fmt.Sprintf("http://%s:%d%s/device_service",
				s.advertiseHost(nil), s.config.Port, s.config.BasePath),
		},
		Message: notificationPayload{Inner: qn.message},
	}
}

// pruneExpiredPullPointsLocked drops pull points past their termination
// time. Callers hold eventsMu.
func (s *Server) pruneExpiredPullPointsLocked(now time.Time) {
	for id, pp := range s.pullPoints {
		if now.After(pp.termination) {
			delete(s.pullPoints, id)
		}
	}
}

// subscriptionIDFromRequest extracts the opaque subscription id from the
// per-subscription URL (.../events_service/sub/<id>).
func subscriptionIDFromRequest(basePath string, rc *soap.RequestContext) (string, error) {
	if rc == nil || rc.Request == nil {
		return "", &soap.SenderFaultError{
			Reason: "Unknown subscription",
			Detail: "no request context",
		}
	}

	prefix := basePath + "/events_service/sub/"
	id := strings.TrimPrefix(rc.Request.URL.Path, prefix)
	if id == rc.Request.URL.Path || id == "" || strings.Contains(id, "/") {
		return "", &soap.SenderFaultError{
			Reason: "Unknown subscription",
			Detail: "not a subscription endpoint: " + rc.Request.URL.Path,
		}
	}

	return id, nil
}

// unknownSubscriptionFault is the Sender fault for gone, unknown, or
// expired pull points.
func unknownSubscriptionFault(id string) error {
	return &soap.SenderFaultError{
		Reason: "Unknown subscription",
		Detail: "no pull point for id " + id,
	}
}

// randomSubscriptionID mints an opaque subscription token.
func randomSubscriptionID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return hex.EncodeToString(buf), nil
}

// simpleItemsToWire converts the host-facing SimpleItem pairs to the wire
// form.
func simpleItemsToWire(items []SimpleItem) eventItemGroup {
	if len(items) == 0 {
		return eventItemGroup{}
	}

	wire := make([]eventSimpleItem, len(items))
	for i, item := range items {
		wire[i] = eventSimpleItem{Name: item.Name, Value: item.Value}
	}

	return eventItemGroup{SimpleItems: wire}
}

// topicFilter is one subscription's topic-expression filter: the
// Concrete dialect (exact topic path) or the ConcreteSet dialect ( '|'
// alternatives with per-segment '*' wildcards), the two dialects
// GetEventProperties advertises.
type topicFilter struct {
	dialect    string
	expression string
}

// parseTopicFilter validates the CreatePullPointSubscription filter.
// Empty dialect defaults to Concrete (the WS-BaseNotification default);
// unsupported dialects fault instead of being silently ignored.
func parseTopicFilter(req *struct {
	TopicExpression *struct {
		Dialect string `xml:"Dialect,attr"`
		Value   string `xml:",chardata"`
	} `xml:"TopicExpression"`
},
) (*topicFilter, error) {
	if req == nil || req.TopicExpression == nil {
		// No filter: every topic is delivered (nilnil-safe sentinel form).
		return (*topicFilter)(nil), nil
	}

	expr := strings.TrimSpace(req.TopicExpression.Value)
	dialect := req.TopicExpression.Dialect
	if dialect == "" {
		dialect = dialectConcrete
	}

	if dialect != dialectConcrete && dialect != dialectConcreteSet {
		return nil, &soap.SenderFaultError{
			Reason: "Unsupported TopicExpression dialect",
			Detail: fmt.Sprintf("got %q, device supports the mandatory Concrete and ConcreteSet dialects", dialect),
		}
	}

	if expr == "" {
		return nil, &soap.SenderFaultError{
			Reason: "Invalid TopicExpression",
			Detail: "filter topic expression must not be empty",
		}
	}

	return &topicFilter{dialect: dialect, expression: expr}, nil
}

// matches reports whether a notification topic satisfies the filter.
// Prefixes are namespace bindings, not identity: matching compares the
// local path segments ("tns1:VideoSource/MotionAlarm" matches
// "tns1:VideoSource/MotionAlarm" and any equivalent binding).
func (f *topicFilter) matches(topic string) bool {
	topicSegs := topicPathSegments(topic)

	for _, alternative := range strings.Split(f.expression, "|") {
		if matchTopicPath(topicPathSegments(alternative), topicSegs) {
			return true
		}
	}

	return false
}

// topicPathSegments splits a topic expression into local-name segments,
// stripping any namespace prefix from each segment.
func topicPathSegments(expr string) []string {
	parts := strings.Split(strings.TrimSpace(expr), "/")
	segs := make([]string, 0, len(parts))

	for _, p := range parts {
		if idx := strings.LastIndex(p, ":"); idx >= 0 {
			p = p[idx+1:]
		}

		segs = append(segs, p)
	}

	return segs
}

// matchTopicPath compares a filter path against a topic path; a "*"
// filter segment matches any single topic segment.
func matchTopicPath(filter, topic []string) bool {
	if len(filter) == 0 || len(filter) != len(topic) {
		return false
	}

	for i := range filter {
		if filter[i] == "*" {
			continue
		}

		if filter[i] != topic[i] {
			return false
		}
	}

	return true
}

// parseISO8601Duration parses the ISO 8601 duration subset ONVIF uses on
// the wire (PnDTnHnMnS, PT0S included). Date-only forms are rejected — a
// lifetime without a time component is not meaningful for subscription
// terminations and pull timeouts.
func parseISO8601Duration(s string) (time.Duration, error) {
	invalid := func() (time.Duration, error) {
		return 0, fmt.Errorf("invalid ISO 8601 duration %q", s)
	}

	if !strings.HasPrefix(s, "P") {
		return invalid()
	}

	rest := s[1:]

	var days, num int64
	var err error

	for rest != "" && rest[0] != 'T' {
		num, rest, err = scanDurationNumber(rest)
		if err != nil {
			return invalid()
		}

		if rest == "" || rest[0] != 'D' {
			return invalid()
		}

		days += num
		rest = rest[1:]
	}

	if rest == "" || len(rest) == 1 {
		// No 'T' section: date-only ("P1D") or a bare "PT".
		return invalid()
	}

	rest = rest[1:] // consume 'T'

	const (
		hourSecs   int64 = 3600
		minuteSecs int64 = 60
		daySecs    int64 = 24 * 3600
	)

	total := days * daySecs

	for rest != "" {
		num, rest, err = scanDurationNumber(rest)
		if err != nil {
			return invalid()
		}

		if rest == "" {
			return invalid()
		}

		switch rest[0] {
		case 'H':
			total += num * hourSecs
		case 'M':
			total += num * minuteSecs
		case 'S':
			total += num
		default:
			return invalid()
		}

		rest = rest[1:]
	}

	return time.Duration(total) * time.Second, nil
}

// scanDurationNumber reads a run of digits off the front of s and returns
// it with the remainder.
func scanDurationNumber(s string) (int64, string, error) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}

	if i == 0 {
		return 0, s, fmt.Errorf("no digits in %q", s)
	}

	num, err := strconv.ParseInt(s[:i], 10, 64)
	if err != nil {
		return 0, s, fmt.Errorf("parse %q: %w", s[:i], err)
	}

	return num, s[i:], nil
}
