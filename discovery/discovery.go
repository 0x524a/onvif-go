package discovery

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	// WS-Discovery multicast address
	multicastAddr = "239.255.255.250:3702"
	
	// WS-Discovery probe message
	probeTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
	<s:Header>
		<a:Action s:mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
		<a:MessageID>uuid:%s</a:MessageID>
		<a:ReplyTo>
			<a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
		</a:ReplyTo>
		<a:To s:mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
	</s:Header>
	<s:Body>
		<Probe xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
			<d:Types xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery" xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">dp0:NetworkVideoTransmitter</d:Types>
		</Probe>
	</s:Body>
</s:Envelope>`
)

// Device represents a discovered ONVIF device
type Device struct {
	// Device endpoint address
	EndpointRef string
	
	// XAddrs contains the device service addresses
	XAddrs []string
	
	// Types contains the device types
	Types []string
	
	// Scopes contains the device scopes (name, location, etc.)
	Scopes []string
	
	// Metadata version
	MetadataVersion int
}

// ProbeMatch represents a WS-Discovery probe match
type ProbeMatch struct {
	XMLName         xml.Name `xml:"ProbeMatch"`
	EndpointRef     string   `xml:"EndpointReference>Address"`
	Types           string   `xml:"Types"`
	Scopes          string   `xml:"Scopes"`
	XAddrs          string   `xml:"XAddrs"`
	MetadataVersion int      `xml:"MetadataVersion"`
}

// ProbeMatches represents WS-Discovery probe matches
type ProbeMatches struct {
	XMLName     xml.Name      `xml:"ProbeMatches"`
	ProbeMatch  []ProbeMatch  `xml:"ProbeMatch"`
}

// Discover discovers ONVIF devices on the network
func Discover(ctx context.Context, timeout time.Duration) ([]*Device, error) {
	// Create UDP connection for multicast
	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on multicast address: %w", err)
	}
	defer conn.Close()

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Generate message ID
	messageID := generateUUID()

	// Send probe message
	probeMsg := fmt.Sprintf(probeTemplate, messageID)
	if _, err := conn.WriteToUDP([]byte(probeMsg), addr); err != nil {
		return nil, fmt.Errorf("failed to send probe message: %w", err)
	}

	// Collect responses
	devices := make(map[string]*Device)
	buffer := make([]byte, 8192)

	// Read responses until timeout or context cancellation
	for {
		select {
		case <-ctx.Done():
			return deviceMapToSlice(devices), ctx.Err()
		default:
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Timeout reached, return collected devices
					return deviceMapToSlice(devices), nil
				}
				return deviceMapToSlice(devices), fmt.Errorf("failed to read UDP response: %w", err)
			}

			// Parse response
			device, err := parseProbeResponse(buffer[:n])
			if err != nil {
				// Skip invalid responses
				continue
			}

			// Add to devices map (deduplicate by endpoint)
			if device != nil && device.EndpointRef != "" {
				devices[device.EndpointRef] = device
			}
		}
	}
}

// parseProbeResponse parses a WS-Discovery probe response
func parseProbeResponse(data []byte) (*Device, error) {
	var envelope struct {
		Body struct {
			ProbeMatches ProbeMatches `xml:"ProbeMatches"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	if len(envelope.Body.ProbeMatches.ProbeMatch) == 0 {
		return nil, fmt.Errorf("no probe matches found")
	}

	// Take the first probe match
	match := envelope.Body.ProbeMatches.ProbeMatch[0]

	device := &Device{
		EndpointRef:     match.EndpointRef,
		XAddrs:          parseSpaceSeparated(match.XAddrs),
		Types:           parseSpaceSeparated(match.Types),
		Scopes:          parseSpaceSeparated(match.Scopes),
		MetadataVersion: match.MetadataVersion,
	}

	return device, nil
}

// parseSpaceSeparated parses a space-separated string into a slice
func parseSpaceSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	return strings.Fields(s)
}

// deviceMapToSlice converts a map of devices to a slice
func deviceMapToSlice(m map[string]*Device) []*Device {
	devices := make([]*Device, 0, len(m))
	for _, device := range m {
		devices = append(devices, device)
	}
	return devices
}

// generateUUID generates a simple UUID (not cryptographically secure)
func generateUUID() string {
	return fmt.Sprintf("%d-%d-%d-%d-%d",
		time.Now().UnixNano(),
		time.Now().Unix(),
		time.Now().UnixNano()%1000,
		time.Now().Unix()%1000,
		time.Now().UnixNano()%10000)
}

// GetDeviceEndpoint extracts the primary device endpoint from XAddrs
func (d *Device) GetDeviceEndpoint() string {
	if len(d.XAddrs) == 0 {
		return ""
	}
	
	// Return the first XAddr
	return d.XAddrs[0]
}

// GetName extracts the device name from scopes
func (d *Device) GetName() string {
	for _, scope := range d.Scopes {
		if strings.Contains(scope, "name") {
			parts := strings.Split(scope, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}

// GetLocation extracts the device location from scopes
func (d *Device) GetLocation() string {
	for _, scope := range d.Scopes {
		if strings.Contains(scope, "location") {
			parts := strings.Split(scope, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}
