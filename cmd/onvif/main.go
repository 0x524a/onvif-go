// Command onvif is the all-in-one CLI for the onvif-go library — one
// binary, four subcommands. It replaces the four separately-shipped
// tools (discover, onvif-quick, onvif-diagnostics, onvif-server):
// `onvif --help` lists everything instead of users having to know four
// binary names, and a release attaches one file per platform instead
// of four.
//
// Usage:
//
//	onvif discover [-interface en0] [-timeout 10s]
//	onvif quick
//	onvif diagnostics -endpoint http://…/device_service -username u -password p
//	onvif server -username u -password p
package main

import (
	"fmt"
	"os"
)

var usageText = `onvif — CLI for onvif-go

Subcommands:

  discover      Scan the LAN for ONVIF cameras (WS-Discovery).
  quick         Interactive sanity-check of one camera: connect, PTZ
                demo, stream URLs.
  diagnostics   Sweep all major operations into a JSON report — the
                tool to run before filing an issue.
  server        Stand up a virtual multi-lens camera (no hardware
                needed).

Run "onvif <subcommand> -help" for its flags.
`

func main() {
	os.Exit(dispatch(os.Args[1:]))
}

// dispatch picks the subcommand; 0 success, 2 usage error. Split from
// main so the table is testable.
func dispatch(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usageText)
		return 2
	}
	sub, rest := args[0], args[1:]
	var fn func([]string)
	switch sub {
	case "discover":
		fn = runDiscover
	case "quick":
		fn = runQuick
	case "diagnostics":
		fn = runDiagnostics
	case "server":
		fn = runServer
	case "help", "-h", "-help", "--help":
		fmt.Print(usageText)
		return 0
	case "version", "-version", "--version":
		fmt.Printf("onvif %s\n", version)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "onvif: unknown subcommand %q\n\n", sub)
		fmt.Fprint(os.Stderr, usageText)
		return 2
	}
	fn(rest)
	return 0
}
