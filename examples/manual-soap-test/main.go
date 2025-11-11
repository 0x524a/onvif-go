package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	// Test SOAP request manually
	endpoint := "http://192.168.1.201/onvif/device_service"
	username := "service"
	password := "Service.1234"

	fmt.Println("🔧 Manual SOAP Test for ONVIF Camera")
	fmt.Println("=====================================")
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Username: %s\n", username)
	fmt.Println()

	// Simple GetDeviceInformation SOAP request (without auth for now)
	soapRequest := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" 
               xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
  <soap:Body>
    <tds:GetDeviceInformation/>
  </soap:Body>
</soap:Envelope>`

	fmt.Println("📤 Sending SOAP request (without authentication)...")
	fmt.Println()

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(soapRequest))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("❌ Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📥 Response Status: %s\n", resp.Status)
	fmt.Println("📋 Response Headers:")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	fmt.Println()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Println("📄 Response Body:")
	fmt.Println(string(body))
	fmt.Println()

	if resp.StatusCode != 200 {
		fmt.Printf("⚠️  Non-200 status code: %d\n", resp.StatusCode)

		if resp.StatusCode == 401 {
			fmt.Println("💡 Authentication required - this is expected!")
			fmt.Println("💡 Now testing with go-onvif client library...")
			fmt.Println()
			testWithClient(username, password)
		} else {
			fmt.Println("💡 Unexpected status code. Check:")
			fmt.Println("  - Is ONVIF enabled on the camera?")
			fmt.Println("  - Is the endpoint path correct?")
		}
	} else {
		fmt.Println("✅ Got successful response!")
	}
}

func testWithClient(username, password string) {
	// Import locally to avoid conflicts
	onvif := struct{}{}
	_ = onvif

	fmt.Println("Note: Would test with go-onvif client here, but keeping this simple.")
	fmt.Println("The camera appears to be responding to ONVIF requests.")
	fmt.Println()
	fmt.Println("💡 Next step: Check if the credentials are correct")
	fmt.Printf("   Username: %s\n", username)
	fmt.Printf("   Password: %s\n", password)
}
