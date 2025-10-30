package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524A/go-onvif/discovery"
)

func main() {
	fmt.Println("Discovering ONVIF devices on the network...")
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Discover devices
	devices, err := discovery.Discover(ctx, 5*time.Second)
	if err != nil {
		log.Fatalf("Discovery failed: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No ONVIF devices found on the network")
		return
	}

	fmt.Printf("\nFound %d device(s):\n\n", len(devices))

	for i, device := range devices {
		fmt.Printf("Device #%d:\n", i+1)
		fmt.Printf("  Endpoint: %s\n", device.GetDeviceEndpoint())
		fmt.Printf("  Name: %s\n", device.GetName())
		fmt.Printf("  Location: %s\n", device.GetLocation())
		fmt.Printf("  Types: %v\n", device.Types)
		fmt.Printf("  Scopes: %v\n", device.Scopes)
		fmt.Printf("  XAddrs: %v\n", device.XAddrs)
		fmt.Println()
	}
}
