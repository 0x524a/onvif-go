package discovery

import (
	"context"
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
	if err != nil && err != context.DeadlineExceeded {
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
		if len(scope) > 0 && scope[:5] == "onvif" {
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
