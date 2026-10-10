// Package discovery finds ONVIF devices on the network with WS-Discovery, and
// can also answer those probes on behalf of devices of your own.
//
// # Finding cameras
//
// [Discover] sends a WS-Discovery Probe to the multicast group and collects the
// ProbeMatches that come back until the timeout expires:
//
//	devices, err := discovery.Discover(ctx, 5*time.Second)
//	for _, d := range devices {
//	    fmt.Println(d.GetName(), d.GetDeviceEndpoint())
//	}
//
// Use [DiscoverWithOptions] to choose the network interface, or to probe a
// single address by unicast with [DiscoverOptions].ProbeAddress, which needs no
// multicast routing and so works in containers and CI.
//
// # Answering probes
//
// A [Responder] is a WS-Discovery target service. Register one or more
// [Device] values and it answers matching Probes, and announces them with Hello
// and Bye messages. The server package builds on this: its DiscoveryDevice
// method describes a virtual camera in the form a Responder expects.
//
//	r := discovery.NewResponder(&discovery.ResponderConfig{})
//	r.Register(device)
//	go r.Start(ctx)
//
// Scope and type matching follow the WS-Discovery 2005/04 rules, including the
// default rfc3986 scope match.
//
// # Stable identities
//
// A device is identified by its endpoint reference, a "urn:uuid:" URI.
// [StableEndpointRef] derives one from a name, so a service that restarts gets
// the same identity back without storing anything.
package discovery
