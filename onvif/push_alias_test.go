package onvif_test

// The push/time-sync capability surfaces through the root-package facade
// (the names the README examples use): identity of the sentinels, the
// NotifyHandler/ParseNotifyEvents re-exports, and the facade types.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/device"
	"github.com/mickeyzzc/onvif-go/v2/events"
	onvif "github.com/mickeyzzc/onvif-go/v2/onvif"
)

func TestPushFacadeAliases(t *testing.T) {
	if !errors.Is(events.ErrPushNotSupported, onvif.ErrPushNotSupported) {
		t.Error("onvif.ErrPushNotSupported must match the events sentinel")
	}

	if !errors.Is(device.ErrNoCredentials, onvif.ErrNoCredentials) {
		t.Error("onvif.ErrNoCredentials must match the device sentinel")
	}

	// ParseNotifyEvents through the facade name decodes a real Notify.
	const body = `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>` +
		`<Notify xmlns="http://docs.oasis-open.org/wsn/b-2">` +
		`<NotificationMessage><Topic>tns1:Device/Humidity</Topic>` +
		`<Message UtcTime="2026-10-12T08:00:03Z"><Data><SimpleItem Name="Value" Value="62"/></Data></Message>` +
		`</NotificationMessage></Notify></s:Body></s:Envelope>`

	msgs, err := onvif.ParseNotifyEvents([]byte(body))
	if err != nil || len(msgs) != 1 || msgs[0].Topic != "tns1:Device/Humidity" {
		t.Fatalf("ParseNotifyEvents via facade = (%+v, %v)", msgs, err)
	}

	// NotifyHandler through the facade name serves requests.
	srv := httptest.NewServer(onvif.NotifyHandler(func([]onvif.NotificationMessage) {}))
	defer srv.Close()

	resp, err := http.Post(srv.URL, "application/soap+xml", strings.NewReader("garbage"))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 through the facade handler", resp.StatusCode)
	}

	// The facade types exist and are the service-package types.
	var _ onvif.PushSubscription
	var _ onvif.SubscribePushOptions
	var _ onvif.PushStream
	var _ onvif.MotionAlarm
	var _ onvif.DeviceTime
}
