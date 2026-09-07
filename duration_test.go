package onvif

import (
	"errors"
	"testing"
	"time"
)

func TestParseXSDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		// The forms this library's own formatDuration emits.
		{input: "PT60S", want: 60 * time.Second},
		{input: "PT1M", want: time.Minute},
		{input: "PT1M30S", want: 90 * time.Second},

		// Forms a camera is free to send that formatDuration never produces,
		// and that server/event.go's deliberately narrow parser rejects.
		{input: "PT0H1M0S", want: time.Minute},
		{input: "PT2H", want: 2 * time.Hour},
		{input: "P1D", want: 24 * time.Hour},
		{input: "P1DT2H3M4S", want: 24*time.Hour + 2*time.Hour + 3*time.Minute + 4*time.Second},
		{input: "PT0S", want: 0},

		// Fractional seconds are permitted by xs:duration.
		{input: "PT1.5S", want: 1500 * time.Millisecond},

		// A negative duration is legal in xs:duration.
		{input: "-PT30S", want: -30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseXSDuration(tt.input)
			if err != nil {
				t.Fatalf("parseXSDuration(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseXSDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseXSDurationRejects(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "no components", input: "P"},
		{name: "empty time part", input: "PT"},
		{name: "not a duration", input: "garbage"},
		{name: "missing P", input: "T30S"},
		{name: "trailing junk", input: "PT30Sx"},
		{name: "unit out of order", input: "PT30S1M"},

		// Years and months have no fixed length, so they are refused rather
		// than approximated against an invented calendar.
		{name: "years", input: "P1Y"},
		{name: "months", input: "P1M"},
		{name: "years and months", input: "P1Y6M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseXSDuration(tt.input)
			if !errors.Is(err, ErrInvalidXSDuration) {
				t.Errorf("parseXSDuration(%q) error = %v, want ErrInvalidXSDuration", tt.input, err)
			}
			if got != 0 {
				t.Errorf("parseXSDuration(%q) = %v, want 0 alongside the error", tt.input, got)
			}
		})
	}
}

// TestParseXSDurationMinuteVersusMonth pins the one ambiguity in the grammar
// that a hand-written parser is most likely to get wrong: "M" means months
// before the T separator and minutes after it.
func TestParseXSDurationMinuteVersusMonth(t *testing.T) {
	if _, err := parseXSDuration("P5M"); !errors.Is(err, ErrInvalidXSDuration) {
		t.Errorf("P5M should be five months and therefore refused, got err = %v", err)
	}

	got, err := parseXSDuration("PT5M")
	if err != nil {
		t.Fatalf("PT5M should be five minutes, got err = %v", err)
	}
	if got != 5*time.Minute {
		t.Errorf("parseXSDuration(\"PT5M\") = %v, want 5m", got)
	}
}

// TestParseXSDurationOrZero covers the response-mapping wrapper: a camera
// sending a malformed timeout must not fail the whole call.
func TestParseXSDurationOrZero(t *testing.T) {
	if got := parseXSDurationOrZero("PT45S"); got != 45*time.Second {
		t.Errorf("parseXSDurationOrZero(\"PT45S\") = %v, want 45s", got)
	}
	if got := parseXSDurationOrZero("nonsense"); got != 0 {
		t.Errorf("parseXSDurationOrZero(\"nonsense\") = %v, want 0", got)
	}
	if got := parseXSDurationOrZero(""); got != 0 {
		t.Errorf("parseXSDurationOrZero(\"\") = %v, want 0", got)
	}
}

// TestParseXSDurationRoundTripsFormatDuration checks the two halves agree, so
// a value written by SetAudioEncoderConfiguration can be read back by
// GetAudioEncoderConfiguration.
func TestParseXSDurationRoundTripsFormatDuration(t *testing.T) {
	for _, want := range []time.Duration{
		1 * time.Second,
		30 * time.Second,
		59 * time.Second,
		time.Minute,
		90 * time.Second,
		10 * time.Minute,
		time.Hour,
	} {
		formatted := formatDuration(want)
		got, err := parseXSDuration(formatted)
		if err != nil {
			t.Errorf("parseXSDuration(formatDuration(%v)) = %q: %v", want, formatted, err)

			continue
		}
		if got != want {
			t.Errorf("round trip of %v through %q gave %v", want, formatted, got)
		}
	}
}
