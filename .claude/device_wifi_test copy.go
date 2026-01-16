package onvif

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockDeviceWiFiServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")

		// Parse request to determine which operation
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		requestBody := string(buf)

		var response string

		switch {
		case strings.Contains(requestBody, "GetDot11Capabilities"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetDot11CapabilitiesResponse>
      <tds:Capabilities>
        <tt:TKIP>true</tt:TKIP>
        <tt:ScanAvailableNetworks>true</tt:ScanAvailableNetworks>
        <tt:MultipleConfiguration>false</tt:MultipleConfiguration>
        <tt:AdHocStationMode>false</tt:AdHocStationMode>
        <tt:WEP>false</tt:WEP>
      </tds:Capabilities>
    </tds:GetDot11CapabilitiesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetDot11Status"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetDot11StatusResponse>
      <tds:Status>
        <tt:SSID>TestNetwork</tt:SSID>
        <tt:BSSID>00:11:22:33:44:55</tt:BSSID>
        <tt:PairCipher>CCMP</tt:PairCipher>
        <tt:GroupCipher>CCMP</tt:GroupCipher>
        <tt:SignalStrength>Good</tt:SignalStrength>
        <tt:ActiveConfigAlias>dot11-config-001</tt:ActiveConfigAlias>
      </tds:Status>
    </tds:GetDot11StatusResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetDot1XConfiguration") && !strings.Contains(requestBody, "GetDot1XConfigurations"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetDot1XConfigurationResponse>
      <tds:Dot1XConfiguration token="dot1x-config-001">
        <tt:Dot1XConfigurationToken>dot1x-config-001</tt:Dot1XConfigurationToken>
        <tt:Identity>device@example.com</tt:Identity>
      </tds:Dot1XConfiguration>
    </tds:GetDot1XConfigurationResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetDot1XConfigurations"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetDot1XConfigurationsResponse>
      <tds:Dot1XConfiguration token="dot1x-config-001">
        <tt:Dot1XConfigurationToken>dot1x-config-001</tt:Dot1XConfigurationToken>
        <tt:Identity>device1@example.com</tt:Identity>
      </tds:Dot1XConfiguration>
      <tds:Dot1XConfiguration token="dot1x-config-002">
        <tt:Dot1XConfigurationToken>dot1x-config-002</tt:Dot1XConfigurationToken>
        <tt:Identity>device2@example.com</tt:Identity>
      </tds:Dot1XConfiguration>
    </tds:GetDot1XConfigurationsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "SetDot1XConfiguration"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetDot1XConfigurationResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "CreateDot1XConfiguration"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:CreateDot1XConfigurationResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "DeleteDot1XConfiguration"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:DeleteDot1XConfigurationResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "ScanAvailableDot11Networks"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:ScanAvailableDot11NetworksResponse>
      <tds:Networks>
        <tt:SSID>Network1</tt:SSID>
        <tt:BSSID>00:11:22:33:44:55</tt:BSSID>
        <tt:AuthAndMangementSuite>PSK</tt:AuthAndMangementSuite>
        <tt:PairCipher>CCMP</tt:PairCipher>
        <tt:GroupCipher>CCMP</tt:GroupCipher>
        <tt:SignalStrength>Very Good</tt:SignalStrength>
      </tds:Networks>
      <tds:Networks>
        <tt:SSID>Network2</tt:SSID>
        <tt:BSSID>AA:BB:CC:DD:EE:FF</tt:BSSID>
        <tt:AuthAndMangementSuite>Dot1X</tt:AuthAndMangementSuite>
        <tt:PairCipher>CCMP</tt:PairCipher>
        <tt:GroupCipher>CCMP</tt:GroupCipher>
        <tt:SignalStrength>Good</tt:SignalStrength>
      </tds:Networks>
    </tds:ScanAvailableDot11NetworksResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		default:
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <SOAP-ENV:Fault>
      <SOAP-ENV:Code><SOAP-ENV:Value>SOAP-ENV:Receiver</SOAP-ENV:Value></SOAP-ENV:Code>
      <SOAP-ENV:Reason><SOAP-ENV:Text>Unknown operation</SOAP-ENV:Text></SOAP-ENV:Reason>
    </SOAP-ENV:Fault>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
		}

		_, _ = w.Write([]byte(response))
	}))
}

func TestGetDot11Capabilities(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	caps, err := client.GetDot11Capabilities(ctx)
	if err != nil {
		t.Fatalf("GetDot11Capabilities failed: %v", err)
	}

	if !caps.TKIP {
		t.Error("Expected TKIP to be supported")
	}

	if !caps.ScanAvailableNetworks {
		t.Error("Expected ScanAvailableNetworks to be supported")
	}

	if caps.MultipleConfiguration {
		t.Error("Expected MultipleConfiguration to be false")
	}

	if caps.WEP {
		t.Error("Expected WEP to be false")
	}
}

func TestGetDot11Status(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	status, err := client.GetDot11Status(ctx, "wifi0")
	if err != nil {
		t.Fatalf("GetDot11Status failed: %v", err)
	}

	if status.SSID != "TestNetwork" {
		t.Errorf("Expected SSID 'TestNetwork', got '%s'", status.SSID)
	}

	if status.BSSID != "00:11:22:33:44:55" {
		t.Errorf("Expected BSSID '00:11:22:33:44:55', got '%s'", status.BSSID)
	}

	if status.PairCipher != Dot11CipherCCMP {
		t.Errorf("Expected PairCipher 'CCMP', got '%s'", status.PairCipher)
	}

	if status.GroupCipher != Dot11CipherCCMP {
		t.Errorf("Expected GroupCipher 'CCMP', got '%s'", status.GroupCipher)
	}

	if status.SignalStrength != Dot11SignalGood {
		t.Errorf("Expected SignalStrength 'Good', got '%s'", status.SignalStrength)
	}

	if status.ActiveConfigAlias != "dot11-config-001" {
		t.Errorf("Expected ActiveConfigAlias 'dot11-config-001', got '%s'", status.ActiveConfigAlias)
	}
}

func TestGetDot1XConfiguration(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	config, err := client.GetDot1XConfiguration(ctx, "dot1x-config-001")
	if err != nil {
		t.Fatalf("GetDot1XConfiguration failed: %v", err)
	}

	if config.Dot1XConfigurationToken != "dot1x-config-001" {
		t.Errorf("Expected Dot1XConfigurationToken 'dot1x-config-001', got '%s'", config.Dot1XConfigurationToken)
	}

	if config.Identity != "device@example.com" {
		t.Errorf("Expected Identity 'device@example.com', got '%s'", config.Identity)
	}
}

func TestGetDot1XConfigurations(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	configs, err := client.GetDot1XConfigurations(ctx)
	if err != nil {
		t.Fatalf("GetDot1XConfigurations failed: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("Expected 2 configurations, got %d", len(configs))
	}

	if configs[0].Dot1XConfigurationToken != "dot1x-config-001" {
		t.Errorf("Expected first config token 'dot1x-config-001', got '%s'", configs[0].Dot1XConfigurationToken)
	}

	if configs[0].Identity != "device1@example.com" {
		t.Errorf("Expected first identity 'device1@example.com', got '%s'", configs[0].Identity)
	}

	if configs[1].Dot1XConfigurationToken != "dot1x-config-002" {
		t.Errorf("Expected second config token 'dot1x-config-002', got '%s'", configs[1].Dot1XConfigurationToken)
	}

	if configs[1].Identity != "device2@example.com" {
		t.Errorf("Expected second identity 'device2@example.com', got '%s'", configs[1].Identity)
	}
}

func TestSetDot1XConfiguration(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	config := &Dot1XConfiguration{
		Dot1XConfigurationToken: "dot1x-config-001",
		Identity:                "updated@example.com",
	}

	err = client.SetDot1XConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetDot1XConfiguration failed: %v", err)
	}
}

func TestCreateDot1XConfiguration(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	config := &Dot1XConfiguration{
		Dot1XConfigurationToken: "dot1x-config-new",
		Identity:                "new@example.com",
	}

	err = client.CreateDot1XConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("CreateDot1XConfiguration failed: %v", err)
	}
}

func TestDeleteDot1XConfiguration(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	err = client.DeleteDot1XConfiguration(ctx, "dot1x-config-001")
	if err != nil {
		t.Fatalf("DeleteDot1XConfiguration failed: %v", err)
	}
}

func TestScanAvailableDot11Networks(t *testing.T) {
	server := newMockDeviceWiFiServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	networks, err := client.ScanAvailableDot11Networks(ctx, "wifi0")
	if err != nil {
		t.Fatalf("ScanAvailableDot11Networks failed: %v", err)
	}

	if len(networks) != 2 {
		t.Fatalf("Expected 2 networks, got %d", len(networks))
	}

	// Test first network
	if networks[0].SSID != "Network1" {
		t.Errorf("Expected first SSID 'Network1', got '%s'", networks[0].SSID)
	}

	if networks[0].BSSID != "00:11:22:33:44:55" {
		t.Errorf("Expected first BSSID '00:11:22:33:44:55', got '%s'", networks[0].BSSID)
	}

	if len(networks[0].AuthAndMangementSuite) == 0 || networks[0].AuthAndMangementSuite[0] != Dot11AuthPSK {
		t.Errorf("Expected first auth suite 'PSK'")
	}

	if len(networks[0].PairCipher) == 0 || networks[0].PairCipher[0] != Dot11CipherCCMP {
		t.Errorf("Expected first pair cipher 'CCMP'")
	}

	if networks[0].SignalStrength != Dot11SignalVeryGood {
		t.Errorf("Expected first signal strength 'VeryGood', got '%s'", networks[0].SignalStrength)
	}

	// Test second network
	if networks[1].SSID != "Network2" {
		t.Errorf("Expected second SSID 'Network2', got '%s'", networks[1].SSID)
	}

	if networks[1].BSSID != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Expected second BSSID 'AA:BB:CC:DD:EE:FF', got '%s'", networks[1].BSSID)
	}

	if len(networks[1].AuthAndMangementSuite) == 0 || networks[1].AuthAndMangementSuite[0] != Dot11AuthDot1X {
		t.Errorf("Expected second auth suite 'Dot1X'")
	}

	if networks[1].SignalStrength != Dot11SignalGood {
		t.Errorf("Expected second signal strength 'Good', got '%s'", networks[1].SignalStrength)
	}
}
