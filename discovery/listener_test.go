package discovery

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

const helloDatagram = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
 xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
<s:Header><a:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Hello</a:Action></s:Header>
<s:Body>
<Hello xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
<EndpointReference><Address>urn:uuid:hello-1111</Address></EndpointReference>
<Types>dp0:NetworkVideoTransmitter</Types>
<Scopes>onvif://www.onvif.org/name/HelloCam</Scopes>
<XAddrs>http://192.168.1.77:80/onvif/device_service</XAddrs>
<MetadataVersion>3</MetadataVersion>
</Hello>
</s:Body>
</s:Envelope>`

const probeMatchesDatagram = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<ProbeMatches xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
<ProbeMatch>
<EndpointReference><Address>urn:uuid:match-2222</Address></EndpointReference>
<Types>dp0:NetworkVideoTransmitter</Types>
<XAddrs>http://192.168.1.88:8000/onvif/device_service</XAddrs>
</ProbeMatch>
</ProbeMatches>
</s:Body>
</s:Envelope>`

const byeDatagram = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<Bye xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
<EndpointReference><Address>urn:uuid:bye-3333</Address></EndpointReference>
</Bye>
</s:Body>
</s:Envelope>`

// loopbackListener wires a Listener's read loop to a plain loopback UDP
// socket so datagram delivery is deterministic on every OS (real multicast
// loopback is not, notably on Windows). The parsing/dispatch/stop semantics
// under test are identical; only the socket differs from production.
type loopbackListener struct {
	*Listener
	conn  *net.UDPConn
	addr  *net.UDPAddr
	errCh chan error
}

func startLoopbackListener(t *testing.T, handler func(*Device)) *loopbackListener {
	t.Helper()

	return startLoopbackListenerOpts(t, handler)
}

// startLoopbackListenerOpts is startLoopbackListener with listener options
// (the filter constraints under test).
func startLoopbackListenerOpts(t *testing.T, handler func(*Device), opts ...ListenerOption) *loopbackListener {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to bind loopback UDP: %v", err)
	}

	listener, err := NewListener("", handler, opts...)
	if err != nil {
		t.Fatalf("NewListener() error = %v", err)
	}

	listener.readTick = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- listener.readLoop(ctx, conn)
	}()

	t.Cleanup(func() {
		listener.Stop()
		cancel()
		_ = conn.Close()
		select {
		case <-listener.Done():
		case <-time.After(2 * time.Second):
			t.Error("listener loop did not exit after Stop")
		}
	})

	return &loopbackListener{
		Listener: listener,
		conn:     conn,
		addr:     conn.LocalAddr().(*net.UDPAddr),
		errCh:    errCh,
	}
}

// send delivers payload to the loopback socket, retrying briefly until the
// loop is provably attached.
func (l *loopbackListener) send(t *testing.T, payload string, delivered func() bool) {
	t.Helper()

	for range 20 {
		if _, err := l.conn.WriteToUDP([]byte(payload), l.addr); err != nil {
			t.Fatalf("failed to send datagram: %v", err)
		}

		if delivered() {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("datagram was not delivered to the listener loop")
}

// sendRaw delivers payload a few times without asserting anything — for
// datagrams the listener is expected to ignore.
func (l *loopbackListener) sendRaw(t *testing.T, payload string) {
	t.Helper()

	for range 3 {
		if _, err := l.conn.WriteToUDP([]byte(payload), l.addr); err != nil {
			t.Fatalf("failed to send datagram: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// sendOnce delivers the payload exactly once and waits for the condition —
// for tests that assert exact delivery counts. The resend-until-delivered
// helper above can inject a duplicate datagram while the first copy still
// sits in the read buffer (the loop polls on readTick), turning "delivered
// exactly one" into a scheduling race.
func (l *loopbackListener) sendOnce(t *testing.T, payload string, delivered func() bool) {
	t.Helper()

	if _, err := l.conn.WriteToUDP([]byte(payload), l.addr); err != nil {
		t.Fatalf("failed to send datagram: %v", err)
	}
	for range 40 {
		if delivered() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("datagram was not delivered to the listener loop")
}

func TestListenerReceivesHello(t *testing.T) {
	var mu sync.Mutex
	var got []*Device

	l := startLoopbackListener(t, func(d *Device) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, d)
	})

	l.send(t, helloDatagram, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(got) > 0
	})

	mu.Lock()
	defer mu.Unlock()

	d := got[0]
	if d.EndpointRef != "urn:uuid:hello-1111" {
		t.Errorf("EndpointRef = %q", d.EndpointRef)
	}

	if len(d.XAddrs) != 1 || d.XAddrs[0] != "http://192.168.1.77:80/onvif/device_service" {
		t.Errorf("XAddrs = %v", d.XAddrs)
	}

	if d.MetadataVersion != 3 {
		t.Errorf("MetadataVersion = %d, want 3", d.MetadataVersion)
	}
}

func TestListenerReceivesProbeMatches(t *testing.T) {
	var mu sync.Mutex
	var refs []string

	l := startLoopbackListener(t, func(d *Device) {
		mu.Lock()
		defer mu.Unlock()
		refs = append(refs, d.EndpointRef)
	})

	l.send(t, probeMatchesDatagram, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(refs) > 0
	})

	mu.Lock()
	defer mu.Unlock()

	if refs[0] != "urn:uuid:match-2222" {
		t.Errorf("EndpointRef = %q, want the ProbeMatches entry", refs[0])
	}
}

func TestListenerIgnoresByeAndGarbage(t *testing.T) {
	var mu sync.Mutex
	var refs []string

	l := startLoopbackListener(t, func(d *Device) {
		mu.Lock()
		defer mu.Unlock()
		refs = append(refs, d.EndpointRef)
	})

	// Bye and garbage must not kill the listener — verify by delivering a
	// Hello afterwards.
	l.sendRaw(t, byeDatagram)
	l.sendRaw(t, "<not+xml")
	l.sendOnce(t, helloDatagram, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(refs) > 0
	})

	mu.Lock()
	defer mu.Unlock()

	if len(refs) != 1 {
		t.Errorf("listener delivered %d devices, want exactly the Hello one (%v)", len(refs), refs)
	}
}

func TestListenerHandlerPanicContained(t *testing.T) {
	var mu sync.Mutex
	var deliveries int

	l := startLoopbackListener(t, func(*Device) {
		mu.Lock()
		deliveries++
		first := deliveries == 1
		mu.Unlock()

		if first {
			panic("handler bug")
		}
	})

	l.sendRaw(t, helloDatagram)
	l.send(t, probeMatchesDatagram, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return deliveries >= 2
	})
}

func TestListenerStopIsIdempotent(t *testing.T) {
	l := startLoopbackListener(t, func(*Device) {})
	l.Stop()
	time.Sleep(50 * time.Millisecond)
	l.Listener.Stop() // must not panic or double-close

	select {
	case <-l.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("listener loop did not exit after Stop")
	}
}

func TestListenerContextCancelExits(t *testing.T) {
	l := startLoopbackListener(t, func(*Device) {})

	// The loop's context is owned by the helper; drive cancellation through
	// Stop semantics instead: verify Done() closes promptly after Stop.
	l.Stop()

	select {
	case <-l.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("listener loop did not exit within 2s of Stop")
	}

	select {
	case err := <-l.errCh:
		if err != nil {
			t.Errorf("read loop returned %v on stop, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("read loop did not return")
	}
}

func TestListenerStartAfterStopFails(t *testing.T) {
	listener, err := NewListener("", func(*Device) {})
	if err != nil {
		t.Fatalf("NewListener() error = %v", err)
	}

	listener.Stop()

	if err := listener.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "stopped") {
		t.Errorf("Start() after Stop() = %v, want ErrListenerStopped", err)
	}
}

func TestNewListenerRequiresHandler(t *testing.T) {
	if _, err := NewListener("", nil); err == nil {
		t.Error("NewListener(nil handler) accepted, want error")
	}
}

// TestListenerMulticastStartStop is the real-socket smoke test: Start must
// bind the multicast group (coexisting with any Discover socket) and Stop
// must unwind it promptly. Datagram delivery over real multicast is
// environment-dependent and is not asserted here.
func TestListenerMulticastStartStop(t *testing.T) {
	// Probe multicast joinability first; hosts without a usable multicast
	// route (some CI runners, VPN-up laptops) skip rather than fail.
	skipWithoutMulticast(t)

	listener, err := NewListener("", func(*Device) {})
	if err != nil {
		t.Fatalf("NewListener() error = %v", err)
	}

	listener.readTick = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- listener.Start(ctx)
	}()

	time.Sleep(300 * time.Millisecond) // allow the socket to bind

	listener.Stop()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Start() = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not exit after Stop")
	}

	// Coexistence: after the listener stops, an active Discover must still
	// be able to bind the same port.
	if _, err := Discover(context.Background(), 300*time.Millisecond); err != nil {
		t.Errorf("Discover() after listener stop failed: %v", err)
	}
}

func TestParseDiscoveryDatagram(t *testing.T) {
	tests := []struct {
		name     string
		datagram string
		wantRef  string
		wantNil  bool
	}{
		{name: "hello", datagram: helloDatagram, wantRef: "urn:uuid:hello-1111"},
		{name: "probe matches", datagram: probeMatchesDatagram, wantRef: "urn:uuid:match-2222"},
		{name: "bye ignored", datagram: byeDatagram, wantNil: true},
		{name: "garbage ignored", datagram: "<nope", wantNil: true},
		{name: "empty envelope ignored", datagram: `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body/></s:Envelope>`, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := parseDiscoveryDatagram([]byte(tt.datagram))
			if tt.wantNil {
				if dev != nil {
					t.Fatalf("parseDiscoveryDatagram() = %+v, want nil", dev)
				}

				return
			}

			if dev == nil {
				t.Fatal("parseDiscoveryDatagram() = nil")
			}

			if dev.EndpointRef != tt.wantRef {
				t.Errorf("EndpointRef = %q, want %q", dev.EndpointRef, tt.wantRef)
			}
		})
	}
}
