package onvif_test

import (
	"context"
	"testing"
	"time"

	"github.com/0x524a/onvif-go"
	onviftesting "github.com/0x524a/onvif-go/testing"
)

// TestBosch_FLEXIDOME_indoor_5100i_IR_8710066 tests ONVIF client against Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066 captured responses
func TestBosch_FLEXIDOME_indoor_5100i_IR_8710066(t *testing.T) {
	// Load capture archive (in same directory as test)
	captureArchive := "Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_xmlcapture_20251110-123259.tar.gz"

	mockServer, err := onviftesting.NewMockSOAPServer(captureArchive)
	if err != nil {
		t.Fatalf("Failed to create mock server: %v", err)
	}
	defer mockServer.Close()

	// Create ONVIF client pointing to mock server
	client, err := onvif.NewClient(
		mockServer.URL()+"/onvif/device_service",
		onvif.WithCredentials("testuser", "testpass"),
	)
	if err != nil {
		t.Fatalf("Failed to create ONVIF client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("GetDeviceInformation", func(t *testing.T) {
		info, err := client.GetDeviceInformation(ctx)
		if err != nil {
			t.Errorf("GetDeviceInformation failed: %v", err)
			return
		}

		// Validate expected values
		if info.Manufacturer == "" {
			t.Error("Manufacturer is empty")
		}
		if info.Model == "" {
			t.Error("Model is empty")
		}
		if info.FirmwareVersion == "" {
			t.Error("FirmwareVersion is empty")
		}

		t.Logf("Device: %s %s (Firmware: %s)", info.Manufacturer, info.Model, info.FirmwareVersion)
	})

	t.Run("GetSystemDateAndTime", func(t *testing.T) {
		_, err := client.GetSystemDateAndTime(ctx)
		if err != nil {
			t.Errorf("GetSystemDateAndTime failed: %v", err)
		}
	})

	t.Run("GetCapabilities", func(t *testing.T) {
		caps, err := client.GetCapabilities(ctx)
		if err != nil {
			t.Errorf("GetCapabilities failed: %v", err)
			return
		}

		if caps.Device == nil {
			t.Error("Device capabilities is nil")
		}
		if caps.Media == nil {
			t.Error("Media capabilities is nil")
		}

		t.Logf("Capabilities: Device=%v, Media=%v, Imaging=%v, PTZ=%v",
			caps.Device != nil, caps.Media != nil, caps.Imaging != nil, caps.PTZ != nil)
	})

	t.Run("GetProfiles", func(t *testing.T) {
		profiles, err := client.GetProfiles(ctx)
		if err != nil {
			t.Errorf("GetProfiles failed: %v", err)
			return
		}

		if len(profiles) == 0 {
			t.Error("No profiles returned")
		}

		t.Logf("Found %d profile(s)", len(profiles))
		for i, profile := range profiles {
			t.Logf("  Profile %d: %s (Token: %s)", i+1, profile.Name, profile.Token)
		}
	})

}
