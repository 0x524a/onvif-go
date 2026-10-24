package main

import (
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server"
)

// The simulator must not ship a default password: an unset flag has to be
// an error (empty credentials would leave every action open), the env var
// is the fallback, and an explicit flag always wins.
func TestResolvePassword(t *testing.T) {
	t.Setenv("ONVIF_SERVER_PASSWORD", "")
	if _, err := resolvePassword(""); err == nil {
		t.Fatal("expected error for empty flag and empty env, got nil")
	}

	t.Setenv("ONVIF_SERVER_PASSWORD", "from-env")
	got, err := resolvePassword("")
	if err != nil || got != "from-env" {
		t.Fatalf("flag empty, env set: got %q, err %v; want from-env", got, err)
	}

	got, err = resolvePassword("from-flag")
	if err != nil || got != "from-flag" {
		t.Fatalf("flag set: got %q, err %v; want from-flag (flag wins over env)", got, err)
	}
}

func TestBuildConfigProfileMatrix(t *testing.T) {
	for _, n := range []int{1, 3, 10} {
		config := buildConfig("127.0.0.1", 0, "admin", "pw", "mfr", "model", "1.0", "SN", n, true, true, false)

		if len(config.Profiles) != n {
			t.Fatalf("n=%d: profile count = %d", n, len(config.Profiles))
		}

		if _, err := server.New(config); err != nil {
			t.Fatalf("n=%d: server.New rejected the built config: %v", n, err)
		}

		if config.Profiles[0].Token != "profile_0" {
			t.Errorf("n=%d: first token = %q", n, config.Profiles[0].Token)
		}
	}

	// PTZ flag off: no profile carries a PTZ config even when the
	// template supports it.
	config := buildConfig("127.0.0.1", 0, "admin", "pw", "mfr", "model", "1.0", "SN", 3, false, true, false)
	for i, p := range config.Profiles {
		if p.PTZ != nil {
			t.Errorf("profile %d carries PTZ despite -ptz=false", i)
		}
	}

	// Template rotation: with PTZ on, profile 0 (Main, hasPTZ) gets PTZ
	// with two presets; profile 1 (Wide Angle, no PTZ) gets none.
	config = buildConfig("127.0.0.1", 0, "admin", "pw", "mfr", "model", "1.0", "SN", 3, true, true, false)
	if config.Profiles[0].PTZ == nil || len(config.Profiles[0].PTZ.Presets) != 2 {
		t.Errorf("profile 0 PTZ = %+v, want 2 presets", config.Profiles[0].PTZ)
	}

	if config.Profiles[1].PTZ != nil {
		t.Errorf("profile 1 (no-PTZ template) carries PTZ: %+v", config.Profiles[1].PTZ)
	}

	eventsConfig := buildConfig("127.0.0.1", 0, "admin", "pw", "mfr", "model", "1.0", "SN", 1, true, true, true)
	if !eventsConfig.SupportEvents {
		t.Error("events flag not propagated")
	}
}

func TestPrintBannerSmoke(t *testing.T) {
	printBanner() // must not panic; output is decorative
}
