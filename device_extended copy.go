package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// SetDNS sets the DNS settings on a device.
func (c *Client) SetDNS(ctx context.Context, fromDHCP bool, searchDomain []string, dnsManual []IPAddress) error {
	type SetDNS struct {
		XMLName      xml.Name `xml:"tds:SetDNS"`
		Xmlns        string   `xml:"xmlns:tds,attr"`
		FromDHCP     bool     `xml:"tds:FromDHCP"`
		SearchDomain []string `xml:"tds:SearchDomain,omitempty"`
		DNSManual    []struct {
			Type        string `xml:"tds:Type"`
			IPv4Address string `xml:"tds:IPv4Address,omitempty"`
			IPv6Address string `xml:"tds:IPv6Address,omitempty"`
		} `xml:"tds:DNSManual,omitempty"`
	}

	req := SetDNS{
		Xmlns:        deviceNamespace,
		FromDHCP:     fromDHCP,
		SearchDomain: searchDomain,
	}

	for _, dns := range dnsManual {
		req.DNSManual = append(req.DNSManual, struct {
			Type        string `xml:"tds:Type"`
			IPv4Address string `xml:"tds:IPv4Address,omitempty"`
			IPv6Address string `xml:"tds:IPv6Address,omitempty"`
		}{
			Type:        dns.Type,
			IPv4Address: dns.IPv4Address,
			IPv6Address: dns.IPv6Address,
		})
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetDNS failed: %w", err)
	}

	return nil
}

// SetNTP sets the NTP settings on a device.
func (c *Client) SetNTP(ctx context.Context, fromDHCP bool, ntpManual []NetworkHost) error {
	type SetNTP struct {
		XMLName   xml.Name `xml:"tds:SetNTP"`
		Xmlns     string   `xml:"xmlns:tds,attr"`
		FromDHCP  bool     `xml:"tds:FromDHCP"`
		NTPManual []struct {
			Type        string `xml:"tds:Type"`
			IPv4Address string `xml:"tds:IPv4Address,omitempty"`
			IPv6Address string `xml:"tds:IPv6Address,omitempty"`
			DNSname     string `xml:"tds:DNSname,omitempty"`
		} `xml:"tds:NTPManual,omitempty"`
	}

	req := SetNTP{
		Xmlns:    deviceNamespace,
		FromDHCP: fromDHCP,
	}

	for _, ntp := range ntpManual {
		req.NTPManual = append(req.NTPManual, struct {
			Type        string `xml:"tds:Type"`
			IPv4Address string `xml:"tds:IPv4Address,omitempty"`
			IPv6Address string `xml:"tds:IPv6Address,omitempty"`
			DNSname     string `xml:"tds:DNSname,omitempty"`
		}{
			Type:        ntp.Type,
			IPv4Address: ntp.IPv4Address,
			IPv6Address: ntp.IPv6Address,
			DNSname:     ntp.DNSname,
		})
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetNTP failed: %w", err)
	}

	return nil
}

// SetHostnameFromDHCP controls whether the hostname is set manually or retrieved via DHCP.
func (c *Client) SetHostnameFromDHCP(ctx context.Context, fromDHCP bool) (bool, error) {
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
		Xmlns:    deviceNamespace,
		FromDHCP: fromDHCP,
	}

	var resp SetHostnameFromDHCPResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return false, fmt.Errorf("SetHostnameFromDHCP failed: %w", err)
	}

	return resp.RebootNeeded, nil
}

// FixedGetSystemDateAndTime retrieves the device's system date and time with proper typing.
func (c *Client) FixedGetSystemDateAndTime(ctx context.Context) (*SystemDateTime, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetSystemDateAndTimeResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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

// SetSystemDateAndTime sets the device system date and time.
func (c *Client) SetSystemDateAndTime(ctx context.Context, dateTime *SystemDateTime) error {
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
		Xmlns:           deviceNamespace,
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

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetSystemDateAndTime failed: %w", err)
	}

	return nil
}

// AddScopes adds new configurable scope parameters to a device.
func (c *Client) AddScopes(ctx context.Context, scopeItems []string) error {
	type AddScopes struct {
		XMLName   xml.Name `xml:"tds:AddScopes"`
		Xmlns     string   `xml:"xmlns:tds,attr"`
		ScopeItem []string `xml:"tds:ScopeItem"`
	}

	req := AddScopes{
		Xmlns:     deviceNamespace,
		ScopeItem: scopeItems,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddScopes failed: %w", err)
	}

	return nil
}

// RemoveScopes deletes scope-configurable scope parameters from a device.
func (c *Client) RemoveScopes(ctx context.Context, scopeItems []string) ([]string, error) {
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
		Xmlns:     deviceNamespace,
		ScopeItem: scopeItems,
	}

	var resp RemoveScopesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("RemoveScopes failed: %w", err)
	}

	return resp.ScopeItem, nil
}

// SetScopes sets the scope parameters of a device.
func (c *Client) SetScopes(ctx context.Context, scopes []string) error {
	type SetScopes struct {
		XMLName xml.Name `xml:"tds:SetScopes"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Scopes  []string `xml:"tds:Scopes"`
	}

	req := SetScopes{
		Xmlns:  deviceNamespace,
		Scopes: scopes,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetScopes failed: %w", err)
	}

	return nil
}

// GetRelayOutputs gets a list of all available relay outputs and their settings.
func (c *Client) GetRelayOutputs(ctx context.Context) ([]*RelayOutput, error) {
	type GetRelayOutputs struct {
		XMLName xml.Name `xml:"tds:GetRelayOutputs"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetRelayOutputsResponse struct {
		XMLName      xml.Name `xml:"GetRelayOutputsResponse"`
		RelayOutputs []struct {
			Token      string `xml:"token,attr"`
			Properties struct {
				Mode      string `xml:"Mode"`
				DelayTime string `xml:"DelayTime"`
				IdleState string `xml:"IdleState"`
			} `xml:"Properties"`
		} `xml:"RelayOutputs"`
	}

	req := GetRelayOutputs{
		Xmlns: deviceNamespace,
	}

	var resp GetRelayOutputsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetRelayOutputs failed: %w", err)
	}

	relays := make([]*RelayOutput, len(resp.RelayOutputs))
	for i, relay := range resp.RelayOutputs {
		relays[i] = &RelayOutput{
			Token: relay.Token,
			Properties: RelayOutputSettings{
				Mode:      RelayMode(relay.Properties.Mode),
				IdleState: RelayIdleState(relay.Properties.IdleState),
				// DelayTime parsing would require duration parsing
			},
		}
	}

	return relays, nil
}

// SetRelayOutputSettings sets the settings of a relay output.
func (c *Client) SetRelayOutputSettings(ctx context.Context, token string, settings *RelayOutputSettings) error {
	type SetRelayOutputSettings struct {
		XMLName          xml.Name `xml:"tds:SetRelayOutputSettings"`
		Xmlns            string   `xml:"xmlns:tds,attr"`
		RelayOutputToken string   `xml:"tds:RelayOutputToken"`
		Properties       struct {
			Mode      string `xml:"tt:Mode"`
			DelayTime string `xml:"tt:DelayTime"`
			IdleState string `xml:"tt:IdleState"`
		} `xml:"tds:Properties"`
	}

	req := SetRelayOutputSettings{
		Xmlns:            deviceNamespace,
		RelayOutputToken: token,
	}
	req.Properties.Mode = string(settings.Mode)
	req.Properties.IdleState = string(settings.IdleState)
	// DelayTime would need duration formatting

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetRelayOutputSettings failed: %w", err)
	}

	return nil
}

// SetRelayOutputState sets the state of a relay output.
func (c *Client) SetRelayOutputState(ctx context.Context, token string, state RelayLogicalState) error {
	type SetRelayOutputState struct {
		XMLName          xml.Name          `xml:"tds:SetRelayOutputState"`
		Xmlns            string            `xml:"xmlns:tds,attr"`
		RelayOutputToken string            `xml:"tds:RelayOutputToken"`
		LogicalState     RelayLogicalState `xml:"tds:LogicalState"`
	}

	req := SetRelayOutputState{
		Xmlns:            deviceNamespace,
		RelayOutputToken: token,
		LogicalState:     state,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetRelayOutputState failed: %w", err)
	}

	return nil
}

// SendAuxiliaryCommand sends an auxiliary command to the device.
func (c *Client) SendAuxiliaryCommand(ctx context.Context, command AuxiliaryData) (AuxiliaryData, error) {
	type SendAuxiliaryCommand struct {
		XMLName          xml.Name      `xml:"tds:SendAuxiliaryCommand"`
		Xmlns            string        `xml:"xmlns:tds,attr"`
		AuxiliaryCommand AuxiliaryData `xml:"tds:AuxiliaryCommand"`
	}

	type SendAuxiliaryCommandResponse struct {
		XMLName                  xml.Name      `xml:"SendAuxiliaryCommandResponse"`
		AuxiliaryCommandResponse AuxiliaryData `xml:"AuxiliaryCommandResponse"`
	}

	req := SendAuxiliaryCommand{
		Xmlns:            deviceNamespace,
		AuxiliaryCommand: command,
	}

	var resp SendAuxiliaryCommandResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return "", fmt.Errorf("SendAuxiliaryCommand failed: %w", err)
	}

	return resp.AuxiliaryCommandResponse, nil
}

// GetSystemLog gets a system log from the device.
func (c *Client) GetSystemLog(ctx context.Context, logType SystemLogType) (*SystemLog, error) {
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
		Xmlns:   deviceNamespace,
		LogType: logType,
	}

	var resp GetSystemLogResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) GetSystemBackup(ctx context.Context) ([]*BackupFile, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetSystemBackupResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) RestoreSystem(ctx context.Context, backupFiles []*BackupFile) error {
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
		Xmlns: deviceNamespace,
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

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RestoreSystem failed: %w", err)
	}

	return nil
}

// GetSystemUris retrieves URIs from which system information may be downloaded.
func (c *Client) GetSystemUris(
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
		Xmlns: deviceNamespace,
	}

	var resp GetSystemUrisResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) GetSystemSupportInformation(ctx context.Context) (*SupportInformation, error) {
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
		Xmlns: deviceNamespace,
	}

	var resp GetSystemSupportInformationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
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
func (c *Client) SetSystemFactoryDefault(ctx context.Context, factoryDefault FactoryDefaultType) error {
	type SetSystemFactoryDefault struct {
		XMLName        xml.Name           `xml:"tds:SetSystemFactoryDefault"`
		Xmlns          string             `xml:"xmlns:tds,attr"`
		FactoryDefault FactoryDefaultType `xml:"tds:FactoryDefault"`
	}

	req := SetSystemFactoryDefault{
		Xmlns:          deviceNamespace,
		FactoryDefault: factoryDefault,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetSystemFactoryDefault failed: %w", err)
	}

	return nil
}

// StartFirmwareUpgrade initiates a firmware upgrade using the HTTP POST mechanism.
func (c *Client) StartFirmwareUpgrade(
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
		Xmlns: deviceNamespace,
	}

	var resp StartFirmwareUpgradeResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return "", "", "", fmt.Errorf("StartFirmwareUpgrade failed: %w", err)
	}

	return resp.UploadURI, resp.UploadDelay, resp.ExpectedDownTime, nil
}

// StartSystemRestore initiates a system restore from backed up configuration data.
func (c *Client) StartSystemRestore(ctx context.Context) (uploadURI, expectedDownTime string, err error) {
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
		Xmlns: deviceNamespace,
	}

	var resp StartSystemRestoreResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return "", "", fmt.Errorf("StartSystemRestore failed: %w", err)
	}

	return resp.UploadURI, resp.ExpectedDownTime, nil
}
