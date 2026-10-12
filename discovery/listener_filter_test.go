package discovery

// Listener filtering (issue #128): the built-in ONVIF signal gate (the
// #555 zombie-device lesson — Windows/WSD hosts announcing generic Hellos
// must not reach the handler) plus the configurable Types/Scopes
// constraints.

import (
	"strings"
	"sync"
	"testing"
)

// wsdWindowsHello is the #555 golden: a Windows host's WS-Discovery Hello —
// generic types, WSD-flavored scopes, an XAddr that will never speak ONVIF.
const wsdWindowsHello = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<Hello xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
<EndpointReference><Address>urn:uuid:wsd-host-4444</Address></EndpointReference>
<Types>wsdiscovery:Device</Types>
<Scopes>http://schemas.xmlsoap.org/ws/2005/04/discovery/stack</Scopes>
<XAddrs>http://192.168.1.50:5000/wsd/</XAddrs>
<MetadataVersion>1</MetadataVersion>
</Hello>
</s:Body>
</s:Envelope>`

// onvifScopeOnlyHello carries no NetworkVideoTransmitter type but an
// onvif:// scope — the second accepted ONVIF signal — plus a location
// scope for the prefix constraint.
const onvifScopeOnlyHello = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<Hello xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
<EndpointReference><Address>urn:uuid:scope-only-5555</Address></EndpointReference>
<Types>tdn:NetworkVideoDisplay</Types>
<Scopes>onvif://www.onvif.org/name/ScopeCam onvif://www.onvif.org/location/Building1/Floor2</Scopes>
<XAddrs>http://192.168.1.99:80/onvif/device_service</XAddrs>
<MetadataVersion>1</MetadataVersion>
</Hello>
</s:Body>
</s:Envelope>`

// nvtWithoutScopes carries the NetworkVideoTransmitter type and no scopes.
const nvtWithoutScopes = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<Hello xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
<EndpointReference><Address>urn:uuid:nvt-6666</Address></EndpointReference>
<Types>dp0:NetworkVideoTransmitter</Types>
<XAddrs>http://192.168.1.66:80/onvif/device_service</XAddrs>
<MetadataVersion>1</MetadataVersion>
</Hello>
</s:Body>
</s:Envelope>`

func TestListenerConfigAcceptsTable(t *testing.T) {
	wsd := parseDiscoveryDatagram([]byte(wsdWindowsHello))
	nvt := parseDiscoveryDatagram([]byte(nvtWithoutScopes))
	scope := parseDiscoveryDatagram([]byte(onvifScopeOnlyHello))

	if wsd == nil || nvt == nil || scope == nil {
		t.Fatalf("fixture parse failed: %v / %v / %v", wsd, nvt, scope)
	}

	cases := []struct {
		name string
		cfg  listenerConfig
		dev  *Device
		want bool
	}{
		{"default drops WSD host", newListenerConfig(), wsd, false},
		{"default keeps NVT", newListenerConfig(), nvt, true},
		{"default keeps onvif scope", newListenerConfig(), scope, true},
		{"filter off admits WSD host", oneOption(WithoutONVIFFilter()), wsd, true},
		{
			"types constraint drops other ONVIF device",
			oneOption(WithListenerTypes("NetworkVideoTransmitter")),
			scope, false,
		},
		{
			"types constraint matches ignoring prefixes",
			oneOption(WithListenerTypes("NetworkVideoTransmitter")),
			nvt, true,
		},
		{
			"scope prefix narrows to Building1",
			oneOption(WithListenerScopePrefix("onvif://www.onvif.org/location/Building1")),
			scope, true,
		},
		{
			"scope prefix rejects non-matching scope",
			oneOption(WithListenerScopePrefix("onvif://www.onvif.org/location/Other")),
			nvt, false,
		},
		{
			"ONVIF gate fires before the types constraint",
			oneOption(WithListenerTypes("Device")),
			wsd, false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.accepts(tc.dev); got != tc.want {
				t.Errorf("accepts = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestListenerDropsNonONVIFHello(t *testing.T) {
	var mu sync.Mutex
	var refs []string

	l := startLoopbackListener(t, func(d *Device) {
		mu.Lock()
		defer mu.Unlock()
		refs = append(refs, d.EndpointRef)
	})

	l.sendRaw(t, wsdWindowsHello)
	l.sendOnce(t, helloDatagram, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(refs) > 0
	})

	mu.Lock()
	defer mu.Unlock()

	if len(refs) != 1 || refs[0] != "urn:uuid:hello-1111" {
		t.Errorf("listener delivered %v, want only the camera Hello (the WSD host must be filtered)", refs)
	}
}

func TestListenerWithoutONVIFFilterAdmitsWSDHost(t *testing.T) {
	var mu sync.Mutex
	var xaddrs []string

	l := startLoopbackListenerOpts(t, func(d *Device) {
		mu.Lock()
		defer mu.Unlock()
		xaddrs = append(xaddrs, d.XAddrs...)
	}, WithoutONVIFFilter())

	l.send(t, wsdWindowsHello, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(xaddrs) > 0
	})

	mu.Lock()
	defer mu.Unlock()

	if len(xaddrs) != 1 || !strings.Contains(xaddrs[0], ":5000/wsd/") {
		t.Errorf("unfiltered listener delivered %v, want the WSD host", xaddrs)
	}
}

func TestListenerTypesConstraintOnTheWire(t *testing.T) {
	var mu sync.Mutex
	var refs []string

	l := startLoopbackListenerOpts(t, func(d *Device) {
		mu.Lock()
		defer mu.Unlock()
		refs = append(refs, d.EndpointRef)
	}, WithListenerTypes("NetworkVideoTransmitter"))

	l.sendRaw(t, onvifScopeOnlyHello) // ONVIF signal via scopes, wrong type
	l.sendOnce(t, nvtWithoutScopes, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(refs) > 0
	})

	mu.Lock()
	defer mu.Unlock()

	if len(refs) != 1 || refs[0] != "urn:uuid:nvt-6666" {
		t.Errorf("types-constrained listener delivered %v, want only the NVT", refs)
	}
}

func TestListenerScopePrefixOnTheWire(t *testing.T) {
	var mu sync.Mutex
	var refs []string

	l := startLoopbackListenerOpts(t, func(d *Device) {
		mu.Lock()
		defer mu.Unlock()
		refs = append(refs, d.EndpointRef)
	}, WithListenerScopePrefix("onvif://www.onvif.org/location/Building1"))

	l.sendRaw(t, nvtWithoutScopes) // no scopes at all
	l.sendOnce(t, onvifScopeOnlyHello, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(refs) > 0
	})

	mu.Lock()
	defer mu.Unlock()

	if len(refs) != 1 || refs[0] != "urn:uuid:scope-only-5555" {
		t.Errorf("scope-constrained listener delivered %v, want only the Building1 device", refs)
	}
}

// oneOption builds a config from a single option for table tests.
func oneOption(opt ListenerOption) listenerConfig {
	cfg := newListenerConfig()
	opt(&cfg)

	return cfg
}
