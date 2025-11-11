package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x524A/go-onvif"
)

func main() {
	// Camera connection details
	endpoint := "http://192.168.1.201/onvif/device_service"
	username := "service"
	password := "Service.1234"

	fmt.Println("=== Comprehensive ONVIF Camera Test ===")
	fmt.Println("Connecting to:", endpoint)
	fmt.Println()

	// Create client
	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test 1: Get Device Information
	fmt.Println("=== Test 1: GetDeviceInformation ===")
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("✓ Manufacturer: %s\n", info.Manufacturer)
		fmt.Printf("✓ Model: %s\n", info.Model)
		fmt.Printf("✓ Firmware: %s\n", info.FirmwareVersion)
		fmt.Printf("✓ Serial Number: %s\n", info.SerialNumber)
		fmt.Printf("✓ Hardware ID: %s\n", info.HardwareID)
	}
	fmt.Println()

	// Test 2: Get System Date and Time
	fmt.Println("=== Test 2: GetSystemDateAndTime ===")
	dateTime, err := client.GetSystemDateAndTime(ctx)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("✓ System Date/Time: %+v\n", dateTime)
	}
	fmt.Println()

	// Test 3: Get Capabilities
	fmt.Println("=== Test 3: GetCapabilities ===")
	capabilities, err := client.GetCapabilities(ctx)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
	} else {
		fmt.Println("✓ Capabilities retrieved successfully:")
		if capabilities.Device != nil {
			fmt.Printf("  - Device: %s\n", capabilities.Device.XAddr)
		}
		if capabilities.Media != nil {
			fmt.Printf("  - Media: %s\n", capabilities.Media.XAddr)
		}
		if capabilities.PTZ != nil {
			fmt.Printf("  - PTZ: %s\n", capabilities.PTZ.XAddr)
		}
		if capabilities.Imaging != nil {
			fmt.Printf("  - Imaging: %s\n", capabilities.Imaging.XAddr)
		}
		if capabilities.Events != nil {
			fmt.Printf("  - Events: %s\n", capabilities.Events.XAddr)
		}
		if capabilities.Analytics != nil {
			fmt.Printf("  - Analytics: %s\n", capabilities.Analytics.XAddr)
		}
	}
	fmt.Println()

	// Initialize client to discover service endpoints
	fmt.Println("=== Test 4: Initialize (Discover Services) ===")
	if err := client.Initialize(ctx); err != nil {
		log.Printf("ERROR: %v\n", err)
	} else {
		fmt.Println("✓ Services discovered successfully")
	}
	fmt.Println()

	// Test 5: Get Media Profiles
	fmt.Println("=== Test 5: GetProfiles ===")
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d profile(s)\n", len(profiles))
		for i, profile := range profiles {
			fmt.Printf("  Profile %d: %s (Token: %s)\n", i+1, profile.Name, profile.Token)
			if profile.VideoEncoderConfiguration != nil {
				fmt.Printf("    - Encoding: %s\n", profile.VideoEncoderConfiguration.Encoding)
				if profile.VideoEncoderConfiguration.Resolution != nil {
					fmt.Printf("    - Resolution: %dx%d\n",
						profile.VideoEncoderConfiguration.Resolution.Width,
						profile.VideoEncoderConfiguration.Resolution.Height)
				}
			}
		}
	}
	fmt.Println()

	// Test 6: Get Stream URIs
	fmt.Println("=== Test 6: GetStreamURI (for first profile) ===")
	if len(profiles) > 0 {
		streamURI, err := client.GetStreamURI(ctx, profiles[0].Token)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("✓ Stream URI: %s\n", streamURI.URI)
			fmt.Printf("  - Invalid After Connect: %v\n", streamURI.InvalidAfterConnect)
			fmt.Printf("  - Invalid After Reboot: %v\n", streamURI.InvalidAfterReboot)
		}
	}
	fmt.Println()

	// Test 7: Get Snapshot URI
	fmt.Println("=== Test 7: GetSnapshotURI (for first profile) ===")
	if len(profiles) > 0 {
		snapshotURI, err := client.GetSnapshotURI(ctx, profiles[0].Token)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("✓ Snapshot URI: %s\n", snapshotURI.URI)
		}
	}
	fmt.Println()

	// Test 8: Get Video Encoder Configuration
	fmt.Println("=== Test 8: GetVideoEncoderConfiguration ===")
	if len(profiles) > 0 && profiles[0].VideoEncoderConfiguration != nil {
		config, err := client.GetVideoEncoderConfiguration(ctx, profiles[0].VideoEncoderConfiguration.Token)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("✓ Video Encoder Configuration:\n")
			fmt.Printf("  - Name: %s\n", config.Name)
			fmt.Printf("  - Encoding: %s\n", config.Encoding)
			if config.Resolution != nil {
				fmt.Printf("  - Resolution: %dx%d\n", config.Resolution.Width, config.Resolution.Height)
			}
			fmt.Printf("  - Quality: %.1f\n", config.Quality)
			if config.RateControl != nil {
				fmt.Printf("  - Frame Rate Limit: %d\n", config.RateControl.FrameRateLimit)
				fmt.Printf("  - Bitrate Limit: %d\n", config.RateControl.BitrateLimit)
			}
		}
	}
	fmt.Println()

	// Test 9: PTZ Operations (if PTZ is available)
	fmt.Println("=== Test 9: PTZ Operations ===")
	if len(profiles) > 0 && profiles[0].PTZConfiguration != nil {
		fmt.Println("PTZ configuration detected, testing PTZ operations...")

		// Get PTZ Status
		ptzStatus, err := client.GetStatus(ctx, profiles[0].Token)
		if err != nil {
			log.Printf("ERROR getting PTZ status: %v\n", err)
		} else {
			fmt.Printf("✓ PTZ Status retrieved\n")
			if ptzStatus.Position != nil {
				if ptzStatus.Position.PanTilt != nil {
					fmt.Printf("  - Pan/Tilt Position: X=%.2f, Y=%.2f\n",
						ptzStatus.Position.PanTilt.X,
						ptzStatus.Position.PanTilt.Y)
				}
				if ptzStatus.Position.Zoom != nil {
					fmt.Printf("  - Zoom Position: %.2f\n", ptzStatus.Position.Zoom.X)
				}
			}
			if ptzStatus.MoveStatus != nil {
				fmt.Printf("  - Pan/Tilt Move Status: %s\n", ptzStatus.MoveStatus.PanTilt)
				fmt.Printf("  - Zoom Move Status: %s\n", ptzStatus.MoveStatus.Zoom)
			}
		}

		// Get PTZ Presets
		presets, err := client.GetPresets(ctx, profiles[0].Token)
		if err != nil {
			log.Printf("ERROR getting PTZ presets: %v\n", err)
		} else {
			fmt.Printf("✓ Found %d PTZ preset(s)\n", len(presets))
			for i, preset := range presets {
				fmt.Printf("  Preset %d: %s (Token: %s)\n", i+1, preset.Name, preset.Token)
			}
		}
	} else {
		fmt.Println("⊘ No PTZ configuration found for this profile")
	}
	fmt.Println()

	// Test 10: Imaging Settings
	fmt.Println("=== Test 10: Imaging Settings ===")
	if len(profiles) > 0 && profiles[0].VideoSourceConfiguration != nil {
		settings, err := client.GetImagingSettings(ctx, profiles[0].VideoSourceConfiguration.SourceToken)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("✓ Imaging Settings:\n")
			if settings.Brightness != nil {
				fmt.Printf("  - Brightness: %.1f\n", *settings.Brightness)
			}
			if settings.ColorSaturation != nil {
				fmt.Printf("  - Color Saturation: %.1f\n", *settings.ColorSaturation)
			}
			if settings.Contrast != nil {
				fmt.Printf("  - Contrast: %.1f\n", *settings.Contrast)
			}
			if settings.Sharpness != nil {
				fmt.Printf("  - Sharpness: %.1f\n", *settings.Sharpness)
			}
			if settings.IrCutFilter != nil {
				fmt.Printf("  - IR Cut Filter: %s\n", *settings.IrCutFilter)
			}
			if settings.BacklightCompensation != nil {
				fmt.Printf("  - Backlight Compensation: %s (Level: %.1f)\n",
					settings.BacklightCompensation.Mode,
					settings.BacklightCompensation.Level)
			}
			if settings.Exposure != nil {
				fmt.Printf("  - Exposure Mode: %s\n", settings.Exposure.Mode)
				fmt.Printf("    Priority: %s\n", settings.Exposure.Priority)
			}
			if settings.Focus != nil {
				fmt.Printf("  - Focus Mode: %s\n", settings.Focus.AutoFocusMode)
			}
			if settings.WhiteBalance != nil {
				fmt.Printf("  - White Balance Mode: %s\n", settings.WhiteBalance.Mode)
			}
			if settings.WideDynamicRange != nil {
				fmt.Printf("  - Wide Dynamic Range: %s (Level: %.1f)\n",
					settings.WideDynamicRange.Mode,
					settings.WideDynamicRange.Level)
			}
		}
	}
	fmt.Println()

	fmt.Println("=== Test Summary ===")
	fmt.Println("All tests completed!")
}
