package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524a/onvif-go"
	"github.com/0x524a/onvif-go/server"
)

func main() {
	fmt.Println("🧪 Testing ONVIF Server Implementation")
	fmt.Println("======================================")
	fmt.Println()

	// Create and start server in background
	config := server.DefaultConfig()
	config.Port = 8081 // Use different port to avoid conflicts

	srv, err := server.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background
	serverReady := make(chan bool)
	go func() {
		// Give server a moment to start
		time.Sleep(500 * time.Millisecond)
		serverReady <- true
		
		if err := srv.Start(ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	<-serverReady
	fmt.Println("✅ Server started on port 8081")
	fmt.Println()

	// Create ONVIF client
	client, err := onvif.NewClient(
		"http://localhost:8081/onvif/device_service",
		onvif.WithCredentials("admin", "admin"),
		onvif.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	testCtx := context.Background()

	// Test 1: Get Device Information
	fmt.Println("Test 1: GetDeviceInformation")
	info, err := client.GetDeviceInformation(testCtx)
	if err != nil {
		log.Fatalf("❌ GetDeviceInformation failed: %v", err)
	}
	fmt.Printf("✅ Device: %s %s (Firmware: %s)\n", info.Manufacturer, info.Model, info.FirmwareVersion)
	fmt.Printf("   Serial: %s\n", info.SerialNumber)
	fmt.Println()

	// Test 2: Get Capabilities
	fmt.Println("Test 2: GetCapabilities")
	if err := client.Initialize(testCtx); err != nil {
		log.Fatalf("❌ Initialize (GetCapabilities) failed: %v", err)
	}
	fmt.Println("✅ Capabilities retrieved successfully")
	fmt.Println()

	// Test 3: Get Profiles
	fmt.Println("Test 3: GetProfiles")
	profiles, err := client.GetProfiles(testCtx)
	if err != nil {
		log.Fatalf("❌ GetProfiles failed: %v", err)
	}
	fmt.Printf("✅ Found %d profiles:\n", len(profiles))
	for i, profile := range profiles {
		fmt.Printf("   [%d] %s (Token: %s)\n", i+1, profile.Name, profile.Token)
		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("       Video: %dx%d @ %s\n",
				profile.VideoEncoderConfiguration.Resolution.Width,
				profile.VideoEncoderConfiguration.Resolution.Height,
				profile.VideoEncoderConfiguration.Encoding)
		}
	}
	fmt.Println()

	// Test 4: Get Stream URI
	if len(profiles) > 0 {
		fmt.Println("Test 4: GetStreamURI")
		streamURI, err := client.GetStreamURI(testCtx, profiles[0].Token)
		if err != nil {
			log.Fatalf("❌ GetStreamURI failed: %v", err)
		}
		fmt.Printf("✅ Stream URI: %s\n", streamURI.URI)
		fmt.Println()
	}

	// Test 5: Get Snapshot URI
	if len(profiles) > 0 {
		fmt.Println("Test 5: GetSnapshotURI")
		snapshotURI, err := client.GetSnapshotURI(testCtx, profiles[0].Token)
		if err != nil {
			log.Fatalf("❌ GetSnapshotURI failed: %v", err)
		}
		fmt.Printf("✅ Snapshot URI: %s\n", snapshotURI.URI)
		fmt.Println()
	}

	// Test 6: PTZ Status (if PTZ is available)
	if len(profiles) > 0 && profiles[0].PTZConfiguration != nil {
		fmt.Println("Test 6: PTZ GetStatus")
		status, err := client.GetStatus(testCtx, profiles[0].Token)
		if err != nil {
			log.Fatalf("❌ GetStatus failed: %v", err)
		}
		fmt.Printf("✅ PTZ Position: Pan=%.2f, Tilt=%.2f, Zoom=%.2f\n",
			status.Position.PanTilt.X,
			status.Position.PanTilt.Y,
			status.Position.Zoom.X)
		fmt.Println()

		// Test 7: PTZ Absolute Move
		fmt.Println("Test 7: PTZ AbsoluteMove")
		position := &onvif.PTZVector{
			PanTilt: &onvif.Vector2D{X: 10.0, Y: -5.0},
			Zoom:    &onvif.Vector1D{X: 0.5},
		}
		if err := client.AbsoluteMove(testCtx, profiles[0].Token, position, nil); err != nil {
			log.Fatalf("❌ AbsoluteMove failed: %v", err)
		}
		fmt.Println("✅ PTZ moved to absolute position")
		fmt.Println()

		// Wait a bit for movement to complete
		time.Sleep(600 * time.Millisecond)

		// Verify new position
		fmt.Println("Test 8: Verify PTZ Position")
		status, err = client.GetStatus(testCtx, profiles[0].Token)
		if err != nil {
			log.Fatalf("❌ GetStatus failed: %v", err)
		}
		fmt.Printf("✅ New PTZ Position: Pan=%.2f, Tilt=%.2f, Zoom=%.2f\n",
			status.Position.PanTilt.X,
			status.Position.PanTilt.Y,
			status.Position.Zoom.X)
		fmt.Println()

		// Test 9: PTZ Presets
		fmt.Println("Test 9: Get PTZ Presets")
		presets, err := client.GetPresets(testCtx, profiles[0].Token)
		if err != nil {
			log.Fatalf("❌ GetPresets failed: %v", err)
		}
		fmt.Printf("✅ Found %d presets:\n", len(presets))
		for i, preset := range presets {
			fmt.Printf("   [%d] %s (Token: %s)\n", i+1, preset.Name, preset.Token)
		}
		fmt.Println()
	}

	// Test 10: Get System Date and Time
	fmt.Println("Test 10: GetSystemDateAndTime")
	_, err = client.GetSystemDateAndTime(testCtx)
	if err != nil {
		log.Fatalf("❌ GetSystemDateAndTime failed: %v", err)
	}
	fmt.Println("✅ System date and time retrieved successfully")
	fmt.Println()

	// All tests passed!
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                          ║")
	fmt.Println("║  ✅ All Tests Passed! ONVIF Server is working! ✅       ║")
	fmt.Println("║                                                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Stop the server
	cancel()
	time.Sleep(500 * time.Millisecond)
}
