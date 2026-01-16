package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524a/onvif-go"
)

func main() {
	// Demonstrates the three different endpoint formats supported by NewClient

	examples := []struct {
		name     string
		endpoint string
		desc     string
	}{
		{
			name:     "Simple IP",
			endpoint: "192.168.1.100",
			desc:     "Just the IP address - automatically adds http:// and /onvif/device_service",
		},
		{
			name:     "IP with Port",
			endpoint: "192.168.1.100:8080",
			desc:     "IP and port - automatically adds http:// and /onvif/device_service",
		},
		{
			name:     "Full URL",
			endpoint: "http://192.168.1.100/onvif/device_service",
			desc:     "Complete URL - used as-is",
		},
	}

	fmt.Println("ONVIF Client - Simplified Endpoint Formats Demo")
	fmt.Println("================================================")
	fmt.Println()

	for _, ex := range examples {
		fmt.Printf("%s:\n", ex.name)
		fmt.Printf("  Input: %s\n", ex.endpoint)
		fmt.Printf("  Description: %s\n", ex.desc)

		// Create client with simplified endpoint
		client, err := onvif.NewClient(
			ex.endpoint,
			onvif.WithCredentials("admin", "password"),
			onvif.WithTimeout(5*time.Second),
		)

		if err != nil {
			log.Printf("  Error: %v\n\n", err)
			continue
		}

		fmt.Printf("  Client created successfully!\n")
		fmt.Printf("  Endpoint will be: %s\n\n", client.Endpoint())

		// Try to get device information (will fail if camera doesn't exist)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		info, err := client.GetDeviceInformation(ctx)
		cancel()

		if err != nil {
			fmt.Printf("  Note: Could not connect to camera (this is expected in demo)\n")
			fmt.Printf("  Error: %v\n\n", err)
		} else {
			fmt.Printf("  Connected to: %s %s\n", info.Manufacturer, info.Model)
			fmt.Printf("  Firmware: %s\n\n", info.FirmwareVersion)
		}
	}

	fmt.Println("Key Benefits:")
	fmt.Println("- Simpler API: Just provide '192.168.1.100' instead of full URL")
	fmt.Println("- Flexible: Works with IP, IP:port, or full URL")
	fmt.Println("- Backward Compatible: Existing code continues to work")
}
