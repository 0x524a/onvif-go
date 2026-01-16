# Camera Data Collection Summary
**Date:** January 13, 2026  
**Collection Time:** 13:40 - 13:42 EST  
**Total Cameras:** 8  
**Successful Collections:** 7  
**Failed Collections:** 1

---

## Collection Results

### ✅ Successfully Collected (7 cameras)

| # | Manufacturer | Model | Firmware | IP:Port | Profiles | PTZ | SOAP Calls |
|---|--------------|-------|----------|---------|----------|-----|------------|
| 1 | REOLINK | E1 Zoom | v3.1.0.2649 | 192.168.2.61:8000 | 2 | ✓ | 16 |
| 2 | Bosch | AUTODOME IP starlight 5000i | 7.80.0128 | 192.168.2.57:80 | 3 | ✓ (2 presets) | 21 |
| 3 | AXIS | P3818-PVE | 11.9.60 | 192.168.2.82:80 | 2 | ✗ | 12 |
| 4 | REOLINK | Reolink TrackMix WiFi | v3.0.0.5428 | 192.168.2.236:8000 | 3 | ✓ (1 preset) | 21 |
| 5 | Bosch | FLEXIDOME IP starlight 8000i | 7.70.0126 | 192.168.2.200:80 | 3 | ✗ | 15 |
| 6 | Bosch | FLEXIDOME panoramic 5100i | 9.00.0210 | 192.168.2.24:80 | 16 | ✗ | 47 |
| 7 | AXIS | Q3819-PVE | 11.11.181 | 192.168.2.190:80 | 2 | ✗ | 12 |

### ❌ Failed Collection (1 camera)

| # | Model | IP | Reason |
|---|-------|-----|--------|
| 8 | AXIS P5655-E | 192.168.2.30:80 | **Authentication Failed** - Credentials "service/Service.1234" not authorized |

---

## Detailed Camera Information

### Camera 1: REOLINK E1 Zoom
- **Resolution:** 2048x1536 (Main), 640x480 (Sub)
- **Encoding:** H264
- **Stream:** rtsp://192.168.2.61:554/
- **Features:** PTZ control, Snapshot support
- **Capture File:** `REOLINK_E1_Zoom_v3.1.0.2649_23083101_xmlcapture_20260113-134015.tar.gz` (13KB)

### Camera 2: Bosch AUTODOME IP starlight 5000i  
- **Resolution:** 1536x864 (H264 profiles), JPEG profile
- **Encoding:** H264 @ 30fps, JPEG @ 1fps
- **Stream:** rtsp://192.168.2.57/rtsp_tunnel
- **Features:** PTZ with 2 presets, HTTPS support
- **Capture File:** `Bosch_AUTODOME_IP_starlight_5000i_7.80.0128_xmlcapture_20260113-134024.tar.gz` (13KB)

### Camera 3: AXIS P3818-PVE
- **Resolution:** 1920x960 (H264), 5120x2560 (JPEG)
- **Encoding:** H264 @ 30fps, JPEG @ 30fps
- **Stream:** rtsp://192.168.2.82/onvif-media/media.amp
- **Features:** High-resolution panoramic, Snapshot, Analytics
- **Capture File:** `AXIS_P3818-PVE_11.9.60_xmlcapture_20260113-134032.tar.gz` (11KB)

### Camera 4: REOLINK Reolink TrackMix WiFi
- **Resolution:** 3840x2160 (Main), 896x512 (Sub), 1920x1080 (Autotrack)
- **Encoding:** H264
- **Stream:** rtsp://192.168.2.236:554/Preview_01_*
- **Features:** 4K main stream, Auto-tracking, PTZ with preset, Analytics
- **Capture File:** `REOLINK_Reolink_TrackMix_WiFi_v3.0.0.5428_2509171974_xmlcapture_20260113-134042.tar.gz` (16KB)

### Camera 5: Bosch FLEXIDOME IP starlight 8000i
- **Resolution:** 1536x864
- **Encoding:** H264 @ 30fps, JPEG @ 1fps
- **Stream:** rtsp://192.168.2.200/rtsp_tunnel
- **Features:** HTTPS support, Multiple encoding profiles
- **Capture File:** `Bosch_FLEXIDOME_IP_starlight_8000i_7.70.0126_xmlcapture_20260113-134051.tar.gz` (10KB)

### Camera 6: Bosch FLEXIDOME panoramic 5100i
- **Resolution:** Multiple (1920x1080, 3072x1728, 2112x2112, etc.)
- **Encoding:** H264 @ 30fps
- **Stream:** rtsp://192.168.2.24/rtsp_tunnel
- **Features:** 16 profiles!, Audio, Metadata, Multi-sensor panoramic
- **Notes:** 3 profiles have incomplete configuration (expected for multi-sensor)
- **Capture File:** `Bosch_FLEXIDOME_panoramic_5100i_9.00.0210_xmlcapture_20260113-134100.tar.gz` (20KB)

### Camera 7: AXIS Q3819-PVE
- **Resolution:** 8192x1728 (panoramic)
- **Encoding:** H264 @ 30fps, JPEG @ 30fps  
- **Stream:** rtsp://192.168.2.190/onvif-media/media.amp
- **Features:** Ultra-wide panoramic (8K), Analytics, Dual IPs (192.168.2.190, 169.254.34.187)
- **Capture File:** `AXIS_Q3819-PVE_11.11.181_xmlcapture_20260113-134111.tar.gz` (11KB)

### Camera 8: AXIS P5655-E ❌
- **Status:** Authentication failed
- **Error:** `ter:NotAuthorized - Sender not authorized`
- **Issue:** The credentials "service/Service.1234" do not have access to this camera
- **Action Required:** Different username/password needed for this camera

---

## Capture Statistics

### By Manufacturer
- **Bosch:** 3 cameras (good enterprise ONVIF support)
- **AXIS:** 2 successful, 1 failed auth (3 total)
- **REOLINK:** 2 cameras (consumer-grade ONVIF)

### Profile Support Summary
- **ONVIF Profile T (Streaming):** 7/7 cameras ✓
- **ONVIF Profile G (Recording):** 5/7 cameras
- **ONVIF Profile M (Metadata):** 3/7 cameras
- **PTZ Support:** 3/7 cameras (Bosch AUTODOME, 2 Reolinks)
- **HTTPS Support:** 3/7 cameras (All Bosch)

### Resolution Capabilities
- **4K (3840x2160):** Reolink TrackMix WiFi
- **Panoramic 8K (8192x1728):** AXIS Q3819-PVE
- **Multi-sensor (16 profiles):** Bosch FLEXIDOME panoramic 5100i
- **High-res snapshot (5120x2560):** AXIS P3818-PVE

### SOAP Operations Captured
- **Total SOAP calls:** 144 across 7 cameras
- **Most comprehensive:** Bosch FLEXIDOME panoramic 5100i (47 calls)
- **Average per camera:** ~20 SOAP operations

---

## Files Generated

### XML Capture Archives (testdata/captures/)
```
✓ REOLINK_E1_Zoom_v3.1.0.2649_23083101_xmlcapture_20260113-134015.tar.gz
✓ Bosch_AUTODOME_IP_starlight_5000i_7.80.0128_xmlcapture_20260113-134024.tar.gz
✓ AXIS_P3818-PVE_11.9.60_xmlcapture_20260113-134032.tar.gz
✓ REOLINK_Reolink_TrackMix_WiFi_v3.0.0.5428_2509171974_xmlcapture_20260113-134042.tar.gz
✓ Bosch_FLEXIDOME_IP_starlight_8000i_7.70.0126_xmlcapture_20260113-134051.tar.gz
✓ Bosch_FLEXIDOME_panoramic_5100i_9.00.0210_xmlcapture_20260113-134100.tar.gz
✓ AXIS_Q3819-PVE_11.11.181_xmlcapture_20260113-134111.tar.gz
⚠ unknown_device_xmlcapture_20260113-134119.tar.gz (AXIS P5655-E - auth failed)
```

### JSON Reports (camera-logs/)
Each archive has a corresponding JSON report with detailed diagnostic information.

---

## Data Contents (Per Camera Archive)

Each `.tar.gz` archive contains:
- **metadata.json** - Camera information, firmware, test summary
- **capture_NNN.json** - Metadata for each SOAP operation
- **capture_NNN_request.xml** - Raw SOAP request
- **capture_NNN_response.xml** - Raw SOAP response

### Operations Captured:
1. GetDeviceInformation
2. GetSystemDateAndTime
3. GetCapabilities
4. GetServices
5. GetProfiles
6. GetStreamURI (per profile)
7. GetSnapshotURI (per profile)
8. GetVideoEncoderConfiguration (per profile)
9. GetImagingSettings (per video source)
10. GetStatus (PTZ, if available)
11. GetPresets (PTZ, if available)

---

## Next Steps

### 1. Generate Tests from Captures
```bash
# Build the test generator
go build -o bin/generate-tests ./cmd/generate-tests

# Generate test for each camera
./bin/generate-tests -capture testdata/captures/REOLINK_E1_Zoom_*.tar.gz -output testdata/captures/
./bin/generate-tests -capture testdata/captures/Bosch_AUTODOME_*.tar.gz -output testdata/captures/
./bin/generate-tests -capture testdata/captures/AXIS_P3818_*.tar.gz -output testdata/captures/
./bin/generate-tests -capture testdata/captures/REOLINK_Reolink_TrackMix_*.tar.gz -output testdata/captures/
./bin/generate-tests -capture testdata/captures/Bosch_FLEXIDOME_IP_starlight_8000i_*.tar.gz -output testdata/captures/
./bin/generate-tests -capture testdata/captures/Bosch_FLEXIDOME_panoramic_*.tar.gz -output testdata/captures/
./bin/generate-tests -capture testdata/captures/AXIS_Q3819_*.tar.gz -output testdata/captures/
```

### 2. Run Generated Tests
```bash
# Run all camera tests
go test -v ./testdata/captures/

# Run specific camera test
go test -v ./testdata/captures/ -run TestREOLINK
go test -v ./testdata/captures/ -run TestBosch
go test -v ./testdata/captures/ -run TestAXIS
```

### 3. Resolve AXIS P5655-E Authentication
- Check camera's ONVIF user accounts
- Try admin credentials if different
- Verify ONVIF is enabled for that user

---

## Usage for Test Development

These captures can be used to:
1. **Generate automated regression tests** - Ensure library changes don't break camera compatibility
2. **Test without hardware** - Mock server replays captured responses
3. **Document camera behavior** - Real-world examples of SOAP responses
4. **Debug issues** - Compare expected vs actual SOAP messages
5. **Contribute to project** - Share camera data to improve library support

---

## Summary

✅ **Success Rate:** 87.5% (7/8 cameras)  
✅ **Total SOAP Operations:** 144  
✅ **Manufacturer Coverage:** Bosch (3), AXIS (2), REOLINK (2)  
✅ **Profile Coverage:** T, G, M profiles tested  
✅ **Resolution Range:** 640x480 to 8192x1728  
✅ **Ready for Test Generation:** All 7 successful captures

The collected data provides comprehensive real-world ONVIF responses across consumer (Reolink), professional (AXIS), and enterprise (Bosch) camera brands, with various resolutions, profiles, and capabilities.
