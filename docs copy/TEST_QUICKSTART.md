# Quick Test Reference

## Running Camera Tests

### Option 1: Using the test script (Recommended)
```bash
# Set credentials
export ONVIF_TEST_ENDPOINT="http://192.168.1.201/onvif/device_service"
export ONVIF_TEST_USERNAME="service"
export ONVIF_TEST_PASSWORD="Service.1234"

# Run all Bosch FLEXIDOME tests
./run-camera-tests.sh

# Run specific test
./run-camera-tests.sh TestBoschFLEXIDOMEIndoor5100iIR_GetDeviceInformation
```

### Option 2: Direct go test commands
```bash
# Run all camera tests
go test -v -run TestBoschFLEXIDOMEIndoor5100iIR

# Run specific test
go test -v -run TestBoschFLEXIDOMEIndoor5100iIR_GetStreamURI

# Run with race detection
go test -v -race -run TestBoschFLEXIDOMEIndoor5100iIR

# Run benchmarks
go test -v -bench=BenchmarkBoschFLEXIDOMEIndoor5100iIR -benchmem
```

### Option 3: One-liner with credentials
```bash
ONVIF_TEST_ENDPOINT="http://192.168.1.201/onvif/device_service" \
ONVIF_TEST_USERNAME="service" \
ONVIF_TEST_PASSWORD="Service.1234" \
go test -v -run TestBoschFLEXIDOMEIndoor5100iIR
```

## Test List

### Device Tests
- `TestBoschFLEXIDOMEIndoor5100iIR_GetDeviceInformation` - Device info retrieval
- `TestBoschFLEXIDOMEIndoor5100iIR_GetSystemDateAndTime` - System time
- `TestBoschFLEXIDOMEIndoor5100iIR_GetCapabilities` - Capability discovery

### Media Tests
- `TestBoschFLEXIDOMEIndoor5100iIR_GetProfiles` - Media profiles (4 expected)
- `TestBoschFLEXIDOMEIndoor5100iIR_GetStreamURI` - RTSP stream URIs
- `TestBoschFLEXIDOMEIndoor5100iIR_GetSnapshotURI` - Snapshot URLs
- `TestBoschFLEXIDOMEIndoor5100iIR_GetVideoEncoderConfiguration` - Encoder settings

### Imaging Tests
- `TestBoschFLEXIDOMEIndoor5100iIR_GetImagingSettings` - Camera imaging parameters

### Integration Tests
- `TestBoschFLEXIDOMEIndoor5100iIR_Initialize` - Service discovery
- `TestBoschFLEXIDOMEIndoor5100iIR_FullWorkflow` - Complete operation sequence

### Performance Tests
- `BenchmarkBoschFLEXIDOMEIndoor5100iIR_GetDeviceInformation` - Device info benchmark
- `BenchmarkBoschFLEXIDOMEIndoor5100iIR_GetStreamURI` - Stream URI benchmark

## Expected Test Results

All tests should **PASS** with the following outputs:

```
✓ Manufacturer: Bosch
✓ Model: FLEXIDOME indoor 5100i IR
✓ 4 Profiles found (1920x1080, 1536x864, 1280x720, 512x288)
✓ All profiles have RTSP stream URIs
✓ Snapshot URI available
✓ Video encoding: H264 @ 30fps, 5200kbps
✓ Default imaging: Brightness 128.0, Saturation 128.0, Contrast 128.0
```

## Troubleshooting

### Tests are skipped
**Solution**: Set environment variables with camera credentials

### Connection timeout
**Solutions**:
- Verify camera IP address
- Check network connectivity
- Ensure firewall allows connection

### Authentication failed
**Solutions**:
- Verify username and password
- Check user permissions on camera

### Unexpected values
**Note**: Camera settings may differ based on:
- Firmware version
- Manual configuration changes
- Update test expectations if needed

## Coverage Report

Generate test coverage:
```bash
go test -coverprofile=coverage.out -run TestBoschFLEXIDOMEIndoor5100iIR
go tool cover -html=coverage.out
```

## Adding New Camera Tests

1. Copy `bosch_flexidome_test.go` to `<vendor>_<model>_test.go`
2. Update test function names
3. Update expected values
4. Run tests to verify
5. Document in CAMERA_TESTS.md
