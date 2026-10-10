package discovery

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	nsDiscovery  = "http://schemas.xmlsoap.org/ws/2005/04/discovery"
	nsAddressing = "http://schemas.xmlsoap.org/ws/2004/08/addressing"
	nsONVIFNet   = "http://www.onvif.org/ver10/network/wsdl"

	actionProbeMatches = nsDiscovery + "/ProbeMatches"
	actionHello        = nsDiscovery + "/Hello"
	actionBye          = nsDiscovery + "/Bye"

	toDiscovery = "urn:schemas-xmlsoap-org:ws:2005:04:discovery"
	toAnonymous = nsAddressing + "/role/anonymous"

	// Scope matching rules from WS-Discovery section 5.1.
	matchByRFC3986 = nsDiscovery + "/rfc3986"
	matchByStrcmp0 = nsDiscovery + "/strcmp0"
	matchByNone    = nsDiscovery + "/none"

	maxDatagram = 65535
)

// DefaultType is the type a registered device advertises when it lists none:
// dn:NetworkVideoTransmitter, which is what ONVIF clients probe for.
const DefaultType = "{" + nsONVIFNet + "}NetworkVideoTransmitter"

// seedNamespace scopes StableEndpointRef so seeds hash to UUIDs unique to this
// library. It must never change: doing so would renumber every device.
var seedNamespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://github.com/0x524a/onvif-go/endpoint"))

// StableEndpointRef derives a "urn:uuid:" endpoint reference from seed. The
// same seed always yields the same reference, on any host and across restarts,
// so a device keeps its identity without anything being persisted. Use a name
// that is stable for the device (for example "farm1/cam-07"); an empty seed is
// valid but gives every caller the same reference.
func StableEndpointRef(seed string) string {
	return "urn:uuid:" + uuid.NewSHA1(seedNamespace, []byte(seed)).String()
}

// ErrInvalidDevice is returned by Responder.Register for a device that cannot
// be advertised.
var ErrInvalidDevice = errors.New("invalid device")

// ErrMalformedProbe marks a Probe the responder cannot interpret.
var ErrMalformedProbe = errors.New("malformed probe")

// errBadTypeName marks a device type string in an unsupported form.
var errBadTypeName = errors.New("unsupported type form")

// ResponderConfig configures a Responder.
type ResponderConfig struct {
	// NetworkInterface selects the interface that joins the multicast group,
	// by name or IP address, as in DiscoverOptions. Ignored when ListenAddr
	// is set.
	NetworkInterface string

	// ListenAddr switches the responder to unicast mode. It binds this UDP
	// address (for example "127.0.0.1:0") instead of joining
	// 239.255.255.250:3702, so it needs no multicast routing and several
	// responders can coexist on one host. Pair it with
	// DiscoverOptions.ProbeAddress. Intended for tests and CI.
	ListenAddr string

	// AnnounceAddr is where Hello and Bye messages are sent. It defaults to
	// the WS-Discovery multicast group in multicast mode. In unicast mode it
	// defaults to empty, which disables announcements.
	AnnounceAddr string

	// MaxResponseDelay bounds a random pause before each ProbeMatches reply,
	// which WS-Discovery recommends (APP_MAX_DELAY, 500ms) to keep a large
	// fleet from answering one probe in the same instant. The default, zero,
	// replies immediately.
	MaxResponseDelay time.Duration

	// Ready, when non-nil, is closed by Start once the socket is bound and
	// Addr is valid. It is closed at most once.
	Ready chan<- struct{}
}

// Responder is a WS-Discovery target service. One Responder owns one socket
// and answers Probe messages on behalf of every registered Device, so a fleet
// of virtual cameras on one host shares a single multicast membership.
//
// Supported: Probe/ProbeMatches with Types and Scopes filtering (MatchBy
// rfc3986, strcmp0 and none), Hello on start, Bye on stop. Not supported:
// Resolve/ResolveMatches, the ldap and uuid scope rules (such a probe matches
// nothing), the 2009/01 (WS-Discovery 1.1) namespace, and managed mode with a
// discovery proxy. Duplicate probes are answered each time.
type Responder struct {
	cfg ResponderConfig

	mu         sync.RWMutex
	devices    map[string]*registered
	conn       *net.UDPConn
	announce   *net.UDPAddr
	readyOnce  sync.Once
	instanceID uint64
	msgNumber  uint64
}

type registered struct {
	device Device
	types  []qname
}

// qname is a namespace-qualified type name.
type qname struct {
	space string
	local string
}

// NewResponder returns a Responder. A nil cfg uses multicast defaults.
func NewResponder(cfg *ResponderConfig) *Responder {
	r := &Responder{
		devices:    make(map[string]*registered),
		instanceID: uint64(time.Now().Unix()), //nolint:gosec // Unix time is positive
	}
	if cfg != nil {
		r.cfg = *cfg
	}

	return r
}

// Register advertises d. EndpointRef is required and identifies the device; a
// second Register with the same EndpointRef replaces the first. Types entries
// are "{namespace}Local" or the prefixed forms "dn:Local" (ONVIF network
// namespace) and "tds:Local" (ONVIF device namespace); no Types means
// DefaultType. If the responder is running, Register sends a Hello.
func (r *Responder) Register(d *Device) error {
	if d == nil || strings.TrimSpace(d.EndpointRef) == "" {
		return fmt.Errorf("%w: EndpointRef is required", ErrInvalidDevice)
	}

	reg := &registered{device: *d}
	reg.device.XAddrs = append([]string(nil), d.XAddrs...)
	reg.device.Scopes = append([]string(nil), d.Scopes...)
	reg.device.Types = append([]string(nil), d.Types...)

	typeNames := d.Types
	if len(typeNames) == 0 {
		typeNames = []string{DefaultType}
	}

	for _, t := range typeNames {
		q, err := parseTypeName(t)
		if err != nil {
			return fmt.Errorf("%w: type %q: %w", ErrInvalidDevice, t, err)
		}

		reg.types = append(reg.types, q)
	}

	r.mu.Lock()
	r.devices[d.EndpointRef] = reg
	running := r.conn != nil
	r.mu.Unlock()

	if running {
		r.send(r.announce, actionHello, reg)
	}

	return nil
}

// Unregister stops advertising the device with the given EndpointRef. If the
// responder is running, it sends a Bye.
func (r *Responder) Unregister(endpointRef string) {
	r.mu.Lock()
	reg := r.devices[endpointRef]
	delete(r.devices, endpointRef)
	running := r.conn != nil
	r.mu.Unlock()

	if reg != nil && running {
		r.send(r.announce, actionBye, reg)
	}
}

// Addr returns the bound socket address, or nil before Start has bound it.
func (r *Responder) Addr() net.Addr {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.conn == nil {
		return nil
	}

	return r.conn.LocalAddr()
}

// Start binds the socket, sends Hello for every registered device, and answers
// probes until ctx is canceled, then sends Bye for every device still
// registered and returns nil. A bind failure is returned immediately.
func (r *Responder) Start(ctx context.Context) error {
	conn, announce, err := r.listen()
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.conn = conn
	r.announce = announce
	r.mu.Unlock()

	if r.cfg.Ready != nil {
		r.readyOnce.Do(func() { close(r.cfg.Ready) })
	}

	for _, reg := range r.snapshot() {
		r.send(announce, actionHello, reg)
	}

	stopped := make(chan struct{})
	byeDone := make(chan struct{})

	go func() {
		defer close(byeDone)

		select {
		case <-ctx.Done():
		case <-stopped:
		}

		for _, reg := range r.snapshot() {
			r.send(announce, actionBye, reg)
		}

		_ = conn.Close()
	}()

	err = r.serve(ctx, conn)

	close(stopped)
	<-byeDone

	r.mu.Lock()
	r.conn = nil
	r.mu.Unlock()

	return err
}

func (r *Responder) serve(ctx context.Context, conn *net.UDPConn) error {
	buf := make([]byte, maxDatagram)

	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}

			return fmt.Errorf("failed to read UDP probe: %w", err)
		}

		req, err := parseProbe(buf[:n])
		if err != nil {
			continue
		}

		for _, reg := range r.snapshot() {
			if !matches(req, reg) {
				continue
			}

			r.reply(ctx, conn, src, req.messageID, reg)
		}
	}
}

func (r *Responder) reply(ctx context.Context, conn *net.UDPConn, dst *net.UDPAddr, relatesTo string, reg *registered) {
	msg := r.buildProbeMatches(relatesTo, reg)

	if r.cfg.MaxResponseDelay <= 0 {
		writeDatagram(conn, msg, dst)

		return
	}

	delay := time.Duration(time.Now().UnixNano() % int64(r.cfg.MaxResponseDelay))
	time.AfterFunc(delay, func() {
		if ctx.Err() == nil {
			writeDatagram(conn, msg, dst)
		}
	})
}

func (r *Responder) listen() (*net.UDPConn, *net.UDPAddr, error) {
	if r.cfg.ListenAddr != "" {
		laddr, err := net.ResolveUDPAddr("udp", r.cfg.ListenAddr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve listen address: %w", err)
		}

		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to listen on %s: %w", r.cfg.ListenAddr, err)
		}

		announce, err := resolveAnnounce(r.cfg.AnnounceAddr)
		if err != nil {
			_ = conn.Close()

			return nil, nil, err
		}

		return conn, announce, nil
	}

	group, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	var iface *net.Interface
	if r.cfg.NetworkInterface != "" {
		iface, err = resolveNetworkInterface(r.cfg.NetworkInterface)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve network interface: %w", err)
		}
	}

	conn, err := net.ListenMulticastUDP("udp", iface, group)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on multicast address: %w", err)
	}

	announce := group
	if r.cfg.AnnounceAddr != "" {
		if announce, err = resolveAnnounce(r.cfg.AnnounceAddr); err != nil {
			_ = conn.Close()

			return nil, nil, err
		}
	}

	return conn, announce, nil
}

func resolveAnnounce(addr string) (*net.UDPAddr, error) {
	if addr == "" {
		return nil, nil //nolint:nilnil // nil means announcements are disabled
	}

	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve announce address: %w", err)
	}

	return a, nil
}

func (r *Responder) snapshot() []*registered {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*registered, 0, len(r.devices))
	for _, reg := range r.devices {
		out = append(out, reg)
	}

	return out
}

// send writes a Hello or Bye for reg to dst; a nil dst is a no-op.
func (r *Responder) send(dst *net.UDPAddr, action string, reg *registered) {
	r.mu.RLock()
	conn := r.conn
	r.mu.RUnlock()

	if dst == nil || conn == nil {
		return
	}

	writeDatagram(conn, r.buildAnnouncement(action, reg), dst)
}

// probe is the subset of a Probe message the responder acts on.
type probe struct {
	messageID string
	types     []qname
	scopes    []string
	matchBy   string
}

// parseProbe extracts a Probe from a SOAP datagram. It errors for anything
// that is not a WS-Discovery 2005/04 Probe.
func parseProbe(data []byte) (*probe, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false

	var (
		frames []map[string]string
		req    probe
		found  bool
	)

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			consumed, err := readProbeElement(dec, &t, frames, &req, &found)
			if err != nil {
				return nil, err
			}

			if !consumed {
				frames = append(frames, declared(t.Attr))
			}
		case xml.EndElement:
			if len(frames) > 0 {
				frames = frames[:len(frames)-1]
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("%w", ErrNoProbeMatches)
	}

	return &req, nil
}

// readProbeElement handles the elements parseProbe cares about. It reports
// whether it consumed the element through its end tag, in which case the
// caller must not track it as an open frame.
func readProbeElement(
	dec *xml.Decoder, start *xml.StartElement, frames []map[string]string, req *probe, found *bool,
) (bool, error) {
	if start.Name.Local == "MessageID" {
		var s string
		if dec.DecodeElement(&s, start) == nil {
			req.messageID = strings.TrimSpace(s)
		}

		return true, nil
	}

	if start.Name.Space != nsDiscovery {
		return false, nil
	}

	switch start.Name.Local {
	case "Probe":
		*found = true
	case "Types":
		if *found {
			return true, readProbeTypes(dec, start, frames, req)
		}
	case "Scopes":
		if *found {
			return true, readProbeScopes(dec, start, req)
		}
	}

	return false, nil
}

func readProbeTypes(dec *xml.Decoder, start *xml.StartElement, frames []map[string]string, req *probe) error {
	ns := map[string]string{}
	for _, f := range frames {
		maps.Copy(ns, f)
	}

	maps.Copy(ns, declared(start.Attr))

	var s string
	if err := dec.DecodeElement(&s, start); err != nil {
		return fmt.Errorf("%w: Types: %w", ErrMalformedProbe, err)
	}

	for name := range strings.FieldsSeq(s) {
		q, ok := resolveQName(name, ns)
		if !ok {
			return fmt.Errorf("%w: unresolvable type %q", ErrMalformedProbe, name)
		}

		req.types = append(req.types, q)
	}

	return nil
}

func readProbeScopes(dec *xml.Decoder, start *xml.StartElement, req *probe) error {
	for _, a := range start.Attr {
		if a.Name.Local == "MatchBy" {
			req.matchBy = strings.TrimSpace(a.Value)
		}
	}

	var s string
	if err := dec.DecodeElement(&s, start); err != nil {
		return fmt.Errorf("%w: Scopes: %w", ErrMalformedProbe, err)
	}

	req.scopes = strings.Fields(s)

	return nil
}

// declared returns the prefix to namespace declarations among attrs. The
// default namespace is stored under the empty prefix.
func declared(attrs []xml.Attr) map[string]string {
	m := map[string]string{}

	for _, a := range attrs {
		switch {
		case a.Name.Space == "xmlns":
			m[a.Name.Local] = a.Value
		case a.Name.Space == "" && a.Name.Local == "xmlns":
			m[""] = a.Value
		}
	}

	return m
}

func resolveQName(name string, ns map[string]string) (qname, bool) {
	prefix, local, hasPrefix := strings.Cut(name, ":")
	if !hasPrefix {
		prefix, local = "", name
	}

	space, ok := ns[prefix]
	if !ok && prefix != "" {
		return qname{}, false
	}

	return qname{space: space, local: local}, local != ""
}

func parseTypeName(s string) (qname, error) {
	switch {
	case strings.HasPrefix(s, "{"):
		space, local, ok := strings.Cut(s[1:], "}")
		if !ok || local == "" {
			return qname{}, fmt.Errorf("%w: malformed {namespace}Local", errBadTypeName)
		}

		return qname{space: space, local: local}, nil
	case strings.HasPrefix(s, "dn:"):
		return qname{space: nsONVIFNet, local: strings.TrimPrefix(s, "dn:")}, nil
	case strings.HasPrefix(s, "tds:"):
		return qname{space: "http://www.onvif.org/ver10/device/wsdl", local: strings.TrimPrefix(s, "tds:")}, nil
	}

	return qname{}, fmt.Errorf("%w: use {namespace}Local, dn:Local or tds:Local", errBadTypeName)
}

// matches applies the WS-Discovery matching rules: every probed type must be
// among the device's, and every probed scope must match one of its scopes.
func matches(p *probe, reg *registered) bool {
	for _, want := range p.types {
		if !containsQName(reg.types, want) {
			return false
		}
	}

	for _, want := range p.scopes {
		if !anyScopeMatches(p.matchBy, want, reg.device.Scopes) {
			return false
		}
	}

	return true
}

func containsQName(list []qname, want qname) bool {
	return slices.Contains(list, want)
}

func anyScopeMatches(matchBy, want string, have []string) bool {
	for _, h := range have {
		switch matchBy {
		case "", matchByRFC3986:
			if rfc3986Match(want, h) {
				return true
			}
		case matchByStrcmp0:
			if want == h {
				return true
			}
		case matchByNone:
			// "none" matches only a probe with no scopes, which never reaches here.
		}
	}

	return false
}

// rfc3986Match reports whether probe scope want matches target scope have:
// equal scheme and authority (case-insensitive) and a path that is a
// segment-wise prefix of the target's.
func rfc3986Match(want, have string) bool {
	w, err := url.Parse(want)
	if err != nil {
		return false
	}

	h, err := url.Parse(have)
	if err != nil {
		return false
	}

	if !strings.EqualFold(w.Scheme, h.Scheme) || !strings.EqualFold(w.Host, h.Host) {
		return false
	}

	wp := splitPath(w.Path)
	hp := splitPath(h.Path)

	if len(wp) > len(hp) {
		return false
	}

	for i := range wp {
		if wp[i] != hp[i] {
			return false
		}
	}

	return true
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}

	return strings.Split(p, "/")
}

// writeDatagram sends msg best-effort: UDP gives no delivery guarantee, and a
// failed reply to one prober must not disturb the others.
func writeDatagram(conn *net.UDPConn, msg []byte, dst *net.UDPAddr) {
	if _, err := conn.WriteToUDP(msg, dst); err != nil {
		return
	}
}

func xmlEscape(s string) string {
	var b bytes.Buffer

	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return s
	}

	return b.String()
}

// typePrefixes assigns a prefix to each type namespace for emission.
func typePrefixes(types []qname) (decls string, names []string) {
	prefixes := map[string]string{}
	names = make([]string, 0, len(types))

	var b strings.Builder

	for _, q := range types {
		p, ok := prefixes[q.space]
		if !ok {
			p = fmt.Sprintf("t%d", len(prefixes))
			if q.space == nsONVIFNet {
				p = "dn"
			}

			prefixes[q.space] = p
			fmt.Fprintf(&b, ` xmlns:%s="%s"`, p, xmlEscape(q.space))
		}

		names = append(names, p+":"+q.local)
	}

	return b.String(), names
}

func (r *Responder) header(action, relatesTo string) string {
	r.mu.Lock()
	r.msgNumber++
	n := r.msgNumber
	r.mu.Unlock()

	to := toDiscovery
	rel := ""

	if relatesTo != "" {
		to = toAnonymous
		rel = "<a:RelatesTo>" + xmlEscape(relatesTo) + "</a:RelatesTo>"
	}

	return fmt.Sprintf(`<s:Header><a:MessageID>urn:uuid:%s</a:MessageID>%s<a:To>%s</a:To>`+
		`<a:Action>%s</a:Action><d:AppSequence InstanceId="%d" MessageNumber="%d"/></s:Header>`,
		generateUUID(), rel, to, action, r.instanceID, n)
}

const envelopeOpen = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="` + nsAddressing + `" xmlns:d="` + nsDiscovery + `">`

// target renders the elements shared by ProbeMatch, Hello and Bye.
func target(reg *registered, withMeta bool) string {
	decls, names := typePrefixes(reg.types)

	var b strings.Builder

	fmt.Fprintf(&b, `<a:EndpointReference><a:Address>%s</a:Address></a:EndpointReference>`,
		xmlEscape(reg.device.EndpointRef))
	fmt.Fprintf(&b, `<d:Types%s>%s</d:Types>`, decls, xmlEscape(strings.Join(names, " ")))

	if withMeta {
		fmt.Fprintf(&b, `<d:Scopes>%s</d:Scopes><d:XAddrs>%s</d:XAddrs><d:MetadataVersion>%d</d:MetadataVersion>`,
			xmlEscape(strings.Join(reg.device.Scopes, " ")),
			xmlEscape(strings.Join(reg.device.XAddrs, " ")),
			reg.device.MetadataVersion)
	}

	return b.String()
}

func (r *Responder) buildProbeMatches(relatesTo string, reg *registered) []byte {
	return []byte(envelopeOpen + r.header(actionProbeMatches, relatesTo) +
		`<s:Body><d:ProbeMatches><d:ProbeMatch>` + target(reg, true) +
		`</d:ProbeMatch></d:ProbeMatches></s:Body></s:Envelope>`)
}

func (r *Responder) buildAnnouncement(action string, reg *registered) []byte {
	name := "Hello"
	if action == actionBye {
		name = "Bye"
	}

	return []byte(envelopeOpen + r.header(action, "") +
		`<s:Body><d:` + name + `>` + target(reg, action == actionHello) + `</d:` + name + `></s:Body></s:Envelope>`)
}
