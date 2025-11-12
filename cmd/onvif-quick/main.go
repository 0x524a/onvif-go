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

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🎥 Quick ONVIF Camera Tool")
	fmt.Println("==========================")
	fmt.Println()

	for {
		fmt.Println("What would you like to do?")
		fmt.Println("1. 🔍 Discover cameras")
		fmt.Println("2. 📹 Connect to camera")
		fmt.Println("3. 🎮 PTZ demo")
		fmt.Println("4. 📡 Get stream URLs")
		fmt.Println("0. Exit")
		fmt.Print("\nChoice: ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			discoverCameras()
		case "2":
			connectAndShowInfo()
		case "3":
			ptzDemo()
		case "4":
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
	fmt.Println("🔍 Discovering cameras on network...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devices, err := discovery.Discover(ctx, 5*time.Second)
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

func connectAndShowInfo() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Camera IP: ")
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Username [admin]: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	endpoint := fmt.Sprintf("http://%s/onvif/device_service", ip)
	fmt.Printf("Connecting to %s...\n", endpoint)

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(30*time.Second),
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
	_ = client.Initialize(ctx) // Ignore initialization errors, we'll catch them on GetProfiles
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

func ptzDemo() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Camera IP: ")
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Username [admin]: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	fmt.Print("Password: ")
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
	_ = client.Initialize(ctx) // Ignore initialization errors, we'll catch them on GetProfiles

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

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var velocity *onvif.PTZSpeed
	var position *onvif.PTZVector

	switch choice {
	case "1":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: 0.5, Y: 0.0}}
	case "2":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: -0.5, Y: 0.0}}
	case "3":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.5}}
	case "4":
		velocity = &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: 0.0, Y: -0.5}}
	case "5":
		position = &onvif.PTZVector{PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.0}}
	default:
		fmt.Println("Invalid choice")
		return
	}

	if velocity != nil {
		timeout := "PT2S"
		err = client.ContinuousMove(ctx, profileToken, velocity, &timeout)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		fmt.Println("✅ Moving for 2 seconds...")
		time.Sleep(2 * time.Second)
		_ = client.Stop(ctx, profileToken, true, false) // Stop PTZ movement
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
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Username [admin]: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	fmt.Print("Password: ")
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
	_ = client.Initialize(ctx) // Ignore initialization errors, we'll catch them on GetProfiles

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