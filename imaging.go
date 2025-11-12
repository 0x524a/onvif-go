package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524A/onvif-go/internal/soap"
)

// Imaging service namespace
const imagingNamespace = "http://www.onvif.org/ver20/imaging/wsdl"

// GetImagingSettings retrieves imaging settings for a video source
func (c *Client) GetImagingSettings(ctx context.Context, videoSourceToken string) (*ImagingSettings, error) {
	endpoint := c.imagingEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

	type GetImagingSettings struct {
		XMLName          xml.Name `xml:"timg:GetImagingSettings"`
		Xmlns            string   `xml:"xmlns:timg,attr"`
		VideoSourceToken string   `xml:"timg:VideoSourceToken"`
	}

	type GetImagingSettingsResponse struct {
		XMLName         xml.Name `xml:"GetImagingSettingsResponse"`
		ImagingSettings struct {
			BacklightCompensation *struct {
				Mode  string  `xml:"Mode"`
				Level float64 `xml:"Level"`
			} `xml:"BacklightCompensation"`
			Brightness      *float64 `xml:"Brightness"`
			ColorSaturation *float64 `xml:"ColorSaturation"`
			Contrast        *float64 `xml:"Contrast"`
			Exposure        *struct {
				Mode            string  `xml:"Mode"`
				Priority        string  `xml:"Priority"`
				MinExposureTime float64 `xml:"MinExposureTime"`
				MaxExposureTime float64 `xml:"MaxExposureTime"`
				MinGain         float64 `xml:"MinGain"`
				MaxGain         float64 `xml:"MaxGain"`
				MinIris         float64 `xml:"MinIris"`
				MaxIris         float64 `xml:"MaxIris"`
				ExposureTime    float64 `xml:"ExposureTime"`
				Gain            float64 `xml:"Gain"`
				Iris            float64 `xml:"Iris"`
			} `xml:"Exposure"`
			Focus *struct {
				AutoFocusMode string  `xml:"AutoFocusMode"`
				DefaultSpeed  float64 `xml:"DefaultSpeed"`
				NearLimit     float64 `xml:"NearLimit"`
				FarLimit      float64 `xml:"FarLimit"`
			} `xml:"Focus"`
			IrCutFilter      *string  `xml:"IrCutFilter"`
			Sharpness        *float64 `xml:"Sharpness"`
			WideDynamicRange *struct {
				Mode  string  `xml:"Mode"`
				Level float64 `xml:"Level"`
			} `xml:"WideDynamicRange"`
			WhiteBalance *struct {
				Mode   string  `xml:"Mode"`
				CrGain float64 `xml:"CrGain"`
				CbGain float64 `xml:"CbGain"`
			} `xml:"WhiteBalance"`
		} `xml:"ImagingSettings"`
	}

	req := GetImagingSettings{
		Xmlns:            imagingNamespace,
		VideoSourceToken: videoSourceToken,
	}

	var resp GetImagingSettingsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetImagingSettings failed: %w", err)
	}

	settings := &ImagingSettings{
		Brightness:      resp.ImagingSettings.Brightness,
		ColorSaturation: resp.ImagingSettings.ColorSaturation,
		Contrast:        resp.ImagingSettings.Contrast,
		IrCutFilter:     resp.ImagingSettings.IrCutFilter,
		Sharpness:       resp.ImagingSettings.Sharpness,
	}

	if resp.ImagingSettings.BacklightCompensation != nil {
		settings.BacklightCompensation = &BacklightCompensation{
			Mode:  resp.ImagingSettings.BacklightCompensation.Mode,
			Level: resp.ImagingSettings.BacklightCompensation.Level,
		}
	}

	if resp.ImagingSettings.Exposure != nil {
		settings.Exposure = &Exposure{
			Mode:            resp.ImagingSettings.Exposure.Mode,
			Priority:        resp.ImagingSettings.Exposure.Priority,
			MinExposureTime: resp.ImagingSettings.Exposure.MinExposureTime,
			MaxExposureTime: resp.ImagingSettings.Exposure.MaxExposureTime,
			MinGain:         resp.ImagingSettings.Exposure.MinGain,
			MaxGain:         resp.ImagingSettings.Exposure.MaxGain,
			MinIris:         resp.ImagingSettings.Exposure.MinIris,
			MaxIris:         resp.ImagingSettings.Exposure.MaxIris,
			ExposureTime:    resp.ImagingSettings.Exposure.ExposureTime,
			Gain:            resp.ImagingSettings.Exposure.Gain,
			Iris:            resp.ImagingSettings.Exposure.Iris,
		}
	}

	if resp.ImagingSettings.Focus != nil {
		settings.Focus = &FocusConfiguration{
			AutoFocusMode: resp.ImagingSettings.Focus.AutoFocusMode,
			DefaultSpeed:  resp.ImagingSettings.Focus.DefaultSpeed,
			NearLimit:     resp.ImagingSettings.Focus.NearLimit,
			FarLimit:      resp.ImagingSettings.Focus.FarLimit,
		}
	}

	if resp.ImagingSettings.WideDynamicRange != nil {
		settings.WideDynamicRange = &WideDynamicRange{
			Mode:  resp.ImagingSettings.WideDynamicRange.Mode,
			Level: resp.ImagingSettings.WideDynamicRange.Level,
		}
	}

	if resp.ImagingSettings.WhiteBalance != nil {
		settings.WhiteBalance = &WhiteBalance{
			Mode:   resp.ImagingSettings.WhiteBalance.Mode,
			CrGain: resp.ImagingSettings.WhiteBalance.CrGain,
			CbGain: resp.ImagingSettings.WhiteBalance.CbGain,
		}
	}

	return settings, nil
}

// SetImagingSettings sets imaging settings for a video source
func (c *Client) SetImagingSettings(ctx context.Context, videoSourceToken string, settings *ImagingSettings, forcePersistence bool) error {
	endpoint := c.imagingEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

	type SetImagingSettings struct {
		XMLName          xml.Name `xml:"timg:SetImagingSettings"`
		Xmlns            string   `xml:"xmlns:timg,attr"`
		VideoSourceToken string   `xml:"timg:VideoSourceToken"`
		ImagingSettings  struct {
			BacklightCompensation *struct {
				Mode  string  `xml:"Mode"`
				Level float64 `xml:"Level"`
			} `xml:"BacklightCompensation,omitempty"`
			Brightness      *float64 `xml:"Brightness,omitempty"`
			ColorSaturation *float64 `xml:"ColorSaturation,omitempty"`
			Contrast        *float64 `xml:"Contrast,omitempty"`
			Exposure        *struct {
				Mode            string  `xml:"Mode"`
				Priority        string  `xml:"Priority,omitempty"`
				MinExposureTime float64 `xml:"MinExposureTime,omitempty"`
				MaxExposureTime float64 `xml:"MaxExposureTime,omitempty"`
				MinGain         float64 `xml:"MinGain,omitempty"`
				MaxGain         float64 `xml:"MaxGain,omitempty"`
				MinIris         float64 `xml:"MinIris,omitempty"`
				MaxIris         float64 `xml:"MaxIris,omitempty"`
				ExposureTime    float64 `xml:"ExposureTime,omitempty"`
				Gain            float64 `xml:"Gain,omitempty"`
				Iris            float64 `xml:"Iris,omitempty"`
			} `xml:"Exposure,omitempty"`
			Focus *struct {
				AutoFocusMode string  `xml:"AutoFocusMode"`
				DefaultSpeed  float64 `xml:"DefaultSpeed,omitempty"`
				NearLimit     float64 `xml:"NearLimit,omitempty"`
				FarLimit      float64 `xml:"FarLimit,omitempty"`
			} `xml:"Focus,omitempty"`
			IrCutFilter      *string  `xml:"IrCutFilter,omitempty"`
			Sharpness        *float64 `xml:"Sharpness,omitempty"`
			WideDynamicRange *struct {
				Mode  string  `xml:"Mode"`
				Level float64 `xml:"Level,omitempty"`
			} `xml:"WideDynamicRange,omitempty"`
			WhiteBalance *struct {
				Mode   string  `xml:"Mode"`
				CrGain float64 `xml:"CrGain,omitempty"`
				CbGain float64 `xml:"CbGain,omitempty"`
			} `xml:"WhiteBalance,omitempty"`
		} `xml:"timg:ImagingSettings"`
		ForcePersistence bool `xml:"timg:ForcePersistence"`
	}

	req := SetImagingSettings{
		Xmlns:            imagingNamespace,
		VideoSourceToken: videoSourceToken,
		ForcePersistence: forcePersistence,
	}

	// Map settings
	if settings.BacklightCompensation != nil {
		req.ImagingSettings.BacklightCompensation = &struct {
			Mode  string  `xml:"Mode"`
			Level float64 `xml:"Level"`
		}{
			Mode:  settings.BacklightCompensation.Mode,
			Level: settings.BacklightCompensation.Level,
		}
	}

	req.ImagingSettings.Brightness = settings.Brightness
	req.ImagingSettings.ColorSaturation = settings.ColorSaturation
	req.ImagingSettings.Contrast = settings.Contrast
	req.ImagingSettings.IrCutFilter = settings.IrCutFilter
	req.ImagingSettings.Sharpness = settings.Sharpness

	if settings.Exposure != nil {
		req.ImagingSettings.Exposure = &struct {
			Mode            string  `xml:"Mode"`
			Priority        string  `xml:"Priority,omitempty"`
			MinExposureTime float64 `xml:"MinExposureTime,omitempty"`
			MaxExposureTime float64 `xml:"MaxExposureTime,omitempty"`
			MinGain         float64 `xml:"MinGain,omitempty"`
			MaxGain         float64 `xml:"MaxGain,omitempty"`
			MinIris         float64 `xml:"MinIris,omitempty"`
			MaxIris         float64 `xml:"MaxIris,omitempty"`
			ExposureTime    float64 `xml:"ExposureTime,omitempty"`
			Gain            float64 `xml:"Gain,omitempty"`
			Iris            float64 `xml:"Iris,omitempty"`
		}{
			Mode:            settings.Exposure.Mode,
			Priority:        settings.Exposure.Priority,
			MinExposureTime: settings.Exposure.MinExposureTime,
			MaxExposureTime: settings.Exposure.MaxExposureTime,
			MinGain:         settings.Exposure.MinGain,
			MaxGain:         settings.Exposure.MaxGain,
			MinIris:         settings.Exposure.MinIris,
			MaxIris:         settings.Exposure.MaxIris,
			ExposureTime:    settings.Exposure.ExposureTime,
			Gain:            settings.Exposure.Gain,
			Iris:            settings.Exposure.Iris,
		}
	}

	if settings.Focus != nil {
		req.ImagingSettings.Focus = &struct {
			AutoFocusMode string  `xml:"AutoFocusMode"`
			DefaultSpeed  float64 `xml:"DefaultSpeed,omitempty"`
			NearLimit     float64 `xml:"NearLimit,omitempty"`
			FarLimit      float64 `xml:"FarLimit,omitempty"`
		}{
			AutoFocusMode: settings.Focus.AutoFocusMode,
			DefaultSpeed:  settings.Focus.DefaultSpeed,
			NearLimit:     settings.Focus.NearLimit,
			FarLimit:      settings.Focus.FarLimit,
		}
	}

	if settings.WideDynamicRange != nil {
		req.ImagingSettings.WideDynamicRange = &struct {
			Mode  string  `xml:"Mode"`
			Level float64 `xml:"Level,omitempty"`
		}{
			Mode:  settings.WideDynamicRange.Mode,
			Level: settings.WideDynamicRange.Level,
		}
	}

	if settings.WhiteBalance != nil {
		req.ImagingSettings.WhiteBalance = &struct {
			Mode   string  `xml:"Mode"`
			CrGain float64 `xml:"CrGain,omitempty"`
			CbGain float64 `xml:"CbGain,omitempty"`
		}{
			Mode:   settings.WhiteBalance.Mode,
			CrGain: settings.WhiteBalance.CrGain,
			CbGain: settings.WhiteBalance.CbGain,
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetImagingSettings failed: %w", err)
	}

	return nil
}

// Move performs a focus move operation
func (c *Client) Move(ctx context.Context, videoSourceToken string, focus *FocusMove) error {
	endpoint := c.imagingEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

	type Move struct {
		XMLName          xml.Name `xml:"timg:Move"`
		Xmlns            string   `xml:"xmlns:timg,attr"`
		VideoSourceToken string   `xml:"timg:VideoSourceToken"`
		Focus            *struct {
			Absolute *struct {
				Position float64 `xml:"Position"`
				Speed    float64 `xml:"Speed,omitempty"`
			} `xml:"Absolute,omitempty"`
			Relative *struct {
				Distance float64 `xml:"Distance"`
				Speed    float64 `xml:"Speed,omitempty"`
			} `xml:"Relative,omitempty"`
			Continuous *struct {
				Speed float64 `xml:"Speed"`
			} `xml:"Continuous,omitempty"`
		} `xml:"timg:Focus"`
	}

	req := Move{
		Xmlns:            imagingNamespace,
		VideoSourceToken: videoSourceToken,
	}

	if focus != nil {
		req.Focus = &struct {
			Absolute *struct {
				Position float64 `xml:"Position"`
				Speed    float64 `xml:"Speed,omitempty"`
			} `xml:"Absolute,omitempty"`
			Relative *struct {
				Distance float64 `xml:"Distance"`
				Speed    float64 `xml:"Speed,omitempty"`
			} `xml:"Relative,omitempty"`
			Continuous *struct {
				Speed float64 `xml:"Speed"`
			} `xml:"Continuous,omitempty"`
		}{}
		// Implementation would add specific focus move types here
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("Move failed: %w", err)
	}

	return nil
}

// FocusMove represents a focus move operation (placeholder for focus move types)
type FocusMove struct {
	// Can be extended with Absolute, Relative, Continuous move types
}

// GetOptions retrieves imaging options for a video source
func (c *Client) GetOptions(ctx context.Context, videoSourceToken string) (*ImagingOptions, error) {
	endpoint := c.imagingEndpoint
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetOptions struct {
		XMLName          xml.Name `xml:"timg:GetOptions"`
		Xmlns            string   `xml:"xmlns:timg,attr"`
		VideoSourceToken string   `xml:"timg:VideoSourceToken"`
	}

	type GetOptionsResponse struct {
		XMLName        xml.Name `xml:"GetOptionsResponse"`
		ImagingOptions struct {
			BacklightCompensation *struct {
				Mode  []string `xml:"Mode"`
				Level struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"Level"`
			} `xml:"BacklightCompensation"`
			Brightness *struct {
				Min float64 `xml:"Min"`
				Max float64 `xml:"Max"`
			} `xml:"Brightness"`
			ColorSaturation *struct {
				Min float64 `xml:"Min"`
				Max float64 `xml:"Max"`
			} `xml:"ColorSaturation"`
			Contrast *struct {
				Min float64 `xml:"Min"`
				Max float64 `xml:"Max"`
			} `xml:"Contrast"`
			Exposure *struct {
				Mode            []string `xml:"Mode"`
				Priority        []string `xml:"Priority"`
				MinExposureTime struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"MinExposureTime"`
				MaxExposureTime struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"MaxExposureTime"`
			} `xml:"Exposure"`
			Focus *struct {
				AutoFocusModes []string `xml:"AutoFocusModes"`
				DefaultSpeed   struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"DefaultSpeed"`
			} `xml:"Focus"`
		} `xml:"ImagingOptions"`
	}

	req := GetOptions{
		Xmlns:            imagingNamespace,
		VideoSourceToken: videoSourceToken,
	}

	var resp GetOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOptions failed: %w", err)
	}

	options := &ImagingOptions{}

	if resp.ImagingOptions.Brightness != nil {
		options.Brightness = &FloatRange{
			Min: resp.ImagingOptions.Brightness.Min,
			Max: resp.ImagingOptions.Brightness.Max,
		}
	}

	if resp.ImagingOptions.ColorSaturation != nil {
		options.ColorSaturation = &FloatRange{
			Min: resp.ImagingOptions.ColorSaturation.Min,
			Max: resp.ImagingOptions.ColorSaturation.Max,
		}
	}

	if resp.ImagingOptions.Contrast != nil {
		options.Contrast = &FloatRange{
			Min: resp.ImagingOptions.Contrast.Min,
			Max: resp.ImagingOptions.Contrast.Max,
		}
	}

	return options, nil
}

// GetMoveOptions retrieves imaging move options for focus
func (c *Client) GetMoveOptions(ctx context.Context, videoSourceToken string) (*MoveOptions, error) {
	endpoint := c.imagingEndpoint
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetMoveOptions struct {
		XMLName          xml.Name `xml:"timg:GetMoveOptions"`
		Xmlns            string   `xml:"xmlns:timg,attr"`
		VideoSourceToken string   `xml:"timg:VideoSourceToken"`
	}

	type GetMoveOptionsResponse struct {
		XMLName     xml.Name `xml:"GetMoveOptionsResponse"`
		MoveOptions struct {
			Absolute *struct {
				Position struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"Position"`
				Speed struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"Speed"`
			} `xml:"Absolute"`
			Relative *struct {
				Distance struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"Distance"`
				Speed struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"Speed"`
			} `xml:"Relative"`
			Continuous *struct {
				Speed struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"Speed"`
			} `xml:"Continuous"`
		} `xml:"MoveOptions"`
	}

	req := GetMoveOptions{
		Xmlns:            imagingNamespace,
		VideoSourceToken: videoSourceToken,
	}

	var resp GetMoveOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMoveOptions failed: %w", err)
	}

	options := &MoveOptions{}

	if resp.MoveOptions.Absolute != nil {
		options.Absolute = &AbsoluteFocusOptions{
			Position: FloatRange{
				Min: resp.MoveOptions.Absolute.Position.Min,
				Max: resp.MoveOptions.Absolute.Position.Max,
			},
			Speed: FloatRange{
				Min: resp.MoveOptions.Absolute.Speed.Min,
				Max: resp.MoveOptions.Absolute.Speed.Max,
			},
		}
	}

	if resp.MoveOptions.Relative != nil {
		options.Relative = &RelativeFocusOptions{
			Distance: FloatRange{
				Min: resp.MoveOptions.Relative.Distance.Min,
				Max: resp.MoveOptions.Relative.Distance.Max,
			},
			Speed: FloatRange{
				Min: resp.MoveOptions.Relative.Speed.Min,
				Max: resp.MoveOptions.Relative.Speed.Max,
			},
		}
	}

	if resp.MoveOptions.Continuous != nil {
		options.Continuous = &ContinuousFocusOptions{
			Speed: FloatRange{
				Min: resp.MoveOptions.Continuous.Speed.Min,
				Max: resp.MoveOptions.Continuous.Speed.Max,
			},
		}
	}

	return options, nil
}

// StopFocus stops focus movement
func (c *Client) StopFocus(ctx context.Context, videoSourceToken string) error {
	endpoint := c.imagingEndpoint
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type Stop struct {
		XMLName          xml.Name `xml:"timg:Stop"`
		Xmlns            string   `xml:"xmlns:timg,attr"`
		VideoSourceToken string   `xml:"timg:VideoSourceToken"`
	}

	req := Stop{
		Xmlns:            imagingNamespace,
		VideoSourceToken: videoSourceToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("Stop failed: %w", err)
	}

	return nil
}

// GetImagingStatus retrieves imaging status
func (c *Client) GetImagingStatus(ctx context.Context, videoSourceToken string) (*ImagingStatus, error) {
	endpoint := c.imagingEndpoint
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetStatus struct {
		XMLName          xml.Name `xml:"timg:GetStatus"`
		Xmlns            string   `xml:"xmlns:timg,attr"`
		VideoSourceToken string   `xml:"timg:VideoSourceToken"`
	}

	type GetStatusResponse struct {
		XMLName       xml.Name `xml:"GetStatusResponse"`
		ImagingStatus struct {
			FocusStatus struct {
				Position   float64 `xml:"Position"`
				MoveStatus string  `xml:"MoveStatus"`
				Error      string  `xml:"Error"`
			} `xml:"FocusStatus"`
		} `xml:"Status"`
	}

	req := GetStatus{
		Xmlns:            imagingNamespace,
		VideoSourceToken: videoSourceToken,
	}

	var resp GetStatusResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetStatus failed: %w", err)
	}

	return &ImagingStatus{
		FocusStatus: &FocusStatus{
			Position:   resp.ImagingStatus.FocusStatus.Position,
			MoveStatus: resp.ImagingStatus.FocusStatus.MoveStatus,
			Error:      resp.ImagingStatus.FocusStatus.Error,
		},
	}, nil
}
