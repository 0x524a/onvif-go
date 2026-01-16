package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// GetGeoLocation retrieves geographic location information. ONVIF Specification: GetGeoLocation operation.
func (c *Client) GetGeoLocation(ctx context.Context) ([]LocationEntity, error) {
	type GetGeoLocationBody struct {
		XMLName xml.Name `xml:"tds:GetGeoLocation"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetGeoLocationResponse struct {
		XMLName  xml.Name         `xml:"GetGeoLocationResponse"`
		Location []LocationEntity `xml:"Location"`
	}

	request := GetGeoLocationBody{
		Xmlns: deviceNamespace,
	}
	var response GetGeoLocationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetGeoLocation failed: %w", err)
	}

	return response.Location, nil
}

// SetGeoLocation sets geographic location information. ONVIF Specification: SetGeoLocation operation.
func (c *Client) SetGeoLocation(ctx context.Context, location []LocationEntity) error {
	type SetGeoLocationBody struct {
		XMLName  xml.Name         `xml:"tds:SetGeoLocation"`
		Xmlns    string           `xml:"xmlns:tds,attr"`
		Location []LocationEntity `xml:"tds:Location"`
	}

	type SetGeoLocationResponse struct {
		XMLName xml.Name `xml:"SetGeoLocationResponse"`
	}

	request := SetGeoLocationBody{
		Xmlns:    deviceNamespace,
		Location: location,
	}
	var response SetGeoLocationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("SetGeoLocation failed: %w", err)
	}

	return nil
}

// DeleteGeoLocation deletes geographic location information. ONVIF Specification: DeleteGeoLocation operation.
func (c *Client) DeleteGeoLocation(ctx context.Context, location []LocationEntity) error {
	type DeleteGeoLocationBody struct {
		XMLName  xml.Name         `xml:"tds:DeleteGeoLocation"`
		Xmlns    string           `xml:"xmlns:tds,attr"`
		Location []LocationEntity `xml:"tds:Location"`
	}

	type DeleteGeoLocationResponse struct {
		XMLName xml.Name `xml:"DeleteGeoLocationResponse"`
	}

	request := DeleteGeoLocationBody{
		Xmlns:    deviceNamespace,
		Location: location,
	}
	var response DeleteGeoLocationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("DeleteGeoLocation failed: %w", err)
	}

	return nil
}

// GetDPAddresses retrieves DP (Device Provisioning) addresses. ONVIF Specification: GetDPAddresses operation.
func (c *Client) GetDPAddresses(ctx context.Context) ([]NetworkHost, error) {
	type GetDPAddressesBody struct {
		XMLName xml.Name `xml:"tds:GetDPAddresses"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetDPAddressesResponse struct {
		XMLName   xml.Name      `xml:"GetDPAddressesResponse"`
		DPAddress []NetworkHost `xml:"DPAddress"`
	}

	request := GetDPAddressesBody{
		Xmlns: deviceNamespace,
	}
	var response GetDPAddressesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetDPAddresses failed: %w", err)
	}

	return response.DPAddress, nil
}

// SetDPAddresses sets DP (Device Provisioning) addresses. ONVIF Specification: SetDPAddresses operation.
func (c *Client) SetDPAddresses(ctx context.Context, dpAddress []NetworkHost) error {
	type SetDPAddressesBody struct {
		XMLName   xml.Name      `xml:"tds:SetDPAddresses"`
		Xmlns     string        `xml:"xmlns:tds,attr"`
		DPAddress []NetworkHost `xml:"tds:DPAddress"`
	}

	type SetDPAddressesResponse struct {
		XMLName xml.Name `xml:"SetDPAddressesResponse"`
	}

	request := SetDPAddressesBody{
		Xmlns:     deviceNamespace,
		DPAddress: dpAddress,
	}
	var response SetDPAddressesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("SetDPAddresses failed: %w", err)
	}

	return nil
}

// GetAccessPolicy retrieves access policy information. ONVIF Specification: GetAccessPolicy operation.
func (c *Client) GetAccessPolicy(ctx context.Context) (*AccessPolicy, error) {
	type GetAccessPolicyBody struct {
		XMLName xml.Name `xml:"tds:GetAccessPolicy"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetAccessPolicyResponse struct {
		XMLName    xml.Name    `xml:"GetAccessPolicyResponse"`
		PolicyFile *BinaryData `xml:"PolicyFile"`
	}

	request := GetAccessPolicyBody{
		Xmlns: deviceNamespace,
	}
	var response GetAccessPolicyResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetAccessPolicy failed: %w", err)
	}

	return &AccessPolicy{PolicyFile: response.PolicyFile}, nil
}

// SetAccessPolicy sets access policy information. ONVIF Specification: SetAccessPolicy operation.
func (c *Client) SetAccessPolicy(ctx context.Context, policy *AccessPolicy) error {
	type SetAccessPolicyBody struct {
		XMLName    xml.Name    `xml:"tds:SetAccessPolicy"`
		Xmlns      string      `xml:"xmlns:tds,attr"`
		PolicyFile *BinaryData `xml:"tds:PolicyFile"`
	}

	type SetAccessPolicyResponse struct {
		XMLName xml.Name `xml:"SetAccessPolicyResponse"`
	}

	request := SetAccessPolicyBody{
		Xmlns:      deviceNamespace,
		PolicyFile: policy.PolicyFile,
	}
	var response SetAccessPolicyResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("SetAccessPolicy failed: %w", err)
	}

	return nil
}

// GetWsdlURL retrieves the WSDL URL (deprecated). ONVIF Specification: GetWsdlUrl operation.
func (c *Client) GetWsdlURL(ctx context.Context) (string, error) {
	type GetWsdlURLBody struct {
		XMLName xml.Name `xml:"tds:GetWsdlUrl"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetWsdlURLResponse struct {
		XMLName xml.Name `xml:"GetWsdlUrlResponse"`
		WsdlURL string   `xml:"WsdlUrl"`
	}

	request := GetWsdlURLBody{
		Xmlns: deviceNamespace,
	}
	var response GetWsdlURLResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return "", fmt.Errorf("GetWsdlURL failed: %w", err)
	}

	return response.WsdlURL, nil
}
