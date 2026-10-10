package main

import "testing"

func TestGoTestIdentifier(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"AXIS_P3818-PVE_11.9.60", "AXIS_P3818PVE_11960"},
		{"Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066", "Bosch_FLEXIDOME_indoor_5100i_IR_8710066"},
		{"unknown_device", "Unknown_device"}, // was "Testunknown_device": not a test
		{"1st-camera", "Camera1stcamera"},
		{"", "Camera"},
		{"---", "Camera"},
	}

	for _, tt := range tests {
		if got := goTestIdentifier(tt.in); got != tt.want {
			t.Errorf("goTestIdentifier(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMakeRelativePath(t *testing.T) {
	// An archive that sits next to the generated test must be referenced by
	// its bare file name: the test runs with its own directory as the working
	// directory. This used to come out as "captures/<file>".
	if got := makeRelativePath("testing/captures/x.tar.gz", "testing/captures/"); got != "x.tar.gz" {
		t.Errorf("same directory: got %q, want %q", got, "x.tar.gz")
	}

	if got := makeRelativePath("camera-logs/x.tar.gz", "testing/captures"); got != "../../camera-logs/x.tar.gz" {
		t.Errorf("different directory: got %q, want %q", got, "../../camera-logs/x.tar.gz")
	}
}
