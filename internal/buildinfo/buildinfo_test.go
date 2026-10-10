package buildinfo

import (
	"runtime/debug"
	"testing"
)

const release = "v1.2.1"

func TestFormat(t *testing.T) {
	tests := []struct {
		name, version, commit, want string
	}{
		{"release", release, "abcdef1234567", "tool " + release + " (abcdef1)"},
		{"short commit", release, "abc", "tool " + release + " (abc)"},
		{"no commit", release, "", "tool " + release},
		{"unknown commit", release, "unknown", "tool " + release},
		{"empty version", "", "", "tool dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format("tool", tt.version, tt.commit); got != tt.want {
				t.Errorf("format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFill(t *testing.T) {
	info := &debug.BuildInfo{
		Main:     debug.Module{Version: "v1.2.0"},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "deadbeefcafe"}},
	}

	t.Run("dev falls back to module version and revision", func(t *testing.T) {
		v, c := fill("dev", "", info)
		if v != "v1.2.0" || c != "deadbeefcafe" {
			t.Errorf("fill() = %q, %q", v, c)
		}
	})

	t.Run("ldflags win", func(t *testing.T) {
		v, c := fill("v9.9.9", "1234567", info)
		if v != "v9.9.9" || c != "1234567" {
			t.Errorf("fill() = %q, %q", v, c)
		}
	})

	t.Run("a development build keeps dev", func(t *testing.T) {
		v, _ := fill("dev", "", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}})
		if v != "dev" {
			t.Errorf("version = %q, want dev", v)
		}
	})
}

func TestStringUsesGivenValues(t *testing.T) {
	if got := String("tool", "v9.9.9", "abcdef1234"); got != "tool v9.9.9 (abcdef1)" {
		t.Errorf("String() = %q", got)
	}
}
