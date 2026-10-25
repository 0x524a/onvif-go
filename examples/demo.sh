#!/bin/bash

# onvif-go demo script — a tour of what the library and its CLIs do.
# Nothing here talks to the network; run the tools for real.

echo "🎥 onvif-go — ONVIF client + virtual-device library (v2 line)"
echo "=============================================================="
echo

echo "📁 Project layout:"
echo "├── onvif/          Client facade (auth ladder, service endpoints)"
echo "├── device/ media/ media2/ ptz/ imaging/ events/ analytics/ deviceio/"
echo "│                   Domain service packages (v2 split)"
echo "├── metadata/       Profile M metadata stream parser"
echo "├── discovery/      WS-Discovery: active probe, passive listener, directed HTTP"
echo "├── server/         Virtual ONVIF camera (HTTP or TLS)"
echo "└── cmd/            CLI tools (also published prebuilt on every release)"
echo

echo "🔧 CLI tools (pick per job):"
echo "   discover          Scan the LAN for ONVIF cameras"
echo "   onvif-quick       Interactive sanity check of one camera (connect/streams/PTZ demo)"
echo "   onvif-diagnostics 11-operation sweep + JSON report — run before filing an issue"
echo "   onvif-server      Stand up a virtual multi-lens camera for testing"
echo "   generate-tests    (dev) turn captured SOAP into regression fixtures"
echo

echo "🚀 Typical commands:"
echo "   make build        # go build ./... (fastest compile check)"
echo "   make test         # go test -race ./..."
echo "   make lint         # golangci-lint run"
echo "   make cross        # CGO_ENABLED=0 linux/arm64 binaries into build/"
echo "   go run ./cmd/discover -timeout 5s"
echo "   go run ./cmd/onvif-server -password change-me -profiles 5"
echo

echo "📖 Library usage (v2 facade API):"
cat << 'EOF'
```go
client, _ := onvif.NewClient(
    "http://192.0.2.100/onvif/device_service",
    onvif.WithCredentials("admin", "password"),
    onvif.WithTimeout(30*time.Second),
)
ctx := context.Background()

info, _ := client.Device().GetDeviceInformation(ctx)
profiles, _ := client.Media().GetProfiles(ctx)
uri, _ := client.Media().GetStreamURI(ctx, profiles[0].Token)

// v2.1.0 surfaces:
opts, _ := client.Media2().GetVideoEncoderConfigurationOptions(ctx, "", "") // H.265/AV1 options
rules, _ := client.Analytics().GetRules(ctx, "VideoAnalytics_1")            // Profile M
stream, _ := metadata.Parse(rtpPayload)                                      // metadata stream
```
EOF

echo
echo "🌟 Highlights:"
echo "✅ Wire contracts grounded in the official ONVIF WSDL/XSD set (onvif/specs)"
echo "✅ Authentication ladder that survives weird firmware (digest, clock skew, retries)"
echo "✅ Media2 client — codec-agnostic configuration model (H.264/H.265/AV1)"
echo "✅ Profile M: analytics client + metadata stream parser + events pull-point"
echo "✅ Virtual camera server, HTTP or TLS, with PTZ presets and 10 lens templates"
echo "✅ Namespace-strict contract suite + client↔simulator conformance loopback + fuzz"
echo "✅ Runnable examples per feature area under examples/"
echo
echo "📚 Manuals: https://github.com/mickeyzzc/onvif-go/tree/main/docs (bilingual: en-US / zh-CN)"
echo "   Releases carry prebuilt CLIs for 6 platforms: https://github.com/mickeyzzc/onvif-go/releases"
