# ONVIF Camera Diagnostic Utility

A comprehensive diagnostic tool for collecting detailed information from ONVIF cameras. This utility helps analyze camera capabilities, troubleshoot issues, and generate reports for creating camera-specific tests.

## Features

✅ **Comprehensive Testing** - Tests all major ONVIF operations:
- Device information and capabilities
- Media profiles and streaming
- Video encoder configurations
- Imaging settings
- PTZ status and presets (if available)
- System date/time

✅ **Detailed Reporting** - Generates JSON reports with:
- All successful operations with response data
- Failed operations with error details
- Response times for performance analysis
- Structured data ready for test generation

✅ **Easy to Use** - Simple command-line interface with minimal requirements

✅ **XML Debugging** - For detailed debugging, see the companion `onvif-xml-capture` utility that captures raw SOAP XML

✅ **Helpful for**:
- Creating camera-specific integration tests
- Troubleshooting ONVIF compatibility issues
- Analyzing camera capabilities
- Debugging connection problems
- Documenting camera configurations

## Installation

### Option 1: Build from source
```bash
cd /path/to/onvif-go
go build -o onvif-diagnostics ./cmd/onvif-diagnostics/
```

### Option 2: Install globally
```bash
go install ./cmd/onvif-diagnostics
```

## Usage

### Basic Usage
```bash
./onvif-diagnostics \
  -endpoint "http://192.168.1.201/onvif/device_service" \
  -username "service" \
  -password "Service.1234"
```

### With XML Capture (for debugging)
```bash
./onvif-diagnostics \
  -endpoint "http://192.168.1.201/onvif/device_service" \
  -username "service" \
  -password "Service.1234" \
  -capture-xml \
  -verbose
```

This creates two files:
- `Manufacturer_Model_Firmware_timestamp.json` - Diagnostic report
- `Manufacturer_Model_Firmware_xmlcapture_timestamp.tar.gz` - Raw SOAP XML archive

### Verbose Output
```bash
./onvif-diagnostics \
  -endpoint "http://192.168.1.201/onvif/device_service" \
  -username "service" \
  -password "Service.1234" \
  -verbose
```

### Capture Raw SOAP XML
```bash
./onvif-diagnostics \
  -endpoint "http://192.168.1.201/onvif/device_service" \
  -username "service" \
  -password "Service.1234" \
  -capture-xml
```

Enables XML traffic capture and creates a compressed tar.gz archive containing all SOAP request/response pairs. Useful for debugging XML parsing issues or analyzing camera behavior.

The archive contains:
- `capture_001_GetDeviceInformation.json` - Request/response metadata with operation name
- `capture_001_GetDeviceInformation_request.xml` - Formatted SOAP request
- `capture_001_GetDeviceInformation_response.xml` - Formatted SOAP response
- `capture_002_GetSystemDateAndTime.json` - Next operation metadata
- ... (one set per SOAP operation, named by operation type)

Each file is named with the SOAP operation (e.g., GetDeviceInformation, GetProfiles) for easy identification.

Extract the archive:
```bash
tar -xzf camera-logs/Camera_Model_xmlcapture_timestamp.tar.gz
```

### Custom Output Directory
```bash
./onvif-diagnostics \
  -endpoint "http://192.168.1.201/onvif/device_service" \
  -username "service" \
  -password "Service.1234" \
  -output ./my-camera-reports
```

### All Options
```
Usage of ./onvif-diagnostics:
  -endpoint string
        ONVIF device endpoint (e.g., http://192.168.1.201/onvif/device_service)
  -username string
        ONVIF username
  -password string
        ONVIF password
  -output string
        Output directory for logs (default "./camera-logs")
  -timeout int
        Request timeout in seconds (default 30)
  -verbose
        Verbose output
  -include-raw
        Include raw SOAP responses (increases file size)
```

## Example Output

```
ONVIF Camera Diagnostic Utility v1.0.0
========================================

Starting diagnostic collection...

→ 1. Getting device information...
  ✓ Manufacturer: Bosch, Model: FLEXIDOME indoor 5100i IR
→ 2. Getting system date and time...
  ✓ Retrieved
→ 3. Getting capabilities...
  ✓ Services: Device, Media, Imaging, Events, Analytics
→ 4. Discovering service endpoints...
  ✓ Service endpoints discovered
→ 5. Getting media profiles...
  ✓ Found 4 profile(s)
→ 6. Getting stream URIs for all profiles...
  ✓ Retrieved 4/4 stream URIs
→ 7. Getting snapshot URIs for all profiles...
  ✓ Retrieved 4/4 snapshot URIs
→ 8. Getting video encoder configurations...
  ✓ Retrieved 4/4 video encoder configs
→ 9. Getting imaging settings...
  ✓ Retrieved 1/1 imaging settings
→ 10. Getting PTZ status...
  ℹ No PTZ configurations found
→ 11. Getting PTZ presets...
  ℹ No PTZ configurations found
→ Saving diagnostic report...

========================================
✓ Diagnostic collection complete!
  Report saved to: camera-logs/Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_20251107-193656.json
  Total errors: 0

  Device: Bosch FLEXIDOME indoor 5100i IR
  Firmware: 8.71.0066
  Profiles: 4

Please share this file for analysis and test creation.
========================================
```

## Report Structure

The generated JSON report includes:

```json
{
  "timestamp": "2025-11-07T19:36:56Z",
  "utility_version": "1.0.0",
  "connection_info": {
    "endpoint": "http://192.168.1.201/onvif/device_service",
    "username": "service",
    "test_date": "2025-11-07"
  },
  "device_info": {
    "success": true,
    "data": {
      "manufacturer": "Bosch",
      "model": "FLEXIDOME indoor 5100i IR",
      "firmware_version": "8.71.0066",
      "serial_number": "404754734001050102",
      "hardware_id": "F000B543"
    },
    "response_time": "21.5ms"
  },
  "profiles": {
    "success": true,
    "count": 4,
    "data": [ /* profile details */ ]
  },
  "stream_uris": [ /* stream URI results for each profile */ ],
  "errors": [ /* any errors encountered */ ]
}
```

## Use Cases

### 1. Creating Camera-Specific Tests
Run the diagnostic on your camera and share the JSON file. The report contains all the information needed to create comprehensive integration tests.

### 2. Troubleshooting Connection Issues
If your camera isn't working, run diagnostics to see exactly which operations fail and what error messages are returned.

### 3. Comparing Cameras
Run diagnostics on multiple cameras to compare capabilities, response times, and compatibility.

### 4. Documentation
Generate detailed reports of camera configurations for documentation purposes.

## Interpreting Results

### Success Indicators
- ✓ Green checkmarks indicate successful operations
- Response times help identify performance issues
- High success rates indicate good compatibility

### Error Indicators
- ✗ Red X marks indicate failed operations
- ℹ Info symbols indicate optional features not available
- Check the `errors` array in JSON for detailed error messages

### Common Issues

**All operations fail:**
- Check network connectivity
- Verify endpoint URL is correct
- Ensure camera is powered on

**Authentication errors:**
- Verify username and password
- Check user permissions on camera

**Some profiles fail:**
- Camera may have different capabilities per profile
- Some operations may not be supported by all profiles

**Timeout errors:**
- Increase timeout with `-timeout 60`
- Check network latency
- Verify camera is responding

## Sharing Reports

When sharing diagnostic reports:

1. **Anonymize if needed** - The report includes:
   - IP addresses (in endpoint)
   - Usernames (not passwords)
   - Serial numbers
   
2. **What to share**:
   - The complete JSON file
   - Any console output showing errors
   - Camera model and firmware version
   
3. **Where to share**:
   - GitHub Issues
   - Email for analysis
   - Pull request descriptions

## Advanced Usage

### Batch Testing Multiple Cameras
Create a script to test multiple cameras:

```bash
#!/bin/bash
cameras=(
  "192.168.1.201:service:password1"
  "192.168.1.202:admin:password2"
  "192.168.1.203:user:password3"
)

for camera in "${cameras[@]}"; do
  IFS=':' read -r ip user pass <<< "$camera"
  echo "Testing camera at $ip..."
  ./onvif-diagnostics \
    -endpoint "http://$ip/onvif/device_service" \
    -username "$user" \
    -password "$pass"
done
```

### Automated Testing
Include in CI/CD pipelines:

```yaml
- name: Run ONVIF Diagnostics
  run: |
    ./onvif-diagnostics \
      -endpoint "${{ secrets.CAMERA_ENDPOINT }}" \
      -username "${{ secrets.CAMERA_USERNAME }}" \
      -password "${{ secrets.CAMERA_PASSWORD }}" \
      -output ./reports
      
- name: Upload Diagnostic Reports
  uses: actions/upload-artifact@v3
  with:
    name: camera-diagnostics
    path: ./reports/
```

## Development

### Adding New Tests

To add new diagnostic tests, edit `cmd/onvif-diagnostics/main.go`:

1. Create a new test function following the pattern:
```go
func testNewOperation(ctx context.Context, client *onvif.Client, report *CameraReport) *NewOperationResult {
    // Implementation
}
```

2. Add result struct to store data
3. Call the test in main()
4. Update report structure

### Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o onvif-diagnostics-linux ./cmd/onvif-diagnostics/

# Windows
GOOS=windows GOARCH=amd64 go build -o onvif-diagnostics.exe ./cmd/onvif-diagnostics/

# macOS ARM
GOOS=darwin GOARCH=arm64 go build -o onvif-diagnostics-mac-arm ./cmd/onvif-diagnostics/
```

## License

Same as parent project.

## Support

For issues or questions:
1. Run diagnostics with `-verbose` flag
2. Share the generated JSON report
3. **For XML parsing issues**: Use `onvif-xml-capture` utility to capture raw SOAP XML
4. Open a GitHub issue with the report attached

## Related Tools

- **onvif-xml-capture** - Captures raw SOAP XML requests/responses for detailed debugging
  - Location: `cmd/onvif-xml-capture/`
  - Use when: Diagnostic report shows errors and you need to see raw XML
  - See: `XML_DEBUGGING_SOLUTION.md` for complete guide

