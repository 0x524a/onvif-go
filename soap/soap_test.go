package soap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "with credentials",
			username: "admin",
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
			name: "with authentication",
			body: &testRequest{Value: "test"},
			username: "admin",
			password: "password",
			wantErr: false,
		},
		{
			name: "without authentication",
			body: &testRequest{Value: "test"},
			username: "",
			password: "",
			wantErr: false,
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
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				}))
			},
			username: "admin",
			password: "wrong",
			wantErr:  true,
		},
		{
			name: "http error status",
			setupServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("Internal Server Error"))
				}))
			},
			username: "admin",
			password: "password",
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

			req := &testRequest{Value: "test"}
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

	req := &testRequest{Value: "test"}
	var resp interface{}

	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Call(ctx, server.URL, "", req, &resp)
	if err == nil {
		t.Error("Expected timeout error, but got none")
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
	req := &testRequest{Value: "test"}

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
