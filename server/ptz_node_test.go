package server

import (
	"context"
	"io"
	"testing"

	onvif "github.com/0x524a/onvif-go"
)

// ptzNodeClient starts the virtual camera from config and returns an
// initialized onvif-go client for it: the real client against the real server.
func ptzNodeClient(t *testing.T, config *Config) *onvif.Client {
	t.Helper()

	config.Host = testLoopbackHost
	config.Output = io.Discard
	_, addr := startTestServer(t, config)

	client, err := onvif.NewClient("http://"+addr+config.BasePath+"/device_service",
		onvif.WithCredentials(config.Username, config.Password))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	return client
}

func wantRange(t *testing.T, what string, r *onvif.FloatRange, lo, hi float64) {
	t.Helper()

	if r == nil || r.Min != lo || r.Max != hi {
		t.Errorf("%s = %+v, want {%g %g}", what, r, lo, hi)
	}
}

// assertGenericSpaces checks that s reports exactly one generic space of every
// kind, with the ranges the ONVIF generic spaces are defined over.
func assertGenericSpaces(t *testing.T, s *onvif.PTZSpaces) {
	t.Helper()

	if s == nil {
		t.Fatal("spaces = nil")
	}
	two := []struct {
		what, uri string
		got       []*onvif.Space2DDescription
	}{
		{"absolute pan/tilt", genericPanTiltPositionSpace, s.AbsolutePanTiltPositionSpace},
		{"relative pan/tilt", genericPanTiltTranslationSpace, s.RelativePanTiltTranslationSpace},
		{"continuous pan/tilt", genericPanTiltVelocitySpace, s.ContinuousPanTiltVelocitySpace},
	}
	for _, c := range two {
		if len(c.got) != 1 || c.got[0].URI != c.uri {
			t.Errorf("%s spaces = %+v, want one %s", c.what, c.got, c.uri)

			continue
		}
		wantRange(t, c.what+" x", c.got[0].XRange, -1, 1)
		wantRange(t, c.what+" y", c.got[0].YRange, -1, 1)
	}
	one := []struct {
		what, uri string
		lo        float64
		got       []*onvif.Space1DDescription
	}{
		{"absolute zoom", genericZoomPositionSpace, 0, s.AbsoluteZoomPositionSpace},
		{"relative zoom", genericZoomTranslationSpace, -1, s.RelativeZoomTranslationSpace},
		{"continuous zoom", genericZoomVelocitySpace, -1, s.ContinuousZoomVelocitySpace},
		{"pan/tilt speed", genericPanTiltSpeedSpace, 0, s.PanTiltSpeedSpace},
		{"zoom speed", genericZoomSpeedSpace, 0, s.ZoomSpeedSpace},
	}
	for _, c := range one {
		if len(c.got) != 1 || c.got[0].URI != c.uri {
			t.Errorf("%s spaces = %+v, want one %s", c.what, c.got, c.uri)

			continue
		}
		wantRange(t, c.what, c.got[0].XRange, c.lo, 1)
	}
}

func TestClientGetNodesAgainstServer(t *testing.T) {
	config := DefaultConfig()
	client := ptzNodeClient(t, config)

	nodes, err := client.GetNodes(context.Background())
	if err != nil {
		t.Fatalf("GetNodes() error = %v", err)
	}
	want := map[string]int{}
	for _, p := range config.Profiles {
		if p.PTZ != nil {
			want[p.PTZ.NodeToken] = len(p.PTZ.Presets)
		}
	}
	if len(nodes) != len(want) {
		t.Fatalf("GetNodes() returned %d nodes, want one per distinct node token: %v", len(nodes), want)
	}
	for _, n := range nodes {
		presets, ok := want[n.Token]
		if !ok {
			t.Errorf("unexpected node %q", n.Token)

			continue
		}
		if n.MaximumNumberOfPresets != presets || n.HomeSupported {
			t.Errorf("node %q: MaximumNumberOfPresets=%d HomeSupported=%v, want %d and false (no home support)",
				n.Token, n.MaximumNumberOfPresets, n.HomeSupported, presets)
		}
		assertGenericSpaces(t, n.SupportedPTZSpaces)
	}
}

func TestClientGetNodeAgainstServer(t *testing.T) {
	config := DefaultConfig()
	client := ptzNodeClient(t, config)
	token := config.Profiles[0].PTZ.NodeToken

	node, err := client.GetNode(context.Background(), token)
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}
	if node.Token != token || node.Name == "" {
		t.Errorf("node = %+v, want token %q and a name", node, token)
	}
	assertGenericSpaces(t, node.SupportedPTZSpaces)

	if _, err := client.GetNode(context.Background(), "no-such-node"); err == nil {
		t.Error("GetNode() of an unknown token = nil error, want a fault")
	}
}

// The configuration token GetProfiles reports is the one the PTZ
// configuration reads accept, so a client can go from a profile to its limits.
func TestClientPTZConfigurationReadsAgainstServer(t *testing.T) {
	config := DefaultConfig()
	client := ptzNodeClient(t, config)
	ctx := context.Background()

	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		t.Fatalf("GetProfiles() error = %v", err)
	}
	var cfgToken string
	for _, p := range profiles {
		if p.PTZConfiguration != nil {
			cfgToken = p.PTZConfiguration.Token

			break
		}
	}
	if cfgToken == "" {
		t.Fatal("no profile carries a PTZ configuration")
	}

	opts, err := client.GetConfigurationOptions(ctx, cfgToken)
	if err != nil {
		t.Fatalf("GetConfigurationOptions() error = %v", err)
	}
	assertGenericSpaces(t, opts.Spaces)
	if opts.PTZTimeout != nil {
		t.Errorf("PTZTimeout = %+v, want none: the virtual camera ignores the continuous move Timeout", opts.PTZTimeout)
	}

	cfg, err := client.GetConfiguration(ctx, cfgToken)
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if cfg.Token != cfgToken || cfg.NodeToken == "" || cfg.PanTiltLimits != nil || cfg.ZoomLimits != nil {
		t.Errorf("configuration = %+v, want token %q, a node token and no configured limits", cfg, cfgToken)
	}

	cfgs, err := client.GetConfigurations(ctx)
	if err != nil || len(cfgs) == 0 {
		t.Fatalf("GetConfigurations() = %d, %v, want at least one", len(cfgs), err)
	}

	if _, err := client.GetConfigurationOptions(ctx, "no-such-configuration"); err == nil {
		t.Error("GetConfigurationOptions() of an unknown token = nil error, want a fault")
	}
}

// A node reports only the move kinds its profile supports.
func TestServerNodeOmitsUnsupportedMoveSpaces(t *testing.T) {
	config := DefaultConfig()
	config.Profiles[0].PTZ.SupportsAbsolute = false
	config.Profiles[0].PTZ.SupportsContinuous = false
	client := ptzNodeClient(t, config)

	node, err := client.GetNode(context.Background(), config.Profiles[0].PTZ.NodeToken)
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}
	s := node.SupportedPTZSpaces
	if s == nil || s.AbsolutePanTiltPositionSpace != nil || s.AbsoluteZoomPositionSpace != nil ||
		s.ContinuousPanTiltVelocitySpace != nil || s.ContinuousZoomVelocitySpace != nil {
		t.Errorf("spaces = %+v, want no absolute or continuous spaces", s)
	}
	if len(s.RelativePanTiltTranslationSpace) != 1 || len(s.PanTiltSpeedSpace) != 1 {
		t.Errorf("spaces = %+v, want the relative and speed spaces kept", s)
	}
}
