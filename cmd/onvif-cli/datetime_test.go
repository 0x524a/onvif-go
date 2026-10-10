package main

import (
	"strings"
	"testing"

	"github.com/0x524a/onvif-go"
)

func TestFormatSystemDateTime(t *testing.T) {
	dt := &onvif.SystemDateTime{
		DateTimeType:    onvif.SetDateTimeNTP,
		DaylightSavings: false,
		TimeZone:        &onvif.TimeZone{TZ: "CET-1CEST"},
		UTCDateTime: &onvif.DateTime{
			Date: onvif.Date{Year: 2026, Month: 10, Day: 5},
			Time: onvif.Time{Hour: 7, Minute: 4, Second: 9},
		},
	}

	got := formatSystemDateTime(dt)

	for _, want := range []string{
		"Set by: NTP",
		"Daylight savings: false",
		"Time zone: CET-1CEST",
		"UTC:   2026-10-05 07:04:09",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}

	if strings.Contains(got, "Local:") {
		t.Errorf("Local line printed although the camera reported none:\n%s", got)
	}

	if strings.Contains(got, "0x") || strings.Contains(got, "&{") {
		t.Errorf("output looks like a printed pointer or struct:\n%s", got)
	}
}

func TestFormatSystemDateTimeNil(t *testing.T) {
	if got := formatSystemDateTime(nil); !strings.Contains(got, "no date/time") {
		t.Errorf("nil input gave %q", got)
	}
}
