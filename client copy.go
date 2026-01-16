package onvif

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 used for ONVIF digest authentication
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
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

// Client represents an ONVIF client for communicating with IP cameras.
type Client struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
	mu         sync.RWMutex

	// Service endpoints
	mediaEndpoint   string
	ptzEndpoint     string
	imagingEndpoint string
	eventEndpoint   string
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

// WithCredentials sets the authentication credentials.
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) {
		c.username = username
		c.password = password
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

	return client, nil
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

// Some cameras incorrectly report localhost (127.0.0.1, 0.0.0.0, localhost) in their capability URLs.
func (c *Client) fixLocalhostURL(serviceURL string) string {
	if serviceURL == "" {
		return serviceURL
	}

	// Parse the service URL
	parsedService, err := url.Parse(serviceURL)
	if err != nil {
		return serviceURL // Return original if parsing fails
	}

	// Check if the service URL has a localhost/loopback address
	host := parsedService.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "::1" {
		// Parse the client's endpoint to get the actual camera address
		parsedClient, err := url.Parse(c.endpoint)
		if err != nil {
			return serviceURL // Return original if parsing fails
		}

		// Replace the host but keep the port from service URL if specified
		servicePort := parsedService.Port()
		if servicePort != "" {
			parsedService.Host = parsedClient.Hostname() + ":" + servicePort
		} else {
			parsedService.Host = parsedClient.Hostname()
			// Use client's port if service doesn't specify one
			if clientPort := parsedClient.Port(); clientPort != "" {
				parsedService.Host = parsedClient.Hostname() + ":" + clientPort
			}
		}

		return parsedService.String()
	}

	return serviceURL
}

// Initialize discovers and initializes service endpoints.
func (c *Client) Initialize(ctx context.Context) error {
	// Get device information and capabilities
	capabilities, err := c.GetCapabilities(ctx)
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %w", err)
	}

	// Extract service endpoints and fix any localhost addresses
	// Some cameras incorrectly report localhost instead of their actual IP
	if capabilities.Media != nil && capabilities.Media.XAddr != "" {
		c.mediaEndpoint = c.fixLocalhostURL(capabilities.Media.XAddr)
	}
	if capabilities.PTZ != nil && capabilities.PTZ.XAddr != "" {
		c.ptzEndpoint = c.fixLocalhostURL(capabilities.PTZ.XAddr)
	}
	if capabilities.Imaging != nil && capabilities.Imaging.XAddr != "" {
		c.imagingEndpoint = c.fixLocalhostURL(capabilities.Imaging.XAddr)
	}
	if capabilities.Events != nil && capabilities.Events.XAddr != "" {
		c.eventEndpoint = c.fixLocalhostURL(capabilities.Events.XAddr)
	}

	return nil
}

// Endpoint returns the device endpoint.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// SetCredentials updates the authentication credentials.
func (c *Client) SetCredentials(username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.username = username
	c.password = password
}

// GetCredentials returns the current credentials.
func (c *Client) GetCredentials() (username, password string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.username, c.password
}

// DownloadFile downloads a file from the given URL with authentication.
// Supports both Basic and Digest authentication (tries basic first, falls back to digest).
func (c *Client) DownloadFile(ctx context.Context, downloadURL string) ([]byte, error) {
	// Try basic auth first
	data, err := c.downloadWithBasicAuth(ctx, downloadURL)
	if err == nil {
		return data, nil
	}

	// If basic auth fails with 401, try digest auth
	if strings.Contains(err.Error(), "401") {
		digestData, digestErr := c.downloadWithDigestAuth(ctx, downloadURL)
		if digestErr == nil {
			return digestData, nil
		}
		// If digest auth also fails, return the original error
		if strings.Contains(digestErr.Error(), "401") {
			return nil, err // Return original error (both auth methods failed)
		}

		return nil, digestErr
	}

	return nil, err
}

// downloadWithBasicAuth performs an HTTP download with Basic authentication.
func (c *Client) downloadWithBasicAuth(ctx context.Context, downloadURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	req.Header.Set("User-Agent", "onvif-go-client")
	req.Header.Set("Connection", "close")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyPreview, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error preview - ignore read errors
		bodyStr := string(bodyPreview)
		const maxBodyPreview = 200
		if len(bodyStr) > maxBodyPreview {
			bodyStr = bodyStr[:maxBodyPreview] + "..."
		}

		// Base error message for programmatic use
		errorMsg := fmt.Sprintf("download failed with status code %d", resp.StatusCode)

		// Add structured error details
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			errorMsg += ": authentication failed (401 Unauthorized); basic auth failed, trying digest auth"
		case http.StatusForbidden:
			errorMsg += ": access denied (403 Forbidden); user may not have permission to download snapshots"
		case http.StatusNotFound:
			errorMsg += ": snapshot URI not found (404); camera may have revoked the URI, try getting a fresh snapshot URI"
		}

		if bodyStr != "" {
			errorMsg += fmt.Sprintf("; response: %s", bodyStr)
		}

		return nil, fmt.Errorf("%w: %s", ErrDownloadFailed, errorMsg)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// downloadWithDigestAuth performs an HTTP download with Digest authentication.
func (c *Client) downloadWithDigestAuth(ctx context.Context, downloadURL string) ([]byte, error) {
	if c.username == "" {
		return nil, fmt.Errorf("%w", ErrDigestAuthRequiresCredentials)
	}

	// Create a custom transport with digest auth
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   DefaultTimeout,
			KeepAlive: DefaultTimeout,
		}).Dial,
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,
	}

	// Create a custom HTTP client for digest auth
	digestClient := &http.Client{
		Transport: &digestAuthTransport{
			transport: tr,
			username:  c.username,
			password:  c.password,
		},
		Timeout: DefaultTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "onvif-go-client")
	req.Header.Set("Connection", "close")

	resp, err := digestClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("digest auth request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyPreview, _ := io.ReadAll(resp.Body) //nolint:errcheck // Error preview - ignore read errors
		bodyStr := string(bodyPreview)
		const maxBodyPreview = 200
		if len(bodyStr) > maxBodyPreview {
			bodyStr = bodyStr[:maxBodyPreview] + "..."
		}

		errorMsg := fmt.Sprintf("download failed with status code %d", resp.StatusCode)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			errorMsg += ": digest authentication failed (401 Unauthorized); check camera credentials (username/password)"
		case http.StatusForbidden:
			errorMsg += ": access denied (403 Forbidden); user may not have permission to download snapshots"
		case http.StatusNotFound:
			errorMsg += ": snapshot URI not found (404); try getting a fresh snapshot URI"
		}

		if bodyStr != "" {
			errorMsg += fmt.Sprintf("; response: %s", bodyStr)
		}

		return nil, fmt.Errorf("%w: %s", ErrDownloadFailed, errorMsg)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// digestAuthTransport implements digest authentication for HTTP transport.
type digestAuthTransport struct {
	transport *http.Transport
	username  string
	password  string
	nc        int
	ncMu      sync.Mutex // Protects nc field from concurrent access
}

// RoundTrip implements http.RoundTripper with digest auth support.
func (d *digestAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// First request without auth to get the challenge
	resp, err := d.transport.RoundTrip(req)
	if err != nil {
		return resp, fmt.Errorf("transport round trip failed: %w", err)
	}

	// If we get 401, handle digest auth challenge
	if resp.StatusCode == http.StatusUnauthorized {
		// Read the WWW-Authenticate header
		authHeader := resp.Header.Get("WWW-Authenticate")
		if strings.Contains(authHeader, "Digest") {
			// Parse digest challenge and create auth header
			authHeaderValue := d.createDigestAuthHeader(req, authHeader)

			// Create new request with auth header
			newReq := req.Clone(req.Context())
			newReq.Header.Set("Authorization", authHeaderValue)

			// Retry with auth
			resp, err = d.transport.RoundTrip(newReq)
			if err != nil {
				return resp, fmt.Errorf("transport round trip with auth failed: %w", err)
			}

			return resp, nil
		}
	}

	return resp, nil
}

// createDigestAuthHeader creates a digest auth header from the challenge.
func (d *digestAuthTransport) createDigestAuthHeader(req *http.Request, authHeader string) string {
	// Simple digest auth implementation - parse challenge and create response
	// This is a basic implementation that handles most ONVIF cameras

	// Extract digest parameters from WWW-Authenticate header
	realm := extractParam(authHeader, "realm")
	nonce := extractParam(authHeader, "nonce")
	qop := extractParam(authHeader, "qop")
	uri := req.URL.Path
	if req.URL.RawQuery != "" {
		uri += "?" + req.URL.RawQuery
	}

	// Generate response hash
	ha1 := md5Hash(d.username + ":" + realm + ":" + d.password)

	method := req.Method
	ha2 := md5Hash(method + ":" + uri)

	// Increment nonce count atomically to prevent race conditions
	// HTTP transports must be safe for concurrent use
	d.ncMu.Lock()
	d.nc++
	nc := d.nc
	d.ncMu.Unlock()
	ncStr := fmt.Sprintf("%08x", nc)
	cnonce := generateNonce()

	var responseStr string
	if qop == "auth" {
		responseStr = md5Hash(ha1 + ":" + nonce + ":" + ncStr + ":" + cnonce + ":auth:" + ha2)
	} else {
		responseStr = md5Hash(ha1 + ":" + nonce + ":" + ha2)
	}

	// Build Authorization header
	authHeaderValue := fmt.Sprintf(`Digest username=%q, realm=%q, nonce=%q, uri=%q, response=%q`,
		d.username, realm, nonce, uri, responseStr)

	if qop == "auth" {
		authHeaderValue += fmt.Sprintf(`, opaque=%q, qop=%s, nc=%s, cnonce=%q`,
			extractParam(authHeader, "opaque"), qop, ncStr, cnonce)
	}

	return authHeaderValue
}

// Helper functions for digest auth.
func extractParam(authHeader, param string) string {
	prefix := param + `="`
	idx := strings.Index(authHeader, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(authHeader[start:], `"`)
	if end == -1 {
		return ""
	}

	return authHeader[start : start+end]
}

func md5Hash(s string) string {
	h := md5.New() //nolint:gosec // MD5 required for ONVIF digest auth
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// generateNonce generates a cryptographically secure random nonce for digest authentication.
func generateNonce() string {
	bytes := make([]byte, NonceSize)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based nonce if crypto/rand fails (shouldn't happen)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(bytes)
}
