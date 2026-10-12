// Device management: system date/time, DNS/NTP, network interfaces and
// gateway, storage.

package device

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// SetDNS sets the DNS settings on a device.
func (s *Service) SetDNS(ctx context.Context, fromDHCP bool, searchDomain []string, dnsManual []types.IPAddress) error {
	type SetDNS struct {
		XMLName      xml.Name `xml:"tds:SetDNS"`
		Xmlns        string   `xml:"xmlns:tds,attr"`
		XmlnsTT      string   `xml:"xmlns:tt,attr"`
		FromDHCP     bool     `xml:"tds:FromDHCP"`
		SearchDomain []string `xml:"tds:SearchDomain,omitempty"`
		DNSManual    []struct {
			Type        string `xml:"tt:Type"`
			IPv4Address string `xml:"tt:IPv4Address,omitempty"`
			IPv6Address string `xml:"tt:IPv6Address,omitempty"`
		} `xml:"tds:DNSManual,omitempty"`
	}

	req := SetDNS{
		Xmlns:        Namespace,
		XmlnsTT:      "http://www.onvif.org/ver10/schema",
		FromDHCP:     fromDHCP,
		SearchDomain: searchDomain,
	}

	for _, dns := range dnsManual {
		req.DNSManual = append(req.DNSManual, struct {
			Type        string `xml:"tt:Type"`
			IPv4Address string `xml:"tt:IPv4Address,omitempty"`
			IPv6Address string `xml:"tt:IPv6Address,omitempty"`
		}{
			Type:        dns.Type,
			IPv4Address: dns.IPv4Address,
			IPv6Address: dns.IPv6Address,
		})
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetDNS failed: %w", err)
	}

	return nil
}

// SetNTP sets the NTP settings on a device.
// SetNTP configures the NTP servers. Admin-level write: a client
// constructed without credentials fails fast with ErrNoCredentials.
func (s *Service) SetNTP(ctx context.Context, fromDHCP bool, ntpManual []NetworkHost) error {
	if err := s.requireCredentials(); err != nil {
		return fmt.Errorf("SetNTP: %w", err)
	}

	type SetNTP struct {
		XMLName   xml.Name `xml:"tds:SetNTP"`
		Xmlns     string   `xml:"xmlns:tds,attr"`
		XmlnsTT   string   `xml:"xmlns:tt,attr"`
		FromDHCP  bool     `xml:"tds:FromDHCP"`
		NTPManual []struct {
			Type        string `xml:"tt:Type"`
			IPv4Address string `xml:"tt:IPv4Address,omitempty"`
			IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			DNSname     string `xml:"tt:DNSname,omitempty"`
		} `xml:"tds:NTPManual,omitempty"`
	}

	req := SetNTP{
		Xmlns:    Namespace,
		XmlnsTT:  "http://www.onvif.org/ver10/schema",
		FromDHCP: fromDHCP,
	}

	for _, ntp := range ntpManual {
		req.NTPManual = append(req.NTPManual, struct {
			Type        string `xml:"tt:Type"`
			IPv4Address string `xml:"tt:IPv4Address,omitempty"`
			IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			DNSname     string `xml:"tt:DNSname,omitempty"`
		}{
			Type:        ntp.Type,
			IPv4Address: ntp.IPv4Address,
			IPv6Address: ntp.IPv6Address,
			DNSname:     ntp.DNSname,
		})
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetNTP failed: %w", err)
	}

	return nil
}

// SetHostnameFromDHCP controls whether the hostname is set manually or retrieved via DHCP.
func (s *Service) SetHostnameFromDHCP(ctx context.Context, fromDHCP bool) (bool, error) {
	type SetHostnameFromDHCP struct {
		XMLName  xml.Name `xml:"tds:SetHostnameFromDHCP"`
		Xmlns    string   `xml:"xmlns:tds,attr"`
		FromDHCP bool     `xml:"tds:FromDHCP"`
	}

	type SetHostnameFromDHCPResponse struct {
		XMLName      xml.Name `xml:"SetHostnameFromDHCPResponse"`
		RebootNeeded bool     `xml:"RebootNeeded"`
	}

	req := SetHostnameFromDHCP{
		Xmlns:    Namespace,
		FromDHCP: fromDHCP,
	}

	var resp SetHostnameFromDHCPResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return false, fmt.Errorf("SetHostnameFromDHCP failed: %w", err)
	}

	return resp.RebootNeeded, nil
}

// FixedGetSystemDateAndTime retrieves the device's system date and time with proper typing.
func (s *Service) FixedGetSystemDateAndTime(ctx context.Context) (*SystemDateTime, error) {
	type GetSystemDateAndTime struct {
		XMLName xml.Name `xml:"tds:GetSystemDateAndTime"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetSystemDateAndTimeResponse struct {
		XMLName           xml.Name `xml:"GetSystemDateAndTimeResponse"`
		SystemDateAndTime struct {
			DateTimeType    string `xml:"DateTimeType"`
			DaylightSavings bool   `xml:"DaylightSavings"`
			TimeZone        struct {
				TZ string `xml:"TZ"`
			} `xml:"TimeZone"`
			UTCDateTime struct {
				Time struct {
					Hour   int `xml:"Hour"`
					Minute int `xml:"Minute"`
					Second int `xml:"Second"`
				} `xml:"Time"`
				Date struct {
					Year  int `xml:"Year"`
					Month int `xml:"Month"`
					Day   int `xml:"Day"`
				} `xml:"Date"`
			} `xml:"UTCDateTime"`
			LocalDateTime struct {
				Time struct {
					Hour   int `xml:"Hour"`
					Minute int `xml:"Minute"`
					Second int `xml:"Second"`
				} `xml:"Time"`
				Date struct {
					Year  int `xml:"Year"`
					Month int `xml:"Month"`
					Day   int `xml:"Day"`
				} `xml:"Date"`
			} `xml:"LocalDateTime"`
		} `xml:"SystemDateAndTime"`
	}

	req := GetSystemDateAndTime{
		Xmlns: Namespace,
	}

	var resp GetSystemDateAndTimeResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSystemDateAndTime failed: %w", err)
	}

	return &SystemDateTime{
		DateTimeType:    SetDateTimeType(resp.SystemDateAndTime.DateTimeType),
		DaylightSavings: resp.SystemDateAndTime.DaylightSavings,
		TimeZone: &TimeZone{
			TZ: resp.SystemDateAndTime.TimeZone.TZ,
		},
		UTCDateTime: &DateTime{
			Time: Time{
				Hour:   resp.SystemDateAndTime.UTCDateTime.Time.Hour,
				Minute: resp.SystemDateAndTime.UTCDateTime.Time.Minute,
				Second: resp.SystemDateAndTime.UTCDateTime.Time.Second,
			},
			Date: Date{
				Year:  resp.SystemDateAndTime.UTCDateTime.Date.Year,
				Month: resp.SystemDateAndTime.UTCDateTime.Date.Month,
				Day:   resp.SystemDateAndTime.UTCDateTime.Date.Day,
			},
		},
		LocalDateTime: &DateTime{
			Time: Time{
				Hour:   resp.SystemDateAndTime.LocalDateTime.Time.Hour,
				Minute: resp.SystemDateAndTime.LocalDateTime.Time.Minute,
				Second: resp.SystemDateAndTime.LocalDateTime.Time.Second,
			},
			Date: Date{
				Year:  resp.SystemDateAndTime.LocalDateTime.Date.Year,
				Month: resp.SystemDateAndTime.LocalDateTime.Date.Month,
				Day:   resp.SystemDateAndTime.LocalDateTime.Date.Day,
			},
		},
	}, nil
}

// SetSystemDateAndTime sets the device system date and time. Admin-level
// write: a client constructed without credentials fails fast with
// ErrNoCredentials instead of a generic device rejection.
func (s *Service) SetSystemDateAndTime(ctx context.Context, dateTime *SystemDateTime) error {
	if err := s.requireCredentials(); err != nil {
		return fmt.Errorf("SetSystemDateAndTime: %w", err)
	}

	type SetSystemDateAndTime struct {
		XMLName         xml.Name `xml:"tds:SetSystemDateAndTime"`
		Xmlns           string   `xml:"xmlns:tds,attr"`
		DateTimeType    string   `xml:"tds:DateTimeType"`
		DaylightSavings bool     `xml:"tds:DaylightSavings"`
		TimeZone        *struct {
			TZ string `xml:"tds:TZ"`
		} `xml:"tds:TimeZone,omitempty"`
		UTCDateTime *struct {
			Time struct {
				Hour   int `xml:"tt:Hour"`
				Minute int `xml:"tt:Minute"`
				Second int `xml:"tt:Second"`
			} `xml:"tt:Time"`
			Date struct {
				Year  int `xml:"tt:Year"`
				Month int `xml:"tt:Month"`
				Day   int `xml:"tt:Day"`
			} `xml:"tt:Date"`
		} `xml:"tds:UTCDateTime,omitempty"`
	}

	req := SetSystemDateAndTime{
		Xmlns:           Namespace,
		DateTimeType:    string(dateTime.DateTimeType),
		DaylightSavings: dateTime.DaylightSavings,
	}

	if dateTime.TimeZone != nil {
		req.TimeZone = &struct {
			TZ string `xml:"tds:TZ"`
		}{
			TZ: dateTime.TimeZone.TZ,
		}
	}

	if dateTime.UTCDateTime != nil {
		req.UTCDateTime = &struct {
			Time struct {
				Hour   int `xml:"tt:Hour"`
				Minute int `xml:"tt:Minute"`
				Second int `xml:"tt:Second"`
			} `xml:"tt:Time"`
			Date struct {
				Year  int `xml:"tt:Year"`
				Month int `xml:"tt:Month"`
				Day   int `xml:"tt:Day"`
			} `xml:"tt:Date"`
		}{}
		req.UTCDateTime.Time.Hour = dateTime.UTCDateTime.Time.Hour
		req.UTCDateTime.Time.Minute = dateTime.UTCDateTime.Time.Minute
		req.UTCDateTime.Time.Second = dateTime.UTCDateTime.Time.Second
		req.UTCDateTime.Date.Year = dateTime.UTCDateTime.Date.Year
		req.UTCDateTime.Date.Month = dateTime.UTCDateTime.Date.Month
		req.UTCDateTime.Date.Day = dateTime.UTCDateTime.Date.Day
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetSystemDateAndTime failed: %w", err)
	}

	return nil
}

// AddScopes adds new configurable scope parameters to a device.
func (s *Service) AddScopes(ctx context.Context, scopeItems []string) error {
	type AddScopes struct {
		XMLName   xml.Name `xml:"tds:AddScopes"`
		Xmlns     string   `xml:"xmlns:tds,attr"`
		ScopeItem []string `xml:"tds:ScopeItem"`
	}

	req := AddScopes{
		Xmlns:     Namespace,
		ScopeItem: scopeItems,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("AddScopes failed: %w", err)
	}

	return nil
}

// RemoveScopes deletes scope-configurable scope parameters from a device.
func (s *Service) RemoveScopes(ctx context.Context, scopeItems []string) ([]string, error) {
	type RemoveScopes struct {
		XMLName   xml.Name `xml:"tds:RemoveScopes"`
		Xmlns     string   `xml:"xmlns:tds,attr"`
		ScopeItem []string `xml:"tds:ScopeItem"`
	}

	type RemoveScopesResponse struct {
		XMLName   xml.Name `xml:"RemoveScopesResponse"`
		ScopeItem []string `xml:"ScopeItem"`
	}

	req := RemoveScopes{
		Xmlns:     Namespace,
		ScopeItem: scopeItems,
	}

	var resp RemoveScopesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("RemoveScopes failed: %w", err)
	}

	return resp.ScopeItem, nil
}

// SetScopes sets the scope parameters of a device.
func (s *Service) SetScopes(ctx context.Context, scopes []string) error {
	type SetScopes struct {
		XMLName xml.Name `xml:"tds:SetScopes"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Scopes  []string `xml:"tds:Scopes"`
	}

	req := SetScopes{
		Xmlns:  Namespace,
		Scopes: scopes,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetScopes failed: %w", err)
	}

	return nil
}

func (s *Service) SendAuxiliaryCommand(ctx context.Context, command ptz.AuxiliaryData) (ptz.AuxiliaryData, error) {
	type SendAuxiliaryCommand struct {
		XMLName          xml.Name          `xml:"tds:SendAuxiliaryCommand"`
		Xmlns            string            `xml:"xmlns:tds,attr"`
		AuxiliaryCommand ptz.AuxiliaryData `xml:"tds:AuxiliaryCommand"`
	}

	type SendAuxiliaryCommandResponse struct {
		XMLName                  xml.Name          `xml:"SendAuxiliaryCommandResponse"`
		AuxiliaryCommandResponse ptz.AuxiliaryData `xml:"AuxiliaryCommandResponse"`
	}

	req := SendAuxiliaryCommand{
		Xmlns:            Namespace,
		AuxiliaryCommand: command,
	}

	var resp SendAuxiliaryCommandResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return "", fmt.Errorf("SendAuxiliaryCommand failed: %w", err)
	}

	return resp.AuxiliaryCommandResponse, nil
}

// GetSystemLog gets a system log from the device.
func (s *Service) GetSystemLog(ctx context.Context, logType SystemLogType) (*SystemLog, error) {
	type GetSystemLog struct {
		XMLName xml.Name      `xml:"tds:GetSystemLog"`
		Xmlns   string        `xml:"xmlns:tds,attr"`
		LogType SystemLogType `xml:"tds:LogType"`
	}

	type GetSystemLogResponse struct {
		XMLName   xml.Name `xml:"GetSystemLogResponse"`
		SystemLog struct {
			Binary *struct {
				ContentType string `xml:"contentType,attr"`
			} `xml:"Binary"`
			String string `xml:"String"`
		} `xml:"SystemLog"`
	}

	req := GetSystemLog{
		Xmlns:   Namespace,
		LogType: logType,
	}

	var resp GetSystemLogResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSystemLog failed: %w", err)
	}

	systemLog := &SystemLog{
		String: resp.SystemLog.String,
	}

	if resp.SystemLog.Binary != nil {
		systemLog.Binary = &AttachmentData{
			ContentType: resp.SystemLog.Binary.ContentType,
		}
	}

	return systemLog, nil
}

// GetSystemBackup retrieves system backup configuration files from a device.
func (s *Service) GetSystemBackup(ctx context.Context) ([]*BackupFile, error) {
	type GetSystemBackup struct {
		XMLName xml.Name `xml:"tds:GetSystemBackup"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetSystemBackupResponse struct {
		XMLName     xml.Name `xml:"GetSystemBackupResponse"`
		BackupFiles []struct {
			Name string `xml:"Name"`
			Data struct {
				ContentType string `xml:"contentType,attr"`
			} `xml:"Data"`
		} `xml:"BackupFiles"`
	}

	req := GetSystemBackup{
		Xmlns: Namespace,
	}

	var resp GetSystemBackupResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSystemBackup failed: %w", err)
	}

	backups := make([]*BackupFile, len(resp.BackupFiles))
	for i, file := range resp.BackupFiles {
		backups[i] = &BackupFile{
			Name: file.Name,
			Data: AttachmentData{
				ContentType: file.Data.ContentType,
			},
		}
	}

	return backups, nil
}

// RestoreSystem restores the system backup configuration files.
func (s *Service) RestoreSystem(ctx context.Context, backupFiles []*BackupFile) error {
	type RestoreSystem struct {
		XMLName     xml.Name `xml:"tds:RestoreSystem"`
		Xmlns       string   `xml:"xmlns:tds,attr"`
		BackupFiles []struct {
			Name string `xml:"tds:Name"`
			Data struct {
				ContentType string `xml:"contentType,attr"`
			} `xml:"tds:Data"`
		} `xml:"tds:BackupFiles"`
	}

	req := RestoreSystem{
		Xmlns: Namespace,
	}

	for _, file := range backupFiles {
		req.BackupFiles = append(req.BackupFiles, struct {
			Name string `xml:"tds:Name"`
			Data struct {
				ContentType string `xml:"contentType,attr"`
			} `xml:"tds:Data"`
		}{
			Name: file.Name,
			Data: struct {
				ContentType string `xml:"contentType,attr"`
			}{
				ContentType: file.Data.ContentType,
			},
		})
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("RestoreSystem failed: %w", err)
	}

	return nil
}

// GetSystemUris retrieves URIs from which system information may be downloaded.
func (s *Service) GetSystemUris(
	ctx context.Context,
) (uriList *SystemLogURIList, systemBackupURI, systemLogURI string, err error) {
	type GetSystemUris struct {
		XMLName xml.Name `xml:"tds:GetSystemUris"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetSystemUrisResponse struct {
		XMLName       xml.Name `xml:"GetSystemUrisResponse"`
		SystemLogUris *struct {
			SystemLog []struct {
				Type string `xml:"Type"`
				URI  string `xml:"Uri"`
			} `xml:"SystemLog"`
		} `xml:"SystemLogUris"`
		SupportInfoURI  string `xml:"SupportInfoUri"`
		SystemBackupURI string `xml:"SystemBackupUri"`
	}

	req := GetSystemUris{
		Xmlns: Namespace,
	}

	var resp GetSystemUrisResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, "", "", fmt.Errorf("GetSystemUris failed: %w", err)
	}

	var logUris *SystemLogURIList
	if resp.SystemLogUris != nil {
		logUris = &SystemLogURIList{}
		for _, log := range resp.SystemLogUris.SystemLog {
			logUris.SystemLog = append(logUris.SystemLog, SystemLogURI{
				Type: SystemLogType(log.Type),
				URI:  log.URI,
			})
		}
	}

	return logUris, resp.SupportInfoURI, resp.SystemBackupURI, nil
}

// GetSystemSupportInformation gets arbitrary device diagnostics information.
func (s *Service) GetSystemSupportInformation(ctx context.Context) (*SupportInformation, error) {
	type GetSystemSupportInformation struct {
		XMLName xml.Name `xml:"tds:GetSystemSupportInformation"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetSystemSupportInformationResponse struct {
		XMLName            xml.Name `xml:"GetSystemSupportInformationResponse"`
		SupportInformation struct {
			Binary *struct {
				ContentType string `xml:"contentType,attr"`
			} `xml:"Binary"`
			String string `xml:"String"`
		} `xml:"SupportInformation"`
	}

	req := GetSystemSupportInformation{
		Xmlns: Namespace,
	}

	var resp GetSystemSupportInformationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSystemSupportInformation failed: %w", err)
	}

	info := &SupportInformation{
		String: resp.SupportInformation.String,
	}

	if resp.SupportInformation.Binary != nil {
		info.Binary = &AttachmentData{
			ContentType: resp.SupportInformation.Binary.ContentType,
		}
	}

	return info, nil
}

// SetSystemFactoryDefault reloads the parameters on the device to their factory default values.
func (s *Service) SetSystemFactoryDefault(ctx context.Context, factoryDefault FactoryDefaultType) error {
	type SetSystemFactoryDefault struct {
		XMLName        xml.Name           `xml:"tds:SetSystemFactoryDefault"`
		Xmlns          string             `xml:"xmlns:tds,attr"`
		FactoryDefault FactoryDefaultType `xml:"tds:FactoryDefault"`
	}

	req := SetSystemFactoryDefault{
		Xmlns:          Namespace,
		FactoryDefault: factoryDefault,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetSystemFactoryDefault failed: %w", err)
	}

	return nil
}

// StartFirmwareUpgrade initiates a firmware upgrade using the HTTP POST mechanism.
func (s *Service) StartFirmwareUpgrade(
	ctx context.Context,
) (uploadURI, uploadDelay, expectedDownTime string, err error) {
	type StartFirmwareUpgrade struct {
		XMLName xml.Name `xml:"tds:StartFirmwareUpgrade"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type StartFirmwareUpgradeResponse struct {
		XMLName          xml.Name `xml:"StartFirmwareUpgradeResponse"`
		UploadURI        string   `xml:"UploadUri"`
		UploadDelay      string   `xml:"UploadDelay"`
		ExpectedDownTime string   `xml:"ExpectedDownTime"`
	}

	req := StartFirmwareUpgrade{
		Xmlns: Namespace,
	}

	var resp StartFirmwareUpgradeResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return "", "", "", fmt.Errorf("StartFirmwareUpgrade failed: %w", err)
	}

	return resp.UploadURI, resp.UploadDelay, resp.ExpectedDownTime, nil
}

// StartSystemRestore initiates a system restore from backed up configuration data.
func (s *Service) StartSystemRestore(ctx context.Context) (uploadURI, expectedDownTime string, err error) {
	type StartSystemRestore struct {
		XMLName xml.Name `xml:"tds:StartSystemRestore"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type StartSystemRestoreResponse struct {
		XMLName          xml.Name `xml:"StartSystemRestoreResponse"`
		UploadURI        string   `xml:"UploadUri"`
		ExpectedDownTime string   `xml:"ExpectedDownTime"`
	}

	req := StartSystemRestore{
		Xmlns: Namespace,
	}

	var resp StartSystemRestoreResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return "", "", fmt.Errorf("StartSystemRestore failed: %w", err)
	}

	return resp.UploadURI, resp.ExpectedDownTime, nil
}

// NetworkInterfaceConfig is the desired configuration for one interface in
// SetNetworkInterfaces.
type NetworkInterfaceConfig struct {
	// Enabled toggles the interface itself.
	Enabled bool
	// IPv4Enabled toggles the interface's IPv4 stack.
	IPv4Enabled bool
	// DHCP enables DHCP address acquisition. When false, ManualAddress and
	// ManualPrefixLength define the static address.
	DHCP bool
	// ManualAddress is the static IPv4 address (ignored when DHCP is true).
	ManualAddress string
	// ManualPrefixLength is the static address prefix length, e.g. 24 for a
	// /24 (ignored when DHCP is true).
	ManualPrefixLength int
}

// SetNetworkInterfaces applies network interface configuration for one
// interface token (the ONVIF operation is per-interface despite its plural
// name). NOTE: most devices apply interface changes only after a reboot —
// check the returned rebootNeeded flag and call SystemReboot accordingly.
func (s *Service) SetNetworkInterfaces(
	ctx context.Context,
	token string,
	cfg NetworkInterfaceConfig,
) (rebootNeeded bool, err error) {
	if token == "" {
		return false, fmt.Errorf("SetNetworkInterfaces: %w: interface token is empty", types.ErrInvalidParameter)
	}

	if !cfg.DHCP && cfg.ManualAddress == "" {
		return false, fmt.Errorf("SetNetworkInterfaces: %w: static configuration needs a manual address",
			types.ErrInvalidParameter)
	}

	type manualEntry struct {
		Address      string `xml:"tt:Address"`
		PrefixLength int    `xml:"tt:PrefixLength"`
	}

	type ipv4Config struct {
		Manual *manualEntry `xml:"tt:Manual,omitempty"`
		DHCP   bool         `xml:"tt:DHCP"`
	}

	type ipv4 struct {
		Enabled bool       `xml:"tt:Enabled"`
		Config  ipv4Config `xml:"tt:Config"`
	}

	type networkInterface struct {
		Enabled bool `xml:"tt:Enabled"`
		IPv4    ipv4 `xml:"tt:IPv4"`
	}

	type SetNetworkInterfaces struct {
		XMLName          xml.Name         `xml:"tds:SetNetworkInterfaces"`
		Xmlns            string           `xml:"xmlns:tds,attr"`
		Xmlnst           string           `xml:"xmlns:tt,attr"`
		InterfaceToken   string           `xml:"tds:InterfaceToken"`
		NetworkInterface networkInterface `xml:"tds:NetworkInterface"`
	}

	type SetNetworkInterfacesResponse struct {
		XMLName      xml.Name `xml:"SetNetworkInterfacesResponse"`
		RebootNeeded bool     `xml:"RebootNeeded"`
	}

	req := SetNetworkInterfaces{
		Xmlns:          Namespace,
		Xmlnst:         "http://www.onvif.org/ver10/schema",
		InterfaceToken: token,
		NetworkInterface: networkInterface{
			Enabled: cfg.Enabled,
			IPv4: ipv4{
				Enabled: cfg.IPv4Enabled,
				Config:  ipv4Config{DHCP: cfg.DHCP},
			},
		},
	}

	if !cfg.DHCP {
		req.NetworkInterface.IPv4.Config.Manual = &manualEntry{
			Address:      cfg.ManualAddress,
			PrefixLength: cfg.ManualPrefixLength,
		}
	}

	var resp SetNetworkInterfacesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return false, fmt.Errorf("SetNetworkInterfaces failed: %w", err)
	}

	return resp.RebootNeeded, nil
}

// NetmaskFromPrefixLength converts a prefix length (0-32) to a dotted netmask
// ("255.255.255.0" for 24). Returns "" for out-of-range prefixes. A
// convenience so every consumer does not reimplement the conversion.
func NetmaskFromPrefixLength(prefixLength int) string {
	if prefixLength < 0 || prefixLength > 32 {
		return ""
	}

	var mask uint32 = 0xFFFFFFFF
	mask <<= 32 - prefixLength

	if prefixLength == 0 {
		mask = 0
	}

	return fmt.Sprintf("%d.%d.%d.%d",
		byte(mask>>24), byte(mask>>16), byte(mask>>8), byte(mask))
}
