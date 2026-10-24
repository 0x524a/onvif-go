# onvif-go

[![CI](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml/badge.svg)](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mickeyzzc/onvif-go/v2.svg)](https://pkg.go.dev/github.com/mickeyzzc/onvif-go/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/mickeyzzc/onvif-go/v2)](https://goreportcard.com/report/github.com/mickeyzzc/onvif-go/v2)
[![License](https://img.shields.io/github/license/mickeyzzc/onvif-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/)

> ONVIF client + virtual camera server for Go — service-facade API, active and
> passive discovery, an authentication ladder that survives weird firmware,
> PTZ / media / events / imaging. Zero third-party dependencies.
> Maintained by [@mickeyzzc](https://github.com/mickeyzzc).

> [**中文**](README.zh-CN.md) — [English](README.md)

A production-hardened Go library for communicating with ONVIF-compliant IP
cameras, NVRs and surveillance devices. Its device compatibility comes from
real deployments: Hikvision, Axis, Dahua, Bosch, Amcrest, HiSilicon-OEM
hardware, minimal embedded implementations (ESP32), and the odd firmware that
answers everything with a SOAP Fault. It powers
[MiBeeNvr](https://github.com/Mi-Bee-Studio/MiBeeNvr), an NVR running
24/7 on ARM64.

## Which onvif-go should I use?

This repository is the continued-maintenance lineage of
[0x524a/onvif-go](https://github.com/0x524a/onvif-go) (its v1.0.0–v1.1.7
tags are preserved in this repo's history; see
[Attribution](#attribution--license)). The upstream project remains
available and maintained — both are MIT, so pick by what you need:

| You want… | Take |
|---|---|
| A virtual camera **server** to embed in your own binary — pluggable state providers, per-action auth policy, a metrics seam, pull-point events | this repo (`/v2`) |
| The service-facade client API, the authentication ladder, active + passive + directed discovery | this repo (`/v2`) |
| An existing codebase already importing `github.com/0x524a/onvif-go` that does everything you need | upstream `v1.1.7` — staying is fine |
| The v2 capabilities above, coming from v1 code | this repo; [MIGRATION.md](MIGRATION.md) maps the move (mostly an import-path swap) |

The module paths differ (`…/onvif-go/v2` vs `…/onvif-go`), so both can even
coexist in one build during a migration.

## Features

**Client** — every ONVIF operation lives on the service object that owns it,
mirroring the ONVIF service model:

- `client.Device()` — info, capabilities (cached, single-flighted,
  degradable to a minimal set for weak devices), network/system/DNS/NTP
  configuration, storage, WiFi, certificates, users; typed time sync
  (`GetDeviceTime` / `SetDeviceTime` / `SetNTPServer` with the
  `ErrNoCredentials` gate — the unauthenticated clock read works exactly
  when drifted clocks have broken digest auth)
- `client.Media()` — profiles with main/sub-stream selection helpers, stream
  URIs (`GetStreamURIWithOptions`: RTSP/HTTP/UDP × unicast/multicast,
  namespace-tolerant response parsing), snapshots, encoder/audio/OSD
  configuration
- `client.PTZ()` — moves, status, presets, configurations + options,
  home position, auxiliary commands (wipers/heaters), nodes, service
  capabilities
- `client.Imaging()` — exposure, focus, imaging settings, move options,
  spec-named `Stop`, service capabilities, imaging presets (18.12)
- `client.Events()` — managed pull-point subscriptions (background polling,
  auto-renewal, `ErrEventsNotSupported` sentinel) plus the raw primitives;
  WS-BaseNotification push (probe with `Subscribe`, fall back to pull-point
  on `ErrPushNotSupported`; Renew/Unsubscribe on the subscription's own
  manager endpoint; `ParseNotifyEvents` + a ready-made `NotifyHandler` for
  the consumer endpoint; the managed `SubscribePushEvents` loop renews at
  half TTL); standard motion-topic helpers (`ParseMotionAlarm` /
  `IsMotionTopic` — MotionAlarm `State` and CellMotionDetector `IsMotion`)
- `client.DeviceIO()`, `client.Security()` — relays, I/O, user management,
  audio output configuration family (audio *decoder* configurations live in
  the Media service per the WSDLs and are covered there)

**Authentication built for real firmware** — `WithAuthMode` selects
digest / password-text / HTTP Basic / none; `WithAuthFallback` adds an
automatic fallback ladder that remembers the first mode the device accepts.
`errors.Is(err, onvif.ErrUnauthorized)` reliably classifies auth failures
(HTTP 401/403, NotAuthorized faults, 200-with-fault). `WithAutoClockSkew`
measures and corrects device clock divergence (the Hikvision "sender not
authorized" trap), and `DiagnoseAuth` tells clock skew, wrong credentials and
auth-less devices apart.

**Discovery, three ways** — `discovery.Discover` (multicast probe),
`discovery.Listener` (passive: hears camera power-on Hello messages and other
probes' answers, coexisting with active discovery; non-ONVIF responders are
dropped before the handler by default, with optional Types/Scopes
constraints), and
`discovery.ProbeEndpoint` / `ProbeSerial` (pure-HTTP directed probing across
subnets for devices multicast can't reach). Post-processing:
`FilterONVIFDevices` (drops Synology/Windows/printer ghost responders),
`ParseScopes`, and parallel `EnrichDevices` (identity fetch with bounded
concurrency).

**Robustness** — fault detection on every call (a 200-with-Fault is never
"success"), structured `FaultError`/`HTTPStatusError`, empty-URI responses
become explicit errors with body summaries, capability XAddr repair (stale
advertised IPs after camera roaming), and a documented, tested concurrency
contract: one `Client`, many goroutines, no external locking.

**Virtual camera server** — `server/` simulates ONVIF cameras for testing
your recorder without hardware: device, media, PTZ, imaging, and the
opt-in events pull-point service (`SupportEvents`), with a
`PublishEvent` seam for injecting notifications from your own analytics.

Scope note: the library handles ONVIF signaling only. It produces and
consumes stream **URIs** (RTSP/HTTP/UDP) but does not transport media
itself — pair it with an RTSP server or client for the media plane.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2
```

Requires Go **1.26+**. The library module has zero third-party dependencies.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mickeyzzc/onvif-go/v2/onvif"
)

func main() {
    client, err := onvif.NewClient("http://192.168.1.100/onvif/device_service",
        onvif.WithCredentials("admin", "pass"),
        onvif.WithAutoClockSkew()) // survive clock-diverged cameras
    if err != nil {
        log.Fatal(err)
    }
    if err := client.Initialize(context.Background()); err != nil {
        log.Fatal(err)
    }

    info, err := client.Device().GetDeviceInformation(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s %s (fw %s, serial %s)\n",
        info.Manufacturer, info.Model, info.FirmwareVersion, info.SerialNumber)

    profiles, err := client.Media().GetProfiles(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    mainToken := onvif.SelectMainProfile(profiles) // highest resolution, not profiles[0]
    uri, err := client.Media().GetStreamURI(context.Background(), mainToken)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("main stream:", uri.URI)
}
```

Firmware rejects digest auth? Add a ladder:

```go
client, err := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAuthFallback(onvif.AuthPasswordText, onvif.AuthHTTPBasic, onvif.AuthNone),
)
```

Discovery — active, passive, or directed:

```go
devices, _ := discovery.Discover(ctx, 5*time.Second)          // multicast probe
usable := discovery.FilterONVIFDevices(devices)               // drop ghost responders
discovery.EnrichDevices(ctx, usable)                          // parallel identity fetch

listener, _ := discovery.NewListener("", func(d *discovery.Device) { // passive
    fmt.Println("camera came online:", d.GetDeviceEndpoint())
})
go listener.Start(ctx)
defer listener.Stop()

dev := discovery.ProbeEndpoint(ctx, "192.168.2.50", 80, 1200*time.Millisecond) // cross-subnet
serial, ok := discovery.ProbeSerial(ctx, "192.168.2.50", nil)                  // identity anchor
```

Events — managed, auto-renewing:

```go
sub, err := client.Events().SubscribeEvents(ctx, func(msg onvif.NotificationMessage) {
    fmt.Println(msg.Topic, msg.Message.Data)
}, nil)
if errors.Is(err, onvif.ErrEventsNotSupported) {
    // device has no events service; cache the negative result
}
defer sub.Unsubscribe(context.Background())
```

Events, push when the device supports it — probe, fall back, serve the
consumer endpoint:

```go
push, err := client.Events().SubscribePushEvents(ctx, "http://nvr:8080/onvif/notify", nil)
if errors.Is(err, onvif.ErrPushNotSupported) {
    // fall back to the managed pull-point subscription above
}
defer push.Unsubscribe(context.Background())

// The endpoint the device POSTs wsnt:Notify bodies to (wrap with your own
// origin/token validation — the push profile carries no credentials):
http.Handle("/onvif/notify", onvif.NotifyHandler(func(msgs []onvif.NotificationMessage) {
    for _, m := range msgs {
        fmt.Println(m.Topic, m.Message.Data)
    }
}))
```

Auth triage when nothing works:

```go
diag, _ := client.DiagnoseAuth(ctx)
// diag.Status: "ok" | "clock-skew" (fix camera NTP) | "bad-credentials"
```

## Project layout

| Path | Purpose |
|---|---|
| `client.go` etc. (root) | `Client` (auth ladder, clock skew, download) + v1 compatibility aliases |
| `device/` `security/` `deviceio/` `media/` `ptz/` `imaging/` `events/` | domain service packages (v2 split, issue #20) |
| `types/` | shared data-model leaf |
| `discovery/` | WS-Discovery: active probe, passive listener, directed HTTP probing, post-processing |
| `internal/soap/` | SOAP transport + WS-Security (digest/text modes, fault detection) |
| `server/` | Virtual camera simulator + embeddable server: DeviceIO (relays / digital inputs, opt-in), Media2 minimal face (opt-in), media OSD loop + audio empty sets |
| `docs/` | Redirect to the documentation hub |
| `testdata/captures/` | Real-camera SOAP captures used as regression fixtures |
| `cmd/` | The unified `onvif` CLI (subcommands `discover`, `quick`, `diagnostics`, `server`) + `generate-tests`, the capture→fixture developer tool |
| `examples/` | Runnable examples per feature area: discovery, device-info, imaging-settings, ptz-control, events (PullPoint subscription), media2 (H.265 options), analytics (Profile M, incl. rules create/modify/delete + options), metadata (stream parsing), simple-server, tls-server, onvif-server, complete-demo |

## Protocol alignment & roadmap

Wire contracts are grounded in the official ONVIF WSDL/XSD set
([onvif/specs](https://github.com/onvif/specs), all schemas
`elementFormDefault="qualified"`): elements locally declared in a service
WSDL resolve to that WSDL's namespace, children of `ver10/schema` types
to `tt`. Every served response surface is pinned by a namespace-strict
contract suite, a client↔simulator conformance loopback drives the full
operation matrix, and hostile-response fuzzers cover the client decode
paths.

Deliberate scope boundaries (tracked for future capability packages):

- **Media2 (ver20/media/wsdl)** — the codec-agnostic configuration model
  is in (`onvif.Client.Media2()`: profiles with inline configurations,
  video encoder configurations and per-encoding options — one entry per
  codec with free-name `Encoding` (H264/H265/AV1/…), set with verbatim
  encoding pass-through, stream URI, snapshot URI, encoder instance
  counts, multicast start/stop, `SetSynchronizationPoint`, service
  capabilities, audio encoder and metadata configuration families).
  `Initialize` discovers Media2-only devices: when GetCapabilities
  carries no ver10 Media XAddr, the ver20/media endpoint is probed via
  GetServices and wired for `Media2()` calls. Profile create/delete and
  the OSD family are not implemented yet; ver10 media remains the
  full-coverage surface.
- **Profile M** — the events PullPoint family, metadata configurations
  in profiles, and the analytics-service client (rule/module
  configuration CRUD, `onvif.Client.Analytics()`, v2.1.0) are in;
  message descriptions in analytics `ConfigDescription` and the
  metadata stream's PTZ/Event/SensorData members are not modeled yet.
- **Profile T** — TLS transport is supported on the server
  (`Config.TLSCertFile`/`TLSKeyFile`); advanced onboarding (PKI,
  802.1X, WS-Security beyond UsernameToken) is out of scope.
- **GetEventProperties** serves the spec-complete answer (topic
  namespace/dialect locations, the two mandatory topic-expression
  dialects — honored by the pull-point filter — and the spec-blessed
  empty message-content dialect); message-content filtering is not
  applied and not advertised.
- **WebRTC** — watched, not implemented.

## CLI tools

One CLI, four subcommands — `onvif discover`, `onvif quick`,
`onvif diagnostics`, `onvif server` — shipped as a single prebuilt,
zero-dependency binary per platform (linux amd64/arm64/arm, macOS
amd64/arm64, Windows amd64) on
[every release](https://github.com/mickeyzzc/onvif-go/releases):
download, `chmod +x`, verify against `SHA256SUMS`. Or
`go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.2.0`.

Which subcommand for which job: `discover` scans the LAN
(WS-Discovery), `quick` interactively sanity-checks one camera
(connect, stream URLs, a guarded PTZ demo), `diagnostics` sweeps all
eleven major operations into a JSON report — the tool to run before
[filing an issue](https://github.com/mickeyzzc/onvif-go/issues) — and
`server` stands up a virtual multi-lens camera for testing without
hardware. Per-subcommand manuals with scenarios live in the
documentation hub (below).

## Documentation

Topic guides now live in the MiBee documentation hub — the single
source of truth for library manuals, bilingual:

> **https://www.mlsbs.top/docs/mibeelibs**

Manual changes go there by PR (review flow in the hub repo's GOVERNANCE).
[`docs/README.md`](docs/README.md) keeps the redirect.
## Development

```bash
make build    # go build ./...
make test     # go test -race ./...
make lint     # golangci-lint run   (install: make lint-install)
make fmt      # gofumpt + goimports via golangci-lint fmt
```

CI ([ci.yml](.github/workflows/ci.yml)) runs lint + format check + race
tests + build on every push to `main`; the branch requires all three jobs
green.

## Library hygiene (rc4 hardening)

- `server.DefaultConfig()` no longer ships `admin`/`admin` credentials
  (empty by default; the no-credentials mode is the documented everything-
  open mode — set `Username`/`Password` explicitly before real networks).
- The public `testing` package moved to `internal/onviftesting` (no more
  `httptest` linked into consumer binaries).
- The exported mutable default slices became functions returning fresh
  slices: `server/soap.DefaultProtectedPrefixes()`, `server.DefaultScopes()`,
  `discovery.DefaultProbePorts()`.
- `testdata/` carries synthetic data only (RFC 5737 ranges); keep real
  captures in the gitignored `tmp/`.
- `hygiene_test.go` at the repo root pins all of the above. See
  [CHANGELOG.md](CHANGELOG.md) and [MIGRATION.md](MIGRATION.md) for
  migration details.

## Attribution & license

This project is a maintained continuation of
[0x524a/onvif-go](https://github.com/0x524a/onvif-go) (itself derived from
earlier ONVIF Go efforts — see git history). The module path changed to
`github.com/mickeyzzc/onvif-go` in **v1.2.0**; before that it was published
as `github.com/0x524a/onvif-go` (tags v1.0.0–v1.1.7 remain on this repo for
history and are never deleted). Continuation-focused changes since v1.2.0
are tracked in the [changelog](CHANGELOG.md).

Licensed under the [MIT License](LICENSE). Contributions from the original
authors are gratefully acknowledged and remain under the same license.
