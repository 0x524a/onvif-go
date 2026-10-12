# Migrating from v1 to v2

v2 (`github.com/mickeyzzc/onvif-go/v2`) is a deliberately breaking
release: the client split into domain packages, the server was rebuilt
for real-device embedding, and the WS-Discovery codec is shared by both
protocol sides. The v1 tags (`v1.0.0`–`v1.2.0`) stay available and
untouched.

Most v1 source compiles against v2 after the import-path swap — the
client package re-exports the moved symbols as aliases. This guide maps
everything that still needs attention.

## 1. Module path

```bash
go get github.com/mickeyzzc/onvif-go/v2@latest
```

```diff
-import "github.com/mickeyzzc/onvif-go"
+import "github.com/mickeyzzc/onvif-go/v2/onvif"
```

Since `v2.0.0-rc2` the client package lives in the `onvif/` subdirectory,
so the repository root contains no Go source. The package identifier stays
`onvif` — code that already used the rc1 root package only needs the
import-path change.

The client facade shape is unchanged:

```go
client.Device().GetDeviceInformation(ctx)
client.Media().GetProfiles(ctx)
client.PTZ().ContinuousMove(ctx, token, speed, timeout)
```

## 2. Client: what moved (and what didn't)

| v1 | v2 home | `onvif` package alias keeps v1 spelling? |
|---|---|---|
| `onvif.Profile`, media types | `media` package | ✅ yes |
| Device-service ops (`GetCapabilities`, …) | `device` package | ✅ yes |
| `GetUsers`/access policy | `security` package | ✅ yes |
| Relay/digital IO ops | `deviceio` package | ✅ yes |
| Events (pull-point + push, managed subs) | `events` package | ✅ yes |
| Imaging / PTZ ops | `imaging` / `ptz` | ✅ yes |
| `IPAddress`, `IntRectangle`, … | `types` leaf | ✅ yes |
| Error sentinels (`ErrUnauthorized`, …) | `onvif` pkg + domain packages | ✅ yes |
| Per-service endpoint setters | **`Client.SetServiceEndpoint(api.Service, endpoint)`** | ❌ replaced — one method, `api.ServiceDevice/Media/PTZ/Imaging/Events` constants |
| capabilities cache invalidation | `client.InvalidateCapabilitiesCache()` | ✅ yes (passthrough) |

New client capabilities worth adopting (non-breaking):
`WithAuthFallback` (sticky auth ladder), `WithAutoClockSkew`,
`media.SelectMain/SelectSub` profile helpers, managed
`Events().SubscribeEvents`, `discovery.Listener` (passive),
`discovery.ProbeEndpoint` (directed HTTP probing); since v2.4.0 also
push events (`Events().Subscribe` probe → `ErrPushNotSupported` → fall
back to pull-point; `NotifyHandler` for the consumer endpoint; managed
`SubscribePushEvents`), the motion helpers (`events.ParseMotionAlarm`),
typed time sync (`Device().GetDeviceTime` / `SetDeviceTime` /
`SetNTPServer`, writes fail fast with `ErrNoCredentials`), and listener
acceptance constraints (`WithoutONVIFFilter` / `WithListenerTypes` /
`WithListenerScopePrefix` — the listener now drops non-ONVIF responders
by default).

Wire-behavior fix you may notice: the WS-Security `Created` element is
now emitted in the correct `…wss-utility-1.0.xsd` namespace (v1 had a
`wssecurity-utility` typo). Lenient devices are unaffected; strict
devices stop dropping the timestamp.

## 3. Server (simulator / embedding)

The SOAP transport was rebuilt (M2–M3). If you embed `server/`:

| v1 | v2 |
|---|---|
| `MessageHandler func(body interface{})` | `RegisterContextHandler(action, func(rc *soap.RequestContext, body []byte))`; legacy `RegisterHandler` wraps a `func([]byte)` |
| request body arrived **always empty** (silent bug) | handlers receive the raw inner XML bytes — `soap.ParseRequest` decodes them |
| all actions forced digest auth when credentials were set | per-action policy: `Set*/Remove*/Create*/Go*` + `SystemReboot` (+ `Config.AuthProtectedActions`) authenticated, reads open; `soap.AuthPolicy{All: true}` restores strict mode |
| PasswordDigest only | PasswordDigest **and** PasswordText accepted |
| `GetStreamURIResponse` / `MediaURI` spellings | canonical WSDL casing: `GetStreamUriResponse` / `MediaUri` (incoming legacy spellings still dispatch) |
| `encoding/xml`-only responses | golden-locked envelope bytes; `Config.ExplicitPrefixes` → `s:/tds:/trt:` form; `soap.RawXML` verbatim passthrough |
| XAddr host from `Config.Host` | requester-IP echo by default; `Config.AdvertiseHost` overrides |
| state inside `Server` | pluggable providers (`server/provider`, default `server/simulator`); inject via `server.New(config, WithStreamURIProvider(…), …)` |
| `GET /snapshot` returned empty body | serves provider JPEGs (simulator generates real JPEGs) |
| `ProfileConfig.ToONVIFProfile()` | `server.ProfileToONVIF(p)` (methods cannot follow type aliases) |
| `Server.Handle*(body)` | `Server.Handle*(rc *soap.RequestContext, body []byte)` |

`server.New(Config)` with no options is still the full simulator — the
CLI (`cmd/onvif-server`) and the examples are unchanged.

## 4. Discovery

`discovery.Discover`, `Listener`, and `ProbeEndpoint`/`ProbeSerial` keep
their signatures. `discovery.ProbeMatch` is now an alias of
`wsdiscovery.Match` (same fields; the unused `XMLName` field is gone —
drop it if you constructed the type yourself).

New on the device side: `server/discovery.Responder` (multicast Probe
answering + Hello/Bye + directed HTTP probe handler) built on the shared
`wsdiscovery` codec.

## 5. v2.0.0-rc4 hygiene changes (library-embedding breaks)

Three small breaks landed with the rc4 hygiene pass; none affect the
client or the wire format:

| Before (≤ rc3) | rc4 |
|---|---|
| `server.DefaultConfig()` shipped `admin`/`admin` credentials | credentials are **empty**; without credentials the server runs in its documented everything-open mode — set `Username`/`Password` explicitly for real networks |
| `github.com/mickeyzzc/onvif-go/v2/testing` (public helper package; linked `httptest` into consumer binaries) | `…/v2/internal/onviftesting` — unreachable from outside the module by design |
| `soap.DefaultProtectedPrefixes`, `server.DefaultScopes`, `discovery.DefaultProbePorts` (exported mutable package vars) | functions returning fresh slices of the same names — appending to a package var no longer mutates process-wide defaults |
| `README.zh.md` | renamed `README.zh-CN.md` |

The repo-root `hygiene_test.go` pins all of these so they cannot
silently regress.

## 6. Quick checklist

1. Bump the import path to `/v2`; run `go build`.
2. Replace per-service endpoint setters with `SetServiceEndpoint`.
3. If you embedded `server/`: adopt the context handler signature and
   review the auth-policy table above; consider swapping in providers
   for your real state sources.
4. If you imported `v2/testing` or mutated the default slice vars: see
   §5 — switch to `internal/onviftesting` (vendor the helpers if you
   need them externally) and call the new functions.
5. Re-run your integration tests; byte-level consumers of server
   responses should pin the golden forms from `server/soap/response_test.go`.
