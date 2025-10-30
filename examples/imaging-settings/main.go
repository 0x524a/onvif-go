package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524A/go-onvif"
)

func main() {
	// Camera connection details
	endpoint := "http://192.168.1.100/onvif/device_service"
	username := "admin"
	password := "password"

	fmt.Println("Connecting to ONVIF camera...")

	// Create a new ONVIF client
	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Initialize client
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	// Get profiles
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Fatalf("Failed to get profiles: %v", err)
	}

	if len(profiles) == 0 {
		log.Fatal("No profiles found")
	}

	// Get video source token from profile
	profile := profiles[0]
	if profile.VideoSourceConfiguration == nil {
		log.Fatal("No video source configuration found")
	}

	videoSourceToken := profile.VideoSourceConfiguration.SourceToken
	fmt.Printf("Using video source: %s\n\n", videoSourceToken)

	// Get current imaging settings
	fmt.Println("Getting current imaging settings...")
	settings, err := client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		log.Fatalf("Failed to get imaging settings: %v", err)
	}

	fmt.Println("\nCurrent Imaging Settings:")
	if settings.Brightness != nil {
		fmt.Printf("  Brightness: %.2f\n", *settings.Brightness)
	}
	if settings.Contrast != nil {
		fmt.Printf("  Contrast: %.2f\n", *settings.Contrast)
	}
	if settings.ColorSaturation != nil {
		fmt.Printf("  Saturation: %.2f\n", *settings.ColorSaturation)
	}
	if settings.Sharpness != nil {
		fmt.Printf("  Sharpness: %.2f\n", *settings.Sharpness)
	}
	if settings.IrCutFilter != nil {
		fmt.Printf("  IR Cut Filter: %s\n", *settings.IrCutFilter)
	}

	if settings.Exposure != nil {
		fmt.Printf("  Exposure Mode: %s\n", settings.Exposure.Mode)
		if settings.Exposure.Mode == "MANUAL" {
			fmt.Printf("    Exposure Time: %.2f\n", settings.Exposure.ExposureTime)
			fmt.Printf("    Gain: %.2f\n", settings.Exposure.Gain)
		}
	}

	if settings.Focus != nil {
		fmt.Printf("  Focus Mode: %s\n", settings.Focus.AutoFocusMode)
	}

	if settings.WhiteBalance != nil {
		fmt.Printf("  White Balance Mode: %s\n", settings.WhiteBalance.Mode)
	}

	if settings.WideDynamicRange != nil {
		fmt.Printf("  WDR Mode: %s\n", settings.WideDynamicRange.Mode)
		fmt.Printf("  WDR Level: %.2f\n", settings.WideDynamicRange.Level)
	}

	// Modify some settings
	fmt.Println("\n\nModifying imaging settings...")
	
	// Increase brightness
	newBrightness := 60.0
	settings.Brightness = &newBrightness
	
	// Increase contrast
	newContrast := 55.0
	settings.Contrast = &newContrast
	
	// Set to auto exposure
	if settings.Exposure != nil {
		settings.Exposure.Mode = "AUTO"
	}

	// Apply new settings
	if err := client.SetImagingSettings(ctx, videoSourceToken, settings, true); err != nil {
		log.Fatalf("Failed to set imaging settings: %v", err)
	}

	fmt.Println("Imaging settings updated successfully!")

	// Verify changes
	fmt.Println("\nVerifying new settings...")
	updatedSettings, err := client.GetImagingSettings(ctx, videoSourceToken)
	if err != nil {
		log.Fatalf("Failed to get updated imaging settings: %v", err)
	}

	fmt.Println("\nUpdated Imaging Settings:")
	if updatedSettings.Brightness != nil {
		fmt.Printf("  Brightness: %.2f\n", *updatedSettings.Brightness)
	}
	if updatedSettings.Contrast != nil {
		fmt.Printf("  Contrast: %.2f\n", *updatedSettings.Contrast)
	}
	if updatedSettings.Exposure != nil {
		fmt.Printf("  Exposure Mode: %s\n", updatedSettings.Exposure.Mode)
	}

	fmt.Println("\nImaging settings demonstration complete!")
}
