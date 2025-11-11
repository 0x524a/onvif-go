# Camera-Specific Integration Tests

This directory contains integration tests for specific ONVIF camera models based on real-world testing.

## Bosch FLEXIDOME indoor 5100i IR Tests

The `bosch_flexidome_test.go` file contains comprehensive tests verified against a real Bosch FLEXIDOME indoor 5100i IR camera running firmware 8.71.0066.

### Running the Tests

Set the following environment variables with your camera credentials:

```bash
export ONVIF_TEST_ENDPOINT="http://192.168.1.201/onvif/device_service"
export ONVIF_TEST_USERNAME="service"
export ONVIF_TEST_PASSWORD="Service.1234"
```

Then run the tests:

```bash
# Run all tests
go test -v ./... -run TestBoschFLEXIDOMEIndoor5100iIR

# Run specific test
go test -v -run TestBoschFLEXIDOMEIndoor5100iIR_GetDeviceInformation

# Run all tests with race detection
go test -v -race -run TestBoschFLEXIDOMEIndoor5100iIR

# Run benchmarks
go test -v -bench=BenchmarkBoschFLEXIDOMEIndoor5100iIR -benchmem

# Run full workflow test
go test -v -run TestBoschFLEXIDOMEIndoor5100iIR_FullWorkflow
```

### Test Coverage

The tests cover the following ONVIF operations:

- ✅ **GetDeviceInformation** - Device identification and firmware info
- ✅ **GetSystemDateAndTime** - System time retrieval
- ✅ **GetCapabilities** - Service capability discovery
- ✅ **Initialize** - Service endpoint initialization
- ✅ **GetProfiles** - Media profile retrieval (4 profiles expected)
- ✅ **GetStreamURI** - RTSP stream URI retrieval for all profiles
- ✅ **GetSnapshotURI** - Snapshot URI retrieval
- ✅ **GetVideoEncoderConfiguration** - Video encoder settings
- ✅ **GetImagingSettings** - Camera imaging parameters
- ✅ **Full Workflow** - Complete operation sequence

### Expected Results for Bosch FLEXIDOME indoor 5100i IR

- **Manufacturer**: Bosch
- **Model**: FLEXIDOME indoor 5100i IR
- **Profiles**: 4 H264 profiles
  - Profile 1: 1920x1080 @ 30fps, 5200 kbps
  - Profile 2: 1536x864
  - Profile 3: 1280x720
  - Profile 4: 512x288
- **Services**: Device, Media, Imaging, Events, Analytics
- **Stream Protocol**: RTSP
- **Snapshot Format**: JPEG
- **Default Imaging Settings**:
  - Brightness: 128.0
  - Color Saturation: 128.0
  - Contrast: 128.0

### Test Without Camera

If environment variables are not set, tests will be automatically skipped:

```bash
go test -v ./...
# Output: SKIP: Skipping test: ONVIF camera credentials not set
```

### Performance Benchmarks

The test suite includes benchmarks for critical operations:

- `BenchmarkBoschFLEXIDOMEIndoor5100iIR_GetDeviceInformation` - Device info retrieval performance
- `BenchmarkBoschFLEXIDOMEIndoor5100iIR_GetStreamURI` - Stream URI retrieval performance

### Adding Tests for Other Camera Models

To add tests for a new camera model:

1. Create a new test file: `<manufacturer>_<model>_test.go`
2. Follow the same pattern as `bosch_flexidome_test.go`
3. Update environment variable names to be model-specific if needed
4. Document expected values and behaviors for the specific model
5. Add README entry with camera-specific details

Example:
```go
// hikvision_ds2cd2xxx_test.go
func TestHikvisionDS2CD_GetDeviceInformation(t *testing.T) {
    // Test implementation
}
```

### Continuous Integration

These tests can be integrated into CI/CD pipelines using secrets management:

```yaml
# GitHub Actions example
- name: Run Camera Integration Tests
  env:
    ONVIF_TEST_ENDPOINT: ${{ secrets.ONVIF_ENDPOINT }}
    ONVIF_TEST_USERNAME: ${{ secrets.ONVIF_USERNAME }}
    ONVIF_TEST_PASSWORD: ${{ secrets.ONVIF_PASSWORD }}
  run: go test -v -run TestBoschFLEXIDOMEIndoor5100iIR
```

### Troubleshooting

**Tests fail with "connection refused":**
- Verify camera IP address and network connectivity
- Check firewall settings
- Ensure camera is powered on

**Tests fail with authentication errors:**
- Verify username and password are correct
- Check if camera requires digest authentication
- Ensure user has appropriate permissions

**Tests fail with unexpected values:**
- Camera firmware may have been updated
- Camera settings may have been changed
- Update expected values in tests to match current configuration

### Notes

- These tests require a physical camera or camera simulator
- Tests modify NO camera settings (read-only operations)
- Some tests may take several seconds due to network communication
- Camera responses may vary based on firmware version and configuration
