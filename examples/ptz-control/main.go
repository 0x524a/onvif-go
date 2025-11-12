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

	profileToken := profiles[0].Token
	fmt.Printf("Using profile: %s\n\n", profiles[0].Name)

	// Demonstrate PTZ controls
	demonstratePTZ(ctx, client, profileToken)
}

func demonstratePTZ(ctx context.Context, client *onvif.Client, profileToken string) {
	// Get current PTZ status
	fmt.Println("Getting current PTZ status...")
	status, err := client.GetStatus(ctx, profileToken)
	if err != nil {
		log.Printf("Warning: Failed to get PTZ status: %v\n", err)
	} else {
		fmt.Printf("Current Position:\n")
		if status.Position != nil {
			if status.Position.PanTilt != nil {
				fmt.Printf("  Pan/Tilt: X=%.2f, Y=%.2f\n",
					status.Position.PanTilt.X,
					status.Position.PanTilt.Y)
			}
			if status.Position.Zoom != nil {
				fmt.Printf("  Zoom: %.2f\n", status.Position.Zoom.X)
			}
		}
		fmt.Println()
	}

	// Get presets
	fmt.Println("Getting PTZ presets...")
	presets, err := client.GetPresets(ctx, profileToken)
	if err != nil {
		log.Printf("Warning: Failed to get presets: %v\n", err)
	} else {
		fmt.Printf("Found %d preset(s):\n", len(presets))
		for _, preset := range presets {
			fmt.Printf("  - %s (Token: %s)\n", preset.Name, preset.Token)
		}
		fmt.Println()
	}

	// Continuous move right for 2 seconds
	fmt.Println("Moving camera right...")
	velocity := &onvif.PTZSpeed{
		PanTilt: &onvif.Vector2D{
			X: 0.5,  // Move right
			Y: 0.0,
		},
	}
	timeout := "PT2S" // 2 seconds
	if err := client.ContinuousMove(ctx, profileToken, velocity, &timeout); err != nil {
		log.Printf("Failed to move: %v\n", err)
	} else {
		time.Sleep(2 * time.Second)
	}

	// Stop movement
	fmt.Println("Stopping camera movement...")
	if err := client.Stop(ctx, profileToken, true, false); err != nil {
		log.Printf("Failed to stop: %v\n", err)
	}

	// Relative move
	fmt.Println("\nPerforming relative move (up and zoom in)...")
	translation := &onvif.PTZVector{
		PanTilt: &onvif.Vector2D{
			X: 0.0,
			Y: 0.1, // Move up
		},
		Zoom: &onvif.Vector1D{
			X: 0.1, // Zoom in
		},
	}
	if err := client.RelativeMove(ctx, profileToken, translation, nil); err != nil {
		log.Printf("Failed to relative move: %v\n", err)
	} else {
		time.Sleep(2 * time.Second)
	}

	// Absolute move to home position
	fmt.Println("\nMoving to home position...")
	homePosition := &onvif.PTZVector{
		PanTilt: &onvif.Vector2D{
			X: 0.0,
			Y: 0.0,
		},
		Zoom: &onvif.Vector1D{
			X: 0.0,
		},
	}
	if err := client.AbsoluteMove(ctx, profileToken, homePosition, nil); err != nil {
		log.Printf("Failed to absolute move: %v\n", err)
	} else {
		time.Sleep(2 * time.Second)
	}

	// Go to preset if available
	if len(presets) > 0 {
		fmt.Printf("\nGoing to preset: %s\n", presets[0].Name)
		if err := client.GotoPreset(ctx, profileToken, presets[0].Token, nil); err != nil {
			log.Printf("Failed to go to preset: %v\n", err)
		} else {
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println("\nPTZ demonstration complete!")
}
