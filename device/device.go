// Device service operations: identity, capabilities, scopes, users,
// geo/access policy, miscellaneous tds operations.

package device

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Device service namespace.
const Namespace = "http://www.onvif.org/ver10/device/wsdl"

// GetDeviceInformation retrieves device information.
func (s *Service) GetDeviceInformation(ctx context.Context) (*DeviceInformation, error) {
	type GetDeviceInformation struct {
		XMLName xml.Name `xml:"tds:GetDeviceInformation"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDeviceInformationResponse struct {
		XMLName         xml.Name `xml:"GetDeviceInformationResponse"`
		Manufacturer    string   `xml:"Manufacturer"`
		Model           string   `xml:"Model"`
		FirmwareVersion string   `xml:"FirmwareVersion"`
		SerialNumber    string   `xml:"SerialNumber"`
		HardwareID      string   `xml:"HardwareId"`
	}

	req := GetDeviceInformation{
		Xmlns: Namespace,
	}

	var resp GetDeviceInformationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDeviceInformation failed: %w", err)
	}

	return &DeviceInformation{
		Manufacturer:    resp.Manufacturer,
		Model:           resp.Model,
		FirmwareVersion: resp.FirmwareVersion,
		SerialNumber:    resp.SerialNumber,
		HardwareID:      resp.HardwareID,
	}, nil
}

// GetCapabilities retrieves device capabilities.
//
//nolint:funlen // GetCapabilities has many statements due to parsing multiple service capabilities
func (s *Service) GetCapabilities(ctx context.Context) (*Capabilities, error) {
	type GetCapabilities struct {
		XMLName  xml.Name `xml:"tds:GetCapabilities"`
		Xmlns    string   `xml:"xmlns:tds,attr"`
		Category []string `xml:"tds:Category,omitempty"`
	}

	type GetCapabilitiesResponse struct {
		XMLName      xml.Name `xml:"GetCapabilitiesResponse"`
		Capabilities struct {
			Analytics *struct {
				XAddr                  string `xml:"XAddr"`
				RuleSupport            bool   `xml:"RuleSupport"`
				AnalyticsModuleSupport bool   `xml:"AnalyticsModuleSupport"`
			} `xml:"Analytics"`
			Device *struct {
				XAddr   string `xml:"XAddr"`
				Network *struct {
					IPFilter          bool `xml:"IPFilter"`
					ZeroConfiguration bool `xml:"ZeroConfiguration"`
					IPVersion6        bool `xml:"IPVersion6"`
					DynDNS            bool `xml:"DynDNS"`
				} `xml:"Network"`
				System *struct {
					DiscoveryResolve  bool     `xml:"DiscoveryResolve"`
					DiscoveryBye      bool     `xml:"DiscoveryBye"`
					RemoteDiscovery   bool     `xml:"RemoteDiscovery"`
					SystemBackup      bool     `xml:"SystemBackup"`
					SystemLogging     bool     `xml:"SystemLogging"`
					FirmwareUpgrade   bool     `xml:"FirmwareUpgrade"`
					SupportedVersions []string `xml:"SupportedVersions>Major"`
				} `xml:"System"`
				IO *struct {
					InputConnectors int `xml:"InputConnectors"`
					RelayOutputs    int `xml:"RelayOutputs"`
				} `xml:"IO"`
				Security *struct {
					TLS11                bool `xml:"TLS1.1"`
					TLS12                bool `xml:"TLS1.2"`
					OnboardKeyGeneration bool `xml:"OnboardKeyGeneration"`
					AccessPolicyConfig   bool `xml:"AccessPolicyConfig"`
					X509Token            bool `xml:"X.509Token"`
					SAMLToken            bool `xml:"SAMLToken"`
					KerberosToken        bool `xml:"KerberosToken"`
					RELToken             bool `xml:"RELToken"`
				} `xml:"Security"`
			} `xml:"Device"`
			Events *struct {
				XAddr                         string `xml:"XAddr"`
				WSSubscriptionPolicySupport   bool   `xml:"WSSubscriptionPolicySupport"`
				WSPullPointSupport            bool   `xml:"WSPullPointSupport"`
				WSPausableSubscriptionSupport bool   `xml:"WSPausableSubscriptionManagerInterfaceSupport"`
			} `xml:"Events"`
			Imaging *struct {
				XAddr string `xml:"XAddr"`
			} `xml:"Imaging"`
			Media *struct {
				XAddr                 string `xml:"XAddr"`
				StreamingCapabilities *struct {
					RTPMulticast bool `xml:"RTPMulticast"`
					RTPTCP       bool `xml:"RTP_TCP"`
					RTPRTSPTCP   bool `xml:"RTP_RTSP_TCP"`
				} `xml:"StreamingCapabilities"`
			} `xml:"Media"`
			PTZ *struct {
				XAddr string `xml:"XAddr"`
			} `xml:"PTZ"`
		} `xml:"Capabilities"`
	}

	req := GetCapabilities{
		Xmlns:    Namespace,
		Category: []string{"All"},
	}

	var resp GetCapabilitiesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCapabilities failed: %w", err)
	}

	capabilities := &Capabilities{}

	// Map Analytics
	if resp.Capabilities.Analytics != nil {
		capabilities.Analytics = &AnalyticsCapabilities{
			XAddr:                  resp.Capabilities.Analytics.XAddr,
			RuleSupport:            resp.Capabilities.Analytics.RuleSupport,
			AnalyticsModuleSupport: resp.Capabilities.Analytics.AnalyticsModuleSupport,
		}
	}

	// Map Device
	if resp.Capabilities.Device != nil {
		capabilities.Device = &DeviceCapabilities{
			XAddr: resp.Capabilities.Device.XAddr,
		}
		if resp.Capabilities.Device.Network != nil {
			capabilities.Device.Network = &NetworkCapabilities{
				IPFilter:          resp.Capabilities.Device.Network.IPFilter,
				ZeroConfiguration: resp.Capabilities.Device.Network.ZeroConfiguration,
				IPVersion6:        resp.Capabilities.Device.Network.IPVersion6,
				DynDNS:            resp.Capabilities.Device.Network.DynDNS,
			}
		}
		if resp.Capabilities.Device.System != nil {
			capabilities.Device.System = &SystemCapabilities{
				DiscoveryResolve:  resp.Capabilities.Device.System.DiscoveryResolve,
				DiscoveryBye:      resp.Capabilities.Device.System.DiscoveryBye,
				RemoteDiscovery:   resp.Capabilities.Device.System.RemoteDiscovery,
				SystemBackup:      resp.Capabilities.Device.System.SystemBackup,
				SystemLogging:     resp.Capabilities.Device.System.SystemLogging,
				FirmwareUpgrade:   resp.Capabilities.Device.System.FirmwareUpgrade,
				SupportedVersions: resp.Capabilities.Device.System.SupportedVersions,
			}
		}
		if resp.Capabilities.Device.IO != nil {
			capabilities.Device.IO = &IOCapabilities{
				InputConnectors: resp.Capabilities.Device.IO.InputConnectors,
				RelayOutputs:    resp.Capabilities.Device.IO.RelayOutputs,
			}
		}
		if resp.Capabilities.Device.Security != nil {
			capabilities.Device.Security = &SecurityCapabilities{
				TLS11:                resp.Capabilities.Device.Security.TLS11,
				TLS12:                resp.Capabilities.Device.Security.TLS12,
				OnboardKeyGeneration: resp.Capabilities.Device.Security.OnboardKeyGeneration,
				AccessPolicyConfig:   resp.Capabilities.Device.Security.AccessPolicyConfig,
				X509Token:            resp.Capabilities.Device.Security.X509Token,
				SAMLToken:            resp.Capabilities.Device.Security.SAMLToken,
				KerberosToken:        resp.Capabilities.Device.Security.KerberosToken,
				RELToken:             resp.Capabilities.Device.Security.RELToken,
			}
		}
	}

	// Map Events
	if resp.Capabilities.Events != nil {
		capabilities.Events = &EventCapabilities{
			XAddr:                         resp.Capabilities.Events.XAddr,
			WSSubscriptionPolicySupport:   resp.Capabilities.Events.WSSubscriptionPolicySupport,
			WSPullPointSupport:            resp.Capabilities.Events.WSPullPointSupport,
			WSPausableSubscriptionSupport: resp.Capabilities.Events.WSPausableSubscriptionSupport,
		}
	}

	// Map Imaging
	if resp.Capabilities.Imaging != nil {
		capabilities.Imaging = &ImagingCapabilities{
			XAddr: resp.Capabilities.Imaging.XAddr,
		}
	}

	// Map Media
	if resp.Capabilities.Media != nil {
		capabilities.Media = &MediaCapabilities{
			XAddr: resp.Capabilities.Media.XAddr,
		}
		if resp.Capabilities.Media.StreamingCapabilities != nil {
			capabilities.Media.StreamingCapabilities = &StreamingCapabilities{
				RTPMulticast: resp.Capabilities.Media.StreamingCapabilities.RTPMulticast,
				RTPTCP:       resp.Capabilities.Media.StreamingCapabilities.RTPTCP,
				RTPRTSPTCP:   resp.Capabilities.Media.StreamingCapabilities.RTPRTSPTCP,
			}
		}
	}

	// Map PTZ
	if resp.Capabilities.PTZ != nil {
		capabilities.PTZ = &PTZCapabilities{
			XAddr: resp.Capabilities.PTZ.XAddr,
		}
	}

	return capabilities, nil
}

// SystemReboot reboots the device.
func (s *Service) SystemReboot(ctx context.Context) (string, error) {
	type SystemReboot struct {
		XMLName xml.Name `xml:"tds:SystemReboot"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type SystemRebootResponse struct {
		XMLName xml.Name `xml:"SystemRebootResponse"`
		Message string   `xml:"Message"`
	}

	req := SystemReboot{
		Xmlns: Namespace,
	}

	var resp SystemRebootResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return "", fmt.Errorf("SystemReboot failed: %w", err)
	}

	return resp.Message, nil
}

// GetSystemDateAndTime retrieves the device's system date and time.
//
// Deprecated: the interface{} return never decoded into a usable shape —
// it produced an opaque map. Use FixedGetSystemDateAndTime (the typed
// *SystemDateTime) or GetDeviceTime (assembled time.Time values); this
// method now delegates to FixedGetSystemDateAndTime so existing callers
// receive the typed struct as the interface{} value.
func (s *Service) GetSystemDateAndTime(ctx context.Context) (interface{}, error) {
	st, err := s.FixedGetSystemDateAndTime(ctx)
	if err != nil {
		return nil, err // already wrapped by FixedGetSystemDateAndTime
	}

	return st, nil
}

// GetHostname retrieves the device's hostname.
func (s *Service) GetHostname(ctx context.Context) (*HostnameInformation, error) {
	type GetHostname struct {
		XMLName xml.Name `xml:"tds:GetHostname"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetHostnameResponse struct {
		XMLName             xml.Name `xml:"GetHostnameResponse"`
		HostnameInformation struct {
			FromDHCP bool   `xml:"FromDHCP"`
			Name     string `xml:"Name"`
		} `xml:"HostnameInformation"`
	}

	req := GetHostname{
		Xmlns: Namespace,
	}

	var resp GetHostnameResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetHostname failed: %w", err)
	}

	return &HostnameInformation{
		FromDHCP: resp.HostnameInformation.FromDHCP,
		Name:     resp.HostnameInformation.Name,
	}, nil
}

// SetHostname sets the device's hostname.
func (s *Service) SetHostname(ctx context.Context, name string) error {
	type SetHostname struct {
		XMLName xml.Name `xml:"tds:SetHostname"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Name    string   `xml:"tds:Name"`
	}

	req := SetHostname{
		Xmlns: Namespace,
		Name:  name,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetHostname failed: %w", err)
	}

	return nil
}

// GetDNS retrieves DNS configuration.
func (s *Service) GetDNS(ctx context.Context) (*DNSInformation, error) {
	type GetDNS struct {
		XMLName xml.Name `xml:"tds:GetDNS"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDNSResponse struct {
		XMLName        xml.Name `xml:"GetDNSResponse"`
		DNSInformation struct {
			FromDHCP     bool     `xml:"FromDHCP"`
			SearchDomain []string `xml:"SearchDomain"`
			DNSFromDHCP  []struct {
				Type        string `xml:"Type"`
				IPv4Address string `xml:"IPv4Address"`
			} `xml:"DNSFromDHCP"`
			DNSManual []struct {
				Type        string `xml:"Type"`
				IPv4Address string `xml:"IPv4Address"`
			} `xml:"DNSManual"`
		} `xml:"DNSInformation"`
	}

	req := GetDNS{
		Xmlns: Namespace,
	}

	var resp GetDNSResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDNS failed: %w", err)
	}

	dns := &DNSInformation{
		FromDHCP:     resp.DNSInformation.FromDHCP,
		SearchDomain: resp.DNSInformation.SearchDomain,
	}

	for _, d := range resp.DNSInformation.DNSFromDHCP {
		dns.DNSFromDHCP = append(dns.DNSFromDHCP, types.IPAddress{
			Type:        d.Type,
			IPv4Address: d.IPv4Address,
		})
	}

	for _, d := range resp.DNSInformation.DNSManual {
		dns.DNSManual = append(dns.DNSManual, types.IPAddress{
			Type:        d.Type,
			IPv4Address: d.IPv4Address,
		})
	}

	return dns, nil
}

// GetNTP retrieves NTP configuration.
func (s *Service) GetNTP(ctx context.Context) (*NTPInformation, error) {
	type GetNTP struct {
		XMLName xml.Name `xml:"tds:GetNTP"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetNTPResponse struct {
		XMLName        xml.Name `xml:"GetNTPResponse"`
		NTPInformation struct {
			FromDHCP    bool `xml:"FromDHCP"`
			NTPFromDHCP []struct {
				Type        string `xml:"Type"`
				IPv4Address string `xml:"IPv4Address"`
				DNSname     string `xml:"DNSname"`
			} `xml:"NTPFromDHCP"`
			NTPManual []struct {
				Type        string `xml:"Type"`
				IPv4Address string `xml:"IPv4Address"`
				DNSname     string `xml:"DNSname"`
			} `xml:"NTPManual"`
		} `xml:"NTPInformation"`
	}

	req := GetNTP{
		Xmlns: Namespace,
	}

	var resp GetNTPResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetNTP failed: %w", err)
	}

	ntp := &NTPInformation{
		FromDHCP: resp.NTPInformation.FromDHCP,
	}

	for _, n := range resp.NTPInformation.NTPFromDHCP {
		ntp.NTPFromDHCP = append(ntp.NTPFromDHCP, NetworkHost{
			Type:        n.Type,
			IPv4Address: n.IPv4Address,
			DNSname:     n.DNSname,
		})
	}

	for _, n := range resp.NTPInformation.NTPManual {
		ntp.NTPManual = append(ntp.NTPManual, NetworkHost{
			Type:        n.Type,
			IPv4Address: n.IPv4Address,
			DNSname:     n.DNSname,
		})
	}

	return ntp, nil
}

// GetNetworkInterfaces retrieves network interface configuration.
func (s *Service) GetNetworkInterfaces(ctx context.Context) ([]*NetworkInterface, error) {
	type GetNetworkInterfaces struct {
		XMLName xml.Name `xml:"tds:GetNetworkInterfaces"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetNetworkInterfacesResponse struct {
		XMLName           xml.Name `xml:"GetNetworkInterfacesResponse"`
		NetworkInterfaces []struct {
			Token   string `xml:"token,attr"`
			Enabled bool   `xml:"Enabled"`
			Info    struct {
				Name      string `xml:"Name"`
				HwAddress string `xml:"HwAddress"`
				MTU       int    `xml:"MTU"`
			} `xml:"Info"`
			IPv4 struct {
				Enabled bool `xml:"Enabled"`
				Config  struct {
					Manual []struct {
						Address      string `xml:"Address"`
						PrefixLength int    `xml:"PrefixLength"`
					} `xml:"Manual"`
					DHCP bool `xml:"DHCP"`
				} `xml:"Config"`
			} `xml:"IPv4"`
		} `xml:"NetworkInterfaces"`
	}

	req := GetNetworkInterfaces{
		Xmlns: Namespace,
	}

	var resp GetNetworkInterfacesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetNetworkInterfaces failed: %w", err)
	}

	interfaces := make([]*NetworkInterface, len(resp.NetworkInterfaces))
	for i, iface := range resp.NetworkInterfaces {
		ni := &NetworkInterface{
			Token:   iface.Token,
			Enabled: iface.Enabled,
			Info: NetworkInterfaceInfo{
				Name:      iface.Info.Name,
				HwAddress: iface.Info.HwAddress,
				MTU:       iface.Info.MTU,
			},
		}

		if iface.IPv4.Enabled {
			ni.IPv4 = &IPv4NetworkInterface{
				Enabled: iface.IPv4.Enabled,
				Config: IPv4Configuration{
					DHCP: iface.IPv4.Config.DHCP,
				},
			}

			for _, m := range iface.IPv4.Config.Manual {
				ni.IPv4.Config.Manual = append(ni.IPv4.Config.Manual, types.PrefixedIPv4Address{
					Address:      m.Address,
					PrefixLength: m.PrefixLength,
					Netmask:      NetmaskFromPrefixLength(m.PrefixLength),
				})
			}
		}

		interfaces[i] = ni
	}

	return interfaces, nil
}

// GetScopes retrieves configured scopes.
func (s *Service) GetScopes(ctx context.Context) ([]*Scope, error) {
	type GetScopes struct {
		XMLName xml.Name `xml:"tds:GetScopes"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetScopesResponse struct {
		XMLName xml.Name `xml:"GetScopesResponse"`
		Scopes  []struct {
			ScopeDef  string `xml:"ScopeDef"`
			ScopeItem string `xml:"ScopeItem"`
		} `xml:"Scopes"`
	}

	req := GetScopes{
		Xmlns: Namespace,
	}

	var resp GetScopesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetScopes failed: %w", err)
	}

	scopes := make([]*Scope, len(resp.Scopes))
	for i, s := range resp.Scopes {
		scopes[i] = &Scope{
			ScopeDef:  s.ScopeDef,
			ScopeItem: s.ScopeItem,
		}
	}

	return scopes, nil
}

func (s *Service) GetServices(ctx context.Context, includeCapability bool) ([]*ServiceEntry, error) {
	type GetServices struct {
		XMLName           xml.Name `xml:"tds:GetServices"`
		Xmlns             string   `xml:"xmlns:tds,attr"`
		IncludeCapability bool     `xml:"tds:IncludeCapability"`
	}

	type GetServicesResponse struct {
		XMLName xml.Name `xml:"GetServicesResponse"`
		Service []struct {
			Namespace    string      `xml:"Namespace"`
			XAddr        string      `xml:"XAddr"`
			Capabilities interface{} `xml:"Capabilities"`
			Version      struct {
				Major int `xml:"Major"`
				Minor int `xml:"Minor"`
			} `xml:"Version"`
		} `xml:"Service"`
	}

	req := GetServices{
		Xmlns:             Namespace,
		IncludeCapability: includeCapability,
	}

	var resp GetServicesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetServices failed: %w", err)
	}

	services := make([]*ServiceEntry, len(resp.Service))
	for i, svc := range resp.Service {
		services[i] = &ServiceEntry{
			Namespace:    svc.Namespace,
			XAddr:        svc.XAddr,
			Capabilities: svc.Capabilities,
			Version: OnvifVersion{
				Major: svc.Version.Major,
				Minor: svc.Version.Minor,
			},
		}
	}

	return services, nil
}

// GetDeviceServiceCapabilities returns the capabilities of the device service.
func (s *Service) GetDeviceServiceCapabilities(ctx context.Context) (*DeviceServiceCapabilities, error) {
	type GetServiceCapabilities struct {
		XMLName xml.Name `xml:"tds:GetServiceCapabilities"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetServiceCapabilitiesResponse struct {
		XMLName      xml.Name `xml:"GetServiceCapabilitiesResponse"`
		Capabilities struct {
			Network struct {
				IPFilter          bool `xml:"IPFilter,attr"`
				ZeroConfiguration bool `xml:"ZeroConfiguration,attr"`
				IPVersion6        bool `xml:"IPVersion6,attr"`
				DynDNS            bool `xml:"DynDNS,attr"`
			} `xml:"Network"`
			Security struct {
				TLS10                bool `xml:"TLS1.0,attr"`
				TLS11                bool `xml:"TLS1.1,attr"`
				TLS12                bool `xml:"TLS1.2,attr"`
				OnboardKeyGeneration bool `xml:"OnboardKeyGeneration,attr"`
				AccessPolicyConfig   bool `xml:"AccessPolicyConfig,attr"`
			} `xml:"Security"`
			System struct {
				DiscoveryResolve bool `xml:"DiscoveryResolve,attr"`
				DiscoveryBye     bool `xml:"DiscoveryBye,attr"`
				RemoteDiscovery  bool `xml:"RemoteDiscovery,attr"`
				SystemBackup     bool `xml:"SystemBackup,attr"`
				SystemLogging    bool `xml:"SystemLogging,attr"`
				FirmwareUpgrade  bool `xml:"FirmwareUpgrade,attr"`
			} `xml:"System"`
		} `xml:"Capabilities"`
	}

	req := GetServiceCapabilities{
		Xmlns: Namespace,
	}

	var resp GetServiceCapabilitiesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDeviceServiceCapabilities failed: %w", err)
	}

	return &DeviceServiceCapabilities{
		Network: &NetworkCapabilities{
			IPFilter:          resp.Capabilities.Network.IPFilter,
			ZeroConfiguration: resp.Capabilities.Network.ZeroConfiguration,
			IPVersion6:        resp.Capabilities.Network.IPVersion6,
			DynDNS:            resp.Capabilities.Network.DynDNS,
		},
		Security: &SecurityCapabilities{
			TLS11:                resp.Capabilities.Security.TLS11,
			TLS12:                resp.Capabilities.Security.TLS12,
			OnboardKeyGeneration: resp.Capabilities.Security.OnboardKeyGeneration,
			AccessPolicyConfig:   resp.Capabilities.Security.AccessPolicyConfig,
		},
		System: &SystemCapabilities{
			DiscoveryResolve: resp.Capabilities.System.DiscoveryResolve,
			DiscoveryBye:     resp.Capabilities.System.DiscoveryBye,
			RemoteDiscovery:  resp.Capabilities.System.RemoteDiscovery,
			SystemBackup:     resp.Capabilities.System.SystemBackup,
			SystemLogging:    resp.Capabilities.System.SystemLogging,
			FirmwareUpgrade:  resp.Capabilities.System.FirmwareUpgrade,
		},
	}, nil
}

// GetDiscoveryMode gets the discovery mode of a device.
func (s *Service) GetDiscoveryMode(ctx context.Context) (DiscoveryMode, error) {
	type GetDiscoveryMode struct {
		XMLName xml.Name `xml:"tds:GetDiscoveryMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDiscoveryModeResponse struct {
		XMLName       xml.Name `xml:"GetDiscoveryModeResponse"`
		DiscoveryMode string   `xml:"DiscoveryMode"`
	}

	req := GetDiscoveryMode{
		Xmlns: Namespace,
	}

	var resp GetDiscoveryModeResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return "", fmt.Errorf("GetDiscoveryMode failed: %w", err)
	}

	return DiscoveryMode(resp.DiscoveryMode), nil
}

// SetDiscoveryMode sets the discovery mode of a device.
func (s *Service) SetDiscoveryMode(ctx context.Context, mode DiscoveryMode) error {
	type SetDiscoveryMode struct {
		XMLName       xml.Name      `xml:"tds:SetDiscoveryMode"`
		Xmlns         string        `xml:"xmlns:tds,attr"`
		DiscoveryMode DiscoveryMode `xml:"tds:DiscoveryMode"`
	}

	req := SetDiscoveryMode{
		Xmlns:         Namespace,
		DiscoveryMode: mode,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetDiscoveryMode failed: %w", err)
	}

	return nil
}

// GetRemoteDiscoveryMode gets the remote discovery mode.
func (s *Service) GetRemoteDiscoveryMode(ctx context.Context) (DiscoveryMode, error) {
	type GetRemoteDiscoveryMode struct {
		XMLName xml.Name `xml:"tds:GetRemoteDiscoveryMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetRemoteDiscoveryModeResponse struct {
		XMLName             xml.Name `xml:"GetRemoteDiscoveryModeResponse"`
		RemoteDiscoveryMode string   `xml:"RemoteDiscoveryMode"`
	}

	req := GetRemoteDiscoveryMode{
		Xmlns: Namespace,
	}

	var resp GetRemoteDiscoveryModeResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return "", fmt.Errorf("GetRemoteDiscoveryMode failed: %w", err)
	}

	return DiscoveryMode(resp.RemoteDiscoveryMode), nil
}

// SetRemoteDiscoveryMode sets the remote discovery mode.
func (s *Service) SetRemoteDiscoveryMode(ctx context.Context, mode DiscoveryMode) error {
	type SetRemoteDiscoveryMode struct {
		XMLName             xml.Name      `xml:"tds:SetRemoteDiscoveryMode"`
		Xmlns               string        `xml:"xmlns:tds,attr"`
		RemoteDiscoveryMode DiscoveryMode `xml:"tds:RemoteDiscoveryMode"`
	}

	req := SetRemoteDiscoveryMode{
		Xmlns:               Namespace,
		RemoteDiscoveryMode: mode,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetRemoteDiscoveryMode failed: %w", err)
	}

	return nil
}

// GetEndpointReference gets the endpoint reference GUID.
func (s *Service) GetEndpointReference(ctx context.Context) (string, error) {
	type GetEndpointReference struct {
		XMLName xml.Name `xml:"tds:GetEndpointReference"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetEndpointReferenceResponse struct {
		XMLName xml.Name `xml:"GetEndpointReferenceResponse"`
		GUID    string   `xml:"GUID"`
	}

	req := GetEndpointReference{
		Xmlns: Namespace,
	}

	var resp GetEndpointReferenceResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return "", fmt.Errorf("GetEndpointReference failed: %w", err)
	}

	return resp.GUID, nil
}

// GetNetworkProtocols gets defined network protocols from a device.
func (s *Service) GetNetworkProtocols(ctx context.Context) ([]*NetworkProtocol, error) {
	type GetNetworkProtocols struct {
		XMLName xml.Name `xml:"tds:GetNetworkProtocols"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetNetworkProtocolsResponse struct {
		XMLName          xml.Name `xml:"GetNetworkProtocolsResponse"`
		NetworkProtocols []struct {
			Name    string `xml:"Name"`
			Enabled bool   `xml:"Enabled"`
			Port    []int  `xml:"Port"`
		} `xml:"NetworkProtocols"`
	}

	req := GetNetworkProtocols{
		Xmlns: Namespace,
	}

	var resp GetNetworkProtocolsResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetNetworkProtocols failed: %w", err)
	}

	protocols := make([]*NetworkProtocol, len(resp.NetworkProtocols))
	for i, proto := range resp.NetworkProtocols {
		protocols[i] = &NetworkProtocol{
			Name:    NetworkProtocolType(proto.Name),
			Enabled: proto.Enabled,
			Port:    proto.Port,
		}
	}

	return protocols, nil
}

// SetNetworkProtocols configures defined network protocols on a device.
func (s *Service) SetNetworkProtocols(ctx context.Context, protocols []*NetworkProtocol) error {
	type SetNetworkProtocols struct {
		XMLName          xml.Name `xml:"tds:SetNetworkProtocols"`
		Xmlns            string   `xml:"xmlns:tds,attr"`
		NetworkProtocols []struct {
			Name    string `xml:"tds:Name"`
			Enabled bool   `xml:"tds:Enabled"`
			Port    []int  `xml:"tds:Port"`
		} `xml:"tds:NetworkProtocols"`
	}

	req := SetNetworkProtocols{
		Xmlns: Namespace,
	}

	for _, proto := range protocols {
		req.NetworkProtocols = append(req.NetworkProtocols, struct {
			Name    string `xml:"tds:Name"`
			Enabled bool   `xml:"tds:Enabled"`
			Port    []int  `xml:"tds:Port"`
		}{
			Name:    string(proto.Name),
			Enabled: proto.Enabled,
			Port:    proto.Port,
		})
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetNetworkProtocols failed: %w", err)
	}

	return nil
}

// GetNetworkDefaultGateway gets the default gateway settings from a device.
func (s *Service) GetNetworkDefaultGateway(ctx context.Context) (*NetworkGateway, error) {
	type GetNetworkDefaultGateway struct {
		XMLName xml.Name `xml:"tds:GetNetworkDefaultGateway"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetNetworkDefaultGatewayResponse struct {
		XMLName        xml.Name `xml:"GetNetworkDefaultGatewayResponse"`
		NetworkGateway struct {
			IPv4Address []string `xml:"IPv4Address"`
			IPv6Address []string `xml:"IPv6Address"`
		} `xml:"NetworkGateway"`
	}

	req := GetNetworkDefaultGateway{
		Xmlns: Namespace,
	}

	var resp GetNetworkDefaultGatewayResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetNetworkDefaultGateway failed: %w", err)
	}

	return &NetworkGateway{
		IPv4Address: resp.NetworkGateway.IPv4Address,
		IPv6Address: resp.NetworkGateway.IPv6Address,
	}, nil
}

// SetNetworkDefaultGateway sets the default gateway settings on a device.
func (s *Service) SetNetworkDefaultGateway(ctx context.Context, gateway *NetworkGateway) error {
	type SetNetworkDefaultGateway struct {
		XMLName     xml.Name `xml:"tds:SetNetworkDefaultGateway"`
		Xmlns       string   `xml:"xmlns:tds,attr"`
		IPv4Address []string `xml:"tds:IPv4Address,omitempty"`
		IPv6Address []string `xml:"tds:IPv6Address,omitempty"`
	}

	req := SetNetworkDefaultGateway{
		Xmlns:       Namespace,
		IPv4Address: gateway.IPv4Address,
		IPv6Address: gateway.IPv6Address,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetNetworkDefaultGateway failed: %w", err)
	}

	return nil
}

// GetGeoLocation retrieves geographic location information. ONVIF Specification: GetGeoLocation operation.
func (s *Service) GetGeoLocation(ctx context.Context) ([]LocationEntity, error) {
	type GetGeoLocationBody struct {
		XMLName xml.Name `xml:"tds:GetGeoLocation"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetGeoLocationResponse struct {
		XMLName  xml.Name         `xml:"GetGeoLocationResponse"`
		Location []LocationEntity `xml:"Location"`
	}

	request := GetGeoLocationBody{
		Xmlns: Namespace,
	}
	var response GetGeoLocationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetGeoLocation failed: %w", err)
	}

	return response.Location, nil
}

// SetGeoLocation sets geographic location information. ONVIF Specification: SetGeoLocation operation.
func (s *Service) SetGeoLocation(ctx context.Context, location []LocationEntity) error {
	type SetGeoLocationBody struct {
		XMLName  xml.Name         `xml:"tds:SetGeoLocation"`
		Xmlns    string           `xml:"xmlns:tds,attr"`
		Location []LocationEntity `xml:"tds:Location"`
	}

	type SetGeoLocationResponse struct {
		XMLName xml.Name `xml:"SetGeoLocationResponse"`
	}

	request := SetGeoLocationBody{
		Xmlns:    Namespace,
		Location: location,
	}
	var response SetGeoLocationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("SetGeoLocation failed: %w", err)
	}

	return nil
}

// DeleteGeoLocation deletes geographic location information. ONVIF Specification: DeleteGeoLocation operation.
func (s *Service) DeleteGeoLocation(ctx context.Context, location []LocationEntity) error {
	type DeleteGeoLocationBody struct {
		XMLName  xml.Name         `xml:"tds:DeleteGeoLocation"`
		Xmlns    string           `xml:"xmlns:tds,attr"`
		Location []LocationEntity `xml:"tds:Location"`
	}

	type DeleteGeoLocationResponse struct {
		XMLName xml.Name `xml:"DeleteGeoLocationResponse"`
	}

	request := DeleteGeoLocationBody{
		Xmlns:    Namespace,
		Location: location,
	}
	var response DeleteGeoLocationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("DeleteGeoLocation failed: %w", err)
	}

	return nil
}

// GetDPAddresses retrieves DP (Device Provisioning) addresses. ONVIF Specification: GetDPAddresses operation.
func (s *Service) GetDPAddresses(ctx context.Context) ([]NetworkHost, error) {
	type GetDPAddressesBody struct {
		XMLName xml.Name `xml:"tds:GetDPAddresses"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDPAddressesResponse struct {
		XMLName   xml.Name      `xml:"GetDPAddressesResponse"`
		DPAddress []NetworkHost `xml:"DPAddress"`
	}

	request := GetDPAddressesBody{
		Xmlns: Namespace,
	}
	var response GetDPAddressesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetDPAddresses failed: %w", err)
	}

	return response.DPAddress, nil
}

// SetDPAddresses sets DP (Device Provisioning) addresses. ONVIF Specification: SetDPAddresses operation.
func (s *Service) SetDPAddresses(ctx context.Context, dpAddress []NetworkHost) error {
	type SetDPAddressesBody struct {
		XMLName   xml.Name      `xml:"tds:SetDPAddresses"`
		Xmlns     string        `xml:"xmlns:tds,attr"`
		DPAddress []NetworkHost `xml:"tds:DPAddress"`
	}

	type SetDPAddressesResponse struct {
		XMLName xml.Name `xml:"SetDPAddressesResponse"`
	}

	request := SetDPAddressesBody{
		Xmlns:     Namespace,
		DPAddress: dpAddress,
	}
	var response SetDPAddressesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("SetDPAddresses failed: %w", err)
	}

	return nil
}

func (s *Service) GetWsdlURL(ctx context.Context) (string, error) {
	type GetWsdlURLBody struct {
		XMLName xml.Name `xml:"tds:GetWsdlUrl"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetWsdlURLResponse struct {
		XMLName xml.Name `xml:"GetWsdlUrlResponse"`
		WsdlURL string   `xml:"WsdlUrl"`
	}

	request := GetWsdlURLBody{
		Xmlns: Namespace,
	}
	var response GetWsdlURLResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return "", fmt.Errorf("GetWsdlURL failed: %w", err)
	}

	return response.WsdlURL, nil
}
