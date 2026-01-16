package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524a/onvif-go"
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

	// Get device information
	fmt.Println("\nRetrieving device information...")
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		log.Fatalf("Failed to get device information: %v", err)
	}

	fmt.Printf("\nDevice Information:\n")
	fmt.Printf("  Manufacturer: %s\n", info.Manufacturer)
	fmt.Printf("  Model: %s\n", info.Model)
	fmt.Printf("  Firmware: %s\n", info.FirmwareVersion)
	fmt.Printf("  Serial Number: %s\n", info.SerialNumber)
	fmt.Printf("  Hardware ID: %s\n", info.HardwareID)

	// Initialize client (discover service endpoints)
	fmt.Println("\nInitializing client and discovering services...")
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	// Get media profiles
	fmt.Println("\nRetrieving media profiles...")
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Fatalf("Failed to get profiles: %v", err)
	}

	fmt.Printf("\nFound %d profile(s):\n", len(profiles))
	for i, profile := range profiles {
		fmt.Printf("\nProfile #%d:\n", i+1)
		fmt.Printf("  Token: %s\n", profile.Token)
		fmt.Printf("  Name: %s\n", profile.Name)

		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("  Video Encoding: %s\n", profile.VideoEncoderConfiguration.Encoding)
			if profile.VideoEncoderConfiguration.Resolution != nil {
				fmt.Printf("  Resolution: %dx%d\n",
					profile.VideoEncoderConfiguration.Resolution.Width,
					profile.VideoEncoderConfiguration.Resolution.Height)
			}
			fmt.Printf("  Quality: %.1f\n", profile.VideoEncoderConfiguration.Quality)
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

	fmt.Println("\nDone!")
}
