// Package discovery provides ONVIF device discovery functionality using WS-Discovery protocol.
package discovery

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	// WS-Discovery multicast address.
	multicastAddr = "239.255.255.250:3702"
	// UUID generation constants.
	uuidMod1000  = 1000
	uuidMod10000 = 10000

	// WS-Discovery probe message.
	probeTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" ` +
		`xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
	<s:Header>
		<a:Action s:mustUnderstand="1">` +
		`http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
		<a:MessageID>uuid:%s</a:MessageID>
		<a:ReplyTo>
			<a:Address>` +
		`http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
		</a:ReplyTo>
		<a:To s:mustUnderstand="1">` +
		`urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
	</s:Header>
	<s:Body>
		<Probe xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
			<d:Types xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery" ` +
		`xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">` +
		`dp0:NetworkVideoTransmitter</d:Types>
		</Probe>
	</s:Body>
</s:Envelope>`
)

// Device represents a discovered ONVIF device.
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

// ProbeMatch represents a WS-Discovery probe match.
type ProbeMatch struct {
	XMLName         xml.Name `xml:"ProbeMatch"`
	EndpointRef     string   `xml:"EndpointReference>Address"`
	Types           string   `xml:"Types"`
	Scopes          string   `xml:"Scopes"`
	XAddrs          string   `xml:"XAddrs"`
	MetadataVersion int      `xml:"MetadataVersion"`
}

// ProbeMatches represents WS-Discovery probe matches.
type ProbeMatches struct {
	XMLName    xml.Name     `xml:"ProbeMatches"`
	ProbeMatch []ProbeMatch `xml:"ProbeMatch"`
}

// DiscoverOptions contains options for device discovery.
type DiscoverOptions struct {
	// NetworkInterface specifies the network interface to use for multicast.
	// If empty, the system will choose the default interface.
	// Examples: "eth0", "wlan0", "192.168.1.100"
	NetworkInterface string

	// Context and timeout are handled by the caller
}

// Discover performs ONVIF device discovery using WS-Discovery protocol.
// For advanced options like specifying a network interface, use DiscoverWithOptions.
func Discover(ctx context.Context, timeout time.Duration) ([]*Device, error) {
	return DiscoverWithOptions(ctx, timeout, &DiscoverOptions{})
}

// DiscoverWithOptions discovers ONVIF devices with custom options.
//
//nolint:gocyclo // Discovery function has high complexity due to multiple network operations
func DiscoverWithOptions(ctx context.Context, timeout time.Duration, opts *DiscoverOptions) ([]*Device, error) {
	if opts == nil {
		opts = &DiscoverOptions{}
	}

	// Create UDP connection for multicast
	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	// Get the network interface to use
	var iface *net.Interface
	if opts.NetworkInterface != "" {
		iface, err = resolveNetworkInterface(opts.NetworkInterface)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve network interface: %w", err)
		}
	}

	conn, err := net.ListenMulticastUDP("udp", iface, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on multicast address: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

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
	const maxUDPPacketSize = 8192
	buffer := make([]byte, maxUDPPacketSize)

	// Read responses until timeout or context cancellation
	for {
		select {
		case <-ctx.Done():
			return deviceMapToSlice(devices), ctx.Err()
		default:
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
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

// parseProbeResponse parses a WS-Discovery probe response.
func parseProbeResponse(data []byte) (*Device, error) {
	var envelope struct {
		Body struct {
			ProbeMatches ProbeMatches `xml:"ProbeMatches"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal probe response: %w", err)
	}

	if len(envelope.Body.ProbeMatches.ProbeMatch) == 0 {
		return nil, fmt.Errorf("%w", ErrNoProbeMatches)
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

// parseSpaceSeparated parses a space-separated string into a slice.
func parseSpaceSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}

	return strings.Fields(s)
}

// deviceMapToSlice converts a map of devices to a slice.
func deviceMapToSlice(m map[string]*Device) []*Device {
	devices := make([]*Device, 0, len(m))
	for _, device := range m {
		devices = append(devices, device)
	}

	return devices
}

// generateUUID generates a simple UUID (not cryptographically secure).
func generateUUID() string {
	now := time.Now()
	nanos := now.UnixNano()
	secs := now.Unix()

	return fmt.Sprintf("%d-%d-%d-%d-%d",
		nanos,
		secs,
		nanos%uuidMod1000,
		secs%uuidMod1000,
		nanos%uuidMod10000)
}

// resolveNetworkInterface resolves a network interface by name or IP address.
//
//nolint:gocyclo,gocognit // Network interface resolution has high complexity due to multiple validation paths
func resolveNetworkInterface(ifaceSpec string) (*net.Interface, error) {
	// Try to get interface by name (e.g., "eth0", "wlan0")
	if iface, err := net.InterfaceByName(ifaceSpec); err == nil {
		return iface, nil
	}

	// Try to parse as IP address and find the interface
	if ip := net.ParseIP(ifaceSpec); ip != nil {
		interfaces, err := net.Interfaces()
		if err != nil {
			return nil, fmt.Errorf("failed to list network interfaces: %w", err)
		}

		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					if v.IP.Equal(ip) {
						return &iface, nil
					}
				case *net.IPAddr:
					if v.IP.Equal(ip) {
						return &iface, nil
					}
				}
			}
		}
	}

	// List available interfaces for error message
	interfaces, err := net.Interfaces()
	if err != nil {
		interfaces = nil // Continue with empty list if we can't get interfaces
	}
	availableInterfaces := make([]string, 0)
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue // Skip this interface if we can't get addresses
		}
		ifaceInfo := iface.Name
		if len(addrs) > 0 {
			var addrStrs []string
			for _, addr := range addrs {
				addrStrs = append(addrStrs, addr.String())
			}
			ifaceInfo += " [" + strings.Join(addrStrs, ", ") + "]"
		}
		availableInterfaces = append(availableInterfaces, ifaceInfo)
	}

	return nil, fmt.Errorf("%w: %q. Available interfaces: %v", ErrNetworkInterfaceNotFound, ifaceSpec, availableInterfaces)
}

// ListNetworkInterfaces returns all available network interfaces with their addresses.
func ListNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list network interfaces: %w", err)
	}

	result := make([]NetworkInterface, 0, len(interfaces))
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ipAddrs []string
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ipAddrs = append(ipAddrs, v.IP.String())
			case *net.IPAddr:
				ipAddrs = append(ipAddrs, v.IP.String())
			}
		}

		result = append(result, NetworkInterface{
			Name:      iface.Name,
			Addresses: ipAddrs,
			Up:        iface.Flags&net.FlagUp != 0,
			Multicast: iface.Flags&net.FlagMulticast != 0,
		})
	}

	return result, nil
}

// NetworkInterface represents a network interface.
type NetworkInterface struct {
	// Name of the interface (e.g., "eth0", "wlan0")
	Name string

	// IP addresses assigned to this interface
	Addresses []string

	// Up indicates if the interface is up
	Up bool

	// Multicast indicates if the interface supports multicast
	Multicast bool
}

// GetDeviceEndpoint extracts the primary device endpoint from XAddrs.
func (d *Device) GetDeviceEndpoint() string {
	if len(d.XAddrs) == 0 {
		return ""
	}

	// Return the first XAddr
	return d.XAddrs[0]
}

// GetName extracts the device name from scopes.
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

// GetLocation extracts the device location from scopes.
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
