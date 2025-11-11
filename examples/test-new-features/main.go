package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/0x524A/go-onvif"
)

var (
	endpoint = flag.String("endpoint", "http://192.168.1.201/onvif/device_service", "ONVIF device endpoint")
	username = flag.String("username", "admin", "Username for authentication")
	password = flag.String("password", "", "Password for authentication")
	verbose  = flag.Bool("verbose", true, "Enable verbose output")
	output   = flag.String("output", "test-results.json", "Output file for results")
)

type TestResults struct {
	Timestamp    time.Time              `json:"timestamp"`
	CameraInfo   *CameraInfo            `json:"camera_info"`
	DeviceTests  map[string]interface{} `json:"device_tests"`
	MediaTests   map[string]interface{} `json:"media_tests"`
	PTZTests     map[string]interface{} `json:"ptz_tests"`
	ImagingTests map[string]interface{} `json:"imaging_tests"`
	Errors       []string               `json:"errors"`
}

type CameraInfo struct {
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmware_version"`
	SerialNumber    string `json:"serial_number"`
	HardwareID      string `json:"hardware_id"`
}

func main() {
	flag.Parse()

	if *password == "" {
		log.Fatal("Password is required. Use -password flag")
	}

	log.Printf("Testing ONVIF camera at: %s", *endpoint)
	log.Printf("Username: %s", *username)

	// Create client
	client, err := onvif.NewClient(
		*endpoint,
		onvif.WithCredentials(*username, *password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	results := &TestResults{
		Timestamp:    time.Now(),
		DeviceTests:  make(map[string]interface{}),
		MediaTests:   make(map[string]interface{}),
		PTZTests:     make(map[string]interface{}),
		ImagingTests: make(map[string]interface{}),
		Errors:       []string{},
	}

	// Initialize client
	log.Println("\n=== Initializing Client ===")
	if err := client.Initialize(ctx); err != nil {
		log.Printf("Warning: Initialize failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("Initialize: %v", err))
	}

	// Get basic device information
	log.Println("\n=== Getting Device Information ===")
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		log.Fatalf("Failed to get device information: %v", err)
	}
	log.Printf("Camera: %s %s", info.Manufacturer, info.Model)
	log.Printf("Firmware: %s", info.FirmwareVersion)
	log.Printf("Serial: %s", info.SerialNumber)

	results.CameraInfo = &CameraInfo{
		Manufacturer:    info.Manufacturer,
		Model:           info.Model,
		FirmwareVersion: info.FirmwareVersion,
		SerialNumber:    info.SerialNumber,
		HardwareID:      info.HardwareID,
	}

	// Test NEW Device Service Methods
	testDeviceService(ctx, client, results)

	// Test NEW Media Service Methods
	testMediaService(ctx, client, results)

	// Test NEW PTZ Service Methods
	testPTZService(ctx, client, results)

	// Test NEW Imaging Service Methods
	testImagingService(ctx, client, results)

	// Save results
	saveResults(results)

	log.Printf("\n=== Test Complete ===")
	log.Printf("Results saved to: %s", *output)
	log.Printf("Total errors: %d", len(results.Errors))
}

func testDeviceService(ctx context.Context, client *onvif.Client, results *TestResults) {
	log.Println("\n=== Testing Device Service (NEW Methods) ===")

	// Test GetHostname
	log.Println("\n--- GetHostname ---")
	if hostname, err := client.GetHostname(ctx); err != nil {
		log.Printf("❌ GetHostname failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetHostname: %v", err))
	} else {
		log.Printf("✅ Hostname: %+v", hostname)
		results.DeviceTests["hostname"] = hostname
	}

	// Test GetDNS
	log.Println("\n--- GetDNS ---")
	if dns, err := client.GetDNS(ctx); err != nil {
		log.Printf("❌ GetDNS failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetDNS: %v", err))
	} else {
		log.Printf("✅ DNS: FromDHCP=%v, SearchDomain=%v", dns.FromDHCP, dns.SearchDomain)
		log.Printf("   DNSFromDHCP: %+v", dns.DNSFromDHCP)
		log.Printf("   DNSManual: %+v", dns.DNSManual)
		results.DeviceTests["dns"] = dns
	}

	// Test GetNTP
	log.Println("\n--- GetNTP ---")
	if ntp, err := client.GetNTP(ctx); err != nil {
		log.Printf("❌ GetNTP failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetNTP: %v", err))
	} else {
		log.Printf("✅ NTP: FromDHCP=%v", ntp.FromDHCP)
		log.Printf("   NTPFromDHCP: %+v", ntp.NTPFromDHCP)
		log.Printf("   NTPManual: %+v", ntp.NTPManual)
		results.DeviceTests["ntp"] = ntp
	}

	// Test GetNetworkInterfaces
	log.Println("\n--- GetNetworkInterfaces ---")
	if interfaces, err := client.GetNetworkInterfaces(ctx); err != nil {
		log.Printf("❌ GetNetworkInterfaces failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetNetworkInterfaces: %v", err))
	} else {
		log.Printf("✅ Found %d network interface(s)", len(interfaces))
		for i, iface := range interfaces {
			log.Printf("   Interface %d: Token=%s, Name=%s, Enabled=%v",
				i+1, iface.Token, iface.Info.Name, iface.Enabled)
			log.Printf("      HwAddress=%s, MTU=%d", iface.Info.HwAddress, iface.Info.MTU)
			if iface.IPv4 != nil {
				log.Printf("      IPv4: Enabled=%v, DHCP=%v", iface.IPv4.Enabled, iface.IPv4.Config.DHCP)
				for _, addr := range iface.IPv4.Config.Manual {
					log.Printf("         Manual: %s/%d", addr.Address, addr.PrefixLength)
				}
			}
		}
		results.DeviceTests["network_interfaces"] = interfaces
	}

	// Test GetScopes
	log.Println("\n--- GetScopes ---")
	if scopes, err := client.GetScopes(ctx); err != nil {
		log.Printf("❌ GetScopes failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetScopes: %v", err))
	} else {
		log.Printf("✅ Found %d scope(s)", len(scopes))
		for i, scope := range scopes {
			log.Printf("   Scope %d: Def=%s, Item=%s", i+1, scope.ScopeDef, scope.ScopeItem)
		}
		results.DeviceTests["scopes"] = scopes
	}

	// Test GetUsers
	log.Println("\n--- GetUsers ---")
	if users, err := client.GetUsers(ctx); err != nil {
		log.Printf("❌ GetUsers failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetUsers: %v", err))
	} else {
		log.Printf("✅ Found %d user(s)", len(users))
		for i, user := range users {
			log.Printf("   User %d: Username=%s, Level=%s", i+1, user.Username, user.UserLevel)
		}
		results.DeviceTests["users"] = users
	}
}

func testMediaService(ctx context.Context, client *onvif.Client, results *TestResults) {
	log.Println("\n=== Testing Media Service (NEW Methods) ===")

	// Test GetVideoSources
	log.Println("\n--- GetVideoSources ---")
	if sources, err := client.GetVideoSources(ctx); err != nil {
		log.Printf("❌ GetVideoSources failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetVideoSources: %v", err))
	} else {
		log.Printf("✅ Found %d video source(s)", len(sources))
		for i, source := range sources {
			log.Printf("   Source %d: Token=%s, Framerate=%.1f",
				i+1, source.Token, source.Framerate)
			if source.Resolution != nil {
				log.Printf("      Resolution: %dx%d", source.Resolution.Width, source.Resolution.Height)
			}
		}
		results.MediaTests["video_sources"] = sources
	}

	// Test GetAudioSources
	log.Println("\n--- GetAudioSources ---")
	if sources, err := client.GetAudioSources(ctx); err != nil {
		log.Printf("❌ GetAudioSources failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetAudioSources: %v", err))
	} else {
		log.Printf("✅ Found %d audio source(s)", len(sources))
		for i, source := range sources {
			log.Printf("   Source %d: Token=%s, Channels=%d",
				i+1, source.Token, source.Channels)
		}
		results.MediaTests["audio_sources"] = sources
	}

	// Test GetAudioOutputs
	log.Println("\n--- GetAudioOutputs ---")
	if outputs, err := client.GetAudioOutputs(ctx); err != nil {
		log.Printf("❌ GetAudioOutputs failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetAudioOutputs: %v", err))
	} else {
		log.Printf("✅ Found %d audio output(s)", len(outputs))
		for i, output := range outputs {
			log.Printf("   Output %d: Token=%s", i+1, output.Token)
		}
		results.MediaTests["audio_outputs"] = outputs
	}

	// Get profiles for further testing
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Printf("⚠️  Could not get profiles: %v", err)
		return
	}

	if len(profiles) > 0 {
		log.Printf("\nUsing profile: %s (%s)", profiles[0].Name, profiles[0].Token)
		results.MediaTests["test_profile_token"] = profiles[0].Token
	}
}

func testPTZService(ctx context.Context, client *onvif.Client, results *TestResults) {
	log.Println("\n=== Testing PTZ Service (NEW Methods) ===")

	// Get profiles to find one with PTZ
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Printf("⚠️  Could not get profiles for PTZ tests: %v", err)
		return
	}

	var ptzProfile *onvif.Profile
	for _, p := range profiles {
		if p.PTZConfiguration != nil {
			ptzProfile = p
			break
		}
	}

	if ptzProfile == nil {
		log.Println("⚠️  No PTZ-enabled profile found, skipping PTZ tests")
		results.PTZTests["skipped"] = "No PTZ profile found"
		return
	}

	log.Printf("Using PTZ profile: %s (%s)", ptzProfile.Name, ptzProfile.Token)
	results.PTZTests["test_profile_token"] = ptzProfile.Token

	// Test GetConfigurations
	log.Println("\n--- GetConfigurations ---")
	if configs, err := client.GetConfigurations(ctx); err != nil {
		log.Printf("❌ GetConfigurations failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetConfigurations: %v", err))
	} else {
		log.Printf("✅ Found %d PTZ configuration(s)", len(configs))
		for i, cfg := range configs {
			log.Printf("   Config %d: Token=%s, Name=%s, NodeToken=%s",
				i+1, cfg.Token, cfg.Name, cfg.NodeToken)
		}
		results.PTZTests["configurations"] = configs
	}

	// Test GetConfiguration
	if ptzProfile.PTZConfiguration != nil {
		log.Println("\n--- GetConfiguration ---")
		if cfg, err := client.GetConfiguration(ctx, ptzProfile.PTZConfiguration.Token); err != nil {
			log.Printf("❌ GetConfiguration failed: %v", err)
			results.Errors = append(results.Errors, fmt.Sprintf("GetConfiguration: %v", err))
		} else {
			log.Printf("✅ Configuration: Token=%s, Name=%s", cfg.Token, cfg.Name)
			results.PTZTests["configuration"] = cfg
		}
	}

	// Test GetPresets
	log.Println("\n--- GetPresets ---")
	if presets, err := client.GetPresets(ctx, ptzProfile.Token); err != nil {
		log.Printf("❌ GetPresets failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetPresets: %v", err))
	} else {
		log.Printf("✅ Found %d preset(s)", len(presets))
		for i, preset := range presets {
			log.Printf("   Preset %d: Token=%s, Name=%s", i+1, preset.Token, preset.Name)
			if preset.PTZPosition != nil {
				if preset.PTZPosition.PanTilt != nil {
					log.Printf("      PanTilt: X=%.2f, Y=%.2f",
						preset.PTZPosition.PanTilt.X, preset.PTZPosition.PanTilt.Y)
				}
				if preset.PTZPosition.Zoom != nil {
					log.Printf("      Zoom: X=%.2f", preset.PTZPosition.Zoom.X)
				}
			}
		}
		results.PTZTests["presets"] = presets
	}

	// Test GetStatus
	log.Println("\n--- GetStatus ---")
	if status, err := client.GetStatus(ctx, ptzProfile.Token); err != nil {
		log.Printf("❌ GetStatus failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("PTZ GetStatus: %v", err))
	} else {
		log.Printf("✅ PTZ Status:")
		if status.Position != nil {
			if status.Position.PanTilt != nil {
				log.Printf("   Position PanTilt: X=%.2f, Y=%.2f",
					status.Position.PanTilt.X, status.Position.PanTilt.Y)
			}
			if status.Position.Zoom != nil {
				log.Printf("   Position Zoom: X=%.2f", status.Position.Zoom.X)
			}
		}
		if status.MoveStatus != nil {
			log.Printf("   MoveStatus: PanTilt=%s, Zoom=%s",
				status.MoveStatus.PanTilt, status.MoveStatus.Zoom)
		}
		results.PTZTests["status"] = status
	}
}

func testImagingService(ctx context.Context, client *onvif.Client, results *TestResults) {
	log.Println("\n=== Testing Imaging Service (NEW Methods) ===")

	// Get video sources first
	sources, err := client.GetVideoSources(ctx)
	if err != nil || len(sources) == 0 {
		log.Printf("⚠️  Could not get video sources for imaging tests: %v", err)
		return
	}

	videoSourceToken := sources[0].Token
	log.Printf("Using video source: %s", videoSourceToken)
	results.ImagingTests["test_video_source_token"] = videoSourceToken

	// Test GetOptions
	log.Println("\n--- GetOptions ---")
	if options, err := client.GetOptions(ctx, videoSourceToken); err != nil {
		log.Printf("❌ GetOptions failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetOptions: %v", err))
	} else {
		log.Printf("✅ Imaging Options:")
		if options.Brightness != nil {
			log.Printf("   Brightness: Min=%.1f, Max=%.1f", options.Brightness.Min, options.Brightness.Max)
		}
		if options.ColorSaturation != nil {
			log.Printf("   ColorSaturation: Min=%.1f, Max=%.1f", options.ColorSaturation.Min, options.ColorSaturation.Max)
		}
		if options.Contrast != nil {
			log.Printf("   Contrast: Min=%.1f, Max=%.1f", options.Contrast.Min, options.Contrast.Max)
		}
		results.ImagingTests["options"] = options
	}

	// Test GetMoveOptions
	log.Println("\n--- GetMoveOptions ---")
	if moveOptions, err := client.GetMoveOptions(ctx, videoSourceToken); err != nil {
		log.Printf("❌ GetMoveOptions failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("GetMoveOptions: %v", err))
	} else {
		log.Printf("✅ Move Options:")
		if moveOptions.Absolute != nil {
			log.Printf("   Absolute Position: Min=%.1f, Max=%.1f",
				moveOptions.Absolute.Position.Min, moveOptions.Absolute.Position.Max)
			log.Printf("   Absolute Speed: Min=%.1f, Max=%.1f",
				moveOptions.Absolute.Speed.Min, moveOptions.Absolute.Speed.Max)
		}
		if moveOptions.Relative != nil {
			log.Printf("   Relative Distance: Min=%.1f, Max=%.1f",
				moveOptions.Relative.Distance.Min, moveOptions.Relative.Distance.Max)
		}
		if moveOptions.Continuous != nil {
			log.Printf("   Continuous Speed: Min=%.1f, Max=%.1f",
				moveOptions.Continuous.Speed.Min, moveOptions.Continuous.Speed.Max)
		}
		results.ImagingTests["move_options"] = moveOptions
	}

	// Test GetImagingStatus
	log.Println("\n--- GetImagingStatus ---")
	if status, err := client.GetImagingStatus(ctx, videoSourceToken); err != nil {
		log.Printf("❌ GetImagingStatus failed: %v", err)
		results.Errors = append(results.Errors, fmt.Sprintf("Imaging GetImagingStatus: %v", err))
	} else {
		log.Printf("✅ Imaging Status:")
		if status.FocusStatus != nil {
			log.Printf("   Focus Position: %.2f", status.FocusStatus.Position)
			log.Printf("   Focus MoveStatus: %s", status.FocusStatus.MoveStatus)
			if status.FocusStatus.Error != "" {
				log.Printf("   Focus Error: %s", status.FocusStatus.Error)
			}
		}
		results.ImagingTests["status"] = status
	}
}

func saveResults(results *TestResults) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}

	if err := os.WriteFile(*output, data, 0644); err != nil {
		log.Fatalf("Failed to write results: %v", err)
	}
}
