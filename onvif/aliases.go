package onvif

// v1 compatibility aliases: the v2 package split moved domain symbols into
// their packages; these aliases keep most v1-style source compiling after
// the import-path change to /v2.

import (
	"github.com/mickeyzzc/onvif-go/v2/device"
	"github.com/mickeyzzc/onvif-go/v2/deviceio"
	"github.com/mickeyzzc/onvif-go/v2/events"
	"github.com/mickeyzzc/onvif-go/v2/imaging"
	"github.com/mickeyzzc/onvif-go/v2/media"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
	"github.com/mickeyzzc/onvif-go/v2/security"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Service facade type aliases (v1 names).
type (
	DeviceService   = device.Service
	MediaService    = media.Service
	PTZService      = ptz.Service
	ImagingService  = imaging.Service
	EventService    = events.Service
	DeviceIOService = deviceio.Service
	SecurityService = security.Service
)

type (
	DeviceInformation          = device.DeviceInformation
	DeviceTime                 = device.DeviceTime
	Capabilities               = device.Capabilities
	AnalyticsCapabilities      = device.AnalyticsCapabilities
	DeviceCapabilities         = device.DeviceCapabilities
	EventCapabilities          = device.EventCapabilities
	ImagingCapabilities        = device.ImagingCapabilities
	MediaCapabilities          = device.MediaCapabilities
	PTZCapabilities            = device.PTZCapabilities
	NetworkCapabilities        = device.NetworkCapabilities
	SystemCapabilities         = device.SystemCapabilities
	IOCapabilities             = device.IOCapabilities
	SecurityCapabilities       = device.SecurityCapabilities
	StreamingCapabilities      = device.StreamingCapabilities
	HostnameInformation        = device.HostnameInformation
	DNSInformation             = device.DNSInformation
	NTPInformation             = device.NTPInformation
	NetworkHost                = device.NetworkHost
	NetworkInterface           = device.NetworkInterface
	NetworkInterfaceInfo       = device.NetworkInterfaceInfo
	IPv4NetworkInterface       = device.IPv4NetworkInterface
	IPv6NetworkInterface       = device.IPv6NetworkInterface
	IPv4Configuration          = device.IPv4Configuration
	Scope                      = device.Scope
	OnvifVersion               = device.OnvifVersion
	DeviceServiceCapabilities  = device.DeviceServiceCapabilities
	MiscCapabilities           = device.MiscCapabilities
	DiscoveryMode              = device.DiscoveryMode
	NetworkProtocol            = device.NetworkProtocol
	NetworkProtocolType        = device.NetworkProtocolType
	NetworkGateway             = device.NetworkGateway
	SystemDateTime             = device.SystemDateTime
	SetDateTimeType            = device.SetDateTimeType
	TimeZone                   = device.TimeZone
	DateTime                   = device.DateTime
	Time                       = device.Time
	Date                       = device.Date
	SystemLogType              = device.SystemLogType
	SystemLog                  = device.SystemLog
	AttachmentData             = device.AttachmentData
	Include                    = device.Include
	BackupFile                 = device.BackupFile
	FactoryDefaultType         = device.FactoryDefaultType
	SystemLogURIList           = device.SystemLogURIList
	SystemLogURI               = device.SystemLogURI
	NetworkZeroConfiguration   = device.NetworkZeroConfiguration
	DynamicDNSInformation      = device.DynamicDNSInformation
	DynamicDNSType             = device.DynamicDNSType
	Dot11Capabilities          = device.Dot11Capabilities
	Dot11Status                = device.Dot11Status
	Dot11Cipher                = device.Dot11Cipher
	Dot11SignalStrength        = device.Dot11SignalStrength
	Dot1XConfiguration         = device.Dot1XConfiguration
	EAPMethodConfiguration     = device.EAPMethodConfiguration
	TLSConfiguration           = device.TLSConfiguration
	Dot11AvailableNetworks     = device.Dot11AvailableNetworks
	Dot11AuthAndMangementSuite = device.Dot11AuthAndMangementSuite
	StorageConfiguration       = device.StorageConfiguration
	StorageConfigurationData   = device.StorageConfigurationData
	UserCredential             = device.UserCredential
	LocationEntity             = device.LocationEntity
	GeoLocation                = device.GeoLocation
	SupportInformation         = device.SupportInformation
	ServiceEntry               = device.ServiceEntry
)

var NetmaskFromPrefixLength = device.NetmaskFromPrefixLength

type (
	User                            = security.User
	IPAddressFilter                 = security.IPAddressFilter
	IPAddressFilterType             = security.IPAddressFilterType
	RemoteUser                      = security.RemoteUser
	Certificate                     = security.Certificate
	BinaryData                      = security.BinaryData
	CertificateStatus               = security.CertificateStatus
	CertificateInformation          = security.CertificateInformation
	CertificateUsage                = security.CertificateUsage
	DateTimeRange                   = security.DateTimeRange
	AccessPolicy                    = security.AccessPolicy
	PasswordComplexityConfiguration = security.PasswordComplexityConfiguration
	PasswordHistoryConfiguration    = security.PasswordHistoryConfiguration
	AuthFailureWarningConfiguration = security.AuthFailureWarningConfiguration
)

type (
	RelayOutput         = deviceio.RelayOutput
	RelayOutputSettings = deviceio.RelayOutputSettings
	RelayMode           = deviceio.RelayMode
	RelayIdleState      = deviceio.RelayIdleState
	RelayLogicalState   = deviceio.RelayLogicalState
)

type (
	Profile                                 = media.Profile
	ProfileExtension                        = media.ProfileExtension
	VideoSourceConfiguration                = media.VideoSourceConfiguration
	AudioSourceConfiguration                = media.AudioSourceConfiguration
	VideoEncoderConfiguration               = media.VideoEncoderConfiguration
	AudioEncoderConfiguration               = media.AudioEncoderConfiguration
	MetadataConfiguration                   = media.MetadataConfiguration
	EventSubscription                       = media.EventSubscription
	FilterType                              = media.FilterType
	VideoResolution                         = media.VideoResolution
	VideoRateControl                        = media.VideoRateControl
	MPEG4Configuration                      = media.MPEG4Configuration
	H264Configuration                       = media.H264Configuration
	MulticastConfiguration                  = media.MulticastConfiguration
	MediaServiceCapabilities                = media.MediaServiceCapabilities
	VideoEncoderConfigurationOptions        = media.VideoEncoderConfigurationOptions
	JPEGOptions                             = media.JPEGOptions
	H264Options                             = media.H264Options
	VideoSourceMode                         = media.VideoSourceMode
	OSDConfiguration                        = media.OSDConfiguration
	AudioEncoderConfigurationOptions        = media.AudioEncoderConfigurationOptions
	MetadataConfigurationOptions            = media.MetadataConfigurationOptions
	AudioOutputConfiguration                = media.AudioOutputConfiguration
	AudioOutputConfigurationOptions         = media.AudioOutputConfigurationOptions
	AudioDecoderConfigurationOptions        = media.AudioDecoderConfigurationOptions
	AudioDecoderOptions                     = media.AudioDecoderOptions
	GuaranteedNumberOfVideoEncoderInstances = media.GuaranteedNumberOfVideoEncoderInstances
	OSDConfigurationOptions                 = media.OSDConfigurationOptions
	VideoSourceConfigurationOptions         = media.VideoSourceConfigurationOptions
	AudioSourceConfigurationOptions         = media.AudioSourceConfigurationOptions
	BoundsRange                             = media.BoundsRange
	AudioDecoderConfiguration               = media.AudioDecoderConfiguration
	VideoAnalyticsConfiguration             = media.VideoAnalyticsConfiguration
	AnalyticsEngineConfiguration            = media.AnalyticsEngineConfiguration
	RuleEngineConfiguration                 = media.RuleEngineConfiguration
	Config                                  = media.Config
	ItemList                                = media.ItemList
	VideoAnalyticsConfigurationOptions      = media.VideoAnalyticsConfigurationOptions
	StreamSetup                             = media.StreamSetup
	Transport                               = media.Transport
	Tunnel                                  = media.Tunnel
	MediaURI                                = media.MediaURI
	VideoSource                             = media.VideoSource
	AudioSource                             = media.AudioSource
	AudioOutput                             = media.AudioOutput
)

const (
	StreamRTPUnicast   = media.StreamRTPUnicast
	StreamRTPMulticast = media.StreamRTPMulticast
	ProtocolRTSP       = media.ProtocolRTSP
	ProtocolHTTP       = media.ProtocolHTTP
	ProtocolUDP        = media.ProtocolUDP
)

var ErrEmptyMediaURI = media.ErrEmptyMediaURI

type (
	PTZConfiguration   = ptz.PTZConfiguration
	PTZSpeed           = ptz.PTZSpeed
	Vector2D           = ptz.Vector2D
	Vector1D           = ptz.Vector1D
	PanTiltLimits      = ptz.PanTiltLimits
	ZoomLimits         = ptz.ZoomLimits
	Space2DDescription = ptz.Space2DDescription
	Space1DDescription = ptz.Space1DDescription
	PTZFilter          = ptz.PTZFilter
	PTZStatus          = ptz.PTZStatus
	PTZVector          = ptz.PTZVector
	PTZMoveStatus      = ptz.PTZMoveStatus
	PTZPreset          = ptz.PTZPreset
	AuxiliaryData      = ptz.AuxiliaryData
)

type (
	ImagingSettings              = imaging.ImagingSettings
	BacklightCompensation        = imaging.BacklightCompensation
	Exposure                     = imaging.Exposure
	FocusConfiguration           = imaging.FocusConfiguration
	WideDynamicRange             = imaging.WideDynamicRange
	WhiteBalance                 = imaging.WhiteBalance
	ImagingSettingsExtension     = imaging.ImagingSettingsExtension
	ImagingOptions               = imaging.ImagingOptions
	BacklightCompensationOptions = imaging.BacklightCompensationOptions
	ExposureOptions              = imaging.ExposureOptions
	FocusOptions                 = imaging.FocusOptions
	WideDynamicRangeOptions      = imaging.WideDynamicRangeOptions
	WhiteBalanceOptions          = imaging.WhiteBalanceOptions
	MoveOptions                  = imaging.MoveOptions
	AbsoluteFocusOptions         = imaging.AbsoluteFocusOptions
	RelativeFocusOptions         = imaging.RelativeFocusOptions
	ContinuousFocusOptions       = imaging.ContinuousFocusOptions
	ImagingStatus                = imaging.ImagingStatus
	FocusStatus                  = imaging.FocusStatus
)

type (
	NotificationMessage      = events.NotificationMessage
	EventMessage             = events.EventMessage
	TopicSet                 = events.TopicSet
	Topic                    = events.Topic
	EventBrokerConfig        = events.EventBrokerConfig
	EventProperties          = events.EventProperties
	EventServiceCapabilities = events.EventServiceCapabilities
	PullPointSubscription    = events.PullPointSubscription
	SubscribeEventsOptions   = events.SubscribeEventsOptions
	EventStream              = events.EventStream
	PushSubscription         = events.PushSubscription
	SubscribePushOptions     = events.SubscribePushOptions
	PushStream               = events.PushStream
	MotionAlarm              = events.MotionAlarm
)

var ErrEventsNotSupported = events.ErrEventsNotSupported

// ErrNoCredentials is the device-service write gate: admin operations fail
// fast when the client was constructed without credentials.
var ErrNoCredentials = device.ErrNoCredentials

// Push receiving (WS-BaseNotification): the probe-fallback sentinel, the
// Notify parser, and the consumer HTTP endpoint.
var (
	ErrPushNotSupported = events.ErrPushNotSupported
	ParseNotifyEvents   = events.ParseNotifyEvents
	NotifyHandler       = events.NotifyHandler
)

type (
	IPAddress    = types.IPAddress
	IntRectangle = types.IntRectangle
	FloatRange   = types.FloatRange
	IntRange     = types.IntRange
	SimpleItem   = types.SimpleItem
	ElementItem  = types.ElementItem
)

var (
	ErrInvalidParameter    = types.ErrInvalidParameter
	ErrServiceNotSupported = types.ErrServiceNotSupported
)

const (
	DiscoveryModeDiscoverable    = device.DiscoveryModeDiscoverable
	DiscoveryModeNonDiscoverable = device.DiscoveryModeNonDiscoverable
	NetworkProtocolHTTP          = device.NetworkProtocolHTTP
	NetworkProtocolHTTPS         = device.NetworkProtocolHTTPS
	NetworkProtocolRTSP          = device.NetworkProtocolRTSP
	SetDateTimeManual            = device.SetDateTimeManual
	SetDateTimeNTP               = device.SetDateTimeNTP
	SystemLogTypeSystem          = device.SystemLogTypeSystem
	SystemLogTypeAccess          = device.SystemLogTypeAccess
	FactoryDefaultHard           = device.FactoryDefaultHard
	FactoryDefaultSoft           = device.FactoryDefaultSoft
	DynamicDNSNoUpdate           = device.DynamicDNSNoUpdate
	DynamicDNSClientUpdates      = device.DynamicDNSClientUpdates
	DynamicDNSServerUpdates      = device.DynamicDNSServerUpdates
	Dot11CipherCCMP              = device.Dot11CipherCCMP
	Dot11CipherTKIP              = device.Dot11CipherTKIP
	Dot11CipherAny               = device.Dot11CipherAny
	Dot11CipherExtended          = device.Dot11CipherExtended
	Dot11SignalNone              = device.Dot11SignalNone
	Dot11SignalVeryBad           = device.Dot11SignalVeryBad
	Dot11SignalBad               = device.Dot11SignalBad
	Dot11SignalGood              = device.Dot11SignalGood
	Dot11SignalVeryGood          = device.Dot11SignalVeryGood
	Dot11SignalExtended          = device.Dot11SignalExtended
	Dot11AuthNone                = device.Dot11AuthNone
	Dot11AuthDot1X               = device.Dot11AuthDot1X
	Dot11AuthPSK                 = device.Dot11AuthPSK
	Dot11AuthExtended            = device.Dot11AuthExtended
	IPAddressFilterAllow         = security.IPAddressFilterAllow
	IPAddressFilterDeny          = security.IPAddressFilterDeny
	RelayModeMonostable          = deviceio.RelayModeMonostable
	RelayModeBistable            = deviceio.RelayModeBistable
	RelayIdleStateClosed         = deviceio.RelayIdleStateClosed
	RelayIdleStateOpen           = deviceio.RelayIdleStateOpen
	RelayLogicalStateActive      = deviceio.RelayLogicalStateActive
	RelayLogicalStateInactive    = deviceio.RelayLogicalStateInactive
)

type (
	CapabilitiesExtension          = device.CapabilitiesExtension
	NetworkCapabilitiesExtension   = device.NetworkCapabilitiesExtension
	SystemCapabilitiesExtension    = device.SystemCapabilitiesExtension
	IOCapabilitiesExtension        = device.IOCapabilitiesExtension
	SecurityCapabilitiesExtension  = device.SecurityCapabilitiesExtension
	StreamingCapabilitiesExtension = device.StreamingCapabilitiesExtension
)

// Function aliases (v1 names moved with their domains).
var (
	SelectMainProfile = media.SelectMainProfile
	SelectSubProfile  = media.SelectSubProfile
)

// Prefixed address aliases (moved to the shared types leaf).
type (
	PrefixedIPv4Address = types.PrefixedIPv4Address
	PrefixedIPv6Address = types.PrefixedIPv6Address
)

// Device-IO symbols that lived outside types.go in v1.
type (
	DigitalInput     = deviceio.DigitalInput
	DigitalIdleState = deviceio.DigitalIdleState
	SerialPortType   = deviceio.SerialPortType
	ParityBit        = deviceio.ParityBit
)

const (
	DigitalIdleOpen       = deviceio.DigitalIdleOpen
	DigitalIdleClosed     = deviceio.DigitalIdleClosed
	SerialPortTypeRS232   = deviceio.SerialPortTypeRS232
	SerialPortTypeRS422   = deviceio.SerialPortTypeRS422
	SerialPortTypeRS485   = deviceio.SerialPortTypeRS485
	SerialPortTypeGeneric = deviceio.SerialPortTypeGeneric
	ParityNone            = deviceio.ParityNone
	ParityOdd             = deviceio.ParityOdd
	ParityEven            = deviceio.ParityEven
	ParityMark            = deviceio.ParityMark
	ParitySpace           = deviceio.ParitySpace
)

// Domain error sentinels that moved with their packages.
var (
	ErrDigitalInputConfigNil        = deviceio.ErrDigitalInputConfigNil
	ErrEventBrokerConfigNil         = events.ErrEventBrokerConfigNil
	ErrInvalidDigitalInputToken     = deviceio.ErrInvalidDigitalInputToken
	ErrInvalidEventBrokerAddress    = events.ErrInvalidEventBrokerAddress
	ErrInvalidFilter                = events.ErrInvalidFilter
	ErrInvalidMessageLimit          = events.ErrInvalidMessageLimit
	ErrInvalidRelayOutputToken      = deviceio.ErrInvalidRelayOutputToken
	ErrInvalidSerialData            = deviceio.ErrInvalidSerialData
	ErrInvalidSerialPortToken       = deviceio.ErrInvalidSerialPortToken
	ErrInvalidSubscriptionReference = events.ErrInvalidSubscriptionReference
	ErrInvalidTerminationTime       = events.ErrInvalidTerminationTime
	ErrInvalidTimeout               = events.ErrInvalidTimeout
	ErrInvalidVideoOutputToken      = deviceio.ErrInvalidVideoOutputToken
	ErrPullPointNotSupported        = events.ErrPullPointNotSupported
	ErrSerialPortConfigNil          = deviceio.ErrSerialPortConfigNil
	ErrVideoOutputConfigNil         = deviceio.ErrVideoOutputConfigNil
)

// Complete export surface of the service packages (v1 names).

type (
	AddAudioDecoderConfiguration                       = media.AddAudioDecoderConfiguration
	AddAudioEncoderConfiguration                       = media.AddAudioEncoderConfiguration
	AddAudioOutputConfiguration                        = media.AddAudioOutputConfiguration
	AddAudioSourceConfiguration                        = media.AddAudioSourceConfiguration
	AddMetadataConfiguration                           = media.AddMetadataConfiguration
	AddPTZConfiguration                                = media.AddPTZConfiguration
	AddVideoAnalyticsConfiguration                     = media.AddVideoAnalyticsConfiguration
	AddVideoEncoderConfiguration                       = media.AddVideoEncoderConfiguration
	AddVideoSourceConfiguration                        = media.AddVideoSourceConfiguration
	CreateOSD                                          = media.CreateOSD
	CreateOSDResponse                                  = media.CreateOSDResponse
	CreateProfile                                      = media.CreateProfile
	CreateProfileResponse                              = media.CreateProfileResponse
	DeleteOSD                                          = media.DeleteOSD
	DeleteProfile                                      = media.DeleteProfile
	DeviceIOServiceCapabilities                        = deviceio.DeviceIOServiceCapabilities
	DigitalInputConfigurationOptions                   = deviceio.DigitalInputConfigurationOptions
	FloatRectangle                                     = deviceio.FloatRectangle
	FocusMove                                          = imaging.FocusMove
	GetAudioDecoderConfiguration                       = media.GetAudioDecoderConfiguration
	GetAudioDecoderConfigurationOptions                = media.GetAudioDecoderConfigurationOptions
	GetAudioDecoderConfigurationOptionsResponse        = media.GetAudioDecoderConfigurationOptionsResponse
	GetAudioDecoderConfigurationResponse               = media.GetAudioDecoderConfigurationResponse
	GetAudioDecoderConfigurations                      = media.GetAudioDecoderConfigurations
	GetAudioDecoderConfigurationsResponse              = media.GetAudioDecoderConfigurationsResponse
	GetAudioEncoderConfiguration                       = media.GetAudioEncoderConfiguration
	GetAudioEncoderConfigurationOptions                = media.GetAudioEncoderConfigurationOptions
	GetAudioEncoderConfigurationOptionsResponse        = media.GetAudioEncoderConfigurationOptionsResponse
	GetAudioEncoderConfigurationResponse               = media.GetAudioEncoderConfigurationResponse
	GetAudioEncoderConfigurations                      = media.GetAudioEncoderConfigurations
	GetAudioEncoderConfigurationsResponse              = media.GetAudioEncoderConfigurationsResponse
	GetAudioOutputConfiguration                        = media.GetAudioOutputConfiguration
	GetAudioOutputConfigurationOptions                 = media.GetAudioOutputConfigurationOptions
	GetAudioOutputConfigurationOptionsResponse         = media.GetAudioOutputConfigurationOptionsResponse
	GetAudioOutputConfigurationResponse                = media.GetAudioOutputConfigurationResponse
	GetAudioOutputConfigurations                       = media.GetAudioOutputConfigurations
	GetAudioOutputConfigurationsResponse               = media.GetAudioOutputConfigurationsResponse
	GetAudioOutputs                                    = media.GetAudioOutputs
	GetAudioOutputsResponse                            = media.GetAudioOutputsResponse
	GetAudioSourceConfiguration                        = media.GetAudioSourceConfiguration
	GetAudioSourceConfigurationOptions                 = media.GetAudioSourceConfigurationOptions
	GetAudioSourceConfigurationOptionsResponse         = media.GetAudioSourceConfigurationOptionsResponse
	GetAudioSourceConfigurationResponse                = media.GetAudioSourceConfigurationResponse
	GetAudioSourceConfigurations                       = media.GetAudioSourceConfigurations
	GetAudioSourceConfigurationsResponse               = media.GetAudioSourceConfigurationsResponse
	GetAudioSources                                    = media.GetAudioSources
	GetAudioSourcesResponse                            = media.GetAudioSourcesResponse
	GetCompatibleAudioDecoderConfigurations            = media.GetCompatibleAudioDecoderConfigurations
	GetCompatibleAudioDecoderConfigurationsResponse    = media.GetCompatibleAudioDecoderConfigurationsResponse
	GetCompatibleAudioEncoderConfigurations            = media.GetCompatibleAudioEncoderConfigurations
	GetCompatibleAudioEncoderConfigurationsResponse    = media.GetCompatibleAudioEncoderConfigurationsResponse
	GetCompatibleAudioOutputConfigurations             = media.GetCompatibleAudioOutputConfigurations
	GetCompatibleAudioOutputConfigurationsResponse     = media.GetCompatibleAudioOutputConfigurationsResponse
	GetCompatibleAudioSourceConfigurations             = media.GetCompatibleAudioSourceConfigurations
	GetCompatibleAudioSourceConfigurationsResponse     = media.GetCompatibleAudioSourceConfigurationsResponse
	GetCompatibleMetadataConfigurations                = media.GetCompatibleMetadataConfigurations
	GetCompatibleMetadataConfigurationsResponse        = media.GetCompatibleMetadataConfigurationsResponse
	GetCompatiblePTZConfigurations                     = media.GetCompatiblePTZConfigurations
	GetCompatiblePTZConfigurationsResponse             = media.GetCompatiblePTZConfigurationsResponse
	GetCompatibleVideoAnalyticsConfigurations          = media.GetCompatibleVideoAnalyticsConfigurations
	GetCompatibleVideoAnalyticsConfigurationsResponse  = media.GetCompatibleVideoAnalyticsConfigurationsResponse
	GetCompatibleVideoEncoderConfigurations            = media.GetCompatibleVideoEncoderConfigurations
	GetCompatibleVideoEncoderConfigurationsResponse    = media.GetCompatibleVideoEncoderConfigurationsResponse
	GetCompatibleVideoSourceConfigurations             = media.GetCompatibleVideoSourceConfigurations
	GetCompatibleVideoSourceConfigurationsResponse     = media.GetCompatibleVideoSourceConfigurationsResponse
	GetGuaranteedNumberOfVideoEncoderInstances         = media.GetGuaranteedNumberOfVideoEncoderInstances
	GetGuaranteedNumberOfVideoEncoderInstancesResponse = media.GetGuaranteedNumberOfVideoEncoderInstancesResponse
	GetMetadataConfiguration                           = media.GetMetadataConfiguration
	GetMetadataConfigurationOptions                    = media.GetMetadataConfigurationOptions
	GetMetadataConfigurationOptionsResponse            = media.GetMetadataConfigurationOptionsResponse
	GetMetadataConfigurationResponse                   = media.GetMetadataConfigurationResponse
	GetMetadataConfigurations                          = media.GetMetadataConfigurations
	GetMetadataConfigurationsResponse                  = media.GetMetadataConfigurationsResponse
	GetOSD                                             = media.GetOSD
	GetOSDOptions                                      = media.GetOSDOptions
	GetOSDOptionsResponse                              = media.GetOSDOptionsResponse
	GetOSDResponse                                     = media.GetOSDResponse
	GetOSDs                                            = media.GetOSDs
	GetOSDsResponse                                    = media.GetOSDsResponse
	GetProfile                                         = media.GetProfile
	GetProfileResponse                                 = media.GetProfileResponse
	GetProfiles                                        = media.GetProfiles
	GetProfilesResponse                                = media.GetProfilesResponse
	GetServiceCapabilities                             = media.GetServiceCapabilities
	GetServiceCapabilitiesResponse                     = media.GetServiceCapabilitiesResponse
	GetSnapshotURI                                     = media.GetSnapshotURI
	GetSnapshotURIResponse                             = media.GetSnapshotURIResponse
	GetStreamURI                                       = media.GetStreamURI
	GetStreamURIResponse                               = media.GetStreamURIResponse
	GetVideoAnalyticsConfiguration                     = media.GetVideoAnalyticsConfiguration
	GetVideoAnalyticsConfigurationOptions              = media.GetVideoAnalyticsConfigurationOptions
	GetVideoAnalyticsConfigurationOptionsResponse      = media.GetVideoAnalyticsConfigurationOptionsResponse
	GetVideoAnalyticsConfigurationResponse             = media.GetVideoAnalyticsConfigurationResponse
	GetVideoAnalyticsConfigurations                    = media.GetVideoAnalyticsConfigurations
	GetVideoAnalyticsConfigurationsResponse            = media.GetVideoAnalyticsConfigurationsResponse
	GetVideoEncoderConfiguration                       = media.GetVideoEncoderConfiguration
	GetVideoEncoderConfigurationOptions                = media.GetVideoEncoderConfigurationOptions
	GetVideoEncoderConfigurationOptionsResponse        = media.GetVideoEncoderConfigurationOptionsResponse
	GetVideoEncoderConfigurationResponse               = media.GetVideoEncoderConfigurationResponse
	GetVideoEncoderConfigurations                      = media.GetVideoEncoderConfigurations
	GetVideoEncoderConfigurationsResponse              = media.GetVideoEncoderConfigurationsResponse
	GetVideoSourceConfiguration                        = media.GetVideoSourceConfiguration
	GetVideoSourceConfigurationOptions                 = media.GetVideoSourceConfigurationOptions
	GetVideoSourceConfigurationOptionsResponse         = media.GetVideoSourceConfigurationOptionsResponse
	GetVideoSourceConfigurationResponse                = media.GetVideoSourceConfigurationResponse
	GetVideoSourceConfigurations                       = media.GetVideoSourceConfigurations
	GetVideoSourceConfigurationsResponse               = media.GetVideoSourceConfigurationsResponse
	GetVideoSourceModes                                = media.GetVideoSourceModes
	GetVideoSourceModesResponse                        = media.GetVideoSourceModesResponse
	GetVideoSources                                    = media.GetVideoSources
	GetVideoSourcesResponse                            = media.GetVideoSourcesResponse
	IPv6Configuration                                  = device.IPv6Configuration
	Layout                                             = deviceio.Layout
	NetworkInterfaceConfig                             = device.NetworkInterfaceConfig
	PaneLayout                                         = deviceio.PaneLayout
	RelayOutputOptions                                 = deviceio.RelayOutputOptions
	RemoveAudioDecoderConfiguration                    = media.RemoveAudioDecoderConfiguration
	RemoveAudioEncoderConfiguration                    = media.RemoveAudioEncoderConfiguration
	RemoveAudioOutputConfiguration                     = media.RemoveAudioOutputConfiguration
	RemoveAudioSourceConfiguration                     = media.RemoveAudioSourceConfiguration
	RemoveMetadataConfiguration                        = media.RemoveMetadataConfiguration
	RemovePTZConfiguration                             = media.RemovePTZConfiguration
	RemoveVideoAnalyticsConfiguration                  = media.RemoveVideoAnalyticsConfiguration
	RemoveVideoEncoderConfiguration                    = media.RemoveVideoEncoderConfiguration
	RemoveVideoSourceConfiguration                     = media.RemoveVideoSourceConfiguration
	SerialPort                                         = deviceio.SerialPort
	SerialPortConfiguration                            = deviceio.SerialPortConfiguration
	SerialPortConfigurationOptions                     = deviceio.SerialPortConfigurationOptions
	SetAudioDecoderConfiguration                       = media.SetAudioDecoderConfiguration
	SetAudioEncoderConfiguration                       = media.SetAudioEncoderConfiguration
	SetAudioOutputConfiguration                        = media.SetAudioOutputConfiguration
	SetAudioSourceConfiguration                        = media.SetAudioSourceConfiguration
	SetMetadataConfiguration                           = media.SetMetadataConfiguration
	SetOSD                                             = media.SetOSD
	SetProfile                                         = media.SetProfile
	SetSynchronizationPoint                            = media.SetSynchronizationPoint
	SetVideoAnalyticsConfiguration                     = media.SetVideoAnalyticsConfiguration
	SetVideoEncoderConfiguration                       = media.SetVideoEncoderConfiguration
	SetVideoSourceConfiguration                        = media.SetVideoSourceConfiguration
	SetVideoSourceMode                                 = media.SetVideoSourceMode
	StartMulticastStreaming                            = media.StartMulticastStreaming
	StopMulticastStreaming                             = media.StopMulticastStreaming
	StringRange                                        = deviceio.StringRange
	VideoOutput                                        = deviceio.VideoOutput
	VideoOutputConfiguration                           = deviceio.VideoOutputConfiguration
	VideoOutputConfigurationOptions                    = deviceio.VideoOutputConfigurationOptions
)
