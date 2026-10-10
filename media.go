package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// Media service namespace.
const mediaNamespace = "http://www.onvif.org/ver10/media/wsdl" // NOSONAR // XML namespace identifier fixed by the ONVIF spec; never dialed

// onvifSchemaNamespace is already defined in deviceio.go and available here

// getMediaEndpoint returns the media endpoint, falling back to the default endpoint if not set.
func (c *Client) getMediaEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

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
		Xmlnst:       onvifSchemaNamespace,
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
		Xmlnst: onvifSchemaNamespace,
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
