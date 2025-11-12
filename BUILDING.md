# Building and Releasing go-onvif

This document describes how to build binaries for multiple platforms and create releases.

## Quick Start

### Build for Your Current Platform

```bash
make build-cli
```

This builds all CLI tools for your current OS/architecture in the `bin/` directory.

### Build for All Platforms

```bash
make build-all
```

This creates binaries for:
- **Linux**: amd64, arm64, arm (32-bit)
- **Windows**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)

Binaries are output to `bin/` directory.

### Create Release Archives

```bash
make release
```

This:
1. Builds for all platforms
2. Creates `.tar.gz` archives (Linux/macOS) and `.zip` files (Windows)
3. Generates SHA256 checksums
4. Places everything in `releases/` directory

## Manual Building

### Using the Build Script

```bash
# Build with automatic version detection
./build-release.sh

# Build with specific version
./build-release.sh v1.0.1
```

### Using Go Directly

```bash
# Set platform and architecture
export GOOS=linux
export GOARCH=amd64

# Build a specific tool
go build -o bin/onvif-cli-linux-amd64 ./cmd/onvif-cli
```

## Supported Platforms

| OS      | Architecture | Binary Suffix          | Notes                      |
|---------|-------------|------------------------|----------------------------|
| Linux   | amd64       | `linux-amd64`          | 64-bit Intel/AMD           |
| Linux   | arm64       | `linux-arm64`          | 64-bit ARM (Raspberry Pi 4)|
| Linux   | arm         | `linux-arm`            | 32-bit ARM (Raspberry Pi 3)|
| Windows | amd64       | `windows-amd64.exe`    | 64-bit Windows             |
| Windows | arm64       | `windows-arm64.exe`    | ARM Windows (Surface Pro X)|
| macOS   | amd64       | `darwin-amd64`         | Intel Macs                 |
| macOS   | arm64       | `darwin-arm64`         | Apple Silicon (M1/M2/M3)   |

## CLI Tools

The following binaries are built:

1. **onvif-cli** - Comprehensive ONVIF client with full feature set
2. **onvif-quick** - Quick tool for common operations
3. **onvif-server** - ONVIF mock server for testing
4. **onvif-diagnostics** - Diagnostic and debugging tools

## Automated Releases via GitHub Actions

Releases are automatically created when you push a tag:

```bash
# Create and push a new version tag
git tag -a v1.0.1 -m "Release version 1.0.1"
git push origin v1.0.1
```

The GitHub Actions workflow will:
1. Build binaries for all platforms
2. Create release archives
3. Generate checksums
4. Create a GitHub release with all artifacts
5. Build and push Docker images (multi-arch)

### Release Workflow Features

- ✅ Builds for 7 platform/architecture combinations
- ✅ Creates compressed archives (`.tar.gz` and `.zip`)
- ✅ Generates SHA256 checksums for verification
- ✅ Auto-generates release notes from commits
- ✅ Supports pre-releases (tags with `-rc`, `-beta`, `-alpha`)
- ✅ Builds multi-architecture Docker images
- ✅ Pushes to GitHub Container Registry

## Docker Images

Docker images are automatically built for:
- `linux/amd64`
- `linux/arm64`
- `linux/arm/v7`

Available at:
```
ghcr.io/0x524a/onvif-go:latest
ghcr.io/0x524a/onvif-go:v1.0.0
```

## Manual GitHub Release

If you prefer to create releases manually:

```bash
# Build release archives
make release

# Create GitHub release using gh CLI
gh release create v1.0.1 releases/* \
  --title "Release v1.0.1" \
  --notes "Release notes here"
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- `v1.0.0` - Major release (breaking changes)
- `v1.1.0` - Minor release (new features, backward compatible)
- `v1.1.1` - Patch release (bug fixes)
- `v1.0.0-rc1` - Release candidate
- `v1.0.0-beta1` - Beta release
- `v1.0.0-alpha1` - Alpha release

## Build Flags

The build process uses the following flags:

```bash
-ldflags="-s -w -X main.Version=<version> -X main.Commit=<sha>"
```

- `-s` - Omit symbol table (smaller binary)
- `-w` - Omit DWARF debug info (smaller binary)
- `-X main.Version` - Inject version string
- `-X main.Commit` - Inject git commit SHA

## Size Optimization

Binaries are built with `CGO_ENABLED=0` and stripped flags, resulting in:
- Smaller binary sizes
- No external dependencies
- Portable across systems

Typical sizes:
- onvif-cli: ~10-15 MB
- onvif-quick: ~8-12 MB
- onvif-server: ~10-14 MB

## Troubleshooting

### Build Fails for Specific Platform

Some platforms may not be supported by all dependencies. Check:
```bash
go tool dist list  # List all supported platforms
```

### Large Binary Sizes

Ensure you're using the build flags:
```bash
go build -ldflags="-s -w" -o binary ./cmd/tool
```

### Missing Dependencies

```bash
go mod download
go mod tidy
```

## Distribution

Once built, binaries can be distributed via:

1. **GitHub Releases** (automatic)
2. **Package managers** (homebrew, apt, etc.)
3. **Container registries** (Docker Hub, GHCR)
4. **Direct download** from your server

## Verification

Users can verify downloads using checksums:

```bash
# Download binary and checksum
wget https://github.com/0x524a/onvif-go/releases/download/v1.0.0/onvif-go-v1.0.0-linux-amd64.tar.gz
wget https://github.com/0x524a/onvif-go/releases/download/v1.0.0/checksums.txt

# Verify
sha256sum -c checksums.txt --ignore-missing
```

## Next Steps

After building:
1. Test binaries on target platforms
2. Update CHANGELOG.md with release notes
3. Create GitHub release
4. Announce on relevant channels
5. Update documentation with new features
