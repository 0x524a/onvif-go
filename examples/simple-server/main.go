package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/0x524a/onvif-go/server"
)

func main() {
	fmt.Println("Starting ONVIF Server on port 8081...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	config := server.DefaultConfig()
	config.Port = 8081
	// The library defaults to silence so it never writes to an importing
	// program's stdout; a console demo wants the banner.
	config.Output = os.Stdout

	srv, err := server.New(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
