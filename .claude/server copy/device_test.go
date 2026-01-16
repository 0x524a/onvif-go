package server

import (
	"encoding/xml"
	"testing"
)

func TestHandleGetDeviceInformation(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetDeviceInformation(nil)
	if err != nil {
		t.Fatalf("HandleGetDeviceInformation() error = %v", err)
	}

	deviceResp, ok := resp.(*GetDeviceInformationResponse)
	if !ok {
		t.Fatalf("Response is not GetDeviceInformationResponse, got %T", resp)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Manufacturer", deviceResp.Manufacturer, config.DeviceInfo.Manufacturer},
		{"Model", deviceResp.Model, config.DeviceInfo.Model},
		{"FirmwareVersion", deviceResp.FirmwareVersion, config.DeviceInfo.FirmwareVersion},
		{"SerialNumber", deviceResp.SerialNumber, config.DeviceInfo.SerialNumber},
		{"HardwareID", deviceResp.HardwareID, config.DeviceInfo.HardwareID},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s mismatch: got %s, want %s", tt.name, tt.got, tt.want)
		}
	}
}

func TestHandleGetCapabilities(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetCapabilities(nil)
	if err != nil {
		t.Fatalf("HandleGetCapabilities() error = %v", err)
	}

	capsResp, ok := resp.(*GetCapabilitiesResponse)
	if !ok {
		t.Fatalf("Response is not GetCapabilitiesResponse, got %T", resp)
	}

	if capsResp.Capabilities == nil {
		t.Error("Capabilities is nil")

		return
	}

	// Check device capabilities
	if capsResp.Capabilities.Device == nil {
		t.Error("Device capabilities is nil")
	}

	// Check media capabilities
	if capsResp.Capabilities.Media == nil {
		t.Error("Media capabilities is nil")
	}

	// Check PTZ capabilities if supported
	if config.SupportPTZ && capsResp.Capabilities.PTZ == nil {
		t.Error("PTZ capabilities is nil but PTZ is supported")
	}

	// Check Imaging capabilities if supported
	if config.SupportImaging && capsResp.Capabilities.Imaging == nil {
		t.Error("Imaging capabilities is nil but Imaging is supported")
	}
}

func TestHandleGetSystemDateAndTime(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetSystemDateAndTime(nil)
	if err != nil {
		t.Fatalf("HandleGetSystemDateAndTime() error = %v", err)
	}

	// Response should be a map or interface
	if resp == nil {
		t.Error("Response is nil")

		return
	}
}

func TestHandleGetServices(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetServices(nil)
	if err != nil {
		t.Fatalf("HandleGetServices() error = %v", err)
	}

	servicesResp, ok := resp.(*GetServicesResponse)
	if !ok {
		t.Fatalf("Response is not GetServicesResponse, got %T", resp)
	}

	if len(servicesResp.Service) == 0 {
		t.Error("No services returned")

		return
	}

	// Check that device and media services are present
	hasDeviceService := false
	hasMediaService := false

	for _, service := range servicesResp.Service {
		if service.Namespace == "http://www.onvif.org/ver10/device/wsdl" {
			hasDeviceService = true
		}
		if service.Namespace == "http://www.onvif.org/ver10/media/wsdl" {
			hasMediaService = true
		}
	}

	if !hasDeviceService {
		t.Error("Device service not found")
	}
	if !hasMediaService {
		t.Error("Media service not found")
	}
}

func TestHandleSystemReboot(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleSystemReboot(nil)
	if err != nil {
		t.Fatalf("HandleSystemReboot() error = %v", err)
	}

	rebootResp, ok := resp.(*SystemRebootResponse)
	if !ok {
		t.Fatalf("Response is not SystemRebootResponse, got %T", resp)
	}

	if rebootResp.Message == "" {
		t.Error("Reboot message is empty")
	}
}

func TestGetDeviceInformationResponseXML(t *testing.T) {
	resp := &GetDeviceInformationResponse{
		Manufacturer:    "TestManu",
		Model:           "TestModel",
		FirmwareVersion: "1.0.0",
		SerialNumber:    "SN123",
		HardwareID:      "HW001",
	}

	// Marshal to XML
	data, err := xml.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal back
	var unmarshaled GetDeviceInformationResponse
	err = xml.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaled.Manufacturer != resp.Manufacturer {
		t.Errorf("Manufacturer mismatch: %s != %s", unmarshaled.Manufacturer, resp.Manufacturer)
	}
	if unmarshaled.Model != resp.Model {
		t.Errorf("Model mismatch: %s != %s", unmarshaled.Model, resp.Model)
	}
}

func TestCapabilitiesStructure(t *testing.T) {
	caps := &Capabilities{
		Device: &DeviceCapabilities{
			XAddr: "http://localhost:8080/onvif/device_service",
			Network: &NetworkCapabilities{
				IPFilter:          true,
				ZeroConfiguration: true,
				IPVersion6:        true,
				DynDNS:            false,
			},
			System: &SystemCapabilities{
				DiscoveryResolve: true,
				DiscoveryBye:     true,
				RemoteDiscovery:  false,
				SystemBackup:     true,
				SystemLogging:    true,
				FirmwareUpgrade:  true,
			},
		},
		Media: &MediaCapabilities{
			XAddr: "http://localhost:8080/onvif/media_service",
			StreamingCapabilities: &StreamingCapabilities{
				RTPMulticast: true,
				RTPTCP:       true,
				RTPRTSPTCP:   true,
			},
		},
	}

	// Test that capabilities are properly structured
	if caps.Device == nil || caps.Device.XAddr == "" {
		t.Error("Device capabilities not properly set")
	}
	if caps.Media == nil || caps.Media.XAddr == "" {
		t.Error("Media capabilities not properly set")
	}

	// Test network capabilities
	if !caps.Device.Network.IPFilter {
		t.Error("IPFilter should be true")
	}

	// Test system capabilities
	if !caps.Device.System.SystemBackup {
		t.Error("SystemBackup should be true")
	}
}

func TestMediaCapabilitiesStructure(t *testing.T) {
	caps := &MediaCapabilities{
		XAddr: "http://localhost:8080/onvif/media_service",
		StreamingCapabilities: &StreamingCapabilities{
			RTPMulticast: true,
			RTPTCP:       true,
			RTPRTSPTCP:   true,
		},
	}

	if caps.StreamingCapabilities == nil {
		t.Error("StreamingCapabilities is nil")
	}

	if !caps.StreamingCapabilities.RTPMulticast {
		t.Error("RTP Multicast should be supported")
	}
	if !caps.StreamingCapabilities.RTPTCP {
		t.Error("RTP TCP should be supported")
	}
	if !caps.StreamingCapabilities.RTPRTSPTCP {
		t.Error("RTSP should be supported")
	}
}

func TestHandleSnapshot(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	// The snapshot handler is tested via HTTP in integration tests
	// Here we just verify the configuration is available
	profiles := server.ListProfiles()
	if len(profiles) == 0 {
		t.Error("No profiles available for snapshot")

		return
	}

	if !profiles[0].Snapshot.Enabled {
		t.Error("Snapshot should be enabled in test config")
	}
}

func TestHandleGetCapabilitiesDetails(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetCapabilities(nil)
	if err != nil {
		t.Fatalf("HandleGetCapabilities error: %v", err)
	}

	capsResp, ok := resp.(*GetCapabilitiesResponse)
	if !ok {
		t.Fatalf("Response is not GetCapabilitiesResponse: %T", resp)
	}

	if capsResp.Capabilities == nil {
		t.Error("Capabilities is nil")

		return
	}

	if capsResp.Capabilities.Device == nil {
		t.Error("Device capabilities is nil")
	}

	if capsResp.Capabilities.Media == nil {
		t.Error("Media capabilities is nil")
	}

	// Check device capabilities structure
	devCaps := capsResp.Capabilities.Device
	if devCaps.XAddr == "" {
		t.Error("Device XAddr is empty")
	}
	if devCaps.Network == nil {
		t.Error("Network capabilities is nil")
	}
	if devCaps.System == nil {
		t.Error("System capabilities is nil")
	}
}

func TestHandleGetServicesDetails(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config)

	resp, err := server.HandleGetServices(nil)
	if err != nil {
		t.Fatalf("HandleGetServices error: %v", err)
	}

	servResp, ok := resp.(*GetServicesResponse)
	if !ok {
		t.Fatalf("Response is not GetServicesResponse: %T", resp)
	}

	if len(servResp.Service) == 0 {
		t.Error("No services returned")

		return
	}

	// Check service structure
	for _, svc := range servResp.Service {
		if svc.Namespace == "" {
			t.Error("Service Namespace is empty")
		}
		if svc.XAddr == "" {
			t.Error("Service XAddr is empty")
		}
	}
}

func TestGetCapabilitiesResponse(t *testing.T) {
	caps := &Capabilities{
		Device: &DeviceCapabilities{
			XAddr: "http://localhost:8080/device",
			Network: &NetworkCapabilities{
				IPFilter:          true,
				ZeroConfiguration: true,
				IPVersion6:        true,
			},
			System: &SystemCapabilities{
				DiscoveryResolve: true,
				DiscoveryBye:     true,
				SystemBackup:     true,
			},
		},
		Media: &MediaCapabilities{
			XAddr: "http://localhost:8080/media",
			StreamingCapabilities: &StreamingCapabilities{
				RTPMulticast: true,
				RTPTCP:       true,
				RTPRTSPTCP:   true,
			},
		},
	}

	resp := &GetCapabilitiesResponse{
		Capabilities: caps,
	}

	if resp.Capabilities == nil {
		t.Error("Capabilities is nil in response")
	}
	if resp.Capabilities.Device == nil {
		t.Error("Device capabilities is nil in response")
	}
}
