package onvif

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidEndpoint is returned when the endpoint is invalid
	ErrInvalidEndpoint = errors.New("invalid endpoint")

	// ErrAuthenticationRequired is returned when authentication is required but not provided
	ErrAuthenticationRequired = errors.New("authentication required")

	// ErrAuthenticationFailed is returned when authentication fails
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrServiceNotSupported is returned when a service is not supported by the device
	ErrServiceNotSupported = errors.New("service not supported")

	// ErrInvalidResponse is returned when the response is invalid
	ErrInvalidResponse = errors.New("invalid response")

	// ErrTimeout is returned when a request times out
	ErrTimeout = errors.New("request timeout")

	// ErrConnectionFailed is returned when connection to the device fails
	ErrConnectionFailed = errors.New("connection failed")

	// ErrInvalidParameter is returned when a parameter is invalid
	ErrInvalidParameter = errors.New("invalid parameter")

	// ErrNotInitialized is returned when the client is not initialized
	ErrNotInitialized = errors.New("client not initialized")
)

// ONVIFError represents an ONVIF-specific error
type ONVIFError struct {
	Code    string
	Reason  string
	Message string
}

// Error implements the error interface
func (e *ONVIFError) Error() string {
	return fmt.Sprintf("ONVIF error [%s]: %s - %s", e.Code, e.Reason, e.Message)
}

// NewONVIFError creates a new ONVIF error
func NewONVIFError(code, reason, message string) *ONVIFError {
	return &ONVIFError{
		Code:    code,
		Reason:  reason,
		Message: message,
	}
}

// IsONVIFError checks if an error is an ONVIF error
func IsONVIFError(err error) bool {
	var onvifErr *ONVIFError
	return errors.As(err, &onvifErr)
}
