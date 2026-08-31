// Package onviftesting provides testing utilities for ONVIF client testing.
package onviftesting

// ONVIF token parameter names, shared between OperationSpec.RequiresToken
// entries below and the parameter extraction logic in mock_server.go.
const (
	tokenProfile       = "ProfileToken"
	tokenConfiguration = "ConfigurationToken"
	tokenVideoSource   = "VideoSourceToken"
	tokenVideoOutput   = "VideoOutputToken"
	tokenSerialPort    = "SerialPortToken"
)

// Operation category labels used across OperationSpec.Category entries.
const (
	categoryCore         = "core"
	categorySystem       = "system"
	categoryNetwork      = "network"
	categorySecurity     = "security"
	categoryCertificates = "certificates"
	categoryAdditional   = "additional"
	categoryWiFi         = "wifi"
	categoryAudio        = "audio"
	categoryAnalytics    = "analytics"
	categoryVideo        = "video"
	categoryEncoder      = "encoder"
	categoryMetadata     = "metadata"
	categoryOSD          = "osd"
	categoryOutputs      = "outputs"
	categorySerial       = "serial"
)

// ONVIF operation names referenced as both an operation's Name and as the
// DependsOn value of operations that require a token it provides.
const (
	opGetProfiles                   = "GetProfiles"
	opGetVideoSources               = "GetVideoSources"
	opGetVideoSourceConfigurations  = "GetVideoSourceConfigurations"
	opGetVideoEncoderConfigurations = "GetVideoEncoderConfigurations"
	opGetAudioSourceConfigurations  = "GetAudioSourceConfigurations"
	opGetAudioEncoderConfigurations = "GetAudioEncoderConfigurations"
	opGetAudioOutputConfigurations  = "GetAudioOutputConfigurations"
	opGetAudioDecoderConfigurations = "GetAudioDecoderConfigurations"
	opGetMetadataConfigurations     = "GetMetadataConfigurations"
	opGetOSDs                       = "GetOSDs"
	opGetVideoOutputs               = "GetVideoOutputs"
	opGetSerialPorts                = "GetSerialPorts"
	opGetStreamURI                  = "GetStreamURI"
	opGetDeviceInformation          = "GetDeviceInformation"
)

// OperationSpec defines how to capture an ONVIF operation.
type OperationSpec struct {
	// Name is the ONVIF operation name (e.g., opGetDeviceInformation)
	Name string

	// Service is the ONVIF service type
	Service ServiceType

	// RequiresInit indicates if Initialize() must be called first
	RequiresInit bool

	// RequiresToken specifies which token parameter is needed (e.g., tokenProfile)
	RequiresToken string

	// DependsOn specifies which operation provides the required token
	DependsOn string

	// Category groups related operations (e.g., categoryCore, categoryNetwork, categorySecurity)
	Category string

	// IsWrite indicates if this operation modifies camera state
	IsWrite bool

	// Description provides a brief description of the operation
	Description string
}

// =============================================================================
// Device Service Operations (97 total, ~35 READ operations)
// =============================================================================

// DeviceReadOperations contains all read-only Device service operations.
var DeviceReadOperations = []OperationSpec{
	// Core operations
	{Name: opGetDeviceInformation, Service: ServiceDevice, Category: categoryCore,
		Description: "Get manufacturer, model, firmware version"},
	{Name: "GetCapabilities", Service: ServiceDevice, Category: categoryCore,
		Description: "Get service capabilities and endpoints"},
	{Name: "GetServices", Service: ServiceDevice, Category: categoryCore,
		Description: "Get list of available services"},
	{Name: "GetServiceCapabilities", Service: ServiceDevice, Category: categoryCore,
		Description: "Get device service capabilities"},

	// System operations
	{Name: "GetSystemDateAndTime", Service: ServiceDevice, Category: categorySystem,
		Description: "Get device date and time settings"},
	{Name: "GetSystemLog", Service: ServiceDevice, Category: categorySystem,
		Description: "Get system log"},
	{Name: "GetSystemUris", Service: ServiceDevice, Category: categorySystem,
		Description: "Get system URIs (support, firmware, logs)"},
	{Name: "GetSystemSupportInformation", Service: ServiceDevice, Category: categorySystem,
		Description: "Get system support information"},
	{Name: "GetEndpointReference", Service: ServiceDevice, Category: categorySystem,
		Description: "Get unique endpoint reference address"},

	// Network operations
	{Name: "GetHostname", Service: ServiceDevice, Category: categoryNetwork,
		Description: "Get device hostname"},
	{Name: "GetDNS", Service: ServiceDevice, Category: categoryNetwork,
		Description: "Get DNS configuration"},
	{Name: "GetNTP", Service: ServiceDevice, Category: categoryNetwork,
		Description: "Get NTP configuration"},
	{Name: "GetNetworkInterfaces", Service: ServiceDevice, Category: categoryNetwork,
		Description: "Get network interface configuration"},
	{Name: "GetNetworkProtocols", Service: ServiceDevice, Category: categoryNetwork,
		Description: "Get enabled network protocols"},
	{Name: "GetNetworkDefaultGateway", Service: ServiceDevice, Category: categoryNetwork,
		Description: "Get default gateway configuration"},

	// Discovery operations
	{Name: "GetDiscoveryMode", Service: ServiceDevice, Category: "discovery",
		Description: "Get WS-Discovery mode"},
	{Name: "GetRemoteDiscoveryMode", Service: ServiceDevice, Category: "discovery",
		Description: "Get remote discovery mode"},

	// Scope operations
	{Name: "GetScopes", Service: ServiceDevice, Category: "scopes",
		Description: "Get device scopes for discovery"},

	// User operations
	{Name: "GetUsers", Service: ServiceDevice, Category: "users",
		Description: "Get list of device users"},

	// Security operations
	{Name: "GetRemoteUser", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get remote user configuration"},
	{Name: "GetIPAddressFilter", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get IP address filter rules"},
	{Name: "GetZeroConfiguration", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get zero configuration (link-local) settings"},
	{Name: "GetDynamicDNS", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get dynamic DNS configuration"},
	{Name: "GetAccessPolicy", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get access policy configuration"},
	{Name: "GetPasswordComplexityConfiguration", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get password complexity requirements"},
	{Name: "GetPasswordHistoryConfiguration", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get password history configuration"},
	{Name: "GetAuthFailureWarningConfiguration", Service: ServiceDevice, Category: categorySecurity,
		Description: "Get authentication failure warning settings"},

	// Certificate operations
	{Name: "GetCertificates", Service: ServiceDevice, Category: categoryCertificates,
		Description: "Get device certificates"},
	{Name: "GetCACertificates", Service: ServiceDevice, Category: categoryCertificates,
		Description: "Get CA certificates"},
	{Name: "GetCertificatesStatus", Service: ServiceDevice, Category: categoryCertificates,
		Description: "Get certificate status"},
	{Name: "GetClientCertificateMode", Service: ServiceDevice, Category: categoryCertificates,
		Description: "Get client certificate mode"},

	// Storage operations
	{Name: "GetStorageConfigurations", Service: ServiceDevice, Category: "storage",
		Description: "Get storage configurations"},

	// Relay operations
	{Name: "GetRelayOutputs", Service: ServiceDevice, Category: "relay",
		Description: "Get relay output states"},

	// Additional operations
	{Name: "GetGeoLocation", Service: ServiceDevice, Category: categoryAdditional,
		Description: "Get geographic location"},
	{Name: "GetDPAddresses", Service: ServiceDevice, Category: categoryAdditional,
		Description: "Get DP (discovery proxy) addresses"},
	{Name: "GetWsdlURL", Service: ServiceDevice, Category: categoryAdditional,
		Description: "Get WSDL URL"},

	// WiFi operations (802.11)
	{Name: "GetDot11Capabilities", Service: ServiceDevice, Category: categoryWiFi,
		Description: "Get 802.11 capabilities"},
	{Name: "GetDot11Status", Service: ServiceDevice, Category: categoryWiFi,
		Description: "Get 802.11 connection status"},
	{Name: "GetDot1XConfigurations", Service: ServiceDevice, Category: categoryWiFi,
		Description: "Get 802.1X configurations"},
	{Name: "ScanAvailableDot11Networks", Service: ServiceDevice, Category: categoryWiFi,
		Description: "Scan for available WiFi networks"},
}

// =============================================================================
// Media Service Operations (82 total, ~45 READ operations)
// =============================================================================

// MediaReadOperations contains all read-only Media service operations.
var MediaReadOperations = []OperationSpec{
	// Service capabilities
	{Name: "GetMediaServiceCapabilities", Service: ServiceMedia, RequiresInit: true, Category: categoryCore,
		Description: "Get media service capabilities"},

	// Profile operations
	{Name: opGetProfiles, Service: ServiceMedia, RequiresInit: true, Category: "profiles",
		Description: "Get all media profiles"},
	{Name: "GetProfile", Service: ServiceMedia, RequiresInit: true, Category: "profiles",
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get specific profile by token"},

	// Video source operations
	{Name: opGetVideoSources, Service: ServiceMedia, RequiresInit: true, Category: categoryVideo,
		Description: "Get video sources"},
	{Name: opGetVideoSourceConfigurations, Service: ServiceMedia, RequiresInit: true, Category: categoryVideo,
		Description: "Get all video source configurations"},
	{Name: "GetVideoSourceConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryVideo,
		RequiresToken: tokenConfiguration, DependsOn: opGetVideoSourceConfigurations,
		Description: "Get specific video source configuration"},
	{Name: "GetVideoSourceConfigurationOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryVideo,
		RequiresToken: tokenConfiguration, DependsOn: opGetVideoSourceConfigurations,
		Description: "Get video source configuration options"},
	{Name: "GetVideoSourceModes", Service: ServiceMedia, RequiresInit: true, Category: categoryVideo,
		RequiresToken: tokenVideoSource, DependsOn: opGetVideoSources,
		Description: "Get video source modes"},
	{Name: "GetCompatibleVideoSourceConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryVideo,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible video source configurations for profile"},

	// Video encoder operations
	{Name: opGetVideoEncoderConfigurations, Service: ServiceMedia, RequiresInit: true, Category: categoryEncoder,
		Description: "Get all video encoder configurations"},
	{Name: "GetVideoEncoderConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryEncoder,
		RequiresToken: tokenConfiguration, DependsOn: opGetVideoEncoderConfigurations,
		Description: "Get specific video encoder configuration"},
	{Name: "GetVideoEncoderConfigurationOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryEncoder,
		RequiresToken: tokenConfiguration, DependsOn: opGetVideoEncoderConfigurations,
		Description: "Get video encoder configuration options"},
	{Name: "GetCompatibleVideoEncoderConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryEncoder,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible video encoder configurations for profile"},
	{Name: "GetGuaranteedNumberOfVideoEncoderInstances", Service: ServiceMedia, RequiresInit: true, Category: categoryEncoder,
		RequiresToken: tokenConfiguration, DependsOn: opGetVideoEncoderConfigurations,
		Description: "Get guaranteed number of video encoder instances"},

	// Audio source operations
	{Name: "GetAudioSources", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		Description: "Get audio sources"},
	{Name: opGetAudioSourceConfigurations, Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		Description: "Get all audio source configurations"},
	{Name: "GetAudioSourceConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioSourceConfigurations,
		Description: "Get specific audio source configuration"},
	{Name: "GetAudioSourceConfigurationOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioSourceConfigurations,
		Description: "Get audio source configuration options"},
	{Name: "GetCompatibleAudioSourceConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible audio source configurations for profile"},

	// Audio encoder operations
	{Name: opGetAudioEncoderConfigurations, Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		Description: "Get all audio encoder configurations"},
	{Name: "GetAudioEncoderConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioEncoderConfigurations,
		Description: "Get specific audio encoder configuration"},
	{Name: "GetAudioEncoderConfigurationOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioEncoderConfigurations,
		Description: "Get audio encoder configuration options"},
	{Name: "GetCompatibleAudioEncoderConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible audio encoder configurations for profile"},

	// Audio output operations
	{Name: "GetAudioOutputs", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		Description: "Get audio outputs"},
	{Name: opGetAudioOutputConfigurations, Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		Description: "Get all audio output configurations"},
	{Name: "GetAudioOutputConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioOutputConfigurations,
		Description: "Get specific audio output configuration"},
	{Name: "GetAudioOutputConfigurationOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioOutputConfigurations,
		Description: "Get audio output configuration options"},
	{Name: "GetCompatibleAudioOutputConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible audio output configurations for profile"},

	// Audio decoder operations
	{Name: opGetAudioDecoderConfigurations, Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		Description: "Get all audio decoder configurations"},
	{Name: "GetAudioDecoderConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioDecoderConfigurations,
		Description: "Get specific audio decoder configuration"},
	{Name: "GetAudioDecoderConfigurationOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenConfiguration, DependsOn: opGetAudioDecoderConfigurations,
		Description: "Get audio decoder configuration options"},
	{Name: "GetCompatibleAudioDecoderConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryAudio,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible audio decoder configurations for profile"},

	// Metadata operations
	{Name: opGetMetadataConfigurations, Service: ServiceMedia, RequiresInit: true, Category: categoryMetadata,
		Description: "Get all metadata configurations"},
	{Name: "GetMetadataConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryMetadata,
		RequiresToken: tokenConfiguration, DependsOn: opGetMetadataConfigurations,
		Description: "Get specific metadata configuration"},
	{Name: "GetMetadataConfigurationOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryMetadata,
		RequiresToken: tokenConfiguration, DependsOn: opGetMetadataConfigurations,
		Description: "Get metadata configuration options"},
	{Name: "GetCompatibleMetadataConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryMetadata,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible metadata configurations for profile"},

	// Video analytics operations
	{Name: "GetVideoAnalyticsConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryAnalytics,
		Description: "Get all video analytics configurations"},
	{Name: "GetVideoAnalyticsConfiguration", Service: ServiceMedia, RequiresInit: true, Category: categoryAnalytics,
		RequiresToken: tokenConfiguration, DependsOn: "GetVideoAnalyticsConfigurations",
		Description: "Get specific video analytics configuration"},
	{Name: "GetCompatibleVideoAnalyticsConfigurations", Service: ServiceMedia, RequiresInit: true, Category: categoryAnalytics,
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible video analytics configurations for profile"},

	// Stream operations
	{Name: opGetStreamURI, Service: ServiceMedia, RequiresInit: true, Category: "streaming",
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get RTSP stream URI"},
	{Name: "GetSnapshotURI", Service: ServiceMedia, RequiresInit: true, Category: "streaming",
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get snapshot URI"},

	// OSD operations
	{Name: opGetOSDs, Service: ServiceMedia, RequiresInit: true, Category: categoryOSD,
		Description: "Get all OSD configurations"},
	{Name: "GetOSD", Service: ServiceMedia, RequiresInit: true, Category: categoryOSD,
		RequiresToken: tokenConfiguration, DependsOn: opGetOSDs,
		Description: "Get specific OSD configuration"},
	{Name: "GetOSDOptions", Service: ServiceMedia, RequiresInit: true, Category: categoryOSD,
		RequiresToken: tokenConfiguration, DependsOn: opGetOSDs,
		Description: "Get OSD configuration options"},

	// PTZ configuration operations (on Media service)
	{Name: "GetCompatiblePTZConfigurations", Service: ServiceMedia, RequiresInit: true, Category: "ptz",
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get compatible PTZ configurations for profile"},
}

// =============================================================================
// PTZ Service Operations (13 total, ~5 READ operations)
// =============================================================================

// PTZReadOperations contains all read-only PTZ service operations.
var PTZReadOperations = []OperationSpec{
	{Name: "GetConfigurations", Service: ServicePTZ, RequiresInit: true, Category: "config",
		Description: "Get all PTZ configurations"},
	{Name: "GetConfiguration", Service: ServicePTZ, RequiresInit: true, Category: "config",
		RequiresToken: "PTZConfigurationToken", DependsOn: "GetConfigurations",
		Description: "Get specific PTZ configuration"},
	{Name: "GetStatus", Service: ServicePTZ, RequiresInit: true, Category: "status",
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get PTZ status (position, move status)"},
	{Name: "GetPresets", Service: ServicePTZ, RequiresInit: true, Category: "presets",
		RequiresToken: tokenProfile, DependsOn: opGetProfiles,
		Description: "Get PTZ presets"},
	{Name: "GetNodes", Service: ServicePTZ, RequiresInit: true, Category: "nodes",
		Description: "Get PTZ nodes"},
	{Name: "GetNode", Service: ServicePTZ, RequiresInit: true, Category: "nodes",
		RequiresToken: "NodeToken", DependsOn: "GetNodes",
		Description: "Get specific PTZ node"},
}

// =============================================================================
// Imaging Service Operations (7 total, ~4 READ operations)
// =============================================================================

// ImagingReadOperations contains all read-only Imaging service operations.
var ImagingReadOperations = []OperationSpec{
	{Name: "GetImagingSettings", Service: ServiceImaging, RequiresInit: true, Category: "settings",
		RequiresToken: tokenVideoSource, DependsOn: opGetVideoSources,
		Description: "Get imaging settings (brightness, contrast, etc.)"},
	{Name: "GetOptions", Service: ServiceImaging, RequiresInit: true, Category: "options",
		RequiresToken: tokenVideoSource, DependsOn: opGetVideoSources,
		Description: "Get imaging options and ranges"},
	{Name: "GetMoveOptions", Service: ServiceImaging, RequiresInit: true, Category: "options",
		RequiresToken: tokenVideoSource, DependsOn: opGetVideoSources,
		Description: "Get focus move options"},
	{Name: "GetImagingStatus", Service: ServiceImaging, RequiresInit: true, Category: "status",
		RequiresToken: tokenVideoSource, DependsOn: opGetVideoSources,
		Description: "Get imaging status (focus status, etc.)"},
}

// =============================================================================
// Event Service Operations (12 total, ~3 READ operations)
// =============================================================================

// EventReadOperations contains all read-only Event service operations.
var EventReadOperations = []OperationSpec{
	{Name: "GetEventServiceCapabilities", Service: ServiceEvent, RequiresInit: true, Category: categoryCore,
		Description: "Get event service capabilities"},
	{Name: "GetEventProperties", Service: ServiceEvent, RequiresInit: true, Category: categoryCore,
		Description: "Get event topic properties"},
	{Name: "GetEventBrokers", Service: ServiceEvent, RequiresInit: true, Category: "brokers",
		Description: "Get event brokers"},
}

// =============================================================================
// DeviceIO Service Operations (14 total, ~11 READ operations)
// =============================================================================

// DeviceIOReadOperations contains all read-only DeviceIO service operations.
var DeviceIOReadOperations = []OperationSpec{
	{Name: "GetDeviceIOServiceCapabilities", Service: ServiceDeviceIO, RequiresInit: true, Category: categoryCore,
		Description: "Get DeviceIO service capabilities"},
	{Name: "GetDigitalInputs", Service: ServiceDeviceIO, RequiresInit: true, Category: "inputs",
		Description: "Get digital inputs"},
	{Name: "GetDigitalInputConfigurationOptions", Service: ServiceDeviceIO, RequiresInit: true, Category: "inputs",
		Description: "Get digital input configuration options"},
	{Name: opGetVideoOutputs, Service: ServiceDeviceIO, RequiresInit: true, Category: categoryOutputs,
		Description: "Get video outputs"},
	{Name: "GetVideoOutputConfiguration", Service: ServiceDeviceIO, RequiresInit: true, Category: categoryOutputs,
		RequiresToken: tokenVideoOutput, DependsOn: opGetVideoOutputs,
		Description: "Get video output configuration"},
	{Name: "GetVideoOutputConfigurationOptions", Service: ServiceDeviceIO, RequiresInit: true, Category: categoryOutputs,
		RequiresToken: tokenVideoOutput, DependsOn: opGetVideoOutputs,
		Description: "Get video output configuration options"},
	{Name: opGetSerialPorts, Service: ServiceDeviceIO, RequiresInit: true, Category: categorySerial,
		Description: "Get serial ports"},
	{Name: "GetSerialPortConfiguration", Service: ServiceDeviceIO, RequiresInit: true, Category: categorySerial,
		RequiresToken: tokenSerialPort, DependsOn: opGetSerialPorts,
		Description: "Get serial port configuration"},
	{Name: "GetSerialPortConfigurationOptions", Service: ServiceDeviceIO, RequiresInit: true, Category: categorySerial,
		RequiresToken: tokenSerialPort, DependsOn: opGetSerialPorts,
		Description: "Get serial port configuration options"},
	{Name: "GetRelayOutputOptions", Service: ServiceDeviceIO, RequiresInit: true, Category: "relay",
		RequiresToken: "RelayOutputToken",
		Description:   "Get relay output options"},
	{Name: "GetAudioOutputs", Service: ServiceDeviceIO, RequiresInit: true, Category: categoryAudio,
		Description: "Get audio outputs (DeviceIO)"},
}

// =============================================================================
// Aggregation Functions
// =============================================================================

// AllReadOperations returns all READ operations across all services.
func AllReadOperations() []OperationSpec {
	all := make([]OperationSpec, 0, len(DeviceReadOperations)+len(MediaReadOperations)+len(PTZReadOperations)+len(ImagingReadOperations)+len(EventReadOperations)+len(DeviceIOReadOperations))
	all = append(all, DeviceReadOperations...)
	all = append(all, MediaReadOperations...)
	all = append(all, PTZReadOperations...)
	all = append(all, ImagingReadOperations...)
	all = append(all, EventReadOperations...)
	all = append(all, DeviceIOReadOperations...)

	return all
}

// ReadOperationsByService returns READ operations for a specific service.
func ReadOperationsByService(service ServiceType) []OperationSpec {
	switch service {
	case ServiceDevice:
		return DeviceReadOperations
	case ServiceMedia:
		return MediaReadOperations
	case ServicePTZ:
		return PTZReadOperations
	case ServiceImaging:
		return ImagingReadOperations
	case ServiceEvent:
		return EventReadOperations
	case ServiceDeviceIO:
		return DeviceIOReadOperations
	case ServiceUnknown:
		return nil
	}

	return nil
}

// IndependentOperations returns operations that don't depend on other operations.
func IndependentOperations() []OperationSpec {
	var independent []OperationSpec
	for _, op := range AllReadOperations() {
		if op.DependsOn == "" {
			independent = append(independent, op)
		}
	}

	return independent
}

// DependentOperations returns operations that depend on other operations.
func DependentOperations() []OperationSpec {
	var dependent []OperationSpec
	for _, op := range AllReadOperations() {
		if op.DependsOn != "" {
			dependent = append(dependent, op)
		}
	}

	return dependent
}

// OperationsByDependency returns operations grouped by their dependency.
func OperationsByDependency(dependsOn string) []OperationSpec {
	var ops []OperationSpec
	for _, op := range AllReadOperations() {
		if op.DependsOn == dependsOn {
			ops = append(ops, op)
		}
	}

	return ops
}

// GetOperationSpec finds an operation by name.
func GetOperationSpec(name string) *OperationSpec {
	for i := range DeviceReadOperations {
		if DeviceReadOperations[i].Name == name {
			return &DeviceReadOperations[i]
		}
	}
	for i := range MediaReadOperations {
		if MediaReadOperations[i].Name == name {
			return &MediaReadOperations[i]
		}
	}
	for i := range PTZReadOperations {
		if PTZReadOperations[i].Name == name {
			return &PTZReadOperations[i]
		}
	}
	for i := range ImagingReadOperations {
		if ImagingReadOperations[i].Name == name {
			return &ImagingReadOperations[i]
		}
	}
	for i := range EventReadOperations {
		if EventReadOperations[i].Name == name {
			return &EventReadOperations[i]
		}
	}
	for i := range DeviceIOReadOperations {
		if DeviceIOReadOperations[i].Name == name {
			return &DeviceIOReadOperations[i]
		}
	}

	return nil
}

// OperationCount returns the count of operations by service.
type OperationCount struct {
	Device   int
	Media    int
	PTZ      int
	Imaging  int
	Event    int
	DeviceIO int
	Total    int
}

// GetOperationCount returns the count of READ operations.
func GetOperationCount() OperationCount {
	return OperationCount{
		Device:   len(DeviceReadOperations),
		Media:    len(MediaReadOperations),
		PTZ:      len(PTZReadOperations),
		Imaging:  len(ImagingReadOperations),
		Event:    len(EventReadOperations),
		DeviceIO: len(DeviceIOReadOperations),
		Total:    len(AllReadOperations()),
	}
}
