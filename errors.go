package onvif

import (
	"errors"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

var (
	// ErrInvalidEndpoint is returned when the endpoint is invalid.
	ErrInvalidEndpoint = errors.New("invalid endpoint")

	// ErrAuthenticationRequired is returned when authentication is required but not provided.
	ErrAuthenticationRequired = errors.New("authentication required")

	// ErrAuthenticationFailed is returned when authentication fails.
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrServiceNotSupported is returned when a service is not supported by the device.
	ErrServiceNotSupported = errors.New("service not supported")

	// ErrInvalidResponse is returned when the response is invalid. Aliased to
	// internal/soap's sentinel of the same name for the same reason as
	// ErrHTTPRequestFailed above.
	ErrInvalidResponse = soap.ErrInvalidResponse

	// ErrTimeout is returned when a request times out.
	ErrTimeout = errors.New("request timeout")

	// ErrConnectionFailed is returned when connection to the device fails.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrInvalidParameter is returned when a parameter is invalid.
	ErrInvalidParameter = errors.New("invalid parameter")

	// ErrNotInitialized is returned by PTZ and Imaging operations when
	// Initialize has not successfully run, so the client does not yet know
	// whether the device offers the service. It is a recoverable client-state
	// problem - call Initialize and retry - and is deliberately distinct from
	// ErrServiceNotSupported, which reports a permanent device limitation.
	ErrNotInitialized = errors.New("client not initialized")

	// ErrNoProbeMatches is returned when no probe matches are found during discovery.
	ErrNoProbeMatches = errors.New("no probe matches found")

	// ErrNetworkInterfaceNotFound is returned when a network interface is not found.
	ErrNetworkInterfaceNotFound = errors.New("network interface not found")

	// ErrHTTPRequestFailed is returned when an HTTP request fails. It is the
	// same sentinel internal/soap.Call returns, so errors.Is matches across
	// the package boundary instead of comparing two distinct errors.New values.
	ErrHTTPRequestFailed = soap.ErrHTTPRequestFailed

	// ErrEmptyResponseBody is returned when a response body is empty. See
	// ErrHTTPRequestFailed above for why this aliases internal/soap's sentinel.
	ErrEmptyResponseBody = soap.ErrEmptyResponseBody

	// ErrVideoSourceNotFound is returned when a video source is not found.
	ErrVideoSourceNotFound = errors.New("video source not found")

	// ErrProfileNotFound is returned when a profile is not found.
	ErrProfileNotFound = errors.New("profile not found")

	// ErrSnapshotNotSupported is returned when snapshot is not supported for a profile.
	ErrSnapshotNotSupported = errors.New("snapshot not supported for profile")

	// ErrPTZNotSupported is returned when PTZ is not supported for a profile.
	ErrPTZNotSupported = errors.New("PTZ not supported for profile")

	// ErrPresetNotFound is returned when a preset is not found.
	ErrPresetNotFound = errors.New("preset not found")

	// ErrTestRequestFailed is returned when a test request fails.
	ErrTestRequestFailed = errors.New("test request failed")

	// ErrTestRequestNewFailed is returned when creating a test request fails.
	ErrTestRequestNewFailed = errors.New("test request creation failed")

	// ErrTestRequestDoFailed is returned when executing a test request fails.
	ErrTestRequestDoFailed = errors.New("test request execution failed")

	// ErrTestRequestUnexpectedStatus is returned when a test request has unexpected status.
	ErrTestRequestUnexpectedStatus = errors.New("test request unexpected status")

	// ErrURLMissingHost is returned when a URL is missing a host.
	ErrURLMissingHost = errors.New("URL missing host")

	// ErrInvalidEndpointFormat is returned when an endpoint format is invalid.
	ErrInvalidEndpointFormat = errors.New("invalid endpoint format")

	// ErrDigestAuthRequiresCredentials is returned when digest auth is attempted without credentials.
	ErrDigestAuthRequiresCredentials = errors.New("digest auth requires credentials")

	// ErrDownloadFailed is returned when a download fails.
	ErrDownloadFailed = errors.New("download failed")
)

// ONVIFError represents an ONVIF-specific error.
type ONVIFError struct {
	Code    string
	Reason  string
	Message string
}

// Error implements the error interface.
func (e *ONVIFError) Error() string {
	return fmt.Sprintf("ONVIF error [%s]: %s - %s", e.Code, e.Reason, e.Message)
}

// NewONVIFError creates a new ONVIF error.
func NewONVIFError(code, reason, message string) *ONVIFError {
	return &ONVIFError{
		Code:    code,
		Reason:  reason,
		Message: message,
	}
}

// IsONVIFError checks if an error is an ONVIF error.
func IsONVIFError(err error) bool {
	var onvifErr *ONVIFError

	return errors.As(err, &onvifErr)
}
