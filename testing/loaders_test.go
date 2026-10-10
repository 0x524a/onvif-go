package onviftesting

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeArchive builds a .tar.gz holding the given name -> content entries.
func writeArchive(t *testing.T, entries map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "cap.tar.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer func() { _ = f.Close() }()

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	for name, body := range entries {
		hdr := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(body))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header: %v", err)
		}

		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatalf("write body: %v", err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}

	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}

	return path
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

const exchangeJSON = `{"timestamp":"t","operation":1,"operation_name":"GetDeviceInformation",` +
	`"endpoint":"http://c/onvif/device_service","request_body":"<req/>","response_body":"<resp/>","status_code":200}`

func TestLoadCaptureFromArchive(t *testing.T) {
	path := writeArchive(t, map[string]string{
		"001_GetDeviceInformation.json": exchangeJSON,
		"001_request.xml":               "<ignored/>",
	})

	capture, err := LoadCaptureFromArchive(path)
	if err != nil {
		t.Fatalf("LoadCaptureFromArchive() error = %v", err)
	}

	if capture.CameraName != "cap.tar.gz" {
		t.Errorf("CameraName = %q", capture.CameraName)
	}

	if len(capture.Exchanges) != 1 || capture.Exchanges[0].OperationName != "GetDeviceInformation" {
		t.Errorf("Exchanges = %+v", capture.Exchanges)
	}
}

func TestLoadCaptureFromArchiveErrors(t *testing.T) {
	if _, err := LoadCaptureFromArchive(filepath.Join(t.TempDir(), "missing.tar.gz")); err == nil {
		t.Error("missing archive: expected error")
	}

	notGzip := filepath.Join(t.TempDir(), "plain.tar.gz")
	writeFile(t, notGzip, "not gzip")

	if _, err := LoadCaptureFromArchive(notGzip); err == nil {
		t.Error("non-gzip archive: expected error")
	}

	bad := writeArchive(t, map[string]string{"bad.json": "{not json"})
	if _, err := LoadCaptureFromArchive(bad); err == nil {
		t.Error("malformed JSON entry: expected error")
	}
}

func TestLoadCaptureFromArchiveV2(t *testing.T) {
	path := writeArchive(t, map[string]string{
		"001_GetDeviceInformation.json": exchangeJSON,
		"metadata.json":                 `{"capture_format_version":"2.0"}`,
	})

	capture, meta, err := LoadCaptureFromArchiveV2(path)
	if err != nil {
		t.Fatalf("LoadCaptureFromArchiveV2() error = %v", err)
	}

	if meta == nil || capture.Metadata != meta {
		t.Errorf("metadata not loaded: %+v", meta)
	}

	if len(capture.Exchanges) != 1 {
		t.Errorf("got %d exchanges, want 1", len(capture.Exchanges))
	}

	if _, _, err := LoadCaptureFromArchiveV2(filepath.Join(t.TempDir(), "missing.tar.gz")); err == nil {
		t.Error("missing archive: expected error")
	}
}

func TestLoadRegistry(t *testing.T) {
	missing, err := LoadRegistry(filepath.Join(t.TempDir(), "none.json"))
	if err != nil {
		t.Fatalf("missing file should yield an empty registry, got %v", err)
	}

	if missing.Version != RegistryVersion || len(missing.Cameras) != 0 {
		t.Errorf("empty registry = %+v", missing)
	}

	path := filepath.Join(t.TempDir(), "registry.json")
	writeFile(t, path, `{"version":"1.0","cameras":[{"name":"cam"}],"coverage":{}}`)

	loaded, err := LoadRegistry(path)
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}

	if len(loaded.Cameras) != 1 {
		t.Errorf("Cameras = %+v", loaded.Cameras)
	}

	writeFile(t, path, "{not json")

	if _, err := LoadRegistry(path); err == nil {
		t.Error("malformed registry: expected error")
	}

	if _, err := LoadRegistry(t.TempDir()); err == nil {
		t.Error("directory as registry: expected read error")
	}
}

func TestLoadGoldenManifestAndFiles(t *testing.T) {
	dir := t.TempDir()

	if _, err := LoadGoldenManifest(dir); err == nil {
		t.Error("missing manifest: expected error")
	}

	writeFile(t, filepath.Join(dir, "manifest.json"), `{"version":"1.0","capture_date":"2026-01-01"}`)
	writeFile(t, filepath.Join(dir, "GetDeviceInformation.json"),
		`{"operation":"GetDeviceInformation","service":"device","request":"<q/>","response":"<r/>"}`)
	writeFile(t, filepath.Join(dir, "notes.txt"), "ignored")

	manifest, err := LoadGoldenManifest(dir)
	if err != nil || manifest.Version != "1.0" {
		t.Fatalf("LoadGoldenManifest() = %+v, %v", manifest, err)
	}

	set, err := LoadGoldenFiles(dir)
	if err != nil {
		t.Fatalf("LoadGoldenFiles() error = %v", err)
	}

	if set.Manifest == nil || len(set.Files) != 1 {
		t.Errorf("set = manifest %v, %d files", set.Manifest, len(set.Files))
	}

	writeFile(t, filepath.Join(dir, "broken.json"), "{not json")

	if _, err := LoadGoldenFiles(dir); err == nil {
		t.Error("malformed golden file: expected error")
	}

	writeFile(t, filepath.Join(dir, "manifest.json"), "{not json")

	if _, err := LoadGoldenManifest(dir); err == nil {
		t.Error("malformed manifest: expected error")
	}
}

// A captured response that advertises the camera's own address must be served
// with the mock server's address instead, or the client would leave the mock.
func TestMockServerRewritesCapturedCameraAddress(t *testing.T) {
	const caps = `<GetCapabilitiesResponse><XAddr>http://192.168.2.9:8000/onvif/media_service</XAddr>` +
		`<XAddr>http://192.168.2.9/onvif/ptz_service</XAddr><Other>http://example.org/keep</Other></GetCapabilitiesResponse>`

	exchange := `{"timestamp":"t","operation":1,"operation_name":"GetCapabilities",` +
		`"endpoint":"http://192.168.2.9:8000/onvif/device_service",` +
		`"request_body":"<Body><GetCapabilities/></Body>","response_body":` + quoteJSON(caps) + `,"status_code":200}`

	path := writeArchive(t, map[string]string{"001_GetCapabilities.json": exchange})

	for name, start := range map[string]func(string) (url string, closeFn func(), err error){
		"v1": func(p string) (string, func(), error) {
			m, err := NewMockSOAPServer(p)
			if err != nil {
				return "", nil, err
			}

			return m.URL(), m.Close, nil
		},
		"v2": func(p string) (string, func(), error) {
			m, err := NewMockSOAPServerV2(p)
			if err != nil {
				return "", nil, err
			}

			return m.URL(), m.Close, nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			base, closeFn, err := start(path)
			if err != nil {
				t.Fatalf("start mock: %v", err)
			}
			defer closeFn()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, base+"/onvif/device_service",
				strings.NewReader("<Body><GetCapabilities/></Body>"))
			if err != nil {
				t.Fatalf("build request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("POST: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			raw, _ := io.ReadAll(resp.Body)
			body := string(raw)

			if strings.Contains(body, "192.168.2.9") {
				t.Errorf("response still points at the captured camera:\n%s", body)
			}

			for _, want := range []string{base + "/onvif/media_service", base + "/onvif/ptz_service", "http://example.org/keep"} {
				if !strings.Contains(body, want) {
					t.Errorf("response missing %q:\n%s", want, body)
				}
			}
		})
	}
}

func TestCapturedCameraPatternNoHosts(t *testing.T) {
	if p := capturedCameraPattern([]string{"", "::bad::"}); p != nil {
		t.Errorf("pattern = %v, want nil when no host is known", p)
	}

	if got := redirectToMock("http://x/y", nil, "http://mock"); got != "http://x/y" {
		t.Errorf("nil pattern must leave the body alone, got %q", got)
	}
}

func quoteJSON(s string) string {
	b, _ := json.Marshal(s)

	return string(b)
}
