package discovery_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/0x524a/onvif-go/discovery"
)

// Probe the local network for five seconds and list what answers.
func ExampleDiscover() {
	devices, err := discovery.Discover(context.Background(), 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range devices {
		fmt.Printf("%s at %s\n", d.GetName(), d.GetDeviceEndpoint())
	}
}

// On a host with several network interfaces, choose which one to probe from,
// by name or by IP address.
func ExampleDiscoverWithOptions() {
	opts := &discovery.DiscoverOptions{NetworkInterface: "eth0"}

	devices, err := discovery.DiscoverWithOptions(context.Background(), 5*time.Second, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(devices), "devices found")
}

// Announce a device and answer probes for it. ListenAddr switches the responder
// to unicast, which needs no multicast routing; leave it empty to join the
// WS-Discovery multicast group on a real network.
func ExampleResponder() {
	ctx, cancel := context.WithCancel(context.Background())

	ready := make(chan struct{})
	responder := discovery.NewResponder(&discovery.ResponderConfig{
		ListenAddr: "127.0.0.1:0",
		Ready:      ready,
	})

	err := responder.Register(&discovery.Device{
		EndpointRef: discovery.StableEndpointRef("lobby-camera"),
		XAddrs:      []string{"http://192.168.1.21:8080/onvif/device_service"},
		Types:       []string{discovery.DefaultType},
		Scopes:      []string{"onvif://www.onvif.org/name/lobby-camera"},
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() { _ = responder.Start(ctx) }()
	<-ready

	// A client probes responder.Addr() with DiscoverOptions.ProbeAddress.
	opts := &discovery.DiscoverOptions{ProbeAddress: responder.Addr().String()}

	devices, err := discovery.DiscoverWithOptions(ctx, 300*time.Millisecond, opts)
	if err != nil {
		log.Fatal(err)
	}

	name := devices[0].GetName()
	cancel() // stops the responder

	fmt.Println(name)
	// Output: lobby-camera
}

// The same name always gives the same identity, so a service that restarts
// keeps the IDs its clients already know.
func ExampleStableEndpointRef() {
	first := discovery.StableEndpointRef("lobby-camera")
	again := discovery.StableEndpointRef("lobby-camera")
	other := discovery.StableEndpointRef("dock-ptz")

	fmt.Println(strings.HasPrefix(first, "urn:uuid:"))
	fmt.Println(first == again)
	fmt.Println(first == other)
	// Output:
	// true
	// true
	// false
}
