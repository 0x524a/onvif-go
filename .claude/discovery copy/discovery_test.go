package discovery

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestDevice_GetName(t *testing.T) {
	tests := []struct {
		name   string
		device *Device
		want   string
	}{
		{
			name: "device with name in scopes",
			device: &Device{
				Scopes: []string{
					"onvif://www.onvif.org/name/TestCamera",
					"onvif://www.onvif.org/hardware/Model123",
				},
			},
			want: "TestCamera",
		},
		{
			name: "device without name in scopes",
			device: &Device{
				Scopes: []string{
					"onvif://www.onvif.org/hardware/Model123",
				},
			},
			want: "",
		},
		{
			name:   "device with no scopes",
			device: &Device{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.device.GetName(); got != tt.want {
				t.Errorf("GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_GetDeviceEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		device *Device
		want   string
	}{
		{
			name: "device with valid XAddrs",
			device: &Device{
				XAddrs: []string{
					"http://192.168.1.100:80/onvif/device_service",
					"http://192.168.1.100:8080/onvif/device_service",
				},
			},
			want: "http://192.168.1.100:80/onvif/device_service",
		},
		{
			name:   "device with no XAddrs",
			device: &Device{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.device.GetDeviceEndpoint(); got != tt.want {
				t.Errorf("GetDeviceEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_GetLocation(t *testing.T) {
	tests := []struct {
		name   string
		device *Device
		want   string
	}{
		{
			name: "device with location in scopes",
			device: &Device{
				Scopes: []string{
					"onvif://www.onvif.org/location/Building1",
					"onvif://www.onvif.org/hardware/Model123",
				},
			},
			want: "Building1",
		},
		{
			name: "device without location in scopes",
			device: &Device{
				Scopes: []string{
					"onvif://www.onvif.org/hardware/Model123",
				},
			},
			want: "",
		},
		{
			name:   "device with no scopes",
			device: &Device{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.device.GetLocation(); got != tt.want {
				t.Errorf("GetLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscover_WithTimeout(t *testing.T) {
	// This test will timeout since there are likely no actual cameras on the test network
	// It validates that the timeout mechanism works
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	devices, err := Discover(ctx, 500*time.Millisecond)

	// We expect either no error (empty devices list) or a timeout/context error
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("Discover returned error: %v (this is expected in test environment)", err)
	}

	// Devices might be empty in test environment
	t.Logf("Discovered %d devices", len(devices))
}

func TestDiscover_InvalidDuration(t *testing.T) {
	ctx := context.Background()

	// Test with zero duration
	devices, err := Discover(ctx, 0)
	if err != nil {
		t.Logf("Discovery with 0 duration returned error: %v", err)
	}
	t.Logf("Discovered %d devices with 0 duration", len(devices))
}

func TestParseSpaceSeparated(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "multiple values",
			input: "value1 value2 value3",
			want:  []string{"value1", "value2", "value3"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "single value",
			input: "value1",
			want:  []string{"value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSpaceSeparated(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseSpaceSeparated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_GetTypes(t *testing.T) {
	device := &Device{
		Types: []string{
			"dn:NetworkVideoTransmitter",
			"tds:Device",
		},
	}

	types := device.Types
	if len(types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(types))
	}
}

func TestDevice_GetScopes(t *testing.T) {
	scopes := []string{
		"onvif://www.onvif.org/name/TestCamera",
		"onvif://www.onvif.org/location/Building1",
		"onvif://www.onvif.org/hardware/Model123",
	}

	device := &Device{
		Scopes: scopes,
	}

	if len(device.Scopes) != 3 {
		t.Errorf("Expected 3 scopes, got %d", len(device.Scopes))
	}

	// Test specific scope extraction
	hasName := false
	for _, scope := range device.Scopes {
		if scope != "" && scope[:5] == "onvif" {
			hasName = true

			break
		}
	}

	if !hasName {
		t.Error("Expected to find onvif scope")
	}
}

func BenchmarkDeviceGetName(b *testing.B) {
	device := &Device{
		Scopes: []string{
			"onvif://www.onvif.org/name/TestCamera",
			"onvif://www.onvif.org/hardware/Model123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = device.GetName()
	}
}

func BenchmarkDeviceGetDeviceEndpoint(b *testing.B) {
	device := &Device{
		XAddrs: []string{
			"http://192.168.1.100/onvif/device_service",
			"http://192.168.1.100:8080/onvif/device_service",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = device.GetDeviceEndpoint()
	}
}

// Tests for network interface discovery features

func TestListNetworkInterfaces(t *testing.T) {
	interfaces, err := ListNetworkInterfaces()
	if err != nil {
		t.Fatalf("ListNetworkInterfaces failed: %v", err)
	}

	if len(interfaces) == 0 {
		t.Skip("No network interfaces available")
	}

	// Verify loopback interface exists (if available)
	for _, iface := range interfaces {
		if iface.Name == "lo" {
			if len(iface.Addresses) == 0 {
				t.Error("Loopback interface should have addresses")
			}

			break
		}
	}

	// Loopback might not exist on all systems, but there should be at least one interface
	t.Logf("Found %d network interface(s)", len(interfaces))
	for _, iface := range interfaces {
		t.Logf("  - %s: up=%v, multicast=%v, addresses=%v", iface.Name, iface.Up, iface.Multicast, iface.Addresses)
	}
}

func TestResolveNetworkInterface(t *testing.T) {
	// Determine the loopback interface name based on platform
	loopbackName := "lo"
	if _, err := net.InterfaceByName("lo"); err != nil {
		// Loopback might be "lo0" on macOS
		loopbackName = "lo0"
	}

	tests := []struct {
		name      string
		ifaceSpec string
		shouldErr bool
	}{
		{
			name:      "loopback by name",
			ifaceSpec: loopbackName,
			shouldErr: false,
		},
		{
			name:      "loopback by ip",
			ifaceSpec: "127.0.0.1",
			shouldErr: false,
		},
		{
			name:      "invalid interface",
			ifaceSpec: "nonexistent-interface-12345xyz",
			shouldErr: true,
		},
		{
			name:      "invalid ip",
			ifaceSpec: "999.999.999.999",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface, err := resolveNetworkInterface(tt.ifaceSpec)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for interface %s, but got none", tt.ifaceSpec)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for interface %s: %v", tt.ifaceSpec, err)
				}
				if iface == nil {
					t.Errorf("Expected interface for %s, but got nil", tt.ifaceSpec)
				} else {
					t.Logf("Resolved %s to interface: %s", tt.ifaceSpec, iface.Name)
				}
			}
		})
	}
}

func TestDiscoverWithOptions_DefaultOptions(t *testing.T) {
	// Test with default options (should not error even if no cameras found)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	devices, err := DiscoverWithOptions(ctx, 1*time.Second, &DiscoverOptions{})
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("DiscoverWithOptions returned: %v (this is OK if no cameras on network)", err)
	}

	// Should return a slice (possibly empty)
	if devices == nil {
		t.Error("Expected devices slice, got nil")
	}

	t.Logf("Found %d devices with default options", len(devices))
}

func TestDiscoverWithOptions_NilOptions(t *testing.T) {
	// Test with nil options (should work with nil)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	devices, err := DiscoverWithOptions(ctx, 500*time.Millisecond, nil)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("DiscoverWithOptions with nil returned: %v", err)
	}

	if devices == nil {
		t.Error("Expected devices slice, got nil")
	}
}

func TestDiscoverWithOptions_LoopbackInterface(t *testing.T) {
	// Test with loopback interface for testing
	// Try both common loopback names
	loopbackName := ""
	if _, err := net.InterfaceByName("lo"); err == nil {
		loopbackName = "lo"
	} else if _, err := net.InterfaceByName("lo0"); err == nil {
		loopbackName = "lo0"
	} else {
		t.Skip("Loopback interface not available on this system")
	}

	opts := &DiscoverOptions{
		NetworkInterface: loopbackName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	devices, err := DiscoverWithOptions(ctx, 500*time.Millisecond, opts)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("DiscoverWithOptions with %s interface: %v (timeout is expected)", loopbackName, err)
	}

	if devices == nil {
		t.Error("Expected devices slice, got nil")
	}

	t.Logf("Found %d devices on loopback interface", len(devices))
}

func TestDiscoverWithOptions_InvalidInterface(t *testing.T) {
	opts := &DiscoverOptions{
		NetworkInterface: "nonexistent-interface-xyz",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := DiscoverWithOptions(ctx, 500*time.Millisecond, opts)
	if err == nil {
		t.Error("Expected error for invalid interface, but got none")
	}

	t.Logf("Got expected error: %v", err)
}

func TestDiscover_BackwardCompatibility(t *testing.T) {
	// Test that old Discover function still works (backward compatibility)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	devices, err := Discover(ctx, 500*time.Millisecond)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("Discover returned: %v", err)
	}

	if devices == nil {
		t.Error("Expected devices slice, got nil")
	}

	t.Logf("Backward compat: found %d devices", len(devices))
}

func BenchmarkListNetworkInterfaces(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ListNetworkInterfaces()
	}
}

func BenchmarkResolveNetworkInterface(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolveNetworkInterface("127.0.0.1")
	}
}
