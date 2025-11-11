package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/0x524A/go-onvif/server/soap"
)

// New creates a new ONVIF server with the given configuration
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
		if host == "0.0.0.0" || host == "" {
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
			Brightness:  50.0,
			Contrast:    50.0,
			Saturation:  50.0,
			Sharpness:   50.0,
			IrCutFilter: "AUTO",
			BacklightComp: BacklightCompensation{
				Mode:  "OFF",
				Level: 0,
			},
			Exposure: ExposureSettings{
				Mode:         "AUTO",
				Priority:     "FrameRate",
				MinExposure:  1,
				MaxExposure:  10000,
				MinGain:      0,
				MaxGain:      100,
				ExposureTime: 100,
				Gain:         50,
			},
			Focus: FocusSettings{
				AutoFocusMode: "AUTO",
				DefaultSpeed:  0.5,
				NearLimit:     0,
				FarLimit:      1,
				CurrentPos:    0.5,
			},
			WhiteBalance: WhiteBalanceSettings{
				Mode:   "AUTO",
				CrGain: 128,
				CbGain: 128,
			},
			WideDynamicRange: WDRSettings{
				Mode:  "OFF",
				Level: 0,
			},
		}
	}

	return server, nil
}

// Start starts the ONVIF server
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

	// Add snapshot endpoint
	mux.HandleFunc(s.config.BasePath+"/snapshot", s.handleSnapshot)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		fmt.Printf("🎥 ONVIF Server starting on %s\n", addr)
		fmt.Printf("📡 Device Service: http://%s%s/device_service\n", addr, s.config.BasePath)
		fmt.Printf("🎬 Media Service: http://%s%s/media_service\n", addr, s.config.BasePath)
		if s.config.SupportPTZ {
			fmt.Printf("🎮 PTZ Service: http://%s%s/ptz_service\n", addr, s.config.BasePath)
		}
		if s.config.SupportImaging {
			fmt.Printf("📷 Imaging Service: http://%s%s/imaging_service\n", addr, s.config.BasePath)
		}
		fmt.Printf("\n🌐 Virtual Camera Profiles:\n")
		for i, profile := range s.config.Profiles {
			stream := s.streams[profile.Token]
			fmt.Printf("   [%d] %s - %s (%dx%d @ %dfps)\n",
				i+1, profile.Name, stream.StreamURI,
				profile.VideoEncoder.Resolution.Width,
				profile.VideoEncoder.Resolution.Height,
				profile.VideoEncoder.Framerate)
		}
		fmt.Printf("\n✅ Server is ready!\n\n")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		fmt.Println("\n🛑 Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// registerDeviceService registers the device service handler
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

// registerMediaService registers the media service handler
func (s *Server) registerMediaService(mux *http.ServeMux) {
	handler := soap.NewHandler(s.config.Username, s.config.Password)

	// Register media service handlers
	handler.RegisterHandler("GetProfiles", s.HandleGetProfiles)
	handler.RegisterHandler("GetStreamURI", s.HandleGetStreamURI)
	handler.RegisterHandler("GetSnapshotURI", s.HandleGetSnapshotURI)
	handler.RegisterHandler("GetVideoSources", s.HandleGetVideoSources)

	mux.Handle(s.config.BasePath+"/media_service", handler)
}

// registerPTZService registers the PTZ service handler
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

// registerImagingService registers the imaging service handler
func (s *Server) registerImagingService(mux *http.ServeMux) {
	handler := soap.NewHandler(s.config.Username, s.config.Password)

	// Register imaging service handlers
	handler.RegisterHandler("GetImagingSettings", s.HandleGetImagingSettings)
	handler.RegisterHandler("SetImagingSettings", s.HandleSetImagingSettings)
	handler.RegisterHandler("GetOptions", s.HandleGetOptions)
	handler.RegisterHandler("Move", s.HandleMove)

	mux.Handle(s.config.BasePath+"/imaging_service", handler)
}

// handleSnapshot handles HTTP snapshot requests
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

// GetConfig returns the server configuration
func (s *Server) GetConfig() *Config {
	return s.config
}

// GetStreamConfig returns the stream configuration for a profile
func (s *Server) GetStreamConfig(profileToken string) (*StreamConfig, bool) {
	stream, ok := s.streams[profileToken]
	return stream, ok
}

// UpdateStreamURI updates the RTSP URI for a profile
func (s *Server) UpdateStreamURI(profileToken, uri string) error {
	stream, ok := s.streams[profileToken]
	if !ok {
		return fmt.Errorf("profile not found: %s", profileToken)
	}
	stream.StreamURI = uri
	return nil
}

// ListProfiles returns all configured profiles
func (s *Server) ListProfiles() []ProfileConfig {
	return s.config.Profiles
}

// GetPTZState returns the current PTZ state for a profile
func (s *Server) GetPTZState(profileToken string) (*PTZState, bool) {
	ptzMutex.RLock()
	defer ptzMutex.RUnlock()
	state, ok := s.ptzState[profileToken]
	return state, ok
}

// GetImagingState returns the current imaging state for a video source
func (s *Server) GetImagingState(videoSourceToken string) (*ImagingState, bool) {
	imagingMutex.RLock()
	defer imagingMutex.RUnlock()
	state, ok := s.imagingState[videoSourceToken]
	return state, ok
}

// ServerInfo returns human-readable server information
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
