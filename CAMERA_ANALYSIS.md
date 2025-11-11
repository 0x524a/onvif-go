# ONVIF Camera Analysis Report

Generated: November 7, 2025

## Executive Summary

Analysis of 5 ONVIF-compliant cameras from 3 manufacturers (REOLINK, AXIS, Bosch) reveals diverse implementations and capabilities. All cameras successfully responded to ONVIF commands with varying feature sets.

---

## Camera Inventory

### 1. REOLINK E1 Zoom
- **Firmware**: v3.1.0.2649_23083101
- **Serial**: 192168261
- **IP**: 192.168.2.61:8000
- **Type**: PTZ Indoor Camera
- **Key Features**: PTZ support, dual stream, basic imaging

### 2. AXIS Q3819-PVE  
- **Firmware**: 10.12.153
- **Serial**: B8A44F9DC7ED
- **IP**: 192.168.2.190
- **Type**: Panoramic Fixed Dome
- **Key Features**: Ultra-wide 8192x1728 resolution, analytics, advanced imaging

### 3. AXIS P3818-PVE
- **Firmware**: 11.9.60
- **Serial**: B8A44FA04F26
- **IP**: 192.168.2.82
- **Type**: Panoramic Fixed Dome
- **Key Features**: 5120x2560 resolution, analytics, dual encoding (H264/JPEG)

### 4. Bosch FLEXIDOME panoramic 5100i
- **Firmware**: 9.00.0210
- **Serial**: 404705923918060213
- **IP**: 192.168.2.24
- **Type**: 360° Panoramic Dome
- **Key Features**: 16 profiles, dewarping, circular image (2112x2112)

### 5. Bosch FLEXIDOME IP starlight 8000i
- **Firmware**: 7.70.0126
- **Serial**: 044518807925140011
- **IP**: 192.168.2.200
- **Type**: Fixed Dome with Low-Light Performance
- **Key Features**: Starlight imaging, I/O connectors, relay output

---

## Comparative Analysis

### Resolution Capabilities

| Camera | Max Resolution | Aspect Ratio | Primary Use Case |
|--------|---------------|--------------|------------------|
| REOLINK E1 Zoom | 2048x1536 | 4:3 | Standard surveillance |
| AXIS Q3819-PVE | 8192x1728 | ~4.7:1 | 180° panoramic |
| AXIS P3818-PVE | 5120x2560 | 2:1 | 180° panoramic |
| Bosch panoramic 5100i | 2112x2112 | 1:1 | 360° fisheye |
| Bosch starlight 8000i | 1536x864 | 16:9 | Low-light environments |

### Profile Count

| Camera | Total Profiles | Video Profiles | Notes |
|--------|----------------|----------------|-------|
| REOLINK E1 Zoom | 2 | 2 | MainStream + SubStream |
| AXIS Q3819-PVE | 2 | 2 | H264 + JPEG |
| AXIS P3818-PVE | 2 | 2 | H264 + JPEG |
| Bosch panoramic 5100i | 16 | 9 valid | Includes metadata/audio profiles |
| Bosch starlight 8000i | 3 | 3 | 2x H264 + 1x JPEG |

### ONVIF Service Support

| Service | REOLINK | AXIS Q3819 | AXIS P3818 | Bosch Panoramic | Bosch Starlight |
|---------|---------|------------|------------|-----------------|-----------------|
| Device | ✓ | ✓ | ✓ | ✓ | ✓ |
| Media | ✓ | ✓ | ✓ | ✓ | ✓ |
| Imaging | ✓ | ✓ | ✓ | ✓ | ✓ |
| Events | ✓ | ✓ | ✓ | ✓ | ✓ |
| Analytics | ✗ | ✓ | ✓ | ✓ | ✗ |
| PTZ | ✓ | ✗ | ✗ | ✓ | ✗ |

### Video Encoding

| Camera | H264 | JPEG | MPEG4 | Notes |
|--------|------|------|-------|-------|
| REOLINK | ✓ | ✗ | ✗ | H264 only |
| AXIS Q3819 | ✓ | ✓ | ✗ | Dual encoding |
| AXIS P3818 | ✓ | ✓ | ✗ | Dual encoding |
| Bosch Panoramic | ✓ | ✗ | ✗ | H264 only |
| Bosch Starlight | ✓ | ✓ | ✗ | Dual encoding |

### Network Capabilities

| Feature | REOLINK | AXIS Q3819 | AXIS P3818 | Bosch Panoramic | Bosch Starlight |
|---------|---------|------------|------------|-----------------|-----------------|
| RTP Multicast | ✗ | ✓ | ✓ | ✓ | ✓ |
| RTP/TCP | ✓ | ✓ | ✓ | ✗ | ✗ |
| RTP/RTSP/TCP | ✓ | ✓ | ✓ | ✓ | ✓ |
| IPv6 Support | ✗ | ✓ | ✓ | ✗ | ✗ |
| TLS 1.2 | ✗ | ✓ | ✓ | ✓ | ✓ |

### Imaging Features

| Feature | REOLINK | AXIS Q3819 | AXIS P3818 | Bosch Panoramic | Bosch Starlight |
|---------|---------|------------|------------|-----------------|-----------------|
| Brightness Control | ✓ (128) | ✓ (50) | ✓ (50) | ✓ (127) | ✓ (128) |
| Saturation Control | ✓ (128) | ✓ (50) | ✓ (50) | ✓ (127) | ✓ (128) |
| Contrast Control | ✓ (128) | ✓ (50) | ✓ (50) | ✓ (127) | ✓ (128) |
| Sharpness Control | ✓ (128) | ✓ (50) | ✓ (50) | ✗ | ✗ |
| IrCutFilter | AUTO | AUTO | AUTO | ✗ | ✗ |
| WDR | ✗ | ON | ON | ✗ | ✗ |
| WhiteBalance | ✗ | AUTO | AUTO | ✗ | ✗ |
| Exposure Control | ✗ | AUTO | AUTO | ✗ | ✗ |

### I/O and Security

| Feature | REOLINK | AXIS Q3819 | AXIS P3818 | Bosch Panoramic | Bosch Starlight |
|---------|---------|------------|------------|-----------------|-----------------|
| Input Connectors | 0 | 2 | 2 | 0 | 2 |
| Relay Outputs | 0 | 0 | 0 | 0 | 1 |
| IP Filter | ✗ | ✓ | ✓ | ✗ | ✗ |
| TLS 1.1 | ✗ | ✓ | ✓ | ✗ | ✓ |
| TLS 1.2 | ✗ | ✓ | ✓ | ✓ | ✓ |

---

## Manufacturer-Specific Findings

### REOLINK
- **Strengths**: 
  - Simple, straightforward ONVIF implementation
  - PTZ support with status reporting
  - Good value camera with basic features
- **Limitations**:
  - Limited imaging controls (no WDR, exposure, focus)
  - Only H264 encoding (no JPEG profile)
  - No analytics support
  - Lower security features (no TLS)
- **RTSP Pattern**: `rtsp://IP:554/` (main), `rtsp://IP:554/h264Preview_01_sub` (sub)
- **Snapshot Pattern**: `http://IP:80/cgi-bin/api.cgi?cmd=onvifSnapPic&channel=0`

### AXIS
- **Strengths**:
  - Excellent ONVIF compliance and feature richness
  - Ultra-high resolution panoramic cameras
  - Advanced imaging with WDR, exposure control, white balance
  - Strong security (TLS 1.1/1.2, IP filtering, access policy)
  - Analytics and rule-based event support
- **Consistent Implementation**:
  - Both cameras share similar ONVIF structure
  - Dual H264/JPEG encoding profiles
  - Same URL patterns and capabilities
- **RTSP Pattern**: `rtsp://IP/onvif-media/media.amp?profile=X&sessiontimeout=60&streamtype=unicast`
- **Snapshot Pattern**: `http://IP/onvif-cgi/jpg/image.cgi?resolution=WxH&compression=30`
- **Notable**: Q3819 has wider aspect ratio (8192x1728 vs 5120x2560)

### Bosch
- **Strengths**:
  - Specialized cameras with unique features
  - Panoramic 5100i has comprehensive dewarping profiles
  - Starlight 8000i optimized for low-light
  - Good I/O options (starlight model has relay output)
- **Quirks**:
  - Panoramic model has 16 profiles (many without video encoders)
  - Some profiles return "IncompleteConfiguration" errors
  - Less standardized RTSP URLs (tunnel-based)
- **RTSP Pattern**: `rtsp://IP/rtsp_tunnel?p=X&line=Y&inst=Z` (various parameters)
- **Snapshot Pattern**: `http://IP/snap.jpg?JpegCam=X`
- **Notable**: 
  - Panoramic uses circular (2112x2112) and dewarped (3072x1728) views
  - 3 profiles failed GetStreamURI with incomplete configuration

---

## Performance Metrics

### Response Times (Average)

| Operation | REOLINK | AXIS Q3819 | AXIS P3818 | Bosch Panoramic | Bosch Starlight |
|-----------|---------|------------|------------|-----------------|-----------------|
| DeviceInfo | 117.7ms | 5.0ms | 4.9ms | 8.5ms | 7.9ms |
| Capabilities | 85.6ms | 72.7ms | 69.3ms | 21.9ms | 27.1ms |
| GetProfiles | 832.1ms | 70.9ms | 8.0ms | 706.2ms | 258.3ms |
| GetStreamURI | ~129ms avg | ~20ms avg | ~4ms avg | ~11ms avg | ~10ms avg |
| GetSnapshot | ~170ms avg | ~20ms avg | ~4ms avg | ~11ms avg | ~6ms avg |
| Imaging | 111.8ms | 55.8ms | 67.2ms | 57.3ms | 14.8ms |

**Key Observations**:
- AXIS cameras have fastest response times overall
- REOLINK has higher latency (likely due to port 8000, may be proxy/gateway)
- Bosch cameras have moderate, consistent response times
- GetProfiles is slowest operation for most cameras

### Error Analysis

| Camera | Total Errors | Error Types |
|--------|--------------|-------------|
| REOLINK E1 Zoom | 0 | None |
| AXIS Q3819-PVE | 0 | None |
| AXIS P3818-PVE | 0 | None |
| Bosch panoramic 5100i | 3 | GetStreamURI: IncompleteConfiguration (profiles 9,10,11) |
| Bosch starlight 8000i | 0 | None |

**Bosch Panoramic Errors**: Profiles 9, 10, 11 have no VideoEncoderConfiguration, causing legitimate failures. These appear to be metadata-only or incomplete profiles.

---

## Stream URI Patterns

### REOLINK Pattern
```
rtsp://192.168.2.61:554/                           # MainStream
rtsp://192.168.2.61:554/h264Preview_01_sub         # SubStream
```

### AXIS Pattern
```
rtsp://IP/onvif-media/media.amp?profile=profile_1_h264&sessiontimeout=60&streamtype=unicast
rtsp://IP/onvif-media/media.amp?profile=profile_1_jpeg&sessiontimeout=60&streamtype=unicast
```

### Bosch Patterns

**Indoor 5100i IR** (from previous report):
```
rtsp://IP/rtsp_tunnel?p=0&line=1&inst=1&vcd=2
```

**Panoramic 5100i**:
```
rtsp://192.168.2.24/rtsp_tunnel?p=0&line=3&inst=4     # E_PTZ view
rtsp://192.168.2.24/rtsp_tunnel?p=1&line=2&inst=1     # Dewarped view
rtsp://192.168.2.24/rtsp_tunnel?p=2&line=1&inst=4     # Full circle
rtsp://192.168.2.24/rtsp_tunnel?von=0&aon=1&aud=1     # Audio only
rtsp://192.168.2.24/rtsp_tunnel?von=0&vcd=2&line=1    # Metadata
```

**Starlight 8000i**:
```
rtsp://192.168.2.200/rtsp_tunnel?p=0&h26x=4&vcd=2
rtsp://192.168.2.200/rtsp_tunnel?p=1&inst=2&h26x=4
rtsp://192.168.2.200/rtsp_tunnel?h26x=0               # JPEG
```

**Parameter Meanings**:
- `p`: Profile index
- `line`: Video line/source (1=full, 2=dewarped, 3=ePTZ)
- `inst`: Instance number
- `vcd`: Video codec (2=metadata)
- `h26x`: H.26x codec (0=JPEG, 4=H264)
- `von`: Video on/off
- `aon`: Audio on/off

---

## PTZ Capabilities

### REOLINK E1 Zoom (PTZ Enabled)
- **PTZ Service**: http://192.168.2.61:8000/onvif/ptz_service
- **Status**: Both profiles report IDLE for PanTilt and Zoom
- **Presets**: 0 configured
- **Configuration**: PTZ config present but with empty position spaces
- **Notes**: PTZ capability exists but requires further testing for movement commands

### Bosch Panoramic 5100i (ePTZ)
- **PTZ Service**: http://192.168.2.24/onvif/ptz_service  
- **Type**: Electronic PTZ (digital zoom/pan on panoramic image)
- **Profile**: Dedicated ePTZ profile (token "0", 1920x1080)
- **Notes**: Digital PTZ on dewarped 360° image, not mechanical movement

### Other Cameras
- AXIS Q3819-PVE, P3818-PVE, Bosch starlight 8000i: No PTZ support

---

## Snapshot URI Patterns

| Manufacturer | Pattern | Authentication Required |
|--------------|---------|------------------------|
| REOLINK | `http://IP:80/cgi-bin/api.cgi?cmd=onvifSnapPic&channel=0` | Yes |
| AXIS | `http://IP/onvif-cgi/jpg/image.cgi?resolution=WxH&compression=30` | Yes |
| Bosch | `http://IP/snap.jpg?JpegCam=N` | Yes |

**InvalidAfterConnect/Reboot**:
- REOLINK: InvalidAfterConnect=true, InvalidAfterReboot=true
- AXIS: All false (persistent URIs)
- Bosch: InvalidAfterReboot=true

---

## Bitrate and Frame Rate Analysis

### REOLINK E1 Zoom
- **MainStream**: 1024 kbps @ 15fps (2048x1536)
- **SubStream**: 512 kbps @ 15fps (640x480)
- **Quality**: 0 (main), 2 (sub)

### AXIS Q3819-PVE
- **H264**: Max bitrate @ 30fps (8192x1728)
- **JPEG**: Max bitrate @ 30fps (8192x1728)
- **Quality**: 70 for both
- **Bitrate Limit**: 2147483647 (max int32 = unlimited)

### AXIS P3818-PVE
- **H264**: Max bitrate @ 30fps (1920x960)
- **JPEG**: Max bitrate @ 30fps (5120x2560)
- **Quality**: 70 for both
- **Bitrate Limit**: 2147483647 (unlimited)

### Bosch Panoramic 5100i
- **Highest**: 13000 kbps @ 30fps (3072x1728 dewarped)
- **Lowest**: 400 kbps @ 30fps (512x288)
- **Standard**: 5200 kbps @ 30fps (1920x1080)
- **Quality**: 50 across all profiles

### Bosch Starlight 8000i
- **H264**: 1400 kbps @ 30fps (1536x864)
- **JPEG**: 6000 kbps @ 1fps (1536x864)
- **Quality**: 50 (H264), 70 (JPEG)

---

## Testing Recommendations

### Priority 1: Create Camera-Specific Tests

Each manufacturer has distinct patterns worthy of dedicated test files:

1. **reolink_e1_zoom_test.go**
   - Test PTZ status retrieval
   - Verify dual-stream profiles
   - Test CGI-based snapshot URLs
   - Validate 15fps frame rate limits

2. **axis_q3819_test.go** 
   - Test ultra-wide resolution (8192x1728)
   - Verify analytics service
   - Test dual H264/JPEG encoding
   - Validate WDR and exposure settings
   - Test multicast support

3. **axis_p3818_test.go**
   - Test 5120x2560 panoramic resolution
   - Similar to Q3819 but different aspect ratio
   - Benchmark performance differences

4. **bosch_panoramic_5100i_test.go**
   - Test circular (2112x2112) image profiles
   - Test dewarped profiles
   - Handle IncompleteConfiguration errors gracefully
   - Test metadata and audio-only profiles
   - Test 16 different profiles

5. **bosch_starlight_8000i_test.go**
   - Test low-light imaging capabilities
   - Test I/O connectors (2 inputs, 1 relay output)
   - Test JPEG motion (1fps) vs H264 (30fps)

### Priority 2: Cross-Manufacturer Tests

Create tests that verify common ONVIF compliance:

1. **stream_uri_compatibility_test.go**
   - Parse and validate different RTSP URL formats
   - Test RTSP connection to each pattern
   - Verify authentication handling

2. **imaging_settings_test.go**
   - Test brightness/contrast/saturation ranges
   - Test optional features (WDR, exposure, white balance)
   - Verify manufacturer-specific defaults

3. **profile_enumeration_test.go**
   - Test handling of 2-16 profiles
   - Verify profile names and tokens
   - Test resolution validation

### Priority 3: Edge Case Tests

1. **incomplete_profile_handling_test.go**
   - Test cameras with profiles lacking video encoders
   - Verify graceful error handling for IncompleteConfiguration
   - Test metadata-only and audio-only profiles

2. **performance_benchmark_test.go**
   - Benchmark GetProfiles (100ms to 800ms variation)
   - Test response time consistency
   - Measure concurrent request handling

---

## Code Patterns for Tests

### Example: Testing AXIS Cameras

```go
func TestAXISQ3819PVE_UltraWideResolution(t *testing.T) {
    skipIfNoCamera(t)
    
    client := createTestClient(t)
    profiles, err := client.GetProfiles()
    require.NoError(t, err)
    
    // AXIS Q3819 should have H264 and JPEG profiles
    assert.Equal(t, 2, len(profiles))
    
    // Find H264 profile
    var h264Profile *onvif.Profile
    for _, p := range profiles {
        if p.VideoEncoderConfiguration != nil && 
           p.VideoEncoderConfiguration.Encoding == "H264" {
            h264Profile = &p
            break
        }
    }
    
    require.NotNil(t, h264Profile, "H264 profile should exist")
    
    // Verify ultra-wide resolution
    assert.Equal(t, 8192, h264Profile.VideoEncoderConfiguration.Resolution.Width)
    assert.Equal(t, 1728, h264Profile.VideoEncoderConfiguration.Resolution.Height)
    
    // Verify 30fps
    assert.Equal(t, 30, h264Profile.VideoEncoderConfiguration.RateControl.FrameRateLimit)
}
```

### Example: Testing Bosch Panoramic Profiles

```go
func TestBoschPanoramic5100i_MultipleProfiles(t *testing.T) {
    skipIfNoCamera(t)
    
    client := createTestClient(t)
    profiles, err := client.GetProfiles()
    require.NoError(t, err)
    
    // Should have 16 profiles
    assert.Equal(t, 16, len(profiles))
    
    // Count profiles with valid video encoders
    validVideoProfiles := 0
    for _, p := range profiles {
        if p.VideoEncoderConfiguration != nil {
            validVideoProfiles++
        }
    }
    
    assert.Equal(t, 9, validVideoProfiles, "Should have 9 video profiles")
    
    // Test that incomplete profiles fail gracefully
    for _, p := range profiles {
        uri, err := client.GetStreamURI(p.Token, "RTP-Unicast")
        
        if p.VideoEncoderConfiguration != nil {
            // Valid profiles should succeed
            if err != nil {
                t.Logf("Profile %s failed: %v", p.Token, err)
            }
        } else {
            // Incomplete profiles should fail
            assert.Error(t, err, "Profile %s should fail (no video encoder)", p.Token)
        }
    }
}
```

### Example: Testing PTZ Status

```go
func TestREOLINKE1Zoom_PTZStatus(t *testing.T) {
    skipIfNoCamera(t)
    
    client := createTestClient(t)
    profiles, err := client.GetProfiles()
    require.NoError(t, err)
    
    for _, profile := range profiles {
        if profile.PTZConfiguration != nil {
            status, err := client.GetPTZStatus(profile.Token)
            require.NoError(t, err)
            
            // Should report IDLE when not moving
            assert.NotNil(t, status.MoveStatus)
            assert.Contains(t, []string{"IDLE", "MOVING"}, status.MoveStatus.PanTilt)
            assert.Contains(t, []string{"IDLE", "MOVING"}, status.MoveStatus.Zoom)
        }
    }
}
```

---

## Integration Test Suite Structure

```
tests/
├── manufacturers/
│   ├── reolink/
│   │   └── e1_zoom_test.go
│   ├── axis/
│   │   ├── q3819_pve_test.go
│   │   └── p3818_pve_test.go
│   └── bosch/
│       ├── flexidome_indoor_5100i_ir_test.go (existing)
│       ├── flexidome_panoramic_5100i_test.go
│       └── flexidome_starlight_8000i_test.go
├── compliance/
│   ├── stream_uri_test.go
│   ├── imaging_test.go
│   └── profile_test.go
├── benchmarks/
│   └── response_time_test.go
└── edge_cases/
    ├── incomplete_profiles_test.go
    └── error_handling_test.go
```

---

## Implementation Insights

### RTSP Tunnel Parameters (Bosch)

Bosch uses a proprietary `rtsp_tunnel` endpoint with various parameters:

- **p**: Profile index (0-15)
- **line**: Video source line
  - 1 = Full image circle
  - 2 = Dewarped view mode  
  - 3 = Electronic PTZ
- **inst**: Stream instance (1-4, corresponds to bitrate tiers)
- **h26x**: Codec selection
  - 0 = JPEG
  - 4 = H.264
- **vcd**: Video coding
  - 2 = Metadata stream
- **von**: Video on (0/1)
- **aon**: Audio on (0/1)
- **aud**: Audio stream identifier
- **JpegCam**: Camera number for snapshots

### AXIS URL Parameters

- **profile**: Profile token
- **sessiontimeout**: Session timeout in seconds
- **streamtype**: unicast or multicast
- **resolution**: Snapshot resolution (WxH)
- **compression**: JPEG compression quality (0-100, lower = better)

### REOLINK CGI API

Uses proprietary CGI commands:
- `cmd=onvifSnapPic`: Get ONVIF-compliant snapshot
- `channel=0`: Camera channel

---

## Security Considerations

### Authentication
All cameras require HTTP Digest Authentication for ONVIF requests.

### TLS Support

| Camera | TLS 1.1 | TLS 1.2 | Notes |
|--------|---------|---------|-------|
| REOLINK E1 Zoom | ✗ | ✗ | HTTP only |
| AXIS Q3819-PVE | ✓ | ✓ | Full TLS support |
| AXIS P3818-PVE | ✓ | ✓ | Full TLS support |
| Bosch Panoramic 5100i | ✗ | ✓ | TLS 1.2 only |
| Bosch Starlight 8000i | ✓ | ✓ | Full TLS support |

**Recommendation**: AXIS cameras provide the strongest security posture with IP filtering, access policy config, and TLS support.

### WS-Security
All cameras support WS-Security UsernameToken with digest authentication, as evidenced by successful ONVIF communication.

---

## Compatibility Matrix

### ONVIF Profile Compliance

Based on feature analysis, likely ONVIF profile compliance:

| Camera | Profile S | Profile T | Profile G | Profile M |
|--------|-----------|-----------|-----------|-----------|
| REOLINK E1 Zoom | ✓ | ✓ (PTZ) | ✗ | ✗ |
| AXIS Q3819-PVE | ✓ | ✗ | ✓ (Analytics) | ✓ (Metadata) |
| AXIS P3818-PVE | ✓ | ✗ | ✓ (Analytics) | ✓ (Metadata) |
| Bosch Panoramic 5100i | ✓ | ✓ (ePTZ) | ✓ (Analytics) | ✓ (Metadata) |
| Bosch Starlight 8000i | ✓ | ✗ | ✗ | Partial |

**Profiles**:
- **S**: Streaming (basic video)
- **T**: PTZ control
- **G**: Video analytics
- **M**: Metadata streaming

---

## Conclusions

### Best Practices Discovered

1. **Profile Enumeration**: Always check VideoEncoderConfiguration before calling GetStreamURI
2. **Error Handling**: Bosch cameras may return IncompleteConfiguration for metadata profiles
3. **Response Times**: Expect 5-800ms for GetProfiles depending on camera complexity
4. **URL Patterns**: Cannot assume consistent RTSP URL format across manufacturers
5. **Imaging Defaults**: Manufacturers use different scales (0-255 vs 0-100 vs 0-128)

### Client Library Improvements Needed

1. **URL Parser**: Helper to parse and validate different RTSP URL formats
2. **Profile Filter**: Method to filter profiles by capability (video, audio, metadata)
3. **Retry Logic**: Handle transient errors and timeouts
4. **TLS Support**: Enable HTTPS for cameras supporting TLS
5. **Batch Operations**: Parallel GetStreamURI calls for cameras with many profiles

### Test Coverage Recommendations

Based on this analysis, create test files covering:

1. ✅ Bosch FLEXIDOME indoor 5100i IR (already exists)
2. 🔲 REOLINK E1 Zoom (PTZ, dual stream)
3. 🔲 AXIS Q3819-PVE (ultra-wide, analytics)
4. 🔲 AXIS P3818-PVE (panoramic, analytics)
5. 🔲 Bosch FLEXIDOME panoramic 5100i (16 profiles, dewarping)
6. 🔲 Bosch FLEXIDOME IP starlight 8000i (low-light, I/O)

### Interoperability Score

Based on ONVIF compliance, feature richness, and ease of integration:

| Camera | Score | Rationale |
|--------|-------|-----------|
| AXIS P3818-PVE | 9.5/10 | Excellent compliance, fast, feature-rich |
| AXIS Q3819-PVE | 9.5/10 | Same as P3818, ultra-wide resolution |
| Bosch Starlight 8000i | 8.0/10 | Good compliance, moderate features |
| Bosch Panoramic 5100i | 7.5/10 | Complex profile structure, some errors |
| REOLINK E1 Zoom | 7.0/10 | Basic features, slower responses, limited imaging |

---

## Next Steps

1. **Create manufacturer-specific test files** for each camera model
2. **Implement helper functions** for common patterns (URL parsing, profile filtering)
3. **Add benchmark tests** to track performance regression
4. **Document manufacturer quirks** in code comments
5. **Create CI/CD pipeline** to test against real cameras (when available)
6. **Expand coverage** for PTZ operations on REOLINK
7. **Test analytics** on AXIS cameras
8. **Validate TLS connections** on supported cameras

---

## Appendix: Raw Data Summary

### REOLINK E1 Zoom
- Profiles: 2
- Stream URIs: 2/2 successful
- Snapshot URIs: 2/2 successful
- Video Encoders: 2/2 successful
- Imaging Settings: 1/1 successful
- PTZ Status: 2/2 successful (both IDLE)
- PTZ Presets: 0
- Total Errors: 0

### AXIS Q3819-PVE
- Profiles: 2
- Stream URIs: 2/2 successful
- Snapshot URIs: 2/2 successful
- Video Encoders: 2/2 successful
- Imaging Settings: 1/1 successful
- Total Errors: 0

### AXIS P3818-PVE
- Profiles: 2
- Stream URIs: 2/2 successful
- Snapshot URIs: 2/2 successful
- Video Encoders: 2/2 successful
- Imaging Settings: 1/1 successful
- Total Errors: 0

### Bosch FLEXIDOME panoramic 5100i
- Profiles: 16
- Stream URIs: 13/16 successful (3 IncompleteConfiguration errors)
- Snapshot URIs: 16/16 successful
- Video Encoders: 9/9 successful (only tested valid profiles)
- Imaging Settings: 1/1 successful
- Total Errors: 3 (expected for incomplete profiles)

### Bosch FLEXIDOME IP starlight 8000i
- Profiles: 3
- Stream URIs: 3/3 successful
- Snapshot URIs: 3/3 successful
- Video Encoders: 3/3 successful
- Imaging Settings: 1/1 successful
- Total Errors: 0

---

**End of Analysis Report**
