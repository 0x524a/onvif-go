package onvif

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockDeviceAdditionalServer() *httptest.Server {
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
		case strings.Contains(bodyContent, "GetGeoLocation"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Body>
		<tds:GetGeoLocationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:Location Lon="-122.4194" Lat="37.7749" Elevation="10.5">
				<tt:Entity>Building A</tt:Entity>
				<tt:Token>location1</tt:Token>
				<tt:Fixed>true</tt:Fixed>
			</tds:Location>
		</tds:GetGeoLocationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetGeoLocation"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetGeoLocationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "DeleteGeoLocation"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:DeleteGeoLocationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetDPAddresses"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetDPAddressesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:DPAddress>
				<tt:Type>IPv4</tt:Type>
				<tt:IPv4Address>239.255.255.250</tt:IPv4Address>
			</tds:DPAddress>
			<tds:DPAddress>
				<tt:Type>IPv6</tt:Type>
				<tt:IPv6Address>ff02::c</tt:IPv6Address>
			</tds:DPAddress>
		</tds:GetDPAddressesResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetDPAddresses"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetDPAddressesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetAccessPolicy"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetAccessPolicyResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:PolicyFile>
				<tt:Data>cG9saWN5IGRhdGE=</tt:Data>
				<tt:ContentType>application/xml</tt:ContentType>
			</tds:PolicyFile>
		</tds:GetAccessPolicyResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetAccessPolicy"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetAccessPolicyResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetWsdlUrl"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetWsdlUrlResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:WsdlUrl>http://192.168.1.100/onvif/device.wsdl</tds:WsdlUrl>
		</tds:GetWsdlUrlResponse>
	</s:Body>
</s:Envelope>`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetGeoLocation(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	locations, err := client.GetGeoLocation(ctx)
	if err != nil {
		t.Fatalf("GetGeoLocation failed: %v", err)
	}

	if len(locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(locations))
	}

	loc := locations[0]
	if loc.Entity != "Building A" {
		t.Errorf("Expected entity 'Building A', got %s", loc.Entity)
	}

	if loc.Token != "location1" {
		t.Errorf("Expected token 'location1', got %s", loc.Token)
	}

	if !loc.Fixed {
		t.Error("Expected Fixed to be true")
	}

	// Check coordinates (approximate comparison due to float precision)
	if loc.Lon < -122.42 || loc.Lon > -122.41 {
		t.Errorf("Expected longitude around -122.4194, got %f", loc.Lon)
	}

	if loc.Lat < 37.77 || loc.Lat > 37.78 {
		t.Errorf("Expected latitude around 37.7749, got %f", loc.Lat)
	}

	if loc.Elevation < 10.0 || loc.Elevation > 11.0 {
		t.Errorf("Expected elevation around 10.5, got %f", loc.Elevation)
	}
}

func TestSetGeoLocation(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	locations := []LocationEntity{
		{
			Entity:    "Main Office",
			Token:     "loc1",
			Fixed:     true,
			Lon:       -122.4194,
			Lat:       37.7749,
			Elevation: 15.0,
		},
	}

	err = client.SetGeoLocation(ctx, locations)
	if err != nil {
		t.Fatalf("SetGeoLocation failed: %v", err)
	}
}

func TestDeleteGeoLocation(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	locations := []LocationEntity{
		{Token: "location1"},
	}

	err = client.DeleteGeoLocation(ctx, locations)
	if err != nil {
		t.Fatalf("DeleteGeoLocation failed: %v", err)
	}
}

func TestGetDPAddresses(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	addresses, err := client.GetDPAddresses(ctx)
	if err != nil {
		t.Fatalf("GetDPAddresses failed: %v", err)
	}

	if len(addresses) != 2 {
		t.Fatalf("Expected 2 addresses, got %d", len(addresses))
	}

	// Check IPv4 address
	if addresses[0].Type != "IPv4" {
		t.Errorf("Expected Type 'IPv4', got %s", addresses[0].Type)
	}
	if addresses[0].IPv4Address != "239.255.255.250" {
		t.Errorf("Expected IPv4 address '239.255.255.250', got %s", addresses[0].IPv4Address)
	}

	// Check IPv6 address
	if addresses[1].Type != "IPv6" {
		t.Errorf("Expected Type 'IPv6', got %s", addresses[1].Type)
	}
	if addresses[1].IPv6Address != "ff02::c" {
		t.Errorf("Expected IPv6 address 'ff02::c', got %s", addresses[1].IPv6Address)
	}
}

func TestSetDPAddresses(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	addresses := []NetworkHost{
		{
			Type:        "IPv4",
			IPv4Address: "239.255.255.250",
		},
	}

	err = client.SetDPAddresses(ctx, addresses)
	if err != nil {
		t.Fatalf("SetDPAddresses failed: %v", err)
	}
}

func TestGetAccessPolicy(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	policy, err := client.GetAccessPolicy(ctx)
	if err != nil {
		t.Fatalf("GetAccessPolicy failed: %v", err)
	}

	if policy == nil || policy.PolicyFile == nil {
		t.Fatal("Expected policy file, got nil")
	}

	if policy.PolicyFile.ContentType != "application/xml" {
		t.Errorf("Expected content type 'application/xml', got %s", policy.PolicyFile.ContentType)
	}
}

func TestSetAccessPolicy(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	policy := &AccessPolicy{
		PolicyFile: &BinaryData{
			Data:        []byte("policy data"),
			ContentType: "application/xml",
		},
	}

	err = client.SetAccessPolicy(ctx, policy)
	if err != nil {
		t.Fatalf("SetAccessPolicy failed: %v", err)
	}
}

func TestGetWsdlUrl(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	url, err := client.GetWsdlURL(ctx)
	if err != nil {
		t.Fatalf("GetWsdlURL failed: %v", err)
	}

	expected := "http://192.168.1.100/onvif/device.wsdl"
	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}
