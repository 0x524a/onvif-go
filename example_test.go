package onvif_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/0x524a/onvif-go"
	"github.com/0x524a/onvif-go/server"
)

// Connect to a camera, find its media profiles and print the RTSP address of
// the first one.
func Example() {
	ctx := context.Background()

	client, err := onvif.NewClient(
		"192.168.1.100",
		onvif.WithCredentials("admin", "password"),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize asks the camera where its media, PTZ and imaging services are.
	if err := client.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Fatal(err)
	}

	uri, err := client.GetStreamURI(ctx, profiles[0].Token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(uri.URI)
}

// The endpoint can be a full URL, host:port, or just an IP address. A bare IP
// gets http:// and the standard /onvif/device_service path added.
func ExampleNewClient() {
	client, err := onvif.NewClient(
		"http://192.168.1.100/onvif/device_service",
		onvif.WithCredentials("admin", "password"),
		onvif.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	info, err := client.GetDeviceInformation(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s %s, firmware %s\n", info.Manufacturer, info.Model, info.FirmwareVersion)
}

func ExampleClient_GetStreamURI() {
	ctx := context.Background()

	client, err := onvif.NewClient("192.168.1.100", onvif.WithCredentials("admin", "password"))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range profiles {
		uri, err := client.GetStreamURI(ctx, p.Token)
		if err != nil {
			log.Printf("profile %s: %v", p.Name, err)

			continue
		}

		fmt.Printf("%s: %s\n", p.Name, uri.URI)
	}
}

// Pan to the right for two seconds, stop, and return to a saved preset.
func ExampleClient_ContinuousMove() {
	ctx := context.Background()

	client, err := onvif.NewClient("192.168.1.100", onvif.WithCredentials("admin", "password"))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	const token = "profile_1" // from GetProfiles

	velocity := &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: 0.5, Y: 0}}
	timeout := "PT2S" // ISO 8601 duration

	if err := client.ContinuousMove(ctx, token, velocity, &timeout); err != nil {
		log.Fatal(err)
	}

	// Stop pan/tilt and zoom now rather than waiting for the timeout.
	if err := client.Stop(ctx, token, true, true); err != nil {
		log.Fatal(err)
	}

	presets, err := client.GetPresets(ctx, token)
	if err != nil {
		log.Fatal(err)
	}
	if len(presets) > 0 {
		if err := client.GotoPreset(ctx, token, presets[0].Token, nil); err != nil {
			log.Fatal(err)
		}
	}
}

// Read the current settings, change two fields and write them back. Fields are
// pointers because ONVIF settings are optional; nil means "leave unchanged".
func ExampleClient_SetImagingSettings() {
	ctx := context.Background()

	client, err := onvif.NewClient("192.168.1.100", onvif.WithCredentials("admin", "password"))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	const videoSource = "VideoSource_1" // from GetVideoSources

	settings, err := client.GetImagingSettings(ctx, videoSource)
	if err != nil {
		log.Fatal(err)
	}

	brightness, contrast := 60.0, 55.0
	settings.Brightness = &brightness
	settings.Contrast = &contrast

	if err := client.SetImagingSettings(ctx, videoSource, settings, true); err != nil {
		log.Fatal(err)
	}
}

// Run the client against the server package's virtual camera. Nothing leaves
// the machine, so this is the shape of a test for code that depends on a
// camera.
func Example_virtualCamera() {
	cfg := server.DefaultConfig()
	cfg.Host = "127.0.0.1"
	cfg.Port = 0 // let the OS pick a free port
	cfg.Username, cfg.Password = "", ""
	cfg.Output = io.Discard

	ready := make(chan struct{})
	cfg.Ready = ready

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() { _ = srv.Start(ctx) }()
	<-ready

	client, err := onvif.NewClient("http://" + srv.Addr() + "/onvif/device_service")
	if err != nil {
		log.Fatal(err)
	}

	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cancel() // stops the server

	fmt.Println(info.Manufacturer, info.Model)
	// Output: onvif-go Virtual Multi-Lens Camera
}
