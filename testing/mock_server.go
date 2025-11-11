package onviftesting

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
)

// CapturedExchange represents a single SOAP request/response pair
type CapturedExchange struct {
	Timestamp     string `json:"timestamp"`
	Operation     int    `json:"operation"`
	OperationName string `json:"operation_name,omitempty"`
	Endpoint      string `json:"endpoint"`
	RequestBody   string `json:"request_body"`
	ResponseBody  string `json:"response_body"`
	StatusCode    int    `json:"status_code"`
	Error         string `json:"error,omitempty"`
}

// CameraCapture holds all captured exchanges for a camera
type CameraCapture struct {
	CameraName string
	Exchanges  []CapturedExchange
}

// LoadCaptureFromArchive loads all captured exchanges from a tar.gz archive
func LoadCaptureFromArchive(archivePath string) (*CameraCapture, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	capture := &CameraCapture{
		CameraName: filepath.Base(archivePath),
		Exchanges:  make([]CapturedExchange, 0),
	}

	// Read all .json files from the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Only process JSON metadata files
		if !strings.HasSuffix(header.Name, ".json") {
			continue
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", header.Name, err)
		}

		var exchange CapturedExchange
		if err := json.Unmarshal(data, &exchange); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %w", header.Name, err)
		}

		capture.Exchanges = append(capture.Exchanges, exchange)
	}

	return capture, nil
}

// MockSOAPServer creates a test HTTP server that replays captured SOAP responses
type MockSOAPServer struct {
	Server  *httptest.Server
	Capture *CameraCapture
}

// NewMockSOAPServer creates a new mock server from a capture archive
func NewMockSOAPServer(archivePath string) (*MockSOAPServer, error) {
	capture, err := LoadCaptureFromArchive(archivePath)
	if err != nil {
		return nil, err
	}

	mock := &MockSOAPServer{
		Capture: capture,
	}

	// Create HTTP test server
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	return mock, nil
}

// handleRequest matches incoming requests to captured responses
func (m *MockSOAPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Read request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	// Extract operation name from request
	operationName := extractOperationFromSOAP(string(reqBody))

	// Find matching response by operation name
	var exchange *CapturedExchange

	if operationName != "" {
		// Try matching by operation_name field if available
		for i := range m.Capture.Exchanges {
			if m.Capture.Exchanges[i].OperationName == operationName {
				exchange = &m.Capture.Exchanges[i]
				break
			}
		}

		// If not found by operation_name, try matching by extracting from request body
		if exchange == nil {
			for i := range m.Capture.Exchanges {
				capturedOp := extractOperationFromSOAP(m.Capture.Exchanges[i].RequestBody)
				if capturedOp == operationName {
					exchange = &m.Capture.Exchanges[i]
					break
				}
			}
		}
	}

	if exchange == nil {
		http.Error(w, fmt.Sprintf("No matching capture found for operation: %s", operationName), http.StatusNotFound)
		return
	}

	// Return the captured response
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(exchange.StatusCode)
	w.Write([]byte(exchange.ResponseBody))
}

// Close shuts down the mock server
func (m *MockSOAPServer) Close() {
	m.Server.Close()
}

// URL returns the mock server's URL
func (m *MockSOAPServer) URL() string {
	return m.Server.URL
}

// extractOperationFromSOAP extracts the SOAP operation name from request body
func extractOperationFromSOAP(soapBody string) string {
	// Find the Body element
	bodyStart := strings.Index(soapBody, "<Body")
	if bodyStart == -1 {
		return ""
	}

	// Find the closing > of the Body opening tag
	bodyOpenEnd := strings.Index(soapBody[bodyStart:], ">")
	if bodyOpenEnd == -1 {
		return ""
	}
	bodyContentStart := bodyStart + bodyOpenEnd + 1

	// Skip whitespace
	for bodyContentStart < len(soapBody) && soapBody[bodyContentStart] <= ' ' {
		bodyContentStart++
	}

	if bodyContentStart >= len(soapBody) || soapBody[bodyContentStart] != '<' {
		return ""
	}

	// Extract the tag name
	tagStart := bodyContentStart + 1
	tagEnd := tagStart
	for tagEnd < len(soapBody) && soapBody[tagEnd] != ' ' && soapBody[tagEnd] != '>' && soapBody[tagEnd] != '/' {
		tagEnd++
	}

	if tagEnd > tagStart {
		tagName := soapBody[tagStart:tagEnd]
		// Remove namespace prefix if present
		if colonIdx := strings.Index(tagName, ":"); colonIdx != -1 {
			return tagName[colonIdx+1:]
		}
		return tagName
	}

	return ""
}
