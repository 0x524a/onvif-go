// Package testdata provides camera configuration data for testing
// Auto-generated from network discovery on 2026-01-13
package testdata

// DiscoveredCamera represents a camera found on the network
type DiscoveredCamera struct {
	ID            int
	Endpoint      string
	XAddrs        []string
	Manufacturer  string
	Model         string
	IP            string
	Port          int
	Profiles      []string
	SupportsHTTPS bool
}

// TestCameras contains the discovered cameras for testing
var TestCameras = []DiscoveredCamera{
	{
		ID:           1,
		Endpoint:     "urn:uuid:15020314-0204-0408-1500-ec71db465af7",
		XAddrs:       []string{"http://192.168.2.61:8000/onvif/device_service"},
		Manufacturer: "Reolink",
		Model:        "E1Zoom",
		IP:           "192.168.2.61",
		Port:         8000,
		Profiles:     []string{"Streaming", "T"},
	},
	{
		ID:            2,
		Endpoint:      "urn:uuid:00075fe0-a604-04a6-e05f-0700075fe05f",
		XAddrs:        []string{"http://192.168.2.57/onvif/device_service", "https://192.168.2.57/onvif/device_service"},
		Manufacturer:  "Bosch",
		Model:         "AUTODOME_IP_starlight_5000i",
		IP:            "192.168.2.57",
		Port:          80,
		Profiles:      []string{"Streaming", "G", "T"},
		SupportsHTTPS: true,
	},
	{
		ID:           3,
		Endpoint:     "urn:uuid:555a3d17-6698-43d9-9a52-2a199ff14dec",
		XAddrs:       []string{"http://192.168.2.82/onvif/device_service"},
		Manufacturer: "AXIS",
		Model:        "P3818-PVE",
		IP:           "192.168.2.82",
		Port:         80,
		Profiles:     []string{"Streaming", "G", "M", "T"},
	},
	{
		ID:           4,
		Endpoint:     "urn:uuid:12060714-0005-0000-0302-ec71dbe838cc",
		XAddrs:       []string{"http://192.168.2.236:8000/onvif/device_service"},
		Manufacturer: "Reolink",
		Model:        "ReolinkTrackMixWiFi",
		IP:           "192.168.2.236",
		Port:         8000,
		Profiles:     []string{"Streaming", "T"},
	},
	{
		ID:            5,
		Endpoint:      "urn:uuid:00075fca-f8fa-faf8-ca5f-0700075fca5f",
		XAddrs:        []string{"http://192.168.2.200/onvif/device_service", "https://192.168.2.200/onvif/device_service"},
		Manufacturer:  "Bosch",
		Model:         "FLEXIDOME_IP_starlight_8000i",
		IP:            "192.168.2.200",
		Port:          80,
		Profiles:      []string{"Streaming", "G", "T"},
		SupportsHTTPS: true,
	},
	{
		ID:            6,
		Endpoint:      "urn:uuid:00075fd5-9fbe-be9f-d55f-0700075fd55f",
		XAddrs:        []string{"http://192.168.2.24/onvif/device_service", "https://192.168.2.24/onvif/device_service"},
		Manufacturer:  "Bosch",
		Model:         "FLEXIDOME_panoramic_5100i",
		IP:            "192.168.2.24",
		Port:          80,
		Profiles:      []string{"Streaming", "G", "T", "M"},
		SupportsHTTPS: true,
	},
	{
		ID:            7,
		Endpoint:      "urn:uuid:cbc93166-2a81-4635-9fe3-dcd5e99528d3",
		XAddrs:        []string{"http://192.168.2.190/onvif/device_service", "https://192.168.2.190/onvif/device_service"},
		Manufacturer:  "AXIS",
		Model:         "Q3819-PVE",
		IP:            "192.168.2.190",
		Port:          80,
		Profiles:      []string{"Streaming", "G", "M", "T"},
		SupportsHTTPS: true,
	},
	{
		ID:            8,
		Endpoint:      "urn:uuid:9e8de0a1-c818-448d-90eb-85670b2b9872",
		XAddrs:        []string{"http://192.168.2.30/onvif/device_service", "https://192.168.2.30/onvif/device_service"},
		Manufacturer:  "AXIS",
		Model:         "P5655-E",
		IP:            "192.168.2.30",
		Port:          80,
		Profiles:      []string{"Streaming", "G", "M", "T"},
		SupportsHTTPS: true,
	},
}

// GetCameraByManufacturer returns cameras filtered by manufacturer
func GetCameraByManufacturer(manufacturer string) []DiscoveredCamera {
	var result []DiscoveredCamera
	for _, cam := range TestCameras {
		if cam.Manufacturer == manufacturer {
			result = append(result, cam)
		}
	}
	return result
}

// GetCameraByProfile returns cameras that support a specific profile
func GetCameraByProfile(profile string) []DiscoveredCamera {
	var result []DiscoveredCamera
	for _, cam := range TestCameras {
		for _, p := range cam.Profiles {
			if p == profile {
				result = append(result, cam)
				break
			}
		}
	}
	return result
}

// GetHTTPSCameras returns cameras that support HTTPS
func GetHTTPSCameras() []DiscoveredCamera {
	var result []DiscoveredCamera
	for _, cam := range TestCameras {
		if cam.SupportsHTTPS {
			result = append(result, cam)
		}
	}
	return result
}
