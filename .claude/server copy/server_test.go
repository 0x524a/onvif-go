package server

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "New with nil config uses default",
			config:      nil,
			expectError: false,
		},
		{
			name:        "New with custom config",
			config:      createTestConfig(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := New(tt.config)
			if (err != nil) != tt.expectError {
				t.Errorf("New() error = %v, expectError %v", err, tt.expectError)

				return
			}
			if server == nil && !tt.expectError {
				t.Error("New() returned nil server")

				return
			}
			if server != nil && server.config == nil {
				t.Error("New() server.config is nil")
			}
		})
	}
}

func TestNewInitializesStreamsAndState(t *testing.T) {
	config := createTestConfig()
	server, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Verify streams are initialized
	if len(server.streams) != len(config.Profiles) {
		t.Errorf("Expected %d streams, got %d", len(config.Profiles), len(server.streams))
	}

	// Verify each stream has correct configuration
	for _, profile := range config.Profiles {
		stream, ok := server.streams[profile.Token]
		if !ok {
			t.Errorf("Stream not found for profile %s", profile.Token)

			continue
		}
		if stream.ProfileToken != profile.Token {
			t.Errorf("Stream profile token mismatch: %s != %s", stream.ProfileToken, profile.Token)
		}
	}

	// Verify PTZ state is initialized for profiles with PTZ
	for _, profile := range config.Profiles {
		if profile.PTZ != nil {
			_, ok := server.ptzState[profile.Token]
			if !ok {
				t.Errorf("PTZ state not found for profile %s", profile.Token)
			}
		}
	}

	// Verify imaging state is initialized
	if len(server.imagingState) != len(config.Profiles) {
		t.Errorf("Expected %d imaging states, got %d", len(config.Profiles), len(server.imagingState))
	}
}

func TestGetConfig(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	got := server.GetConfig()
	if got != config {
		t.Error("GetConfig() returned different config")
	}
	if got.Profiles[0].Name != config.Profiles[0].Name {
		t.Errorf("GetConfig() profile name mismatch: %s != %s", got.Profiles[0].Name, config.Profiles[0].Name)
	}
}

func TestGetStreamConfig(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	profileToken := config.Profiles[0].Token

	tests := []struct {
		name      string
		token     string
		expectOk  bool
		checkFunc func(*StreamConfig) error
	}{
		{
			name:     "Get existing stream",
			token:    profileToken,
			expectOk: true,
			checkFunc: func(sc *StreamConfig) error {
				if sc.ProfileToken != profileToken {
					return errorf("profile token mismatch: %s != %s", sc.ProfileToken, profileToken)
				}
				if sc.StreamURI == "" {
					return errorf("StreamURI is empty")
				}

				return nil
			},
		},
		{
			name:     "Get non-existent stream",
			token:    "invalid-token",
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, ok := server.GetStreamConfig(tt.token)
			if ok != tt.expectOk {
				t.Errorf("GetStreamConfig() ok = %v, expectOk %v", ok, tt.expectOk)

				return
			}
			if ok && tt.checkFunc != nil {
				if err := tt.checkFunc(stream); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestUpdateStreamURI(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	profileToken := config.Profiles[0].Token

	tests := []struct {
		name        string
		token       string
		newURI      string
		expectError bool
	}{
		{
			name:        "Update existing stream URI",
			token:       profileToken,
			newURI:      "rtsp://localhost:8554/newstream",
			expectError: false,
		},
		{
			name:        "Update non-existent stream",
			token:       "invalid-token",
			newURI:      "rtsp://localhost:8554/stream",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.UpdateStreamURI(tt.token, tt.newURI)
			if (err != nil) != tt.expectError {
				t.Errorf("UpdateStreamURI() error = %v, expectError %v", err, tt.expectError)

				return
			}
			if !tt.expectError {
				stream, _ := server.GetStreamConfig(tt.token)
				if stream.StreamURI != tt.newURI {
					t.Errorf("UpdateStreamURI() failed: %s != %s", stream.StreamURI, tt.newURI)
				}
			}
		})
	}
}

func TestListProfiles(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	profiles := server.ListProfiles()

	if len(profiles) != len(config.Profiles) {
		t.Errorf("ListProfiles() length = %d, want %d", len(profiles), len(config.Profiles))
	}

	for i, profile := range profiles {
		if profile.Token != config.Profiles[i].Token {
			t.Errorf("ListProfiles()[%d] token mismatch: %s != %s", i, profile.Token, config.Profiles[i].Token)
		}
		if profile.Name != config.Profiles[i].Name {
			t.Errorf("ListProfiles()[%d] name mismatch: %s != %s", i, profile.Name, config.Profiles[i].Name)
		}
	}
}

func TestGetPTZState(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	// Find a profile with PTZ
	var profileWithPTZ string
	for _, profile := range config.Profiles {
		if profile.PTZ != nil {
			profileWithPTZ = profile.Token

			break
		}
	}

	if profileWithPTZ == "" {
		// Create config with PTZ
		config.Profiles[0].PTZ = &PTZConfig{
			NodeToken: "ptz_node",
			PanRange:  Range{Min: -360, Max: 360},
			TiltRange: Range{Min: -90, Max: 90},
			ZoomRange: Range{Min: 0, Max: 10},
		}
		server, _ = New(config)
		profileWithPTZ = config.Profiles[0].Token
	}

	tests := []struct {
		name     string
		token    string
		expectOk bool
	}{
		{
			name:     "Get PTZ state for profile with PTZ",
			token:    profileWithPTZ,
			expectOk: true,
		},
		{
			name:     "Get PTZ state for non-existent profile",
			token:    "invalid-token",
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, ok := server.GetPTZState(tt.token)
			if ok != tt.expectOk {
				t.Errorf("GetPTZState() ok = %v, expectOk %v", ok, tt.expectOk)

				return
			}
			if ok && state == nil {
				t.Error("GetPTZState() returned nil state")
			}
		})
	}
}

func TestGetImagingState(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	videoSourceToken := config.Profiles[0].VideoSource.Token

	tests := []struct {
		name      string
		token     string
		expectOk  bool
		checkFunc func(*ImagingState) error
	}{
		{
			name:     "Get imaging state for existing source",
			token:    videoSourceToken,
			expectOk: true,
			checkFunc: func(state *ImagingState) error {
				if state.Brightness < 0 || state.Brightness > 100 {
					return errorf("brightness out of range: %f", state.Brightness)
				}
				if state.Contrast < 0 || state.Contrast > 100 {
					return errorf("contrast out of range: %f", state.Contrast)
				}

				return nil
			},
		},
		{
			name:     "Get imaging state for non-existent source",
			token:    "invalid-token",
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, ok := server.GetImagingState(tt.token)
			if ok != tt.expectOk {
				t.Errorf("GetImagingState() ok = %v, expectOk %v", ok, tt.expectOk)

				return
			}
			if ok && tt.checkFunc != nil {
				if err := tt.checkFunc(state); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestServerInfo(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	info := server.ServerInfo()

	if info == "" {
		t.Error("ServerInfo() returned empty string")
	}

	// Check that key information is present
	if !contains(info, config.DeviceInfo.Manufacturer) {
		t.Errorf("ServerInfo() missing manufacturer: %s", config.DeviceInfo.Manufacturer)
	}
	if !contains(info, config.DeviceInfo.Model) {
		t.Errorf("ServerInfo() missing model: %s", config.DeviceInfo.Model)
	}
	if !contains(info, config.Profiles[0].Name) {
		t.Errorf("ServerInfo() missing profile name: %s", config.Profiles[0].Name)
	}
}

func TestStartContextTimeout(t *testing.T) {
	config := createTestConfig()
	config.Port = 0 // Use random port
	server, _ := New(config)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start should return due to context timeout
	err := server.Start(ctx)
	if err != nil {
		t.Logf("Start() error (expected): %v", err)
	}
}

// Helper functions

func createTestConfig() *Config {
	return &Config{
		Host:     "127.0.0.1",
		Port:     8080,
		BasePath: "/onvif",
		Timeout:  30 * time.Second,
		DeviceInfo: DeviceInfo{
			Manufacturer:    "Test",
			Model:           "TestCamera",
			FirmwareVersion: "1.0.0",
			SerialNumber:    "12345",
			HardwareID:      "HW001",
		},
		Username: "admin",
		Password: "password",
		Profiles: []ProfileConfig{
			{
				Token: "profile_token_1",
				Name:  "Profile 1",
				VideoSource: VideoSourceConfig{
					Token:      "video_source_1",
					Name:       "Video Source 1",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Framerate:  30,
					Bounds: Bounds{
						X:      0,
						Y:      0,
						Width:  1920,
						Height: 1080,
					},
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Quality:    80,
					Framerate:  30,
					Bitrate:    2048,
					GovLength:  30,
				},
				PTZ: &PTZConfig{
					NodeToken: "ptz_node_1",
					PanRange:  Range{Min: -360, Max: 360},
					TiltRange: Range{Min: -90, Max: 90},
					ZoomRange: Range{Min: 0, Max: 10},
				},
				Snapshot: SnapshotConfig{
					Enabled:    true,
					Resolution: Resolution{Width: 1920, Height: 1080},
					Quality:    85.0,
				},
			},
		},
		SupportPTZ:     true,
		SupportImaging: true,
		SupportEvents:  false,
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func errorf(format string, args ...interface{}) error {
	return &testError{msg: fmt.Sprintf(format, args...)}
}

func TestServerInfoMethod(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	info := server.ServerInfo()

	if info == "" {
		t.Fatal("ServerInfo() returned empty string")
	}

	// ServerInfo returns a formatted string with server information
	if !strings.Contains(info, "127.0.0.1") && !strings.Contains(info, "localhost") {
		t.Logf("ServerInfo may not contain host: %s", info)
	}
}

func TestGettersAndSetters(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	// Test GetConfig
	cfg := server.GetConfig()
	if cfg == nil {
		t.Error("GetConfig returned nil")
	}

	// Test GetStreamConfig
	streamCfg, _ := server.GetStreamConfig(config.Profiles[0].Token)
	if streamCfg == nil {
		t.Error("GetStreamConfig returned nil")
	}

	// Test UpdateStreamURI
	newURI := "rtsp://example.com/stream"
	server.UpdateStreamURI(config.Profiles[0].Token, newURI)
	updated, _ := server.GetStreamConfig(config.Profiles[0].Token)
	if updated.StreamURI != newURI {
		t.Errorf("UpdateStreamURI failed: got %s, want %s", updated.StreamURI, newURI)
	}

	// Test ListProfiles
	profiles := server.ListProfiles()
	if len(profiles) == 0 {
		t.Error("ListProfiles returned empty list")
	}

	// Test GetPTZState
	ptzState, _ := server.GetPTZState(config.Profiles[0].Token)
	if ptzState == nil {
		t.Error("GetPTZState returned nil")
	}

	// Test GetImagingState
	imgState, _ := server.GetImagingState(config.Profiles[0].VideoSource.Token)
	if imgState == nil {
		t.Error("GetImagingState returned nil")
	}
}
