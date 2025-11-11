package server

import (
	"fmt"
	"time"

	"github.com/0x524A/go-onvif"
)

// Config represents the ONVIF server configuration
type Config struct {
	// Server settings
	Host     string        // Bind address (e.g., "0.0.0.0")
	Port     int           // Server port (default: 8080)
	BasePath string        // Base path for services (default: "/onvif")
	Timeout  time.Duration // Request timeout

	// Device information
	DeviceInfo DeviceInfo

	// Authentication
	Username string
	Password string

	// Camera profiles (supports multi-lens cameras)
	Profiles []ProfileConfig

	// Capabilities
	SupportPTZ     bool
	SupportImaging bool
	SupportEvents  bool
}

// DeviceInfo contains device identification information
type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareID      string
}

// ProfileConfig represents a camera profile configuration
type ProfileConfig struct {
	Token        string              // Profile token (unique identifier)
	Name         string              // Profile name
	VideoSource  VideoSourceConfig   // Video source configuration
	AudioSource  *AudioSourceConfig  // Audio source configuration (optional)
	VideoEncoder VideoEncoderConfig  // Video encoder configuration
	AudioEncoder *AudioEncoderConfig // Audio encoder configuration (optional)
	PTZ          *PTZConfig          // PTZ configuration (optional)
	Snapshot     SnapshotConfig      // Snapshot configuration
}

// VideoSourceConfig represents video source configuration
type VideoSourceConfig struct {
	Token      string // Video source token
	Name       string // Video source name
	Resolution Resolution
	Framerate  int
	Bounds     Bounds
}

// AudioSourceConfig represents audio source configuration
type AudioSourceConfig struct {
	Token      string // Audio source token
	Name       string // Audio source name
	SampleRate int    // Sample rate in Hz (e.g., 8000, 16000, 48000)
	Bitrate    int    // Bitrate in kbps
}

// VideoEncoderConfig represents video encoder configuration
type VideoEncoderConfig struct {
	Encoding   string     // JPEG, H264, H265, MPEG4
	Resolution Resolution // Video resolution
	Quality    float64    // Quality (0-100)
	Framerate  int        // Frames per second
	Bitrate    int        // Bitrate in kbps
	GovLength  int        // GOP length
}

// AudioEncoderConfig represents audio encoder configuration
type AudioEncoderConfig struct {
	Encoding   string // G711, G726, AAC
	Bitrate    int    // Bitrate in kbps
	SampleRate int    // Sample rate in Hz
}

// PTZConfig represents PTZ configuration
type PTZConfig struct {
	NodeToken         string    // PTZ node token
	PanRange          Range     // Pan range in degrees
	TiltRange         Range     // Tilt range in degrees
	ZoomRange         Range     // Zoom range
	DefaultSpeed      PTZSpeed  // Default speed
	SupportsContinuous bool     // Supports continuous move
	SupportsAbsolute   bool     // Supports absolute move
	SupportsRelative   bool     // Supports relative move
	Presets           []Preset  // Predefined presets
}

// SnapshotConfig represents snapshot configuration
type SnapshotConfig struct {
	Enabled    bool       // Whether snapshots are supported
	Resolution Resolution // Snapshot resolution
	Quality    float64    // JPEG quality (0-100)
}

// Resolution represents video resolution
type Resolution struct {
	Width  int
	Height int
}

// Bounds represents video bounds
type Bounds struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Range represents a numeric range
type Range struct {
	Min float64
	Max float64
}

// PTZSpeed represents PTZ movement speed
type PTZSpeed struct {
	Pan  float64 // Pan speed (-1.0 to 1.0)
	Tilt float64 // Tilt speed (-1.0 to 1.0)
	Zoom float64 // Zoom speed (-1.0 to 1.0)
}

// Preset represents a PTZ preset position
type Preset struct {
	Token    string      // Preset token
	Name     string      // Preset name
	Position PTZPosition // Position
}

// PTZPosition represents PTZ position
type PTZPosition struct {
	Pan  float64 // Pan position
	Tilt float64 // Tilt position
	Zoom float64 // Zoom position
}

// StreamConfig represents an RTSP stream configuration
type StreamConfig struct {
	ProfileToken string // Associated profile token
	RTSPPath     string // RTSP path (e.g., "/stream1")
	StreamURI    string // Full RTSP URI
}

// Server represents the ONVIF server
type Server struct {
	config       *Config
	streams      map[string]*StreamConfig // Profile token -> stream config
	ptzState     map[string]*PTZState     // Profile token -> PTZ state
	imagingState map[string]*ImagingState // Video source token -> imaging state
	systemTime   time.Time
}

// PTZState represents the current PTZ state
type PTZState struct {
	Position   PTZPosition
	Moving     bool
	PanMoving  bool
	TiltMoving bool
	ZoomMoving bool
	LastUpdate time.Time
}

// ImagingState represents the current imaging settings state
type ImagingState struct {
	Brightness       float64
	Contrast         float64
	Saturation       float64
	Sharpness        float64
	BacklightComp    BacklightCompensation
	Exposure         ExposureSettings
	Focus            FocusSettings
	WhiteBalance     WhiteBalanceSettings
	WideDynamicRange WDRSettings
	IrCutFilter      string // ON, OFF, AUTO
}

// BacklightCompensation represents backlight compensation settings
type BacklightCompensation struct {
	Mode  string  // OFF, ON
	Level float64 // 0-100
}

// ExposureSettings represents exposure settings
type ExposureSettings struct {
	Mode         string  // AUTO, MANUAL
	Priority     string  // LowNoise, FrameRate
	MinExposure  float64
	MaxExposure  float64
	MinGain      float64
	MaxGain      float64
	ExposureTime float64
	Gain         float64
}

// FocusSettings represents focus settings
type FocusSettings struct {
	AutoFocusMode string  // AUTO, MANUAL
	DefaultSpeed  float64
	NearLimit     float64
	FarLimit      float64
	CurrentPos    float64
}

// WhiteBalanceSettings represents white balance settings
type WhiteBalanceSettings struct {
	Mode   string  // AUTO, MANUAL
	CrGain float64
	CbGain float64
}

// WDRSettings represents wide dynamic range settings
type WDRSettings struct {
	Mode  string  // OFF, ON
	Level float64 // 0-100
}

// DefaultConfig returns a default server configuration with a multi-lens camera setup
func DefaultConfig() *Config {
	return &Config{
		Host:     "0.0.0.0",
		Port:     8080,
		BasePath: "/onvif",
		Timeout:  30 * time.Second,
		DeviceInfo: DeviceInfo{
			Manufacturer:    "go-onvif",
			Model:           "Virtual Multi-Lens Camera",
			FirmwareVersion: "1.0.0",
			SerialNumber:    "SN-12345678",
			HardwareID:      "HW-87654321",
		},
		Username:       "admin",
		Password:       "admin",
		SupportPTZ:     true,
		SupportImaging: true,
		SupportEvents:  false,
		Profiles: []ProfileConfig{
			{
				Token: "profile_0",
				Name:  "Main Camera - High Quality",
				VideoSource: VideoSourceConfig{
					Token:      "video_source_0",
					Name:       "Main Camera",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Framerate:  30,
					Bounds:     Bounds{X: 0, Y: 0, Width: 1920, Height: 1080},
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Quality:    80,
					Framerate:  30,
					Bitrate:    4096,
					GovLength:  30,
				},
				PTZ: &PTZConfig{
					NodeToken:          "ptz_node_0",
					PanRange:           Range{Min: -180, Max: 180},
					TiltRange:          Range{Min: -90, Max: 90},
					ZoomRange:          Range{Min: 0, Max: 1},
					DefaultSpeed:       PTZSpeed{Pan: 0.5, Tilt: 0.5, Zoom: 0.5},
					SupportsContinuous: true,
					SupportsAbsolute:   true,
					SupportsRelative:   true,
					Presets: []Preset{
						{Token: "preset_0", Name: "Home", Position: PTZPosition{Pan: 0, Tilt: 0, Zoom: 0}},
						{Token: "preset_1", Name: "Entrance", Position: PTZPosition{Pan: -45, Tilt: -10, Zoom: 0.5}},
					},
				},
				Snapshot: SnapshotConfig{
					Enabled:    true,
					Resolution: Resolution{Width: 1920, Height: 1080},
					Quality:    85,
				},
			},
			{
				Token: "profile_1",
				Name:  "Wide Angle Camera",
				VideoSource: VideoSourceConfig{
					Token:      "video_source_1",
					Name:       "Wide Angle Camera",
					Resolution: Resolution{Width: 1280, Height: 720},
					Framerate:  30,
					Bounds:     Bounds{X: 0, Y: 0, Width: 1280, Height: 720},
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: 1280, Height: 720},
					Quality:    75,
					Framerate:  30,
					Bitrate:    2048,
					GovLength:  30,
				},
				Snapshot: SnapshotConfig{
					Enabled:    true,
					Resolution: Resolution{Width: 1280, Height: 720},
					Quality:    80,
				},
			},
			{
				Token: "profile_2",
				Name:  "Telephoto Camera",
				VideoSource: VideoSourceConfig{
					Token:      "video_source_2",
					Name:       "Telephoto Camera",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Framerate:  25,
					Bounds:     Bounds{X: 0, Y: 0, Width: 1920, Height: 1080},
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: 1920, Height: 1080},
					Quality:    85,
					Framerate:  25,
					Bitrate:    6144,
					GovLength:  25,
				},
				PTZ: &PTZConfig{
					NodeToken:          "ptz_node_2",
					PanRange:           Range{Min: -180, Max: 180},
					TiltRange:          Range{Min: -90, Max: 90},
					ZoomRange:          Range{Min: 0, Max: 3},
					DefaultSpeed:       PTZSpeed{Pan: 0.3, Tilt: 0.3, Zoom: 0.3},
					SupportsContinuous: true,
					SupportsAbsolute:   true,
					SupportsRelative:   true,
					Presets: []Preset{
						{Token: "preset_2_0", Name: "Home", Position: PTZPosition{Pan: 0, Tilt: 0, Zoom: 0}},
						{Token: "preset_2_1", Name: "Zoom In", Position: PTZPosition{Pan: 0, Tilt: 0, Zoom: 2}},
					},
				},
				Snapshot: SnapshotConfig{
					Enabled:    true,
					Resolution: Resolution{Width: 1920, Height: 1080},
					Quality:    90,
				},
			},
		},
	}
}

// ServiceEndpoints returns the service endpoint URLs
func (c *Config) ServiceEndpoints(host string) map[string]string {
	if host == "" {
		host = c.Host
		if host == "0.0.0.0" || host == "" {
			host = "localhost"
		}
	}

	baseURL := ""
	if c.Port == 80 {
		baseURL = "http://" + host + c.BasePath
	} else {
		// Import fmt at the top to use Sprintf
		baseURL = fmt.Sprintf("http://%s:%d%s", host, c.Port, c.BasePath)
	}

	endpoints := map[string]string{
		"device":  baseURL + "/device_service",
		"media":   baseURL + "/media_service",
		"imaging": baseURL + "/imaging_service",
	}

	if c.SupportPTZ {
		endpoints["ptz"] = baseURL + "/ptz_service"
	}

	if c.SupportEvents {
		endpoints["events"] = baseURL + "/events_service"
	}

	return endpoints
}

// ToONVIFProfile converts a ProfileConfig to an ONVIF Profile
func (p *ProfileConfig) ToONVIFProfile() *onvif.Profile {
	profile := &onvif.Profile{
		Token: p.Token,
		Name:  p.Name,
		VideoSourceConfiguration: &onvif.VideoSourceConfiguration{
			Token:       p.VideoSource.Token,
			Name:        p.VideoSource.Name,
			SourceToken: p.VideoSource.Token,
			Bounds: &onvif.IntRectangle{
				X:      p.VideoSource.Bounds.X,
				Y:      p.VideoSource.Bounds.Y,
				Width:  p.VideoSource.Bounds.Width,
				Height: p.VideoSource.Bounds.Height,
			},
		},
		VideoEncoderConfiguration: &onvif.VideoEncoderConfiguration{
			Token:    p.Token + "_encoder",
			Name:     p.Name + " Encoder",
			Encoding: p.VideoEncoder.Encoding,
			Resolution: &onvif.VideoResolution{
				Width:  p.VideoEncoder.Resolution.Width,
				Height: p.VideoEncoder.Resolution.Height,
			},
			Quality: p.VideoEncoder.Quality,
			RateControl: &onvif.VideoRateControl{
				FrameRateLimit: p.VideoEncoder.Framerate,
				BitrateLimit:   p.VideoEncoder.Bitrate,
			},
		},
	}

	if p.PTZ != nil {
		profile.PTZConfiguration = &onvif.PTZConfiguration{
			Token:     p.PTZ.NodeToken,
			Name:      p.Name + " PTZ",
			NodeToken: p.PTZ.NodeToken,
		}
	}

	return profile
}
