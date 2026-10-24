# v2 Architecture (Breaking)

> Status: implemented — M1–M4 landed (#21 #22 #23 #24), finalization in [#25](https://github.com/mickeyzzc/onvif-go/issues/25), tracked by the [v2 epic](https://github.com/mickeyzzc/onvif-go/issues/20). This document is the plan of record.

v2 is a deliberately breaking release. Two forces drive it:

1. **Client-side scale**: ~24k lines behind a facade API that structurally required a single package, leaving the root directory flat.
2. **Server-side ambitions**: issues #15–#19 need `server/` to graduate from a test simulator into an embeddable device-side framework for real cameras (MiBee Eye / rpi3b-cam).

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## Module versioning

Go modules require the `/v2` suffix for v2+: the module path becomes
`github.com/mickeyzzc/onvif-go/v2`. The client package lives in the `onvif/`
subdirectory (import `github.com/mickeyzzc/onvif-go/v2/onvif` — the package
identifier is unchanged), keeping the repository root free of Go source. The
v1 tags stay untouched for existing
consumers. Breaking within 1.x would forfeit semver credibility for a public
library — not an option.

## The knot v2 cuts: the Caller interface

v1 could not be split into packages because of an import cycle: the facade
accessors (`Client.Media()`) need the service types; the service methods need
`Client`. v2 applies the standard aws-sdk-go-v2 shape:

```go
// internal/api — tiny leaf package
type Caller interface {
    Call(ctx context.Context, endpoint, action string, request, response any) error
    EndpointFor(svc Service) string // resolved endpoint with device-endpoint fallback
}
```

- `Client` (the `onvif` package) implements `Caller` — the auth ladder, clock
  skew, and capability caching stay there.
- Each service package depends only on `internal/api`.
- The `onvif` package imports the service packages and provides the accessors —
  one-directional, cycle-free.
- Services are **long-lived instances** created once by the client; accessors
  return the same pointer, so a service may hold its own state (the
  capabilities cache lives in `device.Service`).

## Target layout

```text
github.com/mickeyzzc/onvif-go/v2
├── onvif/               the client package (package identifier stays `onvif`)
│   ├── client.go        NewClient + Client (implements api.Caller) + options
│   ├── auth.go          auth modes / ladder / clock skew / DiagnoseAuth
│   ├── errors.go        cross-domain sentinels (aliased from leaves)
│   └── aliases.go       v1-compatibility re-exports
├── types/               shared data-model leaf (IPAddress, IntRectangle, ...)
├── device/              tds: identity, capabilities, system, network, DNS/NTP,
│                        storage, WiFi — plus the capabilities cache
├── security/            users, remote user, access policy, certificates
├── media/               trt: profiles, main/sub selection, StreamSetup,
│                        encoder/audio/OSD configuration
├── ptz/   imaging/   events/ (incl. managed subscriptions)   deviceio/
├── media2/              tr2: the codec-agnostic Media2 configuration model —
│                        per-encoding options with free-name Encoding
│                        (the H.265/AV1 answer), verbatim set (v2.1.0)
├── analytics/           tan: Profile M rule / analytics-module configuration
│                        client (v2.1.0)
├── metadata/            tt:MetadataStream parser — the Profile M analytics
│                        output channel (frames, objects, geometry) (v2.1.0)
├── discovery/           client-side discovery (probe / listener / directed / post-processing)
├── wsdiscovery/         WS-Discovery message codec leaf — shared by the client
│                        side and the device-side responder (#15)
├── server/              device-side framework (first-class)
│   ├── server.go        lifecycle, HTTP listener, optional discovery responder
│   ├── provider.go      pluggable state backends (#19): DeviceInfo / StreamURI /
│   │                    Snapshot JPEG / Imaging (range+enum validated) / PTZ
│   ├── simulator/       the current simulator state, as the default provider
│   ├── services/        stateless handlers translating SOAP ↔ provider calls
│   ├── soap/            device-side transport: dispatch, request context (#17),
│   │                    per-action auth policy + PasswordText (#16),
│   │                    XML response writer: canonical casing, explicit
│   │                    prefixes, raw-bytes channel (#18)
│   └── discovery/       device-side WS-Discovery responder (#15)
├── internal/{api,soap,httpdigest}
└── cmd/  examples/  testing/  testdata/captures/  docs/{en,zh}
```

## Server framework principles

- **Stateless services**: `server/services/*` hold no state; they translate
  SOAP operations into provider calls. All state decisions move behind
  interfaces.
- **Simulator as default**: the existing simulator behavior becomes
  `server/simulator`, selected when no provider is injected — `cmd/onvif`
  and the examples keep working unchanged.
- **Transport owns the cross-cutting concerns**: request context (#17),
  per-action authentication policy (#16), and byte-level XML control (#18) all
  live in `server/soap`, implemented once for every service.
- **Byte-predictable XML**: the response writer applies WSDL-canonical element
  casing, optional explicit namespace prefixes, and offers a raw-bytes channel
  for hosts that pre-build envelopes — locked by golden-fixture tests.

## Compatibility strategy

Breaking is allowed, but wasteful breaking is not:

- Root-package **type aliases** (`type Profile = media.Profile`) keep most
  consumer code compiling after the import-path swap.
- Facade call shapes stay: `client.Media().GetProfiles(ctx)`.
- What moves gets documented in `MIGRATION.md` with a complete mapping table
  (e.g. `onvif.SelectMainProfile` → `media.SelectMain`).
- Domain errors move with their packages (`media.ErrEmptyMediaURI`,
  `events.ErrEventsNotSupported`); cross-domain sentinels stay aliased in root.

## Milestones

| Stage | Issue | Scope |
|---|---|---|
| M1 | #21 | ✅ done — client package split, `/v2` path, aliases, zero behavior change |
| M2 | #22 | ✅ done — `server/soap` transport: context, auth policy, XML writer (#17 #16 #18) |
| M3 | #23 | ✅ provider interfaces + simulator extraction (#19) |
| M4 | #24 | ✅ device-side discovery responder + shared codec (#15) |
| M5 | #25 | ✅ docs, migration guide, CHANGELOG, `v2.0.0-rc1`, MiBeeNvr reference migration |

## Process rules

- Every change lands through a PR (main is locked: required checks, PR-only,
  enforced for admins, linear history).
- Each PR updates `docs/{en,zh}` and the CHANGELOG in the same change.
