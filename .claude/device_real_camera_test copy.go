package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test device information from real camera:
// Manufacturer: Bosch
// Model: FLEXIDOME indoor 5100i IR
// Firmware: 8.71.0066
// Serial Number: 404754734001050102
// Hardware ID: F000B543

// TestGetDeviceInformation_Bosch tests GetDeviceInformation with real camera response.
func TestGetDeviceInformation_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Manufacturer>Bosch</tds:Manufacturer>
      <tds:Model>FLEXIDOME indoor 5100i IR</tds:Model>
      <tds:FirmwareVersion>8.71.0066</tds:FirmwareVersion>
      <tds:SerialNumber>404754734001050102</tds:SerialNumber>
      <tds:HardwareId>F000B543</tds:HardwareId>
    </tds:GetDeviceInformationResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetDeviceInformation") {
			t.Errorf("Request should contain GetDeviceInformation, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		t.Fatalf("GetDeviceInformation() failed: %v", err)
	}

	// Validate response matches real camera
	if info.Manufacturer != "Bosch" {
		t.Errorf("Expected Manufacturer=Bosch (Bosch FLEXIDOME), got %s", info.Manufacturer)
	}
	if info.Model != "FLEXIDOME indoor 5100i IR" {
		t.Errorf("Expected Model=FLEXIDOME indoor 5100i IR (Bosch FLEXIDOME), got %s", info.Model)
	}
	if info.FirmwareVersion != "8.71.0066" {
		t.Errorf("Expected FirmwareVersion=8.71.0066 (Bosch FLEXIDOME), got %s", info.FirmwareVersion)
	}
	if info.SerialNumber != "404754734001050102" {
		t.Errorf("Expected SerialNumber=404754734001050102 (Bosch FLEXIDOME), got %s", info.SerialNumber)
	}
	if info.HardwareID != "F000B543" {
		t.Errorf("Expected HardwareID=F000B543 (Bosch FLEXIDOME), got %s", info.HardwareID)
	}
}

// TestGetCapabilities_Bosch tests GetCapabilities with real camera response.
func TestGetCapabilities_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Capabilities>
        <tds:Device>
          <tds:XAddr>http://192.168.1.201/onvif/device_service</tds:XAddr>
          <tds:Network>
            <tt:IPFilter xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:IPFilter>
            <tt:ZeroConfiguration xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:ZeroConfiguration>
            <tt:IPVersion6 xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:IPVersion6>
            <tt:DynDNS xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:DynDNS>
          </tds:Network>
          <tds:System>
            <tt:DiscoveryResolve xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:DiscoveryResolve>
            <tt:DiscoveryBye xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:DiscoveryBye>
            <tt:RemoteDiscovery xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:RemoteDiscovery>
            <tt:SystemBackup xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:SystemBackup>
            <tt:SystemLogging xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:SystemLogging>
            <tt:FirmwareUpgrade xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:FirmwareUpgrade>
            <tt:SupportedVersions xmlns:tt="http://www.onvif.org/ver10/schema">1 2</tt:SupportedVersions>
          </tds:System>
          <tds:IO>
            <tt:InputConnectors xmlns:tt="http://www.onvif.org/ver10/schema">1</tt:InputConnectors>
            <tt:RelayOutputs xmlns:tt="http://www.onvif.org/ver10/schema">1</tt:RelayOutputs>
          </tds:IO>
          <tds:Security>
            <tt:TLS1.1 xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:TLS1.1>
            <tt:TLS1.2 xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:TLS1.2>
            <tt:OnboardKeyGeneration xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:OnboardKeyGeneration>
            <tt:AccessPolicyConfig xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:AccessPolicyConfig>
            <tt:X509Token xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:X509Token>
            <tt:SAMLToken xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:SAMLToken>
            <tt:KerberosToken xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:KerberosToken>
            <tt:RELToken xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:RELToken>
          </tds:Security>
        </tds:Device>
        <tds:Media>
          <tds:XAddr>http://192.168.1.201/onvif/media_service</tds:XAddr>
          <tds:StreamingCapabilities>
            <tt:RTPMulticast xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:RTPMulticast>
            <tt:RTP_TCP xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:RTP_TCP>
            <tt:RTP_RTSP_TCP xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:RTP_RTSP_TCP>
          </tds:StreamingCapabilities>
        </tds:Media>
        <tds:Imaging>
          <tds:XAddr>http://192.168.1.201/onvif/imaging_service</tds:XAddr>
        </tds:Imaging>
        <tds:Events>
          <tds:XAddr>http://192.168.1.201/onvif/event_service</tds:XAddr>
          <tds:WSSubscriptionPolicySupport>false</tds:WSSubscriptionPolicySupport>
          <tds:WSPullPointSupport>false</tds:WSPullPointSupport>
          <tds:WSPausableSubscriptionSupport>false</tds:WSPausableSubscriptionSupport>
        </tds:Events>
        <tds:Analytics>
          <tds:XAddr>http://192.168.1.201/onvif/analytics_service</tds:XAddr>
          <tds:RuleSupport>true</tds:RuleSupport>
          <tds:AnalyticsModuleSupport>true</tds:AnalyticsModuleSupport>
        </tds:Analytics>
      </tds:Capabilities>
    </tds:GetCapabilitiesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetCapabilities") {
			t.Errorf("Request should contain GetCapabilities, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	caps, err := client.GetCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetCapabilities() failed: %v", err)
	}

	// Validate response matches real camera
	if caps.Device == nil {
		t.Fatal("Expected Device capabilities from Bosch FLEXIDOME")
	}
	if !strings.Contains(caps.Device.XAddr, "device_service") {
		t.Errorf("Expected device service XAddr from Bosch FLEXIDOME, got %s", caps.Device.XAddr)
	}
	if caps.Device.Network == nil {
		t.Fatal("Expected Network capabilities from Bosch FLEXIDOME")
	}
	if !caps.Device.Network.ZeroConfiguration {
		t.Error("Expected ZeroConfiguration=true from Bosch FLEXIDOME")
	}
	if caps.Device.Security == nil {
		t.Fatal("Expected Security capabilities from Bosch FLEXIDOME")
	}
	if !caps.Device.Security.TLS12 {
		t.Error("Expected TLS12=true from Bosch FLEXIDOME")
	}
	if caps.Media == nil {
		t.Fatal("Expected Media capabilities from Bosch FLEXIDOME")
	}
	if !strings.Contains(caps.Media.XAddr, "media_service") {
		t.Errorf("Expected media service XAddr from Bosch FLEXIDOME, got %s", caps.Media.XAddr)
	}
	if caps.Media.StreamingCapabilities == nil {
		t.Fatal("Expected StreamingCapabilities from Bosch FLEXIDOME")
	}
	if !caps.Media.StreamingCapabilities.RTPMulticast {
		t.Error("Expected RTPMulticast=true from Bosch FLEXIDOME")
	}
}

// TestGetServices_Bosch tests GetServices with real camera response.
func TestGetServices_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetServicesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Service>
        <tds:Namespace>http://www.onvif.org/ver10/device/wsdl</tds:Namespace>
        <tds:XAddr>http://192.168.1.201/onvif/device_service</tds:XAddr>
        <tds:Version>
          <tt:Major xmlns:tt="http://www.onvif.org/ver10/schema">1</tt:Major>
          <tt:Minor xmlns:tt="http://www.onvif.org/ver10/schema">3</tt:Minor>
        </tds:Version>
      </tds:Service>
      <tds:Service>
        <tds:Namespace>http://www.onvif.org/ver10/media/wsdl</tds:Namespace>
        <tds:XAddr>http://192.168.1.201/onvif/media_service</tds:XAddr>
        <tds:Version>
          <tt:Major xmlns:tt="http://www.onvif.org/ver10/schema">1</tt:Major>
          <tt:Minor xmlns:tt="http://www.onvif.org/ver10/schema">3</tt:Minor>
        </tds:Version>
      </tds:Service>
      <tds:Service>
        <tds:Namespace>http://www.onvif.org/ver10/events/wsdl</tds:Namespace>
        <tds:XAddr>http://192.168.1.201/onvif/event_service</tds:XAddr>
        <tds:Version>
          <tt:Major xmlns:tt="http://www.onvif.org/ver10/schema">1</tt:Major>
          <tt:Minor xmlns:tt="http://www.onvif.org/ver10/schema">4</tt:Minor>
        </tds:Version>
      </tds:Service>
    </tds:GetServicesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetServices") {
			t.Errorf("Request should contain GetServices, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	services, err := client.GetServices(ctx, false)
	if err != nil {
		t.Fatalf("GetServices() failed: %v", err)
	}

	// Validate response matches real camera
	if len(services) == 0 {
		t.Fatal("Expected at least one service from Bosch FLEXIDOME")
	}

	// Check for Device service
	foundDevice := false
	for _, svc := range services {
		if svc.Namespace == "http://www.onvif.org/ver10/device/wsdl" {
			foundDevice = true
			if svc.Version.Major != 1 || svc.Version.Minor != 3 {
				t.Errorf("Expected Device service version 1.3 (Bosch FLEXIDOME), got %d.%d", svc.Version.Major, svc.Version.Minor)
			}
			if !strings.Contains(svc.XAddr, "device_service") {
				t.Errorf("Expected device_service in XAddr (Bosch FLEXIDOME), got %s", svc.XAddr)
			}
		}
	}
	if !foundDevice {
		t.Error("Expected Device service from Bosch FLEXIDOME")
	}
}

// TestGetServiceCapabilities_Bosch tests GetServiceCapabilities with real camera response.
func TestGetServiceCapabilities_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	// Note: Uses attributes, not child elements
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetServiceCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Capabilities>
        <tds:Network IPFilter="false" ZeroConfiguration="true" IPVersion6="false" DynDNS="false"/>
        <tds:System DiscoveryResolve="false" DiscoveryBye="false" RemoteDiscovery="false" SystemBackup="false" SystemLogging="false" FirmwareUpgrade="false"/>
        <tds:Security TLS1.1="false" TLS1.2="true" OnboardKeyGeneration="false" AccessPolicyConfig="false"/>
      </tds:Capabilities>
    </tds:GetServiceCapabilitiesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetServiceCapabilities") {
			t.Errorf("Request should contain GetServiceCapabilities, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	caps, err := client.GetServiceCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetServiceCapabilities() failed: %v", err)
	}

	// Validate response matches real camera
	if caps.Network == nil {
		t.Fatal("Expected Network capabilities from Bosch FLEXIDOME")
	}
	if !caps.Network.ZeroConfiguration {
		t.Error("Expected ZeroConfiguration=true from Bosch FLEXIDOME")
	}
	if caps.Security == nil {
		t.Fatal("Expected Security capabilities from Bosch FLEXIDOME")
	}
	if !caps.Security.TLS12 {
		t.Error("Expected TLS12=true from Bosch FLEXIDOME")
	}
}

// TestGetSystemDateAndTime_Bosch tests GetSystemDateAndTime with real camera response.
func TestGetSystemDateAndTime_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetSystemDateAndTimeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:SystemDateAndTime>
        <tt:DateTimeType xmlns:tt="http://www.onvif.org/ver10/schema">Manual</tt:DateTimeType>
        <tt:DaylightSaving xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:DaylightSaving>
        <tt:TimeZone>
          <tt:TZ xmlns:tt="http://www.onvif.org/ver10/schema">CST6CDT</tt:TZ>
        </tt:TimeZone>
        <tt:UTCDateTime>
          <tt:Time>
            <tt:Hour xmlns:tt="http://www.onvif.org/ver10/schema">4</tt:Hour>
            <tt:Minute xmlns:tt="http://www.onvif.org/ver10/schema">56</tt:Minute>
            <tt:Second xmlns:tt="http://www.onvif.org/ver10/schema">14</tt:Second>
          </tt:Time>
          <tt:Date>
            <tt:Year xmlns:tt="http://www.onvif.org/ver10/schema">2025</tt:Year>
            <tt:Month xmlns:tt="http://www.onvif.org/ver10/schema">12</tt:Month>
            <tt:Day xmlns:tt="http://www.onvif.org/ver10/schema">2</tt:Day>
          </tt:Date>
        </tt:UTCDateTime>
      </tds:SystemDateAndTime>
    </tds:GetSystemDateAndTimeResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetSystemDateAndTime") {
			t.Errorf("Request should contain GetSystemDateAndTime, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	dateTime, err := client.GetSystemDateAndTime(ctx)
	if err != nil {
		t.Fatalf("GetSystemDateAndTime() failed: %v", err)
	}

	// GetSystemDateAndTime returns interface{} - just verify no error
	// The actual structure depends on the camera's response format
	_ = dateTime // Acknowledge we received a response
}

// TestGetHostname_Bosch tests GetHostname with real camera response.
func TestGetHostname_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetHostnameResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:HostnameInformation>
        <tt:FromDHCP xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:FromDHCP>
        <tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">BOSCH-404754734001050102</tt:Name>
      </tds:HostnameInformation>
    </tds:GetHostnameResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetHostname") {
			t.Errorf("Request should contain GetHostname, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	hostname, err := client.GetHostname(ctx)
	if err != nil {
		t.Fatalf("GetHostname() failed: %v", err)
	}

	// Validate response matches real camera
	if hostname == nil {
		t.Fatal("Expected HostnameInformation from Bosch FLEXIDOME")
	}
	if !strings.Contains(hostname.Name, "BOSCH") {
		t.Errorf("Expected hostname to contain BOSCH (Bosch FLEXIDOME), got %s", hostname.Name)
	}
	if hostname.FromDHCP {
		t.Error("Expected FromDHCP=false from Bosch FLEXIDOME")
	}
}

// TestGetScopes_Bosch tests GetScopes with real camera response.
func TestGetScopes_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetScopesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Scopes>
        <tt:ScopeDef xmlns:tt="http://www.onvif.org/ver10/schema">Fixed</tt:ScopeDef>
        <tt:ScopeItem xmlns:tt="http://www.onvif.org/ver10/schema">onvif://www.onvif.org/name/BOSCH-404754734001050102</tt:ScopeItem>
      </tds:Scopes>
      <tds:Scopes>
        <tt:ScopeDef xmlns:tt="http://www.onvif.org/ver10/schema">Fixed</tt:ScopeDef>
        <tt:ScopeItem xmlns:tt="http://www.onvif.org/ver10/schema">onvif://www.onvif.org/location/</tt:ScopeItem>
      </tds:Scopes>
      <tds:Scopes>
        <tt:ScopeDef xmlns:tt="http://www.onvif.org/ver10/schema">Fixed</tt:ScopeDef>
        <tt:ScopeItem xmlns:tt="http://www.onvif.org/ver10/schema">onvif://www.onvif.org/hardware/F000B543</tt:ScopeItem>
      </tds:Scopes>
    </tds:GetScopesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetScopes") {
			t.Errorf("Request should contain GetScopes, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	scopes, err := client.GetScopes(ctx)
	if err != nil {
		t.Fatalf("GetScopes() failed: %v", err)
	}

	// Validate response matches real camera
	if len(scopes) == 0 {
		t.Fatal("Expected at least one scope from Bosch FLEXIDOME")
	}

	// Check for hardware scope
	foundHardware := false
	for _, scope := range scopes {
		if strings.Contains(scope.ScopeItem, "hardware") {
			foundHardware = true
			if !strings.Contains(scope.ScopeItem, "F000B543") {
				t.Errorf("Expected hardware ID F000B543 in scope (Bosch FLEXIDOME), got %s", scope.ScopeItem)
			}
		}
	}
	if !foundHardware {
		t.Error("Expected hardware scope from Bosch FLEXIDOME")
	}
}

// TestGetUsers_Bosch tests GetUsers with real camera response.
func TestGetUsers_Bosch(t *testing.T) {
	// Real SOAP response from Bosch FLEXIDOME indoor 5100i IR (FW: 8.71.0066)
	realResponse := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetUsersResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:User>
        <tt:Username xmlns:tt="http://www.onvif.org/ver10/schema">service</tt:Username>
        <tt:UserLevel xmlns:tt="http://www.onvif.org/ver10/schema">Administrator</tt:UserLevel>
      </tds:User>
    </tds:GetUsersResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "GetUsers") {
			t.Errorf("Request should contain GetUsers, got: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithCredentials("service", "Service.1234"))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	users, err := client.GetUsers(ctx)
	if err != nil {
		t.Fatalf("GetUsers() failed: %v", err)
	}

	// Validate response matches real camera
	if len(users) == 0 {
		t.Fatal("Expected at least one user from Bosch FLEXIDOME")
	}
	if users[0].Username != "service" {
		t.Errorf("Expected username=service (Bosch FLEXIDOME), got %s", users[0].Username)
	}
	if users[0].UserLevel != "Administrator" {
		t.Errorf("Expected UserLevel=Administrator (Bosch FLEXIDOME), got %s", users[0].UserLevel)
	}
}
