package events

// WS-BaseNotification push (issue #127): Subscribe/Renew/Unsubscribe on the
// subscription's own manager endpoint, the wsnt:Notify parser (two-layer
// Message, aligned with the PullMessages decode of #82), the consumer
// http.Handler, and the managed auto-renew loop. Wire goldens derive from
// the MiBeeNvr production envelopes field-proven against Hikvision,
// Uniview, TP-Link and the in-house firmware.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/soap"
	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

const subscribeResponseXML = `<SubscribeResponse xmlns="http://docs.oasis-open.org/wsn/b-2">
<SubscriptionReference><wsa:Address xmlns:wsa="http://www.w3.org/2005/08/addressing">http://192.168.1.64:80/onvif/subscription?idx=7</wsa:Address></SubscriptionReference>
<SubscriptionManager><wsa:Address xmlns:wsa="http://www.w3.org/2005/08/addressing">http://192.168.1.64:80/onvif/subscriptionmanager?idx=7</wsa:Address></SubscriptionManager>
<CurrentTime>2026-10-12T08:00:00Z</CurrentTime>
<TerminationTime>2026-10-12T09:00:00Z</TerminationTime>
</SubscribeResponse>`

func TestSubscribeWireGolden(t *testing.T) {
	svc, caller := newPushService(t, func(action, _ string) (string, error) {
		if action != "wsnt:Subscribe" {
			return "", errors.New("unexpected action " + action)
		}

		return subscribeResponseXML, nil
	})

	sub, err := svc.Subscribe(context.Background(), "http://192.168.1.10:8080/api/onvif/notify", time.Hour)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	reqs := caller.Requests()
	if len(reqs) != 1 {
		t.Fatalf("requests = %d, want 1", len(reqs))
	}

	want := `<wsnt:Subscribe xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2" ` +
		`xmlns:wsa="http://www.w3.org/2005/08/addressing">` +
		`<wsnt:ConsumerReference><wsa:Address>http://192.168.1.10:8080/api/onvif/notify</wsa:Address></wsnt:ConsumerReference>` +
		`<wsnt:InitialTerminationTime>PT60M</wsnt:InitialTerminationTime>` +
		`</wsnt:Subscribe>`
	if reqs[0].Body != want {
		t.Errorf("Subscribe body:\n got %s\nwant %s", reqs[0].Body, want)
	}

	if sub.SubscriptionReference != "http://192.168.1.64:80/onvif/subscription?idx=7" {
		t.Errorf("SubscriptionReference = %q", sub.SubscriptionReference)
	}

	if sub.ManagerAddress != "http://192.168.1.64:80/onvif/subscriptionmanager?idx=7" {
		t.Errorf("ManagerAddress = %q", sub.ManagerAddress)
	}

	if want := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC); !sub.TerminationTime.Equal(want) {
		t.Errorf("TerminationTime = %v, want %v", sub.TerminationTime, want)
	}
}

func TestSubscribeManagerFallsBackToSubscriptionReference(t *testing.T) {
	sameAddr := strings.Replace(subscribeResponseXML,
		"http://192.168.1.64:80/onvif/subscriptionmanager?idx=7",
		"http://192.168.1.64:80/onvif/subscription?idx=7", 1)

	svc, _ := newPushService(t, func(action, _ string) (string, error) {
		return sameAddr, nil
	})

	sub, err := svc.Subscribe(context.Background(), "http://nvr/notify", time.Hour)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	if sub.ManagerAddress != sub.SubscriptionReference {
		t.Errorf("ManagerAddress = %q, want the subscription reference (same-address devices)", sub.ManagerAddress)
	}
}

func TestSubscribeParsesTerminationTimeLayouts(t *testing.T) {
	cases := map[string]string{
		"RFC3339Nano":     `<TerminationTime>2026-10-12T09:00:00.5Z</TerminationTime>`,
		"bare datetime":   `<TerminationTime>2026-10-12T09:00:00</TerminationTime>`,
		"fractional no Z": `<TerminationTime>2026-10-12T09:00:00.25</TerminationTime>`,
		"RFC3339 offset":  `<TerminationTime>2026-10-12T17:00:00+08:00</TerminationTime>`,
	}

	for name, term := range cases {
		t.Run(name, func(t *testing.T) {
			body := strings.Replace(subscribeResponseXML,
				"<TerminationTime>2026-10-12T09:00:00Z</TerminationTime>", term, 1)

			svc, _ := newPushService(t, func(_, _ string) (string, error) { return body, nil })

			sub, err := svc.Subscribe(context.Background(), "http://nvr/notify", time.Hour)
			if err != nil {
				t.Fatalf("Subscribe: %v", err)
			}

			if sub.TerminationTime.IsZero() {
				t.Fatalf("TerminationTime not parsed from %q", term)
			}

			// Fractional layouts keep their sub-second precision; all
			// fixtures denote the same 09:00 UTC instant otherwise.
			var want time.Time
			switch name {
			case "RFC3339Nano":
				want = time.Date(2026, 10, 12, 9, 0, 0, 500000000, time.UTC)
			case "fractional no Z":
				want = time.Date(2026, 10, 12, 9, 0, 0, 250000000, time.UTC)
			default:
				want = time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
			}

			if !sub.TerminationTime.Equal(want) {
				t.Errorf("TerminationTime = %v, want %v", sub.TerminationTime, want)
			}
		})
	}
}

func TestSubscribeValidatesArguments(t *testing.T) {
	svc, _ := newPushService(t, func(_, _ string) (string, error) {
		t.Error("no wire traffic expected")
		return "", nil
	})

	if _, err := svc.Subscribe(context.Background(), "", time.Hour); err == nil {
		t.Error("empty consumer reference accepted")
	}

	if _, err := svc.Subscribe(context.Background(), "http://nvr/notify", 0); err == nil {
		t.Error("zero termination accepted")
	}
}

func TestSubscribeSenderFaultMapsToPushNotSupported(t *testing.T) {
	cases := []struct {
		name            string
		fault           *soap.FaultError
		wantUnsupported bool
	}{
		{
			name:            "SOAP 1.2 Sender",
			fault:           &soap.FaultError{Code: "s:Sender", Reason: "FilteringRequestedNotSupported"},
			wantUnsupported: true,
		},
		{
			name:            "SOAP 1.1 sender",
			fault:           &soap.FaultError{Code: "SOAP-ENV:Client", Subcode: "Sender"},
			wantUnsupported: true,
		},
		{
			name:            "Receiver fault is a real failure",
			fault:           &soap.FaultError{Code: "s:Receiver", Reason: "internal"},
			wantUnsupported: false,
		},
		{
			name:            "NotAuthorized is not push-unsupported",
			fault:           &soap.FaultError{Code: "ter:NotAuthorized"},
			wantUnsupported: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newPushService(t, func(_, _ string) (string, error) {
				return "", tc.fault
			})

			_, err := svc.Subscribe(context.Background(), "http://nvr/notify", time.Hour)

			if err == nil {
				t.Fatal("Subscribe must surface the fault")
			}

			if got := errors.Is(err, ErrPushNotSupported); got != tc.wantUnsupported {
				t.Errorf("errors.Is(ErrPushNotSupported) = %v, want %v (err: %v)", got, tc.wantUnsupported, err)
			}
		})
	}
}

func TestRenewPushTargetsManagerAddress(t *testing.T) {
	svc, caller := newPushService(t, func(action, _ string) (string, error) {
		if action != "wsnt:Renew" {
			return "", errors.New("unexpected action " + action)
		}

		return `<RenewResponse xmlns="http://docs.oasis-open.org/wsn/b-2"><TerminationTime>2026-10-12T10:00:00Z</TerminationTime></RenewResponse>`, nil
	})

	const manager = "http://192.168.1.64:80/onvif/subscriptionmanager?idx=7"

	term, err := svc.RenewPush(context.Background(), manager, time.Hour)
	if err != nil {
		t.Fatalf("RenewPush: %v", err)
	}

	reqs := caller.Requests()
	if len(reqs) != 1 {
		t.Fatalf("requests = %d, want 1", len(reqs))
	}

	if reqs[0].Target != manager {
		t.Errorf("Renew posted to %q, want the subscription manager address", reqs[0].Target)
	}

	want := `<wsnt:Renew xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">` +
		`<wsnt:TerminationTime>PT60M</wsnt:TerminationTime></wsnt:Renew>`
	if reqs[0].Body != want {
		t.Errorf("Renew body:\n got %s\nwant %s", reqs[0].Body, want)
	}

	if want := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC); !term.Equal(want) {
		t.Errorf("termination = %v, want %v", term, want)
	}
}

func TestUnsubscribePushTargetsManagerAddress(t *testing.T) {
	svc, caller := newPushService(t, func(action, _ string) (string, error) {
		if action != "wsnt:Unsubscribe" {
			return "", errors.New("unexpected action " + action)
		}

		return `<UnsubscribeResponse xmlns="http://docs.oasis-open.org/wsn/b-2"/>`, nil
	})

	const manager = "http://192.168.1.64:80/onvif/subscriptionmanager?idx=7"

	if err := svc.UnsubscribePush(context.Background(), manager); err != nil {
		t.Fatalf("UnsubscribePush: %v", err)
	}

	reqs := caller.Requests()
	if len(reqs) != 1 {
		t.Fatalf("requests = %d, want 1", len(reqs))
	}

	if reqs[0].Target != manager {
		t.Errorf("Unsubscribe posted to %q, want the subscription manager address", reqs[0].Target)
	}

	want := `<wsnt:Unsubscribe xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"></wsnt:Unsubscribe>`
	if reqs[0].Body != want {
		t.Errorf("Unsubscribe body:\n got %s\nwant %s", reqs[0].Body, want)
	}
}

// notifyEnvelopeGolden is a dual-message wsnt:Notify POST in the canonical
// two-layer Message form (wsnt:Message wrapping tt:Message), the shape
// Hikvision-class devices deliver.
const notifyEnvelopeGolden = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<wsnt:Notify xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
<wsnt:NotificationMessage>
<wsnt:Topic Dialect="http://www.onvif.org/ver10/tev/topicExpression/ConcreteSet">tns1:VideoSource/MotionAlarm</wsnt:Topic>
<wsnt:Message><tt:Message xmlns:tt="http://www.onvif.org/ver10/schema" UtcTime="2026-10-12T08:00:01Z" PropertyOperation="Changed">
<tt:Source><tt:SimpleItem Name="Source" Value="VideoSource_1"/></tt:Source>
<tt:Data><tt:SimpleItem Name="State" Value="true"/><tt:SimpleItem Name="Score" Value="87"/></tt:Data>
</tt:Message></wsnt:Message>
</wsnt:NotificationMessage>
<wsnt:NotificationMessage>
<wsnt:Topic Dialect="http://www.onvif.org/ver10/tev/topicExpression/ConcreteSet">tns1:RuleEngine/CellMotionDetector/Motion</wsnt:Topic>
<wsnt:Message><tt:Message xmlns:tt="http://www.onvif.org/ver10/schema" UtcTime="2026-10-12T08:00:02Z" PropertyOperation="Changed">
<tt:Source><tt:SimpleItem Name="VideoSourceConfigurationToken" Value="VideoSourceToken"/></tt:Source>
<tt:Data><tt:SimpleItem Name="IsMotion" Value="false"/></tt:Data>
</tt:Message></wsnt:Message>
</wsnt:NotificationMessage>
</wsnt:Notify>
</s:Body>
</s:Envelope>`

// notifySingleLayerGolden is the lax single-layer form: attributes and
// SimpleItems directly on the outer Message element (issue #82's second
// wire shape).
const notifySingleLayerGolden = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<Notify xmlns="http://docs.oasis-open.org/wsn/b-2">
<NotificationMessage>
<Topic>tns1:Device/Humidity</Topic>
<Message UtcTime="2026-10-12T08:00:03Z" PropertyOperation="Changed">
<Data><SimpleItem Name="Value" Value="62"/></Data>
</Message>
</NotificationMessage>
</Notify>
</s:Body>
</s:Envelope>`

func TestParseNotifyEventsTwoLayer(t *testing.T) {
	msgs, err := ParseNotifyEvents([]byte(notifyEnvelopeGolden))
	if err != nil {
		t.Fatalf("ParseNotifyEvents: %v", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("messages = %d, want 2", len(msgs))
	}

	first := msgs[0]
	if first.Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("topic = %q", first.Topic)
	}

	if first.Message.PropertyOperation != "Changed" {
		t.Errorf("PropertyOperation = %q", first.Message.PropertyOperation)
	}

	if want := time.Date(2026, 10, 12, 8, 0, 1, 0, time.UTC); !first.Message.UtcTime.Equal(want) {
		t.Errorf("UtcTime = %v, want %v", first.Message.UtcTime, want)
	}

	if len(first.Message.Source) != 1 || first.Message.Source[0].Value != "VideoSource_1" {
		t.Errorf("Source = %+v", first.Message.Source)
	}

	if len(first.Message.Data) != 2 || first.Message.Data[0].Name != "State" || first.Message.Data[0].Value != "true" {
		t.Errorf("Data = %+v", first.Message.Data)
	}

	if msgs[1].Topic != "tns1:RuleEngine/CellMotionDetector/Motion" {
		t.Errorf("second topic = %q", msgs[1].Topic)
	}
}

func TestParseNotifyEventsSingleLayerTolerated(t *testing.T) {
	msgs, err := ParseNotifyEvents([]byte(notifySingleLayerGolden))
	if err != nil {
		t.Fatalf("ParseNotifyEvents: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("messages = %d, want 1", len(msgs))
	}

	if msgs[0].Topic != "tns1:Device/Humidity" {
		t.Errorf("topic = %q", msgs[0].Topic)
	}

	if len(msgs[0].Message.Data) != 1 || msgs[0].Message.Data[0].Value != "62" {
		t.Errorf("Data = %+v", msgs[0].Message.Data)
	}
}

// TestParseNotifyEventsMatchesPullMessagesDecode pins the "one parser"
// contract: the same NotificationMessage fragment decodes identically
// through the push Notify envelope and the PullMessagesResponse path.
func TestParseNotifyEventsMatchesPullMessagesDecode(t *testing.T) {
	const fragment = `<NotificationMessage>
<Topic>tns1:VideoSource/MotionAlarm</Topic>
<Message><Message UtcTime="2026-10-12T08:00:01Z" PropertyOperation="Changed">
<Source><SimpleItem Name="Source" Value="VideoSource_1"/></Source>
<Data><SimpleItem Name="State" Value="true"/></Data>
</Message></Message>
</NotificationMessage>`

	notifyBody := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>` +
		`<Notify xmlns="http://docs.oasis-open.org/wsn/b-2">` + fragment + `</Notify></s:Body></s:Envelope>`
	pullBody := `<PullMessagesResponse xmlns="http://www.onvif.org/ver10/events/wsdl">` +
		`<CurrentTime>2026-10-12T08:00:00Z</CurrentTime><TerminationTime>2026-10-12T08:01:00Z</TerminationTime>` +
		fragment + `</PullMessagesResponse>`

	pushMsgs, err := ParseNotifyEvents([]byte(notifyBody))
	if err != nil {
		t.Fatalf("ParseNotifyEvents: %v", err)
	}

	pullSvc, _ := newPushService(t, func(action, reqXML string) (string, error) {
		return pullBody, nil
	})
	pullMsgs, err := pullSvc.PullMessages(context.Background(), "sub-ref", time.Second, 5)
	if err != nil {
		t.Fatalf("PullMessages: %v", err)
	}

	if len(pushMsgs) != 1 || len(pullMsgs) != 1 {
		t.Fatalf("push=%d pull=%d messages, want 1 each", len(pushMsgs), len(pullMsgs))
	}

	a, b := pushMsgs[0], pullMsgs[0]

	if a.Topic != b.Topic || a.Message.PropertyOperation != b.Message.PropertyOperation ||
		!a.Message.UtcTime.Equal(b.Message.UtcTime) {
		t.Errorf("decode mismatch: push %+v vs pull %+v", a, b)
	}

	if len(a.Message.Data) != len(b.Message.Data) || a.Message.Data[0] != b.Message.Data[0] {
		t.Errorf("data mismatch: push %+v vs pull %+v", a.Message.Data, b.Message.Data)
	}
}

func TestParseNotifyEventsRejectsGarbage(t *testing.T) {
	for _, body := range []string{"", "not xml", "<s:Envelope></s:Envelope>"} {
		if _, err := ParseNotifyEvents([]byte(body)); err == nil {
			t.Errorf("ParseNotifyEvents(%q) accepted garbage", body)
		}
	}
}

func TestNotifyHandlerServesDevicePosts(t *testing.T) {
	var mu sync.Mutex
	var delivered [][]NotificationMessage

	h := NotifyHandler(func(msgs []NotificationMessage) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, msgs)
	})

	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL, "application/soap+xml", strings.NewReader(notifyEnvelopeGolden))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(delivered) != 1 || len(delivered[0]) != 2 {
		t.Fatalf("delivered = %+v, want one batch of two messages", delivered)
	}
}

func TestNotifyHandlerRejectsBadRequests(t *testing.T) {
	h := NotifyHandler(func([]NotificationMessage) {
		t.Error("no delivery expected")
	})

	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET status = %d, want 405", resp.StatusCode)
	}

	resp, err = http.Post(srv.URL, "application/soap+xml", strings.NewReader("garbage"))
	if err != nil {
		t.Fatalf("POST garbage: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("garbage status = %d, want 400", resp.StatusCode)
	}

	oversize := strings.NewReader("x" + strings.Repeat("0", maxNotifyBodyBytes))
	resp, err = http.Post(srv.URL, "application/soap+xml", oversize)
	if err != nil {
		t.Fatalf("POST oversize: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("oversize status = %d, want 413", resp.StatusCode)
	}
}

func TestSubscribePushEventsRenewsAndStops(t *testing.T) {
	var mu sync.Mutex
	var renews, unsubscribes int

	svc, _ := newPushService(t, func(action, _ string) (string, error) {
		switch action {
		case "wsnt:Subscribe":
			return `<SubscribeResponse xmlns="http://docs.oasis-open.org/wsn/b-2">` +
				`<SubscriptionReference><Address>http://dev/sub</Address></SubscriptionReference>` +
				`<TerminationTime>` + time.Now().Add(200*time.Millisecond).UTC().Format(time.RFC3339) + `</TerminationTime>` +
				`</SubscribeResponse>`, nil
		case "wsnt:Renew":
			mu.Lock()
			renews++
			mu.Unlock()

			return `<RenewResponse xmlns="http://docs.oasis-open.org/wsn/b-2">` +
				`<TerminationTime>` + time.Now().Add(200*time.Millisecond).UTC().Format(time.RFC3339) + `</TerminationTime>` +
				`</RenewResponse>`, nil
		case "wsnt:Unsubscribe":
			mu.Lock()
			unsubscribes++
			mu.Unlock()

			return `<UnsubscribeResponse xmlns="http://docs.oasis-open.org/wsn/b-2"/>`, nil
		default:
			return "", fmt.Errorf("unexpected action %s", action)
		}
	})

	stream, err := svc.SubscribePushEvents(context.Background(), "http://nvr/notify", &SubscribePushOptions{
		Duration:   200 * time.Millisecond,
		RenewFloor: 20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("SubscribePushEvents: %v", err)
	}

	// Bounded wait: at least two renewals must happen within the deadline.
	deadline := time.After(5 * time.Second)
	for {
		mu.Lock()
		r := renews
		mu.Unlock()

		if r >= 2 {
			break
		}

		select {
		case <-deadline:
			t.Fatalf("only %d renewals fired, want >= 2", r)
		case <-stream.Done():
			t.Fatal("stream ended before two renewals")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	if err := stream.Unsubscribe(context.Background()); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}

	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done did not close after Unsubscribe")
	}

	mu.Lock()
	defer mu.Unlock()

	if unsubscribes != 1 {
		t.Errorf("device Unsubscribe sent %d times, want 1", unsubscribes)
	}
}

func TestSubscribePushEventsSenderFaultFailsFast(t *testing.T) {
	svc, _ := newPushService(t, func(_, _ string) (string, error) {
		return "", &soap.FaultError{Code: "s:Sender", Reason: "Subscribe not supported"}
	})

	_, err := svc.SubscribePushEvents(context.Background(), "http://nvr/notify", nil)
	if !errors.Is(err, ErrPushNotSupported) {
		t.Fatalf("err = %v, want ErrPushNotSupported", err)
	}
}

func TestSubscribePushEventsRequiresConsumerURL(t *testing.T) {
	svc, _ := newPushService(t, func(_, _ string) (string, error) {
		t.Error("no wire traffic expected")
		return "", nil
	})

	if _, err := svc.SubscribePushEvents(context.Background(), "", nil); err == nil {
		t.Fatal("empty consumer URL accepted")
	}
}

// newPushService wires an events Service over a recording FakeCaller.
func newPushService(t *testing.T, handler func(action, reqXML string) (string, error)) (*Service, *testutil.FakeCaller) {
	t.Helper()

	caller := testutil.NewFakeCaller("http://fake/events_service", handler)

	return New(caller), caller
}

func TestParseNotifyEventsBareNotifyRoot(t *testing.T) {
	// Some firmwares POST the Notify element itself as the document root,
	// without a SOAP envelope wrapper.
	body := `<wsnt:Notify xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
<wsnt:NotificationMessage><wsnt:Topic>tns1:Device/Humidity</wsnt:Topic></wsnt:NotificationMessage>
</wsnt:Notify>`

	msgs, err := ParseNotifyEvents([]byte(body))
	if err != nil {
		t.Fatalf("ParseNotifyEvents: %v", err)
	}

	if len(msgs) != 1 || msgs[0].Topic != "tns1:Device/Humidity" {
		t.Fatalf("messages = %+v, want the bare-root humidity message", msgs)
	}
}
