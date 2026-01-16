package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// Media service namespace.
const mediaNamespace = "http://www.onvif.org/ver10/media/wsdl"

// getMediaEndpoint returns the media endpoint, falling back to the default endpoint if not set.
func (c *Client) getMediaEndpoint() string {
	if c.mediaEndpoint != "" {
		return c.mediaEndpoint
	}

	return c.endpoint
}

// GetProfiles retrieves all media profiles.
//
//nolint:funlen // GetProfiles has many statements due to parsing complex profile structures
func (c *Client) GetProfiles(ctx context.Context) ([]*Profile, error) {
	endpoint := c.getMediaEndpoint()

	type GetProfiles struct {
		XMLName xml.Name `xml:"trt:GetProfiles"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetProfilesResponse struct {
		XMLName  xml.Name `xml:"GetProfilesResponse"`
		Profiles []struct {
			Token                    string `xml:"token,attr"`
			Name                     string `xml:"Name"`
			VideoSourceConfiguration *struct {
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
			} `xml:"VideoSourceConfiguration"`
			VideoEncoderConfiguration *struct {
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
			} `xml:"VideoEncoderConfiguration"`
			PTZConfiguration *struct {
				Token     string `xml:"token,attr"`
				Name      string `xml:"Name"`
				UseCount  int    `xml:"UseCount"`
				NodeToken string `xml:"NodeToken"`
			} `xml:"PTZConfiguration"`
		} `xml:"Profiles"`
	}

	req := GetProfiles{
		Xmlns: mediaNamespace,
	}

	var resp GetProfilesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetProfiles failed: %w", err)
	}

	profiles := make([]*Profile, len(resp.Profiles))
	for i, p := range resp.Profiles {
		profile := &Profile{
			Token: p.Token,
			Name:  p.Name,
		}

		if p.VideoSourceConfiguration != nil {
			profile.VideoSourceConfiguration = &VideoSourceConfiguration{
				Token:       p.VideoSourceConfiguration.Token,
				Name:        p.VideoSourceConfiguration.Name,
				UseCount:    p.VideoSourceConfiguration.UseCount,
				SourceToken: p.VideoSourceConfiguration.SourceToken,
			}
			if p.VideoSourceConfiguration.Bounds != nil {
				profile.VideoSourceConfiguration.Bounds = &IntRectangle{
					X:      p.VideoSourceConfiguration.Bounds.X,
					Y:      p.VideoSourceConfiguration.Bounds.Y,
					Width:  p.VideoSourceConfiguration.Bounds.Width,
					Height: p.VideoSourceConfiguration.Bounds.Height,
				}
			}
		}

		if p.VideoEncoderConfiguration != nil {
			profile.VideoEncoderConfiguration = &VideoEncoderConfiguration{
				Token:    p.VideoEncoderConfiguration.Token,
				Name:     p.VideoEncoderConfiguration.Name,
				UseCount: p.VideoEncoderConfiguration.UseCount,
				Encoding: p.VideoEncoderConfiguration.Encoding,
				Quality:  p.VideoEncoderConfiguration.Quality,
			}
			if p.VideoEncoderConfiguration.Resolution != nil {
				profile.VideoEncoderConfiguration.Resolution = &VideoResolution{
					Width:  p.VideoEncoderConfiguration.Resolution.Width,
					Height: p.VideoEncoderConfiguration.Resolution.Height,
				}
			}
			if p.VideoEncoderConfiguration.RateControl != nil {
				profile.VideoEncoderConfiguration.RateControl = &VideoRateControl{
					FrameRateLimit:   p.VideoEncoderConfiguration.RateControl.FrameRateLimit,
					EncodingInterval: p.VideoEncoderConfiguration.RateControl.EncodingInterval,
					BitrateLimit:     p.VideoEncoderConfiguration.RateControl.BitrateLimit,
				}
			}
		}

		if p.PTZConfiguration != nil {
			profile.PTZConfiguration = &PTZConfiguration{
				Token:     p.PTZConfiguration.Token,
				Name:      p.PTZConfiguration.Name,
				UseCount:  p.PTZConfiguration.UseCount,
				NodeToken: p.PTZConfiguration.NodeToken,
			}
		}

		profiles[i] = profile
	}

	return profiles, nil
}

// GetStreamURI retrieves the stream URI for a profile.
func (c *Client) GetStreamURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	endpoint := c.getMediaEndpoint()

	type GetStreamURI struct {
		XMLName     xml.Name `xml:"trt:GetStreamUri"`
		Xmlns       string   `xml:"xmlns:trt,attr"`
		Xmlnst      string   `xml:"xmlns:tt,attr"`
		StreamSetup struct {
			Stream    string `xml:"tt:Stream"`
			Transport struct {
				Protocol string `xml:"tt:Protocol"`
			} `xml:"tt:Transport"`
		} `xml:"trt:StreamSetup"`
		ProfileToken string `xml:"trt:ProfileToken"`
	}

	type GetStreamURIResponse struct {
		XMLName  xml.Name `xml:"GetStreamUriResponse"`
		MediaURI struct {
			URI                 string `xml:"Uri"`
			InvalidAfterConnect bool   `xml:"InvalidAfterConnect"`
			InvalidAfterReboot  bool   `xml:"InvalidAfterReboot"`
			Timeout             string `xml:"Timeout"`
		} `xml:"MediaUri"`
	}

	req := GetStreamURI{
		Xmlns:        mediaNamespace,
		Xmlnst:       "http://www.onvif.org/ver10/schema",
		ProfileToken: profileToken,
	}
	req.StreamSetup.Stream = "RTP-Unicast"
	req.StreamSetup.Transport.Protocol = "RTSP"

	var resp GetStreamURIResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetStreamURI failed: %w", err)
	}

	return &MediaURI{
		URI:                 resp.MediaURI.URI,
		InvalidAfterConnect: resp.MediaURI.InvalidAfterConnect,
		InvalidAfterReboot:  resp.MediaURI.InvalidAfterReboot,
	}, nil
}

// GetSnapshotURI retrieves the snapshot URI for a profile.
func (c *Client) GetSnapshotURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	endpoint := c.getMediaEndpoint()

	type GetSnapshotURI struct {
		XMLName      xml.Name `xml:"trt:GetSnapshotUri"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetSnapshotURIResponse struct {
		XMLName  xml.Name `xml:"GetSnapshotUriResponse"`
		MediaURI struct {
			URI                 string `xml:"Uri"`
			InvalidAfterConnect bool   `xml:"InvalidAfterConnect"`
			InvalidAfterReboot  bool   `xml:"InvalidAfterReboot"`
			Timeout             string `xml:"Timeout"`
		} `xml:"MediaUri"`
	}

	req := GetSnapshotURI{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetSnapshotURIResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSnapshotURI failed: %w", err)
	}

	return &MediaURI{
		URI:                 resp.MediaURI.URI,
		InvalidAfterConnect: resp.MediaURI.InvalidAfterConnect,
		InvalidAfterReboot:  resp.MediaURI.InvalidAfterReboot,
	}, nil
}

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
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
		Encoding: resp.Configuration.Encoding,
		Quality:  resp.Configuration.Quality,
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

// GetAudioSources retrieves all audio sources.
func (c *Client) GetAudioSources(ctx context.Context) ([]*AudioSource, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSources struct {
		XMLName xml.Name `xml:"trt:GetAudioSources"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioSourcesResponse struct {
		XMLName      xml.Name `xml:"GetAudioSourcesResponse"`
		AudioSources []struct {
			Token    string `xml:"token,attr"`
			Channels int    `xml:"Channels"`
		} `xml:"AudioSources"`
	}

	req := GetAudioSources{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioSourcesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSources failed: %w", err)
	}

	sources := make([]*AudioSource, len(resp.AudioSources))
	for i, s := range resp.AudioSources {
		sources[i] = &AudioSource{
			Token:    s.Token,
			Channels: s.Channels,
		}
	}

	return sources, nil
}

// GetAudioOutputs retrieves all audio outputs.
func (c *Client) GetAudioOutputs(ctx context.Context) ([]*AudioOutput, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputs struct {
		XMLName xml.Name `xml:"trt:GetAudioOutputs"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioOutputsResponse struct {
		XMLName      xml.Name `xml:"GetAudioOutputsResponse"`
		AudioOutputs []struct {
			Token string `xml:"token,attr"`
		} `xml:"AudioOutputs"`
	}

	req := GetAudioOutputs{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioOutputsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputs failed: %w", err)
	}

	outputs := make([]*AudioOutput, len(resp.AudioOutputs))
	for i, o := range resp.AudioOutputs {
		outputs[i] = &AudioOutput{
			Token: o.Token,
		}
	}

	return outputs, nil
}

// CreateProfile creates a new media profile.
func (c *Client) CreateProfile(ctx context.Context, name, token string) (*Profile, error) {
	endpoint := c.getMediaEndpoint()

	type CreateProfile struct {
		XMLName xml.Name `xml:"trt:CreateProfile"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
		Name    string   `xml:"trt:Name"`
		Token   *string  `xml:"trt:Token,omitempty"`
	}

	type CreateProfileResponse struct {
		XMLName xml.Name `xml:"CreateProfileResponse"`
		Profile struct {
			Token string `xml:"token,attr"`
			Name  string `xml:"Name"`
		} `xml:"Profile"`
	}

	req := CreateProfile{
		Xmlns: mediaNamespace,
		Name:  name,
	}
	if token != "" {
		req.Token = &token
	}

	var resp CreateProfileResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreateProfile failed: %w", err)
	}

	return &Profile{
		Token: resp.Profile.Token,
		Name:  resp.Profile.Name,
	}, nil
}

// DeleteProfile deletes a media profile.
func (c *Client) DeleteProfile(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type DeleteProfile struct {
		XMLName      xml.Name `xml:"trt:DeleteProfile"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := DeleteProfile{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteProfile failed: %w", err)
	}

	return nil
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
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetVideoEncoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
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

// GetMediaServiceCapabilities retrieves media service capabilities.
func (c *Client) GetMediaServiceCapabilities(ctx context.Context) (*MediaServiceCapabilities, error) {
	endpoint := c.getMediaEndpoint()

	type GetServiceCapabilities struct {
		XMLName xml.Name `xml:"trt:GetServiceCapabilities"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetServiceCapabilitiesResponse struct {
		XMLName      xml.Name `xml:"GetServiceCapabilitiesResponse"`
		Capabilities struct {
			SnapshotURI         bool `xml:"SnapshotUri,attr"`
			Rotation            bool `xml:"Rotation,attr"`
			VideoSourceMode     bool `xml:"VideoSourceMode,attr"`
			OSD                 bool `xml:"OSD,attr"`
			TemporaryOSDText    bool `xml:"TemporaryOSDText,attr"`
			EXICompression      bool `xml:"EXICompression,attr"`
			ProfileCapabilities *struct {
				MaximumNumberOfProfiles int `xml:"MaximumNumberOfProfiles,attr"`
			} `xml:"ProfileCapabilities"`
			StreamingCapabilities *struct {
				RTPMulticast bool `xml:"RTPMulticast,attr"`
				RTPTCP       bool `xml:"RTP_TCP,attr"`
				RTPRTSPTCP   bool `xml:"RTP_RTSP_TCP,attr"`
			} `xml:"StreamingCapabilities"`
		} `xml:"Capabilities"`
	}

	req := GetServiceCapabilities{
		Xmlns: mediaNamespace,
	}

	var resp GetServiceCapabilitiesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMediaServiceCapabilities failed: %w", err)
	}

	caps := &MediaServiceCapabilities{
		SnapshotURI:      resp.Capabilities.SnapshotURI,
		Rotation:         resp.Capabilities.Rotation,
		VideoSourceMode:  resp.Capabilities.VideoSourceMode,
		OSD:              resp.Capabilities.OSD,
		TemporaryOSDText: resp.Capabilities.TemporaryOSDText,
		EXICompression:   resp.Capabilities.EXICompression,
	}

	if resp.Capabilities.ProfileCapabilities != nil {
		caps.MaximumNumberOfProfiles = resp.Capabilities.ProfileCapabilities.MaximumNumberOfProfiles
	}

	if resp.Capabilities.StreamingCapabilities != nil {
		caps.RTPMulticast = resp.Capabilities.StreamingCapabilities.RTPMulticast
		caps.RTPTCP = resp.Capabilities.StreamingCapabilities.RTPTCP
		caps.RTPRTSPTCP = resp.Capabilities.StreamingCapabilities.RTPRTSPTCP
	}

	return caps, nil
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

// GetAudioEncoderConfiguration retrieves audio encoder configuration.
func (c *Client) GetAudioEncoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*AudioEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioEncoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioEncoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioEncoderConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioEncoderConfigurationResponse"`
		Configuration struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Bitrate    int    `xml:"Bitrate"`
			SampleRate int    `xml:"SampleRate"`
			Multicast  *struct {
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
		} `xml:"Configuration"`
	}

	req := GetAudioEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioEncoderConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfiguration failed: %w", err)
	}

	config := &AudioEncoderConfiguration{
		Token:      resp.Configuration.Token,
		Name:       resp.Configuration.Name,
		UseCount:   resp.Configuration.UseCount,
		Encoding:   resp.Configuration.Encoding,
		Bitrate:    resp.Configuration.Bitrate,
		SampleRate: resp.Configuration.SampleRate,
	}

	if resp.Configuration.Multicast != nil {
		config.Multicast = &MulticastConfiguration{
			Port:      resp.Configuration.Multicast.Port,
			TTL:       resp.Configuration.Multicast.TTL,
			AutoStart: resp.Configuration.Multicast.AutoStart,
		}
		if resp.Configuration.Multicast.Address != nil {
			config.Multicast.Address = &IPAddress{
				Type:        resp.Configuration.Multicast.Address.Type,
				IPv4Address: resp.Configuration.Multicast.Address.IPv4Address,
				IPv6Address: resp.Configuration.Multicast.Address.IPv6Address,
			}
		}
	}

	return config, nil
}

// SetAudioEncoderConfiguration sets audio encoder configuration.
func (c *Client) SetAudioEncoderConfiguration(
	ctx context.Context,
	config *AudioEncoderConfiguration,
	forcePersistence bool,
) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioEncoderConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioEncoderConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"tt:Name"`
			UseCount   int    `xml:"tt:UseCount"`
			Encoding   string `xml:"tt:Encoding"`
			Bitrate    int    `xml:"tt:Bitrate,omitempty"`
			SampleRate int    `xml:"tt:SampleRate,omitempty"`
			Multicast  *struct {
				Address *struct {
					Type        string `xml:"tt:Type"`
					IPv4Address string `xml:"tt:IPv4Address,omitempty"`
					IPv6Address string `xml:"tt:IPv6Address,omitempty"`
				} `xml:"tt:Address,omitempty"`
				Port      int  `xml:"tt:Port,omitempty"`
				TTL       int  `xml:"tt:TTL,omitempty"`
				AutoStart bool `xml:"tt:AutoStart,omitempty"`
			} `xml:"tt:Multicast,omitempty"`
			SessionTimeout string `xml:"tt:SessionTimeout,omitempty"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetAudioEncoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Encoding = config.Encoding
	if config.Bitrate > 0 {
		req.Configuration.Bitrate = config.Bitrate
	}
	if config.SampleRate > 0 {
		req.Configuration.SampleRate = config.SampleRate
	}

	if config.Multicast != nil {
		req.Configuration.Multicast = &struct {
			Address *struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			} `xml:"tt:Address,omitempty"`
			Port      int  `xml:"tt:Port,omitempty"`
			TTL       int  `xml:"tt:TTL,omitempty"`
			AutoStart bool `xml:"tt:AutoStart,omitempty"`
		}{
			Port:      config.Multicast.Port,
			TTL:       config.Multicast.TTL,
			AutoStart: config.Multicast.AutoStart,
		}
		if config.Multicast.Address != nil {
			req.Configuration.Multicast.Address = &struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			}{
				Type:        config.Multicast.Address.Type,
				IPv4Address: config.Multicast.Address.IPv4Address,
				IPv6Address: config.Multicast.Address.IPv6Address,
			}
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

// GetMetadataConfiguration retrieves metadata configuration.
func (c *Client) GetMetadataConfiguration(
	ctx context.Context,
	configurationToken string,
) (*MetadataConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetMetadataConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetMetadataConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetMetadataConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetMetadataConfigurationResponse"`
		Configuration struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			PTZStatus *struct {
				Status   bool `xml:"Status"`
				Position bool `xml:"Position"`
			} `xml:"PTZStatus"`
			Events    *struct{} `xml:"Events"`
			Analytics bool      `xml:"Analytics"`
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
		} `xml:"Configuration"`
	}

	req := GetMetadataConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetMetadataConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfiguration failed: %w", err)
	}

	config := &MetadataConfiguration{
		Token:     resp.Configuration.Token,
		Name:      resp.Configuration.Name,
		UseCount:  resp.Configuration.UseCount,
		Analytics: resp.Configuration.Analytics,
	}

	if resp.Configuration.PTZStatus != nil {
		config.PTZStatus = &PTZFilter{
			Status:   resp.Configuration.PTZStatus.Status,
			Position: resp.Configuration.PTZStatus.Position,
		}
	}

	if resp.Configuration.Events != nil {
		config.Events = &EventSubscription{}
	}

	if resp.Configuration.Multicast != nil {
		config.Multicast = &MulticastConfiguration{
			Port:      resp.Configuration.Multicast.Port,
			TTL:       resp.Configuration.Multicast.TTL,
			AutoStart: resp.Configuration.Multicast.AutoStart,
		}
		if resp.Configuration.Multicast.Address != nil {
			config.Multicast.Address = &IPAddress{
				Type:        resp.Configuration.Multicast.Address.Type,
				IPv4Address: resp.Configuration.Multicast.Address.IPv4Address,
				IPv6Address: resp.Configuration.Multicast.Address.IPv6Address,
			}
		}
	}

	return config, nil
}

// SetMetadataConfiguration sets metadata configuration.
func (c *Client) SetMetadataConfiguration(
	ctx context.Context,
	config *MetadataConfiguration,
	forcePersistence bool,
) error {
	endpoint := c.getMediaEndpoint()

	type SetMetadataConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetMetadataConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"tt:Name"`
			UseCount  int    `xml:"tt:UseCount"`
			PTZStatus *struct {
				Status   bool `xml:"tt:Status"`
				Position bool `xml:"tt:Position"`
			} `xml:"tt:PTZStatus,omitempty"`
			Events    *struct{} `xml:"tt:Events,omitempty"`
			Analytics bool      `xml:"tt:Analytics,omitempty"`
			Multicast *struct {
				Address *struct {
					Type        string `xml:"tt:Type"`
					IPv4Address string `xml:"tt:IPv4Address,omitempty"`
					IPv6Address string `xml:"tt:IPv6Address,omitempty"`
				} `xml:"tt:Address,omitempty"`
				Port      int  `xml:"tt:Port,omitempty"`
				TTL       int  `xml:"tt:TTL,omitempty"`
				AutoStart bool `xml:"tt:AutoStart,omitempty"`
			} `xml:"tt:Multicast,omitempty"`
			SessionTimeout string `xml:"tt:SessionTimeout,omitempty"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetMetadataConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Analytics = config.Analytics

	if config.PTZStatus != nil {
		req.Configuration.PTZStatus = &struct {
			Status   bool `xml:"tt:Status"`
			Position bool `xml:"tt:Position"`
		}{
			Status:   config.PTZStatus.Status,
			Position: config.PTZStatus.Position,
		}
	}

	if config.Events != nil {
		req.Configuration.Events = &struct{}{}
	}

	if config.Multicast != nil {
		req.Configuration.Multicast = &struct {
			Address *struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			} `xml:"tt:Address,omitempty"`
			Port      int  `xml:"tt:Port,omitempty"`
			TTL       int  `xml:"tt:TTL,omitempty"`
			AutoStart bool `xml:"tt:AutoStart,omitempty"`
		}{
			Port:      config.Multicast.Port,
			TTL:       config.Multicast.TTL,
			AutoStart: config.Multicast.AutoStart,
		}
		if config.Multicast.Address != nil {
			req.Configuration.Multicast.Address = &struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			}{
				Type:        config.Multicast.Address.Type,
				IPv4Address: config.Multicast.Address.IPv4Address,
				IPv6Address: config.Multicast.Address.IPv6Address,
			}
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetMetadataConfiguration failed: %w", err)
	}

	return nil
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

// SetSynchronizationPoint sets a synchronization point for the stream.
func (c *Client) SetSynchronizationPoint(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type SetSynchronizationPoint struct {
		XMLName      xml.Name `xml:"trt:SetSynchronizationPoint"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := SetSynchronizationPoint{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetSynchronizationPoint failed: %w", err)
	}

	return nil
}

// GetOSDs retrieves all OSD configurations.
func (c *Client) GetOSDs(ctx context.Context, configurationToken string) ([]*OSDConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetOSDs struct {
		XMLName            xml.Name `xml:"trt:GetOSDs"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetOSDsResponse struct {
		XMLName xml.Name `xml:"GetOSDsResponse"`
		OSDs    []struct {
			Token string `xml:"token,attr"`
		} `xml:"OSDs"`
	}

	req := GetOSDs{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSDs failed: %w", err)
	}

	osds := make([]*OSDConfiguration, len(resp.OSDs))
	for i, o := range resp.OSDs {
		osds[i] = &OSDConfiguration{
			Token: o.Token,
		}
	}

	return osds, nil
}

// GetOSD retrieves a specific OSD configuration.
func (c *Client) GetOSD(ctx context.Context, osdToken string) (*OSDConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetOSD struct {
		XMLName  xml.Name `xml:"trt:GetOSD"`
		Xmlns    string   `xml:"xmlns:trt,attr"`
		OSDToken string   `xml:"trt:OSDToken"`
	}

	type GetOSDResponse struct {
		XMLName xml.Name `xml:"GetOSDResponse"`
		OSD     struct {
			Token string `xml:"token,attr"`
		} `xml:"OSD"`
	}

	req := GetOSD{
		Xmlns:    mediaNamespace,
		OSDToken: osdToken,
	}

	var resp GetOSDResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

// SetOSD sets OSD configuration.
func (c *Client) SetOSD(ctx context.Context, osd *OSDConfiguration) error {
	endpoint := c.getMediaEndpoint()

	type SetOSD struct {
		XMLName xml.Name `xml:"trt:SetOSD"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
		Xmlnst  string   `xml:"xmlns:tt,attr"`
		OSD     struct {
			Token string `xml:"token,attr"`
		} `xml:"trt:OSD"`
	}

	req := SetOSD{
		Xmlns:  mediaNamespace,
		Xmlnst: "http://www.onvif.org/ver10/schema",
	}
	req.OSD.Token = osd.Token

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetOSD failed: %w", err)
	}

	return nil
}

// CreateOSD creates a new OSD configuration.
func (c *Client) CreateOSD(
	ctx context.Context,
	videoSourceConfigurationToken string,
	osd *OSDConfiguration,
) (*OSDConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type CreateOSD struct {
		XMLName                       xml.Name `xml:"trt:CreateOSD"`
		Xmlns                         string   `xml:"xmlns:trt,attr"`
		Xmlnst                        string   `xml:"xmlns:tt,attr"`
		VideoSourceConfigurationToken string   `xml:"trt:VideoSourceConfigurationToken"`
		OSD                           struct {
			Token string `xml:"token,attr,omitempty"`
		} `xml:"trt:OSD"`
	}

	type CreateOSDResponse struct {
		XMLName xml.Name `xml:"CreateOSDResponse"`
		OSD     struct {
			Token string `xml:"token,attr"`
		} `xml:"OSD"`
	}

	req := CreateOSD{
		Xmlns:                         mediaNamespace,
		Xmlnst:                        "http://www.onvif.org/ver10/schema",
		VideoSourceConfigurationToken: videoSourceConfigurationToken,
	}
	if osd != nil && osd.Token != "" {
		req.OSD.Token = osd.Token
	}

	var resp CreateOSDResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreateOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

// DeleteOSD deletes an OSD configuration.
func (c *Client) DeleteOSD(ctx context.Context, osdToken string) error {
	endpoint := c.getMediaEndpoint()

	type DeleteOSD struct {
		XMLName  xml.Name `xml:"trt:DeleteOSD"`
		Xmlns    string   `xml:"xmlns:trt,attr"`
		OSDToken string   `xml:"trt:OSDToken"`
	}

	req := DeleteOSD{
		Xmlns:    mediaNamespace,
		OSDToken: osdToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteOSD failed: %w", err)
	}

	return nil
}

// StartMulticastStreaming starts multicast streaming.
func (c *Client) StartMulticastStreaming(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type StartMulticastStreaming struct {
		XMLName      xml.Name `xml:"trt:StartMulticastStreaming"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := StartMulticastStreaming{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("StartMulticastStreaming failed: %w", err)
	}

	return nil
}

// StopMulticastStreaming stops multicast streaming.
func (c *Client) StopMulticastStreaming(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type StopMulticastStreaming struct {
		XMLName      xml.Name `xml:"trt:StopMulticastStreaming"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := StopMulticastStreaming{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("StopMulticastStreaming failed: %w", err)
	}

	return nil
}

// GetProfile retrieves a specific media profile.
func (c *Client) GetProfile(ctx context.Context, profileToken string) (*Profile, error) {
	endpoint := c.getMediaEndpoint()

	type GetProfile struct {
		XMLName      xml.Name `xml:"trt:GetProfile"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetProfileResponse struct {
		XMLName xml.Name `xml:"GetProfileResponse"`
		Profile struct {
			Token string `xml:"token,attr"`
			Name  string `xml:"Name"`
		} `xml:"Profile"`
	}

	req := GetProfile{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetProfileResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetProfile failed: %w", err)
	}

	return &Profile{
		Token: resp.Profile.Token,
		Name:  resp.Profile.Name,
	}, nil
}

// SetProfile sets profile configuration.
func (c *Client) SetProfile(ctx context.Context, profile *Profile) error {
	endpoint := c.getMediaEndpoint()

	type SetProfile struct {
		XMLName xml.Name `xml:"trt:SetProfile"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
		Xmlnst  string   `xml:"xmlns:tt,attr"`
		Profile struct {
			Token string `xml:"token,attr"`
			Name  string `xml:"tt:Name"`
		} `xml:"trt:Profile"`
	}

	req := SetProfile{
		Xmlns:  mediaNamespace,
		Xmlnst: "http://www.onvif.org/ver10/schema",
	}
	req.Profile.Token = profile.Token
	req.Profile.Name = profile.Name

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetProfile failed: %w", err)
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

// AddAudioEncoderConfiguration adds audio encoder configuration to a profile.
func (c *Client) AddAudioEncoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioEncoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioEncoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioEncoderConfiguration removes audio encoder configuration from a profile.
func (c *Client) RemoveAudioEncoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioEncoderConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioEncoderConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioEncoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

// AddAudioSourceConfiguration adds audio source configuration to a profile.
func (c *Client) AddAudioSourceConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioSourceConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioSourceConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioSourceConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioSourceConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioSourceConfiguration removes audio source configuration from a profile.
func (c *Client) RemoveAudioSourceConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioSourceConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioSourceConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioSourceConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioSourceConfiguration failed: %w", err)
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

// AddPTZConfiguration adds PTZ configuration to a profile.
func (c *Client) AddPTZConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddPTZConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddPTZConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddPTZConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddPTZConfiguration failed: %w", err)
	}

	return nil
}

// RemovePTZConfiguration removes PTZ configuration from a profile.
func (c *Client) RemovePTZConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemovePTZConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemovePTZConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemovePTZConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemovePTZConfiguration failed: %w", err)
	}

	return nil
}

// AddMetadataConfiguration adds metadata configuration to a profile.
func (c *Client) AddMetadataConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddMetadataConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddMetadataConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddMetadataConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddMetadataConfiguration failed: %w", err)
	}

	return nil
}

// RemoveMetadataConfiguration removes metadata configuration from a profile.
func (c *Client) RemoveMetadataConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveMetadataConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveMetadataConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveMetadataConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveMetadataConfiguration failed: %w", err)
	}

	return nil
}

// GetAudioEncoderConfigurationOptions retrieves available options for audio encoder configuration.
func (c *Client) GetAudioEncoderConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*AudioEncoderConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioEncoderConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioEncoderConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetAudioEncoderConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioEncoderConfigurationOptionsResponse"`
		Options struct {
			EncodingOptions []string `xml:"EncodingOptions"`
			BitrateList     []int    `xml:"BitrateList"`
			SampleRateList  []int    `xml:"SampleRateList"`
		} `xml:"Options"`
	}

	req := GetAudioEncoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetAudioEncoderConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfigurationOptions failed: %w", err)
	}

	return &AudioEncoderConfigurationOptions{
		EncodingOptions: resp.Options.EncodingOptions,
		BitrateList:     resp.Options.BitrateList,
		SampleRateList:  resp.Options.SampleRateList,
	}, nil
}

// GetMetadataConfigurationOptions retrieves available options for metadata configuration.
func (c *Client) GetMetadataConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*MetadataConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetMetadataConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetMetadataConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetMetadataConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetMetadataConfigurationOptionsResponse"`
		Options struct {
			PTZStatusFilterOptions *struct {
				Status   bool `xml:"Status"`
				Position bool `xml:"Position"`
			} `xml:"PTZStatusFilterOptions"`
			Extension struct{} `xml:"Extension"`
		} `xml:"Options"`
	}

	req := GetMetadataConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetMetadataConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfigurationOptions failed: %w", err)
	}

	options := &MetadataConfigurationOptions{}
	if resp.Options.PTZStatusFilterOptions != nil {
		options.PTZStatusFilterOptions = &PTZFilter{
			Status:   resp.Options.PTZStatusFilterOptions.Status,
			Position: resp.Options.PTZStatusFilterOptions.Position,
		}
	}

	return options, nil
}

// GetAudioOutputConfiguration retrieves audio output configuration.
func (c *Client) GetAudioOutputConfiguration(ctx context.Context, configurationToken string) (*AudioOutputConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioOutputConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioOutputConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioOutputConfigurationResponse"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			OutputToken string `xml:"OutputToken"`
		} `xml:"Configuration"`
	}

	req := GetAudioOutputConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioOutputConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfiguration failed: %w", err)
	}

	return &AudioOutputConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		OutputToken: resp.Configuration.OutputToken,
	}, nil
}

// SetAudioOutputConfiguration sets audio output configuration.
func (c *Client) SetAudioOutputConfiguration(ctx context.Context, config *AudioOutputConfiguration, forcePersistence bool) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioOutputConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioOutputConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"tt:Name"`
			UseCount    int    `xml:"tt:UseCount"`
			OutputToken string `xml:"tt:OutputToken"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetAudioOutputConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.OutputToken = config.OutputToken

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

// GetAudioOutputConfigurationOptions retrieves available options for audio output configuration.
func (c *Client) GetAudioOutputConfigurationOptions(
	ctx context.Context,
	configurationToken string,
) (*AudioOutputConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioOutputConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetAudioOutputConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioOutputConfigurationOptionsResponse"`
		Options struct {
			OutputTokensAvailable []string `xml:"OutputTokensAvailable"`
		} `xml:"Options"`
	}

	req := GetAudioOutputConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetAudioOutputConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfigurationOptions failed: %w", err)
	}

	return &AudioOutputConfigurationOptions{
		OutputTokensAvailable: resp.Options.OutputTokensAvailable,
	}, nil
}

// GetAudioDecoderConfigurationOptions retrieves available options for audio decoder configuration.
func (c *Client) GetAudioDecoderConfigurationOptions(
	ctx context.Context,
	configurationToken string,
) (*AudioDecoderConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioDecoderConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioDecoderConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetAudioDecoderConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioDecoderConfigurationOptionsResponse"`
		Options struct {
			AACDecOptions *struct {
				BitrateList    []int `xml:"BitrateList"`
				SampleRateList []int `xml:"SampleRateList"`
			} `xml:"AACDecOptions"`
			G711DecOptions *struct {
				BitrateList []int `xml:"BitrateList"`
			} `xml:"G711DecOptions"`
			G726DecOptions *struct {
				BitrateList []int `xml:"BitrateList"`
			} `xml:"G726DecOptions"`
		} `xml:"Options"`
	}

	req := GetAudioDecoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetAudioDecoderConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfigurationOptions failed: %w", err)
	}

	options := &AudioDecoderConfigurationOptions{}
	if resp.Options.AACDecOptions != nil {
		options.AACDecOptions = &AudioDecoderOptions{
			BitrateList:    resp.Options.AACDecOptions.BitrateList,
			SampleRateList: resp.Options.AACDecOptions.SampleRateList,
		}
	}
	if resp.Options.G711DecOptions != nil {
		options.G711DecOptions = &AudioDecoderOptions{
			BitrateList: resp.Options.G711DecOptions.BitrateList,
		}
	}
	if resp.Options.G726DecOptions != nil {
		options.G726DecOptions = &AudioDecoderOptions{
			BitrateList: resp.Options.G726DecOptions.BitrateList,
		}
	}

	return options, nil
}

// GetGuaranteedNumberOfVideoEncoderInstances retrieves the guaranteed number of video encoder instances.
func (c *Client) GetGuaranteedNumberOfVideoEncoderInstances(
	ctx context.Context,
	configurationToken string,
) (*GuaranteedNumberOfVideoEncoderInstances, error) {
	endpoint := c.getMediaEndpoint()

	type GetGuaranteedNumberOfVideoEncoderInstances struct {
		XMLName            xml.Name `xml:"trt:GetGuaranteedNumberOfVideoEncoderInstances"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetGuaranteedNumberOfVideoEncoderInstancesResponse struct {
		XMLName     xml.Name `xml:"GetGuaranteedNumberOfVideoEncoderInstancesResponse"`
		TotalNumber int      `xml:"TotalNumber"`
		JPEG        int      `xml:"JPEG"`
		H264        int      `xml:"H264"`
		MPEG4       int      `xml:"MPEG4"`
	}

	req := GetGuaranteedNumberOfVideoEncoderInstances{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetGuaranteedNumberOfVideoEncoderInstancesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetGuaranteedNumberOfVideoEncoderInstances failed: %w", err)
	}

	return &GuaranteedNumberOfVideoEncoderInstances{
		TotalNumber: resp.TotalNumber,
		JPEG:        resp.JPEG,
		H264:        resp.H264,
		MPEG4:       resp.MPEG4,
	}, nil
}

// GetOSDOptions retrieves available options for OSD configuration.
func (c *Client) GetOSDOptions(ctx context.Context, configurationToken string) (*OSDConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetOSDOptions struct {
		XMLName            xml.Name `xml:"trt:GetOSDOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetOSDOptionsResponse struct {
		XMLName xml.Name `xml:"GetOSDOptionsResponse"`
		Options struct {
			MaximumNumberOfOSDs int `xml:"MaximumNumberOfOSDs"`
		} `xml:"Options"`
	}

	req := GetOSDOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSDOptions failed: %w", err)
	}

	return &OSDConfigurationOptions{
		MaximumNumberOfOSDs: resp.Options.MaximumNumberOfOSDs,
	}, nil
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

// GetAudioSourceConfigurations retrieves all audio source configurations.
func (c *Client) GetAudioSourceConfigurations(ctx context.Context) ([]*AudioSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSourceConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioSourceConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioSourceConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioSourceConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
		} `xml:"Configurations"`
	}

	req := GetAudioSourceConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioSourceConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfigurations failed: %w", err)
	}

	configs := make([]*AudioSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
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
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
			Encoding: cfg.Encoding,
			Quality:  cfg.Quality,
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

// GetAudioEncoderConfigurations retrieves all audio encoder configurations.
func (c *Client) GetAudioEncoderConfigurations(ctx context.Context) ([]*AudioEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioEncoderConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioEncoderConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioEncoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioEncoderConfigurationsResponse"`
		Configurations []struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Bitrate    int    `xml:"Bitrate"`
			SampleRate int    `xml:"SampleRate"`
			Multicast  *struct {
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

	req := GetAudioEncoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioEncoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &AudioEncoderConfiguration{
			Token:      cfg.Token,
			Name:       cfg.Name,
			UseCount:   cfg.UseCount,
			Encoding:   cfg.Encoding,
			Bitrate:    cfg.Bitrate,
			SampleRate: cfg.SampleRate,
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

// GetAudioSourceConfiguration retrieves a specific audio source configuration.
func (c *Client) GetAudioSourceConfiguration(ctx context.Context, configurationToken string) (*AudioSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSourceConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioSourceConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioSourceConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioSourceConfigurationResponse"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
		} `xml:"Configuration"`
	}

	req := GetAudioSourceConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioSourceConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfiguration failed: %w", err)
	}

	return &AudioSourceConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		SourceToken: resp.Configuration.SourceToken,
	}, nil
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

// GetAudioSourceConfigurationOptions retrieves available options for audio source configuration.
func (c *Client) GetAudioSourceConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*AudioSourceConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSourceConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioSourceConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetAudioSourceConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioSourceConfigurationOptionsResponse"`
		Options struct {
			InputTokensAvailable []string `xml:"InputTokensAvailable"`
		} `xml:"Options"`
	}

	req := GetAudioSourceConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetAudioSourceConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfigurationOptions failed: %w", err)
	}

	return &AudioSourceConfigurationOptions{
		InputTokensAvailable: resp.Options.InputTokensAvailable,
	}, nil
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
		Xmlnst:           "http://www.onvif.org/ver10/schema",
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

// SetAudioSourceConfiguration sets audio source configuration.
func (c *Client) SetAudioSourceConfiguration(ctx context.Context, config *AudioSourceConfiguration, forcePersistence bool) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioSourceConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioSourceConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"tt:Name"`
			UseCount    int    `xml:"tt:UseCount"`
			SourceToken string `xml:"tt:SourceToken"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetAudioSourceConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.SourceToken = config.SourceToken

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioSourceConfiguration failed: %w", err)
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
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
			Encoding: cfg.Encoding,
			Quality:  cfg.Quality,
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

// GetCompatibleAudioEncoderConfigurations retrieves compatible audio encoder configurations for a profile.
func (c *Client) GetCompatibleAudioEncoderConfigurations(
	ctx context.Context,
	profileToken string,
) ([]*AudioEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioEncoderConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioEncoderConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioEncoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioEncoderConfigurationsResponse"`
		Configurations []struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Bitrate    int    `xml:"Bitrate"`
			SampleRate int    `xml:"SampleRate"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioEncoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioEncoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioEncoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioEncoderConfiguration{
			Token:      cfg.Token,
			Name:       cfg.Name,
			UseCount:   cfg.UseCount,
			Encoding:   cfg.Encoding,
			Bitrate:    cfg.Bitrate,
			SampleRate: cfg.SampleRate,
		}
	}

	return configs, nil
}

// GetCompatibleAudioSourceConfigurations retrieves compatible audio source configurations for a profile.
func (c *Client) GetCompatibleAudioSourceConfigurations(ctx context.Context, profileToken string) ([]*AudioSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioSourceConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioSourceConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioSourceConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioSourceConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioSourceConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioSourceConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioSourceConfigurations failed: %w", err)
	}

	configs := make([]*AudioSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
	}

	return configs, nil
}

// GetCompatiblePTZConfigurations retrieves compatible PTZ configurations for a profile.
func (c *Client) GetCompatiblePTZConfigurations(ctx context.Context, profileToken string) ([]*PTZConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatiblePTZConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatiblePTZConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatiblePTZConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatiblePTZConfigurationsResponse"`
		Configurations []struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			NodeToken string `xml:"NodeToken"`
		} `xml:"Configurations"`
	}

	req := GetCompatiblePTZConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatiblePTZConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatiblePTZConfigurations failed: %w", err)
	}

	configs := make([]*PTZConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &PTZConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			NodeToken: cfg.NodeToken,
		}
	}

	return configs, nil
}

// GetCompatibleMetadataConfigurations retrieves compatible metadata configurations for a profile.
func (c *Client) GetCompatibleMetadataConfigurations(ctx context.Context, profileToken string) ([]*MetadataConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleMetadataConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleMetadataConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleMetadataConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleMetadataConfigurationsResponse"`
		Configurations []struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			Analytics bool   `xml:"Analytics"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleMetadataConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleMetadataConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleMetadataConfigurations failed: %w", err)
	}

	configs := make([]*MetadataConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &MetadataConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			Analytics: cfg.Analytics,
		}
	}

	return configs, nil
}

// GetCompatibleAudioOutputConfigurations retrieves compatible audio output configurations for a profile.
func (c *Client) GetCompatibleAudioOutputConfigurations(ctx context.Context, profileToken string) ([]*AudioOutputConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioOutputConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioOutputConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioOutputConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioOutputConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			OutputToken string `xml:"OutputToken"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioOutputConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioOutputConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioOutputConfigurations failed: %w", err)
	}

	configs := make([]*AudioOutputConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioOutputConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			OutputToken: cfg.OutputToken,
		}
	}

	return configs, nil
}

// GetCompatibleAudioDecoderConfigurations retrieves compatible audio decoder configurations for a profile.
func (c *Client) GetCompatibleAudioDecoderConfigurations(ctx context.Context, profileToken string) ([]*AudioDecoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioDecoderConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioDecoderConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioDecoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioDecoderConfigurationsResponse"`
		Configurations []struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioDecoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioDecoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioDecoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioDecoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioDecoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

// GetMetadataConfigurations retrieves all metadata configurations.
func (c *Client) GetMetadataConfigurations(ctx context.Context) ([]*MetadataConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetMetadataConfigurations struct {
		XMLName xml.Name `xml:"trt:GetMetadataConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetMetadataConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetMetadataConfigurationsResponse"`
		Configurations []struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			Analytics bool   `xml:"Analytics"`
		} `xml:"Configurations"`
	}

	req := GetMetadataConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetMetadataConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfigurations failed: %w", err)
	}

	configs := make([]*MetadataConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &MetadataConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			Analytics: cfg.Analytics,
		}
	}

	return configs, nil
}

// GetAudioOutputConfigurations retrieves all audio output configurations.
func (c *Client) GetAudioOutputConfigurations(ctx context.Context) ([]*AudioOutputConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioOutputConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioOutputConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioOutputConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			OutputToken string `xml:"OutputToken"`
		} `xml:"Configurations"`
	}

	req := GetAudioOutputConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioOutputConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfigurations failed: %w", err)
	}

	configs := make([]*AudioOutputConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioOutputConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			OutputToken: cfg.OutputToken,
		}
	}

	return configs, nil
}

// GetAudioDecoderConfigurations retrieves all audio decoder configurations.
func (c *Client) GetAudioDecoderConfigurations(ctx context.Context) ([]*AudioDecoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioDecoderConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioDecoderConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioDecoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioDecoderConfigurationsResponse"`
		Configurations []struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configurations"`
	}

	req := GetAudioDecoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioDecoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioDecoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioDecoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

// GetAudioDecoderConfiguration retrieves a specific audio decoder configuration.
func (c *Client) GetAudioDecoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*AudioDecoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioDecoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioDecoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioDecoderConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioDecoderConfigurationResponse"`
		Configuration struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configuration"`
	}

	req := GetAudioDecoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioDecoderConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfiguration failed: %w", err)
	}

	return &AudioDecoderConfiguration{
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
	}, nil
}

// SetAudioDecoderConfiguration sets audio decoder configuration.
func (c *Client) SetAudioDecoderConfiguration(ctx context.Context, config *AudioDecoderConfiguration, forcePersistence bool) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioDecoderConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioDecoderConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"tt:Name"`
			UseCount int    `xml:"tt:UseCount"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetAudioDecoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioDecoderConfiguration failed: %w", err)
	}

	return nil
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
		Xmlnst:           "http://www.onvif.org/ver10/schema",
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

// AddAudioOutputConfiguration adds an audio output configuration to a profile.
func (c *Client) AddAudioOutputConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioOutputConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioOutputConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioOutputConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioOutputConfiguration removes an audio output configuration from a profile.
func (c *Client) RemoveAudioOutputConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioOutputConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioOutputConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioOutputConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

// AddAudioDecoderConfiguration adds an audio decoder configuration to a profile.
func (c *Client) AddAudioDecoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioDecoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioDecoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioDecoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioDecoderConfiguration removes an audio decoder configuration from a profile.
func (c *Client) RemoveAudioDecoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioDecoderConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioDecoderConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioDecoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}
