package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524a/onvif-go"
	"github.com/0x524a/onvif-go/discovery"
)

// This is a comprehensive demonstration of all go-onvif features
func main() {
	// Step 1: Discover cameras on the network
	fmt.Println("=== Step 1: Discovering ONVIF Cameras ===")
	discoverCameras()

	// Step 2: Connect to a specific camera
	fmt.Println("\n=== Step 2: Connecting to Camera ===")
	client := connectToCamera()

	// Step 3: Get device information
	fmt.Println("\n=== Step 3: Getting Device Information ===")
	getDeviceInfo(client)

	// Step 4: Get media profiles and streams
	fmt.Println("\n=== Step 4: Getting Media Profiles ===")
	profiles := getMediaProfiles(client)

	// Step 5: Control PTZ
	if len(profiles) > 0 {
		fmt.Println("\n=== Step 5: PTZ Control ===")
		controlPTZ(client, profiles[0].Token)
	}

	// Step 6: Adjust imaging settings
	if len(profiles) > 0 && profiles[0].VideoSourceConfiguration != nil {
		fmt.Println("\n=== Step 6: Adjusting Imaging Settings ===")
		adjustImaging(client, profiles[0].VideoSourceConfiguration.SourceToken)
	}

	fmt.Println("\n=== All operations completed successfully! ===")
}

// discoverCameras demonstrates network discovery
func discoverCameras() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devices, err := discovery.Discover(ctx, 5*time.Second)
	if err != nil {
		log.Printf("Discovery error: %v", err)
		return
	}

	fmt.Printf("Found %d device(s):\n", len(devices))
	for i, device := range devices {
		fmt.Printf("  [%d] %s at %s\n", i+1, device.GetName(), device.GetDeviceEndpoint())
	}
}

// connectToCamera creates and initializes a client
func connectToCamera() *onvif.Client {
	// Replace with your camera's details
	endpoint := "http://192.168.1.100/onvif/device_service"
	username := "admin"
	password := "password"

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Initialize to discover service endpoints
	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	fmt.Printf("Connected to: %s\n", endpoint)
	return client
}

// getDeviceInfo retrieves and displays device information
func getDeviceInfo(client *onvif.Client) {
	ctx := context.Background()

	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		log.Printf("Failed to get device info: %v", err)
		return
	}

	fmt.Printf("Manufacturer: %s\n", info.Manufacturer)
	fmt.Printf("Model: %s\n", info.Model)
	fmt.Printf("Firmware: %s\n", info.FirmwareVersion)
	fmt.Printf("Serial: %s\n", info.SerialNumber)

	// Get capabilities
	caps, err := client.GetCapabilities(ctx)
	if err != nil {
		log.Printf("Failed to get capabilities: %v", err)
		return
	}

	fmt.Println("\nSupported Services:")
	if caps.Media != nil {
		fmt.Printf("  ✓ Media (Streaming)\n")
	}
	if caps.PTZ != nil {
		fmt.Printf("  ✓ PTZ (Pan/Tilt/Zoom)\n")
	}
	if caps.Imaging != nil {
		fmt.Printf("  ✓ Imaging (Image Settings)\n")
	}
	if caps.Events != nil {
		fmt.Printf("  ✓ Events\n")
	}
}

// getMediaProfiles retrieves media profiles and stream URIs
func getMediaProfiles(client *onvif.Client) []*onvif.Profile {
	ctx := context.Background()

	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Printf("Failed to get profiles: %v", err)
		return nil
	}

	fmt.Printf("Found %d profile(s):\n", len(profiles))

	for i, profile := range profiles {
		fmt.Printf("\nProfile [%d]: %s\n", i+1, profile.Name)

		// Video configuration
		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("  Encoding: %s\n", profile.VideoEncoderConfiguration.Encoding)
			if profile.VideoEncoderConfiguration.Resolution != nil {
				fmt.Printf("  Resolution: %dx%d\n",
					profile.VideoEncoderConfiguration.Resolution.Width,
					profile.VideoEncoderConfiguration.Resolution.Height)
			}
		}

		// Get stream URI
		streamURI, err := client.GetStreamURI(ctx, profile.Token)
		if err != nil {
			fmt.Printf("  Stream URI: Error - %v\n", err)
		} else {
			fmt.Printf("  Stream URI: %s\n", streamURI.URI)
		}

		// Get snapshot URI
		snapshotURI, err := client.GetSnapshotURI(ctx, profile.Token)
		if err != nil {
			fmt.Printf("  Snapshot URI: Error - %v\n", err)
		} else {
			fmt.Printf("  Snapshot URI: %s\n", snapshotURI.URI)
		}
	}

	return profiles
}

// controlPTZ demonstrates PTZ operations
func controlPTZ(client *onvif.Client, profileToken string) {
	ctx := context.Background()

	// Get current status
	status, err := client.GetStatus(ctx, profileToken)
	if err != nil {
		log.Printf("PTZ not supported: %v", err)
		return
	}

	fmt.Println("PTZ is supported!")

	if status.Position != nil && status.Position.PanTilt != nil {
		fmt.Printf("Current Position: Pan=%.2f, Tilt=%.2f\n",
			status.Position.PanTilt.X,
			status.Position.PanTilt.Y)
	}

	// Get presets
	presets, err := client.GetPresets(ctx, profileToken)
	if err != nil {
		log.Printf("Failed to get presets: %v", err)
	} else {
		fmt.Printf("Available Presets: %d\n", len(presets))
		for _, preset := range presets {
			fmt.Printf("  - %s\n", preset.Name)
		}
	}

	// Demonstrate movement (commented out to avoid camera movement)
	/*
		// Move right
		velocity := &onvif.PTZSpeed{
			PanTilt: &onvif.Vector2D{X: 0.3, Y: 0.0},
		}
		timeout := "PT1S"
		if err := client.ContinuousMove(ctx, profileToken, velocity, &timeout); err != nil {
			log.Printf("Move failed: %v", err)
		}
		time.Sleep(1 * time.Second)
		client.Stop(ctx, profileToken, true, false)

		// Return to home
		home := &onvif.PTZVector{
			PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.0},
		}
		client.AbsoluteMove(ctx, profileToken, home, nil)
	*/

	fmt.Println("PTZ operations available (commented out in demo)")
}

// adjustImaging demonstrates imaging settings
func adjustImaging(client *onvif.Client, videoSourceToken string) {
	ctx := context.Background()

	// Get current settings
	settings, err := client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		log.Printf("Failed to get imaging settings: %v", err)
		return
	}

	fmt.Println("Current Imaging Settings:")
	if settings.Brightness != nil {
		fmt.Printf("  Brightness: %.1f\n", *settings.Brightness)
	}
	if settings.Contrast != nil {
		fmt.Printf("  Contrast: %.1f\n", *settings.Contrast)
	}
	if settings.ColorSaturation != nil {
		fmt.Printf("  Saturation: %.1f\n", *settings.ColorSaturation)
	}
	if settings.Sharpness != nil {
		fmt.Printf("  Sharpness: %.1f\n", *settings.Sharpness)
	}

	if settings.Exposure != nil {
		fmt.Printf("  Exposure Mode: %s\n", settings.Exposure.Mode)
	}

	if settings.Focus != nil {
		fmt.Printf("  Focus Mode: %s\n", settings.Focus.AutoFocusMode)
	}

	if settings.WhiteBalance != nil {
		fmt.Printf("  White Balance: %s\n", settings.WhiteBalance.Mode)
	}

	// Demonstrate setting adjustment (commented out to avoid changes)
	/*
		// Adjust brightness
		newBrightness := 55.0
		settings.Brightness = &newBrightness

		if err := client.SetImagingSettings(ctx, videoSourceToken, settings, true); err != nil {
			log.Printf("Failed to set imaging settings: %v", err)
		} else {
			fmt.Println("\nImaging settings updated!")
		}
	*/

	fmt.Println("Imaging adjustment available (commented out in demo)")
}
