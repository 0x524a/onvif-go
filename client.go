package onvif

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client represents an ONVIF client for communicating with IP cameras
type Client struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
	mu         sync.RWMutex
	
	// Service endpoints
	deviceEndpoint  string
	mediaEndpoint   string
	ptzEndpoint     string
	imagingEndpoint string
	eventEndpoint   string
}

// ClientOption is a functional option for configuring the Client
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithCredentials sets the authentication credentials
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// NewClient creates a new ONVIF client
func NewClient(endpoint string, opts ...ClientOption) (*Client, error) {
	// Validate endpoint
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid endpoint: must include scheme and host")
	}

	client := &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Initialize discovers and initializes service endpoints
func (c *Client) Initialize(ctx context.Context) error {
	// Get device information and capabilities
	capabilities, err := c.GetCapabilities(ctx)
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %w", err)
	}

	// Extract service endpoints
	if capabilities.Media != nil && capabilities.Media.XAddr != "" {
		c.mediaEndpoint = capabilities.Media.XAddr
	}
	if capabilities.PTZ != nil && capabilities.PTZ.XAddr != "" {
		c.ptzEndpoint = capabilities.PTZ.XAddr
	}
	if capabilities.Imaging != nil && capabilities.Imaging.XAddr != "" {
		c.imagingEndpoint = capabilities.Imaging.XAddr
	}
	if capabilities.Events != nil && capabilities.Events.XAddr != "" {
		c.eventEndpoint = capabilities.Events.XAddr
	}

	return nil
}

// Endpoint returns the device endpoint
func (c *Client) Endpoint() string {
	return c.endpoint
}

// SetCredentials updates the authentication credentials
func (c *Client) SetCredentials(username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.username = username
	c.password = password
}

// GetCredentials returns the current credentials
func (c *Client) GetCredentials() (string, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.username, c.password
}
