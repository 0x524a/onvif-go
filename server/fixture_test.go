package server

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	internalsoap "github.com/0x524a/onvif-go/internal/soap"
)

// testLoopbackHost is the bind host for fixture servers. Loopback rather than
// a wildcard, so a test server is not briefly reachable from the network.
const testLoopbackHost = "127.0.0.1"

// startTestServer starts srv in a goroutine and waits for it to signal
// readiness, returning its bound address. This is the whole point of #63: no
// hardcoded port, no sleep, no poll-dial.
func startTestServer(t *testing.T, config *Config) (srv *Server, addr string) {
	t.Helper()

	ready := make(chan struct{})
	config.Port = 0
	config.Ready = ready

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start(ctx) }()

	t.Cleanup(func() {
		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("Start() returned %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Error("Start() did not return after context cancellation")
		}
	})

	select {
	case <-ready:
	case err := <-errCh:
		t.Fatalf("Start() returned before signaling readiness: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("server never signaled readiness")
	}

	addr = srv.Addr()

	return srv, addr
}

// TestStartEphemeralPortReadinessAndAddr covers all three #63 fixes at once,
// because they only pay off together: bind an OS-assigned port, learn it, and
// know when it is reachable.
//
// The readiness contract is asserted the strict way - the very first dial must
// succeed, with no retry loop - since a readiness signal that still needs a
// retry loop behind it has not actually removed the raciness it exists to fix.
func TestStartEphemeralPortReadinessAndAddr(t *testing.T) {
	config := createTestConfig()
	srv, addr := startTestServer(t, config)

	if addr == "" {
		t.Fatal("Addr() is empty after readiness was signaled")
	}

	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("Addr() = %q, not a host:port: %v", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port == 0 {
		t.Fatalf("Addr() = %q, want an OS-assigned non-zero port", addr)
	}
	if port == defaultPort {
		t.Errorf("Addr() = %q, port must not be the hardcoded default with Port: 0", addr)
	}

	// Config.Port stays as the caller set it; Addr is the source of truth for
	// where the server actually is.
	if srv.config.Port != 0 {
		t.Errorf("config.Port = %d, want it left at the requested 0", srv.config.Port)
	}

	// First dial, no retries.
	soapClient := internalsoap.NewClient(&http.Client{Timeout: 5 * time.Second}, config.Username, config.Password)
	endpoint := "http://" + addr + config.BasePath + "/device_service"

	type getDeviceInformation struct {
		XMLName struct{} `xml:"tds:GetDeviceInformation"`
	}
	var resp struct {
		Manufacturer string `xml:"Manufacturer"`
	}
	if err := soapClient.Call(context.Background(), endpoint, "", getDeviceInformation{}, &resp); err != nil {
		t.Fatalf("first request after readiness failed: %v", err)
	}
	if resp.Manufacturer != config.DeviceInfo.Manufacturer {
		t.Errorf("Manufacturer = %q, want %q", resp.Manufacturer, config.DeviceInfo.Manufacturer)
	}
}

// TestAddrBeforeStart asserts Addr reports "" rather than a misleading
// config-derived guess before anything is bound.
func TestAddrBeforeStart(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if got := srv.Addr(); got != "" {
		t.Errorf("Addr() before Start = %q, want \"\"", got)
	}
}

// TestOutputQuietByDefault is the regression test for the stdout problem: a
// program whose stdout carries protocol data must not have banner text
// injected into it just because it imported this package.
func TestOutputQuietByDefault(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if got := srv.output(); got != io.Discard {
		t.Errorf("output() with Config.Output unset = %v, want io.Discard", got)
	}

	// And printBanner must actually respect it rather than reaching for
	// stdout directly.
	srv.printBanner("127.0.0.1:1")
}

// TestOutputWritesBannerWhenSet asserts the banner still gets produced for
// console callers, and lands on the caller's writer rather than stdout.
func TestOutputWritesBannerWhenSet(t *testing.T) {
	var buf bytes.Buffer
	config := createTestConfig()
	config.Output = &buf
	config.SupportPTZ = true
	config.SupportImaging = true
	config.SupportEvents = true

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	const addr = "127.0.0.1:12345"
	srv.printBanner(addr)

	got := buf.String()
	for _, want := range []string{
		addr,
		"/device_service",
		"/media_service",
		"/ptz_service",
		"/imaging_service",
		"/events_service",
		config.Profiles[0].Name,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("banner missing %q\nbanner:\n%s", want, got)
		}
	}
}

// TestBannerReportsBoundAddressNotRequested is why printBanner takes an
// address instead of reading config: with Port: 0 the requested address ends
// in ":0", which would print a URL nobody can use.
func TestBannerReportsBoundAddressNotRequested(t *testing.T) {
	var buf bytes.Buffer
	config := createTestConfig()
	config.Output = &buf

	_, addr := startTestServer(t, config)

	got := buf.String()
	if !strings.Contains(got, addr) {
		t.Errorf("banner does not mention the bound address %q\nbanner:\n%s", addr, got)
	}
	if strings.Contains(got, ":0/") || strings.Contains(got, ":0\n") {
		t.Errorf("banner printed the requested :0 rather than the bound port\nbanner:\n%s", got)
	}
}

// TestStartReturnsBindError asserts a failed bind comes back as Start's own
// error. Previously the listen happened inside a goroutine, so this arrived
// over a channel after Start had already appeared to succeed.
func TestStartReturnsBindError(t *testing.T) {
	// Occupy a port, then ask a second server for the same one.
	first := createTestConfig()
	_, addr := startTestServer(t, first)

	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort(%q) failed: %v", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Atoi(%q) failed: %v", portStr, err)
	}

	second := createTestConfig()
	second.Port = port

	srv, err := New(second)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Start(ctx); err == nil {
		t.Error("Start() on an occupied port returned nil, want a bind error")
	}
}

// TestReadyClosedOnceAcrossRestarts guards the double-close panic: Config.Ready
// is a channel the caller owns, and closing it twice would take the whole
// process down rather than just failing a call.
func TestReadyClosedOnceAcrossRestarts(t *testing.T) {
	ready := make(chan struct{})
	config := createTestConfig()
	config.Port = 0
	config.Ready = ready

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	for i := range 2 {
		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error, 1)
		go func() { errCh <- srv.Start(ctx) }()

		select {
		case <-ready:
		case err := <-errCh:
			t.Fatalf("run %d: Start() returned before readiness: %v", i, err)
		case <-time.After(10 * time.Second):
			t.Fatalf("run %d: never signaled readiness", i)
		}

		cancel()
		if err := <-errCh; err != nil {
			t.Fatalf("run %d: Start() returned %v", i, err)
		}
	}
}
