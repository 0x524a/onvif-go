package hygiene_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestDocsLinksResolve guards the READMEs and the docs/ redirect page:
// every relative markdown link must point at a file that exists. A moved
// or renamed file otherwise 404s silently on GitHub.
//
// The topic manuals moved to the MiBee documentation hub (mibee-docs#6,
// single source of truth); docs/ is intentionally a one-page redirect now.
// The old floors (>=10 markdown files, >=20 links) guarded the removed
// in-repo corpus and no longer apply.
func TestDocsLinksResolve(t *testing.T) {
	files := []string{"README.md", "README.zh-CN.md"}
	err := filepath.WalkDir("docs", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs/: %v", err)
	}
	if len(files) < 3 {
		t.Fatalf("expected READMEs + docs redirect page, found %d", len(files))
	}

	checked := 0
	var broken []string
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, target := range relativeLinkTargets(string(raw)) {
			checked++
			resolved := filepath.Join(filepath.Dir(file), filepath.FromSlash(target))
			if _, err := os.Stat(resolved); err != nil {
				broken = append(broken, file+" -> "+target)
			}
		}
	}
	if checked < 3 {
		t.Fatalf("expected a meaningful number of links, checked %d", checked)
	}
	if len(broken) > 0 {
		t.Fatalf("broken documentation links (%d of %d checked):\n%s",
			len(broken), checked, strings.Join(broken, "\n"))
	}
}

// TestReadmeIndexesStaySingleLanguage pins the language separation of the
// documentation indexes: the English README must not link Chinese guides
// (docs/zh/) and the Chinese README must not link English guides (docs/en/).
func TestReadmeIndexesStaySingleLanguage(t *testing.T) {
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	for _, target := range relativeLinkTargets(string(raw)) {
		if strings.HasPrefix(target, "docs/zh/") {
			t.Fatalf("README.md (English) links the Chinese guide %s — "+
				"each language's index must reference only its own guides", target)
		}
	}
	raw, err = os.ReadFile("README.zh-CN.md")
	if err != nil {
		t.Fatalf("read README.zh-CN.md: %v", err)
	}
	for _, target := range relativeLinkTargets(string(raw)) {
		if strings.HasPrefix(target, "docs/en/") {
			t.Fatalf("README.zh-CN.md (Chinese) links the English guide %s — "+
				"each language's index must reference only its own guides", target)
		}
	}
}

// relativeLinkTargets extracts ](target) link targets from markdown text,
// skipping web links, mailto, and pure-#anchor targets.
func relativeLinkTargets(text string) []string {
	var out []string
	for {
		idx := strings.Index(text, "](")
		if idx < 0 {
			break
		}
		rest := text[idx+2:]
		end := strings.IndexByte(rest, ')')
		if end < 0 {
			break
		}
		target := rest[:end]
		path := target
		if i := strings.IndexByte(target, '#'); i >= 0 {
			path = target[:i]
		}
		if path != "" &&
			!strings.HasPrefix(target, "http://") &&
			!strings.HasPrefix(target, "https://") &&
			!strings.HasPrefix(target, "mailto:") {
			out = append(out, path)
		}
		text = rest[end:]
	}
	return out
}

// TestDocsAreInRepo pins the self-contained-docs contract (2026-10-24
// policy): docs/ holds the bilingual manual set plus its index, and every
// URL in docs/ must point at an allowlisted host — the project's own
// hosting and the standards bodies it cites. Any other host (an external
// docs-site pointer) fails the build, whatever domain it uses.
func TestDocsAreInRepo(t *testing.T) {
	entries, err := os.ReadDir("docs")
	if err != nil {
		t.Fatalf("read docs/: %v", err)
	}
	dirs := map[string]bool{}
	for _, e := range entries {
		dirs[e.Name()] = e.IsDir()
	}
	if !dirs["en-US"] || !dirs["zh-CN"] {
		t.Fatalf("docs/ must hold en-US/ and zh-CN/ manual dirs, found %v", entries)
	}
	for _, lang := range []string{"en-US", "zh-CN"} {
		pages, err := os.ReadDir(filepath.Join("docs", lang))
		if err != nil || len(pages) == 0 {
			t.Fatalf("docs/%s must exist and hold manual pages", lang)
		}
	}
	var walk func(dir string) error
	walk = func(dir string) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			p := filepath.Join(dir, e.Name())
			if e.IsDir() {
				if err := walk(p); err != nil {
					return err
				}
				continue
			}
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			for _, host := range docsLinkHosts(string(raw)) {
				if !docsHostAllowed(host) {
					t.Fatalf("%s links to non-allowlisted host %q — docs must be self-contained", p, host)
				}
			}
		}
		return nil
	}
	if err := walk("docs"); err != nil {
		t.Fatalf("walk docs/: %v", err)
	}
}

// docsLinkHostRe captures the host of every http(s) URL.
var docsLinkHostRe = regexp.MustCompile(`https?://([a-zA-Z0-9.-]+)`)

func docsLinkHosts(text string) []string {
	var out []string
	for _, m := range docsLinkHostRe.FindAllStringSubmatch(text, -1) {
		out = append(out, m[1])
	}
	return out
}

// docsAllowedHosts: the only real hosts documentation may point at.
var docsAllowedHosts = map[string]bool{
	"github.com":    true,
	"onvif.org":     true,
	"www.onvif.org": true,
	"w3.org":        true,
	"www.w3.org":    true,
	"pkg.go.dev":    true,
}

// docsHostAllowed also admits the example endpoint placeholders that
// appear in CLI usage snippets (LAN IPs, literal "host").
func docsHostAllowed(host string) bool {
	h := strings.ToLower(host)
	if docsAllowedHosts[h] {
		return true
	}
	switch h {
	case "host", "localhost":
		return true
	}
	return strings.HasPrefix(h, "192.168.")
}
