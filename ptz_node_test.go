package onvif

import (
	"context"
	"encoding/xml"
	"math"
	"strings"
	"testing"
	"time"
)

// fullSpacesXML lists one entry of every PTZ space kind, each with distinct
// numbers, so a swapped or dropped field is caught.
const fullSpacesXML = `
            <tt:AbsolutePanTiltPositionSpace>
                <tt:URI>urn:abs-pt</tt:URI>
                <tt:XRange><tt:Min>-1</tt:Min><tt:Max>1</tt:Max></tt:XRange>
                <tt:YRange><tt:Min>-0.5</tt:Min><tt:Max>0.75</tt:Max></tt:YRange>
            </tt:AbsolutePanTiltPositionSpace>
            <tt:AbsolutePanTiltPositionSpace>
                <tt:URI>urn:abs-pt-degrees</tt:URI>
                <tt:XRange><tt:Min>-180</tt:Min><tt:Max>180</tt:Max></tt:XRange>
                <tt:YRange><tt:Min>-90</tt:Min><tt:Max>0</tt:Max></tt:YRange>
            </tt:AbsolutePanTiltPositionSpace>
            <tt:AbsoluteZoomPositionSpace>
                <tt:URI>urn:abs-z</tt:URI>
                <tt:XRange><tt:Min>0</tt:Min><tt:Max>0.9</tt:Max></tt:XRange>
            </tt:AbsoluteZoomPositionSpace>
            <tt:RelativePanTiltTranslationSpace>
                <tt:URI>urn:rel-pt</tt:URI>
                <tt:XRange><tt:Min>-2</tt:Min><tt:Max>2</tt:Max></tt:XRange>
                <tt:YRange><tt:Min>-3</tt:Min><tt:Max>3</tt:Max></tt:YRange>
            </tt:RelativePanTiltTranslationSpace>
            <tt:RelativeZoomTranslationSpace>
                <tt:URI>urn:rel-z</tt:URI>
                <tt:XRange><tt:Min>-4</tt:Min><tt:Max>4</tt:Max></tt:XRange>
            </tt:RelativeZoomTranslationSpace>
            <tt:ContinuousPanTiltVelocitySpace>
                <tt:URI>urn:vel-pt</tt:URI>
                <tt:XRange><tt:Min>-5</tt:Min><tt:Max>5</tt:Max></tt:XRange>
                <tt:YRange><tt:Min>-6</tt:Min><tt:Max>6</tt:Max></tt:YRange>
            </tt:ContinuousPanTiltVelocitySpace>
            <tt:ContinuousZoomVelocitySpace>
                <tt:URI>urn:vel-z</tt:URI>
                <tt:XRange><tt:Min>-7</tt:Min><tt:Max>7</tt:Max></tt:XRange>
            </tt:ContinuousZoomVelocitySpace>
            <tt:PanTiltSpeedSpace>
                <tt:URI>urn:speed-pt</tt:URI>
                <tt:XRange><tt:Min>0</tt:Min><tt:Max>0.8</tt:Max></tt:XRange>
            </tt:PanTiltSpeedSpace>
            <tt:ZoomSpeedSpace>
                <tt:URI>urn:speed-z</tt:URI>
                <tt:XRange><tt:Min>0.1</tt:Min><tt:Max>0.6</tt:Max></tt:XRange>
            </tt:ZoomSpeedSpace>`

const fullNodeXML = `
    <tptz:PTZNode token="node1" FixedHomePosition="false" GeoMove="true">
        <tt:Name>Dome</tt:Name>
        <tt:SupportedPTZSpaces>` + fullSpacesXML + `
        </tt:SupportedPTZSpaces>
        <tt:MaximumNumberOfPresets>128</tt:MaximumNumberOfPresets>
        <tt:HomeSupported>true</tt:HomeSupported>
        <tt:AuxiliaryCommands>tt:Wiper|On</tt:AuxiliaryCommands>
        <tt:AuxiliaryCommands>tt:Wiper|Off</tt:AuxiliaryCommands>
    </tptz:PTZNode>`

const ptzReplyNS = `xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"`

func assertRange(t *testing.T, what string, r *FloatRange, wantMin, wantMax float64) {
	t.Helper()

	if r == nil || r.Min != wantMin || r.Max != wantMax {
		t.Errorf("%s = %+v, want {%g %g}", what, r, wantMin, wantMax)
	}
}

func assert2D(t *testing.T, what string, got []*Space2DDescription, uri string, x, y [2]float64) {
	t.Helper()

	if len(got) == 0 || got[0] == nil {
		t.Fatalf("%s = %+v, want one entry", what, got)
	}
	if got[0].URI != uri {
		t.Errorf("%s URI = %q, want %q", what, got[0].URI, uri)
	}
	assertRange(t, what+" XRange", got[0].XRange, x[0], x[1])
	assertRange(t, what+" YRange", got[0].YRange, y[0], y[1])
}

func assert1D(t *testing.T, what string, got []*Space1DDescription, uri string, x [2]float64) {
	t.Helper()

	if len(got) != 1 || got[0] == nil {
		t.Fatalf("%s = %+v, want one entry", what, got)
	}
	if got[0].URI != uri {
		t.Errorf("%s URI = %q, want %q", what, got[0].URI, uri)
	}
	assertRange(t, what+" XRange", got[0].XRange, x[0], x[1])
}

// assertFullSpaces checks every field of fullSpacesXML.
func assertFullSpaces(t *testing.T, s *PTZSpaces) {
	t.Helper()

	if s == nil {
		t.Fatal("spaces = nil, want the full set")
	}
	if len(s.AbsolutePanTiltPositionSpace) != 2 {
		t.Fatalf("AbsolutePanTiltPositionSpace has %d entries, want 2", len(s.AbsolutePanTiltPositionSpace))
	}
	assert2D(t, "AbsolutePanTiltPositionSpace[0]", s.AbsolutePanTiltPositionSpace, "urn:abs-pt", [2]float64{-1, 1}, [2]float64{-0.5, 0.75})
	assert2D(t, "AbsolutePanTiltPositionSpace[1]", s.AbsolutePanTiltPositionSpace[1:], "urn:abs-pt-degrees",
		[2]float64{-180, 180}, [2]float64{-90, 0})
	assert1D(t, "AbsoluteZoomPositionSpace", s.AbsoluteZoomPositionSpace, "urn:abs-z", [2]float64{0, 0.9})
	assert2D(t, "RelativePanTiltTranslationSpace", s.RelativePanTiltTranslationSpace, "urn:rel-pt", [2]float64{-2, 2}, [2]float64{-3, 3})
	assert1D(t, "RelativeZoomTranslationSpace", s.RelativeZoomTranslationSpace, "urn:rel-z", [2]float64{-4, 4})
	assert2D(t, "ContinuousPanTiltVelocitySpace", s.ContinuousPanTiltVelocitySpace, "urn:vel-pt", [2]float64{-5, 5}, [2]float64{-6, 6})
	assert1D(t, "ContinuousZoomVelocitySpace", s.ContinuousZoomVelocitySpace, "urn:vel-z", [2]float64{-7, 7})
	assert1D(t, "PanTiltSpeedSpace", s.PanTiltSpeedSpace, "urn:speed-pt", [2]float64{0, 0.8})
	assert1D(t, "ZoomSpeedSpace", s.ZoomSpeedSpace, "urn:speed-z", [2]float64{0.1, 0.6})
}

func assertFullNode(t *testing.T, n *PTZNode) {
	t.Helper()

	if n == nil {
		t.Fatal("node = nil")
	}
	if n.Token != "node1" || n.Name != "Dome" {
		t.Errorf("Token, Name = %q, %q, want node1, Dome", n.Token, n.Name)
	}
	if n.FixedHomePosition == nil || *n.FixedHomePosition {
		t.Errorf("FixedHomePosition = %v, want a reported false", n.FixedHomePosition)
	}
	if !n.GeoMove || !n.HomeSupported || n.MaximumNumberOfPresets != 128 {
		t.Errorf("GeoMove, HomeSupported, MaximumNumberOfPresets = %v, %v, %d, want true, true, 128",
			n.GeoMove, n.HomeSupported, n.MaximumNumberOfPresets)
	}
	if len(n.AuxiliaryCommands) != 2 || n.AuxiliaryCommands[0] != "tt:Wiper|On" || n.AuxiliaryCommands[1] != "tt:Wiper|Off" {
		t.Errorf("AuxiliaryCommands = %q", n.AuxiliaryCommands)
	}
	assertFullSpaces(t, n.SupportedPTZSpaces)
}

func TestGetNodeMapsEveryField(t *testing.T) {
	body := `<tptz:GetNodeResponse ` + ptzReplyNS + `>` + fullNodeXML + `</tptz:GetNodeResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	node, err := client.GetNode(context.Background(), "node1")
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}
	assertFullNode(t, node)
}

func TestGetNodesMapsEveryNode(t *testing.T) {
	second := `<tptz:PTZNode token="node2"><tt:Name>Fixed</tt:Name></tptz:PTZNode>`
	body := `<tptz:GetNodesResponse ` + ptzReplyNS + `>` + fullNodeXML + second + `</tptz:GetNodesResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	nodes, err := client.GetNodes(context.Background())
	if err != nil {
		t.Fatalf("GetNodes() error = %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("GetNodes() returned %d nodes, want 2", len(nodes))
	}
	assertFullNode(t, nodes[0])
	if nodes[1].Token != "node2" || nodes[1].Name != "Fixed" {
		t.Errorf("nodes[1] = %+v, want node2 Fixed", nodes[1])
	}
}

// A node that reports nothing optional must come back with nil and zero values,
// never with invented ranges: a caller that fails closed on absent limits relies
// on absent meaning absent.
func TestGetNodeOptionalsAbsent(t *testing.T) {
	body := `<tptz:GetNodeResponse ` + ptzReplyNS + `>
        <tptz:PTZNode token="bare">
            <tt:SupportedPTZSpaces>
                <tt:AbsolutePanTiltPositionSpace><tt:URI>urn:no-ranges</tt:URI></tt:AbsolutePanTiltPositionSpace>
            </tt:SupportedPTZSpaces>
        </tptz:PTZNode>
    </tptz:GetNodeResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	node, err := client.GetNode(context.Background(), "bare")
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}
	if node.FixedHomePosition != nil || node.GeoMove || node.HomeSupported || node.MaximumNumberOfPresets != 0 ||
		node.AuxiliaryCommands != nil {
		t.Errorf("node = %+v, want every optional field absent", node)
	}
	s := node.SupportedPTZSpaces
	if s == nil || len(s.AbsolutePanTiltPositionSpace) != 1 {
		t.Fatalf("spaces = %+v, want one absolute pan/tilt space", s)
	}
	if sp := s.AbsolutePanTiltPositionSpace[0]; sp.URI != "urn:no-ranges" || sp.XRange != nil || sp.YRange != nil {
		t.Errorf("space = %+v, want the URI and nil ranges", sp)
	}
	if s.AbsoluteZoomPositionSpace != nil || s.PanTiltSpeedSpace != nil || s.ZoomSpeedSpace != nil {
		t.Errorf("spaces = %+v, want unreported kinds nil", s)
	}
}

func TestGetNodeWithoutSpacesHasNilSpaces(t *testing.T) {
	body := `<tptz:GetNodeResponse ` + ptzReplyNS + `><tptz:PTZNode token="n"/></tptz:GetNodeResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	node, err := client.GetNode(context.Background(), "n")
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}
	if node.SupportedPTZSpaces != nil {
		t.Errorf("SupportedPTZSpaces = %+v, want nil", node.SupportedPTZSpaces)
	}
}

func TestGetConfigurationOptionsMapsEveryField(t *testing.T) {
	body := `<tptz:GetConfigurationOptionsResponse ` + ptzReplyNS + `>
        <tptz:PTZConfigurationOptions>
            <tt:Spaces>` + fullSpacesXML + `</tt:Spaces>
            <tt:PTZTimeout><tt:Min>PT1S</tt:Min><tt:Max>PT1M30S</tt:Max></tt:PTZTimeout>
        </tptz:PTZConfigurationOptions>
    </tptz:GetConfigurationOptionsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	opts, err := client.GetConfigurationOptions(context.Background(), "cfg1")
	if err != nil {
		t.Fatalf("GetConfigurationOptions() error = %v", err)
	}
	assertFullSpaces(t, opts.Spaces)
	if opts.PTZTimeout == nil || opts.PTZTimeout.Min != time.Second || opts.PTZTimeout.Max != 90*time.Second {
		t.Errorf("PTZTimeout = %+v, want {1s 1m30s}", opts.PTZTimeout)
	}
}

func TestGetConfigurationOptionsOptionalsAbsent(t *testing.T) {
	body := `<tptz:GetConfigurationOptionsResponse ` + ptzReplyNS + `>
        <tptz:PTZConfigurationOptions/>
    </tptz:GetConfigurationOptionsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	opts, err := client.GetConfigurationOptions(context.Background(), "cfg1")
	if err != nil {
		t.Fatalf("GetConfigurationOptions() error = %v", err)
	}
	if opts.Spaces != nil || opts.PTZTimeout != nil {
		t.Errorf("options = %+v, want nil spaces and timeout", opts)
	}
}

// Numbers a camera writes as INF or NaN are passed through as such, not turned
// into an error or a zero: the caller decides what a non-finite limit means.
func TestGetConfigurationOptionsPassesNonFiniteNumbersThrough(t *testing.T) {
	body := `<tptz:GetConfigurationOptionsResponse ` + ptzReplyNS + `>
        <tptz:PTZConfigurationOptions><tt:Spaces>
            <tt:AbsoluteZoomPositionSpace><tt:URI>urn:z</tt:URI>
                <tt:XRange><tt:Min>-INF</tt:Min><tt:Max>NaN</tt:Max></tt:XRange>
            </tt:AbsoluteZoomPositionSpace>
        </tt:Spaces></tptz:PTZConfigurationOptions>
    </tptz:GetConfigurationOptionsResponse>`
	client := newPTZTestClient(t, newSOAPTestServer(t, body))

	opts, err := client.GetConfigurationOptions(context.Background(), "cfg1")
	if err != nil {
		t.Fatalf("GetConfigurationOptions() error = %v", err)
	}
	r := opts.Spaces.AbsoluteZoomPositionSpace[0].XRange
	if r == nil || !math.IsInf(r.Min, -1) || !math.IsNaN(r.Max) {
		t.Errorf("XRange = %+v, want {-Inf NaN}", r)
	}
}

// The three new requests must carry their operation and token in the tptz
// namespace: strict (gSOAP) cameras parse namespace-aware and reject anything
// else. This decodes the captured request the way such a camera does.
func TestPTZNodeRequestsAreInThePTZNamespace(t *testing.T) {
	cases := []struct {
		op, child, token string
		call             func(*Client) error
	}{
		{"GetNodes", "", "", func(c *Client) error {
			_, err := c.GetNodes(context.Background())

			return err
		}},
		{"GetNode", "NodeToken", "node-sentinel", func(c *Client) error {
			_, err := c.GetNode(context.Background(), "node-sentinel")

			return err
		}},
		{"GetConfigurationOptions", "ConfigurationToken", "cfg-sentinel", func(c *Client) error {
			_, err := c.GetConfigurationOptions(context.Background(), "cfg-sentinel")

			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.op, func(t *testing.T) {
			var captured string
			reply := `<tptz:` + tc.op + `Response ` + ptzReplyNS + `/>`
			client := newPTZTestClient(t, newRequestCapturingPTZServer(t, reply, &captured))
			_ = tc.call(client)
			assertPTZNamespaced(t, captured, tc.op, tc.child, tc.token)
		})
	}
}

// assertPTZNamespaced decodes body namespace-aware and checks that its root is
// op in the PTZ namespace and that child (when set) is a PTZ-namespace child
// holding token.
func assertPTZNamespaced(t *testing.T, body, op, child, token string) {
	t.Helper()

	dec := xml.NewDecoder(strings.NewReader(body))
	depth, sawRoot, sawChild := 0, false, false
	var inChild bool
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch e := tok.(type) {
		case xml.StartElement:
			depth++
			if e.Name.Space != ptzNamespace {
				t.Errorf("<%s> is in namespace %q, want %q\n%s", e.Name.Local, e.Name.Space, ptzNamespace, body)
			}
			sawRoot = sawRoot || (depth == 1 && e.Name.Local == op)
			inChild = depth == 2 && e.Name.Local == child
		case xml.CharData:
			if inChild && strings.TrimSpace(string(e)) == token {
				sawChild = true
			}
		case xml.EndElement:
			depth--
			inChild = false
		}
	}
	if !sawRoot {
		t.Errorf("request root is not %s:\n%s", op, body)
	}
	if child != "" && !sawChild {
		t.Errorf("request lacks <%s>%s</%s>:\n%s", child, token, child, body)
	}
}
