# Testing

The test suite runs entirely offline by default: every behavioral test works
against `httptest` mock devices or raw XML fixtures. Real-camera integration
tests exist but are gated behind environment variables and never run in CI.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## Running

```bash
make test        # go test -race ./...
make lint        # golangci-lint run (zero findings required)
make fmt         # gofumpt + goimports via golangci-lint fmt
make check       # lint + test
```

CI (`.github/workflows/ci.yml`) runs lint, a format check, race tests, and a
build on every push to `main`; all three jobs are required status checks.

A few discovery tests need the host to join the WS-Discovery multicast
group. On hosts without a usable multicast route (VPN-up laptops, some CI
runners) they skip automatically — the suite stays green.

## Test layers

| Layer | Where | What it covers |
|---|---|---|
| Mock-device tests | `*_test.go` next to the code | Full client behavior against `httptest` SOAP devices: auth ladder transitions, fault handling, response-shape variants, lifecycle of managed subscriptions |
| Raw fixtures | `testdata/captures/*.xml` | Hand-crafted real-shape envelopes replayed through the client (e.g. the GetStreamUri namespace variants behind issue #3) |
| Parser unit tests | e.g. `TestParseScopes`, `TestLooseExtractURI`, `TestSelectMainProfile` | Pure functions, no I/O |
| Conformance loopback | client↔simulator suites | The library's own client and server talk real HTTP across the full operation matrix (device, PTZ, imaging — including the GetScopes element round-trip) |
| Namespace contract suite | wire-namespace tests | Every served response surface marshaled and decoded namespace-strictly — values populate only when elements resolve to the exact WSDL-assigned namespace — plus an explicit-prefix end-to-end check |
| Fuzz targets | native `FuzzXxx` (device/media/events) | The real per-op response parsers driven against hostile bodies, seeded from the contract wire formats |
| Concurrency matrix | `concurrency_test.go` | Mixed operations + config mutation on one shared client, meaningful under `-race` |
| Capture-replay helpers | `internal/onviftesting/` package | Mock server, capture registry, golden files — used by the larger suites |

## Real-camera integration tests

Selected test files (`device_real_camera_test.go`,
`media_real_camera_test.go`) run against actual hardware when (and only
when) you provide connection details:

```bash
export ONVIF_TEST_ENDPOINT="http://192.168.x.x/onvif/device_service"
export ONVIF_TEST_USERNAME="..."
export ONVIF_TEST_PASSWORD="..."
go test -v -run RealCamera ./...
```

Use placeholder credentials in any documentation or issues you file — never
real ones. When a device misbehaves, capture its raw SOAP response and add
it as a fixture under `testdata/captures/` so the regression lives on after
the camera is gone. `cmd/onvif` (see [cli.md](onvif-go-cli.md)) exists
exactly for producing those captures.

## Test-driven development

The library is developed TDD-first: **write the failing test before the
code** — for bug fixes the test reproduces the defect first, for new
service operations the test pins the wire contract (request fields the
device branches on + parsed response fields) before the implementation.
This backfill pass found four real defects that green-looking code was
hiding (`imaging.Move` never encoded its focus parameters,
`imaging.GetOptions` silently dropping most option groups, dead
endpoint-fallback guards, a `Stop` that took a full read-deadline tick
to unblock).

Rules of thumb:

- Service operations are tested through `internal/testutil.FakeCaller`
  — the exact `xml.Unmarshal` decode path of the real transport, no
  sockets, microsecond latency. Reach for `httptest` only when HTTP
  behavior itself is under test.
- **Every channel wait is bounded**: `waitFor`/`mustReceive` helpers
  fail in 5s instead of hanging the suite — a deadlock must surface as
  a test failure, never as a CI timeout.
- Background loops (`events.EventStream`, `discovery.Listener`,
  `server/discovery.Responder`) must observe cancellation **immediately**,
  not on the next timer tick — Stop paths close the underlying socket
  to interrupt blocked reads; the tests assert prompt `Done()` closure.
- Time is injected, never slept on: lifecycle tests pass short
  durations through options; assertions on timestamps render fixtures
  relative to `time.Now()` (timezone-independent).
- Timing budgets: a table case costs microseconds; the whole
  package-level suite stays in single-digit seconds (the one 1s
  backoff-recovery test in `events` is the deliberate exception).
  CI runs the same suite under `-race` — registration maps and shared
  state must be lock-guarded (`server/soap.Handler` is the reference).

## Conventions for new tests

- Prefer `httptest` servers over hand-mocked transports; classify requests
  by body content, not by call order, so tests stay valid when the transport
  details change.
- Assert the *wire shape* for anything a device branches on (e.g. the
  StreamSetup fields of `GetStreamUri` requests).
- Time-sensitive loops (managed subscriptions, discovery listeners) should
  expose deterministic exit signals (`Done()` channels) rather than relying
  on sleeps.
