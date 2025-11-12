# ONVIF Server - Virtual IP Camera Simulator

A complete ONVIF-compliant server implementation that simulates multi-lens IP cameras with full support for Device, Media, PTZ, and Imaging services.

## Features

### 🎥 Multi-Lens Camera Support
- **Multiple Video Profiles**: Support for up to 10 independent camera profiles
- **Different Resolutions**: From 640x480 to 4K (3840x2160)
- **Configurable Framerates**: 25, 30, 60 fps
- **Multiple Encodings**: H.264, H.265, MPEG4, JPEG

### 🎮 PTZ Control
- **Continuous Movement**: Smooth pan, tilt, and zoom control
- **Absolute Positioning**: Move to specific coordinates
- **Relative Movement**: Move relative to current position
- **Preset Positions**: Save and recall camera positions
- **Status Monitoring**: Real-time PTZ state information

### 📷 Imaging Control
- **Brightness, Contrast, Saturation**: Full color control
- **Exposure Settings**: Auto/Manual modes with gain control
- **Focus Control**: Auto-focus and manual focus positioning
- **White Balance**: Auto/Manual white balance adjustment
- **Wide Dynamic Range (WDR)**: Enhanced contrast in challenging lighting
- **IR Cut Filter**: Day/Night mode control

### 🌐 ONVIF Services
- ✅ **Device Service**: Device information, capabilities, system time
- ✅ **Media Service**: Profiles, stream URIs (RTSP), snapshots
- ✅ **PTZ Service**: Full PTZ control and preset management
- ✅ **Imaging Service**: Complete imaging settings control
- ⏳ **Events Service**: (Planned)

### 🔐 Security
- **WS-Security Authentication**: UsernameToken with password digest
- **Configurable Credentials**: Custom username/password
- **SOAP Message Security**: Nonce and timestamp validation

## Installation

```bash
# Clone the repository (if not already done)
git clone https://github.com/0x524a/onvif-go
cd go-onvif

# Build the server CLI
go build -o onvif-server ./cmd/onvif-server

# Or install globally
go install ./cmd/onvif-server
```

## Quick Start

### Basic Usage

Start the server with default settings (3 camera profiles):

```bash
./onvif-server
```

The server will start on `http://0.0.0.0:8080` with:
- Username: `admin`
- Password: `admin`
- 3 camera profiles with different resolutions
- PTZ and Imaging services enabled

### Custom Configuration

```bash
# Custom credentials and port
./onvif-server -username myuser -password mypass -port 9000

# More camera profiles
./onvif-server -profiles 5

# Disable PTZ
./onvif-server -ptz=false

# Custom device information
./onvif-server -manufacturer "Acme Corp" -model "SuperCam 5000"
```

### Command-Line Options

```
  -host string
        Server host address (default "0.0.0.0")
  -port int
        Server port (default 8080)
  -username string
        Authentication username (default "admin")
  -password string
        Authentication password (default "admin")
  -manufacturer string
        Device manufacturer (default "go-onvif")
  -model string
        Device model (default "Virtual Multi-Lens Camera")
  -firmware string
        Firmware version (default "1.0.0")
  -serial string
        Serial number (default "SN-12345678")
  -profiles int
        Number of camera profiles (1-10) (default 3)
  -ptz
        Enable PTZ support (default true)
  -imaging
        Enable Imaging support (default true)
  -events
        Enable Events support (default false)
  -info
        Show server info and exit
  -version
        Show version and exit
```

## Using the Server Library

### Simple Example

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/0x524a/onvif-go/server"
)

func main() {
    // Use default configuration
    config := server.DefaultConfig()
    
    // Or customize
    config.Port = 9000
    config.Username = "myuser"
    config.Password = "mypass"

    // Create server
    srv, err := server.New(config)
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

### Custom Multi-Lens Camera

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/0x524a/onvif-go/server"
)

func main() {
    config := &server.Config{
        Host:     "0.0.0.0",
        Port:     8080,
        BasePath: "/onvif",
        Timeout:  30 * time.Second,
        DeviceInfo: server.DeviceInfo{
            Manufacturer:    "MultiCam Systems",
            Model:           "MC-3000 Pro",
            FirmwareVersion: "2.5.1",
            SerialNumber:    "MC3000-001234",
            HardwareID:      "HW-MC3000",
        },
        Username:       "admin",
        Password:       "SecurePass123",
        SupportPTZ:     true,
        SupportImaging: true,
        SupportEvents:  false,
        Profiles: []server.ProfileConfig{
            {
                Token: "profile_main_4k",
                Name:  "Main Camera 4K",
                VideoSource: server.VideoSourceConfig{
                    Token:      "video_source_main",
                    Name:       "Main Camera",
                    Resolution: server.Resolution{Width: 3840, Height: 2160},
                    Framerate:  30,
                },
                VideoEncoder: server.VideoEncoderConfig{
                    Encoding:   "H264",
                    Resolution: server.Resolution{Width: 3840, Height: 2160},
                    Quality:    90,
                    Framerate:  30,
                    Bitrate:    20480, // 20 Mbps
                    GovLength:  30,
                },
                PTZ: &server.PTZConfig{
                    NodeToken:          "ptz_main",
                    PanRange:           server.Range{Min: -180, Max: 180},
                    TiltRange:          server.Range{Min: -90, Max: 90},
                    ZoomRange:          server.Range{Min: 0, Max: 10},
                    SupportsContinuous: true,
                    SupportsAbsolute:   true,
                    SupportsRelative:   true,
                    Presets: []server.Preset{
                        {Token: "preset_home", Name: "Home", Position: server.PTZPosition{Pan: 0, Tilt: 0, Zoom: 0}},
                        {Token: "preset_entrance", Name: "Entrance", Position: server.PTZPosition{Pan: -45, Tilt: -20, Zoom: 3}},
                    },
                },
                Snapshot: server.SnapshotConfig{
                    Enabled:    true,
                    Resolution: server.Resolution{Width: 3840, Height: 2160},
                    Quality:    95,
                },
            },
            // Add more profiles...
        },
    }

    srv, err := server.New(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := srv.Start(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Testing with ONVIF Client

You can test the server with the included ONVIF client library:

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
    // Connect to the server
    client, err := onvif.NewClient(
        "http://localhost:8080/onvif/device_service",
        onvif.WithCredentials("admin", "admin"),
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
    fmt.Printf("Device: %s %s\n", info.Manufacturer, info.Model)

    // Initialize to discover services
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }

    // Get media profiles
    profiles, err := client.GetProfiles(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d profiles:\n", len(profiles))
    for i, profile := range profiles {
        fmt.Printf("  [%d] %s\n", i+1, profile.Name)
        
        // Get stream URI
        streamURI, err := client.GetStreamURI(ctx, profile.Token)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("      Stream: %s\n", streamURI.URI)
    }

    // PTZ control (if available)
    if len(profiles) > 0 && profiles[0].PTZConfiguration != nil {
        profileToken := profiles[0].Token
        
        // Get PTZ status
        status, err := client.GetStatus(ctx, profileToken)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("PTZ Position: Pan=%.2f, Tilt=%.2f, Zoom=%.2f\n",
            status.Position.PanTilt.X,
            status.Position.PanTilt.Y,
            status.Position.Zoom.X)
        
        // Move to home position
        position := &onvif.PTZVector{
            PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.0},
            Zoom:    &onvif.Vector1D{X: 0.0},
        }
        if err := client.AbsoluteMove(ctx, profileToken, position, nil); err != nil {
            log.Fatal(err)
        }
        fmt.Println("Moved to home position")
    }
}
```

## Examples

See the [examples/onvif-server](../../examples/onvif-server) directory for a complete multi-lens camera configuration example.

```bash
# Run the example
cd examples/onvif-server
go run main.go
```

This example demonstrates:
- 4 different camera profiles (4K main, wide-angle, telephoto, low-light)
- PTZ control with multiple presets
- Different resolutions and framerates
- Custom device information

## Use Cases

### 🧪 Testing & Development
- Test ONVIF client implementations
- Simulate multi-camera setups
- Develop video management systems
- Integration testing without physical cameras

### 📚 Learning & Education
- Understand ONVIF protocol
- Learn SOAP web services
- Study IP camera architectures
- Prototype camera systems

### 🎭 Demonstrations
- Demo video surveillance solutions
- Showcase camera management software
- Present multi-camera scenarios
- Trade show demonstrations

### 🔬 Research & Prototyping
- Computer vision research
- Video analytics development
- Stream processing pipelines
- AI/ML model training

## Architecture

The server is built with a modular architecture:

```
server/
├── types.go           # Core data types and configuration
├── server.go          # Main server implementation
├── device.go          # Device service handlers
├── media.go           # Media service handlers
├── ptz.go             # PTZ service handlers
├── imaging.go         # Imaging service handlers
└── soap/
    └── handler.go     # SOAP message handling
```

### Key Components

1. **Server Core**: HTTP server, request routing, lifecycle management
2. **SOAP Handler**: SOAP message parsing, authentication, response formatting
3. **Service Handlers**: Device, Media, PTZ, Imaging service implementations
4. **State Management**: PTZ positions, imaging settings, stream configurations

## RTSP Streaming

The server provides RTSP URIs for each profile:

```
rtsp://localhost:8554/stream0  # Profile 0
rtsp://localhost:8554/stream1  # Profile 1
rtsp://localhost:8554/stream2  # Profile 2
...
```

**Note**: The current implementation returns RTSP URIs but does not include an actual RTSP server. To provide real video streams, integrate with:

- [RTSPtoWeb](https://github.com/deepch/RTSPtoWeb)
- [MediaMTX](https://github.com/bluenviron/mediamtx)
- [FFmpeg RTSP server](https://ffmpeg.org/)
- Custom RTSP implementation

## Roadmap

- [ ] **Events Service**: Event subscription and notification
- [ ] **Recording Service**: Recording management
- [ ] **Analytics Service**: Video analytics support
- [ ] **Actual RTSP Streaming**: Integrated RTSP server with test patterns
- [ ] **Web UI**: Browser-based configuration and monitoring
- [ ] **Docker Support**: Containerized deployment
- [ ] **Configuration Files**: YAML/JSON configuration support
- [ ] **WS-Discovery**: Automatic device discovery on network
- [ ] **TLS Support**: HTTPS and secure RTSP
- [ ] **Audio Support**: Audio streaming and configuration

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## Acknowledgments

- Built on top of the [go-onvif](https://github.com/0x524a/onvif-go) client library
- ONVIF specifications from [ONVIF.org](https://www.onvif.org)
- Inspired by the need for flexible camera simulation in development workflows

---

**Note**: This is a virtual camera server for testing and development. It simulates ONVIF protocol responses but does not capture or stream real video unless integrated with an RTSP server.
