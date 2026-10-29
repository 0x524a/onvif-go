package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// ptzSpace2DXML and ptzSpace1DXML are the wire forms of a space description.
type ptzSpace2DXML struct {
	URI    string            `xml:"URI"`
	XRange *ptzFloatRangeXML `xml:"XRange"`
	YRange *ptzFloatRangeXML `xml:"YRange"`
}

type ptzSpace1DXML struct {
	URI    string            `xml:"URI"`
	XRange *ptzFloatRangeXML `xml:"XRange"`
}

// ptzSpacesXML is the wire form of tt:PTZSpaces, shared by tt:PTZNode
// (SupportedPTZSpaces) and tt:PTZConfigurationOptions (Spaces).
type ptzSpacesXML struct {
	AbsolutePanTiltPositionSpace    []ptzSpace2DXML `xml:"AbsolutePanTiltPositionSpace"`
	AbsoluteZoomPositionSpace       []ptzSpace1DXML `xml:"AbsoluteZoomPositionSpace"`
	RelativePanTiltTranslationSpace []ptzSpace2DXML `xml:"RelativePanTiltTranslationSpace"`
	RelativeZoomTranslationSpace    []ptzSpace1DXML `xml:"RelativeZoomTranslationSpace"`
	ContinuousPanTiltVelocitySpace  []ptzSpace2DXML `xml:"ContinuousPanTiltVelocitySpace"`
	ContinuousZoomVelocitySpace     []ptzSpace1DXML `xml:"ContinuousZoomVelocitySpace"`
	PanTiltSpeedSpace               []ptzSpace1DXML `xml:"PanTiltSpeedSpace"`
	ZoomSpeedSpace                  []ptzSpace1DXML `xml:"ZoomSpeedSpace"`
}

// ptzNodeXML is the wire form of tt:PTZNode.
type ptzNodeXML struct {
	Token                  string        `xml:"token,attr"`
	FixedHomePosition      *bool         `xml:"FixedHomePosition,attr"`
	GeoMove                bool          `xml:"GeoMove,attr"`
	Name                   string        `xml:"Name"`
	SupportedPTZSpaces     *ptzSpacesXML `xml:"SupportedPTZSpaces"`
	MaximumNumberOfPresets int           `xml:"MaximumNumberOfPresets"`
	HomeSupported          bool          `xml:"HomeSupported"`
	AuxiliaryCommands      []string      `xml:"AuxiliaryCommands"`
}

func toSpaces2D(in []ptzSpace2DXML) []*Space2DDescription {
	if in == nil {
		return nil
	}
	out := make([]*Space2DDescription, len(in))
	for i, s := range in {
		out[i] = &Space2DDescription{URI: s.URI, XRange: toFloatRange(s.XRange), YRange: toFloatRange(s.YRange)}
	}

	return out
}

func toSpaces1D(in []ptzSpace1DXML) []*Space1DDescription {
	if in == nil {
		return nil
	}
	out := make([]*Space1DDescription, len(in))
	for i, s := range in {
		out[i] = &Space1DDescription{URI: s.URI, XRange: toFloatRange(s.XRange)}
	}

	return out
}

// toPTZSpaces converts the wire form, preserving nil for an absent element.
func (x *ptzSpacesXML) toPTZSpaces() *PTZSpaces {
	if x == nil {
		return nil
	}

	return &PTZSpaces{
		AbsolutePanTiltPositionSpace:    toSpaces2D(x.AbsolutePanTiltPositionSpace),
		AbsoluteZoomPositionSpace:       toSpaces1D(x.AbsoluteZoomPositionSpace),
		RelativePanTiltTranslationSpace: toSpaces2D(x.RelativePanTiltTranslationSpace),
		RelativeZoomTranslationSpace:    toSpaces1D(x.RelativeZoomTranslationSpace),
		ContinuousPanTiltVelocitySpace:  toSpaces2D(x.ContinuousPanTiltVelocitySpace),
		ContinuousZoomVelocitySpace:     toSpaces1D(x.ContinuousZoomVelocitySpace),
		PanTiltSpeedSpace:               toSpaces1D(x.PanTiltSpeedSpace),
		ZoomSpeedSpace:                  toSpaces1D(x.ZoomSpeedSpace),
	}
}

func (x *ptzNodeXML) toPTZNode() *PTZNode {
	return &PTZNode{
		Token:                  x.Token,
		Name:                   x.Name,
		FixedHomePosition:      x.FixedHomePosition,
		GeoMove:                x.GeoMove,
		SupportedPTZSpaces:     x.SupportedPTZSpaces.toPTZSpaces(),
		MaximumNumberOfPresets: x.MaximumNumberOfPresets,
		HomeSupported:          x.HomeSupported,
		AuxiliaryCommands:      x.AuxiliaryCommands,
	}
}

// GetNodes retrieves every PTZ node of the device, with the coordinate spaces
// and ranges each supports.
func (c *Client) GetNodes(ctx context.Context) ([]*PTZNode, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return nil, err
	}

	type GetNodes struct {
		XMLName xml.Name `xml:"tptz:GetNodes"`
		Xmlns   string   `xml:"xmlns:tptz,attr"`
	}

	type GetNodesResponse struct {
		XMLName xml.Name     `xml:"GetNodesResponse"`
		PTZNode []ptzNodeXML `xml:"PTZNode"`
	}

	var resp GetNodesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", GetNodes{Xmlns: ptzNamespace}, &resp); err != nil {
		return nil, fmt.Errorf("GetNodes failed: %w", err)
	}

	nodes := make([]*PTZNode, len(resp.PTZNode))
	for i := range resp.PTZNode {
		nodes[i] = resp.PTZNode[i].toPTZNode()
	}

	return nodes, nil
}

// GetNode retrieves one PTZ node by its token (a PTZConfiguration's
// NodeToken), with the coordinate spaces and ranges it supports.
func (c *Client) GetNode(ctx context.Context, nodeToken string) (*PTZNode, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return nil, err
	}

	type GetNode struct {
		XMLName   xml.Name `xml:"tptz:GetNode"`
		Xmlns     string   `xml:"xmlns:tptz,attr"`
		NodeToken string   `xml:"tptz:NodeToken"`
	}

	type GetNodeResponse struct {
		XMLName xml.Name    `xml:"GetNodeResponse"`
		PTZNode *ptzNodeXML `xml:"PTZNode"`
	}

	var resp GetNodeResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", GetNode{Xmlns: ptzNamespace, NodeToken: nodeToken}, &resp); err != nil {
		return nil, fmt.Errorf("GetNode failed: %w", err)
	}

	// The schema requires the element. A reply without it is not a node with
	// no capabilities, and returning a zero-value node would let a caller read
	// it as one.
	if resp.PTZNode == nil {
		return nil, fmt.Errorf("GetNode failed: %w: reply carries no PTZNode", ErrInvalidResponse)
	}

	return resp.PTZNode.toPTZNode(), nil
}

// GetConfigurationOptions retrieves what a PTZ configuration accepts: the
// coordinate spaces valid for it, with their ranges, and the timeout range of
// a continuous move. configurationToken is a PTZConfiguration token, for
// example a profile's PTZConfiguration.Token.
func (c *Client) GetConfigurationOptions(ctx context.Context, configurationToken string) (*PTZConfigurationOptions, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return nil, err
	}

	type GetConfigurationOptions struct {
		XMLName            xml.Name `xml:"tptz:GetConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:tptz,attr"`
		ConfigurationToken string   `xml:"tptz:ConfigurationToken"`
	}

	type GetConfigurationOptionsResponse struct {
		XMLName                 xml.Name `xml:"GetConfigurationOptionsResponse"`
		PTZConfigurationOptions struct {
			Spaces     *ptzSpacesXML `xml:"Spaces"`
			PTZTimeout *struct {
				Min string `xml:"Min"`
				Max string `xml:"Max"`
			} `xml:"PTZTimeout"`
		} `xml:"PTZConfigurationOptions"`
	}

	req := GetConfigurationOptions{Xmlns: ptzNamespace, ConfigurationToken: configurationToken}

	var resp GetConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetConfigurationOptions failed: %w", err)
	}

	opts := &PTZConfigurationOptions{Spaces: resp.PTZConfigurationOptions.Spaces.toPTZSpaces()}
	// PTZTimeout is set only when both bounds parse, are not negative and are
	// ordered. A missing, malformed, overflowing or inverted bound leaves it nil
	// ("not reported") instead of a zero a caller would clamp to.
	if t := resp.PTZConfigurationOptions.PTZTimeout; t != nil {
		lo, errLo := parseXSDuration(t.Min)
		hi, errHi := parseXSDuration(t.Max)
		if errLo == nil && errHi == nil && lo >= 0 && hi >= lo {
			opts.PTZTimeout = &DurationRange{Min: lo, Max: hi}
		}
	}

	return opts, nil
}
