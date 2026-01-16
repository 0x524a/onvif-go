package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// Common XML request/response types for device security operations.
// These are defined at package level to avoid repeated inline struct definitions.

// ipAddressFilterRequest is the common structure for IP address filter SOAP requests.
type ipAddressFilterRequest struct {
	Type        string                   `xml:"tds:Type"`
	IPv4Address []prefixedIPv4AddressXML `xml:"tds:IPv4Address,omitempty"`
	IPv6Address []prefixedIPv6AddressXML `xml:"tds:IPv6Address,omitempty"`
}

// prefixedIPv4AddressXML is the XML representation of a prefixed IPv4 address.
type prefixedIPv4AddressXML struct {
	Address      string `xml:"tds:Address"`
	PrefixLength int    `xml:"tds:PrefixLength"`
}

// prefixedIPv6AddressXML is the XML representation of a prefixed IPv6 address.
type prefixedIPv6AddressXML struct {
	Address      string `xml:"tds:Address"`
	PrefixLength int    `xml:"tds:PrefixLength"`
}

// buildIPAddressFilterRequest converts an IPAddressFilter to the XML request format.
// Pre-allocates slices for efficiency when the source length is known.
func buildIPAddressFilterRequest(filter *IPAddressFilter) ipAddressFilterRequest {
	req := ipAddressFilterRequest{
		Type: string(filter.Type),
	}

	// Pre-allocate slices with known capacity
	if len(filter.IPv4Address) > 0 {
		req.IPv4Address = make([]prefixedIPv4AddressXML, 0, len(filter.IPv4Address))
		for _, addr := range filter.IPv4Address {
			req.IPv4Address = append(req.IPv4Address, prefixedIPv4AddressXML{
				Address:      addr.Address,
				PrefixLength: addr.PrefixLength,
			})
		}
	}

	if len(filter.IPv6Address) > 0 {
		req.IPv6Address = make([]prefixedIPv6AddressXML, 0, len(filter.IPv6Address))
		for _, addr := range filter.IPv6Address {
			req.IPv6Address = append(req.IPv6Address, prefixedIPv6AddressXML{
				Address:      addr.Address,
				PrefixLength: addr.PrefixLength,
			})
		}
	}

	return req
}

// newSOAPClient creates a SOAP client with the current client credentials.
func (c *Client) newSOAPClient() *soap.Client {
	username, password := c.GetCredentials()
	return soap.NewClient(c.httpClient, username, password)
}

// GetRemoteUser returns the configured remote user.
func (c *Client) GetRemoteUser(ctx context.Context) (*RemoteUser, error) {
	type getRemoteUserRequest struct {
		XMLName xml.Name `xml:"tds:GetRemoteUser"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getRemoteUserResponse struct {
		XMLName    xml.Name `xml:"GetRemoteUserResponse"`
		RemoteUser *struct {
			Username           string `xml:"Username"`
			Password           string `xml:"Password"`
			UseDerivedPassword bool   `xml:"UseDerivedPassword"`
		} `xml:"RemoteUser"`
	}

	req := getRemoteUserRequest{
		Xmlns: deviceNamespace,
	}

	var resp getRemoteUserResponse
	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetRemoteUser failed: %w", err)
	}

	if resp.RemoteUser == nil {
		return nil, nil
	}

	return &RemoteUser{
		Username:           resp.RemoteUser.Username,
		Password:           resp.RemoteUser.Password,
		UseDerivedPassword: resp.RemoteUser.UseDerivedPassword,
	}, nil
}

// SetRemoteUser sets the remote user.
func (c *Client) SetRemoteUser(ctx context.Context, remoteUser *RemoteUser) error {
	type remoteUserXML struct {
		Username           string `xml:"tds:Username"`
		Password           string `xml:"tds:Password,omitempty"`
		UseDerivedPassword bool   `xml:"tds:UseDerivedPassword"`
	}

	type setRemoteUserRequest struct {
		XMLName    xml.Name       `xml:"tds:SetRemoteUser"`
		Xmlns      string         `xml:"xmlns:tds,attr"`
		RemoteUser *remoteUserXML `xml:"tds:RemoteUser,omitempty"`
	}

	req := setRemoteUserRequest{
		Xmlns: deviceNamespace,
	}

	if remoteUser != nil {
		req.RemoteUser = &remoteUserXML{
			Username:           remoteUser.Username,
			Password:           remoteUser.Password,
			UseDerivedPassword: remoteUser.UseDerivedPassword,
		}
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetRemoteUser failed: %w", err)
	}

	return nil
}

// GetIPAddressFilter gets the IP address filter settings from a device.
func (c *Client) GetIPAddressFilter(ctx context.Context) (*IPAddressFilter, error) {
	type getIPAddressFilterRequest struct {
		XMLName xml.Name `xml:"tds:GetIPAddressFilter"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type prefixedAddressXML struct {
		Address      string `xml:"Address"`
		PrefixLength int    `xml:"PrefixLength"`
	}

	type getIPAddressFilterResponse struct {
		XMLName         xml.Name `xml:"GetIPAddressFilterResponse"`
		IPAddressFilter struct {
			Type        string               `xml:"Type"`
			IPv4Address []prefixedAddressXML `xml:"IPv4Address"`
			IPv6Address []prefixedAddressXML `xml:"IPv6Address"`
		} `xml:"IPAddressFilter"`
	}

	req := getIPAddressFilterRequest{
		Xmlns: deviceNamespace,
	}

	var resp getIPAddressFilterResponse
	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetIPAddressFilter failed: %w", err)
	}

	filter := &IPAddressFilter{
		Type: IPAddressFilterType(resp.IPAddressFilter.Type),
	}

	// Pre-allocate slices with known capacity
	if len(resp.IPAddressFilter.IPv4Address) > 0 {
		filter.IPv4Address = make([]PrefixedIPv4Address, 0, len(resp.IPAddressFilter.IPv4Address))
		for _, addr := range resp.IPAddressFilter.IPv4Address {
			filter.IPv4Address = append(filter.IPv4Address, PrefixedIPv4Address{
				Address:      addr.Address,
				PrefixLength: addr.PrefixLength,
			})
		}
	}

	if len(resp.IPAddressFilter.IPv6Address) > 0 {
		filter.IPv6Address = make([]PrefixedIPv6Address, 0, len(resp.IPAddressFilter.IPv6Address))
		for _, addr := range resp.IPAddressFilter.IPv6Address {
			filter.IPv6Address = append(filter.IPv6Address, PrefixedIPv6Address{
				Address:      addr.Address,
				PrefixLength: addr.PrefixLength,
			})
		}
	}

	return filter, nil
}

// SetIPAddressFilter sets the IP address filter settings on a device.
func (c *Client) SetIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	type setIPAddressFilterRequest struct {
		XMLName         xml.Name               `xml:"tds:SetIPAddressFilter"`
		Xmlns           string                 `xml:"xmlns:tds,attr"`
		IPAddressFilter ipAddressFilterRequest `xml:"tds:IPAddressFilter"`
	}

	req := setIPAddressFilterRequest{
		Xmlns:           deviceNamespace,
		IPAddressFilter: buildIPAddressFilterRequest(filter),
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetIPAddressFilter failed: %w", err)
	}

	return nil
}

// AddIPAddressFilter adds an IP filter address to a device.
func (c *Client) AddIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	type addIPAddressFilterRequest struct {
		XMLName         xml.Name               `xml:"tds:AddIPAddressFilter"`
		Xmlns           string                 `xml:"xmlns:tds,attr"`
		IPAddressFilter ipAddressFilterRequest `xml:"tds:IPAddressFilter"`
	}

	req := addIPAddressFilterRequest{
		Xmlns:           deviceNamespace,
		IPAddressFilter: buildIPAddressFilterRequest(filter),
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddIPAddressFilter failed: %w", err)
	}

	return nil
}

// RemoveIPAddressFilter deletes an IP filter address from a device.
func (c *Client) RemoveIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	type removeIPAddressFilterRequest struct {
		XMLName         xml.Name               `xml:"tds:RemoveIPAddressFilter"`
		Xmlns           string                 `xml:"xmlns:tds,attr"`
		IPAddressFilter ipAddressFilterRequest `xml:"tds:IPAddressFilter"`
	}

	req := removeIPAddressFilterRequest{
		Xmlns:           deviceNamespace,
		IPAddressFilter: buildIPAddressFilterRequest(filter),
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveIPAddressFilter failed: %w", err)
	}

	return nil
}

// GetZeroConfiguration gets the zero-configuration from a device.
func (c *Client) GetZeroConfiguration(ctx context.Context) (*NetworkZeroConfiguration, error) {
	type getZeroConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetZeroConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getZeroConfigurationResponse struct {
		XMLName           xml.Name `xml:"GetZeroConfigurationResponse"`
		ZeroConfiguration struct {
			InterfaceToken string   `xml:"InterfaceToken"`
			Enabled        bool     `xml:"Enabled"`
			Addresses      []string `xml:"Addresses"`
		} `xml:"ZeroConfiguration"`
	}

	req := getZeroConfigurationRequest{
		Xmlns: deviceNamespace,
	}

	var resp getZeroConfigurationResponse
	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetZeroConfiguration failed: %w", err)
	}

	return &NetworkZeroConfiguration{
		InterfaceToken: resp.ZeroConfiguration.InterfaceToken,
		Enabled:        resp.ZeroConfiguration.Enabled,
		Addresses:      resp.ZeroConfiguration.Addresses,
	}, nil
}

// SetZeroConfiguration sets the zero-configuration.
func (c *Client) SetZeroConfiguration(ctx context.Context, interfaceToken string, enabled bool) error {
	type setZeroConfigurationRequest struct {
		XMLName        xml.Name `xml:"tds:SetZeroConfiguration"`
		Xmlns          string   `xml:"xmlns:tds,attr"`
		InterfaceToken string   `xml:"tds:InterfaceToken"`
		Enabled        bool     `xml:"tds:Enabled"`
	}

	req := setZeroConfigurationRequest{
		Xmlns:          deviceNamespace,
		InterfaceToken: interfaceToken,
		Enabled:        enabled,
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetZeroConfiguration failed: %w", err)
	}

	return nil
}

// GetDynamicDNS gets the dynamic DNS settings from a device.
func (c *Client) GetDynamicDNS(ctx context.Context) (*DynamicDNSInformation, error) {
	type getDynamicDNSRequest struct {
		XMLName xml.Name `xml:"tds:GetDynamicDNS"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getDynamicDNSResponse struct {
		XMLName               xml.Name `xml:"GetDynamicDNSResponse"`
		DynamicDNSInformation struct {
			Type string `xml:"Type"`
			Name string `xml:"Name"`
			TTL  string `xml:"TTL"`
		} `xml:"DynamicDNSInformation"`
	}

	req := getDynamicDNSRequest{
		Xmlns: deviceNamespace,
	}

	var resp getDynamicDNSResponse
	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDynamicDNS failed: %w", err)
	}

	return &DynamicDNSInformation{
		Type: DynamicDNSType(resp.DynamicDNSInformation.Type),
		Name: resp.DynamicDNSInformation.Name,
		// TTL would need duration parsing
	}, nil
}

// SetDynamicDNS sets the dynamic DNS settings on a device.
func (c *Client) SetDynamicDNS(ctx context.Context, dnsType DynamicDNSType, name string) error {
	type setDynamicDNSRequest struct {
		XMLName xml.Name       `xml:"tds:SetDynamicDNS"`
		Xmlns   string         `xml:"xmlns:tds,attr"`
		Type    DynamicDNSType `xml:"tds:Type"`
		Name    string         `xml:"tds:Name,omitempty"`
	}

	req := setDynamicDNSRequest{
		Xmlns: deviceNamespace,
		Type:  dnsType,
		Name:  name,
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetDynamicDNS failed: %w", err)
	}

	return nil
}

// GetPasswordComplexityConfiguration retrieves the current password complexity configuration settings.
func (c *Client) GetPasswordComplexityConfiguration(ctx context.Context) (*PasswordComplexityConfiguration, error) {
	type getPasswordComplexityConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetPasswordComplexityConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getPasswordComplexityConfigurationResponse struct {
		XMLName                   xml.Name `xml:"GetPasswordComplexityConfigurationResponse"`
		MinLen                    int      `xml:"MinLen"`
		Uppercase                 int      `xml:"Uppercase"`
		Number                    int      `xml:"Number"`
		SpecialChars              int      `xml:"SpecialChars"`
		BlockUsernameOccurrence   bool     `xml:"BlockUsernameOccurrence"`
		PolicyConfigurationLocked bool     `xml:"PolicyConfigurationLocked"`
	}

	req := getPasswordComplexityConfigurationRequest{
		Xmlns: deviceNamespace,
	}

	var resp getPasswordComplexityConfigurationResponse
	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetPasswordComplexityConfiguration failed: %w", err)
	}

	return &PasswordComplexityConfiguration{
		MinLen:                    resp.MinLen,
		Uppercase:                 resp.Uppercase,
		Number:                    resp.Number,
		SpecialChars:              resp.SpecialChars,
		BlockUsernameOccurrence:   resp.BlockUsernameOccurrence,
		PolicyConfigurationLocked: resp.PolicyConfigurationLocked,
	}, nil
}

// SetPasswordComplexityConfiguration allows setting of the password complexity configuration.
func (c *Client) SetPasswordComplexityConfiguration(
	ctx context.Context,
	config *PasswordComplexityConfiguration,
) error {
	type setPasswordComplexityConfigurationRequest struct {
		XMLName                   xml.Name `xml:"tds:SetPasswordComplexityConfiguration"`
		Xmlns                     string   `xml:"xmlns:tds,attr"`
		MinLen                    int      `xml:"tds:MinLen,omitempty"`
		Uppercase                 int      `xml:"tds:Uppercase,omitempty"`
		Number                    int      `xml:"tds:Number,omitempty"`
		SpecialChars              int      `xml:"tds:SpecialChars,omitempty"`
		BlockUsernameOccurrence   bool     `xml:"tds:BlockUsernameOccurrence,omitempty"`
		PolicyConfigurationLocked bool     `xml:"tds:PolicyConfigurationLocked,omitempty"`
	}

	req := setPasswordComplexityConfigurationRequest{
		Xmlns:                     deviceNamespace,
		MinLen:                    config.MinLen,
		Uppercase:                 config.Uppercase,
		Number:                    config.Number,
		SpecialChars:              config.SpecialChars,
		BlockUsernameOccurrence:   config.BlockUsernameOccurrence,
		PolicyConfigurationLocked: config.PolicyConfigurationLocked,
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetPasswordComplexityConfiguration failed: %w", err)
	}

	return nil
}

// GetPasswordHistoryConfiguration retrieves the current password history configuration settings.
func (c *Client) GetPasswordHistoryConfiguration(ctx context.Context) (*PasswordHistoryConfiguration, error) {
	type getPasswordHistoryConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetPasswordHistoryConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getPasswordHistoryConfigurationResponse struct {
		XMLName xml.Name `xml:"GetPasswordHistoryConfigurationResponse"`
		Enabled bool     `xml:"Enabled"`
		Length  int      `xml:"Length"`
	}

	req := getPasswordHistoryConfigurationRequest{
		Xmlns: deviceNamespace,
	}

	var resp getPasswordHistoryConfigurationResponse
	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetPasswordHistoryConfiguration failed: %w", err)
	}

	return &PasswordHistoryConfiguration{
		Enabled: resp.Enabled,
		Length:  resp.Length,
	}, nil
}

// SetPasswordHistoryConfiguration allows setting of the password history configuration.
func (c *Client) SetPasswordHistoryConfiguration(ctx context.Context, config *PasswordHistoryConfiguration) error {
	type setPasswordHistoryConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:SetPasswordHistoryConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Enabled bool     `xml:"tds:Enabled"`
		Length  int      `xml:"tds:Length"`
	}

	req := setPasswordHistoryConfigurationRequest{
		Xmlns:   deviceNamespace,
		Enabled: config.Enabled,
		Length:  config.Length,
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetPasswordHistoryConfiguration failed: %w", err)
	}

	return nil
}

// GetAuthFailureWarningConfiguration retrieves the current authentication failure warning configuration.
func (c *Client) GetAuthFailureWarningConfiguration(ctx context.Context) (*AuthFailureWarningConfiguration, error) {
	type getAuthFailureWarningConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetAuthFailureWarningConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getAuthFailureWarningConfigurationResponse struct {
		XMLName         xml.Name `xml:"GetAuthFailureWarningConfigurationResponse"`
		Enabled         bool     `xml:"Enabled"`
		MonitorPeriod   int      `xml:"MonitorPeriod"`
		MaxAuthFailures int      `xml:"MaxAuthFailures"`
	}

	req := getAuthFailureWarningConfigurationRequest{
		Xmlns: deviceNamespace,
	}

	var resp getAuthFailureWarningConfigurationResponse
	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAuthFailureWarningConfiguration failed: %w", err)
	}

	return &AuthFailureWarningConfiguration{
		Enabled:         resp.Enabled,
		MonitorPeriod:   resp.MonitorPeriod,
		MaxAuthFailures: resp.MaxAuthFailures,
	}, nil
}

// SetAuthFailureWarningConfiguration allows setting of the authentication failure warning configuration.
func (c *Client) SetAuthFailureWarningConfiguration(
	ctx context.Context,
	config *AuthFailureWarningConfiguration,
) error {
	type setAuthFailureWarningConfigurationRequest struct {
		XMLName         xml.Name `xml:"tds:SetAuthFailureWarningConfiguration"`
		Xmlns           string   `xml:"xmlns:tds,attr"`
		Enabled         bool     `xml:"tds:Enabled"`
		MonitorPeriod   int      `xml:"tds:MonitorPeriod"`
		MaxAuthFailures int      `xml:"tds:MaxAuthFailures"`
	}

	req := setAuthFailureWarningConfigurationRequest{
		Xmlns:           deviceNamespace,
		Enabled:         config.Enabled,
		MonitorPeriod:   config.MonitorPeriod,
		MaxAuthFailures: config.MaxAuthFailures,
	}

	if err := c.newSOAPClient().Call(ctx, c.endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAuthFailureWarningConfiguration failed: %w", err)
	}

	return nil
}
