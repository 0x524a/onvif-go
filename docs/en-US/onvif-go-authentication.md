# Authentication

Real camera firmware varies wildly in what authentication it accepts — per
device, per service on the same device, even per firmware revision. This
document explains the auth modes, the fallback ladder, clock-skew handling,
and the diagnostics API.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2@v2.2.0
```

## Auth modes

```go
type AuthMode string

const (
    AuthDigest       AuthMode = "digest"        // WS-Security PasswordDigest (default)
    AuthPasswordText AuthMode = "password-text" // cleartext in UsernameToken
    AuthHTTPBasic    AuthMode = "http-basic"    // HTTP Basic, no WS-Security
    AuthNone         AuthMode = "none"          // no auth at all
)
```

Select one with `WithAuthMode`. The default (`AuthDigest`) preserves the
historical behavior exactly.

Where each mode comes from in the field:

| Device behavior | Mode that works |
|---|---|
| Standard ONVIF firmware | `AuthDigest` |
| Rejects digest on specific services (PTZ, GetUsers) but accepts cleartext tokens | `AuthPasswordText` |
| Rejects WS-Security on imaging but accepts HTTP Basic (firmware-dependent) | `AuthHTTPBasic` |
| ESP32-class minimal firmware rejecting every auth-bearing request | `AuthNone` |

## The fallback ladder

When you cannot know in advance what a device accepts, let the client find
out once:

```go
client, _ := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAuthFallback(onvif.AuthPasswordText, onvif.AuthHTTPBasic, onvif.AuthNone),
)
```

On an auth-class failure (see classification below) the next mode is tried;
the first mode that works is **remembered** (sticky), so subsequent calls go
straight to it instead of paying the ladder cost on every request. Changing
credentials (`SetCredentials`) clears the memory — the conclusion was
measured for the old credentials. `ResetAuthLadder()` clears it manually
(after device-side changes); `AuthLadderMode()` reports the effective mode.

Non-auth errors are never retried with another mode: a network failure has
nothing to do with the auth scheme, and blind retries just add latency.

## Classifying auth failures

`errors.Is(err, onvif.ErrUnauthorized)` matches every authentication-shaped
failure:

- HTTP 401 / 403,
- a SOAP Fault with a NotAuthorized code (e.g. `ter:NotAuthorized`),
- a **200-status response carrying such a fault** — a common ONVIF quirk,
- exhaustion of the fallback ladder.

```go
_, err := client.Device().GetDeviceInformation(ctx)
if errors.Is(err, onvif.ErrUnauthorized) {
    // credential problem — not a network bug, not a parsing bug
}
```

## Clock skew (the Hikvision trap)

WS-Security digests embed a `Created` timestamp; devices reject digests
outside their replay window (commonly ±5 min). When the camera's clock
diverges from yours, **every digest looks tampered** and comes back as a
generic "sender not authorized" — the most misleading auth failure there is.

Two APIs deal with it:

```go
// Option: measure once during Initialize (before the first authenticated
// call), apply silently. Measurement failure degrades to the local clock.
client, _ := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAutoClockSkew())

// Manual: RTT-compensated measurement (local reference = round-trip
// midpoint, so network latency does not pollute the value).
skew, err := client.MeasureClockSkew(ctx)
client.SetClockSkew(skew)
```

## DiagnoseAuth — telling the three causes apart

When auth "just fails", there are three very different root causes: clock
skew (fix the camera's NTP), wrong credentials, or a device that rejects
ONVIF auth entirely. `DiagnoseAuth` separates them:

```go
diag, err := client.DiagnoseAuth(ctx)
switch diag.Status {
case onvif.AuthStatusOK:             // digest works
case onvif.AuthStatusClockSkew:      // works with the device clock — fix NTP,
                                     // or keep WithAutoClockSkew
    fmt.Println(diag.ClockSkew)      // measured divergence
case onvif.AuthStatusBadCredentials: // fails even with device time
}
```

The procedure: probe digest auth → on failure, measure the skew → if it
exceeds 2 minutes, retry with the device's time: success pins clock skew,
failure points at credentials.
