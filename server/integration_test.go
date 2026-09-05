package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	internalsoap "github.com/0x524a/onvif-go/internal/soap"
)

// TestServeHTTPEndToEndContinuousMove is a regression test for a bug where
// every parameterized server operation always received a nil request body
// (originsoap.Body.Content was interface{}, which encoding/xml cannot
// populate), so every such operation always failed with a SOAP Fault
// regardless of what a client sent. Prior to this test, no _test.go
// anywhere drove a real HTTP request through the real wire path
// (soap.Handler.ServeHTTP) against a real *Server - every existing test
// called Handle* functions directly with bytes the broken layer never
// actually produces, which is why the bug was invisible.
//
// This test builds a real *Server, wires it up exactly the way Start()
// does, and uses the real internal/soap.Client (the same client the public
// onvif.Client uses against real cameras) to send a real, WS-Security
// authenticated ContinuousMove request over real HTTP. It then asserts
// server-side state actually changed as a result.
func TestServeHTTPEndToEndContinuousMove(t *testing.T) {
	config := createTestConfig()

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	mux := http.NewServeMux()
	srv.registerPTZService(mux)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	profileToken := config.Profiles[0].Token
	soapClient := internalsoap.NewClient(testServer.Client(), config.Username, config.Password)

	req := ContinuousMoveRequest{
		ProfileToken: profileToken,
		Velocity: PTZVector{
			PanTilt: &Vector2D{X: 0.5, Y: 0.2},
		},
	}

	endpoint := testServer.URL + config.BasePath + "/ptz_service"
	if err := soapClient.Call(context.Background(), endpoint, "", req, nil); err != nil {
		t.Fatalf("ContinuousMove over the real wire path failed: %v", err)
	}

	state, ok := srv.GetPTZState(profileToken)
	if !ok {
		t.Fatalf("expected PTZ state for profile %q", profileToken)
	}
	if !state.Moving || !state.PanMoving || !state.TiltMoving {
		t.Errorf("expected PTZ state to reflect the move sent over HTTP, got %+v", state)
	}
}

// TestUpdateStreamURIRace exercises UpdateStreamURI concurrently with reads
// via GetStreamConfig and HandleGetStreamURI, catching the data race that
// existed before streamsMu was added. Run with -race.
func TestUpdateStreamURIRace(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	profileToken := config.Profiles[0].Token

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = srv.UpdateStreamURI(profileToken, "rtsp://example.invalid/stream")
		}()
		go func() {
			defer wg.Done()
			_, _ = srv.GetStreamConfig(profileToken)
		}()
		go func() {
			defer wg.Done()
			type getStreamURIRequest struct {
				ProfileToken string `xml:"ProfileToken"`
			}
			_, _ = srv.HandleGetStreamURI(getStreamURIRequest{ProfileToken: profileToken})
		}()
	}
	wg.Wait()
}

// TestPTZSettleTimerCancelledOnRepeatedMove is a regression test for the
// PTZ handlers spawning an untracked, uncancellable goroutine per move to
// clear the Moving flags after a fixed delay: a second move for the same
// profile before the first one settled left two goroutines racing to clear
// the same state, so the first move's goroutine could clobber the second
// move's Moving flag out from under it.
//
// Timeline: call 1 schedules a settle for t=500ms. At t=300ms, call 2
// happens - this must cancel call 1's pending timer and schedule its own
// for t=800ms. The check at t=600ms falls after call 1's original delay
// but well before call 2's, so if call 1's timer weren't canceled it
// would have already fired and cleared Moving by then.
func TestPTZSettleTimerCancelledOnRepeatedMove(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	profileToken := config.Profiles[0].Token

	move := AbsoluteMoveRequest{
		ProfileToken: profileToken,
		Position:     PTZVector{PanTilt: &Vector2D{X: 0.5, Y: 0.2}},
	}

	if _, err := srv.HandleAbsoluteMove(move); err != nil {
		t.Fatalf("first HandleAbsoluteMove() error = %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	if _, err := srv.HandleAbsoluteMove(move); err != nil {
		t.Fatalf("second HandleAbsoluteMove() error = %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	state, ok := srv.GetPTZState(profileToken)
	if !ok {
		t.Fatalf("expected PTZ state for profile %q", profileToken)
	}
	if !state.Moving {
		t.Errorf("Moving was cleared by a stale settle timer from the first move; " +
			"the second move's own timer should still be pending")
	}
}

// TestPTZStopCancelsSettleTimer is a regression test proving Stop cancels
// any pending settle timer immediately (state.settleTimer is nilled out),
// rather than leaving it live to fire later and clear whatever a
// subsequent move sets. Checked directly on the PTZState field rather than
// via a timing-based wait, since a stale timer clearing already-false
// flags is not observable through Moving alone.
func TestPTZStopCancelsSettleTimer(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	profileToken := config.Profiles[0].Token

	move := AbsoluteMoveRequest{
		ProfileToken: profileToken,
		Position:     PTZVector{PanTilt: &Vector2D{X: 0.5, Y: 0.2}},
	}
	if _, err := srv.HandleAbsoluteMove(move); err != nil {
		t.Fatalf("HandleAbsoluteMove() error = %v", err)
	}

	state, ok := srv.GetPTZState(profileToken)
	if !ok {
		t.Fatalf("expected PTZ state for profile %q", profileToken)
	}
	if state.settleTimer == nil {
		t.Fatalf("expected a pending settle timer after AbsoluteMove")
	}

	if _, err := srv.HandleStop(StopRequest{ProfileToken: profileToken, PanTilt: true, Zoom: true}); err != nil {
		t.Fatalf("HandleStop() error = %v", err)
	}

	if state.settleTimer != nil {
		t.Errorf("expected Stop to cancel and clear the pending settle timer, it's still set")
	}
}

// TestHandleRelativeMove exercises HandleRelativeMove directly with a real
// typed request (unlike the disabled/skipped tests elsewhere in this
// package that hand-build namespace-less XML and silently never reach
// their assertions, this actually runs).
func TestHandleRelativeMove(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	profileToken := config.Profiles[0].Token

	req := RelativeMoveRequest{
		ProfileToken: profileToken,
		Translation: PTZVector{
			PanTilt: &Vector2D{X: 10, Y: 5},
			Zoom:    &Vector1D{X: 0.1},
		},
	}

	resp, err := srv.HandleRelativeMove(req)
	if err != nil {
		t.Fatalf("HandleRelativeMove() error = %v", err)
	}
	if _, ok := resp.(*RelativeMoveResponse); !ok {
		t.Fatalf("expected *RelativeMoveResponse, got %T", resp)
	}

	state, ok := srv.GetPTZState(profileToken)
	if !ok {
		t.Fatalf("expected PTZ state for profile %q", profileToken)
	}
	if !state.Moving {
		t.Errorf("expected Moving to be true after RelativeMove")
	}
	if state.Position.Pan != 10 || state.Position.Tilt != 5 || state.Position.Zoom != 0.1 {
		t.Errorf("expected position to reflect the translation, got %+v", state.Position)
	}

	if _, err := srv.HandleRelativeMove(RelativeMoveRequest{ProfileToken: testInvalidToken}); !errors.Is(err, ErrPTZNotSupported) {
		t.Errorf("expected ErrPTZNotSupported for an unknown profile, got %v", err)
	}
}

// TestHandleMove exercises the imaging HandleMove (focus) handler directly
// with a real typed request - the existing _DisabledTestHandleMove in
// imaging_test.go is disabled for the same namespace-mismatch reason as
// TestHandleGotoPreset above.
func TestHandleMove(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	videoSourceToken := config.Profiles[0].VideoSource.Token

	resp, err := srv.HandleMove(MoveRequest{
		VideoSourceToken: videoSourceToken,
		Focus:            &FocusMove{Absolute: &AbsoluteFocus{Position: 0.75}},
	})
	if err != nil {
		t.Fatalf("HandleMove() error = %v", err)
	}
	if _, ok := resp.(*MoveResponse); !ok {
		t.Fatalf("expected *MoveResponse, got %T", resp)
	}

	state, ok := srv.GetImagingState(videoSourceToken)
	if !ok {
		t.Fatalf("expected imaging state for video source %q", videoSourceToken)
	}
	if state.Focus.CurrentPos != 0.75 {
		t.Errorf("expected Focus.CurrentPos to be 0.75, got %v", state.Focus.CurrentPos)
	}

	if _, err := srv.HandleMove(MoveRequest{VideoSourceToken: testInvalidToken}); !errors.Is(err, ErrVideoSourceNotFound) {
		t.Errorf("expected ErrVideoSourceNotFound for an unknown video source, got %v", err)
	}
}

// TestHandleGetStreamURIUnknownProfile covers the not-found path of
// HandleGetStreamURI, which the new streamsMu-guarded early return
// (server/media.go) needs exercised.
func TestHandleGetStreamURIUnknownProfile(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	type getStreamURIRequest struct {
		ProfileToken string `xml:"ProfileToken"`
	}
	_, err = srv.HandleGetStreamURI(getStreamURIRequest{ProfileToken: testInvalidToken})
	if !errors.Is(err, ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

// TestScheduleSettleFires lets a settle timer actually fire (with a short
// custom delay, rather than waiting out the real 500ms/1s constants) and
// verifies it clears the Moving flags - the one path none of the other PTZ
// timer tests exercise, since they only check cancellation/replacement.
func TestScheduleSettleFires(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	profileToken := config.Profiles[0].Token

	state, ok := srv.GetPTZState(profileToken)
	if !ok {
		t.Fatalf("expected PTZ state for profile %q", profileToken)
	}

	srv.ptzMu.Lock()
	state.Moving = true
	state.PanMoving = true
	state.TiltMoving = true
	state.ZoomMoving = true
	srv.scheduleSettle(state, 10*time.Millisecond)
	srv.ptzMu.Unlock()

	time.Sleep(150 * time.Millisecond)

	state, ok = srv.GetPTZState(profileToken)
	if !ok {
		t.Fatalf("expected PTZ state for profile %q", profileToken)
	}
	if state.Moving || state.PanMoving || state.TiltMoving || state.ZoomMoving {
		t.Errorf("expected the settle timer to have cleared all Moving flags, got %+v", state)
	}
}
