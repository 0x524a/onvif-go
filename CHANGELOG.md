# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of go-onvif library
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

[Unreleased]: https://github.com/0x524A/go-onvif/compare/v0.1.0...HEAD
