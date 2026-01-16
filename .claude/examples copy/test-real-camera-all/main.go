package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/0x524a/onvif-go"
)

const (
	cameraEndpoint = "192.168.1.201"
	username       = "service"
	password       = "Service.1234"
)

type TestResult struct {
	Operation    string      `json:"operation"`
	Success      bool        `json:"success"`
	Error        string      `json:"error,omitempty"`
	Response     interface{} `json:"response,omitempty"`
	ResponseTime string      `json:"response_time"`
}

type CameraTestReport struct {
	DeviceInfo struct {
		Manufacturer    string `json:"manufacturer"`
		Model           string `json:"model"`
		FirmwareVersion string `json:"firmware_version"`
		SerialNumber    string `json:"serial_number"`
		HardwareID      string `json:"hardware_id"`
	} `json:"device_info"`
	TestResults []TestResult `json:"test_results"`
	Timestamp   string       `json:"timestamp"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report := CameraTestReport{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Try different endpoint formats and common ONVIF ports
	endpoints := []string{
		cameraEndpoint,              // http://192.168.1.230/onvif/device_service
		"http://" + cameraEndpoint,  // http://192.168.1.230/onvif/device_service
		"https://" + cameraEndpoint, // https://192.168.1.230/onvif/device_service
		cameraEndpoint + ":80",      // http://192.168.1.230:80/onvif/device_service
		cameraEndpoint + ":443",     // http://192.168.1.230:443/onvif/device_service
		cameraEndpoint + ":8080",    // http://192.168.1.230:8080/onvif/device_service
		cameraEndpoint + ":554",     // http://192.168.1.230:554/onvif/device_service
		cameraEndpoint + ":8000",    // http://192.168.1.230:8000/onvif/device_service
		"http://" + cameraEndpoint + ":80",
		"https://" + cameraEndpoint + ":443",
		"http://" + cameraEndpoint + ":8080",
		"https://" + cameraEndpoint + ":8443",
		"http://" + cameraEndpoint + "/onvif/device_service",
		"https://" + cameraEndpoint + "/onvif/device_service",
		"http://" + cameraEndpoint + ":8080/onvif/device_service",
	}

	var client *onvif.Client
	var deviceInfo *onvif.DeviceInformation
	var err error

	fmt.Println("📡 Trying to connect to camera...")
	for i, endpoint := range endpoints {
		fmt.Printf("  Attempt %d: %s\n", i+1, endpoint)

		opts := []onvif.ClientOption{
			onvif.WithCredentials(username, password),
			onvif.WithTimeout(10 * time.Second),
		}

		// Add insecure skip verify for HTTPS endpoints
		if strings.HasPrefix(endpoint, "https://") {
			opts = append(opts, onvif.WithInsecureSkipVerify())
		}

		client, err = onvif.NewClient(endpoint, opts...)
		if err != nil {
			fmt.Printf("    ❌ Failed to create client: %v\n", err)
			continue
		}

		// Try to get device information
		deviceInfo, err = client.GetDeviceInformation(ctx)
		if err != nil {
			fmt.Printf("    ❌ Failed to connect: %v\n", err)
			continue
		}

		fmt.Printf("    ✅ Connected successfully!\n")
		break
	}

	if err != nil || deviceInfo == nil {
		log.Fatalf("Failed to connect to camera with any endpoint format. Last error: %v", err)
	}

	report.DeviceInfo.Manufacturer = deviceInfo.Manufacturer
	report.DeviceInfo.Model = deviceInfo.Model
	report.DeviceInfo.FirmwareVersion = deviceInfo.FirmwareVersion
	report.DeviceInfo.SerialNumber = deviceInfo.SerialNumber
	report.DeviceInfo.HardwareID = deviceInfo.HardwareID

	fmt.Printf("✅ Camera: %s %s (FW: %s)\n", deviceInfo.Manufacturer, deviceInfo.Model, deviceInfo.FirmwareVersion)

	// Initialize to discover service endpoints
	fmt.Println("🔍 Initializing service endpoints...")
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Test all device operations
	fmt.Println("\n🔧 Testing Device Operations...")
	testDeviceOperations(ctx, client, &report)

	// Test all media operations
	fmt.Println("\n🎬 Testing Media Operations...")
	testMediaOperations(ctx, client, &report)

	// Save report
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal report: %v", err)
	}

	// Create test-reports directory if it doesn't exist
	reportDir := "../../test-reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		log.Fatalf("Failed to create test-reports directory: %v", err)
	}

	filename := fmt.Sprintf("camera_test_report_%s_%s_%s.json",
		sanitizeFilename(deviceInfo.Manufacturer),
		sanitizeFilename(deviceInfo.Model),
		time.Now().Format("20060102_150405"))

	filepath := fmt.Sprintf("%s/%s", reportDir, filename)
	if err := os.WriteFile(filepath, reportJSON, 0644); err != nil {
		log.Fatalf("Failed to write report: %v", err)
	}

	fmt.Printf("\n✅ Test report saved to: %s\n", filepath)
}

func sanitizeFilename(s string) string {
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result += string(r)
		} else {
			result += "_"
		}
	}
	return result
}

func testDeviceOperations(ctx context.Context, client *onvif.Client, report *CameraTestReport) {
	// Test all operations
	testOperation := func(name string, testFn func() (interface{}, error)) {
		fmt.Printf("  Testing %s...", name)
		start := time.Now()
		result, err := testFn()
		duration := time.Since(start)

		testResult := TestResult{
			Operation:    name,
			ResponseTime: duration.String(),
		}

		if err != nil {
			testResult.Success = false
			testResult.Error = err.Error()
			fmt.Printf(" ❌ Error: %v\n", err)
		} else {
			testResult.Success = true
			testResult.Response = result
			fmt.Printf(" ✅\n")
		}

		report.TestResults = append(report.TestResults, testResult)
		time.Sleep(200 * time.Millisecond)
	}

	// Basic device operations
	testOperation("GetDeviceInformation", func() (interface{}, error) {
		return client.GetDeviceInformation(ctx)
	})
	testOperation("GetCapabilities", func() (interface{}, error) {
		return client.GetCapabilities(ctx)
	})
	testOperation("GetServiceCapabilities", func() (interface{}, error) {
		return client.GetServiceCapabilities(ctx)
	})
	testOperation("GetServices", func() (interface{}, error) {
		return client.GetServices(ctx, false)
	})
	testOperation("GetServicesWithCapabilities", func() (interface{}, error) {
		return client.GetServices(ctx, true)
	})

	// System operations
	testOperation("GetSystemDateAndTime", func() (interface{}, error) {
		return client.GetSystemDateAndTime(ctx)
	})
	testOperation("GetHostname", func() (interface{}, error) {
		return client.GetHostname(ctx)
	})
	testOperation("GetDNS", func() (interface{}, error) {
		return client.GetDNS(ctx)
	})
	testOperation("GetNTP", func() (interface{}, error) {
		return client.GetNTP(ctx)
	})

	// Network operations
	testOperation("GetNetworkInterfaces", func() (interface{}, error) {
		return client.GetNetworkInterfaces(ctx)
	})
	testOperation("GetNetworkProtocols", func() (interface{}, error) {
		return client.GetNetworkProtocols(ctx)
	})
	testOperation("GetNetworkDefaultGateway", func() (interface{}, error) {
		return client.GetNetworkDefaultGateway(ctx)
	})

	// Discovery operations
	testOperation("GetDiscoveryMode", func() (interface{}, error) {
		return client.GetDiscoveryMode(ctx)
	})
	testOperation("GetRemoteDiscoveryMode", func() (interface{}, error) {
		return client.GetRemoteDiscoveryMode(ctx)
	})
	testOperation("GetEndpointReference", func() (interface{}, error) {
		return client.GetEndpointReference(ctx)
	})

	// Scope operations
	testOperation("GetScopes", func() (interface{}, error) {
		return client.GetScopes(ctx)
	})

	// User operations (read-only to avoid modifying camera)
	testOperation("GetUsers", func() (interface{}, error) {
		return client.GetUsers(ctx)
	})

	// Set operations - test with caution (may modify camera state)
	// Note: These are commented out to avoid modifying camera during testing
	// Uncomment if you want to test write operations

	// testOperation("SetDiscoveryMode", func() (interface{}, error) {
	// 	currentMode, _ := client.GetDiscoveryMode(ctx)
	// 	err := client.SetDiscoveryMode(ctx, currentMode) // Set to current value
	// 	return nil, err
	// })

	// testOperation("SetRemoteDiscoveryMode", func() (interface{}, error) {
	// 	currentMode, _ := client.GetRemoteDiscoveryMode(ctx)
	// 	err := client.SetRemoteDiscoveryMode(ctx, currentMode) // Set to current value
	// 	return nil, err
	// })

	// System reboot - skip to avoid rebooting camera during testing
	// testOperation("SystemReboot", func() (interface{}, error) {
	// 	return client.SystemReboot(ctx)
	// })
}

func testMediaOperations(ctx context.Context, client *onvif.Client, report *CameraTestReport) {
	// Get profiles and other resources first
	profiles, _ := client.GetProfiles(ctx)
	videoSources, _ := client.GetVideoSources(ctx)
	audioOutputs, _ := client.GetAudioOutputs(ctx)

	var profileToken, videoEncoderToken, audioEncoderToken, videoSourceToken, audioOutputToken string
	if len(profiles) > 0 {
		profileToken = profiles[0].Token
		if profiles[0].VideoEncoderConfiguration != nil {
			videoEncoderToken = profiles[0].VideoEncoderConfiguration.Token
		}
		if profiles[0].AudioEncoderConfiguration != nil {
			audioEncoderToken = profiles[0].AudioEncoderConfiguration.Token
		}
	}
	if len(videoSources) > 0 {
		videoSourceToken = videoSources[0].Token
	}
	if len(audioOutputs) > 0 {
		audioOutputToken = audioOutputs[0].Token
	}

	// Test all operations
	testOperation := func(name string, testFn func() (interface{}, error)) {
		fmt.Printf("  Testing %s...", name)
		start := time.Now()
		result, err := testFn()
		duration := time.Since(start)

		testResult := TestResult{
			Operation:    name,
			ResponseTime: duration.String(),
		}

		if err != nil {
			testResult.Success = false
			testResult.Error = err.Error()
			fmt.Printf(" ❌ Error: %v\n", err)
		} else {
			testResult.Success = true
			testResult.Response = result
			fmt.Printf(" ✅\n")
		}

		report.TestResults = append(report.TestResults, testResult)
		time.Sleep(200 * time.Millisecond)
	}

	// Basic operations
	testOperation("GetMediaServiceCapabilities", func() (interface{}, error) {
		return client.GetMediaServiceCapabilities(ctx)
	})
	testOperation("GetProfiles", func() (interface{}, error) {
		return client.GetProfiles(ctx)
	})
	testOperation("GetVideoSources", func() (interface{}, error) {
		return client.GetVideoSources(ctx)
	})
	testOperation("GetAudioSources", func() (interface{}, error) {
		return client.GetAudioSources(ctx)
	})
	testOperation("GetAudioOutputs", func() (interface{}, error) {
		return client.GetAudioOutputs(ctx)
	})

	// Profile operations
	if profileToken != "" {
		testOperation("GetStreamURI", func() (interface{}, error) {
			return client.GetStreamURI(ctx, profileToken)
		})
		testOperation("GetSnapshotURI", func() (interface{}, error) {
			return client.GetSnapshotURI(ctx, profileToken)
		})
		testOperation("GetProfile", func() (interface{}, error) {
			return client.GetProfile(ctx, profileToken)
		})
		testOperation("SetSynchronizationPoint", func() (interface{}, error) {
			err := client.SetSynchronizationPoint(ctx, profileToken)
			return nil, err
		})
	}

	// Video encoder operations
	if videoEncoderToken != "" {
		testOperation("GetVideoEncoderConfiguration", func() (interface{}, error) {
			return client.GetVideoEncoderConfiguration(ctx, videoEncoderToken)
		})
		testOperation("GetVideoEncoderConfigurationOptions", func() (interface{}, error) {
			return client.GetVideoEncoderConfigurationOptions(ctx, videoEncoderToken)
		})
		testOperation("GetGuaranteedNumberOfVideoEncoderInstances", func() (interface{}, error) {
			return client.GetGuaranteedNumberOfVideoEncoderInstances(ctx, videoEncoderToken)
		})
	}

	// Audio encoder operations
	if audioEncoderToken != "" {
		testOperation("GetAudioEncoderConfiguration", func() (interface{}, error) {
			return client.GetAudioEncoderConfiguration(ctx, audioEncoderToken)
		})
	}
	testOperation("GetAudioEncoderConfigurationOptions", func() (interface{}, error) {
		return client.GetAudioEncoderConfigurationOptions(ctx, audioEncoderToken, profileToken)
	})

	// Video source operations
	if videoSourceToken != "" {
		testOperation("GetVideoSourceModes", func() (interface{}, error) {
			return client.GetVideoSourceModes(ctx, videoSourceToken)
		})
	}

	// Audio output operations
	testOperation("GetAudioOutputConfiguration", func() (interface{}, error) {
		// Try to get audio output config - need to find config token
		// For now, try with empty token or skip if not available
		if audioOutputToken != "" {
			// Try to get configuration - this may require a different approach
			return nil, fmt.Errorf("audio output configuration token lookup not implemented")
		}
		return nil, fmt.Errorf("no audio output available")
	})
	testOperation("GetAudioOutputConfigurationOptions", func() (interface{}, error) {
		return client.GetAudioOutputConfigurationOptions(ctx, "")
	})

	// Metadata operations
	testOperation("GetMetadataConfigurationOptions", func() (interface{}, error) {
		configToken := ""
		if len(profiles) > 0 && profiles[0].MetadataConfiguration != nil {
			configToken = profiles[0].MetadataConfiguration.Token
		}
		return client.GetMetadataConfigurationOptions(ctx, configToken, profileToken)
	})

	// Audio decoder operations
	testOperation("GetAudioDecoderConfigurationOptions", func() (interface{}, error) {
		return client.GetAudioDecoderConfigurationOptions(ctx, "")
	})

	// OSD operations
	testOperation("GetOSDs", func() (interface{}, error) {
		return client.GetOSDs(ctx, "")
	})
	testOperation("GetOSDOptions", func() (interface{}, error) {
		return client.GetOSDOptions(ctx, "")
	})

	// Additional Media operations - test all implemented operations
	if profileToken != "" {
		// Profile management operations
		testOperation("SetProfile", func() (interface{}, error) {
			profile, err := client.GetProfile(ctx, profileToken)
			if err != nil {
				return nil, err
			}
			err = client.SetProfile(ctx, profile)
			return nil, err
		})

		// Profile configuration add/remove operations
		if videoEncoderToken != "" {
			testOperation("AddVideoEncoderConfiguration", func() (interface{}, error) {
				// Try adding to a different profile if available
				if len(profiles) > 1 {
					err := client.AddVideoEncoderConfiguration(ctx, profiles[1].Token, videoEncoderToken)
					return nil, err
				}
				return nil, fmt.Errorf("only one profile available")
			})
			testOperation("RemoveVideoEncoderConfiguration", func() (interface{}, error) {
				// Only test if we have multiple profiles to avoid breaking the main profile
				if len(profiles) > 1 && profiles[1].VideoEncoderConfiguration != nil {
					err := client.RemoveVideoEncoderConfiguration(ctx, profiles[1].Token)
					return nil, err
				}
				return nil, fmt.Errorf("cannot test - would break profile")
			})
		}

		if audioEncoderToken != "" {
			testOperation("AddAudioEncoderConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.AddAudioEncoderConfiguration(ctx, profiles[1].Token, audioEncoderToken)
					return nil, err
				}
				return nil, fmt.Errorf("only one profile available")
			})
			testOperation("RemoveAudioEncoderConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 && profiles[1].AudioEncoderConfiguration != nil {
					err := client.RemoveAudioEncoderConfiguration(ctx, profiles[1].Token)
					return nil, err
				}
				return nil, fmt.Errorf("cannot test - would break profile")
			})
		}

		// Video source configuration operations
		if len(profiles) > 0 && profiles[0].VideoSourceConfiguration != nil {
			videoSourceConfigToken := profiles[0].VideoSourceConfiguration.Token
			testOperation("AddVideoSourceConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.AddVideoSourceConfiguration(ctx, profiles[1].Token, videoSourceConfigToken)
					return nil, err
				}
				return nil, fmt.Errorf("only one profile available")
			})
			testOperation("RemoveVideoSourceConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.RemoveVideoSourceConfiguration(ctx, profiles[1].Token)
					return nil, err
				}
				return nil, fmt.Errorf("cannot test - would break profile")
			})
		}

		// Audio source configuration operations
		if len(profiles) > 0 && profiles[0].AudioSourceConfiguration != nil {
			audioSourceConfigToken := profiles[0].AudioSourceConfiguration.Token
			testOperation("AddAudioSourceConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.AddAudioSourceConfiguration(ctx, profiles[1].Token, audioSourceConfigToken)
					return nil, err
				}
				return nil, fmt.Errorf("only one profile available")
			})
			testOperation("RemoveAudioSourceConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.RemoveAudioSourceConfiguration(ctx, profiles[1].Token)
					return nil, err
				}
				return nil, fmt.Errorf("cannot test - would break profile")
			})
		}

		// Metadata configuration operations
		if len(profiles) > 0 && profiles[0].MetadataConfiguration != nil {
			metadataConfigToken := profiles[0].MetadataConfiguration.Token
			testOperation("GetMetadataConfiguration", func() (interface{}, error) {
				return client.GetMetadataConfiguration(ctx, metadataConfigToken)
			})
			testOperation("AddMetadataConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.AddMetadataConfiguration(ctx, profiles[1].Token, metadataConfigToken)
					return nil, err
				}
				return nil, fmt.Errorf("only one profile available")
			})
			testOperation("RemoveMetadataConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.RemoveMetadataConfiguration(ctx, profiles[1].Token)
					return nil, err
				}
				return nil, fmt.Errorf("cannot test - would break profile")
			})
		}

		// PTZ configuration operations (if available)
		if len(profiles) > 0 && profiles[0].PTZConfiguration != nil {
			ptzConfigToken := profiles[0].PTZConfiguration.Token
			testOperation("AddPTZConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.AddPTZConfiguration(ctx, profiles[1].Token, ptzConfigToken)
					return nil, err
				}
				return nil, fmt.Errorf("only one profile available")
			})
			testOperation("RemovePTZConfiguration", func() (interface{}, error) {
				if len(profiles) > 1 {
					err := client.RemovePTZConfiguration(ctx, profiles[1].Token)
					return nil, err
				}
				return nil, fmt.Errorf("cannot test - would break profile")
			})
		}

		// Multicast streaming operations
		testOperation("StartMulticastStreaming", func() (interface{}, error) {
			err := client.StartMulticastStreaming(ctx, profileToken)
			return nil, err
		})
		testOperation("StopMulticastStreaming", func() (interface{}, error) {
			err := client.StopMulticastStreaming(ctx, profileToken)
			return nil, err
		})

		// OSD operations (if OSD token available)
		osds, _ := client.GetOSDs(ctx, "")
		if len(osds) > 0 {
			osdToken := osds[0].Token
			testOperation("GetOSD", func() (interface{}, error) {
				return client.GetOSD(ctx, osdToken)
			})
		}

		// Video source mode operations
		if videoSourceToken != "" {
			testOperation("SetVideoSourceMode", func() (interface{}, error) {
				modes, err := client.GetVideoSourceModes(ctx, videoSourceToken)
				if err != nil || len(modes) == 0 {
					return nil, fmt.Errorf("no modes available or error getting modes")
				}
				// Try to set to first available mode
				err = client.SetVideoSourceMode(ctx, videoSourceToken, modes[0].Token)
				return nil, err
			})
		}
	}

	// Create/Delete profile operations - test with caution
	// Note: These are commented out to avoid creating test profiles
	// Uncomment if you want to test profile creation/deletion

	// testOperation("CreateProfile", func() (interface{}, error) {
	// 	profile, err := client.CreateProfile(ctx, "TestProfile", "TestToken")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	// Clean up - delete the test profile
	// 	defer func() {
	// 		_ = client.DeleteProfile(ctx, profile.Token)
	// 	}()
	// 	return profile, nil
	// })
}
