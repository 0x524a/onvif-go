package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0x524a/onvif-go/server"
)

func main() {
	// Create a custom multi-lens camera configuration
	config := &server.Config{
		Host:     "0.0.0.0",
		Port:     8080,
		BasePath: "/onvif",
		Timeout:  30 * time.Second,
		DeviceInfo: server.DeviceInfo{
			Manufacturer:    "MultiCam Systems",
			Model:           "MC-3000 Pro",
			FirmwareVersion: "2.5.1",
			SerialNumber:    "MC3000-001234",
			HardwareID:      "HW-MC3000",
		},
		Username:       "admin",
		Password:       "SecurePass123",
		SupportPTZ:     true,
		SupportImaging: true,
		SupportEvents:  false,
		Profiles: []server.ProfileConfig{
			// Profile 1: Main camera with 4K resolution
			{
				Token: "profile_main_4k",
				Name:  "Main Camera 4K",
				VideoSource: server.VideoSourceConfig{
					Token:      "video_source_main",
					Name:       "Main Camera",
					Resolution: server.Resolution{Width: 3840, Height: 2160},
					Framerate:  30,
					Bounds:     server.Bounds{X: 0, Y: 0, Width: 3840, Height: 2160},
				},
				VideoEncoder: server.VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: server.Resolution{Width: 3840, Height: 2160},
					Quality:    90,
					Framerate:  30,
					Bitrate:    20480, // 20 Mbps
					GovLength:  30,
				},
				PTZ: &server.PTZConfig{
					NodeToken:          "ptz_main",
					PanRange:           server.Range{Min: -180, Max: 180},
					TiltRange:          server.Range{Min: -90, Max: 90},
					ZoomRange:          server.Range{Min: 0, Max: 10}, // 10x optical zoom
					DefaultSpeed:       server.PTZSpeed{Pan: 0.5, Tilt: 0.5, Zoom: 0.5},
					SupportsContinuous: true,
					SupportsAbsolute:   true,
					SupportsRelative:   true,
					Presets: []server.Preset{
						{Token: "preset_home", Name: "Home Position", Position: server.PTZPosition{Pan: 0, Tilt: 0, Zoom: 0}},
						{Token: "preset_entrance", Name: "Main Entrance", Position: server.PTZPosition{Pan: -45, Tilt: -20, Zoom: 3}},
						{Token: "preset_parking", Name: "Parking Lot", Position: server.PTZPosition{Pan: 90, Tilt: -30, Zoom: 5}},
						{Token: "preset_perimeter", Name: "Perimeter View", Position: server.PTZPosition{Pan: 180, Tilt: 0, Zoom: 2}},
					},
				},
				Snapshot: server.SnapshotConfig{
					Enabled:    true,
					Resolution: server.Resolution{Width: 3840, Height: 2160},
					Quality:    95,
				},
			},
			// Profile 2: Wide-angle camera for overview
			{
				Token: "profile_wide",
				Name:  "Wide Angle Overview",
				VideoSource: server.VideoSourceConfig{
					Token:      "video_source_wide",
					Name:       "Wide Angle Camera",
					Resolution: server.Resolution{Width: 2560, Height: 1440},
					Framerate:  30,
					Bounds:     server.Bounds{X: 0, Y: 0, Width: 2560, Height: 1440},
				},
				VideoEncoder: server.VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: server.Resolution{Width: 2560, Height: 1440},
					Quality:    85,
					Framerate:  30,
					Bitrate:    8192, // 8 Mbps
					GovLength:  30,
				},
				Snapshot: server.SnapshotConfig{
					Enabled:    true,
					Resolution: server.Resolution{Width: 2560, Height: 1440},
					Quality:    90,
				},
			},
			// Profile 3: Telephoto camera for distant subjects
			{
				Token: "profile_telephoto",
				Name:  "Telephoto Camera",
				VideoSource: server.VideoSourceConfig{
					Token:      "video_source_telephoto",
					Name:       "Telephoto Camera",
					Resolution: server.Resolution{Width: 1920, Height: 1080},
					Framerate:  60, // High framerate for smooth tracking
					Bounds:     server.Bounds{X: 0, Y: 0, Width: 1920, Height: 1080},
				},
				VideoEncoder: server.VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: server.Resolution{Width: 1920, Height: 1080},
					Quality:    88,
					Framerate:  60,
					Bitrate:    10240, // 10 Mbps
					GovLength:  60,
				},
				PTZ: &server.PTZConfig{
					NodeToken:          "ptz_telephoto",
					PanRange:           server.Range{Min: -180, Max: 180},
					TiltRange:          server.Range{Min: -45, Max: 45},
					ZoomRange:          server.Range{Min: 0, Max: 30}, // 30x optical zoom
					DefaultSpeed:       server.PTZSpeed{Pan: 0.3, Tilt: 0.3, Zoom: 0.3},
					SupportsContinuous: true,
					SupportsAbsolute:   true,
					SupportsRelative:   true,
					Presets: []server.Preset{
						{Token: "preset_tel_home", Name: "Home", Position: server.PTZPosition{Pan: 0, Tilt: 0, Zoom: 0}},
						{Token: "preset_tel_far", Name: "Far View", Position: server.PTZPosition{Pan: 0, Tilt: 0, Zoom: 20}},
						{Token: "preset_tel_left", Name: "Left Side", Position: server.PTZPosition{Pan: -90, Tilt: 0, Zoom: 10}},
						{Token: "preset_tel_right", Name: "Right Side", Position: server.PTZPosition{Pan: 90, Tilt: 0, Zoom: 10}},
					},
				},
				Snapshot: server.SnapshotConfig{
					Enabled:    true,
					Resolution: server.Resolution{Width: 1920, Height: 1080},
					Quality:    92,
				},
			},
			// Profile 4: Low-light camera for night vision
			{
				Token: "profile_lowlight",
				Name:  "Low Light Night Camera",
				VideoSource: server.VideoSourceConfig{
					Token:      "video_source_lowlight",
					Name:       "Low Light Camera",
					Resolution: server.Resolution{Width: 1920, Height: 1080},
					Framerate:  30,
					Bounds:     server.Bounds{X: 0, Y: 0, Width: 1920, Height: 1080},
				},
				VideoEncoder: server.VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: server.Resolution{Width: 1920, Height: 1080},
					Quality:    85,
					Framerate:  30,
					Bitrate:    6144, // 6 Mbps
					GovLength:  30,
				},
				Snapshot: server.SnapshotConfig{
					Enabled:    true,
					Resolution: server.Resolution{Width: 1920, Height: 1080},
					Quality:    88,
				},
			},
		},
	}

	// Create and start server
	srv, err := server.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Print configuration
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                                ║")
	fmt.Println("║         🎥 ONVIF Multi-Lens Camera Server Example 🎥          ║")
	fmt.Println("║                                                                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println(srv.ServerInfo())
	fmt.Println()
	fmt.Println("📝 Configuration Details:")
	fmt.Println("   • 4 camera lenses with different capabilities")
	fmt.Println("   • Main camera: 4K resolution with 10x zoom PTZ")
	fmt.Println("   • Wide angle: 1440p for area overview")
	fmt.Println("   • Telephoto: 1080p@60fps with 30x zoom for distant subjects")
	fmt.Println("   • Low light: 1080p optimized for night vision")
	fmt.Println()
	fmt.Println("🔐 Credentials:")
	fmt.Println("   Username: admin")
	fmt.Println("   Password: SecurePass123")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server...")
	fmt.Println()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := srv.Start(ctx); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\n🛑 Shutting down server...")
	cancel()

	time.Sleep(1 * time.Second)
	fmt.Println("✅ Server stopped successfully")
}
