package onvif

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Tests for #86: SessionTimeout was parsed off the wire and then dropped on
// read, and never populated on write, on every Media method carrying it.
//
// Each read test asserts a value that could only have come from the response,
// which is what the previous tests could not do - they asserted the other
// fields and left a zero SessionTimeout looking like a camera that had not
// reported one.

const (
	testSessionTimeoutXML = "PT90S"
	testMetadataCfgToken  = "MetaCfg1"
	testSessionTimeoutDur = 90 * time.Second
)

// newRequestCapturingServer returns a server that records the body of each
// request it receives before replying, for asserting what the client actually
// put on the wire.
func newRequestCapturingServer(t *testing.T, captured *string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		*captured = string(body)

		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
    <soap:Body/>
</soap:Envelope>`))
	}))
	t.Cleanup(server.Close)

	return server
}

func TestGetVideoEncoderConfigurationMapsSessionTimeout(t *testing.T) {
	body := `<GetVideoEncoderConfigurationResponse>
        <Configuration token="VideoEnc1">
            <Name>Encoder</Name>
            <UseCount>1</UseCount>
            <Encoding>H264</Encoding>
            <Quality>5</Quality>
            <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
        </Configuration>
    </GetVideoEncoderConfigurationResponse>`
	client := newMediaConfigClient(t, newSOAPTestServer(t, body).URL)

	config, err := client.GetVideoEncoderConfiguration(context.Background(), testVideoEncToken)
	if err != nil {
		t.Fatalf("GetVideoEncoderConfiguration() error = %v", err)
	}
	if config.SessionTimeout != testSessionTimeoutDur {
		t.Errorf("SessionTimeout = %v, want %v", config.SessionTimeout, testSessionTimeoutDur)
	}
}

func TestGetAudioEncoderConfigurationMapsSessionTimeout(t *testing.T) {
	body := `<GetAudioEncoderConfigurationResponse>
        <Configuration token="AudioEnc1">
            <Name>Encoder</Name>
            <UseCount>1</UseCount>
            <Encoding>AAC</Encoding>
            <Bitrate>64</Bitrate>
            <SampleRate>48000</SampleRate>
            <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
        </Configuration>
    </GetAudioEncoderConfigurationResponse>`
	client := newMediaConfigClient(t, newSOAPTestServer(t, body).URL)

	config, err := client.GetAudioEncoderConfiguration(context.Background(), testAudioEncToken)
	if err != nil {
		t.Fatalf("GetAudioEncoderConfiguration() error = %v", err)
	}
	if config.SessionTimeout != testSessionTimeoutDur {
		t.Errorf("SessionTimeout = %v, want %v", config.SessionTimeout, testSessionTimeoutDur)
	}
}

func TestGetMetadataConfigurationMapsSessionTimeout(t *testing.T) {
	body := `<GetMetadataConfigurationResponse>
        <Configuration token="MetaCfg1">
            <Name>Metadata</Name>
            <UseCount>1</UseCount>
            <Analytics>true</Analytics>
            <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
        </Configuration>
    </GetMetadataConfigurationResponse>`
	client := newMediaConfigClient(t, newSOAPTestServer(t, body).URL)

	config, err := client.GetMetadataConfiguration(context.Background(), testMetadataCfgToken)
	if err != nil {
		t.Fatalf("GetMetadataConfiguration() error = %v", err)
	}
	if config.SessionTimeout != testSessionTimeoutDur {
		t.Errorf("SessionTimeout = %v, want %v", config.SessionTimeout, testSessionTimeoutDur)
	}
}

// TestListGettersMapSessionTimeout covers the plural and compatible-for-profile
// variants, which each had their own copy of the same dropped mapping.
func TestListGettersMapSessionTimeout(t *testing.T) {
	tests := []struct {
		name string
		body string
		call func(*Client) (time.Duration, error)
	}{
		{
			name: "GetVideoEncoderConfigurations",
			body: `<GetVideoEncoderConfigurationsResponse>
                <Configurations token="VideoEnc1">
                    <Name>Encoder</Name><UseCount>1</UseCount><Encoding>H264</Encoding>
                    <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
                </Configurations>
            </GetVideoEncoderConfigurationsResponse>`,
			call: func(c *Client) (time.Duration, error) {
				configs, err := c.GetVideoEncoderConfigurations(context.Background())
				if err != nil {
					return 0, err
				}

				return configs[0].SessionTimeout, nil
			},
		},
		{
			name: "GetAudioEncoderConfigurations",
			body: `<GetAudioEncoderConfigurationsResponse>
                <Configurations token="AudioEnc1">
                    <Name>Encoder</Name><UseCount>1</UseCount><Encoding>AAC</Encoding>
                    <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
                </Configurations>
            </GetAudioEncoderConfigurationsResponse>`,
			call: func(c *Client) (time.Duration, error) {
				configs, err := c.GetAudioEncoderConfigurations(context.Background())
				if err != nil {
					return 0, err
				}

				return configs[0].SessionTimeout, nil
			},
		},
		{
			name: "GetMetadataConfigurations",
			body: `<GetMetadataConfigurationsResponse>
                <Configurations token="MetaCfg1">
                    <Name>Metadata</Name><UseCount>1</UseCount>
                    <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
                </Configurations>
            </GetMetadataConfigurationsResponse>`,
			call: func(c *Client) (time.Duration, error) {
				configs, err := c.GetMetadataConfigurations(context.Background())
				if err != nil {
					return 0, err
				}

				return configs[0].SessionTimeout, nil
			},
		},
		{
			name: "GetCompatibleVideoEncoderConfigurations",
			body: `<GetCompatibleVideoEncoderConfigurationsResponse>
                <Configurations token="VideoEnc1">
                    <Name>Encoder</Name><UseCount>1</UseCount><Encoding>H264</Encoding>
                    <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
                </Configurations>
            </GetCompatibleVideoEncoderConfigurationsResponse>`,
			call: func(c *Client) (time.Duration, error) {
				configs, err := c.GetCompatibleVideoEncoderConfigurations(context.Background(), testProfileToken)
				if err != nil {
					return 0, err
				}

				return configs[0].SessionTimeout, nil
			},
		},
		{
			name: "GetCompatibleAudioEncoderConfigurations",
			body: `<GetCompatibleAudioEncoderConfigurationsResponse>
                <Configurations token="AudioEnc1">
                    <Name>Encoder</Name><UseCount>1</UseCount><Encoding>AAC</Encoding>
                    <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
                </Configurations>
            </GetCompatibleAudioEncoderConfigurationsResponse>`,
			call: func(c *Client) (time.Duration, error) {
				configs, err := c.GetCompatibleAudioEncoderConfigurations(context.Background(), testProfileToken)
				if err != nil {
					return 0, err
				}

				return configs[0].SessionTimeout, nil
			},
		},
		{
			name: "GetCompatibleMetadataConfigurations",
			body: `<GetCompatibleMetadataConfigurationsResponse>
                <Configurations token="MetaCfg1">
                    <Name>Metadata</Name><UseCount>1</UseCount>
                    <SessionTimeout>` + testSessionTimeoutXML + `</SessionTimeout>
                </Configurations>
            </GetCompatibleMetadataConfigurationsResponse>`,
			call: func(c *Client) (time.Duration, error) {
				configs, err := c.GetCompatibleMetadataConfigurations(context.Background(), testProfileToken)
				if err != nil {
					return 0, err
				}

				return configs[0].SessionTimeout, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newMediaConfigClient(t, newSOAPTestServer(t, tt.body).URL)

			got, err := tt.call(client)
			if err != nil {
				t.Fatalf("%s() error = %v", tt.name, err)
			}
			if got != testSessionTimeoutDur {
				t.Errorf("%s() SessionTimeout = %v, want %v", tt.name, got, testSessionTimeoutDur)
			}
		})
	}
}

// TestSettersSendSessionTimeout is the write half of #86: the request structs
// already carried the element, but nothing ever populated it, so a caller's
// value never left the process.
func TestSettersSendSessionTimeout(t *testing.T) {
	tests := []struct {
		name string
		call func(*Client) error
	}{
		{
			name: "SetVideoEncoderConfiguration",
			call: func(c *Client) error {
				return c.SetVideoEncoderConfiguration(context.Background(), &VideoEncoderConfiguration{
					Token:          testVideoEncToken,
					Encoding:       "H264",
					SessionTimeout: testSessionTimeoutDur,
				}, false)
			},
		},
		{
			name: "SetAudioEncoderConfiguration",
			call: func(c *Client) error {
				return c.SetAudioEncoderConfiguration(context.Background(), &AudioEncoderConfiguration{
					Token:          testAudioEncToken,
					Encoding:       testEncodingAAC,
					SessionTimeout: testSessionTimeoutDur,
				}, false)
			},
		},
		{
			name: "SetMetadataConfiguration",
			call: func(c *Client) error {
				return c.SetMetadataConfiguration(context.Background(), &MetadataConfiguration{
					Token:          testMetadataCfgToken,
					SessionTimeout: testSessionTimeoutDur,
				}, false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sent string
			client := newMediaConfigClient(t, newRequestCapturingServer(t, &sent).URL)

			if err := tt.call(client); err != nil {
				t.Fatalf("%s() error = %v", tt.name, err)
			}

			// formatDuration normalizes, so 90s goes out as "PT1M30S" rather
			// than "PT90S". Deriving the expected text from the same formatter
			// the client uses keeps this asserting that the value reached the
			// wire, without also pinning a lexical choice #86 did not make.
			want := "SessionTimeout>" + formatDuration(testSessionTimeoutDur)
			if !strings.Contains(sent, want) {
				t.Errorf("%s() did not send %q\nrequest:\n%s", tt.name, want, sent)
			}
		})
	}
}

// TestSettersOmitZeroSessionTimeout keeps the element off the wire when the
// caller did not ask for one, rather than sending "PT0S" and overwriting a
// camera's own default.
func TestSettersOmitZeroSessionTimeout(t *testing.T) {
	var sent string
	client := newMediaConfigClient(t, newRequestCapturingServer(t, &sent).URL)

	err := client.SetAudioEncoderConfiguration(context.Background(), &AudioEncoderConfiguration{
		Token:    testAudioEncToken,
		Encoding: testEncodingAAC,
	}, false)
	if err != nil {
		t.Fatalf("SetAudioEncoderConfiguration() error = %v", err)
	}
	if strings.Contains(sent, "SessionTimeout") {
		t.Errorf("a zero SessionTimeout was sent anyway\nrequest:\n%s", sent)
	}
}

// TestSessionTimeoutAbsentOrMalformed documents the two cases that legitimately
// yield zero, so that a future reader does not mistake them for the bug #86
// fixed.
func TestSessionTimeoutAbsentOrMalformed(t *testing.T) {
	tests := []struct {
		name    string
		element string
	}{
		{name: "absent", element: ""},
		{name: "malformed", element: "<SessionTimeout>not-a-duration</SessionTimeout>"},
		{name: "months are unrepresentable", element: "<SessionTimeout>P1M</SessionTimeout>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `<GetAudioEncoderConfigurationResponse>
                <Configuration token="AudioEnc1">
                    <Name>Encoder</Name><UseCount>1</UseCount><Encoding>AAC</Encoding>
                    ` + tt.element + `
                </Configuration>
            </GetAudioEncoderConfigurationResponse>`
			client := newMediaConfigClient(t, newSOAPTestServer(t, body).URL)

			// A malformed timeout must not fail the call - every other field
			// in the configuration is still usable.
			config, err := client.GetAudioEncoderConfiguration(context.Background(), testAudioEncToken)
			if err != nil {
				t.Fatalf("GetAudioEncoderConfiguration() error = %v", err)
			}
			if config.SessionTimeout != 0 {
				t.Errorf("SessionTimeout = %v, want 0", config.SessionTimeout)
			}
			if config.Encoding != testEncodingAAC {
				t.Errorf("Encoding = %q, want %q - the rest of the response must survive",
					config.Encoding, testEncodingAAC)
			}
		})
	}
}
