package onvif

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testEventXMLHeader = `<?xml version="1.0" encoding="UTF-8"?>`

func newMockEventServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")

		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		bodyStr := string(body)

		var response string

		switch {
		case strings.Contains(bodyStr, "GetServiceCapabilities"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:GetServiceCapabilitiesResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
      <tev:Capabilities 
        WSSubscriptionPolicySupport="true"
        WSPausableSubscriptionManagerInterfaceSupport="true"
        MaxNotificationProducers="10"
        MaxPullPoints="5"
        PersistentNotificationStorage="true"
        EventBrokerProtocols="mqtt mqtts"
        MaxEventBrokers="3"
        MetadataOverMQTT="true"/>
    </tev:GetServiceCapabilitiesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "CreatePullPointSubscription"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:CreatePullPointSubscriptionResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
      <tev:SubscriptionReference>
        <wsa:Address xmlns:wsa="http://www.w3.org/2005/08/addressing">http://192.168.1.100/onvif/subscription/1</wsa:Address>
      </tev:SubscriptionReference>
      <tev:CurrentTime>2025-01-15T10:30:00Z</tev:CurrentTime>
      <tev:TerminationTime>2025-01-15T11:30:00Z</tev:TerminationTime>
    </tev:CreatePullPointSubscriptionResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "PullMessages"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:PullMessagesResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
      <tev:CurrentTime>2025-01-15T10:30:00Z</tev:CurrentTime>
      <tev:TerminationTime>2025-01-15T11:30:00Z</tev:TerminationTime>
      <wsnt:NotificationMessage xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
        <wsnt:Topic>tns1:VideoSource/MotionAlarm</wsnt:Topic>
        <wsnt:ProducerReference>
          <wsa:Address xmlns:wsa="http://www.w3.org/2005/08/addressing">http://192.168.1.100</wsa:Address>
        </wsnt:ProducerReference>
        <wsnt:Message PropertyOperation="Changed" UtcTime="2025-01-15T10:29:55Z">
          <tt:Source xmlns:tt="http://www.onvif.org/ver10/schema">
            <tt:SimpleItem Name="VideoSourceToken" Value="video_src_001"/>
          </tt:Source>
          <tt:Key xmlns:tt="http://www.onvif.org/ver10/schema">
            <tt:SimpleItem Name="RuleToken" Value="rule_001"/>
          </tt:Key>
          <tt:Data xmlns:tt="http://www.onvif.org/ver10/schema">
            <tt:SimpleItem Name="State" Value="true"/>
          </tt:Data>
        </wsnt:Message>
      </wsnt:NotificationMessage>
    </tev:PullMessagesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "Seek"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:SeekResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl"/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "SetSynchronizationPoint"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:SetSynchronizationPointResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl"/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "Unsubscribe"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <wsnt:UnsubscribeResponse xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "Renew"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <wsnt:RenewResponse xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
      <wsnt:CurrentTime>2025-01-15T10:30:00Z</wsnt:CurrentTime>
      <wsnt:TerminationTime>2025-01-15T12:30:00Z</wsnt:TerminationTime>
    </wsnt:RenewResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "GetEventProperties"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:GetEventPropertiesResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
      <tev:TopicNamespaceLocation>http://www.onvif.org/onvif/ver10/topics/topicns.xml</tev:TopicNamespaceLocation>
      <tev:FixedTopicSet>true</tev:FixedTopicSet>
      <tev:TopicExpressionDialect>http://www.onvif.org/ver10/tev/topicExpression/ConcreteSet</tev:TopicExpressionDialect>
      <tev:MessageContentFilterDialect>http://www.onvif.org/ver10/tev/messageContentFilter/ItemFilter</tev:MessageContentFilterDialect>
      <tev:ProducerPropertiesFilterDialect>http://www.onvif.org/ver10/tev/producerPropertiesFilter</tev:ProducerPropertiesFilterDialect>
      <tev:MessageContentSchemaLocation>http://www.onvif.org/onvif/ver10/schema/onvif.xsd</tev:MessageContentSchemaLocation>
    </tev:GetEventPropertiesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "AddEventBroker"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:AddEventBrokerResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl"/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "DeleteEventBroker"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:DeleteEventBrokerResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl"/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(bodyStr, "GetEventBrokers"):
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tev:GetEventBrokersResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
      <tev:EventBroker>
        <tev:Address>mqtt://broker.example.com:1883</tev:Address>
        <tev:TopicPrefix>onvif/</tev:TopicPrefix>
        <tev:UserName>mqtt_user</tev:UserName>
        <tev:QoS>1</tev:QoS>
        <tev:Status>Connected</tev:Status>
        <tev:CertPathValidation>true</tev:CertPathValidation>
      </tev:EventBroker>
    </tev:GetEventBrokersResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		default:
			response = testEventXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <SOAP-ENV:Fault>
      <SOAP-ENV:Code><SOAP-ENV:Value>SOAP-ENV:Receiver</SOAP-ENV:Value></SOAP-ENV:Code>
      <SOAP-ENV:Reason><SOAP-ENV:Text>Unknown action</SOAP-ENV:Text></SOAP-ENV:Reason>
    </SOAP-ENV:Fault>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
		}

		_, _ = w.Write([]byte(response))
	}))
}

func TestGetEventServiceCapabilities(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	caps, err := client.GetEventServiceCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetEventServiceCapabilities failed: %v", err)
	}

	if !caps.WSSubscriptionPolicySupport {
		t.Error("Expected WSSubscriptionPolicySupport to be true")
	}

	if !caps.WSPausableSubscriptionManagerInterfaceSupport {
		t.Error("Expected WSPausableSubscriptionManagerInterfaceSupport to be true")
	}

	if caps.MaxNotificationProducers != 10 {
		t.Errorf("Expected MaxNotificationProducers to be 10, got %d", caps.MaxNotificationProducers)
	}

	if caps.MaxPullPoints != 5 {
		t.Errorf("Expected MaxPullPoints to be 5, got %d", caps.MaxPullPoints)
	}

	if !caps.PersistentNotificationStorage {
		t.Error("Expected PersistentNotificationStorage to be true")
	}

	if len(caps.EventBrokerProtocols) != 2 {
		t.Errorf("Expected 2 EventBrokerProtocols, got %d", len(caps.EventBrokerProtocols))
	}

	if caps.MaxEventBrokers != 3 {
		t.Errorf("Expected MaxEventBrokers to be 3, got %d", caps.MaxEventBrokers)
	}

	if !caps.MetadataOverMQTT {
		t.Error("Expected MetadataOverMQTT to be true")
	}
}

func TestCreatePullPointSubscription(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with no filter and default termination time.
	sub, err := client.CreatePullPointSubscription(ctx, "", nil, "")
	if err != nil {
		t.Fatalf("CreatePullPointSubscription failed: %v", err)
	}

	if sub.SubscriptionReference == "" {
		t.Error("Expected SubscriptionReference to be set")
	}

	if sub.CurrentTime.IsZero() {
		t.Error("Expected CurrentTime to be set")
	}

	if sub.TerminationTime.IsZero() {
		t.Error("Expected TerminationTime to be set")
	}

	// Test with filter and termination time.
	termTime := 1 * time.Hour
	sub2, err := client.CreatePullPointSubscription(ctx, "tns1:VideoSource/MotionAlarm", &termTime, "policy1")
	if err != nil {
		t.Fatalf("CreatePullPointSubscription with filter failed: %v", err)
	}

	if sub2.SubscriptionReference == "" {
		t.Error("Expected SubscriptionReference to be set")
	}
}

func TestCreatePullPointSubscriptionInvalidTerminationTime(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with invalid (negative) termination time.
	invalidTime := -1 * time.Hour
	_, err = client.CreatePullPointSubscription(ctx, "", &invalidTime, "")
	if !errors.Is(err, ErrInvalidTerminationTime) {
		t.Errorf("Expected ErrInvalidTerminationTime, got %v", err)
	}
}

func TestPullMessages(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	messages, err := client.PullMessages(ctx, server.URL+"/subscription/1", 30*time.Second, 10)
	if err != nil {
		t.Fatalf("PullMessages failed: %v", err)
	}

	if len(messages) == 0 {
		t.Error("Expected at least one notification message")
	}

	if len(messages) > 0 {
		msg := messages[0]
		if msg.Topic == "" {
			t.Error("Expected Topic to be set")
		}

		if msg.Message.PropertyOperation == "" {
			t.Error("Expected PropertyOperation to be set")
		}

		if len(msg.Message.Source) == 0 {
			t.Error("Expected Source items to be present")
		}

		if len(msg.Message.Data) == 0 {
			t.Error("Expected Data items to be present")
		}
	}
}

func TestPullMessagesValidation(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test empty subscription reference.
	_, err = client.PullMessages(ctx, "", 30*time.Second, 10)
	if !errors.Is(err, ErrInvalidSubscriptionReference) {
		t.Errorf("Expected ErrInvalidSubscriptionReference, got %v", err)
	}

	// Test invalid timeout.
	_, err = client.PullMessages(ctx, server.URL+"/subscription/1", 0, 10)
	if !errors.Is(err, ErrInvalidTimeout) {
		t.Errorf("Expected ErrInvalidTimeout, got %v", err)
	}

	// Test invalid message limit.
	_, err = client.PullMessages(ctx, server.URL+"/subscription/1", 30*time.Second, 0)
	if !errors.Is(err, ErrInvalidMessageLimit) {
		t.Errorf("Expected ErrInvalidMessageLimit, got %v", err)
	}
}

func TestSeek(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.Seek(ctx, server.URL+"/subscription/1", time.Now().Add(-1*time.Hour), false)
	if err != nil {
		t.Fatalf("Seek failed: %v", err)
	}

	// Test with reverse.
	err = client.Seek(ctx, server.URL+"/subscription/1", time.Now().Add(-1*time.Hour), true)
	if err != nil {
		t.Fatalf("Seek with reverse failed: %v", err)
	}
}

func TestSeekInvalidSubscriptionReference(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.Seek(ctx, "", time.Now(), false)
	if !errors.Is(err, ErrInvalidSubscriptionReference) {
		t.Errorf("Expected ErrInvalidSubscriptionReference, got %v", err)
	}
}

func TestSetEventSynchronizationPoint(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.SetEventSynchronizationPoint(ctx, server.URL+"/subscription/1")
	if err != nil {
		t.Fatalf("SetEventSynchronizationPoint failed: %v", err)
	}
}

func TestSetEventSynchronizationPointInvalidSubscriptionReference(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.SetEventSynchronizationPoint(ctx, "")
	if !errors.Is(err, ErrInvalidSubscriptionReference) {
		t.Errorf("Expected ErrInvalidSubscriptionReference, got %v", err)
	}
}

func TestUnsubscribe(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.Unsubscribe(ctx, server.URL+"/subscription/1")
	if err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}
}

func TestUnsubscribeInvalidSubscriptionReference(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.Unsubscribe(ctx, "")
	if !errors.Is(err, ErrInvalidSubscriptionReference) {
		t.Errorf("Expected ErrInvalidSubscriptionReference, got %v", err)
	}
}

func TestRenewSubscription(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	currentTime, terminationTime, err := client.RenewSubscription(ctx, server.URL+"/subscription/1", 2*time.Hour)
	if err != nil {
		t.Fatalf("RenewSubscription failed: %v", err)
	}

	if currentTime.IsZero() {
		t.Error("Expected CurrentTime to be set")
	}

	if terminationTime.IsZero() {
		t.Error("Expected TerminationTime to be set")
	}
}

func TestRenewSubscriptionValidation(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test empty subscription reference.
	_, _, err = client.RenewSubscription(ctx, "", time.Hour)
	if !errors.Is(err, ErrInvalidSubscriptionReference) {
		t.Errorf("Expected ErrInvalidSubscriptionReference, got %v", err)
	}

	// Test invalid termination time.
	_, _, err = client.RenewSubscription(ctx, server.URL+"/subscription/1", 0)
	if !errors.Is(err, ErrInvalidTerminationTime) {
		t.Errorf("Expected ErrInvalidTerminationTime, got %v", err)
	}
}

func TestGetEventProperties(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	props, err := client.GetEventProperties(ctx)
	if err != nil {
		t.Fatalf("GetEventProperties failed: %v", err)
	}

	if len(props.TopicNamespaceLocation) == 0 {
		t.Error("Expected TopicNamespaceLocation to be set")
	}

	if !props.FixedTopicSet {
		t.Error("Expected FixedTopicSet to be true")
	}

	if len(props.TopicExpressionDialects) == 0 {
		t.Error("Expected TopicExpressionDialects to be set")
	}

	if len(props.MessageContentFilterDialects) == 0 {
		t.Error("Expected MessageContentFilterDialects to be set")
	}
}

func TestAddEventBroker(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	config := &EventBrokerConfig{
		Address:     "mqtt://broker.example.com:1883",
		TopicPrefix: "onvif/",
		UserName:    "mqtt_user",
		Password:    "mqtt_pass",
		QoS:         1,
	}

	err = client.AddEventBroker(ctx, config)
	if err != nil {
		t.Fatalf("AddEventBroker failed: %v", err)
	}
}

func TestAddEventBrokerValidation(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test nil config.
	err = client.AddEventBroker(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test empty address.
	config := &EventBrokerConfig{Address: ""}
	err = client.AddEventBroker(ctx, config)
	if !errors.Is(err, ErrInvalidEventBrokerAddress) {
		t.Errorf("Expected ErrInvalidEventBrokerAddress, got %v", err)
	}
}

func TestDeleteEventBroker(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.DeleteEventBroker(ctx, "mqtt://broker.example.com:1883")
	if err != nil {
		t.Fatalf("DeleteEventBroker failed: %v", err)
	}
}

func TestDeleteEventBrokerInvalidAddress(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	err = client.DeleteEventBroker(ctx, "")
	if !errors.Is(err, ErrInvalidEventBrokerAddress) {
		t.Errorf("Expected ErrInvalidEventBrokerAddress, got %v", err)
	}
}

func TestGetEventBrokers(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	brokers, err := client.GetEventBrokers(ctx)
	if err != nil {
		t.Fatalf("GetEventBrokers failed: %v", err)
	}

	if len(brokers) == 0 {
		t.Error("Expected at least one event broker")
	}

	if len(brokers) > 0 {
		broker := brokers[0]
		if broker.Address == "" {
			t.Error("Expected Address to be set")
		}

		if broker.TopicPrefix == "" {
			t.Error("Expected TopicPrefix to be set")
		}

		if broker.Status == "" {
			t.Error("Expected Status to be set")
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "PT30S"},
		{60 * time.Second, "PT1M"},
		{90 * time.Second, "PT1M30S"},
		{5 * time.Minute, "PT5M"},
		{65 * time.Second, "PT1M5S"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

func TestSplitSpaceSeparated(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"mqtt", []string{"mqtt"}},
		{"mqtt mqtts", []string{"mqtt", "mqtts"}},
		{"  mqtt   mqtts  ", []string{"mqtt", "mqtts"}},
		{"a b c", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		result := splitSpaceSeparated(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitSpaceSeparated(%q) returned %d items, expected %d", tt.input, len(result), len(tt.expected))

			continue
		}

		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("splitSpaceSeparated(%q)[%d] = %q, expected %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestSetEventEndpoint(t *testing.T) {
	server := newMockEventServer()
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	newEndpoint := "http://192.168.1.100/onvif/events"
	client.SetEventEndpoint(newEndpoint)

	// Verify endpoint was set.
	endpoint := client.getEventEndpoint()
	if endpoint != newEndpoint {
		t.Errorf("Expected event endpoint %s, got %s", newEndpoint, endpoint)
	}
}
