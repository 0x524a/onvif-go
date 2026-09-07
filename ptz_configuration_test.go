package onvif

import (
	"context"
	"testing"
	"time"
)

// Tests for #87 and #89: PTZ responses that parsed fields off the wire and
// then dropped them.
//
// #87: GetConfiguration and GetConfigurations returned a PTZConfiguration with
// 10 of its 14 fields permanently zero, because the response structs declared
// only the other four. #89: GetStatus declared UtcTime and never assigned it.
//
// Every value below is distinct. That is deliberate: a mapping test whose
// fields share values cannot catch two of them being swapped, which is the
// mistake most likely to be made when wiring up a struct this wide.

const ptzConfigurationResponseBody = `
    <PTZConfiguration token="cfg-42">
        <Name>Main PTZ</Name>
        <UseCount>7</UseCount>
        <NodeToken>node-9</NodeToken>
        <DefaultAbsolutePantTiltPositionSpace>space/abs-pantilt</DefaultAbsolutePantTiltPositionSpace>
        <DefaultAbsoluteZoomPositionSpace>space/abs-zoom</DefaultAbsoluteZoomPositionSpace>
        <DefaultRelativePanTiltTranslationSpace>space/rel-pantilt</DefaultRelativePanTiltTranslationSpace>
        <DefaultRelativeZoomTranslationSpace>space/rel-zoom</DefaultRelativeZoomTranslationSpace>
        <DefaultContinuousPanTiltVelocitySpace>space/cont-pantilt</DefaultContinuousPanTiltVelocitySpace>
        <DefaultContinuousZoomVelocitySpace>space/cont-zoom</DefaultContinuousZoomVelocitySpace>
        <DefaultPTZSpeed>
            <PanTilt x="0.11" y="0.22" space="space/speed-pantilt"/>
            <Zoom x="0.33" space="space/speed-zoom"/>
        </DefaultPTZSpeed>
        <DefaultPTZTimeout>PT4M5S</DefaultPTZTimeout>
        <PanTiltLimits>
            <Range>
                <URI>space/pantilt-limit</URI>
                <XRange><Min>-11</Min><Max>12</Max></XRange>
                <YRange><Min>-13</Min><Max>14</Max></YRange>
            </Range>
        </PanTiltLimits>
        <ZoomLimits>
            <Range>
                <URI>space/zoom-limit</URI>
                <XRange><Min>15</Min><Max>16</Max></XRange>
            </Range>
        </ZoomLimits>
    </PTZConfiguration>`

// assertFullPTZConfiguration checks every field of the configuration described
// by ptzConfigurationResponseBody.
func assertFullPTZConfiguration(t *testing.T, config *PTZConfiguration) {
	t.Helper()

	for _, check := range []struct {
		field string
		got   string
		want  string
	}{
		{"Token", config.Token, "cfg-42"},
		{"Name", config.Name, "Main PTZ"},
		{"NodeToken", config.NodeToken, "node-9"},
		{"DefaultAbsolutePantTiltPositionSpace", config.DefaultAbsolutePantTiltPositionSpace, "space/abs-pantilt"},
		{"DefaultAbsoluteZoomPositionSpace", config.DefaultAbsoluteZoomPositionSpace, "space/abs-zoom"},
		{"DefaultRelativePanTiltTranslationSpace", config.DefaultRelativePanTiltTranslationSpace, "space/rel-pantilt"},
		{"DefaultRelativeZoomTranslationSpace", config.DefaultRelativeZoomTranslationSpace, "space/rel-zoom"},
		{"DefaultContinuousPanTiltVelocitySpace", config.DefaultContinuousPanTiltVelocitySpace, "space/cont-pantilt"},
		{"DefaultContinuousZoomVelocitySpace", config.DefaultContinuousZoomVelocitySpace, "space/cont-zoom"},
	} {
		if check.got != check.want {
			t.Errorf("%s = %q, want %q", check.field, check.got, check.want)
		}
	}

	if config.UseCount != 7 {
		t.Errorf("UseCount = %d, want 7", config.UseCount)
	}

	// An xs:duration, and the reason this whole set of fields needed a parser
	// before it could be mapped at all.
	if want := 4*time.Minute + 5*time.Second; config.DefaultPTZTimeout != want {
		t.Errorf("DefaultPTZTimeout = %v, want %v", config.DefaultPTZTimeout, want)
	}

	assertPTZSpeed(t, config.DefaultPTZSpeed)
	assertPanTiltLimits(t, config.PanTiltLimits)
	assertZoomLimits(t, config.ZoomLimits)
}

func assertPTZSpeed(t *testing.T, speed *PTZSpeed) {
	t.Helper()

	if speed == nil {
		t.Fatal("DefaultPTZSpeed = nil, want the speed from the response")
	}
	if speed.PanTilt == nil {
		t.Fatal("DefaultPTZSpeed.PanTilt = nil")
	}
	if speed.PanTilt.X != 0.11 || speed.PanTilt.Y != 0.22 {
		t.Errorf("DefaultPTZSpeed.PanTilt = (%v, %v), want (0.11, 0.22)", speed.PanTilt.X, speed.PanTilt.Y)
	}
	if speed.PanTilt.Space != "space/speed-pantilt" {
		t.Errorf("DefaultPTZSpeed.PanTilt.Space = %q, want %q", speed.PanTilt.Space, "space/speed-pantilt")
	}
	if speed.Zoom == nil {
		t.Fatal("DefaultPTZSpeed.Zoom = nil")
	}
	if speed.Zoom.X != 0.33 {
		t.Errorf("DefaultPTZSpeed.Zoom.X = %v, want 0.33", speed.Zoom.X)
	}
	if speed.Zoom.Space != "space/speed-zoom" {
		t.Errorf("DefaultPTZSpeed.Zoom.Space = %q, want %q", speed.Zoom.Space, "space/speed-zoom")
	}
}

func assertPanTiltLimits(t *testing.T, limits *PanTiltLimits) {
	t.Helper()

	if limits == nil || limits.Range == nil {
		t.Fatal("PanTiltLimits.Range = nil, want the limits from the response")
	}
	if limits.Range.URI != "space/pantilt-limit" {
		t.Errorf("PanTiltLimits.Range.URI = %q, want %q", limits.Range.URI, "space/pantilt-limit")
	}
	if limits.Range.XRange == nil || limits.Range.XRange.Min != -11 || limits.Range.XRange.Max != 12 {
		t.Errorf("PanTiltLimits.Range.XRange = %+v, want {-11 12}", limits.Range.XRange)
	}
	if limits.Range.YRange == nil || limits.Range.YRange.Min != -13 || limits.Range.YRange.Max != 14 {
		t.Errorf("PanTiltLimits.Range.YRange = %+v, want {-13 14}", limits.Range.YRange)
	}
}

func assertZoomLimits(t *testing.T, limits *ZoomLimits) {
	t.Helper()

	if limits == nil || limits.Range == nil {
		t.Fatal("ZoomLimits.Range = nil, want the limits from the response")
	}
	if limits.Range.URI != "space/zoom-limit" {
		t.Errorf("ZoomLimits.Range.URI = %q, want %q", limits.Range.URI, "space/zoom-limit")
	}
	if limits.Range.XRange == nil || limits.Range.XRange.Min != 15 || limits.Range.XRange.Max != 16 {
		t.Errorf("ZoomLimits.Range.XRange = %+v, want {15 16}", limits.Range.XRange)
	}
}

func TestGetConfigurationMapsEveryPTZConfigurationField(t *testing.T) {
	body := `<GetConfigurationResponse>` + ptzConfigurationResponseBody + `</GetConfigurationResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	config, err := client.GetConfiguration(context.Background(), "cfg-42")
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}

	assertFullPTZConfiguration(t, config)
}

func TestGetConfigurationsMapsEveryPTZConfigurationField(t *testing.T) {
	body := `<GetConfigurationsResponse>` + ptzConfigurationResponseBody + `</GetConfigurationsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	configs, err := client.GetConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetConfigurations() error = %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("GetConfigurations() returned %d configurations, want 1", len(configs))
	}

	// The plural getter had its own copy of the truncated response struct, so
	// it needs its own assertion rather than trusting the singular one.
	assertFullPTZConfiguration(t, configs[0])
}

// TestGetConfigurationOptionalsAbsent keeps the optional sub-structures nil
// rather than allocating zero-valued ones, so a caller can tell "the camera
// did not report limits" from "the limits are zero".
func TestGetConfigurationOptionalsAbsent(t *testing.T) {
	body := `<GetConfigurationResponse>
        <PTZConfiguration token="cfg-min">
            <Name>Minimal</Name>
            <UseCount>1</UseCount>
            <NodeToken>node-1</NodeToken>
        </PTZConfiguration>
    </GetConfigurationResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	config, err := client.GetConfiguration(context.Background(), "cfg-min")
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}

	if config.DefaultPTZSpeed != nil {
		t.Errorf("DefaultPTZSpeed = %+v, want nil when absent", config.DefaultPTZSpeed)
	}
	if config.PanTiltLimits != nil {
		t.Errorf("PanTiltLimits = %+v, want nil when absent", config.PanTiltLimits)
	}
	if config.ZoomLimits != nil {
		t.Errorf("ZoomLimits = %+v, want nil when absent", config.ZoomLimits)
	}
	if config.DefaultPTZTimeout != 0 {
		t.Errorf("DefaultPTZTimeout = %v, want 0 when absent", config.DefaultPTZTimeout)
	}

	// The four fields that always worked must keep working.
	if config.Token != "cfg-min" || config.NodeToken != "node-1" {
		t.Errorf("Token/NodeToken = %q/%q, want cfg-min/node-1", config.Token, config.NodeToken)
	}
}

// TestGetConfigurationPartialLimits covers a camera that reports a limit range
// but omits one of its axes. The axis must come back nil, so that "no limit
// reported for this axis" stays distinguishable from "this axis is limited to
// 0..0" - a range a caller would read as the axis being immovable.
func TestGetConfigurationPartialLimits(t *testing.T) {
	body := `<GetConfigurationResponse>
        <PTZConfiguration token="cfg-partial">
            <Name>Partial</Name>
            <UseCount>1</UseCount>
            <NodeToken>node-1</NodeToken>
            <PanTiltLimits>
                <Range>
                    <URI>space/pantilt-limit</URI>
                    <XRange><Min>-1</Min><Max>2</Max></XRange>
                </Range>
            </PanTiltLimits>
        </PTZConfiguration>
    </GetConfigurationResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	config, err := client.GetConfiguration(context.Background(), "cfg-partial")
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}

	if config.PanTiltLimits == nil || config.PanTiltLimits.Range == nil {
		t.Fatalf("PanTiltLimits = %+v, want the range from the response", config.PanTiltLimits)
	}
	if config.PanTiltLimits.Range.XRange == nil {
		t.Error("PanTiltLimits.Range.XRange = nil, want the reported range")
	}
	if config.PanTiltLimits.Range.YRange != nil {
		t.Errorf("PanTiltLimits.Range.YRange = %+v, want nil when the camera omits it",
			config.PanTiltLimits.Range.YRange)
	}
}

// TestGetStatusMapsUTCTime covers #89. The zero time was indistinguishable
// from a camera that reported no timestamp, which is why no existing GetStatus
// assertion could fail on it.
func TestGetStatusMapsUTCTime(t *testing.T) {
	tests := []struct {
		name    string
		element string
		want    time.Time
	}{
		{
			name:    "reported",
			element: "<UtcTime>2026-09-06T14:30:45Z</UtcTime>",
			want:    time.Date(2026, time.September, 6, 14, 30, 45, 0, time.UTC),
		},
		// Both remaining cases yield the zero time, deliberately: a status
		// reading is still worth returning without a usable timestamp.
		{name: "absent", element: ""},
		{name: "malformed", element: "<UtcTime>yesterday</UtcTime>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `<GetStatusResponse>
                <PTZStatus>
                    <Position><Zoom x="0.5"/></Position>
                    ` + tt.element + `
                </PTZStatus>
            </GetStatusResponse>`
			client := newPTZTestClient(t, newSOAPTestServer(t, body))

			status, err := client.GetStatus(context.Background(), testProfileToken)
			if err != nil {
				t.Fatalf("GetStatus() error = %v", err)
			}
			if !status.UTCTime.Equal(tt.want) {
				t.Errorf("UTCTime = %v, want %v", status.UTCTime, tt.want)
			}

			// The rest of the status must survive either way.
			if status.Position == nil || status.Position.Zoom == nil || status.Position.Zoom.X != 0.5 {
				t.Errorf("Position = %+v, want Zoom.X 0.5", status.Position)
			}
		})
	}
}
