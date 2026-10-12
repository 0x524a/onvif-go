package events

import (
	"testing"
	"time"
)

// The dialect vectors below are field captures (vendor prefix dialects,
// wrapped envelopes, literal Name/Value items, heartbeats) from the MiBeeNvr
// reference implementation this parser was ported from — they pin the
// tolerance the original was hardened against.

const notifyWrappedDoubleLayer = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"><soap:Body><?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"><soap:Body>
<wsnt:Notify xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2" xmlns:wsa="http://www.w3.org/2005/08/addressing" xmlns:tt="http://www.onvif.org/ver10/schema">
 <wsnt:NotificationMessage>
  <wsnt:SubscriptionReference><wsa:Address>http://192.0.2.10:8080/onvif/events_service/sub/abc</wsa:Address></wsnt:SubscriptionReference>
  <wsnt:Topic Dialect="http://www.onvif.org/ver10/tev/topicExpression/Concrete">tns1:VideoSource/MotionAlarm</wsnt:Topic>
  <wsnt:Message>
   <tt:Message PropertyOperation="Changed" UtcTime="2026-10-01T07:46:12.123Z">
    <tt:Source><tt:SimpleItem Name="Source" Value="0"/></tt:Source>
    <tt:Key></tt:Key>
    <tt:Data>
     <tt:SimpleItem Name="State" Value="true"/>
     <tt:SimpleItem Name="Targets" Value="person"/>
    </tt:Data>
   </tt:Message>
  </wsnt:Message>
 </wsnt:NotificationMessage>
</wsnt:Notify>
</soap:Body></soap:Envelope>
</soap:Body></soap:Envelope>`

func TestParseNotifyEventsWrappedEnvelopeDoubleLayer(t *testing.T) {
	t.Parallel()
	// The rs-device habit: the whole response envelope wrapped inside
	// another envelope, with the canonical double-layer payload inside.
	msgs, err := ParseNotifyEvents([]byte(notifyWrappedDoubleLayer))
	if err != nil {
		t.Fatalf("ParseNotifyEvents: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("want 1 message, got %d", len(msgs))
	}
	m := msgs[0]
	if m.Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("topic: got %q", m.Topic)
	}
	want := time.Date(2026, 10, 1, 7, 46, 12, 123000000, time.UTC)
	if !m.Message.UtcTime.Equal(want) {
		t.Errorf("inner @UtcTime: got %v, want %v", m.Message.UtcTime, want)
	}
	if m.Message.PropertyOperation != "Changed" {
		t.Errorf("PropertyOperation: got %q", m.Message.PropertyOperation)
	}
	if len(m.Message.Source) != 1 || m.Message.Source[0].Value != "0" {
		t.Errorf("Source items: %+v", m.Message.Source)
	}
	if len(m.Message.Data) != 2 || m.Message.Data[0].Name != "State" || m.Message.Data[0].Value != "true" || m.Message.Data[1].Value != "person" {
		t.Errorf("Data items: %+v", m.Message.Data)
	}
}

func TestParseNotifyEventsLiteralNameValuePairs(t *testing.T) {
	t.Parallel()
	// rs-style dialect: different prefixes and literal Name/Value child
	// pairs instead of SimpleItem attributes.
	body := `<?xml version="1.0"?>
<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope"><env:Body>
<nt:Notify xmlns:nt="http://docs.oasis-open.org/wsn/b-2">
 <nt:NotificationMessage>
  <nt:Topic>tns1:VideoSource/MotionAlarm</nt:Topic>
  <nt:Message UtcTime="2026-09-30T08:00:00.5Z">
   <onvif:Source xmlns:onvif="http://www.onvif.org/ver10/schema"><onvif:Name>Source</onvif:Name><onvif:Value>CSI</onvif:Value></onvif:Source>
   <onvif:Data xmlns:onvif="http://www.onvif.org/ver10/schema"><onvif:Name>State</onvif:Name><onvif:Value>true</onvif:Value></onvif:Data>
  </nt:Message>
 </nt:NotificationMessage>
</nt:Notify>
</env:Body></env:Envelope>`
	msgs, err := ParseNotifyEvents([]byte(body))
	if err != nil {
		t.Fatalf("ParseNotifyEvents: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("want 1 message, got %d", len(msgs))
	}
	if msgs[0].Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("topic: got %q", msgs[0].Topic)
	}
	if len(msgs[0].Message.Source) != 1 || msgs[0].Message.Source[0].Name != "Source" || msgs[0].Message.Source[0].Value != "CSI" {
		t.Errorf("Source items: %+v", msgs[0].Message.Source)
	}
	if len(msgs[0].Message.Data) != 1 || msgs[0].Message.Data[0].Name != "State" || msgs[0].Message.Data[0].Value != "true" {
		t.Errorf("Data items: %+v", msgs[0].Message.Data)
	}
}

func TestParseNotifyEventsEmptyHeartbeatIsValid(t *testing.T) {
	t.Parallel()
	// Devices heartbeat empty Notifies — zero messages is a valid delivery.
	msgs, err := ParseNotifyEvents([]byte(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><wsnt:Notify xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"/></s:Body></s:Envelope>`))
	if err != nil {
		t.Fatalf("empty Notify must be valid, got %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("want 0 messages, got %d", len(msgs))
	}
}

func TestParseNotifyEventsFaultSurfaces(t *testing.T) {
	t.Parallel()
	if _, err := ParseNotifyEvents([]byte(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><s:Fault><s:Code><s:Value>s:Sender</s:Value></s:Code></s:Fault></s:Body></s:Envelope>`)); err == nil {
		t.Fatal("faults inside Notify must surface as errors")
	}
}
