package server

// Events service: the basic-notification interface (wsnt:Subscribe +
// push Notify) — the port of the Rust twin's server-side push half.
//
// A Subscribe registers a consumer (ConsumerReference, http:// URLs only)
// and one sender goroutine drains the subscription's queue, POSTing each
// notification to the consumer as a wsnt:Notify SOAP document. The
// registry entry is the same shape a pull point uses — same topic filter,
// same termination semantics, same cap — so Renew/Unsubscribe operate on
// the returned SubscriptionReference address exactly like a pull point's;
// PullMessages does NOT (it is not a pull point).
//
// Delivery is fire-and-forget: the publish path only queues, this
// goroutine owns the network. After maxPushFailures consecutive delivery
// failures the subscription is auto-unsubscribed (spec-permissible
// housekeeping, same as the Rust twin).

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Push delivery tuning: bounded per-request timeouts so a dead consumer
// cannot pin goroutines, and the failure budget before housekeeping
// removes the subscription.
const (
	maxPushFailures   = 3
	pushIOTimeout     = 5 * time.Second
	pushMaxNotifyBody = 1 << 20
)

// subscribeRequest is the WS-BaseNotification Subscribe body. Both
// termination spellings are accepted: InitialTerminationTime (the spec
// name, emitted by real clients and this library's own client half) and
// the legacy TerminationTime.
type subscribeRequest struct {
	ConsumerReference struct {
		Address string `xml:"http://www.w3.org/2005/08/addressing Address"`
	} `xml:"ConsumerReference"`
	InitialTerminationTime string `xml:"InitialTerminationTime"`
	TerminationTime        string `xml:"TerminationTime"`
	Filter                 *struct {
		TopicExpression *struct {
			Dialect string `xml:"Dialect,attr"`
			Value   string `xml:",chardata"`
		} `xml:"TopicExpression"`
	} `xml:"Filter"`
}

// SubscribeResponse is the basic-notification ack: the
// SubscriptionReference (a pull-style address, so Renew/Unsubscribe
// operate on it through the shared registry) plus the granted termination
// window. Same shape as the create-pull-point response under the wsnt
// root — Subscribe is a WS-BaseNotification action, not a tev one.
type SubscribeResponse struct {
	XMLName               xml.Name          `xml:"http://docs.oasis-open.org/wsn/b-2 SubscribeResponse"`
	SubscriptionReference endpointReference `xml:"http://docs.oasis-open.org/wsn/b-2 SubscriptionReference"`
	CurrentTime           string            `xml:"http://docs.oasis-open.org/wsn/b-2 CurrentTime"`
	TerminationTime       string            `xml:"http://docs.oasis-open.org/wsn/b-2 TerminationTime"`
}

// pushTarget is a parsed consumer address.
type pushTarget struct {
	raw  string
	host string
	port string
	path string
}

// HandleSubscribe registers a basic-notification subscription: the same
// registry a pull point lives in, plus a consumer address and a resident
// sender goroutine (see pushSender).
func (s *Server) HandleSubscribe(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req subscribeRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	consumer := strings.TrimSpace(req.ConsumerReference.Address)
	if consumer == "" {
		return nil, &soap.SenderFaultError{
			Reason: "Invalid Subscribe",
			Detail: "a ConsumerReference/Address is required",
		}
	}

	target, err := parsePushTarget(consumer)
	if err != nil {
		return nil, &soap.SenderFaultError{
			Reason: "Invalid ConsumerReference",
			Detail: err.Error(),
		}
	}

	// Same termination semantics as CreatePullPointSubscription: an
	// absent/empty termination grants the default window; anything else
	// must parse positive and is clamped to maxPullPointTermination.
	termination := defaultPullPointTermination
	raw := req.InitialTerminationTime
	if raw == "" {
		raw = req.TerminationTime
	}
	if raw != "" {
		parsed, err := parseISO8601Duration(raw)
		if err != nil || parsed <= 0 {
			return nil, &soap.SenderFaultError{
				Reason: "Invalid TerminationTime",
				Detail: fmt.Sprintf("got %q, want a positive ISO 8601 duration", raw),
			}
		}

		termination = min(parsed, maxPullPointTermination)
	}

	filter, err := parseTopicFilter(req.Filter)
	if err != nil {
		return nil, err
	}

	id, err := randomSubscriptionID()
	if err != nil {
		return nil, fmt.Errorf("generate subscription id: %w", err)
	}

	now := time.Now()

	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	s.pruneExpiredPullPointsLocked(now)

	// Pull points and basic-notification subscriptions share the cap.
	if len(s.pullPoints) >= defaultMaxPullPoints {
		return nil, &soap.SenderFaultError{
			Reason: "Too many active subscriptions",
			Detail: fmt.Sprintf(
				"device supports at most %d concurrent subscriptions (pull points and basic notification share the cap)",
				defaultMaxPullPoints),
		}
	}

	address := fmt.Sprintf("http://%s:%d%s/events_service/sub/%s",
		s.advertiseHost(rc), s.config.Port, s.config.BasePath, id)

	pp := &pullPoint{
		id:          id,
		termination: now.Add(termination),
		notify:      make(chan struct{}, 1),
		filter:      filter,
		consumer:    target,
	}
	s.pullPoints[id] = pp

	go s.pushSender(pp, address)

	return &SubscribeResponse{
		SubscriptionReference: endpointReference{Address: address},
		CurrentTime:           now.UTC().Format(time.RFC3339),
		TerminationTime:       pp.termination.UTC().Format(time.RFC3339),
	}, nil
}

// pushSender drains the subscription's queue and POSTs every queued
// notification to the consumer. It exits when the subscription is
// unsubscribed or pruned elsewhere, at its termination deadline (re-read
// every wake so a Renew extends it), or — spec-permissible housekeeping —
// after maxPushFailures consecutive delivery failures, removing the
// subscription so a dead consumer cannot pin resources forever.
func (s *Server) pushSender(pp *pullPoint, subscriptionAddress string) {
	client := &http.Client{Timeout: pushIOTimeout}

	for {
		if !s.awaitPushWake(pp) {
			s.removePushSubscription(pp.id)
			return
		}

		s.eventsMu.Lock()
		current, ok := s.pullPoints[pp.id]
		if !ok { // unsubscribed or pruned elsewhere
			s.eventsMu.Unlock()
			return
		}
		queue := current.queue
		current.queue = nil
		s.eventsMu.Unlock()

		for _, qn := range queue {
			if !s.deliverNotify(client, pp, subscriptionAddress, qn) {
				s.removePushSubscription(pp.id)
				return
			}
		}
	}
}

// awaitPushWake blocks until a notification is queued or the subscription
// expires; it reports whether the sender should keep running.
func (s *Server) awaitPushWake(pp *pullPoint) bool {
	s.eventsMu.Lock()
	deadline := pp.termination
	s.eventsMu.Unlock()

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-pp.notify:
		return true
	case <-timer.C:
		return false
	}
}

// deliverNotify POSTs one wsnt:Notify document, tracking the consecutive
// failure budget. Reports whether the sender should keep running.
func (s *Server) deliverNotify(client *http.Client, pp *pullPoint, subscriptionAddress string, qn queuedNotification) bool {
	document := buildNotifyDocument(subscriptionAddress, qn)

	// Delivery is best-effort with a bounded lifetime — a canceled server
	// must not leak sender goroutines waiting on dead consumers.
	ctx, cancel := context.WithTimeout(context.Background(), pushIOTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pp.consumer.raw,
		strings.NewReader(document))
	if err != nil {
		pp.pushFailures++

		return pp.pushFailures < maxPushFailures
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		pp.pushFailures++

		return pp.pushFailures < maxPushFailures
	}

	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, pushMaxNotifyBody))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		pp.pushFailures++

		return pp.pushFailures < maxPushFailures
	}

	pp.pushFailures = 0

	return true
}

// removePushSubscription drops the subscription from the shared registry
// (idempotent).
func (s *Server) removePushSubscription(id string) {
	s.eventsMu.Lock()
	delete(s.pullPoints, id)
	s.eventsMu.Unlock()
}

// parsePushTarget accepts plain http consumer URLs. HTTPS consumers are
// refused: the server carries no TLS client, and silently accepting an
// address it can never deliver to would strand the subscription.
func parsePushTarget(raw string) (*pushTarget, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("unparseable consumer address: %w", err)
	}

	if u.Scheme != "http" {
		return nil, fmt.Errorf("only http:// consumers are supported (got %q)", u.Scheme)
	}

	if u.Hostname() == "" {
		return nil, fmt.Errorf("consumer address carries no host")
	}

	port := u.Port()
	if port == "" {
		port = "80"
	}

	path := u.Path
	if path == "" {
		path = "/"
	}

	return &pushTarget{raw: raw, host: u.Hostname(), port: port, path: path}, nil
}

// pushNotificationMessage is one wsnt:NotificationMessage of a pushed
// Notify: it names the subscription (SubscriptionReference —
// ProducerReference is omitted, WS-BaseNotification makes it optional)
// and carries the same Topic + inner-message payload the pull-point
// response serves.
type pushNotificationMessage struct {
	SubscriptionReference endpointReference   `xml:"http://docs.oasis-open.org/wsn/b-2 SubscriptionReference"`
	Topic                 eventTopic          `xml:"http://docs.oasis-open.org/wsn/b-2 Topic"`
	Message               notificationPayload `xml:"http://docs.oasis-open.org/wsn/b-2 Message"`
}

// pushNotifyBody is the wsnt:Notify body element.
type pushNotifyBody struct {
	XMLName              xml.Name                  `xml:"http://docs.oasis-open.org/wsn/b-2 Notify"`
	NotificationMessages []pushNotificationMessage `xml:"http://docs.oasis-open.org/wsn/b-2 NotificationMessage"`
}

// buildNotifyDocument serializes one queued notification into the full
// SOAP document POSTed to the consumer — the same envelope layout the
// response writer emits (declaration, default SOAP namespace, two-space
// body indent).
func buildNotifyDocument(subscriptionAddress string, qn queuedNotification) string {
	content, err := xml.MarshalIndent(pushNotifyBody{
		NotificationMessages: []pushNotificationMessage{{
			SubscriptionReference: endpointReference{Address: subscriptionAddress},
			Topic:                 eventTopic{Value: qn.topic},
			Message:               notificationPayload{Inner: qn.message},
		}},
	}, "", "  ")
	if err != nil {
		// Static shapes with pre-escaped values — unreachable.
		return ""
	}

	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">`)
	b.WriteString("\n  <Body>\n")

	for _, line := range strings.Split(string(content), "\n") {
		b.WriteString("    ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("  </Body>\n</Envelope>")

	return b.String()
}
