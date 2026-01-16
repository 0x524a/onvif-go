// Package onviftesting provides testing utilities for ONVIF client testing.
package onviftesting

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CapturedExchange represents a single SOAP request/response pair.
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

// CameraCapture holds all captured exchanges for a camera.
type CameraCapture struct {
	CameraName string
	Exchanges  []CapturedExchange
}

// LoadCaptureFromArchive loads all captured exchanges from a tar.gz archive.
func LoadCaptureFromArchive(archivePath string) (*CameraCapture, error) {
	file, err := os.Open(archivePath) //nolint:gosec // File path is from test data, safe
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gzr.Close()
	}()

	tr := tar.NewReader(gzr)

	capture := &CameraCapture{
		CameraName: filepath.Base(archivePath),
		Exchanges:  make([]CapturedExchange, 0),
	}

	// Read all .json files from the archive
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
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

// MockSOAPServer creates a test HTTP server that replays captured SOAP responses.
type MockSOAPServer struct {
	Server  *httptest.Server
	Capture *CameraCapture
}

// NewMockSOAPServer creates a new mock server from a capture archive.
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

// handleRequest matches incoming requests to captured responses.
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
	//nolint:errcheck // Write error is not critical after WriteHeader
	_, _ = w.Write([]byte(exchange.ResponseBody))
}

// Close shuts down the mock server.
func (m *MockSOAPServer) Close() {
	m.Server.Close()
}

// URL returns the mock server's URL.
func (m *MockSOAPServer) URL() string {
	return m.Server.URL
}

// extractOperationFromSOAP extracts the SOAP operation name from request body.
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

// =============================================================================
// Enhanced Mock Server with Parameter-Aware Matching (V2)
// =============================================================================

// MockSOAPServerV2 supports parameter-aware request matching.
// It maintains backward compatibility with V1 captures by falling back to
// operation-name-only matching when parameters don't match.
type MockSOAPServerV2 struct {
	Server      *httptest.Server
	Capture     *CameraCaptureV2
	exchangeMap map[string][]*CapturedExchangeV2 // operationName -> exchanges
	metadata    *CaptureMetadata
}

// NewMockSOAPServerV2 creates an enhanced mock server from a capture archive.
// It supports both V1 and V2 capture formats.
func NewMockSOAPServerV2(archivePath string) (*MockSOAPServerV2, error) {
	capture, metadata, err := LoadCaptureFromArchiveV2(archivePath)
	if err != nil {
		return nil, err
	}

	mock := &MockSOAPServerV2{
		Capture:     capture,
		metadata:    metadata,
		exchangeMap: make(map[string][]*CapturedExchangeV2),
	}

	// Build exchange map for quick lookup
	for i := range capture.Exchanges {
		ex := &capture.Exchanges[i]
		opName := ex.OperationName
		if opName == "" {
			// For V1 captures, extract from request body
			opName = extractOperationFromSOAP(ex.RequestBody)
			ex.OperationName = opName
		}
		mock.exchangeMap[opName] = append(mock.exchangeMap[opName], ex)
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))
	return mock, nil
}

// processArchiveEntry processes a single tar archive entry (JSON file) and adds it to the capture.
// Returns (isMetadata, error).
func processArchiveEntry(header *tar.Header, data []byte, capture *CameraCaptureV2) (*CaptureMetadata, error) {
	// Check for metadata.json (V2 archives)
	if header.Name == "metadata.json" || strings.HasSuffix(header.Name, "/metadata.json") {
		var meta CaptureMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		return &meta, nil
	}

	// Skip files that look like request/response XML stored as JSON
	if strings.Contains(header.Name, "_request") || strings.Contains(header.Name, "_response") {
		return nil, nil
	}

	// Parse exchange from JSON
	exchange, err := parseExchange(header.Name, data)
	if err != nil {
		return nil, err
	}
	if exchange != nil {
		capture.Exchanges = append(capture.Exchanges, *exchange)
	}

	return nil, nil
}

// parseExchange parses a JSON exchange entry, supporting both V1 and V2 formats.
func parseExchange(fileName string, data []byte) (*CapturedExchangeV2, error) {
	version := DetectCaptureVersion(data)
	if version >= "2.0" {
		var exchange CapturedExchangeV2
		if err := json.Unmarshal(data, &exchange); err != nil {
			return nil, fmt.Errorf("failed to unmarshal V2 %s: %w", fileName, err)
		}
		return &exchange, nil
	}

	// V1 format - convert to V2
	var v1Exchange CapturedExchange
	if err := json.Unmarshal(data, &v1Exchange); err != nil {
		return nil, fmt.Errorf("failed to unmarshal V1 %s: %w", fileName, err)
	}
	v2Exchange := ConvertV1ToV2(&v1Exchange)
	// Extract parameters from V1 request body
	v2Exchange.Parameters = ExtractParameters(v2Exchange.OperationName, v2Exchange.RequestBody)
	v2Exchange.ServiceType = DetermineServiceType(v2Exchange.RequestBody)
	return v2Exchange, nil
}

// LoadCaptureFromArchiveV2 loads captures from archive, supporting both V1 and V2 formats.
func LoadCaptureFromArchiveV2(archivePath string) (*CameraCaptureV2, *CaptureMetadata, error) {
	file, err := os.Open(archivePath) //nolint:gosec // File path is from test data, safe
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gzr.Close()
	}()

	tr := tar.NewReader(gzr)

	capture := &CameraCaptureV2{
		Exchanges: make([]CapturedExchangeV2, 0),
	}
	var metadata *CaptureMetadata

	// Read all files from the archive
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Only process JSON files
		if !strings.HasSuffix(header.Name, ".json") {
			continue
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read file %s: %w", header.Name, err)
		}

		// Process the archive entry
		meta, err := processArchiveEntry(header, data, capture)
		if err != nil {
			return nil, nil, err
		}
		if meta != nil {
			metadata = meta
		}
	}

	capture.Metadata = metadata
	return capture, metadata, nil
}

// handleRequest matches incoming requests to captured responses with parameter awareness.
func (m *MockSOAPServerV2) handleRequest(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	operationName := extractOperationFromSOAP(string(reqBody))
	if operationName == "" {
		http.Error(w, "Could not extract operation name from request", http.StatusBadRequest)
		return
	}

	// Get all exchanges for this operation
	exchanges, ok := m.exchangeMap[operationName]
	if !ok || len(exchanges) == 0 {
		http.Error(w, fmt.Sprintf("No capture found for operation: %s", operationName), http.StatusNotFound)
		return
	}

	// Extract parameters from request for matching
	requestParams := ExtractParameters(operationName, string(reqBody))
	requestKey := BuildMatchKey(operationName, requestParams)

	// Find best matching exchange
	var bestMatch *CapturedExchangeV2
	bestScore := -1

	for _, ex := range exchanges {
		exchangeKey := BuildMatchKeyFromExchange(ex)
		score := requestKey.MatchScore(&exchangeKey)
		if score > bestScore {
			bestScore = score
			bestMatch = ex
		}
	}

	if bestMatch == nil {
		// Fall back to first exchange for this operation (V1 behavior)
		bestMatch = exchanges[0]
	}

	// Return the captured response
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(bestMatch.StatusCode)
	//nolint:errcheck // Write error is not critical after WriteHeader
	_, _ = w.Write([]byte(bestMatch.ResponseBody))
}

// Close shuts down the V2 mock server.
func (m *MockSOAPServerV2) Close() {
	m.Server.Close()
}

// URL returns the V2 mock server's URL.
func (m *MockSOAPServerV2) URL() string {
	return m.Server.URL
}

// Metadata returns the capture metadata if available (V2 archives only).
func (m *MockSOAPServerV2) Metadata() *CaptureMetadata {
	return m.metadata
}

// GetExchangeCount returns the total number of captured exchanges.
func (m *MockSOAPServerV2) GetExchangeCount() int {
	return len(m.Capture.Exchanges)
}

// GetOperations returns all unique operation names in the capture.
func (m *MockSOAPServerV2) GetOperations() []string {
	ops := make([]string, 0, len(m.exchangeMap))
	for op := range m.exchangeMap {
		ops = append(ops, op)
	}
	return ops
}

// =============================================================================
// Parameter Extraction
// =============================================================================

// tokenParams are common ONVIF token parameters to extract.
var tokenParams = []string{
	// Core tokens
	"ProfileToken",
	"ConfigurationToken",
	"VideoSourceToken",
	"AudioSourceToken",
	"PresetToken",
	"Token",
	// Configuration tokens
	"VideoSourceConfigurationToken",
	"AudioSourceConfigurationToken",
	"VideoEncoderConfigurationToken",
	"AudioEncoderConfigurationToken",
	"MetadataConfigurationToken",
	"PTZConfigurationToken",
	// Event/subscription tokens
	"SubscriptionReference",
	// Extended tokens (Task 5 additions)
	"OSDToken",
	"NodeToken",
	"RelayOutputToken",
	"VideoOutputToken",
	"DigitalInputToken",
	"SerialPortToken",
	"StorageConfigurationToken",
	"CertificateID",
	"RecordingToken",
	"RecordingJobToken",
	"AnalyticsConfigurationToken",
	"RuleToken",
	"ScheduleToken",
	"SpecialDayGroupToken",
}

// paramRegexes are compiled regexes for extracting parameters.
var paramRegexes = make(map[string]*regexp.Regexp)

func init() {
	// Pre-compile regexes for token extraction
	for _, param := range tokenParams {
		// Match both <ProfileToken>value</ProfileToken> and <trt:ProfileToken>value</trt:ProfileToken>
		pattern := fmt.Sprintf(`<%s[^>]*>([^<]+)</%s>|<[a-z]+:%s[^>]*>([^<]+)</[a-z]+:%s>`,
			param, param, param, param)
		paramRegexes[param] = regexp.MustCompile(pattern)
	}
}

// ExtractParameters extracts key parameters from a SOAP request body.
func ExtractParameters(operationName, soapBody string) map[string]interface{} {
	params := make(map[string]interface{})

	for _, paramName := range tokenParams {
		re := paramRegexes[paramName]
		if re == nil {
			continue
		}

		matches := re.FindStringSubmatch(soapBody)
		if len(matches) > 1 {
			// Get the first non-empty capture group
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					params[paramName] = strings.TrimSpace(matches[i])

					break
				}
			}
		}
	}

	return params
}

// ExtractXMLElement extracts a simple XML element value from a string.
func ExtractXMLElement(xml, element string) string {
	// Try without namespace prefix first
	start := fmt.Sprintf("<%s>", element)
	end := fmt.Sprintf("</%s>", element)

	startIdx := strings.Index(xml, start)
	if startIdx != -1 {
		startIdx += len(start)
		endIdx := strings.Index(xml[startIdx:], end)
		if endIdx != -1 {
			return strings.TrimSpace(xml[startIdx : startIdx+endIdx])
		}
	}

	// Try with namespace prefix pattern :<element>
	pattern := fmt.Sprintf(":%s>", element)
	startIdx = strings.Index(xml, pattern)
	if startIdx != -1 {
		startIdx += len(pattern)
		// Find closing tag with any namespace prefix
		endPattern := fmt.Sprintf("</%s>", element)
		endIdx := strings.Index(xml[startIdx:], endPattern)
		if endIdx == -1 {
			// Try with namespace prefix in closing tag
			for i := startIdx; i < len(xml); i++ {
				if xml[i] == '<' && i+1 < len(xml) && xml[i+1] == '/' {
					// Found potential closing tag
					closeEnd := strings.Index(xml[i:], ">")
					if closeEnd != -1 {
						closeTag := xml[i : i+closeEnd+1]
						if strings.Contains(closeTag, element) {
							return strings.TrimSpace(xml[startIdx:i])
						}
					}
				}
			}
		} else {
			return strings.TrimSpace(xml[startIdx : startIdx+endIdx])
		}
	}

	return ""
}

// =============================================================================
// SOAP Fault Support
// =============================================================================

// SOAPFault represents a SOAP fault for error responses.
type SOAPFault struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
	Detail string `json:"detail,omitempty"`
}

// Common ONVIF SOAP faults.
var (
	FaultActionNotSupported = SOAPFault{
		Code:   "env:Sender/ter:ActionNotSupported",
		Reason: "The requested action is not supported by the service",
	}
	FaultInvalidToken = SOAPFault{
		Code:   "env:Sender/ter:InvalidArgVal/ter:NoProfile",
		Reason: "The requested profile token does not exist",
	}
	FaultNotAuthorized = SOAPFault{
		Code:   "env:Sender/ter:NotAuthorized",
		Reason: "The sender is not authorized to perform the operation",
	}
	FaultInvalidArgument = SOAPFault{
		Code:   "env:Sender/ter:InvalidArgVal",
		Reason: "One or more arguments are invalid",
	}
	FaultOperationFailed = SOAPFault{
		Code:   "env:Receiver/ter:Action",
		Reason: "The operation failed",
	}
)

// GenerateFaultResponse creates a SOAP fault response XML.
func GenerateFaultResponse(fault SOAPFault) string {
	detail := ""
	if fault.Detail != "" {
		detail = fmt.Sprintf("<soap:Detail>%s</soap:Detail>", fault.Detail)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:ter="http://www.onvif.org/ver10/error">
  <soap:Body>
    <soap:Fault>
      <soap:Code>
        <soap:Value>%s</soap:Value>
      </soap:Code>
      <soap:Reason>
        <soap:Text xml:lang="en">%s</soap:Text>
      </soap:Reason>
      %s
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`, fault.Code, fault.Reason, detail)
}

// IsFaultResponse checks if a response body contains a SOAP fault.
func IsFaultResponse(responseBody string) bool {
	return strings.Contains(responseBody, "<soap:Fault>") ||
		strings.Contains(responseBody, "<Fault>") ||
		strings.Contains(responseBody, ":Fault>")
}
