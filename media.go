package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524A/go-onvif/soap"
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
			Token string `xml:"token,attr"`
			Name  string `xml:"Name"`
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
				Token    string `xml:"token,attr"`
				Name     string `xml:"Name"`
				UseCount int    `xml:"UseCount"`
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
		XMLName      xml.Name `xml:"trt:GetStreamUri"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		StreamSetup  struct {
			Stream    string `xml:"Stream"`
			Transport struct {
				Protocol string `xml:"Protocol"`
			} `xml:"Transport"`
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
