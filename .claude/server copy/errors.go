package server

import "errors"

var (
	// ErrVideoSourceNotFound is returned when a video source is not found.
	ErrVideoSourceNotFound = errors.New("video source not found")

	// ErrProfileNotFound is returned when a profile is not found.
	ErrProfileNotFound = errors.New("profile not found")

	// ErrSnapshotNotSupported is returned when snapshot is not supported for a profile.
	ErrSnapshotNotSupported = errors.New("snapshot not supported for profile")

	// ErrPTZNotSupported is returned when PTZ is not supported for a profile.
	ErrPTZNotSupported = errors.New("PTZ not supported for profile")

	// ErrPresetNotFound is returned when a preset is not found.
	ErrPresetNotFound = errors.New("preset not found")
)
