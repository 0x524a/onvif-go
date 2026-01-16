# onvif-go - ONVIF Client and Server Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/0x524a/onvif-go.svg)](https://pkg.go.dev/github.com/0x524a/onvif-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/0x524a/onvif-go)](https://goreportcard.com/report/github.com/0x524a/onvif-go)
[![codecov](https://codecov.io/gh/0x524a/onvif-go/branch/master/graph/badge.svg)](https://codecov.io/gh/0x524a/onvif-go)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=0x524a_onvif-go&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=0x524a_onvif-go)
[![License](https://img.shields.io/github/license/0x524a/onvif-go)](LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/0x524a/onvif-go)](https://github.com/0x524a/onvif-go/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/0x524a/onvif-go)](https://github.com/0x524a/onvif-go/issues)

> **Modern, high-performance Go library for ONVIF IP camera integration** - Control surveillance cameras, NVRs, and video devices with comprehensive ONVIF Profile S/T/G support. Includes both client and server implementations for complete ONVIF camera simulation and testing.

A production-ready, feature-rich Go (Golang) library for communicating with ONVIF-compliant IP cameras, network video recorders (NVR), and surveillance devices. Perfect for building video management systems (VMS), security camera applications, IoT projects, and camera testing frameworks.

## 🎯 Key Features at a Glance

- ✅ **ONVIF Client & Server** - Both client library and virtual camera server
- ✅ **Production Ready** - Battle-tested with multiple camera brands
- ✅ **Full Protocol Support** - Device, Media, PTZ, Imaging, Discovery services
- ✅ **Type Safe** - Comprehensive Go types for all ONVIF operations
- ✅ **Well Documented** - Extensive examples and API documentation
- ✅ **Camera Tested** - Verified with Hikvision, Axis, Dahua, Bosch cameras
- ✅ **Testing Framework** - Built-in mock server and testing utilities

## 🔑 What is ONVIF?

ONVIF (Open Network Video Interface Forum) is an open industry standard for IP-based security products. This library allows you to:

- 🎥 Control IP cameras from any manufacturer (Bosch, Hikvision, Axis, Dahua, etc.)
- 📹 Get RTSP video streams and snapshots
- 🎮 Pan, tilt, and zoom cameras remotely
- 🔧 Configure camera settings (exposure, focus, white balance)
- 🔍 Discover cameras on your network automatically
- 🧪 Test ONVIF implementations without physical hardware

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
go get github.com/0x524a/onvif-go
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

    "github.com/0x524a/onvif-go/discovery"
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

    "github.com/0x524a/onvif-go"
)

func main() {
    // Create client - endpoint can be:
    //   - Full URL: "http://192.168.1.100/onvif/device_service"
    //   - IP with port: "192.168.1.100:8080"
    //   - IP only: "192.168.1.100" (automatically adds http:// and path)
    client, err := onvif.NewClient(
        "192.168.1.100",  // Simple IP address
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

### API Coverage Summary

The onvif-go library provides comprehensive ONVIF protocol support with **200+ implemented APIs** across all major ONVIF services:

- **Device Management**: 98 APIs (100% complete) ✅
- **Media Service**: 14+ APIs (profiles, streams, encoding) ✅
- **PTZ Service**: 13 APIs (movement, presets, status) ✅
- **Imaging Service**: 7 APIs (brightness, contrast, focus control) ✅
- **Discovery Service**: WS-Discovery network scanning ✅

### Client Creation

```go
client, err := onvif.NewClient(
    endpoint,
    onvif.WithCredentials(username, password),
    onvif.WithTimeout(30*time.Second),
    onvif.WithHTTPClient(customHTTPClient),
)
```

### Device Service (98 APIs) - 100% Complete ✅

The Device Service provides comprehensive device management capabilities with **98 fully implemented APIs**:

#### Core Device Information
| Method | Description |
|--------|-------------|
| `GetDeviceInformation()` | Get manufacturer, model, firmware version, serial number, hardware ID |
| `GetCapabilities()` | Get device capabilities and service endpoints (device, media, imaging, PTZ, events, etc.) |
| `GetServices()` | Get list of services with optional capabilities |
| `GetServiceCapabilities()` | Get device service-specific capabilities |
| `GetEndpointReference()` | Get device's WS-Addressing endpoint reference |
| `SystemReboot()` | Reboot the device |
| `Initialize()` | Discover and cache service endpoints |

#### Hostname & Network Discovery
| Method | Description |
|--------|-------------|
| `GetHostname()` | Get device hostname configuration |
| `SetHostname()` | Set device hostname |
| `SetHostnameFromDHCP()` | Enable/disable hostname from DHCP |
| `GetScopes()` | Get configured WS-Discovery scopes |
| `SetScopes()` | Set WS-Discovery scopes |
| `AddScopes()` | Add WS-Discovery scopes |
| `RemoveScopes()` | Remove WS-Discovery scopes |

#### DNS Configuration
| Method | Description |
|--------|-------------|
| `GetDNS()` | Get DNS configuration (DHCP and manual DNS servers) |
| `SetDNS()` | Set DNS configuration (from DHCP, search domains, DNS servers) |

#### NTP Configuration
| Method | Description |
|--------|-------------|
| `GetNTP()` | Get NTP configuration (DHCP and manual NTP servers) |
| `SetNTP()` | Set NTP configuration (from DHCP, NTP servers) |

#### Dynamic DNS
| Method | Description |
|--------|-------------|
| `GetDynamicDNS()` | Get Dynamic DNS configuration |
| `SetDynamicDNS()` | Set Dynamic DNS with type and name |

#### System Date & Time
| Method | Description |
|--------|-------------|
| `GetSystemDateAndTime()` | Get device system date and time (interface{}) |
| `FixedGetSystemDateAndTime()` | Get properly typed system date and time with timezone support |
| `SetSystemDateAndTime()` | Set device system date and time with manual/NTP mode |

#### Network Configuration
| Method | Description |
|--------|-------------|
| `GetNetworkInterfaces()` | Get all network interface configurations |
| `GetNetworkProtocols()` | Get network protocol settings (HTTP, HTTPS, RTSP, RTMP, SSH, etc.) |
| `SetNetworkProtocols()` | Set network protocol settings |
| `GetNetworkDefaultGateway()` | Get default gateway configuration (IPv4 and IPv6) |
| `SetNetworkDefaultGateway()` | Set default gateway configuration |
| `GetZeroConfiguration()` | Get Zero Configuration (zeroconf/Bonjour) status |
| `SetZeroConfiguration()` | Enable/disable Zero Configuration per interface |

#### User Management
| Method | Description |
|--------|-------------|
| `GetUsers()` | Get list of user accounts and credentials |
| `CreateUsers()` | Create new user accounts |
| `SetUser()` | Modify existing user account |
| `DeleteUsers()` | Delete user accounts |
| `GetRemoteUser()` | Get remote user connection status |
| `SetRemoteUser()` | Set remote user connection settings |

#### Security & Access Control
| Method | Description |
|--------|-------------|
| `GetIPAddressFilter()` | Get IP address filter (allow/deny lists) |
| `SetIPAddressFilter()` | Set IP address filtering rules |
| `AddIPAddressFilter()` | Add IP addresses to filter list |
| `RemoveIPAddressFilter()` | Remove IP addresses from filter list |
| `GetPasswordComplexityConfiguration()` | Get password policy settings |
| `SetPasswordComplexityConfiguration()` | Set password policy (length, uppercase, numbers, special chars) |
| `GetPasswordHistoryConfiguration()` | Get password history requirements |
| `SetPasswordHistoryConfiguration()` | Set password history and re-use prevention |
| `GetAuthFailureWarningConfiguration()` | Get failed authentication warning settings |
| `SetAuthFailureWarningConfiguration()` | Set failed authentication thresholds |

#### Discovery Modes
| Method | Description |
|--------|-------------|
| `GetDiscoveryMode()` | Get discovery mode (Discoverable/NonDiscoverable) |
| `SetDiscoveryMode()` | Set discovery mode |
| `GetRemoteDiscoveryMode()` | Get remote discovery mode |
| `SetRemoteDiscoveryMode()` | Set remote discovery mode |

#### Certificate Management
| Method | Description |
|--------|-------------|
| `GetCertificates()` | Get installed certificates |
| `GetCACertificates()` | Get Certificate Authority certificates |
| `LoadCertificates()` | Load/install certificates |
| `LoadCACertificates()` | Load/install CA certificates |
| `CreateCertificate()` | Create self-signed certificate |
| `DeleteCertificates()` | Delete certificates |
| `GetCertificateInformation()` | Get certificate details and validity |
| `GetCertificatesStatus()` | Get certificate usage status |
| `SetCertificatesStatus()` | Set certificate usage (enabled/disabled) |
| `GetPkcs10Request()` | Generate PKCS#10 certificate signing request |
| `LoadCertificateWithPrivateKey()` | Load certificate with private key |
| `GetClientCertificateMode()` | Check if client certificate authentication enabled |
| `SetClientCertificateMode()` | Enable/disable client certificate authentication |

#### WiFi/802.11 Configuration
| Method | Description |
|--------|-------------|
| `GetDot11Capabilities()` | Get WiFi capabilities (cipher suites, auth modes) |
| `GetDot11Status()` | Get WiFi status (SSID, signal strength, link quality) |
| `GetDot1XConfiguration()` | Get 802.1X EAP configuration |
| `GetDot1XConfigurations()` | Get all 802.1X configurations |
| `SetDot1XConfiguration()` | Set 802.1X configuration |
| `CreateDot1XConfiguration()` | Create new 802.1X configuration |
| `DeleteDot1XConfiguration()` | Delete 802.1X configuration |
| `ScanAvailableDot11Networks()` | Scan for available WiFi networks |

#### Storage Configuration
| Method | Description |
|--------|-------------|
| `GetStorageConfigurations()` | Get all storage configurations |
| `GetStorageConfiguration()` | Get specific storage configuration |
| `CreateStorageConfiguration()` | Create new storage configuration |
| `SetStorageConfiguration()` | Update storage configuration |
| `DeleteStorageConfiguration()` | Delete storage configuration |
| `SetHashingAlgorithm()` | Set password hashing algorithm |

#### System Maintenance & Logs
| Method | Description |
|--------|-------------|
| `GetSystemLog()` | Get system logs (boot, security, etc.) |
| `GetSystemBackup()` | Get available system backups |
| `RestoreSystem()` | Restore from backup file |
| `GetSystemUris()` | Get system log and backup URIs |
| `GetSystemSupportInformation()` | Get support information and system details |
| `SetSystemFactoryDefault()` | Reset device to factory defaults |
| `StartFirmwareUpgrade()` | Initiate firmware upgrade |
| `StartSystemRestore()` | Initiate system restore |

#### Relay & Auxiliary I/O
| Method | Description |
|--------|-------------|
| `GetRelayOutputs()` | Get relay outputs and their current state |
| `SetRelayOutputSettings()` | Configure relay output behavior |
| `SetRelayOutputState()` | Set relay output state (active/inactive) |
| `SendAuxiliaryCommand()` | Send auxiliary commands (e.g., IR control) |

#### Additional Features
| Method | Description |
|--------|-------------|
| `GetGeoLocation()` | Get device geographic location |
| `SetGeoLocation()` | Set device geographic location |
| `DeleteGeoLocation()` | Delete geographic location |
| `GetDPAddresses()` | Get WS-Discovery multicast addresses |
| `SetDPAddresses()` | Set WS-Discovery multicast addresses |
| `GetAccessPolicy()` | Get device access policy |
| `SetAccessPolicy()` | Set device access policy |
| `GetWsdlUrl()` | Get device WSDL URL (deprecated) |

## 🔧 Device Management Features

The onvif-go library provides **98 fully-implemented Device Management APIs** for complete device configuration and control. See [DEVICE_API_STATUS.md](DEVICE_API_STATUS.md) for the complete API reference.

### Common Device Management Use Cases

#### Query Device Information
```go
// Get device info (manufacturer, model, firmware)
info, err := client.GetDeviceInformation(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Camera: %s %s (FW: %s)\n", info.Manufacturer, info.Model, info.FirmwareVersion)

// Get capabilities
caps, err := client.GetCapabilities(ctx)
if err != nil {
    log.Fatal(err)
}
```

#### Network Configuration
```go
// Get all network interfaces
interfaces, err := client.GetNetworkInterfaces(ctx)
if err != nil {
    log.Fatal(err)
}

// Get DNS and NTP settings
dns, err := client.GetDNS(ctx)
ntp, err := client.GetNTP(ctx)

// Configure DNS
err = client.SetDNS(ctx, false, []string{"example.com"}, []onvif.IPAddress{
    {Type: "IPv4", IPv4Address: "8.8.8.8"},
})

// Get/Set hostname
hostname, err := client.GetHostname(ctx)
err = client.SetHostname(ctx, "new-camera-name")
```

#### User & Security Management
```go
// Get users
users, err := client.GetUsers(ctx)

// Create new user
err = client.CreateUsers(ctx, []*onvif.User{
    {Username: "operator", Password: "pass123"},
})

// Configure security
err = client.SetPasswordComplexityConfiguration(ctx, &onvif.PasswordComplexityConfiguration{
    MinLen:        8,
    Uppercase:     1,
    Number:        1,
    SpecialChars:  1,
})

// IP address filtering
filter := &onvif.IPAddressFilter{
    Type: onvif.IPAddressFilterAllow,
}
err = client.SetIPAddressFilter(ctx, filter)
```

#### Certificate Management
```go
// Get installed certificates
certs, err := client.GetCertificates(ctx)

// Create self-signed certificate
cert, err := client.CreateCertificate(ctx,
    "cert1",
    "CN=camera.example.com",
    "2024-01-01T00:00:00Z",
    "2025-01-01T00:00:00Z",
)

// Check certificate status
status, err := client.GetCertificatesStatus(ctx)

// Enable client certificate authentication
err = client.SetClientCertificateMode(ctx, true)
```

#### System Maintenance
```go
// Get system logs
log, err := client.GetSystemLog(ctx, onvif.SystemLogTypeBoot)

// Get system backup
backups, err := client.GetSystemBackup(ctx)

// Reboot device
rebootToken, err := client.SystemReboot(ctx)

// Set factory defaults
err = client.SetSystemFactoryDefault(ctx, onvif.FactoryDefaultTypeSoft)

// Firmware upgrade
upgradeToken, err := client.StartFirmwareUpgrade(ctx)
```

#### WiFi Configuration (802.11/802.1X)
```go
// Get WiFi capabilities
caps, err := client.GetDot11Capabilities(ctx)

// Scan available networks
networks, err := client.ScanAvailableDot11Networks(ctx, "interface1")

// Get 802.1X configuration
config, err := client.GetDot1XConfiguration(ctx, "config1")

// Set 802.1X
err = client.SetDot1XConfiguration(ctx, config)
```

#### Relay & I/O Control
```go
// Get relay outputs
relays, err := client.GetRelayOutputs(ctx)

// Control relay state
err = client.SetRelayOutputState(ctx, "relay1", onvif.RelayLogicalStateActive)
err = client.SetRelayOutputState(ctx, "relay1", onvif.RelayLogicalStateInactive)

// Send auxiliary commands (e.g., IR control)
response, err := client.SendAuxiliaryCommand(ctx, "tt:IRLamp|On")
```

### Full API Reference

For complete documentation of all 98 Device Management APIs with detailed descriptions, parameters, and return types, see:
- **[DEVICE_API_STATUS.md](DEVICE_API_STATUS.md)** - Complete API listing with categories and examples

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

    "github.com/0x524a/onvif-go/server"
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
onvif-go/
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

## 🖥️ CLI Tools

### Interactive CLI Tool

Feature-rich command-line interface for camera management and testing:

```bash
go build -o onvif-cli ./cmd/onvif-cli/

# Start interactive menu
./onvif-cli
```

**Features**:
- 🔍 Discover cameras on network with interface selection
- 🌐 View all network interfaces and their capabilities
- 🔗 Connect to cameras with authentication
- 📱 Get device info, capabilities, and system settings
- 📹 Retrieve media profiles and stream URLs
- 🎮 PTZ control (pan, tilt, zoom, presets)
- 🎨 Imaging settings (brightness, contrast, exposure, etc.)
- 📞 Network interface selection for multi-interface systems

**Usage**:
```
📋 Main Menu:
  1. Discover Cameras on Network
  2. Connect to Camera
  3. Device Operations
  4. Media Operations
  5. PTZ Operations
  6. Imaging Operations
  0. Exit
```

Note: The discovery function now intelligently detects multiple interfaces and shows options only when needed - no separate "List Network Interfaces" menu required.

### Quick Demo Tool

Lightweight tool for quick testing and demonstration:

```bash
go build -o onvif-quick ./cmd/onvif-quick/

# Start interactive menu
./onvif-quick
```

**Features**:
- ⚡ Quick camera discovery
- 🌐 List available network interfaces
- 🔗 Quick connection and camera info
- 🎮 PTZ demo with movement examples
- 📡 Stream URL retrieval

### Network Interface Selection

The CLI intelligently handles network interface selection automatically:
- **Single interface**: Auto-discovery works seamlessly
- **Multiple interfaces**: Shows interfaces only if auto-discovery fails
- **Multiple active interfaces**: Tries each one and aggregates results

For programmatic usage:

```go
opts := &discovery.DiscoverOptions{
    NetworkInterface: "eth0",  // By interface name
    // or
    // NetworkInterface: "192.168.1.100",  // By IP address
}
devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second, opts)
```

**See**: 
- `docs/CLI_NETWORK_INTERFACE_USAGE.md` - Detailed CLI guide
- `discovery/NETWORK_INTERFACE_GUIDE.md` - API usage examples
- `DESIGN_REFACTOR.md` - How smart interface detection works

## 🌟 Star History

If you find this project useful, please consider giving it a star! ⭐

[![Star History Chart](https://api.star-history.com/svg?repos=0x524a/onvif-go&type=Date)](https://star-history.com/#0x524a/onvif-go&Date)

## 📊 Project Stats

![GitHub repo size](https://img.shields.io/github/repo-size/0x524a/onvif-go)
![GitHub code size](https://img.shields.io/github/languages/code-size/0x524a/onvif-go)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/0x524a/onvif-go)
![GitHub last commit](https://img.shields.io/github/last-commit/0x524a/onvif-go)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the original [use-go/onvif](https://github.com/use-go/onvif) library
- ONVIF specifications from [ONVIF.org](https://www.onvif.org)
- Thanks to all contributors and the Go community

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/0x524a/onvif-go)
- 🐛 [Issue Tracker](https://github.com/0x524a/onvif-go/issues)
- 💬 [Discussions](https://github.com/0x524a/onvif-go/discussions)
- 🔒 [Security Policy](.github/SECURITY.md)

## Keywords

`onvif` `ip-camera` `surveillance` `golang` `rtsp` `ptz` `camera-control` `video-streaming` `security-camera` `nvr` `vms` `iot` `cctv` `hikvision` `axis` `dahua` `bosch` `camera-sdk` `golang-library` `soap` `ws-discovery`

## Related Projects

- [ONVIF Device Manager](https://sourceforge.net/projects/onvifdm/) - GUI tool for testing ONVIF devices
- [ONVIF Device Tool](https://www.onvif.org/tools/) - Official ONVIF test tool

---

Made with ❤️ for the Go and IoT community