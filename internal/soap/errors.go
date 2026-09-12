package soap

import "errors"

var (
	// ErrHTTPRequestFailed is returned when an HTTP request fails.
	ErrHTTPRequestFailed = errors.New("HTTP request failed")

	// ErrEmptyResponseBody is returned when a response body is empty.
	ErrEmptyResponseBody = errors.New("received empty response body")

	// ErrSOAPFault is returned when a SOAP response's Body contains a Fault
	// element instead of the expected response content.
	ErrSOAPFault = errors.New("SOAP fault")

	// ErrInvalidResponse is returned when a SOAP response's Body cannot be
	// interpreted as either a Fault or response content.
	ErrInvalidResponse = errors.New("invalid SOAP response")
)
