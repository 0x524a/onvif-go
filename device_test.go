package onvif

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDeviceInformation(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "successful device information retrieval",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := `<?xml version="1.0" encoding="UTF-8"?>
				<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
					<s:Body>
						<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
							<tds:Manufacturer>Test Manufacturer</tds:Manufacturer>
							<tds:Model>Test Model</tds:Model>
							<tds:FirmwareVersion>1.0.0</tds:FirmwareVersion>
							<tds:SerialNumber>12345</tds:SerialNumber>
							<tds:HardwareId>HW-001</tds:HardwareId>
						</tds:GetDeviceInformationResponse>
					</s:Body>
				</s:Envelope>`
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(response))
			},
			wantErr: false,
		},
		{
			name: "SOAP fault response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := `<?xml version="1.0" encoding="UTF-8"?>
				<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
					<s:Body>
						<s:Fault>
							<s:Code><s:Value>s:Receiver</s:Value></s:Code>
							<s:Reason><s:Text xml:lang="en">Internal error</s:Text></s:Reason>
						</s:Fault>
					</s:Body>
				</s:Envelope>`
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(response))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			deviceInfo, err := client.GetDeviceInformation(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeviceInformation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && deviceInfo == nil {
				t.Error("Expected device information, got nil")
			}

			if !tt.wantErr && deviceInfo != nil {
				if deviceInfo.Manufacturer != "Test Manufacturer" {
					t.Errorf("Expected manufacturer 'Test Manufacturer', got '%s'", deviceInfo.Manufacturer)
				}
			}
		})
	}
}

func TestGetCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:Capabilities>
						<tds:Device>
							<tds:XAddr>http://example.com/onvif/device_service</tds:XAddr>
						</tds:Device>
						<tds:Media>
							<tds:XAddr>http://example.com/onvif/media_service</tds:XAddr>
						</tds:Media>
					</tds:Capabilities>
				</tds:GetCapabilitiesResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	capabilities, err := client.GetCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	if capabilities == nil {
		t.Fatal("Expected capabilities, got nil")
	}

	if capabilities.Device == nil || capabilities.Device.XAddr == "" {
		t.Error("Expected Device capabilities with XAddr")
	}
}

func TestGetHostname(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetHostnameResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:HostnameInformation>
						<tt:FromDHCP>false</tt:FromDHCP>
						<tt:Name>test-camera</tt:Name>
					</tds:HostnameInformation>
				</tds:GetHostnameResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	hostname, err := client.GetHostname(context.Background())
	if err != nil {
		t.Fatalf("GetHostname() error = %v", err)
	}

	if hostname == nil {
		t.Fatal("Expected hostname information, got nil")
	}

	if hostname.Name != "test-camera" {
		t.Errorf("Expected hostname 'test-camera', got '%s'", hostname.Name)
	}
}

func TestSetHostname(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request body contains the new hostname
		var envelope struct {
			Body struct {
				SetHostname struct {
					XMLName xml.Name `xml:"SetHostname"`
					Name    string   `xml:"Name"`
				} `xml:"SetHostname"`
			} `xml:"Body"`
		}

		if err := xml.NewDecoder(r.Body).Decode(&envelope); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if envelope.Body.SetHostname.Name != "new-hostname" {
			t.Errorf("Expected hostname 'new-hostname', got '%s'", envelope.Body.SetHostname.Name)
		}

		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:SetHostnameResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.SetHostname(context.Background(), "new-hostname")
	if err != nil {
		t.Fatalf("SetHostname() error = %v", err)
	}
}

func TestGetDNS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetDNSResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:DNSInformation>
						<tt:FromDHCP>true</tt:FromDHCP>
						<tt:SearchDomain>example.com</tt:SearchDomain>
						<tt:DNSFromDHCP>
							<tt:Type>IPv4</tt:Type>
							<tt:IPv4Address>8.8.8.8</tt:IPv4Address>
						</tt:DNSFromDHCP>
					</tds:DNSInformation>
				</tds:GetDNSResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	dns, err := client.GetDNS(context.Background())
	if err != nil {
		t.Fatalf("GetDNS() error = %v", err)
	}

	if dns == nil {
		t.Fatal("Expected DNS information, got nil")
	}

	if !dns.FromDHCP {
		t.Error("Expected DNS from DHCP")
	}
}

func TestGetUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetUsersResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:User>
						<tt:Username>admin</tt:Username>
						<tt:UserLevel>Administrator</tt:UserLevel>
					</tds:User>
					<tds:User>
						<tt:Username>user</tt:Username>
						<tt:UserLevel>User</tt:UserLevel>
					</tds:User>
				</tds:GetUsersResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	users, err := client.GetUsers(context.Background())
	if err != nil {
		t.Fatalf("GetUsers() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].Username != "admin" {
		t.Errorf("Expected first user to be 'admin', got '%s'", users[0].Username)
	}
}

func TestCreateUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:CreateUsersResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	users := []*User{
		{
			Username:  "newuser",
			Password:  "password123",
			UserLevel: "User",
		},
	}

	err = client.CreateUsers(context.Background(), users)
	if err != nil {
		t.Fatalf("CreateUsers() error = %v", err)
	}
}

func TestDeleteUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:DeleteUsersResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.DeleteUsers(context.Background(), []string{"testuser"})
	if err != nil {
		t.Fatalf("DeleteUsers() error = %v", err)
	}
}

func TestGetNetworkInterfaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetNetworkInterfacesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:NetworkInterfaces token="eth0">
						<tt:Enabled>true</tt:Enabled>
						<tt:Info>
							<tt:Name>eth0</tt:Name>
							<tt:HwAddress>00:11:22:33:44:55</tt:HwAddress>
							<tt:MTU>1500</tt:MTU>
						</tt:Info>
						<tt:IPv4>
							<tt:Enabled>true</tt:Enabled>
							<tt:Config>
								<tt:DHCP>false</tt:DHCP>
								<tt:Manual>
									<tt:Address>192.168.1.100</tt:Address>
									<tt:PrefixLength>24</tt:PrefixLength>
								</tt:Manual>
							</tt:Config>
						</tt:IPv4>
					</tds:NetworkInterfaces>
				</tds:GetNetworkInterfacesResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	interfaces, err := client.GetNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatalf("GetNetworkInterfaces() error = %v", err)
	}

	if len(interfaces) != 1 {
		t.Errorf("Expected 1 interface, got %d", len(interfaces))
	}

	if interfaces[0].Info.Name != "eth0" {
		t.Errorf("Expected interface name 'eth0', got '%s'", interfaces[0].Info.Name)
	}
}

func BenchmarkDeviceGetDeviceInformation(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:Manufacturer>Test</tds:Manufacturer>
					<tds:Model>Model</tds:Model>
					<tds:FirmwareVersion>1.0</tds:FirmwareVersion>
					<tds:SerialNumber>123</tds:SerialNumber>
					<tds:HardwareId>HW1</tds:HardwareId>
				</tds:GetDeviceInformationResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetDeviceInformation(ctx)
	}
}
