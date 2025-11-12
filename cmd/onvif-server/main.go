package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0x524a/onvif-go/server"
)

var (
	version = "1.0.0"
)

func main() {
	// Define command-line flags
	host := flag.String("host", "0.0.0.0", "Server host address")
	port := flag.Int("port", 8080, "Server port")
	username := flag.String("username", "admin", "Authentication username")
	password := flag.String("password", "admin", "Authentication password")
	manufacturer := flag.String("manufacturer", "go-onvif", "Device manufacturer")
	model := flag.String("model", "Virtual Multi-Lens Camera", "Device model")
	firmware := flag.String("firmware", "1.0.0", "Firmware version")
	serial := flag.String("serial", "SN-12345678", "Serial number")
	profiles := flag.Int("profiles", 3, "Number of camera profiles (1-10)")
	ptz := flag.Bool("ptz", true, "Enable PTZ support")
	imaging := flag.Bool("imaging", true, "Enable Imaging support")
	events := flag.Bool("events", false, "Enable Events support")
	info := flag.Bool("info", false, "Show server info and exit")
	showVersion := flag.Bool("version", false, "Show version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ONVIF Server - Virtual IP Camera Simulator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Start with default settings (3 profiles, PTZ enabled)\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Start with custom credentials and 5 profiles\n")
		fmt.Fprintf(os.Stderr, "  %s -username myuser -password mypass -profiles 5\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Start on specific port without PTZ\n")
		fmt.Fprintf(os.Stderr, "  %s -port 9000 -ptz=false\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Show server information\n")
		fmt.Fprintf(os.Stderr, "  %s -info\n\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("onvif-server version %s\n", version)
		os.Exit(0)
	}

	// Validate profiles count
	if *profiles < 1 || *profiles > 10 {
		log.Fatal("Number of profiles must be between 1 and 10")
	}

	// Create server configuration
	config := buildConfig(*host, *port, *username, *password, *manufacturer, *model,
		*firmware, *serial, *profiles, *ptz, *imaging, *events)

	// Create server
	srv, err := server.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle info flag
	if *info {
		fmt.Println(srv.ServerInfo())
		os.Exit(0)
	}

	// Print banner
	printBanner()

	// Create context that listens for interrupt signals
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
	fmt.Println("\n🛑 Received interrupt signal, shutting down...")
	cancel()

	// Give the server a moment to shut down gracefully
	time.Sleep(1 * time.Second)
	fmt.Println("✅ Server stopped")
}

// buildConfig creates a server configuration from command-line arguments
func buildConfig(host string, port int, username, password, manufacturer, model,
	firmware, serial string, numProfiles int, ptz, imaging, events bool) *server.Config {

	config := &server.Config{
		Host:     host,
		Port:     port,
		BasePath: "/onvif",
		Timeout:  30 * time.Second,
		DeviceInfo: server.DeviceInfo{
			Manufacturer:    manufacturer,
			Model:           model,
			FirmwareVersion: firmware,
			SerialNumber:    serial,
			HardwareID:      "HW-87654321",
		},
		Username:       username,
		Password:       password,
		SupportPTZ:     ptz,
		SupportImaging: imaging,
		SupportEvents:  events,
		Profiles:       make([]server.ProfileConfig, numProfiles),
	}

	// Define profile templates
	templates := []struct {
		name       string
		width      int
		height     int
		framerate  int
		bitrate    int
		quality    float64
		hasPTZ     bool
		ptzZoomMax float64
	}{
		{"Main Camera - High Quality", 1920, 1080, 30, 4096, 80, true, 1},
		{"Wide Angle Camera", 1280, 720, 30, 2048, 75, false, 0},
		{"Telephoto Camera", 1920, 1080, 25, 6144, 85, true, 3},
		{"Low Light Camera", 1920, 1080, 30, 4096, 80, false, 0},
		{"Ultra HD Camera", 3840, 2160, 30, 16384, 90, true, 2},
		{"Compact Camera", 640, 480, 30, 512, 70, false, 0},
		{"PTZ Dome Camera", 1920, 1080, 30, 4096, 80, true, 2},
		{"Fisheye Camera", 1920, 1080, 30, 4096, 80, false, 0},
		{"Thermal Camera", 640, 480, 30, 1024, 75, true, 1},
		{"License Plate Camera", 1920, 1080, 60, 8192, 90, true, 5},
	}

	// Generate profiles
	for i := 0; i < numProfiles; i++ {
		template := templates[i%len(templates)]
		
		profile := server.ProfileConfig{
			Token: fmt.Sprintf("profile_%d", i),
			Name:  template.name,
			VideoSource: server.VideoSourceConfig{
				Token:      fmt.Sprintf("video_source_%d", i),
				Name:       template.name,
				Resolution: server.Resolution{Width: template.width, Height: template.height},
				Framerate:  template.framerate,
				Bounds:     server.Bounds{X: 0, Y: 0, Width: template.width, Height: template.height},
			},
			VideoEncoder: server.VideoEncoderConfig{
				Encoding:   "H264",
				Resolution: server.Resolution{Width: template.width, Height: template.height},
				Quality:    template.quality,
				Framerate:  template.framerate,
				Bitrate:    template.bitrate,
				GovLength:  template.framerate,
			},
			Snapshot: server.SnapshotConfig{
				Enabled:    true,
				Resolution: server.Resolution{Width: template.width, Height: template.height},
				Quality:    template.quality + 5,
			},
		}

		// Add PTZ if enabled and template supports it
		if ptz && template.hasPTZ {
			profile.PTZ = &server.PTZConfig{
				NodeToken:          fmt.Sprintf("ptz_node_%d", i),
				PanRange:           server.Range{Min: -180, Max: 180},
				TiltRange:          server.Range{Min: -90, Max: 90},
				ZoomRange:          server.Range{Min: 0, Max: template.ptzZoomMax},
				DefaultSpeed:       server.PTZSpeed{Pan: 0.5, Tilt: 0.5, Zoom: 0.5},
				SupportsContinuous: true,
				SupportsAbsolute:   true,
				SupportsRelative:   true,
				Presets: []server.Preset{
					{
						Token:    fmt.Sprintf("preset_%d_0", i),
						Name:     "Home",
						Position: server.PTZPosition{Pan: 0, Tilt: 0, Zoom: 0},
					},
					{
						Token:    fmt.Sprintf("preset_%d_1", i),
						Name:     "Entrance",
						Position: server.PTZPosition{Pan: -45, Tilt: -10, Zoom: template.ptzZoomMax * 0.5},
					},
				},
			}
		}

		config.Profiles[i] = profile
	}

	return config
}

// printBanner prints the application banner
func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     🎥  ONVIF Virtual Camera Server  🎥                  ║
║                                                           ║
║     Simulate multi-lens IP cameras with ONVIF support    ║
║     Version: ` + version + `                                        ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
