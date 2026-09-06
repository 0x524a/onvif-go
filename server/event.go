package server

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// defaultSubscriptionDuration is used when a CreatePullPointSubscription or
// Renew request omits its termination-time field.
const defaultSubscriptionDuration = 10 * time.Minute

// Event service SOAP message types.
//
// This is a minimal pull-point implementation: CreatePullPointSubscription,
// PullMessages, Renew, and Unsubscribe only - enough to exercise this
// repo's own client library end-to-end. There is no notification producer,
// so PullMessages always returns zero messages; Seek,
// SetEventSynchronizationPoint, GetEventServiceCapabilities,
// GetEventProperties, and the EventBroker family are out of scope.

// CreatePullPointSubscriptionRequest represents a CreatePullPointSubscription request.
type CreatePullPointSubscriptionRequest struct {
	XMLName                xml.Name `xml:"CreatePullPointSubscription"`
	InitialTerminationTime string   `xml:"InitialTerminationTime"`
}

// CreatePullPointSubscriptionResponse represents a CreatePullPointSubscription response.
type CreatePullPointSubscriptionResponse struct {
	XMLName               xml.Name              `xml:"http://www.onvif.org/ver10/events/wsdl CreatePullPointSubscriptionResponse"`
	SubscriptionReference SubscriptionReference `xml:"SubscriptionReference"`
	CurrentTime           string                `xml:"CurrentTime"`
	TerminationTime       string                `xml:"TerminationTime"`
}

// SubscriptionReference represents a WS-Eventing subscription reference.
type SubscriptionReference struct {
	Address string `xml:"Address"`
}

// PullMessagesRequest represents a PullMessages request.
type PullMessagesRequest struct {
	XMLName      xml.Name `xml:"PullMessages"`
	Timeout      string   `xml:"Timeout"`
	MessageLimit int      `xml:"MessageLimit"`
}

// PullMessagesResponse represents a PullMessages response.
type PullMessagesResponse struct {
	XMLName         xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl PullMessagesResponse"`
	CurrentTime     string   `xml:"CurrentTime"`
	TerminationTime string   `xml:"TerminationTime"`
}

// RenewRequest represents a Renew request.
type RenewRequest struct {
	XMLName         xml.Name `xml:"Renew"`
	TerminationTime string   `xml:"TerminationTime"`
}

// RenewResponse represents a Renew response.
type RenewResponse struct {
	XMLName         xml.Name `xml:"http://docs.oasis-open.org/wsn/b-2 RenewResponse"`
	CurrentTime     string   `xml:"CurrentTime"`
	TerminationTime string   `xml:"TerminationTime"`
}

// UnsubscribeResponse represents an Unsubscribe response.
type UnsubscribeResponse struct {
	XMLName xml.Name `xml:"http://docs.oasis-open.org/wsn/b-2 UnsubscribeResponse"`
}

// Event service handlers

// HandleCreatePullPointSubscription handles CreatePullPointSubscription requests.
// It replaces any existing subscription with a new one - this simulator
// supports exactly one active pull-point subscription at a time (see
// eventSubscription's doc comment for why).
func (s *Server) HandleCreatePullPointSubscription(body interface{}) (interface{}, error) {
	var req CreatePullPointSubscriptionRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	duration := defaultSubscriptionDuration
	if req.InitialTerminationTime != "" {
		d, err := parseISODuration(req.InitialTerminationTime)
		if err != nil {
			return nil, fmt.Errorf("invalid InitialTerminationTime: %w", err)
		}
		duration = d
	}

	now := time.Now()
	sub := &eventSubscription{terminationTime: now.Add(duration)}

	s.subscriptionMu.Lock()
	s.subscription = sub
	s.scheduleExpiry(sub, duration)
	s.subscriptionMu.Unlock()

	return &CreatePullPointSubscriptionResponse{
		SubscriptionReference: SubscriptionReference{Address: s.eventsServiceURL()},
		CurrentTime:           now.UTC().Format(time.RFC3339),
		TerminationTime:       sub.terminationTime.UTC().Format(time.RFC3339),
	}, nil
}

// HandlePullMessages handles PullMessages requests. There is no
// notification producer in this simulator, so a successful call always
// returns zero messages.
func (s *Server) HandlePullMessages(body interface{}) (interface{}, error) {
	var req PullMessagesRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	s.subscriptionMu.RLock()
	sub := s.subscription
	s.subscriptionMu.RUnlock()

	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}

	return &PullMessagesResponse{
		CurrentTime:     time.Now().UTC().Format(time.RFC3339),
		TerminationTime: sub.terminationTime.UTC().Format(time.RFC3339),
	}, nil
}

// HandleRenew handles Renew requests, extending the active subscription's
// termination time.
func (s *Server) HandleRenew(body interface{}) (interface{}, error) {
	var req RenewRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	duration := defaultSubscriptionDuration
	if req.TerminationTime != "" {
		d, err := parseISODuration(req.TerminationTime)
		if err != nil {
			return nil, fmt.Errorf("invalid TerminationTime: %w", err)
		}
		duration = d
	}

	s.subscriptionMu.Lock()
	defer s.subscriptionMu.Unlock()

	sub := s.subscription
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}

	now := time.Now()
	sub.terminationTime = now.Add(duration)
	s.scheduleExpiry(sub, duration)

	return &RenewResponse{
		CurrentTime:     now.UTC().Format(time.RFC3339),
		TerminationTime: sub.terminationTime.UTC().Format(time.RFC3339),
	}, nil
}

// HandleUnsubscribe handles Unsubscribe requests, immediately terminating
// the active subscription. Unsubscribe carries no request fields, so there
// is nothing to decode from body.
func (s *Server) HandleUnsubscribe(_ interface{}) (interface{}, error) {
	s.subscriptionMu.Lock()
	defer s.subscriptionMu.Unlock()

	if s.subscription == nil {
		return nil, ErrSubscriptionNotFound
	}
	if s.subscription.expiryTimer != nil {
		s.subscription.expiryTimer.Stop()
	}
	s.subscription = nil

	return &UnsubscribeResponse{}, nil
}

// scheduleExpiry stops any expiry timer already pending for sub and
// schedules a new one that clears s.subscription after delay - but only if
// s.subscription is still sub by the time the timer fires. That
// pointer-identity check matters because CreatePullPointSubscription
// replaces the whole *eventSubscription rather than mutating one in place
// (unlike PTZState's settleTimer, which mutates fields inside the same
// long-lived struct): without it, a very-late-firing timer from a
// since-replaced subscription could nil out a brand-new one. Must be
// called with s.subscriptionMu held, since it mutates sub.expiryTimer.
func (s *Server) scheduleExpiry(sub *eventSubscription, delay time.Duration) {
	if sub.expiryTimer != nil {
		sub.expiryTimer.Stop()
	}
	sub.expiryTimer = time.AfterFunc(delay, func() {
		s.subscriptionMu.Lock()
		defer s.subscriptionMu.Unlock()
		if s.subscription == sub {
			s.subscription = nil
		}
	})
}

// eventsServiceURL builds the URL this server's events service is reachable
// at, using the same host-normalization and base-URL construction as
// HandleGetCapabilities (device.go).
func (s *Server) eventsServiceURL() string {
	return s.advertisedBaseURL() + "/events_service"
}

// isoDurationPattern matches the limited ISO-8601 duration grammar this
// repo's own client ever produces (event.go's formatDuration): "PT<n>S",
// "PT<n>M", or "PT<n>M<n>S". No other producer of these values exists to
// test against, so a general ISO-8601 duration parser is out of scope.
var isoDurationPattern = regexp.MustCompile(`^PT(?:(\d+)M)?(?:(\d+)S)?$`)

// parseISODuration parses an ISO-8601 duration string in the grammar
// described by isoDurationPattern.
func parseISODuration(s string) (time.Duration, error) {
	m := isoDurationPattern.FindStringSubmatch(s)
	if m == nil || (m[1] == "" && m[2] == "") {
		return 0, fmt.Errorf("%w: %s", ErrInvalidDuration, s)
	}

	var d time.Duration

	if m[1] != "" {
		minutes, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, fmt.Errorf("%w: %s", ErrInvalidDuration, s)
		}
		d += time.Duration(minutes) * time.Minute
	}

	if m[2] != "" {
		seconds, err := strconv.Atoi(m[2])
		if err != nil {
			return 0, fmt.Errorf("%w: %s", ErrInvalidDuration, s)
		}
		d += time.Duration(seconds) * time.Second
	}

	return d, nil
}
