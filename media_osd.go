package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// GetOSDs retrieves all OSD configurations.
func (c *Client) GetOSDs(ctx context.Context, configurationToken string) ([]*OSDConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetOSDs struct {
		XMLName            xml.Name `xml:"trt:GetOSDs"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetOSDsResponse struct {
		XMLName xml.Name `xml:"GetOSDsResponse"`
		OSDs    []struct {
			Token string `xml:"token,attr"`
		} `xml:"OSDs"`
	}

	req := GetOSDs{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSDs failed: %w", err)
	}

	osds := make([]*OSDConfiguration, len(resp.OSDs))
	for i, o := range resp.OSDs {
		osds[i] = &OSDConfiguration{
			Token: o.Token,
		}
	}

	return osds, nil
}

// GetOSD retrieves a specific OSD configuration.
func (c *Client) GetOSD(ctx context.Context, osdToken string) (*OSDConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetOSD struct {
		XMLName  xml.Name `xml:"trt:GetOSD"`
		Xmlns    string   `xml:"xmlns:trt,attr"`
		OSDToken string   `xml:"trt:OSDToken"`
	}

	type GetOSDResponse struct {
		XMLName xml.Name `xml:"GetOSDResponse"`
		OSD     struct {
			Token string `xml:"token,attr"`
		} `xml:"OSD"`
	}

	req := GetOSD{
		Xmlns:    mediaNamespace,
		OSDToken: osdToken,
	}

	var resp GetOSDResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

// SetOSD sets OSD configuration.
func (c *Client) SetOSD(ctx context.Context, osd *OSDConfiguration) error {
	endpoint := c.getMediaEndpoint()

	type SetOSD struct {
		XMLName xml.Name `xml:"trt:SetOSD"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
		Xmlnst  string   `xml:"xmlns:tt,attr"`
		OSD     struct {
			Token string `xml:"token,attr"`
		} `xml:"trt:OSD"`
	}

	req := SetOSD{
		Xmlns:  mediaNamespace,
		Xmlnst: onvifSchemaNamespace,
	}
	req.OSD.Token = osd.Token

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetOSD failed: %w", err)
	}

	return nil
}

// CreateOSD creates a new OSD configuration.
func (c *Client) CreateOSD(
	ctx context.Context,
	videoSourceConfigurationToken string,
	osd *OSDConfiguration,
) (*OSDConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type CreateOSD struct {
		XMLName                       xml.Name `xml:"trt:CreateOSD"`
		Xmlns                         string   `xml:"xmlns:trt,attr"`
		Xmlnst                        string   `xml:"xmlns:tt,attr"`
		VideoSourceConfigurationToken string   `xml:"trt:VideoSourceConfigurationToken"`
		OSD                           struct {
			Token string `xml:"token,attr,omitempty"`
		} `xml:"trt:OSD"`
	}

	type CreateOSDResponse struct {
		XMLName xml.Name `xml:"CreateOSDResponse"`
		OSD     struct {
			Token string `xml:"token,attr"`
		} `xml:"OSD"`
	}

	req := CreateOSD{
		Xmlns:                         mediaNamespace,
		Xmlnst:                        onvifSchemaNamespace,
		VideoSourceConfigurationToken: videoSourceConfigurationToken,
	}
	if osd != nil && osd.Token != "" {
		req.OSD.Token = osd.Token
	}

	var resp CreateOSDResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreateOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

// DeleteOSD deletes an OSD configuration.
func (c *Client) DeleteOSD(ctx context.Context, osdToken string) error {
	endpoint := c.getMediaEndpoint()

	type DeleteOSD struct {
		XMLName  xml.Name `xml:"trt:DeleteOSD"`
		Xmlns    string   `xml:"xmlns:trt,attr"`
		OSDToken string   `xml:"trt:OSDToken"`
	}

	req := DeleteOSD{
		Xmlns:    mediaNamespace,
		OSDToken: osdToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteOSD failed: %w", err)
	}

	return nil
}

// GetOSDOptions retrieves available options for OSD configuration.
func (c *Client) GetOSDOptions(ctx context.Context, configurationToken string) (*OSDConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetOSDOptions struct {
		XMLName            xml.Name `xml:"trt:GetOSDOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetOSDOptionsResponse struct {
		XMLName xml.Name `xml:"GetOSDOptionsResponse"`
		Options struct {
			MaximumNumberOfOSDs int `xml:"MaximumNumberOfOSDs"`
		} `xml:"Options"`
	}

	req := GetOSDOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSDOptions failed: %w", err)
	}

	return &OSDConfigurationOptions{
		MaximumNumberOfOSDs: resp.Options.MaximumNumberOfOSDs,
	}, nil
}
