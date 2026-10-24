# ONVIF Server

`server/` is an ONVIF server implementation you can embed or run
standalone: a virtual-camera simulator with Device, Media, PTZ, and
Imaging services. The v2 transport (`server/soap`) is rebuilt for real
device embedding — request context, per-action authentication, and
byte-predictable XML output.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## Transport architecture

```http
POST /onvif/device_service
        │
        ▼
soap.Handler.ServeHTTP
        │  1. extract action (local name, canonicalized: GetStreamURI → GetStreamUri)
        │  2. decode envelope (WS-Security header + raw inner-XML body)
        │  3. requiresAuth(action)? → authenticate(header)
        │  4. dispatch: ContextHandler(RequestContext, body []byte)
        ▼
response writer (golden-locked byte layout, RawXML passthrough)
```

Request bodies reach handlers as **raw inner XML bytes** — `soap.ParseRequest`
(or plain `encoding/xml`) decodes them into typed requests. The v1 transport
handed handlers a value that `encoding/xml` could never populate, so request
parameters (e.g. `ProfileToken`) silently arrived empty; v2 fixes the pipe.

## Request context

Handlers registered with `RegisterContextHandler` see where requests come
from — the prerequisite for multi-interface address advertising:

```go
handler.RegisterContextHandler("GetStreamUri", func(rc *soap.RequestContext, body []byte) (interface{}, error) {
    ip := rc.RemoteIP              // client IP (host part of RemoteAddr)
    ctx := rc.Context()            // request context: cancellation, deadlines
    _ = rc.Request                 // underlying *http.Request
    ...
})
```

`rc.Action` carries the canonical WSDL action name. The legacy
`RegisterHandler(action, func(body []byte) ...)` signature keeps working as
a thin wrapper.

### XAddr echo

Real cameras answer each client with URLs reachable *from that client's
network*. The simulator does the same by default:

- `Config.AdvertiseHost` empty (default): the requesting client's IP is
  echoed as the host of every advertised URL — GetCapabilities/GetServices
  XAddrs, stream and snapshot URIs.
- `Config.AdvertiseHost` set: that host wins everywhere (fixed DNS name,
  reverse proxy, ...).

## Per-action authentication (#16)

`server.DefaultConfig()` ships with **empty credentials** — a library default
must not carry a guessable pair. Without credentials everything is open
(fine for local simulators); set `Username`/`Password` explicitly before
exposing the server to a real network:

```go
cfg := server.DefaultConfig()
cfg.Username = "admin"      // your choice
cfg.Password = "..."        // from host configuration, never committed
```

With credentials configured, the default policy authenticates **write-style
actions only** — `Set*`, `Remove*`, `Create*`, `Go*`, plus `SystemReboot`
and any names in `Config.AuthProtectedActions`. Read operations stay open
for credential-less discovery clients. Without credentials everything is
open (unchanged).

Both WS-Security UsernameToken password forms are accepted:

| Form | Validation |
|---|---|
| PasswordDigest (default) | `Base64(SHA1(nonce + created + password))`, constant-time compare |
| PasswordText | cleartext compare; disable via `AuthPolicy{AllowPasswordText: false}` |

```go
handler := soap.NewHandlerWithOptions(soap.HandlerOptions{
    Username: "admin",
    Password: "secret",
    Auth: &soap.AuthPolicy{          // nil → DefaultAuthPolicy()
        Prefixes: []string{"Set", "Remove", "Create", "Go"},
        Actions: []string{"SystemReboot"},
        All:     false,              // true = strict mode (v1 behavior)
    },
})
```

## Byte-level XML output (#18)

The response writer builds the envelope by hand; its byte layout is locked
by golden tests, so byte-matching consumers stay stable.

**Canonical casing.** Response elements use the ONVIF WSDL spellings —
`GetStreamUriResponse`/`GetSnapshotUriResponse` (not `...URI...`), inner
`MediaUri`/`Uri`. Incoming legacy spellings (`GetStreamURI`) still dispatch
to the canonical handlers.

**Default form** (matches the historical wire format):

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Body>
    <GetStreamUriResponse xmlns="http://www.onvif.org/ver10/media/wsdl">
      <MediaUri>
        <Uri>rtsp://198.51.100.7:8554/stream0</Uri>
```

**Explicit prefixes** (`Config.ExplicitPrefixes` or
`HandlerOptions.ExplicitPrefixes`):

```xml
<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:MediaUri>
        <trt:Uri>rtsp://198.51.100.7:8554/stream0</trt:Uri>
```

Prefixes: `tds:` `trt:` `tptz:` `timg:` `tev:` `tt:` `trc:` `tan:`.
Namespaces without a conventional prefix keep a default `xmlns` declaration.

**Raw channel.** Return `soap.RawXML` from any handler to embed pre-built
bytes verbatim — bypassing `encoding/xml` entirely for exact element
names, prefixes, or formatting. Raw output is never rewritten by prefix
modes.

## Pluggable providers (M3)

Every state source behind the SOAP layer is an interface in
`server/provider` — the simulator (`server/simulator`) is just the
default implementation:

| Provider | Backs | Swap in for |
|---|---|---|
| `DeviceInfoProvider` | GetDeviceInformation | hardware identity / EEPROM |
| `StreamURIProvider` (+ optional `StreamURISetter`) | GetStreamUri, RTSP paths | real RTSP mount points |
| `SnapshotProvider` | HTTP `/snapshot` endpoint | JPEG frame buffer |
| `ImagingProvider` | Get/SetImagingSettings, GetOptions, Move | ISP / lens controls |
| `PTZProvider` | moves, status, presets | motor drivers (or keep the simulator) |

```go
srv, err := server.New(config,
    server.WithDeviceInfoProvider(hwInfo),
    server.WithStreamURIProvider(rtspMounts),
    server.WithSnapshotProvider(frameBuffer),
    server.WithImagingProvider(isp),
)
```

No options = full simulator behavior (CLI and examples are unchanged).
The SOAP handlers are stateless translators now; `Server.Handle*`
signatures, the exported model types, and the domain error sentinels
keep their `server.*` spellings via aliases.

**Preset store (v2.2.0):** `PTZProvider` keeps serving moves/status/
GotoPreset; the optional `PTZPresetReader` upgrades GetPresets beyond
the static configuration, and `PTZPresetWriter` unlocks SetPreset /
RemovePreset — the simulator implements both with a mutable per-profile
store (set → list → goto → remove works end to end out of the box).
`GetConfigurations` and `GetNodes` answer from the profile
configuration with WSDL-correct namespaces.


## TLS transport (Profile T baseline)

Since v2.1.0 the server serves HTTPS when `Config.TLSCertFile` and
`Config.TLSKeyFile` are set (both, or neither — a lone one fails
validation). No separate start call: TLS is decided inside `Start`.
`Start` binds its listener explicitly, so `Server.ListenAddr()` reports
the bound address as soon as it runs — observable with `Port: 0`
(kernel-assigned) for tests:

```go
config := server.DefaultConfig()
config.TLSCertFile = "cert.pem"
config.TLSKeyFile = "key.pem"

srv, _ := server.New(config)
go srv.Start(ctx)                           // https://…/onvif/device_service
fmt.Println(srv.ListenAddr())               // e.g. 0.0.0.0:8443
```

Scope: plain TLS transport (the Profile T baseline). Advanced
onboarding — PKI, 802.1X, WS-Security beyond UsernameToken — is
deliberately out of scope.

The runnable `onvif server` binary speaks plain HTTP only; TLS is the
embedded package's surface ([onvif server CLI](onvif-go-cli-server.md)).

## Where to go next

Pluggable providers (above) front real camera stacks; the pull-point
events service (`SupportEvents` + the `PublishEvent` seam) is covered in
the events manual. See [v2-architecture.md](onvif-go-v2-architecture.md)
for the layering.
