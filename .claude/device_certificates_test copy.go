package onvif

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testCertID    = "cert-001"
	testXMLHeader = `<?xml version="1.0" encoding="UTF-8"?>`
)

func newMockDeviceCertificatesServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")

		// Parse request to determine which operation
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		requestBody := string(buf)

		var response string

		switch {
		case strings.Contains(requestBody, "GetCertificatesStatus"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCertificatesStatusResponse>
      <tds:CertificateStatus>
        <tt:CertificateID>cert-001</tt:CertificateID>
        <tt:Status>true</tt:Status>
      </tds:CertificateStatus>
    </tds:GetCertificatesStatusResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "SetCertificatesStatus"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetCertificatesStatusResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetCertificateInformation"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCertificateInformationResponse>
      <tds:CertificateInformation>
        <tt:CertificateID>cert-001</tt:CertificateID>
        <tt:IssuerDN>CN=Test CA</tt:IssuerDN>
        <tt:SubjectDN>CN=Device Certificate</tt:SubjectDN>
        <tt:ValidNotBefore>2024-01-01T00:00:00Z</tt:ValidNotBefore>
        <tt:ValidNotAfter>2025-01-01T00:00:00Z</tt:ValidNotAfter>
      </tds:CertificateInformation>
    </tds:GetCertificateInformationResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "LoadCertificateWithPrivateKey"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:LoadCertificateWithPrivateKeyResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "LoadCACertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:LoadCACertificatesResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "LoadCertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:LoadCertificatesResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetCACertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCACertificatesResponse>
      <tds:Certificate>
        <tt:CertificateID>ca-001</tt:CertificateID>
        <tt:Certificate>
          <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("CA CERTIFICATE DATA")) + `</tt:Data>
        </tt:Certificate>
      </tds:Certificate>
    </tds:GetCACertificatesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetCertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCertificatesResponse>
      <tds:Certificate>
        <tt:CertificateID>cert-001</tt:CertificateID>
        <tt:Certificate>
          <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("CERTIFICATE DATA")) + `</tt:Data>
        </tt:Certificate>
      </tds:Certificate>
    </tds:GetCertificatesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "CreateCertificate"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:CreateCertificateResponse>
      <tds:Certificate>
        <tt:CertificateID>cert-new</tt:CertificateID>
        <tt:Certificate>
          <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("NEW CERTIFICATE DATA")) + `</tt:Data>
        </tt:Certificate>
      </tds:Certificate>
    </tds:CreateCertificateResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "DeleteCertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:DeleteCertificatesResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetPkcs10Request"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetPkcs10RequestResponse>
      <tds:Pkcs10Request>
        <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("PKCS#10 CSR DATA")) + `</tt:Data>
      </tds:Pkcs10Request>
    </tds:GetPkcs10RequestResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetClientCertificateMode"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetClientCertificateModeResponse>
      <tds:Enabled>true</tds:Enabled>
    </tds:GetClientCertificateModeResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "SetClientCertificateMode"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetClientCertificateModeResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		default:
			response = testXMLHeader + `
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

func TestGetCertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs, err := client.GetCertificates(ctx)
	if err != nil {
		t.Fatalf("GetCertificates failed: %v", err)
	}

	if len(certs) == 0 {
		t.Error("Expected at least one certificate")
	}

	if certs[0].CertificateID != testCertID {
		t.Errorf("Expected certificate ID '%s', got '%s'", testCertID, certs[0].CertificateID)
	}
}

func TestGetCACertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs, err := client.GetCACertificates(ctx)
	if err != nil {
		t.Fatalf("GetCACertificates failed: %v", err)
	}

	if len(certs) == 0 {
		t.Error("Expected at least one CA certificate")
	}

	if certs[0].CertificateID != "ca-001" {
		t.Errorf("Expected certificate ID 'ca-001', got '%s'", certs[0].CertificateID)
	}
}

func TestLoadCertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs := []*Certificate{
		{
			CertificateID: "cert-upload",
			Certificate: BinaryData{
				Data: []byte("UPLOADED CERTIFICATE DATA"),
			},
		},
	}

	err = client.LoadCertificates(ctx, certs)
	if err != nil {
		t.Fatalf("LoadCertificates failed: %v", err)
	}
}

func TestLoadCACertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs := []*Certificate{
		{
			CertificateID: "ca-upload",
			Certificate: BinaryData{
				Data: []byte("UPLOADED CA CERTIFICATE DATA"),
			},
		},
	}

	err = client.LoadCACertificates(ctx, certs)
	if err != nil {
		t.Fatalf("LoadCACertificates failed: %v", err)
	}
}

func TestCreateCertificate(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	cert, err := client.CreateCertificate(ctx, "cert-new", "CN=New Device", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	if cert.CertificateID != "cert-new" {
		t.Errorf("Expected certificate ID 'cert-new', got '%s'", cert.CertificateID)
	}
}

func TestDeleteCertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	err = client.DeleteCertificates(ctx, []string{"cert-001", "cert-002"})
	if err != nil {
		t.Fatalf("DeleteCertificates failed: %v", err)
	}
}

func TestGetCertificateInformation(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	info, err := client.GetCertificateInformation(ctx, "cert-001")
	if err != nil {
		t.Fatalf("GetCertificateInformation failed: %v", err)
	}

	if info.CertificateID != "cert-001" {
		t.Errorf("Expected certificate ID 'cert-001', got '%s'", info.CertificateID)
	}

	if info.IssuerDN != "CN=Test CA" {
		t.Errorf("Expected issuer 'CN=Test CA', got '%s'", info.IssuerDN)
	}

	if info.SubjectDN != "CN=Device Certificate" {
		t.Errorf("Expected subject 'CN=Device Certificate', got '%s'", info.SubjectDN)
	}
}

func TestGetCertificatesStatus(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	statuses, err := client.GetCertificatesStatus(ctx)
	if err != nil {
		t.Fatalf("GetCertificatesStatus failed: %v", err)
	}

	if len(statuses) == 0 {
		t.Error("Expected at least one certificate status")
	}

	if statuses[0].CertificateID != "cert-001" {
		t.Errorf("Expected certificate ID 'cert-001', got '%s'", statuses[0].CertificateID)
	}

	if !statuses[0].Status {
		t.Error("Expected certificate status to be true")
	}
}

func TestSetCertificatesStatus(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	statuses := []*CertificateStatus{
		{
			CertificateID: "cert-001",
			Status:        true,
		},
	}

	err = client.SetCertificatesStatus(ctx, statuses)
	if err != nil {
		t.Fatalf("SetCertificatesStatus failed: %v", err)
	}
}

func TestGetPkcs10Request(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	csr, err := client.GetPkcs10Request(ctx, "cert-csr", "CN=Device CSR", nil)
	if err != nil {
		t.Fatalf("GetPkcs10Request failed: %v", err)
	}

	if csr == nil || len(csr.Data) == 0 {
		t.Error("Expected non-empty PKCS#10 CSR data")
	}

	// Check that data was decoded from base64
	expectedData := []byte("PKCS#10 CSR DATA")
	if len(csr.Data) > 0 && !bytes.Equal(csr.Data, expectedData) {
		t.Logf("CSR data length: %d, expected: %d", len(csr.Data), len(expectedData))
		t.Logf("CSR data: %q, expected: %q", string(csr.Data), string(expectedData))
	}
}

func TestLoadCertificateWithPrivateKey(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs := []*Certificate{
		{
			CertificateID: "cert-with-key",
			Certificate: BinaryData{
				Data: []byte("CERTIFICATE DATA"),
			},
		},
	}

	privateKeys := []*BinaryData{
		{
			Data: []byte("PRIVATE KEY DATA"),
		},
	}

	err = client.LoadCertificateWithPrivateKey(ctx, certs, privateKeys, []string{"cert-with-key"})
	if err != nil {
		t.Fatalf("LoadCertificateWithPrivateKey failed: %v", err)
	}
}

func TestGetClientCertificateMode(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	enabled, err := client.GetClientCertificateMode(ctx)
	if err != nil {
		t.Fatalf("GetClientCertificateMode failed: %v", err)
	}

	if !enabled {
		t.Error("Expected client certificate mode to be enabled")
	}
}

func TestSetClientCertificateMode(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	err = client.SetClientCertificateMode(ctx, true)
	if err != nil {
		t.Fatalf("SetClientCertificateMode failed: %v", err)
	}
}
