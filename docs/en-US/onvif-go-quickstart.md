# Quick start: discover and pull a stream

`onvif-go/v2` is an ONVIF client (with a server side too). This page is
the minimal compilable example: connect to a camera, list profiles, and
pick the main / sub streams with the built-in heuristics.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

> v2 is stable (v2.0.0 shipped 2026-09-17; the current line is v2.2.0) —
> still pin the exact version in production.

## Minimal example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mickeyzzc/onvif-go/v2/onvif"
)

func main() {
	client, err := onvif.NewClient("192.0.2.10")
	if err != nil {
		log.Fatal(err)
	}
	profiles, err := client.Media().GetProfiles(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	mainToken := onvif.SelectMainProfile(profiles)
	subToken := onvif.SelectSubProfile(profiles, mainToken)
	fmt.Println("main profile:", mainToken, "sub profile:", subToken)
}
```

Notes:

- `NewClient` accepts an IP or URL; add auth via options like
  `WithCredentials` as needed ([authentication](onvif-go-authentication.md)).
- Stream URIs come from `client.Media().GetStreamURI`, parameterizable via
  `StreamSetup` for RTSP / HTTP transports ([media](onvif-go-media.md)).
- Event subscriptions are one call:
  `client.Events().SubscribeEvents` ([events](onvif-go-events.md)).
- H.265/AV1 configuration lives in the Media2 model:
  `client.Media2().GetVideoEncoderConfigurationOptions` reports one entry
  per supported codec with a free-name `Encoding`
  ([media](onvif-go-media.md#media2-h265av1)).
- Profile M surfaces: `client.Analytics()` (rule/analytics-module
  configuration) and the `metadata` package (parsing the analytics
  output stream).

## Next steps

- [CLI tools](onvif-go-cli.md): probe a camera with `discover` /
  `onvif quick` / `onvif diagnostics` before writing any code.
- [Discovery](onvif-go-discovery.md): WS-Discovery probes for cameras on
  the network.
- [Media](onvif-go-media.md): profile-selection heuristics and stream URIs.
- [Architecture](onvif-go-architecture.md) /
  [v2 architecture](onvif-go-v2-architecture.md): layering and migration.
