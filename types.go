package onvif

import "time"

// DeviceInformation contains basic device information
type DeviceInformation struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareID      string
}

// Capabilities represents the device capabilities
type Capabilities struct {
	Analytics *AnalyticsCapabilities
	Device    *DeviceCapabilities
	Events    *EventCapabilities
	Imaging   *ImagingCapabilities
	Media     *MediaCapabilities
	PTZ       *PTZCapabilities
	Extension *CapabilitiesExtension
}

// AnalyticsCapabilities represents analytics service capabilities
type AnalyticsCapabilities struct {
	XAddr                string
	RuleSupport          bool
	AnalyticsModuleSupport bool
}

// DeviceCapabilities represents device service capabilities
type DeviceCapabilities struct {
	XAddr   string
	Network *NetworkCapabilities
	System  *SystemCapabilities
	IO      *IOCapabilities
	Security *SecurityCapabilities
}

// EventCapabilities represents event service capabilities
type EventCapabilities struct {
	XAddr                        string
	WSSubscriptionPolicySupport  bool
	WSPullPointSupport           bool
	WSPausableSubscriptionSupport bool
}

// ImagingCapabilities represents imaging service capabilities
type ImagingCapabilities struct {
	XAddr string
}

// MediaCapabilities represents media service capabilities
type MediaCapabilities struct {
	XAddr            string
	StreamingCapabilities *StreamingCapabilities
}

// PTZCapabilities represents PTZ service capabilities
type PTZCapabilities struct {
	XAddr string
}

// NetworkCapabilities represents network capabilities
type NetworkCapabilities struct {
	IPFilter            bool
	ZeroConfiguration   bool
	IPVersion6          bool
	DynDNS              bool
	Extension           *NetworkCapabilitiesExtension
}

// SystemCapabilities represents system capabilities
type SystemCapabilities struct {
	DiscoveryResolve    bool
	DiscoveryBye        bool
	RemoteDiscovery     bool
	SystemBackup        bool
	SystemLogging       bool
	FirmwareUpgrade     bool
	SupportedVersions   []string
	Extension           *SystemCapabilitiesExtension
}

// IOCapabilities represents I/O capabilities
type IOCapabilities struct {
	InputConnectors  int
	RelayOutputs     int
	Extension        *IOCapabilitiesExtension
}

// SecurityCapabilities represents security capabilities
type SecurityCapabilities struct {
	TLS11            bool
	TLS12            bool
	OnboardKeyGeneration bool
	AccessPolicyConfig   bool
	X509Token        bool
	SAMLToken        bool
	KerberosToken    bool
	RELToken         bool
	Extension        *SecurityCapabilitiesExtension
}

// StreamingCapabilities represents streaming capabilities
type StreamingCapabilities struct {
	RTPMulticast    bool
	RTP_TCP         bool
	RTP_RTSP_TCP    bool
	Extension       *StreamingCapabilitiesExtension
}

// Extension types
type CapabilitiesExtension struct{}
type NetworkCapabilitiesExtension struct{}
type SystemCapabilitiesExtension struct{}
type IOCapabilitiesExtension struct{}
type SecurityCapabilitiesExtension struct{}
type StreamingCapabilitiesExtension struct{}

// Profile represents a media profile
type Profile struct {
	Token             string
	Name              string
	VideoSourceConfiguration    *VideoSourceConfiguration
	AudioSourceConfiguration    *AudioSourceConfiguration
	VideoEncoderConfiguration   *VideoEncoderConfiguration
	AudioEncoderConfiguration   *AudioEncoderConfiguration
	PTZConfiguration            *PTZConfiguration
	MetadataConfiguration       *MetadataConfiguration
	Extension         *ProfileExtension
}

// VideoSourceConfiguration represents video source configuration
type VideoSourceConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	SourceToken string
	Bounds      *IntRectangle
}

// AudioSourceConfiguration represents audio source configuration
type AudioSourceConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	SourceToken string
}

// VideoEncoderConfiguration represents video encoder configuration
type VideoEncoderConfiguration struct {
	Token           string
	Name            string
	UseCount        int
	Encoding        string // JPEG, MPEG4, H264
	Resolution      *VideoResolution
	Quality         float64
	RateControl     *VideoRateControl
	MPEG4           *MPEG4Configuration
	H264            *H264Configuration
	Multicast       *MulticastConfiguration
	SessionTimeout  time.Duration
}

// AudioEncoderConfiguration represents audio encoder configuration
type AudioEncoderConfiguration struct {
	Token          string
	Name           string
	UseCount       int
	Encoding       string // G711, G726, AAC
	Bitrate        int
	SampleRate     int
	Multicast      *MulticastConfiguration
	SessionTimeout time.Duration
}

// PTZConfiguration represents PTZ configuration
type PTZConfiguration struct {
	Token                    string
	Name                     string
	UseCount                 int
	NodeToken                string
	DefaultAbsolutePantTiltPositionSpace string
	DefaultAbsoluteZoomPositionSpace     string
	DefaultRelativePanTiltTranslationSpace string
	DefaultRelativeZoomTranslationSpace    string
	DefaultContinuousPanTiltVelocitySpace  string
	DefaultContinuousZoomVelocitySpace     string
	DefaultPTZSpeed          *PTZSpeed
	DefaultPTZTimeout        time.Duration
	PanTiltLimits            *PanTiltLimits
	ZoomLimits               *ZoomLimits
}

// MetadataConfiguration represents metadata configuration
type MetadataConfiguration struct {
	Token          string
	Name           string
	UseCount       int
	PTZStatus      *PTZFilter
	Events         *EventSubscription
	Analytics      bool
	Multicast      *MulticastConfiguration
	SessionTimeout time.Duration
}

// VideoResolution represents video resolution
type VideoResolution struct {
	Width  int
	Height int
}

// VideoRateControl represents video rate control
type VideoRateControl struct {
	FrameRateLimit      int
	EncodingInterval    int
	BitrateLimit        int
}

// MPEG4Configuration represents MPEG4 configuration
type MPEG4Configuration struct {
	GovLength   int
	MPEG4Profile string
}

// H264Configuration represents H264 configuration
type H264Configuration struct {
	GovLength   int
	H264Profile string
}

// MulticastConfiguration represents multicast configuration
type MulticastConfiguration struct {
	Address   *IPAddress
	Port      int
	TTL       int
	AutoStart bool
}

// IPAddress represents an IP address
type IPAddress struct {
	Type    string // IPv4 or IPv6
	Address string
}

// IntRectangle represents a rectangle with integer coordinates
type IntRectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

// PTZSpeed represents PTZ speed
type PTZSpeed struct {
	PanTilt *Vector2D
	Zoom    *Vector1D
}

// Vector2D represents a 2D vector
type Vector2D struct {
	X     float64
	Y     float64
	Space string
}

// Vector1D represents a 1D vector
type Vector1D struct {
	X     float64
	Space string
}

// PanTiltLimits represents pan/tilt limits
type PanTiltLimits struct {
	Range *Space2DDescription
}

// ZoomLimits represents zoom limits
type ZoomLimits struct {
	Range *Space1DDescription
}

// Space2DDescription represents 2D space description
type Space2DDescription struct {
	URI    string
	XRange *FloatRange
	YRange *FloatRange
}

// Space1DDescription represents 1D space description
type Space1DDescription struct {
	URI   string
	XRange *FloatRange
}

// FloatRange represents a float range
type FloatRange struct {
	Min float64
	Max float64
}

// PTZFilter represents PTZ filter
type PTZFilter struct {
	Status   bool
	Position bool
}

// EventSubscription represents event subscription
type EventSubscription struct {
	Filter *FilterType
}

// FilterType represents filter type
type FilterType struct {
	// Simplified for now
}

// ProfileExtension represents profile extension
type ProfileExtension struct{}

// StreamSetup represents stream setup parameters
type StreamSetup struct {
	Stream    string // RTP-Unicast, RTP-Multicast
	Transport *Transport
}

// Transport represents transport parameters
type Transport struct {
	Protocol string // UDP, TCP, RTSP, HTTP
	Tunnel   *Tunnel
}

// Tunnel represents tunnel parameters
type Tunnel struct{}

// MediaURI represents a media URI
type MediaURI struct {
	URI                 string
	InvalidAfterConnect bool
	InvalidAfterReboot  bool
	Timeout             time.Duration
}

// PTZStatus represents PTZ status
type PTZStatus struct {
	Position   *PTZVector
	MoveStatus *PTZMoveStatus
	Error      string
	UTCTime    time.Time
}

// PTZVector represents PTZ position
type PTZVector struct {
	PanTilt *Vector2D
	Zoom    *Vector1D
}

// PTZMoveStatus represents PTZ movement status
type PTZMoveStatus struct {
	PanTilt string // IDLE, MOVING, UNKNOWN
	Zoom    string // IDLE, MOVING, UNKNOWN
}

// PTZPreset represents a PTZ preset
type PTZPreset struct {
	Token    string
	Name     string
	PTZPosition *PTZVector
}

// ImagingSettings represents imaging settings
type ImagingSettings struct {
	BacklightCompensation *BacklightCompensation
	Brightness            *float64
	ColorSaturation       *float64
	Contrast              *float64
	Exposure              *Exposure
	Focus                 *FocusConfiguration
	IrCutFilter           *string
	Sharpness             *float64
	WideDynamicRange      *WideDynamicRange
	WhiteBalance          *WhiteBalance
	Extension             *ImagingSettingsExtension
}

// BacklightCompensation represents backlight compensation
type BacklightCompensation struct {
	Mode  string // OFF, ON
	Level float64
}

// Exposure represents exposure settings
type Exposure struct {
	Mode        string // AUTO, MANUAL
	Priority    string // LowNoise, FrameRate
	MinExposureTime float64
	MaxExposureTime float64
	MinGain     float64
	MaxGain     float64
	MinIris     float64
	MaxIris     float64
	ExposureTime float64
	Gain        float64
	Iris        float64
}

// FocusConfiguration represents focus configuration
type FocusConfiguration struct {
	AutoFocusMode string // AUTO, MANUAL
	DefaultSpeed  float64
	NearLimit     float64
	FarLimit      float64
}

// WideDynamicRange represents WDR settings
type WideDynamicRange struct {
	Mode  string // OFF, ON
	Level float64
}

// WhiteBalance represents white balance settings
type WhiteBalance struct {
	Mode   string // AUTO, MANUAL
	CrGain float64
	CbGain float64
}

// ImagingSettingsExtension represents imaging settings extension
type ImagingSettingsExtension struct{}
