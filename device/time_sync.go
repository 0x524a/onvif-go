package device

// Typed time-sync primitives (issue #129): a time.Time-shaped read of the
// camera clock, an explicit credentials gate for the admin-level write
// operations, and the SetNTPServer one-shot switch.
//
// The read path rides the unauthenticated GetSystemDateAndTime (one of the
// few ONVIF operations devices must serve without credentials), so camera
// time stays observable even when authentication is broken — exactly the
// situation where it matters: a camera whose clock (and digest window) has
// drifted is the camera most in need of a clock fix, and its writes only
// succeed through the client's skew-corrected digest path anyway.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

// ErrNoCredentials is returned by the write operations when the client was
// constructed without a username. Callers surface it as an explicit
// "credentials required" state instead of a generic device rejection.
var ErrNoCredentials = errors.New("device: credentials required for this operation")

// credentialSource is satisfied by the root Client (GetCredentials). Custom
// api.Caller implementations that do not expose credentials are not
// second-guessed: the gate is skipped and the device is left to judge.
type credentialSource interface {
	GetCredentials() (username, password string)
}

// requireCredentials reports whether the caller can authenticate for an
// admin-level write.
func (s *Service) requireCredentials() error {
	if cs, ok := s.c.(credentialSource); ok {
		if username, _ := cs.GetCredentials(); username == "" {
			return ErrNoCredentials
		}
	}

	return nil
}

// minimumUsableYear rejects clocks from devices that have not set time at
// all (epoch 1970) or lost their RTC battery: anything before 2000 is
// firmware noise, not a clock.
const minimumUsableYear = 2000

// DeviceTime is the camera wall clock in usable Go types.
type DeviceTime struct {
	// UTCDateTime is the camera's clock as a UTC instant (second
	// resolution — the ONVIF struct carries no sub-second field).
	UTCDateTime time.Time
	// LocalDateTime is the camera's wall clock in its own configured
	// zone, assembled as a UTC-tagged instant: a component view for
	// display, not a distinct moment.
	LocalDateTime time.Time
	// TimeZone is the camera's POSIX TZ string (e.g. "CST-8"); empty on
	// devices that do not report one.
	TimeZone string
	// DateTimeType is how the camera sets its clock: Manual or NTP.
	DateTimeType SetDateTimeType
	// DaylightSavings reports the camera's DST flag.
	DaylightSavings bool
}

// GetDeviceTime reads the camera clock via the typed
// GetSystemDateAndTime. It does not require credentials and pairs with the
// root client's MeasureClockSkew / WithAutoClockSkew: this read is the
// same data source the skew measurement uses.
func (s *Service) GetDeviceTime(ctx context.Context) (*DeviceTime, error) {
	st, err := s.FixedGetSystemDateAndTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetDeviceTime: %w", err)
	}

	if st.UTCDateTime == nil {
		return nil, fmt.Errorf("GetDeviceTime: device returned no clock")
	}

	if st.UTCDateTime.Date.Year < minimumUsableYear {
		return nil, fmt.Errorf("GetDeviceTime: device returned no usable clock (year=%d)", st.UTCDateTime.Date.Year)
	}

	utc := st.UTCDateTime
	got := &DeviceTime{
		UTCDateTime: time.Date(
			utc.Date.Year, time.Month(utc.Date.Month), utc.Date.Day,
			utc.Time.Hour, utc.Time.Minute, utc.Time.Second, 0, time.UTC,
		),
		DateTimeType:    st.DateTimeType,
		DaylightSavings: st.DaylightSavings,
	}

	if st.LocalDateTime != nil && st.LocalDateTime.Date.Year >= minimumUsableYear {
		lt := st.LocalDateTime
		got.LocalDateTime = time.Date(
			lt.Date.Year, time.Month(lt.Date.Month), lt.Date.Day,
			lt.Time.Hour, lt.Time.Minute, lt.Time.Second, 0, time.UTC,
		)
	}

	if st.TimeZone != nil {
		got.TimeZone = st.TimeZone.TZ
	}

	return got, nil
}

// SetDeviceTime writes a Manual clock to the camera. The instant is sent
// verbatim (UTC components); an empty TimeZone omits the TimeZone element
// so the camera keeps its zone. Admin credentials required — a camera in
// NTP mode may refuse Manual setting with a Sender fault, which is
// surfaced as-is for the caller to decide (e.g. SetNTPServer instead).
func (s *Service) SetDeviceTime(ctx context.Context, t DeviceTime) error {
	if err := s.requireCredentials(); err != nil {
		return fmt.Errorf("SetDeviceTime: %w", err)
	}

	st := &SystemDateTime{
		DateTimeType:    SetDateTimeManual,
		DaylightSavings: t.DaylightSavings,
		UTCDateTime: &DateTime{
			Time: Time{
				Hour:   t.UTCDateTime.Hour(),
				Minute: t.UTCDateTime.Minute(),
				Second: t.UTCDateTime.Second(),
			},
			Date: Date{
				Year:  t.UTCDateTime.Year(),
				Month: int(t.UTCDateTime.Month()),
				Day:   t.UTCDateTime.Day(),
			},
		},
	}

	if t.TimeZone != "" {
		st.TimeZone = &TimeZone{TZ: t.TimeZone}
	}

	if err := s.SetSystemDateAndTime(ctx, st); err != nil {
		return fmt.Errorf("SetDeviceTime: %w", err)
	}

	return nil
}

// SetNTPServer points the camera at manual NTP servers and flips its
// DateTimeType to NTP so it actually uses them — the long-term
// self-healing path (the camera re-syncs on its own schedule, including
// after a power loss with a dead RTC battery). Admin credentials required.
//
// The flip runs as a second operation because several firmwares accept
// SetNTP alone but never switch modes implicitly: when the flip fails the
// error says so while the server list is already configured.
func (s *Service) SetNTPServer(ctx context.Context, hosts []NetworkHost) error {
	if len(hosts) == 0 {
		return fmt.Errorf("SetNTPServer: %w: at least one NTP server host is required", types.ErrInvalidParameter)
	}

	if err := s.requireCredentials(); err != nil {
		return fmt.Errorf("SetNTPServer: %w", err)
	}

	if err := s.SetNTP(ctx, false, hosts); err != nil {
		return fmt.Errorf("SetNTPServer: %w", err)
	}

	if err := s.SetSystemDateAndTime(ctx, &SystemDateTime{DateTimeType: SetDateTimeNTP}); err != nil {
		return fmt.Errorf("SetNTPServer: NTP server configured, but switching DateTimeType to NTP failed: %w", err)
	}

	return nil
}
