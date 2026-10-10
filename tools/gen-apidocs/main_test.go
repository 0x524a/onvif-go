package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	return dir
}

func renderDir(t *testing.T, dir string) string {
	t.Helper()

	var buf bytes.Buffer
	if err := render(dir, &buf); err != nil {
		t.Fatalf("render() error = %v", err)
	}

	return buf.String()
}

// The real package is the best fixture: the page must list what the client
// actually offers, with the doc comments from the source.
func TestRenderRepository(t *testing.T) {
	out := renderDir(t, "../..")

	for _, want := range []string{
		"<title>API reference | onvif-go</title>",
		`id="GetProfiles"`,
		"GetProfiles retrieves all media profiles.",
		`id="ContinuousMove"`,
		`id="NewClient"`,
		"<h2>PTZ</h2>",
		"<h2>Wi-Fi and 802.1X</h2>",
		"https://pkg.go.dev/github.com/0x524a/onvif-go#Client.GetProfiles",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("page is missing %q", want)
		}
	}

	if n := strings.Count(out, `class="m"`); n < 200 {
		t.Errorf("page lists %d entries, want at least 200", n)
	}
}

func TestRenderGroupsAndFilters(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"client.go": `package onvif

// Client talks to a camera.
type Client struct{}

// NewClient builds a Client.
func NewClient(endpoint string) *Client { return &Client{} }
`,
		"ptz.go": `package onvif

// Stop halts movement.
func (c *Client) Stop() error { return nil }

// ContinuousMove starts moving.
func (c *Client) ContinuousMove(speed float64) error { return nil }

// hidden is not exported.
func (c *Client) hidden() {}
`,
		// A file with no section of its own falls into the Client section.
		"zz_extra.go": `package onvif

// Extra is declared in a file with no section.
func (c *Client) Extra() {}
`,
		// Test files and other packages are ignored.
		"ptz_test.go": "package onvif\n\n// FromTest must not appear.\nfunc (c *Client) FromTest() {}\n",
		"other.go":    "package other\n\n// Other must not appear.\nfunc Other() {}\n",
	})

	out := renderDir(t, dir)

	for _, want := range []string{
		`id="NewClient"`, `id="Stop"`, `id="ContinuousMove"`, `id="Extra"`,
		"func (c *Client) Stop() error", // signature without body
		"Stop halts movement.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("page is missing %q", want)
		}
	}

	for _, unwanted := range []string{"(c *Client) hidden", "FromTest", "Other must not appear", "{ return nil }"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("page should not contain %q", unwanted)
		}
	}

	// Entries are sorted by name within a section.
	if strings.Index(out, `id="ContinuousMove"`) > strings.Index(out, `id="Stop"`) {
		t.Error("PTZ entries are not sorted by name")
	}

	// Extra lands in the Client section, which precedes PTZ.
	if strings.Index(out, `id="Extra"`) > strings.Index(out, "<h2>PTZ</h2>") {
		t.Error("an entry from an unlisted file should appear under Client, before PTZ")
	}
}

func TestRenderErrors(t *testing.T) {
	t.Run("no package", func(t *testing.T) {
		err := render(t.TempDir(), &bytes.Buffer{})
		if !errors.Is(err, errNoPackage) {
			t.Errorf("error = %v, want errNoPackage", err)
		}
	})

	t.Run("missing directory", func(t *testing.T) {
		if err := render(filepath.Join(t.TempDir(), "nope"), &bytes.Buffer{}); err == nil {
			t.Error("expected an error for a missing directory")
		}
	})

	t.Run("syntax error", func(t *testing.T) {
		dir := writeFiles(t, map[string]string{"bad.go": "package onvif\n\nfunc (\n"})
		if err := render(dir, &bytes.Buffer{}); err == nil {
			t.Error("expected an error for unparsable source")
		}
	})

	t.Run("run uses the given directory", func(t *testing.T) {
		if err := run(t.TempDir()); !errors.Is(err, errNoPackage) {
			t.Errorf("run() error = %v, want errNoPackage", err)
		}
	})
}

func TestAnchorFor(t *testing.T) {
	tests := map[string]string{
		"PTZ":                "ptz",
		"Device: additional": "device-additional",
		"Wi-Fi and 802.1X":   "wi-fi-and-802.1x",
		"Device I/O":         "device-i-o",
	}

	for in, want := range tests {
		if got := anchorFor(in); got != want {
			t.Errorf("anchorFor(%q) = %q, want %q", in, got, want)
		}
	}
}
