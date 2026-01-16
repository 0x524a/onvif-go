package soap

import "errors"

var (
	// ErrHTTPRequestFailed is returned when an HTTP request fails.
	ErrHTTPRequestFailed = errors.New("HTTP request failed")

	// ErrEmptyResponseBody is returned when a response body is empty.
	ErrEmptyResponseBody = errors.New("received empty response body")
)
