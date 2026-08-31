package onvif

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
