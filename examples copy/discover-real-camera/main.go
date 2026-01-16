package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524a/onvif-go/discovery"
)

func main() {
	fmt.Println("Discovering ONVIF cameras on the network...")

	ctx := context.Background()

	devices, err := discovery.Discover(ctx, 10*time.Second)
	if err != nil {
		log.Fatalf("Discovery failed: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No ONVIF devices found")
		return
	}

	fmt.Printf("\nFound %d device(s):\n\n", len(devices))
	for i, device := range devices {
		fmt.Printf("Device #%d:\n", i+1)
		fmt.Printf("  Endpoint Ref: %s\n", device.EndpointRef)
		fmt.Printf("  XAddrs: %v\n", device.XAddrs)
		fmt.Printf("  Device Endpoint: %s\n", device.GetDeviceEndpoint())
		fmt.Printf("  Name: %s\n", device.GetName())
		fmt.Printf("  Location: %s\n", device.GetLocation())
		fmt.Printf("  Types: %v\n", device.Types)
		fmt.Printf("  Scopes: %v\n", device.Scopes)
		fmt.Println()
	}
}
