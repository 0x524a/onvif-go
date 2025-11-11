# go-onvif

[![Go Reference](https://pkg.go.dev/badge/github.com/0x524A/go-onvif.svg)](https://pkg.go.dev/github.com/0x524A/go-onvif)
[![Go Report Card](https://goreportcard.com/badge/github.com/0x524A/go-onvif)](https://goreportcard.com/report/github.com/0x524A/go-onvif)
[![License](https://img.shields.io/github/license/0x524A/go-onvif)](LICENSE)

A modern, performant, and easy-to-use Go library for communicating with ONVIF-compliant IP cameras and devices. Includes both **ONVIF client** and **ONVIF server** implementations.

## Features

### 📡 ONVIF Client

✨ **Modern Go Design**
- Context support for cancellation and timeouts
- Concurrent-safe operations
- Type-safe API with comprehensive error handling
- Connection pooling for optimal performance

🎥 **Comprehensive ONVIF Support**
- **Device Management**: Get device info, capabilities, system date/time, reboot
- **Media Services**: Profiles, stream URIs (RTSP/HTTP), snapshot URIs, encoder configuration
- **PTZ Control**: Continuous, absolute, and relative movement, presets, status
- **Imaging**: Get/set brightness, contrast, exposure, focus, white balance, WDR
- **Discovery**: Automatic camera detection via WS-Discovery multicast

### 🎬 ONVIF Server (NEW!)

🎥 **Virtual IP Camera Simulator**
- **Multi-Lens Camera Support**: Simulate up to 10 independent camera profiles
- **Complete ONVIF Implementation**: Device, Media, PTZ, and Imaging services
- **Flexible Configuration**: CLI and library interfaces for easy setup
- **PTZ Simulation**: Full pan-tilt-zoom control with preset positions
- **Imaging Control**: Brightness, contrast, exposure, focus, and more
- **Testing & Development**: Perfect for testing ONVIF clients without physical cameras

🔐 **Security**
- WS-Security with UsernameToken authentication
- Password digest (SHA-1) support
- Configurable timeout and HTTP client options

📦 **Easy Integration**
- Simple, intuitive API
- Well-documented with examples
- No external dependencies beyond Go standard library and golang.org/x/net

## Installation

```bash
go get github.com/0x524A/go-onvif
```

## Quick Start

### Discover Cameras on Network

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/0x524A/go-onvif/discovery"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    devices, err := discovery.Discover(ctx, 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }

    for _, device := range devices {
        fmt.Printf("Found: %s at %s\n", 
            device.GetName(), 
            device.GetDeviceEndpoint())
    }
}
```

### Connect to a Camera

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/0x524A/go-onvif"
)

func main() {
    // Create client
    client, err := onvif.NewClient(
        "http://192.168.1.100/onvif/device_service",
        onvif.WithCredentials("admin", "password"),
        onvif.WithTimeout(30*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get device information
    info, err := client.GetDeviceInformation(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Camera: %s %s\n", info.Manufacturer, info.Model)
    fmt.Printf("Firmware: %s\n", info.FirmwareVersion)

    // Initialize and discover service endpoints
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }

    // Get media profiles
    profiles, err := client.GetProfiles(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Get stream URI
    if len(profiles) > 0 {
        streamURI, err := client.GetStreamURI(ctx, profiles[0].Token)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Stream URI: %s\n", streamURI.URI)
    }
}
```

### PTZ Control

```go
// Continuous movement
velocity := &onvif.PTZSpeed{
    PanTilt: &onvif.Vector2D{X: 0.5, Y: 0.0}, // Move right
}
timeout := "PT2S" // 2 seconds
err := client.ContinuousMove(ctx, profileToken, velocity, &timeout)

// Stop movement
err = client.Stop(ctx, profileToken, true, true)

// Absolute positioning
position := &onvif.PTZVector{
    PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.0}, // Center
    Zoom:    &onvif.Vector1D{X: 0.5},         // 50% zoom
}
err = client.AbsoluteMove(ctx, profileToken, position, nil)

// Go to preset
presets, err := client.GetPresets(ctx, profileToken)
if len(presets) > 0 {
    err = client.GotoPreset(ctx, profileToken, presets[0].Token, nil)
}
```

### Imaging Settings

```go
// Get current settings
settings, err := client.GetImagingSettings(ctx, videoSourceToken)

// Modify settings
brightness := 60.0
settings.Brightness = &brightness

contrast := 55.0
settings.Contrast = &contrast

// Apply settings
err = client.SetImagingSettings(ctx, videoSourceToken, settings, true)
```

## API Overview

### Client Creation

```go
client, err := onvif.NewClient(
    endpoint,
    onvif.WithCredentials(username, password),
    onvif.WithTimeout(30*time.Second),
    onvif.WithHTTPClient(customHTTPClient),
)
```

### Device Service

| Method | Description |
|--------|-------------|
| `GetDeviceInformation()` | Get manufacturer, model, firmware version |
| `GetCapabilities()` | Get device capabilities and service endpoints |
| `GetSystemDateAndTime()` | Get device system time |
| `SystemReboot()` | Reboot the device |
| `Initialize()` | Discover and cache service endpoints |
| `GetHostname()` | Get device hostname configuration |
| `SetHostname()` | Set device hostname |
| `GetDNS()` | Get DNS configuration |
| `GetNTP()` | Get NTP configuration |
| `GetNetworkInterfaces()` | Get network interface configuration |
| `GetScopes()` | Get configured discovery scopes |
| `GetUsers()` | Get list of user accounts |
| `CreateUsers()` | Create new user accounts |
| `DeleteUsers()` | Delete user accounts |
| `SetUser()` | Modify existing user account |

### Media Service

| Method | Description |
|--------|-------------|
| `GetProfiles()` | Get all media profiles |
| `GetStreamURI()` | Get RTSP/HTTP stream URI |
| `GetSnapshotURI()` | Get snapshot image URI |
| `GetVideoEncoderConfiguration()` | Get video encoder settings |
| `GetVideoSources()` | Get all video sources |
| `GetAudioSources()` | Get all audio sources |
| `GetAudioOutputs()` | Get all audio outputs |
| `CreateProfile()` | Create new media profile |
| `DeleteProfile()` | Delete media profile |
| `SetVideoEncoderConfiguration()` | Set video encoder configuration |

### PTZ Service

| Method | Description |
|--------|-------------|
| `ContinuousMove()` | Start continuous PTZ movement |
| `AbsoluteMove()` | Move to absolute position |
| `RelativeMove()` | Move relative to current position |
| `Stop()` | Stop PTZ movement |
| `GetStatus()` | Get current PTZ status and position |
| `GetPresets()` | Get list of PTZ presets |
| `GotoPreset()` | Move to a preset position |
| `SetPreset()` | Save current position as preset |
| `RemovePreset()` | Delete a preset |
| `GotoHomePosition()` | Move to home position |
| `SetHomePosition()` | Set current position as home |
| `GetConfiguration()` | Get PTZ configuration |
| `GetConfigurations()` | Get all PTZ configurations |

### Imaging Service

| Method | Description |
|--------|-------------|
| `GetImagingSettings()` | Get imaging settings (brightness, contrast, etc.) |
| `SetImagingSettings()` | Set imaging settings |
| `Move()` | Perform focus move operations |
| `GetOptions()` | Get available imaging options and ranges |
| `GetMoveOptions()` | Get available focus move options |
| `StopFocus()` | Stop focus movement |
| `GetImagingStatus()` | Get current imaging/focus status |

### Discovery Service

| Method | Description |
|--------|-------------|
| `Discover()` | Discover ONVIF devices on network |

## ONVIF Server

The library now includes a complete ONVIF server implementation that simulates multi-lens IP cameras!

### Quick Start

```bash
# Install the server CLI
go install ./cmd/onvif-server

# Run with default settings (3 camera profiles)
onvif-server

# Or customize
onvif-server -profiles 5 -username admin -password mypass -port 9000
```

### Using the Server Library

```go
package main

import (
    "context"
    "log"

    "github.com/0x524A/go-onvif/server"
)

func main() {
    // Create server with default multi-lens camera configuration
    srv, err := server.New(server.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }

    // Start server
    ctx := context.Background()
    if err := srv.Start(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Server Features

- 🎥 **Multi-Lens Simulation**: Support for up to 10 independent camera profiles
- 🎮 **Full PTZ Control**: Pan, tilt, zoom with preset positions
- 📷 **Imaging Settings**: Brightness, contrast, exposure, focus, white balance
- 🌐 **Complete ONVIF Services**: Device, Media, PTZ, and Imaging services
- 🔐 **WS-Security**: Digest authentication support
- ⚙️ **Flexible Configuration**: CLI and library interfaces

### Use Cases

- Testing ONVIF client implementations
- Developing video management systems
- CI/CD integration testing
- Demonstrations without physical cameras
- Learning ONVIF protocol

For complete documentation, see [server/README.md](server/README.md).

## Examples

The [examples](examples/) directory contains complete working examples:

### Client Examples
- **[discovery](examples/discovery/)**: Discover cameras on the network
- **[device-info](examples/device-info/)**: Get device information and media profiles
- **[ptz-control](examples/ptz-control/)**: Control camera PTZ (pan, tilt, zoom)
- **[imaging-settings](examples/imaging-settings/)**: Adjust imaging settings

### Server Examples
- **[onvif-server](examples/onvif-server/)**: Multi-lens camera server with custom configuration

To run an example:

```bash
cd examples/discovery
go run main.go
```

## Architecture

```
go-onvif/
├── client.go           # Main ONVIF client
├── types.go            # ONVIF data types
├── errors.go           # Error definitions
├── device.go           # Device service implementation
├── media.go            # Media service implementation
├── ptz.go              # PTZ service implementation
├── imaging.go          # Imaging service implementation
├── soap/               # SOAP client with WS-Security
│   └── soap.go
├── discovery/          # WS-Discovery implementation
│   └── discovery.go
├── server/             # ONVIF server implementation
│   ├── server.go       # Main server
│   ├── types.go        # Server types and configuration
│   ├── device.go       # Device service handlers
│   ├── media.go        # Media service handlers
│   ├── ptz.go          # PTZ service handlers
│   ├── imaging.go      # Imaging service handlers
│   └── soap/           # SOAP server handler
│       └── handler.go
├── cmd/
│   ├── onvif-cli/      # Client CLI tool
│   └── onvif-server/   # Server CLI tool
└── examples/           # Usage examples
    ├── discovery/
    ├── device-info/
    ├── ptz-control/
    ├── imaging-settings/
    └── onvif-server/   # Multi-lens camera server example
```

## Design Principles

1. **Context-Aware**: All network operations accept `context.Context` for cancellation and timeouts
2. **Type Safety**: Strong typing with comprehensive struct definitions
3. **Error Handling**: Typed errors with clear error messages
4. **Concurrency Safe**: Thread-safe operations with proper locking
5. **Performance**: Connection pooling and efficient HTTP client reuse
6. **Standards Compliant**: Follows ONVIF specifications for SOAP/XML messaging

## Compatibility

- **Go Version**: 1.21+
- **ONVIF Versions**: Compatible with ONVIF Profile S, Profile T, Profile G
- **Tested Cameras**: Works with most ONVIF-compliant IP cameras including:
  - Axis
  - Hikvision
  - Dahua
  - Bosch
  - Hanwha (Samsung)
  - And many others

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [ ] Event service implementation
- [ ] Analytics service implementation
- [ ] Recording service implementation
- [ ] Replay service implementation
- [ ] Advanced security features (TLS, X.509 certificates)
- [ ] Comprehensive test suite with mock cameras
- [ ] Performance benchmarks
- [ ] CLI tool for camera management

## Debugging Tools

### 🔍 Diagnostic Utility

Comprehensive camera testing and analysis with optional XML capture:

```bash
go build -o onvif-diagnostics ./cmd/onvif-diagnostics/

# Standard diagnostic report
./onvif-diagnostics \
  -endpoint "http://camera-ip/onvif/device_service" \
  -username "admin" \
  -password "pass" \
  -verbose

# With raw SOAP XML capture for debugging
./onvif-diagnostics \
  -endpoint "http://camera-ip/onvif/device_service" \
  -username "admin" \
  -password "pass" \
  -capture-xml \
  -verbose
```

**Generates**:
- `camera-logs/Manufacturer_Model_Firmware_timestamp.json` - Diagnostic report
- `camera-logs/Manufacturer_Model_Firmware_xmlcapture_timestamp.tar.gz` - Raw XML (with `-capture-xml`)

**See**: `XML_DEBUGGING_SOLUTION.md` for complete debugging workflow

### 🧪 Camera Test Framework

Automated regression testing using captured camera responses:

```bash
# 1. Capture from camera
./onvif-diagnostics -endpoint "http://camera/onvif/device_service" \
  -username "user" -password "pass" -capture-xml

# 2. Generate test
go build -o generate-tests ./cmd/generate-tests/
./generate-tests -capture camera-logs/*_xmlcapture_*.tar.gz -output testdata/captures/

# 3. Run tests
go test -v ./testdata/captures/
```

**Benefits**:
- Test without physical cameras
- Prevent regressions across camera models
- Fast CI/CD integration
- Real camera response validation

**See**: `testdata/captures/README.md` for complete testing guide

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the original [use-go/onvif](https://github.com/use-go/onvif) library
- ONVIF specifications from [ONVIF.org](https://www.onvif.org)
- Thanks to all contributors and the Go community

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/0x524A/go-onvif)
- 🐛 [Issue Tracker](https://github.com/0x524A/go-onvif/issues)
- 💬 [Discussions](https://github.com/0x524A/go-onvif/discussions)

## Related Projects

- [ONVIF Device Manager](https://sourceforge.net/projects/onvifdm/) - GUI tool for testing ONVIF devices
- [ONVIF Device Tool](https://www.onvif.org/tools/) - Official ONVIF test tool

---

Made with ❤️ for the Go and IoT community