package onvif

import "time"

// DeviceInformation contains basic device information.
type DeviceInformation struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareID      string
}

// Capabilities represents the device capabilities.
type Capabilities struct {
	Analytics *AnalyticsCapabilities
	Device    *DeviceCapabilities
	Events    *EventCapabilities
	Imaging   *ImagingCapabilities
	Media     *MediaCapabilities
	PTZ       *PTZCapabilities
	Extension *CapabilitiesExtension
}

// AnalyticsCapabilities represents analytics service capabilities.
type AnalyticsCapabilities struct {
	XAddr                  string
	RuleSupport            bool
	AnalyticsModuleSupport bool
}

// DeviceCapabilities represents device service capabilities.
type DeviceCapabilities struct {
	XAddr    string
	Network  *NetworkCapabilities
	System   *SystemCapabilities
	IO       *IOCapabilities
	Security *SecurityCapabilities
}

// EventCapabilities represents event service capabilities.
type EventCapabilities struct {
	XAddr                         string
	WSSubscriptionPolicySupport   bool
	WSPullPointSupport            bool
	WSPausableSubscriptionSupport bool
}

// ImagingCapabilities represents imaging service capabilities.
type ImagingCapabilities struct {
	XAddr string
}

// MediaCapabilities represents media service capabilities.
type MediaCapabilities struct {
	XAddr                 string
	StreamingCapabilities *StreamingCapabilities
}

// PTZCapabilities represents PTZ service capabilities.
type PTZCapabilities struct {
	XAddr string
}

// NetworkCapabilities represents network capabilities.
type NetworkCapabilities struct {
	IPFilter          bool
	ZeroConfiguration bool
	IPVersion6        bool
	DynDNS            bool
	Extension         *NetworkCapabilitiesExtension
}

// SystemCapabilities represents system capabilities.
type SystemCapabilities struct {
	DiscoveryResolve  bool
	DiscoveryBye      bool
	RemoteDiscovery   bool
	SystemBackup      bool
	SystemLogging     bool
	FirmwareUpgrade   bool
	SupportedVersions []string
	Extension         *SystemCapabilitiesExtension
}

// IOCapabilities represents I/O capabilities.
type IOCapabilities struct {
	InputConnectors int
	RelayOutputs    int
	Extension       *IOCapabilitiesExtension
}

// SecurityCapabilities represents security capabilities.
type SecurityCapabilities struct {
	TLS11                bool
	TLS12                bool
	OnboardKeyGeneration bool
	AccessPolicyConfig   bool
	X509Token            bool
	SAMLToken            bool
	KerberosToken        bool
	RELToken             bool
	Extension            *SecurityCapabilitiesExtension
}

// StreamingCapabilities represents streaming capabilities.
type StreamingCapabilities struct {
	RTPMulticast bool
	RTPTCP       bool
	RTPRTSPTCP   bool
	Extension    *StreamingCapabilitiesExtension
}

// CapabilitiesExtension represents extension types for capabilities.
type CapabilitiesExtension struct{}
type NetworkCapabilitiesExtension struct{}
type SystemCapabilitiesExtension struct{}
type IOCapabilitiesExtension struct{}
type SecurityCapabilitiesExtension struct{}
type StreamingCapabilitiesExtension struct{}

// Profile represents a media profile.
type Profile struct {
	Token                     string
	Name                      string
	VideoSourceConfiguration  *VideoSourceConfiguration
	AudioSourceConfiguration  *AudioSourceConfiguration
	VideoEncoderConfiguration *VideoEncoderConfiguration
	AudioEncoderConfiguration *AudioEncoderConfiguration
	PTZConfiguration          *PTZConfiguration
	MetadataConfiguration     *MetadataConfiguration
	Extension                 *ProfileExtension
}

// VideoSourceConfiguration represents video source configuration.
type VideoSourceConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	SourceToken string
	Bounds      *IntRectangle
}

// AudioSourceConfiguration represents audio source configuration.
type AudioSourceConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	SourceToken string
}

// VideoEncoderConfiguration represents video encoder configuration.
type VideoEncoderConfiguration struct {
	Token          string
	Name           string
	UseCount       int
	Encoding       string // JPEG, MPEG4, H264
	Resolution     *VideoResolution
	Quality        float64
	RateControl    *VideoRateControl
	MPEG4          *MPEG4Configuration
	H264           *H264Configuration
	Multicast      *MulticastConfiguration
	SessionTimeout time.Duration
}

// AudioEncoderConfiguration represents audio encoder configuration.
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

// PTZConfiguration represents PTZ configuration.
type PTZConfiguration struct {
	Token                                  string
	Name                                   string
	UseCount                               int
	NodeToken                              string
	DefaultAbsolutePantTiltPositionSpace   string
	DefaultAbsoluteZoomPositionSpace       string
	DefaultRelativePanTiltTranslationSpace string
	DefaultRelativeZoomTranslationSpace    string
	DefaultContinuousPanTiltVelocitySpace  string
	DefaultContinuousZoomVelocitySpace     string
	DefaultPTZSpeed                        *PTZSpeed
	DefaultPTZTimeout                      time.Duration
	PanTiltLimits                          *PanTiltLimits
	ZoomLimits                             *ZoomLimits
}

// MetadataConfiguration represents metadata configuration.
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

// VideoResolution represents video resolution.
type VideoResolution struct {
	Width  int
	Height int
}

// VideoRateControl represents video rate control.
type VideoRateControl struct {
	FrameRateLimit   int
	EncodingInterval int
	BitrateLimit     int
}

// MPEG4Configuration represents MPEG4 configuration.
type MPEG4Configuration struct {
	GovLength    int
	MPEG4Profile string
}

// H264Configuration represents H264 configuration.
type H264Configuration struct {
	GovLength   int
	H264Profile string
}

// MulticastConfiguration represents multicast configuration.
type MulticastConfiguration struct {
	Address   *IPAddress
	Port      int
	TTL       int
	AutoStart bool
}

// IPAddress represents an IP address.
type IPAddress struct {
	Type        string // IPv4 or IPv6
	Address     string
	IPv4Address string
	IPv6Address string
}

// IntRectangle represents a rectangle with integer coordinates.
type IntRectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

// PTZSpeed represents PTZ speed.
type PTZSpeed struct {
	PanTilt *Vector2D
	Zoom    *Vector1D
}

// Vector2D represents a 2D vector.
type Vector2D struct {
	X     float64
	Y     float64
	Space string
}

// Vector1D represents a 1D vector.
type Vector1D struct {
	X     float64
	Space string
}

// PanTiltLimits represents pan/tilt limits.
type PanTiltLimits struct {
	Range *Space2DDescription
}

// ZoomLimits represents zoom limits.
type ZoomLimits struct {
	Range *Space1DDescription
}

// Space2DDescription represents 2D space description.
type Space2DDescription struct {
	URI    string
	XRange *FloatRange
	YRange *FloatRange
}

// Space1DDescription represents 1D space description.
type Space1DDescription struct {
	URI    string
	XRange *FloatRange
}

// FloatRange represents a float range.
type FloatRange struct {
	Min float64
	Max float64
}

// PTZFilter represents PTZ filter.
type PTZFilter struct {
	Status   bool
	Position bool
}

// EventSubscription represents event subscription.
type EventSubscription struct {
	Filter *FilterType
}

// FilterType represents filter type.
type FilterType struct {
	// Simplified for now
}

// ProfileExtension represents profile extension.
type ProfileExtension struct{}

// MediaServiceCapabilities represents media service capabilities.
type MediaServiceCapabilities struct {
	SnapshotURI             bool
	Rotation                bool
	VideoSourceMode         bool
	OSD                     bool
	TemporaryOSDText        bool
	EXICompression          bool
	MaximumNumberOfProfiles int
	RTPMulticast            bool
	RTPTCP                  bool
	RTPRTSPTCP              bool
}

// VideoEncoderConfigurationOptions represents available options for video encoder configuration.
type VideoEncoderConfigurationOptions struct {
	QualityRange *FloatRange
	JPEG         *JPEGOptions
	H264         *H264Options
}

// JPEGOptions represents JPEG encoder options.
type JPEGOptions struct {
	ResolutionsAvailable  []*VideoResolution
	FrameRateRange        *FloatRange
	EncodingIntervalRange *IntRange
}

// H264Options represents H264 encoder options.
type H264Options struct {
	ResolutionsAvailable  []*VideoResolution
	GovLengthRange        *IntRange
	FrameRateRange        *FloatRange
	EncodingIntervalRange *IntRange
	H264ProfilesSupported []string
}

// VideoSourceMode represents a video source mode.
type VideoSourceMode struct {
	Token      string
	Enabled    bool
	Resolution *VideoResolution
}

// OSDConfiguration represents OSD (On-Screen Display) configuration.
type OSDConfiguration struct {
	Token string
	// Additional fields can be added based on ONVIF spec
}

// AudioEncoderConfigurationOptions represents available options for audio encoder configuration.
type AudioEncoderConfigurationOptions struct {
	EncodingOptions []string
	BitrateList     []int
	SampleRateList  []int
}

// MetadataConfigurationOptions represents available options for metadata configuration.
type MetadataConfigurationOptions struct {
	PTZStatusFilterOptions *PTZFilter
}

// AudioOutputConfiguration represents audio output configuration.
type AudioOutputConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	OutputToken string
}

// AudioOutputConfigurationOptions represents available options for audio output configuration.
type AudioOutputConfigurationOptions struct {
	OutputTokensAvailable []string
}

// AudioDecoderConfigurationOptions represents available options for audio decoder configuration.
type AudioDecoderConfigurationOptions struct {
	AACDecOptions  *AudioDecoderOptions
	G711DecOptions *AudioDecoderOptions
	G726DecOptions *AudioDecoderOptions
}

// AudioDecoderOptions represents audio decoder options.
type AudioDecoderOptions struct {
	BitrateList    []int
	SampleRateList []int
}

// GuaranteedNumberOfVideoEncoderInstances represents guaranteed number of video encoder instances.
type GuaranteedNumberOfVideoEncoderInstances struct {
	TotalNumber int
	JPEG        int
	H264        int
	MPEG4       int
}

// OSDConfigurationOptions represents available options for OSD configuration.
type OSDConfigurationOptions struct {
	MaximumNumberOfOSDs int
}

// VideoSourceConfigurationOptions represents available options for video source configuration.
type VideoSourceConfigurationOptions struct {
	BoundsRange                *BoundsRange
	VideoSourceTokensAvailable []string
}

// AudioSourceConfigurationOptions represents available options for audio source configuration.
type AudioSourceConfigurationOptions struct {
	InputTokensAvailable []string
}

// BoundsRange represents bounds range for video source configuration.
type BoundsRange struct {
	X      *IntRange
	Y      *IntRange
	Width  *IntRange
	Height *IntRange
}

// AudioDecoderConfiguration represents audio decoder configuration.
type AudioDecoderConfiguration struct {
	Token    string
	Name     string
	UseCount int
}

// VideoAnalyticsConfiguration represents video analytics configuration.
type VideoAnalyticsConfiguration struct {
	Token                        string
	Name                         string
	UseCount                     int
	AnalyticsEngineConfiguration *AnalyticsEngineConfiguration
	RuleEngineConfiguration      *RuleEngineConfiguration
}

// AnalyticsEngineConfiguration represents analytics engine configuration.
type AnalyticsEngineConfiguration struct {
	AnalyticsEngine *Config
	Parameters      *ItemList
}

// RuleEngineConfiguration represents rule engine configuration.
type RuleEngineConfiguration struct {
	Rule *Config
}

// Config represents a generic configuration.
type Config struct {
	Parameters *ItemList
}

// ItemList represents a list of configuration items.
type ItemList struct {
	SimpleItem  []SimpleItem
	ElementItem []ElementItem
}

// SimpleItem represents a simple configuration item.
type SimpleItem struct {
	Name  string
	Value string
}

// ElementItem represents an element configuration item.
type ElementItem struct {
	Name string
}

// VideoAnalyticsConfigurationOptions represents available options for video analytics configuration.
type VideoAnalyticsConfigurationOptions struct {
	// Simplified for now - can be expanded based on ONVIF spec
}

// StreamSetup represents stream setup parameters.
type StreamSetup struct {
	Stream    string // RTP-Unicast, RTP-Multicast
	Transport *Transport
}

// Transport represents transport parameters.
type Transport struct {
	Protocol string // UDP, TCP, RTSP, HTTP
	Tunnel   *Tunnel
}

// Tunnel represents tunnel parameters.
type Tunnel struct{}

// MediaURI represents a media URI.
type MediaURI struct {
	URI                 string
	InvalidAfterConnect bool
	InvalidAfterReboot  bool
	Timeout             time.Duration
}

// PTZStatus represents PTZ status.
type PTZStatus struct {
	Position   *PTZVector
	MoveStatus *PTZMoveStatus
	Error      string
	UTCTime    time.Time
}

// PTZVector represents PTZ position.
type PTZVector struct {
	PanTilt *Vector2D
	Zoom    *Vector1D
}

// PTZMoveStatus represents PTZ movement status.
type PTZMoveStatus struct {
	PanTilt string // IDLE, MOVING, UNKNOWN
	Zoom    string // IDLE, MOVING, UNKNOWN
}

// PTZPreset represents a PTZ preset.
type PTZPreset struct {
	Token       string
	Name        string
	PTZPosition *PTZVector
}

// ImagingSettings represents imaging settings.
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

// BacklightCompensation represents backlight compensation.
type BacklightCompensation struct {
	Mode  string // OFF, ON
	Level float64
}

// Exposure represents exposure settings.
type Exposure struct {
	Mode            string // AUTO, MANUAL
	Priority        string // LowNoise, FrameRate
	MinExposureTime float64
	MaxExposureTime float64
	MinGain         float64
	MaxGain         float64
	MinIris         float64
	MaxIris         float64
	ExposureTime    float64
	Gain            float64
	Iris            float64
}

// FocusConfiguration represents focus configuration.
type FocusConfiguration struct {
	AutoFocusMode string // AUTO, MANUAL
	DefaultSpeed  float64
	NearLimit     float64
	FarLimit      float64
}

// WideDynamicRange represents WDR settings.
type WideDynamicRange struct {
	Mode  string // OFF, ON
	Level float64
}

// WhiteBalance represents white balance settings.
type WhiteBalance struct {
	Mode   string // AUTO, MANUAL
	CrGain float64
	CbGain float64
}

// ImagingSettingsExtension represents imaging settings extension.
type ImagingSettingsExtension struct{}

// HostnameInformation represents hostname configuration.
type HostnameInformation struct {
	FromDHCP bool
	Name     string
}

// DNSInformation represents DNS configuration.
type DNSInformation struct {
	FromDHCP     bool
	SearchDomain []string
	DNSFromDHCP  []IPAddress
	DNSManual    []IPAddress
}

// NTPInformation represents NTP configuration.
type NTPInformation struct {
	FromDHCP    bool
	NTPFromDHCP []NetworkHost
	NTPManual   []NetworkHost
}

// NetworkHost represents a network host.
type NetworkHost struct {
	Type        string // IPv4, IPv6, DNS
	IPv4Address string
	IPv6Address string
	DNSname     string
}

// NetworkInterface represents a network interface.
type NetworkInterface struct {
	Token   string
	Enabled bool
	Info    NetworkInterfaceInfo
	IPv4    *IPv4NetworkInterface
	IPv6    *IPv6NetworkInterface
}

// NetworkInterfaceInfo represents network interface info.
type NetworkInterfaceInfo struct {
	Name      string
	HwAddress string
	MTU       int
}

// IPv4NetworkInterface represents IPv4 configuration.
type IPv4NetworkInterface struct {
	Enabled bool
	Config  IPv4Configuration
}

// IPv6NetworkInterface represents IPv6 configuration.
type IPv6NetworkInterface struct {
	Enabled bool
	Config  IPv6Configuration
}

// IPv4Configuration represents IPv4 configuration.
type IPv4Configuration struct {
	Manual []PrefixedIPv4Address
	DHCP   bool
}

// IPv6Configuration represents IPv6 configuration.
type IPv6Configuration struct {
	Manual []PrefixedIPv6Address
	DHCP   bool
}

// PrefixedIPv4Address represents an IPv4 address with prefix.
type PrefixedIPv4Address struct {
	Address      string
	PrefixLength int
}

// PrefixedIPv6Address represents an IPv6 address with prefix.
type PrefixedIPv6Address struct {
	Address      string
	PrefixLength int
}

// Scope represents a device scope.
type Scope struct {
	ScopeDef  string
	ScopeItem string
}

// User represents a user account.
type User struct {
	Username  string
	Password  string
	UserLevel string // Administrator, Operator, User
}

// VideoSource represents a video source.
type VideoSource struct {
	Token      string
	Framerate  float64
	Resolution *VideoResolution
	Imaging    *ImagingSettings
}

// AudioSource represents an audio source.
type AudioSource struct {
	Token    string
	Channels int
}

// AudioOutput represents an audio output.
type AudioOutput struct {
	Token string
}

// ImagingOptions represents available imaging options.
type ImagingOptions struct {
	BacklightCompensation *BacklightCompensationOptions
	Brightness            *FloatRange
	ColorSaturation       *FloatRange
	Contrast              *FloatRange
	Exposure              *ExposureOptions
	Focus                 *FocusOptions
	IrCutFilterModes      []string
	Sharpness             *FloatRange
	WideDynamicRange      *WideDynamicRangeOptions
	WhiteBalance          *WhiteBalanceOptions
}

// BacklightCompensationOptions represents backlight compensation options.
type BacklightCompensationOptions struct {
	Mode  []string
	Level *FloatRange
}

// ExposureOptions represents exposure options.
type ExposureOptions struct {
	Mode            []string
	Priority        []string
	MinExposureTime *FloatRange
	MaxExposureTime *FloatRange
	MinGain         *FloatRange
	MaxGain         *FloatRange
	MinIris         *FloatRange
	MaxIris         *FloatRange
	ExposureTime    *FloatRange
	Gain            *FloatRange
	Iris            *FloatRange
}

// FocusOptions represents focus options.
type FocusOptions struct {
	AutoFocusModes []string
	DefaultSpeed   *FloatRange
	NearLimit      *FloatRange
	FarLimit       *FloatRange
}

// WideDynamicRangeOptions represents WDR options.
type WideDynamicRangeOptions struct {
	Mode  []string
	Level *FloatRange
}

// WhiteBalanceOptions represents white balance options.
type WhiteBalanceOptions struct {
	Mode   []string
	YrGain *FloatRange
	YbGain *FloatRange
}

// MoveOptions represents imaging move options.
type MoveOptions struct {
	Absolute   *AbsoluteFocusOptions
	Relative   *RelativeFocusOptions
	Continuous *ContinuousFocusOptions
}

// AbsoluteFocusOptions represents absolute focus options.
type AbsoluteFocusOptions struct {
	Position FloatRange
	Speed    FloatRange
}

// RelativeFocusOptions represents relative focus options.
type RelativeFocusOptions struct {
	Distance FloatRange
	Speed    FloatRange
}

// ContinuousFocusOptions represents continuous focus options.
type ContinuousFocusOptions struct {
	Speed FloatRange
}

// ImagingStatus represents imaging status.
type ImagingStatus struct {
	FocusStatus *FocusStatus
}

// FocusStatus represents focus status.
type FocusStatus struct {
	Position   float64
	MoveStatus string
	Error      string
}

// Service represents an ONVIF service.
type Service struct {
	Namespace    string
	XAddr        string
	Capabilities interface{}
	Version      OnvifVersion
}

// OnvifVersion represents ONVIF version.
type OnvifVersion struct {
	Major int
	Minor int
}

// DeviceServiceCapabilities represents device service capabilities.
type DeviceServiceCapabilities struct {
	Network  *NetworkCapabilities
	Security *SecurityCapabilities
	System   *SystemCapabilities
	Misc     *MiscCapabilities
}

// MiscCapabilities represents miscellaneous capabilities.
type MiscCapabilities struct {
	AuxiliaryCommands []string
}

// DiscoveryMode represents discovery mode.
type DiscoveryMode string

const (
	DiscoveryModeDiscoverable    DiscoveryMode = "Discoverable"
	DiscoveryModeNonDiscoverable DiscoveryMode = "NonDiscoverable"
)

// NetworkProtocol represents network protocol configuration.
type NetworkProtocol struct {
	Name    NetworkProtocolType
	Enabled bool
	Port    []int
}

// NetworkProtocolType represents protocol type.
type NetworkProtocolType string

const (
	NetworkProtocolHTTP  NetworkProtocolType = "HTTP"
	NetworkProtocolHTTPS NetworkProtocolType = "HTTPS"
	NetworkProtocolRTSP  NetworkProtocolType = "RTSP"
)

// NetworkGateway represents default gateway.
type NetworkGateway struct {
	IPv4Address []string
	IPv6Address []string
}

// SystemDateTime represents system date and time.
type SystemDateTime struct {
	DateTimeType    SetDateTimeType
	DaylightSavings bool
	TimeZone        *TimeZone
	UTCDateTime     *DateTime
	LocalDateTime   *DateTime
}

// SetDateTimeType represents date/time set method.
type SetDateTimeType string

const (
	SetDateTimeManual SetDateTimeType = "Manual"
	SetDateTimeNTP    SetDateTimeType = "NTP"
)

// TimeZone represents timezone.
type TimeZone struct {
	TZ string // POSIX format
}

// DateTime represents date and time.
type DateTime struct {
	Time Time
	Date Date
}

// Time represents time.
type Time struct {
	Hour   int
	Minute int
	Second int
}

// Date represents date.
type Date struct {
	Year  int
	Month int
	Day   int
}

// SystemLogType represents system log type.
type SystemLogType string

const (
	SystemLogTypeSystem SystemLogType = "System"
	SystemLogTypeAccess SystemLogType = "Access"
)

// SystemLog represents system log data.
type SystemLog struct {
	Binary *AttachmentData
	String string
}

// AttachmentData represents attachment/binary data.
type AttachmentData struct {
	ContentType string
	Include     *Include
}

// Include represents XOP include.
type Include struct {
	Href string
}

// BackupFile represents backup file.
type BackupFile struct {
	Name string
	Data AttachmentData
}

// FactoryDefaultType represents factory default type.
type FactoryDefaultType string

const (
	FactoryDefaultHard FactoryDefaultType = "Hard"
	FactoryDefaultSoft FactoryDefaultType = "Soft"
)

// RelayOutput represents relay output.
type RelayOutput struct {
	Token      string
	Properties RelayOutputSettings
}

// RelayOutputSettings represents relay output settings.
type RelayOutputSettings struct {
	Mode      RelayMode
	DelayTime time.Duration
	IdleState RelayIdleState
}

// RelayMode represents relay mode.
type RelayMode string

const (
	RelayModeMonostable RelayMode = "Monostable"
	RelayModeBistable   RelayMode = "Bistable"
)

// RelayIdleState represents relay idle state.
type RelayIdleState string

const (
	RelayIdleStateClosed RelayIdleState = "closed"
	RelayIdleStateOpen   RelayIdleState = "open"
)

// RelayLogicalState represents relay logical state.
type RelayLogicalState string

const (
	RelayLogicalStateActive   RelayLogicalState = "active"
	RelayLogicalStateInactive RelayLogicalState = "inactive"
)

// AuxiliaryData represents auxiliary command data.
type AuxiliaryData string

// SupportInformation represents support information.
type SupportInformation struct {
	Binary *AttachmentData
	String string
}

// SystemLogURIList represents system log URIs.
type SystemLogURIList struct {
	SystemLog []SystemLogURI
}

// SystemLogURI represents system log URI.
type SystemLogURI struct {
	Type SystemLogType
	URI  string
}

// NetworkZeroConfiguration represents zero-configuration.
type NetworkZeroConfiguration struct {
	InterfaceToken string
	Enabled        bool
	Addresses      []string
}

// DynamicDNSInformation represents dynamic DNS info.
type DynamicDNSInformation struct {
	Type DynamicDNSType
	Name string
	TTL  time.Duration
}

// DynamicDNSType represents dynamic DNS type.
type DynamicDNSType string

const (
	DynamicDNSNoUpdate      DynamicDNSType = "NoUpdate"
	DynamicDNSClientUpdates DynamicDNSType = "ClientUpdates"
	DynamicDNSServerUpdates DynamicDNSType = "ServerUpdates"
)

// IPAddressFilter represents IP address filter.
type IPAddressFilter struct {
	Type        IPAddressFilterType
	IPv4Address []PrefixedIPv4Address
	IPv6Address []PrefixedIPv6Address
}

// IPAddressFilterType represents filter type.
type IPAddressFilterType string

const (
	IPAddressFilterAllow IPAddressFilterType = "Allow"
	IPAddressFilterDeny  IPAddressFilterType = "Deny"
)

// RemoteUser represents remote user configuration.
type RemoteUser struct {
	Username           string
	Password           string
	UseDerivedPassword bool
}

// Certificate represents a certificate.
type Certificate struct {
	CertificateID string
	Certificate   BinaryData
}

// BinaryData represents binary data.
type BinaryData struct {
	ContentType string
	Data        []byte
}

// CertificateStatus represents certificate status.
type CertificateStatus struct {
	CertificateID string
	Status        bool
}

// CertificateInformation represents certificate information.
type CertificateInformation struct {
	CertificateID      string
	IssuerDN           string
	SubjectDN          string
	KeyUsage           *CertificateUsage
	ExtendedKeyUsage   *CertificateUsage
	KeyLength          int
	Version            string
	SerialNum          string
	SignatureAlgorithm string
	Validity           *DateTimeRange
}

// CertificateUsage represents certificate usage.
type CertificateUsage struct {
	Critical bool
	Value    string
}

// DateTimeRange represents date/time range.
type DateTimeRange struct {
	From  time.Time
	Until time.Time
}

// Dot11Capabilities represents 802.11 capabilities.
type Dot11Capabilities struct {
	TKIP                  bool
	ScanAvailableNetworks bool
	MultipleConfiguration bool
	AdHocStationMode      bool
	WEP                   bool
}

// Dot11Status represents 802.11 status.
type Dot11Status struct {
	SSID              string
	BSSID             string
	PairCipher        Dot11Cipher
	GroupCipher       Dot11Cipher
	SignalStrength    Dot11SignalStrength
	ActiveConfigAlias string
}

// Dot11Cipher represents 802.11 cipher.
type Dot11Cipher string

const (
	Dot11CipherCCMP     Dot11Cipher = "CCMP"
	Dot11CipherTKIP     Dot11Cipher = "TKIP"
	Dot11CipherAny      Dot11Cipher = "Any"
	Dot11CipherExtended Dot11Cipher = "Extended"
)

// Dot11SignalStrength represents signal strength.
type Dot11SignalStrength string

const (
	Dot11SignalNone     Dot11SignalStrength = "None"
	Dot11SignalVeryBad  Dot11SignalStrength = "Very Bad"
	Dot11SignalBad      Dot11SignalStrength = "Bad"
	Dot11SignalGood     Dot11SignalStrength = "Good"
	Dot11SignalVeryGood Dot11SignalStrength = "Very Good"
	Dot11SignalExtended Dot11SignalStrength = "Extended"
)

// Dot1XConfiguration represents 802.1X configuration.
type Dot1XConfiguration struct {
	Dot1XConfigurationToken string
	Identity                string
	AnonymousID             string
	EAPMethod               int
	CACertificateID         []string
	EAPMethodConfiguration  *EAPMethodConfiguration
}

// EAPMethodConfiguration represents EAP method configuration.
type EAPMethodConfiguration struct {
	TLSConfiguration *TLSConfiguration
	Password         string
}

// TLSConfiguration represents TLS configuration.
type TLSConfiguration struct {
	CertificateID string
}

// Dot11AvailableNetworks represents available 802.11 networks.
type Dot11AvailableNetworks struct {
	SSID                  string
	BSSID                 string
	AuthAndMangementSuite []Dot11AuthAndMangementSuite
	PairCipher            []Dot11Cipher
	GroupCipher           []Dot11Cipher
	SignalStrength        Dot11SignalStrength
}

// Dot11AuthAndMangementSuite represents auth suite.
type Dot11AuthAndMangementSuite string

const (
	Dot11AuthNone     Dot11AuthAndMangementSuite = "None"
	Dot11AuthDot1X    Dot11AuthAndMangementSuite = "Dot1X"
	Dot11AuthPSK      Dot11AuthAndMangementSuite = "PSK"
	Dot11AuthExtended Dot11AuthAndMangementSuite = "Extended"
)

// StorageConfiguration represents storage configuration.
type StorageConfiguration struct {
	Token string
	Data  StorageConfigurationData
}

// StorageConfigurationData represents storage configuration data.
type StorageConfigurationData struct {
	Type                       string
	LocalPath                  string
	StorageURI                 string
	User                       *UserCredential
	CertPathValidationPolicyID string
}

// UserCredential represents user credentials.
type UserCredential struct {
	UserName string
	Password string
	Token    string
}

// LocationEntity represents geo location.
type LocationEntity struct {
	Entity    string  `xml:"Entity"`
	Token     string  `xml:"Token"`
	Fixed     bool    `xml:"Fixed"`
	Lon       float64 `xml:"Lon,attr"`
	Lat       float64 `xml:"Lat,attr"`
	Elevation float64 `xml:"Elevation,attr"`
}

// GeoLocation represents geographic location coordinates.
type GeoLocation struct {
	Lon       float64 `xml:"lon,attr,omitempty"`       // Longitude in degrees
	Lat       float64 `xml:"lat,attr,omitempty"`       // Latitude in degrees
	Elevation float64 `xml:"elevation,attr,omitempty"` // Elevation in meters
}

// AccessPolicy represents device access policy configuration.
type AccessPolicy struct {
	PolicyFile *BinaryData
}

// PasswordComplexityConfiguration represents password complexity config.
type PasswordComplexityConfiguration struct {
	MinLen                    int
	Uppercase                 int
	Number                    int
	SpecialChars              int
	BlockUsernameOccurrence   bool
	PolicyConfigurationLocked bool
}

// PasswordHistoryConfiguration represents password history config.
type PasswordHistoryConfiguration struct {
	Enabled bool
	Length  int
}

// AuthFailureWarningConfiguration represents auth failure warning config.
type AuthFailureWarningConfiguration struct {
	Enabled         bool
	MonitorPeriod   int
	MaxAuthFailures int
}

// IntRange represents integer range.
type IntRange struct {
	Min int
	Max int
}
