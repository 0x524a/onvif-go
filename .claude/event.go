package onvif

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0x524a/onvif-go/internal/soap"
)

// Event service namespace.
const eventNamespace = "http://www.onvif.org/ver10/events/wsdl"

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
	Source            []SimpleItem
	Key               []SimpleItem
	Data              []SimpleItem
}

// EventSimpleItem represents a simple name-value pair in an event message.
// Note: Uses SimpleItem from types.go which has the same structure.

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

// getEventEndpoint returns the event endpoint, falling back to the default endpoint if not set.
func (c *Client) getEventEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.eventEndpoint != "" {
		return c.eventEndpoint
	}

	return c.endpoint
}

// SetEventEndpoint sets the event service endpoint.
func (c *Client) SetEventEndpoint(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventEndpoint = endpoint
}

// GetEventServiceCapabilities retrieves the capabilities of the event service.
func (c *Client) GetEventServiceCapabilities(ctx context.Context) (*EventServiceCapabilities, error) {
	endpoint := c.getEventEndpoint()

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
		Xmlns: eventNamespace,
	}

	var resp GetServiceCapabilitiesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
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
func (c *Client) CreatePullPointSubscription(
	ctx context.Context,
	filter string,
	initialTerminationTime *time.Duration,
	subscriptionPolicy string,
) (*PullPointSubscription, error) {
	endpoint := c.getEventEndpoint()

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
		XmlnsTev:  eventNamespace,
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

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
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
func (c *Client) PullMessages(
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

	type SimpleItemXML struct {
		Name  string `xml:"Name,attr"`
		Value string `xml:"Value,attr"`
	}

	type PullMessagesResponse struct {
		XMLName              xml.Name `xml:"PullMessagesResponse"`
		CurrentTime          string   `xml:"CurrentTime"`
		TerminationTime      string   `xml:"TerminationTime"`
		NotificationMessages []struct {
			Topic struct {
				Value string `xml:",chardata"`
			} `xml:"Topic"`
			ProducerReference struct {
				Address string `xml:"Address"`
			} `xml:"ProducerReference"`
			Message struct {
				PropertyOperation string `xml:"PropertyOperation,attr"`
				UtcTime           string `xml:"UtcTime,attr"`
				Source            struct {
					SimpleItems []SimpleItemXML `xml:"SimpleItem"`
				} `xml:"Source"`
				Key struct {
					SimpleItems []SimpleItemXML `xml:"SimpleItem"`
				} `xml:"Key"`
				Data struct {
					SimpleItems []SimpleItemXML `xml:"SimpleItem"`
				} `xml:"Data"`
			} `xml:"Message"`
		} `xml:"NotificationMessage"`
	}

	req := PullMessages{
		Xmlns:        eventNamespace,
		Timeout:      formatDuration(timeout),
		MessageLimit: messageLimit,
	}

	var resp PullMessagesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return nil, fmt.Errorf("PullMessages failed: %w", err)
	}

	messages := make([]NotificationMessage, len(resp.NotificationMessages))
	for i := range resp.NotificationMessages {
		nm := &resp.NotificationMessages[i]
		msg := NotificationMessage{
			Topic:           nm.Topic.Value,
			ProducerAddress: nm.ProducerReference.Address,
		}

		msg.Message.PropertyOperation = nm.Message.PropertyOperation

		if nm.Message.UtcTime != "" {
			if t, err := time.Parse(time.RFC3339, nm.Message.UtcTime); err == nil {
				msg.Message.UtcTime = t
			}
		}

		// Convert source items.
		msg.Message.Source = make([]SimpleItem, len(nm.Message.Source.SimpleItems))
		for j, item := range nm.Message.Source.SimpleItems {
			msg.Message.Source[j] = SimpleItem{Name: item.Name, Value: item.Value}
		}

		// Convert key items.
		msg.Message.Key = make([]SimpleItem, len(nm.Message.Key.SimpleItems))
		for j, item := range nm.Message.Key.SimpleItems {
			msg.Message.Key[j] = SimpleItem{Name: item.Name, Value: item.Value}
		}

		// Convert data items.
		msg.Message.Data = make([]SimpleItem, len(nm.Message.Data.SimpleItems))
		for j, item := range nm.Message.Data.SimpleItems {
			msg.Message.Data[j] = SimpleItem{Name: item.Name, Value: item.Value}
		}

		messages[i] = msg
	}

	return messages, nil
}

// Seek seeks to a specific position in the event stream.
func (c *Client) Seek(ctx context.Context, subscriptionReference string, utcTime time.Time, reverse bool) error {
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
		Xmlns:   eventNamespace,
		UtcTime: utcTime.Format(time.RFC3339),
		Reverse: reverse,
	}

	var resp SeekResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return fmt.Errorf("Seek failed: %w", err)
	}

	return nil
}

// SetEventSynchronizationPoint instructs the device to send a synchronization point for events.
func (c *Client) SetEventSynchronizationPoint(ctx context.Context, subscriptionReference string) error {
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
		Xmlns: eventNamespace,
	}

	var resp SetSynchronizationPointResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return fmt.Errorf("SetSynchronizationPoint failed: %w", err)
	}

	return nil
}

// Unsubscribe terminates a subscription.
func (c *Client) Unsubscribe(ctx context.Context, subscriptionReference string) error {
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

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
		return fmt.Errorf("Unsubscribe failed: %w", err)
	}

	return nil
}

// RenewSubscription renews a subscription with a new termination time.
func (c *Client) RenewSubscription(
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

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, subscriptionReference, "", req, &resp); err != nil {
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
func (c *Client) GetEventProperties(ctx context.Context) (*EventProperties, error) {
	endpoint := c.getEventEndpoint()

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
		Xmlns: eventNamespace,
	}

	var resp GetEventPropertiesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
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
func (c *Client) AddEventBroker(ctx context.Context, config *EventBrokerConfig) error {
	if config == nil {
		return ErrEventBrokerConfigNil
	}

	if config.Address == "" {
		return ErrInvalidEventBrokerAddress
	}

	endpoint := c.getEventEndpoint()

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
		Xmlns: eventNamespace,
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

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return fmt.Errorf("AddEventBroker failed: %w", err)
	}

	return nil
}

// DeleteEventBroker deletes an event broker configuration.
func (c *Client) DeleteEventBroker(ctx context.Context, address string) error {
	if address == "" {
		return ErrInvalidEventBrokerAddress
	}

	endpoint := c.getEventEndpoint()

	type DeleteEventBroker struct {
		XMLName xml.Name `xml:"tev:DeleteEventBroker"`
		Xmlns   string   `xml:"xmlns:tev,attr"`
		Address string   `xml:"tev:Address"`
	}

	type DeleteEventBrokerResponse struct {
		XMLName xml.Name `xml:"DeleteEventBrokerResponse"`
	}

	req := DeleteEventBroker{
		Xmlns:   eventNamespace,
		Address: address,
	}

	var resp DeleteEventBrokerResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return fmt.Errorf("DeleteEventBroker failed: %w", err)
	}

	return nil
}

// GetEventBrokers retrieves all event broker configurations.
func (c *Client) GetEventBrokers(ctx context.Context) ([]*EventBrokerConfig, error) {
	endpoint := c.getEventEndpoint()

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
		Xmlns: eventNamespace,
	}

	var resp GetEventBrokersResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
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
