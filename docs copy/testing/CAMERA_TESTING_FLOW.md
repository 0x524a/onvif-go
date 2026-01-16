# Camera Testing Flow - How to Add Your Camera Tests

This guide explains how public users can contribute camera-specific tests to onvif-go by capturing their camera's SOAP responses and generating automated tests.

## 🎯 Overview

The testing flow consists of:

1. **Capture** - Run diagnostics to collect SOAP XML from your camera
2. **Archive** - Generated tar.gz file with all SOAP exchanges
3. **Contribute** - Submit capture as test data via Pull Request
4. **Generate** - Tool auto-creates test file from capture
5. **Verify** - Tests validate against your camera

## 📋 Prerequisites

- Access to an ONVIF-compatible camera
- Camera credentials (username/password)
- onvif-go tools (diagnostics and test generator)
- Git and GitHub account (for contribution)

## 🔄 Step-by-Step Flow

### Step 1: Build Required Tools

```bash
# Clone the repository
git clone https://github.com/0x524a/onvif-go.git
cd onvif-go

# Build the diagnostics tool
go build -o onvif-diagnostics ./cmd/onvif-diagnostics

# Build the test generator
go build -o generate-tests ./cmd/generate-tests
```

### Step 2: Run Camera Diagnostics

The `onvif-diagnostics` tool connects to your camera and captures all SOAP exchanges:

```bash
./onvif-diagnostics \
  -endpoint "http://192.168.1.100/onvif/device_service" \
  -username "admin" \
  -password "password123" \
  -capture-xml \
  -verbose
```

**Parameters:**
- `-endpoint`: Your camera's ONVIF device service URL
- `-username`: Camera authentication username
- `-password`: Camera authentication password
- `-capture-xml`: Capture raw SOAP XML (required for tests)
- `-verbose`: Show detailed output

**Output:**
```
camera-logs/
├── Manufacturer_Model_Firmware_timestamp.json
└── Manufacturer_Model_Firmware_xmlcapture_timestamp.tar.gz  ← THIS is the capture
```

### Step 3: Review Captured Data

Inspect what was captured:

```bash
# List archive contents
tar -tzf camera-logs/Manufacturer_Model_*_xmlcapture_*.tar.gz | head -20

# Extract to review (optional)
tar -xzf camera-logs/Manufacturer_Model_*_xmlcapture_*.tar.gz -C /tmp
```

**Expected contents:**
```
capture_001.json                           # Metadata for 1st operation
capture_001_request.xml                    # SOAP request
capture_001_response.xml                   # SOAP response
capture_002.json                           # Metadata for 2nd operation
capture_002_request.xml
capture_002_response.xml
... (one set per ONVIF operation)
```

### Step 4: Copy to testdata/captures

```bash
# Copy archive to test data directory
cp camera-logs/Manufacturer_Model_*_xmlcapture_*.tar.gz testdata/captures/
```

### Step 5: Generate Test File

The `generate-tests` tool creates a Go test file from the capture:

```bash
./generate-tests \
  -capture testdata/captures/Manufacturer_Model_*_xmlcapture_*.tar.gz \
  -output testdata/captures/
```

**Output:**
```
testdata/captures/manufacturer_model_firmware_test.go
```

### Step 6: Run the Generated Test

Verify the test works with your camera data:

```bash
# Run your camera's test
go test -v ./testdata/captures/ -run TestManufacturer

# Or run all camera tests
go test -v ./testdata/captures/
```

**Expected output:**
```
=== RUN   TestManufacturer
    --- Camera: Manufacturer_Model_Firmware
    mock_server_test.go:XX: Operations tested: 15
    ✓ Device Information captured
    ✓ Profiles captured
    ✓ Stream URIs captured
    --- PASS: TestManufacturer (0.25s)
PASS
ok      github.com/0x524a/onvif-go/testdata/captures  0.25s
```

### Step 7: Customize Test (Optional)

Edit the generated test file to add camera-specific validations:

```go
// In testdata/captures/manufacturer_model_firmware_test.go

t.Run("CustomValidations", func(t *testing.T) {
    info, err := client.GetDeviceInformation(ctx)
    if err != nil {
        t.Fatalf("GetDeviceInformation failed: %v", err)
    }
    
    // Add your specific assertions
    if !strings.Contains(info.Manufacturer, "YourManufacturer") {
        t.Errorf("Expected manufacturer, got %s", info.Manufacturer)
    }
    
    if !strings.Contains(info.Model, "YourModel") {
        t.Errorf("Expected model, got %s", info.Model)
    }
})
```

### Step 8: Submit Pull Request

Contribute your camera test to the project:

```bash
# Create a branch
git checkout -b add/camera-tests-manufacturer-model

# Stage the test files
git add testdata/captures/
git add camera-logs/  # Optional: include diagnostic report too

# Commit with descriptive message
git commit -m "test: add Manufacturer Model camera tests

- Captured SOAP XML from firmware version X.Y.Z
- Generated test validates all ONVIF services
- Tests Device, Media, PTZ, and Imaging operations"

# Push to your fork
git push origin add/camera-tests-manufacturer-model
```

Then create a Pull Request on GitHub with:
- **Title:** `test: add Manufacturer Model camera tests`
- **Description:**
  ```
  ## Camera Details
  - Manufacturer: [Name]
  - Model: [Model]
  - Firmware: [Version]
  - ONVIF Version: [Version, if known]
  
  ## Features Tested
  - Device management
  - Media profiles and streaming
  - PTZ control (if applicable)
  - Imaging settings (if applicable)
  
  ## Files
  - Capture: `testdata/captures/Manufacturer_Model_Firmware_xmlcapture_*.tar.gz`
  - Test: `testdata/captures/manufacturer_model_firmware_test.go`
  
  Resolves #[issue-number] (if applicable)
  ```

## 📊 What Gets Tested

Each camera test automatically validates:

✅ **Device Management**
- GetDeviceInformation
- GetCapabilities
- GetSystemDateAndTime

✅ **Media Services**
- GetProfiles
- GetStreamUri
- GetSnapshotUri
- GetVideoEncoderConfiguration

✅ **PTZ Control** (if available)
- GetPTZStatus
- GetPresets
- GetTurns

✅ **Imaging** (if available)
- GetImagingSettings
- GetOptions

✅ **Response Validation**
- Correct structure
- Required fields populated
- Proper data types
- No parsing errors

## 🎥 Example Workflow

Complete example adding a **Hikvision DS-2CD2143G2-I** camera:

```bash
# 1. Build tools
cd onvif-go
go build -o onvif-diagnostics ./cmd/onvif-diagnostics
go build -o generate-tests ./cmd/generate-tests

# 2. Capture from camera
./onvif-diagnostics \
  -endpoint "http://192.168.1.50/onvif/device_service" \
  -username "admin" \
  -password "Hikvision123" \
  -capture-xml \
  -verbose

# Output: camera-logs/Hikvision_DS-2CD2143G2-I_V5.5.61_xmlcapture_20251117-143022.tar.gz

# 3. Copy to testdata
cp camera-logs/Hikvision_DS-2CD2143G2-I_V5.5.61_xmlcapture_*.tar.gz testdata/captures/

# 4. Generate test
./generate-tests \
  -capture testdata/captures/Hikvision_DS-2CD2143G2-I_V5.5.61_xmlcapture_*.tar.gz \
  -output testdata/captures/

# Output: testdata/captures/hikvision_ds-2cd2143g2-i_v5.5.61_test.go

# 5. Run test
go test -v ./testdata/captures/ -run TestHikvision

# Output: PASS ✓

# 6. Submit PR
git checkout -b add/hikvision-ds-2cd2143g2-i-tests
git add testdata/captures/hikvision_ds-2cd2143g2-i_v5.5.61_test.go
git add testdata/captures/Hikvision_DS-2CD2143G2-I_V5.5.61_xmlcapture_*.tar.gz
git commit -m "test: add Hikvision DS-2CD2143G2-I camera tests (v5.5.61)"
git push origin add/hikvision-ds-2cd2143g2-i-tests
```

Then open PR on GitHub!

## 🛠️ Troubleshooting

### Diagnostics Tool Can't Connect

```
Error: dial tcp 192.168.1.100:80: connect: connection refused
```

**Solutions:**
- Verify camera IP address is correct
- Check camera is online: `ping 192.168.1.100`
- Ensure camera ONVIF port (typically 80 or 8080)
- Try full URL: `-endpoint "http://192.168.1.100:8080/onvif/device_service"`

### Authentication Failed

```
Error: 401 Unauthorized - invalid credentials
```

**Solutions:**
- Verify username and password
- Try single quotes for special characters: `-password 'pass!word'`
- Check if camera requires different username format
- Verify camera admin access level is enabled

### No XML Captured

```
diagnostics: Error: -capture-xml flag requires -endpoint
```

**Solution:** Use all required flags:
```bash
./onvif-diagnostics \
  -endpoint "..." \
  -username "..." \
  -password "..." \
  -capture-xml
```

### Test Generation Fails

```
Error: failed to open archive
```

**Solutions:**
- Verify archive file exists and is valid
- Check filename matches pattern: `*_xmlcapture_*.tar.gz`
- Ensure archive is in `testdata/captures/` directory
- Try extracting manually: `tar -tzf file.tar.gz`

### Generated Test Won't Compile

```
error: undefined: t
```

**Solution:** Ensure generated file is in `testdata/captures/` and has `_test.go` suffix.

## 📈 Benefits of Contributing

✅ **Improve Library** - Help catch bugs with real camera data  
✅ **Prevent Regressions** - Ensure future changes don't break your camera  
✅ **Community** - Help other users with same camera  
✅ **Recognition** - Your camera is now tested in CI/CD  
✅ **Better Support** - Maintainers understand your camera better  

## 🔒 Privacy & Security

**What's in the capture:**
- SOAP XML request/response pairs
- Device information (manufacturer, model, firmware)
- Configuration data (profiles, presets, etc.)

**What's NOT included:**
- Video streams
- Actual video data
- Personal information
- Credentials (unless you include them - they're stripped by default)

**Before submitting:**
1. Review captured XML for sensitive data
2. Remove any custom configurations if desired
3. Ensure camera is on a test network, not production

## 📚 Related Documentation

- **[onvif-diagnostics README](cmd/onvif-diagnostics/README.md)** - Detailed tool usage
- **[Camera Test Framework](testdata/captures/README.md)** - How tests work
- **[Contributing Guide](CONTRIBUTING.md)** - General contribution guidelines
- **[QUICKSTART](QUICKSTART.md)** - Library basics

## 💬 Getting Help

- **Questions?** Open an issue on GitHub
- **Need guidance?** Check existing camera tests: `testdata/captures/*_test.go`
- **Found a bug?** Report it with your camera model and firmware version

---

**Thank you for contributing! Your camera tests help make onvif-go better for everyone.** 🎉
