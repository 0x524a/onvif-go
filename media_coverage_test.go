package onvif

import (
	"context"
	"strings"
	"testing"
)

// This file closes the remaining gap in media.go's coverage (#10). Before it,
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

// --- Fault-path coverage: every method below 100% before this file --------

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
// (in this file or media_test.go) populates.
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
