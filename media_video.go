package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// GetVideoEncoderConfiguration retrieves video encoder configuration.
func (c *Client) GetVideoEncoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*VideoEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoEncoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetVideoEncoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetVideoEncoderConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetVideoEncoderConfigurationResponse"`
		Configuration struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Resolution *struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"Resolution"`
			Quality     float64 `xml:"Quality"`
			RateControl *struct {
				FrameRateLimit   int `xml:"FrameRateLimit"`
				EncodingInterval int `xml:"EncodingInterval"`
				BitrateLimit     int `xml:"BitrateLimit"`
			} `xml:"RateControl"`
			SessionTimeout string `xml:"SessionTimeout"`
		} `xml:"Configuration"`
	}

	req := GetVideoEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetVideoEncoderConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoEncoderConfiguration failed: %w", err)
	}

	config := &VideoEncoderConfiguration{
		Token:          resp.Configuration.Token,
		Name:           resp.Configuration.Name,
		UseCount:       resp.Configuration.UseCount,
		Encoding:       resp.Configuration.Encoding,
		Quality:        resp.Configuration.Quality,
		SessionTimeout: parseXSDurationOrZero(resp.Configuration.SessionTimeout),
	}

	if resp.Configuration.Resolution != nil {
		config.Resolution = &VideoResolution{
			Width:  resp.Configuration.Resolution.Width,
			Height: resp.Configuration.Resolution.Height,
		}
	}

	if resp.Configuration.RateControl != nil {
		config.RateControl = &VideoRateControl{
			FrameRateLimit:   resp.Configuration.RateControl.FrameRateLimit,
			EncodingInterval: resp.Configuration.RateControl.EncodingInterval,
			BitrateLimit:     resp.Configuration.RateControl.BitrateLimit,
		}
	}

	return config, nil
}

// GetVideoSources retrieves all video sources.
func (c *Client) GetVideoSources(ctx context.Context) ([]*VideoSource, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoSources struct {
		XMLName xml.Name `xml:"trt:GetVideoSources"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetVideoSourcesResponse struct {
		XMLName      xml.Name `xml:"GetVideoSourcesResponse"`
		VideoSources []struct {
			Token      string  `xml:"token,attr"`
			Framerate  float64 `xml:"Framerate"`
			Resolution struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"Resolution"`
		} `xml:"VideoSources"`
	}

	req := GetVideoSources{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoSourcesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSources failed: %w", err)
	}

	sources := make([]*VideoSource, len(resp.VideoSources))
	for i, s := range resp.VideoSources {
		sources[i] = &VideoSource{
			Token:     s.Token,
			Framerate: s.Framerate,
			Resolution: &VideoResolution{
				Width:  s.Resolution.Width,
				Height: s.Resolution.Height,
			},
		}
	}

	return sources, nil
}

// SetVideoEncoderConfiguration sets video encoder configuration.
func (c *Client) SetVideoEncoderConfiguration(
	ctx context.Context,
	config *VideoEncoderConfiguration,
	forcePersistence bool,
) error {
	endpoint := c.getMediaEndpoint()

	type SetVideoEncoderConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetVideoEncoderConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"tt:Name"`
			UseCount   int    `xml:"tt:UseCount"`
			Encoding   string `xml:"tt:Encoding"`
			Resolution *struct {
				Width  int `xml:"tt:Width"`
				Height int `xml:"tt:Height"`
			} `xml:"tt:Resolution,omitempty"`
			Quality     *float64 `xml:"tt:Quality,omitempty"`
			RateControl *struct {
				FrameRateLimit   int `xml:"tt:FrameRateLimit"`
				EncodingInterval int `xml:"tt:EncodingInterval"`
				BitrateLimit     int `xml:"tt:BitrateLimit"`
			} `xml:"tt:RateControl,omitempty"`
			SessionTimeout string `xml:"tt:SessionTimeout,omitempty"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetVideoEncoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Encoding = config.Encoding

	if config.Resolution != nil {
		req.Configuration.Resolution = &struct {
			Width  int `xml:"tt:Width"`
			Height int `xml:"tt:Height"`
		}{
			Width:  config.Resolution.Width,
			Height: config.Resolution.Height,
		}
	}

	if config.Quality > 0 {
		req.Configuration.Quality = &config.Quality
	}

	if config.SessionTimeout > 0 {
		req.Configuration.SessionTimeout = formatDuration(config.SessionTimeout)
	}

	if config.RateControl != nil {
		req.Configuration.RateControl = &struct {
			FrameRateLimit   int `xml:"tt:FrameRateLimit"`
			EncodingInterval int `xml:"tt:EncodingInterval"`
			BitrateLimit     int `xml:"tt:BitrateLimit"`
		}{
			FrameRateLimit:   config.RateControl.FrameRateLimit,
			EncodingInterval: config.RateControl.EncodingInterval,
			BitrateLimit:     config.RateControl.BitrateLimit,
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoEncoderConfiguration failed: %w", err)
	}

	return nil
}

// GetVideoEncoderConfigurationOptions retrieves available options for video encoder configuration.
//
//nolint:funlen // GetVideoEncoderConfigurationOptions has many statements due to parsing complex encoder options
func (c *Client) GetVideoEncoderConfigurationOptions(
	ctx context.Context, configurationToken string,
) (*VideoEncoderConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoEncoderConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetVideoEncoderConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetVideoEncoderConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetVideoEncoderConfigurationOptionsResponse"`
		Options struct {
			QualityRange *struct {
				Min float64 `xml:"Min"`
				Max float64 `xml:"Max"`
			} `xml:"QualityRange"`
			JPEG *struct {
				ResolutionsAvailable []struct {
					Width  int `xml:"Width"`
					Height int `xml:"Height"`
				} `xml:"ResolutionsAvailable"`
				FrameRateRange *struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"FrameRateRange"`
				EncodingIntervalRange *struct {
					Min int `xml:"Min"`
					Max int `xml:"Max"`
				} `xml:"EncodingIntervalRange"`
			} `xml:"JPEG"`
			H264 *struct {
				ResolutionsAvailable []struct {
					Width  int `xml:"Width"`
					Height int `xml:"Height"`
				} `xml:"ResolutionsAvailable"`
				GovLengthRange *struct {
					Min int `xml:"Min"`
					Max int `xml:"Max"`
				} `xml:"GovLengthRange"`
				FrameRateRange *struct {
					Min float64 `xml:"Min"`
					Max float64 `xml:"Max"`
				} `xml:"FrameRateRange"`
				EncodingIntervalRange *struct {
					Min int `xml:"Min"`
					Max int `xml:"Max"`
				} `xml:"EncodingIntervalRange"`
				H264ProfilesSupported []string `xml:"H264ProfilesSupported"`
			} `xml:"H264"`
			Extension struct{} `xml:"Extension"`
		} `xml:"Options"`
	}

	req := GetVideoEncoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetVideoEncoderConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoEncoderConfigurationOptions failed: %w", err)
	}

	options := &VideoEncoderConfigurationOptions{}

	if resp.Options.QualityRange != nil {
		options.QualityRange = &FloatRange{
			Min: resp.Options.QualityRange.Min,
			Max: resp.Options.QualityRange.Max,
		}
	}

	if resp.Options.JPEG != nil {
		jpegOpts := &JPEGOptions{}
		if resp.Options.JPEG.FrameRateRange != nil {
			jpegOpts.FrameRateRange = &FloatRange{
				Min: resp.Options.JPEG.FrameRateRange.Min,
				Max: resp.Options.JPEG.FrameRateRange.Max,
			}
		}
		if resp.Options.JPEG.EncodingIntervalRange != nil {
			jpegOpts.EncodingIntervalRange = &IntRange{
				Min: resp.Options.JPEG.EncodingIntervalRange.Min,
				Max: resp.Options.JPEG.EncodingIntervalRange.Max,
			}
		}
		for _, res := range resp.Options.JPEG.ResolutionsAvailable {
			jpegOpts.ResolutionsAvailable = append(jpegOpts.ResolutionsAvailable, &VideoResolution{
				Width:  res.Width,
				Height: res.Height,
			})
		}
		options.JPEG = jpegOpts
	}

	if resp.Options.H264 != nil {
		h264Opts := &H264Options{}
		if resp.Options.H264.FrameRateRange != nil {
			h264Opts.FrameRateRange = &FloatRange{
				Min: resp.Options.H264.FrameRateRange.Min,
				Max: resp.Options.H264.FrameRateRange.Max,
			}
		}
		if resp.Options.H264.GovLengthRange != nil {
			h264Opts.GovLengthRange = &IntRange{
				Min: resp.Options.H264.GovLengthRange.Min,
				Max: resp.Options.H264.GovLengthRange.Max,
			}
		}
		if resp.Options.H264.EncodingIntervalRange != nil {
			h264Opts.EncodingIntervalRange = &IntRange{
				Min: resp.Options.H264.EncodingIntervalRange.Min,
				Max: resp.Options.H264.EncodingIntervalRange.Max,
			}
		}
		for _, res := range resp.Options.H264.ResolutionsAvailable {
			h264Opts.ResolutionsAvailable = append(h264Opts.ResolutionsAvailable, &VideoResolution{
				Width:  res.Width,
				Height: res.Height,
			})
		}
		h264Opts.H264ProfilesSupported = resp.Options.H264.H264ProfilesSupported
		options.H264 = h264Opts
	}

	return options, nil
}

// GetVideoSourceModes retrieves available video source modes.
func (c *Client) GetVideoSourceModes(ctx context.Context, videoSourceToken string) ([]*VideoSourceMode, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoSourceModes struct {
		XMLName          xml.Name `xml:"trt:GetVideoSourceModes"`
		Xmlns            string   `xml:"xmlns:trt,attr"`
		VideoSourceToken string   `xml:"trt:VideoSourceToken"`
	}

	type GetVideoSourceModesResponse struct {
		XMLName          xml.Name `xml:"GetVideoSourceModesResponse"`
		VideoSourceModes []struct {
			Token      string `xml:"token,attr"`
			Enabled    bool   `xml:"Enabled"`
			Resolution struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"Resolution"`
		} `xml:"VideoSourceModes"`
	}

	req := GetVideoSourceModes{
		Xmlns:            mediaNamespace,
		VideoSourceToken: videoSourceToken,
	}

	var resp GetVideoSourceModesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceModes failed: %w", err)
	}

	modes := make([]*VideoSourceMode, len(resp.VideoSourceModes))
	for i, m := range resp.VideoSourceModes {
		modes[i] = &VideoSourceMode{
			Token:   m.Token,
			Enabled: m.Enabled,
			Resolution: &VideoResolution{
				Width:  m.Resolution.Width,
				Height: m.Resolution.Height,
			},
		}
	}

	return modes, nil
}

// SetVideoSourceMode sets the video source mode.
func (c *Client) SetVideoSourceMode(ctx context.Context, videoSourceToken, modeToken string) error {
	endpoint := c.getMediaEndpoint()

	type SetVideoSourceMode struct {
		XMLName          xml.Name `xml:"trt:SetVideoSourceMode"`
		Xmlns            string   `xml:"xmlns:trt,attr"`
		VideoSourceToken string   `xml:"trt:VideoSourceToken"`
		ModeToken        string   `xml:"trt:ModeToken"`
	}

	req := SetVideoSourceMode{
		Xmlns:            mediaNamespace,
		VideoSourceToken: videoSourceToken,
		ModeToken:        modeToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoSourceMode failed: %w", err)
	}

	return nil
}

// AddVideoEncoderConfiguration adds video encoder configuration to a profile.
func (c *Client) AddVideoEncoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddVideoEncoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddVideoEncoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddVideoEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddVideoEncoderConfiguration failed: %w", err)
	}

	return nil
}

// RemoveVideoEncoderConfiguration removes video encoder configuration from a profile.
func (c *Client) RemoveVideoEncoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveVideoEncoderConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveVideoEncoderConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveVideoEncoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveVideoEncoderConfiguration failed: %w", err)
	}

	return nil
}

// AddVideoSourceConfiguration adds video source configuration to a profile.
func (c *Client) AddVideoSourceConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddVideoSourceConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddVideoSourceConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddVideoSourceConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddVideoSourceConfiguration failed: %w", err)
	}

	return nil
}

// RemoveVideoSourceConfiguration removes video source configuration from a profile.
func (c *Client) RemoveVideoSourceConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveVideoSourceConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveVideoSourceConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveVideoSourceConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveVideoSourceConfiguration failed: %w", err)
	}

	return nil
}

// GetVideoSourceConfigurations retrieves all video source configurations.
func (c *Client) GetVideoSourceConfigurations(ctx context.Context) ([]*VideoSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoSourceConfigurations struct {
		XMLName xml.Name `xml:"trt:GetVideoSourceConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetVideoSourceConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetVideoSourceConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
			Bounds      *struct {
				X      int `xml:"x,attr"`
				Y      int `xml:"y,attr"`
				Width  int `xml:"width,attr"`
				Height int `xml:"height,attr"`
			} `xml:"Bounds"`
		} `xml:"Configurations"`
	}

	req := GetVideoSourceConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoSourceConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceConfigurations failed: %w", err)
	}

	configs := make([]*VideoSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
		if cfg.Bounds != nil {
			config.Bounds = &IntRectangle{
				X:      cfg.Bounds.X,
				Y:      cfg.Bounds.Y,
				Width:  cfg.Bounds.Width,
				Height: cfg.Bounds.Height,
			}
		}
		configs[i] = config
	}

	return configs, nil
}

// GetVideoEncoderConfigurations retrieves all video encoder configurations.
func (c *Client) GetVideoEncoderConfigurations(ctx context.Context) ([]*VideoEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoEncoderConfigurations struct {
		XMLName xml.Name `xml:"trt:GetVideoEncoderConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetVideoEncoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetVideoEncoderConfigurationsResponse"`
		Configurations []struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Resolution *struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"Resolution"`
			Quality     float64 `xml:"Quality"`
			RateControl *struct {
				FrameRateLimit   int `xml:"FrameRateLimit"`
				EncodingInterval int `xml:"EncodingInterval"`
				BitrateLimit     int `xml:"BitrateLimit"`
			} `xml:"RateControl"`
			MPEG4 *struct {
				GovLength    int    `xml:"GovLength"`
				MPEG4Profile string `xml:"MPEG4Profile"`
			} `xml:"MPEG4"`
			H264 *struct {
				GovLength   int    `xml:"GovLength"`
				H264Profile string `xml:"H264Profile"`
			} `xml:"H264"`
			Multicast *struct {
				Address *struct {
					Type        string `xml:"Type"`
					IPv4Address string `xml:"IPv4Address"`
					IPv6Address string `xml:"IPv6Address"`
				} `xml:"Address"`
				Port      int  `xml:"Port"`
				TTL       int  `xml:"TTL"`
				AutoStart bool `xml:"AutoStart"`
			} `xml:"Multicast"`
			SessionTimeout string `xml:"SessionTimeout"`
		} `xml:"Configurations"`
	}

	req := GetVideoEncoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoEncoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoEncoderConfigurations failed: %w", err)
	}

	configs := make([]*VideoEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoEncoderConfiguration{
			Token:          cfg.Token,
			Name:           cfg.Name,
			UseCount:       cfg.UseCount,
			Encoding:       cfg.Encoding,
			Quality:        cfg.Quality,
			SessionTimeout: parseXSDurationOrZero(cfg.SessionTimeout),
		}

		if cfg.Resolution != nil {
			config.Resolution = &VideoResolution{
				Width:  cfg.Resolution.Width,
				Height: cfg.Resolution.Height,
			}
		}

		if cfg.RateControl != nil {
			config.RateControl = &VideoRateControl{
				FrameRateLimit:   cfg.RateControl.FrameRateLimit,
				EncodingInterval: cfg.RateControl.EncodingInterval,
				BitrateLimit:     cfg.RateControl.BitrateLimit,
			}
		}

		if cfg.MPEG4 != nil {
			config.MPEG4 = &MPEG4Configuration{
				GovLength:    cfg.MPEG4.GovLength,
				MPEG4Profile: cfg.MPEG4.MPEG4Profile,
			}
		}

		if cfg.H264 != nil {
			config.H264 = &H264Configuration{
				GovLength:   cfg.H264.GovLength,
				H264Profile: cfg.H264.H264Profile,
			}
		}

		if cfg.Multicast != nil {
			config.Multicast = &MulticastConfiguration{
				Port:      cfg.Multicast.Port,
				TTL:       cfg.Multicast.TTL,
				AutoStart: cfg.Multicast.AutoStart,
			}
			if cfg.Multicast.Address != nil {
				config.Multicast.Address = &IPAddress{
					Type:        cfg.Multicast.Address.Type,
					IPv4Address: cfg.Multicast.Address.IPv4Address,
					IPv6Address: cfg.Multicast.Address.IPv6Address,
				}
			}
		}

		configs[i] = config
	}

	return configs, nil
}

// GetVideoSourceConfiguration retrieves a specific video source configuration.
func (c *Client) GetVideoSourceConfiguration(
	ctx context.Context,
	configurationToken string,
) (*VideoSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoSourceConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetVideoSourceConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetVideoSourceConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetVideoSourceConfigurationResponse"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
			Bounds      *struct {
				X      int `xml:"x,attr"`
				Y      int `xml:"y,attr"`
				Width  int `xml:"width,attr"`
				Height int `xml:"height,attr"`
			} `xml:"Bounds"`
		} `xml:"Configuration"`
	}

	req := GetVideoSourceConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetVideoSourceConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceConfiguration failed: %w", err)
	}

	config := &VideoSourceConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		SourceToken: resp.Configuration.SourceToken,
	}

	if resp.Configuration.Bounds != nil {
		config.Bounds = &IntRectangle{
			X:      resp.Configuration.Bounds.X,
			Y:      resp.Configuration.Bounds.Y,
			Width:  resp.Configuration.Bounds.Width,
			Height: resp.Configuration.Bounds.Height,
		}
	}

	return config, nil
}

// GetVideoSourceConfigurationOptions retrieves available options for video source configuration.
func (c *Client) GetVideoSourceConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*VideoSourceConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoSourceConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetVideoSourceConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetVideoSourceConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetVideoSourceConfigurationOptionsResponse"`
		Options struct {
			BoundsRange *struct {
				X      *IntRange `xml:"X"`
				Y      *IntRange `xml:"Y"`
				Width  *IntRange `xml:"Width"`
				Height *IntRange `xml:"Height"`
			} `xml:"BoundsRange"`
			VideoSourceTokensAvailable []string `xml:"VideoSourceTokensAvailable"`
		} `xml:"Options"`
	}

	req := GetVideoSourceConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetVideoSourceConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceConfigurationOptions failed: %w", err)
	}

	options := &VideoSourceConfigurationOptions{}
	if resp.Options.BoundsRange != nil {
		options.BoundsRange = &BoundsRange{
			X:      resp.Options.BoundsRange.X,
			Y:      resp.Options.BoundsRange.Y,
			Width:  resp.Options.BoundsRange.Width,
			Height: resp.Options.BoundsRange.Height,
		}
	}
	options.VideoSourceTokensAvailable = resp.Options.VideoSourceTokensAvailable

	return options, nil
}

// SetVideoSourceConfiguration sets video source configuration.
func (c *Client) SetVideoSourceConfiguration(
	ctx context.Context,
	config *VideoSourceConfiguration,
	forcePersistence bool,
) error {
	endpoint := c.getMediaEndpoint()

	type SetVideoSourceConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetVideoSourceConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"tt:Name"`
			UseCount    int    `xml:"tt:UseCount"`
			SourceToken string `xml:"tt:SourceToken"`
			Bounds      *struct {
				X      int `xml:"x,attr"`
				Y      int `xml:"y,attr"`
				Width  int `xml:"width,attr"`
				Height int `xml:"height,attr"`
			} `xml:"tt:Bounds,omitempty"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetVideoSourceConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.SourceToken = config.SourceToken

	if config.Bounds != nil {
		req.Configuration.Bounds = &struct {
			X      int `xml:"x,attr"`
			Y      int `xml:"y,attr"`
			Width  int `xml:"width,attr"`
			Height int `xml:"height,attr"`
		}{
			X:      config.Bounds.X,
			Y:      config.Bounds.Y,
			Width:  config.Bounds.Width,
			Height: config.Bounds.Height,
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoSourceConfiguration failed: %w", err)
	}

	return nil
}

// GetCompatibleVideoEncoderConfigurations retrieves compatible video encoder configurations for a profile.
func (c *Client) GetCompatibleVideoEncoderConfigurations(
	ctx context.Context,
	profileToken string,
) ([]*VideoEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleVideoEncoderConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleVideoEncoderConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleVideoEncoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleVideoEncoderConfigurationsResponse"`
		Configurations []struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Resolution *struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"Resolution"`
			Quality     float64 `xml:"Quality"`
			RateControl *struct {
				FrameRateLimit   int `xml:"FrameRateLimit"`
				EncodingInterval int `xml:"EncodingInterval"`
				BitrateLimit     int `xml:"BitrateLimit"`
			} `xml:"RateControl"`
			SessionTimeout string `xml:"SessionTimeout"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleVideoEncoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleVideoEncoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleVideoEncoderConfigurations failed: %w", err)
	}

	configs := make([]*VideoEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoEncoderConfiguration{
			Token:          cfg.Token,
			Name:           cfg.Name,
			UseCount:       cfg.UseCount,
			Encoding:       cfg.Encoding,
			Quality:        cfg.Quality,
			SessionTimeout: parseXSDurationOrZero(cfg.SessionTimeout),
		}

		if cfg.Resolution != nil {
			config.Resolution = &VideoResolution{
				Width:  cfg.Resolution.Width,
				Height: cfg.Resolution.Height,
			}
		}

		if cfg.RateControl != nil {
			config.RateControl = &VideoRateControl{
				FrameRateLimit:   cfg.RateControl.FrameRateLimit,
				EncodingInterval: cfg.RateControl.EncodingInterval,
				BitrateLimit:     cfg.RateControl.BitrateLimit,
			}
		}

		configs[i] = config
	}

	return configs, nil
}

// GetCompatibleVideoSourceConfigurations retrieves compatible video source configurations for a profile.
func (c *Client) GetCompatibleVideoSourceConfigurations(
	ctx context.Context,
	profileToken string,
) ([]*VideoSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleVideoSourceConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleVideoSourceConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleVideoSourceConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleVideoSourceConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
			Bounds      *struct {
				X      int `xml:"x,attr"`
				Y      int `xml:"y,attr"`
				Width  int `xml:"width,attr"`
				Height int `xml:"height,attr"`
			} `xml:"Bounds"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleVideoSourceConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleVideoSourceConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleVideoSourceConfigurations failed: %w", err)
	}

	configs := make([]*VideoSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
		if cfg.Bounds != nil {
			config.Bounds = &IntRectangle{
				X:      cfg.Bounds.X,
				Y:      cfg.Bounds.Y,
				Width:  cfg.Bounds.Width,
				Height: cfg.Bounds.Height,
			}
		}
		configs[i] = config
	}

	return configs, nil
}

// GetVideoAnalyticsConfigurations retrieves all video analytics configurations.
func (c *Client) GetVideoAnalyticsConfigurations(ctx context.Context) ([]*VideoAnalyticsConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoAnalyticsConfigurations struct {
		XMLName xml.Name `xml:"trt:GetVideoAnalyticsConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetVideoAnalyticsConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetVideoAnalyticsConfigurationsResponse"`
		Configurations []struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configurations"`
	}

	req := GetVideoAnalyticsConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoAnalyticsConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfigurations failed: %w", err)
	}

	configs := make([]*VideoAnalyticsConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &VideoAnalyticsConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

// GetVideoAnalyticsConfiguration retrieves a specific video analytics configuration.
func (c *Client) GetVideoAnalyticsConfiguration(
	ctx context.Context,
	configurationToken string,
) (*VideoAnalyticsConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoAnalyticsConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetVideoAnalyticsConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetVideoAnalyticsConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetVideoAnalyticsConfigurationResponse"`
		Configuration struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configuration"`
	}

	req := GetVideoAnalyticsConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetVideoAnalyticsConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfiguration failed: %w", err)
	}

	return &VideoAnalyticsConfiguration{
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
	}, nil
}

// GetCompatibleVideoAnalyticsConfigurations retrieves compatible video analytics configurations for a profile.
func (c *Client) GetCompatibleVideoAnalyticsConfigurations(ctx context.Context, profileToken string) ([]*VideoAnalyticsConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleVideoAnalyticsConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleVideoAnalyticsConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleVideoAnalyticsConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleVideoAnalyticsConfigurationsResponse"`
		Configurations []struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleVideoAnalyticsConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleVideoAnalyticsConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleVideoAnalyticsConfigurations failed: %w", err)
	}

	configs := make([]*VideoAnalyticsConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &VideoAnalyticsConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

// SetVideoAnalyticsConfiguration sets video analytics configuration.
func (c *Client) SetVideoAnalyticsConfiguration(ctx context.Context, config *VideoAnalyticsConfiguration, forcePersistence bool) error {
	endpoint := c.getMediaEndpoint()

	type SetVideoAnalyticsConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetVideoAnalyticsConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"tt:Name"`
			UseCount int    `xml:"tt:UseCount"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetVideoAnalyticsConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}

// GetVideoAnalyticsConfigurationOptions retrieves available options for video analytics configuration.
func (c *Client) GetVideoAnalyticsConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*VideoAnalyticsConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetVideoAnalyticsConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetVideoAnalyticsConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetVideoAnalyticsConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetVideoAnalyticsConfigurationOptionsResponse"`
		Options struct{} `xml:"Options"`
	}

	req := GetVideoAnalyticsConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetVideoAnalyticsConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfigurationOptions failed: %w", err)
	}

	return &VideoAnalyticsConfigurationOptions{}, nil
}

// AddVideoAnalyticsConfiguration adds a video analytics configuration to a profile.
func (c *Client) AddVideoAnalyticsConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddVideoAnalyticsConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddVideoAnalyticsConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddVideoAnalyticsConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}

// RemoveVideoAnalyticsConfiguration removes a video analytics configuration from a profile.
func (c *Client) RemoveVideoAnalyticsConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveVideoAnalyticsConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveVideoAnalyticsConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveVideoAnalyticsConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}
