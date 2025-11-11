package onvif
package captures

import (
	"context"
	"testing"
	"time"

	"github.com/0x524A/go-onvif"
)

// TestEnhancedDeviceFeatures tests new Device service methods with real camera data
// Based on test results from Bosch FLEXIDOME indoor 5100i IR (8.71.0066)
func TestEnhancedDeviceFeatures(t *testing.T) {
	// Create client with test credentials
	client, err := onvif.NewClient(
		"http://192.168.1.201/onvif/device_service",
		onvif.WithCredentials("service", "Service.1234"),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("GetHostname", func(t *testing.T) {
		hostname, err := client.GetHostname(ctx)
		if err != nil {
			t.Fatalf("GetHostname failed: %v", err)
		}

		// Bosch camera has hostname configuration
		if hostname == nil {
			t.Fatal("Expected hostname information, got nil")
		}

		t.Logf("Hostname: FromDHCP=%v, Name=%q", hostname.FromDHCP, hostname.Name)
	})

	t.Run("GetDNS", func(t *testing.T) {
		dns, err := client.GetDNS(ctx)
		if err != nil {
			t.Fatalf("GetDNS failed: %v", err)
		}

		if dns == nil {
			t.Fatal("Expected DNS information, got nil")
		}

		// Bosch camera uses DHCP for DNS
		if !dns.FromDHCP {
			t.Logf("Note: Camera not using DHCP for DNS")
		}

		// Should have at least one DNS server
		if len(dns.DNSFromDHCP) == 0 && len(dns.DNSManual) == 0 {
			t.Error("Expected at least one DNS server")
		}

		t.Logf("DNS: FromDHCP=%v, Servers=%d (DHCP) + %d (Manual)",
			dns.FromDHCP, len(dns.DNSFromDHCP), len(dns.DNSManual))
	})

	t.Run("GetNTP", func(t *testing.T) {
		ntp, err := client.GetNTP(ctx)
		if err != nil {
			t.Fatalf("GetNTP failed: %v", err)
		}

		if ntp == nil {
			t.Fatal("Expected NTP information, got nil")
		}

		// Bosch camera uses DHCP for NTP
		if !ntp.FromDHCP {
			t.Logf("Note: Camera not using DHCP for NTP")
		}

		t.Logf("NTP: FromDHCP=%v, Servers=%d (DHCP) + %d (Manual)",
			ntp.FromDHCP, len(ntp.NTPFromDHCP), len(ntp.NTPManual))
	})

	t.Run("GetNetworkInterfaces", func(t *testing.T) {
		interfaces, err := client.GetNetworkInterfaces(ctx)
		if err != nil {
			t.Fatalf("GetNetworkInterfaces failed: %v", err)
		}

		// Bosch camera has 1 network interface
		if len(interfaces) == 0 {
			t.Fatal("Expected at least one network interface")
		}

		iface := interfaces[0]
		if iface.Token == "" {
			t.Error("Expected interface to have token")
		}

		if iface.Info.Name == "" {
			t.Error("Expected interface to have name")
		}

		if iface.Info.HwAddress == "" {
			t.Error("Expected interface to have hardware address")
		}

		// Bosch camera has MTU of 1514
		if iface.Info.MTU == 0 {
			t.Error("Expected interface to have MTU")
		}

		t.Logf("Interface: Token=%s, Name=%s, HwAddr=%s, MTU=%d",
			iface.Token, iface.Info.Name, iface.Info.HwAddress, iface.Info.MTU)

		if iface.IPv4 != nil {
			t.Logf("  IPv4: Enabled=%v, DHCP=%v",
				iface.IPv4.Enabled, iface.IPv4.Config.DHCP)
		}
	})

	t.Run("GetScopes", func(t *testing.T) {
		scopes, err := client.GetScopes(ctx)
		if err != nil {
			t.Fatalf("GetScopes failed: %v", err)
		}

		// Bosch camera has 8 scopes
		if len(scopes) == 0 {
			t.Fatal("Expected at least one scope")
		}

		// Check for expected scopes
		foundManufacturer := false
		foundType := false
		foundProfiles := 0

		for _, scope := range scopes {
			if scope.ScopeItem == "onvif://www.onvif.org/name/Bosch" {
				foundManufacturer = true
			}
			if scope.ScopeItem == "onvif://www.onvif.org/type/Network_Video_Transmitter" {
				foundType = true
			}
			// Count ONVIF profiles
			if len(scope.ScopeItem) > 30 && scope.ScopeItem[:30] == "onvif://www.onvif.org/Profile/" {
				foundProfiles++
			}
		}

		if !foundManufacturer {
			t.Error("Expected to find manufacturer scope")
		}
		if !foundType {
			t.Error("Expected to find device type scope")
		}

		t.Logf("Scopes: Total=%d, Manufacturer=%v, Type=%v, Profiles=%d",
			len(scopes), foundManufacturer, foundType, foundProfiles)
	})

	t.Run("GetUsers", func(t *testing.T) {
		users, err := client.GetUsers(ctx)
		if err != nil {
			t.Fatalf("GetUsers failed: %v", err)
		}

		// Bosch camera has 3 users
		if len(users) == 0 {
			t.Fatal("Expected at least one user")
		}

		// Verify user levels
		userLevels := make(map[string]int)
		for _, user := range users {
			if user.Username == "" {
				t.Error("Expected user to have username")
			}
			if user.UserLevel == "" {
				t.Error("Expected user to have level")
			}
			userLevels[user.UserLevel]++
		}

		t.Logf("Users: Total=%d, Administrator=%d, Operator=%d, User=%d",
			len(users),
			userLevels["Administrator"],
			userLevels["Operator"],
			userLevels["User"])
	})
}

// TestEnhancedMediaFeatures tests new Media service methods
func TestEnhancedMediaFeatures(t *testing.T) {
	client, err := onvif.NewClient(
		"http://192.168.1.201/onvif/device_service",
		onvif.WithCredentials("service", "Service.1234"),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Initialize to get media endpoint
	if err := client.Initialize(ctx); err != nil {
		t.Logf("Warning: Initialize failed: %v", err)
	}

	t.Run("GetVideoSources", func(t *testing.T) {
		sources, err := client.GetVideoSources(ctx)
		if err != nil {
			t.Fatalf("GetVideoSources failed: %v", err)
		}

		// Bosch camera has 1 video source
		if len(sources) == 0 {
			t.Fatal("Expected at least one video source")
		}

		source := sources[0]
		if source.Token == "" {
			t.Error("Expected source to have token")
		}

		// Bosch camera supports 30fps
		if source.Framerate == 0 {
			t.Error("Expected source to have framerate")
		}

		// Bosch camera has 1920x1080 resolution
		if source.Resolution == nil {
			t.Error("Expected source to have resolution")
		} else {
			if source.Resolution.Width == 0 || source.Resolution.Height == 0 {
				t.Error("Expected valid resolution dimensions")
			}
			t.Logf("Video Source: Token=%s, Framerate=%.1ffps, Resolution=%dx%d",
				source.Token, source.Framerate,
				source.Resolution.Width, source.Resolution.Height)
		}
	})

	t.Run("GetAudioSources", func(t *testing.T) {
		sources, err := client.GetAudioSources(ctx)
		if err != nil {
			t.Fatalf("GetAudioSources failed: %v", err)
		}

		// Bosch camera has 1 audio source with 2 channels
		if len(sources) == 0 {
			t.Fatal("Expected at least one audio source")
		}

		source := sources[0]
		if source.Token == "" {
			t.Error("Expected source to have token")
		}

		t.Logf("Audio Source: Token=%s, Channels=%d",
			source.Token, source.Channels)
	})

	t.Run("GetAudioOutputs", func(t *testing.T) {
		outputs, err := client.GetAudioOutputs(ctx)
		if err != nil {
			t.Fatalf("GetAudioOutputs failed: %v", err)
		}

		// Bosch camera has 1 audio output
		if len(outputs) == 0 {
			t.Fatal("Expected at least one audio output")
		}

		output := outputs[0]
		if output.Token == "" {
			t.Error("Expected output to have token")
		}

		t.Logf("Audio Output: Token=%s", output.Token)
	})
}

// TestEnhancedImagingFeatures tests new Imaging service methods
func TestEnhancedImagingFeatures(t *testing.T) {
	client, err := onvif.NewClient(
		"http://192.168.1.201/onvif/device_service",
		onvif.WithCredentials("service", "Service.1234"),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Initialize to get imaging endpoint
	if err := client.Initialize(ctx); err != nil {
		t.Logf("Warning: Initialize failed: %v", err)
	}

	// Get video source token
	sources, err := client.GetVideoSources(ctx)
	if err != nil || len(sources) == 0 {
		t.Skip("No video sources available for imaging tests")
	}

	videoSourceToken := sources[0].Token

	t.Run("GetOptions", func(t *testing.T) {
		options, err := client.GetOptions(ctx, videoSourceToken)
		if err != nil {
			t.Fatalf("GetOptions failed: %v", err)
		}

		if options == nil {
			t.Fatal("Expected imaging options, got nil")
		}

		// Bosch camera supports brightness (0-255)
		if options.Brightness != nil {
			if options.Brightness.Min > options.Brightness.Max {
				t.Error("Expected Min <= Max for brightness")
			}
			t.Logf("Brightness range: %.0f - %.0f",
				options.Brightness.Min, options.Brightness.Max)
		}

		// Bosch camera supports color saturation (0-255)
		if options.ColorSaturation != nil {
			if options.ColorSaturation.Min > options.ColorSaturation.Max {
				t.Error("Expected Min <= Max for color saturation")
			}
			t.Logf("ColorSaturation range: %.0f - %.0f",
				options.ColorSaturation.Min, options.ColorSaturation.Max)
		}

		// Bosch camera supports contrast (0-255)
		if options.Contrast != nil {
			if options.Contrast.Min > options.Contrast.Max {
				t.Error("Expected Min <= Max for contrast")
			}
			t.Logf("Contrast range: %.0f - %.0f",
				options.Contrast.Min, options.Contrast.Max)
		}
	})

	t.Run("GetMoveOptions", func(t *testing.T) {
		moveOptions, err := client.GetMoveOptions(ctx, videoSourceToken)
		if err != nil {
			t.Fatalf("GetMoveOptions failed: %v", err)
		}

		if moveOptions == nil {
			t.Fatal("Expected move options, got nil")
		}

		// Log available move options
		hasAbsolute := moveOptions.Absolute != nil
		hasRelative := moveOptions.Relative != nil
		hasContinuous := moveOptions.Continuous != nil

		t.Logf("Move Options: Absolute=%v, Relative=%v, Continuous=%v",
			hasAbsolute, hasRelative, hasContinuous)
	})
}
