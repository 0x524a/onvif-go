// Package discovery provides error definitions for the discovery package.
package discovery

import "errors"

var (
	// ErrNoProbeMatches is returned when no probe matches are found during discovery.
	ErrNoProbeMatches = errors.New("no probe matches found")

	// ErrNetworkInterfaceNotFound is returned when a network interface is not found.
	ErrNetworkInterfaceNotFound = errors.New("network interface not found")
)
