// Package onviftesting provides testing utilities for ONVIF client testing.
package onviftesting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Registry holds information about all available camera captures.
type Registry struct {
	Version     string            `json:"version"`
	LastUpdated time.Time         `json:"last_updated"`
	Cameras     []CameraEntry     `json:"cameras"`
	Coverage    map[string]Coverage `json:"coverage"`
}

// CameraEntry represents a single camera in the registry.
type CameraEntry struct {
	ID                string   `json:"id"`
	Manufacturer      string   `json:"manufacturer"`
	Model             string   `json:"model"`
	Firmware          string   `json:"firmware"`
	CaptureFile       string   `json:"capture_file"`
	CaptureVersion    string   `json:"capture_version,omitempty"`
	Capabilities      []string `json:"capabilities"`
	OperationsCaptured int     `json:"operations_captured"`
	ProfileCompliance []string `json:"profile_compliance,omitempty"`
	TestFile          string   `json:"test_file,omitempty"`
	Notes             string   `json:"notes,omitempty"`
	AddedDate         string   `json:"added_date,omitempty"`
}

// Coverage tracks operation coverage per service.
type Coverage struct {
	Total    int `json:"total"`
	Captured int `json:"captured"`
}

// RegistryVersion is the current registry format version.
const RegistryVersion = "1.0"

// DefaultRegistryPath is the default path for the registry file.
const DefaultRegistryPath = "testdata/captures/registry.json"

// LoadRegistry loads the capture registry from a file.
func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty registry if file doesn't exist
			return &Registry{
				Version:     RegistryVersion,
				LastUpdated: time.Now(),
				Cameras:     []CameraEntry{},
				Coverage:    make(map[string]Coverage),
			}, nil
		}
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registry: %w", err)
	}

	return &registry, nil
}

// SaveRegistry saves the registry to a file.
func SaveRegistry(registry *Registry, path string) error {
	registry.LastUpdated = time.Now()

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil { //nolint:mnd
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// AddCamera adds a new camera to the registry.
func (r *Registry) AddCamera(entry CameraEntry) {
	// Check if camera already exists
	for i, cam := range r.Cameras {
		if cam.ID == entry.ID {
			// Update existing entry
			r.Cameras[i] = entry
			return
		}
	}

	// Add new entry
	if entry.AddedDate == "" {
		entry.AddedDate = time.Now().Format("2006-01-02")
	}
	r.Cameras = append(r.Cameras, entry)
}

// GetCamera retrieves a camera entry by ID.
func (r *Registry) GetCamera(id string) *CameraEntry {
	for i := range r.Cameras {
		if r.Cameras[i].ID == id {
			return &r.Cameras[i]
		}
	}
	return nil
}

// RemoveCamera removes a camera from the registry.
func (r *Registry) RemoveCamera(id string) bool {
	for i, cam := range r.Cameras {
		if cam.ID == id {
			r.Cameras = append(r.Cameras[:i], r.Cameras[i+1:]...)
			return true
		}
	}
	return false
}

// GetCamerasByManufacturer returns all cameras from a specific manufacturer.
func (r *Registry) GetCamerasByManufacturer(manufacturer string) []CameraEntry {
	var cameras []CameraEntry
	for _, cam := range r.Cameras {
		if cam.Manufacturer == manufacturer {
			cameras = append(cameras, cam)
		}
	}
	return cameras
}

// UpdateCoverage updates the coverage statistics based on registered cameras.
func (r *Registry) UpdateCoverage() {
	// Define total operations per service
	totals := map[string]int{
		"Device":   len(DeviceReadOperations),
		"Media":    len(MediaReadOperations),
		"PTZ":      len(PTZReadOperations),
		"Imaging":  len(ImagingReadOperations),
		"Event":    len(EventReadOperations),
		"DeviceIO": len(DeviceIOReadOperations),
	}

	// Initialize coverage
	r.Coverage = make(map[string]Coverage)
	for service, total := range totals {
		r.Coverage[service] = Coverage{
			Total:    total,
			Captured: 0, // Would need to analyze captures to determine actual coverage
		}
	}
}

// GetTotalCoverage returns the total coverage across all services.
func (r *Registry) GetTotalCoverage() (total int, captured int) {
	for _, cov := range r.Coverage {
		total += cov.Total
		captured += cov.Captured
	}
	return total, captured
}

// GenerateCameraID generates a unique ID for a camera.
func GenerateCameraID(manufacturer, model, firmware string) string {
	// Sanitize and combine
	id := fmt.Sprintf("%s_%s_%s", manufacturer, model, firmware)
	id = sanitizeID(id)
	return id
}

// sanitizeID removes or replaces invalid characters in an ID.
func sanitizeID(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z':
			result = append(result, c)
		case c >= 'A' && c <= 'Z':
			result = append(result, c+'a'-'A') // lowercase
		case c >= '0' && c <= '9':
			result = append(result, c)
		case c == ' ' || c == '-' || c == '_' || c == '.':
			result = append(result, '_')
		}
	}
	return string(result)
}

// ValidateRegistry checks if all referenced capture files exist.
func ValidateRegistry(registry *Registry, basePath string) []string {
	var errors []string

	for _, cam := range registry.Cameras {
		capturePath := filepath.Join(basePath, cam.CaptureFile)
		if _, err := os.Stat(capturePath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("camera %s: capture file not found: %s", cam.ID, cam.CaptureFile))
		}

		if cam.TestFile != "" {
			testPath := filepath.Join(basePath, cam.TestFile)
			if _, err := os.Stat(testPath); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("camera %s: test file not found: %s", cam.ID, cam.TestFile))
			}
		}
	}

	return errors
}

// CreateCameraEntryFromCapture creates a registry entry from a capture archive.
func CreateCameraEntryFromCapture(archivePath string) (*CameraEntry, error) {
	capture, metadata, err := LoadCaptureFromArchiveV2(archivePath)
	if err != nil {
		return nil, err
	}

	// Extract camera info
	var cameraInfo CameraInfo
	if metadata != nil {
		cameraInfo = metadata.CameraInfo
	} else {
		// Try to extract from GetDeviceInformation response
		for _, ex := range capture.Exchanges {
			if ex.OperationName == "GetDeviceInformation" {
				cameraInfo.Manufacturer = ExtractXMLElement(ex.ResponseBody, "Manufacturer")
				cameraInfo.Model = ExtractXMLElement(ex.ResponseBody, "Model")
				cameraInfo.FirmwareVersion = ExtractXMLElement(ex.ResponseBody, "FirmwareVersion")
				break
			}
		}
	}

	// Determine capabilities from captured operations
	capabilities := detectCapabilities(capture)

	entry := &CameraEntry{
		ID:                GenerateCameraID(cameraInfo.Manufacturer, cameraInfo.Model, cameraInfo.FirmwareVersion),
		Manufacturer:      cameraInfo.Manufacturer,
		Model:             cameraInfo.Model,
		Firmware:          cameraInfo.FirmwareVersion,
		CaptureFile:       filepath.Base(archivePath),
		OperationsCaptured: len(capture.Exchanges),
		Capabilities:      capabilities,
		AddedDate:         time.Now().Format("2006-01-02"),
	}

	if metadata != nil {
		entry.CaptureVersion = metadata.Version
	}

	return entry, nil
}

// detectCapabilities determines which services are captured.
func detectCapabilities(capture *CameraCaptureV2) []string {
	services := make(map[string]bool)

	for _, ex := range capture.Exchanges {
		if ex.ServiceType != "" {
			services[string(ex.ServiceType)] = true
		} else {
			// Infer from operation name
			svc := inferServiceFromOperation(ex.OperationName)
			if svc != "" {
				services[svc] = true
			}
		}
	}

	var result []string
	for svc := range services {
		result = append(result, svc)
	}
	return result
}

// inferServiceFromOperation guesses the service type from an operation name.
func inferServiceFromOperation(op string) string {
	// Media operations typically have these patterns
	mediaOps := []string{"Profile", "Stream", "Encoder", "VideoSource", "AudioSource", "OSD", "Metadata"}
	for _, pattern := range mediaOps {
		if containsSubstring(op, pattern) {
			return "Media"
		}
	}

	// PTZ operations
	if containsSubstring(op, "PTZ") || containsSubstring(op, "Preset") || containsSubstring(op, "Move") {
		return "PTZ"
	}

	// Imaging operations
	if containsSubstring(op, "Imaging") || op == "GetOptions" || op == "GetMoveOptions" {
		return "Imaging"
	}

	// Event operations
	if containsSubstring(op, "Event") || containsSubstring(op, "Subscription") {
		return "Event"
	}

	// Default to Device
	return "Device"
}

// containsSubstring checks if s contains substr (case-sensitive).
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring finds substr in s, returns -1 if not found.
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// RegistrySummary provides a summary of the registry.
type RegistrySummary struct {
	TotalCameras      int
	TotalOperations   int
	CapturedOperations int
	ManufacturerCount map[string]int
	ServiceCoverage   map[string]float64
}

// GetSummary generates a summary of the registry.
func (r *Registry) GetSummary() RegistrySummary {
	summary := RegistrySummary{
		TotalCameras:      len(r.Cameras),
		ManufacturerCount: make(map[string]int),
		ServiceCoverage:   make(map[string]float64),
	}

	// Count by manufacturer
	for _, cam := range r.Cameras {
		summary.ManufacturerCount[cam.Manufacturer]++
	}

	// Calculate coverage percentages
	for service, cov := range r.Coverage {
		summary.TotalOperations += cov.Total
		summary.CapturedOperations += cov.Captured
		if cov.Total > 0 {
			summary.ServiceCoverage[service] = float64(cov.Captured) / float64(cov.Total) * 100
		}
	}

	return summary
}
