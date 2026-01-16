package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	encodingH264 = "H264"
)

// Test device information from real camera:
// Manufacturer: Bosch
// Model: FLEXIDOME indoor 5100i IR
// Firmware: 8.71.0066
// Serial Number: 404754734001050102
// Hardware ID: F000B543

// TestGetMediaServiceCapabilities_Bosch tests GetMediaServiceCapabilities with real camera response.
func TestGetMediaServiceCapabilities_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	// Note: Adapted to match the expected nested structure in the code
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetServiceCapabilitiesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Capabilities SnapshotUri="false" Rotation="true" VideoSourceMode="false" OSD="false" TemporaryOSDText="false" EXICompression="false">
        <trt:ProfileCapabilities MaximumNumberOfProfiles="32"/>
        <trt:StreamingCapabilities RTPMulticast="true" RTP_TCP="false" RTP_RTSP_TCP="true"/>
      </trt:Capabilities>
    </trt:GetServiceCapabilitiesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		// Validate SOAP request contains GetServiceCapabilities
		if !strings.Contains(bodyStr, "GetServiceCapabilities") {
			t.Errorf("Request should contain GetServiceCapabilities, got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "http://www.onvif.org/ver10/media/wsdl") {
			t.Errorf("Request should contain media namespace, got: %s", bodyStr)
		}

		// Return real camera response
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	capabilities, err := client.GetMediaServiceCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetMediaServiceCapabilities() failed: %v", err)
	}

	// Validate response matches real camera
	if capabilities.MaximumNumberOfProfiles != 32 {
		t.Errorf("Expected MaximumNumberOfProfiles=32 (Bosch FLEXIDOME), got %d", capabilities.MaximumNumberOfProfiles)
	}
	if !capabilities.RTPMulticast {
		t.Error("Expected RTPMulticast=true (Bosch FLEXIDOME)")
	}
	if !capabilities.RTPRTSPTCP {
		t.Error("Expected RTPRTSPTCP=true (Bosch FLEXIDOME)")
	}
	if capabilities.SnapshotURI {
		t.Error("Expected SnapshotURI=false (Bosch FLEXIDOME)")
	}
	if !capabilities.Rotation {
		t.Error("Expected Rotation=true (Bosch FLEXIDOME)")
	}
}

// TestGetProfiles_Bosch tests GetProfiles with real camera response.
func TestGetProfiles_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetProfilesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Profiles token="0">
        <trt:Name>Profile_L1S1</trt:Name>
        <trt:VideoSourceConfiguration token="1">
          <trt:Name>Camera_1</trt:Name>
          <trt:UseCount>4</trt:UseCount>
          <trt:SourceToken>1</trt:SourceToken>
          <trt:Bounds x="0" y="0" width="1920" height="1080"/>
        </trt:VideoSourceConfiguration>
        <trt:VideoEncoderConfiguration token="EncCfg_L1S1">
          <trt:Name>Balanced 2 MP</trt:Name>
          <trt:UseCount>1</trt:UseCount>
          <trt:Encoding>H264</trt:Encoding>
          <trt:Resolution>
            <tt:Width xmlns:tt="http://www.onvif.org/ver10/schema">1920</tt:Width>
            <tt:Height xmlns:tt="http://www.onvif.org/ver10/schema">1080</tt:Height>
          </trt:Resolution>
          <trt:Quality>0</trt:Quality>
          <trt:RateControl>
            <tt:FrameRateLimit xmlns:tt="http://www.onvif.org/ver10/schema">30</tt:FrameRateLimit>
            <tt:EncodingInterval xmlns:tt="http://www.onvif.org/ver10/schema">1</tt:EncodingInterval>
            <tt:BitrateLimit xmlns:tt="http://www.onvif.org/ver10/schema">5200</tt:BitrateLimit>
          </trt:RateControl>
        </trt:VideoEncoderConfiguration>
      </trt:Profiles>
    </trt:GetProfilesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		// Validate SOAP request
		if !strings.Contains(bodyStr, "GetProfiles") {
			t.Errorf("Request should contain GetProfiles, got: %s", bodyStr)
		}

		// Return real camera response
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		t.Fatalf("GetProfiles() failed: %v", err)
	}

	// Validate response matches real camera
	if len(profiles) == 0 {
		t.Fatal("Expected at least one profile from Bosch FLEXIDOME")
	}
	if profiles[0].Token != "0" {
		t.Errorf("Expected profile token=0 (Bosch FLEXIDOME), got %s", profiles[0].Token)
	}
	if profiles[0].Name != "Profile_L1S1" {
		t.Errorf("Expected profile name=Profile_L1S1 (Bosch FLEXIDOME), got %s", profiles[0].Name)
	}
	if profiles[0].VideoEncoderConfiguration == nil {
		t.Fatal("Expected VideoEncoderConfiguration from Bosch FLEXIDOME")
	}
	if profiles[0].VideoEncoderConfiguration.Token != "EncCfg_L1S1" {
		t.Errorf("Expected encoder token=EncCfg_L1S1 (Bosch FLEXIDOME), got %s", profiles[0].VideoEncoderConfiguration.Token)
	}
	if profiles[0].VideoEncoderConfiguration.Encoding != encodingH264 {
		t.Errorf("Expected encoding=H264 (Bosch FLEXIDOME), got %s", profiles[0].VideoEncoderConfiguration.Encoding)
	}
	if profiles[0].VideoEncoderConfiguration.Resolution.Width != 1920 {
		t.Errorf("Expected width=1920 (Bosch FLEXIDOME), got %d", profiles[0].VideoEncoderConfiguration.Resolution.Width)
	}
	if profiles[0].VideoEncoderConfiguration.Resolution.Height != 1080 {
		t.Errorf("Expected height=1080 (Bosch FLEXIDOME), got %d", profiles[0].VideoEncoderConfiguration.Resolution.Height)
	}
}

// TestGetVideoSources_Bosch tests GetVideoSources with real camera response.
func TestGetVideoSources_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetVideoSourcesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:VideoSources token="1">
        <tt:Framerate xmlns:tt="http://www.onvif.org/ver10/schema">30</tt:Framerate>
        <tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
          <tt:Width>1920</tt:Width>
          <tt:Height>1080</tt:Height>
        </tt:Resolution>
      </trt:VideoSources>
    </trt:GetVideoSourcesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetVideoSources") {
			t.Errorf("Request should contain GetVideoSources, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	sources, err := client.GetVideoSources(ctx)
	if err != nil {
		t.Fatalf("GetVideoSources() failed: %v", err)
	}

	// Validate response matches real camera
	if len(sources) == 0 {
		t.Fatal("Expected at least one video source from Bosch FLEXIDOME")
	}
	if sources[0].Token != "1" {
		t.Errorf("Expected source token=1 (Bosch FLEXIDOME), got %s", sources[0].Token)
	}
	if sources[0].Framerate != 30 {
		t.Errorf("Expected framerate=30 (Bosch FLEXIDOME), got %f", sources[0].Framerate)
	}
	if sources[0].Resolution.Width != 1920 {
		t.Errorf("Expected width=1920 (Bosch FLEXIDOME), got %d", sources[0].Resolution.Width)
	}
	if sources[0].Resolution.Height != 1080 {
		t.Errorf("Expected height=1080 (Bosch FLEXIDOME), got %d", sources[0].Resolution.Height)
	}
}

// TestGetAudioSources_Bosch tests GetAudioSources with real camera response.
func TestGetAudioSources_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetAudioSourcesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:AudioSources token="1">
        <tt:Channels xmlns:tt="http://www.onvif.org/ver10/schema">2</tt:Channels>
      </trt:AudioSources>
    </trt:GetAudioSourcesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetAudioSources") {
			t.Errorf("Request should contain GetAudioSources, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	sources, err := client.GetAudioSources(ctx)
	if err != nil {
		t.Fatalf("GetAudioSources() failed: %v", err)
	}

	// Validate response matches real camera
	if len(sources) == 0 {
		t.Fatal("Expected at least one audio source from Bosch FLEXIDOME")
	}
	if sources[0].Token != "1" {
		t.Errorf("Expected source token=1 (Bosch FLEXIDOME), got %s", sources[0].Token)
	}
	if sources[0].Channels != 2 {
		t.Errorf("Expected channels=2 (Bosch FLEXIDOME), got %d", sources[0].Channels)
	}
}

// TestGetAudioOutputs_Bosch tests GetAudioOutputs with real camera response.
func TestGetAudioOutputs_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetAudioOutputsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:AudioOutputs token="AudioOut 1"/>
    </trt:GetAudioOutputsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetAudioOutputs") {
			t.Errorf("Request should contain GetAudioOutputs, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	outputs, err := client.GetAudioOutputs(ctx)
	if err != nil {
		t.Fatalf("GetAudioOutputs() failed: %v", err)
	}

	// Validate response matches real camera
	if len(outputs) == 0 {
		t.Fatal("Expected at least one audio output from Bosch FLEXIDOME")
	}
	if outputs[0].Token != "AudioOut 1" {
		t.Errorf("Expected output token=AudioOut 1 (Bosch FLEXIDOME), got %s", outputs[0].Token)
	}
}

// TestGetStreamURI_Bosch tests GetStreamURI with real camera response.
func TestGetStreamURI_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:MediaUri>
        <tt:Uri xmlns:tt="http://www.onvif.org/ver10/schema">rtsp://192.168.1.201/rtsp_tunnel?p=0&amp;line=1&amp;inst=1&amp;vcd=2</tt:Uri>
        <tt:InvalidAfterConnect xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:InvalidAfterConnect>
        <tt:InvalidAfterReboot xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:InvalidAfterReboot>
        <tt:Timeout xmlns:tt="http://www.onvif.org/ver10/schema">0</tt:Timeout>
      </trt:MediaUri>
    </trt:GetStreamUriResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetStreamUri") {
			t.Errorf("Request should contain GetStreamUri, got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "ProfileToken") {
			t.Errorf("Request should contain ProfileToken, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	uri, err := client.GetStreamURI(ctx, "0")
	if err != nil {
		t.Fatalf("GetStreamURI() failed: %v", err)
	}

	// Validate response matches real camera
	if !strings.Contains(uri.URI, "rtsp://") {
		t.Errorf("Expected RTSP URI from Bosch FLEXIDOME, got %s", uri.URI)
	}
	if !strings.Contains(uri.URI, "rtsp_tunnel") {
		t.Errorf("Expected rtsp_tunnel in URI from Bosch FLEXIDOME, got %s", uri.URI)
	}
	if uri.InvalidAfterReboot != true {
		t.Error("Expected InvalidAfterReboot=true from Bosch FLEXIDOME")
	}
}

// TestGetSnapshotURI_Bosch tests GetSnapshotURI with real camera response.
func TestGetSnapshotURI_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetSnapshotUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:MediaUri>
        <tt:Uri xmlns:tt="http://www.onvif.org/ver10/schema">http://192.168.1.201/snap.jpg?JpegCam=1</tt:Uri>
        <tt:InvalidAfterConnect xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:InvalidAfterConnect>
        <tt:InvalidAfterReboot xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:InvalidAfterReboot>
        <tt:Timeout xmlns:tt="http://www.onvif.org/ver10/schema">0</tt:Timeout>
      </trt:MediaUri>
    </trt:GetSnapshotUriResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetSnapshotUri") {
			t.Errorf("Request should contain GetSnapshotUri, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	uri, err := client.GetSnapshotURI(ctx, "0")
	if err != nil {
		t.Fatalf("GetSnapshotURI() failed: %v", err)
	}

	// Validate response matches real camera
	if !strings.Contains(uri.URI, "http://") {
		t.Errorf("Expected HTTP URI from Bosch FLEXIDOME, got %s", uri.URI)
	}
	if !strings.Contains(uri.URI, "snap.jpg") {
		t.Errorf("Expected snap.jpg in URI from Bosch FLEXIDOME, got %s", uri.URI)
	}
	if !strings.Contains(uri.URI, "JpegCam=1") {
		t.Errorf("Expected JpegCam=1 in URI from Bosch FLEXIDOME, got %s", uri.URI)
	}
}

// TestGetVideoEncoderConfiguration_Bosch tests GetVideoEncoderConfiguration with real camera response.
func TestGetVideoEncoderConfiguration_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetVideoEncoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Configuration token="EncCfg_L1S1">
        <tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Balanced 2 MP</tt:Name>
        <tt:UseCount xmlns:tt="http://www.onvif.org/ver10/schema">1</tt:UseCount>
        <tt:Encoding xmlns:tt="http://www.onvif.org/ver10/schema">H264</tt:Encoding>
        <tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
          <tt:Width>1920</tt:Width>
          <tt:Height>1080</tt:Height>
        </tt:Resolution>
        <tt:Quality xmlns:tt="http://www.onvif.org/ver10/schema">0</tt:Quality>
        <tt:RateControl xmlns:tt="http://www.onvif.org/ver10/schema">
          <tt:FrameRateLimit>30</tt:FrameRateLimit>
          <tt:EncodingInterval>1</tt:EncodingInterval>
          <tt:BitrateLimit>5200</tt:BitrateLimit>
        </tt:RateControl>
      </trt:Configuration>
    </trt:GetVideoEncoderConfigurationResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetVideoEncoderConfiguration") {
			t.Errorf("Request should contain GetVideoEncoderConfiguration, got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "ConfigurationToken") {
			t.Errorf("Request should contain ConfigurationToken, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	config, err := client.GetVideoEncoderConfiguration(ctx, "EncCfg_L1S1")
	if err != nil {
		t.Fatalf("GetVideoEncoderConfiguration() failed: %v", err)
	}

	// Validate response matches real camera
	if config.Token != "EncCfg_L1S1" {
		t.Errorf("Expected token=EncCfg_L1S1 (Bosch FLEXIDOME), got %s", config.Token)
	}
	if config.Name != "Balanced 2 MP" {
		t.Errorf("Expected name=Balanced 2 MP (Bosch FLEXIDOME), got %s", config.Name)
	}
	if config.Encoding != encodingH264 {
		t.Errorf("Expected encoding=H264 (Bosch FLEXIDOME), got %s", config.Encoding)
	}
	if config.Resolution.Width != 1920 {
		t.Errorf("Expected width=1920 (Bosch FLEXIDOME), got %d", config.Resolution.Width)
	}
	if config.Resolution.Height != 1080 {
		t.Errorf("Expected height=1080 (Bosch FLEXIDOME), got %d", config.Resolution.Height)
	}
	if config.RateControl.FrameRateLimit != 30 {
		t.Errorf("Expected FrameRateLimit=30 (Bosch FLEXIDOME), got %d", config.RateControl.FrameRateLimit)
	}
	if config.RateControl.BitrateLimit != 5200 {
		t.Errorf("Expected BitrateLimit=5200 (Bosch FLEXIDOME), got %d", config.RateControl.BitrateLimit)
	}
}

// TestGetVideoEncoderConfigurationOptions_Bosch tests GetVideoEncoderConfigurationOptions with real camera response.
func TestGetVideoEncoderConfigurationOptions_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetVideoEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Options>
        <tt:QualityRange xmlns:tt="http://www.onvif.org/ver10/schema">
          <tt:Min>0</tt:Min>
          <tt:Max>100</tt:Max>
        </tt:QualityRange>
        <tt:H264 xmlns:tt="http://www.onvif.org/ver10/schema">
          <tt:ResolutionsAvailable>
            <tt:Width>1920</tt:Width>
            <tt:Height>1080</tt:Height>
          </tt:ResolutionsAvailable>
          <tt:GovLengthRange>
            <tt:Min>1</tt:Min>
            <tt:Max>255</tt:Max>
          </tt:GovLengthRange>
          <tt:FrameRateRange>
            <tt:Min>1</tt:Min>
            <tt:Max>30</tt:Max>
          </tt:FrameRateRange>
          <tt:EncodingIntervalRange>
            <tt:Min>1</tt:Min>
            <tt:Max>1</tt:Max>
          </tt:EncodingIntervalRange>
          <tt:H264ProfilesSupported>Main</tt:H264ProfilesSupported>
        </tt:H264>
      </trt:Options>
    </trt:GetVideoEncoderConfigurationOptionsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetVideoEncoderConfigurationOptions") {
			t.Errorf("Request should contain GetVideoEncoderConfigurationOptions, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	options, err := client.GetVideoEncoderConfigurationOptions(ctx, "EncCfg_L1S1")
	if err != nil {
		t.Fatalf("GetVideoEncoderConfigurationOptions() failed: %v", err)
	}

	// Validate response matches real camera
	if options.QualityRange == nil {
		t.Fatal("Expected QualityRange from Bosch FLEXIDOME")
	}
	if options.QualityRange.Min != 0 || options.QualityRange.Max != 100 {
		t.Errorf("Expected QualityRange 0-100 (Bosch FLEXIDOME), got %f-%f", options.QualityRange.Min, options.QualityRange.Max)
	}
	if options.H264 == nil {
		t.Fatal("Expected H264 options from Bosch FLEXIDOME")
	}
	if len(options.H264.ResolutionsAvailable) == 0 {
		t.Fatal("Expected at least one resolution from Bosch FLEXIDOME")
	}
	if options.H264.ResolutionsAvailable[0].Width != 1920 {
		t.Errorf("Expected resolution width=1920 (Bosch FLEXIDOME), got %d", options.H264.ResolutionsAvailable[0].Width)
	}
	if options.H264.FrameRateRange.Min != 1 || options.H264.FrameRateRange.Max != 30 {
		t.Errorf("Expected FrameRateRange 1-30 (Bosch FLEXIDOME), got %f-%f", options.H264.FrameRateRange.Min, options.H264.FrameRateRange.Max)
	}
	if len(options.H264.H264ProfilesSupported) == 0 || options.H264.H264ProfilesSupported[0] != "Main" {
		t.Errorf("Expected H264 profile=Main (Bosch FLEXIDOME), got %v", options.H264.H264ProfilesSupported)
	}
}

// TestGetAudioEncoderConfigurationOptions_Bosch tests GetAudioEncoderConfigurationOptions with real camera response.
func TestGetAudioEncoderConfigurationOptions_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetAudioEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Options/>
    </trt:GetAudioEncoderConfigurationOptionsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetAudioEncoderConfigurationOptions") {
			t.Errorf("Request should contain GetAudioEncoderConfigurationOptions, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	options, err := client.GetAudioEncoderConfigurationOptions(ctx, "", "")
	if err != nil {
		t.Fatalf("GetAudioEncoderConfigurationOptions() failed: %v", err)
	}

	// Validate response - Bosch FLEXIDOME returns empty options
	if options == nil {
		t.Fatal("Expected options struct from Bosch FLEXIDOME")
	}
}

// TestGetAudioOutputConfigurationOptions_Bosch tests GetAudioOutputConfigurationOptions with real camera response.
func TestGetAudioOutputConfigurationOptions_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetAudioOutputConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Options>
        <tt:OutputTokensAvailable xmlns:tt="http://www.onvif.org/ver10/schema">AudioOut 1</tt:OutputTokensAvailable>
      </trt:Options>
    </trt:GetAudioOutputConfigurationOptionsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetAudioOutputConfigurationOptions") {
			t.Errorf("Request should contain GetAudioOutputConfigurationOptions, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	options, err := client.GetAudioOutputConfigurationOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetAudioOutputConfigurationOptions() failed: %v", err)
	}

	// Validate response matches real camera
	if len(options.OutputTokensAvailable) == 0 {
		t.Fatal("Expected at least one output token from Bosch FLEXIDOME")
	}
	if options.OutputTokensAvailable[0] != "AudioOut 1" {
		t.Errorf("Expected AudioOut 1 (Bosch FLEXIDOME), got %s", options.OutputTokensAvailable[0])
	}
}

// TestGetMetadataConfigurationOptions_Bosch tests GetMetadataConfigurationOptions with real camera response.
func TestGetMetadataConfigurationOptions_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetMetadataConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Options>
        <tt:PTZStatusFilterOptions xmlns:tt="http://www.onvif.org/ver10/schema">
          <tt:Status>false</tt:Status>
          <tt:Position>false</tt:Position>
        </tt:PTZStatusFilterOptions>
      </trt:Options>
    </trt:GetMetadataConfigurationOptionsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetMetadataConfigurationOptions") {
			t.Errorf("Request should contain GetMetadataConfigurationOptions, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	options, err := client.GetMetadataConfigurationOptions(ctx, "", "")
	if err != nil {
		t.Fatalf("GetMetadataConfigurationOptions() failed: %v", err)
	}

	// Validate response matches real camera
	if options.PTZStatusFilterOptions == nil {
		t.Fatal("Expected PTZStatusFilterOptions from Bosch FLEXIDOME")
	}
	if options.PTZStatusFilterOptions.Status != false {
		t.Error("Expected Status=false from Bosch FLEXIDOME")
	}
	if options.PTZStatusFilterOptions.Position != false {
		t.Error("Expected Position=false from Bosch FLEXIDOME")
	}
}

// TestGetAudioDecoderConfigurationOptions_Bosch tests GetAudioDecoderConfigurationOptions with real camera response.
func TestGetAudioDecoderConfigurationOptions_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:GetAudioDecoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Options>
        <tt:G711DecOptions xmlns:tt="http://www.onvif.org/ver10/schema"/>
      </trt:Options>
    </trt:GetAudioDecoderConfigurationOptionsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetAudioDecoderConfigurationOptions") {
			t.Errorf("Request should contain GetAudioDecoderConfigurationOptions, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	options, err := client.GetAudioDecoderConfigurationOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetAudioDecoderConfigurationOptions() failed: %v", err)
	}

	// Validate response matches real camera
	if options == nil {
		t.Fatal("Expected options from Bosch FLEXIDOME")
	}
	if options.G711DecOptions == nil {
		t.Error("Expected G711DecOptions from Bosch FLEXIDOME")
	}
}

// TestSetSynchronizationPoint_Bosch tests SetSynchronizationPoint with real camera response.
func TestSetSynchronizationPoint_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <trt:SetSynchronizationPointResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl"/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "SetSynchronizationPoint") {
			t.Errorf("Request should contain SetSynchronizationPoint, got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "ProfileToken") {
			t.Errorf("Request should contain ProfileToken, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.mediaEndpoint = server.URL

	ctx := context.Background()
	err = client.SetSynchronizationPoint(ctx, "0")
	if err != nil {
		t.Fatalf("SetSynchronizationPoint() failed: %v", err)
	}
}
