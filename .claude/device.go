package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// Device service namespace.
const deviceNamespace = "http://www.onvif.org/ver10/device/wsdl"

// GetDeviceInformation retrieves device information.
func (c *Client) GetDeviceInformation(ctx context.Context) (*DeviceInformation, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetDeviceInformationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) GetCapabilities(ctx context.Context) (*Capabilities, error) {
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
		Xmlns:    deviceNamespace,
		Category: []string{"All"},
	}

	var resp GetCapabilitiesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) SystemReboot(ctx context.Context) (string, error) {
	type SystemReboot struct {
		XMLName xml.Name `xml:"tds:SystemReboot"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type SystemRebootResponse struct {
		XMLName xml.Name `xml:"SystemRebootResponse"`
		Message string   `xml:"Message"`
	}

	req := SystemReboot{
		Xmlns: deviceNamespace,
	}

	var resp SystemRebootResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return "", fmt.Errorf("SystemReboot failed: %w", err)
	}

	return resp.Message, nil
}

// GetSystemDateAndTime retrieves the device's system date and time.
func (c *Client) GetSystemDateAndTime(ctx context.Context) (interface{}, error) {
	type GetSystemDateAndTime struct {
		XMLName xml.Name `xml:"tds:GetSystemDateAndTime"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	req := GetSystemDateAndTime{
		Xmlns: deviceNamespace,
	}

	var resp interface{}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSystemDateAndTime failed: %w", err)
	}

	return resp, nil
}

// GetHostname retrieves the device's hostname.
func (c *Client) GetHostname(ctx context.Context) (*HostnameInformation, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetHostnameResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetHostname failed: %w", err)
	}

	return &HostnameInformation{
		FromDHCP: resp.HostnameInformation.FromDHCP,
		Name:     resp.HostnameInformation.Name,
	}, nil
}

// SetHostname sets the device's hostname.
func (c *Client) SetHostname(ctx context.Context, name string) error {
	type SetHostname struct {
		XMLName xml.Name `xml:"tds:SetHostname"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Name    string   `xml:"tds:Name"`
	}

	req := SetHostname{
		Xmlns: deviceNamespace,
		Name:  name,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetHostname failed: %w", err)
	}

	return nil
}

// GetDNS retrieves DNS configuration.
func (c *Client) GetDNS(ctx context.Context) (*DNSInformation, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetDNSResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDNS failed: %w", err)
	}

	dns := &DNSInformation{
		FromDHCP:     resp.DNSInformation.FromDHCP,
		SearchDomain: resp.DNSInformation.SearchDomain,
	}

	for _, d := range resp.DNSInformation.DNSFromDHCP {
		dns.DNSFromDHCP = append(dns.DNSFromDHCP, IPAddress{
			Type:        d.Type,
			IPv4Address: d.IPv4Address,
		})
	}

	for _, d := range resp.DNSInformation.DNSManual {
		dns.DNSManual = append(dns.DNSManual, IPAddress{
			Type:        d.Type,
			IPv4Address: d.IPv4Address,
		})
	}

	return dns, nil
}

// GetNTP retrieves NTP configuration.
func (c *Client) GetNTP(ctx context.Context) (*NTPInformation, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetNTPResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) GetNetworkInterfaces(ctx context.Context) ([]*NetworkInterface, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetNetworkInterfacesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
				ni.IPv4.Config.Manual = append(ni.IPv4.Config.Manual, PrefixedIPv4Address{
					Address:      m.Address,
					PrefixLength: m.PrefixLength,
				})
			}
		}

		interfaces[i] = ni
	}

	return interfaces, nil
}

// GetScopes retrieves configured scopes.
func (c *Client) GetScopes(ctx context.Context) ([]*Scope, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetScopesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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

// GetUsers retrieves user accounts.
func (c *Client) GetUsers(ctx context.Context) ([]*User, error) {
	type GetUsers struct {
		XMLName xml.Name `xml:"tds:GetUsers"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetUsersResponse struct {
		XMLName xml.Name `xml:"GetUsersResponse"`
		User    []struct {
			Username  string `xml:"Username"`
			UserLevel string `xml:"UserLevel"`
		} `xml:"User"`
	}

	req := GetUsers{
		Xmlns: deviceNamespace,
	}

	var resp GetUsersResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetUsers failed: %w", err)
	}

	users := make([]*User, len(resp.User))
	for i, u := range resp.User {
		users[i] = &User{
			Username:  u.Username,
			UserLevel: u.UserLevel,
		}
	}

	return users, nil
}

// CreateUsers creates new user accounts.
func (c *Client) CreateUsers(ctx context.Context, users []*User) error {
	type CreateUsers struct {
		XMLName xml.Name `xml:"tds:CreateUsers"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		User    []struct {
			Username  string `xml:"tds:Username"`
			Password  string `xml:"tds:Password"`
			UserLevel string `xml:"tds:UserLevel"`
		} `xml:"tds:User"`
	}

	req := CreateUsers{
		Xmlns: deviceNamespace,
	}

	for _, user := range users {
		req.User = append(req.User, struct {
			Username  string `xml:"tds:Username"`
			Password  string `xml:"tds:Password"`
			UserLevel string `xml:"tds:UserLevel"`
		}{
			Username:  user.Username,
			Password:  user.Password,
			UserLevel: user.UserLevel,
		})
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("CreateUsers failed: %w", err)
	}

	return nil
}

// DeleteUsers deletes user accounts.
func (c *Client) DeleteUsers(ctx context.Context, usernames []string) error {
	type DeleteUsers struct {
		XMLName  xml.Name `xml:"tds:DeleteUsers"`
		Xmlns    string   `xml:"xmlns:tds,attr"`
		Username []string `xml:"tds:Username"`
	}

	req := DeleteUsers{
		Xmlns:    deviceNamespace,
		Username: usernames,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteUsers failed: %w", err)
	}

	return nil
}

// SetUser modifies an existing user account.
func (c *Client) SetUser(ctx context.Context, user *User) error {
	type SetUser struct {
		XMLName xml.Name `xml:"tds:SetUser"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		User    struct {
			Username  string  `xml:"tds:Username"`
			Password  *string `xml:"tds:Password,omitempty"`
			UserLevel string  `xml:"tds:UserLevel"`
		} `xml:"tds:User"`
	}

	req := SetUser{
		Xmlns: deviceNamespace,
	}
	req.User.Username = user.Username
	if user.Password != "" {
		req.User.Password = &user.Password
	}
	req.User.UserLevel = user.UserLevel

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetUser failed: %w", err)
	}

	return nil
}

// GetServices returns information about services on the device.
func (c *Client) GetServices(ctx context.Context, includeCapability bool) ([]*Service, error) {
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
		Xmlns:             deviceNamespace,
		IncludeCapability: includeCapability,
	}

	var resp GetServicesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetServices failed: %w", err)
	}

	services := make([]*Service, len(resp.Service))
	for i, svc := range resp.Service {
		services[i] = &Service{
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

// GetServiceCapabilities returns the capabilities of the device service.
func (c *Client) GetServiceCapabilities(ctx context.Context) (*DeviceServiceCapabilities, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetServiceCapabilitiesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetServiceCapabilities failed: %w", err)
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
func (c *Client) GetDiscoveryMode(ctx context.Context) (DiscoveryMode, error) {
	type GetDiscoveryMode struct {
		XMLName xml.Name `xml:"tds:GetDiscoveryMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDiscoveryModeResponse struct {
		XMLName       xml.Name `xml:"GetDiscoveryModeResponse"`
		DiscoveryMode string   `xml:"DiscoveryMode"`
	}

	req := GetDiscoveryMode{
		Xmlns: deviceNamespace,
	}

	var resp GetDiscoveryModeResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return "", fmt.Errorf("GetDiscoveryMode failed: %w", err)
	}

	return DiscoveryMode(resp.DiscoveryMode), nil
}

// SetDiscoveryMode sets the discovery mode of a device.
func (c *Client) SetDiscoveryMode(ctx context.Context, mode DiscoveryMode) error {
	type SetDiscoveryMode struct {
		XMLName       xml.Name      `xml:"tds:SetDiscoveryMode"`
		Xmlns         string        `xml:"xmlns:tds,attr"`
		DiscoveryMode DiscoveryMode `xml:"tds:DiscoveryMode"`
	}

	req := SetDiscoveryMode{
		Xmlns:         deviceNamespace,
		DiscoveryMode: mode,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetDiscoveryMode failed: %w", err)
	}

	return nil
}

// GetRemoteDiscoveryMode gets the remote discovery mode.
func (c *Client) GetRemoteDiscoveryMode(ctx context.Context) (DiscoveryMode, error) {
	type GetRemoteDiscoveryMode struct {
		XMLName xml.Name `xml:"tds:GetRemoteDiscoveryMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetRemoteDiscoveryModeResponse struct {
		XMLName             xml.Name `xml:"GetRemoteDiscoveryModeResponse"`
		RemoteDiscoveryMode string   `xml:"RemoteDiscoveryMode"`
	}

	req := GetRemoteDiscoveryMode{
		Xmlns: deviceNamespace,
	}

	var resp GetRemoteDiscoveryModeResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return "", fmt.Errorf("GetRemoteDiscoveryMode failed: %w", err)
	}

	return DiscoveryMode(resp.RemoteDiscoveryMode), nil
}

// SetRemoteDiscoveryMode sets the remote discovery mode.
func (c *Client) SetRemoteDiscoveryMode(ctx context.Context, mode DiscoveryMode) error {
	type SetRemoteDiscoveryMode struct {
		XMLName             xml.Name      `xml:"tds:SetRemoteDiscoveryMode"`
		Xmlns               string        `xml:"xmlns:tds,attr"`
		RemoteDiscoveryMode DiscoveryMode `xml:"tds:RemoteDiscoveryMode"`
	}

	req := SetRemoteDiscoveryMode{
		Xmlns:               deviceNamespace,
		RemoteDiscoveryMode: mode,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetRemoteDiscoveryMode failed: %w", err)
	}

	return nil
}

// GetEndpointReference gets the endpoint reference GUID.
func (c *Client) GetEndpointReference(ctx context.Context) (string, error) {
	type GetEndpointReference struct {
		XMLName xml.Name `xml:"tds:GetEndpointReference"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetEndpointReferenceResponse struct {
		XMLName xml.Name `xml:"GetEndpointReferenceResponse"`
		GUID    string   `xml:"GUID"`
	}

	req := GetEndpointReference{
		Xmlns: deviceNamespace,
	}

	var resp GetEndpointReferenceResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return "", fmt.Errorf("GetEndpointReference failed: %w", err)
	}

	return resp.GUID, nil
}

// GetNetworkProtocols gets defined network protocols from a device.
func (c *Client) GetNetworkProtocols(ctx context.Context) ([]*NetworkProtocol, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetNetworkProtocolsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) SetNetworkProtocols(ctx context.Context, protocols []*NetworkProtocol) error {
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
		Xmlns: deviceNamespace,
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

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetNetworkProtocols failed: %w", err)
	}

	return nil
}

// GetNetworkDefaultGateway gets the default gateway settings from a device.
func (c *Client) GetNetworkDefaultGateway(ctx context.Context) (*NetworkGateway, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetNetworkDefaultGatewayResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetNetworkDefaultGateway failed: %w", err)
	}

	return &NetworkGateway{
		IPv4Address: resp.NetworkGateway.IPv4Address,
		IPv6Address: resp.NetworkGateway.IPv6Address,
	}, nil
}

// SetNetworkDefaultGateway sets the default gateway settings on a device.
func (c *Client) SetNetworkDefaultGateway(ctx context.Context, gateway *NetworkGateway) error {
	type SetNetworkDefaultGateway struct {
		XMLName     xml.Name `xml:"tds:SetNetworkDefaultGateway"`
		Xmlns       string   `xml:"xmlns:tds,attr"`
		IPv4Address []string `xml:"tds:IPv4Address,omitempty"`
		IPv6Address []string `xml:"tds:IPv6Address,omitempty"`
	}

	req := SetNetworkDefaultGateway{
		Xmlns:       deviceNamespace,
		IPv4Address: gateway.IPv4Address,
		IPv6Address: gateway.IPv6Address,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetNetworkDefaultGateway failed: %w", err)
	}

	return nil
}
