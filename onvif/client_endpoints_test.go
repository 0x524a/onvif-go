package onvif

import (
	"context"
	"testing"
)

// Endpoint-resolution tests for Initialize's GetServices-first precedence
// (MiBeeNvr#1075): the namespace → XAddr enumeration is authoritative, with
// GetCapabilities XAddrs and the device endpoint as fallbacks.

// capsWithXAddrs builds a GetCapabilities response advertising each service's
// conventional path on the given base URL (scheme + host[:port]).
func capsWithXAddrs(base string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
 <s:Body>
  <tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
   <tds:Capabilities>
    <tt:Media xmlns:tt="http://www.onvif.org/ver10/schema"><tt:XAddr>` + base + `/onvif/media_service</tt:XAddr></tt:Media>
    <tt:PTZ xmlns:tt="http://www.onvif.org/ver10/schema"><tt:XAddr>` + base + `/onvif/ptz_service</tt:XAddr></tt:PTZ>
    <tt:Imaging xmlns:tt="http://www.onvif.org/ver10/schema"><tt:XAddr>` + base + `/onvif/imaging_service</tt:XAddr></tt:Imaging>
    <tt:Events xmlns:tt="http://www.onvif.org/ver10/schema"><tt:XAddr>` + base + `/onvif/events_service</tt:XAddr></tt:Events>
   </tds:Capabilities>
  </tds:GetCapabilitiesResponse>
 </s:Body>
</s:Envelope>`
}

// servicesSingleEntry builds a GetServices response routing every namespace to
// one endpoint — the layout of single-entry firmwares whose HTTP server has
// no /onvif/<x>_service paths at all.
func servicesSingleEntry(endpoint string) string {
	ns := func(namespace string) string {
		return `<tds:Service><tds:Namespace>` + namespace + `</tds:Namespace>` +
			`<tds:XAddr>` + endpoint + `</tds:XAddr></tds:Service>`
	}
	return `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
 <s:Body>
  <tds:GetServicesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">` +
		ns("http://www.onvif.org/ver10/device/wsdl") +
		ns("http://www.onvif.org/ver10/media/wsdl") +
		ns("http://www.onvif.org/ver20/ptz/wsdl") +
		ns("http://www.onvif.org/ver20/imaging/wsdl") +
		ns("http://www.onvif.org/ver10/events/wsdl") +
		`</tds:GetServicesResponse>
 </s:Body>
</s:Envelope>`
}

// servicesWithStaleMediaHost builds a GetServices response whose device entry
// carries base (the address the client uses) while the media XAddr carries a
// stale host with a service-specific port.
func servicesWithStaleMediaHost(base, staleHost string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
 <s:Body>
  <tds:GetServicesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
   <tds:Service><tds:Namespace>http://www.onvif.org/ver10/device/wsdl</tds:Namespace><tds:XAddr>` + base + `/onvif/device_service</tds:XAddr></tds:Service>
   <tds:Service><tds:Namespace>http://www.onvif.org/ver10/media/wsdl</tds:Namespace><tds:XAddr>http://` + staleHost + `:8080/onvif/media_service</tds:XAddr></tds:Service>
  </tds:GetServicesResponse>
 </s:Body>
</s:Envelope>`
}

// initializeWith spins up a mock device on 127.0.0.1 whose advertised hosts
// are built from its real address (via the caps/services builder callbacks),
// runs Initialize, and returns the client plus the mock's base URL.
func initializeWith(t *testing.T, capsFn, servicesFn func(mockBase string) string) (*Client, string) {
	t.Helper()
	mock := NewMockONVIFServer()
	defer mock.Close()

	if capsFn != nil {
		mock.SetResponse("GetCapabilities", capsFn(mock.URL()))
	}
	if servicesFn != nil {
		mock.SetResponse("GetServices", servicesFn(mock.URL()))
	}

	client, err := NewClient(
		mock.URL()+"/onvif/device_service",
		WithCredentials(testUsername, testUsername),
	)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	if err := client.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	return client, mock.URL()
}

// TestInitializePrefersGetServicesXAddr: single-entry firmware advertises
// every service on the device endpoint while GetCapabilities carries the
// conventional paths — GetServices must win, or requests to
// /onvif/<x>_service 404 (MiBeeNvr#1075).
func TestInitializePrefersGetServicesXAddr(t *testing.T) {
	client, base := initializeWith(t,
		capsWithXAddrs, // conventional /onvif/<x>_service paths, same host
		func(mockBase string) string { return servicesSingleEntry(mockBase + "/onvif/device_service") },
	)

	want := base + "/onvif/device_service"
	for name, got := range map[string]string{
		"media":   client.mediaEndpoint,
		"ptz":     client.ptzEndpoint,
		"imaging": client.imagingEndpoint,
		"events":  client.eventEndpoint,
	} {
		if got != want {
			t.Errorf("%s endpoint = %q, want single-entry %q", name, got, want)
		}
	}
}

// TestInitializeRewritesStaleGetServicesHost: an advertised XAddr host that
// disagrees with the endpoint the client actually reached is rewritten to it,
// preserving the service-specific port.
func TestInitializeRewritesStaleGetServicesHost(t *testing.T) {
	client, _ := initializeWith(t,
		nil,
		func(mockBase string) string { return servicesWithStaleMediaHost(mockBase, "192.0.2.200") },
	)

	if want := "http://127.0.0.1:8080/onvif/media_service"; client.mediaEndpoint != want {
		t.Errorf("media endpoint = %q, want stale host rewritten to %q", client.mediaEndpoint, want)
	}
}

// TestInitializeCapsFallbackWithoutGetServices: when the GetServices probe
// fails, endpoints still resolve from GetCapabilities XAddrs.
func TestInitializeCapsFallbackWithoutGetServices(t *testing.T) {
	client, base := initializeWith(t, capsWithXAddrs, nil)

	cases := []struct{ name, got, want string }{
		{"media", client.mediaEndpoint, base + "/onvif/media_service"},
		{"imaging", client.imagingEndpoint, base + "/onvif/imaging_service"},
		{"events", client.eventEndpoint, base + "/onvif/events_service"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s endpoint = %q, want %q", c.name, c.got, c.want)
		}
	}
}
