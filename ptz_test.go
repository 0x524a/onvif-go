package onvif

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newSOAPTestServer returns an httptest.Server that responds to every
// request with a fixed SOAP envelope wrapping body. Each test below only
// ever calls one client method against it, so no action-based dispatch
// (unlike MockONVIFServer's) is needed.
func newSOAPTestServer(t *testing.T, body string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
    <soap:Body>` + body + `</soap:Body>
</soap:Envelope>`))
	}))
	t.Cleanup(server.Close)

	return server
}

// newPTZTestClient returns a Client whose ptzEndpoint points at server.
func newPTZTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.ptzEndpoint = server.URL

	return client
}

const soapAckOnlyResponse = ""

func TestContinuousMove(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.ContinuousMove(context.Background(), "profile1", &PTZSpeed{PanTilt: &Vector2D{X: 0.5}}, nil); err != nil {
		t.Errorf("ContinuousMove() error = %v", err)
	}
}

func TestAbsoluteMove(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.AbsoluteMove(context.Background(), "profile1", &PTZVector{PanTilt: &Vector2D{X: 1, Y: 2}}, nil); err != nil {
		t.Errorf("AbsoluteMove() error = %v", err)
	}
}

func TestRelativeMove(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.RelativeMove(context.Background(), "profile1", &PTZVector{PanTilt: &Vector2D{X: 1, Y: 2}}, nil); err != nil {
		t.Errorf("RelativeMove() error = %v", err)
	}
}

func TestPTZStop(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.Stop(context.Background(), "profile1", true, true); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestPTZGetStatus(t *testing.T) {
	body := `<GetStatusResponse>
        <PTZStatus>
            <Position>
                <PanTilt x="0.1" y="0.2" space="PanTiltSpace"/>
                <Zoom x="0.3" space="ZoomSpace"/>
            </Position>
            <MoveStatus>
                <PanTilt>IDLE</PanTilt>
                <Zoom>IDLE</Zoom>
            </MoveStatus>
            <UtcTime>2026-01-01T00:00:00Z</UtcTime>
        </PTZStatus>
    </GetStatusResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	status, err := client.GetStatus(context.Background(), "profile1")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Position == nil || status.Position.PanTilt == nil || status.Position.PanTilt.X != 0.1 {
		t.Errorf("GetStatus() Position = %+v, want PanTilt.X = 0.1", status.Position)
	}
	if status.MoveStatus == nil || status.MoveStatus.PanTilt != "IDLE" {
		t.Errorf("GetStatus() MoveStatus = %+v, want PanTilt = IDLE", status.MoveStatus)
	}
}

func TestGetPresets(t *testing.T) {
	body := `<GetPresetsResponse>
        <Preset token="preset1">
            <Name>Home</Name>
            <PTZPosition>
                <PanTilt x="1" y="2"/>
                <Zoom x="3"/>
            </PTZPosition>
        </Preset>
    </GetPresetsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	presets, err := client.GetPresets(context.Background(), "profile1")
	if err != nil {
		t.Fatalf("GetPresets() error = %v", err)
	}
	if len(presets) != 1 || presets[0].Token != "preset1" || presets[0].Name != "Home" {
		t.Errorf("GetPresets() = %+v, want one preset {preset1 Home}", presets)
	}
}

func TestGotoPreset(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.GotoPreset(context.Background(), "profile1", "preset1", nil); err != nil {
		t.Errorf("GotoPreset() error = %v", err)
	}
}

func TestSetPreset(t *testing.T) {
	body := `<SetPresetResponse><PresetToken>preset42</PresetToken></SetPresetResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	token, err := client.SetPreset(context.Background(), "profile1", "Home", "")
	if err != nil {
		t.Fatalf("SetPreset() error = %v", err)
	}
	if token != "preset42" {
		t.Errorf("SetPreset() = %q, want %q", token, "preset42")
	}
}

func TestRemovePreset(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.RemovePreset(context.Background(), "profile1", "preset1"); err != nil {
		t.Errorf("RemovePreset() error = %v", err)
	}
}

func TestGotoHomePosition(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.GotoHomePosition(context.Background(), "profile1", nil); err != nil {
		t.Errorf("GotoHomePosition() error = %v", err)
	}
}

func TestSetHomePosition(t *testing.T) {
	client := newPTZTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))
	if err := client.SetHomePosition(context.Background(), "profile1"); err != nil {
		t.Errorf("SetHomePosition() error = %v", err)
	}
}

func TestGetConfiguration(t *testing.T) {
	body := `<GetConfigurationResponse>
        <PTZConfiguration token="cfg1">
            <Name>PTZ Config</Name>
            <UseCount>1</UseCount>
            <NodeToken>node1</NodeToken>
        </PTZConfiguration>
    </GetConfigurationResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	cfg, err := client.GetConfiguration(context.Background(), "cfg1")
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if cfg.Token != "cfg1" || cfg.NodeToken != "node1" {
		t.Errorf("GetConfiguration() = %+v, want Token=cfg1 NodeToken=node1", cfg)
	}
}

func TestGetConfigurations(t *testing.T) {
	body := `<GetConfigurationsResponse>
        <PTZConfiguration token="cfg1">
            <Name>PTZ Config</Name>
            <UseCount>1</UseCount>
            <NodeToken>node1</NodeToken>
        </PTZConfiguration>
    </GetConfigurationsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	cfgs, err := client.GetConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetConfigurations() error = %v", err)
	}
	if len(cfgs) != 1 || cfgs[0].Token != "cfg1" {
		t.Errorf("GetConfigurations() = %+v, want one configuration with Token=cfg1", cfgs)
	}
}

func TestConvertToPTZVectorXML(t *testing.T) {
	if got := convertToPTZVectorXML(nil); got != nil {
		t.Errorf("convertToPTZVectorXML(nil) = %+v, want nil", got)
	}

	v := &PTZVector{
		PanTilt: &Vector2D{X: 0.1, Y: -0.2, Space: "PanTiltSpace"},
		Zoom:    &Vector1D{X: 0.5, Space: "ZoomSpace"},
	}
	got := convertToPTZVectorXML(v)
	if got == nil {
		t.Fatal("convertToPTZVectorXML() = nil, want non-nil")
	}
	if got.PanTilt == nil || got.PanTilt.X != 0.1 || got.PanTilt.Y != -0.2 || got.PanTilt.Space != "PanTiltSpace" {
		t.Errorf("PanTilt = %+v, want {0.1 -0.2 PanTiltSpace}", got.PanTilt)
	}
	if got.Zoom == nil || got.Zoom.X != 0.5 || got.Zoom.Space != "ZoomSpace" {
		t.Errorf("Zoom = %+v, want {0.5 ZoomSpace}", got.Zoom)
	}

	partial := convertToPTZVectorXML(&PTZVector{PanTilt: &Vector2D{X: 1, Y: 2}})
	if partial.PanTilt == nil {
		t.Error("PanTilt = nil, want non-nil")
	}
	if partial.Zoom != nil {
		t.Errorf("Zoom = %+v, want nil when source Zoom is nil", partial.Zoom)
	}
}

func TestConvertToPTZSpeedXML(t *testing.T) {
	if got := convertToPTZSpeedXML(nil); got != nil {
		t.Errorf("convertToPTZSpeedXML(nil) = %+v, want nil", got)
	}

	s := &PTZSpeed{
		PanTilt: &Vector2D{X: 0.3, Y: 0.4, Space: "PanTiltSpeedSpace"},
		Zoom:    &Vector1D{X: 0.6, Space: "ZoomSpeedSpace"},
	}
	got := convertToPTZSpeedXML(s)
	if got == nil {
		t.Fatal("convertToPTZSpeedXML() = nil, want non-nil")
	}
	if got.PanTilt == nil || got.PanTilt.X != 0.3 || got.PanTilt.Y != 0.4 || got.PanTilt.Space != "PanTiltSpeedSpace" {
		t.Errorf("PanTilt = %+v, want {0.3 0.4 PanTiltSpeedSpace}", got.PanTilt)
	}
	if got.Zoom == nil || got.Zoom.X != 0.6 || got.Zoom.Space != "ZoomSpeedSpace" {
		t.Errorf("Zoom = %+v, want {0.6 ZoomSpeedSpace}", got.Zoom)
	}

	partial := convertToPTZSpeedXML(&PTZSpeed{Zoom: &Vector1D{X: 9}})
	if partial.Zoom == nil {
		t.Error("Zoom = nil, want non-nil")
	}
	if partial.PanTilt != nil {
		t.Errorf("PanTilt = %+v, want nil when source PanTilt is nil", partial.PanTilt)
	}
}
