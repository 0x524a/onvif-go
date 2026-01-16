# Test Generator

Automatically generate Go tests from captured ONVIF camera XML traffic.

## Overview

This tool reads XML capture archives (created by `onvif-diagnostics -capture-xml`) and generates complete Go test files that replay the captured SOAP traffic through a mock server.

## Usage

### Basic Usage

```bash
./generate-tests \
  -capture camera-logs/Camera_Model_xmlcapture_timestamp.tar.gz \
  -output testdata/captures/
```

### Options

```
-capture string
    Path to XML capture archive (.tar.gz) (required)
    
-output string
    Output directory for generated test file (default: "./")
    
-package string
    Package name for generated test (default: "onvif_test")
```

## Example

```bash
# Generate test from Bosch camera capture
./generate-tests \
  -capture camera-logs/Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_xmlcapture_20251110-120000.tar.gz \
  -output testdata/captures/

# Output:
# ✓ Generated test file: testdata/captures/bosch_flexidome_indoor_5100i_ir_8.71.0066_test.go
#   Camera: Bosch FLEXIDOME indoor 5100i IR (Firmware: 8.71.0066)
#   Captured operations: 18
```

## Generated Test Structure

The tool creates a complete test file with:

### Test Function

```go
func Test<CameraName>(t *testing.T)
```

Named based on camera manufacturer, model, and firmware.

### Subtests

- `GetDeviceInformation` - Validates device info parsing
- `GetSystemDateAndTime` - Tests date/time operation
- `GetCapabilities` - Verifies capability discovery
- `GetProfiles` - Tests media profile enumeration

### Assertions

Each subtest includes:
- Error checking
- Nil validation
- Basic field validation
- Informative logging

## How It Works

1. **Load Capture** - Reads all SOAP exchanges from tar.gz archive
2. **Extract Metadata** - Gets camera manufacturer, model, firmware from responses
3. **Generate Name** - Creates valid Go identifier from camera info
4. **Render Template** - Fills in test template with camera-specific data
5. **Write File** - Saves test to output directory

## Template

The generator uses an embedded Go template that creates:

```go
package onvif_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/0x524a/onvif-go"
    onviftesting "github.com/0x524a/onvif-go/testing"
)

func Test<CameraName>(t *testing.T) {
    captureArchive := "<archive-file>.tar.gz"
    
    mockServer, err := onviftesting.NewMockSOAPServer(captureArchive)
    if err != nil {
        t.Fatalf("Failed to create mock server: %v", err)
    }
    defer mockServer.Close()
    
    client, err := onvif.NewClient(
        mockServer.URL()+"/onvif/device_service",
        onvif.WithCredentials("testuser", "testpass"),
    )
    // ... test operations
}
```

## Workflow

### 1. Capture from Camera

```bash
./onvif-diagnostics \
  -endpoint "http://camera/onvif/device_service" \
  -username "user" \
  -password "pass" \
  -capture-xml
```

### 2. Generate Test

```bash
./generate-tests \
  -capture camera-logs/Camera_*_xmlcapture_*.tar.gz \
  -output testdata/captures/
```

### 3. Run Test

```bash
go test -v ./testdata/captures/ -run TestCamera
```

## Customization

After generation, you can customize the test:

### Add Camera-Specific Tests

```go
t.Run("CustomFeature", func(t *testing.T) {
    // Add custom test for camera-specific features
})
```

### Add Detailed Assertions

```go
t.Run("GetDeviceInformation", func(t *testing.T) {
    info, err := client.GetDeviceInformation(ctx)
    if err != nil {
        t.Errorf("GetDeviceInformation failed: %v", err)
        return
    }
    
    // Add specific assertions
    if info.Manufacturer != "ExpectedManufacturer" {
        t.Errorf("Expected manufacturer X, got %s", info.Manufacturer)
    }
})
```

## Building

```bash
go build -o generate-tests ./cmd/generate-tests/
```

## Dependencies

- `github.com/0x524a/onvif-go/testing` - Mock server and capture loader

## Output File Naming

Generated test files are named:

```
<manufacturer>_<model>_<firmware>_test.go
```

Examples:
- `bosch_flexidome_indoor_5100i_ir_8.71.0066_test.go`
- `axis_q3626-ve_12.6.104_test.go`
- `reolink_e1_zoom_v3.1.0.2649_test.go`

All special characters converted to underscores or removed.

## Archive Path Handling

The generator automatically handles archive paths:

- If archive is in output directory, uses filename only
- Otherwise uses relative path from output directory
- Tests can find archives when run with `go test ./testdata/captures/`

## Troubleshooting

### "Failed to load capture"

Archive file not found or corrupted.

**Solution**: Verify archive path and ensure it's a valid tar.gz file.

### "Failed to extract device info"

Archive doesn't contain GetDeviceInformation response.

**Solution**: Re-capture from camera, ensuring diagnostic runs fully.

### Generated test won't compile

Usually due to invalid characters in camera names.

**Solution**: The generator should handle this, but you can manually edit the test function name.

## Future Enhancements

Potential improvements:

- [ ] Detect camera-specific operations (PTZ, audio, etc.)
- [ ] Generate profile-specific tests
- [ ] Add benchmarking subtests
- [ ] Support custom test templates
- [ ] Batch generation from multiple captures

## See Also

- `testdata/captures/README.md` - Using generated tests
- `testing/mock_server.go` - Mock server implementation
- `cmd/onvif-diagnostics/` - Capturing tool
