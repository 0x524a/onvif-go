// Package buildinfo formats the version line printed by the command-line tools.
package buildinfo

import (
	"runtime/debug"
	"strings"
)

const devVersion = "dev"

// String returns "<tool> <version> (<commit>)".
//
// version and commit are normally set with -ldflags "-X main.Version=..." at
// release time. When they are not, as with "go install ...@v1.2.0" or a plain
// "go build", the module version and VCS revision that the Go toolchain embeds
// in the binary are used instead.
func String(tool, version, commit string) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		version, commit = fill(version, commit, info)
	}

	return format(tool, version, commit)
}

func fill(version, commit string, info *debug.BuildInfo) (outVersion, outCommit string) {
	if (version == "" || version == devVersion) && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}

	if commit == "" {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				commit = s.Value
			}
		}
	}

	return version, commit
}

func format(tool, version, commit string) string {
	if version == "" {
		version = devVersion
	}

	const shortLen = 7
	if len(commit) > shortLen {
		commit = commit[:shortLen]
	}

	var b strings.Builder
	b.WriteString(tool + " " + version)

	if commit != "" && commit != "unknown" {
		b.WriteString(" (" + commit + ")")
	}

	return b.String()
}
