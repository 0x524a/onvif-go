# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Simplified Endpoint API**: `NewClient()` now accepts multiple endpoint formats
  - Simple IP address: `"192.168.1.100"`
  - IP with port: `"192.168.1.100:8080"`
  - Full URL: `"http://192.168.1.100/onvif/device_service"` (backward compatible)
  - Automatically adds `http://` scheme and `/onvif/device_service` path when needed
  - See `docs/SIMPLIFIED_ENDPOINT.md` for details
- Comprehensive test coverage for endpoint normalization (12 test cases)
- New example: `examples/simplified-endpoint/` demonstrating all endpoint formats
- Documentation: `docs/PROJECT_STRUCTURE.md` explaining project organization
- Initial release of go-onvif library

### Changed
- **Project Structure**: Implemented ideal Go project layout
  - Moved `soap/` to `internal/soap/` (private implementation)
  - Public API remains at root level for clean imports
  - Follows Standard Go Project Layout for libraries
  - Updated all imports throughout codebase
  - See `docs/PROJECT_STRUCTURE.md` and `docs/ARCHITECTURE.md` for details
- Updated `docs/ARCHITECTURE.md` to reflect new project structure
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

[Unreleased]: https://github.com/0x524A/onvif-go/compare/v0.1.0...HEAD
