# Concurrency Model

A `Client` is safe for concurrent use by multiple goroutines. This document
explains what backs that guarantee and how to use it.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## The contract

- **One `Client`, many goroutines, no external locking.** Share a single
  client across components (recording, snapshots, PTZ, events) — that is the
  intended usage, and it matters for weak devices where every extra HTTP
  connection counts.
- All mutable client state — credentials, clock skew, auth-ladder stickiness,
  the capabilities cache, service endpoints — is guarded by an internal
  `sync.RWMutex`.
- Every operation builds its own stateless SOAP exchange over the shared
  `*http.Client` (itself concurrency-safe); nothing per-request lives on the
  `Client` between calls.
- Configuration setters (`SetCredentials`, `SetClockSkew`,
  `InvalidateCapabilitiesCache`, `ResetAuthLadder`) may run concurrently with
  in-flight calls; each call applies a consistent snapshot taken at dispatch.

## What the audit fixed

The guarantee used to be *almost* true, with three gaps (issue #12):

1. **Endpoint races** — the Media/PTZ/Imaging service endpoint reads were
   unguarded while `Initialize` wrote them.
2. **Raw reads without fallback** — PTZ and Imaging read their endpoint
   fields directly, so calling them before `Initialize` sent requests to an
   empty URL. All endpoint access now goes through lock-guarded getters that
   fall back to the device endpoint.
3. **Unguarded writes** — `Initialize` now takes the write lock.

## How it is pinned down

`concurrency_test.go` runs a mixed-operation matrix on one shared client —
Device / Media / PTZ / Events operations concurrently with
`SetCredentials` / `SetClockSkew` / `InvalidateCapabilitiesCache` /
`Initialize` — and the CI test job runs it under `-race` on every push.

## Guidance

- **Do** share one `Client` per device. Per-call clients re-handshake and
  can overwhelm ESP32-class firmware.
- **Do** use `GetCapabilitiesCached` instead of `GetCapabilities` in hot
  paths — it single-flights concurrent first fetches (see the
  [architecture overview](onvif-go-architecture.md)).
- **Don't** share a `Client` between *different devices*; endpoint identity
  and auth conclusions are per device.
- Managed event subscriptions own their background goroutine and stop via
  `Unsubscribe` / context cancellation (see [events.md](onvif-go-events.md)).
