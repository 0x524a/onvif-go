package server_test

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/0x524a/onvif-go/discovery"
	"github.com/0x524a/onvif-go/server"
)

// Start from the default three-lens camera, change what you need, and run it.
func ExampleNew() {
	cfg := server.DefaultConfig()
	cfg.Port = 8080
	cfg.Username, cfg.Password = "admin", "s3cret"

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Start blocks until the context is canceled.
	log.Fatal(srv.Start(context.Background()))
}

// An endpoint seed gives the camera the same identity every time it starts,
// which is what lets a client remember it across restarts.
func ExampleConfig_endpointSeed() {
	build := func() *server.Server {
		cfg := server.DefaultConfig()
		cfg.Output = io.Discard
		cfg.EndpointSeed = "lobby-camera"

		srv, err := server.New(cfg)
		if err != nil {
			log.Fatal(err)
		}

		return srv
	}

	first, second := build().EndpointReference(), build().EndpointReference()

	fmt.Println(first == second)
	// Output: true
}

// Make a virtual camera discoverable. Call DiscoveryDevice after Start has
// bound the listener, so the advertised address is the real one.
func ExampleServer_DiscoveryDevice() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := server.DefaultConfig()
	cfg.EndpointSeed = "lobby-camera"

	ready := make(chan struct{})
	cfg.Ready = ready

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	go func() { _ = srv.Start(ctx) }()
	<-ready

	responder := discovery.NewResponder(&discovery.ResponderConfig{})
	if err := responder.Register(srv.DiscoveryDevice()); err != nil {
		log.Fatal(err)
	}

	// Answers Probes, and sends Hello now and Bye when ctx is canceled.
	err = responder.Start(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}
