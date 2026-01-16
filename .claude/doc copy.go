// Package onvif provides a modern, performant Go library for communicating with ONVIF-compliant IP cameras.
//
// This package implements the ONVIF (Open Network Video Interface Forum) specification,
// providing a simple and type-safe API for controlling IP cameras and video devices.
//
// # Features
//
//   - Device Management: Get device information, capabilities, system settings
//   - Media Services: Access video streams, snapshots, and encoder configurations
//   - PTZ Control: Pan, tilt, and zoom control with presets
//   - Imaging: Adjust brightness, contrast, exposure, focus, and other image settings
//   - Discovery: Automatic device discovery via WS-Discovery
//   - Security: WS-Security authentication with password digest
//
// # Basic Usage
//
// Create a client and connect to a camera:
//
//	client, err := onvif.NewClient(
//	    "http://192.168.1.100/onvif/device_service",
//	    onvif.WithCredentials("admin", "password"),
//	    onvif.WithTimeout(30*time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx := context.Background()
//
//	// Get device information
//	info, err := client.GetDeviceInformation(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Camera: %s %s\n", info.Manufacturer, info.Model)
//
// # Discovery
//
// Discover ONVIF devices on the network:
//
//	devices, err := discovery.Discover(ctx, 5*time.Second)
//	for _, device := range devices {
//	    fmt.Printf("Found: %s at %s\n",
//	        device.GetName(),
//	        device.GetDeviceEndpoint())
//	}
//
// # Media Streaming
//
// Get stream URIs for video playback:
//
//	profiles, err := client.GetProfiles(ctx)
//	if len(profiles) > 0 {
//	    streamURI, err := client.GetStreamURI(ctx, profiles[0].Token)
//	    fmt.Printf("RTSP Stream: %s\n", streamURI.URI)
//	}
//
// # PTZ Control
//
// Control camera movement:
//
//	// Continuous movement
//	velocity := &onvif.PTZSpeed{
//	    PanTilt: &onvif.Vector2D{X: 0.5, Y: 0.0},
//	}
//	timeout := "PT2S"
//	client.ContinuousMove(ctx, profileToken, velocity, &timeout)
//
//	// Go to preset
//	presets, _ := client.GetPresets(ctx, profileToken)
//	client.GotoPreset(ctx, profileToken, presets[0].Token, nil)
//
// # Imaging Settings
//
// Adjust camera image settings:
//
//	settings, err := client.GetImagingSettings(ctx, videoSourceToken)
//	brightness := 60.0
//	settings.Brightness = &brightness
//	client.SetImagingSettings(ctx, videoSourceToken, settings, true)
//
// For more examples, see the examples directory in the repository.
package onvif
