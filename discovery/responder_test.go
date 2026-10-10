package discovery

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testEndpoint = "urn:uuid:12345678-1234-1234-1234-123456789abc"

func startResponder(t *testing.T, devices ...*Device) *Responder {
	t.Helper()

	ready := make(chan struct{})
	r := NewResponder(&ResponderConfig{ListenAddr: "127.0.0.1:0", Ready: ready})
	for _, d := range devices {
		if err := r.Register(d); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() { errCh <- r.Start(ctx) }()

	t.Cleanup(func() {
		cancel()

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("Start() returned %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Start() did not return after cancel")
		}
	})

	select {
	case <-ready:
	case err := <-errCh:
		t.Fatalf("Start() returned early: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("responder never became ready")
	}

	return r
}

func testDevice(endpoint string, scopes ...string) *Device {
	return &Device{
		EndpointRef: endpoint,
		XAddrs:      []string{"http://127.0.0.1:8080/onvif/device_service"},
		Scopes:      scopes,
	}
}

// TestClientDiscoversResponderOverUnicast is the #117 acceptance test: the
// module's own client finds a responder-backed device with no hardware and no
// multicast.
func TestClientDiscoversResponderOverUnicast(t *testing.T) {
	r := startResponder(t, testDevice(testEndpoint,
		"onvif://www.onvif.org/name/Cam1", "onvif://www.onvif.org/type/video_encoder"))

	devices, err := DiscoverWithOptions(context.Background(), 500*time.Millisecond,
		&DiscoverOptions{ProbeAddress: r.Addr().String()})
	if err != nil {
		t.Fatalf("DiscoverWithOptions() error = %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("found %d devices, want 1", len(devices))
	}

	d := devices[0]
	if d.EndpointRef != testEndpoint {
		t.Errorf("EndpointRef = %q, want %q", d.EndpointRef, testEndpoint)
	}

	if len(d.XAddrs) != 1 || d.XAddrs[0] != "http://127.0.0.1:8080/onvif/device_service" {
		t.Errorf("XAddrs = %v", d.XAddrs)
	}

	if len(d.Types) != 1 || d.Types[0] != "dn:NetworkVideoTransmitter" {
		t.Errorf("Types = %v", d.Types)
	}

	if len(d.Scopes) != 2 {
		t.Errorf("Scopes = %v", d.Scopes)
	}
}

// TestResponderMultiplexesDevices: dozens of cameras behind one socket, each
// found separately.
func TestResponderMultiplexesDevices(t *testing.T) {
	const fleet = 24

	devs := make([]*Device, 0, fleet)
	for i := range fleet {
		devs = append(devs, testDevice(fmt.Sprintf("urn:uuid:00000000-0000-0000-0000-%012d", i)))
	}

	r := startResponder(t, devs...)

	found, err := DiscoverWithOptions(context.Background(), 800*time.Millisecond,
		&DiscoverOptions{ProbeAddress: r.Addr().String()})
	if err != nil {
		t.Fatalf("DiscoverWithOptions() error = %v", err)
	}

	if len(found) != fleet {
		t.Fatalf("found %d devices, want %d", len(found), fleet)
	}
}

func probeXML(types, scopes, matchBy string) string {
	scopeEl := ""
	if scopes != "" {
		attr := ""
		if matchBy != "" {
			attr = ` MatchBy="` + matchBy + `"`
		}

		scopeEl = `<d:Scopes` + attr + `>` + scopes + `</d:Scopes>`
	}

	typeEl := ""
	if types != "" {
		typeEl = `<d:Types xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">` + types + `</d:Types>`
	}

	return `<?xml version="1.0"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" ` +
		`xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" ` +
		`xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"><s:Header>` +
		`<a:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>` +
		`<a:MessageID>uuid:probe-1</a:MessageID></s:Header><s:Body><d:Probe>` +
		typeEl + scopeEl + `</d:Probe></s:Body></s:Envelope>`
}

// probeCount sends raw and counts the replies that arrive within a short wait.
func probeCount(t *testing.T, r *Responder, raw string) []string {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.WriteToUDP([]byte(raw), r.Addr().(*net.UDPAddr)); err != nil {
		t.Fatalf("write: %v", err)
	}

	var replies []string

	buf := make([]byte, maxDatagram)

	for {
		_ = conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))

		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return replies
		}

		replies = append(replies, string(buf[:n]))
	}
}

func TestProbeFiltering(t *testing.T) {
	r := startResponder(t,
		testDevice("urn:uuid:a", "onvif://www.onvif.org/name/Cam1", "onvif://www.onvif.org/location/lab/rack1"),
		&Device{EndpointRef: "urn:uuid:b", Types: []string{"{urn:other}Thing"},
			Scopes: []string{"onvif://www.onvif.org/name/Cam2"}},
	)

	tests := []struct {
		name string
		raw  string
		want int
	}{
		{"no filter matches all", probeXML("", "", ""), 2},
		{"onvif type", probeXML("dp0:NetworkVideoTransmitter", "", ""), 1},
		{"unknown type", probeXML("dp0:Nope", "", ""), 0},
		{"scope prefix", probeXML("", "onvif://www.onvif.org/location", ""), 1},
		{"scope path prefix by segment", probeXML("", "onvif://www.onvif.org/location/lab", ""), 1},
		{"scope partial segment no match", probeXML("", "onvif://www.onvif.org/location/la", ""), 0},
		{"scope authority case-insensitive", probeXML("", "onvif://WWW.ONVIF.ORG/name", ""), 2},
		{"scope exact via strcmp0", probeXML("", "onvif://www.onvif.org/name/Cam1", matchByStrcmp0), 1},
		{"strcmp0 rejects prefix", probeXML("", "onvif://www.onvif.org/name", matchByStrcmp0), 0},
		{"ldap unsupported matches nothing", probeXML("", "onvif://www.onvif.org/name", nsDiscovery+"/ldap"), 0},
		{"type and scope combined", probeXML("dp0:NetworkVideoTransmitter", "onvif://www.onvif.org/name/Cam2", ""), 0},
		{"not a probe", `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body/></s:Envelope>`, 0},
		{"garbage", "not xml", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(probeCount(t, r, tt.raw)); got != tt.want {
				t.Errorf("got %d replies, want %d", got, tt.want)
			}
		})
	}
}

func TestProbeMatchesEnvelope(t *testing.T) {
	r := startResponder(t, testDevice(testEndpoint, "onvif://www.onvif.org/name/Cam1"))

	replies := probeCount(t, r, probeXML("", "", ""))
	if len(replies) != 1 {
		t.Fatalf("got %d replies", len(replies))
	}

	for _, want := range []string{
		"<a:RelatesTo>uuid:probe-1</a:RelatesTo>",
		nsDiscovery + "/ProbeMatches",
		"<d:AppSequence ",
		"<d:MetadataVersion>0</d:MetadataVersion>",
	} {
		if !strings.Contains(replies[0], want) {
			t.Errorf("reply missing %q:\n%s", want, replies[0])
		}
	}
}

func TestHelloAndByeAnnounced(t *testing.T) {
	sink, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer func() { _ = sink.Close() }()

	ready := make(chan struct{})
	r := NewResponder(&ResponderConfig{
		ListenAddr:   "127.0.0.1:0",
		AnnounceAddr: sink.LocalAddr().String(),
		Ready:        ready,
	})

	if err := r.Register(testDevice(testEndpoint)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() { done <- r.Start(ctx) }()

	<-ready

	read := func() string {
		t.Helper()

		_ = sink.SetReadDeadline(time.Now().Add(3 * time.Second))
		buf := make([]byte, maxDatagram)

		n, _, err := sink.ReadFromUDP(buf)
		if err != nil {
			t.Fatalf("no announcement: %v", err)
		}

		return string(buf[:n])
	}

	if hello := read(); !strings.Contains(hello, "<d:Hello>") || !strings.Contains(hello, testEndpoint) {
		t.Errorf("expected Hello for %s, got:\n%s", testEndpoint, hello)
	}

	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Start() returned %v", err)
	}

	if bye := read(); !strings.Contains(bye, "<d:Bye>") || !strings.Contains(bye, testEndpoint) {
		t.Errorf("expected Bye for %s, got:\n%s", testEndpoint, bye)
	}
}

func TestRegisterWhileRunningAndUnregister(t *testing.T) {
	r := startResponder(t)

	if got := len(probeCount(t, r, probeXML("", "", ""))); got != 0 {
		t.Fatalf("empty responder replied %d times", got)
	}

	if err := r.Register(testDevice(testEndpoint)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if got := len(probeCount(t, r, probeXML("", "", ""))); got != 1 {
		t.Errorf("after Register got %d replies, want 1", got)
	}

	r.Unregister(testEndpoint)

	if got := len(probeCount(t, r, probeXML("", "", ""))); got != 0 {
		t.Errorf("after Unregister got %d replies, want 0", got)
	}
}

func TestRegisterValidation(t *testing.T) {
	r := NewResponder(nil)

	if err := r.Register(nil); err == nil {
		t.Error("Register(nil) succeeded")
	}

	if err := r.Register(&Device{}); err == nil {
		t.Error("Register without EndpointRef succeeded")
	}

	if err := r.Register(&Device{EndpointRef: "x", Types: []string{"bogus"}}); err == nil {
		t.Error("Register with unparseable type succeeded")
	}
}

func TestStartReportsBindError(t *testing.T) {
	r := NewResponder(&ResponderConfig{ListenAddr: "256.0.0.1:0"})
	if err := r.Start(context.Background()); err == nil {
		t.Fatal("Start() succeeded with an unusable address")
	}
}

func TestDiscoverOptionsBadProbeAddress(t *testing.T) {
	_, err := DiscoverWithOptions(context.Background(), 50*time.Millisecond,
		&DiscoverOptions{ProbeAddress: "not an address"})
	if err == nil {
		t.Fatal("expected error for bad ProbeAddress")
	}
}

func TestStableEndpointRef(t *testing.T) {
	a, b := StableEndpointRef("farm1/cam-07"), StableEndpointRef("farm1/cam-07")
	if a != b {
		t.Errorf("same seed gave %q and %q", a, b)
	}

	if a == StableEndpointRef("farm1/cam-08") {
		t.Error("different seeds collided")
	}

	// Pinned: changing seedNamespace would renumber every seeded device.
	if want := "urn:uuid:" + uuid.NewSHA1(seedNamespace, []byte("farm1/cam-07")).String(); a != want || !strings.HasPrefix(a, "urn:uuid:") {
		t.Errorf("unexpected form %q", a)
	}
}
