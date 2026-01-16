# ONVIF Debugging Solution

## Problem

The diagnostic utility (`onvif-diagnostics`) logs only parsed JSON results. When XML parsing fails or responses are unexpected, you can't see the raw SOAP XML to debug the issue.

## Solution

The `onvif-diagnostics` utility now includes built-in XML capture functionality via the `-capture-xml` flag. This captures raw SOAP request/response XML and creates a compressed tar.gz archive.

## What Changed

### 1. Enhanced SOAP Client (`soap/soap.go`)

Added debug logging capability:

```go
type Client struct {
    httpClient *http.Client
    username   string
    password   string
    debug      bool                                    // NEW
    logger     func(format string, args ...interface{}) // NEW
}

// New methods:
func (c *Client) SetDebug(enabled bool, logger func(format string, args ...interface{}))
func (c *Client) logDebug(format string, args ...interface{})
```

The SOAP client now logs requests/responses when debug mode is enabled.

### 2. Integrated XML Capture in `onvif-diagnostics`

Location: `cmd/onvif-diagnostics/main.go`

Features:
- Single command for both diagnostic report and XML capture
- `-capture-xml` flag enables raw SOAP traffic capture
- Creates compressed tar.gz archive with camera identification
- Archive naming: `Manufacturer_Model_Firmware_xmlcapture_timestamp.tar.gz`
- Saves to `camera-logs/` directory (same as diagnostic report)
- Automatic cleanup of temporary files

## Usage

### Quick Start

```bash
# Build the utility
go build -o onvif-diagnostics ./cmd/onvif-diagnostics/

# Run with XML capture enabled
./onvif-diagnostics \
  -endpoint "http://192.168.1.164/onvif/device_service" \
  -username "admin" \
  -password "password" \
  -capture-xml \
  -verbose
```

This creates two files:
- `Manufacturer_Model_Firmware_timestamp.json` - Diagnostic report
- `Manufacturer_Model_Firmware_xmlcapture_timestamp.tar.gz` - Raw SOAP XML archive

### Without XML Capture (Faster)

```bash
# Just diagnostic report
./onvif-diagnostics \
  -endpoint "http://192.168.1.164/onvif/device_service" \
  -username "admin" \
  -password "password" \
  -verbose
```

### Extract and Analyze XML

```bash
# Extract the archive
tar -xzf camera-logs/Camera_Model_xmlcapture_timestamp.tar.gz -C /tmp/xml-debug

# View files (now with operation names)
ls /tmp/xml-debug/
# capture_001_GetDeviceInformation.json
# capture_001_GetDeviceInformation_request.xml
# capture_001_GetDeviceInformation_response.xml
# capture_002_GetSystemDateAndTime.json
# ...
```

## Workflow

### 1. Run Diagnostic with XML Capture

```bash
./onvif-diagnostics \
  -endpoint "http://camera-ip/onvif/device_service" \
  -username "user" \
  -password "pass" \
  -capture-xml \
  -verbose
```

This generates both:
- JSON diagnostic report
- tar.gz XML capture archive

### 2. Review Diagnostic Report

Check the JSON file for errors:
```bash
cat camera-logs/Camera_Model_Firmware_timestamp.json | jq '.errors'
```

### 3. Analyze Raw XML (if needed)

Extract and inspect the XML archive:
```bash
tar -xzf camera-logs/Camera_Model_xmlcapture_timestamp.tar.gz -C /tmp/xml-debug
```

### 3. Analyze Raw XML

```bash
# Extract the archive
tar -xzf camera-logs/Camera_Model_xmlcapture_timestamp.tar.gz -C /tmp/xml-debug

# View specific operation (now easier to find)
cat /tmp/xml-debug/capture_*_GetCapabilities_response.xml

# Search for errors
grep "Fault" /tmp/xml-debug/capture_*_response.xml

# Pretty-print (XML is already formatted with indentation)
cat /tmp/xml-debug/capture_001_GetDeviceInformation_response.xml
```

## Example: Debugging AXIS Q3626-VE Localhost Issue

### Problem (from diagnostic report)

```json
{
  "operation": "GetProfiles",
  "error": "Post \"http://127.0.0.1/onvif/services\": EOF"
}
```

### Capture XML

```bash
### Capture XML

```bash
./onvif-diagnostics \
  -endpoint "http://192.168.1.164/onvif/device_service" \
  -username "admin" \
  -password "password" \
  -capture-xml \
  -verbose
```

Result: 
- `camera-logs/AXIS_Q3626-VE_12.6.104_20251110-120000.json`
- `camera-logs/AXIS_Q3626-VE_12.6.104_xmlcapture_20251110-120000.tar.gz`
```

Result: `camera-logs/AXIS_Q3626-VE_12.6.104_xmlcapture_20251110-120000.tar.gz`

### Analyze Response

```bash
tar -xzf camera-logs/AXIS_Q3626-VE_12.6.104_xmlcapture_20251110-120000.tar.gz
cat capture_*_GetCapabilities_response.xml | grep XAddr
```

Shows:

```xml
<Media>
  <XAddr>http://127.0.0.1/onvif/services</XAddr>
</Media>
```

### Root Cause

Camera returns `127.0.0.1` instead of actual IP `192.168.1.164`, causing client to connect to localhost.

### Solution Required

Client needs to rewrite localhost addresses:

```go
if strings.Contains(xAddr, "127.0.0.1") || strings.Contains(xAddr, "localhost") {
    // Replace with actual camera IP from original endpoint
}
```

## Example: Debugging Bosch Panoramic "Incomplete Configuration"

### Problem (from diagnostic report)

```json
{
  "operation": "GetStreamURI[9]",
  "error": "ter:IncompleteConfiguration - Configuration not complete"
}
```

### Capture XML

```bash
### Capture XML

```bash
./onvif-diagnostics \
  -endpoint "http://192.168.2.24/onvif/device_service" \
  -username "service" \
  -password "Service.1234" \
  -capture-xml \
  -verbose
```

Result:
- `camera-logs/Bosch_FLEXIDOME_panoramic_5100i_9.00.0210_20251110.json`
- `camera-logs/Bosch_FLEXIDOME_panoramic_5100i_9.00.0210_xmlcapture_20251110.tar.gz`
```

### Analyze Response

```bash
tar -xzf camera-logs/Bosch_FLEXIDOME_panoramic_5100i_*_xmlcapture_*.tar.gz
# Look for GetStreamUri operation (easy to find by name)
cat capture_*_GetStreamUri_response.xml
```

Result:

```xml
<SOAP-ENV:Fault>
  <SOAP-ENV:Code>
    <SOAP-ENV:Subcode>
      <SOAP-ENV:Value>ter:IncompleteConfiguration</SOAP-ENV:Value>
    </SOAP-ENV:Subcode>
  </SOAP-ENV:Code>
  <SOAP-ENV:Reason>
    <SOAP-ENV:Text>Configuration not complete</SOAP-ENV:Text>
  </SOAP-ENV:Reason>
</SOAP-ENV:Fault>
```

### Root Cause

Profile 9 has `VideoEncoderConfiguration: null` in the profiles response. Can't get stream URI for profile without video encoder.

### Solution

Skip GetStreamURI for profiles without VideoEncoderConfiguration:

```go
if profile.VideoEncoderConfiguration == nil {
    // Skip - this is audio-only or metadata-only profile
    continue
}
```

## Files Created

### SOAP Client Enhancement
- `soap/soap.go` - Added debug logging capability

### Diagnostic Utility Enhancement
- `cmd/onvif-diagnostics/main.go` - Added XML capture functionality with `-capture-xml` flag

## Output Organization

All debugging files are saved to the same `camera-logs/` directory:

```
camera-logs/
├── Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_20251107-193656.json      # Diagnostic report
├── Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_xmlcapture_20251110.tar.gz # XML capture archive
├── AXIS_Q3626-VE_12.6.104_20251108-212157.json
├── AXIS_Q3626-VE_12.6.104_xmlcapture_20251108-213000.tar.gz
└── Bosch_FLEXIDOME_panoramic_5100i_9.00.0210_20251107-195636.json
```

### Archive Contents

Each tar.gz archive contains the captured XML files with descriptive operation names:

```bash
$ tar -tzf camera-logs/Bosch_FLEXIDOME_indoor_5100i_IR_*_xmlcapture_*.tar.gz
capture_001_GetDeviceInformation.json
capture_001_GetDeviceInformation_request.xml
capture_001_GetDeviceInformation_response.xml
capture_002_GetSystemDateAndTime.json
capture_002_GetSystemDateAndTime_request.xml
capture_002_GetSystemDateAndTime_response.xml
capture_003_GetCapabilities.json
capture_003_GetCapabilities_request.xml
capture_003_GetCapabilities_response.xml
...
```

Each file is named with both a sequence number and the SOAP operation name for easy identification.

## Benefits

1. **Complete Visibility**: See exact SOAP XML sent/received
2. **Namespace Debugging**: Identify namespace mismatches
3. **Fault Analysis**: See detailed SOAP fault information
4. **Comparison**: Compare working vs failing cameras
5. **Easy Sharing**: Compressed archives (< 10KB) easy to share via email
6. **Organized**: All camera logs in one directory with consistent naming
7. **Privacy**: Review and sanitize XML before sharing archives

## Next Steps

When you encounter errors in the diagnostic report:

1. ✅ Run `onvif-diagnostics` to identify which operations fail
2. ✅ Re-run with `-capture-xml` flag to capture raw XML
3. ✅ Extract and analyze the tar.gz archive
4. ✅ Share both files (JSON report + tar.gz archive) for debugging assistance

## Command-Line Flags

```
-endpoint string
    ONVIF device endpoint (required)

-username string
    Username for authentication (required)

-password string
    Password for authentication (required)

-output string
    Output directory (default: "./camera-logs")

-timeout int
    Request timeout in seconds (default: 30)

-verbose
    Enable verbose output

-capture-xml
    Capture raw SOAP XML traffic and create tar.gz archive
```

## Output Structure

### Before (separate files):
```
xml-captures/
└── 20251110-095000/
    ├── capture_001.json
    ├── capture_001_request.xml
    ├── capture_001_response.xml
    └── ...
```

### Now (compressed archives):
```
camera-logs/
├── Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_20251107-193656.json
├── Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_xmlcapture_20251110-115830.tar.gz (5KB)
├── AXIS_Q3626-VE_12.6.104_20251108-212157.json
└── AXIS_Q3626-VE_12.6.104_xmlcapture_20251110-120000.tar.gz (3KB)
```

## Tips

- Use `-operation` to test specific failing operations
- Check response XML for `<Fault>` elements
- Compare namespace prefixes (tds, trt, tt, etc.)
- Look for XAddr values in capabilities response
- Verify authentication headers in request XML
