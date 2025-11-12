package soap

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	originsoap "github.com/0x524a/onvif-go/internal/soap"
)

// Handler handles incoming SOAP requests
type Handler struct {
	username string
	password string
	handlers map[string]MessageHandler
}

// MessageHandler is a function that handles a specific SOAP message
type MessageHandler func(body interface{}) (interface{}, error)

// NewHandler creates a new SOAP handler
func NewHandler(username, password string) *Handler {
	return &Handler{
		username: username,
		password: password,
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler registers a handler for a specific action/message type
func (h *Handler) RegisterHandler(action string, handler MessageHandler) {
	h.handlers[action] = handler
}

// ServeHTTP implements http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.sendFault(w, "Receiver", "Failed to read request body", err.Error())
		return
	}
	_ = r.Body.Close()

	// Extract action from raw XML first (before parsing)
	action := h.extractAction(body)
	if action == "" {
		h.sendFault(w, "Sender", "Unknown action", "Could not determine request action")
		return
	}

	// Parse SOAP envelope
	var envelope originsoap.Envelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		h.sendFault(w, "Sender", "Invalid SOAP envelope", err.Error())
		return
	}

	// Authenticate if credentials are configured
	if h.username != "" && h.password != "" {
		if !h.authenticate(&envelope) {
			h.sendFault(w, "Sender", "Authentication failed", "Invalid username or password")
			return
		}
	}

	// Find and execute handler
	handler, ok := h.handlers[action]
	if !ok {
		h.sendFault(w, "Receiver", "Action not supported", fmt.Sprintf("No handler for action: %s", action))
		return
	}

	// Execute handler
	response, err := handler(envelope.Body.Content)
	if err != nil {
		h.sendFault(w, "Receiver", "Handler error", err.Error())
		return
	}

	// Send response
	h.sendResponse(w, response)
}

// authenticate verifies the WS-Security credentials
func (h *Handler) authenticate(envelope *originsoap.Envelope) bool {
	if envelope.Header == nil || envelope.Header.Security == nil || envelope.Header.Security.UsernameToken == nil {
		return false
	}

	token := envelope.Header.Security.UsernameToken

	// Check username
	if token.Username != h.username {
		return false
	}

	// Decode nonce
	nonce, err := base64.StdEncoding.DecodeString(token.Nonce.Nonce)
	if err != nil {
		return false
	}

	// Calculate expected digest
	hash := sha1.New()
	hash.Write(nonce)
	hash.Write([]byte(token.Created))
	hash.Write([]byte(h.password))
	expectedDigest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Compare digests
	return token.Password.Password == expectedDigest
}

// extractAction extracts the action/message type from the SOAP body
func (h *Handler) extractAction(bodyXML []byte) string {
	// Parse XML to find the first element inside the Body element
	decoder := xml.NewDecoder(bytes.NewReader(bodyXML))
	inBody := false
	depth := 0
	
	for {
		token, err := decoder.Token()
		if err != nil {
			return ""
		}

		switch t := token.(type) {
		case xml.StartElement:
			depth++
			// Check if we're entering the Body element
			if t.Name.Local == "Body" {
				inBody = true
			} else if inBody && depth > 2 {
				// Found the first element inside Body
				return t.Name.Local
			}
		case xml.EndElement:
			depth--
			if t.Name.Local == "Body" {
				inBody = false
			}
		}
	}
}

// sendResponse sends a SOAP response
func (h *Handler) sendResponse(w http.ResponseWriter, response interface{}) {
	envelope := &originsoap.Envelope{
		Body: originsoap.Body{
			Content: response,
		},
	}

	// Marshal to XML
	body, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		h.sendFault(w, "Receiver", "Failed to marshal response", err.Error())
		return
	}

	// Add XML declaration
	xmlBody := append([]byte(xml.Header), body...)

	// Send response
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(xmlBody)
}

// sendFault sends a SOAP fault response
func (h *Handler) sendFault(w http.ResponseWriter, code, reason, detail string) {
	fault := &originsoap.Fault{
		Code:   code,
		Reason: reason,
		Detail: detail,
	}

	envelope := &originsoap.Envelope{
		Body: originsoap.Body{
			Fault: fault,
		},
	}

	// Marshal to XML
	body, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add XML declaration
	xmlBody := append([]byte(xml.Header), body...)

	// Send fault response
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(xmlBody)
}

// RequestWrapper wraps incoming SOAP request structures
type RequestWrapper struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
}

// ParseRequest parses a SOAP request into a specific structure
func ParseRequest(bodyContent interface{}, target interface{}) error {
	// Marshal the body content back to XML
	bodyXML, err := xml.Marshal(bodyContent)
	if err != nil {
		return fmt.Errorf("failed to marshal body content: %w", err)
	}

	// Unmarshal into target structure
	if err := xml.Unmarshal(bodyXML, target); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return nil
}

// Common SOAP request/response structures for ONVIF

// GetSystemDateAndTimeRequest represents GetSystemDateAndTime request
type GetSystemDateAndTimeRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetSystemDateAndTime"`
}

// GetSystemDateAndTimeResponse represents GetSystemDateAndTime response
type GetSystemDateAndTimeResponse struct {
	XMLName        xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetSystemDateAndTimeResponse"`
	SystemDateAndTime SystemDateAndTime `xml:"SystemDateAndTime"`
}

// SystemDateAndTime represents system date and time
type SystemDateAndTime struct {
	DateTimeType    string    `xml:"DateTimeType"`
	DaylightSavings bool      `xml:"DaylightSavings"`
	TimeZone        TimeZone  `xml:"TimeZone,omitempty"`
	UTCDateTime     DateTime  `xml:"UTCDateTime,omitempty"`
	LocalDateTime   DateTime  `xml:"LocalDateTime,omitempty"`
}

// TimeZone represents timezone information
type TimeZone struct {
	TZ string `xml:"TZ"`
}

// DateTime represents date and time
type DateTime struct {
	Time Time `xml:"Time"`
	Date Date `xml:"Date"`
}

// Time represents time components
type Time struct {
	Hour   int `xml:"Hour"`
	Minute int `xml:"Minute"`
	Second int `xml:"Second"`
}

// Date represents date components
type Date struct {
	Year  int `xml:"Year"`
	Month int `xml:"Month"`
	Day   int `xml:"Day"`
}

// ToDateTime converts time.Time to DateTime structure
func ToDateTime(t time.Time) DateTime {
	return DateTime{
		Date: Date{
			Year:  t.Year(),
			Month: int(t.Month()),
			Day:   t.Day(),
		},
		Time: Time{
			Hour:   t.Hour(),
			Minute: t.Minute(),
			Second: t.Second(),
		},
	}
}

// GetCapabilitiesRequest represents GetCapabilities request
type GetCapabilitiesRequest struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetCapabilities"`
	Category []string `xml:"Category,omitempty"`
}

// GetDeviceInformationRequest represents GetDeviceInformation request
type GetDeviceInformationRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetDeviceInformation"`
}

// GetServicesRequest represents GetServices request
type GetServicesRequest struct {
	XMLName           xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetServices"`
	IncludeCapability bool     `xml:"IncludeCapability"`
}

// GetProfilesRequest represents GetProfiles request
type GetProfilesRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetProfiles"`
}

// GetStreamURIRequest represents GetStreamURI request
type GetStreamURIRequest struct {
	XMLName      xml.Name    `xml:"http://www.onvif.org/ver10/media/wsdl GetStreamURI"`
	StreamSetup  StreamSetup `xml:"StreamSetup"`
	ProfileToken string      `xml:"ProfileToken"`
}

// StreamSetup represents stream setup parameters
type StreamSetup struct {
	Stream    string    `xml:"Stream"`
	Transport Transport `xml:"Transport"`
}

// Transport represents transport parameters
type Transport struct {
	Protocol string `xml:"Protocol"`
}

// GetSnapshotURIRequest represents GetSnapshotURI request
type GetSnapshotURIRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetSnapshotURI"`
	ProfileToken string   `xml:"ProfileToken"`
}

// NormalizeAction normalizes SOAP action names
func NormalizeAction(action string) string {
	// Remove namespace prefixes
	if idx := strings.LastIndex(action, ":"); idx != -1 {
		action = action[idx+1:]
	}
	return action
}
