package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// GetDot11Capabilities retrieves 802.11 capabilities. ONVIF Specification: GetDot11Capabilities operation.
func (c *Client) GetDot11Capabilities(ctx context.Context) (*Dot11Capabilities, error) {
	type GetDot11CapabilitiesBody struct {
		XMLName xml.Name `xml:"tds:GetDot11Capabilities"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDot11CapabilitiesResponse struct {
		XMLName      xml.Name           `xml:"GetDot11CapabilitiesResponse"`
		Capabilities *Dot11Capabilities `xml:"Capabilities"`
	}

	request := GetDot11CapabilitiesBody{
		Xmlns: deviceNamespace,
	}
	var response GetDot11CapabilitiesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetDot11Capabilities failed: %w", err)
	}

	return response.Capabilities, nil
}

// GetDot11Status retrieves 802.11 status. ONVIF Specification: GetDot11Status operation.
func (c *Client) GetDot11Status(ctx context.Context, interfaceToken string) (*Dot11Status, error) {
	type GetDot11StatusBody struct {
		XMLName        xml.Name `xml:"tds:GetDot11Status"`
		Xmlns          string   `xml:"xmlns:tds,attr"`
		InterfaceToken string   `xml:"tds:InterfaceToken"`
	}

	type GetDot11StatusResponse struct {
		XMLName xml.Name     `xml:"GetDot11StatusResponse"`
		Status  *Dot11Status `xml:"Status"`
	}

	request := GetDot11StatusBody{
		Xmlns:          deviceNamespace,
		InterfaceToken: interfaceToken,
	}
	var response GetDot11StatusResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetDot11Status failed: %w", err)
	}

	return response.Status, nil
}

// GetDot1XConfiguration retrieves an 802.1X configuration. ONVIF Specification: GetDot1XConfiguration operation.
func (c *Client) GetDot1XConfiguration(ctx context.Context, configToken string) (*Dot1XConfiguration, error) {
	type GetDot1XConfigurationBody struct {
		XMLName                 xml.Name `xml:"tds:GetDot1XConfiguration"`
		Xmlns                   string   `xml:"xmlns:tds,attr"`
		Dot1XConfigurationToken string   `xml:"tds:Dot1XConfigurationToken"`
	}

	type GetDot1XConfigurationResponse struct {
		XMLName            xml.Name            `xml:"GetDot1XConfigurationResponse"`
		Dot1XConfiguration *Dot1XConfiguration `xml:"Dot1XConfiguration"`
	}

	request := GetDot1XConfigurationBody{
		Xmlns:                   deviceNamespace,
		Dot1XConfigurationToken: configToken,
	}
	var response GetDot1XConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetDot1XConfiguration failed: %w", err)
	}

	return response.Dot1XConfiguration, nil
}

// GetDot1XConfigurations retrieves all 802.1X configurations. ONVIF Specification: GetDot1XConfigurations operation.
func (c *Client) GetDot1XConfigurations(ctx context.Context) ([]*Dot1XConfiguration, error) {
	type GetDot1XConfigurationsBody struct {
		XMLName xml.Name `xml:"tds:GetDot1XConfigurations"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDot1XConfigurationsResponse struct {
		XMLName            xml.Name              `xml:"GetDot1XConfigurationsResponse"`
		Dot1XConfiguration []*Dot1XConfiguration `xml:"Dot1XConfiguration"`
	}

	request := GetDot1XConfigurationsBody{
		Xmlns: deviceNamespace,
	}
	var response GetDot1XConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetDot1XConfigurations failed: %w", err)
	}

	return response.Dot1XConfiguration, nil
}

// SetDot1XConfiguration sets an 802.1X configuration. ONVIF Specification: SetDot1XConfiguration operation.
func (c *Client) SetDot1XConfiguration(ctx context.Context, config *Dot1XConfiguration) error {
	type SetDot1XConfigurationBody struct {
		XMLName            xml.Name            `xml:"tds:SetDot1XConfiguration"`
		Xmlns              string              `xml:"xmlns:tds,attr"`
		Dot1XConfiguration *Dot1XConfiguration `xml:"tds:Dot1XConfiguration"`
	}

	type SetDot1XConfigurationResponse struct {
		XMLName xml.Name `xml:"SetDot1XConfigurationResponse"`
	}

	request := SetDot1XConfigurationBody{
		Xmlns:              deviceNamespace,
		Dot1XConfiguration: config,
	}
	var response SetDot1XConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("SetDot1XConfiguration failed: %w", err)
	}

	return nil
}

// CreateDot1XConfiguration creates an 802.1X configuration. ONVIF Specification: CreateDot1XConfiguration operation.
func (c *Client) CreateDot1XConfiguration(ctx context.Context, config *Dot1XConfiguration) error {
	type CreateDot1XConfigurationBody struct {
		XMLName            xml.Name            `xml:"tds:CreateDot1XConfiguration"`
		Xmlns              string              `xml:"xmlns:tds,attr"`
		Dot1XConfiguration *Dot1XConfiguration `xml:"tds:Dot1XConfiguration"`
	}

	type CreateDot1XConfigurationResponse struct {
		XMLName xml.Name `xml:"CreateDot1XConfigurationResponse"`
	}

	request := CreateDot1XConfigurationBody{
		Xmlns:              deviceNamespace,
		Dot1XConfiguration: config,
	}
	var response CreateDot1XConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("CreateDot1XConfiguration failed: %w", err)
	}

	return nil
}

// DeleteDot1XConfiguration deletes an 802.1X configuration. ONVIF Specification: DeleteDot1XConfiguration operation.
func (c *Client) DeleteDot1XConfiguration(ctx context.Context, configToken string) error {
	type DeleteDot1XConfigurationBody struct {
		XMLName                 xml.Name `xml:"tds:DeleteDot1XConfiguration"`
		Xmlns                   string   `xml:"xmlns:tds,attr"`
		Dot1XConfigurationToken string   `xml:"tds:Dot1XConfigurationToken"`
	}

	type DeleteDot1XConfigurationResponse struct {
		XMLName xml.Name `xml:"DeleteDot1XConfigurationResponse"`
	}

	request := DeleteDot1XConfigurationBody{
		Xmlns:                   deviceNamespace,
		Dot1XConfigurationToken: configToken,
	}
	var response DeleteDot1XConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("DeleteDot1XConfiguration failed: %w", err)
	}

	return nil
}

// ScanAvailableDot11Networks scans for available 802.11 networks.
// ONVIF Specification: ScanAvailableDot11Networks operation.
func (c *Client) ScanAvailableDot11Networks(
	ctx context.Context,
	interfaceToken string,
) ([]*Dot11AvailableNetworks, error) {
	type ScanAvailableDot11NetworksBody struct {
		XMLName        xml.Name `xml:"tds:ScanAvailableDot11Networks"`
		Xmlns          string   `xml:"xmlns:tds,attr"`
		InterfaceToken string   `xml:"tds:InterfaceToken"`
	}

	type ScanAvailableDot11NetworksResponse struct {
		XMLName  xml.Name                  `xml:"ScanAvailableDot11NetworksResponse"`
		Networks []*Dot11AvailableNetworks `xml:"Networks"`
	}

	request := ScanAvailableDot11NetworksBody{
		Xmlns:          deviceNamespace,
		InterfaceToken: interfaceToken,
	}
	var response ScanAvailableDot11NetworksResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("ScanAvailableDot11Networks failed: %w", err)
	}

	return response.Networks, nil
}
