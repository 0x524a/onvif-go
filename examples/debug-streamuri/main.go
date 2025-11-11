package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// SOAP Envelope structures
type Envelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  *Header  `xml:"http://www.w3.org/2003/05/soap-envelope Header,omitempty"`
	Body    Body     `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

type Header struct {
	Security *Security `xml:"Security,omitempty"`
}

type Body struct {
	Content interface{} `xml:",omitempty"`
}

type Security struct {
	XMLName        xml.Name       `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd Security"`
	MustUnderstand string         `xml:"http://www.w3.org/2003/05/soap-envelope mustUnderstand,attr,omitempty"`
	UsernameToken  *UsernameToken `xml:"UsernameToken,omitempty"`
}

type UsernameToken struct {
	XMLName  xml.Name `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd UsernameToken"`
	Username string   `xml:"Username"`
	Password Password `xml:"Password"`
	Nonce    Nonce    `xml:"Nonce"`
	Created  string   `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd Created"`
}

type Password struct {
	Type     string `xml:"Type,attr"`
	Password string `xml:",chardata"`
}

type Nonce struct {
	Type  string `xml:"EncodingType,attr"`
	Nonce string `xml:",chardata"`
}

type GetStreamUri struct {
	XMLName     xml.Name `xml:"trt:GetStreamUri"`
	Xmlns       string   `xml:"xmlns:trt,attr"`
	Xmlnst      string   `xml:"xmlns:tt,attr"`
	StreamSetup struct {
		Stream    string `xml:"tt:Stream"`
		Transport struct {
			Protocol string `xml:"tt:Protocol"`
		} `xml:"tt:Transport"`
	} `xml:"trt:StreamSetup"`
	ProfileToken string `xml:"trt:ProfileToken"`
}

func createSecurityHeader(username, password string) *Security {
	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)

	created := time.Now().UTC().Format(time.RFC3339)

	hash := sha1.New()
	hash.Write(nonceBytes)
	hash.Write([]byte(created))
	hash.Write([]byte(password))
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return &Security{
		MustUnderstand: "1",
		UsernameToken: &UsernameToken{
			Username: username,
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

func main() {
	// Using the media service endpoint
	endpoint := "http://192.168.1.201/onvif/media_service"
	username := "service"
	password := "Service.1234"
	profileToken := "0"

	fmt.Println("Testing GetStreamUri SOAP request...")

	// Build request
	req := GetStreamUri{
		Xmlns:        "http://www.onvif.org/ver10/media/wsdl",
		Xmlnst:       "http://www.onvif.org/ver10/schema",
		ProfileToken: profileToken,
	}
	req.StreamSetup.Stream = "RTP-Unicast"
	req.StreamSetup.Transport.Protocol = "RTSP"

	envelope := &Envelope{
		Header: &Header{
			Security: createSecurityHeader(username, password),
		},
		Body: Body{
			Content: req,
		},
	}

	// Marshal to XML
	body, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal: %v", err)
	}

	xmlBody := append([]byte(xml.Header), body...)

	fmt.Println("\n=== Request XML ===")
	fmt.Println(string(xmlBody))

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", endpoint, bytes.NewReader(xmlBody))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("\n=== HTTP Status: %d ===\n", resp.StatusCode)
	fmt.Printf("\n=== Response Body ===\n")
	fmt.Println(string(respBody))
}
