package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	onviftesting "github.com/0x524a/onvif-go/testing"
)

var (
	captureArchive = flag.String("capture", "", "Path to XML capture archive (.tar.gz)")
	outputDir      = flag.String("output", "./", "Output directory for generated test file")
	packageName    = flag.String("package", "onvif_test", "Package name for generated test")
	updateRegistry = flag.Bool("update-registry", true, "Update registry.json with camera info")
	registryPath   = flag.String("registry", "", "Path to registry.json (default: testdata/captures/registry.json)")
	coverageReport = flag.Bool("coverage-report", false, "Generate coverage report from registry")
	coverageOutput = flag.String("coverage-output", "", "Output path for coverage report (default: stdout)")
)

const testTemplate = `package {{.PackageName}}

import (
	"context"
	"testing"
	"time"

	"github.com/0x524a/onvif-go"
	onviftesting "github.com/0x524a/onvif-go/testing"
)

// Test{{.CameraName}} tests ONVIF client against {{.CameraDescription}} captured responses.
// Capture format: V2 with parameter-aware matching
// Total captured operations: {{.TotalExchanges}}
func Test{{.CameraName}}(t *testing.T) {
	// Load capture archive (relative to project root)
	captureArchive := "{{.CaptureArchiveRelPath}}"

	mockServer, err := onviftesting.NewMockSOAPServerV2(captureArchive)
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

	// =========================================================================
	// Device Service Operations
	// =========================================================================
{{range .DeviceTests}}
	t.Run("{{.Name}}", func(t *testing.T) {
		{{.Code}}
	})
{{end}}
	// =========================================================================
	// Media Service Operations
	// =========================================================================
{{if .NeedsInit}}
	// Initialize to discover service endpoints (required for Media/PTZ/Imaging)
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}
{{end}}
{{range .MediaTests}}
	t.Run("{{.Name}}", func(t *testing.T) {
		{{.Code}}
	})
{{end}}
	// =========================================================================
	// Profile-Dependent Operations
	// =========================================================================
{{range .ProfileTests}}
	t.Run("{{.Name}}", func(t *testing.T) {
		{{.Code}}
	})
{{end}}
	// =========================================================================
	// PTZ Operations
	// =========================================================================
{{range .PTZTests}}
	t.Run("{{.Name}}", func(t *testing.T) {
		{{.Code}}
	})
{{end}}
	// =========================================================================
	// Imaging Operations
	// =========================================================================
{{range .ImagingTests}}
	t.Run("{{.Name}}", func(t *testing.T) {
		{{.Code}}
	})
{{end}}
}
`

type TestData struct {
	PackageName           string
	CameraName            string
	CameraDescription     string
	CaptureArchiveRelPath string
	TotalExchanges        int
	NeedsInit             bool
	DeviceTests           []GeneratedTest
	MediaTests            []GeneratedTest
	ProfileTests          []GeneratedTest
	PTZTests              []GeneratedTest
	ImagingTests          []GeneratedTest
}

type GeneratedTest struct {
	Name string
	Code string
}

// operationInfo holds info about captured operations
type operationInfo struct {
	OperationName string
	ServiceType   onviftesting.ServiceType
	Parameters    map[string]interface{}
	Success       bool
}

func main() {
	flag.Parse()

	// Set default registry path
	regPath := *registryPath
	if regPath == "" {
		regPath = onviftesting.DefaultRegistryPath
	}

	// Handle coverage report mode
	if *coverageReport {
		generateCoverageReport(regPath)
		return
	}

	if *captureArchive == "" {
		fmt.Println("Error: -capture flag is required")
		fmt.Println()
		fmt.Println("Usage:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  ./generate-tests -capture camera-logs/Bosch_FLEXIDOME_indoor_5100i_IR_8.71.0066_xmlcapture_*.tar.gz")
		fmt.Println()
		fmt.Println("Coverage report:")
		fmt.Println("  ./generate-tests -coverage-report")
		os.Exit(1)
	}

	outputFile := generateTests()

	// Update registry if requested
	if *updateRegistry {
		updateCameraRegistry(regPath, *captureArchive, outputFile)
	}
}

func generateTests() string {
	// Load capture with V2 support
	capture, metadata, err := onviftesting.LoadCaptureFromArchiveV2(*captureArchive)
	if err != nil {
		log.Fatalf("Failed to load capture: %v", err)
	}

	// Extract camera name from archive filename
	baseName := filepath.Base(*captureArchive)
	parts := strings.Split(baseName, "_xmlcapture_")
	cameraID := parts[0]

	// Convert to valid Go identifier
	cameraName := strings.ReplaceAll(cameraID, "-", "")
	cameraName = strings.ReplaceAll(cameraName, ".", "")
	cameraName = strings.ReplaceAll(cameraName, " ", "")

	// Get camera description from metadata or extract from captures
	cameraDesc := cameraID
	if metadata != nil && metadata.CameraInfo.Manufacturer != "" {
		cameraDesc = fmt.Sprintf("%s %s (Firmware: %s)",
			metadata.CameraInfo.Manufacturer,
			metadata.CameraInfo.Model,
			metadata.CameraInfo.FirmwareVersion)
	} else {
		// Try to extract from GetDeviceInformation response
		for _, ex := range capture.Exchanges {
			if ex.OperationName == "GetDeviceInformation" && ex.Success {
				manufacturer := extractXMLValue(ex.ResponseBody, "Manufacturer")
				model := extractXMLValue(ex.ResponseBody, "Model")
				firmware := extractXMLValue(ex.ResponseBody, "FirmwareVersion")
				if manufacturer != "" && model != "" {
					cameraDesc = fmt.Sprintf("%s %s (Firmware: %s)", manufacturer, model, firmware)
				}
				break
			}
		}
	}

	// Analyze captured operations
	ops := analyzeOperations(capture)

	// Generate tests by service type
	testData := TestData{
		PackageName:           *packageName,
		CameraName:            cameraName,
		CameraDescription:     cameraDesc,
		CaptureArchiveRelPath: makeRelativePath(*captureArchive, *outputDir),
		TotalExchanges:        len(capture.Exchanges),
		NeedsInit:             hasNonDeviceOperations(ops),
		DeviceTests:           generateDeviceTests(ops),
		MediaTests:            generateMediaTests(ops),
		ProfileTests:          generateProfileDependentTests(ops),
		PTZTests:              generatePTZTests(ops),
		ImagingTests:          generateImagingTests(ops),
	}

	// Generate test file
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	outputFile := filepath.Join(*outputDir, fmt.Sprintf("%s_test.go", strings.ToLower(cameraID)))
	f, err := os.Create(outputFile) //nolint:gosec // Filename is generated from test data, safe
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := tmpl.Execute(f, testData); err != nil {
		_ = f.Close()
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Printf("✓ Generated test file: %s\n", outputFile)
	fmt.Printf("  Camera: %s\n", cameraDesc)
	fmt.Printf("  Captured operations: %d\n", len(capture.Exchanges))
	fmt.Printf("  Generated subtests: Device=%d, Media=%d, Profile=%d, PTZ=%d, Imaging=%d\n",
		len(testData.DeviceTests), len(testData.MediaTests), len(testData.ProfileTests),
		len(testData.PTZTests), len(testData.ImagingTests))
	fmt.Println()
	fmt.Println("Run tests with:")
	fmt.Printf("  go test -v %s\n", outputFile)

	return outputFile
}

func analyzeOperations(capture *onviftesting.CameraCaptureV2) []operationInfo {
	var ops []operationInfo
	seen := make(map[string]bool)

	for _, ex := range capture.Exchanges {
		// Create unique key for deduplication
		key := ex.OperationName
		if token := ex.GetProfileToken(); token != "" {
			key += "_" + token
		} else if token := ex.GetConfigurationToken(); token != "" {
			key += "_" + token
		} else if token := ex.GetVideoSourceToken(); token != "" {
			key += "_" + token
		}

		if seen[key] {
			continue
		}
		seen[key] = true

		ops = append(ops, operationInfo{
			OperationName: ex.OperationName,
			ServiceType:   ex.ServiceType,
			Parameters:    ex.Parameters,
			Success:       ex.Success,
		})
	}

	return ops
}

func hasNonDeviceOperations(ops []operationInfo) bool {
	for _, op := range ops {
		switch op.ServiceType {
		case onviftesting.ServiceMedia, onviftesting.ServicePTZ, onviftesting.ServiceImaging:
			return true
		}
	}
	return false
}

func generateDeviceTests(ops []operationInfo) []GeneratedTest {
	var tests []GeneratedTest

	// Standard device tests
	deviceOps := map[string]string{
		"GetDeviceInformation": `info, err := client.GetDeviceInformation(ctx)
		if err != nil {
			t.Errorf("GetDeviceInformation failed: %v", err)
			return
		}
		if info.Manufacturer == "" {
			t.Error("Manufacturer is empty")
		}
		if info.Model == "" {
			t.Error("Model is empty")
		}
		t.Logf("Device: %s %s (Firmware: %s)", info.Manufacturer, info.Model, info.FirmwareVersion)`,

		"GetSystemDateAndTime": `_, err := client.GetSystemDateAndTime(ctx)
		if err != nil {
			t.Errorf("GetSystemDateAndTime failed: %v", err)
		}`,

		"GetCapabilities": `caps, err := client.GetCapabilities(ctx)
		if err != nil {
			t.Errorf("GetCapabilities failed: %v", err)
			return
		}
		t.Logf("Capabilities: Device=%v, Media=%v, Imaging=%v, PTZ=%v",
			caps.Device != nil, caps.Media != nil, caps.Imaging != nil, caps.PTZ != nil)`,

		"GetHostname": `hostname, err := client.GetHostname(ctx)
		if err != nil {
			t.Errorf("GetHostname failed: %v", err)
			return
		}
		t.Logf("Hostname: %s", hostname)`,

		"GetScopes": `scopes, err := client.GetScopes(ctx)
		if err != nil {
			t.Errorf("GetScopes failed: %v", err)
			return
		}
		t.Logf("Scopes: %d", len(scopes))`,

		"GetNetworkInterfaces": `interfaces, err := client.GetNetworkInterfaces(ctx)
		if err != nil {
			t.Errorf("GetNetworkInterfaces failed: %v", err)
			return
		}
		t.Logf("Network interfaces: %d", len(interfaces))`,

		"GetServices": `services, err := client.GetServices(ctx, true)
		if err != nil {
			t.Errorf("GetServices failed: %v", err)
			return
		}
		t.Logf("Services: %d", len(services))`,
	}

	// Generate tests for captured operations
	for _, op := range ops {
		if op.ServiceType != onviftesting.ServiceDevice && op.ServiceType != onviftesting.ServiceUnknown {
			continue
		}
		if code, ok := deviceOps[op.OperationName]; ok {
			tests = append(tests, GeneratedTest{
				Name: op.OperationName,
				Code: code,
			})
			delete(deviceOps, op.OperationName) // Don't duplicate
		}
	}

	// Sort by name for consistent output
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].Name < tests[j].Name
	})

	return tests
}

func generateMediaTests(ops []operationInfo) []GeneratedTest {
	var tests []GeneratedTest

	mediaOps := map[string]string{
		"GetProfiles": `profiles, err := client.GetProfiles(ctx)
		if err != nil {
			t.Errorf("GetProfiles failed: %v", err)
			return
		}
		if len(profiles) == 0 {
			t.Error("No profiles returned")
		}
		t.Logf("Found %d profile(s)", len(profiles))`,

		"GetVideoSources": `sources, err := client.GetVideoSources(ctx)
		if err != nil {
			t.Errorf("GetVideoSources failed: %v", err)
			return
		}
		t.Logf("Video sources: %d", len(sources))`,

		"GetVideoSourceConfigurations": `configs, err := client.GetVideoSourceConfigurations(ctx)
		if err != nil {
			t.Errorf("GetVideoSourceConfigurations failed: %v", err)
			return
		}
		t.Logf("Video source configs: %d", len(configs))`,

		"GetVideoEncoderConfigurations": `configs, err := client.GetVideoEncoderConfigurations(ctx)
		if err != nil {
			t.Errorf("GetVideoEncoderConfigurations failed: %v", err)
			return
		}
		t.Logf("Video encoder configs: %d", len(configs))`,

		"GetAudioSources": `sources, err := client.GetAudioSources(ctx)
		if err != nil {
			t.Errorf("GetAudioSources failed: %v", err)
			return
		}
		t.Logf("Audio sources: %d", len(sources))`,

		"GetAudioSourceConfigurations": `configs, err := client.GetAudioSourceConfigurations(ctx)
		if err != nil {
			t.Errorf("GetAudioSourceConfigurations failed: %v", err)
			return
		}
		t.Logf("Audio source configs: %d", len(configs))`,

		"GetMetadataConfigurations": `configs, err := client.GetMetadataConfigurations(ctx)
		if err != nil {
			t.Errorf("GetMetadataConfigurations failed: %v", err)
			return
		}
		t.Logf("Metadata configs: %d", len(configs))`,
	}

	for _, op := range ops {
		if op.ServiceType != onviftesting.ServiceMedia {
			continue
		}
		if code, ok := mediaOps[op.OperationName]; ok {
			tests = append(tests, GeneratedTest{
				Name: op.OperationName,
				Code: code,
			})
			delete(mediaOps, op.OperationName)
		}
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].Name < tests[j].Name
	})

	return tests
}

func generateProfileDependentTests(ops []operationInfo) []GeneratedTest {
	var tests []GeneratedTest

	// Group operations by profile token
	profileOps := make(map[string][]operationInfo)
	for _, op := range ops {
		if token, ok := op.Parameters["ProfileToken"].(string); ok && token != "" {
			profileOps[token] = append(profileOps[token], op)
		}
	}

	// Generate GetStreamURI tests for each profile
	for token, opList := range profileOps {
		for _, op := range opList {
			switch op.OperationName {
			case "GetStreamURI":
				testName := fmt.Sprintf("GetStreamURI_%s", sanitizeToken(token))
				tests = append(tests, GeneratedTest{
					Name: testName,
					Code: fmt.Sprintf(`uri, err := client.GetStreamURI(ctx, "%s")
		if err != nil {
			t.Errorf("GetStreamURI failed: %%v", err)
			return
		}
		if uri.URI == "" {
			t.Error("Stream URI is empty")
		}
		t.Logf("Stream URI: %%s", uri.URI)`, token),
				})

			case "GetSnapshotURI":
				testName := fmt.Sprintf("GetSnapshotURI_%s", sanitizeToken(token))
				tests = append(tests, GeneratedTest{
					Name: testName,
					Code: fmt.Sprintf(`uri, err := client.GetSnapshotURI(ctx, "%s")
		if err != nil {
			t.Errorf("GetSnapshotURI failed: %%v", err)
			return
		}
		if uri.URI == "" {
			t.Error("Snapshot URI is empty")
		}
		t.Logf("Snapshot URI: %%s", uri.URI)`, token),
				})

			case "GetProfile":
				testName := fmt.Sprintf("GetProfile_%s", sanitizeToken(token))
				tests = append(tests, GeneratedTest{
					Name: testName,
					Code: fmt.Sprintf(`profile, err := client.GetProfile(ctx, "%s")
		if err != nil {
			t.Errorf("GetProfile failed: %%v", err)
			return
		}
		if profile.Token != "%s" {
			t.Errorf("Expected token %%s, got %%s", "%s", profile.Token)
		}
		t.Logf("Profile: %%s", profile.Name)`, token, token, token),
				})
			}
		}
	}

	// Deduplicate tests
	seen := make(map[string]bool)
	var uniqueTests []GeneratedTest
	for _, t := range tests {
		if !seen[t.Name] {
			seen[t.Name] = true
			uniqueTests = append(uniqueTests, t)
		}
	}

	sort.Slice(uniqueTests, func(i, j int) bool {
		return uniqueTests[i].Name < uniqueTests[j].Name
	})

	return uniqueTests
}

func generatePTZTests(ops []operationInfo) []GeneratedTest {
	var tests []GeneratedTest

	ptzOps := map[string]string{
		"GetNodes": `nodes, err := client.GetNodes(ctx)
		if err != nil {
			t.Errorf("GetNodes failed: %v", err)
			return
		}
		t.Logf("PTZ nodes: %d", len(nodes))`,

		"GetConfigurations": `configs, err := client.GetConfigurations(ctx)
		if err != nil {
			t.Errorf("GetConfigurations failed: %v", err)
			return
		}
		t.Logf("PTZ configs: %d", len(configs))`,
	}

	// Group by profile token for status and presets
	profileOps := make(map[string][]operationInfo)
	for _, op := range ops {
		if op.ServiceType != onviftesting.ServicePTZ {
			continue
		}
		if code, ok := ptzOps[op.OperationName]; ok {
			tests = append(tests, GeneratedTest{
				Name: op.OperationName,
				Code: code,
			})
			delete(ptzOps, op.OperationName)
			continue
		}
		if token, ok := op.Parameters["ProfileToken"].(string); ok && token != "" {
			profileOps[token] = append(profileOps[token], op)
		}
	}

	// Generate profile-specific PTZ tests
	for token, opList := range profileOps {
		for _, op := range opList {
			switch op.OperationName {
			case "GetStatus":
				testName := fmt.Sprintf("PTZ_GetStatus_%s", sanitizeToken(token))
				tests = append(tests, GeneratedTest{
					Name: testName,
					Code: fmt.Sprintf(`status, err := client.GetStatus(ctx, "%s")
		if err != nil {
			t.Errorf("GetStatus failed: %%v", err)
			return
		}
		t.Logf("PTZ Status retrieved for profile %s")
		_ = status`, token, token),
				})

			case "GetPresets":
				testName := fmt.Sprintf("PTZ_GetPresets_%s", sanitizeToken(token))
				tests = append(tests, GeneratedTest{
					Name: testName,
					Code: fmt.Sprintf(`presets, err := client.GetPresets(ctx, "%s")
		if err != nil {
			t.Errorf("GetPresets failed: %%v", err)
			return
		}
		t.Logf("Found %%d preset(s) for profile %s", len(presets))`, token, token),
				})
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var uniqueTests []GeneratedTest
	for _, t := range tests {
		if !seen[t.Name] {
			seen[t.Name] = true
			uniqueTests = append(uniqueTests, t)
		}
	}

	sort.Slice(uniqueTests, func(i, j int) bool {
		return uniqueTests[i].Name < uniqueTests[j].Name
	})

	return uniqueTests
}

func generateImagingTests(ops []operationInfo) []GeneratedTest {
	var tests []GeneratedTest

	// Group by video source token
	sourceOps := make(map[string][]operationInfo)
	for _, op := range ops {
		if op.ServiceType != onviftesting.ServiceImaging {
			continue
		}
		if token, ok := op.Parameters["VideoSourceToken"].(string); ok && token != "" {
			sourceOps[token] = append(sourceOps[token], op)
		}
	}

	for token, opList := range sourceOps {
		for _, op := range opList {
			switch op.OperationName {
			case "GetImagingSettings":
				testName := fmt.Sprintf("GetImagingSettings_%s", sanitizeToken(token))
				tests = append(tests, GeneratedTest{
					Name: testName,
					Code: fmt.Sprintf(`settings, err := client.GetImagingSettings(ctx, "%s")
		if err != nil {
			t.Errorf("GetImagingSettings failed: %%v", err)
			return
		}
		t.Logf("Imaging settings retrieved for source %s")
		_ = settings`, token, token),
				})

			case "GetOptions":
				testName := fmt.Sprintf("GetImagingOptions_%s", sanitizeToken(token))
				tests = append(tests, GeneratedTest{
					Name: testName,
					Code: fmt.Sprintf(`options, err := client.GetOptions(ctx, "%s")
		if err != nil {
			t.Errorf("GetOptions failed: %%v", err)
			return
		}
		t.Logf("Imaging options retrieved for source %s")
		_ = options`, token, token),
				})
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var uniqueTests []GeneratedTest
	for _, t := range tests {
		if !seen[t.Name] {
			seen[t.Name] = true
			uniqueTests = append(uniqueTests, t)
		}
	}

	sort.Slice(uniqueTests, func(i, j int) bool {
		return uniqueTests[i].Name < uniqueTests[j].Name
	})

	return uniqueTests
}

func sanitizeToken(token string) string {
	// Make token safe for test name
	token = strings.ReplaceAll(token, "-", "_")
	token = strings.ReplaceAll(token, ".", "_")
	token = strings.ReplaceAll(token, " ", "_")
	// Truncate if too long
	if len(token) > 20 {
		token = token[:20]
	}
	return token
}

func makeRelativePath(archivePath, outputDir string) string {
	if absOutput, err := filepath.Abs(outputDir); err == nil {
		if absArchive, err := filepath.Abs(archivePath); err == nil {
			if rel, err := filepath.Rel(filepath.Dir(absOutput), absArchive); err == nil {
				return rel
			}
		}
	}
	return archivePath
}

func extractXMLValue(xmlStr, tagName string) string {
	start := fmt.Sprintf("<%s>", tagName)
	end := fmt.Sprintf("</%s>", tagName)

	startIdx := strings.Index(xmlStr, start)
	if startIdx == -1 {
		start = fmt.Sprintf(":%s>", tagName)
		startIdx = strings.Index(xmlStr, start)
		if startIdx == -1 {
			return ""
		}
		startIdx += len(start)
	} else {
		startIdx += len(start)
	}

	endIdx := strings.Index(xmlStr[startIdx:], end)
	if endIdx == -1 {
		end = fmt.Sprintf(":/%s>", tagName)
		endIdx = strings.Index(xmlStr[startIdx:], end)
		if endIdx == -1 {
			return ""
		}
	}

	return strings.TrimSpace(xmlStr[startIdx : startIdx+endIdx])
}

// updateCameraRegistry updates the registry with camera information from the capture.
func updateCameraRegistry(regPath, archivePath, testFile string) {
	registry, err := onviftesting.LoadRegistry(regPath)
	if err != nil {
		log.Printf("Warning: Failed to load registry: %v", err)
		return
	}

	entry, err := onviftesting.CreateCameraEntryFromCapture(archivePath)
	if err != nil {
		log.Printf("Warning: Failed to create registry entry: %v", err)
		return
	}

	// Set the test file path (relative to registry directory)
	if testFile != "" {
		regDir := filepath.Dir(regPath)
		if absTest, err := filepath.Abs(testFile); err == nil {
			if absRegDir, err := filepath.Abs(regDir); err == nil {
				if rel, err := filepath.Rel(absRegDir, absTest); err == nil {
					entry.TestFile = rel
				}
			}
		}
		if entry.TestFile == "" {
			entry.TestFile = filepath.Base(testFile)
		}
	}

	// Add or update the camera entry
	registry.AddCamera(*entry)

	// Update coverage statistics
	updateRegistryCoverage(registry, archivePath)

	// Save registry
	if err := onviftesting.SaveRegistry(registry, regPath); err != nil {
		log.Printf("Warning: Failed to save registry: %v", err)
		return
	}

	fmt.Printf("✓ Registry updated: %s\n", regPath)
	fmt.Printf("  Camera ID: %s\n", entry.ID)
	fmt.Printf("  Total cameras in registry: %d\n", len(registry.Cameras))
}

// updateRegistryCoverage calculates coverage from captured operations.
func updateRegistryCoverage(registry *onviftesting.Registry, archivePath string) {
	capture, _, err := onviftesting.LoadCaptureFromArchiveV2(archivePath)
	if err != nil {
		return
	}

	// Count unique operations per service
	serviceCounts := make(map[string]map[string]bool)
	for _, ex := range capture.Exchanges {
		service := string(ex.ServiceType)
		if service == "" || service == "Unknown" {
			continue
		}
		if serviceCounts[service] == nil {
			serviceCounts[service] = make(map[string]bool)
		}
		serviceCounts[service][ex.OperationName] = true
	}

	// Get totals from operations registry
	opCounts := onviftesting.GetOperationCount()

	// Update coverage
	registry.Coverage = make(map[string]onviftesting.Coverage)
	for service, ops := range serviceCounts {
		total := 0
		switch service {
		case "Device":
			total = opCounts.Device
		case "Media":
			total = opCounts.Media
		case "PTZ":
			total = opCounts.PTZ
		case "Imaging":
			total = opCounts.Imaging
		case "Event":
			total = opCounts.Event
		case "DeviceIO":
			total = opCounts.DeviceIO
		}

		registry.Coverage[service] = onviftesting.Coverage{
			Total:    total,
			Captured: len(ops),
		}
	}
}

// generateCoverageReport generates a coverage report from the registry.
func generateCoverageReport(regPath string) {
	registry, err := onviftesting.LoadRegistry(regPath)
	if err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	// Generate markdown report
	report := generateCoverageMarkdown(registry)

	// Output to file or stdout
	if *coverageOutput != "" {
		if err := os.WriteFile(*coverageOutput, []byte(report), 0600); err != nil { //nolint:mnd
			log.Fatalf("Failed to write coverage report: %v", err)
		}
		fmt.Printf("✓ Coverage report written to: %s\n", *coverageOutput)
	} else {
		fmt.Println(report)
	}
}

// generateCoverageMarkdown creates a markdown coverage report.
func generateCoverageMarkdown(registry *onviftesting.Registry) string {
	var sb strings.Builder

	sb.WriteString("# ONVIF Operation Coverage Report\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Cameras**: %d\n", len(registry.Cameras)))

	total, captured := registry.GetTotalCoverage()
	if total > 0 {
		sb.WriteString(fmt.Sprintf("- **Overall Coverage**: %.1f%% (%d/%d operations)\n\n",
			float64(captured)/float64(total)*100, captured, total))
	}

	// Cameras
	if len(registry.Cameras) > 0 {
		sb.WriteString("## Registered Cameras\n\n")
		sb.WriteString("| Manufacturer | Model | Firmware | Operations | Capabilities |\n")
		sb.WriteString("|--------------|-------|----------|------------|---------------|\n")

		for _, cam := range registry.Cameras {
			caps := strings.Join(cam.Capabilities, ", ")
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
				cam.Manufacturer, cam.Model, cam.Firmware, cam.OperationsCaptured, caps))
		}
		sb.WriteString("\n")
	}

	// Coverage by service
	if len(registry.Coverage) > 0 {
		sb.WriteString("## Coverage by Service\n\n")
		sb.WriteString("| Service | Total | Captured | Coverage |\n")
		sb.WriteString("|---------|-------|----------|----------|\n")

		services := []string{"Device", "Media", "PTZ", "Imaging", "Event", "DeviceIO"}
		for _, service := range services {
			if cov, ok := registry.Coverage[service]; ok {
				pct := 0.0
				if cov.Total > 0 {
					pct = float64(cov.Captured) / float64(cov.Total) * 100
				}
				sb.WriteString(fmt.Sprintf("| %s | %d | %d | %.1f%% |\n",
					service, cov.Total, cov.Captured, pct))
			}
		}
		sb.WriteString("\n")
	}

	// Missing operations
	sb.WriteString("## Operation Specifications\n\n")
	opCounts := onviftesting.GetOperationCount()
	sb.WriteString(fmt.Sprintf("- Device: %d operations defined\n", opCounts.Device))
	sb.WriteString(fmt.Sprintf("- Media: %d operations defined\n", opCounts.Media))
	sb.WriteString(fmt.Sprintf("- PTZ: %d operations defined\n", opCounts.PTZ))
	sb.WriteString(fmt.Sprintf("- Imaging: %d operations defined\n", opCounts.Imaging))
	sb.WriteString(fmt.Sprintf("- Event: %d operations defined\n", opCounts.Event))
	sb.WriteString(fmt.Sprintf("- DeviceIO: %d operations defined\n", opCounts.DeviceIO))
	sb.WriteString(fmt.Sprintf("\n**Total**: %d read-only operations tracked\n", opCounts.Total))

	return sb.String()
}
