# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Capture replay (`testing`): the mock SOAP servers now rewrite the captured camera's own address in replayed responses to the mock's address. Before, a client followed the `GetCapabilities`/`GetServices` URLs out of the mock and tried to reach the original camera, so replay tests timed out.
- `generate-tests`: the generated test referenced its archive as `captures/<file>` instead of the bare file name, could emit an invalid test name such as `Testunknown_device`, and wrote unformatted code. Output is now gofmt-clean and marked as generated.
- The camera replay tests under `testdata/captures/` were never run, because Go ignores `testdata/`. They now live in `testing/captures/` and run with `go test ./...`; tests for the January captures were added (the broken `real_camera` test moved to the repository root and compiles).

### Changed
- Repository layout: debug and test programs moved from `examples/` to `tools/`, shell scripts to `scripts/`, 19 redundant status documents removed from `docs/`, and a committed 8.7 MB example binary removed.
- Repository root tidied: `media.go` (3,900 lines) split into `media.go`, `media_video.go`, `media_audio.go`, `media_metadata.go` and `media_osd.go` with no API change; coverage and configuration tests merged into `imaging_test.go`, `media_test.go` and `ptz_test.go`; `BUILDING.md` moved to `docs/`, the `Dockerfile` to `build/`, `.codecov.yml` to `.github/codecov.yml`; the duplicate root `CONTRIBUTING.md` removed in favour of `.github/CONTRIBUTING.md`; `test-reports/` and the January discovery data from `testdata/` moved to `testing/reports/`; the unused `testdata` Go package removed.

## [1.2.0] - 2026-10-10

### Added
- Virtual server: `GetEndpointReference` handler and `Config.EndpointUUID` (bare or `urn:uuid:`, validated in `New`), reported in the `urn:uuid:` form real cameras return. Left empty, a UUID is generated once per server. `Server.EndpointReference()` exposes it (#116, #118).
- Stable device identity across restarts: `Config.EndpointSeed` derives the endpoint UUID from a stable name, and `discovery.StableEndpointRef(seed)` does the same for any device, so a restarted service keeps its IDs without persisting them (#119).
- WS-Discovery responder: `discovery.Responder` answers `Probe` with `ProbeMatches` for any number of registered devices behind one socket, and sends `Hello` on start and `Bye` on stop. Types and Scopes filtering supports MatchBy `rfc3986` (default), `strcmp0` and `none`; Resolve, MatchBy `ldap`/`uuid`, the WS-Discovery 1.1 namespace and discovery proxies are not supported. Opt-in, with no change for existing users (#117, #119).
- Unicast discovery for tests and CI: `ResponderConfig.ListenAddr` and `DiscoverOptions.ProbeAddress` need no multicast routing, so the module's own client can find a server-backed device on loopback (#119).
- `Server.DiscoveryDevice()` builds the responder entry for a running server: endpoint reference, device-service XAddr and standard ONVIF scopes (#119).

### Changed
- README: Release and CI badges (#115). Doc comments on capabilities extension types (#114).
- Sonar `go:S5332` findings on ONVIF namespace identifiers and the scheme-less default URL are marked `NOSONAR` with the reason; no behaviour change (#120, #121).

## [1.1.8] - 2026-10-04

### Fixed
- `GetOptions` (imaging) parsed `BacklightCompensation`, `Exposure` and `Focus` and then discarded them, and never parsed `Sharpness`, `WideDynamicRange`, `WhiteBalance` or `IrCutFilterModes`, so callers could not learn the valid exposure or focus ranges before `SetImagingSettings`. All option blocks are now returned, including the Exposure gain/iris/time ranges and Focus near/far limits. A range the camera omits is now `nil` instead of a zero-valued `{0,0}` (#90, #112).
- A SOAP fault carried on a non-200 response (HTTP 500 is the SOAP 1.2 norm) was reported as a generic `HTTP request failed with status N: <raw XML>` error, so `errors.Is(err, ErrSOAPFault)` was false. It is now wrapped together with `ErrHTTPRequestFailed`, so both match and the camera's reason text is in the message. Faults on HTTP 200 were already handled (#94, #111).
- Release workflow: upgraded `softprops/action-gh-release` to v3.0.3 (asset uploads failed under GitHub's Node 24 runtime), pass `tag_name` explicitly so `workflow_dispatch` runs work, and fixed the Dockerfile builder image, which was older than the Go version `go.mod` requires.

### Changed
- Dependencies: `github.com/0x524A/rtspeek` 0.0.1 -> 0.2.0, which moves the indirect RTSP dependency from `gortsplib/v4` to `gortsplib/v5`. Docker builder image `golang:1.26-alpine` -> `1.27-alpine`. GitHub Actions bumped across all workflows (#106, #105, #110).
- Added Dependabot version updates for Go modules, GitHub Actions and Docker, and corrected version comments on five SHA-pinned actions that named the wrong release (#104, #108).

## [1.1.7] - 2026-09-13

### Added
- `GetSystemDateAndTimeTyped` for inspecting typed clock, timezone, and optional timestamp data without breaking existing callers (#61).
- Minimal pull-point event service in the virtual server (`CreatePullPointSubscription`/`PullMessages`/`RenewSubscription`/`Unsubscribe`), so `Config.SupportEvents` no longer advertises a 404 endpoint (#62).
- Ephemeral-port binding, `Addr()`, and a `Config.Ready` channel for the virtual server, so it can be used as an in-process test fixture without hardcoding a port (#63).
- Non-interactive mode for `onvif-cli` (`-op`/`-endpoint`/`-username`/`-password`/`-interface`/`-timeout`), for scripts and CI (#97).

### Fixed
- `soap.Body.Content` (`interface{}`) was never populated by `encoding/xml` on unmarshal, so every parameterized client request and every server-side request with body parameters silently received a nil body regardless of what was sent (#80, #95).
- Data races: `Client`'s endpoint/credential fields were written without locking (including a PTZ endpoint lock that was dropped after landing), and `Server.ServerInfo()` read the streams map without holding its mutex (#60, #95).
- PTZ/imaging mutexes in the virtual server were package-global, serializing every `Server` instance in a process through one lock each; PTZ move handlers also leaked an untracked, uncancellable goroutine per call (#80).
- `GetStreamUri`/`GetSnapshotUri` action-name casing on the virtual server didn't match the real ONVIF Media WSDL spelling the client sends, making both operations unreachable (#79).
- Four public config duration fields (e.g. `SessionTimeout`, `DefaultPTZTimeout`) and nine `PTZConfiguration` fields (default position/velocity spaces, pan/tilt/zoom limits) were parsed by nothing and silently dropped on every read and write (#86, #87).
- `PTZStatus.UtcTime` was never mapped onto the status returned by `GetStatus` (#89).
- `Config.Output` defaulted to `io.Discard` instead of `os.Stdout`, silencing the virtual server's startup banner for existing callers.
- WS-Security authentication hardening: constant-time digest comparison, a clock-skew check on `Created`, a per-handler nonce replay cache, and auth is no longer silently skipped when only one of username/password is configured (#80).
- `onvif-cli -op ...` and every other flag were never parsed at all, so the tool always fell through to the interactive menu and hung on stdin regardless of arguments (#97).
- Storage configuration XML tags/types, `PullMessages` XML parsing, and several SonarCloud/lint false positives.
- Release workflow: publish the GitHub Release as a draft and only mark it public after all binary assets finish uploading. The prior `draft: false` regressed a fix already noted below (v1.1.3) and broke the `v1.1.5`/`v1.1.6` releases with "Cannot upload assets to an immutable release".

### Changed
- Unified `ErrHTTPRequestFailed`/`ErrEmptyResponseBody`/`ErrInvalidResponse` and five duplicated sentinels between `server/errors.go`, the root package, and `internal/soap`, so `errors.Is` matches across package boundaries instead of comparing distinct `errors.New` values with the same text. Removed the leaked test-only `ErrRegularError` from the public API (#95).
- PTZ/Imaging now return `ErrNotInitialized` vs `ErrServiceNotSupported` distinctly depending on whether `Initialize` has run. **Behavior change:** `GetOptions`/`GetMoveOptions`/`StopFocus`/`GetImagingStatus` no longer fall back to the device endpoint before `Initialize` has been called (#64).
- Discovery's UUID generation now uses `google/uuid` (RFC 4122) instead of a wall-clock-derived string; `google/uuid` is now a direct dependency.
- README accuracy fixes: dependency list, architecture diagram, roadmap status, and CLI documentation (#98).

### Deprecated
- `GetSystemDateAndTime` and `FixedGetSystemDateAndTime` in favor of `GetSystemDateAndTimeTyped`.

### Notes
- **Behavior change:** When a device response omits the `SystemDateAndTime` element entirely, `GetSystemDateAndTimeTyped` now returns an error. Previously, `FixedGetSystemDateAndTime` returned a zero-value struct with no error in this case.
- `v1.1.6` was never published: GitHub rejected re-creation of that tag after its broken release was deleted, for reasons not exposed via any rules/tag-protection API. This release covers the same changes under `v1.1.7`.

## [1.1.3] - 2025-11-18

### Changed
- **Release Workflow**: Create releases as draft initially
  - Fixes "Cannot upload assets to an immutable release" error
  - Releases must be manually published after assets upload
  - Prevents race condition where release publishes before all assets finish uploading

## [1.1.2] - 2025-11-18

### Changed
- **Release Workflow**: Upgraded to `softprops/action-gh-release@v2`
  - Fixes asset upload race condition in v1
  - Better handling of concurrent file uploads
  - Added `fail_on_unmatched_files` and `make_latest` flags

## [1.1.1] - 2025-11-18

### Added
- **RTSPeek Library Integration**: RTSP stream inspection using `github.com/0x524A/rtspeek`
  - Replaced command-line `ffprobe` execution with library-based approach
  - Enhanced stream inspection with codec, resolution, and framerate detection
  - 5-second timeout for stream DESCRIBE operations
  - TCP fallback for basic connectivity checks
  - See `cmd/onvif-cli/main.go` for implementation

### Changed
- **Code Quality Improvements**: Fixed all linting errors
  - Removed unused `generateDemoASCII()` function
  - Fixed dynamic format strings (SA1006 errors)
  - Added proper error handling for Close() operations
  - Migrated to golangci-lint v2 configuration
  - CI/CD pipeline excludes utility tools and examples from linting
- **golangci-lint v2**: Updated configuration and GitHub Actions workflow
  - Created `.golangci.yml` with v2 schema
  - Updated CI to use golangci-lint-action@v8 with v2.2
  - Scoped linting to main packages only

## [1.1.0] - 2025-11-18

### Added
- **Simplified Endpoint API**: `NewClient()` now accepts multiple endpoint formats
  - Simple IP address: `"192.168.1.100"`
  - IP with port: `"192.168.1.100:8080"`
  - Full URL: `"http://192.168.1.100/onvif/device_service"` (backward compatible)
  - Automatically adds `http://` scheme and `/onvif/device_service` path when needed
  - See `docs/SIMPLIFIED_ENDPOINT.md` for details
- **Localhost URL Fix**: Automatic handling of cameras that report localhost addresses
  - Detects and fixes localhost/127.0.0.1/0.0.0.0/::1 in GetCapabilities response
  - Replaces with actual camera IP address
  - Preserves service-specific ports when specified
  - Handles common camera firmware bugs transparently
- Comprehensive test coverage for endpoint normalization (12 test cases)
- Comprehensive test coverage for localhost URL handling (10 test cases)
- New example: `examples/simplified-endpoint/` demonstrating all endpoint formats
- Documentation: `docs/PROJECT_STRUCTURE.md` explaining project organization
- Initial release of onvif-go library

### Changed
- **Project Structure**: Implemented ideal Go project layout
  - Moved `soap/` to `internal/soap/` (private implementation)
  - Moved `test/test-server.go` to `examples/test-server/` for clarity
  - Removed empty `test/` directory
  - Public API remains at root level for clean imports
  - Follows Standard Go Project Layout for libraries
  - Updated all imports throughout codebase
  - See `docs/PROJECT_STRUCTURE.md` and `docs/ARCHITECTURE.md` for details
- Updated `docs/ARCHITECTURE.md` to reflect new project structure
- Updated module path from `github.com/0x524A/onvif-go` to `github.com/0x524a/onvif-go` (lowercase)
- ONVIF Client with context support
- Device service implementation
  - GetDeviceInformation
  - GetCapabilities
  - GetSystemDateAndTime
  - SystemReboot
- Media service implementation
  - GetProfiles
  - GetStreamURI (RTSP/HTTP)
  - GetSnapshotURI
  - GetVideoEncoderConfiguration
- PTZ service implementation
  - ContinuousMove
  - AbsoluteMove
  - RelativeMove
  - Stop
  - GetStatus
  - GetPresets
  - GotoPreset
- Imaging service implementation
  - GetImagingSettings
  - SetImagingSettings
  - Move (focus control)
- WS-Discovery implementation
  - Automatic device discovery via multicast
- SOAP client with WS-Security
  - UsernameToken authentication
  - Password digest (SHA-1)
- Comprehensive type definitions
- Error handling with typed errors
- Connection pooling for performance
- Complete examples
  - Discovery
  - Device information
  - PTZ control
  - Imaging settings
- Comprehensive documentation
- README with usage guide

[Unreleased]: https://github.com/0x524a/onvif-go/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/0x524a/onvif-go/compare/v1.1.8...v1.2.0
[1.1.8]: https://github.com/0x524a/onvif-go/compare/v1.1.7...v1.1.8
[1.1.7]: https://github.com/0x524a/onvif-go/compare/v1.1.5...v1.1.7
[1.1.3]: https://github.com/0x524a/onvif-go/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/0x524a/onvif-go/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/0x524a/onvif-go/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/0x524a/onvif-go/compare/v1.0.3...v1.1.0
