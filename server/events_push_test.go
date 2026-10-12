package server

// Basic-notification interface (wsnt:Subscribe + push Notify) — the
// scenarios mirror the Rust twin's events_subscribe.rs: subscribe over
// the wire, Notify POSTs arrive at a real loopback consumer, Renew /
// Unsubscribe share the pull-point lifecycle, PullMessages refuses push
// subscriptions, and a dead consumer is auto-unsubscribed after the
// failure budget.

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// pushConsumer is a wsnt consumer on an ephemeral loopback port: it
// captures every Notify POST body and answers `status`.
type pushConsumer struct {
	server *httptest.Server

	mu     sync.Mutex
	bodies []string
}

func newPushConsumer(t *testing.T, status int) *pushConsumer {
	t.Helper()

	c := &pushConsumer{}
	c.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		c.mu.Lock()
		c.bodies = append(c.bodies, string(body))
		c.mu.Unlock()
		w.WriteHeader(status)
	}))
	t.Cleanup(c.server.Close)

	return c
}

func (c *pushConsumer) captured() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]string(nil), c.bodies...)
}

// subscribe issues a wsnt:Subscribe for the consumer URL with the given
// termination and returns (status, body, subscriptionPath).
func subscribe(t *testing.T, mux *http.ServeMux, consumerURL, termination string) (int, string, string) {
	t.Helper()

	body := `<Subscribe xmlns="http://docs.oasis-open.org/wsn/b-2">` +
		`<ConsumerReference><wsa:Address xmlns:wsa="http://www.w3.org/2005/08/addressing">` + consumerURL + `</wsa:Address></ConsumerReference>` +
		`<InitialTerminationTime>` + termination + `</InitialTerminationTime>` +
		`</Subscribe>`

	status, resp := postEventsSOAP(t, mux, "/onvif/events_service", body)

	subPath := ""
	if addr := xmlFieldText(resp, "Address"); addr != "" {
		if i := strings.Index(addr, "/onvif/events_service/sub/"); i >= 0 {
			subPath = addr[i:]
		}
	}

	return status, resp, subPath
}

// xmlFieldText extracts the text of the first element whose local name
// matches (prefix-agnostic).
func xmlFieldText(xmlBody, tag string) string {
	decoder := xml.NewDecoder(strings.NewReader(xmlBody))
	for {
		tok, err := decoder.Token()
		if err != nil {
			return ""
		}

		if start, ok := tok.(xml.StartElement); ok && start.Name.Local == tag {
			var text string
			if err := decoder.DecodeElement(&text, &start); err == nil {
				return text
			}
		}
	}
}

func TestSubscribePushesNotifyToConsumer(t *testing.T) {
	srv, mux := eventsTestServer(t)
	consumer := newPushConsumer(t, http.StatusOK)

	status, resp, subPath := subscribe(t, mux, consumer.server.URL, "PT10M")
	if status != http.StatusOK {
		t.Fatalf("subscribe status = %d, body: %s", status, resp)
	}

	if !strings.Contains(resp, "SubscribeResponse") || !strings.Contains(resp, "TerminationTime") {
		t.Fatalf("subscribe response shape:\n%s", resp)
	}

	if subPath == "" {
		t.Fatalf("no SubscriptionReference address in:\n%s", resp)
	}

	srv.PublishEvent(Event{
		Topic:  "tns1:VideoSource/MotionAlarm",
		Source: []SimpleItem{{Name: "Source", Value: "0"}},
		Data:   []SimpleItem{{Name: "State", Value: "true"}, {Name: "Targets", Value: "3"}},
	})

	deadline := time.Now().Add(3 * time.Second)
	for len(consumer.captured()) == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}

	bodies := consumer.captured()
	if len(bodies) == 0 {
		t.Fatal("Notify POST never arrived at the consumer")
	}

	notify := bodies[0]
	for _, want := range []string{
		"<Notify",
		"tns1:VideoSource/MotionAlarm",
		"SubscriptionReference",
		`Name="State"`,
		`Value="true"`,
		`Name="Targets" Value="3"`,
	} {
		if !strings.Contains(notify, want) {
			t.Errorf("Notify document missing %q:\n%s", want, notify)
		}
	}

	// The Notify names the subscription it was delivered for.
	if !strings.Contains(notify, subPath[strings.LastIndex(subPath, "/")+1:]) {
		t.Errorf("Notify SubscriptionReference does not address the subscription:\n%s", notify)
	}

	// Renew / Unsubscribe share the pull-point lifecycle on the sub address.
	status, resp = postEventsSOAP(t, mux, subPath, `<Renew xmlns="http://docs.oasis-open.org/wsn/b-2"><TerminationTime>PT5M</TerminationTime></Renew>`)
	if status != http.StatusOK || !strings.Contains(resp, "RenewResponse") {
		t.Fatalf("renew status = %d body:\n%s", status, resp)
	}

	status, resp = postEventsSOAP(t, mux, subPath, `<Unsubscribe xmlns="http://docs.oasis-open.org/wsn/b-2"></Unsubscribe>`)
	if status != http.StatusOK || !strings.Contains(resp, "UnsubscribeResponse") {
		t.Fatalf("unsubscribe status = %d body:\n%s", status, resp)
	}
}

func TestPullMessagesRefusesPushSubscription(t *testing.T) {
	srv, mux := eventsTestServer(t)
	_ = srv
	consumer := newPushConsumer(t, http.StatusOK)

	_, _, subPath := subscribe(t, mux, consumer.server.URL, "PT10M")
	if subPath == "" {
		t.Fatal("no subscription address")
	}

	status, resp := postEventsSOAP(t, mux, subPath,
		`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl"><Timeout>PT0S</Timeout><MessageLimit>5</MessageLimit></PullMessages>`)
	if status == http.StatusOK || !strings.Contains(resp, "Not a pull point") {
		t.Fatalf("PullMessages on a push subscription: status=%d body:\n%s", status, resp)
	}
}

func TestSubscribeValidationFaults(t *testing.T) {
	_, mux := eventsTestServer(t)

	// No ConsumerReference.
	status, resp := postEventsSOAP(t, mux, "/onvif/events_service",
		`<Subscribe xmlns="http://docs.oasis-open.org/wsn/b-2"></Subscribe>`)
	if status == http.StatusOK || !strings.Contains(resp, "ConsumerReference") {
		t.Fatalf("missing consumer: status=%d body:\n%s", status, resp)
	}

	// HTTPS consumers are refused (the server carries no TLS client).
	status, resp, _ = subscribe(t, mux, "https://consumer/notify", "PT10M")
	if status == http.StatusOK || !strings.Contains(resp, "http") {
		t.Fatalf("https consumer: status=%d body:\n%s", status, resp)
	}

	// Unparsable termination.
	status, resp, _ = subscribe(t, mux, "http://127.0.0.1:1/notify", "not-a-duration")
	if status == http.StatusOK || !strings.Contains(resp, "Invalid TerminationTime") {
		t.Fatalf("bad termination: status=%d body:\n%s", status, resp)
	}
}

func TestDeadConsumerAutoUnsubscribed(t *testing.T) {
	srv, mux := eventsTestServer(t)
	consumer := newPushConsumer(t, http.StatusInternalServerError)

	_, _, subPath := subscribe(t, mux, consumer.server.URL, "PT10M")
	if subPath == "" {
		t.Fatal("no subscription address")
	}

	// Publish enough events to burn the failure budget (one POST per
	// event, three consecutive failures remove the subscription).
	for range 5 {
		srv.PublishEvent(Event{Topic: "tns1:VideoSource/SignalLoss"})
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(consumer.captured()) >= maxPushFailures {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// The subscription is gone: Renew on its address faults unknown-subscription.
	status, resp := postEventsSOAP(t, mux, subPath, `<Renew xmlns="http://docs.oasis-open.org/wsn/b-2"><TerminationTime>PT5M</TerminationTime></Renew>`)
	if status == http.StatusOK {
		t.Fatalf("renew on auto-unsubscribed subscription unexpectedly succeeded:\n%s", resp)
	}

	// And no further deliveries are attempted after removal.
	after := len(consumer.captured())
	srv.PublishEvent(Event{Topic: "tns1:VideoSource/SignalLoss"})
	time.Sleep(200 * time.Millisecond)
	if got := len(consumer.captured()); got != after {
		t.Fatalf("deliveries continued after auto-unsubscribe: %d → %d", after, got)
	}
}

func TestBuildNotifyDocumentShape(t *testing.T) {
	doc := buildNotifyDocument("http://192.168.1.30:8080/onvif/events_service/sub/abc",
		queuedNotification{
			topic: "tns1:VideoSource/MotionAlarm",
			message: eventMessage{
				PropertyOperation: "Changed",
				UtcTime:           "2026-10-12T08:57:05Z",
				Source:            eventItemGroup{SimpleItems: []eventSimpleItem{{Name: "Source", Value: "0"}}},
				Data:              eventItemGroup{SimpleItems: []eventSimpleItem{{Name: "State", Value: "true"}}},
			},
		})

	for _, want := range []string{
		`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">`,
		"<Notify",
		"<SubscriptionReference",
		"http://192.168.1.30:8080/onvif/events_service/sub/abc",
		"tns1:VideoSource/MotionAlarm",
		`PropertyOperation="Changed"`,
		`UtcTime="2026-10-12T08:57:05Z"`,
		"<SimpleItem",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("document missing %q:\n%s", want, doc)
		}
	}

	if !strings.HasPrefix(doc, "<?xml") {
		t.Error("document lacks the XML declaration")
	}
}

func TestParsePushTarget(t *testing.T) {
	target, err := parsePushTarget("http://192.168.1.10:8080/api/onvif/notify")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if target.host != "192.168.1.10" || target.port != "8080" || target.path != "/api/onvif/notify" {
		t.Errorf("target = %+v", target)
	}

	// Default port.
	target, err = parsePushTarget("http://consumer/notify")
	if err != nil || target.port != "80" {
		t.Errorf("default port: %+v err=%v", target, err)
	}

	if _, err := parsePushTarget("https://consumer/notify"); err == nil {
		t.Error("https accepted")
	}

	if _, err := parsePushTarget("ftp://consumer/notify"); err == nil {
		t.Error("ftp accepted")
	}

	if _, err := parsePushTarget("http://"); err == nil {
		t.Error("hostless accepted")
	}
}

func TestSubscribeSharesThePullPointCap(t *testing.T) {
	_, mux := eventsTestServer(t)
	consumer := newPushConsumer(t, http.StatusOK)

	for range defaultMaxPullPoints {
		status, _, _ := subscribe(t, mux, consumer.server.URL, "PT10M")
		if status != http.StatusOK {
			t.Fatalf("subscription below cap failed")
		}
	}

	status, resp, _ := subscribe(t, mux, consumer.server.URL, "PT10M")
	if status == http.StatusOK || !strings.Contains(resp, "at most") {
		t.Fatalf("cap not enforced: status=%d body:\n%s", status, resp)
	}

	_ = fmt.Sprint() // keep fmt import stable if asserts above shrink
}
