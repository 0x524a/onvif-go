package server

import (
	"bytes"
	"context"
	"image/jpeg"
	"io"
	"net/http"
	"testing"
	"time"
)

func httpGet(t *testing.T, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}

	return resp
}

func snapshotConfig(t *testing.T) *Config {
	t.Helper()

	cfg := DefaultConfig()
	cfg.Host = testLoopbackHost
	cfg.Output = io.Discard
	cfg.Username, cfg.Password = "", ""

	return cfg
}

func TestSnapshotServesDecodableJPEG(t *testing.T) {
	cfg := snapshotConfig(t)
	cfg.Profiles[1].Snapshot.Resolution = Resolution{Width: 320, Height: 180}

	_, addr := startTestServer(t, cfg)

	resp := httpGet(t, "http://"+addr+"/onvif/snapshot?profile="+cfg.Profiles[1].Token)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", ct)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if len(body) == 0 {
		t.Fatal("snapshot body is empty")
	}

	img, err := jpeg.Decode(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("body is not a valid JPEG: %v", err)
	}

	if b := img.Bounds(); b.Dx() != 320 || b.Dy() != 180 {
		t.Errorf("image is %dx%d, want 320x180", b.Dx(), b.Dy())
	}
}

func TestSnapshotErrors(t *testing.T) {
	cfg := snapshotConfig(t)
	cfg.Profiles[2].Snapshot.Enabled = false

	_, addr := startTestServer(t, cfg)

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"missing profile parameter", "", http.StatusBadRequest},
		{"unknown profile", "?profile=nope", http.StatusNotFound},
		{"snapshots disabled", "?profile=" + cfg.Profiles[2].Token, http.StatusNotImplemented},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := httpGet(t, "http://"+addr+"/onvif/snapshot"+tt.query)
			_ = resp.Body.Close()

			if resp.StatusCode != tt.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestRenderSnapshotChangesWithTime(t *testing.T) {
	profile := &ProfileConfig{Snapshot: SnapshotConfig{
		Enabled: true, Resolution: Resolution{Width: 200, Height: 120}, Quality: 90,
	}}

	base := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

	first, err := renderSnapshot(profile, base)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	again, _ := renderSnapshot(profile, base)
	if !bytes.Equal(first, again) {
		t.Error("same profile and time rendered different images")
	}

	later, _ := renderSnapshot(profile, base.Add(30*time.Second))
	if bytes.Equal(first, later) {
		t.Error("snapshots 30s apart are identical, so they cannot show the camera is live")
	}
}

func TestSnapshotSizeFallbacks(t *testing.T) {
	tests := []struct {
		name         string
		profile      ProfileConfig
		wantW, wantH int
	}{
		{"snapshot resolution", ProfileConfig{Snapshot: SnapshotConfig{Resolution: Resolution{Width: 100, Height: 50}}}, 100, 50},
		{"video source resolution", ProfileConfig{VideoSource: VideoSourceConfig{Resolution: Resolution{Width: 800, Height: 600}}}, 800, 600},
		{"built-in default", ProfileConfig{}, fallbackSnapshotWidth, fallbackSnapshotHeight},
		{"clamped", ProfileConfig{Snapshot: SnapshotConfig{Resolution: Resolution{Width: 100000, Height: 100000}}}, maxSnapshotDimension, maxSnapshotDimension},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := snapshotSize(&tt.profile)
			if w != tt.wantW || h != tt.wantH {
				t.Errorf("snapshotSize = %dx%d, want %dx%d", w, h, tt.wantW, tt.wantH)
			}
		})
	}
}
