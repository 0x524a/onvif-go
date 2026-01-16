package server

import (
	"encoding/xml"
	"testing"
)

func TestHandleGetProfiles(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetProfiles(nil)
	if err != nil {
		t.Fatalf("HandleGetProfiles() error = %v", err)
	}

	profilesResp, ok := resp.(*GetProfilesResponse)
	if !ok {
		t.Fatalf("Response is not GetProfilesResponse, got %T", resp)
	}

	if len(profilesResp.Profiles) != len(config.Profiles) {
		t.Errorf("Profile count mismatch: got %d, want %d", len(profilesResp.Profiles), len(config.Profiles))
	}

	// Check first profile
	if len(profilesResp.Profiles) > 0 {
		profile := profilesResp.Profiles[0]
		if profile.Token != config.Profiles[0].Token {
			t.Errorf("Profile token mismatch: got %s, want %s", profile.Token, config.Profiles[0].Token)
		}
		if profile.Name != config.Profiles[0].Name {
			t.Errorf("Profile name mismatch: got %s, want %s", profile.Name, config.Profiles[0].Name)
		}
	}
}

func TestHandleGetStreamURI(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	profileToken := config.Profiles[0].Token

	// Create SOAP body with profile token
	reqXML := `<GetStreamURI><ProfileToken>` + profileToken + `</ProfileToken></GetStreamURI>`
	resp, err := server.HandleGetStreamURI([]byte(reqXML))
	if err != nil {
		t.Fatalf("HandleGetStreamURI() error = %v", err)
	}

	streamResp, ok := resp.(*GetStreamURIResponse)
	if !ok {
		t.Fatalf("Response is not GetStreamURIResponse, got %T", resp)
	}

	if streamResp.MediaURI.URI == "" {
		t.Error("Stream URI is empty")

		return
	}

	// URI should contain stream path
	if !contains(streamResp.MediaURI.URI, "rtsp://") {
		t.Errorf("Invalid stream URI format: %s", streamResp.MediaURI.URI)
	}
}

func TestHandleGetSnapshotURI(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)
	profileToken := config.Profiles[0].Token

	reqXML := `<GetSnapshotURI><ProfileToken>` + profileToken + `</ProfileToken></GetSnapshotURI>`
	resp, err := server.HandleGetSnapshotURI([]byte(reqXML))
	if err != nil {
		t.Fatalf("HandleGetSnapshotURI() error = %v", err)
	}

	snapResp, ok := resp.(*GetSnapshotURIResponse)
	if !ok {
		t.Fatalf("Response is not GetSnapshotURIResponse, got %T", resp)
	}

	if snapResp.MediaURI.URI == "" {
		t.Error("Snapshot URI is empty")
	}
}

func TestHandleGetVideoSources(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetVideoSources(nil)
	if err != nil {
		t.Fatalf("HandleGetVideoSources() error = %v", err)
	}

	sourcesResp, ok := resp.(*GetVideoSourcesResponse)
	if !ok {
		t.Fatalf("Response is not GetVideoSourcesResponse, got %T", resp)
	}

	if len(sourcesResp.VideoSources) == 0 {
		t.Error("No video sources returned")

		return
	}

	source := sourcesResp.VideoSources[0]
	if source.Token != config.Profiles[0].VideoSource.Token {
		t.Errorf("Video source token mismatch: got %s, want %s",
			source.Token, config.Profiles[0].VideoSource.Token)
	}

	// Check resolution
	if source.Resolution.Width != config.Profiles[0].VideoSource.Resolution.Width {
		t.Errorf("Width mismatch: got %d, want %d",
			source.Resolution.Width, config.Profiles[0].VideoSource.Resolution.Width)
	}
	if source.Resolution.Height != config.Profiles[0].VideoSource.Resolution.Height {
		t.Errorf("Height mismatch: got %d, want %d",
			source.Resolution.Height, config.Profiles[0].VideoSource.Resolution.Height)
	}

	// Check framerate
	if source.Framerate != float64(config.Profiles[0].VideoSource.Framerate) {
		t.Errorf("Framerate mismatch: got %f, want %d",
			source.Framerate, config.Profiles[0].VideoSource.Framerate)
	}
}

func TestMediaProfileStructure(t *testing.T) {
	profile := MediaProfile{
		Token: "profile_1",
		Fixed: true,
		Name:  "Profile 1",
		VideoSourceConfiguration: &VideoSourceConfiguration{
			Token:       "vs_1",
			SourceToken: "vs_1",
			Bounds: IntRectangle{
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			},
		},
		VideoEncoderConfiguration: &VideoEncoderConfiguration{
			Token:    "ve_1",
			Encoding: "H264",
			Resolution: VideoResolution{
				Width:  1920,
				Height: 1080,
			},
			Quality: 80,
		},
	}

	if profile.Token == "" {
		t.Error("Profile token is empty")
	}
	if profile.VideoSourceConfiguration == nil {
		t.Error("VideoSourceConfiguration is nil")
	}
	if profile.VideoEncoderConfiguration == nil {
		t.Error("VideoEncoderConfiguration is nil")
	}
	if profile.VideoEncoderConfiguration.Encoding == "" {
		t.Error("Video encoding is empty")
	}
}

func TestVideoEncoderConfigurationStructure(t *testing.T) {
	cfg := VideoEncoderConfiguration{
		Token:      "ve_1",
		Name:       "Video Encoder 1",
		Encoding:   "H264",
		Quality:    80,
		Resolution: VideoResolution{Width: 1920, Height: 1080},
		RateControl: &VideoRateControl{
			FrameRateLimit:   30,
			EncodingInterval: 1,
			BitrateLimit:     2048,
		},
	}

	if cfg.Token == "" {
		t.Error("Encoder token is empty")
	}
	if cfg.Encoding != "H264" {
		t.Errorf("Expected H264, got %s", cfg.Encoding)
	}
	if cfg.RateControl == nil {
		t.Error("RateControl is nil")
	}
	if cfg.RateControl.FrameRateLimit != 30 {
		t.Errorf("FrameRateLimit mismatch: got %d, want 30", cfg.RateControl.FrameRateLimit)
	}
}

func TestGetProfilesResponseXML(t *testing.T) {
	resp := &GetProfilesResponse{
		Profiles: []MediaProfile{
			{
				Token: "profile_1",
				Name:  "Profile 1",
			},
		},
	}

	// Marshal to XML
	data, err := xml.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Should contain necessary XML elements
	xmlStr := string(data)
	if !contains(xmlStr, "GetProfilesResponse") {
		t.Error("Response element not in XML")
	}
	if !contains(xmlStr, "Profiles") {
		t.Error("Profiles element not in XML")
	}
	if !contains(xmlStr, "profile_1") {
		t.Error("Profile token not in XML")
	}
}

func TestIntRectangle(t *testing.T) {
	tests := []struct {
		name        string
		rect        IntRectangle
		expectValid bool
	}{
		{
			name:        "Valid rectangle",
			rect:        IntRectangle{X: 0, Y: 0, Width: 100, Height: 100},
			expectValid: true,
		},
		{
			name:        "Zero width",
			rect:        IntRectangle{X: 0, Y: 0, Width: 0, Height: 100},
			expectValid: false,
		},
		{
			name:        "Zero height",
			rect:        IntRectangle{X: 0, Y: 0, Width: 100, Height: 0},
			expectValid: false,
		},
		{
			name:        "Negative dimensions",
			rect:        IntRectangle{X: -10, Y: -10, Width: 100, Height: 100},
			expectValid: true, // Negative coordinates may be valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.rect.Width > 0 && tt.rect.Height > 0
			if isValid != tt.expectValid {
				t.Errorf("Rectangle validation failed: Width=%d, Height=%d", tt.rect.Width, tt.rect.Height)
			}
		})
	}
}

func TestVideoResolution(t *testing.T) {
	tests := []struct {
		name        string
		resolution  VideoResolution
		expectValid bool
	}{
		{
			name:        "1080p",
			resolution:  VideoResolution{Width: 1920, Height: 1080},
			expectValid: true,
		},
		{
			name:        "720p",
			resolution:  VideoResolution{Width: 1280, Height: 720},
			expectValid: true,
		},
		{
			name:        "VGA",
			resolution:  VideoResolution{Width: 640, Height: 480},
			expectValid: true,
		},
		{
			name:        "4K",
			resolution:  VideoResolution{Width: 3840, Height: 2160},
			expectValid: true,
		},
		{
			name:        "Zero width",
			resolution:  VideoResolution{Width: 0, Height: 1080},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.resolution.Width > 0 && tt.resolution.Height > 0
			if isValid != tt.expectValid {
				t.Errorf("Resolution validation failed: %dx%d", tt.resolution.Width, tt.resolution.Height)
			}
		})
	}
}

func TestMulticastConfiguration(t *testing.T) {
	cfg := MulticastConfiguration{
		Address:   IPAddress{IPv4Address: "239.255.255.250"},
		Port:      1900,
		TTL:       128,
		AutoStart: true,
	}

	if cfg.Address.IPv4Address == "" && cfg.Address.IPv6Address == "" {
		t.Error("Multicast address is empty")
	}
	if cfg.Port == 0 {
		t.Error("Multicast port is 0")
	}
	if cfg.TTL < 1 {
		t.Error("TTL is invalid")
	}
}

func TestHandleGetProfilesDetails(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetProfiles(nil)
	if err != nil {
		t.Fatalf("HandleGetProfiles error: %v", err)
	}

	profilesResp, ok := resp.(*GetProfilesResponse)
	if !ok {
		t.Fatalf("Response is not GetProfilesResponse: %T", resp)
	}

	if len(profilesResp.Profiles) == 0 {
		t.Error("No profiles returned")
	}

	// Check profile structure
	for _, profile := range profilesResp.Profiles {
		if profile.Token == "" {
			t.Error("Profile token is empty")
		}
		if profile.Name == "" {
			t.Error("Profile name is empty")
		}
		if profile.VideoSourceConfiguration == nil {
			t.Error("VideoSourceConfiguration is nil")
		}
		if profile.VideoEncoderConfiguration == nil {
			t.Error("VideoEncoderConfiguration is nil")
		}
	}
}

func TestHandleGetVideoSourcesDetails(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetVideoSources(nil)
	if err != nil {
		t.Fatalf("HandleGetVideoSources error: %v", err)
	}

	sourcesResp, ok := resp.(*GetVideoSourcesResponse)
	if !ok {
		t.Fatalf("Response is not GetVideoSourcesResponse: %T", resp)
	}

	if len(sourcesResp.VideoSources) == 0 {
		t.Error("No video sources returned")
	}

	for _, source := range sourcesResp.VideoSources {
		if source.Token == "" {
			t.Error("VideoSource token is empty")
		}
	}
}

func TestStreamURIEdgeCases(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	// Test with invalid profile token
	reqXML := `<GetStreamURI><ProfileToken>invalid_token</ProfileToken></GetStreamURI>`
	resp, err := server.HandleGetStreamURI([]byte(reqXML))

	if err == nil {
		t.Error("Expected error for invalid profile token")
	}
	if resp != nil {
		t.Error("Expected nil response for error case")
	}
}

func TestSnapshotURIEdgeCases(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	// Test with invalid profile token
	reqXML := `<GetSnapshotURI><ProfileToken>invalid_token</ProfileToken></GetSnapshotURI>`
	resp, err := server.HandleGetSnapshotURI([]byte(reqXML))

	if err == nil {
		t.Error("Expected error for invalid profile token")
	}
	if resp != nil {
		t.Error("Expected nil response for error case")
	}
}
