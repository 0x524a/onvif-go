package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524A/go-onvif/soap"
)

// PTZ service namespace
const ptzNamespace = "http://www.onvif.org/ver20/ptz/wsdl"

// ContinuousMove starts continuous PTZ movement
func (c *Client) ContinuousMove(ctx context.Context, profileToken string, velocity *PTZSpeed, timeout *string) error {
	endpoint := c.ptzEndpoint
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type ContinuousMove struct {
		XMLName      xml.Name `xml:"tptz:ContinuousMove"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		Velocity     *struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		} `xml:"tptz:Velocity"`
		Timeout *string `xml:"tptz:Timeout,omitempty"`
	}

	req := ContinuousMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Timeout:      timeout,
	}

	if velocity != nil {
		req.Velocity = &struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		}{}

		if velocity.PanTilt != nil {
			req.Velocity.PanTilt = &struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     velocity.PanTilt.X,
				Y:     velocity.PanTilt.Y,
				Space: velocity.PanTilt.Space,
			}
		}

		if velocity.Zoom != nil {
			req.Velocity.Zoom = &struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     velocity.Zoom.X,
				Space: velocity.Zoom.Space,
			}
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)
	
	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("ContinuousMove failed: %w", err)
	}

	return nil
}

// AbsoluteMove moves PTZ to an absolute position
func (c *Client) AbsoluteMove(ctx context.Context, profileToken string, position *PTZVector, speed *PTZSpeed) error {
	endpoint := c.ptzEndpoint
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type AbsoluteMove struct {
		XMLName      xml.Name `xml:"tptz:AbsoluteMove"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		Position     *struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		} `xml:"tptz:Position"`
		Speed *struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		} `xml:"tptz:Speed,omitempty"`
	}

	req := AbsoluteMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	if position != nil {
		req.Position = &struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		}{}

		if position.PanTilt != nil {
			req.Position.PanTilt = &struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     position.PanTilt.X,
				Y:     position.PanTilt.Y,
				Space: position.PanTilt.Space,
			}
		}

		if position.Zoom != nil {
			req.Position.Zoom = &struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     position.Zoom.X,
				Space: position.Zoom.Space,
			}
		}
	}

	if speed != nil {
		req.Speed = &struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		}{}

		if speed.PanTilt != nil {
			req.Speed.PanTilt = &struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     speed.PanTilt.X,
				Y:     speed.PanTilt.Y,
				Space: speed.PanTilt.Space,
			}
		}

		if speed.Zoom != nil {
			req.Speed.Zoom = &struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     speed.Zoom.X,
				Space: speed.Zoom.Space,
			}
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)
	
	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AbsoluteMove failed: %w", err)
	}

	return nil
}

// RelativeMove moves PTZ relative to current position
func (c *Client) RelativeMove(ctx context.Context, profileToken string, translation *PTZVector, speed *PTZSpeed) error {
	endpoint := c.ptzEndpoint
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type RelativeMove struct {
		XMLName      xml.Name `xml:"tptz:RelativeMove"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		Translation  *struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		} `xml:"tptz:Translation"`
		Speed *struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		} `xml:"tptz:Speed,omitempty"`
	}

	req := RelativeMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	if translation != nil {
		req.Translation = &struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		}{}

		if translation.PanTilt != nil {
			req.Translation.PanTilt = &struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     translation.PanTilt.X,
				Y:     translation.PanTilt.Y,
				Space: translation.PanTilt.Space,
			}
		}

		if translation.Zoom != nil {
			req.Translation.Zoom = &struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     translation.Zoom.X,
				Space: translation.Zoom.Space,
			}
		}
	}

	if speed != nil {
		req.Speed = &struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		}{}

		if speed.PanTilt != nil {
			req.Speed.PanTilt = &struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     speed.PanTilt.X,
				Y:     speed.PanTilt.Y,
				Space: speed.PanTilt.Space,
			}
		}

		if speed.Zoom != nil {
			req.Speed.Zoom = &struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     speed.Zoom.X,
				Space: speed.Zoom.Space,
			}
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)
	
	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RelativeMove failed: %w", err)
	}

	return nil
}

// Stop stops PTZ movement
func (c *Client) Stop(ctx context.Context, profileToken string, panTilt, zoom bool) error {
	endpoint := c.ptzEndpoint
	if endpoint == "" {
		return ErrServiceNotSupported
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

// GetStatus retrieves PTZ status
func (c *Client) GetStatus(ctx context.Context, profileToken string) (*PTZStatus, error) {
	endpoint := c.ptzEndpoint
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetStatus struct {
		XMLName      xml.Name `xml:"tptz:GetStatus"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	type GetStatusResponse struct {
		XMLName    xml.Name `xml:"GetStatusResponse"`
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

	return status, nil
}

// GetPresets retrieves PTZ presets
func (c *Client) GetPresets(ctx context.Context, profileToken string) ([]*PTZPreset, error) {
	endpoint := c.ptzEndpoint
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetPresets struct {
		XMLName      xml.Name `xml:"tptz:GetPresets"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	type GetPresetsResponse struct {
		XMLName xml.Name `xml:"GetPresetsResponse"`
		Preset  []struct {
			Token string `xml:"token,attr"`
			Name  string `xml:"Name"`
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

// GotoPreset moves PTZ to a preset position
func (c *Client) GotoPreset(ctx context.Context, profileToken, presetToken string, speed *PTZSpeed) error {
	endpoint := c.ptzEndpoint
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type GotoPreset struct {
		XMLName      xml.Name `xml:"tptz:GotoPreset"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		PresetToken  string   `xml:"tptz:PresetToken"`
		Speed        *struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		} `xml:"tptz:Speed,omitempty"`
	}

	req := GotoPreset{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		PresetToken:  presetToken,
	}

	if speed != nil {
		req.Speed = &struct {
			PanTilt *struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"PanTilt,omitempty"`
			Zoom *struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			} `xml:"Zoom,omitempty"`
		}{}

		if speed.PanTilt != nil {
			req.Speed.PanTilt = &struct {
				X     float64 `xml:"x,attr"`
				Y     float64 `xml:"y,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     speed.PanTilt.X,
				Y:     speed.PanTilt.Y,
				Space: speed.PanTilt.Space,
			}
		}

		if speed.Zoom != nil {
			req.Speed.Zoom = &struct {
				X     float64 `xml:"x,attr"`
				Space string  `xml:"space,attr,omitempty"`
			}{
				X:     speed.Zoom.X,
				Space: speed.Zoom.Space,
			}
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)
	
	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("GotoPreset failed: %w", err)
	}

	return nil
}
