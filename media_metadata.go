package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// GetMetadataConfiguration retrieves metadata configuration.
func (c *Client) GetMetadataConfiguration(
	ctx context.Context,
	configurationToken string,
) (*MetadataConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetMetadataConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetMetadataConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetMetadataConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetMetadataConfigurationResponse"`
		Configuration struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			PTZStatus *struct {
				Status   bool `xml:"Status"`
				Position bool `xml:"Position"`
			} `xml:"PTZStatus"`
			Events    *struct{} `xml:"Events"`
			Analytics bool      `xml:"Analytics"`
			Multicast *struct {
				Address *struct {
					Type        string `xml:"Type"`
					IPv4Address string `xml:"IPv4Address"`
					IPv6Address string `xml:"IPv6Address"`
				} `xml:"Address"`
				Port      int  `xml:"Port"`
				TTL       int  `xml:"TTL"`
				AutoStart bool `xml:"AutoStart"`
			} `xml:"Multicast"`
			SessionTimeout string `xml:"SessionTimeout"`
		} `xml:"Configuration"`
	}

	req := GetMetadataConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetMetadataConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfiguration failed: %w", err)
	}

	config := &MetadataConfiguration{
		Token:          resp.Configuration.Token,
		Name:           resp.Configuration.Name,
		UseCount:       resp.Configuration.UseCount,
		Analytics:      resp.Configuration.Analytics,
		SessionTimeout: parseXSDurationOrZero(resp.Configuration.SessionTimeout),
	}

	if resp.Configuration.PTZStatus != nil {
		config.PTZStatus = &PTZFilter{
			Status:   resp.Configuration.PTZStatus.Status,
			Position: resp.Configuration.PTZStatus.Position,
		}
	}

	if resp.Configuration.Events != nil {
		config.Events = &EventSubscription{}
	}

	if resp.Configuration.Multicast != nil {
		config.Multicast = &MulticastConfiguration{
			Port:      resp.Configuration.Multicast.Port,
			TTL:       resp.Configuration.Multicast.TTL,
			AutoStart: resp.Configuration.Multicast.AutoStart,
		}
		if resp.Configuration.Multicast.Address != nil {
			config.Multicast.Address = &IPAddress{
				Type:        resp.Configuration.Multicast.Address.Type,
				IPv4Address: resp.Configuration.Multicast.Address.IPv4Address,
				IPv6Address: resp.Configuration.Multicast.Address.IPv6Address,
			}
		}
	}

	return config, nil
}

// SetMetadataConfiguration sets metadata configuration.
func (c *Client) SetMetadataConfiguration(
	ctx context.Context,
	config *MetadataConfiguration,
	forcePersistence bool,
) error {
	endpoint := c.getMediaEndpoint()

	type SetMetadataConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetMetadataConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"tt:Name"`
			UseCount  int    `xml:"tt:UseCount"`
			PTZStatus *struct {
				Status   bool `xml:"tt:Status"`
				Position bool `xml:"tt:Position"`
			} `xml:"tt:PTZStatus,omitempty"`
			Events    *struct{} `xml:"tt:Events,omitempty"`
			Analytics bool      `xml:"tt:Analytics,omitempty"`
			Multicast *struct {
				Address *struct {
					Type        string `xml:"tt:Type"`
					IPv4Address string `xml:"tt:IPv4Address,omitempty"`
					IPv6Address string `xml:"tt:IPv6Address,omitempty"`
				} `xml:"tt:Address,omitempty"`
				Port      int  `xml:"tt:Port,omitempty"`
				TTL       int  `xml:"tt:TTL,omitempty"`
				AutoStart bool `xml:"tt:AutoStart,omitempty"`
			} `xml:"tt:Multicast,omitempty"`
			SessionTimeout string `xml:"tt:SessionTimeout,omitempty"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetMetadataConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Analytics = config.Analytics
	if config.SessionTimeout > 0 {
		req.Configuration.SessionTimeout = formatDuration(config.SessionTimeout)
	}

	if config.PTZStatus != nil {
		req.Configuration.PTZStatus = &struct {
			Status   bool `xml:"tt:Status"`
			Position bool `xml:"tt:Position"`
		}{
			Status:   config.PTZStatus.Status,
			Position: config.PTZStatus.Position,
		}
	}

	if config.Events != nil {
		req.Configuration.Events = &struct{}{}
	}

	if config.Multicast != nil {
		req.Configuration.Multicast = &struct {
			Address *struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			} `xml:"tt:Address,omitempty"`
			Port      int  `xml:"tt:Port,omitempty"`
			TTL       int  `xml:"tt:TTL,omitempty"`
			AutoStart bool `xml:"tt:AutoStart,omitempty"`
		}{
			Port:      config.Multicast.Port,
			TTL:       config.Multicast.TTL,
			AutoStart: config.Multicast.AutoStart,
		}
		if config.Multicast.Address != nil {
			req.Configuration.Multicast.Address = &struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			}{
				Type:        config.Multicast.Address.Type,
				IPv4Address: config.Multicast.Address.IPv4Address,
				IPv6Address: config.Multicast.Address.IPv6Address,
			}
		}
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetMetadataConfiguration failed: %w", err)
	}

	return nil
}

// AddPTZConfiguration adds PTZ configuration to a profile.
func (c *Client) AddPTZConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddPTZConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddPTZConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddPTZConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddPTZConfiguration failed: %w", err)
	}

	return nil
}

// RemovePTZConfiguration removes PTZ configuration from a profile.
func (c *Client) RemovePTZConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemovePTZConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemovePTZConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemovePTZConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemovePTZConfiguration failed: %w", err)
	}

	return nil
}

// AddMetadataConfiguration adds metadata configuration to a profile.
func (c *Client) AddMetadataConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddMetadataConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddMetadataConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddMetadataConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddMetadataConfiguration failed: %w", err)
	}

	return nil
}

// RemoveMetadataConfiguration removes metadata configuration from a profile.
func (c *Client) RemoveMetadataConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveMetadataConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveMetadataConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveMetadataConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveMetadataConfiguration failed: %w", err)
	}

	return nil
}

// GetMetadataConfigurationOptions retrieves available options for metadata configuration.
func (c *Client) GetMetadataConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*MetadataConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetMetadataConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetMetadataConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetMetadataConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetMetadataConfigurationOptionsResponse"`
		Options struct {
			PTZStatusFilterOptions *struct {
				Status   bool `xml:"Status"`
				Position bool `xml:"Position"`
			} `xml:"PTZStatusFilterOptions"`
			Extension struct{} `xml:"Extension"`
		} `xml:"Options"`
	}

	req := GetMetadataConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetMetadataConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfigurationOptions failed: %w", err)
	}

	options := &MetadataConfigurationOptions{}
	if resp.Options.PTZStatusFilterOptions != nil {
		options.PTZStatusFilterOptions = &PTZFilter{
			Status:   resp.Options.PTZStatusFilterOptions.Status,
			Position: resp.Options.PTZStatusFilterOptions.Position,
		}
	}

	return options, nil
}

// GetCompatiblePTZConfigurations retrieves compatible PTZ configurations for a profile.
func (c *Client) GetCompatiblePTZConfigurations(ctx context.Context, profileToken string) ([]*PTZConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatiblePTZConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatiblePTZConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatiblePTZConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatiblePTZConfigurationsResponse"`
		Configurations []struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			NodeToken string `xml:"NodeToken"`
		} `xml:"Configurations"`
	}

	req := GetCompatiblePTZConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatiblePTZConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatiblePTZConfigurations failed: %w", err)
	}

	configs := make([]*PTZConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &PTZConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			NodeToken: cfg.NodeToken,
		}
	}

	return configs, nil
}

// GetCompatibleMetadataConfigurations retrieves compatible metadata configurations for a profile.
func (c *Client) GetCompatibleMetadataConfigurations(ctx context.Context, profileToken string) ([]*MetadataConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleMetadataConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleMetadataConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleMetadataConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleMetadataConfigurationsResponse"`
		Configurations []struct {
			Token          string `xml:"token,attr"`
			Name           string `xml:"Name"`
			UseCount       int    `xml:"UseCount"`
			Analytics      bool   `xml:"Analytics"`
			SessionTimeout string `xml:"SessionTimeout"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleMetadataConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleMetadataConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleMetadataConfigurations failed: %w", err)
	}

	configs := make([]*MetadataConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &MetadataConfiguration{
			Token:          cfg.Token,
			Name:           cfg.Name,
			UseCount:       cfg.UseCount,
			Analytics:      cfg.Analytics,
			SessionTimeout: parseXSDurationOrZero(cfg.SessionTimeout),
		}
	}

	return configs, nil
}

// GetMetadataConfigurations retrieves all metadata configurations.
func (c *Client) GetMetadataConfigurations(ctx context.Context) ([]*MetadataConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetMetadataConfigurations struct {
		XMLName xml.Name `xml:"trt:GetMetadataConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetMetadataConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetMetadataConfigurationsResponse"`
		Configurations []struct {
			Token          string `xml:"token,attr"`
			Name           string `xml:"Name"`
			UseCount       int    `xml:"UseCount"`
			Analytics      bool   `xml:"Analytics"`
			SessionTimeout string `xml:"SessionTimeout"`
		} `xml:"Configurations"`
	}

	req := GetMetadataConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetMetadataConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfigurations failed: %w", err)
	}

	configs := make([]*MetadataConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &MetadataConfiguration{
			Token:          cfg.Token,
			Name:           cfg.Name,
			UseCount:       cfg.UseCount,
			Analytics:      cfg.Analytics,
			SessionTimeout: parseXSDurationOrZero(cfg.SessionTimeout),
		}
	}

	return configs, nil
}
