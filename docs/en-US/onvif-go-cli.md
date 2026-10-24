# CLI Tools

One CLI, four subcommands — `onvif discover`, `onvif quick`, `onvif
diagnostics`, and `onvif server` (unified into a single binary at
v2.6.0; previously four separate tools). Static, zero-dependency, one
prebuilt file per platform on every release.

## Which tool for which job?

| You want to… | Run |
|---|---|
| See every ONVIF camera on the LAN (script-friendly output) | `discover` |
| Poke at one camera interactively — connect, PTZ nudge, stream URLs | `onvif quick` |
| Collect a full per-operation report of one camera (file an issue, capture fixtures) | `onvif diagnostics` |
| Test a recorder/NVR against virtual cameras without hardware | `onvif server` |
| Turn captured SOAP exchanges into regression tests (developer) | `generate-tests` |

Per-tool manuals: [discover](onvif-go-cli-discover.md) ·
[onvif quick](onvif-go-cli-quick.md) ·
[onvif diagnostics](onvif-go-cli-diagnostics.md) ·
[onvif server](onvif-go-cli-server.md)

A typical first contact with an unknown camera runs the first three in
sequence: `discover` finds it, `onvif quick` proves the library talks to
it, `onvif diagnostics -capture-xml` records everything else you need.

## Install

Since **v2.1.0** every [release](https://github.com/mickeyzzc/onvif-go/releases)
attaches prebuilt binaries for linux/amd64, linux/arm64, linux/arm,
darwin/amd64, darwin/arm64, and windows/amd64, with `SHA256SUMS` —
download and run, no Go toolchain needed:

```bash
gh release download v2.6.0 -R mickeyzzc/onvif-go -p 'onvif_linux_amd64'
```

Or install from source at a pinned version:

```bash
go install github.com/mickeyzzc/onvif-go/v2/cmd/onvif@v2.6.0   # discover/quick/diagnostics/server 全在一个二进制
```

Or build from a checkout (`make build` compiles everything including
`generate-tests`, which is a developer tool and not part of release
artifacts).

## Shared conventions

- Exit code 0 = success; nonzero with a message on stderr = failure.
- Camera credentials are never read from a config file — flags (or the
  `ONVIF_SERVER_PASSWORD` env var for `onvif server`) only, so secrets
  stay out of dotfiles.
- What you see is what the library does: the tools are thin shells over
  the public API, with no tool-side special cases — behavior you observe
  reproduces in your own integration.

For the discovery mechanics behind `discover` and `onvif quick`, see
[discovery.md](onvif-go-discovery.md); for embedding the simulator
instead of running the binary, [server.md](onvif-go-server.md).
