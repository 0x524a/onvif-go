package soap

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Wire-format tests for the canonical request shape (MiBeeNvr#1074): compact,
// s:-prefixed envelope, operation namespaces hoisted onto s:Body. Strict
// parsers in old camera firmware answer any other shape with 200 + empty body.

func TestRenderEnvelopeCanonicalShape(t *testing.T) {
	type op struct {
		XMLName  xml.Name `xml:"tds:GetCapabilities"`
		Xmlns    string   `xml:"xmlns:tds,attr"`
		Category string   `xml:"tds:Category"`
	}
	env := &Envelope{Body: Body{Content: &op{
		Xmlns:    "http://www.onvif.org/ver10/device/wsdl",
		Category: "All",
	}}}

	out, err := renderEnvelope(env)
	if err != nil {
		t.Fatalf("renderEnvelope: %v", err)
	}

	want := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">` +
		`<s:Body xmlns:tds="http://www.onvif.org/ver10/device/wsdl">` +
		`<tds:GetCapabilities><tds:Category>All</tds:Category></tds:GetCapabilities>` +
		`</s:Body></s:Envelope>`
	if string(out) != want {
		t.Errorf("wire mismatch:\n got: %s\nwant: %s", out, want)
	}
}

func TestRenderEnvelopeCompactNoWhitespace(t *testing.T) {
	type op struct {
		XMLName  xml.Name `xml:"tds:GetCapabilities"`
		Xmlns    string   `xml:"xmlns:tds,attr"`
		Category string   `xml:"tds:Category"`
	}
	env := &Envelope{Body: Body{Content: &op{
		Xmlns:    "http://www.onvif.org/ver10/device/wsdl",
		Category: "All",
	}}}
	out, err := renderEnvelope(env)
	if err != nil {
		t.Fatalf("renderEnvelope: %v", err)
	}
	if strings.ContainsRune(string(out), '\n') || strings.Contains(string(out), "  ") {
		t.Errorf("wire output must be compact, got: %s", out)
	}
}

func TestRenderEnvelopeHoistsMultipleNamespaces(t *testing.T) {
	type op struct {
		XMLName xml.Name `xml:"trt:GetStreamUri"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
		Xmlnst  string   `xml:"xmlns:tt,attr"`
		Setup   struct {
			Stream string `xml:"tt:Stream"`
		} `xml:"trt:StreamSetup"`
	}
	env := &Envelope{Body: Body{Content: &op{
		Xmlns:  "http://www.onvif.org/ver10/media/wsdl",
		Xmlnst: "http://www.onvif.org/ver10/schema",
	}}}
	env.Body.Content.(*op).Setup.Stream = "RTP-Unicast"

	out, err := renderEnvelope(env)
	if err != nil {
		t.Fatalf("renderEnvelope: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, `<s:Body xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">`) {
		t.Errorf("both namespaces must be hoisted onto s:Body, got: %s", got)
	}
	if !strings.Contains(got, `<trt:GetStreamUri>`) {
		t.Errorf("operation element must carry no namespace declaration after hoisting, got: %s", got)
	}
	if !strings.Contains(got, `<trt:GetStreamUri><trt:StreamSetup><tt:Stream>RTP-Unicast</tt:Stream></trt:StreamSetup></trt:GetStreamUri>`) {
		t.Errorf("prefixed descendants must survive hoisting unchanged, got: %s", got)
	}
}

func TestRenderEnvelopeNilBody(t *testing.T) {
	out, err := renderEnvelope(&Envelope{})
	if err != nil {
		t.Fatalf("renderEnvelope: %v", err)
	}
	want := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body/></s:Envelope>`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRenderSecurityHeaderShape(t *testing.T) {
	env := &Envelope{
		Header: &Header{Security: &Security{
			MustUnderstand: "1",
			UsernameToken: &UsernameToken{
				Username: "admin",
				Password: Password{Type: passwordDigestType, Password: "digest=="},
				Nonce:    Nonce{Type: nonceEncodingType, Nonce: "nonce=="},
				Created:  "2026-10-27T00:00:00Z",
			},
		}},
	}
	out, err := renderEnvelope(env)
	if err != nil {
		t.Fatalf("renderEnvelope: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		`<s:Header>`,
		`<wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd" s:mustUnderstand="1">`,
		`<wsse:UsernameToken>`,
		`<wsse:Username>admin</wsse:Username>`,
		`Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">digest==<`,
		`<wsse:Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">nonce==</wsse:Nonce>`,
		`<wsu:Created>2026-10-27T00:00:00Z</wsu:Created>`,
		`</wsse:UsernameToken></wsse:Security></s:Header>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("security header missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderSecurityHeaderEscapesUsername(t *testing.T) {
	env := &Envelope{
		Header: &Header{Security: &Security{
			UsernameToken: &UsernameToken{
				Username: `a<b&c>"d`,
				Password: Password{Type: passwordTextType, Password: "x"},
			},
		}},
	}
	out, err := renderEnvelope(env)
	if err != nil {
		t.Fatalf("renderEnvelope: %v", err)
	}
	if !strings.Contains(string(out), "<wsse:Username>a&lt;b&amp;c&gt;&#34;d</wsse:Username>") {
		t.Errorf("username must be XML-escaped, got: %s", out)
	}
}

// TestCallWireFormat end-to-end: the HTTP request body produced by Call must
// carry the canonical compact shape.
func TestCallWireFormat(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		captured = string(b)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><TestResponse><Value>ok</Value></TestResponse></Body></Envelope>`))
	}))
	defer srv.Close()

	type op struct {
		XMLName xml.Name `xml:"tds:GetCapabilities"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}
	type resp struct {
		Value string `xml:"TestResponse>Value"`
	}
	c := NewClient(&http.Client{Timeout: 5 * time.Second}, "", "")
	var out resp
	if err := c.Call(context.Background(), srv.URL, "", &op{Xmlns: "http://www.onvif.org/ver10/device/wsdl"}, &out); err != nil {
		t.Fatalf("Call: %v", err)
	}
	body := strings.TrimPrefix(captured, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	if !strings.HasPrefix(body, "<s:Envelope ") {
		t.Errorf("request body must open with s:Envelope, got: %s", captured)
	}
	if !strings.Contains(captured, `<s:Body xmlns:tds="http://www.onvif.org/ver10/device/wsdl"><tds:GetCapabilities>`) {
		t.Errorf("operation namespace must be hoisted onto s:Body, got: %s", captured)
	}
	if strings.Contains(body, "\n") {
		t.Errorf("request body must be compact, got: %s", captured)
	}
}
