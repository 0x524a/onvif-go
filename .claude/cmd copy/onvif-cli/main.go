package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	sd "github.com/0x524A/rtspeek/pkg/rtspeek"

	"github.com/0x524a/onvif-go"
	"github.com/0x524a/onvif-go/discovery"
)

const (
	defaultTimeoutSeconds = 10
	defaultRetryDelay     = 5
	ptzTimeoutSeconds     = 30
	maxRetries            = 3
	readBufferSize        = 5
	defaultBrightness     = "50.0"
)

type CLI struct {
	client *onvif.Client
	reader *bufio.Reader
}

func main() {
	fmt.Println("🎥 ONVIF Camera CLI Tool")
	fmt.Println("=======================")
	fmt.Println()

	cli := &CLI{
		reader: bufio.NewReader(os.Stdin),
	}

	// Main menu loop
	for {
		cli.showMainMenu()
		choice := cli.readInput("Select an option: ")

		switch choice {
		case "1":
			cli.discoverCameras()
		case "2":
			cli.connectToCamera()
		case "3":
			cli.deviceOperations()
		case "4":
			cli.mediaOperations()
		case "5":
			cli.ptzOperations()
		case "6":
			cli.imagingOperations()
		case "7":
			cli.eventOperations()
		case "8":
			cli.deviceIOOperations()
		case "0", "q", "quit", "exit":
			fmt.Println("Goodbye! 👋")

			return
		default:
			fmt.Println("❌ Invalid option. Please try again.")
		}
		fmt.Println()
	}
}

func (c *CLI) showMainMenu() {
	fmt.Println("📋 Main Menu:")
	fmt.Println("  1. Discover Cameras on Network")
	fmt.Println("  2. Connect to Camera")
	if c.client != nil {
		fmt.Println("  3. Device Operations")
		fmt.Println("  4. Media Operations")
		fmt.Println("  5. PTZ Operations")
		fmt.Println("  6. Imaging Operations")
		fmt.Println("  7. Event Operations")
		fmt.Println("  8. Device IO Operations")
	} else {
		fmt.Println("  3-8. (Connect to camera first)")
	}
	fmt.Println("  0. Exit")
	fmt.Println()
}

func (c *CLI) readInput(prompt string) string {
	fmt.Print(prompt)
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	input, _ := c.reader.ReadString('\n')

	return strings.TrimSpace(input)
}

func (c *CLI) readInputWithDefault(prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	input, _ := c.reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}

	return input
}

func (c *CLI) discoverCameras() {
	fmt.Println("🔍 Discovering ONVIF cameras...")
	fmt.Println("This may take a few seconds...")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutSeconds*time.Second)
	defer cancel()

	// Try auto-discovery first (no specific interface)
	fmt.Println("⏳ Attempting auto-discovery on default interface...")
	devices, err := discovery.DiscoverWithOptions(ctx, defaultRetryDelay*time.Second, &discovery.DiscoverOptions{})

	// If auto-discovery fails or finds nothing, offer interface selection
	if err != nil || len(devices) == 0 {
		if err != nil {
			fmt.Printf("⚠️  Auto-discovery failed: %v\n", err)
		} else {
			fmt.Println("⚠️  No cameras found on default interface")
		}

		fmt.Println()
		fmt.Println("💡 Trying specific network interfaces...")
		fmt.Println()

		// Get available interfaces and let user select
		devices, err = c.discoverWithInterfaceSelection()
		if err != nil {
			fmt.Printf("❌ Discovery failed: %v\n", err)

			return
		}
	}

	if len(devices) == 0 {
		fmt.Println("❌ No ONVIF cameras found on the network")
		fmt.Println()
		fmt.Println("� Troubleshooting tips:")
		fmt.Println("   - Make sure cameras are powered on and connected to the network")
		fmt.Println("   - Verify ONVIF is enabled on the cameras")
		fmt.Println("   - Ensure you're on the same network segment as the cameras")
		fmt.Println("   - Note: ONVIF requires multicast support (not available on WiFi)")
		fmt.Println("   - Try discovering on wired Ethernet interfaces instead")

		return
	}

	fmt.Printf("✅ Found %d camera(s):\n\n", len(devices))

	for i, device := range devices {
		fmt.Printf("📹 Camera #%d:\n", i+1)
		fmt.Printf("   Endpoint: %s\n", device.GetDeviceEndpoint())

		name := device.GetName()
		if name != "" {
			fmt.Printf("   Name: %s\n", name)
		}

		location := device.GetLocation()
		if location != "" {
			fmt.Printf("   Location: %s\n", location)
		}

		fmt.Printf("   Types: %v\n", device.Types)
		fmt.Printf("   XAddrs: %v\n", device.XAddrs)
		fmt.Println()
	}

	// Ask if user wants to connect to one of the discovered cameras
	if len(devices) > 0 {
		connect := c.readInput("Do you want to connect to one of these cameras? (y/n): ")
		if strings.EqualFold(connect, "y") || strings.EqualFold(connect, "yes") {
			if len(devices) == 1 {
				c.connectToDiscoveredCamera(devices[0])
			} else {
				c.selectAndConnectCamera(devices)
			}
		}
	}
}

// discoverWithInterfaceSelection shows available network interfaces and lets user select one.
//
//nolint:gocyclo // Interface selection has high complexity due to multiple user interaction paths
func (c *CLI) discoverWithInterfaceSelection() ([]*discovery.Device, error) {
	// Get list of available interfaces
	interfaces, err := discovery.ListNetworkInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list network interfaces: %w", err)
	}

	if len(interfaces) == 0 {
		return nil, fmt.Errorf("%w", ErrNoNetworkInterfaces)
	}

	// Check how many interfaces are usable (UP and with addresses)
	activeInterfaces := make([]discovery.NetworkInterface, 0)
	for _, iface := range interfaces {
		if iface.Up && len(iface.Addresses) > 0 {
			activeInterfaces = append(activeInterfaces, iface)
		}
	}

	// If only one active interface, use it automatically
	if len(activeInterfaces) == 1 {
		fmt.Printf("📡 Using only active interface: %s\n", activeInterfaces[0].Name)

		return c.performDiscoveryOnInterface(activeInterfaces[0].Name)
	}

	// If multiple interfaces, show list for user selection
	if len(activeInterfaces) > 1 {
		fmt.Println("📡 Multiple active network interfaces detected. Trying each one...")
		fmt.Println()

		// Try each interface and collect results
		allDevices := make([]*discovery.Device, 0)
		for _, iface := range activeInterfaces {
			fmt.Printf("🔄 Scanning interface: %s\n", iface.Name)
			for _, addr := range iface.Addresses {
				fmt.Printf("   └─ %s", addr)
				if !iface.Multicast {
					fmt.Printf(" (⚠️  No multicast)")
				}
				fmt.Println()
			}

			devices, err := c.performDiscoveryOnInterface(iface.Name)
			if err == nil && len(devices) > 0 {
				fmt.Printf("   ✅ Found %d camera(s) on this interface\n", len(devices))
				allDevices = append(allDevices, devices...)
			} else {
				fmt.Println("   ❌ No cameras found")
			}
			fmt.Println()
		}

		if len(allDevices) > 0 {
			return allDevices, nil
		}

		return nil, fmt.Errorf("%w", ErrNoCamerasFound)
	}

	// If no active interfaces found
	fmt.Println("❌ No active network interfaces with assigned addresses")
	fmt.Println()
	fmt.Println("📡 All available interfaces:")
	for _, iface := range interfaces {
		upStr := "⬆️  Up"
		if !iface.Up {
			upStr = "⬇️  Down"
		}
		multicastStr := "✓"
		if !iface.Multicast {
			multicastStr = "✗"
		}
		fmt.Printf("   %s (%s, Multicast: %s)\n", iface.Name, upStr, multicastStr)
	}

	return nil, fmt.Errorf("%w", ErrNoActiveInterfaces)
}

// performDiscoveryOnInterface performs discovery on a specific network interface.
func (c *CLI) performDiscoveryOnInterface(interfaceName string) ([]*discovery.Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutSeconds*time.Second)
	defer cancel()

	opts := &discovery.DiscoverOptions{
		NetworkInterface: interfaceName,
	}

	devices, err := discovery.DiscoverWithOptions(ctx, defaultRetryDelay*time.Second, opts)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	return devices, nil
}

func (c *CLI) selectAndConnectCamera(devices []*discovery.Device) {
	fmt.Println("Select a camera to connect to:")
	for i, device := range devices {
		name := device.GetName()
		if name == "" {
			name = "Unknown"
		}
		fmt.Printf("  %d. %s (%s)\n", i+1, name, device.GetDeviceEndpoint())
	}

	choice := c.readInput("Enter camera number: ")
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(devices) {
		fmt.Println("❌ Invalid selection")

		return
	}

	c.connectToDiscoveredCamera(devices[index-1])
}

func (c *CLI) connectToDiscoveredCamera(device *discovery.Device) {
	endpoint := device.GetDeviceEndpoint()

	fmt.Printf("Connecting to: %s\n", endpoint)

	// Warn if using HTTPS
	if strings.HasPrefix(endpoint, "https://") {
		fmt.Println("⚠️  HTTPS endpoint detected - you may need to skip TLS verification for self-signed certificates")
	}

	username := c.readInputWithDefault("Username", "admin")

	fmt.Print("Password: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Ask about TLS verification only for HTTPS
	insecure := false
	if strings.HasPrefix(endpoint, "https://") {
		skipTLS := c.readInputWithDefault("Skip TLS certificate verification? (y/N)", "N")
		insecure = strings.EqualFold(skipTLS, "y") || strings.EqualFold(skipTLS, "yes")
	}

	c.createClient(endpoint, username, password, insecure)
}

func (c *CLI) connectToCamera() {
	fmt.Println("🔗 Connect to Camera")
	fmt.Println("===================")

	endpoint := c.readInputWithDefault(
		"Camera endpoint (http://ip:port/onvif/device_service)",
		"http://192.168.1.100/onvif/device_service")

	// Warn if using HTTPS
	if strings.HasPrefix(endpoint, "https://") {
		fmt.Println("⚠️  HTTPS endpoint detected - you may need to skip TLS verification for self-signed certificates")
	}

	username := c.readInputWithDefault("Username", "admin")

	fmt.Print("Password: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Ask about TLS verification only for HTTPS
	insecure := false
	if strings.HasPrefix(endpoint, "https://") {
		skipTLS := c.readInputWithDefault("Skip TLS certificate verification? (y/N)", "N")
		insecure = strings.EqualFold(skipTLS, "y") || strings.EqualFold(skipTLS, "yes")
	}

	c.createClient(endpoint, username, password, insecure)
}

func (c *CLI) createClient(endpoint, username, password string, insecure bool) {
	fmt.Println("⏳ Connecting...")

	opts := []onvif.ClientOption{
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(ptzTimeoutSeconds * time.Second),
	}

	if insecure {
		fmt.Println("⚠️  TLS certificate verification disabled")
		opts = append(opts, onvif.WithInsecureSkipVerify())
	}

	client, err := onvif.NewClient(endpoint, opts...)
	if err != nil {
		fmt.Printf("❌ Failed to create client: %v\n", err)

		return
	}

	ctx := context.Background()

	// Test connection by getting device information
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		fmt.Println("💡 Check:")
		fmt.Println("   - Endpoint URL is correct")
		fmt.Println("   - Username and password are correct")
		fmt.Println("   - Camera is accessible from this network")
		if strings.Contains(err.Error(), "tls") ||
			strings.Contains(err.Error(), "certificate") ||
			strings.Contains(err.Error(), "x509") {
			fmt.Println("   - For HTTPS cameras with self-signed certificates, answer 'y' to skip TLS verification")
		}

		return
	}

	fmt.Printf("✅ Connected successfully!\n")
	fmt.Printf("📹 Camera: %s %s\n", info.Manufacturer, info.Model)
	fmt.Printf("🔧 Firmware: %s\n", info.FirmwareVersion)

	// Initialize to discover service endpoints
	fmt.Println("⏳ Discovering services...")
	if err := client.Initialize(ctx); err != nil {
		fmt.Printf("⚠️  Service discovery failed: %v\n", err)
		fmt.Println("Some features may not be available.")
	} else {
		fmt.Println("✅ Services discovered")
	}

	c.client = client
}

func (c *CLI) deviceOperations() {
	if c.client == nil {
		fmt.Println("❌ Not connected to any camera")

		return
	}

	fmt.Println("🔧 Device Operations")
	fmt.Println("===================")
	fmt.Println("  1. Get Device Information")
	fmt.Println("  2. Get Capabilities")
	fmt.Println("  3. Get System Date and Time")
	fmt.Println("  4. Reboot Device")
	fmt.Println("  0. Back to Main Menu")

	choice := c.readInput("Select operation: ")
	ctx := context.Background()

	switch choice {
	case "1":
		c.getDeviceInformation(ctx)
	case "2":
		c.getCapabilities(ctx)
	case "3":
		c.getSystemDateTime(ctx)
	case "4":
		c.rebootDevice(ctx)
	case "0":
		return
	default:
		fmt.Println("❌ Invalid option")
	}
}

func (c *CLI) getDeviceInformation(ctx context.Context) {
	fmt.Println("⏳ Getting device information...")

	info, err := c.client.GetDeviceInformation(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Device Information:")
	fmt.Printf("   Manufacturer: %s\n", info.Manufacturer)
	fmt.Printf("   Model: %s\n", info.Model)
	fmt.Printf("   Firmware Version: %s\n", info.FirmwareVersion)
	fmt.Printf("   Serial Number: %s\n", info.SerialNumber)
	fmt.Printf("   Hardware ID: %s\n", info.HardwareID)
}

func (c *CLI) getCapabilities(ctx context.Context) {
	fmt.Println("⏳ Getting capabilities...")

	caps, err := c.client.GetCapabilities(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Device Capabilities:")

	if caps.Device != nil {
		fmt.Printf("   ✓ Device Service\n")
	}
	if caps.Media != nil {
		fmt.Printf("   ✓ Media Service (Streaming)\n")
	}
	if caps.PTZ != nil {
		fmt.Printf("   ✓ PTZ Service (Pan/Tilt/Zoom)\n")
	}
	if caps.Imaging != nil {
		fmt.Printf("   ✓ Imaging Service\n")
	}
	if caps.Events != nil {
		fmt.Printf("   ✓ Event Service\n")
	}
	if caps.Analytics != nil {
		fmt.Printf("   ✓ Analytics Service\n")
	}
}

func (c *CLI) getSystemDateTime(ctx context.Context) {
	fmt.Println("⏳ Getting system date and time...")

	dateTime, err := c.client.GetSystemDateAndTime(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ System Date/Time: %v\n", dateTime)
}

func (c *CLI) rebootDevice(ctx context.Context) {
	confirm := c.readInput("⚠️  Are you sure you want to reboot the device? (y/N): ")
	if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
		fmt.Println("Reboot canceled")

		return
	}

	fmt.Println("⏳ Rebooting device...")

	message, err := c.client.SystemReboot(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Reboot initiated: %s\n", message)
	fmt.Println("💡 The camera will be unavailable for a few minutes")
}

func (c *CLI) mediaOperations() {
	if c.client == nil {
		fmt.Println("❌ Not connected to any camera")

		return
	}

	fmt.Println("🎬 Media Operations")
	fmt.Println("==================")
	fmt.Println("  1. Get Media Profiles")
	fmt.Println("  2. Get Stream URIs")
	fmt.Println("  3. Get Snapshot URIs")
	fmt.Println("  4. Get Video Encoder Configuration")
	fmt.Println("  0. Back to Main Menu")

	choice := c.readInput("Select operation: ")
	ctx := context.Background()

	switch choice {
	case "1":
		c.getMediaProfiles(ctx)
	case "2":
		c.getStreamURIs(ctx)
	case "3":
		c.getSnapshotURIs(ctx)
	case "4":
		c.getVideoEncoderConfig(ctx)
	case "0":
		return
	default:
		fmt.Println("❌ Invalid option")
	}
}

func (c *CLI) getMediaProfiles(ctx context.Context) {
	fmt.Println("⏳ Getting media profiles...")

	profiles, err := c.client.GetProfiles(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Found %d profile(s):\n\n", len(profiles))

	for i, profile := range profiles {
		fmt.Printf("📹 Profile #%d: %s\n", i+1, profile.Name)
		fmt.Printf("   Token: %s\n", profile.Token)

		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("   Video Encoding: %s\n", profile.VideoEncoderConfiguration.Encoding)
			if profile.VideoEncoderConfiguration.Resolution != nil {
				fmt.Printf("   Resolution: %dx%d\n",
					profile.VideoEncoderConfiguration.Resolution.Width,
					profile.VideoEncoderConfiguration.Resolution.Height)
			}
			fmt.Printf("   Quality: %.1f\n", profile.VideoEncoderConfiguration.Quality)
		}

		if profile.PTZConfiguration != nil {
			fmt.Printf("   PTZ: Enabled\n")
		}

		fmt.Println()
	}
}

// inspectRTSPStream probes an RTSP URI to get stream details using rtspeek library.
func (c *CLI) inspectRTSPStream(streamURI string) map[string]interface{} {
	details := map[string]interface{}{
		"uri":        streamURI,
		"reachable":  false,
		"codec":      "unknown",
		"resolution": "unknown",
	}

	// Use rtspeek library for detailed stream inspection
	ctx, cancel := context.WithTimeout(
		context.Background(),
		defaultRetryDelay*time.Second,
	)
	defer cancel()

	streamInfo, err := sd.DescribeStream(
		ctx, streamURI, defaultRetryDelay*time.Second,
	)
	if err == nil && streamInfo != nil {
		details["reachable"] = streamInfo.IsReachable()

		if streamInfo.IsDescribeSucceeded() && streamInfo.HasVideo() {
			// Extract codec information from first video media
			if firstVideo := streamInfo.GetFirstVideoMedia(); firstVideo != nil {
				// Get codec format (H264, H265, MJPEG, etc.)
				details["codec"] = firstVideo.Format

				// Extract resolution directly from the video media
				if firstVideo.Resolution != nil {
					details["resolution"] = fmt.Sprintf("%dx%d",
						firstVideo.Resolution.Width,
						firstVideo.Resolution.Height)
				} else {
					// Fallback to resolution strings
					resolutions := streamInfo.GetVideoResolutionStrings()
					if len(resolutions) > 0 {
						details["resolution"] = resolutions[0]
					}
				}
			}

			return details
		}

		// Describe failed but connection was reachable - try TCP fallback
		if streamInfo.IsReachable() {
			details["reachable"] = true

			return details
		}
	}

	// Fallback: try basic TCP connection to RTSP port for connectivity check
	if details := c.tryRTSPConnection(streamURI); details != nil {
		return details
	}

	return details
}

// tryRTSPConnection attempts to connect to RTSP port and grab basic info.
func (c *CLI) tryRTSPConnection(streamURI string) map[string]interface{} {
	details := map[string]interface{}{
		"uri":       streamURI,
		"reachable": false,
	}

	// Parse URL to get host and port
	rtspURL := streamURI
	if !strings.HasPrefix(rtspURL, "rtsp://") {
		return details
	}

	// Extract host:port from rtsp://host:port/path
	parts := strings.TrimPrefix(rtspURL, "rtsp://")
	hostParts := strings.Split(parts, "/")
	hostPort := hostParts[0]

	// Default RTSP port if not specified
	if !strings.Contains(hostPort, ":") {
		hostPort += ":554"
	}

	// Try to connect
	conn, err := net.DialTimeout("tcp", hostPort, maxRetries*time.Second)
	if err == nil {
		_ = conn.Close()
		details["reachable"] = true
		details["port"] = strings.Split(hostPort, ":")[1]

		return details
	}

	return details
}

func (c *CLI) getStreamURIs(ctx context.Context) {
	profiles, err := c.client.GetProfiles(ctx)
	if err != nil {
		fmt.Printf("❌ Error getting profiles: %v\n", err)

		return
	}

	if len(profiles) == 0 {
		fmt.Println("❌ No profiles found")

		return
	}

	fmt.Println("📡 Stream URIs:")
	fmt.Println()

	for i, profile := range profiles {
		fmt.Printf("Profile #%d: %s\n", i+1, profile.Name)

		streamURI, err := c.client.GetStreamURI(ctx, profile.Token)
		if err != nil {
			fmt.Printf("   Stream URI: ❌ Error - %v\n", err)
		} else {
			fmt.Printf("   Stream URI: %s\n", streamURI.URI)

			// Warn if camera returns HTTPS when we connected via HTTP
			if strings.HasPrefix(c.client.Endpoint(), "http://") && strings.HasPrefix(streamURI.URI, "https://") {
				fmt.Printf("   ⚠️  WARNING: Camera returned HTTPS URL but you connected via HTTP\n")
				fmt.Printf("   💡 Stream may fail due to TLS certificate issues\n")
				fmt.Printf("   💡 Consider reconnecting with https:// endpoint and skip TLS verification\n")
			}

			// Inspect RTSP stream details
			fmt.Print("   ⏳ Inspecting stream details...")
			details := c.inspectRTSPStream(streamURI.URI)
			fmt.Print("\r")
			fmt.Print("   ✅ Stream inspection complete  \n")

			// Display stream details
			if reachable, ok := details["reachable"].(bool); ok && reachable {
				fmt.Printf("      Status: ✅ Stream is reachable\n")
			} else {
				fmt.Printf("      Status: ⚠️  Stream connectivity check skipped\n")
			}

			if codec, ok := details["codec"].(string); ok && codec != "unknown" {
				fmt.Printf("      Video Codec: %s\n", codec)
			}

			if resolution, ok := details["resolution"].(string); ok && resolution != "unknown" {
				fmt.Printf("      Resolution: %s\n", resolution)
			}

			if port, ok := details["port"].(string); ok {
				fmt.Printf("      RTSP Port: %s\n", port)
			}

			fmt.Printf("   📱 Use this URL in VLC or other RTSP player\n")
		}
		fmt.Println()
	}
}

func (c *CLI) getSnapshotURIs(ctx context.Context) {
	profiles, err := c.client.GetProfiles(ctx)
	if err != nil {
		fmt.Printf("❌ Error getting profiles: %v\n", err)

		return
	}

	if len(profiles) == 0 {
		fmt.Println("❌ No profiles found")

		return
	}

	fmt.Println("📸 Snapshot URIs:")
	fmt.Println()

	for i, profile := range profiles {
		fmt.Printf("Profile #%d: %s\n", i+1, profile.Name)

		snapshotURI, err := c.client.GetSnapshotURI(ctx, profile.Token)
		if err != nil {
			fmt.Printf("   Snapshot URI: ❌ Error - %v\n", err)
		} else {
			fmt.Printf("   Snapshot URI: %s\n", snapshotURI.URI)

			// Warn if camera returns HTTPS when we connected via HTTP
			if strings.HasPrefix(c.client.Endpoint(), "http://") && strings.HasPrefix(snapshotURI.URI, "https://") {
				fmt.Printf("   ⚠️  WARNING: Camera returned HTTPS URL but you connected via HTTP\n")
				fmt.Printf("   💡 Snapshot may fail due to TLS certificate issues\n")
				fmt.Printf("   💡 Consider reconnecting with https:// endpoint and skip TLS verification\n")
			}

			fmt.Printf("   🌐 Open this URL in a browser to see the snapshot\n")
		}
		fmt.Println()
	}
}

func (c *CLI) getVideoEncoderConfig(ctx context.Context) {
	profiles, err := c.client.GetProfiles(ctx)
	if err != nil {
		fmt.Printf("❌ Error getting profiles: %v\n", err)

		return
	}

	if len(profiles) == 0 {
		fmt.Println("❌ No profiles found")

		return
	}

	fmt.Println("Available profiles:")
	for i, profile := range profiles {
		fmt.Printf("  %d. %s\n", i+1, profile.Name)
	}

	choice := c.readInput("Select profile number: ")
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(profiles) {
		fmt.Println("❌ Invalid selection")

		return
	}

	profile := profiles[index-1]
	if profile.VideoEncoderConfiguration == nil {
		fmt.Println("❌ No video encoder configuration found")

		return
	}

	fmt.Println("⏳ Getting video encoder configuration...")

	config, err := c.client.GetVideoEncoderConfiguration(ctx, profile.VideoEncoderConfiguration.Token)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Video Encoder Configuration:\n")
	fmt.Printf("   Name: %s\n", config.Name)
	fmt.Printf("   Token: %s\n", config.Token)
	fmt.Printf("   Use Count: %d\n", config.UseCount)
	fmt.Printf("   Encoding: %s\n", config.Encoding)

	if config.Resolution != nil {
		fmt.Printf("   Resolution: %dx%d\n", config.Resolution.Width, config.Resolution.Height)
	}

	fmt.Printf("   Quality: %.1f\n", config.Quality)

	if config.RateControl != nil {
		fmt.Printf("   Frame Rate Limit: %d\n", config.RateControl.FrameRateLimit)
		fmt.Printf("   Encoding Interval: %d\n", config.RateControl.EncodingInterval)
		fmt.Printf("   Bitrate Limit: %d\n", config.RateControl.BitrateLimit)
	}
}

func (c *CLI) ptzOperations() {
	if c.client == nil {
		fmt.Println("❌ Not connected to any camera")

		return
	}

	fmt.Println("🎮 PTZ Operations")
	fmt.Println("================")
	fmt.Println("  1. Get PTZ Status")
	fmt.Println("  2. Continuous Move")
	fmt.Println("  3. Absolute Move")
	fmt.Println("  4. Relative Move")
	fmt.Println("  5. Stop Movement")
	fmt.Println("  6. Get Presets")
	fmt.Println("  7. Go to Preset")
	fmt.Println("  0. Back to Main Menu")

	choice := c.readInput("Select operation: ")
	ctx := context.Background()

	// Get profile token for PTZ operations
	profileToken, err := c.getPTZProfileToken(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	switch choice {
	case "1":
		c.getPTZStatus(ctx, profileToken)
	case "2":
		c.continuousMove(ctx, profileToken)
	case "3":
		c.absoluteMove(ctx, profileToken)
	case "4":
		c.relativeMove(ctx, profileToken)
	case "5":
		c.stopMovement(ctx, profileToken)
	case "6":
		c.getPTZPresets(ctx, profileToken)
	case "7":
		c.gotoPreset(ctx, profileToken)
	case "0":
		return
	default:
		fmt.Println("❌ Invalid option")
	}
}

func (c *CLI) getPTZProfileToken(ctx context.Context) (string, error) {
	profiles, err := c.client.GetProfiles(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get profiles: %w", err)
	}

	if len(profiles) == 0 {
		return "", fmt.Errorf("%w", ErrNoProfilesFound)
	}

	// Find a profile with PTZ configuration
	for _, profile := range profiles {
		if profile.PTZConfiguration != nil {
			return profile.Token, nil
		}
	}

	// If no PTZ profile found, use the first profile
	fmt.Println("⚠️  No PTZ-specific profile found, using first profile")

	return profiles[0].Token, nil
}

func (c *CLI) getPTZStatus(ctx context.Context, profileToken string) {
	fmt.Println("⏳ Getting PTZ status...")

	status, err := c.client.GetStatus(ctx, profileToken)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("💡 PTZ might not be supported on this camera")

		return
	}

	fmt.Println("✅ PTZ Status:")

	if status.Position != nil {
		if status.Position.PanTilt != nil {
			fmt.Printf("   Pan: %.3f\n", status.Position.PanTilt.X)
			fmt.Printf("   Tilt: %.3f\n", status.Position.PanTilt.Y)
		}
		if status.Position.Zoom != nil {
			fmt.Printf("   Zoom: %.3f\n", status.Position.Zoom.X)
		}
	}

	if status.MoveStatus != nil {
		fmt.Printf("   Pan/Tilt Status: %s\n", status.MoveStatus.PanTilt)
		fmt.Printf("   Zoom Status: %s\n", status.MoveStatus.Zoom)
	}

	if status.Error != "" {
		fmt.Printf("   Error: %s\n", status.Error)
	}
}

func (c *CLI) continuousMove(ctx context.Context, profileToken string) {
	fmt.Println("🎮 Continuous Move")
	fmt.Println("Pan/Tilt values: -1.0 to 1.0 (negative = left/down, positive = right/up)")
	fmt.Println("Zoom values: -1.0 to 1.0 (negative = zoom out, positive = zoom in)")

	panStr := c.readInputWithDefault("Pan speed (-1.0 to 1.0)", "0.0")
	tiltStr := c.readInputWithDefault("Tilt speed (-1.0 to 1.0)", "0.0")
	zoomStr := c.readInputWithDefault("Zoom speed (-1.0 to 1.0)", "0.0")
	timeoutStr := c.readInputWithDefault("Timeout (seconds)", "2")

	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	pan, _ := strconv.ParseFloat(panStr, 64)
	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	tilt, _ := strconv.ParseFloat(tiltStr, 64)
	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	zoom, _ := strconv.ParseFloat(zoomStr, 64)

	velocity := &onvif.PTZSpeed{
		PanTilt: &onvif.Vector2D{X: pan, Y: tilt},
		Zoom:    &onvif.Vector1D{X: zoom},
	}

	timeout := fmt.Sprintf("PT%sS", timeoutStr)

	fmt.Println("⏳ Moving camera...")

	err := c.client.ContinuousMove(ctx, profileToken, velocity, &timeout)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Movement started")
}

func (c *CLI) absoluteMove(ctx context.Context, profileToken string) {
	fmt.Println("🎯 Absolute Move")
	fmt.Println("Position values: -1.0 to 1.0")

	panStr := c.readInputWithDefault("Pan position (-1.0 to 1.0)", "0.0")
	tiltStr := c.readInputWithDefault("Tilt position (-1.0 to 1.0)", "0.0")
	zoomStr := c.readInputWithDefault("Zoom position (-1.0 to 1.0)", "0.0")

	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	pan, _ := strconv.ParseFloat(panStr, 64)
	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	tilt, _ := strconv.ParseFloat(tiltStr, 64)
	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	zoom, _ := strconv.ParseFloat(zoomStr, 64)

	position := &onvif.PTZVector{
		PanTilt: &onvif.Vector2D{X: pan, Y: tilt},
		Zoom:    &onvif.Vector1D{X: zoom},
	}

	fmt.Println("⏳ Moving to position...")

	err := c.client.AbsoluteMove(ctx, profileToken, position, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Moving to absolute position")
}

func (c *CLI) relativeMove(ctx context.Context, profileToken string) {
	fmt.Println("↗️ Relative Move")
	fmt.Println("Translation values: -1.0 to 1.0 (relative to current position)")

	panStr := c.readInputWithDefault("Pan translation (-1.0 to 1.0)", "0.0")
	tiltStr := c.readInputWithDefault("Tilt translation (-1.0 to 1.0)", "0.0")
	zoomStr := c.readInputWithDefault("Zoom translation (-1.0 to 1.0)", "0.0")

	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	pan, _ := strconv.ParseFloat(panStr, 64)
	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	tilt, _ := strconv.ParseFloat(tiltStr, 64)
	//nolint:errcheck // ParseFloat errors default to 0.0 which is acceptable for CLI input
	zoom, _ := strconv.ParseFloat(zoomStr, 64)

	translation := &onvif.PTZVector{
		PanTilt: &onvif.Vector2D{X: pan, Y: tilt},
		Zoom:    &onvif.Vector1D{X: zoom},
	}

	fmt.Println("⏳ Moving relative to current position...")

	err := c.client.RelativeMove(ctx, profileToken, translation, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Moving relative to current position")
}

func (c *CLI) stopMovement(ctx context.Context, profileToken string) {
	stopPanTilt := c.readInputWithDefault("Stop Pan/Tilt? (y/n)", "y")
	stopZoom := c.readInputWithDefault("Stop Zoom? (y/n)", "y")

	panTilt := strings.EqualFold(stopPanTilt, "y") || strings.EqualFold(stopPanTilt, "yes")
	zoom := strings.EqualFold(stopZoom, "y") || strings.EqualFold(stopZoom, "yes")

	fmt.Println("⏳ Stopping movement...")

	err := c.client.Stop(ctx, profileToken, panTilt, zoom)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Movement stopped")
}

func (c *CLI) getPTZPresets(ctx context.Context, profileToken string) {
	fmt.Println("⏳ Getting PTZ presets...")

	presets, err := c.client.GetPresets(ctx, profileToken)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(presets) == 0 {
		fmt.Println("📝 No presets found")

		return
	}

	fmt.Printf("✅ Found %d preset(s):\n\n", len(presets))

	for i, preset := range presets {
		fmt.Printf("📍 Preset #%d:\n", i+1)
		fmt.Printf("   Name: %s\n", preset.Name)
		fmt.Printf("   Token: %s\n", preset.Token)

		if preset.PTZPosition != nil {
			if preset.PTZPosition.PanTilt != nil {
				fmt.Printf("   Pan: %.3f, Tilt: %.3f\n",
					preset.PTZPosition.PanTilt.X,
					preset.PTZPosition.PanTilt.Y)
			}
			if preset.PTZPosition.Zoom != nil {
				fmt.Printf("   Zoom: %.3f\n", preset.PTZPosition.Zoom.X)
			}
		}
		fmt.Println()
	}
}

func (c *CLI) gotoPreset(ctx context.Context, profileToken string) {
	presets, err := c.client.GetPresets(ctx, profileToken)
	if err != nil {
		fmt.Printf("❌ Error getting presets: %v\n", err)

		return
	}

	if len(presets) == 0 {
		fmt.Println("📝 No presets available")

		return
	}

	fmt.Println("Available presets:")
	for i, preset := range presets {
		fmt.Printf("  %d. %s\n", i+1, preset.Name)
	}

	choice := c.readInput("Select preset number: ")
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(presets) {
		fmt.Println("❌ Invalid selection")

		return
	}

	preset := presets[index-1]

	fmt.Printf("⏳ Going to preset '%s'...\n", preset.Name)

	err = c.client.GotoPreset(ctx, profileToken, preset.Token, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Moving to preset '%s'\n", preset.Name)
}

func (c *CLI) imagingOperations() {
	if c.client == nil {
		fmt.Println("❌ Not connected to any camera")

		return
	}

	fmt.Println("🎨 Imaging Operations")
	fmt.Println("====================")
	fmt.Println("  1. Get Imaging Settings")
	fmt.Println("  2. Set Brightness")
	fmt.Println("  3. Set Contrast")
	fmt.Println("  4. Set Saturation")
	fmt.Println("  5. Set Sharpness")
	fmt.Println("  6. Advanced Settings")
	fmt.Println("  7. Capture Snapshot (ASCII Preview)")
	fmt.Println("  0. Back to Main Menu")

	choice := c.readInput("Select operation: ")
	ctx := context.Background()

	// Get video source token
	videoSourceToken, err := c.getVideoSourceToken(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	switch choice {
	case "1":
		c.getImagingSettings(ctx, videoSourceToken)
	case "2":
		c.setBrightness(ctx, videoSourceToken)
	case "3":
		c.setContrast(ctx, videoSourceToken)
	case "4":
		c.setSaturation(ctx, videoSourceToken)
	case "5":
		c.setSharpness(ctx, videoSourceToken)
	case "6":
		c.advancedImagingSettings(ctx, videoSourceToken)
	case "7":
		c.captureAndDisplaySnapshot(ctx)
	case "0":
		return
	default:
		fmt.Println("❌ Invalid option")
	}
}

func (c *CLI) getVideoSourceToken(ctx context.Context) (string, error) {
	profiles, err := c.client.GetProfiles(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get profiles: %w", err)
	}

	if len(profiles) == 0 {
		return "", fmt.Errorf("%w", ErrNoProfilesFound)
	}

	for _, profile := range profiles {
		if profile.VideoSourceConfiguration != nil {
			return profile.VideoSourceConfiguration.SourceToken, nil
		}
	}

	return "", fmt.Errorf("%w", ErrNoVideoSourceConfiguration)
}

func (c *CLI) getImagingSettings(ctx context.Context, videoSourceToken string) {
	fmt.Println("⏳ Getting imaging settings...")

	settings, err := c.client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Current Imaging Settings:")

	if settings.Brightness != nil {
		fmt.Printf("   Brightness: %.1f\n", *settings.Brightness)
	}
	if settings.Contrast != nil {
		fmt.Printf("   Contrast: %.1f\n", *settings.Contrast)
	}
	if settings.ColorSaturation != nil {
		fmt.Printf("   Saturation: %.1f\n", *settings.ColorSaturation)
	}
	if settings.Sharpness != nil {
		fmt.Printf("   Sharpness: %.1f\n", *settings.Sharpness)
	}
	if settings.IrCutFilter != nil {
		fmt.Printf("   IR Cut Filter: %s\n", *settings.IrCutFilter)
	}

	if settings.Exposure != nil {
		fmt.Printf("   Exposure Mode: %s\n", settings.Exposure.Mode)
		if settings.Exposure.Mode == "MANUAL" {
			fmt.Printf("     Exposure Time: %.2f\n", settings.Exposure.ExposureTime)
			fmt.Printf("     Gain: %.2f\n", settings.Exposure.Gain)
		}
	}

	if settings.Focus != nil {
		fmt.Printf("   Focus Mode: %s\n", settings.Focus.AutoFocusMode)
	}

	if settings.WhiteBalance != nil {
		fmt.Printf("   White Balance: %s\n", settings.WhiteBalance.Mode)
	}

	if settings.WideDynamicRange != nil {
		fmt.Printf("   WDR Mode: %s\n", settings.WideDynamicRange.Mode)
		fmt.Printf("   WDR Level: %.1f\n", settings.WideDynamicRange.Level)
	}
}

func (c *CLI) setBrightness(ctx context.Context, videoSourceToken string) {
	currentSettings, err := c.client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		fmt.Printf("❌ Error getting current settings: %v\n", err)

		return
	}

	currentValue := defaultBrightness
	if currentSettings.Brightness != nil {
		currentValue = fmt.Sprintf("%.1f", *currentSettings.Brightness)
	}

	brightnessStr := c.readInputWithDefault(fmt.Sprintf("Brightness (0-100, current: %s)", currentValue), currentValue)
	brightness, err := strconv.ParseFloat(brightnessStr, 64)
	if err != nil {
		fmt.Println("❌ Invalid brightness value")

		return
	}

	currentSettings.Brightness = &brightness

	fmt.Println("⏳ Setting brightness...")

	err = c.client.SetImagingSettings(ctx, videoSourceToken, currentSettings, true)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Brightness set to %.1f\n", brightness)
}

func (c *CLI) setContrast(ctx context.Context, videoSourceToken string) {
	currentSettings, err := c.client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		fmt.Printf("❌ Error getting current settings: %v\n", err)

		return
	}

	currentValue := defaultBrightness
	if currentSettings.Contrast != nil {
		currentValue = fmt.Sprintf("%.1f", *currentSettings.Contrast)
	}

	contrastStr := c.readInputWithDefault(fmt.Sprintf("Contrast (0-100, current: %s)", currentValue), currentValue)
	contrast, err := strconv.ParseFloat(contrastStr, 64)
	if err != nil {
		fmt.Println("❌ Invalid contrast value")

		return
	}

	currentSettings.Contrast = &contrast

	fmt.Println("⏳ Setting contrast...")

	err = c.client.SetImagingSettings(ctx, videoSourceToken, currentSettings, true)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Contrast set to %.1f\n", contrast)
}

func (c *CLI) setSaturation(ctx context.Context, videoSourceToken string) {
	currentSettings, err := c.client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		fmt.Printf("❌ Error getting current settings: %v\n", err)

		return
	}

	currentValue := defaultBrightness
	if currentSettings.ColorSaturation != nil {
		currentValue = fmt.Sprintf("%.1f", *currentSettings.ColorSaturation)
	}

	saturationStr := c.readInputWithDefault(fmt.Sprintf("Saturation (0-100, current: %s)", currentValue), currentValue)
	saturation, err := strconv.ParseFloat(saturationStr, 64)
	if err != nil {
		fmt.Println("❌ Invalid saturation value")

		return
	}

	currentSettings.ColorSaturation = &saturation

	fmt.Println("⏳ Setting saturation...")

	err = c.client.SetImagingSettings(ctx, videoSourceToken, currentSettings, true)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Saturation set to %.1f\n", saturation)
}

func (c *CLI) setSharpness(ctx context.Context, videoSourceToken string) {
	currentSettings, err := c.client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		fmt.Printf("❌ Error getting current settings: %v\n", err)

		return
	}

	currentValue := defaultBrightness
	if currentSettings.Sharpness != nil {
		currentValue = fmt.Sprintf("%.1f", *currentSettings.Sharpness)
	}

	sharpnessStr := c.readInputWithDefault(fmt.Sprintf("Sharpness (0-100, current: %s)", currentValue), currentValue)
	sharpness, err := strconv.ParseFloat(sharpnessStr, 64)
	if err != nil {
		fmt.Println("❌ Invalid sharpness value")

		return
	}

	currentSettings.Sharpness = &sharpness

	fmt.Println("⏳ Setting sharpness...")

	err = c.client.SetImagingSettings(ctx, videoSourceToken, currentSettings, true)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Sharpness set to %.1f\n", sharpness)
}

func (c *CLI) advancedImagingSettings(ctx context.Context, videoSourceToken string) {
	fmt.Println("🔧 Advanced Imaging Settings")
	fmt.Println("This feature allows you to modify multiple settings at once")
	fmt.Println("Leave empty to keep current value")

	currentSettings, err := c.client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		fmt.Printf("❌ Error getting current settings: %v\n", err)

		return
	}

	// Show current values and ask for new ones
	fmt.Println("\nCurrent settings:")
	c.getImagingSettings(ctx, videoSourceToken)
	fmt.Println()

	if input := c.readInput("New brightness (0-100, empty to keep current): "); input != "" {
		if val, err := strconv.ParseFloat(input, 64); err == nil {
			currentSettings.Brightness = &val
		}
	}

	if input := c.readInput("New contrast (0-100, empty to keep current): "); input != "" {
		if val, err := strconv.ParseFloat(input, 64); err == nil {
			currentSettings.Contrast = &val
		}
	}

	if input := c.readInput("New saturation (0-100, empty to keep current): "); input != "" {
		if val, err := strconv.ParseFloat(input, 64); err == nil {
			currentSettings.ColorSaturation = &val
		}
	}

	if input := c.readInput("New sharpness (0-100, empty to keep current): "); input != "" {
		if val, err := strconv.ParseFloat(input, 64); err == nil {
			currentSettings.Sharpness = &val
		}
	}

	confirm := c.readInput("Apply these settings? (y/N): ")
	if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
		fmt.Println("Settings not applied")

		return
	}

	fmt.Println("⏳ Applying settings...")

	err = c.client.SetImagingSettings(ctx, videoSourceToken, currentSettings, true)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Settings applied successfully!")
	fmt.Println("\nNew settings:")
	c.getImagingSettings(ctx, videoSourceToken)
}

//nolint:gocyclo // Snapshot capture and display has high complexity due to multiple error handling paths
func (c *CLI) captureAndDisplaySnapshot(ctx context.Context) { //nolint:funlen // Many statements due to error handling
	fmt.Println("📷 Capture Snapshot as ASCII Preview")
	fmt.Println("===================================")
	fmt.Println()

	// Get media profiles to find snapshot URI
	profiles, err := c.client.GetProfiles(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get profiles: %v\n", err)

		return
	}

	if len(profiles) == 0 {
		fmt.Println("❌ No profiles found")

		return
	}

	profile := profiles[0]

	fmt.Println("⏳ Getting snapshot URI...")

	// Get snapshot URI from camera
	snapshotURI, err := c.client.GetSnapshotURI(ctx, profile.Token)
	if err != nil {
		fmt.Printf("❌ Failed to get snapshot URI: %v\n", err)

		return
	}

	if snapshotURI == nil || snapshotURI.URI == "" {
		fmt.Println("❌ No snapshot URI available")

		return
	}

	fmt.Printf("📸 Snapshot URI: %s\n", snapshotURI.URI)
	fmt.Println()

	// Display ASCII preview with quality options
	fmt.Println("Select preview quality:")
	fmt.Println("  1. Low (60 chars wide, faster)")
	fmt.Println("  2. Medium (100 chars wide, balanced)")
	fmt.Println("  3. High (140 chars wide, detailed)")
	fmt.Println("  4. Block characters (compact)")

	choice := c.readInput("Select quality (1-4) [2]: ")
	if choice == "" {
		choice = "2"
	}

	config := DefaultASCIIConfig()
	switch choice {
	case "1":
		config.Width = 60
		config.Height = 20
		config.Quality = "low"
	case "2":
		config.Width = 100
		config.Height = 30
		config.Quality = defaultQuality
	case "3":
		config.Width = 140
		config.Height = 40
		config.Quality = "high"
	case "4":
		config.Width = 100
		config.Height = 30
		config.Quality = "block"
	default:
		config.Width = 100
		config.Height = 30
		config.Quality = defaultQuality
	}

	// Download actual snapshot
	fmt.Println("⏳ Downloading snapshot...")
	snapshotData, err := c.client.DownloadFile(ctx, snapshotURI.URI)
	if err != nil {
		fmt.Printf("❌ Failed to download snapshot: %v\n", err)
		fmt.Println("\n💡 Try using curl directly:")
		fmt.Printf("   curl -u username:password '%s' > snapshot.jpg\n", snapshotURI.URI)

		return
	}

	fmt.Printf("✅ Snapshot downloaded (%d bytes)\n", len(snapshotData))
	fmt.Println()

	// Convert to ASCII
	fmt.Println("⏳ Converting to ASCII art...")
	asciiArt, err := ImageToASCII(snapshotData, config)
	if err != nil {
		fmt.Printf("❌ Failed to convert image: %v\n", err)
		fmt.Println("\n💡 Image might not be JPEG/PNG. Try downloading manually:")
		fmt.Printf("   curl -u username:password '%s' > snapshot.jpg\n", snapshotURI.URI)

		return
	}

	// Detect image format and get dimensions
	format := "JPEG"
	if bytes.Contains(snapshotData[:20], []byte("\x89PNG")) {
		format = "PNG"
	}

	imageInfo := ImageInfo{
		SizeBytes:   int64(len(snapshotData)),
		Format:      format,
		CaptureTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	output := FormatASCIIOutput(asciiArt, imageInfo)
	fmt.Print(output)

	// Offer to save the snapshot
	fmt.Println()
	save := c.readInput("💾 Save snapshot to file? (y/n) [n]: ")
	if strings.EqualFold(save, "y") {
		filename := c.readInput("📝 Filename [snapshot.jpg]: ")
		if filename == "" {
			filename = "snapshot.jpg"
		}
		if err := os.WriteFile(
			filename, snapshotData, 0600, //nolint:mnd // 0600 appropriate for CLI output files
		); err != nil {
			fmt.Printf("❌ Failed to save file: %v\n", err)
		} else {
			fmt.Printf("✅ Snapshot saved to %s\n", filename)
		}
	}
}

// ============================================
// Event Operations
// ============================================

func (c *CLI) eventOperations() {
	if c.client == nil {
		fmt.Println("❌ Not connected to any camera")

		return
	}

	fmt.Println("📡 Event Operations")
	fmt.Println("==================")
	fmt.Println("  1. Get Event Service Capabilities")
	fmt.Println("  2. Get Event Properties")
	fmt.Println("  3. Create Pull Point Subscription")
	fmt.Println("  4. Get Event Brokers")
	fmt.Println("  0. Back to Main Menu")

	choice := c.readInput("Select operation: ")
	ctx := context.Background()

	switch choice {
	case "1":
		c.getEventServiceCapabilities(ctx)
	case "2":
		c.getEventProperties(ctx)
	case "3":
		c.createPullPointSubscription(ctx)
	case "4":
		c.getEventBrokers(ctx)
	case "0":
		return
	default:
		fmt.Println("❌ Invalid option")
	}
}

func (c *CLI) getEventServiceCapabilities(ctx context.Context) {
	fmt.Println("⏳ Getting event service capabilities...")

	caps, err := c.client.GetEventServiceCapabilities(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Event Service Capabilities:")
	fmt.Printf("   WS Subscription Policy Support: %v\n", caps.WSSubscriptionPolicySupport)
	fmt.Printf("   WS Pausable Subscription: %v\n", caps.WSPausableSubscriptionManagerInterfaceSupport)
	fmt.Printf("   Max Notification Producers: %d\n", caps.MaxNotificationProducers)
	fmt.Printf("   Max Pull Points: %d\n", caps.MaxPullPoints)
	fmt.Printf("   Persistent Notification Storage: %v\n", caps.PersistentNotificationStorage)
	fmt.Printf("   Event Broker Protocols: %v\n", caps.EventBrokerProtocols)
	fmt.Printf("   Max Event Brokers: %d\n", caps.MaxEventBrokers)
	fmt.Printf("   Metadata Over MQTT: %v\n", caps.MetadataOverMQTT)
}

func (c *CLI) getEventProperties(ctx context.Context) {
	fmt.Println("⏳ Getting event properties...")

	props, err := c.client.GetEventProperties(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Event Properties:")
	fmt.Printf("   Fixed Topic Set: %v\n", props.FixedTopicSet)
	fmt.Printf("   Topic Namespace Locations: %d\n", len(props.TopicNamespaceLocation))
	for i, loc := range props.TopicNamespaceLocation {
		fmt.Printf("     %d. %s\n", i+1, loc)
	}
	fmt.Printf("   Topic Expression Dialects: %d\n", len(props.TopicExpressionDialects))
	fmt.Printf("   Message Content Filter Dialects: %d\n", len(props.MessageContentFilterDialects))
}

func (c *CLI) createPullPointSubscription(ctx context.Context) {
	fmt.Println("⏳ Creating pull point subscription...")

	termTimeStr := c.readInputWithDefault("Subscription duration (seconds)", "60")
	termTimeSec, err := strconv.Atoi(termTimeStr)
	if err != nil || termTimeSec <= 0 {
		termTimeSec = 60
	}

	termTime := time.Duration(termTimeSec) * time.Second

	sub, err := c.client.CreatePullPointSubscription(ctx, "", &termTime, "")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Pull Point Subscription Created:")
	fmt.Printf("   Subscription Reference: %s\n", sub.SubscriptionReference)
	fmt.Printf("   Current Time: %v\n", sub.CurrentTime)
	fmt.Printf("   Termination Time: %v\n", sub.TerminationTime)

	// Offer to pull messages
	pull := c.readInput("📨 Pull messages now? (y/n) [y]: ")
	if pull == "" || strings.EqualFold(pull, "y") {
		c.pullMessagesFromSubscription(ctx, sub.SubscriptionReference)
	}

	// Offer to unsubscribe
	unsub := c.readInput("🔌 Unsubscribe? (y/n) [y]: ")
	if unsub == "" || strings.EqualFold(unsub, "y") {
		if err := c.client.Unsubscribe(ctx, sub.SubscriptionReference); err != nil {
			fmt.Printf("❌ Unsubscribe error: %v\n", err)
		} else {
			fmt.Println("✅ Unsubscribed successfully")
		}
	}
}

func (c *CLI) pullMessagesFromSubscription(ctx context.Context, subscriptionRef string) {
	fmt.Println("⏳ Pulling messages (5 second timeout)...")

	messages, err := c.client.PullMessages(ctx, subscriptionRef, 5*time.Second, 100) //nolint:mnd // 100 max messages
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(messages) == 0 {
		fmt.Println("📭 No messages available")

		return
	}

	fmt.Printf("✅ Received %d message(s):\n", len(messages))
	for i := range messages {
		msg := &messages[i]
		if i >= 10 { //nolint:mnd // Show max 10 messages
			fmt.Printf("   ... and %d more\n", len(messages)-10) //nolint:mnd // Show remaining count

			break
		}
		fmt.Printf("   %d. Topic: %s\n", i+1, msg.Topic)
		if msg.Message.PropertyOperation != "" {
			fmt.Printf("      Operation: %s\n", msg.Message.PropertyOperation)
		}
		if !msg.Message.UtcTime.IsZero() {
			fmt.Printf("      Time: %v\n", msg.Message.UtcTime)
		}
		if len(msg.Message.Source) > 0 {
			fmt.Printf("      Source: %s=%s\n", msg.Message.Source[0].Name, msg.Message.Source[0].Value)
		}
		if len(msg.Message.Data) > 0 {
			fmt.Printf("      Data: %s=%s\n", msg.Message.Data[0].Name, msg.Message.Data[0].Value)
		}
	}
}

func (c *CLI) getEventBrokers(ctx context.Context) {
	fmt.Println("⏳ Getting event brokers...")

	brokers, err := c.client.GetEventBrokers(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(brokers) == 0 {
		fmt.Println("📭 No event brokers configured")

		return
	}

	fmt.Printf("✅ Found %d Event Broker(s):\n", len(brokers))
	for i, broker := range brokers {
		fmt.Printf("   %d. Address: %s\n", i+1, broker.Address)
		if broker.TopicPrefix != "" {
			fmt.Printf("      Topic Prefix: %s\n", broker.TopicPrefix)
		}
		if broker.Status != "" {
			fmt.Printf("      Status: %s\n", broker.Status)
		}
		fmt.Printf("      QoS: %d\n", broker.QoS)
	}
}

// ============================================
// Device IO Operations
// ============================================

func (c *CLI) deviceIOOperations() {
	if c.client == nil {
		fmt.Println("❌ Not connected to any camera")

		return
	}

	fmt.Println("🔌 Device IO Operations")
	fmt.Println("======================")
	fmt.Println("  1. Get Device IO Capabilities")
	fmt.Println("  2. Get Digital Inputs")
	fmt.Println("  3. Get Relay Outputs")
	fmt.Println("  4. Set Relay Output State")
	fmt.Println("  5. Get Relay Output Options")
	fmt.Println("  6. Get Video Outputs")
	fmt.Println("  7. Get Video Output Configuration")
	fmt.Println("  8. Get Video Output Configuration Options")
	fmt.Println("  9. Get Serial Ports")
	fmt.Println("  0. Back to Main Menu")

	choice := c.readInput("Select operation: ")
	ctx := context.Background()

	switch choice {
	case "1":
		c.getDeviceIOCapabilities(ctx)
	case "2":
		c.getDigitalInputs(ctx)
	case "3":
		c.getRelayOutputsCLI(ctx)
	case "4":
		c.setRelayOutputStateCLI(ctx)
	case "5":
		c.getRelayOutputOptionsCLI(ctx)
	case "6":
		c.getVideoOutputsCLI(ctx)
	case "7":
		c.getVideoOutputConfigurationCLI(ctx)
	case "8":
		c.getVideoOutputConfigurationOptionsCLI(ctx)
	case "9":
		c.getSerialPortsCLI(ctx)
	case "0":
		return
	default:
		fmt.Println("❌ Invalid option")
	}
}

func (c *CLI) getDeviceIOCapabilities(ctx context.Context) {
	fmt.Println("⏳ Getting Device IO capabilities...")

	caps, err := c.client.GetDeviceIOServiceCapabilities(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Device IO Capabilities:")
	fmt.Printf("   Video Sources: %d\n", caps.VideoSources)
	fmt.Printf("   Video Outputs: %d\n", caps.VideoOutputs)
	fmt.Printf("   Audio Sources: %d\n", caps.AudioSources)
	fmt.Printf("   Audio Outputs: %d\n", caps.AudioOutputs)
	fmt.Printf("   Relay Outputs: %d\n", caps.RelayOutputs)
	fmt.Printf("   Digital Inputs: %d\n", caps.DigitalInputs)
	fmt.Printf("   Serial Ports: %d\n", caps.SerialPorts)
	fmt.Printf("   Digital Input Options: %v\n", caps.DigitalInputOptions)
	fmt.Printf("   Serial Port Configuration: %v\n", caps.SerialPortConfiguration)
}

func (c *CLI) getDigitalInputs(ctx context.Context) {
	fmt.Println("⏳ Getting digital inputs...")

	inputs, err := c.client.GetDigitalInputs(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(inputs) == 0 {
		fmt.Println("📭 No digital inputs found")

		return
	}

	fmt.Printf("✅ Found %d Digital Input(s):\n", len(inputs))
	for i, input := range inputs {
		fmt.Printf("   %d. Token: %s\n", i+1, input.Token)
		fmt.Printf("      Idle State: %s\n", input.IdleState)
	}
}

func (c *CLI) getRelayOutputsCLI(ctx context.Context) {
	fmt.Println("⏳ Getting relay outputs...")

	relays, err := c.client.GetRelayOutputs(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(relays) == 0 {
		fmt.Println("📭 No relay outputs found")

		return
	}

	fmt.Printf("✅ Found %d Relay Output(s):\n", len(relays))
	for i, relay := range relays {
		fmt.Printf("   %d. Token: %s\n", i+1, relay.Token)
		fmt.Printf("      Mode: %s\n", relay.Properties.Mode)
		fmt.Printf("      Idle State: %s\n", relay.Properties.IdleState)
		if relay.Properties.DelayTime > 0 {
			fmt.Printf("      Delay Time: %v\n", relay.Properties.DelayTime)
		}
	}
}

func (c *CLI) setRelayOutputStateCLI(ctx context.Context) {
	// First get available relay outputs
	relays, err := c.client.GetRelayOutputs(ctx)
	if err != nil {
		fmt.Printf("❌ Error getting relays: %v\n", err)

		return
	}

	if len(relays) == 0 {
		fmt.Println("📭 No relay outputs available")

		return
	}

	fmt.Println("Available relay outputs:")
	for i, relay := range relays {
		fmt.Printf("  %d. %s (Mode: %s)\n", i+1, relay.Token, relay.Properties.Mode)
	}

	choice := c.readInput("Select relay (1-" + strconv.Itoa(len(relays)) + "): ")
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(relays) {
		fmt.Println("❌ Invalid selection")

		return
	}

	selectedRelay := relays[idx-1]

	fmt.Println("Select state:")
	fmt.Println("  1. Active")
	fmt.Println("  2. Inactive")
	stateChoice := c.readInput("State: ")

	var state onvif.RelayLogicalState
	switch stateChoice {
	case "1":
		state = onvif.RelayLogicalStateActive
	case "2":
		state = onvif.RelayLogicalStateInactive
	default:
		fmt.Println("❌ Invalid state")

		return
	}

	fmt.Println("⏳ Setting relay output state...")

	if err := c.client.SetRelayOutputState(ctx, selectedRelay.Token, state); err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Printf("✅ Relay %s set to %s\n", selectedRelay.Token, state)
}

func (c *CLI) getVideoOutputsCLI(ctx context.Context) {
	fmt.Println("⏳ Getting video outputs...")

	outputs, err := c.client.GetVideoOutputs(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(outputs) == 0 {
		fmt.Println("📭 No video outputs found")

		return
	}

	fmt.Printf("✅ Found %d Video Output(s):\n", len(outputs))
	for i, output := range outputs {
		fmt.Printf("   %d. Token: %s\n", i+1, output.Token)
		if output.Resolution != nil {
			fmt.Printf("      Resolution: %dx%d\n", output.Resolution.Width, output.Resolution.Height)
		}
		if output.RefreshRate > 0 {
			fmt.Printf("      Refresh Rate: %.1f Hz\n", output.RefreshRate)
		}
		if output.AspectRatio != "" {
			fmt.Printf("      Aspect Ratio: %s\n", output.AspectRatio)
		}
	}
}

func (c *CLI) getSerialPortsCLI(ctx context.Context) {
	fmt.Println("⏳ Getting serial ports...")

	ports, err := c.client.GetSerialPorts(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(ports) == 0 {
		fmt.Println("📭 No serial ports found")

		return
	}

	fmt.Printf("✅ Found %d Serial Port(s):\n", len(ports))
	for i, port := range ports {
		fmt.Printf("   %d. Token: %s\n", i+1, port.Token)
		fmt.Printf("      Type: %s\n", port.Type)

		// Get configuration if available
		config, err := c.client.GetSerialPortConfiguration(ctx, port.Token)
		if err == nil {
			fmt.Printf("      Baud Rate: %d\n", config.BaudRate)
			fmt.Printf("      Parity: %s\n", config.ParityBit)
			fmt.Printf("      Data Bits: %d\n", config.CharacterLength)
			fmt.Printf("      Stop Bits: %.1f\n", config.StopBit)
		}
	}
}

func (c *CLI) getRelayOutputOptionsCLI(ctx context.Context) {
	// First get available relay outputs
	relays, err := c.client.GetRelayOutputs(ctx)
	if err != nil {
		fmt.Printf("❌ Error getting relays: %v\n", err)

		return
	}

	if len(relays) == 0 {
		fmt.Println("📭 No relay outputs available")

		return
	}

	fmt.Println("Available relay outputs:")
	for i, relay := range relays {
		fmt.Printf("  %d. %s\n", i+1, relay.Token)
	}

	choice := c.readInput("Select relay (1-" + strconv.Itoa(len(relays)) + "): ")
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(relays) {
		fmt.Println("❌ Invalid selection")

		return
	}

	selectedRelay := relays[idx-1]
	fmt.Printf("⏳ Getting relay output options for %s...\n", selectedRelay.Token)

	options, err := c.client.GetRelayOutputOptions(ctx, selectedRelay.Token)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Relay Output Options:")
	fmt.Printf("   Token: %s\n", options.Token)
	if len(options.Mode) > 0 {
		fmt.Println("   Supported Modes:")
		for _, mode := range options.Mode {
			fmt.Printf("      - %s\n", mode)
		}
	}
	if len(options.DelayTimes) > 0 {
		fmt.Println("   Supported Delay Times:")
		for _, dt := range options.DelayTimes {
			fmt.Printf("      - %s\n", dt)
		}
	}
	fmt.Printf("   Discrete: %v\n", options.Discrete)
}

func (c *CLI) getVideoOutputConfigurationCLI(ctx context.Context) {
	// First get available video outputs
	outputs, err := c.client.GetVideoOutputs(ctx)
	if err != nil {
		fmt.Printf("❌ Error getting video outputs: %v\n", err)

		return
	}

	if len(outputs) == 0 {
		fmt.Println("📭 No video outputs available")

		return
	}

	fmt.Println("Available video outputs:")
	for i, output := range outputs {
		fmt.Printf("  %d. %s\n", i+1, output.Token)
	}

	choice := c.readInput("Select video output (1-" + strconv.Itoa(len(outputs)) + "): ")
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(outputs) {
		fmt.Println("❌ Invalid selection")

		return
	}

	selectedOutput := outputs[idx-1]
	fmt.Printf("⏳ Getting video output configuration for %s...\n", selectedOutput.Token)

	config, err := c.client.GetVideoOutputConfiguration(ctx, selectedOutput.Token)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Video Output Configuration:")
	fmt.Printf("   Token: %s\n", config.Token)
	fmt.Printf("   Name: %s\n", config.Name)
	fmt.Printf("   Use Count: %d\n", config.UseCount)
	fmt.Printf("   Output Token: %s\n", config.OutputToken)
}

func (c *CLI) getVideoOutputConfigurationOptionsCLI(ctx context.Context) {
	// First get available video outputs
	outputs, err := c.client.GetVideoOutputs(ctx)
	if err != nil {
		fmt.Printf("❌ Error getting video outputs: %v\n", err)

		return
	}

	if len(outputs) == 0 {
		fmt.Println("📭 No video outputs available")

		return
	}

	fmt.Println("Available video outputs:")
	for i, output := range outputs {
		fmt.Printf("  %d. %s\n", i+1, output.Token)
	}

	choice := c.readInput("Select video output (1-" + strconv.Itoa(len(outputs)) + "): ")
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(outputs) {
		fmt.Println("❌ Invalid selection")

		return
	}

	selectedOutput := outputs[idx-1]
	fmt.Printf("⏳ Getting video output configuration options for %s...\n", selectedOutput.Token)

	options, err := c.client.GetVideoOutputConfigurationOptions(ctx, selectedOutput.Token)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	fmt.Println("✅ Video Output Configuration Options:")
	fmt.Printf("   Name Length: Min=%d, Max=%d\n", options.Name.Min, options.Name.Max)
	if len(options.OutputTokensAvailable) > 0 {
		fmt.Println("   Available Output Tokens:")
		for _, token := range options.OutputTokensAvailable {
			fmt.Printf("      - %s\n", token)
		}
	}
}
