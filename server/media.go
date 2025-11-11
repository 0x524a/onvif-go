package server

import (
	"encoding/xml"
	"fmt"
)

// Media service SOAP message types

// GetProfilesResponse represents GetProfiles response
type GetProfilesResponse struct {
	XMLName  xml.Name          `xml:"http://www.onvif.org/ver10/media/wsdl GetProfilesResponse"`
	Profiles []MediaProfile `xml:"Profiles"`
}

// MediaProfile represents a media profile
type MediaProfile struct {
	Token                       string                       `xml:"token,attr"`
	Fixed                       bool                         `xml:"fixed,attr"`
	Name                        string                       `xml:"Name"`
	VideoSourceConfiguration    *VideoSourceConfiguration    `xml:"VideoSourceConfiguration"`
	AudioSourceConfiguration    *AudioSourceConfiguration    `xml:"AudioSourceConfiguration,omitempty"`
	VideoEncoderConfiguration   *VideoEncoderConfiguration   `xml:"VideoEncoderConfiguration"`
	AudioEncoderConfiguration   *AudioEncoderConfiguration   `xml:"AudioEncoderConfiguration,omitempty"`
	VideoAnalyticsConfiguration *VideoAnalyticsConfiguration `xml:"VideoAnalyticsConfiguration,omitempty"`
	PTZConfiguration            *PTZConfiguration            `xml:"PTZConfiguration,omitempty"`
	MetadataConfiguration       *MetadataConfiguration       `xml:"MetadataConfiguration,omitempty"`
}

// VideoSourceConfiguration represents video source configuration
type VideoSourceConfiguration struct {
	Token       string       `xml:"token,attr"`
	Name        string       `xml:"Name"`
	UseCount    int          `xml:"UseCount"`
	SourceToken string       `xml:"SourceToken"`
	Bounds      IntRectangle `xml:"Bounds"`
}

// AudioSourceConfiguration represents audio source configuration
type AudioSourceConfiguration struct {
	Token       string `xml:"token,attr"`
	Name        string `xml:"Name"`
	UseCount    int    `xml:"UseCount"`
	SourceToken string `xml:"SourceToken"`
}

// VideoEncoderConfiguration represents video encoder configuration
type VideoEncoderConfiguration struct {
	Token          string            `xml:"token,attr"`
	Name           string            `xml:"Name"`
	UseCount       int               `xml:"UseCount"`
	Encoding       string            `xml:"Encoding"`
	Resolution     VideoResolution   `xml:"Resolution"`
	Quality        float64           `xml:"Quality"`
	RateControl    *VideoRateControl `xml:"RateControl,omitempty"`
	H264           *H264Configuration `xml:"H264,omitempty"`
	Multicast      *MulticastConfiguration `xml:"Multicast,omitempty"`
	SessionTimeout string            `xml:"SessionTimeout"`
}

// AudioEncoderConfiguration represents audio encoder configuration
type AudioEncoderConfiguration struct {
	Token          string                  `xml:"token,attr"`
	Name           string                  `xml:"Name"`
	UseCount       int                     `xml:"UseCount"`
	Encoding       string                  `xml:"Encoding"`
	Bitrate        int                     `xml:"Bitrate"`
	SampleRate     int                     `xml:"SampleRate"`
	Multicast      *MulticastConfiguration `xml:"Multicast,omitempty"`
	SessionTimeout string                  `xml:"SessionTimeout"`
}

// VideoAnalyticsConfiguration represents video analytics configuration
type VideoAnalyticsConfiguration struct {
	Token    string `xml:"token,attr"`
	Name     string `xml:"Name"`
	UseCount int    `xml:"UseCount"`
}

// PTZConfiguration represents PTZ configuration
type PTZConfiguration struct {
	Token     string `xml:"token,attr"`
	Name      string `xml:"Name"`
	UseCount  int    `xml:"UseCount"`
	NodeToken string `xml:"NodeToken"`
}

// MetadataConfiguration represents metadata configuration
type MetadataConfiguration struct {
	Token          string `xml:"token,attr"`
	Name           string `xml:"Name"`
	UseCount       int    `xml:"UseCount"`
	SessionTimeout string `xml:"SessionTimeout"`
}

// IntRectangle represents a rectangle with integer coordinates
type IntRectangle struct {
	X      int `xml:"x,attr"`
	Y      int `xml:"y,attr"`
	Width  int `xml:"width,attr"`
	Height int `xml:"height,attr"`
}

// VideoResolution represents video resolution
type VideoResolution struct {
	Width  int `xml:"Width"`
	Height int `xml:"Height"`
}

// VideoRateControl represents video rate control
type VideoRateControl struct {
	FrameRateLimit   int `xml:"FrameRateLimit"`
	EncodingInterval int `xml:"EncodingInterval"`
	BitrateLimit     int `xml:"BitrateLimit"`
}

// H264Configuration represents H264 configuration
type H264Configuration struct {
	GovLength   int    `xml:"GovLength"`
	H264Profile string `xml:"H264Profile"`
}

// MulticastConfiguration represents multicast configuration
type MulticastConfiguration struct {
	Address   IPAddress `xml:"Address"`
	Port      int       `xml:"Port"`
	TTL       int       `xml:"TTL"`
	AutoStart bool      `xml:"AutoStart"`
}

// IPAddress represents an IP address
type IPAddress struct {
	Type    string `xml:"Type"`
	IPv4Address string `xml:"IPv4Address,omitempty"`
	IPv6Address string `xml:"IPv6Address,omitempty"`
}

// GetStreamURIResponse represents GetStreamURI response
type GetStreamURIResponse struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetStreamURIResponse"`
	MediaUri MediaUri `xml:"MediaUri"`
}

// MediaUri represents a media URI
type MediaUri struct {
	Uri                 string `xml:"Uri"`
	InvalidAfterConnect bool   `xml:"InvalidAfterConnect"`
	InvalidAfterReboot  bool   `xml:"InvalidAfterReboot"`
	Timeout             string `xml:"Timeout"`
}

// GetSnapshotURIResponse represents GetSnapshotURI response
type GetSnapshotURIResponse struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetSnapshotURIResponse"`
	MediaUri MediaUri `xml:"MediaUri"`
}

// GetVideoSourcesResponse represents GetVideoSources response
type GetVideoSourcesResponse struct {
	XMLName      xml.Name      `xml:"http://www.onvif.org/ver10/media/wsdl GetVideoSourcesResponse"`
	VideoSources []VideoSource `xml:"VideoSources"`
}

// VideoSource represents a video source
type VideoSource struct {
	Token      string          `xml:"token,attr"`
	Framerate  float64         `xml:"Framerate"`
	Resolution VideoResolution `xml:"Resolution"`
}

// Media service handlers

// HandleGetProfiles handles GetProfiles request
func (s *Server) HandleGetProfiles(body interface{}) (interface{}, error) {
	profiles := make([]MediaProfile, len(s.config.Profiles))

	for i, profileCfg := range s.config.Profiles {
		profile := MediaProfile{
			Token: profileCfg.Token,
			Fixed: true,
			Name:  profileCfg.Name,
			VideoSourceConfiguration: &VideoSourceConfiguration{
				Token:       profileCfg.VideoSource.Token,
				Name:        profileCfg.VideoSource.Name,
				UseCount:    1,
				SourceToken: profileCfg.VideoSource.Token,
				Bounds: IntRectangle{
					X:      profileCfg.VideoSource.Bounds.X,
					Y:      profileCfg.VideoSource.Bounds.Y,
					Width:  profileCfg.VideoSource.Bounds.Width,
					Height: profileCfg.VideoSource.Bounds.Height,
				},
			},
			VideoEncoderConfiguration: &VideoEncoderConfiguration{
				Token:    profileCfg.Token + "_encoder",
				Name:     profileCfg.Name + " Encoder",
				UseCount: 1,
				Encoding: profileCfg.VideoEncoder.Encoding,
				Resolution: VideoResolution{
					Width:  profileCfg.VideoEncoder.Resolution.Width,
					Height: profileCfg.VideoEncoder.Resolution.Height,
				},
				Quality: profileCfg.VideoEncoder.Quality,
				RateControl: &VideoRateControl{
					FrameRateLimit:   profileCfg.VideoEncoder.Framerate,
					EncodingInterval: 1,
					BitrateLimit:     profileCfg.VideoEncoder.Bitrate,
				},
				SessionTimeout: "PT60S",
			},
		}

		// Add H264 configuration if encoding is H264
		if profileCfg.VideoEncoder.Encoding == "H264" {
			profile.VideoEncoderConfiguration.H264 = &H264Configuration{
				GovLength:   profileCfg.VideoEncoder.GovLength,
				H264Profile: "Main",
			}
		}

		// Add audio configuration if present
		if profileCfg.AudioSource != nil {
			profile.AudioSourceConfiguration = &AudioSourceConfiguration{
				Token:       profileCfg.AudioSource.Token,
				Name:        profileCfg.AudioSource.Name,
				UseCount:    1,
				SourceToken: profileCfg.AudioSource.Token,
			}
		}

		if profileCfg.AudioEncoder != nil {
			profile.AudioEncoderConfiguration = &AudioEncoderConfiguration{
				Token:          profileCfg.Token + "_audio_encoder",
				Name:           profileCfg.Name + " Audio Encoder",
				UseCount:       1,
				Encoding:       profileCfg.AudioEncoder.Encoding,
				Bitrate:        profileCfg.AudioEncoder.Bitrate,
				SampleRate:     profileCfg.AudioEncoder.SampleRate,
				SessionTimeout: "PT60S",
			}
		}

		// Add PTZ configuration if present
		if profileCfg.PTZ != nil {
			profile.PTZConfiguration = &PTZConfiguration{
				Token:     profileCfg.PTZ.NodeToken,
				Name:      profileCfg.Name + " PTZ",
				UseCount:  1,
				NodeToken: profileCfg.PTZ.NodeToken,
			}
		}

		profiles[i] = profile
	}

	return &GetProfilesResponse{
		Profiles: profiles,
	}, nil
}

// HandleGetStreamURI handles GetStreamURI request
func (s *Server) HandleGetStreamURI(body interface{}) (interface{}, error) {
	var req struct {
		ProfileToken string `xml:"ProfileToken"`
	}

	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Find the stream configuration for this profile
	streamCfg, ok := s.streams[req.ProfileToken]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", req.ProfileToken)
	}

	// Build RTSP URI
	uri := streamCfg.StreamURI
	if uri == "" {
		// Default URI construction
		host := s.config.Host
		if host == "0.0.0.0" || host == "" {
			host = "localhost"
		}
		uri = fmt.Sprintf("rtsp://%s:8554%s", host, streamCfg.RTSPPath)
	}

	return &GetStreamURIResponse{
		MediaUri: MediaUri{
			Uri:                 uri,
			InvalidAfterConnect: false,
			InvalidAfterReboot:  true,
			Timeout:             "PT60S",
		},
	}, nil
}

// HandleGetSnapshotURI handles GetSnapshotURI request
func (s *Server) HandleGetSnapshotURI(body interface{}) (interface{}, error) {
	var req struct {
		ProfileToken string `xml:"ProfileToken"`
	}

	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Find the profile
	var profileCfg *ProfileConfig
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Token == req.ProfileToken {
			profileCfg = &s.config.Profiles[i]
			break
		}
	}

	if profileCfg == nil {
		return nil, fmt.Errorf("profile not found: %s", req.ProfileToken)
	}

	if !profileCfg.Snapshot.Enabled {
		return nil, fmt.Errorf("snapshot not supported for profile: %s", req.ProfileToken)
	}

	// Build snapshot URI
	host := s.config.Host
	if host == "0.0.0.0" || host == "" {
		host = "localhost"
	}
	uri := fmt.Sprintf("http://%s:%d%s/snapshot?profile=%s",
		host, s.config.Port, s.config.BasePath, req.ProfileToken)

	return &GetSnapshotURIResponse{
		MediaUri: MediaUri{
			Uri:                 uri,
			InvalidAfterConnect: false,
			InvalidAfterReboot:  true,
			Timeout:             "PT5S",
		},
	}, nil
}

// HandleGetVideoSources handles GetVideoSources request
func (s *Server) HandleGetVideoSources(body interface{}) (interface{}, error) {
	sources := make([]VideoSource, 0)

	// Collect unique video sources from profiles
	seenSources := make(map[string]bool)
	for _, profileCfg := range s.config.Profiles {
		if !seenSources[profileCfg.VideoSource.Token] {
			sources = append(sources, VideoSource{
				Token:     profileCfg.VideoSource.Token,
				Framerate: float64(profileCfg.VideoSource.Framerate),
				Resolution: VideoResolution{
					Width:  profileCfg.VideoSource.Resolution.Width,
					Height: profileCfg.VideoSource.Resolution.Height,
				},
			})
			seenSources[profileCfg.VideoSource.Token] = true
		}
	}

	return &GetVideoSourcesResponse{
		VideoSources: sources,
	}, nil
}

// unmarshalBody is a helper to unmarshal SOAP body content
func unmarshalBody(body interface{}, target interface{}) error {
	bodyXML, err := xml.Marshal(body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(bodyXML, target)
}
