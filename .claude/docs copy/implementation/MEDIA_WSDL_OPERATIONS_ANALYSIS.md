# ONVIF Media Service WSDL Operations Analysis

## Total Operations in WSDL: 79

Based on the official ONVIF Media Service WSDL at https://www.onvif.org/ver10/media/wsdl/media.wsdl, there are **79 operations** defined.

## Operations Breakdown

### 1. Service Capabilities (1 operation)
1. ✅ `GetServiceCapabilities` / `GetMediaServiceCapabilities` - **IMPLEMENTED**

### 2. Profile Management (5 operations)
2. ✅ `GetProfiles` - **IMPLEMENTED**
3. ✅ `GetProfile` - **IMPLEMENTED**
4. ✅ `SetProfile` - **IMPLEMENTED**
5. ✅ `CreateProfile` - **IMPLEMENTED**
6. ✅ `DeleteProfile` - **IMPLEMENTED**

### 3. Stream Operations (4 operations)
7. ✅ `GetStreamUri` - **IMPLEMENTED**
8. ✅ `GetSnapshotUri` - **IMPLEMENTED**
9. ✅ `StartMulticastStreaming` - **IMPLEMENTED**
10. ✅ `StopMulticastStreaming` - **IMPLEMENTED**
11. ✅ `SetSynchronizationPoint` - **IMPLEMENTED**

### 4. Source Operations (2 operations)
12. ✅ `GetVideoSources` - **IMPLEMENTED**
13. ✅ `GetAudioSources` - **IMPLEMENTED**

### 5. Configuration Retrieval - Plural Forms (8 operations)
14. ❌ `GetVideoSourceConfigurations` - **NOT IMPLEMENTED**
15. ❌ `GetAudioSourceConfigurations` - **NOT IMPLEMENTED**
16. ❌ `GetVideoEncoderConfigurations` - **NOT IMPLEMENTED**
17. ❌ `GetAudioEncoderConfigurations` - **NOT IMPLEMENTED**
18. ❌ `GetVideoAnalyticsConfigurations` - **NOT IMPLEMENTED**
19. ❌ `GetMetadataConfigurations` - **NOT IMPLEMENTED**
20. ❌ `GetAudioOutputConfigurations` - **NOT IMPLEMENTED**
21. ❌ `GetAudioDecoderConfigurations` - **NOT IMPLEMENTED**

### 6. Configuration Retrieval - Singular Forms (8 operations)
22. ❌ `GetVideoSourceConfiguration` - **NOT IMPLEMENTED**
23. ❌ `GetAudioSourceConfiguration` - **NOT IMPLEMENTED**
24. ✅ `GetVideoEncoderConfiguration` - **IMPLEMENTED**
25. ✅ `GetAudioEncoderConfiguration` - **IMPLEMENTED**
26. ❌ `GetVideoAnalyticsConfiguration` - **NOT IMPLEMENTED**
27. ✅ `GetMetadataConfiguration` - **IMPLEMENTED**
28. ✅ `GetAudioOutputConfiguration` - **IMPLEMENTED**
29. ❌ `GetAudioDecoderConfiguration` - **NOT IMPLEMENTED**

### 7. Compatible Configuration Operations (8 operations)
30. ❌ `GetCompatibleVideoEncoderConfigurations` - **NOT IMPLEMENTED**
31. ❌ `GetCompatibleVideoSourceConfigurations` - **NOT IMPLEMENTED**
32. ❌ `GetCompatibleAudioEncoderConfigurations` - **NOT IMPLEMENTED**
33. ❌ `GetCompatibleAudioSourceConfigurations` - **NOT IMPLEMENTED**
34. ❌ `GetCompatiblePTZConfigurations` - **NOT IMPLEMENTED**
35. ❌ `GetCompatibleVideoAnalyticsConfigurations` - **NOT IMPLEMENTED**
36. ❌ `GetCompatibleMetadataConfigurations` - **NOT IMPLEMENTED**
37. ❌ `GetCompatibleAudioOutputConfigurations` - **NOT IMPLEMENTED**
38. ❌ `GetCompatibleAudioDecoderConfigurations` - **NOT IMPLEMENTED**

### 8. Configuration Setting Operations (8 operations)
39. ❌ `SetVideoSourceConfiguration` - **NOT IMPLEMENTED**
40. ✅ `SetVideoEncoderConfiguration` - **IMPLEMENTED**
41. ❌ `SetAudioSourceConfiguration` - **NOT IMPLEMENTED**
42. ✅ `SetAudioEncoderConfiguration` - **IMPLEMENTED**
43. ❌ `SetVideoAnalyticsConfiguration` - **NOT IMPLEMENTED**
44. ✅ `SetMetadataConfiguration` - **IMPLEMENTED**
45. ✅ `SetAudioOutputConfiguration` - **IMPLEMENTED**
46. ❌ `SetAudioDecoderConfiguration` - **NOT IMPLEMENTED**

### 9. Configuration Options Operations (8 operations)
47. ❌ `GetVideoSourceConfigurationOptions` - **NOT IMPLEMENTED**
48. ✅ `GetVideoEncoderConfigurationOptions` - **IMPLEMENTED**
49. ❌ `GetAudioSourceConfigurationOptions` - **NOT IMPLEMENTED**
50. ✅ `GetAudioEncoderConfigurationOptions` - **IMPLEMENTED**
51. ❌ `GetVideoAnalyticsConfigurationOptions` - **NOT IMPLEMENTED**
52. ✅ `GetMetadataConfigurationOptions` - **IMPLEMENTED**
53. ✅ `GetAudioOutputConfigurationOptions` - **IMPLEMENTED**
54. ✅ `GetAudioDecoderConfigurationOptions` - **IMPLEMENTED**

### 10. Profile Configuration Add Operations (9 operations)
55. ✅ `AddVideoEncoderConfiguration` - **IMPLEMENTED**
56. ✅ `AddVideoSourceConfiguration` - **IMPLEMENTED**
57. ✅ `AddAudioEncoderConfiguration` - **IMPLEMENTED**
58. ✅ `AddAudioSourceConfiguration` - **IMPLEMENTED**
59. ✅ `AddPTZConfiguration` - **IMPLEMENTED**
60. ❌ `AddVideoAnalyticsConfiguration` - **NOT IMPLEMENTED**
61. ✅ `AddMetadataConfiguration` - **IMPLEMENTED**
62. ❌ `AddAudioOutputConfiguration` - **NOT IMPLEMENTED**
63. ❌ `AddAudioDecoderConfiguration` - **NOT IMPLEMENTED**

### 11. Profile Configuration Remove Operations (9 operations)
64. ✅ `RemoveVideoEncoderConfiguration` - **IMPLEMENTED**
65. ✅ `RemoveVideoSourceConfiguration` - **IMPLEMENTED**
66. ✅ `RemoveAudioEncoderConfiguration` - **IMPLEMENTED**
67. ✅ `RemoveAudioSourceConfiguration` - **IMPLEMENTED**
68. ✅ `RemovePTZConfiguration` - **IMPLEMENTED**
69. ❌ `RemoveVideoAnalyticsConfiguration` - **NOT IMPLEMENTED**
70. ✅ `RemoveMetadataConfiguration` - **IMPLEMENTED**
71. ❌ `RemoveAudioOutputConfiguration` - **NOT IMPLEMENTED**
72. ❌ `RemoveAudioDecoderConfiguration` - **NOT IMPLEMENTED**

### 12. Video Source Mode Operations (2 operations)
73. ✅ `GetVideoSourceModes` - **IMPLEMENTED**
74. ✅ `SetVideoSourceMode` - **IMPLEMENTED**

### 13. OSD Operations (6 operations)
75. ✅ `GetOSDs` - **IMPLEMENTED**
76. ✅ `GetOSD` - **IMPLEMENTED**
77. ✅ `GetOSDOptions` - **IMPLEMENTED**
78. ✅ `SetOSD` - **IMPLEMENTED**
79. ✅ `CreateOSD` - **IMPLEMENTED**
80. ✅ `DeleteOSD` - **IMPLEMENTED**

### 14. Advanced Operations (1 operation)
81. ✅ `GetGuaranteedNumberOfVideoEncoderInstances` - **IMPLEMENTED**

---

## Summary

### Implementation Status

| Category | Total | Implemented | Missing |
|----------|-------|-------------|---------|
| Service Capabilities | 1 | 1 | 0 |
| Profile Management | 5 | 5 | 0 |
| Stream Operations | 5 | 5 | 0 |
| Source Operations | 2 | 2 | 0 |
| Config Retrieval (Plural) | 8 | 0 | 8 |
| Config Retrieval (Singular) | 8 | 4 | 4 |
| Compatible Configs | 9 | 0 | 9 |
| Config Setting | 8 | 4 | 4 |
| Config Options | 8 | 5 | 3 |
| Profile Add Config | 9 | 6 | 3 |
| Profile Remove Config | 9 | 6 | 3 |
| Video Source Modes | 2 | 2 | 0 |
| OSD Operations | 6 | 6 | 0 |
| Advanced Operations | 1 | 1 | 0 |
| **TOTAL** | **79** | **47** | **32** |

### Current Implementation: 47/79 = 59.5%

### Missing Operations: 32 operations

#### High Priority (Commonly Used)
1. `GetVideoSourceConfigurations` (plural)
2. `GetAudioSourceConfigurations` (plural)
3. `GetVideoEncoderConfigurations` (plural)
4. `GetAudioEncoderConfigurations` (plural)
5. `GetVideoSourceConfiguration` (singular)
6. `GetAudioSourceConfiguration` (singular)
7. `GetVideoSourceConfigurationOptions`
8. `GetAudioSourceConfigurationOptions`
9. `SetVideoSourceConfiguration`
10. `SetAudioSourceConfiguration`

#### Medium Priority (Useful for Discovery)
11. `GetCompatibleVideoEncoderConfigurations`
12. `GetCompatibleVideoSourceConfigurations`
13. `GetCompatibleAudioEncoderConfigurations`
14. `GetCompatibleAudioSourceConfigurations`
15. `GetCompatibleMetadataConfigurations`
16. `GetCompatibleAudioOutputConfigurations`
17. `GetCompatiblePTZConfigurations`

#### Lower Priority (Video Analytics - Less Common)
18. `GetVideoAnalyticsConfigurations`
19. `GetVideoAnalyticsConfiguration`
20. `GetCompatibleVideoAnalyticsConfigurations`
21. `SetVideoAnalyticsConfiguration`
22. `GetVideoAnalyticsConfigurationOptions`
23. `AddVideoAnalyticsConfiguration`
24. `RemoveVideoAnalyticsConfiguration`

#### Lower Priority (Audio Decoder - Less Common)
25. `GetAudioDecoderConfiguration`
26. `SetAudioDecoderConfiguration`
27. `AddAudioDecoderConfiguration`
28. `RemoveAudioDecoderConfiguration`

#### Lower Priority (Metadata/Audio Output Plural - May be Redundant)
29. `GetMetadataConfigurations` (plural)
30. `GetAudioOutputConfigurations` (plural)
31. `AddAudioOutputConfiguration`
32. `RemoveAudioOutputConfiguration`

---

## Recommendations

### Phase 1: High Priority (10 operations)
Implement the most commonly used operations:
- Plural form retrievals for Video/Audio Source/Encoder configurations
- Singular form retrievals for Video/Audio Source configurations
- Configuration options for Video/Audio Source
- Set operations for Video/Audio Source configurations

### Phase 2: Medium Priority (7 operations)
Implement compatible configuration discovery operations for better profile building support.

### Phase 3: Lower Priority (15 operations)
Implement Video Analytics and Audio Decoder operations if needed for specific use cases.

---

*Analysis based on ONVIF Media Service WSDL v1.0*  
*Reference: https://www.onvif.org/ver10/media/wsdl/media.wsdl*  
*Last Updated: December 2, 2025*

