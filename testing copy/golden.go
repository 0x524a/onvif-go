// Package onviftesting provides testing utilities for ONVIF client testing.
package onviftesting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoldenManifest describes a camera's golden file set.
type GoldenManifest struct {
	Version        string         `json:"version"`
	Camera         CameraInfo     `json:"camera"`
	CaptureDate    string         `json:"capture_date"`
	Capabilities   []string       `json:"capabilities"`
	OperationCount map[string]int `json:"operation_count"`
	Notes          string         `json:"notes,omitempty"`
}

// GoldenFile represents a single operation's expected result.
type GoldenFile struct {
	Operation      string                 `json:"operation"`
	Service        string                 `json:"service"`
	Parameters     map[string]string      `json:"parameters,omitempty"`
	Request        string                 `json:"request"`
	Response       string                 `json:"response"`
	ExpectedFields map[string]interface{} `json:"expected_fields,omitempty"`
	VariableFields []string               `json:"variable_fields,omitempty"`
}

// GoldenFileSet holds all golden files for a camera.
type GoldenFileSet struct {
	Manifest *GoldenManifest
	Files    map[string]*GoldenFile // key is operation + params
	BasePath string
}

// LoadGoldenManifest loads a manifest.json from a golden directory.
func LoadGoldenManifest(goldenDir string) (*GoldenManifest, error) {
	manifestPath := filepath.Join(goldenDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest GoldenManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// LoadGoldenFiles loads all golden files from a camera directory.
func LoadGoldenFiles(goldenDir string) (*GoldenFileSet, error) {
	set := &GoldenFileSet{
		Files:    make(map[string]*GoldenFile),
		BasePath: goldenDir,
	}

	// Load manifest if it exists
	manifest, err := LoadGoldenManifest(goldenDir)
	if err == nil {
		set.Manifest = manifest
	}

	// Walk through all JSON files in the directory
	err = filepath.Walk(goldenDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JSON files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		// Skip manifest.json
		if info.Name() == "manifest.json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var golden GoldenFile
		if err := json.Unmarshal(data, &golden); err != nil {
			return fmt.Errorf("failed to unmarshal %s: %w", path, err)
		}

		// Build key from operation and parameters
		key := buildGoldenKey(&golden)
		set.Files[key] = &golden

		return nil
	})

	if err != nil {
		return nil, err
	}

	return set, nil
}

// buildGoldenKey creates a unique key for a golden file.
func buildGoldenKey(g *GoldenFile) string {
	key := g.Operation
	if g.Parameters != nil {
		// Sort parameters for consistent keys
		for k, v := range g.Parameters {
			key += "_" + k + "_" + v
		}
	}
	return key
}

// GetGoldenFile retrieves a golden file by operation name and parameters.
func (s *GoldenFileSet) GetGoldenFile(operation string, params map[string]string) *GoldenFile {
	// Try exact match first
	golden := &GoldenFile{Operation: operation, Parameters: params}
	key := buildGoldenKey(golden)
	if g, ok := s.Files[key]; ok {
		return g
	}

	// Fall back to operation-only match
	for _, g := range s.Files {
		if g.Operation == operation {
			return g
		}
	}

	return nil
}

// GetOperations returns all unique operations in the golden file set.
func (s *GoldenFileSet) GetOperations() []string {
	seen := make(map[string]bool)
	var ops []string

	for _, g := range s.Files {
		if !seen[g.Operation] {
			seen[g.Operation] = true
			ops = append(ops, g.Operation)
		}
	}

	return ops
}

// ValidateResponse validates a response against expected fields in a golden file.
func ValidateResponse(response interface{}, golden *GoldenFile) []string {
	if golden.ExpectedFields == nil {
		return nil
	}

	var errors []string

	// Convert response to map for comparison
	responseData, err := toMap(response)
	if err != nil {
		return []string{fmt.Sprintf("failed to convert response: %v", err)}
	}

	// Check each expected field
	for field, expected := range golden.ExpectedFields {
		actual, ok := responseData[field]
		if !ok {
			errors = append(errors, fmt.Sprintf("missing field: %s", field))
			continue
		}

		// Skip variable fields (like timestamps)
		if isVariableField(field, golden.VariableFields) {
			continue
		}

		// Compare values
		if !valuesEqual(expected, actual) {
			errors = append(errors, fmt.Sprintf("field %s: expected %v, got %v", field, expected, actual))
		}
	}

	return errors
}

// toMap converts a struct to a map for field comparison.
func toMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// isVariableField checks if a field should be skipped during validation.
func isVariableField(field string, variableFields []string) bool {
	for _, v := range variableFields {
		if v == field {
			return true
		}
	}
	return false
}

// valuesEqual compares two values for equality.
func valuesEqual(expected, actual interface{}) bool {
	// Handle nil comparison
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	// Convert to JSON for deep comparison
	e, err1 := json.Marshal(expected)
	a, err2 := json.Marshal(actual)
	if err1 != nil || err2 != nil {
		return false
	}

	return string(e) == string(a)
}

// SaveGoldenFile saves a golden file to disk.
func SaveGoldenFile(golden *GoldenFile, outputPath string) error {
	data, err := json.MarshalIndent(golden, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal golden file: %w", err)
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0750); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil { //nolint:mnd
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SaveGoldenManifest saves a manifest file to disk.
func SaveGoldenManifest(manifest *GoldenManifest, outputPath string) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil { //nolint:mnd
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// GenerateGoldenFileName generates a filename for a golden file.
func GenerateGoldenFileName(operation string, params map[string]string) string {
	name := operation
	if params != nil {
		for k, v := range params {
			// Sanitize parameter value for filename
			v = strings.ReplaceAll(v, "/", "_")
			v = strings.ReplaceAll(v, "\\", "_")
			name += "_" + k + "_" + v
		}
	}
	return name + ".json"
}

// CreateGoldenFromCapture creates a golden file from a captured exchange.
func CreateGoldenFromCapture(exchange *CapturedExchangeV2) *GoldenFile {
	params := make(map[string]string)
	if exchange.Parameters != nil {
		for k, v := range exchange.Parameters {
			if s, ok := v.(string); ok {
				params[k] = s
			}
		}
	}

	return &GoldenFile{
		Operation:  exchange.OperationName,
		Service:    string(exchange.ServiceType),
		Parameters: params,
		Request:    exchange.RequestBody,
		Response:   exchange.ResponseBody,
	}
}

// GoldenTestRunner helps run tests against golden files.
type GoldenTestRunner struct {
	GoldenSet *GoldenFileSet
}

// NewGoldenTestRunner creates a new golden test runner.
func NewGoldenTestRunner(goldenDir string) (*GoldenTestRunner, error) {
	set, err := LoadGoldenFiles(goldenDir)
	if err != nil {
		return nil, err
	}

	return &GoldenTestRunner{GoldenSet: set}, nil
}

// ValidateOperation validates a response against the golden file for an operation.
func (r *GoldenTestRunner) ValidateOperation(operation string, params map[string]string, response interface{}) []string {
	golden := r.GoldenSet.GetGoldenFile(operation, params)
	if golden == nil {
		return []string{fmt.Sprintf("no golden file found for operation: %s", operation)}
	}

	return ValidateResponse(response, golden)
}
