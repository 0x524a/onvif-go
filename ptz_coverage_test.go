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

// This file closes the ptz.go coverage gaps left by ptz_test.go: the
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
// which is what keeps this file from re-triggering it on every name ptzOps()
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
