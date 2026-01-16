# Release v1.0.1

## 🎉 What's New

### ✨ Features

#### Simplified Endpoint API
The `NewClient()` function now accepts multiple endpoint formats for easier camera connection:

```go
// Simple IP address - automatically adds http:// and path
client, _ := onvif.NewClient("192.168.1.100")

// IP with custom port
client, _ := onvif.NewClient("192.168.1.100:8080")

// Full URL (backward compatible)
client, _ := onvif.NewClient("http://192.168.1.100/onvif/device_service")
```

**Benefits:**
- 🎯 More intuitive API - just provide the camera IP
- 🔄 Backward compatible - existing code works unchanged
- 📝 Less boilerplate code required

#### Localhost URL Fix (Camera Firmware Bug Workaround)
Automatic handling of cameras that incorrectly report localhost addresses in their GetCapabilities response.

**Problem Solved:**
Some camera firmwares have bugs where they report `localhost`, `127.0.0.1`, `0.0.0.0`, or `::1` in service endpoint URLs instead of their actual IP address, making services unreachable.

**Solution:**
The library now automatically detects and fixes these addresses:

```go
client, _ := onvif.NewClient("192.168.1.100")
client.Initialize(ctx)
// Service endpoints are automatically corrected:
// http://localhost/onvif/media_service → http://192.168.1.100/onvif/media_service
// http://127.0.0.1:8080/onvif/ptz → http://192.168.1.100:8080/onvif/ptz
```

**Handled Cases:**
- ✅ localhost → actual camera IP
- ✅ 127.0.0.1 → actual camera IP
- ✅ 0.0.0.0 → actual camera IP
- ✅ ::1 (IPv6) → actual camera IP
- ✅ Port numbers preserved
- ✅ HTTPS supported
- ✅ Transparent - no code changes needed

### 🏗️ Project Structure Improvements

#### Internal Package Organization
- Moved `soap/` to `internal/soap/` following Go best practices
- SOAP implementation is now private (not part of public API)
- Allows refactoring without breaking changes
- Cleaner separation of public vs private code

#### Examples Organization
- Moved `test/test-server.go` to `examples/test-server/`
- Better clarity - all examples in one place
- Removed empty `test/` directory
- Consistent project structure

#### Module Path Update
- Updated from `github.com/0x524A/onvif-go` to `github.com/0x524a/onvif-go` (lowercase)
- Consistent with GitHub username conventions
- All imports updated across the codebase

### 📚 Documentation

- ✅ Created comprehensive `docs/PROJECT_STRUCTURE.md`
- ✅ Updated `docs/ARCHITECTURE.md` with new structure
- ✅ Added `docs/SIMPLIFIED_ENDPOINT.md` with endpoint format examples
- ✅ Updated CHANGELOG.md with all changes

### 🧪 Testing

**New Test Coverage:**
- 12 test cases for endpoint normalization
- 10 test cases for localhost URL handling
- Integration tests with mock ONVIF server
- Edge case handling verified

**Current Coverage:**
- Main package: 21.2%
- Discovery: 67.2%
- Internal/SOAP: 81.5%
- Overall: ~56%

## 📦 Installation

### Go Module

```bash
go get github.com/0x524a/onvif-go@v1.0.1
```

### Pre-built Binaries

Download platform-specific binaries from the [Releases page](https://github.com/0x524a/onvif-go/releases/tag/v1.0.1).

**Available platforms:**
- Linux: amd64, arm64, arm/v7
- Windows: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)

**Tools included:**
- `onvif-cli` - Interactive CLI tool
- `onvif-quick` - Quick test utility
- `onvif-server` - Virtual ONVIF camera server
- `onvif-diagnostics` - Network diagnostics tool

#### Linux/macOS Installation

```bash
# Download
wget https://github.com/0x524a/onvif-go/releases/download/v1.0.1/onvif-go-v1.0.1-linux-amd64.tar.gz

# Extract
tar xzf onvif-go-v1.0.1-linux-amd64.tar.gz

# Install
chmod +x onvif-cli-linux-amd64
sudo mv onvif-cli-linux-amd64 /usr/local/bin/onvif-cli
```

#### Windows Installation

1. Download `onvif-go-v1.0.1-windows-amd64.zip`
2. Extract the ZIP file
3. Add the extracted directory to your PATH

### Docker Image

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/0x524a/onvif-go:v1.0.1
docker pull ghcr.io/0x524a/onvif-go:latest

# Run ONVIF server
docker run -p 8080:8080 ghcr.io/0x524a/onvif-go:v1.0.1 onvif-server
```

**Multi-architecture support:**
- linux/amd64
- linux/arm64
- linux/arm/v7

## 🔄 Migration Guide

### From v1.0.0

No breaking changes! All existing code continues to work.

**Optional improvements you can make:**

#### Simplify endpoint format:
```go
// Before (still works)
client, _ := onvif.NewClient(
    "http://192.168.1.100/onvif/device_service",
    onvif.WithCredentials("admin", "password"),
)

// After (simpler)
client, _ := onvif.NewClient(
    "192.168.1.100",
    onvif.WithCredentials("admin", "password"),
)
```

#### Update module path (if using lowercase):
```go
// Old import (still works)
import "github.com/0x524A/onvif-go"

// New import (recommended)
import "github.com/0x524a/onvif-go"
```

## 🐛 Bug Fixes

- Fixed cameras with localhost addresses in GetCapabilities response
- Improved URL parsing for edge cases
- Better error messages for invalid endpoints

## 🔗 Links

- 📖 [Documentation](https://pkg.go.dev/github.com/0x524a/onvif-go)
- 💬 [Discussions](https://github.com/0x524a/onvif-go/discussions)
- 🐛 [Issue Tracker](https://github.com/0x524a/onvif-go/issues)
- 📦 [Go Package](https://pkg.go.dev/github.com/0x524a/onvif-go)
- 🐳 [Docker Hub](https://github.com/0x524a/onvif-go/pkgs/container/onvif-go)

## 📊 Stats

- **28 binaries** across 7 platforms
- **4 command-line tools**
- **56% test coverage**
- **Zero external dependencies** (pure Go standard library)

## 🙏 Contributors

Thank you to all contributors who helped make this release possible!

## 📝 Full Changelog

See [CHANGELOG.md](https://github.com/0x524a/onvif-go/blob/master/CHANGELOG.md) for complete details.

---

**Full Changelog**: https://github.com/0x524a/onvif-go/compare/v1.0.0...v1.0.1
