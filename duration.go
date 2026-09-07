package onvif

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidXSDuration is returned by parseXSDuration for input that is not a
// duration it can represent as a time.Duration.
var ErrInvalidXSDuration = errors.New("invalid xs:duration")

// xsDurationPattern matches the xs:duration lexical form used by ONVIF for
// timeout-valued elements: an optional sign, then P, then an optional date
// part, then an optional T-prefixed time part.
//
// Deliberately more permissive than server/event.go's isoDurationPattern,
// which covers only the "PT<n>M<n>S" subset this library's own formatDuration
// emits. That narrowness is correct there - the server parses values produced
// by an ONVIF client, and #62 scoped it to what could actually be tested
// against. Here the producer is a real camera's firmware, which is free to use
// any lexical form the schema allows, so "PT60S", "PT1M", "PT0H1M0S" and "P1D"
// all have to be understood.
var xsDurationPattern = regexp.MustCompile(
	`^(-)?P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`,
)

// Submatch indices into xsDurationPattern.
const (
	xsDurSign = 1
	xsDurYear = 2
	xsDurMon  = 3
	xsDurDay  = 4
	xsDurHour = 5
	xsDurMin  = 6
	xsDurSec  = 7
)

const hoursPerDay = 24

// parseXSDuration parses an xs:duration into a time.Duration.
//
// Years and months are rejected rather than approximated: neither has a fixed
// length, so "P1M" cannot be converted without inventing a calendar context
// the caller never supplied. Every ONVIF element this is used for is a
// timeout, where years and months do not occur in practice, so rejecting is
// preferable to silently returning a number that is wrong by up to three days.
func parseXSDuration(s string) (time.Duration, error) {
	m := xsDurationPattern.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidXSDuration, s)
	}

	// "P" and "PT" match the pattern but carry no components, and xs:duration
	// requires at least one. Likewise a "T" with nothing after it.
	if !hasAnyComponent(m) {
		return 0, fmt.Errorf("%w: %q has no components", ErrInvalidXSDuration, s)
	}
	if strings.Contains(s, "T") && m[xsDurHour] == "" && m[xsDurMin] == "" && m[xsDurSec] == "" {
		return 0, fmt.Errorf("%w: %q has an empty time part", ErrInvalidXSDuration, s)
	}

	if m[xsDurYear] != "" || m[xsDurMon] != "" {
		return 0, fmt.Errorf(
			"%w: %q uses years or months, which have no fixed length", ErrInvalidXSDuration, s,
		)
	}

	total, ok := sumDurationParts(m)
	if !ok {
		return 0, fmt.Errorf("%w: %q is out of range", ErrInvalidXSDuration, s)
	}

	if m[xsDurSign] == "-" {
		total = -total
	}

	return total, nil
}

// hasAnyComponent reports whether the match carries at least one numeric
// component.
func hasAnyComponent(m []string) bool {
	for _, i := range []int{xsDurYear, xsDurMon, xsDurDay, xsDurHour, xsDurMin, xsDurSec} {
		if m[i] != "" {
			return true
		}
	}

	return false
}

// sumDurationParts adds up the day, hour, minute and second components,
// reporting false if any of them will not fit.
//
// The regex has already established that every component is a run of digits,
// so the only way a conversion fails here is overflow. That makes a bool the
// honest return: there is no underlying error worth propagating, and the
// caller builds its own ErrInvalidXSDuration message either way.
func sumDurationParts(m []string) (total time.Duration, ok bool) {
	for _, part := range []struct {
		text string
		unit time.Duration
	}{
		{m[xsDurDay], hoursPerDay * time.Hour},
		{m[xsDurHour], time.Hour},
		{m[xsDurMin], time.Minute},
	} {
		if part.text == "" {
			continue
		}
		n, err := strconv.Atoi(part.text)
		if err != nil {
			return 0, false
		}
		total += time.Duration(n) * part.unit
	}

	// Seconds may carry a fractional part, so they go through ParseFloat.
	if m[xsDurSec] != "" {
		secs, err := strconv.ParseFloat(m[xsDurSec], 64)
		if err != nil {
			return 0, false
		}
		total += time.Duration(secs * float64(time.Second))
	}

	return total, true
}

// parseXSDurationOrZero parses an xs:duration, returning zero for input that
// cannot be parsed.
//
// Response mapping uses this rather than parseXSDuration directly. A camera
// reporting a malformed timeout should not fail the whole call - the caller
// asked for a configuration, and every other field in it is still good - and
// Client has no logger to report the discrepancy through. The zero value is
// indistinguishable from an absent element, which is the one wart here; the
// alternative of failing the call would make an unrelated field's formatting
// quirk break an entire operation against otherwise-working firmware.
func parseXSDurationOrZero(s string) time.Duration {
	d, err := parseXSDuration(s)
	if err != nil {
		return 0
	}

	return d
}
