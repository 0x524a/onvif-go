package main

import "errors"

var (
	// ErrNoNetworkInterfaces is returned when no network interfaces are found.
	ErrNoNetworkInterfaces = errors.New("no network interfaces found")

	// ErrNoCamerasFound is returned when no cameras are found on any interface.
	ErrNoCamerasFound = errors.New("no cameras found on any interface")

	// ErrNoActiveInterfaces is returned when no active interfaces are available for discovery.
	ErrNoActiveInterfaces = errors.New("no active interfaces available for discovery")

	// ErrNoProfilesFound is returned when no profiles are found.
	ErrNoProfilesFound = errors.New("no profiles found")

	// ErrNoVideoSourceConfiguration is returned when no video source configuration is found.
	ErrNoVideoSourceConfiguration = errors.New("no video source configuration found")
)
