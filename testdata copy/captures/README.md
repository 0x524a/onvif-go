# Camera Test Framework

This directory contains camera-specific tests generated from real camera XML captures. These tests ensure the ONVIF client works correctly with various camera models and prevents regressions when making changes.

## Overview

The test framework consists of:

1. **Captured XML Archives** (`*.tar.gz`) - Real SOAP XML request/response pairs from cameras
2. **Generated Tests** (`*_test.go`) - Automated tests that replay captures through a mock server
3. **Test Generator** (`cmd/generate-tests`) - Tool to create tests from captures
4. **Mock Server** (`testing/mock_server.go`) - HTTP server that replays captured responses

## Benefits

✅ **Test Without Hardware** - Run ONVIF tests without needing physical cameras  
✅ **Prevent Regressions** - Catch breaking changes before they affect real deployments  
✅ **Camera Coverage** - Test against multiple camera manufacturers and models  
✅ **Fast Feedback** - Tests complete in milliseconds vs. minutes with real cameras  
✅ **CI/CD Ready** - Automated tests that can run in continuous integration

## Running Tests

### Run All Camera Tests

```bash
go test -v ./testdata/captures/
```

### Run Specific Camera

```bash
go test -v ./testdata/captures/ -run TestBosch
```

### Run from Project Root

```bash
go test -v ./...
```

## Adding New Camera Tests

### 1. Capture Camera XML

First, capture SOAP XML from your camera:

```bash
# Run diagnostic with XML capture
./onvif-diagnostics \
  -endpoint "http://camera-ip/onvif/device_service" \
  -username "user" \
  -password "pass" \
  -capture-xml \
  -verbose
```

This creates an archive like:
```
camera-logs/Manufacturer_Model_Firmware_xmlcapture_timestamp.tar.gz
```

### 2. Copy to testdata/captures

```bash
cp camera-logs/Manufacturer_Model_*_xmlcapture_*.tar.gz testdata/captures/
```

### 3. Generate Test

```bash
./generate-tests \
  -capture testdata/captures/Manufacturer_Model_*_xmlcapture_*.tar.gz \
  -output testdata/captures/
```

This generates:
```
testdata/captures/manufacturer_model_firmware_test.go
```

### 4. Run the Test

```bash
go test -v ./testdata/captures/ -run TestManufacturerModel
```

## Example Workflow

Complete example adding an AXIS camera:

```bash
# 1. Capture from camera
./onvif-diagnostics \
  -endpoint "http://192.168.1.100/onvif/device_service" \
  -username "root" \
  -password "pass" \
  -capture-xml

# Output: camera-logs/AXIS_Q3626-VE_12.6.104_xmlcapture_20251110-130000.tar.gz

# 2. Copy to testdata
cp camera-logs/AXIS_Q3626-VE_12.6.104_xmlcapture_20251110-130000.tar.gz testdata/captures/

# 3. Generate test
./generate-tests \
  -capture testdata/captures/AXIS_Q3626-VE_12.6.104_xmlcapture_20251110-130000.tar.gz \
  -output testdata/captures/

# Output: testdata/captures/axis_q3626-ve_12.6.104_test.go

# 4. Run test
go test -v ./testdata/captures/ -run TestAXIS
```

## Directory Structure

```
testdata/captures/
├── README.md                                                      # This file
├── Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_xmlcapture_*.tar.gz # Capture archive
├── bosch_flexidome_indoor_5100i_ir_8.71.0066_test.go             # Generated test
├── AXIS_Q3626-VE_12.6.104_xmlcapture_*.tar.gz                    # Another camera
└── axis_q3626-ve_12.6.104_test.go                                # Its test
```

## How It Works

### Capture Archive Contents

Each `*.tar.gz` archive contains:

```
capture_001.json                        # Request/response metadata
capture_001_request.xml                 # SOAP request
capture_001_response.xml                # SOAP response
capture_002.json
capture_002_request.xml
capture_002_response.xml
...
```

### Mock Server

The test framework includes a mock HTTP server that:

1. Loads all captured exchanges from the archive
2. Extracts SOAP operation names from requests (GetDeviceInformation, GetProfiles, etc.)
3. Matches incoming test requests to captured responses by operation name
4. Returns the exact SOAP response the real camera sent

This allows the ONVIF client to interact with "virtual cameras" that behave exactly like the real ones.

### Generated Test

Each generated test:

1. Creates a mock server from the capture archive
2. Creates an ONVIF client pointing to the mock server
3. Runs common ONVIF operations (GetDeviceInformation, GetProfiles, etc.)
4. Validates responses match expected values

## Customizing Tests

### Adding Custom Assertions

Edit the generated test file to add camera-specific validations:

```go
t.Run("GetDeviceInformation", func(t *testing.T) {
    info, err := client.GetDeviceInformation(ctx)
    if err != nil {
        t.Errorf("GetDeviceInformation failed: %v", err)
        return
    }

    // Add custom assertions
    if info.Manufacturer != "Bosch" {
        t.Errorf("Expected Bosch, got %s", info.Manufacturer)
    }
    if !strings.Contains(info.Model, "FLEXIDOME") {
        t.Errorf("Expected FLEXIDOME model, got %s", info.Model)
    }
})
```

### Testing Specific Operations

Add tests for camera-specific features:

```go
t.Run("PTZPresets", func(t *testing.T) {
    // Only for PTZ cameras
    presets, err := client.GetPresets(ctx, "profile_token")
    if err != nil {
        t.Errorf("GetPresets failed: %v", err)
        return
    }
    
    if len(presets) == 0 {
        t.Error("Expected at least one preset")
    }
})
```

## Troubleshooting

### Test Fails: "No matching capture found"

The mock server couldn't find a captured response for the operation.

**Solution**: Re-capture from the camera to include all operations.

### Test Fails: Unexpected Response

The client is receiving the wrong SOAP response.

**Solution**: Check that operation names match. The mock server matches by SOAP operation name extracted from the `<Body>` element.

### Archive Not Found

```
Failed to create mock server: failed to open archive: no such file or directory
```

**Solution**: Ensure the capture archive is in `testdata/captures/` directory.

## Maintenance

### Updating Captures

When camera firmware changes:

1. Re-run diagnostics with `-capture-xml`
2. Replace old capture archive
3. Regenerate test (or manually update paths)
4. Re-run tests to verify

### Cleaning Up

Remove old captures and tests:

```bash
rm testdata/captures/old_camera_*.tar.gz
rm testdata/captures/old_camera_test.go
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Camera Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run Camera Tests
        run: go test -v ./testdata/captures/
```

### Benefits in CI

- Tests run on every commit
- Prevents merging code that breaks camera compatibility
- No need for test cameras in CI environment
- Fast execution (< 1 second for all cameras)

## Best Practices

1. **Capture from latest firmware** - Use up-to-date camera firmware
2. **Include all operations** - Run full diagnostic to capture all SOAP operations
3. **Document camera models** - Add comments in tests noting camera specifics
4. **Version control captures** - Commit `.tar.gz` files to track camera behavior over time
5. **Test before changes** - Run tests before making client changes to establish baseline
6. **Test after changes** - Verify all camera tests pass after modifications

## Related Tools

- **onvif-diagnostics** - Captures XML from cameras (`cmd/onvif-diagnostics`)
- **generate-tests** - Creates tests from captures (`cmd/generate-tests`)
- **mock_server** - Test server implementation (`testing/mock_server.go`)

## Support

For issues or questions:

1. Check that capture archive is valid (can extract with `tar -xzf`)
2. Verify test file package is `onvif_test`
3. Run with `-v` flag for verbose output
4. Check `testing/mock_server.go` logs for operation matching details
