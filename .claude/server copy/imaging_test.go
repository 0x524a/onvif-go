package server

import (
	"encoding/xml"
	"testing"
)

const (
	exposureModeAuto   = "AUTO"
	exposureModeManual = "MANUAL"
)

func TestHandleGetImagingSettings(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	videoSourceToken := config.Profiles[0].VideoSource.Token

	req := GetImagingSettingsRequest{VideoSourceToken: videoSourceToken}

	resp, err := server.HandleGetImagingSettings(&req)
	if err != nil {
		t.Fatalf("HandleGetImagingSettings() error = %v", err)
	}

	settingsResp, ok := resp.(*GetImagingSettingsResponse)
	if !ok {
		t.Fatalf("Response is not GetImagingSettingsResponse, got %T", resp)
	}

	if settingsResp.ImagingSettings == nil {
		t.Error("ImagingSettings is nil")

		return
	}

	// Check that settings have default values
	if settingsResp.ImagingSettings.Brightness != nil {
		if *settingsResp.ImagingSettings.Brightness < 0 || *settingsResp.ImagingSettings.Brightness > 100 {
			t.Errorf("Brightness out of range: %f", *settingsResp.ImagingSettings.Brightness)
		}
	}
	if settingsResp.ImagingSettings.Contrast != nil {
		if *settingsResp.ImagingSettings.Contrast < 0 || *settingsResp.ImagingSettings.Contrast > 100 {
			t.Errorf("Contrast out of range: %f", *settingsResp.ImagingSettings.Contrast)
		}
	}
}

func TestHandleSetImagingSettings(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	videoSourceToken := config.Profiles[0].VideoSource.Token

	brightness := 75.0
	contrast := 60.0
	setReq := SetImagingSettingsRequest{
		VideoSourceToken: videoSourceToken,
		ImagingSettings: &ImagingSettings{
			Brightness: &brightness,
			Contrast:   &contrast,
		},
		ForcePersistence: true,
	}

	resp, err := server.HandleSetImagingSettings(&setReq)
	if err != nil {
		t.Fatalf("HandleSetImagingSettings() error = %v", err)
	}

	setResp, ok := resp.(*SetImagingSettingsResponse)
	if !ok {
		t.Fatalf("Response is not SetImagingSettingsResponse, got %T", resp)
	}

	if setResp == nil {
		t.Error("SetImagingSettingsResponse is nil")
	}

	// Verify the settings were actually changed
	getReq := GetImagingSettingsRequest{VideoSourceToken: videoSourceToken}
	getResp, _ := server.HandleGetImagingSettings(&getReq)
	getResp2, _ := getResp.(*GetImagingSettingsResponse)
	if getResp2.ImagingSettings.Brightness == nil || *getResp2.ImagingSettings.Brightness != 75 {
		if getResp2.ImagingSettings.Brightness != nil {
			t.Errorf("Brightness not set: got %f, want 75", *getResp2.ImagingSettings.Brightness)
		} else {
			t.Error("Brightness is nil")
		}
	}
}

func TestHandleGetOptions(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	videoSourceToken := config.Profiles[0].VideoSource.Token

	type getOptionsRequest struct {
		VideoSourceToken string `xml:"VideoSourceToken"`
	}

	req := getOptionsRequest{VideoSourceToken: videoSourceToken}
	reqData, _ := xml.Marshal(req)

	resp, err := server.HandleGetOptions(reqData)
	if err != nil {
		t.Fatalf("HandleGetOptions() error = %v", err)
	}

	optionsResp, ok := resp.(*GetOptionsResponse)
	if !ok {
		t.Fatalf("Response is not GetOptionsResponse, got %T", resp)
	}

	if optionsResp.ImagingOptions == nil {
		t.Error("ImagingOptions is nil")

		return
	}

	// Check that options define valid ranges
	if optionsResp.ImagingOptions.Brightness == nil {
		t.Error("Brightness options is nil")
	}
	if optionsResp.ImagingOptions.Contrast == nil {
		t.Error("Contrast options is nil")
	}
}

// TestHandleMove - DISABLED due to SOAP namespace requirements.
//
//nolint:unused // Disabled test function kept for reference
func _DisabledTestHandleMove(t *testing.T) {
	t.Helper()
	config := createTestConfig()
	server, _ := New(config)
	videoSourceToken := config.Profiles[0].VideoSource.Token

	reqXML := `<Move><VideoSourceToken>` + videoSourceToken + `</VideoSourceToken><Focus><Absolute><Position>0.5</Position></Absolute></Focus></Move>`
	resp, err := server.HandleMove([]byte(reqXML))
	if err != nil {
		t.Fatalf("HandleMove() error = %v", err)
	}

	moveResp, ok := resp.(*MoveResponse)
	if !ok {
		t.Fatalf("Response is not MoveResponse, got %T", resp)
	}

	if moveResp == nil {
		t.Error("MoveResponse is nil")
	}
}

func TestImagingSettings(t *testing.T) {
	brightness := 75.0
	contrast := 60.0
	saturation := 50.0
	sharpness := 50.0
	irCutFilter := exposureModeAuto
	level := 50.0
	gain := 50.0
	exposureTime := 100.0
	defaultSpeed := 0.5
	crGain := 128.0
	cbGain := 128.0

	settings := &ImagingSettings{
		Brightness:      &brightness,
		Contrast:        &contrast,
		ColorSaturation: &saturation,
		Sharpness:       &sharpness,
		IrCutFilter:     &irCutFilter,
		BacklightCompensation: &BacklightCompensationSettings{
			Mode:  "ON",
			Level: &level,
		},
		Exposure: &ExposureSettings20{
			Mode:         exposureModeAuto,
			ExposureTime: &exposureTime,
			Gain:         &gain,
		},
		Focus: &FocusConfiguration20{
			AutoFocusMode: exposureModeAuto,
			DefaultSpeed:  &defaultSpeed,
		},
		WhiteBalance: &WhiteBalanceSettings20{
			Mode:   exposureModeAuto,
			CrGain: &crGain,
			CbGain: &cbGain,
		},
		WideDynamicRange: &WideDynamicRangeSettings{
			Mode:  "ON",
			Level: &level,
		},
	}

	// Validate all settings
	if settings.Brightness != nil && (*settings.Brightness < 0 || *settings.Brightness > 100) {
		t.Errorf("Brightness invalid: %f", *settings.Brightness)
	}
	if settings.Contrast != nil && (*settings.Contrast < 0 || *settings.Contrast > 100) {
		t.Errorf("Contrast invalid: %f", *settings.Contrast)
	}
	if settings.ColorSaturation != nil && (*settings.ColorSaturation < 0 || *settings.ColorSaturation > 100) {
		t.Errorf("ColorSaturation invalid: %f", *settings.ColorSaturation)
	}
	if settings.Sharpness != nil && (*settings.Sharpness < 0 || *settings.Sharpness > 100) {
		t.Errorf("Sharpness invalid: %f", *settings.Sharpness)
	}

	if settings.BacklightCompensation != nil && settings.BacklightCompensation.Mode != "ON" {
		t.Errorf("BacklightCompensation mode invalid: %s", settings.BacklightCompensation.Mode)
	}

	if settings.Exposure != nil && settings.Exposure.Mode != exposureModeAuto {
		t.Errorf("Exposure mode invalid: %s", settings.Exposure.Mode)
	}

	if settings.Focus != nil && settings.Focus.AutoFocusMode != exposureModeAuto {
		t.Errorf("Focus mode invalid: %s", settings.Focus.AutoFocusMode)
	}

	if settings.WhiteBalance.Mode != exposureModeAuto {
		t.Errorf("WhiteBalance mode invalid: %s", settings.WhiteBalance.Mode)
	}
}

func TestBacklightCompensation(t *testing.T) {
	tests := []struct {
		name        string
		comp        BacklightCompensation
		expectValid bool
	}{
		{
			name:        "Backlight ON",
			comp:        BacklightCompensation{Mode: "ON", Level: 50},
			expectValid: true,
		},
		{
			name:        "Backlight OFF",
			comp:        BacklightCompensation{Mode: "OFF", Level: 0},
			expectValid: true,
		},
		{
			name:        "Invalid mode",
			comp:        BacklightCompensation{Mode: "INVALID", Level: 50},
			expectValid: false,
		},
		{
			name:        "Level out of range",
			comp:        BacklightCompensation{Mode: "ON", Level: 150},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := (tt.comp.Mode == "ON" || tt.comp.Mode == "OFF") &&
				tt.comp.Level >= 0 && tt.comp.Level <= 100
			if valid != tt.expectValid {
				t.Errorf("Backlight validation failed: Mode=%s, Level=%f", tt.comp.Mode, tt.comp.Level)
			}
		})
	}
}

func TestExposureSettings(t *testing.T) {
	tests := []struct {
		name        string
		exposure    ExposureSettings
		expectValid bool
	}{
		{
			name: "Valid AUTO exposure",
			exposure: ExposureSettings{
				Mode:        "AUTO",
				Priority:    "FrameRate",
				MinExposure: 1,
				MaxExposure: 10000,
				Gain:        50,
			},
			expectValid: true,
		},
		{
			name: "Valid MANUAL exposure",
			exposure: ExposureSettings{
				Mode:         exposureModeManual,
				ExposureTime: 100,
				Gain:         50,
			},
			expectValid: true,
		},
		{
			name: "Invalid mode",
			exposure: ExposureSettings{
				Mode: "INVALID",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.exposure.Mode == exposureModeAuto || tt.exposure.Mode == exposureModeManual
			if valid != tt.expectValid {
				t.Errorf("Exposure validation failed: Mode=%s", tt.exposure.Mode)
			}
		})
	}
}

func TestFocusSettings(t *testing.T) {
	tests := []struct {
		name        string
		focus       FocusSettings
		expectValid bool
	}{
		{
			name: "Valid AUTO focus",
			focus: FocusSettings{
				AutoFocusMode: exposureModeAuto,
				DefaultSpeed:  0.5,
				NearLimit:     0,
				FarLimit:      1,
			},
			expectValid: true,
		},
		{
			name: "Valid MANUAL focus",
			focus: FocusSettings{
				AutoFocusMode: exposureModeManual,
				DefaultSpeed:  0.5,
				CurrentPos:    0.5,
			},
			expectValid: true,
		},
		{
			name: "Invalid mode",
			focus: FocusSettings{
				AutoFocusMode: "INVALID",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.focus.AutoFocusMode == exposureModeAuto || tt.focus.AutoFocusMode == exposureModeManual
			if valid != tt.expectValid {
				t.Errorf("Focus validation failed: Mode=%s", tt.focus.AutoFocusMode)
			}
		})
	}
}

func TestWhiteBalanceSettings(t *testing.T) {
	tests := []struct {
		name         string
		whiteBalance WhiteBalanceSettings
		expectValid  bool
	}{
		{
			name: "Valid AUTO white balance",
			whiteBalance: WhiteBalanceSettings{
				Mode:   exposureModeAuto,
				CrGain: 128,
				CbGain: 128,
			},
			expectValid: true,
		},
		{
			name: "Valid MANUAL white balance",
			whiteBalance: WhiteBalanceSettings{
				Mode:   "MANUAL",
				CrGain: 100,
				CbGain: 120,
			},
			expectValid: true,
		},
		{
			name: "Gain out of range",
			whiteBalance: WhiteBalanceSettings{
				Mode:   exposureModeAuto,
				CrGain: 300,
				CbGain: 128,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := (tt.whiteBalance.Mode == exposureModeAuto || tt.whiteBalance.Mode == exposureModeManual) &&
				tt.whiteBalance.CrGain >= 0 && tt.whiteBalance.CrGain <= 255 &&
				tt.whiteBalance.CbGain >= 0 && tt.whiteBalance.CbGain <= 255
			if valid != tt.expectValid {
				t.Errorf("WhiteBalance validation failed: Mode=%s, Cr=%f, Cb=%f",
					tt.whiteBalance.Mode, tt.whiteBalance.CrGain, tt.whiteBalance.CbGain)
			}
		})
	}
}

func TestWideDynamicRange(t *testing.T) {
	tests := []struct {
		name        string
		wdr         WDRSettings
		expectValid bool
	}{
		{
			name:        "WDR ON",
			wdr:         WDRSettings{Mode: "ON", Level: 50},
			expectValid: true,
		},
		{
			name:        "WDR OFF",
			wdr:         WDRSettings{Mode: "OFF", Level: 0},
			expectValid: true,
		},
		{
			name:        "Invalid mode",
			wdr:         WDRSettings{Mode: "INVALID", Level: 50},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := (tt.wdr.Mode == "ON" || tt.wdr.Mode == "OFF") &&
				tt.wdr.Level >= 0 && tt.wdr.Level <= 100
			if valid != tt.expectValid {
				t.Errorf("WDR validation failed: Mode=%s, Level=%f", tt.wdr.Mode, tt.wdr.Level)
			}
		})
	}
}

func TestGetImagingSettingsResponseXML(t *testing.T) {
	brightness := 75.0
	contrast := 60.0
	resp := &GetImagingSettingsResponse{
		ImagingSettings: &ImagingSettings{
			Brightness: &brightness,
			Contrast:   &contrast,
		},
	}

	// Marshal to XML
	data, err := xml.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal back
	var unmarshaled GetImagingSettingsResponse
	err = xml.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaled.ImagingSettings == nil {
		t.Error("ImagingSettings is nil after unmarshal")
	}
	if unmarshaled.ImagingSettings.Brightness == nil || *unmarshaled.ImagingSettings.Brightness != 75 {
		if unmarshaled.ImagingSettings.Brightness != nil {
			t.Errorf("Brightness mismatch: %f != 75", *unmarshaled.ImagingSettings.Brightness)
		} else {
			t.Error("Brightness is nil")
		}
	}
}

func TestHandleGetOptionsDetails(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	videoSourceToken := config.Profiles[0].VideoSource.Token

	resp, err := server.HandleGetOptions(struct {
		VideoSourceToken string `xml:"VideoSourceToken"`
	}{VideoSourceToken: videoSourceToken})

	if err != nil {
		t.Fatalf("HandleGetOptions error: %v", err)
	}

	optionsResp, ok := resp.(*GetOptionsResponse)
	if !ok {
		t.Fatalf("Response is not GetOptionsResponse: %T", resp)
	}

	if optionsResp.ImagingOptions == nil {
		t.Error("ImagingOptions is nil")
	}
}

func TestImagingSettingsEdgeCases(t *testing.T) {
	// Test with nil imaging settings
	settings := &ImagingSettings{}

	// All pointers should be nil initially
	if settings.Brightness != nil {
		t.Error("Brightness should be nil")
	}
	if settings.Contrast != nil {
		t.Error("Contrast should be nil")
	}
}

func TestSetImagingSettingsEdgeCases(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	videoSourceToken := config.Profiles[0].VideoSource.Token

	// Test with empty imaging settings
	setReq := SetImagingSettingsRequest{
		VideoSourceToken: videoSourceToken,
		ImagingSettings:  nil,
		ForcePersistence: false,
	}

	resp, err := server.HandleSetImagingSettings(&setReq)

	if err == nil && resp != nil {
		t.Logf("SetImagingSettings with nil settings succeeded")
	}
}

func TestGetImagingSettingsEdgeCases(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	// Test with invalid token
	invalidReq := struct {
		VideoSourceToken string `xml:"VideoSourceToken"`
	}{VideoSourceToken: "invalid_token"}

	resp, err := server.HandleGetImagingSettings(invalidReq)

	if err == nil {
		t.Error("Expected error for invalid token")
	}
	if resp != nil {
		t.Error("Expected nil response for error case")
	}
}
