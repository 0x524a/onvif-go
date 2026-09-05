package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	onvif "github.com/0x524a/onvif-go"
)

// TestMediaServiceActionNamesMatchRealClient is a regression test for #79:
// the simulator registered GetStreamURI/GetSnapshotURI under the wrong
// action-name casing ("GetStreamURI"/"GetSnapshotURI") while the real
// client - and the ONVIF Media WSDL itself - spell them
// "GetStreamUri"/"GetSnapshotUri". soap.Handler.extractAction dispatches on
// the literal element name case-sensitively, so every call from a real
// onvif.Client to either operation always faulted with "Action not
// supported", regardless of what the client sent.
//
// This drives the real, public onvif.Client (the same one used against
// real cameras) against a real httptest server wrapping the actual
// registerMediaService mux - the only kind of test that would have caught
// this, since every existing direct-handler test in this package calls
// Handle* functions directly and bypasses SOAP action dispatch entirely.
func TestMediaServiceActionNamesMatchRealClient(t *testing.T) {
	config := createTestConfig()
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	mux := http.NewServeMux()
	srv.registerMediaService(mux)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client, err := onvif.NewClient(
		testServer.URL+config.BasePath+"/media_service",
		onvif.WithCredentials(config.Username, config.Password),
		onvif.WithHTTPClient(testServer.Client()),
	)
	if err != nil {
		t.Fatalf("onvif.NewClient() failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	profileToken := config.Profiles[0].Token

	if _, err := client.GetStreamURI(ctx, profileToken); err != nil {
		t.Errorf("GetStreamURI over the real wire path failed: %v", err)
	}
	if _, err := client.GetSnapshotURI(ctx, profileToken); err != nil {
		t.Errorf("GetSnapshotURI over the real wire path failed: %v", err)
	}
}
