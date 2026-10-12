package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

// ErrListenerStopped is returned by Start after the listener was stopped —
// a stopped listener cannot be restarted; build a new one.
var ErrListenerStopped = errors.New("discovery listener already stopped")

// listenerReadTick is the read-deadline granularity: the loop re-arms its
// deadline every tick so cancellation (Stop / ctx) is noticed promptly even
// while blocked in ReadFromUDP. A blocked read would otherwise hold the
// close sequence hostage.
const listenerReadTick = 5 * time.Second

// listenerBufferSize absorbs power-on bursts when a camera rack boots at
// once; Hello messages are small but arrive in volleys.
const listenerBufferSize = 256 * 1024

// Listener is a passive WS-Discovery listener: a resident UDP :3702 socket
// that receives the Hello messages cameras broadcast the moment they power
// on — the "plug in and it appears instantly" experience — plus the
// ProbeMatches other devices send in answer to someone else's probe (a free
// second discovery source). Bye messages and anything unparseable are
// ignored.
//
// Coexistence: the listener binds the multicast group exactly like
// DiscoverWithOptions does; on Linux and macOS the kernel delivers a copy of
// each multicast datagram to every bound socket, so a passive listener and
// on-demand Discover() calls coexist in one process without stealing each
// other's traffic.
type Listener struct {
	ifaceName string
	handler   func(*Device)
	cfg       listenerConfig

	readTick time.Duration

	// conn is the live multicast socket (nil before Start binds it);
	// Stop closes it so a blocked ReadFromUDP wakes immediately
	// instead of waiting out the read-deadline tick.
	conn     *net.UDPConn
	connMu   sync.Mutex
	stopOnce sync.Once
	stopped  chan struct{}
	done     chan struct{}
}

// listenerConfig is the compiled set of listener constraints.
type listenerConfig struct {
	// onvifFilter drops announcements carrying no ONVIF signal (see
	// IsONVIFResponder) — generic WS-Discovery responders (Windows hosts,
	// NAS boxes, printers) otherwise surface as devices that can never
	// connect.
	onvifFilter  bool
	requireTypes []string
	scopePrefix  string
}

// ListenerOption configures NewListener.
type ListenerOption func(*listenerConfig)

// newListenerConfig is the default configuration: the ONVIF signal gate on.
func newListenerConfig() listenerConfig {
	return listenerConfig{onvifFilter: true}
}

// WithoutONVIFFilter disables the built-in ONVIF signal gate. By default the
// listener drops Hello/ProbeMatch announcements that carry no ONVIF signal
// (no NetworkVideoTransmitter type and no onvif://www.onvif.org/ scope) —
// without that gate, the Windows/WSD hosts on a LAN announce generic Hellos
// that surface as forever-pending ghost devices in inventories. Turn it off
// only when the handler does its own filtering.
func WithoutONVIFFilter() ListenerOption {
	return func(c *listenerConfig) { c.onvifFilter = false }
}

// WithListenerTypes adds an exact-match Types constraint: an announcement
// must advertise at least one of the given types to reach the handler.
// Comparison is by local name, so namespace prefix dialects (dp0:, tdn:,
// tns:) do not matter.
func WithListenerTypes(types ...string) ListenerOption {
	return func(c *listenerConfig) {
		c.requireTypes = append(c.requireTypes, types...)
	}
}

// WithListenerScopePrefix adds a Scopes constraint: at least one advertised
// scope must carry the given prefix (e.g. "onvif://www.onvif.org/location/Building1"
// to hear only one site's cameras).
func WithListenerScopePrefix(prefix string) ListenerOption {
	return func(c *listenerConfig) { c.scopePrefix = prefix }
}

// accepts applies the compiled constraints to a parsed announcement.
func (c listenerConfig) accepts(d *Device) bool {
	if c.onvifFilter && !IsONVIFResponder(d) {
		return false
	}

	if len(c.requireTypes) > 0 && !deviceHasAnyType(d.Types, c.requireTypes) {
		return false
	}

	if c.scopePrefix != "" && !deviceHasScopePrefix(d.Scopes, c.scopePrefix) {
		return false
	}

	return true
}

// deviceHasAnyType reports whether advertised and required type lists share
// a local name.
func deviceHasAnyType(advertised, required []string) bool {
	for _, have := range advertised {
		for _, want := range required {
			if localPart(have) == localPart(want) {
				return true
			}
		}
	}

	return false
}

// deviceHasScopePrefix reports whether any advertised scope carries prefix.
func deviceHasScopePrefix(scopes []string, prefix string) bool {
	for _, scope := range scopes {
		if strings.HasPrefix(scope, prefix) {
			return true
		}
	}

	return false
}

// localPart strips a qname's namespace prefix ("dp0:NetworkVideoTransmitter"
// → "NetworkVideoTransmitter"; bare names pass through).
func localPart(qname string) string {
	if idx := strings.LastIndex(qname, ":"); idx >= 0 {
		return qname[idx+1:]
	}

	return qname
}

// NewListener creates a passive discovery listener. ifaceName selects the
// multicast interface ("" = kernel default, recommended on single-NIC hosts;
// see DiscoverOptions.NetworkInterface for accepted forms). The handler runs
// on the listener goroutine: return quickly and run slow work on your own
// goroutine — a panicking handler is contained and does not stop the
// listener.
//
// By default announcements that carry no ONVIF signal are dropped before
// the handler sees them (see WithoutONVIFFilter); WithListenerTypes and
// WithListenerScopePrefix narrow further.
func NewListener(ifaceName string, handler func(*Device), opts ...ListenerOption) (*Listener, error) {
	if handler == nil {
		return nil, fmt.Errorf("discovery listener: handler is nil")
	}

	cfg := newListenerConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Listener{
		ifaceName: ifaceName,
		handler:   handler,
		cfg:       cfg,
		readTick:  listenerReadTick,
		stopped:   make(chan struct{}),
		done:      make(chan struct{}),
	}, nil
}

// Start blocks listening for Hello/ProbeMatches until Stop is called or ctx
// is cancelled. Call it on its own goroutine. Returns ErrListenerStopped when
// called after Stop.
func (l *Listener) Start(ctx context.Context) error {
	select {
	case <-l.stopped:
		return ErrListenerStopped
	default:
	}

	var iface *net.Interface
	if l.ifaceName != "" {
		var err error
		iface, err = resolveNetworkInterface(l.ifaceName)
		if err != nil {
			return fmt.Errorf("discovery listener: %w", err)
		}
	}

	group, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return fmt.Errorf("discovery listener: failed to resolve multicast address: %w", err)
	}

	conn, err := net.ListenMulticastUDP("udp", iface, group)
	if err != nil {
		return fmt.Errorf("discovery listener: failed to listen on multicast address: %w", err)
	}

	l.connMu.Lock()
	l.conn = conn
	l.connMu.Unlock()

	// Cover the Stop-before-bind race (Stop found no socket to close).
	select {
	case <-l.stopped:
		_ = conn.Close()

		return nil
	default:
	}

	defer func() {
		_ = conn.Close()
	}()

	return l.readLoop(ctx, conn)
}

// readLoop listens on conn until stopped/cancelled; it owns the Done signal
// so both Start and direct-loop consumers (tests) observe loop exit.
func (l *Listener) readLoop(ctx context.Context, conn *net.UDPConn) error {
	defer close(l.done)

	buffer := make([]byte, listenerBufferSize)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(l.readTick)); err != nil {
			return fmt.Errorf("discovery listener: failed to set read deadline: %w", err)
		}

		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				// Deadline tick: check for cancellation, keep listening.
				select {
				case <-l.stopped:
					return nil
				case <-ctx.Done():
					return nil
				default:
					continue
				}
			}

			// A closed connection (Stop on another path) surfaces as a
			// non-timeout error; treat every error as terminal but check
			// whether it was an intentional stop for the return value.
			select {
			case <-l.stopped:
				return nil
			default:
				return fmt.Errorf("discovery listener: read failed: %w", err)
			}
		}

		if device := parseDiscoveryDatagram(buffer[:n]); device != nil && l.cfg.accepts(device) {
			l.dispatch(device)
		}

		select {
		case <-l.stopped:
			return nil
		case <-ctx.Done():
			return nil
		default:
		}
	}
}

// dispatch hands a device to the handler with panic isolation.
func (l *Listener) dispatch(device *Device) {
	defer func() {
		_ = recover() // a panicking handler must not stop the listener
	}()

	l.handler(device)
}

// Stop stops the listener and unblocks Start immediately (the multicast
// socket is closed, interrupting any blocked read). Idempotent; a stopped
// listener cannot be restarted (create a new one).
func (l *Listener) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopped)

		l.connMu.Lock()
		conn := l.conn
		l.connMu.Unlock()

		if conn != nil {
			_ = conn.Close()
		}
	})
}

// Done returns a channel closed when the read loop has fully exited.
func (l *Listener) Done() <-chan struct{} {
	return l.done
}

// parseDiscoveryDatagram extracts a device from a Hello or ProbeMatches
// datagram through the shared wsdiscovery codec. Returns nil for Bye
// messages, responses with no match, and anything unparseable — a
// listener must survive garbage on the wire.
func parseDiscoveryDatagram(data []byte) *Device {
	match := wsdiscovery.ParseAnnouncement(data)
	if match == nil {
		return nil
	}

	device := &Device{
		EndpointRef:     match.EndpointRef,
		XAddrs:          parseSpaceSeparated(match.XAddrs),
		Types:           parseSpaceSeparated(match.Types),
		Scopes:          parseSpaceSeparated(match.Scopes),
		MetadataVersion: match.MetadataVersion,
	}
	device.Name, device.Hardware, device.Location = scopeFields(device.Scopes)

	return device
}
