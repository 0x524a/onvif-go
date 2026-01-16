package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/0x524a/onvif-go"
	"github.com/0x524a/onvif-go/discovery"
)

const (
	defaultUsername   = "admin"
	defaultTimeout    = 10
	defaultRetryDelay = 5
	ptzTimeout        = 30
	ptzStepSize       = 2
	ptzSpeed          = 0.5
	maxBodyPreview    = 200
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🎥 Quick ONVIF Camera Tool")
	fmt.Println("==========================")
	fmt.Println()

	for {
		fmt.Println("What would you like to do?")
		fmt.Println("1. 🔍 Discover cameras")
		fmt.Println("2. 🌐 List network interfaces")
		fmt.Println("3. 📹 Connect to camera")
		fmt.Println("4. 🎮 PTZ demo")
		fmt.Println("5. 📡 Get stream URLs")
		fmt.Println("0. Exit")
		fmt.Print("\nChoice: ")

		//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			discoverCameras()
		case "2":
			listNetworkInterfaces()
		case "3":
			connectAndShowInfo()
		case "4":
			ptzDemo()
		case "5":
			getStreamURLs()
		case "0", "q", "quit":
			fmt.Println("Goodbye! 👋")

			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
		fmt.Println()
	}
}

func discoverCameras() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🔍 Discovering cameras on network...")

	// Ask if user wants to use a specific interface
	fmt.Print("Use specific network interface? (y/n) [n]: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	useInterface, _ := reader.ReadString('\n')
	useInterface = strings.ToLower(strings.TrimSpace(useInterface))

	var opts *discovery.DiscoverOptions
	if useInterface == "y" || useInterface == "yes" {
		// List interfaces
		interfaces, err := discovery.ListNetworkInterfaces()
		if err != nil {
			fmt.Printf("Error: %v\n", err)

			return
		}

		fmt.Println("\nAvailable interfaces:")
		for i, iface := range interfaces {
			fmt.Printf("  %d. %s (%v)\n", i+1, iface.Name, iface.Addresses)
		}

		fmt.Print("\nEnter interface name or IP: ")
		//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
		ifaceInput, _ := reader.ReadString('\n')
		ifaceInput = strings.TrimSpace(ifaceInput)

		if ifaceInput != "" {
			opts = &discovery.DiscoverOptions{
				NetworkInterface: ifaceInput,
			}
		}
	}

	if opts == nil {
		opts = &discovery.DiscoverOptions{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout*time.Second)
	defer cancel()

	devices, err := discovery.DiscoverWithOptions(ctx, defaultRetryDelay*time.Second, opts)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(devices) == 0 {
		fmt.Println("No cameras found")

		return
	}

	fmt.Printf("✅ Found %d camera(s):\n", len(devices))
	for i, device := range devices {
		fmt.Printf("  %d. %s (%s)\n", i+1, device.GetName(), device.GetDeviceEndpoint())
	}
}

func listNetworkInterfaces() {
	fmt.Println("🌐 Network Interfaces")
	fmt.Println("====================")

	interfaces, err := discovery.ListNetworkInterfaces()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		return
	}

	if len(interfaces) == 0 {
		fmt.Println("No network interfaces found")

		return
	}

	fmt.Printf("✅ Found %d interface(s):\n\n", len(interfaces))

	for _, iface := range interfaces {
		upStr := "Up"
		if !iface.Up {
			upStr = "Down"
		}

		multicastStr := "Yes"
		if !iface.Multicast {
			multicastStr = "No"
		}

		fmt.Printf("📡 %s (%s, Multicast: %s)\n", iface.Name, upStr, multicastStr)

		if len(iface.Addresses) > 0 {
			for _, addr := range iface.Addresses {
				fmt.Printf("   └─ %s\n", addr)
			}
		}
	}
}

func connectAndShowInfo() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Camera IP: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Username [admin]: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = defaultUsername
	}

	fmt.Print("Password: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	endpoint := fmt.Sprintf("http://%s/onvif/device_service", ip)
	fmt.Printf("Connecting to %s...\n", endpoint)

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(ptzTimeout*time.Second),
	)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	ctx := context.Background()

	// Get device info
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)

		return
	}

	fmt.Printf("✅ Connected!\n")
	fmt.Printf("📹 %s %s\n", info.Manufacturer, info.Model)
	fmt.Printf("🔧 Firmware: %s\n", info.FirmwareVersion)

	// Initialize and get profiles
	//nolint:errcheck // Ignore initialization errors, we'll catch them on GetProfiles
	_ = client.Initialize(ctx)
	profiles, err := client.GetProfiles(ctx)
	if err == nil && len(profiles) > 0 {
		fmt.Printf("📺 %d profile(s) available\n", len(profiles))

		// Show first stream URL
		streamURI, err := client.GetStreamURI(ctx, profiles[0].Token)
		if err == nil {
			fmt.Printf("📡 Stream: %s\n", streamURI.URI)
		}
	}
}

func ptzDemo() { //nolint:funlen,gocyclo // Many statements and high complexity due to user interaction
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Camera IP: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Username [admin]: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = defaultUsername
	}

	fmt.Print("Password: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	endpoint := fmt.Sprintf("http://%s/onvif/device_service", ip)

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
	)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	ctx := context.Background()
	//nolint:errcheck // Ignore initialization errors, we'll catch them on GetProfiles
	_ = client.Initialize(ctx)

	profiles, err := client.GetProfiles(ctx)
	if err != nil || len(profiles) == 0 {
		fmt.Println("❌ No profiles found")

		return
	}

	profileToken := profiles[0].Token

	// Check PTZ status
	status, err := client.GetStatus(ctx, profileToken)
	if err != nil {
		fmt.Printf("❌ PTZ not supported: %v\n", err)

		return
	}

	fmt.Println("✅ PTZ is supported!")
	if status.Position != nil && status.Position.PanTilt != nil {
		fmt.Printf("Current position: Pan=%.2f, Tilt=%.2f\n",
			status.Position.PanTilt.X, status.Position.PanTilt.Y)
	}

	fmt.Println("\n🎮 PTZ Demo - Choose movement:")
	fmt.Println("1. Move right")
	fmt.Println("2. Move left")
	fmt.Println("3. Move up")
	fmt.Println("4. Move down")
	fmt.Println("5. Go to center")
	fmt.Print("Choice: ")

	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var velocity *onvif.PTZSpeed
	var position *onvif.PTZVector

	switch choice {
	case "1":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: ptzSpeed, Y: 0.0}}
	case "2":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: -ptzSpeed, Y: 0.0}}
	case "3":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: 0.0, Y: ptzSpeed}}
	case "4":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: 0.0, Y: -ptzSpeed}}
	case "5":
		position = &onvif.PTZVector{PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.0}}
	default:
		fmt.Println("Invalid choice")

		return
	}

	if velocity != nil {
		timeout := fmt.Sprintf("PT%dS", ptzStepSize)
		err = client.ContinuousMove(ctx, profileToken, velocity, &timeout)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)

			return
		}
		fmt.Println("✅ Moving for 2 seconds...")
		time.Sleep(ptzStepSize * time.Second)
		//nolint:errcheck // Stop error is not critical for demo
		_ = client.Stop(ctx, profileToken, true, false)
	} else if position != nil {
		err = client.AbsoluteMove(ctx, profileToken, position, nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)

			return
		}
		fmt.Println("✅ Moving to center...")
	}

	fmt.Println("Demo complete!")
}

func getStreamURLs() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Camera IP: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Username [admin]: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = defaultUsername
	}

	fmt.Print("Password: ")
	//nolint:errcheck // ReadString error on stdin is rare and not critical for CLI
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	endpoint := fmt.Sprintf("http://%s/onvif/device_service", ip)

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
	)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	ctx := context.Background()
	//nolint:errcheck // Ignore initialization errors, we'll catch them on GetProfiles
	_ = client.Initialize(ctx)

	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)

		return
	}

	if len(profiles) == 0 {
		fmt.Println("❌ No profiles found")

		return
	}

	fmt.Printf("✅ Found %d profile(s):\n\n", len(profiles))

	for i, profile := range profiles {
		fmt.Printf("📹 Profile %d: %s\n", i+1, profile.Name)

		// Stream URI
		streamURI, err := client.GetStreamURI(ctx, profile.Token)
		if err != nil {
			fmt.Printf("   Stream: ❌ Error\n")
		} else {
			fmt.Printf("   📡 Stream: %s\n", streamURI.URI)
		}

		// Snapshot URI
		snapshotURI, err := client.GetSnapshotURI(ctx, profile.Token)
		if err != nil {
			fmt.Printf("   Snapshot: ❌ Error\n")
		} else {
			fmt.Printf("   📸 Snapshot: %s\n", snapshotURI.URI)
		}

		// Video info
		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("   🎬 Encoding: %s", profile.VideoEncoderConfiguration.Encoding)
			if profile.VideoEncoderConfiguration.Resolution != nil {
				fmt.Printf(" (%dx%d)",
					profile.VideoEncoderConfiguration.Resolution.Width,
					profile.VideoEncoderConfiguration.Resolution.Height)
			}
			fmt.Println()
		}

		fmt.Println()
	}

	fmt.Println("💡 Tips:")
	fmt.Println("   - Use VLC to open RTSP streams")
	fmt.Println("   - Open snapshot URLs in a web browser")
	fmt.Println("   - Some cameras may require authentication in the URL")
}
