# CLI: onvif diagnostics

`onvif diagnostics` runs the library's full read surface against one
camera — eleven numbered operations, per-operation success/error and
response time — and writes a JSON report. With `-capture-xml` it also
records every raw SOAP exchange into a `tar.gz` archive designed to
become a regression fixture. This is **the tool to run before filing an
issue**: the JSON plus the archive carry everything a maintainer needs.

## Install

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
```

## Usage

```bash
onvif diagnostics -endpoint http://192.168.1.201/onvif/device_service \
    -username service -password '***' [-output ./camera-logs] \
    [-timeout 30] [-verbose] [-capture-xml] [-capture-all]
```

| Flag | Default | Meaning |
|---|---|---|
| `-endpoint` | — (required) | Device service endpoint URL |
| `-username` / `-password` | — (required) | Camera credentials |
| `-output` | `./camera-logs` | Output directory |
| `-timeout` | `30` | Per-request timeout, seconds |
| `-verbose` | off | Progress detail per operation |
| `-capture-xml` | off | Record raw SOAP request/response pairs → `tar.gz` |
| `-capture-all` | off | Comprehensive mode: every read operation (implies `-capture-xml`) |

## What it runs

The standard pass executes, in order: `GetDeviceInformation`,
`GetSystemDateAndTime`, `GetCapabilities`, service-endpoint discovery
(`Initialize`), `GetProfiles`, then per profile `GetStreamUri`,
`GetSnapshotUri`, `GetVideoEncoderConfigurations`,
`GetImagingSettings`, `GetStatus` (PTZ), and `GetPresets` (PTZ). Each
result lands in the report with `success`, `data`, `error`, and
`response_time`; failures also append to a global `errors` log.

Outputs, named from the device's own identity when available:

```text
camera-logs/
├── Hikvision_DS-2CD2143G2-I_5.7.1_diag_20260920-1030.json
└── Hikvision_DS-2CD2143G2-I_5.7.1_xmlcapture_20260920-1030.tar.gz
```

The archive contains one file per SOAP exchange plus a `metadata.json`
(V2 format) describing the device and capture context.

## Scenarios

**Filing an issue.** Run the standard pass with `-capture-xml`, redact
credentials from the JSON, and attach both files. The report pinpoints
which operation failed and how (fault code, HTTP status, timeout), and
the archive lets maintainers replay your camera's exact wire behavior.

**Becoming a regression fixture.** Captured archives feed
`generate-tests`, which converts them into Go tests against recorded
responses and registers the camera in `testdata/captures/registry.json`
— this is how real-firmware quirks become permanent regression coverage
(see [testing.md](onvif-go-testing.md)).

**Firmware comparison.** Run the same pass before and after a camera
firmware upgrade; diff the two JSON reports to see exactly which
operations changed response shape or timing.

**Slow or flaky cameras.** Raise `-timeout` (some embedded firmware
takes tens of seconds on `GetCapabilities`); `-verbose` shows each
operation as it happens instead of only the summary.

## Privacy notes

The JSON contains your endpoint and username; the XML archive contains
full SOAP bodies — including whatever the camera echoes. Redact
credentials and internal addresses before sharing (the hub's desensitization
rules apply to anything pasted into issues).
