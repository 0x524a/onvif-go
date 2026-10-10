package onvif

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/0x524a/onvif-go/internal/soap"
)

// GetAudioSources retrieves all audio sources.
func (c *Client) GetAudioSources(ctx context.Context) ([]*AudioSource, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSources struct {
		XMLName xml.Name `xml:"trt:GetAudioSources"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioSourcesResponse struct {
		XMLName      xml.Name `xml:"GetAudioSourcesResponse"`
		AudioSources []struct {
			Token    string `xml:"token,attr"`
			Channels int    `xml:"Channels"`
		} `xml:"AudioSources"`
	}

	req := GetAudioSources{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioSourcesResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSources failed: %w", err)
	}

	sources := make([]*AudioSource, len(resp.AudioSources))
	for i, s := range resp.AudioSources {
		sources[i] = &AudioSource{
			Token:    s.Token,
			Channels: s.Channels,
		}
	}

	return sources, nil
}

// GetAudioOutputs retrieves all audio outputs.
func (c *Client) GetAudioOutputs(ctx context.Context) ([]*AudioOutput, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputs struct {
		XMLName xml.Name `xml:"trt:GetAudioOutputs"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioOutputsResponse struct {
		XMLName      xml.Name `xml:"GetAudioOutputsResponse"`
		AudioOutputs []struct {
			Token string `xml:"token,attr"`
		} `xml:"AudioOutputs"`
	}

	req := GetAudioOutputs{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioOutputsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputs failed: %w", err)
	}

	outputs := make([]*AudioOutput, len(resp.AudioOutputs))
	for i, o := range resp.AudioOutputs {
		outputs[i] = &AudioOutput{
			Token: o.Token,
		}
	}

	return outputs, nil
}

// GetAudioEncoderConfiguration retrieves audio encoder configuration.
func (c *Client) GetAudioEncoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*AudioEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioEncoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioEncoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioEncoderConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioEncoderConfigurationResponse"`
		Configuration struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Bitrate    int    `xml:"Bitrate"`
			SampleRate int    `xml:"SampleRate"`
			Multicast  *struct {
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

	req := GetAudioEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioEncoderConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfiguration failed: %w", err)
	}

	config := &AudioEncoderConfiguration{
		Token:          resp.Configuration.Token,
		Name:           resp.Configuration.Name,
		UseCount:       resp.Configuration.UseCount,
		Encoding:       resp.Configuration.Encoding,
		Bitrate:        resp.Configuration.Bitrate,
		SampleRate:     resp.Configuration.SampleRate,
		SessionTimeout: parseXSDurationOrZero(resp.Configuration.SessionTimeout),
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

// SetAudioEncoderConfiguration sets audio encoder configuration.
func (c *Client) SetAudioEncoderConfiguration(
	ctx context.Context,
	config *AudioEncoderConfiguration,
	forcePersistence bool,
) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioEncoderConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioEncoderConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"tt:Name"`
			UseCount   int    `xml:"tt:UseCount"`
			Encoding   string `xml:"tt:Encoding"`
			Bitrate    int    `xml:"tt:Bitrate,omitempty"`
			SampleRate int    `xml:"tt:SampleRate,omitempty"`
			Multicast  *struct {
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

	req := SetAudioEncoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Encoding = config.Encoding
	if config.Bitrate > 0 {
		req.Configuration.Bitrate = config.Bitrate
	}
	if config.SampleRate > 0 {
		req.Configuration.SampleRate = config.SampleRate
	}
	if config.SessionTimeout > 0 {
		req.Configuration.SessionTimeout = formatDuration(config.SessionTimeout)
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
		return fmt.Errorf("SetAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

// AddAudioEncoderConfiguration adds audio encoder configuration to a profile.
func (c *Client) AddAudioEncoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioEncoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioEncoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioEncoderConfiguration removes audio encoder configuration from a profile.
func (c *Client) RemoveAudioEncoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioEncoderConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioEncoderConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioEncoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

// AddAudioSourceConfiguration adds audio source configuration to a profile.
func (c *Client) AddAudioSourceConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioSourceConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioSourceConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioSourceConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioSourceConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioSourceConfiguration removes audio source configuration from a profile.
func (c *Client) RemoveAudioSourceConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioSourceConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioSourceConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioSourceConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioSourceConfiguration failed: %w", err)
	}

	return nil
}

// GetAudioEncoderConfigurationOptions retrieves available options for audio encoder configuration.
func (c *Client) GetAudioEncoderConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*AudioEncoderConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioEncoderConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioEncoderConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetAudioEncoderConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioEncoderConfigurationOptionsResponse"`
		Options struct {
			EncodingOptions []string `xml:"EncodingOptions"`
			BitrateList     []int    `xml:"BitrateList"`
			SampleRateList  []int    `xml:"SampleRateList"`
		} `xml:"Options"`
	}

	req := GetAudioEncoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetAudioEncoderConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfigurationOptions failed: %w", err)
	}

	return &AudioEncoderConfigurationOptions{
		EncodingOptions: resp.Options.EncodingOptions,
		BitrateList:     resp.Options.BitrateList,
		SampleRateList:  resp.Options.SampleRateList,
	}, nil
}

// GetAudioOutputConfiguration retrieves audio output configuration.
func (c *Client) GetAudioOutputConfiguration(ctx context.Context, configurationToken string) (*AudioOutputConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioOutputConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioOutputConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioOutputConfigurationResponse"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			OutputToken string `xml:"OutputToken"`
		} `xml:"Configuration"`
	}

	req := GetAudioOutputConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioOutputConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfiguration failed: %w", err)
	}

	return &AudioOutputConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		OutputToken: resp.Configuration.OutputToken,
	}, nil
}

// SetAudioOutputConfiguration sets audio output configuration.
func (c *Client) SetAudioOutputConfiguration(ctx context.Context, config *AudioOutputConfiguration, forcePersistence bool) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioOutputConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioOutputConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"tt:Name"`
			UseCount    int    `xml:"tt:UseCount"`
			OutputToken string `xml:"tt:OutputToken"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetAudioOutputConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.OutputToken = config.OutputToken

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

// GetAudioOutputConfigurationOptions retrieves available options for audio output configuration.
func (c *Client) GetAudioOutputConfigurationOptions(
	ctx context.Context,
	configurationToken string,
) (*AudioOutputConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioOutputConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetAudioOutputConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioOutputConfigurationOptionsResponse"`
		Options struct {
			OutputTokensAvailable []string `xml:"OutputTokensAvailable"`
		} `xml:"Options"`
	}

	req := GetAudioOutputConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetAudioOutputConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfigurationOptions failed: %w", err)
	}

	return &AudioOutputConfigurationOptions{
		OutputTokensAvailable: resp.Options.OutputTokensAvailable,
	}, nil
}

// GetAudioDecoderConfigurationOptions retrieves available options for audio decoder configuration.
func (c *Client) GetAudioDecoderConfigurationOptions(
	ctx context.Context,
	configurationToken string,
) (*AudioDecoderConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioDecoderConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioDecoderConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	}

	type GetAudioDecoderConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioDecoderConfigurationOptionsResponse"`
		Options struct {
			AACDecOptions *struct {
				BitrateList    []int `xml:"BitrateList"`
				SampleRateList []int `xml:"SampleRateList"`
			} `xml:"AACDecOptions"`
			G711DecOptions *struct {
				BitrateList []int `xml:"BitrateList"`
			} `xml:"G711DecOptions"`
			G726DecOptions *struct {
				BitrateList []int `xml:"BitrateList"`
			} `xml:"G726DecOptions"`
		} `xml:"Options"`
	}

	req := GetAudioDecoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetAudioDecoderConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfigurationOptions failed: %w", err)
	}

	options := &AudioDecoderConfigurationOptions{}
	if resp.Options.AACDecOptions != nil {
		options.AACDecOptions = &AudioDecoderOptions{
			BitrateList:    resp.Options.AACDecOptions.BitrateList,
			SampleRateList: resp.Options.AACDecOptions.SampleRateList,
		}
	}
	if resp.Options.G711DecOptions != nil {
		options.G711DecOptions = &AudioDecoderOptions{
			BitrateList: resp.Options.G711DecOptions.BitrateList,
		}
	}
	if resp.Options.G726DecOptions != nil {
		options.G726DecOptions = &AudioDecoderOptions{
			BitrateList: resp.Options.G726DecOptions.BitrateList,
		}
	}

	return options, nil
}

// GetAudioSourceConfigurations retrieves all audio source configurations.
func (c *Client) GetAudioSourceConfigurations(ctx context.Context) ([]*AudioSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSourceConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioSourceConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioSourceConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioSourceConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
		} `xml:"Configurations"`
	}

	req := GetAudioSourceConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioSourceConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfigurations failed: %w", err)
	}

	configs := make([]*AudioSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
	}

	return configs, nil
}

// GetAudioEncoderConfigurations retrieves all audio encoder configurations.
func (c *Client) GetAudioEncoderConfigurations(ctx context.Context) ([]*AudioEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioEncoderConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioEncoderConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioEncoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioEncoderConfigurationsResponse"`
		Configurations []struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Bitrate    int    `xml:"Bitrate"`
			SampleRate int    `xml:"SampleRate"`
			Multicast  *struct {
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
		} `xml:"Configurations"`
	}

	req := GetAudioEncoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioEncoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &AudioEncoderConfiguration{
			Token:          cfg.Token,
			Name:           cfg.Name,
			UseCount:       cfg.UseCount,
			Encoding:       cfg.Encoding,
			Bitrate:        cfg.Bitrate,
			SampleRate:     cfg.SampleRate,
			SessionTimeout: parseXSDurationOrZero(cfg.SessionTimeout),
		}

		if cfg.Multicast != nil {
			config.Multicast = &MulticastConfiguration{
				Port:      cfg.Multicast.Port,
				TTL:       cfg.Multicast.TTL,
				AutoStart: cfg.Multicast.AutoStart,
			}
			if cfg.Multicast.Address != nil {
				config.Multicast.Address = &IPAddress{
					Type:        cfg.Multicast.Address.Type,
					IPv4Address: cfg.Multicast.Address.IPv4Address,
					IPv6Address: cfg.Multicast.Address.IPv6Address,
				}
			}
		}

		configs[i] = config
	}

	return configs, nil
}

// GetAudioSourceConfiguration retrieves a specific audio source configuration.
func (c *Client) GetAudioSourceConfiguration(ctx context.Context, configurationToken string) (*AudioSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSourceConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioSourceConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioSourceConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioSourceConfigurationResponse"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
		} `xml:"Configuration"`
	}

	req := GetAudioSourceConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioSourceConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfiguration failed: %w", err)
	}

	return &AudioSourceConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		SourceToken: resp.Configuration.SourceToken,
	}, nil
}

// GetAudioSourceConfigurationOptions retrieves available options for audio source configuration.
func (c *Client) GetAudioSourceConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*AudioSourceConfigurationOptions, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioSourceConfigurationOptions struct {
		XMLName            xml.Name `xml:"trt:GetAudioSourceConfigurationOptions"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
		ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
	}

	type GetAudioSourceConfigurationOptionsResponse struct {
		XMLName xml.Name `xml:"GetAudioSourceConfigurationOptionsResponse"`
		Options struct {
			InputTokensAvailable []string `xml:"InputTokensAvailable"`
		} `xml:"Options"`
	}

	req := GetAudioSourceConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetAudioSourceConfigurationOptionsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfigurationOptions failed: %w", err)
	}

	return &AudioSourceConfigurationOptions{
		InputTokensAvailable: resp.Options.InputTokensAvailable,
	}, nil
}

// SetAudioSourceConfiguration sets audio source configuration.
func (c *Client) SetAudioSourceConfiguration(ctx context.Context, config *AudioSourceConfiguration, forcePersistence bool) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioSourceConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioSourceConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"tt:Name"`
			UseCount    int    `xml:"tt:UseCount"`
			SourceToken string `xml:"tt:SourceToken"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetAudioSourceConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.SourceToken = config.SourceToken

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioSourceConfiguration failed: %w", err)
	}

	return nil
}

// GetCompatibleAudioEncoderConfigurations retrieves compatible audio encoder configurations for a profile.
func (c *Client) GetCompatibleAudioEncoderConfigurations(
	ctx context.Context,
	profileToken string,
) ([]*AudioEncoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioEncoderConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioEncoderConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioEncoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioEncoderConfigurationsResponse"`
		Configurations []struct {
			Token          string `xml:"token,attr"`
			Name           string `xml:"Name"`
			UseCount       int    `xml:"UseCount"`
			Encoding       string `xml:"Encoding"`
			Bitrate        int    `xml:"Bitrate"`
			SampleRate     int    `xml:"SampleRate"`
			SessionTimeout string `xml:"SessionTimeout"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioEncoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioEncoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioEncoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioEncoderConfiguration{
			Token:          cfg.Token,
			Name:           cfg.Name,
			UseCount:       cfg.UseCount,
			Encoding:       cfg.Encoding,
			Bitrate:        cfg.Bitrate,
			SampleRate:     cfg.SampleRate,
			SessionTimeout: parseXSDurationOrZero(cfg.SessionTimeout),
		}
	}

	return configs, nil
}

// GetCompatibleAudioSourceConfigurations retrieves compatible audio source configurations for a profile.
func (c *Client) GetCompatibleAudioSourceConfigurations(ctx context.Context, profileToken string) ([]*AudioSourceConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioSourceConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioSourceConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioSourceConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioSourceConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioSourceConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioSourceConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioSourceConfigurations failed: %w", err)
	}

	configs := make([]*AudioSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
	}

	return configs, nil
}

// GetCompatibleAudioOutputConfigurations retrieves compatible audio output configurations for a profile.
func (c *Client) GetCompatibleAudioOutputConfigurations(ctx context.Context, profileToken string) ([]*AudioOutputConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioOutputConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioOutputConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioOutputConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioOutputConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			OutputToken string `xml:"OutputToken"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioOutputConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioOutputConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioOutputConfigurations failed: %w", err)
	}

	configs := make([]*AudioOutputConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioOutputConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			OutputToken: cfg.OutputToken,
		}
	}

	return configs, nil
}

// GetCompatibleAudioDecoderConfigurations retrieves compatible audio decoder configurations for a profile.
func (c *Client) GetCompatibleAudioDecoderConfigurations(ctx context.Context, profileToken string) ([]*AudioDecoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetCompatibleAudioDecoderConfigurations struct {
		XMLName      xml.Name `xml:"trt:GetCompatibleAudioDecoderConfigurations"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	type GetCompatibleAudioDecoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetCompatibleAudioDecoderConfigurationsResponse"`
		Configurations []struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configurations"`
	}

	req := GetCompatibleAudioDecoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioDecoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioDecoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioDecoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioDecoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

// GetAudioOutputConfigurations retrieves all audio output configurations.
func (c *Client) GetAudioOutputConfigurations(ctx context.Context) ([]*AudioOutputConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioOutputConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioOutputConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioOutputConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioOutputConfigurationsResponse"`
		Configurations []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			OutputToken string `xml:"OutputToken"`
		} `xml:"Configurations"`
	}

	req := GetAudioOutputConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioOutputConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfigurations failed: %w", err)
	}

	configs := make([]*AudioOutputConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioOutputConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			OutputToken: cfg.OutputToken,
		}
	}

	return configs, nil
}

// GetAudioDecoderConfigurations retrieves all audio decoder configurations.
func (c *Client) GetAudioDecoderConfigurations(ctx context.Context) ([]*AudioDecoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioDecoderConfigurations struct {
		XMLName xml.Name `xml:"trt:GetAudioDecoderConfigurations"`
		Xmlns   string   `xml:"xmlns:trt,attr"`
	}

	type GetAudioDecoderConfigurationsResponse struct {
		XMLName        xml.Name `xml:"GetAudioDecoderConfigurationsResponse"`
		Configurations []struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configurations"`
	}

	req := GetAudioDecoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioDecoderConfigurationsResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioDecoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioDecoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

// GetAudioDecoderConfiguration retrieves a specific audio decoder configuration.
func (c *Client) GetAudioDecoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*AudioDecoderConfiguration, error) {
	endpoint := c.getMediaEndpoint()

	type GetAudioDecoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:GetAudioDecoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	type GetAudioDecoderConfigurationResponse struct {
		XMLName       xml.Name `xml:"GetAudioDecoderConfigurationResponse"`
		Configuration struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"Name"`
			UseCount int    `xml:"UseCount"`
		} `xml:"Configuration"`
	}

	req := GetAudioDecoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioDecoderConfigurationResponse

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfiguration failed: %w", err)
	}

	return &AudioDecoderConfiguration{
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
	}, nil
}

// SetAudioDecoderConfiguration sets audio decoder configuration.
func (c *Client) SetAudioDecoderConfiguration(ctx context.Context, config *AudioDecoderConfiguration, forcePersistence bool) error {
	endpoint := c.getMediaEndpoint()

	type SetAudioDecoderConfiguration struct {
		XMLName       xml.Name `xml:"trt:SetAudioDecoderConfiguration"`
		Xmlns         string   `xml:"xmlns:trt,attr"`
		Xmlnst        string   `xml:"xmlns:tt,attr"`
		Configuration struct {
			Token    string `xml:"token,attr"`
			Name     string `xml:"tt:Name"`
			UseCount int    `xml:"tt:UseCount"`
		} `xml:"trt:Configuration"`
		ForcePersistence bool `xml:"trt:ForcePersistence"`
	}

	req := SetAudioDecoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           onvifSchemaNamespace,
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}

// AddAudioOutputConfiguration adds an audio output configuration to a profile.
func (c *Client) AddAudioOutputConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioOutputConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioOutputConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioOutputConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioOutputConfiguration removes an audio output configuration from a profile.
func (c *Client) RemoveAudioOutputConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioOutputConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioOutputConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioOutputConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

// AddAudioDecoderConfiguration adds an audio decoder configuration to a profile.
func (c *Client) AddAudioDecoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := c.getMediaEndpoint()

	type AddAudioDecoderConfiguration struct {
		XMLName            xml.Name `xml:"trt:AddAudioDecoderConfiguration"`
		Xmlns              string   `xml:"xmlns:trt,attr"`
		ProfileToken       string   `xml:"trt:ProfileToken"`
		ConfigurationToken string   `xml:"trt:ConfigurationToken"`
	}

	req := AddAudioDecoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}

// RemoveAudioDecoderConfiguration removes an audio decoder configuration from a profile.
func (c *Client) RemoveAudioDecoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := c.getMediaEndpoint()

	type RemoveAudioDecoderConfiguration struct {
		XMLName      xml.Name `xml:"trt:RemoveAudioDecoderConfiguration"`
		Xmlns        string   `xml:"xmlns:trt,attr"`
		ProfileToken string   `xml:"trt:ProfileToken"`
	}

	req := RemoveAudioDecoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := c.GetCredentials()
	soapClient := soap.NewClient(c.httpClient, username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}
