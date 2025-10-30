package soap

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Envelope represents a SOAP envelope
type Envelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  *Header  `xml:"http://www.w3.org/2003/05/soap-envelope Header,omitempty"`
	Body    Body     `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

// Header represents a SOAP header
type Header struct {
	Security *Security `xml:"Security,omitempty"`
}

// Body represents a SOAP body
type Body struct {
	Content interface{} `xml:",omitempty"`
	Fault   *Fault      `xml:"Fault,omitempty"`
}

// Fault represents a SOAP fault
type Fault struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Fault"`
	Code    string   `xml:"Code>Value"`
	Reason  string   `xml:"Reason>Text"`
	Detail  string   `xml:"Detail,omitempty"`
}

// Security represents WS-Security header
type Security struct {
	XMLName   xml.Name       `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd Security"`
	MustUnderstand string    `xml:"http://www.w3.org/2003/05/soap-envelope mustUnderstand,attr,omitempty"`
	UsernameToken *UsernameToken `xml:"UsernameToken,omitempty"`
}

// UsernameToken represents a WS-Security username token
type UsernameToken struct {
	XMLName   xml.Name `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd UsernameToken"`
	Username  string   `xml:"Username"`
	Password  Password `xml:"Password"`
	Nonce     Nonce    `xml:"Nonce"`
	Created   string   `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd Created"`
}

// Password represents a WS-Security password
type Password struct {
	Type     string `xml:"Type,attr"`
	Password string `xml:",chardata"`
}

// Nonce represents a WS-Security nonce
type Nonce struct {
	Type  string `xml:"EncodingType,attr"`
	Nonce string `xml:",chardata"`
}

// Client represents a SOAP client
type Client struct {
	httpClient *http.Client
	username   string
	password   string
}

// NewClient creates a new SOAP client
func NewClient(httpClient *http.Client, username, password string) *Client {
	return &Client{
		httpClient: httpClient,
		username:   username,
		password:   password,
	}
}

// Call makes a SOAP call to the specified endpoint
func (c *Client) Call(ctx context.Context, endpoint string, action string, request interface{}, response interface{}) error {
	// Build SOAP envelope
	envelope := &Envelope{
		Body: Body{
			Content: request,
		},
	}

	// Add security header if credentials are provided
	if c.username != "" && c.password != "" {
		envelope.Header = &Header{
			Security: c.createSecurityHeader(),
		}
	}

	// Marshal envelope to XML
	body, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	// Add XML declaration
	xmlBody := append([]byte(xml.Header), body...)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(xmlBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	if action != "" {
		req.Header.Set("SOAPAction", action)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var respEnvelope Envelope
	if err := xml.Unmarshal(respBody, &respEnvelope); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP response: %w", err)
	}

	// Check for SOAP fault
	if respEnvelope.Body.Fault != nil {
		return fmt.Errorf("SOAP fault: [%s] %s - %s",
			respEnvelope.Body.Fault.Code,
			respEnvelope.Body.Fault.Reason,
			respEnvelope.Body.Fault.Detail)
	}

	// Unmarshal response content
	if response != nil {
		// Re-marshal the body content and unmarshal into the response struct
		bodyXML, err := xml.Marshal(respEnvelope.Body.Content)
		if err != nil {
			return fmt.Errorf("failed to marshal response body: %w", err)
		}
		if err := xml.Unmarshal(bodyXML, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// createSecurityHeader creates a WS-Security header with username token digest
func (c *Client) createSecurityHeader() *Security {
	// Generate nonce
	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)

	// Get current timestamp
	created := time.Now().UTC().Format(time.RFC3339)

	// Calculate password digest: Base64(SHA1(nonce + created + password))
	hash := sha1.New()
	hash.Write(nonceBytes)
	hash.Write([]byte(created))
	hash.Write([]byte(c.password))
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return &Security{
		MustUnderstand: "1",
		UsernameToken: &UsernameToken{
			Username: c.username,
			Password: Password{
				Type:     "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest",
				Password: digest,
			},
			Nonce: Nonce{
				Type:  "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary",
				Nonce: nonce,
			},
			Created: created,
		},
	}
}

// BuildEnvelope builds a SOAP envelope with the given body content
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
