<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/brand/logo-dark.png">
    <img src="assets/brand/logo-light.png" alt="onvif-go" width="360">
  </picture>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/0x524a/onvif-go"><img src="https://pkg.go.dev/badge/github.com/0x524a/onvif-go.svg" alt="Go Reference"></a>
  <a href="https://github.com/0x524a/onvif-go/releases/latest"><img src="https://img.shields.io/github/v/release/0x524a/onvif-go" alt="Release"></a>
  <a href="https://github.com/0x524a/onvif-go/actions/workflows/ci.yml"><img src="https://github.com/0x524a/onvif-go/actions/workflows/ci.yml/badge.svg?branch=master" alt="CI"></a>
  <a href="https://codecov.io/gh/0x524a/onvif-go"><img src="https://codecov.io/gh/0x524a/onvif-go/branch/master/graph/badge.svg" alt="codecov"></a>
  <a href="https://sonarcloud.io/summary/new_code?id=0x524a_onvif-go"><img src="https://sonarcloud.io/api/project_badges/measure?project=0x524a_onvif-go&metric=alert_status" alt="Quality Gate Status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/0x524a/onvif-go" alt="License"></a>
  <a href="https://github.com/0x524a/onvif-go/stargazers"><img src="https://img.shields.io/github/stars/0x524a/onvif-go" alt="GitHub stars"></a>
  <a href="https://github.com/0x524a/onvif-go/issues"><img src="https://img.shields.io/github/issues/0x524a/onvif-go" alt="GitHub issues"></a>
  <br>
  <img src="https://img.shields.io/github/repo-size/0x524a/onvif-go" alt="GitHub repo size">
  <img src="https://img.shields.io/github/languages/code-size/0x524a/onvif-go" alt="GitHub code size">
  <img src="https://img.shields.io/github/go-mod/go-version/0x524a/onvif-go" alt="GitHub go.mod Go version">
  <img src="https://img.shields.io/github/last-commit/0x524a/onvif-go" alt="GitHub last commit">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/0x524a/onvif-go">Go docs</a> ·
  <a href="https://0x524a.github.io/onvif-go/">Project site</a> ·
  <a href="https://0x524a.github.io/onvif-go/api.html">API reference</a> ·
  <a href="CHANGELOG.md">Changelog</a>
</p>

> **Modern, high-performance Go library for ONVIF IP camera integration** - Control surveillance cameras, NVRs, and video devices with comprehensive ONVIF Profile S/T/G support. Includes both client and server implementations for complete ONVIF camera simulation and testing.

---

onvif-go is a Go library for IP cameras, NVRs and other devices that speak [ONVIF](https://www.onvif.org). Use it to find cameras on the network, read their RTSP streams and snapshots, move PTZ cameras, change image settings and manage the device. It also includes a virtual camera server, so you can test code that depends on a camera without owning one.

- **Battle-tested:** verified with multiple ONVIF-compliant cameras, including Hikvision, Dahua, Axis, Bosch and Reolink.
- **Client:** over 200 typed operations across Device, Media, PTZ, Imaging, Events and Device I/O, with WS-Security digest authentication.
- **Discovery:** find devices with WS-Discovery, or answer probes for your own.
- **Virtual cameras:** a multi-lens camera simulator with stable identities that survive restarts.
- **Command-line tools:** prebuilt binaries for discovery, diagnostics and scripting.
- **Small footprint:** the standard library plus [rtspeek](https://github.com/0x524A/rtspeek) and [google/uuid](https://github.com/google/uuid).

## Install

```bash
go get github.com/0x524a/onvif-go
```

Requires Go 1.25 or newer. Prebuilt command-line tools are attached to every [release](https://github.com/0x524a/onvif-go/releases/latest).

## Quick start

### Find cameras

```go
devices, err := discovery.Discover(ctx, 5*time.Second)
if err != nil {
    log.Fatal(err)
}

for _, d := range devices {
    fmt.Println(d.GetName(), d.GetDeviceEndpoint())
}
```

### Connect and get a stream

The endpoint can be a full URL, `host:port`, or a bare IP address.

```go
client, err := onvif.NewClient(
    "192.168.1.100",
    onvif.WithCredentials("admin", "password"),
    onvif.WithTimeout(30*time.Second),
)
if err != nil {
    log.Fatal(err)
}

// Initialize asks the camera where its media, PTZ and imaging services are.
if err := client.Initialize(ctx); err != nil {
    log.Fatal(err)
}

profiles, err := client.GetProfiles(ctx)
if err != nil {
    log.Fatal(err)
}

uri, err := client.GetStreamURI(ctx, profiles[0].Token)
if err != nil {
    log.Fatal(err)
}

fmt.Println(uri.URI) // rtsp://192.168.1.100:554/...
```

### Move a PTZ camera

```go
velocity := &onvif.PTZSpeed{PanTilt: &onvif.Vector2D{X: 0.5, Y: 0}} // pan right
timeout := "PT2S"                                                     // for two seconds

err := client.ContinuousMove(ctx, token, velocity, &timeout)
err = client.Stop(ctx, token, true, true)

presets, err := client.GetPresets(ctx, token)
err = client.GotoPreset(ctx, token, presets[0].Token, nil)
```

Every method takes a `context.Context` first, so timeouts and cancellation work everywhere. More runnable snippets are in the [Go docs](https://pkg.go.dev/github.com/0x524a/onvif-go#pkg-examples) and the [examples](examples/) directory.

## Test without a camera

The `server` package is a virtual ONVIF camera with Device, Media, PTZ and Imaging services:

```go
cfg := server.DefaultConfig()
cfg.Port = 0                      // let the OS pick a free port
cfg.EndpointSeed = "lobby-camera" // same identity on every start

srv, err := server.New(cfg)
if err != nil {
    log.Fatal(err)
}

go srv.Start(ctx)
```

Or from the shell:

```bash
go install github.com/0x524a/onvif-go/cmd/onvif-server@latest
onvif-server -profiles 3 -username admin -password admin -port 8080
```

To make it discoverable on the network, register `srv.DiscoveryDevice()` with a `discovery.Responder`. See the [server README](server/README.md) for configuration, stable identities and the WS-Discovery responder.

Captured responses from real cameras are replayed in the test suite through a mock server (`testing/captures`), so a fix for one camera is checked against the others. See the [camera testing guide](docs/testing/CAMERA_TESTING_FLOW.md) to add yours.

## What it covers

| Service | Methods | Includes |
|---|---|---|
| Device | 97 | Information, capabilities, network, users, certificates, Wi-Fi and 802.1X, storage, logs, firmware |
| Media | 82 | Profiles, RTSP and HTTP stream URIs, snapshots, video, audio and metadata configuration |
| PTZ | 13 | Continuous, absolute and relative moves, presets, home position, status |
| Imaging | 7 | Brightness, contrast, exposure, focus, white balance, wide dynamic range |
| Events | 12 | Subscriptions and pull points |
| Device I/O | 14 | Relays, digital inputs, video outputs |
| Discovery | | Probe the network or answer probes, by interface or by unicast address |

The full list, with signatures and descriptions, is the [API reference](https://0x524a.github.io/onvif-go/api.html), generated from the source. Device Management status is tracked in [DEVICE_API_STATUS.md](docs/api/DEVICE_API_STATUS.md).

## Command-line tools

| Tool | Use it to |
|---|---|
| `onvif-cli` | Discover cameras and read or change settings, interactively or with one-shot flags for scripts (`-op info`, `-op stream`, `-op discover`). See [non-interactive mode](docs/CLI_NON_INTERACTIVE_MODE.md). |
| `onvif-diagnostics` | Test a real camera and capture its raw SOAP, then turn the capture into a regression test. See [XML debugging](docs/XML_DEBUGGING_SOLUTION.md). |
| `onvif-server` | Run virtual cameras from the shell. |
| `onvif-quick` | Discover, connect and try PTZ in a few keystrokes. |
| `generate-tests` | Generate a replay test from a captured archive. |

```bash
go install github.com/0x524a/onvif-go/cmd/onvif-cli@latest
onvif-cli -endpoint http://camera/onvif/device_service -username admin -password pass -op info
```

On a machine with several network interfaces, pick the one to probe from with `-i eth0`, or in code with `discovery.DiscoverOptions{NetworkInterface: "eth0"}`. See [network interface selection](docs/CLI_NETWORK_INTERFACE_USAGE.md).

## Compatibility

- **Go:** 1.25 or newer; CI also runs the latest release.
- **ONVIF:** Profile S, T and G operations.
- **Cameras:** verified with multiple ONVIF-compliant cameras, including Hikvision, Dahua, Axis, Bosch, Reolink and others. Responses captured from Axis, Bosch and Reolink devices are also replayed in the automated test suite. If yours misbehaves, please open an issue; a capture from it (see the [camera testing guide](docs/testing/CAMERA_TESTING_FLOW.md)) is the most useful thing you can attach.

## Documentation

- [Quick start guide](docs/QUICKSTART.md) and [architecture](docs/ARCHITECTURE.md)
- [All guides](docs/README.md)
- [Changelog](CHANGELOG.md)

## Contributing

Contributions are welcome. For anything larger than a fix, please open an issue first. See [CONTRIBUTING.md](CONTRIBUTING.md) for the workflow, and run `make check` before you push.

## License

MIT, see [LICENSE](LICENSE). Inspired by [use-go/onvif](https://github.com/use-go/onvif); specifications from [ONVIF.org](https://www.onvif.org).
