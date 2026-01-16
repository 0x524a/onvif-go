package onvif

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockDeviceSecurityServer() *httptest.Server {
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
		case strings.Contains(bodyContent, "GetRemoteUser"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetRemoteUserResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:RemoteUser>
				<tt:Username>remote_admin</tt:Username>
				<tt:Password></tt:Password>
				<tt:UseDerivedPassword>true</tt:UseDerivedPassword>
			</tds:RemoteUser>
		</tds:GetRemoteUserResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetRemoteUser"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetRemoteUserResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetIPAddressFilter"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetIPAddressFilterResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:IPAddressFilter>
				<tt:Type>Allow</tt:Type>
				<tt:IPv4Address>
					<tt:Address>192.168.1.0</tt:Address>
					<tt:PrefixLength>24</tt:PrefixLength>
				</tt:IPv4Address>
			</tds:IPAddressFilter>
		</tds:GetIPAddressFilterResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetIPAddressFilter"),
			strings.Contains(bodyContent, "AddIPAddressFilter"),
			strings.Contains(bodyContent, "RemoveIPAddressFilter"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetIPAddressFilterResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetZeroConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetZeroConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:ZeroConfiguration>
				<tt:InterfaceToken>eth0</tt:InterfaceToken>
				<tt:Enabled>true</tt:Enabled>
				<tt:Addresses>169.254.1.100</tt:Addresses>
			</tds:ZeroConfiguration>
		</tds:GetZeroConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetZeroConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetZeroConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetPasswordComplexityConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetPasswordComplexityConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:MinLen>8</tds:MinLen>
			<tds:Uppercase>1</tds:Uppercase>
			<tds:Number>1</tds:Number>
			<tds:SpecialChars>1</tds:SpecialChars>
			<tds:BlockUsernameOccurrence>true</tds:BlockUsernameOccurrence>
			<tds:PolicyConfigurationLocked>false</tds:PolicyConfigurationLocked>
		</tds:GetPasswordComplexityConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetPasswordComplexityConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetPasswordComplexityConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetPasswordHistoryConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetPasswordHistoryConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:Enabled>true</tds:Enabled>
			<tds:Length>5</tds:Length>
		</tds:GetPasswordHistoryConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetPasswordHistoryConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetPasswordHistoryConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetAuthFailureWarningConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetAuthFailureWarningConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:Enabled>true</tds:Enabled>
			<tds:MonitorPeriod>60</tds:MonitorPeriod>
			<tds:MaxAuthFailures>5</tds:MaxAuthFailures>
		</tds:GetAuthFailureWarningConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetAuthFailureWarningConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetAuthFailureWarningConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetRemoteUser(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	remoteUser, err := client.GetRemoteUser(ctx)
	if err != nil {
		t.Fatalf("GetRemoteUser failed: %v", err)
	}

	if remoteUser.Username != "remote_admin" {
		t.Errorf("Expected username 'remote_admin', got %s", remoteUser.Username)
	}

	if !remoteUser.UseDerivedPassword {
		t.Error("UseDerivedPassword should be true")
	}
}

func TestSetRemoteUser(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	remoteUser := &RemoteUser{
		Username:           "new_remote",
		Password:           "password123",
		UseDerivedPassword: true,
	}

	err = client.SetRemoteUser(ctx, remoteUser)
	if err != nil {
		t.Fatalf("SetRemoteUser failed: %v", err)
	}
}

func TestGetIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter, err := client.GetIPAddressFilter(ctx)
	if err != nil {
		t.Fatalf("GetIPAddressFilter failed: %v", err)
	}

	if filter.Type != IPAddressFilterAllow {
		t.Errorf("Expected Allow filter type, got %s", filter.Type)
	}

	if len(filter.IPv4Address) != 1 {
		t.Fatalf("Expected 1 IPv4 address, got %d", len(filter.IPv4Address))
	}

	if filter.IPv4Address[0].Address != "192.168.1.0" {
		t.Errorf("Expected address 192.168.1.0, got %s", filter.IPv4Address[0].Address)
	}

	if filter.IPv4Address[0].PrefixLength != 24 {
		t.Errorf("Expected prefix length 24, got %d", filter.IPv4Address[0].PrefixLength)
	}
}

func TestSetIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "10.0.0.0", PrefixLength: 8},
		},
	}

	err = client.SetIPAddressFilter(ctx, filter)
	if err != nil {
		t.Fatalf("SetIPAddressFilter failed: %v", err)
	}
}

func TestAddIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	err = client.AddIPAddressFilter(ctx, filter)
	if err != nil {
		t.Fatalf("AddIPAddressFilter failed: %v", err)
	}
}

func TestRemoveIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	err = client.RemoveIPAddressFilter(ctx, filter)
	if err != nil {
		t.Fatalf("RemoveIPAddressFilter failed: %v", err)
	}
}

func TestGetZeroConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	zeroConf, err := client.GetZeroConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetZeroConfiguration failed: %v", err)
	}

	if zeroConf.InterfaceToken != "eth0" {
		t.Errorf("Expected interface token 'eth0', got %s", zeroConf.InterfaceToken)
	}

	if !zeroConf.Enabled {
		t.Error("Zero configuration should be enabled")
	}

	if len(zeroConf.Addresses) != 1 || zeroConf.Addresses[0] != "169.254.1.100" {
		t.Errorf("Expected address 169.254.1.100, got %v", zeroConf.Addresses)
	}
}

func TestSetZeroConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.SetZeroConfiguration(ctx, "eth0", true)
	if err != nil {
		t.Fatalf("SetZeroConfiguration failed: %v", err)
	}
}

func TestGetPasswordComplexityConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config, err := client.GetPasswordComplexityConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetPasswordComplexityConfiguration failed: %v", err)
	}

	if config.MinLen != 8 {
		t.Errorf("Expected MinLen 8, got %d", config.MinLen)
	}

	if config.Uppercase != 1 {
		t.Errorf("Expected Uppercase 1, got %d", config.Uppercase)
	}

	if config.Number != 1 {
		t.Errorf("Expected Number 1, got %d", config.Number)
	}

	if config.SpecialChars != 1 {
		t.Errorf("Expected SpecialChars 1, got %d", config.SpecialChars)
	}

	if !config.BlockUsernameOccurrence {
		t.Error("BlockUsernameOccurrence should be true")
	}

	if config.PolicyConfigurationLocked {
		t.Error("PolicyConfigurationLocked should be false")
	}
}

func TestSetPasswordComplexityConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config := &PasswordComplexityConfiguration{
		MinLen:                    10,
		Uppercase:                 2,
		Number:                    2,
		SpecialChars:              1,
		BlockUsernameOccurrence:   true,
		PolicyConfigurationLocked: false,
	}

	err = client.SetPasswordComplexityConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetPasswordComplexityConfiguration failed: %v", err)
	}
}

func TestGetPasswordHistoryConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config, err := client.GetPasswordHistoryConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetPasswordHistoryConfiguration failed: %v", err)
	}

	if !config.Enabled {
		t.Error("Password history should be enabled")
	}

	if config.Length != 5 {
		t.Errorf("Expected Length 5, got %d", config.Length)
	}
}

func TestSetPasswordHistoryConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config := &PasswordHistoryConfiguration{
		Enabled: true,
		Length:  10,
	}

	err = client.SetPasswordHistoryConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetPasswordHistoryConfiguration failed: %v", err)
	}
}

func TestGetAuthFailureWarningConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config, err := client.GetAuthFailureWarningConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetAuthFailureWarningConfiguration failed: %v", err)
	}

	if !config.Enabled {
		t.Error("Auth failure warning should be enabled")
	}

	if config.MonitorPeriod != 60 {
		t.Errorf("Expected MonitorPeriod 60, got %d", config.MonitorPeriod)
	}

	if config.MaxAuthFailures != 5 {
		t.Errorf("Expected MaxAuthFailures 5, got %d", config.MaxAuthFailures)
	}
}

func TestSetAuthFailureWarningConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config := &AuthFailureWarningConfiguration{
		Enabled:         true,
		MonitorPeriod:   120,
		MaxAuthFailures: 3,
	}

	err = client.SetAuthFailureWarningConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetAuthFailureWarningConfiguration failed: %v", err)
	}
}

func TestIPAddressFilterTypeConstants(t *testing.T) {
	if IPAddressFilterAllow != "Allow" {
		t.Errorf("IPAddressFilterAllow should be 'Allow', got %s", IPAddressFilterAllow)
	}

	if IPAddressFilterDeny != "Deny" {
		t.Errorf("IPAddressFilterDeny should be 'Deny', got %s", IPAddressFilterDeny)
	}
}

// Benchmarks for device security operations.

func BenchmarkGetRemoteUser(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetRemoteUser(ctx)
	}
}

func BenchmarkSetRemoteUser(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	remoteUser := &RemoteUser{
		Username:           "test_user",
		Password:           "password123",
		UseDerivedPassword: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetRemoteUser(ctx, remoteUser)
	}
}

func BenchmarkGetIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetIPAddressFilter(ctx)
	}
}

func BenchmarkSetIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "192.168.1.0", PrefixLength: 24},
			{Address: "10.0.0.0", PrefixLength: 8},
		},
		IPv6Address: []PrefixedIPv6Address{
			{Address: "fe80::", PrefixLength: 64},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetIPAddressFilter(ctx, filter)
	}
}

func BenchmarkAddIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AddIPAddressFilter(ctx, filter)
	}
}

func BenchmarkRemoveIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.RemoveIPAddressFilter(ctx, filter)
	}
}

func BenchmarkGetZeroConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetZeroConfiguration(ctx)
	}
}

func BenchmarkSetZeroConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetZeroConfiguration(ctx, "eth0", true)
	}
}

func BenchmarkGetPasswordComplexityConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetPasswordComplexityConfiguration(ctx)
	}
}

func BenchmarkSetPasswordComplexityConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	config := &PasswordComplexityConfiguration{
		MinLen:                    10,
		Uppercase:                 2,
		Number:                    2,
		SpecialChars:              1,
		BlockUsernameOccurrence:   true,
		PolicyConfigurationLocked: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetPasswordComplexityConfiguration(ctx, config)
	}
}

func BenchmarkGetPasswordHistoryConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetPasswordHistoryConfiguration(ctx)
	}
}

func BenchmarkSetPasswordHistoryConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	config := &PasswordHistoryConfiguration{
		Enabled: true,
		Length:  10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetPasswordHistoryConfiguration(ctx, config)
	}
}

func BenchmarkGetAuthFailureWarningConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetAuthFailureWarningConfiguration(ctx)
	}
}

func BenchmarkSetAuthFailureWarningConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	config := &AuthFailureWarningConfiguration{
		Enabled:         true,
		MonitorPeriod:   120,
		MaxAuthFailures: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetAuthFailureWarningConfiguration(ctx, config)
	}
}

// BenchmarkIPAddressFilterWithManyAddresses tests performance with larger address lists.
func BenchmarkIPAddressFilterWithManyAddresses(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	// Create filter with many addresses to test pre-allocation efficiency
	filter := &IPAddressFilter{
		Type:        IPAddressFilterAllow,
		IPv4Address: make([]PrefixedIPv4Address, 100),
		IPv6Address: make([]PrefixedIPv6Address, 50),
	}

	for i := 0; i < 100; i++ {
		filter.IPv4Address[i] = PrefixedIPv4Address{
			Address:      "192.168.1.0",
			PrefixLength: 24,
		}
	}

	for i := 0; i < 50; i++ {
		filter.IPv6Address[i] = PrefixedIPv6Address{
			Address:      "fe80::",
			PrefixLength: 64,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetIPAddressFilter(ctx, filter)
	}
}
