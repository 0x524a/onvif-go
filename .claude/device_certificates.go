package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// GetCertificates retrieves certificates. ONVIF Specification: GetCertificates operation.
func (c *Client) GetCertificates(ctx context.Context) ([]*Certificate, error) {
	type GetCertificatesBody struct {
		XMLName xml.Name `xml:"tds:GetCertificates"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetCertificatesResponse struct {
		XMLName      xml.Name       `xml:"GetCertificatesResponse"`
		Certificates []*Certificate `xml:"Certificate"`
	}

	request := GetCertificatesBody{
		Xmlns: deviceNamespace,
	}
	var response GetCertificatesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCertificates failed: %w", err)
	}

	return response.Certificates, nil
}

// GetCACertificates retrieves CA certificates. ONVIF Specification: GetCACertificates operation.
func (c *Client) GetCACertificates(ctx context.Context) ([]*Certificate, error) {
	type GetCACertificatesBody struct {
		XMLName xml.Name `xml:"tds:GetCACertificates"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetCACertificatesResponse struct {
		XMLName      xml.Name       `xml:"GetCACertificatesResponse"`
		Certificates []*Certificate `xml:"Certificate"`
	}

	request := GetCACertificatesBody{
		Xmlns: deviceNamespace,
	}
	var response GetCACertificatesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCACertificates failed: %w", err)
	}

	return response.Certificates, nil
}

// LoadCertificates loads certificates. ONVIF Specification: LoadCertificates operation.
func (c *Client) LoadCertificates(ctx context.Context, certificates []*Certificate) error {
	type LoadCertificatesBody struct {
		XMLName     xml.Name       `xml:"tds:LoadCertificates"`
		Xmlns       string         `xml:"xmlns:tds,attr"`
		Certificate []*Certificate `xml:"tds:Certificate"`
	}

	type LoadCertificatesResponse struct {
		XMLName xml.Name `xml:"LoadCertificatesResponse"`
	}

	request := LoadCertificatesBody{
		Xmlns:       deviceNamespace,
		Certificate: certificates,
	}
	var response LoadCertificatesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("LoadCertificates failed: %w", err)
	}

	return nil
}

// LoadCACertificates loads CA certificates. ONVIF Specification: LoadCACertificates operation.
func (c *Client) LoadCACertificates(ctx context.Context, certificates []*Certificate) error {
	type LoadCACertificatesBody struct {
		XMLName     xml.Name       `xml:"tds:LoadCACertificates"`
		Xmlns       string         `xml:"xmlns:tds,attr"`
		Certificate []*Certificate `xml:"tds:Certificate"`
	}

	type LoadCACertificatesResponse struct {
		XMLName xml.Name `xml:"LoadCACertificatesResponse"`
	}

	request := LoadCACertificatesBody{
		Xmlns:       deviceNamespace,
		Certificate: certificates,
	}
	var response LoadCACertificatesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("LoadCACertificates failed: %w", err)
	}

	return nil
}

// CreateCertificate creates a certificate. ONVIF Specification: CreateCertificate operation.
func (c *Client) CreateCertificate(
	ctx context.Context,
	certificateID, subject, validNotBefore, validNotAfter string,
) (*Certificate, error) {
	type CreateCertificateBody struct {
		XMLName        xml.Name `xml:"tds:CreateCertificate"`
		Xmlns          string   `xml:"xmlns:tds,attr"`
		CertificateID  string   `xml:"tds:CertificateID,omitempty"`
		Subject        string   `xml:"tds:Subject"`
		ValidNotBefore string   `xml:"tds:ValidNotBefore"`
		ValidNotAfter  string   `xml:"tds:ValidNotAfter"`
	}

	type CreateCertificateResponse struct {
		XMLName     xml.Name     `xml:"CreateCertificateResponse"`
		Certificate *Certificate `xml:"Certificate"`
	}

	request := CreateCertificateBody{
		Xmlns:          deviceNamespace,
		CertificateID:  certificateID,
		Subject:        subject,
		ValidNotBefore: validNotBefore,
		ValidNotAfter:  validNotAfter,
	}
	var response CreateCertificateResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("CreateCertificate failed: %w", err)
	}

	return response.Certificate, nil
}

// DeleteCertificates deletes certificates. ONVIF Specification: DeleteCertificates operation.
func (c *Client) DeleteCertificates(ctx context.Context, certificateIDs []string) error {
	type DeleteCertificatesBody struct {
		XMLName       xml.Name `xml:"tds:DeleteCertificates"`
		Xmlns         string   `xml:"xmlns:tds,attr"`
		CertificateID []string `xml:"tds:CertificateID"`
	}

	type DeleteCertificatesResponse struct {
		XMLName xml.Name `xml:"DeleteCertificatesResponse"`
	}

	request := DeleteCertificatesBody{
		Xmlns:         deviceNamespace,
		CertificateID: certificateIDs,
	}
	var response DeleteCertificatesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("DeleteCertificates failed: %w", err)
	}

	return nil
}

// GetCertificateInformation retrieves certificate information.
// ONVIF Specification: GetCertificateInformation operation.
func (c *Client) GetCertificateInformation(ctx context.Context, certificateID string) (*CertificateInformation, error) {
	type GetCertificateInformationBody struct {
		XMLName       xml.Name `xml:"tds:GetCertificateInformation"`
		Xmlns         string   `xml:"xmlns:tds,attr"`
		CertificateID string   `xml:"tds:CertificateID"`
	}

	type GetCertificateInformationResponse struct {
		XMLName                xml.Name                `xml:"GetCertificateInformationResponse"`
		CertificateInformation *CertificateInformation `xml:"CertificateInformation"`
	}

	request := GetCertificateInformationBody{
		Xmlns:         deviceNamespace,
		CertificateID: certificateID,
	}
	var response GetCertificateInformationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCertificateInformation failed: %w", err)
	}

	return response.CertificateInformation, nil
}

// GetCertificatesStatus retrieves certificate status. ONVIF Specification: GetCertificatesStatus operation.
func (c *Client) GetCertificatesStatus(ctx context.Context) ([]*CertificateStatus, error) {
	type GetCertificatesStatusBody struct {
		XMLName xml.Name `xml:"tds:GetCertificatesStatus"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetCertificatesStatusResponse struct {
		XMLName           xml.Name             `xml:"GetCertificatesStatusResponse"`
		CertificateStatus []*CertificateStatus `xml:"CertificateStatus"`
	}

	request := GetCertificatesStatusBody{
		Xmlns: deviceNamespace,
	}
	var response GetCertificatesStatusResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCertificatesStatus failed: %w", err)
	}

	return response.CertificateStatus, nil
}

// SetCertificatesStatus sets certificate status. ONVIF Specification: SetCertificatesStatus operation.
func (c *Client) SetCertificatesStatus(ctx context.Context, statuses []*CertificateStatus) error {
	type SetCertificatesStatusBody struct {
		XMLName           xml.Name             `xml:"tds:SetCertificatesStatus"`
		Xmlns             string               `xml:"xmlns:tds,attr"`
		CertificateStatus []*CertificateStatus `xml:"tds:CertificateStatus"`
	}

	type SetCertificatesStatusResponse struct {
		XMLName xml.Name `xml:"SetCertificatesStatusResponse"`
	}

	request := SetCertificatesStatusBody{
		Xmlns:             deviceNamespace,
		CertificateStatus: statuses,
	}
	var response SetCertificatesStatusResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("SetCertificatesStatus failed: %w", err)
	}

	return nil
}

// GetPkcs10Request retrieves a PKCS10 certificate request. ONVIF Specification: GetPkcs10Request operation.
func (c *Client) GetPkcs10Request(
	ctx context.Context,
	certificateID, subject string,
	attributes *BinaryData,
) (*BinaryData, error) {
	type GetPkcs10RequestBody struct {
		XMLName       xml.Name    `xml:"tds:GetPkcs10Request"`
		Xmlns         string      `xml:"xmlns:tds,attr"`
		CertificateID string      `xml:"tds:CertificateID,omitempty"`
		Subject       string      `xml:"tds:Subject"`
		Attributes    *BinaryData `xml:"tds:Attributes,omitempty"`
	}

	type GetPkcs10RequestResponse struct {
		XMLName       xml.Name    `xml:"GetPkcs10RequestResponse"`
		Pkcs10Request *BinaryData `xml:"Pkcs10Request"`
	}

	request := GetPkcs10RequestBody{
		Xmlns:         deviceNamespace,
		CertificateID: certificateID,
		Subject:       subject,
		Attributes:    attributes,
	}
	var response GetPkcs10RequestResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return nil, fmt.Errorf("GetPkcs10Request failed: %w", err)
	}

	return response.Pkcs10Request, nil
}

// LoadCertificateWithPrivateKey loads a certificate with its private key.
// ONVIF Specification: LoadCertificateWithPrivateKey operation.
func (c *Client) LoadCertificateWithPrivateKey(
	ctx context.Context,
	certificates []*Certificate,
	privateKey []*BinaryData,
	certificateIDs []string,
) error {
	type LoadCertificateWithPrivateKeyBody struct {
		XMLName                   xml.Name `xml:"tds:LoadCertificateWithPrivateKey"`
		Xmlns                     string   `xml:"xmlns:tds,attr"`
		CertificateWithPrivateKey []struct {
			CertificateID string       `xml:"CertificateID"`
			Certificate   *Certificate `xml:"Certificate"`
			PrivateKey    *BinaryData  `xml:"PrivateKey"`
		} `xml:"tds:CertificateWithPrivateKey"`
	}

	type LoadCertificateWithPrivateKeyResponse struct {
		XMLName xml.Name `xml:"LoadCertificateWithPrivateKeyResponse"`
	}

	request := LoadCertificateWithPrivateKeyBody{
		Xmlns: deviceNamespace,
	}

	// Build certificate with private key array
	for i := 0; i < len(certificates); i++ {
		item := struct {
			CertificateID string       `xml:"CertificateID"`
			Certificate   *Certificate `xml:"Certificate"`
			PrivateKey    *BinaryData  `xml:"PrivateKey"`
		}{
			CertificateID: certificateIDs[i],
			Certificate:   certificates[i],
		}
		if i < len(privateKey) {
			item.PrivateKey = privateKey[i]
		}
		request.CertificateWithPrivateKey = append(request.CertificateWithPrivateKey, item)
	}

	var response LoadCertificateWithPrivateKeyResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("LoadCertificateWithPrivateKey failed: %w", err)
	}

	return nil
}

// GetClientCertificateMode retrieves the client certificate mode.
// ONVIF Specification: GetClientCertificateMode operation.
func (c *Client) GetClientCertificateMode(ctx context.Context) (bool, error) {
	type GetClientCertificateModeBody struct {
		XMLName xml.Name `xml:"tds:GetClientCertificateMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetClientCertificateModeResponse struct {
		XMLName xml.Name `xml:"GetClientCertificateModeResponse"`
		Enabled bool     `xml:"Enabled"`
	}

	request := GetClientCertificateModeBody{
		Xmlns: deviceNamespace,
	}
	var response GetClientCertificateModeResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return false, fmt.Errorf("GetClientCertificateMode failed: %w", err)
	}

	return response.Enabled, nil
}

// SetClientCertificateMode sets the client certificate mode. ONVIF Specification: SetClientCertificateMode operation.
func (c *Client) SetClientCertificateMode(ctx context.Context, enabled bool) error {
	type SetClientCertificateModeBody struct {
		XMLName xml.Name `xml:"tds:SetClientCertificateMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Enabled bool     `xml:"tds:Enabled"`
	}

	type SetClientCertificateModeResponse struct {
		XMLName xml.Name `xml:"SetClientCertificateModeResponse"`
	}

	request := SetClientCertificateModeBody{
		Xmlns:   deviceNamespace,
		Enabled: enabled,
	}
	var response SetClientCertificateModeResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, c.endpoint, "", request, &response); err != nil {
		return fmt.Errorf("SetClientCertificateMode failed: %w", err)
	}

	return nil
}
