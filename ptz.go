package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/0x524a/onvif-go/internal/soap"
)

// PTZ service namespace.
const ptzNamespace = "http://www.onvif.org/ver20/ptz/wsdl"

// getPTZEndpoint returns the discovered PTZ service endpoint.
//
// When no endpoint is known it reports which of the two possible reasons
// applies, rather than conflating them: ErrNotInitialized if Initialize has
// never succeeded - the device may well support PTZ, the client simply has
// not looked yet, and the caller's recovery is to call Initialize and retry -
// or ErrServiceNotSupported if capabilities were fetched and PTZ was absent
// from them, which is a permanent property of the device and means the caller
// should stop offering PTZ at all.
//
// Both checks are made under one lock so the pair is always consistent.
func (c *Client) getPTZEndpoint() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.ptzEndpoint == "" {
		if !c.initialized {
			return "", ErrNotInitialized
		}

		return "", ErrServiceNotSupported
	}

	return c.ptzEndpoint, nil
}

// ptzPanTiltXML is a shared type for PTZ pan/tilt XML serialization.
type ptzPanTiltXML struct {
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// ptzZoomXML is a shared type for PTZ zoom XML serialization.
type ptzZoomXML struct {
	X     float64 `xml:"x,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// ptzVectorXML is a shared type for PTZ position/velocity XML serialization.
type ptzVectorXML struct {
	PanTilt *ptzPanTiltXML `xml:"PanTilt,omitempty"`
	Zoom    *ptzZoomXML    `xml:"Zoom,omitempty"`
}

// ptzSpeedXML is a shared type for PTZ speed XML serialization.
type ptzSpeedXML struct {
	PanTilt *ptzPanTiltXML `xml:"PanTilt,omitempty"`
	Zoom    *ptzZoomXML    `xml:"Zoom,omitempty"`
}

// convertToPTZVectorXML converts PTZVector to XML struct.
func convertToPTZVectorXML(v *PTZVector) *ptzVectorXML {
	if v == nil {
		return nil
	}
	result := &ptzVectorXML{}
	if v.PanTilt != nil {
		result.PanTilt = &ptzPanTiltXML{X: v.PanTilt.X, Y: v.PanTilt.Y, Space: v.PanTilt.Space}
	}
	if v.Zoom != nil {
		result.Zoom = &ptzZoomXML{X: v.Zoom.X, Space: v.Zoom.Space}
	}

	return result
}

// convertToPTZSpeedXML converts PTZSpeed to XML struct.
func convertToPTZSpeedXML(s *PTZSpeed) *ptzSpeedXML {
	if s == nil {
		return nil
	}
	result := &ptzSpeedXML{}
	if s.PanTilt != nil {
		result.PanTilt = &ptzPanTiltXML{X: s.PanTilt.X, Y: s.PanTilt.Y, Space: s.PanTilt.Space}
	}
	if s.Zoom != nil {
		result.Zoom = &ptzZoomXML{X: s.Zoom.X, Space: s.Zoom.Space}
	}

	return result
}

// ptzFloatRangeXML is the wire form of a float range, used for the pan/tilt
// and zoom limit descriptions.
type ptzFloatRangeXML struct {
	Min float64 `xml:"Min"`
	Max float64 `xml:"Max"`
}

// ptzConfigurationXML is the wire form of a PTZConfiguration.
//
// Shared by GetConfiguration and GetConfigurations, which return the same
// element and previously each declared their own four-field subset of it.
//
// The misspelling in DefaultAbsolutePantTiltPositionSpace ("Pant") is the
// ONVIF WSDL's own and must be reproduced exactly for the element to match.
type ptzConfigurationXML struct {
	Token                                  string       `xml:"token,attr"`
	Name                                   string       `xml:"Name"`
	UseCount                               int          `xml:"UseCount"`
	NodeToken                              string       `xml:"NodeToken"`
	DefaultAbsolutePantTiltPositionSpace   string       `xml:"DefaultAbsolutePantTiltPositionSpace"`
	DefaultAbsoluteZoomPositionSpace       string       `xml:"DefaultAbsoluteZoomPositionSpace"`
	DefaultRelativePanTiltTranslationSpace string       `xml:"DefaultRelativePanTiltTranslationSpace"`
	DefaultRelativeZoomTranslationSpace    string       `xml:"DefaultRelativeZoomTranslationSpace"`
	DefaultContinuousPanTiltVelocitySpace  string       `xml:"DefaultContinuousPanTiltVelocitySpace"`
	DefaultContinuousZoomVelocitySpace     string       `xml:"DefaultContinuousZoomVelocitySpace"`
	DefaultPTZSpeed                        *ptzSpeedXML `xml:"DefaultPTZSpeed"`
	DefaultPTZTimeout                      string       `xml:"DefaultPTZTimeout"`
	PanTiltLimits                          *struct {
		Range *struct {
			URI    string            `xml:"URI"`
			XRange *ptzFloatRangeXML `xml:"XRange"`
			YRange *ptzFloatRangeXML `xml:"YRange"`
		} `xml:"Range"`
	} `xml:"PanTiltLimits"`
	ZoomLimits *struct {
		Range *struct {
			URI    string            `xml:"URI"`
			XRange *ptzFloatRangeXML `xml:"XRange"`
		} `xml:"Range"`
	} `xml:"ZoomLimits"`
}

// toPTZConfiguration converts the wire form into the library's type.
//
// DefaultPTZTimeout is an xs:duration, which is why it needs
// parseXSDurationOrZero rather than a direct assignment - the reason it, and
// the rest of these fields, went unmapped until #87.
func (x *ptzConfigurationXML) toPTZConfiguration() *PTZConfiguration {
	config := &PTZConfiguration{
		Token:                                  x.Token,
		Name:                                   x.Name,
		UseCount:                               x.UseCount,
		NodeToken:                              x.NodeToken,
		DefaultAbsolutePantTiltPositionSpace:   x.DefaultAbsolutePantTiltPositionSpace,
		DefaultAbsoluteZoomPositionSpace:       x.DefaultAbsoluteZoomPositionSpace,
		DefaultRelativePanTiltTranslationSpace: x.DefaultRelativePanTiltTranslationSpace,
		DefaultRelativeZoomTranslationSpace:    x.DefaultRelativeZoomTranslationSpace,
		DefaultContinuousPanTiltVelocitySpace:  x.DefaultContinuousPanTiltVelocitySpace,
		DefaultContinuousZoomVelocitySpace:     x.DefaultContinuousZoomVelocitySpace,
		DefaultPTZTimeout:                      parseXSDurationOrZero(x.DefaultPTZTimeout),
	}

	if x.DefaultPTZSpeed != nil {
		config.DefaultPTZSpeed = &PTZSpeed{}
		if x.DefaultPTZSpeed.PanTilt != nil {
			config.DefaultPTZSpeed.PanTilt = &Vector2D{
				X:     x.DefaultPTZSpeed.PanTilt.X,
				Y:     x.DefaultPTZSpeed.PanTilt.Y,
				Space: x.DefaultPTZSpeed.PanTilt.Space,
			}
		}
		if x.DefaultPTZSpeed.Zoom != nil {
			config.DefaultPTZSpeed.Zoom = &Vector1D{
				X:     x.DefaultPTZSpeed.Zoom.X,
				Space: x.DefaultPTZSpeed.Zoom.Space,
			}
		}
	}

	if x.PanTiltLimits != nil {
		config.PanTiltLimits = &PanTiltLimits{}
		if r := x.PanTiltLimits.Range; r != nil {
			config.PanTiltLimits.Range = &Space2DDescription{
				URI:    r.URI,
				XRange: toFloatRange(r.XRange),
				YRange: toFloatRange(r.YRange),
			}
		}
	}

	if x.ZoomLimits != nil {
		config.ZoomLimits = &ZoomLimits{}
		if r := x.ZoomLimits.Range; r != nil {
			config.ZoomLimits.Range = &Space1DDescription{
				URI:    r.URI,
				XRange: toFloatRange(r.XRange),
			}
		}
	}

	return config
}

// toFloatRange converts a wire float range, preserving nil.
func toFloatRange(r *ptzFloatRangeXML) *FloatRange {
	if r == nil {
		return nil
	}

	return &FloatRange{Min: r.Min, Max: r.Max}
}

// ContinuousMove starts continuous PTZ movement.
func (c *Client) ContinuousMove(ctx context.Context, profileToken string, velocity *PTZSpeed, timeout *string) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type ContinuousMove struct {
		XMLName      xml.Name     `xml:"tptz:ContinuousMove"`
		Xmlns        string       `xml:"xmlns:tptz,attr"`
		ProfileToken string       `xml:"tptz:ProfileToken"`
		Velocity     *ptzSpeedXML `xml:"tptz:Velocity"`
		Timeout      *string      `xml:"tptz:Timeout,omitempty"`
	}

	req := ContinuousMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Velocity:     convertToPTZSpeedXML(velocity),
		Timeout:      timeout,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("ContinuousMove failed: %w", err)
	}

	return nil
}

// AbsoluteMove moves PTZ to an absolute position.
func (c *Client) AbsoluteMove(ctx context.Context, profileToken string, position *PTZVector, speed *PTZSpeed) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type AbsoluteMove struct {
		XMLName      xml.Name      `xml:"tptz:AbsoluteMove"`
		Xmlns        string        `xml:"xmlns:tptz,attr"`
		ProfileToken string        `xml:"tptz:ProfileToken"`
		Position     *ptzVectorXML `xml:"tptz:Position"`
		Speed        *ptzSpeedXML  `xml:"tptz:Speed,omitempty"`
	}

	req := AbsoluteMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Position:     convertToPTZVectorXML(position),
		Speed:        convertToPTZSpeedXML(speed),
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AbsoluteMove failed: %w", err)
	}

	return nil
}

// RelativeMove moves PTZ relative to current position.
func (c *Client) RelativeMove(ctx context.Context, profileToken string, translation *PTZVector, speed *PTZSpeed) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type RelativeMove struct {
		XMLName      xml.Name      `xml:"tptz:RelativeMove"`
		Xmlns        string        `xml:"xmlns:tptz,attr"`
		ProfileToken string        `xml:"tptz:ProfileToken"`
		Translation  *ptzVectorXML `xml:"tptz:Translation"`
		Speed        *ptzSpeedXML  `xml:"tptz:Speed,omitempty"`
	}

	req := RelativeMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Translation:  convertToPTZVectorXML(translation),
		Speed:        convertToPTZSpeedXML(speed),
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RelativeMove failed: %w", err)
	}

	return nil
}

// Stop stops PTZ movement.
func (c *Client) Stop(ctx context.Context, profileToken string, panTilt, zoom bool) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type Stop struct {
		XMLName      xml.Name `xml:"tptz:Stop"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		PanTilt      *bool    `xml:"tptz:PanTilt,omitempty"`
		Zoom         *bool    `xml:"tptz:Zoom,omitempty"`
	}

	req := Stop{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	if panTilt {
		req.PanTilt = &panTilt
	}
	if zoom {
		req.Zoom = &zoom
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("Stop failed: %w", err)
	}

	return nil
}

// GetStatus retrieves PTZ status.
func (c *Client) GetStatus(ctx context.Context, profileToken string) (*PTZStatus, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return nil, err
	}

	type GetStatus struct {
		XMLName      xml.Name `xml:"tptz:GetStatus"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	type GetStatusResponse struct {
		XMLName   xml.Name `xml:"GetStatusResponse"`
		PTZStatus struct {
			Position *struct {
				PanTilt *struct {
					X     float64 `xml:"x,attr"`
					Y     float64 `xml:"y,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"PanTilt"`
				Zoom *struct {
					X     float64 `xml:"x,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"Zoom"`
			} `xml:"Position"`
			MoveStatus *struct {
				PanTilt string `xml:"PanTilt"`
				Zoom    string `xml:"Zoom"`
			} `xml:"MoveStatus"`
			Error   string `xml:"Error"`
			UTCTime string `xml:"UtcTime"`
		} `xml:"PTZStatus"`
	}

	req := GetStatus{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	var resp GetStatusResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetStatus failed: %w", err)
	}

	status := &PTZStatus{
		Error: resp.PTZStatus.Error,
	}

	if resp.PTZStatus.Position != nil {
		status.Position = &PTZVector{}
		if resp.PTZStatus.Position.PanTilt != nil {
			status.Position.PanTilt = &Vector2D{
				X:     resp.PTZStatus.Position.PanTilt.X,
				Y:     resp.PTZStatus.Position.PanTilt.Y,
				Space: resp.PTZStatus.Position.PanTilt.Space,
			}
		}
		if resp.PTZStatus.Position.Zoom != nil {
			status.Position.Zoom = &Vector1D{
				X:     resp.PTZStatus.Position.Zoom.X,
				Space: resp.PTZStatus.Position.Zoom.Space,
			}
		}
	}

	if resp.PTZStatus.MoveStatus != nil {
		status.MoveStatus = &PTZMoveStatus{
			PanTilt: resp.PTZStatus.MoveStatus.PanTilt,
			Zoom:    resp.PTZStatus.MoveStatus.Zoom,
		}
	}

	// A timestamp the camera formats unexpectedly leaves UTCTime zero rather
	// than failing the call, matching how event.go treats the same field: the
	// position and move status are the point of GetStatus and are still good.
	if t, err := time.Parse(time.RFC3339, resp.PTZStatus.UTCTime); err == nil {
		status.UTCTime = t
	}

	return status, nil
}

// GetPresets retrieves PTZ presets.
func (c *Client) GetPresets(ctx context.Context, profileToken string) ([]*PTZPreset, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return nil, err
	}

	type GetPresets struct {
		XMLName      xml.Name `xml:"tptz:GetPresets"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	type GetPresetsResponse struct {
		XMLName xml.Name `xml:"GetPresetsResponse"`
		Preset  []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			PTZPosition *struct {
				PanTilt *struct {
					X     float64 `xml:"x,attr"`
					Y     float64 `xml:"y,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"PanTilt"`
				Zoom *struct {
					X     float64 `xml:"x,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"Zoom"`
			} `xml:"PTZPosition"`
		} `xml:"Preset"`
	}

	req := GetPresets{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	var resp GetPresetsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetPresets failed: %w", err)
	}

	presets := make([]*PTZPreset, len(resp.Preset))
	for i, p := range resp.Preset {
		preset := &PTZPreset{
			Token: p.Token,
			Name:  p.Name,
		}

		if p.PTZPosition != nil {
			preset.PTZPosition = &PTZVector{}
			if p.PTZPosition.PanTilt != nil {
				preset.PTZPosition.PanTilt = &Vector2D{
					X:     p.PTZPosition.PanTilt.X,
					Y:     p.PTZPosition.PanTilt.Y,
					Space: p.PTZPosition.PanTilt.Space,
				}
			}
			if p.PTZPosition.Zoom != nil {
				preset.PTZPosition.Zoom = &Vector1D{
					X:     p.PTZPosition.Zoom.X,
					Space: p.PTZPosition.Zoom.Space,
				}
			}
		}

		presets[i] = preset
	}

	return presets, nil
}

// GotoPreset moves PTZ to a preset position.
func (c *Client) GotoPreset(ctx context.Context, profileToken, presetToken string, speed *PTZSpeed) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type GotoPreset struct {
		XMLName      xml.Name     `xml:"tptz:GotoPreset"`
		Xmlns        string       `xml:"xmlns:tptz,attr"`
		ProfileToken string       `xml:"tptz:ProfileToken"`
		PresetToken  string       `xml:"tptz:PresetToken"`
		Speed        *ptzSpeedXML `xml:"tptz:Speed,omitempty"`
	}

	req := GotoPreset{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		PresetToken:  presetToken,
		Speed:        convertToPTZSpeedXML(speed),
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("GotoPreset failed: %w", err)
	}

	return nil
}

// SetPreset sets a preset position.
func (c *Client) SetPreset(ctx context.Context, profileToken, presetName, presetToken string) (string, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return "", err
	}

	type SetPreset struct {
		XMLName      xml.Name `xml:"tptz:SetPreset"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		PresetName   *string  `xml:"tptz:PresetName,omitempty"`
		PresetToken  *string  `xml:"tptz:PresetToken,omitempty"`
	}

	type SetPresetResponse struct {
		XMLName     xml.Name `xml:"SetPresetResponse"`
		PresetToken string   `xml:"PresetToken"`
	}

	req := SetPreset{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	if presetName != "" {
		req.PresetName = &presetName
	}
	if presetToken != "" {
		req.PresetToken = &presetToken
	}

	var resp SetPresetResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return "", fmt.Errorf("SetPreset failed: %w", err)
	}

	return resp.PresetToken, nil
}

// RemovePreset removes a preset.
func (c *Client) RemovePreset(ctx context.Context, profileToken, presetToken string) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type RemovePreset struct {
		XMLName      xml.Name `xml:"tptz:RemovePreset"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		PresetToken  string   `xml:"tptz:PresetToken"`
	}

	req := RemovePreset{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		PresetToken:  presetToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemovePreset failed: %w", err)
	}

	return nil
}

// GotoHomePosition moves PTZ to home position.
func (c *Client) GotoHomePosition(ctx context.Context, profileToken string, speed *PTZSpeed) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type GotoHomePosition struct {
		XMLName      xml.Name     `xml:"tptz:GotoHomePosition"`
		Xmlns        string       `xml:"xmlns:tptz,attr"`
		ProfileToken string       `xml:"tptz:ProfileToken"`
		Speed        *ptzSpeedXML `xml:"tptz:Speed,omitempty"`
	}

	req := GotoHomePosition{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Speed:        convertToPTZSpeedXML(speed),
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("GotoHomePosition failed: %w", err)
	}

	return nil
}

// SetHomePosition sets the current position as home position.
func (c *Client) SetHomePosition(ctx context.Context, profileToken string) error {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return err
	}

	type SetHomePosition struct {
		XMLName      xml.Name `xml:"tptz:SetHomePosition"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	req := SetHomePosition{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetHomePosition failed: %w", err)
	}

	return nil
}

// GetConfiguration retrieves PTZ configuration.
func (c *Client) GetConfiguration(ctx context.Context, configurationToken string) (*PTZConfiguration, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return nil, err
	}

	type GetConfiguration struct {
		XMLName               xml.Name `xml:"tptz:GetConfiguration"`
		Xmlns                 string   `xml:"xmlns:tptz,attr"`
		PTZConfigurationToken string   `xml:"tptz:PTZConfigurationToken"`
	}

	type GetConfigurationResponse struct {
		XMLName          xml.Name            `xml:"GetConfigurationResponse"`
		PTZConfiguration ptzConfigurationXML `xml:"PTZConfiguration"`
	}

	req := GetConfiguration{
		Xmlns:                 ptzNamespace,
		PTZConfigurationToken: configurationToken,
	}

	var resp GetConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetConfiguration failed: %w", err)
	}

	return resp.PTZConfiguration.toPTZConfiguration(), nil
}

// GetConfigurations retrieves all PTZ configurations.
func (c *Client) GetConfigurations(ctx context.Context) ([]*PTZConfiguration, error) {
	endpoint, err := c.getPTZEndpoint()
	if err != nil {
		return nil, err
	}

	type GetConfigurations struct {
		XMLName xml.Name `xml:"tptz:GetConfigurations"`
		Xmlns   string   `xml:"xmlns:tptz,attr"`
	}

	type GetConfigurationsResponse struct {
		XMLName          xml.Name              `xml:"GetConfigurationsResponse"`
		PTZConfiguration []ptzConfigurationXML `xml:"PTZConfiguration"`
	}

	req := GetConfigurations{
		Xmlns: ptzNamespace,
	}

	var resp GetConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetConfigurations failed: %w", err)
	}

	configs := make([]*PTZConfiguration, len(resp.PTZConfiguration))
	for i := range resp.PTZConfiguration {
		configs[i] = resp.PTZConfiguration[i].toPTZConfiguration()
	}

	return configs, nil
}
