package server

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/0x524a/onvif-go/server/soap"
)

// Device service SOAP message types

// GetDeviceInformationResponse represents GetDeviceInformation response
type GetDeviceInformationResponse struct {
	XMLName         xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetDeviceInformationResponse"`
	Manufacturer    string   `xml:"Manufacturer"`
	Model           string   `xml:"Model"`
	FirmwareVersion string   `xml:"FirmwareVersion"`
	SerialNumber    string   `xml:"SerialNumber"`
	HardwareId      string   `xml:"HardwareId"`
}

// GetCapabilitiesResponse represents GetCapabilities response
type GetCapabilitiesResponse struct {
	XMLName      xml.Name      `xml:"http://www.onvif.org/ver10/device/wsdl GetCapabilitiesResponse"`
	Capabilities *Capabilities `xml:"Capabilities"`
}

// Capabilities represents device capabilities
type Capabilities struct {
	Analytics *AnalyticsCapabilities `xml:"Analytics,omitempty"`
	Device    *DeviceCapabilities    `xml:"Device"`
	Events    *EventCapabilities     `xml:"Events,omitempty"`
	Imaging   *ImagingCapabilities   `xml:"Imaging,omitempty"`
	Media     *MediaCapabilities     `xml:"Media"`
	PTZ       *PTZCapabilities       `xml:"PTZ,omitempty"`
}

// AnalyticsCapabilities represents analytics service capabilities
type AnalyticsCapabilities struct {
	XAddr                  string `xml:"XAddr"`
	RuleSupport            bool   `xml:"RuleSupport,attr"`
	AnalyticsModuleSupport bool   `xml:"AnalyticsModuleSupport,attr"`
}

// DeviceCapabilities represents device service capabilities
type DeviceCapabilities struct {
	XAddr   string                `xml:"XAddr"`
	Network *NetworkCapabilities  `xml:"Network,omitempty"`
	System  *SystemCapabilities   `xml:"System,omitempty"`
	IO      *IOCapabilities       `xml:"IO,omitempty"`
	Security *SecurityCapabilities `xml:"Security,omitempty"`
}

// NetworkCapabilities represents network capabilities
type NetworkCapabilities struct {
	IPFilter          bool `xml:"IPFilter,attr"`
	ZeroConfiguration bool `xml:"ZeroConfiguration,attr"`
	IPVersion6        bool `xml:"IPVersion6,attr"`
	DynDNS            bool `xml:"DynDNS,attr"`
}

// SystemCapabilities represents system capabilities
type SystemCapabilities struct {
	DiscoveryResolve  bool `xml:"DiscoveryResolve,attr"`
	DiscoveryBye      bool `xml:"DiscoveryBye,attr"`
	RemoteDiscovery   bool `xml:"RemoteDiscovery,attr"`
	SystemBackup      bool `xml:"SystemBackup,attr"`
	SystemLogging     bool `xml:"SystemLogging,attr"`
	FirmwareUpgrade   bool `xml:"FirmwareUpgrade,attr"`
}

// IOCapabilities represents I/O capabilities
type IOCapabilities struct {
	InputConnectors int `xml:"InputConnectors,attr"`
	RelayOutputs    int `xml:"RelayOutputs,attr"`
}

// SecurityCapabilities represents security capabilities
type SecurityCapabilities struct {
	TLS11                bool `xml:"TLS1.1,attr"`
	TLS12                bool `xml:"TLS1.2,attr"`
	OnboardKeyGeneration bool `xml:"OnboardKeyGeneration,attr"`
	AccessPolicyConfig   bool `xml:"AccessPolicyConfig,attr"`
	X509Token            bool `xml:"X.509Token,attr"`
	SAMLToken            bool `xml:"SAMLToken,attr"`
	KerberosToken        bool `xml:"KerberosToken,attr"`
	RELToken             bool `xml:"RELToken,attr"`
}

// EventCapabilities represents event service capabilities
type EventCapabilities struct {
	XAddr                         string `xml:"XAddr"`
	WSSubscriptionPolicySupport   bool   `xml:"WSSubscriptionPolicySupport,attr"`
	WSPullPointSupport            bool   `xml:"WSPullPointSupport,attr"`
	WSPausableSubscriptionSupport bool   `xml:"WSPausableSubscriptionManagerInterfaceSupport,attr"`
}

// ImagingCapabilities represents imaging service capabilities
type ImagingCapabilities struct {
	XAddr string `xml:"XAddr"`
}

// MediaCapabilities represents media service capabilities
type MediaCapabilities struct {
	XAddr                 string                 `xml:"XAddr"`
	StreamingCapabilities *StreamingCapabilities `xml:"StreamingCapabilities"`
}

// StreamingCapabilities represents streaming capabilities
type StreamingCapabilities struct {
	RTPMulticast bool `xml:"RTPMulticast,attr"`
	RTP_TCP      bool `xml:"RTP_TCP,attr"`
	RTP_RTSP_TCP bool `xml:"RTP_RTSP_TCP,attr"`
}

// PTZCapabilities represents PTZ service capabilities
type PTZCapabilities struct {
	XAddr string `xml:"XAddr"`
}

// GetServicesResponse represents GetServices response
type GetServicesResponse struct {
	XMLName xml.Name  `xml:"http://www.onvif.org/ver10/device/wsdl GetServicesResponse"`
	Service []Service `xml:"Service"`
}

// Service represents a service
type Service struct {
	Namespace string `xml:"Namespace"`
	XAddr     string `xml:"XAddr"`
	Version   Version `xml:"Version"`
}

// Version represents service version
type Version struct {
	Major int `xml:"Major"`
	Minor int `xml:"Minor"`
}

// SystemRebootResponse represents SystemReboot response
type SystemRebootResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl SystemRebootResponse"`
	Message string   `xml:"Message"`
}

// Device service handlers

// HandleGetDeviceInformation handles GetDeviceInformation request
func (s *Server) HandleGetDeviceInformation(body interface{}) (interface{}, error) {
	return &GetDeviceInformationResponse{
		Manufacturer:    s.config.DeviceInfo.Manufacturer,
		Model:           s.config.DeviceInfo.Model,
		FirmwareVersion: s.config.DeviceInfo.FirmwareVersion,
		SerialNumber:    s.config.DeviceInfo.SerialNumber,
		HardwareId:      s.config.DeviceInfo.HardwareID,
	}, nil
}

// HandleGetCapabilities handles GetCapabilities request
func (s *Server) HandleGetCapabilities(body interface{}) (interface{}, error) {
	// Get the host from the request (in a real implementation)
	// For now, use a placeholder
	host := s.config.Host
	if host == "0.0.0.0" || host == "" {
		host = "localhost"
	}

	baseURL := fmt.Sprintf("http://%s:%d%s", host, s.config.Port, s.config.BasePath)

	capabilities := &Capabilities{
		Device: &DeviceCapabilities{
			XAddr: baseURL + "/device_service",
			Network: &NetworkCapabilities{
				IPFilter:          false,
				ZeroConfiguration: false,
				IPVersion6:        false,
				DynDNS:            false,
			},
			System: &SystemCapabilities{
				DiscoveryResolve: true,
				DiscoveryBye:     true,
				RemoteDiscovery:  true,
				SystemBackup:     false,
				SystemLogging:    false,
				FirmwareUpgrade:  false,
			},
			IO: &IOCapabilities{
				InputConnectors: 0,
				RelayOutputs:    0,
			},
			Security: &SecurityCapabilities{
				TLS11:                false,
				TLS12:                false,
				OnboardKeyGeneration: false,
				AccessPolicyConfig:   false,
				X509Token:            false,
				SAMLToken:            false,
				KerberosToken:        false,
				RELToken:             false,
			},
		},
		Media: &MediaCapabilities{
			XAddr: baseURL + "/media_service",
			StreamingCapabilities: &StreamingCapabilities{
				RTPMulticast: false,
				RTP_TCP:      true,
				RTP_RTSP_TCP: true,
			},
		},
	}

	if s.config.SupportPTZ {
		capabilities.PTZ = &PTZCapabilities{
			XAddr: baseURL + "/ptz_service",
		}
	}

	if s.config.SupportImaging {
		capabilities.Imaging = &ImagingCapabilities{
			XAddr: baseURL + "/imaging_service",
		}
	}

	if s.config.SupportEvents {
		capabilities.Events = &EventCapabilities{
			XAddr:                         baseURL + "/events_service",
			WSSubscriptionPolicySupport:   false,
			WSPullPointSupport:            false,
			WSPausableSubscriptionSupport: false,
		}
	}

	return &GetCapabilitiesResponse{
		Capabilities: capabilities,
	}, nil
}

// HandleGetSystemDateAndTime handles GetSystemDateAndTime request
func (s *Server) HandleGetSystemDateAndTime(body interface{}) (interface{}, error) {
	now := time.Now().UTC()

	return &soap.GetSystemDateAndTimeResponse{
		SystemDateAndTime: soap.SystemDateAndTime{
			DateTimeType:    "NTP",
			DaylightSavings: false,
			TimeZone: soap.TimeZone{
				TZ: "UTC",
			},
			UTCDateTime:   soap.ToDateTime(now),
			LocalDateTime: soap.ToDateTime(now.Local()),
		},
	}, nil
}

// HandleGetServices handles GetServices request
func (s *Server) HandleGetServices(body interface{}) (interface{}, error) {
	host := s.config.Host
	if host == "0.0.0.0" || host == "" {
		host = "localhost"
	}

	baseURL := fmt.Sprintf("http://%s:%d%s", host, s.config.Port, s.config.BasePath)

	services := []Service{
		{
			Namespace: "http://www.onvif.org/ver10/device/wsdl",
			XAddr:     baseURL + "/device_service",
			Version:   Version{Major: 2, Minor: 5},
		},
		{
			Namespace: "http://www.onvif.org/ver10/media/wsdl",
			XAddr:     baseURL + "/media_service",
			Version:   Version{Major: 2, Minor: 5},
		},
	}

	if s.config.SupportPTZ {
		services = append(services, Service{
			Namespace: "http://www.onvif.org/ver20/ptz/wsdl",
			XAddr:     baseURL + "/ptz_service",
			Version:   Version{Major: 2, Minor: 5},
		})
	}

	if s.config.SupportImaging {
		services = append(services, Service{
			Namespace: "http://www.onvif.org/ver20/imaging/wsdl",
			XAddr:     baseURL + "/imaging_service",
			Version:   Version{Major: 2, Minor: 5},
		})
	}

	return &GetServicesResponse{
		Service: services,
	}, nil
}

// HandleSystemReboot handles SystemReboot request
func (s *Server) HandleSystemReboot(body interface{}) (interface{}, error) {
	return &SystemRebootResponse{
		Message: "Device rebooting",
	}, nil
}
