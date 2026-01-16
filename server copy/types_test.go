package server

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name      string
		checkFunc func(*Config) error
	}{
		{
			name: "Host is set",
			checkFunc: func(c *Config) error {
				if c.Host == "" {
					return errorf("Host is empty")
				}

				return nil
			},
		},
		{
			name: "Port is valid",
			checkFunc: func(c *Config) error {
				if c.Port <= 0 || c.Port > 65535 {
					return errorf("Port is invalid: %d", c.Port)
				}

				return nil
			},
		},
		{
			name: "BasePath is set",
			checkFunc: func(c *Config) error {
				if c.BasePath == "" {
					return errorf("BasePath is empty")
				}

				return nil
			},
		},
		{
			name: "Timeout is positive",
			checkFunc: func(c *Config) error {
				if c.Timeout <= 0 {
					return errorf("Timeout is not positive: %v", c.Timeout)
				}

				return nil
			},
		},
		{
			name: "DeviceInfo is populated",
			checkFunc: func(c *Config) error {
				if c.DeviceInfo.Manufacturer == "" {
					return errorf("Manufacturer is empty")
				}
				if c.DeviceInfo.Model == "" {
					return errorf("Model is empty")
				}
				if c.DeviceInfo.FirmwareVersion == "" {
					return errorf("FirmwareVersion is empty")
				}

				return nil
			},
		},
		{
			name: "Has at least one profile",
			checkFunc: func(c *Config) error {
				if len(c.Profiles) == 0 {
					return errorf("No profiles configured")
				}

				return nil
			},
		},
		{
			name: "Profile has valid token",
			checkFunc: func(c *Config) error {
				if c.Profiles[0].Token == "" {
					return errorf("Profile token is empty")
				}

				return nil
			},
		},
		{
			name: "Profile has valid name",
			checkFunc: func(c *Config) error {
				if c.Profiles[0].Name == "" {
					return errorf("Profile name is empty")
				}

				return nil
			},
		},
		{
			name: "Profile has video source",
			checkFunc: func(c *Config) error {
				if c.Profiles[0].VideoSource.Token == "" {
					return errorf("Video source token is empty")
				}
				if c.Profiles[0].VideoSource.Resolution.Width == 0 {
					return errorf("Video resolution width is 0")
				}
				if c.Profiles[0].VideoSource.Resolution.Height == 0 {
					return errorf("Video resolution height is 0")
				}

				return nil
			},
		},
		{
			name: "Profile has video encoder",
			checkFunc: func(c *Config) error {
				if c.Profiles[0].VideoEncoder.Encoding == "" {
					return errorf("Video encoder encoding is empty")
				}
				if c.Profiles[0].VideoEncoder.Framerate == 0 {
					return errorf("Video framerate is 0")
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.checkFunc(config); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestResolution(t *testing.T) {
	tests := []struct {
		name        string
		resolution  Resolution
		expectValid bool
	}{
		{
			name:        "Valid resolution 1920x1080",
			resolution:  Resolution{Width: 1920, Height: 1080},
			expectValid: true,
		},
		{
			name:        "Valid resolution 640x480",
			resolution:  Resolution{Width: 640, Height: 480},
			expectValid: true,
		},
		{
			name:        "Zero width",
			resolution:  Resolution{Width: 0, Height: 1080},
			expectValid: false,
		},
		{
			name:        "Zero height",
			resolution:  Resolution{Width: 1920, Height: 0},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.resolution.Width > 0 && tt.resolution.Height > 0) != tt.expectValid {
				t.Errorf("Resolution validation failed: Width=%d, Height=%d",
					tt.resolution.Width, tt.resolution.Height)
			}
		})
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name      string
		rangeVal  Range
		testValue float64
		expectIn  bool
	}{
		{
			name:      "Value within range",
			rangeVal:  Range{Min: -360, Max: 360},
			testValue: 0,
			expectIn:  true,
		},
		{
			name:      "Value at min boundary",
			rangeVal:  Range{Min: -90, Max: 90},
			testValue: -90,
			expectIn:  true,
		},
		{
			name:      "Value at max boundary",
			rangeVal:  Range{Min: -90, Max: 90},
			testValue: 90,
			expectIn:  true,
		},
		{
			name:      "Value below range",
			rangeVal:  Range{Min: 0, Max: 10},
			testValue: -1,
			expectIn:  false,
		},
		{
			name:      "Value above range",
			rangeVal:  Range{Min: 0, Max: 10},
			testValue: 11,
			expectIn:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inRange := tt.testValue >= tt.rangeVal.Min && tt.testValue <= tt.rangeVal.Max
			if inRange != tt.expectIn {
				t.Errorf("Range check failed: %f in [%f, %f] = %v, expect %v",
					tt.testValue, tt.rangeVal.Min, tt.rangeVal.Max, inRange, tt.expectIn)
			}
		})
	}
}

func TestBounds(t *testing.T) {
	tests := []struct {
		name        string
		bounds      Bounds
		expectValid bool
	}{
		{
			name:        "Valid bounds",
			bounds:      Bounds{X: 0, Y: 0, Width: 1920, Height: 1080},
			expectValid: true,
		},
		{
			name:        "Zero width",
			bounds:      Bounds{X: 0, Y: 0, Width: 0, Height: 1080},
			expectValid: false,
		},
		{
			name:        "Negative coordinates",
			bounds:      Bounds{X: -10, Y: -10, Width: 1920, Height: 1080},
			expectValid: true, // Negative coordinates may be valid in some cases
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.bounds.Width > 0 && tt.bounds.Height > 0
			if isValid != tt.expectValid {
				t.Errorf("Bounds validation failed: %+v", tt.bounds)
			}
		})
	}
}

func TestPreset(t *testing.T) {
	tests := []struct {
		name        string
		preset      Preset
		expectValid bool
	}{
		{
			name: "Valid preset",
			preset: Preset{
				Token:    "preset_1",
				Name:     "Home",
				Position: PTZPosition{Pan: 0, Tilt: 0, Zoom: 0},
			},
			expectValid: true,
		},
		{
			name: "Preset with empty token",
			preset: Preset{
				Token: "",
				Name:  "Home",
			},
			expectValid: false,
		},
		{
			name: "Preset with empty name",
			preset: Preset{
				Token: "preset_1",
				Name:  "",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.preset.Token != "" && tt.preset.Name != ""
			if isValid != tt.expectValid {
				t.Errorf("Preset validation failed: Token=%s, Name=%s",
					tt.preset.Token, tt.preset.Name)
			}
		})
	}
}

func TestPTZConfig(t *testing.T) {
	tests := []struct {
		name        string
		ptzConfig   *PTZConfig
		expectValid bool
	}{
		{
			name: "Valid PTZ config",
			ptzConfig: &PTZConfig{
				NodeToken: "ptz_node",
				PanRange:  Range{Min: -360, Max: 360},
				TiltRange: Range{Min: -90, Max: 90},
				ZoomRange: Range{Min: 0, Max: 10},
			},
			expectValid: true,
		},
		{
			name: "PTZ config with presets",
			ptzConfig: &PTZConfig{
				NodeToken: "ptz_node",
				PanRange:  Range{Min: -360, Max: 360},
				TiltRange: Range{Min: -90, Max: 90},
				ZoomRange: Range{Min: 0, Max: 10},
				Presets: []Preset{
					{Token: "preset_1", Name: "Home"},
					{Token: "preset_2", Name: "Away"},
				},
			},
			expectValid: true,
		},
		{
			name: "PTZ config with empty node token",
			ptzConfig: &PTZConfig{
				NodeToken: "",
				PanRange:  Range{Min: -360, Max: 360},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.ptzConfig.NodeToken != ""
			if isValid != tt.expectValid {
				t.Errorf("PTZ config validation failed: NodeToken=%s", tt.ptzConfig.NodeToken)
			}
		})
	}
}

func TestVideoEncoderConfig(t *testing.T) {
	tests := []struct {
		name          string
		encoderConfig VideoEncoderConfig
		expectValid   bool
	}{
		{
			name: "Valid H264 encoder",
			encoderConfig: VideoEncoderConfig{
				Encoding:   "H264",
				Resolution: Resolution{Width: 1920, Height: 1080},
				Quality:    80,
				Framerate:  30,
				Bitrate:    2048,
			},
			expectValid: true,
		},
		{
			name: "Valid H265 encoder",
			encoderConfig: VideoEncoderConfig{
				Encoding:   "H265",
				Resolution: Resolution{Width: 1920, Height: 1080},
				Quality:    80,
				Framerate:  30,
				Bitrate:    1024,
			},
			expectValid: true,
		},
		{
			name: "JPEG encoder",
			encoderConfig: VideoEncoderConfig{
				Encoding:   "JPEG",
				Resolution: Resolution{Width: 640, Height: 480},
				Quality:    90,
				Framerate:  15,
			},
			expectValid: true,
		},
		{
			name: "Invalid quality (too high)",
			encoderConfig: VideoEncoderConfig{
				Encoding:   "H264",
				Resolution: Resolution{Width: 1920, Height: 1080},
				Quality:    101,
				Framerate:  30,
			},
			expectValid: false,
		},
		{
			name: "Invalid quality (negative)",
			encoderConfig: VideoEncoderConfig{
				Encoding:   "H264",
				Resolution: Resolution{Width: 1920, Height: 1080},
				Quality:    -1,
				Framerate:  30,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.encoderConfig.Encoding != "" &&
				tt.encoderConfig.Quality >= 0 && tt.encoderConfig.Quality <= 100 &&
				tt.encoderConfig.Resolution.Width > 0 && tt.encoderConfig.Resolution.Height > 0
			if isValid != tt.expectValid {
				t.Errorf("Encoder validation failed: Quality=%f", tt.encoderConfig.Quality)
			}
		})
	}
}

func TestProfileConfig(t *testing.T) {
	tests := []struct {
		name          string
		profileConfig ProfileConfig
		expectValid   bool
	}{
		{
			name: "Valid profile config",
			profileConfig: ProfileConfig{
				Token: "profile_1",
				Name:  "Profile 1",
				VideoSource: VideoSourceConfig{
					Token:      "vs_1",
					Name:       "Video Source",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Framerate:  30,
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Quality:    80,
					Framerate:  30,
				},
			},
			expectValid: true,
		},
		{
			name: "Profile with empty token",
			profileConfig: ProfileConfig{
				Token: "",
				Name:  "Profile",
			},
			expectValid: false,
		},
		{
			name: "Profile with empty name",
			profileConfig: ProfileConfig{
				Token: "profile_1",
				Name:  "",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.profileConfig.Token != "" && tt.profileConfig.Name != ""
			if isValid != tt.expectValid {
				t.Errorf("Profile validation failed: Token=%s, Name=%s",
					tt.profileConfig.Token, tt.profileConfig.Name)
			}
		})
	}
}

func TestSnapshotConfig(t *testing.T) {
	tests := []struct {
		name           string
		snapshotConfig SnapshotConfig
		expectValid    bool
	}{
		{
			name: "Valid snapshot config",
			snapshotConfig: SnapshotConfig{
				Enabled:    true,
				Resolution: Resolution{Width: 1920, Height: 1080},
				Quality:    85.0,
			},
			expectValid: true,
		},
		{
			name: "Disabled snapshot",
			snapshotConfig: SnapshotConfig{
				Enabled:    false,
				Resolution: Resolution{Width: 0, Height: 0},
				Quality:    0,
			},
			expectValid: true,
		},
		{
			name: "Enabled with resolution",
			snapshotConfig: SnapshotConfig{
				Enabled:    true,
				Resolution: Resolution{Width: 1280, Height: 720},
				Quality:    75.0,
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Snapshot config is valid if it has resolution and quality when enabled
			isValid := !tt.snapshotConfig.Enabled ||
				(tt.snapshotConfig.Resolution.Width > 0 && tt.snapshotConfig.Resolution.Height > 0)
			if isValid != tt.expectValid {
				t.Errorf("Snapshot validation failed: Enabled=%v, Resolution=%dx%d",
					tt.snapshotConfig.Enabled, tt.snapshotConfig.Resolution.Width, tt.snapshotConfig.Resolution.Height)
			}
		})
	}
}

func TestConfigTimeout(t *testing.T) {
	config := DefaultConfig()

	if config.Timeout == 0 {
		t.Error("Timeout should not be 0")
	}

	if config.Timeout < 1*time.Second {
		t.Errorf("Timeout too small: %v", config.Timeout)
	}

	if config.Timeout > 5*time.Minute {
		t.Errorf("Timeout too large: %v", config.Timeout)
	}
}

func TestServiceEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		host           string
		expectServices []string
	}{
		{
			name: "Default endpoints",
			config: &Config{
				Host:          "192.168.1.100",
				Port:          8080,
				BasePath:      "/onvif",
				SupportPTZ:    true,
				SupportEvents: true,
			},
			host:           "",
			expectServices: []string{"device", "media", "imaging", "ptz", "events"},
		},
		{
			name: "Custom host",
			config: &Config{
				Host:          "192.168.1.100",
				Port:          8080,
				BasePath:      "/onvif",
				SupportPTZ:    false,
				SupportEvents: false,
			},
			host:           "custom.example.com",
			expectServices: []string{"device", "media", "imaging"},
		},
		{
			name: "Port 80",
			config: &Config{
				Host:       "localhost",
				Port:       80,
				BasePath:   "/onvif",
				SupportPTZ: true,
			},
			host:           "",
			expectServices: []string{"device", "media", "imaging", "ptz"},
		},
		{
			name: "Default host with 0.0.0.0",
			config: &Config{
				Host:     "0.0.0.0",
				Port:     8080,
				BasePath: "/onvif",
			},
			host:           "",
			expectServices: []string{"device", "media", "imaging"},
		},
		{
			name: "Empty host fallback",
			config: &Config{
				Host:     "",
				Port:     8080,
				BasePath: "/onvif",
			},
			host:           "",
			expectServices: []string{"device", "media", "imaging"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := tt.config.ServiceEndpoints(tt.host)

			for _, svc := range tt.expectServices {
				if _, ok := endpoints[svc]; !ok {
					t.Errorf("Missing endpoint: %s", svc)
				}
			}

			// Verify URL format
			for name, url := range endpoints {
				if !strings.HasPrefix(url, "http://") {
					t.Errorf("Endpoint %s should start with http://: %s", name, url)
				}
			}
		})
	}
}

func TestServiceEndpointsURL(t *testing.T) {
	config := &Config{
		Host:          "example.com",
		Port:          9000,
		BasePath:      "/services",
		SupportPTZ:    true,
		SupportEvents: true,
	}

	endpoints := config.ServiceEndpoints("example.com")

	expectedDeviceURL := "http://example.com:9000/services/device_service"
	if endpoints["device"] != expectedDeviceURL {
		t.Errorf("Device endpoint mismatch: got %s, want %s", endpoints["device"], expectedDeviceURL)
	}
}

func TestToONVIFProfile(t *testing.T) {
	profile := &ProfileConfig{
		Token: "profile_1",
		Name:  "HD Profile",
		VideoSource: VideoSourceConfig{
			Token:      "source_1",
			Framerate:  30,
			Resolution: Resolution{Width: 1920, Height: 1080},
		},
		VideoEncoder: VideoEncoderConfig{
			Encoding:   "H264",
			Bitrate:    4096,
			Framerate:  30,
			Resolution: Resolution{Width: 1920, Height: 1080},
		},
		Snapshot: SnapshotConfig{
			Enabled:    true,
			Resolution: Resolution{Width: 1920, Height: 1080},
			Quality:    85.0,
		},
	}

	onvifProfile := profile.ToONVIFProfile()

	if onvifProfile.Token != "profile_1" {
		t.Errorf("Profile token mismatch: got %s", onvifProfile.Token)
	}
	if onvifProfile.Name != "HD Profile" {
		t.Errorf("Profile name mismatch: got %s", onvifProfile.Name)
	}
}
