package soap

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrHTTPRequestFailed is returned when an HTTP request fails.
	ErrHTTPRequestFailed = errors.New("HTTP request failed")

	// ErrEmptyResponseBody is returned when a response body is empty.
	ErrEmptyResponseBody = errors.New("received empty response body")

	// ErrUnauthorized is a sentinel matching every authentication-shaped
	// failure: HTTP 401/403, a SOAP Fault with a NotAuthorized code, or a
	// 200-status response carrying a NotAuthorized fault. Test with
	// errors.Is(err, ErrUnauthorized); it is re-exported by the root package.
	ErrUnauthorized = errors.New("ONVIF authentication failed")
)

// maxErrorBodyLen caps how much of a raw response body is embedded in
// structured errors; device fault pages can be huge and useless.
const maxErrorBodyLen = 2048

// truncateBody shortens a raw body for inclusion in an error message.
func truncateBody(body string) string {
	if len(body) <= maxErrorBodyLen {
		return body
	}

	return body[:maxErrorBodyLen] + "...(truncated)"
}

// FaultError represents a SOAP Fault received with HTTP 200 (ONVIF devices
// commonly report faults this way) or with a 4xx/5xx status.
type FaultError struct {
	// Code is the SOAP 1.2 fault code value (e.g. "ter:NotAuthorized") or the
	// SOAP 1.1 faultcode.
	Code string
	// Subcode is the SOAP 1.2 fault subcode value, when present.
	Subcode string
	// Reason is the human-readable fault reason (SOAP 1.2 Reason/Text or
	// SOAP 1.1 faultstring).
	Reason string
	// Detail is the raw fault detail element content, when present.
	Detail string
	// HTTPStatus is the HTTP status the fault arrived with (200 is common).
	HTTPStatus int
}

// Error implements the error interface.
func (e *FaultError) Error() string {
	msg := fmt.Sprintf("SOAP fault [%s]: %s", e.Code, e.Reason)
	if e.Subcode != "" {
		msg = fmt.Sprintf("SOAP fault [%s/%s]: %s", e.Code, e.Subcode, e.Reason)
	}

	return msg
}

// IsAuthFailure reports whether the fault represents an authentication or
// authorization rejection (ter:NotAuthorized and common firmware phrasings).
func (e *FaultError) IsAuthFailure() bool {
	return containsFold(e.Code, "notauthorized") ||
		containsFold(e.Subcode, "notauthorized") ||
		containsFold(e.Reason, "not authorized") ||
		containsFold(e.Reason, "unauthorized") ||
		containsFold(e.Reason, "sender not authorized")
}

// HTTPStatusError represents a non-200 HTTP response from a device.
type HTTPStatusError struct {
	// Status is the HTTP status code.
	Status int
	// Body is the (possibly truncated) response body.
	Body string
}

// Error implements the error interface.
func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("%s with status %d: %s",
		ErrHTTPRequestFailed.Error(), e.Status, truncateBody(e.Body))
}

// Unwrap keeps errors.Is(err, ErrHTTPRequestFailed) working for callers that
// matched the legacy string-wrapped error.
func (e *HTTPStatusError) Unwrap() error {
	return ErrHTTPRequestFailed
}

// IsAuthFailure reports whether the status is an authentication rejection.
func (e *HTTPStatusError) IsAuthFailure() bool {
	return e.Status == 401 || e.Status == 403 //nolint:mnd // HTTP status codes
}

// IsAuthFailure reports whether err (or anything in its chain) is an
// authentication-shaped failure: HTTP 401/403, a NotAuthorized SOAP fault, or
// a 200-with-fault of that class.
func IsAuthFailure(err error) bool {
	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) && statusErr.IsAuthFailure() {
		return true
	}

	var faultErr *FaultError
	if errors.As(err, &faultErr) {
		return faultErr.IsAuthFailure()
	}

	return false
}

// IsAuthLadderRetryable reports whether err should advance an auth fallback
// ladder: an auth-class failure, or an empty response body. Old firmwares
// answer a rejected WS-Security digest with HTTP 200 and an empty body — no
// 401, no fault (issue #140) — and the next auth mode often succeeds where
// the digest died. Empty bodies stay non-auth for classification (see
// IsAuthFailure): only the ladder decision treats them as retryable.
func IsAuthLadderRetryable(err error) bool {
	return IsAuthFailure(err) || errors.Is(err, ErrEmptyResponseBody)
}

// containsFold reports whether s contains substr, case-insensitively.
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
