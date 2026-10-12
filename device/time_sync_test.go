package device

// Typed time-sync primitives (issue #129): DeviceTime read, the
// ErrNoCredentials write gate, and the SetNTPServer one-shot switch.
// Wire goldens ride the internal/testutil.FakeCaller, which records the
// marshaled request bodies through the exact decode path of the real
// transport.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

// credCaller augments the FakeCaller with the root Client's credential
// accessor, so the ErrNoCredentials gate can be exercised the same way the
// production client triggers it.
type credCaller struct {
	*testutil.FakeCaller
	username string
}

func (c *credCaller) GetCredentials() (string, string) {
	return c.username, "secret"
}

// newDeviceHandlerService wires a Service over a multi-action handler
// (newDeviceOpsService pins a single action; the two-step SetNTPServer
// flow needs dispatch by request). The tds: prefix is stripped so the
// handler dispatches on the bare action name.
func newDeviceHandlerService(t *testing.T, handler func(action, reqXML string) (string, error)) (*Service, *testutil.FakeCaller) {
	t.Helper()

	caller := testutil.NewFakeCaller("http://fake/device_service", func(action, reqXML string) (string, error) {
		return handler(strings.TrimPrefix(action, "tds:"), reqXML)
	})

	return New(caller), caller
}

// hikvisionClockResponse is a Hikvision-style GetSystemDateAndTimeResponse:
// UTC + local wall clocks, POSIX TZ, NTP mode, DST off.
const hikvisionClockResponse = `<GetSystemDateAndTimeResponse xmlns="http://www.onvif.org/ver10/device/wsdl">
<SystemDateAndTime>
<DateTimeType>NTP</DateTimeType>
<DaylightSavings>false</DaylightSavings>
<TimeZone><TZ>CST-8</TZ></TimeZone>
<UTCDateTime>
<Time><Hour>9</Hour><Minute>58</Minute><Second>41</Second></Time>
<Date><Year>2026</Year><Month>10</Month><Day>12</Day></Date>
</UTCDateTime>
<LocalDateTime>
<Time><Hour>17</Hour><Minute>58</Minute><Second>41</Second></Time>
<Date><Year>2026</Year><Month>10</Month><Day>12</Day></Date>
</LocalDateTime>
</SystemDateAndTime>
</GetSystemDateAndTimeResponse>`

func TestGetDeviceTimeDecodesClock(t *testing.T) {
	svc, caller := newDeviceOpsService(t, "GetSystemDateAndTime", hikvisionClockResponse)

	got, err := svc.GetDeviceTime(context.Background())
	if err != nil {
		t.Fatalf("GetDeviceTime: %v", err)
	}

	wantUTC := time.Date(2026, 10, 12, 9, 58, 41, 0, time.UTC)
	if !got.UTCDateTime.Equal(wantUTC) {
		t.Errorf("UTCDateTime = %v, want %v", got.UTCDateTime, wantUTC)
	}

	// LocalDateTime carries the camera's own wall clock (components
	// assembled as a UTC-tagged instant — a display view, not a moment).
	wantLocal := time.Date(2026, 10, 12, 17, 58, 41, 0, time.UTC)
	if !got.LocalDateTime.Equal(wantLocal) {
		t.Errorf("LocalDateTime = %v, want %v", got.LocalDateTime, wantLocal)
	}

	if got.TimeZone != "CST-8" {
		t.Errorf("TimeZone = %q, want CST-8", got.TimeZone)
	}

	if got.DateTimeType != SetDateTimeNTP {
		t.Errorf("DateTimeType = %q, want NTP", got.DateTimeType)
	}

	if got.DaylightSavings {
		t.Error("DaylightSavings = true, want false")
	}

	if got := caller.CountAction("tds:GetSystemDateAndTime"); got != 1 {
		t.Errorf("GetSystemDateAndTime sent %d times, want 1", got)
	}
}

func TestGetDeviceTimeRejectsUnusableClock(t *testing.T) {
	unusable := strings.Replace(hikvisionClockResponse, "<Year>2026</Year>", "<Year>1999</Year>", 1)
	svc, _ := newDeviceOpsService(t, "GetSystemDateAndTime", unusable)

	if _, err := svc.GetDeviceTime(context.Background()); err == nil {
		t.Fatal("GetDeviceTime accepted a pre-2000 clock (uninitialized device)")
	}
}

func TestGetDeviceTimeRejectsMissingClock(t *testing.T) {
	empty := `<GetSystemDateAndTimeResponse xmlns="http://www.onvif.org/ver10/device/wsdl"/>`
	svc, _ := newDeviceOpsService(t, "GetSystemDateAndTime", empty)

	if _, err := svc.GetDeviceTime(context.Background()); err == nil {
		t.Fatal("GetDeviceTime accepted a response without any clock")
	}
}

func TestWriteOpsRequireCredentials(t *testing.T) {
	caller := &credCaller{FakeCaller: testutil.NewFakeCaller("http://fake/device_service", func(_, _ string) (string, error) {
		t.Error("write operation must not reach the wire without credentials")
		return "", nil
	}), username: ""}

	svc := New(caller)

	if err := svc.SetSystemDateAndTime(context.Background(), &SystemDateTime{DateTimeType: SetDateTimeManual}); !errors.Is(err, ErrNoCredentials) {
		t.Errorf("SetSystemDateAndTime without credentials: err = %v, want ErrNoCredentials", err)
	}

	if err := svc.SetNTP(context.Background(), false, []NetworkHost{{Type: "IPv4", IPv4Address: "10.1.1.1"}}); !errors.Is(err, ErrNoCredentials) {
		t.Errorf("SetNTP without credentials: err = %v, want ErrNoCredentials", err)
	}

	if err := svc.SetNTPServer(context.Background(), []NetworkHost{{Type: "IPv4", IPv4Address: "10.1.1.1"}}); !errors.Is(err, ErrNoCredentials) {
		t.Errorf("SetNTPServer without credentials: err = %v, want ErrNoCredentials", err)
	}
}

// TestWriteOpsSkipCredentialGateOnOpaqueCaller pins the fail-open contract:
// a custom api.Caller without GetCredentials is not second-guessed — the
// device is left to judge (e.g. proxies that inject credentials upstream).
func TestWriteOpsSkipCredentialGateOnOpaqueCaller(t *testing.T) {
	svc, _ := newDeviceOpsService(t, "SetSystemDateAndTime", `<SetSystemDateAndTimeResponse xmlns="http://www.onvif.org/ver10/device/wsdl"/>`)

	st := &SystemDateTime{
		DateTimeType: SetDateTimeManual,
		TimeZone:     &TimeZone{TZ: "CST-8"},
	}

	if err := svc.SetSystemDateAndTime(context.Background(), st); err != nil {
		t.Errorf("SetSystemDateAndTime over opaque caller: %v", err)
	}
}

func TestSetNTPServerTwoStepWire(t *testing.T) {
	var bodies []string

	svc, _ := newDeviceHandlerService(t, func(action, reqXML string) (string, error) {
		bodies = append(bodies, strings.TrimPrefix(action, "tds:")+"\x00"+reqXML)

		switch strings.TrimPrefix(action, "tds:") {
		case "SetNTP":
			return `<SetNTPResponse xmlns="http://www.onvif.org/ver10/device/wsdl"/>`, nil
		case "SetSystemDateAndTime":
			return `<SetSystemDateAndTimeResponse xmlns="http://www.onvif.org/ver10/device/wsdl"/>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	hosts := []NetworkHost{
		{Type: "IPv4", IPv4Address: "10.1.1.53"},
		{Type: "DNS", DNSname: "ntp.example.org"},
	}

	if err := svc.SetNTPServer(context.Background(), hosts); err != nil {
		t.Fatalf("SetNTPServer: %v", err)
	}

	if len(bodies) != 2 {
		t.Fatalf("SetNTPServer sent %d requests, want 2 (SetNTP + DateTimeType flip)", len(bodies))
	}

	setNTP, flip := splitRecorded(bodies[0]), splitRecorded(bodies[1])

	if setNTP.action != "SetNTP" {
		t.Errorf("first request = %s, want SetNTP", setNTP.action)
	}

	for _, want := range []string{
		"<tds:FromDHCP>false</tds:FromDHCP>",
		"<tt:Type>IPv4</tt:Type>",
		"<tt:IPv4Address>10.1.1.53</tt:IPv4Address>",
		"<tt:Type>DNS</tt:Type>",
		"<tt:DNSname>ntp.example.org</tt:DNSname>",
	} {
		if !strings.Contains(setNTP.body, want) {
			t.Errorf("SetNTP body missing %q:\n%s", want, setNTP.body)
		}
	}

	if flip.action != "SetSystemDateAndTime" {
		t.Errorf("second request = %s, want SetSystemDateAndTime", flip.action)
	}

	for _, want := range []string{">NTP</tds:DateTimeType>", "<tds:DaylightSavings>false</tds:DaylightSavings>"} {
		if !strings.Contains(flip.body, want) {
			t.Errorf("DateTimeType flip body missing %q:\n%s", want, flip.body)
		}
	}
}

func TestSetNTPServerFlipFailureKeepsServerMessage(t *testing.T) {
	svc, _ := newDeviceHandlerService(t, func(action, _ string) (string, error) {
		if strings.TrimPrefix(action, "tds:") == "SetNTP" {
			return `<SetNTPResponse xmlns="http://www.onvif.org/ver10/device/wsdl"/>`, nil
		}

		return "", errors.New("device refused mode switch")
	})

	err := svc.SetNTPServer(context.Background(), []NetworkHost{{Type: "IPv4", IPv4Address: "10.1.1.53"}})
	if err == nil {
		t.Fatal("SetNTPServer must report the flip failure")
	}

	if !strings.Contains(err.Error(), "NTP server configured") {
		t.Errorf("flip-failure error must say the server is configured but the mode did not switch: %v", err)
	}
}

func TestSetNTPServerRejectsEmptyHostList(t *testing.T) {
	svc, _ := newDeviceHandlerService(t, func(action, _ string) (string, error) {
		t.Errorf("no wire traffic expected, got %s", action)
		return "", nil
	})

	if err := svc.SetNTPServer(context.Background(), nil); err == nil {
		t.Fatal("SetNTPServer accepted an empty host list")
	}
}

func TestSetDeviceTimeWritesManualClock(t *testing.T) {
	svc, caller := newDeviceHandlerService(t, func(action, _ string) (string, error) {
		if strings.TrimPrefix(action, "tds:") != "SetSystemDateAndTime" {
			return "", errors.New("unexpected action " + action)
		}

		return `<SetSystemDateAndTimeResponse xmlns="http://www.onvif.org/ver10/device/wsdl"/>`, nil
	})

	dt := DeviceTime{
		UTCDateTime:  time.Date(2026, 10, 12, 9, 58, 41, 0, time.UTC),
		TimeZone:     "CST-8",
		DateTimeType: SetDateTimeManual,
	}

	if err := svc.SetDeviceTime(context.Background(), dt); err != nil {
		t.Fatalf("SetDeviceTime: %v", err)
	}

	reqs := caller.Requests()
	if len(reqs) != 1 || reqs[0].Action != "tds:SetSystemDateAndTime" {
		t.Fatalf("requests = %+v, want one SetSystemDateAndTime", reqs)
	}

	for _, want := range []string{
		">Manual</tds:DateTimeType>",
		">CST-8</tds:TZ>",
		"<tt:Year>2026</tt:Year>",
		"<tt:Month>10</tt:Month>",
		"<tt:Day>12</tt:Day>",
		"<tt:Hour>9</tt:Hour>",
		"<tt:Minute>58</tt:Minute>",
		"<tt:Second>41</tt:Second>",
	} {
		if !strings.Contains(reqs[0].Body, want) {
			t.Errorf("SetDeviceTime body missing %q:\n%s", want, reqs[0].Body)
		}
	}
}

// TestGetSystemDateAndTimeDelegates pins the deprecation shim: the legacy
// interface{} getter now returns the typed *SystemDateTime instead of the
// unusable decoded map it produced historically.
func TestGetSystemDateAndTimeDelegates(t *testing.T) {
	svc, _ := newDeviceOpsService(t, "GetSystemDateAndTime", hikvisionClockResponse)

	got, err := svc.GetSystemDateAndTime(context.Background())
	if err != nil {
		t.Fatalf("GetSystemDateAndTime: %v", err)
	}

	st, ok := got.(*SystemDateTime)
	if !ok {
		t.Fatalf("GetSystemDateAndTime returned %T, want *SystemDateTime", got)
	}

	if st.UTCDateTime == nil || st.UTCDateTime.Date.Year != 2026 {
		t.Errorf("delegated clock = %+v, want the typed 2026 UTC clock", st.UTCDateTime)
	}
}

type recordedRequest struct {
	action string
	body   string
}

func splitRecorded(rec string) recordedRequest {
	action, body, _ := strings.Cut(rec, "\x00")

	return recordedRequest{action: action, body: body}
}
