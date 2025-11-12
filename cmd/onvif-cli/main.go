package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/0x524a/onvif-go"
	"github.com/0x524a/onvif-go/discovery"
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
	} else {
		fmt.Println("  3-6. (Connect to camera first)")
	}
	fmt.Println("  0. Exit")
	fmt.Println()
}

func (c *CLI) readInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CLI) readInputWithDefault(prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devices, err := discovery.Discover(ctx, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ Discovery failed: %v\n", err)
		return
	}

	if len(devices) == 0 {
		fmt.Println("❌ No ONVIF cameras found on the network")
		fmt.Println("💡 Make sure:")
		fmt.Println("   - Cameras are powered on and connected")
		fmt.Println("   - ONVIF is enabled on the cameras")
		fmt.Println("   - You're on the same network segment")
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
		if strings.ToLower(connect) == "y" || strings.ToLower(connect) == "yes" {
			if len(devices) == 1 {
				c.connectToDiscoveredCamera(devices[0])
			} else {
				c.selectAndConnectCamera(devices)
			}
		}
	}
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
	username := c.readInputWithDefault("Username", "admin")
	
	fmt.Print("Password: ")
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	c.createClient(endpoint, username, password)
}

func (c *CLI) connectToCamera() {
	fmt.Println("🔗 Connect to Camera")
	fmt.Println("===================")

	endpoint := c.readInputWithDefault("Camera endpoint (http://ip:port/onvif/device_service)", "http://192.168.1.100/onvif/device_service")
	username := c.readInputWithDefault("Username", "admin")
	
	fmt.Print("Password: ")
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	c.createClient(endpoint, username, password)
}

func (c *CLI) createClient(endpoint, username, password string) {
	fmt.Println("⏳ Connecting...")

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(30*time.Second),
	)
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
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("Reboot cancelled")
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
		return "", fmt.Errorf("no profiles found")
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

	pan, _ := strconv.ParseFloat(panStr, 64)
	tilt, _ := strconv.ParseFloat(tiltStr, 64)
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

	pan, _ := strconv.ParseFloat(panStr, 64)
	tilt, _ := strconv.ParseFloat(tiltStr, 64)
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

	pan, _ := strconv.ParseFloat(panStr, 64)
	tilt, _ := strconv.ParseFloat(tiltStr, 64)
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

	panTilt := strings.ToLower(stopPanTilt) == "y" || strings.ToLower(stopPanTilt) == "yes"
	zoom := strings.ToLower(stopZoom) == "y" || strings.ToLower(stopZoom) == "yes"

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
		return "", fmt.Errorf("no profiles found")
	}

	for _, profile := range profiles {
		if profile.VideoSourceConfiguration != nil {
			return profile.VideoSourceConfiguration.SourceToken, nil
		}
	}

	return "", fmt.Errorf("no video source configuration found")
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

	currentValue := "50.0"
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

	currentValue := "50.0"
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

	currentValue := "50.0"
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

	currentValue := "50.0"
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
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
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