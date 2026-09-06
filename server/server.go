// Package server provides ONVIF server implementation for testing and simulation.
package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/0x524a/onvif-go/server/soap"
)

// New creates a new ONVIF server with the given configuration.
func New(config *Config) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	}

	server := &Server{
		config:       config,
		streams:      make(map[string]*StreamConfig),
		ptzState:     make(map[string]*PTZState),
		imagingState: make(map[string]*ImagingState),
		systemTime:   time.Now(),
	}

	// Initialize streams for each profile
	for i := range config.Profiles {
		profile := &config.Profiles[i]
		streamPath := fmt.Sprintf("/stream%d", i)

		host := config.Host
		if host == defaultHost || host == "" {
			host = "localhost"
		}

		streamURI := fmt.Sprintf("rtsp://%s:8554%s", host, streamPath)

		server.streams[profile.Token] = &StreamConfig{
			ProfileToken: profile.Token,
			RTSPPath:     streamPath,
			StreamURI:    streamURI,
		}

		// Initialize PTZ state if PTZ is supported
		if profile.PTZ != nil {
			server.ptzState[profile.Token] = &PTZState{
				Position:   PTZPosition{Pan: 0, Tilt: 0, Zoom: 0},
				Moving:     false,
				PanMoving:  false,
				TiltMoving: false,
				ZoomMoving: false,
				LastUpdate: time.Now(),
			}
		}

		// Initialize imaging state
		server.imagingState[profile.VideoSource.Token] = &ImagingState{
			Brightness:  50.0, //nolint:mnd // Default imaging value
			Contrast:    50.0, //nolint:mnd // Default imaging value
			Saturation:  50.0, //nolint:mnd // Default imaging value
			Sharpness:   50.0, //nolint:mnd // Default imaging value
			IrCutFilter: autoMode,
			BacklightComp: BacklightCompensation{
				Mode:  offMode,
				Level: 0,
			},
			Exposure: ExposureSettings{
				Mode:         autoMode,
				Priority:     "FrameRate",
				MinExposure:  1,
				MaxExposure:  10000, //nolint:mnd // Exposure time in microseconds
				MinGain:      0,
				MaxGain:      100, //nolint:mnd // Gain value
				ExposureTime: 100, //nolint:mnd // Exposure time
				Gain:         50,  //nolint:mnd // Gain value
			},
			Focus: FocusSettings{
				AutoFocusMode: autoMode,
				DefaultSpeed:  0.5, //nolint:mnd // Focus speed
				NearLimit:     0,
				FarLimit:      1,
				CurrentPos:    0.5, //nolint:mnd // Focus position
			},
			WhiteBalance: WhiteBalanceSettings{
				Mode:   autoMode,
				CrGain: 128, //nolint:mnd // White balance gain
				CbGain: 128, //nolint:mnd // White balance gain
			},
			WideDynamicRange: WDRSettings{
				Mode:  offMode,
				Level: 0,
			},
		}
	}

	return server, nil
}

// Start starts the ONVIF server.
func (s *Server) Start(ctx context.Context) error {
	// Create HTTP server
	mux := http.NewServeMux()

	// Register service handlers
	s.registerDeviceService(mux)
	s.registerMediaService(mux)

	if s.config.SupportPTZ {
		s.registerPTZService(mux)
	}

	if s.config.SupportImaging {
		s.registerImagingService(mux)
	}

	if s.config.SupportEvents {
		s.registerEventService(mux)
	}

	// Add snapshot endpoint
	mux.HandleFunc(s.config.BasePath+"/snapshot", s.handleSnapshot)

	// Bind before serving, as a distinct step. ListenAndServe resolves the
	// address internally, which left a caller unable to learn the port an
	// OS-assigned Port: 0 produced, and unable to know when the server had
	// started accepting - forcing tests to hardcode a port and then sleep or
	// poll-dial. Binding here also surfaces an "address already in use"
	// failure as Start's own return value instead of racing through errChan.
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listenerMu.Lock()
	s.listener = listener
	s.listenerMu.Unlock()

	httpServer := &http.Server{
		Handler:      mux,
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
	}

	// Report the address actually bound, not the requested one, so Port: 0
	// prints a URL that can be used.
	s.printBanner(listener.Addr().String())

	// The listener is already queueing connections, so anything dialing after
	// this point cannot be refused.
	s.signalReady()

	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		fmt.Fprintf(s.output(), "\n🛑 Shutting down server...\n")
		const shutdownTimeout = 5 // Server shutdown timeout in seconds
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		return nil
	case err := <-errChan:
		return err
	}
}

// output returns the writer for the startup banner and shutdown notice,
// defaulting to io.Discard so that merely importing this package never writes
// to a program's stdout. See Config.Output.
func (s *Server) output() io.Writer {
	if s.config.Output == nil {
		return io.Discard
	}

	return s.config.Output
}

// signalReady closes Config.Ready if one was supplied, at most once.
func (s *Server) signalReady() {
	if s.config.Ready == nil {
		return
	}

	s.readyOnce.Do(func() {
		close(s.config.Ready)
	})
}

// Addr returns the address the server is listening on, or "" before Start has
// bound one. With Config.Port set to 0 this is how a caller learns the port the
// OS assigned.
func (s *Server) Addr() string {
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()

	if s.listener == nil {
		return ""
	}

	return s.listener.Addr().String()
}

// advertisedHost returns the host to put in URLs the server hands out. A
// wildcard bind has no single reachable host, so it degrades to localhost.
func (s *Server) advertisedHost() string {
	host := s.config.Host
	if host == defaultHost || host == "" {
		host = defaultHostname
	}

	return host
}

// advertisedHostPort returns the host and port for URLs the server hands out.
//
// The port comes from the bound listener when there is one, in preference to
// Config.Port. With Config.Port 0 the configured value stays 0, which would
// make every advertised URL point at port 0 - including the
// SubscriptionReference that pull-point clients genuinely call back on, and the
// XAddrs a client uses to reach every other service after GetCapabilities. The
// host still comes from config, for the wildcard-bind reason in
// advertisedHost.
func (s *Server) advertisedHostPort() (host string, port int) {
	port = s.config.Port

	s.listenerMu.RLock()
	listener := s.listener
	s.listenerMu.RUnlock()

	if listener != nil {
		if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
			port = tcpAddr.Port
		}
	}

	return s.advertisedHost(), port
}

// advertisedBaseURL returns the http://host:port/basepath prefix shared by the
// service XAddrs and the snapshot URI.
func (s *Server) advertisedBaseURL() string {
	host, port := s.advertisedHostPort()

	return fmt.Sprintf("http://%s:%d%s", host, port, s.config.BasePath)
}

// printBanner writes the human-readable startup summary for a server listening
// on addr. It writes nothing unless Config.Output is set.
func (s *Server) printBanner(addr string) {
	out := s.output()
	if out == io.Discard {
		return
	}

	fmt.Fprintf(out, "🎥 ONVIF Server starting on %s\n", addr)
	fmt.Fprintf(out, "📡 Device Service: http://%s%s/device_service\n", addr, s.config.BasePath)
	fmt.Fprintf(out, "🎬 Media Service: http://%s%s/media_service\n", addr, s.config.BasePath)
	if s.config.SupportPTZ {
		fmt.Fprintf(out, "🎮 PTZ Service: http://%s%s/ptz_service\n", addr, s.config.BasePath)
	}
	if s.config.SupportImaging {
		fmt.Fprintf(out, "📷 Imaging Service: http://%s%s/imaging_service\n", addr, s.config.BasePath)
	}
	if s.config.SupportEvents {
		fmt.Fprintf(out, "🔔 Event Service: http://%s%s/events_service\n", addr, s.config.BasePath)
	}
	fmt.Fprintf(out, "\n🌐 Virtual Camera Profiles:\n")
	s.streamsMu.RLock()
	//nolint:gocritic // Range value copy is acceptable for small structs
	for i, profile := range s.config.Profiles {
		stream := s.streams[profile.Token]
		fmt.Fprintf(out, "   [%d] %s - %s (%dx%d @ %dfps)\n",
			i+1, profile.Name, stream.StreamURI,
			profile.VideoEncoder.Resolution.Width,
			profile.VideoEncoder.Resolution.Height,
			profile.VideoEncoder.Framerate)
	}
	s.streamsMu.RUnlock()
	fmt.Fprintf(out, "\n✅ Server is ready!\n\n")
}

// registerDeviceService registers the device service handler.
func (s *Server) registerDeviceService(mux *http.ServeMux) {
	handler := soap.NewHandler(s.config.Username, s.config.Password)

	// Register device service handlers
	handler.RegisterHandler("GetDeviceInformation", s.HandleGetDeviceInformation)
	handler.RegisterHandler("GetCapabilities", s.HandleGetCapabilities)
	handler.RegisterHandler("GetSystemDateAndTime", s.HandleGetSystemDateAndTime)
	handler.RegisterHandler("GetServices", s.HandleGetServices)
	handler.RegisterHandler("SystemReboot", s.HandleSystemReboot)

	mux.Handle(s.config.BasePath+"/device_service", handler)
}

// registerMediaService registers the media service handler.
func (s *Server) registerMediaService(mux *http.ServeMux) {
	handler := soap.NewHandler(s.config.Username, s.config.Password)

	// Register media service handlers. GetStreamUri/GetSnapshotUri are
	// registered under the ONVIF Media WSDL's actual spelling (lowercase
	// "Uri") - the real onvif.Client sends exactly that, and action
	// dispatch in soap.Handler.extractAction is case-sensitive, so a
	// "GetStreamURI" registration here silently never matches (#79).
	handler.RegisterHandler("GetProfiles", s.HandleGetProfiles)
	handler.RegisterHandler("GetStreamUri", s.HandleGetStreamURI)
	handler.RegisterHandler("GetSnapshotUri", s.HandleGetSnapshotURI)
	handler.RegisterHandler("GetVideoSources", s.HandleGetVideoSources)

	mux.Handle(s.config.BasePath+"/media_service", handler)
}

// registerPTZService registers the PTZ service handler.
func (s *Server) registerPTZService(mux *http.ServeMux) {
	handler := soap.NewHandler(s.config.Username, s.config.Password)

	// Register PTZ service handlers
	handler.RegisterHandler("ContinuousMove", s.HandleContinuousMove)
	handler.RegisterHandler("AbsoluteMove", s.HandleAbsoluteMove)
	handler.RegisterHandler("RelativeMove", s.HandleRelativeMove)
	handler.RegisterHandler("Stop", s.HandleStop)
	handler.RegisterHandler("GetStatus", s.HandleGetStatus)
	handler.RegisterHandler("GetPresets", s.HandleGetPresets)
	handler.RegisterHandler("GotoPreset", s.HandleGotoPreset)

	mux.Handle(s.config.BasePath+"/ptz_service", handler)
}

// registerImagingService registers the imaging service handler.
func (s *Server) registerImagingService(mux *http.ServeMux) {
	handler := soap.NewHandler(s.config.Username, s.config.Password)

	// Register imaging service handlers
	handler.RegisterHandler("GetImagingSettings", s.HandleGetImagingSettings)
	handler.RegisterHandler("SetImagingSettings", s.HandleSetImagingSettings)
	handler.RegisterHandler("GetOptions", s.HandleGetOptions)
	handler.RegisterHandler("Move", s.HandleMove)

	mux.Handle(s.config.BasePath+"/imaging_service", handler)
}

// registerEventService registers the event service handler.
func (s *Server) registerEventService(mux *http.ServeMux) {
	handler := soap.NewHandler(s.config.Username, s.config.Password)

	// Register event service handlers
	handler.RegisterHandler("CreatePullPointSubscription", s.HandleCreatePullPointSubscription)
	handler.RegisterHandler("PullMessages", s.HandlePullMessages)
	handler.RegisterHandler("Renew", s.HandleRenew)
	handler.RegisterHandler("Unsubscribe", s.HandleUnsubscribe)

	mux.Handle(s.config.BasePath+"/events_service", handler)
}

// handleSnapshot handles HTTP snapshot requests.
func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	// Get profile token from query parameter
	profileToken := r.URL.Query().Get("profile")
	if profileToken == "" {
		http.Error(w, "Missing profile parameter", http.StatusBadRequest)

		return
	}

	// Find the profile
	var profileCfg *ProfileConfig
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Token == profileToken {
			profileCfg = &s.config.Profiles[i]

			break
		}
	}

	if profileCfg == nil {
		http.Error(w, "Profile not found", http.StatusNotFound)

		return
	}

	if !profileCfg.Snapshot.Enabled {
		http.Error(w, "Snapshot not supported", http.StatusNotImplemented)

		return
	}

	// In a real implementation, this would capture a frame from the video source
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(http.StatusOK)

	// TODO: Generate or capture actual JPEG snapshot
}

// GetConfig returns the server configuration.
func (s *Server) GetConfig() *Config {
	return s.config
}

// GetStreamConfig returns the stream configuration for a profile.
func (s *Server) GetStreamConfig(profileToken string) (*StreamConfig, bool) {
	s.streamsMu.RLock()
	defer s.streamsMu.RUnlock()
	stream, ok := s.streams[profileToken]

	return stream, ok
}

// UpdateStreamURI updates the RTSP URI for a profile.
func (s *Server) UpdateStreamURI(profileToken, uri string) error {
	s.streamsMu.Lock()
	defer s.streamsMu.Unlock()

	stream, ok := s.streams[profileToken]
	if !ok {
		return fmt.Errorf("%w: %s", ErrProfileNotFound, profileToken)
	}
	stream.StreamURI = uri

	return nil
}

// ListProfiles returns all configured profiles.
func (s *Server) ListProfiles() []ProfileConfig {
	return s.config.Profiles
}

// GetPTZState returns the current PTZ state for a profile.
func (s *Server) GetPTZState(profileToken string) (*PTZState, bool) {
	s.ptzMu.RLock()
	defer s.ptzMu.RUnlock()
	state, ok := s.ptzState[profileToken]

	return state, ok
}

// GetImagingState returns the current imaging state for a video source.
func (s *Server) GetImagingState(videoSourceToken string) (*ImagingState, bool) {
	s.imagingMu.RLock()
	defer s.imagingMu.RUnlock()
	state, ok := s.imagingState[videoSourceToken]

	return state, ok
}

// ServerInfo returns human-readable server information.
func (s *Server) ServerInfo() string {
	var info string
	info += "ONVIF Server Configuration\n"
	info += "==========================\n"
	info += fmt.Sprintf("Device: %s %s\n", s.config.DeviceInfo.Manufacturer, s.config.DeviceInfo.Model)
	info += fmt.Sprintf("Firmware: %s\n", s.config.DeviceInfo.FirmwareVersion)
	info += fmt.Sprintf("Serial: %s\n", s.config.DeviceInfo.SerialNumber)
	info += fmt.Sprintf("\nServer Address: %s:%d\n", s.config.Host, s.config.Port)
	info += fmt.Sprintf("Base Path: %s\n", s.config.BasePath)
	info += fmt.Sprintf("\nProfiles (%d):\n", len(s.config.Profiles))
	//nolint:gocritic // Range value copy is acceptable for small structs
	for i, profile := range s.config.Profiles {
		info += fmt.Sprintf("  [%d] %s (%s)\n", i+1, profile.Name, profile.Token)
		info += fmt.Sprintf("      Video: %dx%d @ %dfps (%s)\n",
			profile.VideoEncoder.Resolution.Width,
			profile.VideoEncoder.Resolution.Height,
			profile.VideoEncoder.Framerate,
			profile.VideoEncoder.Encoding)
		if stream, ok := s.streams[profile.Token]; ok {
			info += fmt.Sprintf("      RTSP: %s\n", stream.StreamURI)
		}
		if profile.PTZ != nil {
			info += "      PTZ: Enabled\n"
		}
	}
	info += "\nCapabilities:\n"
	info += fmt.Sprintf("  PTZ: %v\n", s.config.SupportPTZ)
	info += fmt.Sprintf("  Imaging: %v\n", s.config.SupportImaging)
	info += fmt.Sprintf("  Events: %v\n", s.config.SupportEvents)

	return info
}
