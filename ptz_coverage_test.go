package onvif

import (
	"context"
	"strings"
	"testing"
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
