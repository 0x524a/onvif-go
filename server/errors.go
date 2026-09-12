package server

import (
	"errors"

	"github.com/0x524a/onvif-go"
)

var (
	// ErrVideoSourceNotFound is returned when a video source is not found.
	// Aliased to the client package's sentinel so errors.Is matches
	// regardless of whether the caller is comparing against the server or
	// client package's error variable.
	ErrVideoSourceNotFound = onvif.ErrVideoSourceNotFound

	// ErrProfileNotFound is returned when a profile is not found.
	ErrProfileNotFound = onvif.ErrProfileNotFound

	// ErrSnapshotNotSupported is returned when snapshot is not supported for a profile.
	ErrSnapshotNotSupported = onvif.ErrSnapshotNotSupported

	// ErrPTZNotSupported is returned when PTZ is not supported for a profile.
	ErrPTZNotSupported = onvif.ErrPTZNotSupported

	// ErrPresetNotFound is returned when a preset is not found.
	ErrPresetNotFound = onvif.ErrPresetNotFound

	// ErrSubscriptionNotFound is returned when there is no active pull-point
	// event subscription.
	ErrSubscriptionNotFound = errors.New("no active pull-point subscription")

	// ErrInvalidDuration is returned when an ISO-8601 duration string cannot
	// be parsed.
	ErrInvalidDuration = errors.New("invalid duration")
)
