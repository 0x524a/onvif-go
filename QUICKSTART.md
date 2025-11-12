# Quick Start Guide

Get up and running with go-onvif in 5 minutes!

## Installation

```bash
go get github.com/0x524a/onvif-go
```

## Step 1: Discover Cameras

Find ONVIF cameras on your network:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/0x524a/onvif-go/discovery"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    devices, err := discovery.Discover(ctx, 5*time.Second)
    if err != nil {
        panic(err)
    }
    
    for _, device := range devices {
        fmt.Printf("Found: %s at %s\n", 
            device.GetName(), 
            device.GetDeviceEndpoint())
    }
}
```

## Step 2: Connect to Camera

Create a client and get basic information. The endpoint can be specified in multiple formats:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/0x524a/onvif-go"
)

func main() {
    // Create client - endpoint accepts multiple formats:
    //   - Simple IP: "192.168.1.100"
    //   - IP with port: "192.168.1.100:8080"  
    //   - Full URL: "http://192.168.1.100/onvif/device_service"
    client, err := onvif.NewClient(
        "192.168.1.100",  // Simple IP address works!
        onvif.WithCredentials("admin", "password"),
        onvif.WithTimeout(30*time.Second),
    )
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // Get device info
    info, err := client.GetDeviceInformation(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Camera: %s %s (Firmware: %s)\n",
        info.Manufacturer,
        info.Model,
        info.FirmwareVersion)
}
```

## Step 3: Get Stream URL

Retrieve RTSP stream URLs:

```go
// Initialize client (discovers service endpoints)
if err := client.Initialize(ctx); err != nil {
    panic(err)
}

// Get profiles
profiles, err := client.GetProfiles(ctx)
if err != nil {
    panic(err)
}

// Get stream URI for first profile
if len(profiles) > 0 {
    streamURI, err := client.GetStreamURI(ctx, profiles[0].Token)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Stream URL: %s\n", streamURI.URI)
    // Example: rtsp://192.168.1.100/stream1
}
```

## Step 4: Control PTZ

Move the camera:

```go
profileToken := profiles[0].Token

// Move right for 2 seconds
velocity := &onvif.PTZSpeed{
    PanTilt: &onvif.Vector2D{X: 0.5, Y: 0.0},
}
timeout := "PT2S"
client.ContinuousMove(ctx, profileToken, velocity, &timeout)

time.Sleep(2 * time.Second)

// Stop movement
client.Stop(ctx, profileToken, true, false)

// Go to home position
homePosition := &onvif.PTZVector{
    PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.0},
}
client.AbsoluteMove(ctx, profileToken, homePosition, nil)
```

## Step 5: Adjust Image Settings

Modify camera imaging settings:

```go
// Get video source token
videoSourceToken := profiles[0].VideoSourceConfiguration.SourceToken

// Get current settings
settings, err := client.GetImagingSettings(ctx, videoSourceToken)
if err != nil {
    panic(err)
}

// Modify brightness and contrast
brightness := 60.0
settings.Brightness = &brightness

contrast := 55.0
settings.Contrast = &contrast

// Apply settings
err = client.SetImagingSettings(ctx, videoSourceToken, settings, true)
if err != nil {
    panic(err)
}

fmt.Println("Imaging settings updated!")
```

## Complete Example

Here's a complete program that does everything:

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
    // Configuration
    endpoint := "http://192.168.1.100/onvif/device_service"
    username := "admin"
    password := "password"
    
    // Create client
    client, err := onvif.NewClient(
        endpoint,
        onvif.WithCredentials(username, password),
        onvif.WithTimeout(30*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Get device information
    fmt.Println("Getting device information...")
    info, err := client.GetDeviceInformation(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Camera: %s %s\n", info.Manufacturer, info.Model)
    
    // Initialize client
    fmt.Println("\nInitializing client...")
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Get profiles
    fmt.Println("Getting media profiles...")
    profiles, err := client.GetProfiles(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    if len(profiles) == 0 {
        log.Fatal("No profiles found")
    }
    
    profile := profiles[0]
    fmt.Printf("Using profile: %s\n", profile.Name)
    
    // Get stream URI
    streamURI, err := client.GetStreamURI(ctx, profile.Token)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Stream URI: %s\n", streamURI.URI)
    
    // Get snapshot URI
    snapshotURI, err := client.GetSnapshotURI(ctx, profile.Token)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Snapshot URI: %s\n", snapshotURI.URI)
    
    // PTZ control (if supported)
    fmt.Println("\nTesting PTZ control...")
    status, err := client.GetStatus(ctx, profile.Token)
    if err != nil {
        fmt.Printf("PTZ not supported or error: %v\n", err)
    } else {
        fmt.Println("PTZ is supported!")
        if status.Position != nil && status.Position.PanTilt != nil {
            fmt.Printf("Current position: X=%.2f, Y=%.2f\n",
                status.Position.PanTilt.X,
                status.Position.PanTilt.Y)
        }
    }
    
    fmt.Println("\nSetup complete!")
}
```

## Next Steps

1. **Explore Examples**: Check out the `examples/` directory for more detailed use cases
2. **Read Documentation**: Visit [pkg.go.dev](https://pkg.go.dev/github.com/0x524a/onvif-go)
3. **Review Architecture**: See [ARCHITECTURE.md](ARCHITECTURE.md) for design details
4. **Check Issues**: Look at [GitHub Issues](https://github.com/0x524a/onvif-go/issues) for known issues

## Common Patterns

### Error Handling

```go
info, err := client.GetDeviceInformation(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    } else if onvif.IsONVIFError(err) {
        // Handle SOAP fault
    } else {
        // Handle other errors
    }
    return err
}
```

### Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := client.SomeOperation(ctx)
```

### Checking Service Support

```go
status, err := client.GetStatus(ctx, profileToken)
if errors.Is(err, onvif.ErrServiceNotSupported) {
    fmt.Println("PTZ not supported on this camera")
} else if err != nil {
    return err
}
```

## Tips & Tricks

1. **Always Initialize**: Call `client.Initialize(ctx)` before using service-specific methods
2. **Use Timeouts**: Always use contexts with timeouts for network operations
3. **Reuse Clients**: Create one client per camera and reuse it
4. **Check Capabilities**: Use `GetCapabilities()` to check what the camera supports
5. **Handle Errors**: Check for `ErrServiceNotSupported` when using optional services

## Troubleshooting

### Camera Not Found During Discovery
- Check network connectivity
- Ensure camera is on the same subnet
- Verify ONVIF is enabled on the camera
- Check firewall settings (UDP port 3702)

### Authentication Failed
- Verify username and password
- Check if camera requires admin privileges
- Some cameras need authentication enabled

### Connection Timeout
- Increase timeout duration
- Check network latency
- Verify endpoint URL is correct
- Test with ping/curl first

### Service Not Supported
- Check camera capabilities with `GetCapabilities()`
- Update camera firmware if needed
- Some features require specific ONVIF profiles

## Additional Resources

- [ONVIF Official Site](https://www.onvif.org)
- [ONVIF Core Specification](https://www.onvif.org/specs/core/ONVIF-Core-Specification.pdf)
- [ONVIF Device Test Tool](https://www.onvif.org/tools/)

Happy coding! 🎥📹
