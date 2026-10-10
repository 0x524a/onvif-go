#!/bin/bash

# Go ONVIF Library Demo Script
# This script demonstrates the capabilities of the Go ONVIF library

# Run from the repository root whatever the caller's working directory.
cd "$(dirname "${BASH_SOURCE[0]}")/.."

echo "🎥 Go ONVIF Library - Complete Implementation Demo"
echo "=================================================="
echo

echo "📁 Project Structure:"
echo "├── Core Library (client.go, types.go, device.go, media.go, ptz.go, imaging.go)"
echo "├── SOAP Client (soap/soap.go) with WS-Security authentication"
echo "├── Discovery Service (discovery/discovery.go) for network camera detection"
echo "├── Examples (examples/*) showing various use cases"
echo "├── CLI Tools:"
echo "│   ├── 🔧 onvif-cli - Comprehensive interactive tool"
echo "│   └── ⚡ onvif-quick - Simple quick-start tool"
echo "└── Tests with mock ONVIF server"
echo

echo "🚀 Available Commands:"
echo

echo "1. Build & Test:"
echo "   make build        # Build both CLI tools"
echo "   make test         # Run test suite"
echo "   make examples     # Build example programs"
echo "   make build-all    # Build for multiple platforms"
echo

echo "2. CLI Tools:"
echo "   ./bin/onvif-cli   # Interactive comprehensive tool"
echo "   ./bin/onvif-quick # Simple quick-start tool"
echo

echo "3. Library Usage Example:"
cat << 'EOF'
```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/0x524A/onvif-go"
)

func main() {
    // Create client with credentials
    client, err := onvif.NewClient(
        "http://192.168.1.100/onvif/device_service",
        onvif.WithCredentials("admin", "password"),
        onvif.WithTimeout(30*time.Second),
    )
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // Get device information
    info, err := client.GetDeviceInformation(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Camera: %s %s\n", info.Manufacturer, info.Model)
    
    // Initialize for additional services
    client.Initialize(ctx)
    
    // Get media profiles
    profiles, err := client.GetProfiles(ctx)
    if err != nil {
        panic(err)
    }
    
    // Get stream URI
    streamURI, err := client.GetStreamURI(ctx, profiles[0].Token)
    if err == nil {
        fmt.Printf("Stream: %s\n", streamURI.URI)
    }
    
    // PTZ Control (if supported)
    velocity := &onvif.PTZSpeed{
        PanTilt: &onvif.Vector2D{X: 0.5, Y: 0.0},
    }
    timeout := "PT5S"
    client.ContinuousMove(ctx, profiles[0].Token, velocity, &timeout)
}
```
EOF

echo
echo "🌟 Key Features:"
echo "✅ Complete ONVIF Profile S implementation"
echo "✅ WS-Discovery for automatic camera detection"
echo "✅ WS-Security authentication with digest"
echo "✅ PTZ control (pan, tilt, zoom)"
echo "✅ Media profile management"
echo "✅ Imaging settings control"
echo "✅ Device information and capabilities"
echo "✅ Stream URI generation (RTSP/HTTP)"
echo "✅ Context-based timeout and cancellation"
echo "✅ Comprehensive error handling"
echo "✅ Thread-safe credential management"
echo "✅ Interactive CLI tools"
echo "✅ Docker support"
echo "✅ Cross-platform builds"
echo "✅ Extensive test coverage"
echo

echo "🛠️  Development Features:"
echo "✅ Modern Go 1.21+ with generics support"
echo "✅ Functional options pattern"
echo "✅ Comprehensive type definitions"
echo "✅ Mock server for testing"
echo "✅ Benchmark tests"
echo "✅ CI/CD ready"
echo "✅ Docker containerization"
echo "✅ Multi-platform builds"
echo

echo "📋 Quick Start:"
echo "1. go mod tidy                    # Install dependencies"
echo "2. make build                     # Build CLI tools"
echo "3. ./bin/onvif-quick             # Run quick tool"
echo "4. ./bin/onvif-cli               # Run comprehensive tool"
echo

echo "🔗 For real camera testing:"
echo "- Set up a test camera with known IP/credentials"
echo "- Run discovery to find cameras: ./bin/onvif-quick"
echo "- Use device info to verify connection"
echo "- Test PTZ movements if camera supports it"
echo "- Get stream URLs for media playback"
echo

echo "🎯 This implementation provides a production-ready,"
echo "   comprehensive ONVIF library with full CLI tooling!"

echo
echo "Run 'make help' for all available commands."