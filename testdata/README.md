# Test Data for ONVIF Camera Testing

This directory contains discovered camera data for testing the onvif-go library.

## Files

### discovered_cameras_20260113.json
JSON file containing structured data for all 8 cameras discovered on the network:
- Complete endpoint information
- XAddrs (service URLs)
- Manufacturer and model details
- Supported ONVIF profiles
- Network configuration (IP, port)
- HTTPS support status

### test_cameras_config.go
Go package providing programmatic access to test camera data:
- `TestCameras` slice with all discovered cameras
- `GetCameraByManufacturer()` - filter by manufacturer
- `GetCameraByProfile()` - filter by ONVIF profile support
- `GetHTTPSCameras()` - get cameras with HTTPS support

## Discovery Summary (2026-01-13)

**Total Cameras Found:** 8

### By Manufacturer:
- **AXIS:** 3 cameras (P3818-PVE, Q3819-PVE, P5655-E)
- **Bosch:** 3 cameras (AUTODOME IP starlight 5000i, FLEXIDOME IP starlight 8000i, FLEXIDOME panoramic 5100i)
- **Reolink:** 2 cameras (E1Zoom, ReolinkTrackMixWiFi)

### By ONVIF Profile Support:
- **Profile Streaming:** 8/8 (100%)
- **Profile T (Streaming):** 8/8 (100%)
- **Profile G (Recording):** 6/8 (75%)
- **Profile M (Metadata):** 4/8 (50%)

### Network Configuration:
- Network: 192.168.2.0/24
- HTTPS Support: 6/8 cameras
- Port 80: 6 cameras
- Port 8000: 2 cameras (Reolink)

## Usage in Tests

### Example 1: Using JSON Data
```go
import (
    "encoding/json"
    "os"
)

type CameraData struct {
    Cameras []struct {
        IP           string   `json:"ip"`
        XAddrs       []string `json:"xaddrs"`
        Manufacturer string   `json:"manufacturer"`
        Model        string   `json:"model"`
    } `json:"cameras"`
}

func loadTestCameras() (*CameraData, error) {
    data, err := os.ReadFile("testdata/discovered_cameras_20260113.json")
    if err != nil {
        return nil, err
    }
    var cameras CameraData
    err = json.Unmarshal(data, &cameras)
    return &cameras, err
}
```

### Example 2: Using Go Package
```go
import "github.com/yourusername/onvif-go/testdata"

func TestWithAxisCameras(t *testing.T) {
    axisCameras := testdata.GetCameraByManufacturer("AXIS")
    for _, cam := range axisCameras {
        t.Logf("Testing with %s %s at %s", cam.Manufacturer, cam.Model, cam.IP)
        // Run your tests...
    }
}

func TestProfileM(t *testing.T) {
    metadataCameras := testdata.GetCameraByProfile("M")
    if len(metadataCameras) == 0 {
        t.Skip("No cameras with Profile M support")
    }
    // Test metadata operations...
}

func TestHTTPS(t *testing.T) {
    httpsCameras := testdata.GetHTTPSCameras()
    for _, cam := range httpsCameras {
        // Test HTTPS connections...
    }
}
```

## Camera Details

### High-End Cameras (Profile G + M)
- AXIS P3818-PVE (192.168.2.82)
- AXIS Q3819-PVE (192.168.2.190) - Dual network interfaces
- AXIS P5655-E (192.168.2.30)
- Bosch FLEXIDOME panoramic 5100i (192.168.2.24)

### Mid-Range Cameras (Profile G)
- Bosch AUTODOME IP starlight 5000i (192.168.2.57)
- Bosch FLEXIDOME IP starlight 8000i (192.168.2.200)

### Basic Cameras (Profile T only)
- Reolink E1Zoom (192.168.2.61:8000)
- Reolink ReolinkTrackMixWiFi (192.168.2.236:8000)

## Notes

1. **Credentials Required:** These endpoints require authentication. Set test credentials using environment variables:
   ```bash
   export ONVIF_TEST_USERNAME="your_username"
   export ONVIF_TEST_PASSWORD="your_password"
   ```

2. **Network Access:** Tests require access to the 192.168.2.0/24 network.

3. **Camera Availability:** Ensure cameras are powered on and network-accessible before running tests.

4. **HTTPS Certificates:** AXIS and Bosch cameras use self-signed certificates. Tests may need to skip certificate verification:
   ```go
   client.HTTPClient = &http.Client{
       Transport: &http.Transport{
           TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
       },
   }
   ```

5. **Rate Limiting:** Some cameras may rate-limit requests. Add delays between test runs if needed.

## Updating Test Data

To refresh the discovered camera data:

```bash
# Run discovery and save output
./bin/discover 2>&1 | tee camera-discovery-$(date +%Y%m%d-%H%M%S).log

# Discovery will run for ~10 seconds
# Press Ctrl+C to stop when cameras are found

# Update JSON and Go files with new data as needed
```

## See Also

- [Main Testing Documentation](../docs/testing/)
- [Camera Test Reports](../CAMERA_TEST_REPORT.md)
- [Quick Start Guide](../docs/QUICKSTART.md)
