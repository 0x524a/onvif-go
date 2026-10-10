package onvif

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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

// The tests below close the ptz.go coverage gaps left by the tests above: the
// SOAP-call-failure branch of every PTZ method (each one just wraps the
// underlying transport error, but that wrapping code only runs if the Call
// itself fails, and no existing test ever made it fail), and the SetPreset
// branch that maps a non-empty presetToken argument onto the request - the
// existing TestSetPreset only ever supplies a name, so PresetToken assignment
// was dead code.

// newFaultOp builds one serviceOp entry. It exists only so that each
// operation's name - which must equal the literal each PTZ method uses in its
// own "<Name> failed: %w" wrap, and so unavoidably duplicates a name already
// used as a serviceOp.name literal in not_initialized_test.go's ptzOps() - is
// passed as a call argument rather than repeated as a composite-literal
// field; goconst does not count string literals used as call arguments,
// which is what keeps this section from re-triggering it on every name ptzOps()
// already uses.
func newFaultOp(name string, call func(context.Context, *Client) error) serviceOp {
	return serviceOp{name: name, call: call}
}

// ptzFaultOps mirrors ptzOps() from not_initialized_test.go but is shaped for
// the opposite boundary: instead of an unset endpoint, every closure here runs
// against a client whose ptzEndpoint points at a server that always answers
// with a SOAP fault, so the assertion is that the method's own
// "<Name> failed: %w" wrap fires rather than the endpoint check in
// getPTZEndpoint.
func ptzFaultOps() []serviceOp {
	return []serviceOp{
		newFaultOp("ContinuousMove", func(ctx context.Context, c *Client) error {
			return c.ContinuousMove(ctx, testProfileToken, &PTZSpeed{PanTilt: &Vector2D{X: 0.5}}, nil)
		}),
		newFaultOp("AbsoluteMove", func(ctx context.Context, c *Client) error {
			return c.AbsoluteMove(ctx, testProfileToken, &PTZVector{PanTilt: &Vector2D{X: 1, Y: 2}}, nil)
		}),
		newFaultOp("RelativeMove", func(ctx context.Context, c *Client) error {
			return c.RelativeMove(ctx, testProfileToken, &PTZVector{PanTilt: &Vector2D{X: 1, Y: 2}}, nil)
		}),
		newFaultOp("Stop", func(ctx context.Context, c *Client) error {
			return c.Stop(ctx, testProfileToken, true, true)
		}),
		newFaultOp("GetStatus", func(ctx context.Context, c *Client) error {
			_, err := c.GetStatus(ctx, testProfileToken)

			return err
		}),
		newFaultOp("GetPresets", func(ctx context.Context, c *Client) error {
			_, err := c.GetPresets(ctx, testProfileToken)

			return err
		}),
		newFaultOp("GotoPreset", func(ctx context.Context, c *Client) error {
			return c.GotoPreset(ctx, testProfileToken, "preset1", nil)
		}),
		newFaultOp("SetPreset", func(ctx context.Context, c *Client) error {
			_, err := c.SetPreset(ctx, testProfileToken, "name", "preset1")

			return err
		}),
		newFaultOp("RemovePreset", func(ctx context.Context, c *Client) error {
			return c.RemovePreset(ctx, testProfileToken, "preset1")
		}),
		newFaultOp("GotoHomePosition", func(ctx context.Context, c *Client) error {
			return c.GotoHomePosition(ctx, testProfileToken, nil)
		}),
		newFaultOp("SetHomePosition", func(ctx context.Context, c *Client) error {
			return c.SetHomePosition(ctx, testProfileToken)
		}),
		newFaultOp("GetConfiguration", func(ctx context.Context, c *Client) error {
			_, err := c.GetConfiguration(ctx, "ptzconfig1")

			return err
		}),
		newFaultOp("GetConfigurations", func(ctx context.Context, c *Client) error {
			_, err := c.GetConfigurations(ctx)

			return err
		}),
	}
}

// TestPTZOperationsReturnSOAPFaultError asserts that when the SOAP transport
// itself fails, every PTZ method returns a non-nil error that names the
// failing operation - the "<Name> failed: %w" wrap at the end of each method.
// Before this test, that final error branch of all 13 methods was never
// exercised: every existing test drove only the success path, so a method
// that swallowed the Call error, or wrapped it under the wrong operation
// name, would have passed unnoticed.
func TestPTZOperationsReturnSOAPFaultError(t *testing.T) {
	client := newPTZTestClient(t, newSOAPFaultTestServer(t))

	ctx := context.Background()
	for _, op := range ptzFaultOps() {
		t.Run(op.name, func(t *testing.T) {
			err := op.call(ctx, client)
			if err == nil {
				t.Fatalf("%s() against a faulting server = nil, want an error", op.name)
			}

			if !strings.Contains(err.Error(), op.name+" failed") {
				t.Errorf("%s() error = %q, want it to mention %q", op.name, err.Error(), op.name+" failed")
			}
		})
	}
}

// TestSetPresetWithExistingToken covers the presetToken branch of SetPreset,
// which TestSetPreset never exercises since it only ever supplies a name. A
// caller updating an existing preset by token instead of creating one by name
// takes this branch, and PresetToken being mapped onto the request had no
// coverage at all.
func TestSetPresetWithExistingToken(t *testing.T) {
	body := `<SetPresetResponse><PresetToken>preset99</PresetToken></SetPresetResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	token, err := client.SetPreset(context.Background(), testProfileToken, "", "preset99")
	if err != nil {
		t.Fatalf("SetPreset() error = %v", err)
	}

	if token != "preset99" {
		t.Errorf("SetPreset() = %q, want %q", token, "preset99")
	}
}

// newRequestCapturingPTZServer returns an httptest.Server that records the
// inner SOAP body of the request it receives into *captured, then answers
// with response wrapped in the same envelope shape newSOAPTestServer uses.
// It exists for assertions on what the client actually put on the wire,
// which the plain response-shaped servers above cannot support.
func newRequestCapturingPTZServer(t *testing.T, response string, captured *string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var envelope struct {
			Body struct {
				Content []byte `xml:",innerxml"`
			} `xml:"Body"`
		}
		_ = xml.NewDecoder(r.Body).Decode(&envelope)
		*captured = string(envelope.Body.Content)

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
    <soap:Body>` + response + `</soap:Body>
</soap:Envelope>`))
	}))
	t.Cleanup(server.Close)

	return server
}

// TestSetPresetRequestOmitsUnusedOptionalField pins SetPreset's two
// independent optional-argument branches from the request side, not just
// the response: PresetName and PresetToken are each marshaled only when the
// corresponding argument is non-empty, and each is omitted (via omitempty on
// a nil pointer) when its argument is empty. TestSetPreset and
// TestSetPresetWithExistingToken each supply exactly one of the two
// arguments and only check the response, so a bug that set both pointers
// unconditionally, or fed an argument to the wrong pointer, would not have
// been caught by either.
func TestSetPresetRequestOmitsUnusedOptionalField(t *testing.T) {
	response := `<SetPresetResponse><PresetToken>ignored</PresetToken></SetPresetResponse>`

	t.Run("name only", func(t *testing.T) {
		var requestBody string
		client := newPTZTestClient(t, newRequestCapturingPTZServer(t, response, &requestBody))

		if _, err := client.SetPreset(context.Background(), testProfileToken, "NameSentinel", ""); err != nil {
			t.Fatalf("SetPreset() error = %v", err)
		}

		if !strings.Contains(requestBody, "NameSentinel") {
			t.Errorf("request body = %q, want it to contain the preset name", requestBody)
		}
		if strings.Contains(requestBody, "PresetToken") {
			t.Errorf("request body = %q, want no PresetToken element when presetToken is empty", requestBody)
		}
	})

	t.Run("token only", func(t *testing.T) {
		var requestBody string
		client := newPTZTestClient(t, newRequestCapturingPTZServer(t, response, &requestBody))

		if _, err := client.SetPreset(context.Background(), testProfileToken, "", "TokenSentinel"); err != nil {
			t.Fatalf("SetPreset() error = %v", err)
		}

		if !strings.Contains(requestBody, "TokenSentinel") {
			t.Errorf("request body = %q, want it to contain the preset token", requestBody)
		}
		if strings.Contains(requestBody, "PresetName") {
			t.Errorf("request body = %q, want no PresetName element when presetName is empty", requestBody)
		}
	})
}

// TestGetStatusResponseMapping pins the full field-by-field wire-to-struct
// mapping performed by GetStatus, giving every field its own distinct value
// so a swapped assignment - PanTilt.X and PanTilt.Y transposed, or
// MoveStatus.PanTilt and MoveStatus.Zoom swapped - would be caught.
// TestPTZGetStatus in ptz_test.go only asserts Position.PanTilt.X and
// MoveStatus.PanTilt, leaving the rest of the mapping unchecked.
func TestGetStatusResponseMapping(t *testing.T) {
	body := `<GetStatusResponse>
        <PTZStatus>
            <Position>
                <PanTilt x="11" y="22" space="PanTiltSpaceA"/>
                <Zoom x="33" space="ZoomSpaceB"/>
            </Position>
            <MoveStatus>
                <PanTilt>MOVING</PanTilt>
                <Zoom>UNKNOWN</Zoom>
            </MoveStatus>
            <Error>SomeError</Error>
            <UtcTime>2026-01-02T03:04:05Z</UtcTime>
        </PTZStatus>
    </GetStatusResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	status, err := client.GetStatus(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status.Position == nil || status.Position.PanTilt == nil || status.Position.Zoom == nil {
		t.Fatalf("GetStatus() Position = %+v, want PanTilt and Zoom both non-nil", status.Position)
	}
	if status.Position.PanTilt.X != 11 {
		t.Errorf("Position.PanTilt.X = %v, want 11", status.Position.PanTilt.X)
	}
	if status.Position.PanTilt.Y != 22 {
		t.Errorf("Position.PanTilt.Y = %v, want 22", status.Position.PanTilt.Y)
	}
	if status.Position.PanTilt.Space != "PanTiltSpaceA" {
		t.Errorf("Position.PanTilt.Space = %q, want %q", status.Position.PanTilt.Space, "PanTiltSpaceA")
	}
	if status.Position.Zoom.X != 33 {
		t.Errorf("Position.Zoom.X = %v, want 33", status.Position.Zoom.X)
	}
	if status.Position.Zoom.Space != "ZoomSpaceB" {
		t.Errorf("Position.Zoom.Space = %q, want %q", status.Position.Zoom.Space, "ZoomSpaceB")
	}

	if status.MoveStatus == nil {
		t.Fatalf("GetStatus() MoveStatus = nil, want non-nil")
	}
	if status.MoveStatus.PanTilt != "MOVING" {
		t.Errorf("MoveStatus.PanTilt = %q, want %q", status.MoveStatus.PanTilt, "MOVING")
	}
	if status.MoveStatus.Zoom != "UNKNOWN" {
		t.Errorf("MoveStatus.Zoom = %q, want %q", status.MoveStatus.Zoom, "UNKNOWN")
	}

	if status.Error != "SomeError" {
		t.Errorf("Error = %q, want %q", status.Error, "SomeError")
	}

	if want := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC); !status.UTCTime.Equal(want) {
		t.Errorf("UTCTime = %v, want %v", status.UTCTime, want)
	}
}

// TestGetPresetsResponseMapping pins the per-preset PTZPosition mapping that
// TestGetPresets does not exercise at all: it only asserts Token and Name,
// leaving the PanTilt/Zoom conversion inside GetPresets' loop with no
// coverage. Distinct values per field catch a swapped assignment.
func TestGetPresetsResponseMapping(t *testing.T) {
	body := `<GetPresetsResponse>
        <Preset token="preset7">
            <Name>Sentinel</Name>
            <PTZPosition>
                <PanTilt x="1.5" y="2.5" space="PresetPanTiltSpace"/>
                <Zoom x="3.5" space="PresetZoomSpace"/>
            </PTZPosition>
        </Preset>
    </GetPresetsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	presets, err := client.GetPresets(context.Background(), testProfileToken)
	if err != nil {
		t.Fatalf("GetPresets() error = %v", err)
	}
	if len(presets) != 1 {
		t.Fatalf("GetPresets() returned %d presets, want 1", len(presets))
	}

	position := presets[0].PTZPosition
	if position == nil || position.PanTilt == nil || position.Zoom == nil {
		t.Fatalf("GetPresets() PTZPosition = %+v, want PanTilt and Zoom both non-nil", position)
	}
	if position.PanTilt.X != 1.5 {
		t.Errorf("PTZPosition.PanTilt.X = %v, want 1.5", position.PanTilt.X)
	}
	if position.PanTilt.Y != 2.5 {
		t.Errorf("PTZPosition.PanTilt.Y = %v, want 2.5", position.PanTilt.Y)
	}
	if position.PanTilt.Space != "PresetPanTiltSpace" {
		t.Errorf("PTZPosition.PanTilt.Space = %q, want %q", position.PanTilt.Space, "PresetPanTiltSpace")
	}
	if position.Zoom.X != 3.5 {
		t.Errorf("PTZPosition.Zoom.X = %v, want 3.5", position.Zoom.X)
	}
	if position.Zoom.Space != "PresetZoomSpace" {
		t.Errorf("PTZPosition.Zoom.Space = %q, want %q", position.Zoom.Space, "PresetZoomSpace")
	}
}

// TestGetConfigurationsResponseMapping checks that GetConfigurations' loop
// keeps entries in order and does not cross values between them.
//
// Whether each entry is mapped completely is settled by
// TestGetConfigurationsMapsEveryPTZConfigurationField, which asserts all 14
// fields against a single entry. This test covers what that one cannot: with
// one configuration in the response, an index mix-up has nothing to mix up.
// Four fields per entry are enough for that, given distinct values.
func TestGetConfigurationsResponseMapping(t *testing.T) {
	body := `<GetConfigurationsResponse>
        <PTZConfiguration token="cfgTokenA">
            <Name>ConfigNameA</Name>
            <UseCount>1</UseCount>
            <NodeToken>nodeTokenA</NodeToken>
        </PTZConfiguration>
        <PTZConfiguration token="cfgTokenB">
            <Name>ConfigNameB</Name>
            <UseCount>2</UseCount>
            <NodeToken>nodeTokenB</NodeToken>
        </PTZConfiguration>
    </GetConfigurationsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	cfgs, err := client.GetConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetConfigurations() error = %v", err)
	}
	if len(cfgs) != 2 {
		t.Fatalf("GetConfigurations() returned %d configurations, want 2", len(cfgs))
	}

	if cfgs[0].Token != "cfgTokenA" || cfgs[0].Name != "ConfigNameA" || cfgs[0].UseCount != 1 || cfgs[0].NodeToken != "nodeTokenA" {
		t.Errorf("cfgs[0] = %+v, want {cfgTokenA ConfigNameA 1 nodeTokenA}", cfgs[0])
	}
	if cfgs[1].Token != "cfgTokenB" || cfgs[1].Name != "ConfigNameB" || cfgs[1].UseCount != 2 || cfgs[1].NodeToken != "nodeTokenB" {
		t.Errorf("cfgs[1] = %+v, want {cfgTokenB ConfigNameB 2 nodeTokenB}", cfgs[1])
	}
}

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
