# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

onvif-go is a production-ready Go library for communicating with ONVIF-compliant IP cameras. It provides both a client library for camera control and a server implementation for camera simulation/testing.

**Key Features:**
- ONVIF client with 200+ APIs across Device, Media, PTZ, and Imaging services
- ONVIF server for virtual camera simulation
- WS-Discovery for network camera detection
- WS-Security authentication with digest passwords
- Multiple CLI tools for camera interaction and diagnostics

## Essential Commands

### Build
```bash
# Build all CLI tools for current platform
make build

# Build for multiple platforms (Linux, Windows, macOS)
make build-all

# Build specific CLI tool
go build -o bin/onvif-cli ./cmd/onvif-cli
```

### Test
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
make test-coverage

# Run benchmarks
make bench
go test -bench=. -benchmem ./...

# Run specific package tests
go test -v ./discovery
go test -v ./server
```

### Lint and Format
```bash
# Run all checks (fmt, vet, lint)
make check

# Format code
make fmt

# Run linter
make lint  # Requires golangci-lint
```

### Development
```bash
# Install dependencies
make deps

# Clean build artifacts
make clean

# Build examples
make examples

# Run CLI tools
./bin/onvif-cli
./bin/onvif-quick
```

### CLI Tools

**onvif-cli**: Comprehensive ONVIF client with interactive and non-interactive modes
```bash
# Interactive menu
./bin/onvif-cli

# Discover cameras
./bin/onvif-cli discover -interface eth0 -timeout 5

# Get device info
./bin/onvif-cli -op info -endpoint http://camera-ip/onvif/device_service -username admin -password pass
```

**onvif-diagnostics**: Camera testing and XML capture for debugging
```bash
./bin/onvif-diagnostics -endpoint http://camera-ip/onvif/device_service -username admin -password pass -verbose

# Capture raw SOAP XML
./bin/onvif-diagnostics ... -capture-xml
```

**onvif-server**: Virtual camera server for testing
```bash
./bin/onvif-server -profiles 5 -username admin -password mypass -port 9000
```

## Architecture

### Package Structure

```
onvif-go/
├── *.go                    # Core client library (client.go, device.go, media.go, ptz.go, imaging.go, etc.)
├── types.go                # ONVIF type definitions (all SOAP XML structures)
├── internal/soap/          # SOAP client with WS-Security (NOT exported)
├── discovery/              # WS-Discovery implementation (exported package)
├── server/                 # ONVIF server implementation (exported package)
├── cmd/                    # CLI tools
│   ├── onvif-cli/         # Full-featured client
│   ├── onvif-quick/       # Lightweight tool
│   ├── onvif-diagnostics/ # Debugging and XML capture
│   ├── onvif-server/      # Server CLI
│   └── generate-tests/    # Test generation from XML captures
├── testing/               # Test utilities (mock_server.go)
├── testdata/captures/     # Real camera SOAP response captures
└── examples/              # Usage examples
```

### Key Components

**Client Layer** (`client.go`):
- Main `Client` struct with HTTP connection pooling
- Functional options pattern for configuration (WithCredentials, WithTimeout, WithHTTPClient)
- Context-aware operations throughout
- Thread-safe credential management with sync.RWMutex

**Service Implementations**:
- `device.go` + `device_*.go`: 98 Device Management APIs (configuration, users, network, certificates, WiFi, storage)
- `media.go`: Media profiles, stream URIs (RTSP/HTTP), snapshots, encoder configuration
- `ptz.go`: PTZ control (continuous, absolute, relative movement, presets)
- `imaging.go`: Image settings (brightness, contrast, exposure, focus, white balance)
- `event.go`: Event service (subscriptions, pull-point)
- `deviceio.go`: Device I/O and relay control

**SOAP Layer** (`internal/soap/`):
- WS-Security UsernameToken authentication with password digest (SHA-1)
- XML marshaling/unmarshaling for ONVIF SOAP messages
- Error handling with ONVIFError type
- NOT exported - internal implementation detail

**Discovery** (`discovery/`):
- WS-Discovery multicast probe on 239.255.255.250:3702
- Network interface selection support
- Device deduplication by endpoint reference

**Server** (`server/`):
- Virtual multi-lens camera simulator
- Implements Device, Media, PTZ, and Imaging services
- Configurable number of camera profiles (up to 10)
- WS-Security authentication support

### Type System

All ONVIF types are defined in `types.go` (~30,000+ lines). Key patterns:

- XML struct tags for SOAP serialization
- Pointer fields for optional values (ONVIF convention)
- Namespace-aware XML marshaling
- Comprehensive coverage of ONVIF Core, Device, Media, PTZ, Imaging specs

## Development Patterns

### Client Usage Pattern
```go
// 1. Create client with options
client, err := onvif.NewClient(
    endpoint,
    onvif.WithCredentials(username, password),
    onvif.WithTimeout(30*time.Second),
)

// 2. Initialize to discover service endpoints
if err := client.Initialize(ctx); err != nil {
    return err
}

// 3. Use service methods
profiles, err := client.GetProfiles(ctx)
```

### Context Usage
All network operations require `context.Context` as first parameter:
- Enables timeouts: `context.WithTimeout()`
- Enables cancellation: `context.WithCancel()`
- No blocking indefinitely

### Error Handling
- Sentinel errors: `ErrServiceNotSupported`, `ErrAuthenticationFailed`
- Typed errors: `ONVIFError` for SOAP faults
- Use `errors.Is()` and `errors.As()` for error checking
- Always wrap errors with context: `fmt.Errorf("operation failed: %w", err)`

### Testing Strategy
- Unit tests alongside implementation files (`*_test.go`)
- Real camera tests in `*_real_camera_test.go` (skipped without `-tags=real_camera`)
- Mock server in `testing/mock_server.go` for integration tests
- XML captures in `testdata/captures/` for regression testing
- Comprehensive test coverage tracked in `docs/testing/`

### Authentication Implementation
WS-Security digest authentication requires:
1. Generate 16-byte random nonce
2. Get UTC timestamp
3. Calculate: `Base64(SHA1(nonce + timestamp + password))`
4. Include Username, Password (digest), Nonce, Created in SOAP header

## Critical Implementation Details

### SOAP Message Structure
All ONVIF operations use SOAP 1.2 over HTTP POST:
- Envelope with WS-Security header (if authenticated)
- Body contains operation-specific request
- Response parsed from SOAP envelope body
- SOAP faults mapped to Go errors

### Service Endpoint Discovery
The `Initialize()` method discovers service endpoints:
1. Calls `GetCapabilities()` to get service URLs
2. Caches endpoints (media, PTZ, imaging, event)
3. Falls back to device service endpoint if not found
4. Subsequent operations use cached endpoints

### Connection Pooling
HTTP client configured for optimal performance:
- Idle connection timeout: 90s
- Max idle connections: 10
- Max idle per host: 5
- Custom transport for TLS control

### Network Interface Selection (Discovery)
Discovery supports binding to specific interfaces:
- By interface name: `"eth0"`, `"en0"`
- By IP address: `"192.168.1.100"`
- Auto-detection tries all active interfaces if not specified
- Uses `golang.org/x/net/ipv4` for multicast control

## File Organization

- **Root `*.go`**: Public API and implementation
- **`*_test.go`**: Unit tests (run with `go test`)
- **`*_real_camera_test.go`**: Integration tests requiring real cameras
- **`docs/`**: Comprehensive documentation organized by category
- **`test-reports/`**: JSON reports from real camera testing
- **`examples/`**: Standalone example programs

## Build System

**Makefile targets**:
- `make all`: deps + check + test + build
- `make build`: Build CLI tools for current platform
- `make build-all`: Cross-compile for all platforms (Linux, Windows, macOS - amd64, arm64, arm)
- `make release`: Build + create archives + checksums
- `make test`: Run tests with race detection
- `make bench`: Run benchmarks
- `make check`: fmt + vet + lint
- `make clean`: Remove build artifacts

**Build flags**:
- `CGO_ENABLED=0`: Static binaries
- `-ldflags="-s -w"`: Strip symbols for smaller size
- Version injection: `-X main.Version=$(VERSION)`

## Testing Without Real Cameras

Use the diagnostic tool to capture real camera responses:
```bash
# 1. Capture XML from real camera
./onvif-diagnostics -endpoint http://camera/onvif/device_service -username user -password pass -capture-xml

# 2. Generate test from capture
./generate-tests -capture camera-logs/*_xmlcapture_*.tar.gz -output testdata/captures/

# 3. Run generated tests
go test -v ./testdata/captures/
```

This allows testing library changes against real camera behavior without physical hardware.

## Important Notes

- **ONVIF specification compliance**: Follows ONVIF Core, Device, Media, PTZ, Imaging specs
- **WS-Security**: Digest authentication (SHA-1) per ONVIF requirements
- **Concurrency**: All operations are thread-safe
- **XML namespaces**: Critical for ONVIF - handled in types.go struct tags
- **Pointer semantics**: Optional fields use pointers (ONVIF convention)
- **Service support detection**: Always check capabilities before calling service-specific methods
- **Endpoint flexibility**: Accepts full URLs, IP:port, or bare IPs (auto-adds http:// and /onvif/device_service)

## Common Development Tasks

**Adding a new ONVIF operation**:
1. Define request/response types in `types.go` with XML tags
2. Implement method in appropriate service file (`device.go`, `media.go`, etc.)
3. Use `callMethod()` helper for SOAP invocation
4. Add unit test in corresponding `*_test.go`
5. Update documentation in `docs/api/`

**Adding a new CLI command**:
1. Add command/flags in `cmd/onvif-cli/main.go`
2. Implement handler function
3. Update CLI help text
4. Add example to `docs/CLI_*.md`

**Adding server functionality**:
1. Implement handler in `server/*.go`
2. Register handler in SOAP router
3. Add test in `server/*_test.go`
4. Update `server/README.md`

## Dependencies

Minimal dependencies (see `go.mod`):
- `golang.org/x/net`: HTTP/2 and IDNA support
- `github.com/0x524A/rtspeek`: RTSP stream validation (diagnostics tool)
- Standard library for everything else

Go version: 1.21+ (currently 1.24)
