// Package onviftesting provides testing utilities for ONVIF client testing.
package onviftesting

import (
	"encoding/json"
	"time"
)

// CaptureVersion is the current capture format version.
const CaptureVersion = "2.0"

// ServiceType categorizes ONVIF services.
type ServiceType string

const (
	ServiceDevice   ServiceType = "Device"
	ServiceMedia    ServiceType = "Media"
	ServicePTZ      ServiceType = "PTZ"
	ServiceImaging  ServiceType = "Imaging"
	ServiceEvent    ServiceType = "Event"
	ServiceDeviceIO ServiceType = "DeviceIO"
	ServiceUnknown  ServiceType = "Unknown"
)

// CameraInfo stores camera identification information.
type CameraInfo struct {
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmware_version"`
	SerialNumber    string `json:"serial_number,omitempty"`
	HardwareID      string `json:"hardware_id,omitempty"`
}

// CaptureMetadata contains versioned capture archive metadata.
// This is stored as metadata.json in V2 archives.
type CaptureMetadata struct {
	Version        string            `json:"version"`
	CreatedAt      time.Time         `json:"created_at"`
	ToolVersion    string            `json:"tool_version"`
	CameraInfo     CameraInfo        `json:"camera_info"`
	TotalExchanges int               `json:"total_exchanges"`
	ServiceMap     map[string]string `json:"service_map,omitempty"` // operation -> service type
	Tags           []string          `json:"tags,omitempty"`
}

// CapturedExchangeV2 extends the original CapturedExchange with parameter awareness
// and additional metadata for smarter request matching.
type CapturedExchangeV2 struct {
	// Version indicates the capture format version (empty for V1, "2.0" for V2)
	Version string `json:"version,omitempty"`

	// Timestamp is when the exchange was captured (RFC3339 format)
	Timestamp string `json:"timestamp"`

	// Sequence is the capture order (1-indexed for V2, 0-indexed for V1)
	Sequence int `json:"sequence,omitempty"`

	// Operation is deprecated in V2, kept for V1 compatibility
	Operation int `json:"operation,omitempty"`

	// OperationName is the SOAP operation name (e.g., "GetDeviceInformation")
	OperationName string `json:"operation_name,omitempty"`

	// ServiceType categorizes which ONVIF service handles this operation
	ServiceType ServiceType `json:"service_type,omitempty"`

	// Parameters contains extracted key parameters from the request
	// Common keys: ProfileToken, ConfigurationToken, VideoSourceToken, etc.
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Endpoint is the URL the request was sent to
	Endpoint string `json:"endpoint"`

	// RequestBody is the full SOAP request XML
	RequestBody string `json:"request_body"`

	// ResponseBody is the full SOAP response XML
	ResponseBody string `json:"response_body"`

	// StatusCode is the HTTP response status code
	StatusCode int `json:"status_code"`

	// DurationNs is the request duration in nanoseconds
	DurationNs int64 `json:"duration_ns,omitempty"`

	// Success indicates if the operation succeeded (no SOAP fault)
	Success bool `json:"success,omitempty"`

	// Error contains error message if the operation failed
	Error string `json:"error,omitempty"`
}

// IsV2 returns true if this exchange is in V2 format.
func (e *CapturedExchangeV2) IsV2() bool {
	return e.Version != "" && e.Version >= "2.0"
}

// GetProfileToken returns the ProfileToken parameter if present.
func (e *CapturedExchangeV2) GetProfileToken() string {
	if e.Parameters == nil {
		return ""
	}
	if token, ok := e.Parameters["ProfileToken"].(string); ok {
		return token
	}
	return ""
}

// GetConfigurationToken returns the ConfigurationToken parameter if present.
func (e *CapturedExchangeV2) GetConfigurationToken() string {
	if e.Parameters == nil {
		return ""
	}
	if token, ok := e.Parameters["ConfigurationToken"].(string); ok {
		return token
	}
	// Also check for Token (some operations use just "Token")
	if token, ok := e.Parameters["Token"].(string); ok {
		return token
	}
	return ""
}

// GetVideoSourceToken returns the VideoSourceToken parameter if present.
func (e *CapturedExchangeV2) GetVideoSourceToken() string {
	if e.Parameters == nil {
		return ""
	}
	if token, ok := e.Parameters["VideoSourceToken"].(string); ok {
		return token
	}
	return ""
}

// GetAudioSourceToken returns the AudioSourceToken parameter if present.
func (e *CapturedExchangeV2) GetAudioSourceToken() string {
	if e.Parameters == nil {
		return ""
	}
	if token, ok := e.Parameters["AudioSourceToken"].(string); ok {
		return token
	}
	return ""
}

// GetPresetToken returns the PresetToken parameter if present.
func (e *CapturedExchangeV2) GetPresetToken() string {
	if e.Parameters == nil {
		return ""
	}
	if token, ok := e.Parameters["PresetToken"].(string); ok {
		return token
	}
	return ""
}

// GetNodeToken returns the NodeToken parameter if present.
func (e *CapturedExchangeV2) GetNodeToken() string {
	if e.Parameters == nil {
		return ""
	}
	if token, ok := e.Parameters["NodeToken"].(string); ok {
		return token
	}
	return ""
}

// GetOSDToken returns the OSDToken parameter if present.
func (e *CapturedExchangeV2) GetOSDToken() string {
	if e.Parameters == nil {
		return ""
	}
	if token, ok := e.Parameters["OSDToken"].(string); ok {
		return token
	}
	return ""
}

// CameraCaptureV2 holds all captured exchanges for a camera with metadata.
type CameraCaptureV2 struct {
	Metadata  *CaptureMetadata     `json:"metadata,omitempty"`
	Exchanges []CapturedExchangeV2 `json:"exchanges"`
}

// MatchKey uniquely identifies a capture for parameter-aware matching.
type MatchKey struct {
	OperationName      string
	ProfileToken       string
	ConfigurationToken string
	VideoSourceToken   string
	// Extended fields for better matching
	AudioSourceToken string
	PresetToken      string
	NodeToken        string
	OSDToken         string
}

// String returns a string representation of the match key for debugging.
func (k MatchKey) String() string {
	s := k.OperationName
	if k.ProfileToken != "" {
		s += "[Profile:" + k.ProfileToken + "]"
	}
	if k.ConfigurationToken != "" {
		s += "[Config:" + k.ConfigurationToken + "]"
	}
	if k.VideoSourceToken != "" {
		s += "[VideoSource:" + k.VideoSourceToken + "]"
	}
	if k.AudioSourceToken != "" {
		s += "[AudioSource:" + k.AudioSourceToken + "]"
	}
	if k.PresetToken != "" {
		s += "[Preset:" + k.PresetToken + "]"
	}
	if k.NodeToken != "" {
		s += "[Node:" + k.NodeToken + "]"
	}
	if k.OSDToken != "" {
		s += "[OSD:" + k.OSDToken + "]"
	}
	return s
}

// BuildMatchKey creates a MatchKey from an operation name and parameters.
func BuildMatchKey(operationName string, params map[string]interface{}) MatchKey {
	key := MatchKey{
		OperationName: operationName,
	}

	if params == nil {
		return key
	}

	if token, ok := params["ProfileToken"].(string); ok {
		key.ProfileToken = token
	}
	if token, ok := params["ConfigurationToken"].(string); ok {
		key.ConfigurationToken = token
	} else if token, ok := params["Token"].(string); ok {
		key.ConfigurationToken = token
	}
	if token, ok := params["VideoSourceToken"].(string); ok {
		key.VideoSourceToken = token
	}
	if token, ok := params["AudioSourceToken"].(string); ok {
		key.AudioSourceToken = token
	}
	if token, ok := params["PresetToken"].(string); ok {
		key.PresetToken = token
	}
	if token, ok := params["NodeToken"].(string); ok {
		key.NodeToken = token
	}
	if token, ok := params["OSDToken"].(string); ok {
		key.OSDToken = token
	}

	return key
}

// BuildMatchKeyFromExchange creates a MatchKey from a captured exchange.
func BuildMatchKeyFromExchange(exchange *CapturedExchangeV2) MatchKey {
	return MatchKey{
		OperationName:      exchange.OperationName,
		ProfileToken:       exchange.GetProfileToken(),
		ConfigurationToken: exchange.GetConfigurationToken(),
		VideoSourceToken:   exchange.GetVideoSourceToken(),
		AudioSourceToken:   exchange.GetAudioSourceToken(),
		PresetToken:        exchange.GetPresetToken(),
		NodeToken:          exchange.GetNodeToken(),
		OSDToken:           exchange.GetOSDToken(),
	}
}

// MatchScore returns how well two MatchKeys match (higher is better).
// Returns -1 if operation names don't match.
func (k MatchKey) MatchScore(other MatchKey) int {
	if k.OperationName != other.OperationName {
		return -1
	}

	score := 1 // Base score for matching operation

	// Bonus points for matching parameters
	if k.ProfileToken != "" && k.ProfileToken == other.ProfileToken {
		score += 10
	}
	if k.ConfigurationToken != "" && k.ConfigurationToken == other.ConfigurationToken {
		score += 10
	}
	if k.VideoSourceToken != "" && k.VideoSourceToken == other.VideoSourceToken {
		score += 10
	}
	if k.AudioSourceToken != "" && k.AudioSourceToken == other.AudioSourceToken {
		score += 10
	}
	if k.PresetToken != "" && k.PresetToken == other.PresetToken {
		score += 10
	}
	if k.NodeToken != "" && k.NodeToken == other.NodeToken {
		score += 10
	}
	if k.OSDToken != "" && k.OSDToken == other.OSDToken {
		score += 10
	}

	return score
}

// DetectCaptureVersion determines if JSON data is V1 or V2 format.
func DetectCaptureVersion(data []byte) string {
	var probe struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return "1.0"
	}
	if probe.Version == "" {
		return "1.0"
	}
	return probe.Version
}

// ConvertV1ToV2 converts a V1 CapturedExchange to V2 format.
func ConvertV1ToV2(v1 *CapturedExchange) *CapturedExchangeV2 {
	return &CapturedExchangeV2{
		Version:       "", // Keep empty to indicate V1 origin
		Timestamp:     v1.Timestamp,
		Operation:     v1.Operation,
		OperationName: v1.OperationName,
		Endpoint:      v1.Endpoint,
		RequestBody:   v1.RequestBody,
		ResponseBody:  v1.ResponseBody,
		StatusCode:    v1.StatusCode,
		Error:         v1.Error,
		Success:       v1.StatusCode >= 200 && v1.StatusCode < 300 && v1.Error == "",
	}
}

// serviceNamespaces maps ONVIF service namespaces to ServiceType.
var serviceNamespaces = map[string]ServiceType{
	"http://www.onvif.org/ver10/device/wsdl":   ServiceDevice,
	"http://www.onvif.org/ver10/media/wsdl":    ServiceMedia,
	"http://www.onvif.org/ver20/media/wsdl":    ServiceMedia,
	"http://www.onvif.org/ver20/ptz/wsdl":      ServicePTZ,
	"http://www.onvif.org/ver10/ptz/wsdl":      ServicePTZ,
	"http://www.onvif.org/ver20/imaging/wsdl":  ServiceImaging,
	"http://www.onvif.org/ver10/imaging/wsdl":  ServiceImaging,
	"http://www.onvif.org/ver10/events/wsdl":   ServiceEvent,
	"http://www.onvif.org/ver10/deviceIO/wsdl": ServiceDeviceIO,
}

// DetermineServiceType determines the service type from a SOAP request body.
func DetermineServiceType(soapBody string) ServiceType {
	for ns, svc := range serviceNamespaces {
		if containsString(soapBody, ns) {
			return svc
		}
	}
	return ServiceUnknown
}

// containsString is a simple string contains check.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findString(s, substr) >= 0
}

// findString finds substr in s, returns -1 if not found.
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
