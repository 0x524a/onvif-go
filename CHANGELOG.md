# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.6] - 2026-09-13

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

[Unreleased]: https://github.com/0x524a/onvif-go/compare/v1.1.3...HEAD
[1.1.3]: https://github.com/0x524a/onvif-go/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/0x524a/onvif-go/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/0x524a/onvif-go/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/0x524a/onvif-go/compare/v1.0.3...v1.1.0
