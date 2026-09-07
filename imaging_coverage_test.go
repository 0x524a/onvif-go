package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// This file closes the coverage gap on imaging.go described in issue #12.
//
// Two things were untested before this file existed:
//
//  1. The SOAP-call-failure branch of every one of the 7 Imaging operations.
//  2. The optional-field response/request mapping that GetImagingSettings,
//     SetImagingSettings, GetOptions and GetMoveOptions each perform through
//     nil-guarded blocks. None of that mapping was verified beyond a single
//     field (Brightness) - a mis-assignment such as swapping MinGain and
//     MaxGain would have passed every existing test silently.

// newCapturingSOAPTestServer behaves like newSOAPTestServer but also records
// the raw bytes of the last request body it received, so a test can assert on
// what the client actually put on the wire rather than only on the error it
// got back.
func newCapturingSOAPTestServer(t *testing.T, respBody string) (server *httptest.Server, lastRequestBody func() string) {
	t.Helper()

	var captured []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		captured = body

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
    <soap:Body>` + respBody + `</soap:Body>
</soap:Envelope>`))
	}))
	t.Cleanup(srv.Close)

	return srv, func() string { return string(captured) }
}

// imagingFaultCase pairs one exported Imaging operation with the operation
// name its own fmt.Errorf wrap uses in the returned message. Most methods
// wrap using their own Go name, but StopFocus and GetImagingStatus build a
// request type named Stop/GetStatus and wrap the error under that name
// instead, so the two cannot be assumed equal.
type imagingFaultCase struct {
	name      string
	call      func(context.Context, *Client) error
	wantInMsg string
}

// imagingFaultCases enumerates every exported Imaging operation for the
// SOAP-call-failure test below. It intentionally mirrors the shape of
// imagingOps in not_initialized_test.go rather than reusing it directly,
// since that slice's contract is "the error result before Initialize" and
// carries no information about wrapped error message text.
func imagingFaultCases() []imagingFaultCase {
	return []imagingFaultCase{
		{"GetImagingSettings", func(ctx context.Context, c *Client) error {
			_, err := c.GetImagingSettings(ctx, testVideoSourceToken)

			return err
		}, "GetImagingSettings failed"},
		{"SetImagingSettings", func(ctx context.Context, c *Client) error {
			return c.SetImagingSettings(ctx, testVideoSourceToken, &ImagingSettings{}, true)
		}, "SetImagingSettings failed"},
		{"Move", func(ctx context.Context, c *Client) error {
			return c.Move(ctx, testVideoSourceToken, &FocusMove{})
		}, "Move failed"},
		{"GetOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetOptions(ctx, testVideoSourceToken)

			return err
		}, "GetOptions failed"},
		{"GetMoveOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetMoveOptions(ctx, testVideoSourceToken)

			return err
		}, "GetMoveOptions failed"},
		{"StopFocus", func(ctx context.Context, c *Client) error {
			return c.StopFocus(ctx, testVideoSourceToken)
		}, "Stop failed"},
		{"GetImagingStatus", func(ctx context.Context, c *Client) error {
			_, err := c.GetImagingStatus(ctx, testVideoSourceToken)

			return err
		}, "GetStatus failed"},
	}
}

// TestImagingSOAPCallFailure drives every Imaging operation against a server
// that answers with a SOAP fault, so the soapClient.Call error branch that
// each method ends with actually runs. Before this test, none of those 7
// branches had ever been exercised.
func TestImagingSOAPCallFailure(t *testing.T) {
	for _, tc := range imagingFaultCases() {
		t.Run(tc.name, func(t *testing.T) {
			client := newImagingTestClient(t, newSOAPFaultTestServer(t))

			err := tc.call(context.Background(), client)
			if err == nil {
				t.Fatalf("%s() against a faulting server = nil error, want an error", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantInMsg) {
				t.Errorf("%s() error = %q, want it to contain %q", tc.name, err.Error(), tc.wantInMsg)
			}
		})
	}
}

// TestGetImagingSettingsMapsAllOptionalBlocks sends one response carrying
// every optional field GetImagingSettings knows how to map, each with its own
// sentinel value, and checks each destination field individually. Values are
// distinct across the whole response (1-20 for numbers, a unique string per
// string field) specifically so that swapping two sibling assignments - e.g.
// MinGain and MaxGain in the Exposure block - lands a wrong value in a field
// this test checks, rather than passing by coincidence.
func TestGetImagingSettingsMapsAllOptionalBlocks(t *testing.T) {
	body := `<GetImagingSettingsResponse>
        <ImagingSettings>
            <BacklightCompensation>
                <Mode>BLC_ON</Mode>
                <Level>5</Level>
            </BacklightCompensation>
            <Brightness>1</Brightness>
            <ColorSaturation>2</ColorSaturation>
            <Contrast>3</Contrast>
            <Exposure>
                <Mode>EXP_AUTO</Mode>
                <Priority>EXP_LOWNOISE</Priority>
                <MinExposureTime>6</MinExposureTime>
                <MaxExposureTime>7</MaxExposureTime>
                <MinGain>8</MinGain>
                <MaxGain>9</MaxGain>
                <MinIris>10</MinIris>
                <MaxIris>11</MaxIris>
                <ExposureTime>12</ExposureTime>
                <Gain>13</Gain>
                <Iris>14</Iris>
            </Exposure>
            <Focus>
                <AutoFocusMode>FOCUS_AUTO</AutoFocusMode>
                <DefaultSpeed>15</DefaultSpeed>
                <NearLimit>16</NearLimit>
                <FarLimit>17</FarLimit>
            </Focus>
            <IrCutFilter>IRCUT_ON</IrCutFilter>
            <Sharpness>4</Sharpness>
            <WideDynamicRange>
                <Mode>WDR_ON</Mode>
                <Level>18</Level>
            </WideDynamicRange>
            <WhiteBalance>
                <Mode>WB_AUTO</Mode>
                <CrGain>19</CrGain>
                <CbGain>20</CbGain>
            </WhiteBalance>
        </ImagingSettings>
    </GetImagingSettingsResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	settings, err := client.GetImagingSettings(context.Background(), testVideoSourceToken)
	if err != nil {
		t.Fatalf("GetImagingSettings() error = %v", err)
	}

	checkGetImagingSettingsTopLevel(t, settings)
	checkGetImagingSettingsBacklightCompensation(t, settings)
	checkGetImagingSettingsExposure(t, settings)
	checkGetImagingSettingsFocus(t, settings)
	checkGetImagingSettingsWideDynamicRange(t, settings)
	checkGetImagingSettingsWhiteBalance(t, settings)
}

func checkGetImagingSettingsTopLevel(t *testing.T, settings *ImagingSettings) {
	t.Helper()

	if settings.Brightness == nil || *settings.Brightness != 1 {
		t.Errorf("Brightness = %v, want 1", settings.Brightness)
	}
	if settings.ColorSaturation == nil || *settings.ColorSaturation != 2 {
		t.Errorf("ColorSaturation = %v, want 2", settings.ColorSaturation)
	}
	if settings.Contrast == nil || *settings.Contrast != 3 {
		t.Errorf("Contrast = %v, want 3", settings.Contrast)
	}
	if settings.Sharpness == nil || *settings.Sharpness != 4 {
		t.Errorf("Sharpness = %v, want 4", settings.Sharpness)
	}
	if settings.IrCutFilter == nil || *settings.IrCutFilter != "IRCUT_ON" {
		t.Errorf("IrCutFilter = %v, want IRCUT_ON", settings.IrCutFilter)
	}
}

func checkGetImagingSettingsBacklightCompensation(t *testing.T, settings *ImagingSettings) {
	t.Helper()

	if settings.BacklightCompensation == nil {
		t.Fatal("BacklightCompensation = nil, want non-nil")
	}
	if settings.BacklightCompensation.Mode != "BLC_ON" {
		t.Errorf("BacklightCompensation.Mode = %q, want BLC_ON", settings.BacklightCompensation.Mode)
	}
	if settings.BacklightCompensation.Level != 5 {
		t.Errorf("BacklightCompensation.Level = %v, want 5", settings.BacklightCompensation.Level)
	}
}

func checkGetImagingSettingsExposure(t *testing.T, settings *ImagingSettings) {
	t.Helper()

	if settings.Exposure == nil {
		t.Fatal("Exposure = nil, want non-nil")
	}

	e := settings.Exposure
	if e.Mode != "EXP_AUTO" {
		t.Errorf("Exposure.Mode = %q, want EXP_AUTO", e.Mode)
	}
	if e.Priority != "EXP_LOWNOISE" {
		t.Errorf("Exposure.Priority = %q, want EXP_LOWNOISE", e.Priority)
	}
	if e.MinExposureTime != 6 {
		t.Errorf("Exposure.MinExposureTime = %v, want 6", e.MinExposureTime)
	}
	if e.MaxExposureTime != 7 {
		t.Errorf("Exposure.MaxExposureTime = %v, want 7", e.MaxExposureTime)
	}
	if e.MinGain != 8 {
		t.Errorf("Exposure.MinGain = %v, want 8", e.MinGain)
	}
	if e.MaxGain != 9 {
		t.Errorf("Exposure.MaxGain = %v, want 9", e.MaxGain)
	}
	if e.MinIris != 10 {
		t.Errorf("Exposure.MinIris = %v, want 10", e.MinIris)
	}
	if e.MaxIris != 11 {
		t.Errorf("Exposure.MaxIris = %v, want 11", e.MaxIris)
	}
	if e.ExposureTime != 12 {
		t.Errorf("Exposure.ExposureTime = %v, want 12", e.ExposureTime)
	}
	if e.Gain != 13 {
		t.Errorf("Exposure.Gain = %v, want 13", e.Gain)
	}
	if e.Iris != 14 {
		t.Errorf("Exposure.Iris = %v, want 14", e.Iris)
	}
}

func checkGetImagingSettingsFocus(t *testing.T, settings *ImagingSettings) {
	t.Helper()

	if settings.Focus == nil {
		t.Fatal("Focus = nil, want non-nil")
	}
	if settings.Focus.AutoFocusMode != "FOCUS_AUTO" {
		t.Errorf("Focus.AutoFocusMode = %q, want FOCUS_AUTO", settings.Focus.AutoFocusMode)
	}
	if settings.Focus.DefaultSpeed != 15 {
		t.Errorf("Focus.DefaultSpeed = %v, want 15", settings.Focus.DefaultSpeed)
	}
	if settings.Focus.NearLimit != 16 {
		t.Errorf("Focus.NearLimit = %v, want 16", settings.Focus.NearLimit)
	}
	if settings.Focus.FarLimit != 17 {
		t.Errorf("Focus.FarLimit = %v, want 17", settings.Focus.FarLimit)
	}
}

func checkGetImagingSettingsWideDynamicRange(t *testing.T, settings *ImagingSettings) {
	t.Helper()

	if settings.WideDynamicRange == nil {
		t.Fatal("WideDynamicRange = nil, want non-nil")
	}
	if settings.WideDynamicRange.Mode != "WDR_ON" {
		t.Errorf("WideDynamicRange.Mode = %q, want WDR_ON", settings.WideDynamicRange.Mode)
	}
	if settings.WideDynamicRange.Level != 18 {
		t.Errorf("WideDynamicRange.Level = %v, want 18", settings.WideDynamicRange.Level)
	}
}

func checkGetImagingSettingsWhiteBalance(t *testing.T, settings *ImagingSettings) {
	t.Helper()

	if settings.WhiteBalance == nil {
		t.Fatal("WhiteBalance = nil, want non-nil")
	}
	if settings.WhiteBalance.Mode != "WB_AUTO" {
		t.Errorf("WhiteBalance.Mode = %q, want WB_AUTO", settings.WhiteBalance.Mode)
	}
	if settings.WhiteBalance.CrGain != 19 {
		t.Errorf("WhiteBalance.CrGain = %v, want 19", settings.WhiteBalance.CrGain)
	}
	if settings.WhiteBalance.CbGain != 20 {
		t.Errorf("WhiteBalance.CbGain = %v, want 20", settings.WhiteBalance.CbGain)
	}
}

// TestGetImagingSettingsOmitsAbsentOptionalBlocks is the mirror image of
// TestGetImagingSettingsMapsAllOptionalBlocks: a response with none of the
// five optional blocks present must leave all five nil pointers, rather than
// the nil-guard allocating a zero-valued struct regardless of the response.
func TestGetImagingSettingsOmitsAbsentOptionalBlocks(t *testing.T) {
	body := `<GetImagingSettingsResponse>
        <ImagingSettings>
        </ImagingSettings>
    </GetImagingSettingsResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	settings, err := client.GetImagingSettings(context.Background(), testVideoSourceToken)
	if err != nil {
		t.Fatalf("GetImagingSettings() error = %v", err)
	}

	if settings.BacklightCompensation != nil {
		t.Errorf("BacklightCompensation = %+v, want nil", settings.BacklightCompensation)
	}
	if settings.Exposure != nil {
		t.Errorf("Exposure = %+v, want nil", settings.Exposure)
	}
	if settings.Focus != nil {
		t.Errorf("Focus = %+v, want nil", settings.Focus)
	}
	if settings.WideDynamicRange != nil {
		t.Errorf("WideDynamicRange = %+v, want nil", settings.WideDynamicRange)
	}
	if settings.WhiteBalance != nil {
		t.Errorf("WhiteBalance = %+v, want nil", settings.WhiteBalance)
	}
	if settings.Brightness != nil {
		t.Errorf("Brightness = %v, want nil", settings.Brightness)
	}
	if settings.IrCutFilter != nil {
		t.Errorf("IrCutFilter = %v, want nil", settings.IrCutFilter)
	}
}

// TestSetImagingSettingsSendsOptionalBlocks builds an ImagingSettings with
// every optional block populated with its own sentinel value and inspects the
// XML actually sent over the wire, not just the returned error. A "set"
// method whose fields never reach the request is invisible to a test that
// only checks err == nil, which is exactly why TestSetImagingSettings alone
// was not enough.
func TestSetImagingSettingsSendsOptionalBlocks(t *testing.T) {
	server, lastRequestBody := newCapturingSOAPTestServer(t, soapAckOnlyResponse)
	client := newImagingTestClient(t, server)

	brightness := 101.0
	colorSaturation := 102.0
	contrast := 103.0
	sharpness := 104.0
	irCutFilter := "SET_IRCUT_ON"

	settings := &ImagingSettings{
		Brightness:      &brightness,
		ColorSaturation: &colorSaturation,
		Contrast:        &contrast,
		Sharpness:       &sharpness,
		IrCutFilter:     &irCutFilter,
		BacklightCompensation: &BacklightCompensation{
			Mode:  "SET_BLC_ON",
			Level: 105,
		},
		Exposure: &Exposure{
			Mode:            "SET_EXP_AUTO",
			Priority:        "SET_EXP_LOWNOISE",
			MinExposureTime: 106,
			MaxExposureTime: 107,
			MinGain:         108,
			MaxGain:         109,
			MinIris:         110,
			MaxIris:         111,
			ExposureTime:    112,
			Gain:            113,
			Iris:            114,
		},
		Focus: &FocusConfiguration{
			AutoFocusMode: "SET_FOCUS_AUTO",
			DefaultSpeed:  115,
			NearLimit:     116,
			FarLimit:      117,
		},
		WideDynamicRange: &WideDynamicRange{
			Mode:  "SET_WDR_ON",
			Level: 118,
		},
		WhiteBalance: &WhiteBalance{
			Mode:   "SET_WB_AUTO",
			CrGain: 119,
			CbGain: 120,
		},
	}

	if err := client.SetImagingSettings(context.Background(), testVideoSourceToken, settings, true); err != nil {
		t.Fatalf("SetImagingSettings() error = %v", err)
	}

	sent := lastRequestBody()

	checkSetImagingSettingsTopLevelSent(t, sent)
	checkSetImagingSettingsExposureSent(t, sent)
	checkSetImagingSettingsFocusSent(t, sent)
	checkSetImagingSettingsOtherBlocksSent(t, sent)
}

func checkSetImagingSettingsTopLevelSent(t *testing.T, sent string) {
	t.Helper()

	wants := []string{
		"<Brightness>101</Brightness>",
		"<ColorSaturation>102</ColorSaturation>",
		"<Contrast>103</Contrast>",
		"<Sharpness>104</Sharpness>",
		"<IrCutFilter>SET_IRCUT_ON</IrCutFilter>",
		"<Mode>SET_BLC_ON</Mode>",
		"<Level>105</Level>",
	}
	for _, want := range wants {
		if !strings.Contains(sent, want) {
			t.Errorf("request body missing %q\nfull body:\n%s", want, sent)
		}
	}
}

func checkSetImagingSettingsExposureSent(t *testing.T, sent string) {
	t.Helper()

	wants := []string{
		"<Mode>SET_EXP_AUTO</Mode>",
		"<Priority>SET_EXP_LOWNOISE</Priority>",
		"<MinExposureTime>106</MinExposureTime>",
		"<MaxExposureTime>107</MaxExposureTime>",
		"<MinGain>108</MinGain>",
		"<MaxGain>109</MaxGain>",
		"<MinIris>110</MinIris>",
		"<MaxIris>111</MaxIris>",
		"<ExposureTime>112</ExposureTime>",
		"<Gain>113</Gain>",
		"<Iris>114</Iris>",
	}
	for _, want := range wants {
		if !strings.Contains(sent, want) {
			t.Errorf("request body missing %q\nfull body:\n%s", want, sent)
		}
	}
}

func checkSetImagingSettingsFocusSent(t *testing.T, sent string) {
	t.Helper()

	wants := []string{
		"<AutoFocusMode>SET_FOCUS_AUTO</AutoFocusMode>",
		"<DefaultSpeed>115</DefaultSpeed>",
		"<NearLimit>116</NearLimit>",
		"<FarLimit>117</FarLimit>",
	}
	for _, want := range wants {
		if !strings.Contains(sent, want) {
			t.Errorf("request body missing %q\nfull body:\n%s", want, sent)
		}
	}
}

func checkSetImagingSettingsOtherBlocksSent(t *testing.T, sent string) {
	t.Helper()

	wants := []string{
		"<Mode>SET_WDR_ON</Mode>",
		"<Level>118</Level>",
		"<Mode>SET_WB_AUTO</Mode>",
		"<CrGain>119</CrGain>",
		"<CbGain>120</CbGain>",
	}
	for _, want := range wants {
		if !strings.Contains(sent, want) {
			t.Errorf("request body missing %q\nfull body:\n%s", want, sent)
		}
	}
}

// TestSetImagingSettingsOmitsNilOptionalBlocks checks the other direction: an
// ImagingSettings with only Brightness set must not put any of the other
// optional blocks on the wire at all, proving the ",omitempty" request tags
// actually suppress them rather than just being decorative.
func TestSetImagingSettingsOmitsNilOptionalBlocks(t *testing.T) {
	server, lastRequestBody := newCapturingSOAPTestServer(t, soapAckOnlyResponse)
	client := newImagingTestClient(t, server)

	brightness := 42.0
	settings := &ImagingSettings{Brightness: &brightness}

	if err := client.SetImagingSettings(context.Background(), testVideoSourceToken, settings, false); err != nil {
		t.Fatalf("SetImagingSettings() error = %v", err)
	}

	sent := lastRequestBody()
	if !strings.Contains(sent, "<Brightness>42</Brightness>") {
		t.Errorf("request body missing Brightness\nfull body:\n%s", sent)
	}

	unwanted := []string{
		"<BacklightCompensation", "<Exposure", "<Focus", "<WideDynamicRange", "<WhiteBalance",
		"<ColorSaturation", "<Contrast", "<Sharpness", "<IrCutFilter",
	}
	for _, tag := range unwanted {
		if strings.Contains(sent, tag) {
			t.Errorf("request body unexpectedly contains %q when the field was nil\nfull body:\n%s", tag, sent)
		}
	}
}

// TestGetOptionsMapsColorSaturationAndContrast covers the two ImagingOptions
// fields the existing TestImagingGetOptions did not: ColorSaturation and
// Contrast, alongside a re-check of Brightness with a distinct value so a
// three-way swap between the three FloatRange blocks would be caught.
//
// Three fields is the whole of what GetOptions returns today, not the whole
// of ImagingOptions: it parses BacklightCompensation, Exposure and Focus and
// then discards them, and never parses Sharpness, WideDynamicRange,
// WhiteBalance or IrCutFilterModes at all. That is #90. Asserting those here
// would either fail or, worse, pin their zero values as though a camera had
// reported nothing - so this test deliberately covers only what is mapped,
// and #90 carries the rest.
func TestGetOptionsMapsColorSaturationAndContrast(t *testing.T) {
	body := `<GetOptionsResponse>
        <ImagingOptions>
            <Brightness><Min>1</Min><Max>2</Max></Brightness>
            <ColorSaturation><Min>3</Min><Max>4</Max></ColorSaturation>
            <Contrast><Min>5</Min><Max>6</Max></Contrast>
        </ImagingOptions>
    </GetOptionsResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	options, err := client.GetOptions(context.Background(), testVideoSourceToken)
	if err != nil {
		t.Fatalf("GetOptions() error = %v", err)
	}

	if options.Brightness == nil || options.Brightness.Min != 1 || options.Brightness.Max != 2 {
		t.Errorf("Brightness = %+v, want {Min:1 Max:2}", options.Brightness)
	}
	if options.ColorSaturation == nil || options.ColorSaturation.Min != 3 || options.ColorSaturation.Max != 4 {
		t.Errorf("ColorSaturation = %+v, want {Min:3 Max:4}", options.ColorSaturation)
	}
	if options.Contrast == nil || options.Contrast.Min != 5 || options.Contrast.Max != 6 {
		t.Errorf("Contrast = %+v, want {Min:5 Max:6}", options.Contrast)
	}
}

// TestGetMoveOptionsMapsRelativeAndContinuous covers the two MoveOptions
// blocks the existing TestGetMoveOptions did not: Relative and Continuous,
// each with distinct sentinels across all three blocks so a mix-up between
// Absolute/Relative/Continuous, or between their Position/Distance/Speed
// sub-fields, would show up as a wrong value rather than passing silently.
func TestGetMoveOptionsMapsRelativeAndContinuous(t *testing.T) {
	body := `<GetMoveOptionsResponse>
        <MoveOptions>
            <Absolute>
                <Position><Min>1</Min><Max>2</Max></Position>
                <Speed><Min>3</Min><Max>4</Max></Speed>
            </Absolute>
            <Relative>
                <Distance><Min>5</Min><Max>6</Max></Distance>
                <Speed><Min>7</Min><Max>8</Max></Speed>
            </Relative>
            <Continuous>
                <Speed><Min>9</Min><Max>10</Max></Speed>
            </Continuous>
        </MoveOptions>
    </GetMoveOptionsResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	options, err := client.GetMoveOptions(context.Background(), testVideoSourceToken)
	if err != nil {
		t.Fatalf("GetMoveOptions() error = %v", err)
	}

	checkMoveOptionsAbsolute(t, options)
	checkMoveOptionsRelative(t, options)
	checkMoveOptionsContinuous(t, options)
}

func checkMoveOptionsAbsolute(t *testing.T, options *MoveOptions) {
	t.Helper()

	if options.Absolute == nil {
		t.Fatal("Absolute = nil, want non-nil")
	}
	if options.Absolute.Position.Min != 1 || options.Absolute.Position.Max != 2 {
		t.Errorf("Absolute.Position = %+v, want {Min:1 Max:2}", options.Absolute.Position)
	}
	if options.Absolute.Speed.Min != 3 || options.Absolute.Speed.Max != 4 {
		t.Errorf("Absolute.Speed = %+v, want {Min:3 Max:4}", options.Absolute.Speed)
	}
}

func checkMoveOptionsRelative(t *testing.T, options *MoveOptions) {
	t.Helper()

	if options.Relative == nil {
		t.Fatal("Relative = nil, want non-nil")
	}
	if options.Relative.Distance.Min != 5 || options.Relative.Distance.Max != 6 {
		t.Errorf("Relative.Distance = %+v, want {Min:5 Max:6}", options.Relative.Distance)
	}
	if options.Relative.Speed.Min != 7 || options.Relative.Speed.Max != 8 {
		t.Errorf("Relative.Speed = %+v, want {Min:7 Max:8}", options.Relative.Speed)
	}
}

func checkMoveOptionsContinuous(t *testing.T, options *MoveOptions) {
	t.Helper()

	if options.Continuous == nil {
		t.Fatal("Continuous = nil, want non-nil")
	}
	if options.Continuous.Speed.Min != 9 || options.Continuous.Speed.Max != 10 {
		t.Errorf("Continuous.Speed = %+v, want {Min:9 Max:10}", options.Continuous.Speed)
	}
}
