package soap

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testSoapUsername = "admin"
	testSoapPassword = "password"
	testSoapValue    = "test"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "with credentials",
			username: testSoapUsername,
			password: "password123",
		},
		{
			name:     "without credentials",
			username: "",
			password: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := &http.Client{Timeout: 10 * time.Second}
			client := NewClient(httpClient, tt.username, tt.password)

			if client == nil {
				t.Fatal("NewClient() returned nil")
			}

			if client.username != tt.username {
				t.Errorf("username = %v, want %v", client.username, tt.username)
			}

			if client.password != tt.password {
				t.Errorf("password = %v, want %v", client.password, tt.password)
			}

			if client.httpClient != httpClient {
				t.Error("httpClient not set correctly")
			}
		})
	}
}

func TestBuildEnvelope(t *testing.T) {
	type testRequest struct {
		Value string `xml:"Value"`
	}

	tests := []struct {
		name     string
		body     interface{}
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "with authentication",
			body:     &testRequest{Value: testSoapValue},
			username: testSoapUsername,
			password: testSoapPassword,
			wantErr:  false,
		},
		{
			name:     "without authentication",
			body:     &testRequest{Value: testSoapValue},
			username: "",
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope, err := BuildEnvelope(tt.body, tt.username, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildEnvelope() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if envelope == nil {
				t.Fatal("BuildEnvelope() returned nil envelope")
			}

			if tt.username != "" && envelope.Header == nil {
				t.Error("Expected Header to be set with credentials")
			}

			if tt.username == "" && envelope.Header != nil {
				t.Error("Expected Header to be nil without credentials")
			}
		})
	}
}

func TestClientCall(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func(*testing.T) *httptest.Server
		username       string
		password       string
		wantErr        bool
		wantStatusCode int
	}{
		{
			name: "successful request",
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/soap+xml")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`<?xml version="1.0"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
	<Body>
		<TestResponse>
			<Value>success</Value>
		</TestResponse>
	</Body>
</Envelope>`))
				}))
			},
			username:       "admin",
			password:       "password",
			wantErr:        false,
			wantStatusCode: http.StatusOK,
		},
		{
			name: "unauthorized request",
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				}))
			},
			username: testSoapUsername,
			password: "wrong",
			wantErr:  true,
		},
		{
			name: "http error status",
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("Internal Server Error"))
				}))
			},
			username: testSoapUsername,
			password: testSoapPassword,
			wantErr:  true,
		},
		{
			name: "malformed envelope",
			setupServer: func(t *testing.T) *httptest.Server {
				t.Helper()

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("not xml at all"))
				}))
			},
			username: testSoapUsername,
			password: testSoapPassword,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer(t)
			defer server.Close()

			httpClient := &http.Client{Timeout: 5 * time.Second}
			client := NewClient(httpClient, tt.username, tt.password)

			type testRequest struct {
				Value string `xml:"Value"`
			}

			type testResponse struct {
				Value string `xml:"Value"`
			}

			req := &testRequest{Value: testSoapValue}
			var resp testResponse

			ctx := context.Background()
			err := client.Call(ctx, server.URL, "", req, &resp)

			if (err != nil) != tt.wantErr {
				t.Errorf("Call() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientCallWithTimeout(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := NewClient(httpClient, "admin", "password")

	type testRequest struct {
		Value string `xml:"Value"`
	}

	req := &testRequest{Value: testSoapValue}
	var resp interface{}

	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Call(ctx, server.URL, "", req, &resp)
	if err == nil {
		t.Error("Expected timeout error, but got none")
	}
}

func TestClientCallResponseUnmarshalFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
	<Body>
		<TestResponse>
			<Value>not-a-number</Value>
		</TestResponse>
	</Body>
</Envelope>`))
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := NewClient(httpClient, "", "")

	type testResponse struct {
		Value int `xml:"Value"`
	}
	var resp testResponse

	err := client.Call(context.Background(), server.URL, "", struct{}{}, &resp)
	if err == nil {
		t.Fatal("Call() error = nil, want an unmarshal error for a non-numeric Value")
	}
}

func TestClientCallSOAPFault(t *testing.T) {
	tests := []struct {
		name      string
		faultBody string
		wantInErr []string
	}{
		{
			name: "fault with detail",
			faultBody: `<?xml version="1.0"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
	<Body>
		<Fault>
			<Code><Value>Sender</Value></Code>
			<Reason><Text>Invalid argument</Text></Reason>
			<Detail>ProfileToken not found</Detail>
		</Fault>
	</Body>
</Envelope>`,
			wantInErr: []string{"Sender", "Invalid argument", "ProfileToken not found"},
		},
		{
			name: "fault without detail",
			faultBody: `<?xml version="1.0"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
	<Body>
		<Fault>
			<Code><Value>Receiver</Value></Code>
			<Reason><Text>Internal error</Text></Reason>
		</Fault>
	</Body>
</Envelope>`,
			wantInErr: []string{"Receiver", "Internal error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/soap+xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.faultBody))
			}))
			defer server.Close()

			httpClient := &http.Client{Timeout: 5 * time.Second}
			client := NewClient(httpClient, "", "")

			type testResponse struct {
				Value string `xml:"Value"`
			}
			var resp testResponse

			err := client.Call(context.Background(), server.URL, "", struct{}{}, &resp)
			if err == nil {
				t.Fatal("Call() error = nil, want a SOAP fault error")
			}

			if !errors.Is(err, ErrSOAPFault) {
				t.Errorf("Call() error = %v, want errors.Is(err, ErrSOAPFault)", err)
			}

			for _, want := range tt.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Call() error = %q, want it to contain %q", err.Error(), want)
				}
			}
		})
	}
}

func TestBodyUnmarshalXMLFault(t *testing.T) {
	data := []byte(`<Body xmlns="http://www.w3.org/2003/05/soap-envelope">
	<Fault>
		<Code><Value>Sender</Value></Code>
		<Reason><Text>bad request</Text></Reason>
	</Fault>
</Body>`)

	var body Body
	if err := xml.Unmarshal(data, &body); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if body.Fault == nil {
		t.Fatal("body.Fault = nil, want non-nil Fault")
	}

	if body.Fault.Code != "Sender" || body.Fault.Reason != "bad request" {
		t.Errorf("body.Fault = %+v, want Code=Sender Reason=%q", body.Fault, "bad request")
	}

	if body.Content != nil {
		t.Errorf("body.Content = %v, want nil when Fault is present", body.Content)
	}
}

func TestSecurityHeaderCreation(t *testing.T) {
	httpClient := &http.Client{}
	client := NewClient(httpClient, "testuser", "testpass")

	security := client.createSecurityHeader()

	if security == nil {
		t.Fatal("createSecurityHeader() returned nil")
	}

	if security.UsernameToken == nil {
		t.Fatal("UsernameToken is nil")
	}

	if security.UsernameToken.Username != "testuser" {
		t.Errorf("Username = %v, want %v", security.UsernameToken.Username, "testuser")
	}

	if security.UsernameToken.Password.Type != "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest" {
		t.Error("Password type not set correctly")
	}

	if security.UsernameToken.Nonce.Type != "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary" {
		t.Error("Nonce type not set correctly")
	}

	if security.UsernameToken.Created == "" {
		t.Error("Created timestamp is empty")
	}

	if security.UsernameToken.Password.Password == "" {
		t.Error("Password digest is empty")
	}

	if security.UsernameToken.Nonce.Nonce == "" {
		t.Error("Nonce is empty")
	}
}

func BenchmarkNewClient(b *testing.B) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewClient(httpClient, "admin", "password")
	}
}

func BenchmarkBuildEnvelope(b *testing.B) {
	type testRequest struct {
		Value string `xml:"Value"`
	}
	req := &testRequest{Value: testSoapValue}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildEnvelope(req, "admin", "password")
	}
}

func BenchmarkCreateSecurityHeader(b *testing.B) {
	httpClient := &http.Client{}
	client := NewClient(httpClient, "admin", "password")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.createSecurityHeader()
	}
}
