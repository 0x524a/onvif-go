package main

// discover subcommand: ONVIF camera discovery on the local network.

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/discovery"
)

const defaultDiscoveryTimeout = 10 * time.Second

func runDiscover(args []string) {
	fs := flag.NewFlagSet("discover", flag.ExitOnError)
	iface := fs.String("interface", "", "Network interface to use (e.g., en0, en11)")
	timeout := fs.Duration("timeout", defaultDiscoveryTimeout, "Discovery timeout")
	_ = fs.Parse(args)

	opts := &discovery.DiscoverOptions{
		NetworkInterface: *iface,
	}

	fmt.Printf("Discovering ONVIF cameras on the network")
	if *iface != "" {
		fmt.Printf(" (interface: %s)", *iface)
	}
	fmt.Println("...")

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	devices, err := discovery.DiscoverWithOptions(ctx, *timeout, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Discovery error: %v\n", err)
		os.Exit(1) //nolint:gocritic // defer cancel() is still executed by runtime on exit
	}

	if len(devices) == 0 {
		fmt.Println("No cameras found.")
		os.Exit(0)
	}

	fmt.Printf("\nFound %d camera(s):\n\n", len(devices))
	for i, d := range devices {
		fmt.Printf("Camera %d:\n", i+1)
		fmt.Printf("  Endpoint: %s\n", d.EndpointRef)
		for _, addr := range d.XAddrs {
			fmt.Printf("  XAddr: %s\n", addr)
		}
		if len(d.Scopes) > 0 {
			fmt.Printf("  Scopes:\n")
			for _, s := range d.Scopes {
				fmt.Printf("    - %s\n", s)
			}
		}
		fmt.Println()
	}
}
