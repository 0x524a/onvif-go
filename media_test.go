package onvif

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testProfileToken = "Profile1"
)

// TestGetProfiles tests GetProfiles operation.
func TestGetProfiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetProfilesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Profiles token="Profile1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Main Profile</tt:Name>
				<tt:VideoEncoderConfiguration xmlns:tt="http://www.onvif.org/ver10/schema" token="VideoEnc1">
					<tt:Encoding>H264</tt:Encoding>
					<tt:Resolution>
						<tt:Width>1920</tt:Width>
						<tt:Height>1080</tt:Height>
					</tt:Resolution>
					<tt:Quality>5.0</tt:Quality>
				</tt:VideoEncoderConfiguration>
			</trt:Profiles>
		</trt:GetProfilesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		t.Fatalf("GetProfiles() failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}

	if profiles[0].Token != testProfileToken {
		t.Errorf("Expected token %s, got %s", testProfileToken, profiles[0].Token)
	}

	if profiles[0].Name != "Main Profile" {
		t.Errorf("Expected name 'Main Profile', got %s", profiles[0].Name)
	}
}

// TestGetProfile tests GetProfile operation.
func TestGetProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetProfileResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Profile token="Profile1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Main Profile</tt:Name>
			</trt:Profile>
		</trt:GetProfileResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profile, err := client.GetProfile(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("GetProfile() failed: %v", err)
	}

	if profile.Token != testProfileToken {
		t.Errorf("Expected token Profile1, got %s", profile.Token)
	}
}

// TestSetProfile tests SetProfile operation.
func TestSetProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetProfileResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profile := &Profile{
		Token: testProfileToken,
		Name:  "Updated Profile",
	}

	err = client.SetProfile(ctx, profile)
	if err != nil {
		t.Fatalf("SetProfile() failed: %v", err)
	}
}

// TestGetStreamURI tests GetStreamURI operation.
func TestGetStreamURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:MediaUri>
				<tt:Uri xmlns:tt="http://www.onvif.org/ver10/schema">rtsp://192.168.1.100:554/stream1</tt:Uri>
				<tt:InvalidAfterConnect xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:InvalidAfterConnect>
				<tt:InvalidAfterReboot xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:InvalidAfterReboot>
			</trt:MediaUri>
		</trt:GetStreamUriResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	uri, err := client.GetStreamURI(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("GetStreamURI() failed: %v", err)
	}

	if uri.URI != "rtsp://192.168.1.100:554/stream1" {
		t.Errorf("Expected URI 'rtsp://192.168.1.100:554/stream1', got %s", uri.URI)
	}
}

// TestGetSnapshotURI tests GetSnapshotURI operation.
func TestGetSnapshotURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetSnapshotUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:MediaUri>
				<tt:Uri xmlns:tt="http://www.onvif.org/ver10/schema">http://192.168.1.100/snapshot.jpg</tt:Uri>
			</trt:MediaUri>
		</trt:GetSnapshotUriResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	uri, err := client.GetSnapshotURI(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("GetSnapshotURI() failed: %v", err)
	}

	if !strings.Contains(uri.URI, "snapshot") {
		t.Errorf("Expected snapshot URI, got %s", uri.URI)
	}
}

// TestGetVideoSources tests GetVideoSources operation.
func TestGetVideoSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoSourcesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:VideoSources token="VideoSource1">
				<tt:Framerate xmlns:tt="http://www.onvif.org/ver10/schema">30.0</tt:Framerate>
				<tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Width>1920</tt:Width>
					<tt:Height>1080</tt:Height>
				</tt:Resolution>
			</trt:VideoSources>
		</trt:GetVideoSourcesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	sources, err := client.GetVideoSources(ctx)
	if err != nil {
		t.Fatalf("GetVideoSources() failed: %v", err)
	}

	if len(sources) != 1 {
		t.Errorf("Expected 1 video source, got %d", len(sources))
	}

	if sources[0].Token != "VideoSource1" {
		t.Errorf("Expected token VideoSource1, got %s", sources[0].Token)
	}
}

// TestGetAudioSources tests GetAudioSources operation.
func TestGetAudioSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioSourcesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:AudioSources token="AudioSource1">
				<tt:Channels xmlns:tt="http://www.onvif.org/ver10/schema">2</tt:Channels>
			</trt:AudioSources>
		</trt:GetAudioSourcesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	sources, err := client.GetAudioSources(ctx)
	if err != nil {
		t.Fatalf("GetAudioSources() failed: %v", err)
	}

	if len(sources) != 1 {
		t.Errorf("Expected 1 audio source, got %d", len(sources))
	}
}

// TestGetAudioOutputs tests GetAudioOutputs operation.
func TestGetAudioOutputs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioOutputsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:AudioOutputs token="AudioOutput1"/>
		</trt:GetAudioOutputsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	outputs, err := client.GetAudioOutputs(ctx)
	if err != nil {
		t.Fatalf("GetAudioOutputs() failed: %v", err)
	}

	if len(outputs) != 1 {
		t.Errorf("Expected 1 audio output, got %d", len(outputs))
	}
}

// TestCreateProfile tests CreateProfile operation.
func TestCreateProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:CreateProfileResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Profile token="NewProfile1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">New Profile</tt:Name>
			</trt:Profile>
		</trt:CreateProfileResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profile, err := client.CreateProfile(ctx, "New Profile", "")
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	if profile.Token != "NewProfile1" {
		t.Errorf("Expected token NewProfile1, got %s", profile.Token)
	}
}

// TestDeleteProfile tests DeleteProfile operation.
func TestDeleteProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:DeleteProfileResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.DeleteProfile(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("DeleteProfile() failed: %v", err)
	}
}

// TestGetVideoEncoderConfiguration tests GetVideoEncoderConfiguration operation.
func TestGetVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoEncoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="VideoEnc1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">H264 Config</tt:Name>
				<tt:Encoding xmlns:tt="http://www.onvif.org/ver10/schema">H264</tt:Encoding>
				<tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Width>1920</tt:Width>
					<tt:Height>1080</tt:Height>
				</tt:Resolution>
				<tt:Quality xmlns:tt="http://www.onvif.org/ver10/schema">5.0</tt:Quality>
			</trt:Configuration>
		</trt:GetVideoEncoderConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.GetVideoEncoderConfiguration(ctx, testVideoEncToken)
	if err != nil {
		t.Fatalf("GetVideoEncoderConfiguration() failed: %v", err)
	}

	if config.Token != testVideoEncToken {
		t.Errorf("Expected token VideoEnc1, got %s", config.Token)
	}

	if config.Encoding != encodingH264 {
		t.Errorf("Expected encoding H264, got %s", config.Encoding)
	}
}

// TestSetVideoEncoderConfiguration tests SetVideoEncoderConfiguration operation.
func TestSetVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetVideoEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &VideoEncoderConfiguration{
		Token:    testVideoEncToken,
		Name:     "H264 Config",
		Encoding: encodingH264,
		Resolution: &VideoResolution{
			Width:  1920,
			Height: 1080,
		},
		Quality: 5.0,
	}

	err = client.SetVideoEncoderConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetVideoEncoderConfiguration() failed: %v", err)
	}
}

// TestGetMediaServiceCapabilities tests GetMediaServiceCapabilities operation.
func TestGetMediaServiceCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetServiceCapabilitiesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Capabilities SnapshotUri="true" Rotation="true" OSD="true">
				<trt:ProfileCapabilities MaximumNumberOfProfiles="10"/>
				<trt:StreamingCapabilities RTPMulticast="true" RTP_TCP="true"/>
			</trt:Capabilities>
		</trt:GetServiceCapabilitiesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	caps, err := client.GetMediaServiceCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetMediaServiceCapabilities() failed: %v", err)
	}

	if !caps.SnapshotURI {
		t.Error("Expected SnapshotURI to be true")
	}

	if caps.MaximumNumberOfProfiles != 10 {
		t.Errorf("Expected MaximumNumberOfProfiles 10, got %d", caps.MaximumNumberOfProfiles)
	}
}

// TestGetVideoEncoderConfigurationOptions tests GetVideoEncoderConfigurationOptions operation.
func TestGetVideoEncoderConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:QualityRange xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Min>1.0</tt:Min>
					<tt:Max>10.0</tt:Max>
				</tt:QualityRange>
				<tt:H264 xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:ResolutionsAvailable>
						<tt:Width>1920</tt:Width>
						<tt:Height>1080</tt:Height>
					</tt:ResolutionsAvailable>
					<tt:H264ProfilesSupported>Baseline</tt:H264ProfilesSupported>
				</tt:H264>
			</trt:Options>
		</trt:GetVideoEncoderConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.GetVideoEncoderConfigurationOptions(ctx, testVideoEncToken)
	if err != nil {
		t.Fatalf("GetVideoEncoderConfigurationOptions() failed: %v", err)
	}

	if options.QualityRange == nil {
		t.Error("Expected QualityRange to be set")
	}

	if options.H264 == nil {
		t.Error("Expected H264 options to be set")
	}
}

// TestGetAudioEncoderConfiguration tests GetAudioEncoderConfiguration operation.
func TestGetAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioEncoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="AudioEnc1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">AAC Config</tt:Name>
				<tt:Encoding xmlns:tt="http://www.onvif.org/ver10/schema">AAC</tt:Encoding>
				<tt:Bitrate xmlns:tt="http://www.onvif.org/ver10/schema">128000</tt:Bitrate>
				<tt:SampleRate xmlns:tt="http://www.onvif.org/ver10/schema">48000</tt:SampleRate>
			</trt:Configuration>
		</trt:GetAudioEncoderConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.GetAudioEncoderConfiguration(ctx, testAudioEncToken)
	if err != nil {
		t.Fatalf("GetAudioEncoderConfiguration() failed: %v", err)
	}

	if config.Token != testAudioEncToken {
		t.Errorf("Expected token AudioEnc1, got %s", config.Token)
	}

	if config.Encoding != testEncodingAAC {
		t.Errorf("Expected encoding AAC, got %s", config.Encoding)
	}
}

// TestSetAudioEncoderConfiguration tests SetAudioEncoderConfiguration operation.
func TestSetAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetAudioEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &AudioEncoderConfiguration{
		Token:      testAudioEncToken,
		Name:       "AAC Config",
		Encoding:   testEncodingAAC,
		Bitrate:    128000,
		SampleRate: 48000,
	}

	err = client.SetAudioEncoderConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetAudioEncoderConfiguration() failed: %v", err)
	}
}

// TestGetMetadataConfiguration tests GetMetadataConfiguration operation.
func TestGetMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetMetadataConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="Metadata1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Metadata Config</tt:Name>
				<tt:PTZStatus xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Status>true</tt:Status>
					<tt:Position>true</tt:Position>
				</tt:PTZStatus>
				<tt:Analytics xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:Analytics>
			</trt:Configuration>
		</trt:GetMetadataConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.GetMetadataConfiguration(ctx, "Metadata1")
	if err != nil {
		t.Fatalf("GetMetadataConfiguration() failed: %v", err)
	}

	if config.Token != "Metadata1" {
		t.Errorf("Expected token Metadata1, got %s", config.Token)
	}

	if config.PTZStatus == nil {
		t.Error("Expected PTZStatus to be set")
	}
}

// TestSetMetadataConfiguration tests SetMetadataConfiguration operation.
func TestSetMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetMetadataConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &MetadataConfiguration{
		Token:     "Metadata1",
		Name:      "Metadata Config",
		Analytics: false,
		PTZStatus: &PTZFilter{
			Status:   true,
			Position: true,
		},
	}

	err = client.SetMetadataConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetMetadataConfiguration() failed: %v", err)
	}
}

// TestGetVideoSourceModes tests GetVideoSourceModes operation.
func TestGetVideoSourceModes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoSourceModesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:VideoSourceModes token="Mode1">
				<tt:Enabled xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:Enabled>
				<tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Width>1920</tt:Width>
					<tt:Height>1080</tt:Height>
				</tt:Resolution>
			</trt:VideoSourceModes>
		</trt:GetVideoSourceModesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	modes, err := client.GetVideoSourceModes(ctx, "VideoSource1")
	if err != nil {
		t.Fatalf("GetVideoSourceModes() failed: %v", err)
	}

	if len(modes) != 1 {
		t.Errorf("Expected 1 mode, got %d", len(modes))
	}

	if modes[0].Token != "Mode1" {
		t.Errorf("Expected token Mode1, got %s", modes[0].Token)
	}
}

// TestSetVideoSourceMode tests SetVideoSourceMode operation.
func TestSetVideoSourceMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetVideoSourceModeResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.SetVideoSourceMode(ctx, "VideoSource1", "Mode1")
	if err != nil {
		t.Fatalf("SetVideoSourceMode() failed: %v", err)
	}
}

// TestSetSynchronizationPoint tests SetSynchronizationPoint operation.
func TestSetSynchronizationPoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetSynchronizationPointResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.SetSynchronizationPoint(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("SetSynchronizationPoint() failed: %v", err)
	}
}

// TestGetOSDs tests GetOSDs operation.
func TestGetOSDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetOSDsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:OSDs token="OSD1"/>
			<trt:OSDs token="OSD2"/>
		</trt:GetOSDsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osds, err := client.GetOSDs(ctx, "")
	if err != nil {
		t.Fatalf("GetOSDs() failed: %v", err)
	}

	if len(osds) != 2 {
		t.Errorf("Expected 2 OSDs, got %d", len(osds))
	}
}

// TestGetOSD tests GetOSD operation.
func TestGetOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetOSDResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:OSD token="OSD1"/>
		</trt:GetOSDResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osd, err := client.GetOSD(ctx, "OSD1")
	if err != nil {
		t.Fatalf("GetOSD() failed: %v", err)
	}

	if osd.Token != "OSD1" {
		t.Errorf("Expected token OSD1, got %s", osd.Token)
	}
}

// TestSetOSD tests SetOSD operation.
func TestSetOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetOSDResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osd := &OSDConfiguration{
		Token: "OSD1",
	}

	err = client.SetOSD(ctx, osd)
	if err != nil {
		t.Fatalf("SetOSD() failed: %v", err)
	}
}

// TestCreateOSD tests CreateOSD operation.
func TestCreateOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:CreateOSDResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:OSD token="NewOSD1"/>
		</trt:CreateOSDResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osd, err := client.CreateOSD(ctx, "VideoSourceConfig1", nil)
	if err != nil {
		t.Fatalf("CreateOSD() failed: %v", err)
	}

	if osd.Token != "NewOSD1" {
		t.Errorf("Expected token NewOSD1, got %s", osd.Token)
	}
}

// TestDeleteOSD tests DeleteOSD operation.
func TestDeleteOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:DeleteOSDResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.DeleteOSD(ctx, "OSD1")
	if err != nil {
		t.Fatalf("DeleteOSD() failed: %v", err)
	}
}

// TestStartMulticastStreaming tests StartMulticastStreaming operation.
func TestStartMulticastStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:StartMulticastStreamingResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.StartMulticastStreaming(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("StartMulticastStreaming() failed: %v", err)
	}
}

// TestStopMulticastStreaming tests StopMulticastStreaming operation.
func TestStopMulticastStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:StopMulticastStreamingResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.StopMulticastStreaming(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("StopMulticastStreaming() failed: %v", err)
	}
}

// TestAddVideoEncoderConfiguration tests AddVideoEncoderConfiguration operation.
func TestAddVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddVideoEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.AddVideoEncoderConfiguration(ctx, testProfileToken, testVideoEncToken)
	if err != nil {
		t.Fatalf("AddVideoEncoderConfiguration() failed: %v", err)
	}
}

// TestRemoveVideoEncoderConfiguration tests RemoveVideoEncoderConfiguration operation.
func TestRemoveVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveVideoEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.RemoveVideoEncoderConfiguration(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("RemoveVideoEncoderConfiguration() failed: %v", err)
	}
}

// TestAddAudioEncoderConfiguration tests AddAudioEncoderConfiguration operation.
func TestAddAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddAudioEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.AddAudioEncoderConfiguration(ctx, testProfileToken, testAudioEncToken)
	if err != nil {
		t.Fatalf("AddAudioEncoderConfiguration() failed: %v", err)
	}
}

// TestRemoveAudioEncoderConfiguration tests RemoveAudioEncoderConfiguration operation.
func TestRemoveAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveAudioEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.RemoveAudioEncoderConfiguration(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("RemoveAudioEncoderConfiguration() failed: %v", err)
	}
}

// TestAddAudioSourceConfiguration tests AddAudioSourceConfiguration operation.
func TestAddAudioSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddAudioSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.AddAudioSourceConfiguration(ctx, testProfileToken, "AudioSourceConfig1")
	if err != nil {
		t.Fatalf("AddAudioSourceConfiguration() failed: %v", err)
	}
}

// TestRemoveAudioSourceConfiguration tests RemoveAudioSourceConfiguration operation.
func TestRemoveAudioSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveAudioSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.RemoveAudioSourceConfiguration(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("RemoveAudioSourceConfiguration() failed: %v", err)
	}
}

// TestAddVideoSourceConfiguration tests AddVideoSourceConfiguration operation.
func TestAddVideoSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddVideoSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.AddVideoSourceConfiguration(ctx, testProfileToken, "VideoSourceConfig1")
	if err != nil {
		t.Fatalf("AddVideoSourceConfiguration() failed: %v", err)
	}
}

// TestRemoveVideoSourceConfiguration tests RemoveVideoSourceConfiguration operation.
func TestRemoveVideoSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveVideoSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.RemoveVideoSourceConfiguration(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("RemoveVideoSourceConfiguration() failed: %v", err)
	}
}

// TestAddPTZConfiguration tests AddPTZConfiguration operation.
func TestAddPTZConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddPTZConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.AddPTZConfiguration(ctx, testProfileToken, "PTZConfig1")
	if err != nil {
		t.Fatalf("AddPTZConfiguration() failed: %v", err)
	}
}

// TestRemovePTZConfiguration tests RemovePTZConfiguration operation.
func TestRemovePTZConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemovePTZConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.RemovePTZConfiguration(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("RemovePTZConfiguration() failed: %v", err)
	}
}

// TestAddMetadataConfiguration tests AddMetadataConfiguration operation.
func TestAddMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddMetadataConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.AddMetadataConfiguration(ctx, testProfileToken, "Metadata1")
	if err != nil {
		t.Fatalf("AddMetadataConfiguration() failed: %v", err)
	}
}

// TestRemoveMetadataConfiguration tests RemoveMetadataConfiguration operation.
func TestRemoveMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveMetadataConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.RemoveMetadataConfiguration(ctx, testProfileToken)
	if err != nil {
		t.Fatalf("RemoveMetadataConfiguration() failed: %v", err)
	}
}

// TestGetAudioEncoderConfigurationOptions tests GetAudioEncoderConfigurationOptions operation.
func TestGetAudioEncoderConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:EncodingOptions xmlns:tt="http://www.onvif.org/ver10/schema">AAC</tt:EncodingOptions>
				<tt:EncodingOptions xmlns:tt="http://www.onvif.org/ver10/schema">G711</tt:EncodingOptions>
				<tt:BitrateList xmlns:tt="http://www.onvif.org/ver10/schema">64000</tt:BitrateList>
				<tt:BitrateList xmlns:tt="http://www.onvif.org/ver10/schema">128000</tt:BitrateList>
				<tt:SampleRateList xmlns:tt="http://www.onvif.org/ver10/schema">44100</tt:SampleRateList>
				<tt:SampleRateList xmlns:tt="http://www.onvif.org/ver10/schema">48000</tt:SampleRateList>
			</trt:Options>
		</trt:GetAudioEncoderConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.GetAudioEncoderConfigurationOptions(ctx, testAudioEncToken, "")
	if err != nil {
		t.Fatalf("GetAudioEncoderConfigurationOptions() failed: %v", err)
	}

	if len(options.EncodingOptions) == 0 {
		t.Error("Expected encoding options to be set")
	}
}

// TestGetMetadataConfigurationOptions tests GetMetadataConfigurationOptions operation.
func TestGetMetadataConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetMetadataConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:PTZStatusFilterOptions xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Status>true</tt:Status>
					<tt:Position>true</tt:Position>
				</tt:PTZStatusFilterOptions>
			</trt:Options>
		</trt:GetMetadataConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.GetMetadataConfigurationOptions(ctx, "Metadata1", "")
	if err != nil {
		t.Fatalf("GetMetadataConfigurationOptions() failed: %v", err)
	}

	if options.PTZStatusFilterOptions == nil {
		t.Error("Expected PTZStatusFilterOptions to be set")
	}
}

// TestGetAudioOutputConfiguration tests GetAudioOutputConfiguration operation.
func TestGetAudioOutputConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioOutputConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="AudioOutputConfig1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Audio Output Config</tt:Name>
				<tt:OutputToken xmlns:tt="http://www.onvif.org/ver10/schema">AudioOutput1</tt:OutputToken>
			</trt:Configuration>
		</trt:GetAudioOutputConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.GetAudioOutputConfiguration(ctx, "AudioOutputConfig1")
	if err != nil {
		t.Fatalf("GetAudioOutputConfiguration() failed: %v", err)
	}

	if config.Token != "AudioOutputConfig1" {
		t.Errorf("Expected token AudioOutputConfig1, got %s", config.Token)
	}
}

// TestSetAudioOutputConfiguration tests SetAudioOutputConfiguration operation.
func TestSetAudioOutputConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetAudioOutputConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &AudioOutputConfiguration{
		Token:       "AudioOutputConfig1",
		Name:        "Audio Output Config",
		OutputToken: "AudioOutput1",
	}

	err = client.SetAudioOutputConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetAudioOutputConfiguration() failed: %v", err)
	}
}

// TestGetAudioOutputConfigurationOptions tests GetAudioOutputConfigurationOptions operation.
func TestGetAudioOutputConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioOutputConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:OutputTokensAvailable xmlns:tt="http://www.onvif.org/ver10/schema">AudioOutput1</tt:OutputTokensAvailable>
				<tt:OutputTokensAvailable xmlns:tt="http://www.onvif.org/ver10/schema">AudioOutput2</tt:OutputTokensAvailable>
			</trt:Options>
		</trt:GetAudioOutputConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.GetAudioOutputConfigurationOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetAudioOutputConfigurationOptions() failed: %v", err)
	}

	if len(options.OutputTokensAvailable) == 0 {
		t.Error("Expected output tokens to be available")
	}
}

// TestGetAudioDecoderConfigurationOptions tests GetAudioDecoderConfigurationOptions operation.
func TestGetAudioDecoderConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioDecoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:AACDecOptions xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:BitrateList>64000</tt:BitrateList>
					<tt:BitrateList>128000</tt:BitrateList>
					<tt:SampleRateList>44100</tt:SampleRateList>
					<tt:SampleRateList>48000</tt:SampleRateList>
				</tt:AACDecOptions>
			</trt:Options>
		</trt:GetAudioDecoderConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.GetAudioDecoderConfigurationOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetAudioDecoderConfigurationOptions() failed: %v", err)
	}

	if options.AACDecOptions == nil {
		t.Error("Expected AACDecOptions to be set")
	}
}

// TestGetGuaranteedNumberOfVideoEncoderInstances tests GetGuaranteedNumberOfVideoEncoderInstances operation.
func TestGetGuaranteedNumberOfVideoEncoderInstances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetGuaranteedNumberOfVideoEncoderInstancesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:TotalNumber>4</trt:TotalNumber>
			<trt:JPEG>2</trt:JPEG>
			<trt:H264>2</trt:H264>
			<trt:MPEG4>0</trt:MPEG4>
		</trt:GetGuaranteedNumberOfVideoEncoderInstancesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	instances, err := client.GetGuaranteedNumberOfVideoEncoderInstances(ctx, testVideoEncToken)
	if err != nil {
		t.Fatalf("GetGuaranteedNumberOfVideoEncoderInstances() failed: %v", err)
	}

	if instances.TotalNumber != 4 {
		t.Errorf("Expected TotalNumber 4, got %d", instances.TotalNumber)
	}

	if instances.H264 != 2 {
		t.Errorf("Expected H264 2, got %d", instances.H264)
	}
}

// TestGetOSDOptions tests GetOSDOptions operation.
func TestGetOSDOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetOSDOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:MaximumNumberOfOSDs xmlns:tt="http://www.onvif.org/ver10/schema">10</tt:MaximumNumberOfOSDs>
			</trt:Options>
		</trt:GetOSDOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.GetOSDOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetOSDOptions() failed: %v", err)
	}

	if options.MaximumNumberOfOSDs != 10 {
		t.Errorf("Expected MaximumNumberOfOSDs 10, got %d", options.MaximumNumberOfOSDs)
	}
}

// The tests below close the remaining gap in media.go's coverage (#10). Before it,
// every method added after around line 2242 had a fault-path test, but the 49
// methods above that line did not: their "<Name> failed: %w" wrapping around
// soapClient.Call had never actually been driven by a failing server. A wrong
// verb, a dropped %w, or a copy-pasted wrong name in that wrapping would have
// compiled and passed every existing test while quietly misleading callers.
//
// It also adds full optional-field mapping coverage for the five methods that
// sat furthest under 100%: GetVideoEncoderConfigurationOptions,
// GetAudioEncoderConfiguration, SetAudioEncoderConfiguration,
// GetMetadataConfiguration and SetMetadataConfiguration. Each optional branch
// is exercised both present and absent, and every field in a present branch
// carries its own distinct value so a mapping bug that swaps two same-typed
// fields (Width/Height, Port/TTL, ...) actually fails the test instead of
// passing by coincidence.
//
// The Set* mapping tests here assert what the client marshals rather than what
// it parses back, so they use newRequestCapturingServer (session_timeout_test.go)
// instead of the response-shaped servers in soap_harness_test.go.
//
// The one field an AllOptionalsPresent subtest here does not assert is
// SessionTimeout, and only because session_timeout_test.go already asserts it
// across all twelve methods that carry it - including these - as part of the
// #86 fix. Duplicating it here would add nothing.

// osdTestToken and ptzConfigTestToken back the fault-path ops below; they are
// named constants rather than repeated literals only because each is used by
// more than one op (GetOSD/SetOSD/CreateOSD/DeleteOSD, AddPTZConfiguration/
// RemovePTZConfiguration).
const (
	osdTestToken       = "osd1"
	ptzConfigTestToken = "ptzconfig1"

	// h264ProfileMain, addressTypeIPv4 and addressTypeIPv6 name values that
	// recur across several mapping assertions below.
	h264ProfileMain = "Main"
	addressTypeIPv4 = "IPv4"
	addressTypeIPv6 = "IPv6"
)

// --- Fault-path coverage: every method below 100% before these tests --------

// mediaFaultOpsProfiles covers profile lifecycle and streaming operations.
func mediaFaultOpsProfiles() []serviceOp {
	return []serviceOp{
		{"GetProfiles", func(ctx context.Context, c *Client) error {
			_, err := c.GetProfiles(ctx)

			return err
		}},
		{"GetStreamURI", func(ctx context.Context, c *Client) error {
			_, err := c.GetStreamURI(ctx, testProfileToken)

			return err
		}},
		{"GetSnapshotURI", func(ctx context.Context, c *Client) error {
			_, err := c.GetSnapshotURI(ctx, testProfileToken)

			return err
		}},
		{"CreateProfile", func(ctx context.Context, c *Client) error {
			_, err := c.CreateProfile(ctx, "profile-name", testProfileToken)

			return err
		}},
		{"DeleteProfile", func(ctx context.Context, c *Client) error {
			return c.DeleteProfile(ctx, testProfileToken)
		}},
		{"GetProfile", func(ctx context.Context, c *Client) error {
			_, err := c.GetProfile(ctx, testProfileToken)

			return err
		}},
		{"SetProfile", func(ctx context.Context, c *Client) error {
			return c.SetProfile(ctx, &Profile{Token: testProfileToken})
		}},
		{"SetSynchronizationPoint", func(ctx context.Context, c *Client) error {
			return c.SetSynchronizationPoint(ctx, testProfileToken)
		}},
		{"StartMulticastStreaming", func(ctx context.Context, c *Client) error {
			return c.StartMulticastStreaming(ctx, testProfileToken)
		}},
		{"StopMulticastStreaming", func(ctx context.Context, c *Client) error {
			return c.StopMulticastStreaming(ctx, testProfileToken)
		}},
	}
}

// mediaFaultOpsSources covers source enumeration and video source modes.
func mediaFaultOpsSources() []serviceOp {
	return []serviceOp{
		{"GetVideoSources", func(ctx context.Context, c *Client) error {
			_, err := c.GetVideoSources(ctx)

			return err
		}},
		{"GetAudioSources", func(ctx context.Context, c *Client) error {
			_, err := c.GetAudioSources(ctx)

			return err
		}},
		{"GetAudioOutputs", func(ctx context.Context, c *Client) error {
			_, err := c.GetAudioOutputs(ctx)

			return err
		}},
		{"GetMediaServiceCapabilities", func(ctx context.Context, c *Client) error {
			_, err := c.GetMediaServiceCapabilities(ctx)

			return err
		}},
		{"GetVideoSourceModes", func(ctx context.Context, c *Client) error {
			_, err := c.GetVideoSourceModes(ctx, testVideoSourceToken)

			return err
		}},
		{"SetVideoSourceMode", func(ctx context.Context, c *Client) error {
			return c.SetVideoSourceMode(ctx, testVideoSourceToken, "mode1")
		}},
	}
}

// mediaFaultOpsEncoderConfigs covers video and audio encoder configuration.
func mediaFaultOpsEncoderConfigs() []serviceOp {
	return []serviceOp{
		{"GetVideoEncoderConfiguration", func(ctx context.Context, c *Client) error {
			_, err := c.GetVideoEncoderConfiguration(ctx, testVideoEncToken)

			return err
		}},
		{"SetVideoEncoderConfiguration", func(ctx context.Context, c *Client) error {
			return c.SetVideoEncoderConfiguration(ctx, &VideoEncoderConfiguration{Token: testVideoEncToken}, true)
		}},
		{"GetVideoEncoderConfigurationOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetVideoEncoderConfigurationOptions(ctx, testVideoEncToken)

			return err
		}},
		{"GetAudioEncoderConfiguration", func(ctx context.Context, c *Client) error {
			_, err := c.GetAudioEncoderConfiguration(ctx, testAudioEncToken)

			return err
		}},
		{"SetAudioEncoderConfiguration", func(ctx context.Context, c *Client) error {
			return c.SetAudioEncoderConfiguration(ctx, &AudioEncoderConfiguration{Token: testAudioEncToken}, true)
		}},
		{"GetAudioEncoderConfigurationOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetAudioEncoderConfigurationOptions(ctx, testAudioEncToken, testProfileToken)

			return err
		}},
		{"GetVideoEncoderConfigurations", func(ctx context.Context, c *Client) error {
			_, err := c.GetVideoEncoderConfigurations(ctx)

			return err
		}},
		{"GetGuaranteedNumberOfVideoEncoderInstances", func(ctx context.Context, c *Client) error {
			_, err := c.GetGuaranteedNumberOfVideoEncoderInstances(ctx, testVideoEncToken)

			return err
		}},
	}
}

// mediaFaultOpsMetadataAndAudioOutput covers metadata and audio output
// configuration.
func mediaFaultOpsMetadataAndAudioOutput() []serviceOp {
	return []serviceOp{
		{"GetMetadataConfiguration", func(ctx context.Context, c *Client) error {
			_, err := c.GetMetadataConfiguration(ctx, testAnalyticsCfgToken)

			return err
		}},
		{"SetMetadataConfiguration", func(ctx context.Context, c *Client) error {
			return c.SetMetadataConfiguration(ctx, &MetadataConfiguration{Token: testAnalyticsCfgToken}, true)
		}},
		{"GetMetadataConfigurationOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetMetadataConfigurationOptions(ctx, testAnalyticsCfgToken, testProfileToken)

			return err
		}},
		{"GetAudioOutputConfiguration", func(ctx context.Context, c *Client) error {
			_, err := c.GetAudioOutputConfiguration(ctx, testAudioOutputToken)

			return err
		}},
		{"SetAudioOutputConfiguration", func(ctx context.Context, c *Client) error {
			return c.SetAudioOutputConfiguration(ctx, &AudioOutputConfiguration{Token: testAudioOutputToken}, true)
		}},
		{"GetAudioOutputConfigurationOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetAudioOutputConfigurationOptions(ctx, testAudioOutputToken)

			return err
		}},
		{"GetAudioDecoderConfigurationOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetAudioDecoderConfigurationOptions(ctx, testAudioDecCfgToken)

			return err
		}},
	}
}

// mediaFaultOpsOSD covers on-screen-display configuration.
func mediaFaultOpsOSD() []serviceOp {
	return []serviceOp{
		{"GetOSDs", func(ctx context.Context, c *Client) error {
			_, err := c.GetOSDs(ctx, testVideoSrcCfgToken)

			return err
		}},
		{"GetOSD", func(ctx context.Context, c *Client) error {
			_, err := c.GetOSD(ctx, osdTestToken)

			return err
		}},
		{"SetOSD", func(ctx context.Context, c *Client) error {
			return c.SetOSD(ctx, &OSDConfiguration{Token: osdTestToken})
		}},
		{"CreateOSD", func(ctx context.Context, c *Client) error {
			_, err := c.CreateOSD(ctx, testVideoSrcCfgToken, &OSDConfiguration{Token: osdTestToken})

			return err
		}},
		{"DeleteOSD", func(ctx context.Context, c *Client) error {
			return c.DeleteOSD(ctx, osdTestToken)
		}},
		{"GetOSDOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetOSDOptions(ctx, testVideoSrcCfgToken)

			return err
		}},
	}
}

// mediaFaultOpsProfileLinks covers the Add/Remove pairs that attach an
// existing configuration to a profile by token.
func mediaFaultOpsProfileLinks() []serviceOp {
	return []serviceOp{
		{"AddVideoEncoderConfiguration", func(ctx context.Context, c *Client) error {
			return c.AddVideoEncoderConfiguration(ctx, testProfileToken, testVideoEncToken)
		}},
		{"RemoveVideoEncoderConfiguration", func(ctx context.Context, c *Client) error {
			return c.RemoveVideoEncoderConfiguration(ctx, testProfileToken)
		}},
		{"AddAudioEncoderConfiguration", func(ctx context.Context, c *Client) error {
			return c.AddAudioEncoderConfiguration(ctx, testProfileToken, testAudioEncToken)
		}},
		{"RemoveAudioEncoderConfiguration", func(ctx context.Context, c *Client) error {
			return c.RemoveAudioEncoderConfiguration(ctx, testProfileToken)
		}},
		{"AddAudioSourceConfiguration", func(ctx context.Context, c *Client) error {
			return c.AddAudioSourceConfiguration(ctx, testProfileToken, testAudioSrcCfgToken)
		}},
		{"RemoveAudioSourceConfiguration", func(ctx context.Context, c *Client) error {
			return c.RemoveAudioSourceConfiguration(ctx, testProfileToken)
		}},
		{"AddVideoSourceConfiguration", func(ctx context.Context, c *Client) error {
			return c.AddVideoSourceConfiguration(ctx, testProfileToken, testVideoSrcCfgToken)
		}},
		{"RemoveVideoSourceConfiguration", func(ctx context.Context, c *Client) error {
			return c.RemoveVideoSourceConfiguration(ctx, testProfileToken)
		}},
		{"AddPTZConfiguration", func(ctx context.Context, c *Client) error {
			return c.AddPTZConfiguration(ctx, testProfileToken, ptzConfigTestToken)
		}},
		{"RemovePTZConfiguration", func(ctx context.Context, c *Client) error {
			return c.RemovePTZConfiguration(ctx, testProfileToken)
		}},
		{"AddMetadataConfiguration", func(ctx context.Context, c *Client) error {
			return c.AddMetadataConfiguration(ctx, testProfileToken, testAnalyticsCfgToken)
		}},
		{"RemoveMetadataConfiguration", func(ctx context.Context, c *Client) error {
			return c.RemoveMetadataConfiguration(ctx, testProfileToken)
		}},
	}
}

// TestMediaOperationsSOAPFault drives every media method that was below 100%
// coverage against a server that answers with a SOAP fault, and asserts each
// one surfaces a non-nil error naming itself - per its own
// "<Name> failed: %w" wrapping around soapClient.Call. A method whose wrapping
// dropped the name, dropped %w, or named the wrong operation would still pass
// every other existing test in this package; this is the only test that would
// catch it.
func TestMediaOperationsSOAPFault(t *testing.T) {
	ops := make([]serviceOp, 0, 49)
	ops = append(ops, mediaFaultOpsProfiles()...)
	ops = append(ops, mediaFaultOpsSources()...)
	ops = append(ops, mediaFaultOpsEncoderConfigs()...)
	ops = append(ops, mediaFaultOpsMetadataAndAudioOutput()...)
	ops = append(ops, mediaFaultOpsOSD()...)
	ops = append(ops, mediaFaultOpsProfileLinks()...)

	server := newSOAPFaultTestServer(t)

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			err := op.call(ctx, client)
			if err == nil {
				t.Fatalf("%s() against a faulting server = nil error, want non-nil", op.name)
			}
			if !strings.Contains(err.Error(), op.name+" failed") {
				t.Errorf("%s() error = %q, want it to name the operation via %q", op.name, err, op.name+" failed")
			}
		})
	}
}

// --- Optional-field mapping coverage ----------------------------------------

// TestGetVideoEncoderConfigurationOptionsMapping covers every optional branch
// of VideoEncoderConfigurationOptions: QualityRange, JPEG (with its nested
// ResolutionsAvailable/FrameRateRange/EncodingIntervalRange) and H264 (with
// its nested ResolutionsAvailable/GovLengthRange/FrameRateRange/
// EncodingIntervalRange/H264ProfilesSupported).
func TestGetVideoEncoderConfigurationOptionsMapping(t *testing.T) {
	t.Run("AllOptionalsPresent", func(t *testing.T) {
		const response = `<trt:GetVideoEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Options>
		<trt:QualityRange>
			<trt:Min>1</trt:Min>
			<trt:Max>2</trt:Max>
		</trt:QualityRange>
		<trt:JPEG>
			<trt:ResolutionsAvailable>
				<trt:Width>3</trt:Width>
				<trt:Height>4</trt:Height>
			</trt:ResolutionsAvailable>
			<trt:FrameRateRange>
				<trt:Min>5</trt:Min>
				<trt:Max>6</trt:Max>
			</trt:FrameRateRange>
			<trt:EncodingIntervalRange>
				<trt:Min>7</trt:Min>
				<trt:Max>8</trt:Max>
			</trt:EncodingIntervalRange>
		</trt:JPEG>
		<trt:H264>
			<trt:ResolutionsAvailable>
				<trt:Width>9</trt:Width>
				<trt:Height>10</trt:Height>
			</trt:ResolutionsAvailable>
			<trt:GovLengthRange>
				<trt:Min>11</trt:Min>
				<trt:Max>12</trt:Max>
			</trt:GovLengthRange>
			<trt:FrameRateRange>
				<trt:Min>13</trt:Min>
				<trt:Max>14</trt:Max>
			</trt:FrameRateRange>
			<trt:EncodingIntervalRange>
				<trt:Min>15</trt:Min>
				<trt:Max>16</trt:Max>
			</trt:EncodingIntervalRange>
			<trt:H264ProfilesSupported>Baseline</trt:H264ProfilesSupported>
			<trt:H264ProfilesSupported>Main</trt:H264ProfilesSupported>
		</trt:H264>
	</trt:Options>
</trt:GetVideoEncoderConfigurationOptionsResponse>`

		server := newSOAPTestServer(t, response)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		opts, err := client.GetVideoEncoderConfigurationOptions(context.Background(), testVideoEncToken)
		if err != nil {
			t.Fatalf("GetVideoEncoderConfigurationOptions() failed: %v", err)
		}

		assertVideoEncoderConfigurationOptionsPresent(t, opts)
	})

	t.Run("OptionalsAbsent", func(t *testing.T) {
		const response = `<trt:GetVideoEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Options></trt:Options>
</trt:GetVideoEncoderConfigurationOptionsResponse>`

		server := newSOAPTestServer(t, response)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		opts, err := client.GetVideoEncoderConfigurationOptions(context.Background(), testVideoEncToken)
		if err != nil {
			t.Fatalf("GetVideoEncoderConfigurationOptions() failed: %v", err)
		}

		if opts.QualityRange != nil {
			t.Errorf("QualityRange = %+v, want nil", opts.QualityRange)
		}
		if opts.JPEG != nil {
			t.Errorf("JPEG = %+v, want nil", opts.JPEG)
		}
		if opts.H264 != nil {
			t.Errorf("H264 = %+v, want nil", opts.H264)
		}
	})
}

// assertVideoEncoderConfigurationOptionsPresent checks every field of opts
// against the sentinel values encoded in the "AllOptionalsPresent" response
// XML above: QualityRange={1,2}, JPEG={{3,4}},{5,6},{7,8},
// H264={{9,10}},{11,12},{13,14},{15,16},[Baseline,Main].
func assertVideoEncoderConfigurationOptionsPresent(t *testing.T, opts *VideoEncoderConfigurationOptions) {
	t.Helper()

	if opts.QualityRange == nil || opts.QualityRange.Min != 1 || opts.QualityRange.Max != 2 {
		t.Errorf("QualityRange = %+v, want {Min:1 Max:2}", opts.QualityRange)
	}

	if opts.JPEG == nil {
		t.Fatal("JPEG = nil, want populated")
	}
	if len(opts.JPEG.ResolutionsAvailable) != 1 ||
		opts.JPEG.ResolutionsAvailable[0].Width != 3 || opts.JPEG.ResolutionsAvailable[0].Height != 4 {
		t.Errorf("JPEG.ResolutionsAvailable = %+v, want [{Width:3 Height:4}]", opts.JPEG.ResolutionsAvailable)
	}
	if opts.JPEG.FrameRateRange == nil || opts.JPEG.FrameRateRange.Min != 5 || opts.JPEG.FrameRateRange.Max != 6 {
		t.Errorf("JPEG.FrameRateRange = %+v, want {Min:5 Max:6}", opts.JPEG.FrameRateRange)
	}
	if opts.JPEG.EncodingIntervalRange == nil ||
		opts.JPEG.EncodingIntervalRange.Min != 7 || opts.JPEG.EncodingIntervalRange.Max != 8 {
		t.Errorf("JPEG.EncodingIntervalRange = %+v, want {Min:7 Max:8}", opts.JPEG.EncodingIntervalRange)
	}

	if opts.H264 == nil {
		t.Fatal("H264 = nil, want populated")
	}
	if len(opts.H264.ResolutionsAvailable) != 1 ||
		opts.H264.ResolutionsAvailable[0].Width != 9 || opts.H264.ResolutionsAvailable[0].Height != 10 {
		t.Errorf("H264.ResolutionsAvailable = %+v, want [{Width:9 Height:10}]", opts.H264.ResolutionsAvailable)
	}
	if opts.H264.GovLengthRange == nil || opts.H264.GovLengthRange.Min != 11 || opts.H264.GovLengthRange.Max != 12 {
		t.Errorf("H264.GovLengthRange = %+v, want {Min:11 Max:12}", opts.H264.GovLengthRange)
	}
	if opts.H264.FrameRateRange == nil || opts.H264.FrameRateRange.Min != 13 || opts.H264.FrameRateRange.Max != 14 {
		t.Errorf("H264.FrameRateRange = %+v, want {Min:13 Max:14}", opts.H264.FrameRateRange)
	}
	if opts.H264.EncodingIntervalRange == nil ||
		opts.H264.EncodingIntervalRange.Min != 15 || opts.H264.EncodingIntervalRange.Max != 16 {
		t.Errorf("H264.EncodingIntervalRange = %+v, want {Min:15 Max:16}", opts.H264.EncodingIntervalRange)
	}

	wantProfiles := []string{"Baseline", h264ProfileMain}
	gotProfiles := opts.H264.H264ProfilesSupported
	if len(gotProfiles) != len(wantProfiles) || gotProfiles[0] != wantProfiles[0] || gotProfiles[1] != wantProfiles[1] {
		t.Errorf("H264.H264ProfilesSupported = %v, want %v", gotProfiles, wantProfiles)
	}
}

// TestGetAudioEncoderConfigurationMapping covers Multicast and its nested
// Address, the one optional branch of AudioEncoderConfiguration.
func TestGetAudioEncoderConfigurationMapping(t *testing.T) {
	t.Run("AllOptionalsPresent", func(t *testing.T) {
		const response = `<trt:GetAudioEncoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Configuration token="AECToken1">
		<trt:Name>AECName2</trt:Name>
		<trt:UseCount>3</trt:UseCount>
		<trt:Encoding>AAC</trt:Encoding>
		<trt:Bitrate>4</trt:Bitrate>
		<trt:SampleRate>5</trt:SampleRate>
		<trt:Multicast>
			<trt:Address>
				<trt:Type>IPv4</trt:Type>
				<trt:IPv4Address>10.0.0.6</trt:IPv4Address>
				<trt:IPv6Address>::7</trt:IPv6Address>
			</trt:Address>
			<trt:Port>8</trt:Port>
			<trt:TTL>9</trt:TTL>
			<trt:AutoStart>true</trt:AutoStart>
		</trt:Multicast>
	</trt:Configuration>
</trt:GetAudioEncoderConfigurationResponse>`

		server := newSOAPTestServer(t, response)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		cfg, err := client.GetAudioEncoderConfiguration(context.Background(), testAudioEncToken)
		if err != nil {
			t.Fatalf("GetAudioEncoderConfiguration() failed: %v", err)
		}

		if cfg.Token != "AECToken1" || cfg.Name != "AECName2" {
			t.Errorf("Token/Name = %q/%q, want %q/%q", cfg.Token, cfg.Name, "AECToken1", "AECName2")
		}
		if cfg.UseCount != 3 || cfg.Bitrate != 4 || cfg.SampleRate != 5 {
			t.Errorf("UseCount/Bitrate/SampleRate = %d/%d/%d, want 3/4/5", cfg.UseCount, cfg.Bitrate, cfg.SampleRate)
		}
		if cfg.Encoding != testEncodingAAC {
			t.Errorf("Encoding = %q, want %q", cfg.Encoding, testEncodingAAC)
		}

		if cfg.Multicast == nil {
			t.Fatal("Multicast = nil, want populated")
		}
		if cfg.Multicast.Port != 8 || cfg.Multicast.TTL != 9 || !cfg.Multicast.AutoStart {
			t.Errorf("Multicast = %+v, want {Port:8 TTL:9 AutoStart:true ...}", cfg.Multicast)
		}
		if cfg.Multicast.Address == nil {
			t.Fatal("Multicast.Address = nil, want populated")
		}
		if cfg.Multicast.Address.Type != addressTypeIPv4 ||
			cfg.Multicast.Address.IPv4Address != "10.0.0.6" || cfg.Multicast.Address.IPv6Address != "::7" {
			t.Errorf("Multicast.Address = %+v, want {Type:IPv4 IPv4Address:10.0.0.6 IPv6Address:::7}", cfg.Multicast.Address)
		}
	})

	t.Run("OptionalsAbsent", func(t *testing.T) {
		const response = `<trt:GetAudioEncoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Configuration token="AECToken1">
		<trt:Name>AECName2</trt:Name>
	</trt:Configuration>
</trt:GetAudioEncoderConfigurationResponse>`

		server := newSOAPTestServer(t, response)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		cfg, err := client.GetAudioEncoderConfiguration(context.Background(), testAudioEncToken)
		if err != nil {
			t.Fatalf("GetAudioEncoderConfiguration() failed: %v", err)
		}

		if cfg.Multicast != nil {
			t.Errorf("Multicast = %+v, want nil", cfg.Multicast)
		}
	})
}

// TestGetMetadataConfigurationMapping covers all three optional branches of
// MetadataConfiguration: PTZStatus, Events (a presence-only marker with no
// fields of its own) and Multicast with its nested Address.
func TestGetMetadataConfigurationMapping(t *testing.T) {
	t.Run("AllOptionalsPresent", func(t *testing.T) {
		const response = `<trt:GetMetadataConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Configuration token="MDCToken1">
		<trt:Name>MDCName2</trt:Name>
		<trt:UseCount>3</trt:UseCount>
		<trt:PTZStatus>
			<trt:Status>true</trt:Status>
			<trt:Position>false</trt:Position>
		</trt:PTZStatus>
		<trt:Events></trt:Events>
		<trt:Analytics>true</trt:Analytics>
		<trt:Multicast>
			<trt:Address>
				<trt:Type>IPv6</trt:Type>
				<trt:IPv4Address>10.0.0.4</trt:IPv4Address>
				<trt:IPv6Address>::5</trt:IPv6Address>
			</trt:Address>
			<trt:Port>6</trt:Port>
			<trt:TTL>7</trt:TTL>
			<trt:AutoStart>true</trt:AutoStart>
		</trt:Multicast>
	</trt:Configuration>
</trt:GetMetadataConfigurationResponse>`

		server := newSOAPTestServer(t, response)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		cfg, err := client.GetMetadataConfiguration(context.Background(), testAnalyticsCfgToken)
		if err != nil {
			t.Fatalf("GetMetadataConfiguration() failed: %v", err)
		}

		if cfg.Token != "MDCToken1" || cfg.Name != "MDCName2" || cfg.UseCount != 3 {
			t.Errorf("Token/Name/UseCount = %q/%q/%d, want %q/%q/3", cfg.Token, cfg.Name, cfg.UseCount, "MDCToken1", "MDCName2")
		}
		if !cfg.Analytics {
			t.Error("Analytics = false, want true")
		}

		if cfg.PTZStatus == nil {
			t.Fatal("PTZStatus = nil, want populated")
		}
		if !cfg.PTZStatus.Status {
			t.Error("PTZStatus.Status = false, want true")
		}
		if cfg.PTZStatus.Position {
			t.Error("PTZStatus.Position = true, want false")
		}

		if cfg.Events == nil {
			t.Error("Events = nil, want populated (the response element was present)")
		}

		if cfg.Multicast == nil {
			t.Fatal("Multicast = nil, want populated")
		}
		if cfg.Multicast.Port != 6 || cfg.Multicast.TTL != 7 || !cfg.Multicast.AutoStart {
			t.Errorf("Multicast = %+v, want {Port:6 TTL:7 AutoStart:true ...}", cfg.Multicast)
		}
		if cfg.Multicast.Address == nil ||
			cfg.Multicast.Address.Type != addressTypeIPv6 ||
			cfg.Multicast.Address.IPv4Address != "10.0.0.4" ||
			cfg.Multicast.Address.IPv6Address != "::5" {
			t.Errorf("Multicast.Address = %+v, want {Type:IPv6 IPv4Address:10.0.0.4 IPv6Address:::5}", cfg.Multicast.Address)
		}
	})

	t.Run("OptionalsAbsent", func(t *testing.T) {
		const response = `<trt:GetMetadataConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Configuration token="MDCToken1">
		<trt:Name>MDCName2</trt:Name>
	</trt:Configuration>
</trt:GetMetadataConfigurationResponse>`

		server := newSOAPTestServer(t, response)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		cfg, err := client.GetMetadataConfiguration(context.Background(), testAnalyticsCfgToken)
		if err != nil {
			t.Fatalf("GetMetadataConfiguration() failed: %v", err)
		}

		if cfg.PTZStatus != nil {
			t.Errorf("PTZStatus = %+v, want nil", cfg.PTZStatus)
		}
		if cfg.Events != nil {
			t.Errorf("Events = %+v, want nil", cfg.Events)
		}
		if cfg.Multicast != nil {
			t.Errorf("Multicast = %+v, want nil", cfg.Multicast)
		}
	})
}

// TestSetAudioEncoderConfigurationMapping covers how SetAudioEncoderConfiguration
// marshals Multicast (and its nested Address) into the outgoing request, and
// that Bitrate/SampleRate/Multicast are omitted rather than zero-valued when
// unset on the source config.
func TestSetAudioEncoderConfigurationMapping(t *testing.T) {
	t.Run("AllOptionalsPresent", func(t *testing.T) {
		var gotBody string

		server := newRequestCapturingServer(t, &gotBody)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		config := &AudioEncoderConfiguration{
			Token:      "SAECToken1",
			Name:       "SAECName2",
			UseCount:   3,
			Encoding:   testEncodingAAC,
			Bitrate:    4,
			SampleRate: 5,
			Multicast: &MulticastConfiguration{
				Address: &IPAddress{
					Type:        addressTypeIPv4,
					IPv4Address: "10.0.0.6",
					IPv6Address: "::7",
				},
				Port:      8,
				TTL:       9,
				AutoStart: true,
			},
		}

		if err := client.SetAudioEncoderConfiguration(context.Background(), config, true); err != nil {
			t.Fatalf("SetAudioEncoderConfiguration() failed: %v", err)
		}

		for _, want := range []string{
			`token="SAECToken1"`,
			"<tt:Name>SAECName2</tt:Name>",
			"<tt:UseCount>3</tt:UseCount>",
			"<tt:Encoding>AAC</tt:Encoding>",
			"<tt:Bitrate>4</tt:Bitrate>",
			"<tt:SampleRate>5</tt:SampleRate>",
			"<tt:Type>IPv4</tt:Type>",
			"<tt:IPv4Address>10.0.0.6</tt:IPv4Address>",
			"<tt:IPv6Address>::7</tt:IPv6Address>",
			"<tt:Port>8</tt:Port>",
			"<tt:TTL>9</tt:TTL>",
			"<tt:AutoStart>true</tt:AutoStart>",
			"<trt:ForcePersistence>true</trt:ForcePersistence>",
		} {
			if !strings.Contains(gotBody, want) {
				t.Errorf("request body missing %q\ngot: %s", want, gotBody)
			}
		}
	})

	t.Run("OptionalsAbsent", func(t *testing.T) {
		var gotBody string

		server := newRequestCapturingServer(t, &gotBody)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		config := &AudioEncoderConfiguration{
			Token:    "SAECToken1",
			Name:     "SAECName2",
			Encoding: testEncodingAAC,
		}

		if err := client.SetAudioEncoderConfiguration(context.Background(), config, false); err != nil {
			t.Fatalf("SetAudioEncoderConfiguration() failed: %v", err)
		}

		for _, unwanted := range []string{"Bitrate", "SampleRate", "Multicast"} {
			if strings.Contains(gotBody, unwanted) {
				t.Errorf("request body contains %q, want omitted when unset on the source config\ngot: %s", unwanted, gotBody)
			}
		}
	})
}

// TestSetMetadataConfigurationMapping covers how SetMetadataConfiguration
// marshals PTZStatus, Events and Multicast (with its nested Address) into the
// outgoing request, and that all three are omitted when unset on the source
// config.
func TestSetMetadataConfigurationMapping(t *testing.T) {
	t.Run("AllOptionalsPresent", func(t *testing.T) {
		var gotBody string

		server := newRequestCapturingServer(t, &gotBody)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		config := &MetadataConfiguration{
			Token:     "SMDCToken1",
			Name:      "SMDCName2",
			UseCount:  3,
			Analytics: true,
			PTZStatus: &PTZFilter{
				Status:   true,
				Position: false,
			},
			Events: &EventSubscription{},
			Multicast: &MulticastConfiguration{
				Address: &IPAddress{
					Type:        addressTypeIPv6,
					IPv4Address: "10.0.0.4",
					IPv6Address: "::5",
				},
				Port:      6,
				TTL:       7,
				AutoStart: true,
			},
		}

		if err := client.SetMetadataConfiguration(context.Background(), config, true); err != nil {
			t.Fatalf("SetMetadataConfiguration() failed: %v", err)
		}

		for _, want := range []string{
			`token="SMDCToken1"`,
			"<tt:Name>SMDCName2</tt:Name>",
			"<tt:UseCount>3</tt:UseCount>",
			"<tt:Analytics>true</tt:Analytics>",
			"<tt:Status>true</tt:Status>",
			"<tt:Position>false</tt:Position>",
			"<tt:Type>IPv6</tt:Type>",
			"<tt:IPv4Address>10.0.0.4</tt:IPv4Address>",
			"<tt:IPv6Address>::5</tt:IPv6Address>",
			"<tt:Port>6</tt:Port>",
			"<tt:TTL>7</tt:TTL>",
			"<tt:AutoStart>true</tt:AutoStart>",
			"<trt:ForcePersistence>true</trt:ForcePersistence>",
		} {
			if !strings.Contains(gotBody, want) {
				t.Errorf("request body missing %q\ngot: %s", want, gotBody)
			}
		}
		if !strings.Contains(gotBody, "tt:Events") {
			t.Errorf("request body missing an Events element for a non-nil config.Events\ngot: %s", gotBody)
		}
	})

	t.Run("OptionalsAbsent", func(t *testing.T) {
		var gotBody string

		server := newRequestCapturingServer(t, &gotBody)

		client, err := NewClient(server.URL + "/onvif/media_service")
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		config := &MetadataConfiguration{
			Token: "SMDCToken1",
			Name:  "SMDCName2",
		}

		if err := client.SetMetadataConfiguration(context.Background(), config, false); err != nil {
			t.Fatalf("SetMetadataConfiguration() failed: %v", err)
		}

		for _, unwanted := range []string{"PTZStatus", "Events", "Multicast"} {
			if strings.Contains(gotBody, unwanted) {
				t.Errorf("request body contains %q, want omitted when unset on the source config\ngot: %s", unwanted, gotBody)
			}
		}
	})
}

// --- Remaining single-branch gaps -------------------------------------------
//
// The four tests below each cover the one optional branch that was still
// uncovered in an otherwise-100% method once the fault-path and priority
// mapping tests above landed: GetProfiles.PTZConfiguration,
// SetVideoEncoderConfiguration.RateControl,
// GetAudioDecoderConfigurationOptions.G726DecOptions and
// GetVideoEncoderConfigurations.MPEG4. None of these were in the original
// priority list, but each is a one nil-guard fix once identified from
// go tool cover -func output, so there was no reason to leave them out.

// TestGetProfilesPTZConfigurationMapping covers the PTZConfiguration branch of
// GetProfiles - the only optional field of a profile that no existing test
// (in this file) populates.
func TestGetProfilesPTZConfigurationMapping(t *testing.T) {
	const response = `<trt:GetProfilesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Profiles token="ProfileToken1">
		<trt:Name>ProfileName2</trt:Name>
		<trt:PTZConfiguration token="PTZCfgToken3">
			<trt:Name>PTZCfgName4</trt:Name>
			<trt:UseCount>5</trt:UseCount>
			<trt:NodeToken>PTZNode6</trt:NodeToken>
		</trt:PTZConfiguration>
	</trt:Profiles>
</trt:GetProfilesResponse>`

	server := newSOAPTestServer(t, response)

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	profiles, err := client.GetProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetProfiles() failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("len(profiles) = %d, want 1", len(profiles))
	}

	ptz := profiles[0].PTZConfiguration
	if ptz == nil {
		t.Fatal("PTZConfiguration = nil, want populated")
	}
	if ptz.Token != "PTZCfgToken3" || ptz.Name != "PTZCfgName4" || ptz.UseCount != 5 || ptz.NodeToken != "PTZNode6" {
		t.Errorf("PTZConfiguration = %+v, want {Token:PTZCfgToken3 Name:PTZCfgName4 UseCount:5 NodeToken:PTZNode6}", ptz)
	}
}

// TestSetVideoEncoderConfigurationRateControlMapping covers the RateControl
// branch of SetVideoEncoderConfiguration's request marshaling - the only
// optional field of that request no existing test exercises.
func TestSetVideoEncoderConfigurationRateControlMapping(t *testing.T) {
	var gotBody string

	server := newRequestCapturingServer(t, &gotBody)

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	config := &VideoEncoderConfiguration{
		Token: testVideoEncToken,
		RateControl: &VideoRateControl{
			FrameRateLimit:   1,
			EncodingInterval: 2,
			BitrateLimit:     3,
		},
	}

	if err := client.SetVideoEncoderConfiguration(context.Background(), config, true); err != nil {
		t.Fatalf("SetVideoEncoderConfiguration() failed: %v", err)
	}

	for _, want := range []string{
		"<tt:FrameRateLimit>1</tt:FrameRateLimit>",
		"<tt:EncodingInterval>2</tt:EncodingInterval>",
		"<tt:BitrateLimit>3</tt:BitrateLimit>",
	} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("request body missing %q\ngot: %s", want, gotBody)
		}
	}
}

// TestGetAudioDecoderConfigurationOptionsG726Mapping covers the G726DecOptions
// branch - AACDecOptions and G711DecOptions are already covered elsewhere in
// this package, but nothing populates G726DecOptions.
func TestGetAudioDecoderConfigurationOptionsG726Mapping(t *testing.T) {
	const response = `<trt:GetAudioDecoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Options>
		<trt:G726DecOptions>
			<trt:BitrateList>16</trt:BitrateList>
			<trt:BitrateList>32</trt:BitrateList>
		</trt:G726DecOptions>
	</trt:Options>
</trt:GetAudioDecoderConfigurationOptionsResponse>`

	server := newSOAPTestServer(t, response)

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	opts, err := client.GetAudioDecoderConfigurationOptions(context.Background(), testAudioDecCfgToken)
	if err != nil {
		t.Fatalf("GetAudioDecoderConfigurationOptions() failed: %v", err)
	}

	if opts.G726DecOptions == nil {
		t.Fatal("G726DecOptions = nil, want populated")
	}

	want := []int{16, 32}
	got := opts.G726DecOptions.BitrateList
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("G726DecOptions.BitrateList = %v, want %v", got, want)
	}
}

// TestGetVideoEncoderConfigurationsMPEG4Mapping covers the MPEG4 branch of
// GetVideoEncoderConfigurations - Resolution, RateControl, H264 and Multicast
// are already covered elsewhere in this package, but nothing populates MPEG4.
func TestGetVideoEncoderConfigurationsMPEG4Mapping(t *testing.T) {
	const response = `<trt:GetVideoEncoderConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
	<trt:Configurations token="VECToken1">
		<trt:Name>VECName2</trt:Name>
		<trt:MPEG4>
			<trt:GovLength>3</trt:GovLength>
			<trt:MPEG4Profile>SP</trt:MPEG4Profile>
		</trt:MPEG4>
	</trt:Configurations>
</trt:GetVideoEncoderConfigurationsResponse>`

	server := newSOAPTestServer(t, response)

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	configs, err := client.GetVideoEncoderConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetVideoEncoderConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("len(configs) = %d, want 1", len(configs))
	}

	mpeg4 := configs[0].MPEG4
	if mpeg4 == nil {
		t.Fatal("MPEG4 = nil, want populated")
	}
	if mpeg4.GovLength != 3 || mpeg4.MPEG4Profile != "SP" {
		t.Errorf("MPEG4 = %+v, want {GovLength:3 MPEG4Profile:SP}", mpeg4)
	}
}

// Test tokens shared across the configuration-related media tests below. Centralizing them as
// constants avoids goconst churn from the many happy-path/fault-path test pairs that reference
// the same tokens.
const (
	testVideoSrcCfgToken  = "VideoSrcCfg1"
	testVideoSourceToken  = "VideoSource1"
	testAudioSrcCfgToken  = "AudioSrcCfg1"
	testAudioSourceToken  = "AudioSource1"
	testAudioDecCfgToken  = "AudioDecCfg1"
	testAnalyticsCfgToken = "AnalyticsCfg1"
	testAudioOutputToken  = "AudioOutput1"
	testEncodingAAC       = "AAC"
	testVideoEncToken     = "VideoEnc1"
	testAudioEncToken     = "AudioEnc1"
)

// newMediaConfigClient creates a Client pointed at the given test server's media service endpoint.
func newMediaConfigClient(t *testing.T, serverURL string) *Client {
	t.Helper()

	client, err := NewClient(serverURL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	return client
}

// newMediaConfigServer starts a test server that always responds with the given SOAP body.
func newMediaConfigServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
}

// newMediaConfigFaultServer starts a test server that always responds with a SOAP fault and a
// non-200 HTTP status, which is how this library's SOAP client surfaces errors to callers.
func newMediaConfigFaultServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<soap:Fault>
			<soap:Code><soap:Value>soap:Receiver</soap:Value></soap:Code>
			<soap:Reason><soap:Text>Internal error</soap:Text></soap:Reason>
		</soap:Fault>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(response))
	}))
}

// ---- GetVideoSourceConfigurations ----

func TestGetVideoSourceConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoSourceConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="VideoSrcCfg1">
				<trt:Name>Video Source Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:SourceToken>VideoSource1</trt:SourceToken>
				<trt:Bounds x="0" y="0" width="1920" height="1080"/>
			</trt:Configurations>
		</trt:GetVideoSourceConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetVideoSourceConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetVideoSourceConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testVideoSrcCfgToken {
		t.Errorf("Expected token VideoSrcCfg1, got %s", configs[0].Token)
	}

	if configs[0].SourceToken != testVideoSourceToken {
		t.Errorf("Expected source token VideoSource1, got %s", configs[0].SourceToken)
	}

	if configs[0].Bounds == nil {
		t.Fatal("Expected Bounds to be set")
	}

	if configs[0].Bounds.Width != 1920 || configs[0].Bounds.Height != 1080 {
		t.Errorf("Expected bounds 1920x1080, got %dx%d", configs[0].Bounds.Width, configs[0].Bounds.Height)
	}
}

func TestGetVideoSourceConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetVideoSourceConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetVideoSourceConfigurations(), got nil")
	}
}

// ---- GetAudioSourceConfigurations ----

func TestGetAudioSourceConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioSourceConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioSrcCfg1">
				<trt:Name>Audio Source Config</trt:Name>
				<trt:UseCount>2</trt:UseCount>
				<trt:SourceToken>AudioSource1</trt:SourceToken>
			</trt:Configurations>
		</trt:GetAudioSourceConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetAudioSourceConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetAudioSourceConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testAudioSrcCfgToken {
		t.Errorf("Expected token AudioSrcCfg1, got %s", configs[0].Token)
	}

	if configs[0].UseCount != 2 {
		t.Errorf("Expected UseCount 2, got %d", configs[0].UseCount)
	}

	if configs[0].SourceToken != testAudioSourceToken {
		t.Errorf("Expected source token AudioSource1, got %s", configs[0].SourceToken)
	}
}

func TestGetAudioSourceConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetAudioSourceConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetAudioSourceConfigurations(), got nil")
	}
}

// ---- GetVideoEncoderConfigurations ----

func TestGetVideoEncoderConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoEncoderConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="VideoEnc1">
				<trt:Name>H264 Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:Encoding>H264</trt:Encoding>
				<trt:Resolution>
					<trt:Width>1920</trt:Width>
					<trt:Height>1080</trt:Height>
				</trt:Resolution>
				<trt:Quality>5.0</trt:Quality>
				<trt:RateControl>
					<trt:FrameRateLimit>30</trt:FrameRateLimit>
					<trt:EncodingInterval>1</trt:EncodingInterval>
					<trt:BitrateLimit>4096</trt:BitrateLimit>
				</trt:RateControl>
				<trt:H264>
					<trt:GovLength>60</trt:GovLength>
					<trt:H264Profile>Main</trt:H264Profile>
				</trt:H264>
				<trt:Multicast>
					<trt:Address>
						<trt:Type>IPv4</trt:Type>
						<trt:IPv4Address>239.0.0.1</trt:IPv4Address>
					</trt:Address>
					<trt:Port>10000</trt:Port>
					<trt:TTL>16</trt:TTL>
					<trt:AutoStart>false</trt:AutoStart>
				</trt:Multicast>
			</trt:Configurations>
		</trt:GetVideoEncoderConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetVideoEncoderConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetVideoEncoderConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	cfg := configs[0]

	if cfg.Token != testVideoEncToken || cfg.Encoding != encodingH264 {
		t.Errorf("Expected token VideoEnc1/H264, got %s/%s", cfg.Token, cfg.Encoding)
	}

	if cfg.Resolution == nil || cfg.Resolution.Width != 1920 || cfg.Resolution.Height != 1080 {
		t.Errorf("Expected resolution 1920x1080, got %+v", cfg.Resolution)
	}

	if cfg.RateControl == nil || cfg.RateControl.BitrateLimit != 4096 {
		t.Errorf("Expected BitrateLimit 4096, got %+v", cfg.RateControl)
	}

	if cfg.H264 == nil || cfg.H264.H264Profile != "Main" {
		t.Errorf("Expected H264Profile Main, got %+v", cfg.H264)
	}

	if cfg.Multicast == nil || cfg.Multicast.Port != 10000 {
		t.Errorf("Expected multicast port 10000, got %+v", cfg.Multicast)
	}

	if cfg.Multicast.Address == nil || cfg.Multicast.Address.IPv4Address != "239.0.0.1" {
		t.Errorf("Expected multicast address 239.0.0.1, got %+v", cfg.Multicast.Address)
	}
}

func TestGetVideoEncoderConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetVideoEncoderConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetVideoEncoderConfigurations(), got nil")
	}
}

// ---- GetAudioEncoderConfigurations ----

func TestGetAudioEncoderConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioEncoderConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioEnc1">
				<trt:Name>AAC Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:Encoding>AAC</trt:Encoding>
				<trt:Bitrate>128</trt:Bitrate>
				<trt:SampleRate>48</trt:SampleRate>
				<trt:Multicast>
					<trt:Address>
						<trt:Type>IPv4</trt:Type>
						<trt:IPv4Address>239.0.0.2</trt:IPv4Address>
					</trt:Address>
					<trt:Port>10002</trt:Port>
					<trt:TTL>8</trt:TTL>
					<trt:AutoStart>true</trt:AutoStart>
				</trt:Multicast>
			</trt:Configurations>
		</trt:GetAudioEncoderConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetAudioEncoderConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetAudioEncoderConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	cfg := configs[0]

	if cfg.Token != testAudioEncToken || cfg.Encoding != testEncodingAAC {
		t.Errorf("Expected token AudioEnc1/AAC, got %s/%s", cfg.Token, cfg.Encoding)
	}

	if cfg.Bitrate != 128 || cfg.SampleRate != 48 {
		t.Errorf("Expected bitrate 128 / sample rate 48, got %d/%d", cfg.Bitrate, cfg.SampleRate)
	}

	if cfg.Multicast == nil || !cfg.Multicast.AutoStart || cfg.Multicast.Port != 10002 {
		t.Errorf("Expected multicast autostart true / port 10002, got %+v", cfg.Multicast)
	}
}

func TestGetAudioEncoderConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetAudioEncoderConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetAudioEncoderConfigurations(), got nil")
	}
}

// ---- GetVideoSourceConfiguration ----

func TestGetVideoSourceConfiguration(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoSourceConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="VideoSrcCfg1">
				<trt:Name>Video Source Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:SourceToken>VideoSource1</trt:SourceToken>
				<trt:Bounds x="1" y="2" width="640" height="480"/>
			</trt:Configuration>
		</trt:GetVideoSourceConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	cfg, err := client.GetVideoSourceConfiguration(context.Background(), testVideoSrcCfgToken)
	if err != nil {
		t.Fatalf("GetVideoSourceConfiguration() failed: %v", err)
	}

	if cfg.Token != testVideoSrcCfgToken || cfg.SourceToken != testVideoSourceToken {
		t.Errorf("Expected token/source VideoSrcCfg1/VideoSource1, got %s/%s", cfg.Token, cfg.SourceToken)
	}

	if cfg.Bounds == nil || cfg.Bounds.Width != 640 || cfg.Bounds.Height != 480 {
		t.Errorf("Expected bounds 640x480, got %+v", cfg.Bounds)
	}
}

func TestGetVideoSourceConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetVideoSourceConfiguration(context.Background(), testVideoSrcCfgToken)
	if err == nil {
		t.Fatal("Expected error from GetVideoSourceConfiguration(), got nil")
	}
}

// ---- GetAudioSourceConfiguration ----

func TestGetAudioSourceConfiguration(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioSourceConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="AudioSrcCfg1">
				<trt:Name>Audio Source Config</trt:Name>
				<trt:UseCount>3</trt:UseCount>
				<trt:SourceToken>AudioSource1</trt:SourceToken>
			</trt:Configuration>
		</trt:GetAudioSourceConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	cfg, err := client.GetAudioSourceConfiguration(context.Background(), testAudioSrcCfgToken)
	if err != nil {
		t.Fatalf("GetAudioSourceConfiguration() failed: %v", err)
	}

	if cfg.Token != testAudioSrcCfgToken || cfg.SourceToken != testAudioSourceToken || cfg.UseCount != 3 {
		t.Errorf("Unexpected configuration: %+v", cfg)
	}
}

func TestGetAudioSourceConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetAudioSourceConfiguration(context.Background(), testAudioSrcCfgToken)
	if err == nil {
		t.Fatal("Expected error from GetAudioSourceConfiguration(), got nil")
	}
}

// ---- GetVideoSourceConfigurationOptions ----

func TestGetVideoSourceConfigurationOptions(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoSourceConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<trt:BoundsRange>
					<trt:X><trt:Min>0</trt:Min><trt:Max>1920</trt:Max></trt:X>
					<trt:Y><trt:Min>0</trt:Min><trt:Max>1080</trt:Max></trt:Y>
					<trt:Width><trt:Min>1</trt:Min><trt:Max>1920</trt:Max></trt:Width>
					<trt:Height><trt:Min>1</trt:Min><trt:Max>1080</trt:Max></trt:Height>
				</trt:BoundsRange>
				<trt:VideoSourceTokensAvailable>VideoSource1</trt:VideoSourceTokensAvailable>
				<trt:VideoSourceTokensAvailable>VideoSource2</trt:VideoSourceTokensAvailable>
			</trt:Options>
		</trt:GetVideoSourceConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	opts, err := client.GetVideoSourceConfigurationOptions(context.Background(), testVideoSrcCfgToken, testProfileToken)
	if err != nil {
		t.Fatalf("GetVideoSourceConfigurationOptions() failed: %v", err)
	}

	if opts.BoundsRange == nil || opts.BoundsRange.Width == nil || opts.BoundsRange.Width.Max != 1920 {
		t.Errorf("Expected bounds range width max 1920, got %+v", opts.BoundsRange)
	}

	if len(opts.VideoSourceTokensAvailable) != 2 || opts.VideoSourceTokensAvailable[0] != testVideoSourceToken {
		t.Errorf("Expected 2 video source tokens starting with VideoSource1, got %v", opts.VideoSourceTokensAvailable)
	}
}

func TestGetVideoSourceConfigurationOptionsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetVideoSourceConfigurationOptions(context.Background(), testVideoSrcCfgToken, testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetVideoSourceConfigurationOptions(), got nil")
	}
}

// ---- GetAudioSourceConfigurationOptions ----

func TestGetAudioSourceConfigurationOptions(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioSourceConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<trt:InputTokensAvailable>AudioSource1</trt:InputTokensAvailable>
			</trt:Options>
		</trt:GetAudioSourceConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	opts, err := client.GetAudioSourceConfigurationOptions(context.Background(), testAudioSrcCfgToken, testProfileToken)
	if err != nil {
		t.Fatalf("GetAudioSourceConfigurationOptions() failed: %v", err)
	}

	if len(opts.InputTokensAvailable) != 1 || opts.InputTokensAvailable[0] != testAudioSourceToken {
		t.Errorf("Expected [AudioSource1], got %v", opts.InputTokensAvailable)
	}
}

func TestGetAudioSourceConfigurationOptionsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetAudioSourceConfigurationOptions(context.Background(), testAudioSrcCfgToken, testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetAudioSourceConfigurationOptions(), got nil")
	}
}

// ---- SetVideoSourceConfiguration ----

func TestSetVideoSourceConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetVideoSourceConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &VideoSourceConfiguration{
		Token:       testVideoSrcCfgToken,
		Name:        "Video Source Config",
		SourceToken: testVideoSourceToken,
		Bounds: &IntRectangle{
			X: 0, Y: 0, Width: 1920, Height: 1080,
		},
	}

	if err := client.SetVideoSourceConfiguration(context.Background(), config, true); err != nil {
		t.Fatalf("SetVideoSourceConfiguration() failed: %v", err)
	}
}

func TestSetVideoSourceConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &VideoSourceConfiguration{Token: testVideoSrcCfgToken}

	if err := client.SetVideoSourceConfiguration(context.Background(), config, true); err == nil {
		t.Fatal("Expected error from SetVideoSourceConfiguration(), got nil")
	}
}

// ---- SetAudioSourceConfiguration ----

func TestSetAudioSourceConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetAudioSourceConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &AudioSourceConfiguration{
		Token:       testAudioSrcCfgToken,
		Name:        "Audio Source Config",
		SourceToken: testAudioSourceToken,
	}

	if err := client.SetAudioSourceConfiguration(context.Background(), config, false); err != nil {
		t.Fatalf("SetAudioSourceConfiguration() failed: %v", err)
	}
}

func TestSetAudioSourceConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &AudioSourceConfiguration{Token: testAudioSrcCfgToken}

	if err := client.SetAudioSourceConfiguration(context.Background(), config, false); err == nil {
		t.Fatal("Expected error from SetAudioSourceConfiguration(), got nil")
	}
}

// ---- GetCompatibleVideoEncoderConfigurations ----

func TestGetCompatibleVideoEncoderConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleVideoEncoderConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="VideoEnc1">
				<trt:Name>H264 Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:Encoding>H264</trt:Encoding>
				<trt:Resolution>
					<trt:Width>1280</trt:Width>
					<trt:Height>720</trt:Height>
				</trt:Resolution>
				<trt:Quality>4.0</trt:Quality>
				<trt:RateControl>
					<trt:FrameRateLimit>25</trt:FrameRateLimit>
					<trt:EncodingInterval>1</trt:EncodingInterval>
					<trt:BitrateLimit>2048</trt:BitrateLimit>
				</trt:RateControl>
			</trt:Configurations>
		</trt:GetCompatibleVideoEncoderConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleVideoEncoderConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleVideoEncoderConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testVideoEncToken || configs[0].Resolution == nil || configs[0].Resolution.Width != 1280 {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}

	if configs[0].RateControl == nil || configs[0].RateControl.BitrateLimit != 2048 {
		t.Errorf("Expected BitrateLimit 2048, got %+v", configs[0].RateControl)
	}
}

func TestGetCompatibleVideoEncoderConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleVideoEncoderConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleVideoEncoderConfigurations(), got nil")
	}
}

// ---- GetCompatibleVideoSourceConfigurations ----

func TestGetCompatibleVideoSourceConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleVideoSourceConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="VideoSrcCfg1">
				<trt:Name>Video Source Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:SourceToken>VideoSource1</trt:SourceToken>
				<trt:Bounds x="0" y="0" width="800" height="600"/>
			</trt:Configurations>
		</trt:GetCompatibleVideoSourceConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleVideoSourceConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleVideoSourceConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testVideoSrcCfgToken || configs[0].Bounds == nil || configs[0].Bounds.Width != 800 {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatibleVideoSourceConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleVideoSourceConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleVideoSourceConfigurations(), got nil")
	}
}

// ---- GetCompatibleAudioEncoderConfigurations ----

func TestGetCompatibleAudioEncoderConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleAudioEncoderConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioEnc1">
				<trt:Name>AAC Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:Encoding>AAC</trt:Encoding>
				<trt:Bitrate>96</trt:Bitrate>
				<trt:SampleRate>44</trt:SampleRate>
			</trt:Configurations>
		</trt:GetCompatibleAudioEncoderConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleAudioEncoderConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleAudioEncoderConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testAudioEncToken || configs[0].Bitrate != 96 || configs[0].SampleRate != 44 {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatibleAudioEncoderConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleAudioEncoderConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleAudioEncoderConfigurations(), got nil")
	}
}

// ---- GetCompatibleAudioSourceConfigurations ----

func TestGetCompatibleAudioSourceConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleAudioSourceConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioSrcCfg1">
				<trt:Name>Audio Source Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:SourceToken>AudioSource1</trt:SourceToken>
			</trt:Configurations>
		</trt:GetCompatibleAudioSourceConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleAudioSourceConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleAudioSourceConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testAudioSrcCfgToken || configs[0].SourceToken != testAudioSourceToken {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatibleAudioSourceConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleAudioSourceConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleAudioSourceConfigurations(), got nil")
	}
}

// ---- GetCompatiblePTZConfigurations ----

func TestGetCompatiblePTZConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatiblePTZConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="PTZCfg1">
				<trt:Name>PTZ Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:NodeToken>PTZNode1</trt:NodeToken>
			</trt:Configurations>
		</trt:GetCompatiblePTZConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatiblePTZConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatiblePTZConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != "PTZCfg1" || configs[0].NodeToken != "PTZNode1" {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatiblePTZConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatiblePTZConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatiblePTZConfigurations(), got nil")
	}
}

// ---- GetCompatibleMetadataConfigurations ----

func TestGetCompatibleMetadataConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleMetadataConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="MetaCfg1">
				<trt:Name>Metadata Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:Analytics>true</trt:Analytics>
			</trt:Configurations>
		</trt:GetCompatibleMetadataConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleMetadataConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleMetadataConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != "MetaCfg1" || !configs[0].Analytics {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatibleMetadataConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleMetadataConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleMetadataConfigurations(), got nil")
	}
}

// ---- GetCompatibleAudioOutputConfigurations ----

func TestGetCompatibleAudioOutputConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleAudioOutputConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioOutCfg1">
				<trt:Name>Audio Output Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:OutputToken>AudioOutput1</trt:OutputToken>
			</trt:Configurations>
		</trt:GetCompatibleAudioOutputConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleAudioOutputConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleAudioOutputConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != "AudioOutCfg1" || configs[0].OutputToken != testAudioOutputToken {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatibleAudioOutputConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleAudioOutputConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleAudioOutputConfigurations(), got nil")
	}
}

// ---- GetCompatibleAudioDecoderConfigurations ----

func TestGetCompatibleAudioDecoderConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleAudioDecoderConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioDecCfg1">
				<trt:Name>Audio Decoder Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
			</trt:Configurations>
		</trt:GetCompatibleAudioDecoderConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleAudioDecoderConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleAudioDecoderConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testAudioDecCfgToken {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatibleAudioDecoderConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleAudioDecoderConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleAudioDecoderConfigurations(), got nil")
	}
}

// ---- GetMetadataConfigurations ----

func TestGetMetadataConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetMetadataConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="MetaCfg1">
				<trt:Name>Metadata Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:Analytics>true</trt:Analytics>
			</trt:Configurations>
		</trt:GetMetadataConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetMetadataConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetMetadataConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != "MetaCfg1" || !configs[0].Analytics {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetMetadataConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetMetadataConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetMetadataConfigurations(), got nil")
	}
}

// ---- GetAudioOutputConfigurations ----

func TestGetAudioOutputConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioOutputConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioOutCfg1">
				<trt:Name>Audio Output Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
				<trt:OutputToken>AudioOutput1</trt:OutputToken>
			</trt:Configurations>
		</trt:GetAudioOutputConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetAudioOutputConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetAudioOutputConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != "AudioOutCfg1" || configs[0].OutputToken != testAudioOutputToken {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetAudioOutputConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetAudioOutputConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetAudioOutputConfigurations(), got nil")
	}
}

// ---- GetAudioDecoderConfigurations ----

func TestGetAudioDecoderConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioDecoderConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AudioDecCfg1">
				<trt:Name>Audio Decoder Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
			</trt:Configurations>
		</trt:GetAudioDecoderConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetAudioDecoderConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetAudioDecoderConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testAudioDecCfgToken {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetAudioDecoderConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetAudioDecoderConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetAudioDecoderConfigurations(), got nil")
	}
}

// ---- GetAudioDecoderConfiguration ----

func TestGetAudioDecoderConfiguration(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioDecoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="AudioDecCfg1">
				<trt:Name>Audio Decoder Config</trt:Name>
				<trt:UseCount>2</trt:UseCount>
			</trt:Configuration>
		</trt:GetAudioDecoderConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	cfg, err := client.GetAudioDecoderConfiguration(context.Background(), testAudioDecCfgToken)
	if err != nil {
		t.Fatalf("GetAudioDecoderConfiguration() failed: %v", err)
	}

	if cfg.Token != testAudioDecCfgToken || cfg.UseCount != 2 {
		t.Errorf("Unexpected configuration: %+v", cfg)
	}
}

func TestGetAudioDecoderConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetAudioDecoderConfiguration(context.Background(), testAudioDecCfgToken)
	if err == nil {
		t.Fatal("Expected error from GetAudioDecoderConfiguration(), got nil")
	}
}

// ---- SetAudioDecoderConfiguration ----

func TestSetAudioDecoderConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetAudioDecoderConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &AudioDecoderConfiguration{
		Token: testAudioDecCfgToken,
		Name:  "Audio Decoder Config",
	}

	if err := client.SetAudioDecoderConfiguration(context.Background(), config, true); err != nil {
		t.Fatalf("SetAudioDecoderConfiguration() failed: %v", err)
	}
}

func TestSetAudioDecoderConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &AudioDecoderConfiguration{Token: testAudioDecCfgToken}

	if err := client.SetAudioDecoderConfiguration(context.Background(), config, true); err == nil {
		t.Fatal("Expected error from SetAudioDecoderConfiguration(), got nil")
	}
}

// ---- GetVideoAnalyticsConfigurations ----

func TestGetVideoAnalyticsConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoAnalyticsConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AnalyticsCfg1">
				<trt:Name>Analytics Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
			</trt:Configurations>
		</trt:GetVideoAnalyticsConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetVideoAnalyticsConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetVideoAnalyticsConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testAnalyticsCfgToken {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetVideoAnalyticsConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetVideoAnalyticsConfigurations(context.Background())
	if err == nil {
		t.Fatal("Expected error from GetVideoAnalyticsConfigurations(), got nil")
	}
}

// ---- GetVideoAnalyticsConfiguration ----

func TestGetVideoAnalyticsConfiguration(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoAnalyticsConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="AnalyticsCfg1">
				<trt:Name>Analytics Config</trt:Name>
				<trt:UseCount>5</trt:UseCount>
			</trt:Configuration>
		</trt:GetVideoAnalyticsConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	cfg, err := client.GetVideoAnalyticsConfiguration(context.Background(), testAnalyticsCfgToken)
	if err != nil {
		t.Fatalf("GetVideoAnalyticsConfiguration() failed: %v", err)
	}

	if cfg.Token != testAnalyticsCfgToken || cfg.UseCount != 5 {
		t.Errorf("Unexpected configuration: %+v", cfg)
	}
}

func TestGetVideoAnalyticsConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetVideoAnalyticsConfiguration(context.Background(), testAnalyticsCfgToken)
	if err == nil {
		t.Fatal("Expected error from GetVideoAnalyticsConfiguration(), got nil")
	}
}

// ---- GetCompatibleVideoAnalyticsConfigurations ----

func TestGetCompatibleVideoAnalyticsConfigurations(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetCompatibleVideoAnalyticsConfigurationsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configurations token="AnalyticsCfg1">
				<trt:Name>Analytics Config</trt:Name>
				<trt:UseCount>1</trt:UseCount>
			</trt:Configurations>
		</trt:GetCompatibleVideoAnalyticsConfigurationsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	configs, err := client.GetCompatibleVideoAnalyticsConfigurations(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetCompatibleVideoAnalyticsConfigurations() failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(configs))
	}

	if configs[0].Token != testAnalyticsCfgToken {
		t.Errorf("Unexpected configuration: %+v", configs[0])
	}
}

func TestGetCompatibleVideoAnalyticsConfigurationsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetCompatibleVideoAnalyticsConfigurations(context.Background(), testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetCompatibleVideoAnalyticsConfigurations(), got nil")
	}
}

// ---- SetVideoAnalyticsConfiguration ----

func TestSetVideoAnalyticsConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetVideoAnalyticsConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &VideoAnalyticsConfiguration{
		Token: testAnalyticsCfgToken,
		Name:  "Analytics Config",
	}

	if err := client.SetVideoAnalyticsConfiguration(context.Background(), config, true); err != nil {
		t.Fatalf("SetVideoAnalyticsConfiguration() failed: %v", err)
	}
}

func TestSetVideoAnalyticsConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	config := &VideoAnalyticsConfiguration{Token: testAnalyticsCfgToken}

	if err := client.SetVideoAnalyticsConfiguration(context.Background(), config, true); err == nil {
		t.Fatal("Expected error from SetVideoAnalyticsConfiguration(), got nil")
	}
}

// ---- GetVideoAnalyticsConfigurationOptions ----

func TestGetVideoAnalyticsConfigurationOptions(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoAnalyticsConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options/>
		</trt:GetVideoAnalyticsConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
	server := newMediaConfigServer(response)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	opts, err := client.GetVideoAnalyticsConfigurationOptions(context.Background(), testAnalyticsCfgToken, testProfileToken)
	if err != nil {
		t.Fatalf("GetVideoAnalyticsConfigurationOptions() failed: %v", err)
	}

	if opts == nil {
		t.Fatal("Expected non-nil options")
	}
}

func TestGetVideoAnalyticsConfigurationOptionsFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	_, err := client.GetVideoAnalyticsConfigurationOptions(context.Background(), testAnalyticsCfgToken, testProfileToken)
	if err == nil {
		t.Fatal("Expected error from GetVideoAnalyticsConfigurationOptions(), got nil")
	}
}

// ---- AddVideoAnalyticsConfiguration ----

func TestAddVideoAnalyticsConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddVideoAnalyticsConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.AddVideoAnalyticsConfiguration(context.Background(), testProfileToken, testAnalyticsCfgToken); err != nil {
		t.Fatalf("AddVideoAnalyticsConfiguration() failed: %v", err)
	}
}

func TestAddVideoAnalyticsConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.AddVideoAnalyticsConfiguration(context.Background(), testProfileToken, testAnalyticsCfgToken); err == nil {
		t.Fatal("Expected error from AddVideoAnalyticsConfiguration(), got nil")
	}
}

// ---- RemoveVideoAnalyticsConfiguration ----

func TestRemoveVideoAnalyticsConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveVideoAnalyticsConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.RemoveVideoAnalyticsConfiguration(context.Background(), testProfileToken); err != nil {
		t.Fatalf("RemoveVideoAnalyticsConfiguration() failed: %v", err)
	}
}

func TestRemoveVideoAnalyticsConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.RemoveVideoAnalyticsConfiguration(context.Background(), testProfileToken); err == nil {
		t.Fatal("Expected error from RemoveVideoAnalyticsConfiguration(), got nil")
	}
}

// ---- AddAudioOutputConfiguration ----

func TestAddAudioOutputConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddAudioOutputConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.AddAudioOutputConfiguration(context.Background(), testProfileToken, "AudioOutCfg1"); err != nil {
		t.Fatalf("AddAudioOutputConfiguration() failed: %v", err)
	}
}

func TestAddAudioOutputConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.AddAudioOutputConfiguration(context.Background(), testProfileToken, "AudioOutCfg1"); err == nil {
		t.Fatal("Expected error from AddAudioOutputConfiguration(), got nil")
	}
}

// ---- RemoveAudioOutputConfiguration ----

func TestRemoveAudioOutputConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveAudioOutputConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.RemoveAudioOutputConfiguration(context.Background(), testProfileToken); err != nil {
		t.Fatalf("RemoveAudioOutputConfiguration() failed: %v", err)
	}
}

func TestRemoveAudioOutputConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.RemoveAudioOutputConfiguration(context.Background(), testProfileToken); err == nil {
		t.Fatal("Expected error from RemoveAudioOutputConfiguration(), got nil")
	}
}

// ---- AddAudioDecoderConfiguration ----

func TestAddAudioDecoderConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddAudioDecoderConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.AddAudioDecoderConfiguration(context.Background(), testProfileToken, testAudioDecCfgToken); err != nil {
		t.Fatalf("AddAudioDecoderConfiguration() failed: %v", err)
	}
}

func TestAddAudioDecoderConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.AddAudioDecoderConfiguration(context.Background(), testProfileToken, testAudioDecCfgToken); err == nil {
		t.Fatal("Expected error from AddAudioDecoderConfiguration(), got nil")
	}
}

// ---- RemoveAudioDecoderConfiguration ----

func TestRemoveAudioDecoderConfiguration(t *testing.T) {
	server := newMediaConfigServer(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveAudioDecoderConfigurationResponse/></soap:Body></soap:Envelope>`)
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.RemoveAudioDecoderConfiguration(context.Background(), testProfileToken); err != nil {
		t.Fatalf("RemoveAudioDecoderConfiguration() failed: %v", err)
	}
}

func TestRemoveAudioDecoderConfigurationFault(t *testing.T) {
	server := newMediaConfigFaultServer()
	defer server.Close()

	client := newMediaConfigClient(t, server.URL)

	if err := client.RemoveAudioDecoderConfiguration(context.Background(), testProfileToken); err == nil {
		t.Fatal("Expected error from RemoveAudioDecoderConfiguration(), got nil")
	}
}
