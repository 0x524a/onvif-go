package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockDeviceStorageServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")

		// Parse request to determine which operation
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		requestBody := string(buf)

		var response string

		switch {
		case strings.Contains(requestBody, "GetStorageConfigurations"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetStorageConfigurationsResponse>
      <tds:StorageConfigurations token="storage-001">
        <tt:Data type="NFS">
          <tt:LocalPath>/var/media/storage1</tt:LocalPath>
          <tt:StorageUri>file:///var/media/storage1</tt:StorageUri>
        </tt:Data>
      </tds:StorageConfigurations>
      <tds:StorageConfigurations token="storage-002">
        <tt:Data type="CIFS">
          <tt:LocalPath>/var/media/storage2</tt:LocalPath>
          <tt:StorageUri>cifs://nas.local/recordings</tt:StorageUri>
        </tt:Data>
      </tds:StorageConfigurations>
    </tds:GetStorageConfigurationsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetStorageConfiguration"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetStorageConfigurationResponse>
      <tds:StorageConfiguration token="storage-001">
        <tt:Data type="NFS">
          <tt:LocalPath>/var/media/storage1</tt:LocalPath>
          <tt:StorageUri>file:///var/media/storage1</tt:StorageUri>
        </tt:Data>
      </tds:StorageConfiguration>
    </tds:GetStorageConfigurationResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "CreateStorageConfiguration"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:CreateStorageConfigurationResponse>
      <tds:Token>storage-new</tds:Token>
    </tds:CreateStorageConfigurationResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "SetStorageConfiguration"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetStorageConfigurationResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "DeleteStorageConfiguration"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:DeleteStorageConfigurationResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "SetHashingAlgorithm"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetHashingAlgorithmResponse/>
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

func TestGetStorageConfigurations(t *testing.T) {
	server := newMockDeviceStorageServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	configs, err := client.GetStorageConfigurations(ctx)
	if err != nil {
		t.Fatalf("GetStorageConfigurations failed: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("Expected 2 storage configurations, got %d", len(configs))
	}

	if configs[0].Token != "storage-001" {
		t.Errorf("Expected first config token 'storage-001', got '%s'", configs[0].Token)
	}

	if configs[0].Data.LocalPath != "/var/media/storage1" {
		t.Errorf("Expected first config path '/var/media/storage1', got '%s'", configs[0].Data.LocalPath)
	}

	if configs[0].Data.Type != "NFS" {
		t.Errorf("Expected first config type 'NFS', got '%s'", configs[0].Data.Type)
	}

	if configs[1].Token != "storage-002" {
		t.Errorf("Expected second config token 'storage-002', got '%s'", configs[1].Token)
	}

	if configs[1].Data.StorageURI != "cifs://nas.local/recordings" {
		t.Errorf("Expected second config URI 'cifs://nas.local/recordings', got '%s'", configs[1].Data.StorageURI)
	}
}

func TestGetStorageConfiguration(t *testing.T) {
	server := newMockDeviceStorageServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	config, err := client.GetStorageConfiguration(ctx, "storage-001")
	if err != nil {
		t.Fatalf("GetStorageConfiguration failed: %v", err)
	}

	if config.Token != "storage-001" {
		t.Errorf("Expected config token 'storage-001', got '%s'", config.Token)
	}

	if config.Data.LocalPath != "/var/media/storage1" {
		t.Errorf("Expected config path '/var/media/storage1', got '%s'", config.Data.LocalPath)
	}

	if config.Data.StorageURI != "file:///var/media/storage1" {
		t.Errorf("Expected config URI 'file:///var/media/storage1', got '%s'", config.Data.StorageURI)
	}

	if config.Data.Type != "NFS" {
		t.Errorf("Expected config type 'NFS', got '%s'", config.Data.Type)
	}
}

func TestCreateStorageConfiguration(t *testing.T) {
	server := newMockDeviceStorageServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	config := &StorageConfiguration{
		Token: "storage-new",
		Data: StorageConfigurationData{
			LocalPath:  "/var/media/storage3",
			StorageURI: "file:///var/media/storage3",
			Type:       "Local",
		},
	}

	token, err := client.CreateStorageConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("CreateStorageConfiguration failed: %v", err)
	}

	if token != "storage-new" {
		t.Errorf("Expected token 'storage-new', got '%s'", token)
	}
}

func TestSetStorageConfiguration(t *testing.T) {
	server := newMockDeviceStorageServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	config := &StorageConfiguration{
		Token: "storage-001",
		Data: StorageConfigurationData{
			LocalPath:  "/var/media/updated",
			StorageURI: "file:///var/media/updated",
			Type:       "NFS",
		},
	}

	err = client.SetStorageConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetStorageConfiguration failed: %v", err)
	}
}

// TestSetStorageConfigurationWireFormat asserts the outbound payload matches the
// ONVIF schema: token and type are attributes, and the URI element is StorageUri.
func TestSetStorageConfigurationWireFormat(t *testing.T) {
	var requestBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		requestBody = string(body)

		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetStorageConfigurationResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	config := &StorageConfiguration{
		Token: "storage-001",
		Data: StorageConfigurationData{
			LocalPath:  "/var/media/updated",
			StorageURI: "file:///var/media/updated",
			Type:       "NFS",
		},
	}

	if err := client.SetStorageConfiguration(context.Background(), config); err != nil {
		t.Fatalf("SetStorageConfiguration failed: %v", err)
	}

	for _, want := range []string{
		`token="storage-001"`,
		`type="NFS"`,
		"<StorageUri>file:///var/media/updated</StorageUri>",
	} {
		if !strings.Contains(requestBody, want) {
			t.Errorf("request body missing %s\nbody:\n%s", want, requestBody)
		}
	}

	for _, unwanted := range []string{
		"<Token>",
		"<Type>",
		"<StorageURI>",
	} {
		if strings.Contains(requestBody, unwanted) {
			t.Errorf("request body contains non-spec element %s\nbody:\n%s", unwanted, requestBody)
		}
	}
}

func TestDeleteStorageConfiguration(t *testing.T) {
	server := newMockDeviceStorageServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	err = client.DeleteStorageConfiguration(ctx, "storage-old")
	if err != nil {
		t.Fatalf("DeleteStorageConfiguration failed: %v", err)
	}
}

func TestSetHashingAlgorithm(t *testing.T) {
	server := newMockDeviceStorageServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	err = client.SetHashingAlgorithm(ctx, "SHA-256")
	if err != nil {
		t.Fatalf("SetHashingAlgorithm failed: %v", err)
	}
}
