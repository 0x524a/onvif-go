// Package server is a virtual ONVIF camera. It serves the Device, Media, PTZ
// and Imaging services over SOAP so that clients, video management systems and
// CI jobs can be tested without hardware.
//
// # Running a camera
//
// Start from [DefaultConfig], which describes a three-lens camera, adjust it,
// and pass it to [New]:
//
//	cfg := server.DefaultConfig()
//	cfg.Port = 8080
//	srv, err := server.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	log.Fatal(srv.Start(ctx))
//
// Start blocks until the context is canceled. Set [Config].Ready to learn when
// the listener is bound, and use Port 0 to let the OS pick a free port, which
// [Server.Addr] then reports.
//
// # Identity
//
// Each camera has an endpoint reference, returned by the Device service's
// GetEndpointReference. Set [Config].EndpointUUID to pin it, or
// [Config].EndpointSeed to derive it from a stable name, and the camera keeps
// the same identity across restarts. With neither set, one is generated.
//
// # Discovery
//
// [Server.DiscoveryDevice] describes the camera for the discovery package's
// Responder, so that WS-Discovery clients find it on the network. Call it after
// Start has bound the listener.
package server
