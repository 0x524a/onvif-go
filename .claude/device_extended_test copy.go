package onvif

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockDeviceExtendedServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := xml.NewDecoder(r.Body)
		var envelope struct {
			Body struct {
				Content []byte `xml:",innerxml"`
			} `xml:"Body"`
		}
		_ = decoder.Decode(&envelope)
		bodyContent := string(envelope.Body.Content)

		w.Header().Set("Content-Type", "application/soap+xml")

		switch {
		case strings.Contains(bodyContent, "AddScopes"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:AddScopesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "RemoveScopes"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:RemoveScopesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:ScopeItem>onvif://www.onvif.org/location/test</tds:ScopeItem>
		</tds:RemoveScopesResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetScopes"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetScopesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetRelayOutputs"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetRelayOutputsResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:RelayOutputs token="relay1">
				<tt:Properties>
					<tt:Mode>Bistable</tt:Mode>
					<tt:DelayTime>PT0S</tt:DelayTime>
					<tt:IdleState>closed</tt:IdleState>
				</tt:Properties>
			</tds:RelayOutputs>
		</tds:GetRelayOutputsResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetRelayOutputSettings"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetRelayOutputSettingsResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetRelayOutputState"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetRelayOutputStateResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SendAuxiliaryCommand"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SendAuxiliaryCommandResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:AuxiliaryCommandResponse>tt:IRLamp|On</tds:AuxiliaryCommandResponse>
		</tds:SendAuxiliaryCommandResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetSystemLog"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetSystemLogResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:SystemLog>
				<tt:String>System log content here</tt:String>
			</tds:SystemLog>
		</tds:GetSystemLogResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetSystemFactoryDefault"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetSystemFactoryDefaultResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "StartFirmwareUpgrade"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:StartFirmwareUpgradeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:UploadUri>http://192.168.1.100/upload</tds:UploadUri>
			<tds:UploadDelay>PT5S</tds:UploadDelay>
			<tds:ExpectedDownTime>PT60S</tds:ExpectedDownTime>
		</tds:StartFirmwareUpgradeResponse>
	</s:Body>
</s:Envelope>`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestAddScopes(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	scopes := []string{
		"onvif://www.onvif.org/location/building/floor1",
		"onvif://www.onvif.org/name/camera-entrance",
	}

	err = client.AddScopes(ctx, scopes)
	if err != nil {
		t.Fatalf("AddScopes failed: %v", err)
	}
}

func TestRemoveScopes(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	scopes := []string{"onvif://www.onvif.org/location/test"}

	removed, err := client.RemoveScopes(ctx, scopes)
	if err != nil {
		t.Fatalf("RemoveScopes failed: %v", err)
	}

	if len(removed) != 1 {
		t.Fatalf("Expected 1 removed scope, got %d", len(removed))
	}

	if removed[0] != "onvif://www.onvif.org/location/test" {
		t.Errorf("Expected removed scope to match, got %s", removed[0])
	}
}

func TestSetScopes(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	scopes := []string{"scope1", "scope2"}

	err = client.SetScopes(ctx, scopes)
	if err != nil {
		t.Fatalf("SetScopes failed: %v", err)
	}
}

func TestGetRelayOutputs(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	relays, err := client.GetRelayOutputs(ctx)
	if err != nil {
		t.Fatalf("GetRelayOutputs failed: %v", err)
	}

	if len(relays) != 1 {
		t.Fatalf("Expected 1 relay, got %d", len(relays))
	}

	if relays[0].Token != "relay1" {
		t.Errorf("Expected relay token 'relay1', got %s", relays[0].Token)
	}

	if relays[0].Properties.Mode != RelayModeBistable {
		t.Errorf("Expected Bistable mode, got %s", relays[0].Properties.Mode)
	}

	if relays[0].Properties.IdleState != RelayIdleStateClosed {
		t.Errorf("Expected closed idle state, got %s", relays[0].Properties.IdleState)
	}
}

func TestSetRelayOutputSettings(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	settings := &RelayOutputSettings{
		Mode:      RelayModeBistable,
		IdleState: RelayIdleStateClosed,
	}

	err = client.SetRelayOutputSettings(ctx, "relay1", settings)
	if err != nil {
		t.Fatalf("SetRelayOutputSettings failed: %v", err)
	}
}

func TestSetRelayOutputState(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test active state
	err = client.SetRelayOutputState(ctx, "relay1", RelayLogicalStateActive)
	if err != nil {
		t.Fatalf("SetRelayOutputState (active) failed: %v", err)
	}

	// Test inactive state
	err = client.SetRelayOutputState(ctx, "relay1", RelayLogicalStateInactive)
	if err != nil {
		t.Fatalf("SetRelayOutputState (inactive) failed: %v", err)
	}
}

func TestSendAuxiliaryCommand(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	response, err := client.SendAuxiliaryCommand(ctx, "tt:IRLamp|On")
	if err != nil {
		t.Fatalf("SendAuxiliaryCommand failed: %v", err)
	}

	if response != "tt:IRLamp|On" {
		t.Errorf("Expected response 'tt:IRLamp|On', got %s", response)
	}
}

func TestGetSystemLog(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	log, err := client.GetSystemLog(ctx, SystemLogTypeSystem)
	if err != nil {
		t.Fatalf("GetSystemLog failed: %v", err)
	}

	if log.String != "System log content here" {
		t.Errorf("Expected system log content, got %s", log.String)
	}
}

func TestSetSystemFactoryDefault(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test soft reset
	err = client.SetSystemFactoryDefault(ctx, FactoryDefaultSoft)
	if err != nil {
		t.Fatalf("SetSystemFactoryDefault (soft) failed: %v", err)
	}

	// Test hard reset
	err = client.SetSystemFactoryDefault(ctx, FactoryDefaultHard)
	if err != nil {
		t.Fatalf("SetSystemFactoryDefault (hard) failed: %v", err)
	}
}

func TestStartFirmwareUpgrade(t *testing.T) {
	server := newMockDeviceExtendedServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	uploadURI, delay, downtime, err := client.StartFirmwareUpgrade(ctx)
	if err != nil {
		t.Fatalf("StartFirmwareUpgrade failed: %v", err)
	}

	if uploadURI != "http://192.168.1.100/upload" {
		t.Errorf("Expected upload URI http://192.168.1.100/upload, got %s", uploadURI)
	}

	if delay != "PT5S" {
		t.Errorf("Expected delay PT5S, got %s", delay)
	}

	if downtime != "PT60S" {
		t.Errorf("Expected downtime PT60S, got %s", downtime)
	}
}

func TestRelayModeConstants(t *testing.T) {
	if RelayModeMonostable != "Monostable" {
		t.Errorf("RelayModeMonostable should be 'Monostable', got %s", RelayModeMonostable)
	}

	if RelayModeBistable != "Bistable" {
		t.Errorf("RelayModeBistable should be 'Bistable', got %s", RelayModeBistable)
	}
}

func TestRelayIdleStateConstants(t *testing.T) {
	if RelayIdleStateClosed != "closed" {
		t.Errorf("RelayIdleStateClosed should be 'closed', got %s", RelayIdleStateClosed)
	}

	if RelayIdleStateOpen != "open" {
		t.Errorf("RelayIdleStateOpen should be 'open', got %s", RelayIdleStateOpen)
	}
}

func TestRelayLogicalStateConstants(t *testing.T) {
	if RelayLogicalStateActive != "active" {
		t.Errorf("RelayLogicalStateActive should be 'active', got %s", RelayLogicalStateActive)
	}

	if RelayLogicalStateInactive != "inactive" {
		t.Errorf("RelayLogicalStateInactive should be 'inactive', got %s", RelayLogicalStateInactive)
	}
}

func TestSystemLogTypeConstants(t *testing.T) {
	if SystemLogTypeSystem != "System" {
		t.Errorf("SystemLogTypeSystem should be 'System', got %s", SystemLogTypeSystem)
	}

	if SystemLogTypeAccess != "Access" {
		t.Errorf("SystemLogTypeAccess should be 'Access', got %s", SystemLogTypeAccess)
	}
}

func TestFactoryDefaultTypeConstants(t *testing.T) {
	if FactoryDefaultHard != "Hard" {
		t.Errorf("FactoryDefaultHard should be 'Hard', got %s", FactoryDefaultHard)
	}

	if FactoryDefaultSoft != "Soft" {
		t.Errorf("FactoryDefaultSoft should be 'Soft', got %s", FactoryDefaultSoft)
	}
}
