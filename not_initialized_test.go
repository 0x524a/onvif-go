package onvif

import (
	"context"
	"errors"
	"testing"
)

// This file covers the error contract described in #64: an empty service
// endpoint has two entirely different causes, and callers need to tell them
// apart because the recovery differs. ErrNotInitialized means "call Initialize
// and retry"; ErrServiceNotSupported means "this device will never do this,
// stop offering it". Before #64 both cases returned ErrServiceNotSupported, so
// a missed setup step was indistinguishable from a device limitation.

// serviceOp is one PTZ or Imaging operation, reduced to just its error result.
// Argument values are irrelevant: every one of these must fail on the endpoint
// check before it builds a request or touches the network.
type serviceOp struct {
	name string
	call func(context.Context, *Client) error
}

// ptzOps is every exported PTZ operation. All 13 must agree on the contract -
// a single one still returning the old error would leave callers unable to
// trust errors.Is on any of them.
func ptzOps() []serviceOp {
	return []serviceOp{
		{"ContinuousMove", func(ctx context.Context, c *Client) error {
			return c.ContinuousMove(ctx, testProfileToken, nil, nil)
		}},
		{"AbsoluteMove", func(ctx context.Context, c *Client) error {
			return c.AbsoluteMove(ctx, testProfileToken, nil, nil)
		}},
		{"RelativeMove", func(ctx context.Context, c *Client) error {
			return c.RelativeMove(ctx, testProfileToken, nil, nil)
		}},
		{"Stop", func(ctx context.Context, c *Client) error {
			return c.Stop(ctx, testProfileToken, true, true)
		}},
		{"GetStatus", func(ctx context.Context, c *Client) error {
			_, err := c.GetStatus(ctx, testProfileToken)

			return err
		}},
		{"GetPresets", func(ctx context.Context, c *Client) error {
			_, err := c.GetPresets(ctx, testProfileToken)

			return err
		}},
		{"GotoPreset", func(ctx context.Context, c *Client) error {
			return c.GotoPreset(ctx, testProfileToken, "preset1", nil)
		}},
		{"SetPreset", func(ctx context.Context, c *Client) error {
			_, err := c.SetPreset(ctx, testProfileToken, "name", "preset1")

			return err
		}},
		{"RemovePreset", func(ctx context.Context, c *Client) error {
			return c.RemovePreset(ctx, testProfileToken, "preset1")
		}},
		{"GotoHomePosition", func(ctx context.Context, c *Client) error {
			return c.GotoHomePosition(ctx, testProfileToken, nil)
		}},
		{"SetHomePosition", func(ctx context.Context, c *Client) error {
			return c.SetHomePosition(ctx, testProfileToken)
		}},
		{"GetConfiguration", func(ctx context.Context, c *Client) error {
			_, err := c.GetConfiguration(ctx, "ptzconfig1")

			return err
		}},
		{"GetConfigurations", func(ctx context.Context, c *Client) error {
			_, err := c.GetConfigurations(ctx)

			return err
		}},
	}
}

// imagingOps is every exported Imaging operation.
//
// GetImagingSettings, SetImagingSettings and Move are the three that used to
// silently fall back to the device endpoint while the other four returned an
// error, so the same client state produced a working call or a "not supported"
// error depending only on which method the caller reached for. They are
// included here precisely because that split is what #64 removed.
func imagingOps() []serviceOp {
	return []serviceOp{
		{"GetImagingSettings", func(ctx context.Context, c *Client) error {
			_, err := c.GetImagingSettings(ctx, testVideoSourceToken)

			return err
		}},
		{"SetImagingSettings", func(ctx context.Context, c *Client) error {
			return c.SetImagingSettings(ctx, testVideoSourceToken, &ImagingSettings{}, true)
		}},
		{"Move", func(ctx context.Context, c *Client) error {
			return c.Move(ctx, testVideoSourceToken, &FocusMove{})
		}},
		{"GetOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetOptions(ctx, testVideoSourceToken)

			return err
		}},
		{"GetMoveOptions", func(ctx context.Context, c *Client) error {
			_, err := c.GetMoveOptions(ctx, testVideoSourceToken)

			return err
		}},
		{"StopFocus", func(ctx context.Context, c *Client) error {
			return c.StopFocus(ctx, testVideoSourceToken)
		}},
		{"GetImagingStatus", func(ctx context.Context, c *Client) error {
			_, err := c.GetImagingStatus(ctx, testVideoSourceToken)

			return err
		}},
	}
}

// capabilitiesResponse builds a GetCapabilities response advertising only the
// services named in extra, so a test can produce a client that has genuinely
// discovered capabilities yet still has no endpoint for a given service.
func capabilitiesResponse(deviceURL, extra string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
    <soap:Body>
        <tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
            <tds:Capabilities>
                <tt:Device xmlns:tt="http://www.onvif.org/ver10/schema">
                    <tt:XAddr>` + deviceURL + `/onvif/device_service</tt:XAddr>
                </tt:Device>` + extra + `
            </tds:Capabilities>
        </tds:GetCapabilitiesResponse>
    </soap:Body>
</soap:Envelope>`
}

// TestPTZOperationsBeforeInitialize asserts every PTZ operation reports
// ErrNotInitialized - not ErrServiceNotSupported - on a client that has never
// run Initialize. The device in this test does support PTZ; only the client is
// unaware, which is exactly the state the old error misdescribed.
func TestPTZOperationsBeforeInitialize(t *testing.T) {
	client, err := NewClient("http://192.0.2.1/onvif/device_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	for _, op := range ptzOps() {
		t.Run(op.name, func(t *testing.T) {
			err := op.call(ctx, client)
			if !errors.Is(err, ErrNotInitialized) {
				t.Errorf("%s() before Initialize = %v, want ErrNotInitialized", op.name, err)
			}
			// The two must stay distinguishable; a sentinel that satisfied
			// both errors.Is checks would defeat the whole point.
			if errors.Is(err, ErrServiceNotSupported) {
				t.Errorf("%s() must not also report ErrServiceNotSupported", op.name)
			}
		})
	}
}

// TestImagingOperationsBeforeInitialize is TestPTZOperationsBeforeInitialize
// for Imaging, including the three operations that previously reached the
// device endpoint instead of failing.
func TestImagingOperationsBeforeInitialize(t *testing.T) {
	client, err := NewClient("http://192.0.2.1/onvif/device_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	for _, op := range imagingOps() {
		t.Run(op.name, func(t *testing.T) {
			err := op.call(ctx, client)
			if !errors.Is(err, ErrNotInitialized) {
				t.Errorf("%s() before Initialize = %v, want ErrNotInitialized", op.name, err)
			}
			if errors.Is(err, ErrServiceNotSupported) {
				t.Errorf("%s() must not also report ErrServiceNotSupported", op.name)
			}
		})
	}
}

// TestPTZOperationsAfterInitializeWithoutPTZ asserts that once capabilities
// have been fetched, a service the device did not advertise yields
// ErrServiceNotSupported. This is the case ErrServiceNotSupported is actually
// for, and it must survive the #64 change - a caller may reasonably use it to
// decide not to offer PTZ controls at all.
func TestPTZOperationsAfterInitializeWithoutPTZ(t *testing.T) {
	mock := NewMockONVIFServer()
	defer mock.Close()

	// Media only: the device reports no PTZ service.
	mock.SetResponse("GetCapabilities", capabilitiesResponse(mock.URL(),
		`<tt:Media xmlns:tt="http://www.onvif.org/ver10/schema">
                    <tt:XAddr>`+mock.URL()+`/onvif/media_service</tt:XAddr>
                </tt:Media>`))

	client, err := NewClient(mock.URL(), WithCredentials(testUsername, testPassword))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	for _, op := range ptzOps() {
		t.Run(op.name, func(t *testing.T) {
			err := op.call(ctx, client)
			if !errors.Is(err, ErrServiceNotSupported) {
				t.Errorf("%s() after Initialize without PTZ = %v, want ErrServiceNotSupported", op.name, err)
			}
			if errors.Is(err, ErrNotInitialized) {
				t.Errorf("%s() must not report ErrNotInitialized after a successful Initialize", op.name)
			}
		})
	}
}

// TestImagingOperationsAfterInitializeWithoutImaging is the Imaging
// counterpart. The default mock's capabilities advertise Device, Media and PTZ
// but no Imaging, so no override is needed here.
func TestImagingOperationsAfterInitializeWithoutImaging(t *testing.T) {
	mock := NewMockONVIFServer()
	defer mock.Close()

	client, err := NewClient(mock.URL(), WithCredentials(testUsername, testPassword))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	for _, op := range imagingOps() {
		t.Run(op.name, func(t *testing.T) {
			err := op.call(ctx, client)
			if !errors.Is(err, ErrServiceNotSupported) {
				t.Errorf("%s() after Initialize without Imaging = %v, want ErrServiceNotSupported", op.name, err)
			}
			if errors.Is(err, ErrNotInitialized) {
				t.Errorf("%s() must not report ErrNotInitialized after a successful Initialize", op.name)
			}
		})
	}
}

// TestFailedInitializeLeavesClientUninitialized covers the boundary between
// the two errors: Initialize returning an error means capabilities were never
// fetched, so PTZ must still report the recoverable ErrNotInitialized rather
// than claiming the device lacks the service.
func TestFailedInitializeLeavesClientUninitialized(t *testing.T) {
	mock := NewMockONVIFServer()
	defer mock.Close()

	// A SOAP Fault for GetCapabilities: reachable device, failed discovery.
	mock.SetResponse("GetCapabilities", mock.responses["default"])

	client, err := NewClient(mock.URL(), WithCredentials(testUsername, testPassword))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	if err := client.Initialize(ctx); err == nil {
		t.Fatal("Initialize() succeeded against a faulting GetCapabilities, want error")
	}

	if err := client.SetHomePosition(ctx, testProfileToken); !errors.Is(err, ErrNotInitialized) {
		t.Errorf("SetHomePosition() after a failed Initialize = %v, want ErrNotInitialized", err)
	}
}

// TestMediaAndEventsStillFallBackToDeviceEndpoint pins the asymmetry that #64
// deliberately left in place and documented on Initialize, rather than
// silently relying on it: Media and Events degrade to the device endpoint, so
// they keep working without Initialize on the many cameras that serve those
// operations there. Only PTZ and Imaging hard-require Initialize.
//
// Without this test, "make everything consistent" is an easy future change to
// make by accident.
func TestMediaAndEventsStillFallBackToDeviceEndpoint(t *testing.T) {
	mock := NewMockONVIFServer()
	defer mock.Close()

	client, err := NewClient(mock.URL(), WithCredentials(testUsername, testPassword))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Compared against Endpoint() rather than mock.URL(): NewClient
	// normalizes a bare URL by appending /onvif/device_service.
	deviceEndpoint := client.Endpoint()

	// No Initialize call: mediaEndpoint and eventEndpoint are both empty.
	if got := client.getMediaEndpoint(); got != deviceEndpoint {
		t.Errorf("getMediaEndpoint() without Initialize = %q, want the device endpoint %q", got, deviceEndpoint)
	}
	if got := client.getEventEndpoint(); got != deviceEndpoint {
		t.Errorf("getEventEndpoint() without Initialize = %q, want the device endpoint %q", got, deviceEndpoint)
	}

	// And the fallback is load-bearing, not just cosmetic: GetProfiles
	// actually completes against the device endpoint.
	if _, err := client.GetProfiles(context.Background()); err != nil {
		t.Errorf("GetProfiles() without Initialize failed: %v", err)
	}
}
