package onvif

import (
	"encoding/xml"
	"strings"
	"testing"
)

// Strict (gSOAP) cameras reject PanTilt and Zoom written without the tt
// namespace: "tag name namespace mismatch in element 'PanTilt'". This parses
// the marshaled request namespace-aware, like such a camera does.
func TestPTZRequestChildrenAreInTheSchemaNamespace(t *testing.T) {
	type req struct {
		XMLName     xml.Name         `xml:"tptz:RelativeMove"`
		Xmlns       string           `xml:"xmlns:tptz,attr"`
		Translation *ptzVectorReqXML `xml:"tptz:Translation"`
		Speed       *ptzSpeedReqXML  `xml:"tptz:Speed,omitempty"`
	}
	v := &PTZVector{PanTilt: &Vector2D{X: 0.05, Y: 0, Space: "urn:space"}, Zoom: &Vector1D{X: 0.1}}
	s := &PTZSpeed{PanTilt: &Vector2D{X: 0.5, Y: 0.5}}
	out, err := xml.Marshal(req{Xmlns: ptzNamespace, Translation: convertToPTZVectorXML(v), Speed: convertToPTZSpeedXML(s)})
	if err != nil {
		t.Fatal(err)
	}
	dec := xml.NewDecoder(strings.NewReader(string(out)))
	seen := map[string]bool{}
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		if se, ok := tok.(xml.StartElement); ok && (se.Name.Local == "PanTilt" || se.Name.Local == "Zoom") {
			if se.Name.Space != schemaNamespace {
				t.Errorf("%s is in namespace %q, want %q\n%s", se.Name.Local, se.Name.Space, schemaNamespace, out)
			}
			seen[se.Name.Local] = true
		}
	}
	if !seen["PanTilt"] || !seen["Zoom"] {
		t.Fatalf("missing children in %s", out)
	}
}
