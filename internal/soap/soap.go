// Package soap provides SOAP client functionality for ONVIF communication.
package soap

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // SHA1 used for ONVIF digest authentication
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// AuthMode selects how requests are authenticated.
type AuthMode int

const (
	// AuthModeDigest is the ONVIF default: WS-Security UsernameToken with
	// PasswordDigest (Base64(SHA1(nonce + created + password))).
	AuthModeDigest AuthMode = iota
	// AuthModePasswordText sends the password in cleartext inside the
	// WS-Security UsernameToken (PasswordText profile). Some firmwares
	// reject digest on specific services while accepting this.
	AuthModePasswordText
	// AuthModeHTTPBasic sends no WS-Security header and authenticates with
	// an HTTP Basic Authorization header instead.
	AuthModeHTTPBasic
	// AuthModeNone sends no authentication at all. Minimal embedded devices
	// (e.g. ESP32 firmwares) may reject every auth-bearing request.
	AuthModeNone
)

// Password type URIs from the WS-Security UsernameToken profile.
const (
	passwordDigestType = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest" //nolint:lll // Long XML namespace
	passwordTextType   = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText"   //nolint:lll // Long XML namespace
	nonceEncodingType  = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"    //nolint:lll // Long XML namespace
)

// Exported aliases for consumers inside this module (the server-side
// UsernameToken validator matches on these URIs).
const (
	// PasswordDigestType marks a WS-Security Password element carrying
	// Base64(SHA1(nonce + created + password)).
	PasswordDigestType = passwordDigestType
	// PasswordTextType marks a WS-Security Password element carrying the
	// cleartext password.
	PasswordTextType = passwordTextType
)

// Envelope represents a SOAP envelope.
type Envelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  *Header  `xml:"http://www.w3.org/2003/05/soap-envelope Header,omitempty"`
	Body    Body     `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

// Header represents a SOAP header.
type Header struct {
	Security *Security `xml:"Security,omitempty"`
}

// Body represents a SOAP body.
type Body struct {
	Content interface{} `xml:",omitempty"`
	Fault   *Fault      `xml:"Fault,omitempty"`
}

// Fault represents a SOAP fault.
type Fault struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Fault"`
	Code    string   `xml:"Code>Value"`
	Reason  string   `xml:"Reason>Text"`
	Detail  string   `xml:"Detail,omitempty"`
}

// Security represents WS-Security header.
type Security struct {
	XMLName        xml.Name       `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd Security"` //nolint:lll // Long XML namespace
	MustUnderstand string         `xml:"http://www.w3.org/2003/05/soap-envelope mustUnderstand,attr,omitempty"`
	UsernameToken  *UsernameToken `xml:"UsernameToken,omitempty"`
}

// UsernameToken represents a WS-Security username token.
type UsernameToken struct {
	XMLName  xml.Name `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd UsernameToken"` //nolint:lll // Long XML namespace
	Username string   `xml:"Username"`
	Password Password `xml:"Password"`
	Nonce    Nonce    `xml:"Nonce"`
	Created  string   `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-utility-1.0.xsd Created"`

	// CreatedVariant captures wsu:Created sent under the common
	// misspelled utility namespace (oasis-200401-wss-wssecurity-utility-…)
	// used by many community clients; encoding/xml namespace-qualifies
	// field matching, so the variant needs its own field. Emission always
	// uses the canonical Created above (omitempty keeps this one off the
	// wire when empty).
	CreatedVariant string `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd Created,omitempty"` //nolint:lll // Long XML namespace
}

// CreatedValue returns the token's creation timestamp, accepting either
// the canonical or the misspelled utility namespace on decode.
func (t *UsernameToken) CreatedValue() string {
	if t.Created != "" {
		return t.Created
	}

	return t.CreatedVariant
}

// Password represents a WS-Security password.
type Password struct {
	Type     string `xml:"Type,attr"`
	Password string `xml:",chardata"`
}

// Nonce represents a WS-Security nonce.
type Nonce struct {
	Type  string `xml:"EncodingType,attr"`
	Nonce string `xml:",chardata"`
}

// Client represents a SOAP client.
type Client struct {
	httpClient *http.Client
	username   string
	password   string
	authMode   AuthMode
	debug      bool
	logger     func(format string, args ...interface{})
	// clockSkew is the offset (deviceTime - localTime) applied to the Created
	// timestamp in WS-Security UsernameToken digest auth. ONVIF cameras (notably
	// Hikvision) reject digests whose Created timestamp falls outside their
	// replay window (commonly ±5 min). When the NVR's clock and the camera's
	// clock diverge, every digest is rejected as "sender not authorized" — a
	// generic-looking auth failure. Setting the skew (measured via the
	// unauthenticated GetSystemDateAndTime) makes the digest use the device's
	// view of "now". Default 0 = use local clock (legacy behavior).
	clockSkew time.Duration

	// slogger receives structured request/response/fault summaries
	// (issue #62). nil = silent, the default.
	slogger *slog.Logger
}

// NewClient creates a new SOAP client.
func NewClient(httpClient *http.Client, username, password string) *Client {
	return &Client{
		httpClient: httpClient,
		username:   username,
		password:   password,
		authMode:   AuthModeDigest,
		debug:      false,
		logger:     nil,
	}
}

// SetAuthMode selects the authentication mode used for subsequent calls.
// The zero value AuthModeDigest preserves the historical behavior.
func (c *Client) SetAuthMode(mode AuthMode) {
	c.authMode = mode
}

// SetDebug enables debug logging with a custom logger.
func (c *Client) SetDebug(enabled bool, logger func(format string, args ...interface{})) {
	c.debug = enabled
	c.logger = logger
}

// SetLogger installs a structured logger for request/response/fault
// summaries and clock-skew correction notes (issue #62). nil restores the
// silent default. It is independent of the legacy SetDebug printf hook.
func (c *Client) SetLogger(l *slog.Logger) {
	c.slogger = l
}

// SetClockSkew sets the clock offset (deviceTime - localTime) used when building
// WS-Security UsernameToken digest timestamps. Call this after measuring the
// device's clock via GetSystemDateAndTime to fix auth failures caused by clock
// divergence (Hikvision time-skew issue). Pass 0 to use the local clock.
func (c *Client) SetClockSkew(skew time.Duration) {
	c.clockSkew = skew
}

// authModeName renders the auth mode for log summaries.
func authModeName(m AuthMode) string {
	switch m {
	case AuthModeDigest:
		return "digest"
	case AuthModePasswordText:
		return "password-text"
	case AuthModeNone:
		return "none"
	default:
		return "unknown"
	}
}

// logDebugf logs debug information if debug mode is enabled.
func (c *Client) logDebugf(format string, args ...interface{}) {
	if c.debug && c.logger != nil {
		c.logger(format, args...)
	}
}

// Call makes a SOAP call to the specified endpoint.
//
// Faults are detected regardless of the HTTP status: ONVIF devices frequently
// return SOAP Faults (including NotAuthorized) with HTTP 200, and before fault
// detection existed such responses were mistaken for successful — or worse,
// silently dropped — operation results.
func (c *Client) Call(ctx context.Context, endpoint, action string, request, response interface{}) error {
	// Build SOAP envelope
	envelope := &Envelope{
		Body: Body{
			Content: request,
		},
	}

	// Add the WS-Security header for the header-based modes when credentials
	// are present (HTTP Basic and None need no SOAP header).
	if c.username != "" && c.password != "" &&
		(c.authMode == AuthModeDigest || c.authMode == AuthModePasswordText) {
		envelope.Header = &Header{
			Security: c.createSecurityHeader(),
		}
	}

	// Marshal envelope to XML in the canonical compact wire form (see
	// renderEnvelope for why encoding/xml's default output is not used).
	body, err := renderEnvelope(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	// Add XML declaration
	xmlBody := append([]byte(xml.Header), body...)

	// Log request if debug is enabled
	c.logDebugf("=== SOAP Request ===\nEndpoint: %s\nAction: %s\n%s\n", endpoint, action, string(xmlBody))
	if c.slogger != nil {
		c.slogger.Debug("soap request", "action", action, "endpoint", endpoint, "body_bytes", len(xmlBody), "auth_mode", authModeName(c.authMode))
		if c.clockSkew != 0 {
			c.slogger.Debug("clock skew correction applied to UsernameToken timestamps", "skew_ms", c.clockSkew.Milliseconds())
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(xmlBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	if action != "" {
		req.Header.Set("SOAPAction", action)
	}

	if c.authMode == AuthModeHTTPBasic && c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body, bounded: a hostile or broken device must not be
	// able to exhaust the client's memory (issue #59). 1 MiB is ample for
	// real ONVIF responses (profiles, capabilities); larger bodies error.
	const maxResponseBytes = 1 << 20
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if len(respBody) > maxResponseBytes {
		return fmt.Errorf("response body exceeds %d bytes", maxResponseBytes)
	}

	// Log response if debug is enabled
	c.logDebugf("=== SOAP Response ===\nStatus: %d\n%s\n", resp.StatusCode, string(respBody))
	if c.slogger != nil {
		c.slogger.Debug("soap response", "status", resp.StatusCode, "body_bytes", len(respBody), "action", action)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return &HTTPStatusError{Status: resp.StatusCode, Body: string(respBody)}
	}

	// If response is empty, return immediately
	if len(respBody) == 0 {
		return fmt.Errorf("%w", ErrEmptyResponseBody)
	}

	// Extract the raw Body content first: fault detection must happen for
	// every call (including void operations whose response is nil), because a
	// 200-with-Fault must never be reported as success.
	var envelopeResp struct {
		Body struct {
			Content []byte `xml:",innerxml"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(respBody, &envelopeResp); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP envelope: %w", err)
	}

	if fault := parseFault(envelopeResp.Body.Content, resp.StatusCode); fault != nil {
		if c.slogger != nil {
			c.slogger.Warn("soap fault", "action", action, "http_status", resp.StatusCode, "fault_code", fault.Code, "fault_reason", fault.Reason)
		}
		return fault
	}

	// Unmarshal response content if response is provided
	if response != nil {
		if err := xml.Unmarshal(envelopeResp.Body.Content, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// faultXML parses a Fault element by local name so any namespace prefix
// (s:, soapenv:, SOAP-ENV:, ...) and either SOAP version matches.
type faultXML struct {
	XMLName xml.Name `xml:"Fault"`

	// SOAP 1.2 shape.
	Code struct {
		Value   string `xml:"Value"`
		Subcode struct {
			Value string `xml:"Value"`
		} `xml:"Subcode"`
	} `xml:"Code"`
	Reason struct {
		Text string `xml:"Text"`
	} `xml:"Reason"`
	Detail struct {
		Inner string `xml:",innerxml"`
	} `xml:"Detail"`

	// SOAP 1.1 shape.
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

// parseFault detects a SOAP Fault in the raw body content and converts it to a
// FaultError. Returns nil when the content is not a fault: a failed unmarshal
// here simply means the body holds an operation response instead of a Fault
// element. Matching is by local element name, so it tolerates arbitrary
// prefixes and both SOAP 1.1 and 1.2 fault layouts.
func parseFault(bodyContent []byte, httpStatus int) *FaultError {
	var f faultXML
	if err := xml.Unmarshal(bodyContent, &f); err != nil {
		return nil //nolint:nilerr // unmarshal failure = body is not a fault
	}

	fault := &FaultError{
		Code:       f.Code.Value,
		Subcode:    f.Code.Subcode.Value,
		Reason:     f.Reason.Text,
		Detail:     f.Detail.Inner,
		HTTPStatus: httpStatus,
	}

	// Map SOAP 1.1 fields when the 1.2 shape is empty.
	if fault.Code == "" {
		fault.Code = f.FaultCode
	}

	if fault.Reason == "" {
		fault.Reason = f.FaultString
	}

	return fault
}

// createSecurityHeader creates a WS-Security header with a UsernameToken for
// the configured auth mode (digest or cleartext password).
func (c *Client) createSecurityHeader() *Security {
	if c.authMode == AuthModePasswordText {
		return &Security{
			MustUnderstand: "1",
			UsernameToken: &UsernameToken{
				Username: c.username,
				Password: Password{
					Type:     passwordTextType,
					Password: c.password,
				},
			},
		}
	}

	// Generate nonce
	const nonceSize = 16
	nonceBytes := make([]byte, nonceSize)

	_, _ = rand.Read(nonceBytes)
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)

	// Get current timestamp, adjusted by the device's clock skew so the digest's
	// Created field matches the camera's view of "now" (fixes Hikvision time-skew
	// auth rejections). clockSkew is 0 when unset (legacy local-time behavior).
	created := time.Now().UTC().Add(c.clockSkew).Format(time.RFC3339)

	// Calculate password digest: Base64(SHA1(nonce + created + password))
	hash := sha1.New() //nolint:gosec // SHA1 required for ONVIF digest auth
	hash.Write(nonceBytes)
	hash.Write([]byte(created))
	hash.Write([]byte(c.password))
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return &Security{
		MustUnderstand: "1",
		UsernameToken: &UsernameToken{
			Username: c.username,
			Password: Password{
				Type:     passwordDigestType,
				Password: digest,
			},
			Nonce: Nonce{
				Type:  nonceEncodingType,
				Nonce: nonce,
			},
			Created: created,
		},
	}
}

// BuildEnvelope builds a SOAP envelope with the given body content.
func BuildEnvelope(body interface{}, username, password string) (*Envelope, error) {
	envelope := &Envelope{
		Body: Body{
			Content: body,
		},
	}

	if username != "" && password != "" {
		client := &Client{username: username, password: password}
		envelope.Header = &Header{
			Security: client.createSecurityHeader(),
		}
	}

	return envelope, nil
}

// Wire prefixes for the canonical request form. s:/wsse:/wsu: mirror the
// gSOAP output most device firmware is tested against (ONVIF Device Manager
// and friends); mustUnderstand is namespace-qualified per SOAP 1.2.
const (
	soapEnvNamespace   = "http://www.w3.org/2003/05/soap-envelope"
	wsseNamespace      = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"  //nolint:lll // Long XML namespace
	wsuNamespace       = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd" //nolint:lll // Long XML namespace
	soapEnvelopeOpen   = `<s:Envelope xmlns:s="` + soapEnvNamespace + `">`
	soapEnvelopeClose  = "</s:Envelope>"
	securityOpenPrefix = `<s:Header><wsse:Security xmlns:wsse="` + wsseNamespace + `"`
)

// renderEnvelope serializes the envelope into the canonical request wire
// format: s:-prefixed Envelope/Body/Header, compact (no inter-element
// whitespace), and operation namespaces declared on s:Body instead of the
// operation element.
//
// encoding/xml's own output — default-namespace Envelope/Body, indented,
// namespaces on the operation element — is semantically equivalent but is
// rejected by strict XML parsers in old camera firmware, which answer
// HTTP 200 with an empty body ("received empty response body" downstream;
// MiBeeNvr#1074 has the packet captures and the per-characteristic matrix).
// The compact prefixed form is what mainstream gSOAP clients emit and every
// conformant parser accepts, so it is the only shape this client produces.
func renderEnvelope(envelope *Envelope) ([]byte, error) {
	var b strings.Builder

	b.WriteString(soapEnvelopeOpen)
	if envelope.Header != nil && envelope.Header.Security != nil {
		renderSecurityHeader(&b, envelope.Header.Security)
	}
	if envelope.Body.Content != nil {
		fragment, err := xml.Marshal(envelope.Body.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal SOAP body: %w", err)
		}
		inner, decls := hoistNamespaceDecls(fragment)
		b.WriteString("<s:Body")
		for _, d := range decls {
			b.WriteByte(' ')
			b.WriteString(d)
		}
		b.WriteByte('>')
		b.WriteString(inner)
		b.WriteString("</s:Body>")
	} else {
		b.WriteString("<s:Body/>")
	}
	b.WriteString(soapEnvelopeClose)

	return []byte(b.String()), nil
}

// hoistNamespaceDecls strips the xmlns:* declarations from the fragment's
// root element and returns them as complete attribute strings (e.g.
// `xmlns:tds="http://..."`) for re-declaration on an ancestor. Descendant
// elements using those prefixes stay in scope when the caller re-declares
// them on the ancestor. The fragment is encoding/xml output whose root
// attribute values are namespace URLs, which never contain double quotes,
// so scanning to the next double quote is unambiguous.
func hoistNamespaceDecls(fragment []byte) (string, []string) {
	s := string(fragment)
	end := strings.IndexByte(s, '>')
	if end < 0 {
		return s, nil // malformed fragment; marshal output always has one
	}
	tag := s[:end]
	var decls []string
	for {
		i := strings.Index(tag, " xmlns:")
		if i < 0 {
			break
		}
		start := i + 1
		open := strings.IndexByte(tag[start:], '"')
		if open < 0 {
			break
		}
		open += start
		closeq := strings.IndexByte(tag[open+1:], '"')
		if closeq < 0 {
			break
		}
		closeq += open + 1
		decls = append(decls, tag[start:closeq+1])
		tag = tag[:i] + tag[closeq+1:]
	}
	return tag + s[end:], decls
}

// renderSecurityHeader writes the WS-Security UsernameToken header with the
// conventional wsse:/wsu: prefixes. Attribute values written here (Type,
// EncodingType) are the fixed profile URIs from this file's constants, which
// contain no characters needing attribute escaping; element text goes through
// xml.EscapeText.
func renderSecurityHeader(b *strings.Builder, sec *Security) {
	b.WriteString(securityOpenPrefix)
	if sec.MustUnderstand != "" {
		b.WriteString(` s:mustUnderstand="`)
		writeXMLText(b, sec.MustUnderstand)
		b.WriteString(`"`)
	}
	b.WriteString(">")

	if ut := sec.UsernameToken; ut != nil {
		b.WriteString("<wsse:UsernameToken>")
		b.WriteString("<wsse:Username>")
		writeXMLText(b, ut.Username)
		b.WriteString("</wsse:Username>")
		b.WriteString(`<wsse:Password Type="`)
		writeXMLText(b, ut.Password.Type)
		b.WriteString(`">`)
		writeXMLText(b, ut.Password.Password)
		b.WriteString("</wsse:Password>")
		if ut.Nonce.Nonce != "" || ut.Nonce.Type != "" {
			b.WriteString(`<wsse:Nonce`)
			if ut.Nonce.Type != "" {
				b.WriteString(` EncodingType="`)
				writeXMLText(b, ut.Nonce.Type)
				b.WriteString(`"`)
			}
			b.WriteString(">")
			writeXMLText(b, ut.Nonce.Nonce)
			b.WriteString("</wsse:Nonce>")
		}
		if created := ut.CreatedValue(); created != "" {
			b.WriteString(`<wsu:Created xmlns:wsu="` + wsuNamespace + `">`)
			writeXMLText(b, created)
			b.WriteString("</wsu:Created>")
		}
		b.WriteString("</wsse:UsernameToken>")
	}

	b.WriteString("</wsse:Security></s:Header>")
}

// writeXMLText appends s to b with XML text escaping.
func writeXMLText(b *strings.Builder, s string) {
	// Error is impossible for strings.Builder writes; EscapeText's only
	// failure mode is the underlying writer failing.
	_ = xml.EscapeText(b, []byte(s)) //nolint:errcheck // strings.Builder writes cannot fail
}
