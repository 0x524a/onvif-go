package onvif

import (
	"context"
	"crypto/sha1" //nolint:gosec // SHA1 required for ONVIF digest auth
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// authFlavor describes how a test server decides which authentication a
// request carried, mirroring the real-device compatibility matrix from
// issue #1.
type authFlavor string

const (
	flavorDigestOnly  authFlavor = "digest-only"
	flavorTextOnly    authFlavor = "text-only"
	flavorBasicOnly   authFlavor = "basic-only"
	flavorNoneOnly    authFlavor = "none-only"
	flavorRejectAll   authFlavor = "reject-all"
	flavorEmptyDigest authFlavor = "empty-on-digest"
	flavorEmptyAll    authFlavor = "empty-always"
)

// classifiedAuth records what one incoming request looked like.
type classifiedAuth struct {
	hasDigest bool
	hasText   bool
	hasBasic  bool
	hasNone   bool
}

// classify inspects a SOAP request and reports which auth shapes it carried.
func classify(r *http.Request) classifiedAuth {
	var c classifiedAuth

	body, _ := io.ReadAll(r.Body)
	bodyStr := string(body)

	c.hasDigest = strings.Contains(bodyStr, "#PasswordDigest")
	c.hasText = strings.Contains(bodyStr, "#PasswordText") &&
		!strings.Contains(bodyStr, "#PasswordDigest")
	if user, pass, ok := r.BasicAuth(); ok && user != "" && pass != "" {
		c.hasBasic = true
	}

	c.hasNone = !c.hasDigest && !c.hasText && !c.hasBasic

	return c
}

// deviceInfoBody is a minimal GetDeviceInformationResponse envelope.
const deviceInfoBody = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Manufacturer>TestCam</tds:Manufacturer>
      <tds:Model>T-1000</tds:Model>
      <tds:FirmwareVersion>1.0</tds:FirmwareVersion>
      <tds:SerialNumber>SER-42</tds:SerialNumber>
    </tds:GetDeviceInformationResponse>
  </s:Body>
</s:Envelope>`

const notAuthorizedFaultBody = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <s:Fault>
      <s:Code>
        <s:Value>s:Sender</s:Value>
        <s:Subcode>
          <s:Value>ter:NotAuthorized</s:Value>
        </s:Subcode>
      </s:Code>
      <s:Reason><s:Text xml:lang="en">Sender not authorized</s:Text></s:Reason>
    </s:Fault>
  </s:Body>
</s:Envelope>`

// newAuthTestServer starts a device that accepts exactly one auth flavor and
// counts the classified attempts it saw.
func newAuthTestServer(t *testing.T, flavor authFlavor) (*httptest.Server, *authAttemptLog) {
	t.Helper()

	attempts := &authAttemptLog{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := classify(r)
		attempts.add(c)

		accepted := false
		switch flavor {
		case flavorDigestOnly:
			accepted = c.hasDigest
		case flavorTextOnly:
			accepted = c.hasText
		case flavorBasicOnly:
			accepted = c.hasBasic
		case flavorNoneOnly:
			accepted = c.hasNone
		case flavorRejectAll:
			accepted = false
		case flavorEmptyDigest:
			// 2016-era firmware (issue #140): a rejected digest gets HTTP 200
			// with an empty body — no fault, no status hint.
			if c.hasDigest {
				w.WriteHeader(http.StatusOK)

				return
			}
			accepted = true
		case flavorEmptyAll:
			w.WriteHeader(http.StatusOK)

			return
		}

		if !accepted {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(notAuthorizedFaultBody))

			return
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(deviceInfoBody))
	}))
	t.Cleanup(server.Close)

	return server, attempts
}

// authAttemptLog is a concurrency-safe log of classified requests.
type authAttemptLog struct {
	mu   sync.Mutex
	logs []classifiedAuth
}

func (l *authAttemptLog) add(c classifiedAuth) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, c)
}

func (l *authAttemptLog) snapshot() []classifiedAuth {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append([]classifiedAuth(nil), l.logs...)
}

// newAuthTestClient builds a client pointed at the test server.
func newAuthTestClient(t *testing.T, server *httptest.Server, opts ...ClientOption) *Client {
	t.Helper()

	allOpts := append([]ClientOption{
		WithCredentials("admin", "secret"),
		WithHTTPClient(server.Client()),
	}, opts...)

	client, err := NewClient(server.URL, allOpts...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func TestAuthDefaultBehaviorIsDigest(t *testing.T) {
	server, attempts := newAuthTestServer(t, flavorDigestOnly)
	client := newAuthTestClient(t, server)

	info, err := client.Device().GetDeviceInformation(context.Background())
	if err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	if info.Manufacturer != "TestCam" {
		t.Errorf("Manufacturer = %q, want %q", info.Manufacturer, "TestCam")
	}

	for i, a := range attempts.snapshot() {
		if !a.hasDigest {
			t.Errorf("attempt %d was not digest-authenticated: %+v", i, a)
		}
	}
}

func TestAuthModePasswordText(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorTextOnly)
	client := newAuthTestClient(t, server, WithAuthMode(AuthPasswordText))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}
}

func TestAuthModeHTTPBasic(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorBasicOnly)
	client := newAuthTestClient(t, server, WithAuthMode(AuthHTTPBasic))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}
}

func TestAuthModeNoneSendsNoAuthHeaders(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorNoneOnly)
	// Credentials remain configured; AuthNone must suppress them anyway
	// (ESP32-style devices reject every auth-bearing request).
	client := newAuthTestClient(t, server, WithAuthMode(AuthNone))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}
}

func TestAuthFallbackLadderAndSticky(t *testing.T) {
	// Device accepts only HTTP Basic; the ladder is digest -> basic.
	server, attempts := newAuthTestServer(t, flavorBasicOnly)
	client := newAuthTestClient(t, server,
		WithAuthFallback(AuthHTTPBasic, AuthNone))

	// First call: digest rejected, basic accepted.
	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() (ladder) error = %v", err)
	}

	first := attempts.snapshot()
	if len(first) != 2 { //nolint:mnd // digest attempt + basic attempt
		t.Fatalf("first call made %d attempts, want 2: %+v", len(first), first)
	}
	if !first[0].hasDigest || !first[1].hasBasic {
		t.Errorf("ladder order wrong: %+v", first)
	}

	// Sticky: the second call must go straight to basic.
	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() (sticky) error = %v", err)
	}

	if got := client.AuthLadderMode(); got != AuthHTTPBasic {
		t.Errorf("AuthLadderMode() = %q, want %q", got, AuthHTTPBasic)
	}

	second := attempts.snapshot()[len(first):]
	if len(second) != 1 || !second[0].hasBasic {
		t.Errorf("sticky call made %+v, want a single basic attempt", second)
	}
}

func TestAuthFallbackExhaustedIsErrUnauthorized(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorRejectAll)
	client := newAuthTestClient(t, server,
		WithAuthFallback(AuthPasswordText, AuthHTTPBasic))

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want auth failure")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestAuthFallbackAdvancesPastEmptyResponseBody(t *testing.T) {
	// Old firmware answers a digest request with 200 + empty body instead of
	// a 401/fault (issue #140); the ladder must still reach PasswordText.
	server, attempts := newAuthTestServer(t, flavorEmptyDigest)
	client := newAuthTestClient(t, server, WithAuthFallback(AuthPasswordText))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	if got := client.AuthLadderMode(); got != AuthPasswordText {
		t.Errorf("AuthLadderMode() = %q, want sticky %q", got, AuthPasswordText)
	}

	logs := attempts.snapshot()
	if len(logs) != 2 { //nolint:mnd // empty-bodied digest + password-text
		t.Fatalf("call made %d attempts, want 2: %+v", len(logs), logs)
	}
	if !logs[0].hasDigest || !logs[1].hasText {
		t.Errorf("ladder order wrong: %+v", logs)
	}
}

func TestAuthLadderEmptyBodyExhaustedKeepsOriginalError(t *testing.T) {
	// Every rung gets 200 + empty body: the exhausted ladder must surface the
	// empty-body cause, not a misleading ErrUnauthorized (issue #140).
	server, attempts := newAuthTestServer(t, flavorEmptyAll)
	client := newAuthTestClient(t, server,
		WithAuthFallback(AuthPasswordText, AuthHTTPBasic))

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want empty-body failure")
	}

	if !errors.Is(err, ErrEmptyResponseBody) {
		t.Errorf("errors.Is(err, ErrEmptyResponseBody) = false, err = %v", err)
	}
	if errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = true, err = %v", err)
	}

	if got := len(attempts.snapshot()); got != 3 { //nolint:mnd // one per ladder rung
		t.Errorf("ladder made %d attempts, want 3", got)
	}
}

func TestEmptyBodyWithoutFallbackPropagatesUnchanged(t *testing.T) {
	// No fallback rungs: the empty-body error propagates as-is, still not
	// classified as unauthorized (issue #140).
	server, _ := newAuthTestServer(t, flavorEmptyDigest)
	client := newAuthTestClient(t, server)

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want empty-body failure")
	}

	if !errors.Is(err, ErrEmptyResponseBody) {
		t.Errorf("errors.Is(err, ErrEmptyResponseBody) = false, err = %v", err)
	}
	if errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = true, err = %v", err)
	}
}

func TestErrUnauthorizedOnPlainHTTP401(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorRejectAll)
	client := newAuthTestClient(t, server) // no fallback configured

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want auth failure")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestErrUnauthorizedOnFaultWithHTTP200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(notAuthorizedFaultBody)) // HTTP 200 + fault
	}))
	t.Cleanup(server.Close)

	client := newAuthTestClient(t, server)

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want fault error")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestVoidOperationSurfacesFaultWithHTTP200(t *testing.T) {
	// Regression: void operations (response struct without payload fields)
	// previously reported success for 200-with-Fault responses.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(notAuthorizedFaultBody))
	}))
	t.Cleanup(server.Close)

	client := newAuthTestClient(t, server, WithAuthMode(AuthNone))

	err := client.Events().Unsubscribe(context.Background(), server.URL+"/subscription")
	if err == nil {
		t.Fatal("Unsubscribe() succeeded on a fault response, want error")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestSetCredentialsResetsStickyAuth(t *testing.T) {
	server, attempts := newAuthTestServer(t, flavorBasicOnly)
	client := newAuthTestClient(t, server, WithAuthFallback(AuthHTTPBasic))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	if got := client.AuthLadderMode(); got != AuthHTTPBasic {
		t.Fatalf("AuthLadderMode() = %q before reset, want %q", got, AuthHTTPBasic)
	}

	client.SetCredentials("admin", "newsecret")
	if got := client.AuthLadderMode(); got != AuthDigest {
		t.Errorf("AuthLadderMode() = %q after SetCredentials, want primary %q", got, AuthDigest)
	}

	// Ladder must run again for the new credentials.
	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	logs := attempts.snapshot()
	if len(logs) < 3 { //nolint:mnd // digest+basic, then digest+basic again
		t.Errorf("expected ladder to re-run after credential change, attempts = %+v", logs)
	}
}

func TestNewClientRejectsUnknownAuthMode(t *testing.T) {
	_, err := NewClient("http://192.168.1.100", WithAuthMode("bogus"))
	if err == nil {
		t.Fatal("NewClient() accepted unknown auth mode, want error")
	}

	_, err = NewClient("http://192.168.1.100", WithAuthFallback("bogus"))
	if err == nil {
		t.Fatal("NewClient() accepted unknown fallback mode, want error")
	}
}

func TestNonAuthErrorsAreNotRetried(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	t.Cleanup(server.Close)

	client := newAuthTestClient(t, server, WithAuthFallback(AuthHTTPBasic))

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want error")
	}

	if calls != 1 {
		t.Errorf("server called %d times, want 1 (non-auth errors must not re-run the ladder)", calls)
	}

	if errors.Is(err, ErrUnauthorized) {
		t.Errorf("500 must not classify as ErrUnauthorized, err = %v", err)
	}
}

// skewDevice is a test camera whose clock runs `offset` away from real time
// and which fully verifies WS-Security digests (digest value AND Created
// within its ±5min replay window, evaluated against the DEVICE clock).
type skewDevice struct {
	offset     time.Duration // deviceTime = now + offset
	password   string
	rttDelay   time.Duration // simulated processing delay before answering
	alwaysFail bool          // reject every authenticated call regardless
}

var (
	reUsername = regexp.MustCompile(`<Username>([^<]*)</Username>`)
	rePassword = regexp.MustCompile(`<Password[^>]*>([^<]*)</Password>`)
	reNonce    = regexp.MustCompile(`<Nonce[^>]*>([^<]*)</Nonce>`)
	reCreated  = regexp.MustCompile(`<Created[^>]*>([^<]*)</Created>`)
)

func (d *skewDevice) handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	bodyStr := string(body)
	deviceNow := time.Now().UTC().Add(d.offset)

	if strings.Contains(bodyStr, "GetSystemDateAndTime") {
		// Sample the device clock at the RTT midpoint (half the delay before,
		// half after): the library's midpoint compensation then measures ~0
		// error even when CI scheduling stretches the round trip, because the
		// stretch hits both halves symmetrically.
		if d.rttDelay > 0 {
			time.Sleep(d.rttDelay / 2)
		}

		deviceNow = time.Now().UTC().Add(d.offset)
		t := deviceNow

		if d.rttDelay > 0 {
			time.Sleep(d.rttDelay / 2)
		}
		_, _ = w.Write([]byte(xmlEnvelope(`
<tds:GetSystemDateAndTimeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:SystemDateAndTime>
<tds:UTCDateTime><tds:Time><tds:Hour>` + itoa(t.Hour()) + `</tds:Hour><tds:Minute>` +
			itoa(t.Minute()) + `</tds:Minute><tds:Second>` + itoa(t.Second()) +
			`</tds:Second></tds:Time><tds:Date><tds:Year>` + itoa(t.Year()) +
			`</tds:Year><tds:Month>` + itoa(int(t.Month())) + `</tds:Month><tds:Day>` +
			itoa(t.Day()) + `</tds:Day></tds:Date></tds:UTCDateTime>
</tds:SystemDateAndTime>
</tds:GetSystemDateAndTimeResponse>`)))

		return
	}

	if !d.alwaysFail && verifyDigest(bodyStr, deviceNow, d.password) {
		if strings.Contains(bodyStr, "GetCapabilities") {
			_, _ = w.Write([]byte(xmlEnvelope(`
<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Capabilities><tds:Device><tds:XAddr>http://device/onvif/device_service</tds:XAddr>
</tds:Device></tds:Capabilities>
</tds:GetCapabilitiesResponse>`)))

			return
		}

		_, _ = w.Write([]byte(xmlEnvelope(`
<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Manufacturer>SkewCam</tds:Manufacturer>
</tds:GetDeviceInformationResponse>`)))

		return
	}

	_, _ = w.Write([]byte(notAuthorizedFaultBody))
}

func xmlEnvelope(body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>` + body + `</s:Body></s:Envelope>`
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

// verifyDigest recomputes the WS-Security digest from the request body and
// checks it against the device password and clock window.
func verifyDigest(body string, deviceNow time.Time, password string) bool {
	user := reUsername.FindStringSubmatch(body)
	pwd := rePassword.FindStringSubmatch(body)
	nonce := reNonce.FindStringSubmatch(body)
	created := reCreated.FindStringSubmatch(body)
	if user == nil || pwd == nil || nonce == nil || created == nil {
		return false
	}

	createdTime, err := time.Parse(time.RFC3339, created[1])
	if err != nil {
		return false
	}

	// Replay window evaluated on the DEVICE clock (Hikvision behavior).
	if diff := createdTime.Sub(deviceNow); diff < -5*time.Minute || diff > 5*time.Minute {
		return false
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonce[1])
	if err != nil {
		return false
	}

	hash := sha1.New() //nolint:gosec // ONVIF digest algorithm
	hash.Write(nonceBytes)
	hash.Write([]byte(created[1]))
	hash.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)) == pwd[1]
}

func newSkewTestClient(t *testing.T, dev *skewDevice, opts ...ClientOption) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(dev.handler))
	t.Cleanup(server.Close)

	allOpts := append([]ClientOption{
		WithCredentials("admin", dev.password),
		WithHTTPClient(server.Client()),
	}, opts...)

	client, err := NewClient(server.URL, allOpts...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

// TestMeasureClockSkewRTTCompensation verifies that a device answering with a
// delay does not pollute the measurement: the skew must land near the true
// offset, not offset±RTT/2.
func TestMeasureClockSkewRTTCompensation(t *testing.T) {
	const trueOffset = 90 * time.Minute

	dev := &skewDevice{offset: trueOffset, password: "secret", rttDelay: 300 * time.Millisecond}
	client := newSkewTestClient(t, dev)

	skew, err := client.MeasureClockSkew(context.Background())
	if err != nil {
		t.Fatalf("MeasureClockSkew() error = %v", err)
	}

	if diff := skew - trueOffset; diff < -time.Second || diff > time.Second {
		t.Errorf("skew = %v, want ~%v (RTT compensation off by %v)", skew, trueOffset, diff)
	}
}

// TestMeasureClockSkewSafeFallbacks covers zero-time and malformed responses.
func TestMeasureClockSkewSafeFallbacks(t *testing.T) {
	bodies := map[string]string{
		"zero time": xmlEnvelope(`
<tds:GetSystemDateAndTimeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:SystemDateAndTime><tds:UTCDateTime><tds:Time><tds:Hour>0</tds:Hour>
</tds:Time><tds:Date><tds:Year>0</tds:Year></tds:Date></tds:UTCDateTime>
</tds:SystemDateAndTime></tds:GetSystemDateAndTimeResponse>`),
		"malformed": `<not-a-response/>`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(body))
			}))
			t.Cleanup(server.Close)

			client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			skew, err := client.MeasureClockSkew(context.Background())
			if err == nil {
				t.Fatalf("MeasureClockSkew() = %v, nil; want error", skew)
			}

			if skew != 0 {
				t.Errorf("skew = %v on failure, want 0", skew)
			}
		})
	}
}

// TestWithAutoClockSkewOnInitialize shows the end-to-end fix: a Hikvision-
// style clock-diverged camera rejects local-time digests; auto skew makes
// Initialize + calls succeed.
func TestWithAutoClockSkewOnInitialize(t *testing.T) {
	dev := &skewDevice{offset: 42 * time.Minute, password: "secret"}

	t.Run("without auto skew digest is rejected", func(t *testing.T) {
		client := newSkewTestClient(t, dev)
		if _, err := client.Device().GetDeviceInformation(context.Background()); err == nil {
			t.Fatal("GetDeviceInformation() succeeded against a 42-minute-diverged clock, want rejection")
		}
	})

	t.Run("with auto skew initialize and calls succeed", func(t *testing.T) {
		client := newSkewTestClient(t, dev, WithAutoClockSkew())

		if err := client.Initialize(context.Background()); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
			t.Fatalf("GetDeviceInformation() error = %v", err)
		}
	})
}

func TestDiagnoseAuth(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		client := newSkewTestClient(t, &skewDevice{password: "secret"})

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusOK {
			t.Errorf("Status = %q, want %q (%s)", diag.Status, AuthStatusOK, diag.Detail)
		}
	})

	t.Run("clock skew confirmed by device-time retry", func(t *testing.T) {
		client := newSkewTestClient(t, &skewDevice{offset: 10 * time.Minute, password: "secret"})

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusClockSkew {
			t.Fatalf("Status = %q, want %q (%s)", diag.Status, AuthStatusClockSkew, diag.Detail)
		}

		if diag.ClockSkew < 9*time.Minute || diag.ClockSkew > 11*time.Minute {
			t.Errorf("ClockSkew = %v, want ~10m", diag.ClockSkew)
		}
	})

	t.Run("bad credentials despite skew", func(t *testing.T) {
		// Device clock diverged AND digest always rejected: the retry with
		// device time must still fail, pointing at credentials.
		client := newSkewTestClient(t, &skewDevice{offset: 10 * time.Minute, password: "right", alwaysFail: true})
		client.SetCredentials("admin", "wrong")

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusBadCredentials {
			t.Errorf("Status = %q, want %q (%s)", diag.Status, AuthStatusBadCredentials, diag.Detail)
		}
	})

	t.Run("no significant skew means credentials", func(t *testing.T) {
		client := newSkewTestClient(t, &skewDevice{password: "right", alwaysFail: true})
		client.SetCredentials("admin", "wrong")

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusBadCredentials {
			t.Errorf("Status = %q, want %q (%s)", diag.Status, AuthStatusBadCredentials, diag.Detail)
		}
	})
}
