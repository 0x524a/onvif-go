package main

import (
	"bytes"
	"errors"
	"go/ast"
	"go/token"
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
		// The discovery and server packages have their own sections.
		"<h2>Discovery package</h2>", `id="discovery.Discover"`, `id="discovery.Responder.Start"`,
		"<h2>Server package</h2>", `id="server.New"`, `id="server.Config"`,
		`id="server.Server.Start"`, `id="server.Server.EndpointReference"`,
		"https://pkg.go.dev/github.com/0x524a/onvif-go/server#Server.Start",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("page is missing %q", want)
		}
	}

	// SOAP wire types and handlers are plumbing, not API.
	for _, unwanted := range []string{`id="server.GetProfilesResponse"`, `id="server.Server.HandleGetProfiles"`} {
		if strings.Contains(out, unwanted) {
			t.Errorf("page should not contain %q", unwanted)
		}
	}

	if n := strings.Count(out, `class="m"`); n < 250 {
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

func TestRenderSubpackages(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"client.go": "package onvif\n\n// Client talks to a camera.\ntype Client struct{}\n",
	})

	sub := filepath.Join(dir, "discovery")
	if err := os.Mkdir(sub, 0o750); err != nil {
		t.Fatal(err)
	}

	src := `package discovery

// ErrNone says nothing was found.
var ErrNone = errors.New("none")

// Find looks for devices.
func Find() {}

// Device is a thing that answers.
type Device struct{ Name string }

// Name returns the name.
func (d *Device) Label() string { return d.Name }

// NewDevice builds a Device.
func NewDevice() *Device { return nil }
`
	if err := os.WriteFile(filepath.Join(sub, "d.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	out := renderDir(t, dir)

	for _, want := range []string{
		"<h2>Discovery package</h2>", `id="discovery.ErrNone"`, `id="discovery.Find"`,
		`id="discovery.Device"`, `id="discovery.Device.Label"`, `id="discovery.NewDevice"`,
		"type Device struct", "Find looks for devices.",
		"https://pkg.go.dev/github.com/0x524a/onvif-go/discovery#Device.Label",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("page is missing %q", want)
		}
	}

	// The server directory does not exist, so it gets no section.
	if strings.Contains(out, "Server package") {
		t.Error("a missing subpackage should have no section")
	}
}

func TestSubpackageErrors(t *testing.T) {
	dir := writeFiles(t, map[string]string{"client.go": "package onvif\n"})

	sub := filepath.Join(dir, "server")
	if err := os.Mkdir(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "bad.go"), []byte("package server\n\nfunc (\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := render(dir, &bytes.Buffer{}); err == nil {
		t.Error("expected an error for an unparsable subpackage")
	}
}

func TestNodeTextRejectsOtherNodes(t *testing.T) {
	if _, err := nodeText(token.NewFileSet(), &ast.Ident{Name: "x"}); !errors.Is(err, errUnsupportedNode) {
		t.Errorf("error = %v, want errUnsupportedNode", err)
	}
}

func TestInternalFilter(t *testing.T) {
	b := &sectionBuilder{fset: token.NewFileSet(), files: []string{"keep.go"}, also: []string{"T.M"}}
	b.fset.AddFile("other.go", -1, 10)
	n := &ast.Ident{NamePos: token.Pos(1)}

	if !b.internal("X", n) {
		t.Error("a declaration outside the listed files should be internal")
	}
	if b.internal("T.M", n) {
		t.Error("a name in Also should be public")
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
