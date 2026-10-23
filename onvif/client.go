package onvif

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/analytics"
	"github.com/mickeyzzc/onvif-go/v2/device"
	"github.com/mickeyzzc/onvif-go/v2/deviceio"
	"github.com/mickeyzzc/onvif-go/v2/events"
	"github.com/mickeyzzc/onvif-go/v2/imaging"
	"github.com/mickeyzzc/onvif-go/v2/internal/soap"
	"github.com/mickeyzzc/onvif-go/v2/media"
	"github.com/mickeyzzc/onvif-go/v2/media2"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
	"github.com/mickeyzzc/onvif-go/v2/security"
)

// Default client configuration constants.
const (
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
	// DefaultIdleConnTimeout is the default idle connection timeout.
	DefaultIdleConnTimeout = 90 * time.Second
	// DefaultMaxIdleConns is the default maximum idle connections.
	DefaultMaxIdleConns = 10
	// DefaultMaxIdleConnsPerHost is the default maximum idle connections per host.
	DefaultMaxIdleConnsPerHost = 5
	// NonceSize is the size of the nonce for digest authentication.
	NonceSize = 16
)

// AuthMode selects how the client authenticates its SOAP requests. Real-world
// device compatibility varies wildly per firmware and even per service on the
// same device, which is why the client also supports an auth fallback ladder
// (see WithAuthFallback).
type AuthMode string

const (
	// AuthDigest is the ONVIF default: WS-Security UsernameToken with
	// PasswordDigest. Behavior matches the historical client exactly.
	AuthDigest AuthMode = "digest"

	// AuthPasswordText sends the password cleartext inside the WS-Security
	// UsernameToken. Some firmwares reject digest on specific services (PTZ,
	// GetUsers) while accepting this.
	AuthPasswordText AuthMode = "password-text"

	// AuthHTTPBasic authenticates with an HTTP Basic Authorization header and
	// no WS-Security header.
	AuthHTTPBasic AuthMode = "http-basic"

	// AuthNone sends no authentication at all. Minimal embedded devices
	// (e.g. ESP32 firmwares) may reject every auth-bearing request.
	AuthNone AuthMode = "none"
)

// validate reports whether the mode is one of the known AuthMode values.
func (m AuthMode) validate() error {
	switch m {
	case AuthDigest, AuthPasswordText, AuthHTTPBasic, AuthNone:
		return nil
	default:
		return fmt.Errorf("%w: unknown auth mode %q", ErrInvalidParameter, string(m))
	}
}

// soapMode maps the public AuthMode onto the internal soap layer enum.
func (m AuthMode) soapMode() soap.AuthMode {
	switch m {
	case AuthPasswordText:
		return soap.AuthModePasswordText
	case AuthHTTPBasic:
		return soap.AuthModeHTTPBasic
	case AuthNone:
		return soap.AuthModeNone
	default:
		return soap.AuthModeDigest
	}
}

// Client represents an ONVIF client for communicating with IP cameras.
//
// A Client is safe for concurrent use by multiple goroutines: every mutable
// field (credentials, clock skew, auth ladder state, capabilities cache,
// service endpoints) is guarded by an internal mutex, and each operation
// builds its own stateless SOAP exchange over the shared *http.Client. Callers
// sharing one Client across goroutines (recording, snapshot, PTZ components)
// do not need external locking.
type Client struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
	mu         sync.RWMutex

	// Service endpoints
	mediaEndpoint     string
	ptzEndpoint       string
	imagingEndpoint   string
	eventEndpoint     string
	analyticsEndpoint string
	media2Endpoint    string

	// Auth configuration. authMode is the primary mode; authFallback lists
	// modes to try (in order) when the primary fails with an auth-class error.
	// stickyAuth/stickySet remember the first mode the ladder found working so
	// later calls skip the ladder's failed attempts.
	authMode     AuthMode
	authFallback []AuthMode
	stickyAuth   AuthMode
	stickySet    bool

	// autoClockSkew makes Initialize measure and apply the device clock skew
	// before its first authenticated call (WithAutoClockSkew).
	autoClockSkew bool

	// Long-lived service instances (v2): accessors return the same pointers,
	// so services may hold their own state — the capabilities cache lives on
	// the device service.
	deviceSvc    *device.Service
	mediaSvc     *media.Service
	ptzSvc       *ptz.Service
	imagingSvc   *imaging.Service
	eventsSvc    *events.Service
	analyticsSvc *analytics.Service
	media2Svc    *media2.Service
	deviceioSvc  *deviceio.Service
	securitySvc  *security.Service

	// minimalCapsFallback configures the device service's cached-capabilities
	// degradation (WithMinimalCapsFallback).
	minimalCapsFallback bool

	// clockSkew is the offset (deviceTime - localTime) applied to WS-Security
	// digest timestamps. Set via SetClockSkew after measuring the device's clock
	// via GetSystemDateAndTime. Fixes Hikvision time-skew auth rejections.
	clockSkew time.Duration
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
// WARNING: Only use this for testing or with trusted cameras on private networks.
func WithInsecureSkipVerify() ClientOption {
	return func(c *Client) {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = &tls.Config{ //nolint:gosec // InsecureSkipVerify is intentional for testing
				}
			}
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}
}

// WithMinimalCapsFallback makes the cached capabilities accessor degrade
// gracefully: when GetCapabilities fails (minimal embedded devices can fault
// on it), a minimal all-off capability set is returned AND cached instead of
// an error. Callers gate advanced calls on it (no PTZ capability → no PTZ
// calls), and a weak device is never hammered with retries. Without this
// option the conservative default applies: failures surface as errors and
// are not cached.
func WithMinimalCapsFallback() ClientOption {
	return func(c *Client) {
		c.minimalCapsFallback = true
	}
}

// WithCredentials sets the authentication credentials.
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// WithAuthMode sets the primary authentication mode (default AuthDigest,
// which preserves the historical behavior). Unknown modes are rejected by
// NewClient.
func WithAuthMode(mode AuthMode) ClientOption {
	return func(c *Client) {
		c.authMode = mode
	}
}

// WithAuthFallback configures an authentication fallback ladder: when a call
// fails with an auth-class error (HTTP 401/403 or a NotAuthorized SOAP fault)
// under the primary mode, the listed modes are tried in order. The first mode
// that works is remembered (sticky) so subsequent calls do not pay the ladder
// cost again; changing credentials clears the memory. Setting a fallback also
// makes errors.Is(err, ErrUnauthorized) meaningful for exhausted ladders.
func WithAuthFallback(modes ...AuthMode) ClientOption {
	return func(c *Client) {
		c.authFallback = append(c.authFallback, modes...)
	}
}

// NewClient creates a new ONVIF client
// The endpoint can be provided in multiple formats:
//   - Full URL: "http://192.168.1.100/onvif/device_service"
//   - IP with port: "192.168.1.100:80" (http assumed, /onvif/device_service added)
//   - IP only: "192.168.1.100" (http://IP:80/onvif/device_service used)
func NewClient(endpoint string, opts ...ClientOption) (*Client, error) {
	// Normalize endpoint to full URL
	normalizedEndpoint, err := normalizeEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	client := &Client{
		endpoint: normalizedEndpoint,
		authMode: AuthDigest,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        DefaultMaxIdleConns,
				MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
				IdleConnTimeout:     DefaultIdleConnTimeout,
			},
			// Don't follow redirects automatically
			// This prevents http:// from being silently upgraded to https://
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	if err := client.validateAuthConfig(); err != nil {
		return nil, err
	}

	client.deviceSvc = device.NewWithFallback(client, client.minimalCapsFallback)
	client.mediaSvc = media.New(client)
	client.ptzSvc = ptz.New(client)
	client.imagingSvc = imaging.New(client)
	client.eventsSvc = events.New(client)
	client.analyticsSvc = analytics.New(client)
	client.media2Svc = media2.New(client)
	client.deviceioSvc = deviceio.New(client)
	client.securitySvc = security.New(client)

	return client, nil
}

// validateAuthConfig checks the auth mode and fallback ladder configured via
// options, deduplicating repeated fallback entries.
func (c *Client) validateAuthConfig() error {
	if err := c.authMode.validate(); err != nil {
		return fmt.Errorf("invalid auth configuration: %w", err)
	}

	seen := map[AuthMode]bool{c.authMode: true}
	deduped := make([]AuthMode, 0, len(c.authFallback))
	for _, mode := range c.authFallback {
		if err := mode.validate(); err != nil {
			return fmt.Errorf("invalid auth fallback: %w", err)
		}

		if seen[mode] {
			continue
		}

		seen[mode] = true
		deduped = append(deduped, mode)
	}

	c.authFallback = deduped

	return nil
}

// normalizeEndpoint converts various endpoint formats to a full ONVIF URL.
func normalizeEndpoint(endpoint string) (string, error) {
	// Check if endpoint starts with a scheme
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		// Parse as full URL
		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to parse endpoint URL: %w", err)
		}
		if parsedURL.Host == "" {
			return "", fmt.Errorf("%w", ErrURLMissingHost)
		}
		// If path is empty or just "/", add default ONVIF path
		if parsedURL.Path == "" || parsedURL.Path == "/" {
			parsedURL.Path = "/onvif/device_service"
		}

		return parsedURL.String(), nil
	}

	// No scheme - treat as IP, IP:port, hostname, or hostname:port
	// Add http:// scheme and validate
	fullURL := "http://" + endpoint + "/onvif/device_service"
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid IP address or hostname: %w", err)
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("%w", ErrInvalidEndpointFormat)
	}

	return fullURL, nil
}

// fixServiceURL corrects capability XAddr URLs that a camera reports
// incorrectly. Two cases are handled:
//
//  1. Localhost/loopback (127.0.0.1, 0.0.0.0, localhost, ::1): some cameras
//     report these instead of their real IP.
//
//  2. Stale routable IP: after a camera's IP changes (DHCP reassignment, AP
//     roaming), the camera's internal network config may lag and it continues
//     to advertise its OLD IP in GetCapabilities XAddrs (e.g. it is reachable
//     at .199 but reports media_service at .200). Since we reached the camera
//     via c.endpoint (the device_service URL), c.endpoint's host is guaranteed
//     to be the current correct address — so when a capability XAddr host
//     disagrees with it, we rewrite the XAddr host to match c.endpoint's host.
//
// In both cases the port from the service URL is preserved (the camera's
// service-specific port is authoritative); if the service URL omits a port,
// the client endpoint's port is used as a fallback.
func (c *Client) fixServiceURL(serviceURL string) string {
	if serviceURL == "" {
		return serviceURL
	}

	// Parse the service URL.
	parsedService, err := url.Parse(serviceURL)
	if err != nil {
		return serviceURL // Return original if parsing fails.
	}

	// Parse the client's endpoint to get the authoritative camera address.
	parsedClient, err := url.Parse(c.endpoint)
	if err != nil {
		return serviceURL // Cannot determine the correct host.
	}

	serviceHost := parsedService.Hostname()
	clientHost := parsedClient.Hostname()

	// Determine whether the service URL host needs correction. It needs
	// correction if it is a loopback address OR if it differs from the
	// device_service host we used to reach the camera (stale advertised IP).
	needsFix := isLoopbackHost(serviceHost) || serviceHost != clientHost
	if !needsFix {
		return serviceURL
	}

	// Replace the host with the client endpoint's (authoritative) host,
	// preserving the service URL's port if specified.
	servicePort := parsedService.Port()
	if servicePort != "" {
		parsedService.Host = clientHost + ":" + servicePort
	} else {
		parsedService.Host = clientHost
		// Use client's port if the service doesn't specify one.
		if clientPort := parsedClient.Port(); clientPort != "" {
			parsedService.Host = clientHost + ":" + clientPort
		}
	}

	return parsedService.String()
}

// isLoopbackHost reports whether the given host is a loopback/localhost address.
func isLoopbackHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "::1"
}

// Initialize discovers and initializes service endpoints. With
// WithAutoClockSkew configured, the device clock is measured (via an
// unauthenticated GetSystemDateAndTime) and applied BEFORE the capabilities
// call, so the very first authenticated request already carries
// device-correct digest timestamps; measurement failure is silently skipped.
func (c *Client) Initialize(ctx context.Context) error {
	c.mu.RLock()
	autoSkew := c.autoClockSkew
	c.mu.RUnlock()

	if autoSkew {
		if skew, err := c.MeasureClockSkew(ctx); err == nil {
			c.SetClockSkew(skew)
		}
	}

	// Get device information and capabilities
	capabilities, err := c.Device().GetCapabilities(ctx)
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %w", err)
	}

	// Probe GetServices for the ver20/media endpoint before taking the
	// write lock (the call itself needs the read lock). Media2-only
	// (Profile T) devices have no ver10 Media XAddr in GetCapabilities,
	// and dual-face devices advertise a DEDICATED ver20/media address —
	// both need the GetServices entry, which is authoritative for Media2
	// (the Media1 endpoint dispatches by action local name, so a tr2
	// request posted there gets the Media1-shaped answer). A probe
	// failure is non-fatal: Initialize tolerates partial endpoint
	// knowledge, matching the capabilities handling above.
	var media2FromServices string
	c.mu.RLock()
	media2Pinned := c.media2Endpoint != ""
	c.mu.RUnlock()
	if !media2Pinned {
		if services, serr := c.Device().GetServices(ctx, false); serr == nil {
			for _, svc := range services {
				if svc.Namespace == media2.Namespace && svc.XAddr != "" {
					media2FromServices = svc.XAddr
					break
				}
			}
		}
	}

	// Extract service endpoints and fix incorrect addresses (localhost or stale
	// advertised IPs after the camera roamed to a new address). The writes are
	// lock-guarded so Initialize is safe to run while other goroutines call
	// through the same client (issue #12).
	c.mu.Lock()
	defer c.mu.Unlock()

	if capabilities.Media != nil && capabilities.Media.XAddr != "" {
		c.mediaEndpoint = c.fixServiceURL(capabilities.Media.XAddr)
	}
	if capabilities.PTZ != nil && capabilities.PTZ.XAddr != "" {
		c.ptzEndpoint = c.fixServiceURL(capabilities.PTZ.XAddr)
	}
	if capabilities.Imaging != nil && capabilities.Imaging.XAddr != "" {
		c.imagingEndpoint = c.fixServiceURL(capabilities.Imaging.XAddr)
	}
	if capabilities.Events != nil && capabilities.Events.XAddr != "" {
		c.eventEndpoint = c.fixServiceURL(capabilities.Events.XAddr)
	}
	if capabilities.Analytics != nil && capabilities.Analytics.XAddr != "" {
		c.analyticsEndpoint = c.fixServiceURL(capabilities.Analytics.XAddr)
	}
	if media2FromServices != "" {
		// GetServices carried the ver20/media XAddr — authoritative for
		// dual-face and Media2-only devices alike.
		c.media2Endpoint = c.fixServiceURL(media2FromServices)
	} else if c.media2Endpoint == "" && c.mediaEndpoint != "" {
		// Legacy single-face device: Media2 rides the media endpoint.
		c.media2Endpoint = c.mediaEndpoint
	}

	return nil
}

// InvalidateCapabilitiesCache drops the cached capabilities so the next
// GetCapabilitiesCached call re-fetches them. Call after a firmware upgrade
// or any change that could alter the device's advertised services.
func (c *Client) InvalidateCapabilitiesCache() {
	c.deviceSvc.InvalidateCapsCache()
}

// Endpoint returns the device endpoint.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// SetCredentials updates the authentication credentials. Any auth-mode
// conclusion remembered by the fallback ladder is cleared, since the ladder's
// stickiness is only valid for the credentials it was measured with.
func (c *Client) SetCredentials(username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.username = username
	c.password = password
	c.stickySet = false
	c.stickyAuth = ""
}

// ResetAuthLadder clears the remembered (sticky) auth mode so the next call
// re-runs the full fallback ladder. Useful after device-side credential or
// firmware changes.
func (c *Client) ResetAuthLadder() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stickySet = false
	c.stickyAuth = ""
}

// GetCredentials returns the current credentials.
func (c *Client) GetCredentials() (username, password string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.username, c.password
}

// AuthMode returns the primary authentication mode (not the sticky ladder
// result; use AuthLadderMode for the effective mode).
func (c *Client) AuthMode() AuthMode {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.authMode
}

// AuthLadderMode returns the auth mode currently in effect: the sticky ladder
// result when one was established, otherwise the configured primary mode.
func (c *Client) AuthLadderMode() AuthMode {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.stickySet {
		return c.stickyAuth
	}

	return c.authMode
}

// SetClockSkew sets the clock offset (deviceTime - localTime) used for WS-Security
// digest auth timestamps. Call this after measuring the device's clock via
// GetSystemDateAndTime to fix auth failures from clock divergence (Hikvision).
func (c *Client) SetClockSkew(skew time.Duration) {
	c.mu.Lock()
	c.clockSkew = skew
	c.mu.Unlock()
}

// Call performs a SOAP call through the client's auth configuration: it
// applies the configured mode, retries through the fallback ladder on
// auth-class errors, and remembers the first working mode. All service
// operations go through this single audited path (issue #12).
//
// Empty 200 response bodies also advance the ladder (issue #140): old
// firmwares answer a rejected digest with an empty body instead of a
// 401/fault, so the next auth mode must still get its chance.
//
// Auth-class failures are wrapped so errors.Is(err, ErrUnauthorized) holds
// regardless of how many modes were tried; other failures propagate unchanged.
func (c *Client) Call(ctx context.Context, endpoint, action string, request, response interface{}) error {
	modes := c.authLadder()

	var lastErr error

	for i, mode := range modes {
		lastErr = c.callWithMode(ctx, endpoint, action, request, response, mode)
		if lastErr == nil {
			if i > 0 {
				c.rememberStickyAuth(mode)
			}

			return nil
		}

		if !soap.IsAuthLadderRetryable(lastErr) {
			// Failures that are neither auth-class nor empty-bodied are never
			// retried with another mode.
			return lastErr
		}
	}

	if !soap.IsAuthFailure(lastErr) {
		// Ladder exhausted on a non-auth failure (e.g. empty response body):
		// the original cause must stay diagnosable, so no unauthorized wrap.
		return lastErr
	}

	return fmt.Errorf("%w: %w", ErrUnauthorized, lastErr)
}

// callWithMode executes one SOAP attempt with a specific auth mode.
func (c *Client) callWithMode(ctx context.Context, endpoint, action string, request, response interface{}, mode AuthMode) error {
	c.mu.RLock()
	username, password := c.username, c.password
	skew := c.clockSkew
	httpClient := c.httpClient
	c.mu.RUnlock()

	sc := soap.NewClient(httpClient, username, password)
	sc.SetClockSkew(skew)
	sc.SetAuthMode(mode.soapMode())

	return sc.Call(ctx, endpoint, action, request, response)
}

// authLadder snapshots the modes to try: the sticky result when established,
// otherwise the primary mode followed by the (already validated, deduplicated)
// fallback list.
func (c *Client) authLadder() []AuthMode {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.stickySet {
		return []AuthMode{c.stickyAuth}
	}

	if len(c.authFallback) == 0 {
		return []AuthMode{c.authMode}
	}

	modes := make([]AuthMode, 0, 1+len(c.authFallback))
	modes = append(modes, c.authMode)

	return append(modes, c.authFallback...)
}

// rememberStickyAuth records the first ladder mode that succeeded.
func (c *Client) rememberStickyAuth(mode AuthMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stickyAuth = mode
	c.stickySet = true
}
