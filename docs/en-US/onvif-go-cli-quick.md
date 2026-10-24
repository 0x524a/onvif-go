# CLI: onvif quick

`onvif quick` is the interactive five-menu companion for first contact
with a camera: discover it, list NICs, connect and identify it, nudge
the PTZ, and print every stream/snapshot URL. No flags, no config —
run it and answer prompts. It exists to answer "does this library talk
to my camera at all?" in under a minute.

## Install

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0
```

## Usage

```bash
onvif quick
```

The menu loops until you exit:

| Choice | Action |
|---|---|
| 1 — Discover cameras | Multicast probe (optionally pinned to one interface, with a listed picker); prints name + endpoint per camera |
| 2 — List network interfaces | Every NIC with up/down state, multicast capability, and addresses — the pre-flight for choice 1 on multi-NIC hosts |
| 3 — Connect to camera | Prompts IP / username (default `admin`) / password → device information, profile count, first stream URI |
| 4 — PTZ demo | Connects, checks `PTZ GetStatus`, then one canned move: right / left / up / down (continuous for 2 s, then stop) or absolute go-to-center |
| 5 — Get stream URLs | Per profile: RTSP stream URI, snapshot URI, encoding + resolution |

Camera endpoints are built as `http://<ip>/onvif/device_service` — the
conventional ONVIF path most devices expose.

## Scenarios

**Bring-up in front of a new camera.** Choices in order: 1 (is it on
the network?), 3 (do the credentials work and what is it?), 5 (grab the
RTSP URL for a VLC sanity check). Choice 5 ends with exactly the URLs
you need: open the stream in VLC, the snapshot in a browser.

**PTZ wiring check.** Choice 4 proves the whole chain — auth, service
discovery, `GetProfiles`, PTZ `GetStatus` — and physically moves the
camera. If it reports "PTZ not supported", the profile has no PTZ
configuration; try another profile before concluding the camera has no
mechanics.

**Interface triage before discovery.** On VPN-heavy laptops, run choice
2 first: a NIC listed as `Multicast: No` will not carry a probe. Feed
the interface name into choice 1's picker (or `discover -interface`).

## Reading the results

- "Connected!" on choice 3 means digest/none auth succeeded with the
  given credentials; for auth-ladder behavior (password-text, HTTP
  Basic fallbacks) see [authentication.md](onvif-go-authentication.md).
- Stream URLs print verbatim from `GetStreamUri` — some cameras embed
  credentials in the URL, others require them via RTSP; the tool's
  closing tips note both.

## Limits

Everything is one-shot and interactive — for repeatable runs, scripted
collection, or attaching evidence to an issue, use
[onvif diagnostics](onvif-go-cli-diagnostics.md).
