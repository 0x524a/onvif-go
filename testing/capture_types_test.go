package onviftesting

import (
	"encoding/json"
	"testing"
)

func TestDetectCaptureVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "V1 format (no version)",
			input:    `{"timestamp":"2025-01-01T00:00:00Z","operation":1}`,
			expected: "1.0",
		},
		{
			name:     "V2 format",
			input:    `{"version":"2.0","timestamp":"2025-01-01T00:00:00Z"}`,
			expected: "2.0",
		},
		{
			name:     "Empty object",
			input:    `{}`,
			expected: "1.0",
		},
		{
			name:     "Invalid JSON",
			input:    `{invalid}`,
			expected: "1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectCaptureVersion([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("DetectCaptureVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCapturedExchangeV2_IsV2(t *testing.T) {
	tests := []struct {
		name     string
		exchange CapturedExchangeV2
		expected bool
	}{
		{
			name:     "V2 exchange",
			exchange: CapturedExchangeV2{Version: "2.0"},
			expected: true,
		},
		{
			name:     "V1 exchange (empty version)",
			exchange: CapturedExchangeV2{Version: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.exchange.IsV2(); result != tt.expected {
				t.Errorf("IsV2() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCapturedExchangeV2_GetTokens(t *testing.T) {
	exchange := CapturedExchangeV2{
		Parameters: map[string]interface{}{
			"ProfileToken":       "profile1",
			"ConfigurationToken": "config1",
			"VideoSourceToken":   "video1",
		},
	}

	if token := exchange.GetProfileToken(); token != "profile1" {
		t.Errorf("GetProfileToken() = %v, want %v", token, "profile1")
	}

	if token := exchange.GetConfigurationToken(); token != "config1" {
		t.Errorf("GetConfigurationToken() = %v, want %v", token, "config1")
	}

	if token := exchange.GetVideoSourceToken(); token != "video1" {
		t.Errorf("GetVideoSourceToken() = %v, want %v", token, "video1")
	}
}

func TestCapturedExchangeV2_GetTokens_Empty(t *testing.T) {
	exchange := CapturedExchangeV2{}

	if token := exchange.GetProfileToken(); token != "" {
		t.Errorf("GetProfileToken() should return empty string for nil parameters")
	}
}

func TestBuildMatchKey(t *testing.T) {
	params := map[string]interface{}{
		"ProfileToken":       "profile1",
		"ConfigurationToken": "config1",
	}

	key := BuildMatchKey("GetStreamURI", params)

	if key.OperationName != "GetStreamURI" {
		t.Errorf("OperationName = %v, want %v", key.OperationName, "GetStreamURI")
	}

	if key.ProfileToken != "profile1" {
		t.Errorf("ProfileToken = %v, want %v", key.ProfileToken, "profile1")
	}

	if key.ConfigurationToken != "config1" {
		t.Errorf("ConfigurationToken = %v, want %v", key.ConfigurationToken, "config1")
	}
}

func TestMatchKey_MatchScore(t *testing.T) {
	tests := []struct {
		name     string
		key1     MatchKey
		key2     MatchKey
		expected int
	}{
		{
			name:     "Different operations",
			key1:     MatchKey{OperationName: "GetProfiles"},
			key2:     MatchKey{OperationName: "GetStreamURI"},
			expected: -1,
		},
		{
			name:     "Same operation only",
			key1:     MatchKey{OperationName: "GetProfiles"},
			key2:     MatchKey{OperationName: "GetProfiles"},
			expected: 1,
		},
		{
			name:     "Same operation with matching profile",
			key1:     MatchKey{OperationName: "GetStreamURI", ProfileToken: "profile1"},
			key2:     MatchKey{OperationName: "GetStreamURI", ProfileToken: "profile1"},
			expected: 11, // 1 + 10
		},
		{
			name:     "Same operation with non-matching profile",
			key1:     MatchKey{OperationName: "GetStreamURI", ProfileToken: "profile1"},
			key2:     MatchKey{OperationName: "GetStreamURI", ProfileToken: "profile2"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.key1.MatchScore(&tt.key2); result != tt.expected {
				t.Errorf("MatchScore() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetermineServiceType(t *testing.T) {
	tests := []struct {
		name     string
		soapBody string
		expected ServiceType
	}{
		{
			name:     "Device service",
			soapBody: `xmlns="http://www.onvif.org/ver10/device/wsdl"`,
			expected: ServiceDevice,
		},
		{
			name:     "Media service",
			soapBody: `xmlns="http://www.onvif.org/ver10/media/wsdl"`,
			expected: ServiceMedia,
		},
		{
			name:     "PTZ service",
			soapBody: `xmlns="http://www.onvif.org/ver20/ptz/wsdl"`,
			expected: ServicePTZ,
		},
		{
			name:     "Unknown namespace",
			soapBody: `xmlns="http://example.com/unknown"`,
			expected: ServiceUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := DetermineServiceType(tt.soapBody); result != tt.expected {
				t.Errorf("DetermineServiceType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertV1ToV2(t *testing.T) {
	v1 := &CapturedExchange{
		Timestamp:     "2025-01-01T00:00:00Z",
		Operation:     1,
		OperationName: "GetDeviceInformation",
		Endpoint:      "http://camera/onvif/device_service",
		RequestBody:   "<request/>",
		ResponseBody:  "<response/>",
		StatusCode:    200,
	}

	v2 := ConvertV1ToV2(v1)

	if v2.Version != "" {
		t.Errorf("Version should be empty for converted V1, got %v", v2.Version)
	}

	if v2.OperationName != v1.OperationName {
		t.Errorf("OperationName = %v, want %v", v2.OperationName, v1.OperationName)
	}

	if v2.StatusCode != v1.StatusCode {
		t.Errorf("StatusCode = %v, want %v", v2.StatusCode, v1.StatusCode)
	}

	if !v2.Success {
		t.Errorf("Success should be true for 200 status")
	}
}

func TestCaptureMetadata_JSON(t *testing.T) {
	metadata := CaptureMetadata{
		Version:     CaptureVersion,
		ToolVersion: "1.0.0",
		CameraInfo: CameraInfo{
			Manufacturer:    "Bosch",
			Model:           "FLEXIDOME",
			FirmwareVersion: "8.71.0066",
		},
		TotalExchanges: 100,
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed CaptureMetadata
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed.Version != CaptureVersion {
		t.Errorf("Version = %v, want %v", parsed.Version, CaptureVersion)
	}

	if parsed.CameraInfo.Manufacturer != "Bosch" {
		t.Errorf("Manufacturer = %v, want %v", parsed.CameraInfo.Manufacturer, "Bosch")
	}
}
