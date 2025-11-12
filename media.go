package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524A/onvif-go/internal/soap"
)

// Media service namespace
const mediaNamespace = "http://www.onvif.org/ver10/media/wsdl"

// GetProfiles retrieves all media profiles
func (c *Client) GetProfiles(ctx context.Context) ([]*Profile, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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

// GetStreamURI retrieves the stream URI for a profile
func (c *Client) GetStreamURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

	type GetStreamUri struct {
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

	type GetStreamUriResponse struct {
		XMLName  xml.Name `xml:"GetStreamUriResponse"`
		MediaUri struct {
			Uri                 string `xml:"Uri"`
			InvalidAfterConnect bool   `xml:"InvalidAfterConnect"`
			InvalidAfterReboot  bool   `xml:"InvalidAfterReboot"`
			Timeout             string `xml:"Timeout"`
		} `xml:"MediaUri"`
	}

	req := GetStreamUri{
		Xmlns:        mediaNamespace,
		Xmlnst:       "http://www.onvif.org/ver10/schema",
		ProfileToken: profileToken,
	}
	req.StreamSetup.Stream = "RTP-Unicast"
	req.StreamSetup.Transport.Protocol = "RTSP"

	var resp GetStreamUriResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetStreamUri failed: %w", err)
	}

	return &MediaURI{
		URI:                 resp.MediaUri.Uri,
		InvalidAfterConnect: resp.MediaUri.InvalidAfterConnect,
		InvalidAfterReboot:  resp.MediaUri.InvalidAfterReboot,
	}, nil
}

// GetSnapshotURI retrieves the snapshot URI for a profile
func (c *Client) GetSnapshotURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

	type GetSnapshotUri struct {
		XMLName      xml.Name `xml:"trt:GetSnapshotUri"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetSnapshotUriResponse struct {
		XMLName  xml.Name `xml:"GetSnapshotUriResponse"`
		MediaUri struct {
			Uri                 string `xml:"Uri"`
			InvalidAfterConnect bool   `xml:"InvalidAfterConnect"`
			InvalidAfterReboot  bool   `xml:"InvalidAfterReboot"`
			Timeout             string `xml:"Timeout"`
		} `xml:"MediaUri"`
	}

	req := GetSnapshotUri{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetSnapshotUriResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSnapshotUri failed: %w", err)
	}

	return &MediaURI{
		URI:                 resp.MediaUri.Uri,
		InvalidAfterConnect: resp.MediaUri.InvalidAfterConnect,
		InvalidAfterReboot:  resp.MediaUri.InvalidAfterReboot,
	}, nil
}

// GetVideoEncoderConfiguration retrieves video encoder configuration
func (c *Client) GetVideoEncoderConfiguration(ctx context.Context, configurationToken string) (*VideoEncoderConfiguration, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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

// GetVideoSources retrieves all video sources
func (c *Client) GetVideoSources(ctx context.Context) ([]*VideoSource, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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

// GetAudioSources retrieves all audio sources
func (c *Client) GetAudioSources(ctx context.Context) ([]*AudioSource, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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

// GetAudioOutputs retrieves all audio outputs
func (c *Client) GetAudioOutputs(ctx context.Context) ([]*AudioOutput, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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

// CreateProfile creates a new media profile
func (c *Client) CreateProfile(ctx context.Context, name, token string) (*Profile, error) {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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

// DeleteProfile deletes a media profile
func (c *Client) DeleteProfile(ctx context.Context, profileToken string) error {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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

// SetVideoEncoderConfiguration sets video encoder configuration
func (c *Client) SetVideoEncoderConfiguration(ctx context.Context, config *VideoEncoderConfiguration, forcePersistence bool) error {
	endpoint := c.mediaEndpoint
	if endpoint == "" {
		endpoint = c.endpoint
	}

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
