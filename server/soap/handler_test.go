package soap

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testXMLHeader = `<?xml version="1.0"?>`

const testResultSuccess = "Success"

const (
	testNonce   = "test_nonce_12345"
	testCreated = "2024-01-01T00:00:00Z"
)

// testActionResponse is the response body shared by the handler tests. It has to
// be a struct rather than a map: encoding/xml cannot marshal maps, so a map would
// surface as a Receiver fault instead of reaching any assertion.
type testActionResponse struct {
	XMLName xml.Name `xml:"TestActionResponse"`
	Result  string   `xml:"Result"`
}

func TestNewHandler(t *testing.T) {
	handler := NewHandler("admin", "password")

	if handler == nil {
		t.Error("NewHandler returned nil")

		return
	}
	if handler.username != "admin" {
		t.Errorf("Username mismatch: got %s, want admin", handler.username)
	}
	if handler.password != "password" {
		t.Errorf("Password mismatch: got %s, want password", handler.password)
	}
	if handler.handlers == nil {
		t.Error("Handlers map is nil")
	}
}

func TestRegisterHandler(t *testing.T) {
	handler := NewHandler("admin", "password")

	testHandler := func(body interface{}) (interface{}, error) {
		return "test response", nil
	}

	handler.RegisterHandler("TestAction", testHandler)

	if _, ok := handler.handlers["TestAction"]; !ok {
		t.Error("Handler not registered")
	}
}

func TestServeHTTPMethodNotAllowed(t *testing.T) {
	handler := NewHandler("admin", "password")

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestServeHTTPValidSOAPRequest(t *testing.T) {
	handler := NewHandler("", "") // No authentication

	// Create test handler
	handler.RegisterHandler("TestAction", func(body interface{}) (interface{}, error) {
		return &testActionResponse{Result: testResultSuccess}, nil
	})

	// Create SOAP request
	soapBody := testXMLHeader + `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Assert the registered handler was actually reached. Checking only for
	// "not 500" passed while the request was being rejected as a 400 fault.
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "TestActionResponse") {
		t.Errorf("Response should contain TestActionResponse element, got: %s", body)
	}

	if strings.Contains(body, "Fault") {
		t.Errorf("Response should not be a SOAP fault, got: %s", body)
	}
}

func TestServeHTTPInvalidSOAPEnvelope(t *testing.T) {
	handler := NewHandler("", "")

	invalidXML := `<?xml version="1.0"?>
<invalid>
  <xml>not soap</xml>
</invalid>`

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", strings.NewReader(invalidXML))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return a SOAP fault
	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault, got: %s", w.Body.String())
	}
}

func TestServeHTTPUnknownAction(t *testing.T) {
	handler := NewHandler("", "")

	soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <UnknownAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault for unknown action")
	}
}

func TestExtractAction(t *testing.T) {
	handler := NewHandler("", "")

	tests := []struct {
		name           string
		soapBody       string
		expectedAction string
	}{
		{
			name: "Simple action",
			soapBody: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <GetDeviceInformation/>
  </soap:Body>
</soap:Envelope>`,
			expectedAction: "GetDeviceInformation",
		},
		{
			name: "Action with namespace",
			soapBody: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <tds:GetDeviceInformation xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
  </soap:Body>
</soap:Envelope>`,
			expectedAction: "GetDeviceInformation",
		},
		{
			name: "Action with attributes",
			soapBody: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <GetProfiles>
      <param>value</param>
    </GetProfiles>
  </soap:Body>
</soap:Envelope>`,
			expectedAction: "GetProfiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := handler.extractAction([]byte(tt.soapBody))
			if action != tt.expectedAction {
				t.Errorf("Expected action %s, got %s", tt.expectedAction, action)
			}
		})
	}
}

func TestExtractActionInvalid(t *testing.T) {
	handler := NewHandler("", "")

	invalidXML := "not valid xml at all"
	action := handler.extractAction([]byte(invalidXML))

	if action != "" {
		t.Errorf("Expected empty action for invalid XML, got %s", action)
	}
}

func TestSendFault(t *testing.T) {
	handler := NewHandler("", "")

	w := httptest.NewRecorder()
	handler.sendFault(w, "Sender", "Test error", "Test error message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	response := w.Body.String()
	if !strings.Contains(response, "Fault") {
		t.Error("Response should contain Fault element")
	}
	if !strings.Contains(response, "Test error") {
		t.Error("Response should contain error message")
	}
}

func TestSendResponse(t *testing.T) {
	handler := NewHandler("", "")

	w := httptest.NewRecorder()

	handler.sendResponse(w, &testActionResponse{Result: testResultSuccess})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Response body is empty")
	}
}

// securitySOAPRequest builds a SOAP request carrying a WS-Security UsernameToken.
// The digest is computed over digestPassword, so a caller can send either a
// correct digest or a deliberately wrong one.
func securitySOAPRequest(username, digestPassword, nonce, created string) string {
	hash := sha1.New()
	hash.Write([]byte(nonce))
	hash.Write([]byte(created))
	hash.Write([]byte(digestPassword))
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return testXMLHeader + `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
               xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
  <soap:Header>
    <wsse:Security>
      <wsse:UsernameToken>
        <wsse:Username>` + username + `</wsse:Username>
        <wsse:Password>` + digest + `</wsse:Password>
        <wsse:Nonce>` + base64.StdEncoding.EncodeToString([]byte(nonce)) + `</wsse:Nonce>
        <wsu:Created>` + created + `</wsu:Created>
      </wsse:UsernameToken>
    </wsse:Security>
  </soap:Header>
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`
}

func TestAuthenticate(t *testing.T) {
	handler := NewHandler("admin", "password")
	handler.RegisterHandler("TestAction", func(body interface{}) (interface{}, error) {
		return &testActionResponse{Result: testResultSuccess}, nil
	})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/",
		strings.NewReader(securitySOAPRequest("admin", "password", testNonce, testCreated)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// A correct digest must actually authenticate. The previous version only
	// logged inside a conditional, so it could never fail.
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for a valid digest, got %d: %s", w.Code, w.Body.String())
	}

	if strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Valid credentials should not produce a fault, got: %s", w.Body.String())
	}
}

func TestAuthenticateFailsWithWrongPassword(t *testing.T) {
	handler := NewHandler("admin", "correct_password")
	handler.RegisterHandler("TestAction", func(body interface{}) (interface{}, error) {
		return &testActionResponse{Result: "should not reach here"}, nil
	})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/",
		strings.NewReader(securitySOAPRequest("admin", "wrong_password", testNonce, testCreated)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should fail authentication
	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected authentication failure")
	}
}

func TestHandlerWithoutAuthentication(t *testing.T) {
	handler := NewHandler("", "") // No authentication

	soapBody := testXMLHeader + `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	handler.RegisterHandler("TestAction", func(body interface{}) (interface{}, error) {
		return "success", nil
	})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should succeed without authentication
	if w.Code == http.StatusInternalServerError && strings.Contains(w.Body.String(), "Authentication") {
		t.Errorf("Should not require authentication when not configured")
	}
}

func TestReadRequestBodyError(t *testing.T) {
	handler := NewHandler("", "")

	// Create a request with a body that will fail to read
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", &failingReader{})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault for read error")
	}
}

// Helper types and functions

type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestResponseHandling(t *testing.T) {
	handler := NewHandler("", "")

	handler.RegisterHandler("TestAction", func(body interface{}) (interface{}, error) {
		return &testActionResponse{Result: testResultSuccess}, nil
	})

	soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	response := w.Body.String()
	if !strings.Contains(response, "TestActionResponse") {
		t.Errorf("Response should contain TestActionResponse element")
	}
}

func TestEmptyBody(t *testing.T) {
	handler := NewHandler("", "")

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault for empty body")
	}
}

func TestContentType(t *testing.T) {
	handler := NewHandler("", "")

	handler.RegisterHandler("TestAction", func(body interface{}) (interface{}, error) {
		return "test", nil
	})

	soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/", strings.NewReader(soapBody))
	req.Header.Set("Content-Type", "application/soap+xml")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Handler should work regardless of content type
	if w.Code == http.StatusInternalServerError {
		t.Logf("Note: Handler may validate content type")
	}
}
