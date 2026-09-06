package server

import (
	"errors"
	"testing"
	"time"
)

const (
	testDuration1Sec = "PT1S"
	testDuration5Sec = "PT5S"
	testDuration1Min = "PT1M"
)

func TestHandleCreatePullPointSubscriptionDefaultDuration(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	before := time.Now()

	resp, err := srv.HandleCreatePullPointSubscription(CreatePullPointSubscriptionRequest{})
	if err != nil {
		t.Fatalf("HandleCreatePullPointSubscription() error = %v", err)
	}

	createResp, ok := resp.(*CreatePullPointSubscriptionResponse)
	if !ok {
		t.Fatalf("Response is not CreatePullPointSubscriptionResponse, got %T", resp)
	}

	if createResp.SubscriptionReference.Address == "" {
		t.Error("expected a non-empty SubscriptionReference.Address")
	}

	term, err := time.Parse(time.RFC3339, createResp.TerminationTime)
	if err != nil {
		t.Fatalf("TerminationTime is not RFC3339: %v", err)
	}
	if diff := term.Sub(before); diff < defaultSubscriptionDuration-time.Second || diff > defaultSubscriptionDuration+time.Second {
		t.Errorf("expected TerminationTime ~%v after now, got diff %v", defaultSubscriptionDuration, diff)
	}

	srv.subscriptionMu.RLock()
	sub := srv.subscription
	srv.subscriptionMu.RUnlock()
	if sub == nil {
		t.Error("expected srv.subscription to be populated")
	}
}

func TestHandleCreatePullPointSubscriptionExplicitDuration(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	before := time.Now()

	resp, err := srv.HandleCreatePullPointSubscription(CreatePullPointSubscriptionRequest{
		InitialTerminationTime: testDuration5Sec,
	})
	if err != nil {
		t.Fatalf("HandleCreatePullPointSubscription() error = %v", err)
	}

	createResp, ok := resp.(*CreatePullPointSubscriptionResponse)
	if !ok {
		t.Fatalf("Response is not CreatePullPointSubscriptionResponse, got %T", resp)
	}

	term, err := time.Parse(time.RFC3339, createResp.TerminationTime)
	if err != nil {
		t.Fatalf("TerminationTime is not RFC3339: %v", err)
	}
	if diff := term.Sub(before); diff < 4*time.Second || diff > 6*time.Second {
		t.Errorf("expected TerminationTime ~5s after now, got diff %v", diff)
	}
}

func TestHandleCreatePullPointSubscriptionReplacesExisting(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if _, err := srv.HandleCreatePullPointSubscription(CreatePullPointSubscriptionRequest{InitialTerminationTime: testDuration1Sec}); err != nil {
		t.Fatalf("first HandleCreatePullPointSubscription() error = %v", err)
	}

	srv.subscriptionMu.RLock()
	first := srv.subscription
	srv.subscriptionMu.RUnlock()

	if _, err := srv.HandleCreatePullPointSubscription(CreatePullPointSubscriptionRequest{InitialTerminationTime: testDuration1Min}); err != nil {
		t.Fatalf("second HandleCreatePullPointSubscription() error = %v", err)
	}

	srv.subscriptionMu.RLock()
	second := srv.subscription
	srv.subscriptionMu.RUnlock()

	if first == second {
		t.Fatalf("expected the second call to replace the subscription pointer")
	}

	// Wait past the first subscription's 1s expiry. The pointer-identity
	// guard in scheduleExpiry must stop the stale timer from clearing the
	// second (still-active, 1-minute) subscription.
	time.Sleep(1500 * time.Millisecond)

	srv.subscriptionMu.RLock()
	current := srv.subscription
	srv.subscriptionMu.RUnlock()

	if current == nil {
		t.Errorf("expected the second subscription to still be active; a stale timer from the first cleared it")
	}
}

func TestHandlePullMessagesNoActiveSubscription(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if _, err := srv.HandlePullMessages(PullMessagesRequest{Timeout: testDuration1Sec, MessageLimit: 10}); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
	}
}

func TestHandlePullMessagesReturnsEmpty(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if _, err := srv.HandleCreatePullPointSubscription(CreatePullPointSubscriptionRequest{}); err != nil {
		t.Fatalf("HandleCreatePullPointSubscription() error = %v", err)
	}

	resp, err := srv.HandlePullMessages(PullMessagesRequest{Timeout: testDuration1Sec, MessageLimit: 10})
	if err != nil {
		t.Fatalf("HandlePullMessages() error = %v", err)
	}

	if _, ok := resp.(*PullMessagesResponse); !ok {
		t.Fatalf("expected *PullMessagesResponse, got %T", resp)
	}
}

func TestHandleRenewExtendsTerminationTime(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	createResp, err := srv.HandleCreatePullPointSubscription(CreatePullPointSubscriptionRequest{InitialTerminationTime: testDuration5Sec})
	if err != nil {
		t.Fatalf("HandleCreatePullPointSubscription() error = %v", err)
	}
	originalTerm, err := time.Parse(time.RFC3339, createResp.(*CreatePullPointSubscriptionResponse).TerminationTime)
	if err != nil {
		t.Fatalf("failed to parse original TerminationTime: %v", err)
	}

	renewResp, err := srv.HandleRenew(RenewRequest{TerminationTime: testDuration1Min})
	if err != nil {
		t.Fatalf("HandleRenew() error = %v", err)
	}

	newTerm, err := time.Parse(time.RFC3339, renewResp.(*RenewResponse).TerminationTime)
	if err != nil {
		t.Fatalf("failed to parse renewed TerminationTime: %v", err)
	}

	if !newTerm.After(originalTerm) {
		t.Errorf("expected renewed TerminationTime %v to be after original %v", newTerm, originalTerm)
	}
}

func TestHandleRenewNoActiveSubscription(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if _, err := srv.HandleRenew(RenewRequest{TerminationTime: testDuration1Min}); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
	}
}

func TestHandleUnsubscribeClearsSubscription(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if _, err := srv.HandleCreatePullPointSubscription(CreatePullPointSubscriptionRequest{}); err != nil {
		t.Fatalf("HandleCreatePullPointSubscription() error = %v", err)
	}

	if _, err := srv.HandleUnsubscribe(nil); err != nil {
		t.Fatalf("HandleUnsubscribe() error = %v", err)
	}

	srv.subscriptionMu.RLock()
	sub := srv.subscription
	srv.subscriptionMu.RUnlock()
	if sub != nil {
		t.Error("expected srv.subscription to be nil after Unsubscribe")
	}

	if _, err := srv.HandlePullMessages(PullMessagesRequest{Timeout: testDuration1Sec, MessageLimit: 10}); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected PullMessages after Unsubscribe to fail with ErrSubscriptionNotFound, got %v", err)
	}
}

func TestHandleUnsubscribeNoActiveSubscription(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if _, err := srv.HandleUnsubscribe(nil); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
	}
}

// TestEventSubscriptionExpires lets a subscription's expiry timer actually
// fire (with a short custom delay, rather than waiting out the real
// 10-minute default) and verifies subsequent PullMessages/Renew calls see
// it as gone. Mirrors TestScheduleSettleFires in integration_test.go.
func TestEventSubscriptionExpires(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	sub := &eventSubscription{terminationTime: time.Now().Add(10 * time.Millisecond)}
	srv.subscriptionMu.Lock()
	srv.subscription = sub
	srv.scheduleExpiry(sub, 10*time.Millisecond)
	srv.subscriptionMu.Unlock()

	time.Sleep(150 * time.Millisecond)

	if _, err := srv.HandlePullMessages(PullMessagesRequest{Timeout: testDuration1Sec, MessageLimit: 10}); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected the expiry timer to have cleared the subscription, got %v", err)
	}
}

func TestParseISODuration(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{input: testDuration5Sec, want: 5 * time.Second},
		{input: testDuration1Min, want: 1 * time.Minute},
		{input: "PT1M30S", want: 90 * time.Second},
		{input: "", wantErr: true},
		{input: "PT", wantErr: true},
		{input: "garbage", wantErr: true},
		{input: "P1D", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseISODuration(tt.input)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidDuration) {
					t.Errorf("parseISODuration(%q): expected ErrInvalidDuration, got %v", tt.input, err)
				}

				return
			}
			if err != nil {
				t.Fatalf("parseISODuration(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseISODuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
