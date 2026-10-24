# CLI: discover

`discover` answers one question in one command: *which ONVIF devices are
on this network right now?* It sends a multicast WS-Discovery Probe and
prints every responder's endpoint reference, XAddrs, and scopes — the
script-friendly starting point before any credentials are known.

## Install

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
# or grab the prebuilt binary from the release assets (six platforms)
```

## Usage

```bash
discover [-interface eth0] [-timeout 10s]
```

| Flag | Default | Meaning |
|---|---|---|
| `-interface` | all | Network interface to probe out of — by name or address (`eth0`, `en0`, `192.0.2.10`) |
| `-timeout` | `10s` | Overall discovery window (Go duration: `5s`, `30s`, `1m`) |

Output is one block per camera: `Endpoint` (the discovery endpoint
reference), `XAddr` lines (the actual service addresses — a device may
advertise several), and `Scopes` (ONVIF scope URIs carrying
manufacturer, model, and hardware hints). Zero finds exit 0 with
"No cameras found."; transport failures exit 1.

```text
Found 2 camera(s):

Camera 1:
  Endpoint: uuid:1b2e3d47-0000-0000-9cfa-5e0a9c4b2d1f
  XAddr: http://192.168.1.201:80/onvif/device_service
  Scopes:
    - onvif://www.onvif.org/Manufacturer/Hikvision
    - onvif://www.onvif.org/Model/DS-2CD2143G2-I
```

## Scenarios

**Multi-NIC hosts.** On a machine with VPN + LAN + docker bridges, a
probe out of the wrong interface finds nothing or gets answered by
virtual networks. Pin the physical NIC:

```bash
discover -interface en0 -timeout 5s
```

**Inventory sweeps.** The fixed-format output makes the tool a one-liner
front end for inventory scripts (parse the `XAddr:` lines), or pair it
with `onvif quick` / `onvif diagnostics` once credentials exist.

**Slow networks.** Cameras boot slowly after power cycles; extend the
window instead of looping the tool — responders keep arriving during the
whole timeout:

```bash
discover -timeout 30s
```

## Limits and next steps

Multicast reaches the local subnet only; for cameras across routed
subnets use directed probing from a host that shares their LAN — see
[discovery.md](onvif-go-discovery.md) for the `ProbeEndpoint` /
`ProbeSerial` API, and `discovery.FilterONVIFDevices` for dropping the
Synology/Windows/printer ghost responders that also answer probes.
