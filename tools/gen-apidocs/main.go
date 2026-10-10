// Command gen-apidocs writes the API reference page for the project site.
//
// It reads the exported API of the client (onvif.Client) and of the discovery
// and server packages straight from the source, so the page lists exactly what
// the code offers and cannot drift from it. Run from the repository root:
//
//	go run ./tools/gen-apidocs > site/api.html
package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const (
	importPath = "github.com/0x524a/onvif-go"
	pkgDocsURL = "https://pkg.go.dev/" + importPath
)

// group maps source files to the section their methods appear under. Files not
// listed here fall into the "Client" section.
var groups = []struct {
	Files []string
	Title string
	Blurb string
}{
	{[]string{"client.go"}, "Client", "Create a client and discover the service endpoints."},
	{[]string{"device.go"}, "Device", "Information, capabilities, hostname, DNS, NTP, network and users."},
	{[]string{"device_additional.go"}, "Device: additional", "Geolocation, access policy and discovery addresses."},
	{[]string{"device_extended.go"}, "Device: extended", "Further device configuration."},
	{[]string{"device_security.go"}, "Device: security", "IP filters, password policy and authentication settings."},
	{[]string{"device_certificates.go"}, "Certificates", "Install, create and manage device certificates."},
	{[]string{"device_wifi.go"}, "Wi-Fi and 802.1X", "Wireless status, scans and 802.1X configuration."},
	{[]string{"device_storage.go"}, "Storage", "Storage configurations."},
	{[]string{"media.go", "media_video.go", "media_audio.go", "media_metadata.go", "media_osd.go"}, "Media", "Profiles, stream and snapshot URIs, and encoder, source and metadata configuration."},
	{[]string{"ptz.go"}, "PTZ", "Pan, tilt and zoom movement, presets and status."},
	{[]string{"imaging.go"}, "Imaging", "Image settings and focus control."},
	{[]string{"event.go"}, "Events", "Subscriptions and pull points."},
	{[]string{"deviceio.go"}, "Device I/O", "Relays, digital inputs and video outputs."},
}

// subPackage is a package whose whole exported API is listed in one section.
type subPackage struct {
	Dir   string
	Name  string
	Title string
	Blurb string
	Files []string // source files that hold the public API; empty means all
	Also  []string // declarations from other files that are public API too
}

var subPackages = []subPackage{
	{"discovery", "discovery", "Discovery package", "Find cameras with WS-Discovery, or answer probes for your own devices.", nil, nil},
	{"server", "server", "Server package", "A virtual ONVIF camera: Device, Media, PTZ and Imaging services, with stable identities.",
		[]string{"types.go", "server.go", "errors.go", "snapshot.go", "discovery.go"},
		[]string{"Server.EndpointReference"}},
}

type entry struct {
	ID        string
	Name      string
	Signature string
	Doc       string
	URL       string
}

type section struct {
	Anchor  string
	Title   string
	Blurb   string
	Entries []entry
}

type page struct {
	Sections []section
	Total    int
}

func main() {
	if err := run("."); err != nil {
		fmt.Fprintln(os.Stderr, "gen-apidocs:", err)
		os.Exit(1)
	}
}

var errUnsupportedNode = errors.New("unsupported declaration")

var errNoPackage = errors.New("package onvif not found; run from the repository root")

// run writes the API reference for the package in dir to standard output.
func run(dir string) error {
	return render(dir, os.Stdout)
}

func render(dir string, w io.Writer) error {
	fset := token.NewFileSet()

	files, err := parseSources(fset, dir, "onvif")
	if err != nil {
		return err
	}

	docPkg, err := doc.NewFromFiles(fset, files, importPath)
	if err != nil {
		return fmt.Errorf("build package docs: %w", err)
	}

	byFile, err := collect(fset, docPkg)
	if err != nil {
		return err
	}

	pg := buildPage(byFile)

	for i := range subPackages {
		sub := &subPackages[i]
		sec, err := packageSection(fset, dir, sub)
		if err != nil {
			return err
		}

		if len(sec.Entries) > 0 {
			pg.Sections = append(pg.Sections, sec)
			pg.Total += len(sec.Entries)
		}
	}

	if err := tmpl.Execute(w, pg); err != nil {
		return fmt.Errorf("render page: %w", err)
	}

	return nil
}

// parseSources parses the non-test Go files of the package called name in dir.
func parseSources(fset *token.FileSet, dir, name string) ([]*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var files []*ast.File
	for _, e := range entries {
		file := e.Name()
		if e.IsDir() || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, filepath.Join(dir, file), nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
		if f.Name.Name == name {
			files = append(files, f)
		}
	}

	if len(files) == 0 {
		return nil, errNoPackage
	}

	return files, nil
}

// collect gathers the exported Client constructors and methods, keyed by the
// source file that declares them.
func collect(fset *token.FileSet, docPkg *doc.Package) (map[string][]entry, error) {
	byFile := map[string][]entry{}

	add := func(f *doc.Func, recv string) error {
		if f.Decl == nil || !ast.IsExported(f.Name) {
			return nil
		}

		sig, err := signature(fset, f.Decl)
		if err != nil {
			return err
		}

		anchor := f.Name
		if recv != "" {
			anchor = recv + "." + f.Name
		}

		file := filepath.Base(fset.Position(f.Decl.Pos()).Filename)
		byFile[file] = append(byFile[file], entry{
			ID:        f.Name,
			Name:      f.Name,
			Signature: sig,
			Doc:       strings.TrimSpace(f.Doc),
			URL:       pkgDocsURL + "#" + anchor,
		})

		return nil
	}

	for _, t := range docPkg.Types {
		if t.Name != "Client" {
			continue
		}

		for _, f := range t.Funcs { // constructors such as NewClient
			if err := add(f, ""); err != nil {
				return nil, err
			}
		}
		for _, m := range t.Methods {
			if err := add(m, "Client"); err != nil {
				return nil, err
			}
		}
	}

	return byFile, nil
}

// packageSection lists the exported constants, variables, functions, types and
// methods of a subpackage, in source order within each kind. A directory that
// is missing or holds no such package yields an empty section.
func packageSection(fset *token.FileSet, root string, sub *subPackage) (section, error) {
	sec := section{Anchor: anchorFor(sub.Title), Title: sub.Title, Blurb: sub.Blurb}

	files, err := parseSources(fset, filepath.Join(root, sub.Dir), sub.Name)
	if errors.Is(err, errNoPackage) || errors.Is(err, fs.ErrNotExist) {
		return sec, nil
	}
	if err != nil {
		return sec, err
	}

	docPkg, err := doc.NewFromFiles(fset, files, importPath+"/"+sub.Dir)
	if err != nil {
		return sec, fmt.Errorf("build %s docs: %w", sub.Name, err)
	}

	b := &sectionBuilder{fset: fset, pkg: sub.Name, files: sub.Files, also: sub.Also}

	b.values(docPkg.Consts)
	b.values(docPkg.Vars)
	b.funcs(docPkg.Funcs, "")

	for _, t := range docPkg.Types {
		b.node(t.Name, t.Name, t.Decl, t.Doc)
		b.values(t.Consts)
		b.values(t.Vars)
		b.funcs(t.Funcs, "")
		b.funcs(t.Methods, t.Name+".")
	}

	if b.err != nil {
		return sec, b.err
	}

	sec.Entries = b.entries

	return sec, nil
}

// sectionBuilder accumulates the entries of one package, remembering the first
// error so callers check it once.
type sectionBuilder struct {
	fset    *token.FileSet
	pkg     string
	files   []string
	also    []string
	entries []entry
	err     error
}

func (b *sectionBuilder) node(name, anchor string, n ast.Node, docText string) {
	if b.err != nil || b.internal(name, n) {
		return
	}

	text, err := nodeText(b.fset, n)
	if err != nil {
		b.err = err

		return
	}

	b.entries = append(b.entries, entry{
		ID:        b.pkg + "." + name,
		Name:      name,
		Signature: text,
		Doc:       strings.TrimSpace(docText),
		URL:       pkgDocsURL + "/" + b.pkg + "#" + anchor,
	})
}

// internal reports whether a declaration lies outside the files that hold the
// package's public API. The server keeps its SOAP wire types and the handlers
// that serve them in per-service files, which are plumbing rather than API.
func (b *sectionBuilder) internal(name string, n ast.Node) bool {
	if len(b.files) == 0 || slices.Contains(b.also, name) {
		return false
	}

	file := filepath.Base(b.fset.Position(n.Pos()).Filename)

	return !slices.Contains(b.files, file)
}

func (b *sectionBuilder) values(vs []*doc.Value) {
	for _, v := range vs {
		if len(v.Names) > 0 {
			b.node(v.Names[0], v.Names[0], v.Decl, v.Doc)
		}
	}
}

// funcs adds functions, or methods when prefix is the receiver name and a dot.
func (b *sectionBuilder) funcs(list []*doc.Func, prefix string) {
	for _, f := range list {
		if ast.IsExported(f.Name) {
			b.node(prefix+f.Name, prefix+f.Name, f.Decl, f.Doc)
		}
	}
}

// nodeText prints a declaration without its body or doc comment.
func nodeText(fset *token.FileSet, n ast.Node) (string, error) {
	switch d := n.(type) {
	case *ast.FuncDecl:
		return signature(fset, d)
	case *ast.GenDecl:
		c := *d
		c.Doc = nil

		var buf bytes.Buffer
		if err := printer.Fprint(&buf, fset, &c); err != nil {
			return "", fmt.Errorf("print declaration: %w", err)
		}

		return buf.String(), nil
	}

	return "", errUnsupportedNode
}

// buildPage orders the entries into the sections listed in groups.
func buildPage(byFile map[string][]entry) page {
	known := map[string]bool{}
	p := page{}

	for _, g := range groups {
		var es []entry
		for _, f := range g.Files {
			known[f] = true
			es = append(es, byFile[f]...)
		}

		if g.Files[0] == "client.go" {
			es = append(es, otherFiles(byFile, known)...)
		}
		if len(es) == 0 {
			continue
		}

		sort.Slice(es, func(i, j int) bool { return es[i].Name < es[j].Name })
		p.Sections = append(p.Sections, section{
			Anchor:  anchorFor(g.Title),
			Title:   g.Title,
			Blurb:   g.Blurb,
			Entries: es,
		})
		p.Total += len(es)
	}

	return p
}

// otherFiles returns entries from source files with no section of their own.
// known must already list every file that has one.
func otherFiles(byFile map[string][]entry, known map[string]bool) []entry {
	for _, g := range groups {
		for _, f := range g.Files {
			known[f] = true
		}
	}

	var files []string
	for f := range byFile {
		if !known[f] {
			files = append(files, f)
		}
	}
	sort.Strings(files)

	var out []entry
	for _, f := range files {
		out = append(out, byFile[f]...)
	}

	return out
}

func anchorFor(title string) string {
	r := strings.NewReplacer(" ", "-", ":", "", "/", "-")

	return strings.ToLower(r.Replace(title))
}

// signature prints a function declaration without its body or doc comment.
func signature(fset *token.FileSet, decl *ast.FuncDecl) (string, error) {
	d := *decl
	d.Body = nil
	d.Doc = nil

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, &d); err != nil {
		return "", fmt.Errorf("print %s: %w", decl.Name.Name, err)
	}

	return buf.String(), nil
}

var tmpl = template.Must(template.New("api").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>API reference | onvif-go</title>
<meta name="description" content="Every exported method of the onvif-go client, plus the discovery and server packages, generated from the source.">
<link rel="icon" href="icon.svg" type="image/svg+xml">
<link rel="apple-touch-icon" href="icon-180.png">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,400;12..96,500;12..96,700&family=Martian+Mono:wght@400;500&display=swap">
<link rel="stylesheet" href="style.css">
<style>
.api { display: grid; grid-template-columns: minmax(0, 1fr); gap: 32px; padding: 32px 0 80px; }
@media (min-width: 900px) { .api { grid-template-columns: 220px minmax(0, 1fr); gap: 56px; } }
.api h1 { font-size: clamp(2rem, 4.5vw, 3rem); margin: 0 0 12px; }
.side { font-weight: 500; }
@media (min-width: 900px) { .side { position: sticky; top: 24px; align-self: start; max-height: calc(100vh - 48px); overflow-y: auto; } }
.side ul { list-style: none; margin: 0; padding: 0; display: flex; flex-wrap: wrap; gap: 4px 16px; }
@media (min-width: 900px) { .side ul { display: block; } }
.side li a { display: flex; justify-content: space-between; gap: 12px; padding: 5px 0; text-decoration: none; color: var(--slate); }
.side li a:hover { color: var(--ink); }
.side .n { font-family: var(--mono); font-size: 0.75rem; }
.find { width: 100%; max-width: 440px; font: 400 1rem var(--display); color: var(--ink); background: var(--paper-2); border: 1.5px solid var(--line); border-radius: 8px; padding: 11px 14px; margin: 8px 0 28px; }
.find:focus-visible { border-color: var(--lens-bright); }
.grp { padding: 0; border: 0; margin: 0 0 56px; scroll-margin-top: 24px; }
.grp h2 { font-size: 1.75rem; margin: 0 0 4px; }
.grp > p { color: var(--slate); margin: 0 0 12px; }
.m { overflow-wrap: anywhere; padding: 20px 0; border-top: 1px solid var(--line); scroll-margin-top: 24px; }
.m h3 { font: 700 1.0625rem var(--display); margin: 0 0 8px; overflow-wrap: anywhere; }
.m h3 a { text-decoration: none; }
.m h3 a:hover { text-decoration: underline; text-underline-offset: 4px; }
.m pre { margin: 0 0 10px; padding: 12px 14px; background: var(--ink-2); color: var(--code-fg); border-radius: 8px; overflow-x: auto; font: 400 0.8125rem/1.6 var(--mono); }
.m p { margin: 0; color: var(--slate); max-width: 70ch; white-space: pre-line; }
.m[hidden], .grp[hidden], .side li[hidden] { display: none; }
.none { color: var(--slate); }
</style>
</head>
<body>
<a class="skip" href="#main">Skip to content</a>
<header>
  <div class="wrap">
    <a class="brand" href="./"><img src="icon.svg" alt="" width="36" height="36">onvif-go</a>
    <nav aria-label="Primary">
      <a href="./">Home</a>
      <a href="https://pkg.go.dev/github.com/0x524a/onvif-go">pkg.go.dev</a>
      <a class="gh" href="https://github.com/0x524a/onvif-go">GitHub</a>
    </nav>
  </div>
</header>
<main id="main" class="wrap">
  <div class="api">
    <nav class="side" aria-label="Sections">
      <ul>
      {{- range .Sections}}
        <li><a href="#{{.Anchor}}">{{.Title}} <span class="n">{{len .Entries}}</span></a></li>
      {{- end}}
      </ul>
    </nav>
    <div>
      <h1>API reference</h1>
      <p class="sub">{{.Total}} entries across the client, the discovery package and the server package, generated from the source on every deploy. Client methods take a <code>context.Context</code> first.</p>
      <label for="find" class="sub" style="margin:0">Filter by name or description</label><br>
      <input id="find" class="find" type="search" placeholder="GetProfiles, preset, Responder, Config" autocomplete="off">
      <p id="none" class="none" hidden>No method matches that. Try a shorter word.</p>
      {{range .Sections}}
      <section class="grp" id="{{.Anchor}}">
        <h2>{{.Title}}</h2>
        <p>{{.Blurb}}</p>
        {{range .Entries}}
        <article class="m" id="{{.ID}}" data-q="{{.Name}} {{.Doc}}">
          <h3><a href="{{.URL}}">{{.Name}}</a></h3>
          <pre><code>{{.Signature}}</code></pre>
          {{if .Doc}}<p>{{.Doc}}</p>{{end}}
        </article>
        {{end}}
      </section>
      {{end}}
    </div>
  </div>
</main>
<footer>
  <div class="wrap">
    <span>Generated from the source with <code>go run ./tools/gen-apidocs</code>.</span>
    <span><a href="./">Home</a> · <a href="https://github.com/0x524a/onvif-go">GitHub</a> · <a href="https://pkg.go.dev/github.com/0x524a/onvif-go">pkg.go.dev</a></span>
  </div>
</footer>
<script>
(function () {
  var box = document.getElementById('find');
  var items = [].slice.call(document.querySelectorAll('.m'));
  var groups = [].slice.call(document.querySelectorAll('.grp'));
  var none = document.getElementById('none');
  box.addEventListener('input', function () {
    var q = box.value.trim().toLowerCase();
    items.forEach(function (m) { m.hidden = q && m.dataset.q.toLowerCase().indexOf(q) < 0; });
    var shown = 0;
    groups.forEach(function (g) {
      var any = g.querySelector('.m:not([hidden])');
      g.hidden = !any;
      if (any) shown++;
      var li = document.querySelector('.side a[href="#' + g.id + '"]');
      if (li) li.parentNode.hidden = !any;
    });
    none.hidden = shown > 0;
  });
})();
</script>
</body>
</html>
`))
