// Events service: raw pull-point primitives and the managed
// subscription loop.

package events

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/types"

	"github.com/mickeyzzc/onvif-go/v2/internal/soap"
)

// Event service namespace.
const Namespace = "http://www.onvif.org/ver10/events/wsdl"

// Event service errors.
var (
	// ErrInvalidSubscriptionReference is returned when subscription reference is invalid.
	ErrInvalidSubscriptionReference = errors.New("invalid subscription reference")
	// ErrInvalidTerminationTime is returned when termination time is invalid.
	ErrInvalidTerminationTime = errors.New("invalid termination time")
	// ErrInvalidMessageLimit is returned when message limit is invalid.
	ErrInvalidMessageLimit = errors.New("invalid message limit: must be positive")
	// ErrInvalidTimeout is returned when timeout is invalid.
	ErrInvalidTimeout = errors.New("invalid timeout: must be positive")
	// ErrInvalidFilter is returned when filter expression is invalid.
	ErrInvalidFilter = errors.New("invalid filter expression")
	// ErrInvalidEventBrokerAddress is returned when event broker address is empty.
	ErrInvalidEventBrokerAddress = errors.New("invalid event broker address: cannot be empty")
	// ErrPullPointNotSupported is returned when pull point is not supported.
	ErrPullPointNotSupported = errors.New("pull point subscription not supported")
	// ErrEventBrokerConfigNil is returned when event broker config is nil.
	ErrEventBrokerConfigNil = errors.New("event broker config cannot be nil")
)

// EventServiceCapabilities represents the capabilities of the event service.
type EventServiceCapabilities struct {
	WSSubscriptionPolicySupport                   bool
	WSPausableSubscriptionManagerInterfaceSupport bool
	MaxNotificationProducers                      int
	MaxPullPoints                                 int
	PersistentNotificationStorage                 bool
	EventBrokerProtocols                          []string
	MaxEventBrokers                               int
	MetadataOverMQTT                              bool
}

// PullPointSubscription represents a pull point subscription.
type PullPointSubscription struct {
	SubscriptionReference string
	CurrentTime           time.Time
	TerminationTime       time.Time
}

// NotificationMessage represents a notification message from an event.
type NotificationMessage struct {
	Topic           string
	Message         EventMessage
	ProducerAddress string
	SubscriptionID  string
}

// EventMessage represents the content of an event message.
type EventMessage struct {
	PropertyOperation string
	UtcTime           time.Time
	Source            []types.SimpleItem
	Key               []types.SimpleItem
	Data              []types.SimpleItem
}

// EventSimpleItem represents a simple name-value pair in an event message.
// Note: Uses types.SimpleItem from types.go which has the same structure.

// TopicSet represents the set of topics supported by the device.
type TopicSet struct {
	Topics []Topic
}

// Topic represents an event topic.
type Topic struct {
	Name        string
	Description string
	Children    []Topic
}

// EventBrokerConfig represents an event broker configuration.
type EventBrokerConfig struct {
	Address            string
	TopicPrefix        string
	UserName           string
	Password           string
	CertificateID      string
	PublishFilter      string
	QoS                int
	Status             string
	CertPathValidation bool
	MetadataFilter     string
}

// EventProperties represents the event properties of the device.
type EventProperties struct {
	TopicNamespaceLocation           []string
	FixedTopicSet                    bool
	TopicSet                         TopicSet
	TopicExpressionDialects          []string
	MessageContentFilterDialects     []string
	ProducerPropertiesFilterDialects []string
	MessageContentSchemaLocation     []string
}

// GetEventServiceCapabilities retrieves the capabilities of the event service.
func (s *Service) GetEventServiceCapabilities(ctx context.Context) (*EventServiceCapabilities, error) {
	endpoint := s.c.EndpointFor(api.ServiceEvents)

	type GetServiceCapabilities struct {
		XMLName xml.Name `xml:"tev:GetServiceCapabilities"`
		Xmlns   string   `xml:"xmlns:tev,attr"`
	}

	type GetServiceCapabilitiesResponse struct {
		XMLName      xml.Name `xml:"GetServiceCapabilitiesResponse"`
		Capabilities struct {
			WSSubscriptionPolicySupport                   bool   `xml:"WSSubscriptionPolicySupport,attr"`
			WSPausableSubscriptionManagerInterfaceSupport bool   `xml:"WSPausableSubscriptionManagerInterfaceSupport,attr"`
			MaxNotificationProducers                      int    `xml:"MaxNotificationProducers,attr"`
			MaxPullPoints                                 int    `xml:"MaxPullPoints,attr"`
			PersistentNotificationStorage                 bool   `xml:"PersistentNotificationStorage,attr"`
			EventBrokerProtocols                          string `xml:"EventBrokerProtocols,attr"`
			MaxEventBrokers                               int    `xml:"MaxEventBrokers,attr"`
			MetadataOverMQTT                              bool   `xml:"MetadataOverMQTT,attr"`
		} `xml:"Capabilities"`
	}

	req := GetServiceCapabilities{
		Xmlns: Namespace,
	}

	var resp GetServiceCapabilitiesResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetEventServiceCapabilities failed: %w", err)
	}

	caps := &EventServiceCapabilities{
		WSSubscriptionPolicySupport:                   resp.Capabilities.WSSubscriptionPolicySupport,
		WSPausableSubscriptionManagerInterfaceSupport: resp.Capabilities.WSPausableSubscriptionManagerInterfaceSupport,
		MaxNotificationProducers:                      resp.Capabilities.MaxNotificationProducers,
		MaxPullPoints:                                 resp.Capabilities.MaxPullPoints,
		PersistentNotificationStorage:                 resp.Capabilities.PersistentNotificationStorage,
		MaxEventBrokers:                               resp.Capabilities.MaxEventBrokers,
		MetadataOverMQTT:                              resp.Capabilities.MetadataOverMQTT,
	}

	// Parse event broker protocols from space-separated string.
	if resp.Capabilities.EventBrokerProtocols != "" {
		caps.EventBrokerProtocols = splitSpaceSeparated(resp.Capabilities.EventBrokerProtocols)
	}

	return caps, nil
}

// CreatePullPointSubscription creates a new pull point subscription.
func (s *Service) CreatePullPointSubscription(
	ctx context.Context,
	filter string,
	initialTerminationTime *time.Duration,
	subscriptionPolicy string,
) (*PullPointSubscription, error) {
	endpoint := s.c.EndpointFor(api.ServiceEvents)

	type Filter struct {
		TopicExpression string `xml:"wsnt:TopicExpression,omitempty"`
	}

	type CreatePullPointSubscription struct {
		XMLName                xml.Name `xml:"tev:CreatePullPointSubscription"`
		XmlnsTev               string   `xml:"xmlns:tev,attr"`
		XmlnsWsnt              string   `xml:"xmlns:wsnt,attr"`
		Filter                 *Filter  `xml:"tev:Filter,omitempty"`
		InitialTerminationTime string   `xml:"tev:InitialTerminationTime,omitempty"`
		SubscriptionPolicy     string   `xml:"tev:SubscriptionPolicy,omitempty"`
	}

	type CreatePullPointSubscriptionResponse struct {
		XMLName               xml.Name `xml:"CreatePullPointSubscriptionResponse"`
		SubscriptionReference struct {
			Address string `xml:"Address"`
		} `xml:"SubscriptionReference"`
		CurrentTime     string `xml:"CurrentTime"`
		TerminationTime string `xml:"TerminationTime"`
	}

	req := CreatePullPointSubscription{
		XmlnsTev:  Namespace,
		XmlnsWsnt: "http://docs.oasis-open.org/wsn/b-2",
	}

	if filter != "" {
		req.Filter = &Filter{
			TopicExpression: filter,
		}
	}

	if initialTerminationTime != nil {
		if *initialTerminationTime <= 0 {
			return nil, ErrInvalidTerminationTime
		}
		req.InitialTerminationTime = formatDuration(*initialTerminationTime)
	}

	if subscriptionPolicy != "" {
		req.SubscriptionPolicy = subscriptionPolicy
	}

	var resp CreatePullPointSubscriptionResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreatePullPointSubscription failed: %w", err)
	}

	subscription := &PullPointSubscription{
		SubscriptionReference: resp.SubscriptionReference.Address,
	}

	if resp.CurrentTime != "" {
		if t, err := time.Parse(time.RFC3339, resp.CurrentTime); err == nil {
			subscription.CurrentTime = t
		}
	}

	if resp.TerminationTime != "" {
		if t, err := time.Parse(time.RFC3339, resp.TerminationTime); err == nil {
			subscription.TerminationTime = t
		}
	}

	return subscription, nil
}

// PullMessages pulls notification messages from a pull point subscription.
func (s *Service) PullMessages(
	ctx context.Context,
	subscriptionReference string,
	timeout time.Duration,
	messageLimit int,
) ([]NotificationMessage, error) {
	if subscriptionReference == "" {
		return nil, ErrInvalidSubscriptionReference
	}

	if timeout <= 0 {
		return nil, ErrInvalidTimeout
	}

	if messageLimit <= 0 {
		return nil, ErrInvalidMessageLimit
	}

	type PullMessages struct {
		XMLName      xml.Name `xml:"tev:PullMessages"`
		Xmlns        string   `xml:"xmlns:tev,attr"`
		Timeout      string   `xml:"tev:Timeout"`
		MessageLimit int      `xml:"tev:MessageLimit"`
	}

	type PullMessagesResponse struct {
		XMLName              xml.Name                 `xml:"PullMessagesResponse"`
		CurrentTime          string                   `xml:"CurrentTime"`
		TerminationTime      string                   `xml:"TerminationTime"`
		NotificationMessages []notificationMessageXML `xml:"NotificationMessage"`
	}

	req := PullMessages{
		Xmlns:        Namespace,
		Timeout:      formatDuration(timeout),
		MessageLimit: messageLimit,
	}

	var resp PullMessagesResponse

	if err := s.c.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return nil, fmt.Errorf("PullMessages failed: %w", err)
	}

	return toNotificationMessages(resp.NotificationMessages), nil
}

// toNotificationMessages maps the shared wire shape onto the public
// NotificationMessage. The spec's double-layer form keeps everything on
// the inner tt:Message; lax single-layer devices put it on the outer
// element. The inner layer wins when present (issue #82: MotionAlarm
// State/Score were dropped when only the outer layer was read). Shared by
// the PullMessages response and the push Notify document — one parser,
// both transports.
func toNotificationMessages(nms []notificationMessageXML) []NotificationMessage {
	messages := make([]NotificationMessage, len(nms))

	for i := range nms {
		nm := &nms[i]
		msg := NotificationMessage{
			Topic:           nm.Topic.Value,
			ProducerAddress: nm.ProducerReference.Address,
		}

		operation := nm.Message.PropertyOperation
		utcTime := nm.Message.UtcTime
		source := nm.Message.Source.SimpleItems
		key := nm.Message.Key.SimpleItems
		data := nm.Message.Data.SimpleItems

		if nm.Message.Inner.PropertyOperation != "" || nm.Message.Inner.UtcTime != "" ||
			len(nm.Message.Inner.Source.SimpleItems) > 0 ||
			len(nm.Message.Inner.Key.SimpleItems) > 0 ||
			len(nm.Message.Inner.Data.SimpleItems) > 0 {
			operation = nm.Message.Inner.PropertyOperation
			utcTime = nm.Message.Inner.UtcTime
			source = nm.Message.Inner.Source.SimpleItems
			key = nm.Message.Inner.Key.SimpleItems
			data = nm.Message.Inner.Data.SimpleItems
		}

		msg.Message.PropertyOperation = operation

		if utcTime != "" {
			if t, err := time.Parse(time.RFC3339, utcTime); err == nil {
				msg.Message.UtcTime = t
			}
		}

		msg.Message.Source = make([]types.SimpleItem, len(source))
		for j, item := range source {
			msg.Message.Source[j] = types.SimpleItem(item)
		}

		msg.Message.Key = make([]types.SimpleItem, len(key))
		for j, item := range key {
			msg.Message.Key[j] = types.SimpleItem(item)
		}

		msg.Message.Data = make([]types.SimpleItem, len(data))
		for j, item := range data {
			msg.Message.Data[j] = types.SimpleItem(item)
		}

		messages[i] = msg
	}

	return messages
}

// Seek seeks to a specific position in the event stream.
func (s *Service) Seek(ctx context.Context, subscriptionReference string, utcTime time.Time, reverse bool) error {
	if subscriptionReference == "" {
		return ErrInvalidSubscriptionReference
	}

	type Seek struct {
		XMLName xml.Name `xml:"tev:Seek"`
		Xmlns   string   `xml:"xmlns:tev,attr"`
		UtcTime string   `xml:"tev:UtcTime"`
		Reverse bool     `xml:"tev:Reverse,omitempty"`
	}

	type SeekResponse struct {
		XMLName xml.Name `xml:"SeekResponse"`
	}

	req := Seek{
		Xmlns:   Namespace,
		UtcTime: utcTime.Format(time.RFC3339),
		Reverse: reverse,
	}

	var resp SeekResponse

	if err := s.c.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return fmt.Errorf("Seek failed: %w", err)
	}

	return nil
}

// SetEventSynchronizationPoint instructs the device to send a synchronization point for events.
func (s *Service) SetEventSynchronizationPoint(ctx context.Context, subscriptionReference string) error {
	if subscriptionReference == "" {
		return ErrInvalidSubscriptionReference
	}

	type SetSynchronizationPoint struct {
		XMLName xml.Name `xml:"tev:SetSynchronizationPoint"`
		Xmlns   string   `xml:"xmlns:tev,attr"`
	}

	type SetSynchronizationPointResponse struct {
		XMLName xml.Name `xml:"SetSynchronizationPointResponse"`
	}

	req := SetSynchronizationPoint{
		Xmlns: Namespace,
	}

	var resp SetSynchronizationPointResponse

	if err := s.c.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return fmt.Errorf("SetSynchronizationPoint failed: %w", err)
	}

	return nil
}

// Unsubscribe terminates a subscription.
func (s *Service) Unsubscribe(ctx context.Context, subscriptionReference string) error {
	if subscriptionReference == "" {
		return ErrInvalidSubscriptionReference
	}

	type Unsubscribe struct {
		XMLName xml.Name `xml:"wsnt:Unsubscribe"`
		Xmlns   string   `xml:"xmlns:wsnt,attr"`
	}

	type UnsubscribeResponse struct {
		XMLName xml.Name `xml:"UnsubscribeResponse"`
	}

	req := Unsubscribe{
		Xmlns: "http://docs.oasis-open.org/wsn/b-2",
	}

	var resp UnsubscribeResponse

	if err := s.c.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return fmt.Errorf("Unsubscribe failed: %w", err)
	}

	return nil
}

// RenewSubscription renews a subscription with a new termination time.
func (s *Service) RenewSubscription(
	ctx context.Context,
	subscriptionReference string,
	terminationTime time.Duration,
) (time.Time, time.Time, error) {
	if subscriptionReference == "" {
		return time.Time{}, time.Time{}, ErrInvalidSubscriptionReference
	}

	if terminationTime <= 0 {
		return time.Time{}, time.Time{}, ErrInvalidTerminationTime
	}

	type Renew struct {
		XMLName         xml.Name `xml:"wsnt:Renew"`
		Xmlns           string   `xml:"xmlns:wsnt,attr"`
		TerminationTime string   `xml:"wsnt:TerminationTime"`
	}

	type RenewResponse struct {
		XMLName         xml.Name `xml:"RenewResponse"`
		CurrentTime     string   `xml:"CurrentTime"`
		TerminationTime string   `xml:"TerminationTime"`
	}

	req := Renew{
		Xmlns:           "http://docs.oasis-open.org/wsn/b-2",
		TerminationTime: formatDuration(terminationTime),
	}

	var resp RenewResponse

	if err := s.c.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("RenewSubscription failed: %w", err)
	}

	var currentTime, newTerminationTime time.Time

	if resp.CurrentTime != "" {
		if t, err := time.Parse(time.RFC3339, resp.CurrentTime); err == nil {
			currentTime = t
		}
	}

	if resp.TerminationTime != "" {
		if t, err := time.Parse(time.RFC3339, resp.TerminationTime); err == nil {
			newTerminationTime = t
		}
	}

	return currentTime, newTerminationTime, nil
}

// GetEventProperties retrieves the event properties of the device.
func (s *Service) GetEventProperties(ctx context.Context) (*EventProperties, error) {
	endpoint := s.c.EndpointFor(api.ServiceEvents)

	type GetEventProperties struct {
		XMLName xml.Name `xml:"tev:GetEventProperties"`
		Xmlns   string   `xml:"xmlns:tev,attr"`
	}

	type GetEventPropertiesResponse struct {
		XMLName                         xml.Name `xml:"GetEventPropertiesResponse"`
		TopicNamespaceLocation          []string `xml:"TopicNamespaceLocation"`
		FixedTopicSet                   bool     `xml:"FixedTopicSet"`
		TopicExpressionDialect          []string `xml:"TopicExpressionDialect"`
		MessageContentFilterDialect     []string `xml:"MessageContentFilterDialect"`
		ProducerPropertiesFilterDialect []string `xml:"ProducerPropertiesFilterDialect"`
		MessageContentSchemaLocation    []string `xml:"MessageContentSchemaLocation"`
	}

	req := GetEventProperties{
		Xmlns: Namespace,
	}

	var resp GetEventPropertiesResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetEventProperties failed: %w", err)
	}

	properties := &EventProperties{
		TopicNamespaceLocation:           resp.TopicNamespaceLocation,
		FixedTopicSet:                    resp.FixedTopicSet,
		TopicExpressionDialects:          resp.TopicExpressionDialect,
		MessageContentFilterDialects:     resp.MessageContentFilterDialect,
		ProducerPropertiesFilterDialects: resp.ProducerPropertiesFilterDialect,
		MessageContentSchemaLocation:     resp.MessageContentSchemaLocation,
	}

	return properties, nil
}

// AddEventBroker adds an event broker configuration.
func (s *Service) AddEventBroker(ctx context.Context, config *EventBrokerConfig) error {
	if config == nil {
		return ErrEventBrokerConfigNil
	}

	if config.Address == "" {
		return ErrInvalidEventBrokerAddress
	}

	endpoint := s.c.EndpointFor(api.ServiceEvents)

	type EventBrokerConfigXML struct {
		Address            string `xml:"tev:Address"`
		TopicPrefix        string `xml:"tev:TopicPrefix,omitempty"`
		UserName           string `xml:"tev:UserName,omitempty"`
		Password           string `xml:"tev:Password,omitempty"`
		CertificateID      string `xml:"tev:CertificateID,omitempty"`
		PublishFilter      string `xml:"tev:PublishFilter,omitempty"`
		QoS                int    `xml:"tev:QoS,omitempty"`
		CertPathValidation bool   `xml:"tev:CertPathValidation,omitempty"`
		MetadataFilter     string `xml:"tev:MetadataFilter,omitempty"`
	}

	type AddEventBroker struct {
		XMLName           xml.Name             `xml:"tev:AddEventBroker"`
		Xmlns             string               `xml:"xmlns:tev,attr"`
		EventBrokerConfig EventBrokerConfigXML `xml:"tev:EventBrokerConfig"`
	}

	type AddEventBrokerResponse struct {
		XMLName xml.Name `xml:"AddEventBrokerResponse"`
	}

	req := AddEventBroker{
		Xmlns: Namespace,
		EventBrokerConfig: EventBrokerConfigXML{
			Address:            config.Address,
			TopicPrefix:        config.TopicPrefix,
			UserName:           config.UserName,
			Password:           config.Password,
			CertificateID:      config.CertificateID,
			PublishFilter:      config.PublishFilter,
			QoS:                config.QoS,
			CertPathValidation: config.CertPathValidation,
			MetadataFilter:     config.MetadataFilter,
		},
	}

	var resp AddEventBrokerResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return fmt.Errorf("AddEventBroker failed: %w", err)
	}

	return nil
}

// DeleteEventBroker deletes an event broker configuration.
func (s *Service) DeleteEventBroker(ctx context.Context, address string) error {
	if address == "" {
		return ErrInvalidEventBrokerAddress
	}

	endpoint := s.c.EndpointFor(api.ServiceEvents)

	type DeleteEventBroker struct {
		XMLName xml.Name `xml:"tev:DeleteEventBroker"`
		Xmlns   string   `xml:"xmlns:tev,attr"`
		Address string   `xml:"tev:Address"`
	}

	type DeleteEventBrokerResponse struct {
		XMLName xml.Name `xml:"DeleteEventBrokerResponse"`
	}

	req := DeleteEventBroker{
		Xmlns:   Namespace,
		Address: address,
	}

	var resp DeleteEventBrokerResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return fmt.Errorf("DeleteEventBroker failed: %w", err)
	}

	return nil
}

// GetEventBrokers retrieves all event broker configurations.
func (s *Service) GetEventBrokers(ctx context.Context) ([]*EventBrokerConfig, error) {
	endpoint := s.c.EndpointFor(api.ServiceEvents)

	type GetEventBrokers struct {
		XMLName xml.Name `xml:"tev:GetEventBrokers"`
		Xmlns   string   `xml:"xmlns:tev,attr"`
	}

	type GetEventBrokersResponse struct {
		XMLName      xml.Name `xml:"GetEventBrokersResponse"`
		EventBrokers []struct {
			Address            string `xml:"Address"`
			TopicPrefix        string `xml:"TopicPrefix"`
			UserName           string `xml:"UserName"`
			Password           string `xml:"Password"`
			CertificateID      string `xml:"CertificateID"`
			PublishFilter      string `xml:"PublishFilter"`
			QoS                int    `xml:"QoS"`
			Status             string `xml:"Status"`
			CertPathValidation bool   `xml:"CertPathValidation"`
			MetadataFilter     string `xml:"MetadataFilter"`
		} `xml:"EventBroker"`
	}

	req := GetEventBrokers{
		Xmlns: Namespace,
	}

	var resp GetEventBrokersResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetEventBrokers failed: %w", err)
	}

	brokers := make([]*EventBrokerConfig, len(resp.EventBrokers))
	for i := range resp.EventBrokers {
		eb := &resp.EventBrokers[i]
		brokers[i] = &EventBrokerConfig{
			Address:            eb.Address,
			TopicPrefix:        eb.TopicPrefix,
			UserName:           eb.UserName,
			Password:           eb.Password,
			CertificateID:      eb.CertificateID,
			PublishFilter:      eb.PublishFilter,
			QoS:                eb.QoS,
			Status:             eb.Status,
			CertPathValidation: eb.CertPathValidation,
			MetadataFilter:     eb.MetadataFilter,
		}
	}

	return brokers, nil
}

// formatDuration formats a duration as an ISO 8601 duration string.
func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 { //nolint:mnd // 60 seconds in a minute
		return fmt.Sprintf("PT%dS", seconds)
	}

	minutes := seconds / 60 //nolint:mnd // 60 seconds in a minute
	seconds %= 60

	if seconds == 0 {
		return fmt.Sprintf("PT%dM", minutes)
	}

	return fmt.Sprintf("PT%dM%dS", minutes, seconds)
}

// splitSpaceSeparated splits a space-separated string into a slice.
func splitSpaceSeparated(s string) []string {
	if s == "" {
		return nil
	}

	return strings.Fields(s)
}

// ErrEventsNotSupported is returned when the device does not implement the
// ONVIF events service (some cameras advertise it in GetCapabilities but
// fault on CreatePullPointSubscription with "Action Not Implemented").
// Without this sentinel, callers would retry an impossible subscription
// forever. Match with errors.Is(err, ErrEventsNotSupported) and cache the
// negative result.
var ErrEventsNotSupported = errors.New("device does not support the ONVIF events service")

// Default managed-subscription tuning.
const (
	defaultEventSubscriptionDuration = time.Hour
	defaultEventRenewMargin          = 5 * time.Minute
	defaultEventPullTimeout          = 30 * time.Second
	defaultEventMessageLimit         = 10
	eventErrorBackoffStart           = time.Second
	eventErrorBackoffMax             = 30 * time.Second
)

// SubscribeEventsOptions tunes the managed event loop. The zero value is
// valid and selects the defaults above.
type SubscribeEventsOptions struct {
	// Filter is a topic expression filter passed to CreatePullPointSubscription.
	Filter string
	// SubscriptionDuration is how long each subscription grant lasts before
	// renewal (default 1h).
	SubscriptionDuration time.Duration
	// RenewMargin is how early before expiry the subscription is renewed
	// (default 5m). Renewal failure terminates the loop.
	RenewMargin time.Duration
	// PullTimeout is the long-poll timeout of each PullMessages call
	// (default 30s).
	PullTimeout time.Duration
	// MessageLimit caps messages per pull (default 10).
	MessageLimit int
}

// EventStream is a managed pull-point subscription: a background
// goroutine long-polls PullMessages, delivers every notification to the
// handler, renews the subscription before it expires, and stops cleanly on
// Unsubscribe or context cancellation.
type EventStream struct {
	svc     *Service
	handler func(NotificationMessage)
	opts    SubscribeEventsOptions

	ref      string
	cancel   context.CancelFunc
	done     chan struct{}
	doneOnce sync.Once

	mu           sync.Mutex
	expiresAt    time.Time
	explicitStop bool
}

// SubscribeEvents creates a pull-point subscription and manages its whole
// lifecycle. The handler runs on the polling goroutine: return quickly, run
// slow work on your own goroutine (a panicking handler is isolated and does
// not kill the loop). The subscription lives until Unsubscribe is called or
// ctx is cancelled; a renewal failure also terminates it (Done closes).
//
// Devices that do not implement the events service yield an error matching
// errors.Is(err, ErrEventsNotSupported).
func (s *Service) SubscribeEvents(
	ctx context.Context,
	handler func(NotificationMessage),
	opts *SubscribeEventsOptions,
) (*EventStream, error) {
	if handler == nil {
		return nil, fmt.Errorf("SubscribeEvents: %w: handler is nil", types.ErrInvalidParameter)
	}

	effective := SubscribeEventsOptions{
		SubscriptionDuration: defaultEventSubscriptionDuration,
		RenewMargin:          defaultEventRenewMargin,
		PullTimeout:          defaultEventPullTimeout,
		MessageLimit:         defaultEventMessageLimit,
	}
	if opts != nil {
		if opts.Filter != "" {
			effective.Filter = opts.Filter
		}

		if opts.SubscriptionDuration > 0 {
			effective.SubscriptionDuration = opts.SubscriptionDuration
		}

		if opts.RenewMargin > 0 {
			effective.RenewMargin = opts.RenewMargin
		}

		if opts.PullTimeout > 0 {
			effective.PullTimeout = opts.PullTimeout
		}

		if opts.MessageLimit > 0 {
			effective.MessageLimit = opts.MessageLimit
		}
	}

	duration := effective.SubscriptionDuration
	sub, err := s.CreatePullPointSubscription(ctx, effective.Filter, &duration, "")
	if err != nil {
		if classifyEventsNotSupported(err) {
			return nil, fmt.Errorf("SubscribeEvents: %w: %w", ErrEventsNotSupported, err)
		}

		return nil, fmt.Errorf("SubscribeEvents: %w", err)
	}

	loopCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	es := &EventStream{
		svc:     s,
		handler: handler,
		opts:    effective,
		ref:     sub.SubscriptionReference,
		cancel:  cancel,
		done:    make(chan struct{}),
	}

	es.mu.Lock()
	if !sub.TerminationTime.IsZero() {
		es.expiresAt = sub.TerminationTime
	} else {
		es.expiresAt = time.Now().Add(duration)
	}
	es.mu.Unlock()

	// Dying with the caller's context is part of the contract.
	go func() {
		<-ctx.Done()
		es.stopLoop()
	}()

	go es.run(loopCtx)

	return es, nil
}

// Done returns a channel that closes when the managed loop has fully exited
// (Unsubscribe, cancellation, or renewal failure). Useful for deterministic
// cleanup in tests and shutdown paths.
func (es *EventStream) Done() <-chan struct{} {
	return es.done
}

// Unsubscribe stops the managed loop and best-effort unsubscribes on the
// device: the loop is terminated even when the SOAP unsubscribe fails (the
// error is then returned, cleanup is not blocked by it).
func (es *EventStream) Unsubscribe(ctx context.Context) error {
	es.mu.Lock()
	es.explicitStop = true
	es.mu.Unlock()
	es.stopLoop()

	err := es.svc.Unsubscribe(ctx, es.ref)
	if err != nil {
		return fmt.Errorf("Unsubscribe: device unsubscribe failed (loop stopped anyway): %w", err)
	}

	return nil
}

// stopLoop cancels the loop exactly once.
func (es *EventStream) stopLoop() {
	es.cancel()
}

// run is the managed polling loop.
func (es *EventStream) run(ctx context.Context) {
	defer es.doneOnce.Do(func() { close(es.done) })

	backoff := eventErrorBackoffStart

	for {
		if ctx.Err() != nil {
			es.cleanupUnlessExplicit(ctx)

			return
		}

		// Renew before the subscription runs out.
		es.mu.Lock()
		renewIn := time.Until(es.expiresAt) - es.opts.RenewMargin
		es.mu.Unlock()

		if renewIn <= 0 && !es.renew(ctx) {
			// Renewal failed: the subscription is gone or the device is
			// refusing; retrying forever would just spam — terminate.
			es.cleanupUnlessExplicit(ctx)

			return
		}

		messages, err := es.svc.PullMessages(ctx, es.ref, es.opts.PullTimeout, es.opts.MessageLimit)
		if err != nil {
			if ctx.Err() != nil {
				es.cleanupUnlessExplicit(ctx)

				return
			}

			// Transient failure: back off, keep the loop alive.
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
			}

			backoff = min(2*backoff, eventErrorBackoffMax)

			continue
		}

		backoff = eventErrorBackoffStart

		es.deliver(messages)

		if len(messages) == 0 {
			// Devices that honor long-polling only get here once per
			// PullTimeout; devices that return instantly-empty would
			// otherwise spin the loop at full speed — throttle them.
			select {
			case <-time.After(eventIdleGap):
			case <-ctx.Done():
			}
		}
	}
}

// eventIdleGap throttles the loop against devices that ignore the long-poll
// timeout and answer immediately with no messages.
const eventIdleGap = 100 * time.Millisecond

// renew extends the subscription; reports success.
func (es *EventStream) renew(ctx context.Context) bool {
	_, termination, err := es.svc.RenewSubscription(ctx, es.ref, es.opts.SubscriptionDuration)
	if err != nil {
		return false
	}

	es.mu.Lock()
	if termination.IsZero() {
		es.expiresAt = time.Now().Add(es.opts.SubscriptionDuration)
	} else {
		es.expiresAt = termination
	}
	es.mu.Unlock()

	return true
}

// deliver hands messages to the handler with panic isolation and missing
// timestamp defaults.
func (es *EventStream) deliver(messages []NotificationMessage) {
	for i := range messages {
		msg := messages[i]

		if msg.Message.UtcTime.IsZero() {
			msg.Message.UtcTime = time.Now().UTC()
		}

		es.safeHandle(msg)
	}
}

// cleanupUnlessExplicit sends a best-effort SOAP unsubscribe when the loop
// ended on its own (renewal failure, caller-context death). When Unsubscribe()
// triggered the stop it already sent the SOAP request itself — a second one
// would be noise.
func (es *EventStream) cleanupUnlessExplicit(parent context.Context) {
	es.mu.Lock()
	explicit := es.explicitStop
	es.mu.Unlock()

	if explicit {
		return
	}

	// The loop context is already cancelled here; detach cancellation but
	// keep it bounded so cleanup cannot hang.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), es.opts.PullTimeout)
	defer cancel()

	_ = es.svc.Unsubscribe(ctx, es.ref)
}

// safeHandle invokes the handler with panic isolation: a panicking handler
// must not kill the polling loop.
func (es *EventStream) safeHandle(msg NotificationMessage) {
	defer func() {
		_ = recover()
	}()

	es.handler(msg)
}

// classifyEventsNotSupported recognizes the "events service not implemented"
// fault family across firmware phrasings.
func classifyEventsNotSupported(err error) bool {
	var faultErr *soap.FaultError
	if !errors.As(err, &faultErr) {
		return false
	}

	normalized := strings.ToLower(strings.NewReplacer(" ", "", "-", "", "_", "").
		Replace(faultErr.Code + "|" + faultErr.Subcode + "|" + faultErr.Reason))

	return strings.Contains(normalized, "notimplemented") ||
		strings.Contains(normalized, "actionnotsupported") ||
		strings.Contains(normalized, "notsupportedbydevice")
}
