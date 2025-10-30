package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524A/go-onvif/soap"
)

// Device service namespace
const deviceNamespace = "http://www.onvif.org/ver10/device/wsdl"

// GetDeviceInformation retrieves device information
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

// GetCapabilities retrieves device capabilities
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
					RTP_TCP      bool `xml:"RTP_TCP"`
					RTP_RTSP_TCP bool `xml:"RTP_RTSP_TCP"`
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
				RTP_TCP:      resp.Capabilities.Media.StreamingCapabilities.RTP_TCP,
				RTP_RTSP_TCP: resp.Capabilities.Media.StreamingCapabilities.RTP_RTSP_TCP,
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

// SystemReboot reboots the device
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

// GetSystemDateAndTime retrieves the device's system date and time
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
