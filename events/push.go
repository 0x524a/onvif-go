package events

// WS-BaseNotification push path (issue #127): the device POSTs wsnt:Notify
// bodies to a consumer HTTP endpoint instead of being polled. Subscribe
// probes the events service; Renew and Unsubscribe live on the
// subscription's OWN manager endpoint (not the events service) — a
// distinction real devices enforce. The managed PushStream keeps the
// subscription alive with half-TTL renewals.
//
// Envelope shapes derive from the MiBeeNvr production implementation,
// field-proven against Hikvision, Uniview, TP-Link and in-house firmware.

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/internal/soap"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// WS-BaseNotification and WS-Addressing namespaces.
const (
	wsntNamespace = "http://docs.oasis-open.org/wsn/b-2"
	wsaNamespace  = "http://www.w3.org/2005/08/addressing"
)

// ErrPushNotSupported reports that the device answered the Subscribe probe
// with a sender-side SOAP fault — the caller falls back to the Pull-Point
// path (SubscribeEvents) instead.
var ErrPushNotSupported = errors.New("push event subscription not supported")

// defaultPushRenewFloor bounds the half-TTL renew interval from below:
// devices that grant very short terms must not turn the renew loop into a
// busy cycle.
const defaultPushRenewFloor = 2 * time.Second

// wsnTimeLayouts covers the xsd:dateTime shapes devices actually emit for
// CurrentTime / TerminationTime.
var wsnTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.999999999",
}

// parseWSNTime decodes a device-emitted xsd:dateTime (UTC when it carries
// no zone).
func parseWSNTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)

	for _, layout := range wsnTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}

	return time.Time{}, false
}

// PushSubscription is the result of a successful push Subscribe.
type PushSubscription struct {
	// SubscriptionReference identifies the subscription on the device (the
	// address its Notify messages relate to).
	SubscriptionReference string
	// ManagerAddress is the endpoint Renew/Unsubscribe must be posted to —
	// the subscription's own manager, NOT the events service.
	ManagerAddress  string
	CurrentTime     time.Time
	TerminationTime time.Time
}

// Subscribe probes the device's event service with a wsnt:Subscribe for
// consumerReference (the HTTP endpoint this client will receive Notify
// POSTs on). A sender-side SOAP fault maps to ErrPushNotSupported so
// callers can fall back to the Pull-Point path; everything else is
// surfaced as-is.
func (s *Service) Subscribe(
	ctx context.Context,
	consumerReference string,
	initialTermination time.Duration,
) (*PushSubscription, error) {
	if consumerReference == "" {
		return nil, fmt.Errorf("Subscribe: %w: consumer reference is empty", types.ErrInvalidParameter)
	}

	if initialTermination <= 0 {
		return nil, fmt.Errorf("Subscribe: %w: termination must be positive", types.ErrInvalidParameter)
	}

	type subscribeRequest struct {
		XMLName           xml.Name `xml:"wsnt:Subscribe"`
		XmlnsWSNT         string   `xml:"xmlns:wsnt,attr"`
		XmlnsWSA          string   `xml:"xmlns:wsa,attr"`
		ConsumerReference struct {
			Address string `xml:"wsa:Address"`
		} `xml:"wsnt:ConsumerReference"`
		InitialTerminationTime string `xml:"wsnt:InitialTerminationTime"`
	}

	type subscribeResponse struct {
		XMLName               xml.Name `xml:"SubscribeResponse"`
		SubscriptionReference struct {
			Address string `xml:"Address"`
		} `xml:"SubscriptionReference"`
		SubscriptionManager struct {
			Address string `xml:"Address"`
		} `xml:"SubscriptionManager"`
		CurrentTime     string `xml:"CurrentTime"`
		TerminationTime string `xml:"TerminationTime"`
	}

	req := subscribeRequest{
		XmlnsWSNT:              wsntNamespace,
		XmlnsWSA:               wsaNamespace,
		InitialTerminationTime: formatDuration(initialTermination),
	}
	req.ConsumerReference.Address = consumerReference

	var resp subscribeResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceEvents), "", req, &resp); err != nil {
		if classifyPushUnsupported(err) {
			return nil, fmt.Errorf("Subscribe: %w: %w", ErrPushNotSupported, err)
		}

		return nil, fmt.Errorf("Subscribe: %w", err)
	}

	subRef := strings.TrimSpace(resp.SubscriptionReference.Address)
	if subRef == "" {
		return nil, fmt.Errorf("Subscribe: no SubscriptionReference address in response")
	}

	// Plenty of devices expose the same address for both roles.
	manager := strings.TrimSpace(resp.SubscriptionManager.Address)
	if manager == "" {
		manager = subRef
	}

	sub := &PushSubscription{
		SubscriptionReference: subRef,
		ManagerAddress:        manager,
	}

	if t, ok := parseWSNTime(resp.CurrentTime); ok {
		sub.CurrentTime = t
	}

	if t, ok := parseWSNTime(resp.TerminationTime); ok {
		sub.TerminationTime = t
	}

	return sub, nil
}

// classifyPushUnsupported recognizes sender-side refusals: SOAP 1.2
// s:Sender and the SOAP 1.1 Client equivalent. Receiver faults and
// authorization failures are real errors, not "push not supported".
func classifyPushUnsupported(err error) bool {
	var fault *soap.FaultError
	if !errors.As(err, &fault) {
		return false
	}

	joined := strings.ToLower(fault.Code + "|" + fault.Subcode)

	return strings.Contains(joined, "sender") || strings.Contains(joined, "client")
}

// RenewPush extends the push subscription by posting wsnt:Renew to the
// subscription's manager address (see PushSubscription.ManagerAddress) and
// returns the newly granted termination time.
func (s *Service) RenewPush(ctx context.Context, managerAddress string, termination time.Duration) (time.Time, error) {
	if managerAddress == "" {
		return time.Time{}, fmt.Errorf("RenewPush: %w: manager address is empty", types.ErrInvalidParameter)
	}

	if termination <= 0 {
		return time.Time{}, fmt.Errorf("RenewPush: %w: termination must be positive", types.ErrInvalidParameter)
	}

	type renewRequest struct {
		XMLName         xml.Name `xml:"wsnt:Renew"`
		Xmlns           string   `xml:"xmlns:wsnt,attr"`
		TerminationTime string   `xml:"wsnt:TerminationTime"`
	}

	type renewResponse struct {
		XMLName         xml.Name `xml:"RenewResponse"`
		CurrentTime     string   `xml:"CurrentTime"`
		TerminationTime string   `xml:"TerminationTime"`
	}

	req := renewRequest{
		Xmlns:           wsntNamespace,
		TerminationTime: formatDuration(termination),
	}

	var resp renewResponse

	if err := s.c.Call(ctx, managerAddress, "", req, &resp); err != nil {
		return time.Time{}, fmt.Errorf("RenewPush: %w", err)
	}

	term, ok := parseWSNTime(resp.TerminationTime)
	if !ok {
		return time.Time{}, fmt.Errorf("RenewPush: no TerminationTime in response")
	}

	return term, nil
}

// UnsubscribePush terminates the push subscription by posting
// wsnt:Unsubscribe to the subscription's manager address. Best-effort
// semantics stay with the caller: the device reclaims the subscription at
// its TTL regardless.
func (s *Service) UnsubscribePush(ctx context.Context, managerAddress string) error {
	if managerAddress == "" {
		return fmt.Errorf("UnsubscribePush: %w: manager address is empty", types.ErrInvalidParameter)
	}

	type unsubscribeRequest struct {
		XMLName xml.Name `xml:"wsnt:Unsubscribe"`
		Xmlns   string   `xml:"xmlns:wsnt,attr"`
	}

	type unsubscribeResponse struct {
		XMLName xml.Name `xml:"UnsubscribeResponse"`
	}

	req := unsubscribeRequest{Xmlns: wsntNamespace}

	var resp unsubscribeResponse

	if err := s.c.Call(ctx, managerAddress, "", req, &resp); err != nil {
		return fmt.Errorf("UnsubscribePush: %w", err)
	}

	return nil
}

// notificationMessageXML is the wire shape of one NotificationMessage,
// shared by the PullMessages response and the push Notify document (issue
// #82's two-layer Message: the spec form wraps the ONVIF tt:Message inside
// the outer wsnt:Message — Inner reaches through the wrapper — while lax
// firmwares put the attributes and SimpleItems on the outer element).
type notificationMessageXML struct {
	Topic struct {
		Value string `xml:",chardata"`
	} `xml:"Topic"`
	ProducerReference struct {
		Address string `xml:"Address"`
	} `xml:"ProducerReference"`
	Message struct {
		Inner struct {
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
}

// SimpleItemXML is the wire shape of a SimpleItem (Name/Value attributes).
type SimpleItemXML struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:"Value,attr"`
}

// ParseNotifyEvents parses a wsnt:Notify SOAP body — what the device POSTs
// to the consumer endpoint — into NotificationMessages, via the tolerant
// local-name tree decoder (see notify_parse.go): prefix dialects, wrapped
// envelopes, literal Name/Value item pairs and heartbeats (zero messages)
// are all handled. A SOAP Fault inside the body surfaces as an error.
func ParseNotifyEvents(body []byte) ([]NotificationMessage, error) {
	return parseNotifyTree(body)
}

// maxNotifyBodyBytes caps a Notify POST: real notifications carry a handful
// of messages; the cap keeps a misbehaving sender from pinning memory with
// an unbounded body.
const maxNotifyBodyBytes = 1 << 20 // 1 MiB

// NotifyHandler returns the consumer HTTP endpoint for push events: it
// accepts device wsnt:Notify POSTs (capped at 1 MiB), parses them with
// ParseNotifyEvents, hands each parsed batch to deliver and answers 200.
// Mount it under your own path and wrap it with whatever origin/token
// validation your deployment needs — the devices POSTing cannot
// authenticate (the ONVIF push profile carries no credentials).
func NotifyHandler(deliver func([]NotificationMessage)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "Notify endpoint accepts POST only", http.StatusMethodNotAllowed)

			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxNotifyBodyBytes+1))
		if err != nil {
			http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)

			return
		}

		if len(body) > maxNotifyBodyBytes {
			http.Error(w, "Notify body too large", http.StatusRequestEntityTooLarge)

			return
		}

		messages, err := ParseNotifyEvents(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		deliver(messages)
		w.WriteHeader(http.StatusOK)
	})
}

// SubscribePushOptions tunes the managed push loop. The zero value selects
// the defaults.
type SubscribePushOptions struct {
	// Duration is the requested subscription lifetime per grant (default
	// 1h).
	Duration time.Duration
	// RenewFloor is the minimum spacing between renewals (default 2s):
	// devices granting very short terms must not turn the loop into a
	// busy cycle.
	RenewFloor time.Duration
}

// PushStream is a managed push subscription: a background goroutine
// renews the subscription at half its granted TTL and stops cleanly on
// Unsubscribe or context cancellation. Events themselves arrive on your
// NotifyHandler endpoint — the stream only manages the subscription
// lifecycle.
type PushStream struct {
	svc     *Service
	manager string
	opts    SubscribePushOptions

	mu           sync.Mutex
	granted      time.Duration
	explicitStop bool

	cancel   context.CancelFunc
	done     chan struct{}
	doneOnce sync.Once
}

// SubscribePushEvents creates a push subscription and manages its
// lifetime. A device that refuses push (sender fault) fails fast with an
// error matching errors.Is(err, ErrPushNotSupported) — fall back to
// SubscribeEvents (Pull-Point) in that case. The subscription lives until
// Unsubscribe is called or ctx is cancelled; a renewal failure also
// terminates it (Done closes).
func (s *Service) SubscribePushEvents(
	ctx context.Context,
	consumerURL string,
	opts *SubscribePushOptions,
) (*PushStream, error) {
	if consumerURL == "" {
		return nil, fmt.Errorf("SubscribePushEvents: %w: consumer URL is empty", types.ErrInvalidParameter)
	}

	effective := SubscribePushOptions{
		Duration:   defaultEventSubscriptionDuration,
		RenewFloor: defaultPushRenewFloor,
	}
	if opts != nil {
		if opts.Duration > 0 {
			effective.Duration = opts.Duration
		}

		if opts.RenewFloor > 0 {
			effective.RenewFloor = opts.RenewFloor
		}
	}

	sub, err := s.Subscribe(ctx, consumerURL, effective.Duration)
	if err != nil {
		return nil, fmt.Errorf("SubscribePushEvents: %w", err)
	}

	granted := effective.Duration
	if !sub.TerminationTime.IsZero() {
		if g := time.Until(sub.TerminationTime); g > 0 {
			granted = g
		}
	}

	loopCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	ps := &PushStream{
		svc:     s,
		manager: sub.ManagerAddress,
		opts:    effective,
		granted: granted,
		cancel:  cancel,
		done:    make(chan struct{}),
	}

	// Dying with the caller's context is part of the contract.
	go func() {
		<-ctx.Done()
		ps.stopLoop()
	}()

	go ps.run(loopCtx)

	return ps, nil
}

// Done returns a channel that closes when the managed loop has fully
// exited (Unsubscribe, cancellation, or renewal failure).
func (ps *PushStream) Done() <-chan struct{} {
	return ps.done
}

// Unsubscribe stops the managed loop and best-effort unsubscribes on the
// device: the loop is terminated even when the SOAP unsubscribe fails (the
// error is then returned, cleanup is not blocked by it).
func (ps *PushStream) Unsubscribe(ctx context.Context) error {
	ps.mu.Lock()
	ps.explicitStop = true
	ps.mu.Unlock()
	ps.stopLoop()

	if err := ps.svc.UnsubscribePush(ctx, ps.manager); err != nil {
		return fmt.Errorf("Unsubscribe: device unsubscribe failed (loop stopped anyway): %w", err)
	}

	return nil
}

// stopLoop cancels the loop exactly once.
func (ps *PushStream) stopLoop() {
	ps.cancel()
}

// run renews at half the granted TTL until the subscription dies.
func (ps *PushStream) run(ctx context.Context) {
	defer ps.doneOnce.Do(func() { close(ps.done) })

	for {
		ps.mu.Lock()
		delay := ps.granted / 2
		ps.mu.Unlock()

		if delay < ps.opts.RenewFloor {
			delay = ps.opts.RenewFloor
		}

		select {
		case <-ctx.Done():
			ps.cleanupUnlessExplicit(ctx)

			return
		case <-time.After(delay):
		}

		term, err := ps.svc.RenewPush(ctx, ps.manager, ps.opts.Duration)
		if err != nil {
			// The subscription is gone or the device refuses; retrying
			// forever would just spam — terminate (Done closes).
			return
		}

		granted := time.Until(term)
		if granted <= 0 {
			granted = ps.opts.Duration
		}

		ps.mu.Lock()
		ps.granted = granted
		ps.mu.Unlock()
	}
}

// cleanupUnlessExplicit unsubscribes on the device when the loop died for
// a reason other than an explicit Unsubscribe (the caller's context was
// cancelled): a dangling server-side subscription would otherwise live
// until its TTL.
func (ps *PushStream) cleanupUnlessExplicit(parent context.Context) {
	ps.mu.Lock()
	explicit := ps.explicitStop
	ps.mu.Unlock()

	if explicit {
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)
	defer cancel()

	_ = ps.svc.UnsubscribePush(ctx, ps.manager)
}
