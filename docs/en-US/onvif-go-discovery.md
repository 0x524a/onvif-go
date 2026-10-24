# Device Discovery

onvif-go ships three complementary discovery modes plus a post-processing
layer. All live in the self-contained `discovery` package.

| Mode | Function | Reaches |
|---|---|---|
| Active probe | `Discover` / `DiscoverWithOptions` | Same subnet (UDP multicast) |
| Passive listener | `Listener` | Same subnet, zero latency |
| Directed probe | `ProbeEndpoint` / `ProbeSerial` | Any address, cross-subnet, pure HTTP |


## Discovery Exchange

```mermaid
sequenceDiagram
    participant C as discovery client
    participant D as ONVIF device
    C->>D: UDP Probe (multicast 3702)
    D-->>C: ProbeMatches (XAddrs)
    C->>D: HTTP POST GetCapabilities
    D-->>C: Capability document (Media/Device addresses)
    C->>D: GetProfiles / GetStreamUri
    D-->>C: Profiles and stream URI
```

Active and directed probes share the same Probe/ProbeMatches parsing; the passive `Listener` only observes existing probes on the network and never transmits.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## Active: multicast probe

```go
devices, err := discovery.Discover(ctx, 5*time.Second)
// or pin the multicast interface (name or IP) on multi-NIC hosts:
devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second,
    &discovery.DiscoverOptions{NetworkInterface: "eth0"})
```

Polling this once a minute is the usual pattern — which wastes the Hello a
camera broadcasts at power-on unless something is listening. That is what
the passive listener is for.

## Passive: the resident listener

```go
listener, _ := discovery.NewListener("", func(d *discovery.Device) {
    fmt.Println("camera online:", d.GetDeviceEndpoint())
})
go func() { _ = listener.Start(ctx) }()
defer listener.Stop()
```

- Hears **Hello** (power-on self-announcement) and **ProbeMatches** (other
  devices answering someone else's probe — a free second discovery source);
  Bye and garbage are ignored.
- **Coexists with active discovery**: it joins the WS-Discovery multicast
  group exactly like `DiscoverWithOptions` does (SO_REUSEADDR), and on
  Linux/macOS the kernel delivers a copy of each datagram to every bound
  socket — a listener plus on-demand probes in one process do not steal each
  other's traffic.
- `ifaceName` `""` = kernel default interface (recommended on single-NIC
  hosts).
- The handler runs on the listener goroutine: return quickly, offload slow
  work. Panics in the handler are contained.
- `Stop()` is idempotent; a stopped listener refuses restart. `Done()`
  closes when the loop has fully exited.

## Directed: pure-HTTP probing

Multicast does not route across subnets, and some devices never answer
WS-Discovery probes. The camera-rediscovery pattern (a camera changed IP and
must be found again by serial) needs pure HTTP:

```go
// Two strategies, in order: a WS-Discovery Probe POSTed to
// http://host:port/onvif/device_service, then an unauthenticated
// GetDeviceInformation. nil = not ONVIF / offline (the two are
// indistinguishable from outside; both mean "not found").
dev := discovery.ProbeEndpoint(ctx, "192.0.2.50", 80, 1200*time.Millisecond)

// Serial number via common-port scan (80, 8080, 8000 by default). The
// serial is extracted namespace-agnostically — it is the stable identity
// anchor for correlating one physical camera across protocols.
serial, ok := discovery.ProbeSerial(ctx, "192.0.2.50", nil)
```

Malformed responses from arbitrary hosts are contained (never a panic);
401/405-class answers count as "not ONVIF for probing purposes".

## Post-processing

Three helpers turn raw discovery results into usable device lists:

```go
usable := discovery.FilterONVIFDevices(devices) // drop ghost responders
info := discovery.ParseScopes(dev.Scopes)       // name / hardware / location
discovery.EnrichDevices(ctx, usable)            // parallel identity fetch
```

- **`FilterONVIFDevices`** — generic WS-Discovery responders (Synology DSM,
  Windows hosts, printers) answer any Probe regardless of its Types filter
  and otherwise become forever-pending ghosts in camera inventories. A
  device is kept when its Types contain `NetworkVideoTransmitter` (local
  part, prefix-agnostic) **or** any scope starts with `onvif://www.onvif.org/`
  — deliberately lenient so marginal implementations survive.
- **`ParseScopes`** — the ONVIF scope conventions (`onvif://www.onvif.org/
  name/X`, `.../hardware/Y`, `.../location/Z`), percent-decoded. The
  structured fields are also filled directly on every `Device` the package
  produces (`d.Name`, `d.Hardware`, `d.Location`).
- **`EnrichDevices`** — fetches unauthenticated `GetDeviceInformation` for
  every device in parallel (bounded concurrency via `WithEnrichConcurrency`,
  per-device timeout via `WithEnrichTimeout`), filling `d.Info`
  (manufacturer / model / firmware / serial). Best-effort: unreachable
  devices are skipped, never fatal; devices already carrying `Info` are
  untouched.

## A note on identity

`Device.EndpointRef` is the WS-Discovery endpoint address (a `urn:uuid:...`
form) — **not** the device serial number. Correlating a camera across
protocols (ONVIF vs GB28181) must use `Device.Info.SerialNumber`; comparing
`EndpointRef` against a serial silently never matches.

## Device side: the WS-Discovery responder

The same codec powers the device side (`server/discovery`): a resident
multicast `:3702` listener that answers Probe messages with
ProbeMatches (unicast to the probing socket), announces Hello on start
and Bye on stop, and answers directed Probe-over-HTTP POSTs through an
`http.Handler` — the transport `ProbeEndpoint` uses across subnets.

```go
responder := serverdiscovery.NewResponder(serverdiscovery.Config{
    Types:   []string{"tds:Device", "dp0:NetworkVideoTransmitter"},
    Scopes:  []string{"onvif://www.onvif.org/name/MiBeeEye"},
    Port:    8080, // derived XAddrs: http://<requester IP>:8080/onvif/device_service
})
err := responder.Start(ctx)
defer responder.Stop()
mux.Handle("/onvif/device_service", soapHandler)   // SOAP
mux.Handle("/onvif/discovery", responder)          // directed HTTP probes
```

With `XAddrs` empty the responder echoes the requesting peer's IP — the
same per-peer reachability rule the SOAP layer applies to XAddr
responses. Answer content (Types/Scopes/XAddrs/MetadataVersion) is fully
configurable; probes whose Types filter does not select the configured
types are ignored, per WS-Discovery semantics.

Coexistence: the responder, `Discover()`, and the passive `Listener`
all bind the multicast group the same way; on Linux and macOS the
kernel delivers a copy of each multicast datagram to every bound
socket, so device side and client side coexist in one process without
stealing each other's traffic. Answers are always unicast.

The wire codec itself lives in `wsdiscovery/` (shared by both sides):
`BuildProbe`/`ParseProbe`, `BuildProbeMatches`/`ParseProbeMatches`,
`BuildHello`/`BuildBye`, and `ParseAnnouncement` — one definition, no
private parsing drift between client and device.
