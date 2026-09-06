package onvif

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test servers shared by the per-service test files. A client method has two
// outcomes worth pinning: it maps a response into the library's own types, or
// it returns an error. newSOAPTestServer drives the first, newSOAPFaultTestServer
// the second.

// newSOAPTestServer returns an httptest.Server that responds to every
// request with a fixed SOAP envelope wrapping body. Tests that only ever
// call one client method against it need no action-based dispatch (unlike
// MockONVIFServer's).
func newSOAPTestServer(t *testing.T, body string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
    <soap:Body>` + body + `</soap:Body>
</soap:Envelope>`))
	}))
	t.Cleanup(server.Close)

	return server
}

// newSOAPFaultTestServer returns an httptest.Server that answers every request
// with a SOAP fault, for exercising the error return of a client method rather
// than its response mapping.
//
// The fault is carried on HTTP 500 deliberately. internal/soap.Client.Call
// decides success from the HTTP status and from whether the body parses into
// the response type - it never consults Body.Fault - so the same fault served
// with a 200 would unmarshal into an all-zero response and be reported to the
// caller as success. 500 is also what SOAP 1.2 prescribes for a fault and what
// real cameras send, so it is both the realistic and the effective choice.
func newSOAPFaultTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
    <soap:Body>
        <soap:Fault>
            <soap:Code><soap:Value>soap:Receiver</soap:Value></soap:Code>
            <soap:Reason><soap:Text>Internal error</soap:Text></soap:Reason>
        </soap:Fault>
    </soap:Body>
</soap:Envelope>`))
	}))
	t.Cleanup(server.Close)

	return server
}
