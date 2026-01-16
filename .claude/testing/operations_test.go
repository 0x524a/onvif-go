package onviftesting

import (
	"testing"
)

func TestAllReadOperations(t *testing.T) {
	ops := AllReadOperations()

	if len(ops) == 0 {
		t.Error("AllReadOperations should return operations")
	}

	// Check we have significant coverage
	if len(ops) < 100 {
		t.Errorf("Expected at least 100 READ operations, got %d", len(ops))
	}

	// Verify all operations have names
	for i, op := range ops {
		if op.Name == "" {
			t.Errorf("Operation %d has empty name", i)
		}
		if op.Service == "" {
			t.Errorf("Operation %s has empty service", op.Name)
		}
	}
}

func TestGetOperationCount(t *testing.T) {
	count := GetOperationCount()

	if count.Total == 0 {
		t.Error("Total should be greater than 0")
	}

	expectedTotal := count.Device + count.Media + count.PTZ + count.Imaging + count.Event + count.DeviceIO
	if count.Total != expectedTotal {
		t.Errorf("Total = %d, but sum of services = %d", count.Total, expectedTotal)
	}

	// Verify we have operations in major services
	if count.Device == 0 {
		t.Error("Device operations should be > 0")
	}
	if count.Media == 0 {
		t.Error("Media operations should be > 0")
	}
}

func TestReadOperationsByService(t *testing.T) {
	tests := []struct {
		service ServiceType
		minOps  int
	}{
		{ServiceDevice, 30},
		{ServiceMedia, 40},
		{ServicePTZ, 4},
		{ServiceImaging, 3},
		{ServiceEvent, 2},
		{ServiceDeviceIO, 8},
	}

	for _, tt := range tests {
		t.Run(string(tt.service), func(t *testing.T) {
			ops := ReadOperationsByService(tt.service)
			if len(ops) < tt.minOps {
				t.Errorf("ReadOperationsByService(%s) returned %d ops, want at least %d",
					tt.service, len(ops), tt.minOps)
			}
		})
	}
}

func TestIndependentOperations(t *testing.T) {
	independent := IndependentOperations()

	if len(independent) == 0 {
		t.Error("IndependentOperations should return operations")
	}

	// Verify all are actually independent
	for _, op := range independent {
		if op.DependsOn != "" {
			t.Errorf("Operation %s has DependsOn=%s but returned as independent",
				op.Name, op.DependsOn)
		}
	}
}

func TestDependentOperations(t *testing.T) {
	dependent := DependentOperations()

	if len(dependent) == 0 {
		t.Error("DependentOperations should return operations")
	}

	// Verify all are actually dependent
	for _, op := range dependent {
		if op.DependsOn == "" {
			t.Errorf("Operation %s has empty DependsOn but returned as dependent", op.Name)
		}
	}
}

func TestOperationsByDependency(t *testing.T) {
	// GetProfiles is a common dependency
	ops := OperationsByDependency("GetProfiles")

	if len(ops) == 0 {
		t.Error("Operations depending on GetProfiles should exist")
	}

	for _, op := range ops {
		if op.DependsOn != "GetProfiles" {
			t.Errorf("Operation %s has DependsOn=%s, want GetProfiles",
				op.Name, op.DependsOn)
		}
	}
}

func TestGetOperationSpec(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"GetDeviceInformation", true},
		{"GetProfiles", true},
		{"GetStreamURI", true},
		{"GetStatus", true},
		{"NonExistentOperation", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := GetOperationSpec(tt.name)
			if tt.expected && op == nil {
				t.Errorf("GetOperationSpec(%s) returned nil, expected operation", tt.name)
			}
			if !tt.expected && op != nil {
				t.Errorf("GetOperationSpec(%s) returned operation, expected nil", tt.name)
			}
		})
	}
}

func TestOperationSpec_DependencyChain(t *testing.T) {
	// Test that dependent operations reference valid dependencies
	dependent := DependentOperations()

	for _, op := range dependent {
		depOp := GetOperationSpec(op.DependsOn)
		if depOp == nil {
			t.Errorf("Operation %s depends on %s which doesn't exist",
				op.Name, op.DependsOn)
		}
	}
}

func TestDeviceReadOperations(t *testing.T) {
	// Check for expected core operations
	expectedOps := []string{
		"GetDeviceInformation",
		"GetCapabilities",
		"GetSystemDateAndTime",
		"GetHostname",
		"GetDNS",
		"GetNTP",
		"GetNetworkInterfaces",
		"GetScopes",
		"GetUsers",
	}

	ops := DeviceReadOperations
	opMap := make(map[string]bool)
	for _, op := range ops {
		opMap[op.Name] = true
	}

	for _, expected := range expectedOps {
		if !opMap[expected] {
			t.Errorf("Expected DeviceReadOperations to contain %s", expected)
		}
	}
}

func TestMediaReadOperations(t *testing.T) {
	// Check for expected core operations
	expectedOps := []string{
		"GetProfiles",
		"GetProfile",
		"GetVideoSources",
		"GetAudioSources",
		"GetStreamURI",
		"GetSnapshotURI",
		"GetVideoEncoderConfigurations",
	}

	ops := MediaReadOperations
	opMap := make(map[string]bool)
	for _, op := range ops {
		opMap[op.Name] = true
	}

	for _, expected := range expectedOps {
		if !opMap[expected] {
			t.Errorf("Expected MediaReadOperations to contain %s", expected)
		}
	}
}

func TestOperationCategories(t *testing.T) {
	ops := AllReadOperations()

	// Check that all operations have categories
	for _, op := range ops {
		if op.Category == "" {
			t.Errorf("Operation %s has empty category", op.Name)
		}
	}

	// Check for common categories
	categories := make(map[string]int)
	for _, op := range ops {
		categories[op.Category]++
	}

	expectedCategories := []string{"core", "network", "profiles", "streaming"}
	for _, cat := range expectedCategories {
		if categories[cat] == 0 {
			t.Errorf("Expected category %s to have operations", cat)
		}
	}
}
