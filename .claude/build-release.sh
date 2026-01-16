#!/bin/bash
# build-release.sh - Build release binaries locally

set -e

VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
echo "Building release binaries for version: $VERSION"

# Clean previous builds
rm -rf bin releases
mkdir -p bin releases

# Platforms to build
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# Binaries to build
BINARIES=(
    "onvif-cli"
    "onvif-quick"
    "onvif-server"
    "onvif-diagnostics"
)

LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

echo "Building binaries..."
for platform in "${PLATFORMS[@]}"; do
    OS="${platform%/*}"
    ARCH="${platform#*/}"
    
    echo ""
    echo "Building for $OS/$ARCH..."
    
    for binary in "${BINARIES[@]}"; do
        OUTPUT="bin/${binary}-${OS}-${ARCH}"
        
        if [ "$OS" = "windows" ]; then
            OUTPUT="${OUTPUT}.exe"
        fi
        
        echo "  - ${binary}"
        GOOS=$OS GOARCH=$ARCH CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o "$OUTPUT" "./cmd/${binary}" 2>/dev/null || {
            echo "    ⚠️  Skipped (build failed)"
            continue
        }
    done
done

echo ""
echo "Creating release archives..."

cd bin

for platform in "${PLATFORMS[@]}"; do
    OS="${platform%/*}"
    ARCH="${platform#*/}"
    ARCHIVE_NAME="onvif-go-${VERSION}-${OS}-${ARCH}"
    
    # Check if any binary exists for this platform
    if [ "$OS" = "windows" ]; then
        FILES=(*-${OS}-${ARCH}.exe)
    else
        FILES=(*-${OS}-${ARCH})
    fi
    
    # Skip if no files found
    if [ "${FILES[0]}" = "*-${OS}-${ARCH}" ] || [ "${FILES[0]}" = "*-${OS}-${ARCH}.exe" ]; then
        continue
    fi
    
    echo "  Creating archive for ${OS}/${ARCH}..."
    
    if [ "$OS" = "windows" ]; then
        # ZIP for Windows
        zip -q "../releases/${ARCHIVE_NAME}.zip" *-${OS}-${ARCH}.exe ../README.md ../LICENSE
    else
        # tar.gz for Unix-like
        tar czf "../releases/${ARCHIVE_NAME}.tar.gz" *-${OS}-${ARCH} -C .. README.md LICENSE
    fi
done

cd ..

echo ""
echo "Generating checksums..."
cd releases
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum * > checksums.txt
else
    shasum -a 256 * > checksums.txt
fi
cd ..

echo ""
echo "✅ Build complete!"
echo ""
echo "Binaries in:  $(pwd)/bin/"
echo "Archives in:  $(pwd)/releases/"
echo ""
ls -lh releases/

echo ""
echo "To create a GitHub release, run:"
echo "  gh release create ${VERSION} releases/* --title \"Release ${VERSION}\" --notes \"Release notes here\""
