# CLI: onvif server

`onvif server` is the `server/` package as a runnable binary: a virtual
multi-lens ONVIF camera. Start it and your recorder, NVR, or ONVIF
client sees 1–10 configurable camera profiles — with PTZ nodes, imaging,
and opt-in pull-point events — no hardware required. Add it by address:
the binary serves SOAP over HTTP and does not answer WS-Discovery
multicast probes. It is the same simulator the library's own conformance
loopback tests drive.

## Install

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
```

## Usage

```bash
onvif server -password 'test-pass' [-port 8080] [-profiles 3] \
    [-manufacturer onvif-go] [-model 'Virtual Multi-Lens Camera'] \
    [-ptz=true] [-imaging=true] [-events=false] [-info] [-version]
```

| Flag | Default | Meaning |
|---|---|---|
| `-host` / `-port` | `0.0.0.0` / `8080` | Listen address |
| `-username` / `-password` | `admin` / **required** | Device credentials |
| `-manufacturer` `-model` `-firmware` `-serial` | `onvif-go` / `Virtual Multi-Lens Camera` / `1.0.0` / `SN-12345678` | Advertised identity |
| `-profiles` | `3` | Number of profiles, 1–10 |
| `-ptz` `-imaging` `-events` | on / on / off | Service toggles |
| `-info` | — | Print the resolved configuration and exit |
| `-version` | — | Print version and exit |

A password is mandatory — flag or `ONVIF_SERVER_PASSWORD` env var. The
simulator refuses to start with empty credentials because the empty
credentials mode is the documented everything-open mode of the embedded
package, inappropriate for a demo binary on real networks.

Profiles are generated from ten rotating templates (main 1080p, wide
720p, telephoto, low-light, 4K UHD, compact VGA, PTZ dome, fisheye,
thermal, license-plate 60 fps), each with an H264 encoder and snapshot
configuration; PTZ-capable templates get pan/tilt/zoom nodes with two
presets. The service tree mounts under `/onvif` on the given port.
`SIGINT`/`SIGTERM` shut down gracefully.

## Scenarios

**Recorder onboarding without hardware.** Develop and demo a recorder's
add-camera-by-address flow — the endpoint is
`http://<host>:<port>/onvif/device_service`, `GetDeviceInformation`
returns the identity you set, profiles and stream URIs answer — before
any camera is on the desk:

```bash
onvif server -password demo -manufacturer TestCam -model VC-1 -serial SIM-0001
```

**A multi-camera rack from one process.** `-profiles 10` yields ten
distinct virtual cameras (the template rotation) — one device address,
`GetProfiles` shows the rack; ideal for UI layout and
main/sub-stream selection testing.

**PTZ UI testing.** With `-ptz`, PTZ services answer `GetStatus`,
continuous/absolute moves, and presets — enough to exercise a joystick
widget or preset-tour logic against something that never wears out.

**Events integration.** `-events` enables the pull-point service; point
your subscription code at the simulator to develop MotionAlarm handling
offline (see [events.md](onvif-go-events.md) for the managed client API).

**CI soak target.** The binary is a single process with no state on
disk; recorders under test can add/remove it freely. Run it in a
container next to your integration suite for a deterministic ONVIF
peer.

## TLS and embedding

The runnable binary speaks plain HTTP. For HTTPS use the embedded
package's `Config.TLSCertFile`/`TLSKeyFile` (TLS engages inside
`Start`) — see [server.md](onvif-go-server.md), which also covers
pluggable state providers, per-action auth policy, and the
`PublishEvent` seam that the binary does not expose.
