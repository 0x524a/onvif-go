package server

import (
	"net/url"

	"github.com/0x524a/onvif-go/discovery"
)

// DiscoveryDevice describes this server as a WS-Discovery target, ready to
// pass to discovery.Responder.Register. It carries the stable endpoint
// reference (see Config.EndpointUUID), the device service XAddr, and the
// standard ONVIF name, hardware, type and Streaming-profile scopes derived
// from Config.DeviceInfo.
//
// The XAddr uses the bound port, so call it after Start has bound the
// listener (Config.Ready signals this); before that it reports Config.Port,
// which is wrong for an OS-assigned port 0.
func (s *Server) DiscoveryDevice() *discovery.Device {
	info := s.config.DeviceInfo

	hardware := info.HardwareID
	if hardware == "" {
		hardware = info.Model
	}

	scopes := []string{
		"onvif://www.onvif.org/type/video_encoder",
		"onvif://www.onvif.org/Profile/Streaming",
	}
	if info.Model != "" {
		scopes = append(scopes, "onvif://www.onvif.org/name/"+url.PathEscape(info.Model))
	}

	if hardware != "" {
		scopes = append(scopes, "onvif://www.onvif.org/hardware/"+url.PathEscape(hardware))
	}

	return &discovery.Device{
		EndpointRef:     s.endpointURN,
		XAddrs:          []string{s.advertisedBaseURL() + "/device_service"},
		Scopes:          scopes,
		MetadataVersion: 1,
	}
}
