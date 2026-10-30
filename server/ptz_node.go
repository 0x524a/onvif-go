package server

import (
	"encoding/xml"
	"fmt"
)

// The ONVIF generic coordinate spaces. The virtual camera reports these and
// only these: it keeps one position per profile and applies whatever numbers
// a move carries, without converting between spaces.
const (
	genericPanTiltPositionSpace    = "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace"    // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
	genericZoomPositionSpace       = "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace"       // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
	genericPanTiltTranslationSpace = "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace" // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
	genericZoomTranslationSpace    = "http://www.onvif.org/ver10/tptz/ZoomSpaces/TranslationGenericSpace"    // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
	genericPanTiltVelocitySpace    = "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace"    // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
	genericZoomVelocitySpace       = "http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace"       // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
	genericPanTiltSpeedSpace       = "http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"       // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
	genericZoomSpeedSpace          = "http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"      // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed
)

// GetNodesRequest represents GetNodes request.
type GetNodesRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetNodes"`
}

// GetNodesResponse represents GetNodes response.
type GetNodesResponse struct {
	XMLName xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl GetNodesResponse"`
	PTZNode []PTZNode `xml:"PTZNode"`
}

// GetNodeRequest represents GetNode request.
type GetNodeRequest struct {
	XMLName   xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetNode"`
	NodeToken string   `xml:"NodeToken"`
}

// GetNodeResponse represents GetNode response.
type GetNodeResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetNodeResponse"`
	PTZNode PTZNode  `xml:"PTZNode"`
}

// GetConfigurationOptionsRequest represents GetConfigurationOptions request.
type GetConfigurationOptionsRequest struct {
	XMLName            xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurationOptions"`
	ConfigurationToken string   `xml:"ConfigurationToken"`
}

// GetConfigurationOptionsResponse represents GetConfigurationOptions response.
type GetConfigurationOptionsResponse struct {
	XMLName                 xml.Name                `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurationOptionsResponse"`
	PTZConfigurationOptions PTZConfigurationOptions `xml:"PTZConfigurationOptions"`
}

// GetConfigurationRequest represents GetConfiguration request.
type GetConfigurationRequest struct {
	XMLName               xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfiguration"`
	PTZConfigurationToken string   `xml:"PTZConfigurationToken"`
}

// GetConfigurationResponse represents GetConfiguration response.
type GetConfigurationResponse struct {
	XMLName          xml.Name            `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurationResponse"`
	PTZConfiguration PTZConfigurationExt `xml:"PTZConfiguration"`
}

// PTZNode represents a PTZ node.
type PTZNode struct {
	Token                  string    `xml:"token,attr"`
	Name                   string    `xml:"Name"`
	SupportedPTZSpaces     PTZSpaces `xml:"SupportedPTZSpaces"`
	MaximumNumberOfPresets int       `xml:"MaximumNumberOfPresets"`
	HomeSupported          bool      `xml:"HomeSupported"`
}

// PTZConfigurationOptions represents the options of a PTZ configuration.
type PTZConfigurationOptions struct {
	Spaces PTZSpaces `xml:"Spaces"`
	// PTZTimeout is required by the schema. The virtual camera reports a
	// nominal range and ignores a continuous move's Timeout.
	PTZTimeout PTZTimeoutRange `xml:"PTZTimeout"`
}

// PTZTimeoutRange is the xs:duration range of a continuous move's Timeout.
type PTZTimeoutRange struct {
	Min string `xml:"Min"`
	Max string `xml:"Max"`
}

// PTZSpaces lists supported coordinate spaces with their ranges.
type PTZSpaces struct {
	AbsolutePanTiltPositionSpace    []Space2DDescription `xml:"AbsolutePanTiltPositionSpace,omitempty"`
	AbsoluteZoomPositionSpace       []Space1DDescription `xml:"AbsoluteZoomPositionSpace,omitempty"`
	RelativePanTiltTranslationSpace []Space2DDescription `xml:"RelativePanTiltTranslationSpace,omitempty"`
	RelativeZoomTranslationSpace    []Space1DDescription `xml:"RelativeZoomTranslationSpace,omitempty"`
	ContinuousPanTiltVelocitySpace  []Space2DDescription `xml:"ContinuousPanTiltVelocitySpace,omitempty"`
	ContinuousZoomVelocitySpace     []Space1DDescription `xml:"ContinuousZoomVelocitySpace,omitempty"`
	PanTiltSpeedSpace               []Space1DDescription `xml:"PanTiltSpeedSpace,omitempty"`
	ZoomSpeedSpace                  []Space1DDescription `xml:"ZoomSpeedSpace,omitempty"`
}

// genericSpaces describes the generic spaces over the ranges the ONVIF PTZ
// specification defines them on, for the move kinds ptz supports.
func genericSpaces(ptz *PTZConfig) PTZSpaces {
	full := FloatRange{Min: -1, Max: 1}
	unit := FloatRange{Min: 0, Max: 1}
	pt := func(uri string) []Space2DDescription {
		return []Space2DDescription{{URI: uri, XRange: full, YRange: full}}
	}
	z := func(uri string, r FloatRange) []Space1DDescription {
		return []Space1DDescription{{URI: uri, XRange: r}}
	}

	s := PTZSpaces{
		PanTiltSpeedSpace: z(genericPanTiltSpeedSpace, unit),
		ZoomSpeedSpace:    z(genericZoomSpeedSpace, unit),
	}
	if ptz.SupportsAbsolute {
		s.AbsolutePanTiltPositionSpace = pt(genericPanTiltPositionSpace)
		s.AbsoluteZoomPositionSpace = z(genericZoomPositionSpace, unit)
	}
	if ptz.SupportsRelative {
		s.RelativePanTiltTranslationSpace = pt(genericPanTiltTranslationSpace)
		s.RelativeZoomTranslationSpace = z(genericZoomTranslationSpace, full)
	}
	if ptz.SupportsContinuous {
		s.ContinuousPanTiltVelocitySpace = pt(genericPanTiltVelocitySpace)
		s.ContinuousZoomVelocitySpace = z(genericZoomVelocitySpace, full)
	}

	return s
}

// ptzProfileByNode returns the first profile whose PTZ node (which is also its
// PTZ configuration token, see HandleGetProfiles) is token. Profiles that share
// a node token should share their Supports* flags: the first one wins, so a
// later profile's differing flags are not reported.
func (s *Server) ptzProfileByNode(token string) *ProfileConfig {
	for i := range s.config.Profiles {
		if p := &s.config.Profiles[i]; p.PTZ != nil && p.PTZ.NodeToken == token {
			return p
		}
	}

	return nil
}

func nodeOf(p *ProfileConfig) PTZNode {
	// The virtual camera stores no new presets and has no home position, so
	// its capacity is what it holds and HomeSupported is false.
	return PTZNode{
		Token:                  p.PTZ.NodeToken,
		Name:                   p.Name + " PTZ Node",
		SupportedPTZSpaces:     genericSpaces(p.PTZ),
		MaximumNumberOfPresets: len(p.PTZ.Presets),
	}
}

func configurationOf(p *ProfileConfig) PTZConfigurationExt {
	return PTZConfigurationExt{
		Token:     p.PTZ.NodeToken,
		Name:      p.Name + " PTZ",
		UseCount:  1,
		NodeToken: p.PTZ.NodeToken,
	}
}

// HandleGetNodes handles GetNodes request: one node per distinct node token.
func (s *Server) HandleGetNodes(body interface{}) (interface{}, error) {
	var req GetNodesRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp := &GetNodesResponse{}
	seen := map[string]bool{}
	for i := range s.config.Profiles {
		p := &s.config.Profiles[i]
		if p.PTZ == nil || seen[p.PTZ.NodeToken] {
			continue
		}
		seen[p.PTZ.NodeToken] = true
		resp.PTZNode = append(resp.PTZNode, nodeOf(p))
	}

	return resp, nil
}

// HandleGetNode handles GetNode request.
func (s *Server) HandleGetNode(body interface{}) (interface{}, error) {
	var req GetNodeRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	p := s.ptzProfileByNode(req.NodeToken)
	if p == nil {
		return nil, fmt.Errorf("%w: node %s", ErrPTZNotSupported, req.NodeToken)
	}

	return &GetNodeResponse{PTZNode: nodeOf(p)}, nil
}

// HandleGetConfigurationOptions handles GetConfigurationOptions request. It
// reports a nominal PTZTimeout range (PT1S to PT60S) because the schema requires
// one; the virtual camera ignores a continuous move's Timeout.
func (s *Server) HandleGetConfigurationOptions(body interface{}) (interface{}, error) {
	var req GetConfigurationOptionsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	p := s.ptzProfileByNode(req.ConfigurationToken)
	if p == nil {
		return nil, fmt.Errorf("%w: configuration %s", ErrPTZNotSupported, req.ConfigurationToken)
	}

	return &GetConfigurationOptionsResponse{
		PTZConfigurationOptions: PTZConfigurationOptions{
			Spaces:     genericSpaces(p.PTZ),
			PTZTimeout: PTZTimeoutRange{Min: "PT1S", Max: "PT60S"},
		},
	}, nil
}

// HandleGetConfiguration handles GetConfiguration request. The virtual camera
// has no configurable limits, so none are reported.
func (s *Server) HandleGetConfiguration(body interface{}) (interface{}, error) {
	var req GetConfigurationRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	p := s.ptzProfileByNode(req.PTZConfigurationToken)
	if p == nil {
		return nil, fmt.Errorf("%w: configuration %s", ErrPTZNotSupported, req.PTZConfigurationToken)
	}

	return &GetConfigurationResponse{PTZConfiguration: configurationOf(p)}, nil
}

// HandleGetConfigurations handles GetConfigurations request: one configuration
// per distinct node token.
func (s *Server) HandleGetConfigurations(_ interface{}) (interface{}, error) {
	resp := &GetConfigurationsResponse{}
	seen := map[string]bool{}
	for i := range s.config.Profiles {
		p := &s.config.Profiles[i]
		if p.PTZ == nil || seen[p.PTZ.NodeToken] {
			continue
		}
		seen[p.PTZ.NodeToken] = true
		resp.PTZConfiguration = append(resp.PTZConfiguration, configurationOf(p))
	}

	return resp, nil
}
