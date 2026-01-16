package server

import (
	"encoding/xml"
	"fmt"
	"sync"
	"time"
)

// PTZ service SOAP message types

// ContinuousMoveRequest represents ContinuousMove request.
type ContinuousMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl ContinuousMove"`
	ProfileToken string    `xml:"ProfileToken"`
	Velocity     PTZVector `xml:"Velocity"`
	Timeout      string    `xml:"Timeout,omitempty"`
}

// ContinuousMoveResponse represents ContinuousMove response.
type ContinuousMoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl ContinuousMoveResponse"`
}

// AbsoluteMoveRequest represents AbsoluteMove request.
type AbsoluteMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl AbsoluteMove"`
	ProfileToken string    `xml:"ProfileToken"`
	Position     PTZVector `xml:"Position"`
	Speed        PTZVector `xml:"Speed,omitempty"`
}

// AbsoluteMoveResponse represents AbsoluteMove response.
type AbsoluteMoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl AbsoluteMoveResponse"`
}

// RelativeMoveRequest represents RelativeMove request.
type RelativeMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl RelativeMove"`
	ProfileToken string    `xml:"ProfileToken"`
	Translation  PTZVector `xml:"Translation"`
	Speed        PTZVector `xml:"Speed,omitempty"`
}

// RelativeMoveResponse represents RelativeMove response.
type RelativeMoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl RelativeMoveResponse"`
}

// StopRequest represents Stop request.
type StopRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl Stop"`
	ProfileToken string   `xml:"ProfileToken"`
	PanTilt      bool     `xml:"PanTilt,omitempty"`
	Zoom         bool     `xml:"Zoom,omitempty"`
}

// StopResponse represents Stop response.
type StopResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl StopResponse"`
}

// GetStatusRequest represents GetStatus request.
type GetStatusRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetStatus"`
	ProfileToken string   `xml:"ProfileToken"`
}

// GetStatusResponse represents GetStatus response.
type GetStatusResponse struct {
	XMLName   xml.Name   `xml:"http://www.onvif.org/ver20/ptz/wsdl GetStatusResponse"`
	PTZStatus *PTZStatus `xml:"PTZStatus"`
}

// PTZStatus represents PTZ status.
type PTZStatus struct {
	Position   PTZVector     `xml:"Position"`
	MoveStatus PTZMoveStatus `xml:"MoveStatus"`
	UTCTime    string        `xml:"UtcTime"`
}

// PTZMoveStatus represents PTZ movement status.
type PTZMoveStatus struct {
	PanTilt string `xml:"PanTilt,omitempty"`
	Zoom    string `xml:"Zoom,omitempty"`
}

// PTZVector represents PTZ position/velocity.
type PTZVector struct {
	PanTilt *Vector2D `xml:"PanTilt,omitempty"`
	Zoom    *Vector1D `xml:"Zoom,omitempty"`
}

// Vector2D represents a 2D vector.
type Vector2D struct {
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// Vector1D represents a 1D vector.
type Vector1D struct {
	X     float64 `xml:"x,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// GetPresetsRequest represents GetPresets request.
type GetPresetsRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetPresets"`
	ProfileToken string   `xml:"ProfileToken"`
}

// GetPresetsResponse represents GetPresets response.
type GetPresetsResponse struct {
	XMLName xml.Name    `xml:"http://www.onvif.org/ver20/ptz/wsdl GetPresetsResponse"`
	Preset  []PTZPreset `xml:"Preset"`
}

// PTZPreset represents a PTZ preset.
type PTZPreset struct {
	Token       string     `xml:"token,attr"`
	Name        string     `xml:"Name"`
	PTZPosition *PTZVector `xml:"PTZPosition,omitempty"`
}

// GotoPresetRequest represents GotoPreset request.
type GotoPresetRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl GotoPreset"`
	ProfileToken string    `xml:"ProfileToken"`
	PresetToken  string    `xml:"PresetToken"`
	Speed        PTZVector `xml:"Speed,omitempty"`
}

// GotoPresetResponse represents GotoPreset response.
type GotoPresetResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GotoPresetResponse"`
}

// SetPresetRequest represents SetPreset request.
type SetPresetRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl SetPreset"`
	ProfileToken string   `xml:"ProfileToken"`
	PresetName   string   `xml:"PresetName,omitempty"`
	PresetToken  string   `xml:"PresetToken,omitempty"`
}

// SetPresetResponse represents SetPreset response.
type SetPresetResponse struct {
	XMLName     xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl SetPresetResponse"`
	PresetToken string   `xml:"PresetToken"`
}

// GetConfigurationsResponse represents GetConfigurations response.
type GetConfigurationsResponse struct {
	XMLName          xml.Name              `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurationsResponse"`
	PTZConfiguration []PTZConfigurationExt `xml:"PTZConfiguration"`
}

// PTZConfigurationExt represents PTZ configuration with extensions.
type PTZConfigurationExt struct {
	Token         string         `xml:"token,attr"`
	Name          string         `xml:"Name"`
	UseCount      int            `xml:"UseCount"`
	NodeToken     string         `xml:"NodeToken"`
	PanTiltLimits *PanTiltLimits `xml:"PanTiltLimits,omitempty"`
	ZoomLimits    *ZoomLimits    `xml:"ZoomLimits,omitempty"`
}

// PanTiltLimits represents pan/tilt limits.
type PanTiltLimits struct {
	Range Space2DDescription `xml:"Range"`
}

// ZoomLimits represents zoom limits.
type ZoomLimits struct {
	Range Space1DDescription `xml:"Range"`
}

// Space2DDescription represents 2D space description.
type Space2DDescription struct {
	URI    string     `xml:"URI"`
	XRange FloatRange `xml:"XRange"`
	YRange FloatRange `xml:"YRange"`
}

// Space1DDescription represents 1D space description.
type Space1DDescription struct {
	URI    string     `xml:"URI"`
	XRange FloatRange `xml:"XRange"`
}

// FloatRange represents a float range.
type FloatRange struct {
	Min float64 `xml:"Min"`
	Max float64 `xml:"Max"`
}

// PTZ service handlers

var ptzMutex sync.RWMutex

// HandleContinuousMove handles ContinuousMove request.
func (s *Server) HandleContinuousMove(body interface{}) (interface{}, error) {
	var req ContinuousMoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get PTZ state
	ptzMutex.Lock()
	defer ptzMutex.Unlock()

	state, ok := s.ptzState[req.ProfileToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Set movement state
	state.Moving = true
	if req.Velocity.PanTilt != nil {
		state.PanMoving = req.Velocity.PanTilt.X != 0 || req.Velocity.PanTilt.Y != 0
		state.TiltMoving = state.PanMoving
	}
	if req.Velocity.Zoom != nil {
		state.ZoomMoving = req.Velocity.Zoom.X != 0
	}
	state.LastUpdate = time.Now()

	// In a real implementation, this would start a background task to
	// simulate movement and update position over time

	return &ContinuousMoveResponse{}, nil
}

// HandleAbsoluteMove handles AbsoluteMove request.
func (s *Server) HandleAbsoluteMove(body interface{}) (interface{}, error) {
	var req AbsoluteMoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get PTZ state
	ptzMutex.Lock()
	defer ptzMutex.Unlock()

	state, ok := s.ptzState[req.ProfileToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Update position
	if req.Position.PanTilt != nil {
		state.Position.Pan = req.Position.PanTilt.X
		state.Position.Tilt = req.Position.PanTilt.Y
	}
	if req.Position.Zoom != nil {
		state.Position.Zoom = req.Position.Zoom.X
	}

	// Set moving state temporarily
	state.Moving = true
	state.PanMoving = req.Position.PanTilt != nil
	state.TiltMoving = req.Position.PanTilt != nil
	state.ZoomMoving = req.Position.Zoom != nil
	state.LastUpdate = time.Now()

	// In a real implementation, simulate movement over time
	// For now, we'll stop immediately
	go func() {
		time.Sleep(500 * time.Millisecond) //nolint:mnd // PTZ movement delay
		ptzMutex.Lock()
		state.Moving = false
		state.PanMoving = false
		state.TiltMoving = false
		state.ZoomMoving = false
		ptzMutex.Unlock()
	}()

	return &AbsoluteMoveResponse{}, nil
}

// HandleRelativeMove handles RelativeMove request.
func (s *Server) HandleRelativeMove(body interface{}) (interface{}, error) {
	var req RelativeMoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get PTZ state
	ptzMutex.Lock()
	defer ptzMutex.Unlock()

	state, ok := s.ptzState[req.ProfileToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Update position relatively
	if req.Translation.PanTilt != nil {
		state.Position.Pan += req.Translation.PanTilt.X
		state.Position.Tilt += req.Translation.PanTilt.Y
	}
	if req.Translation.Zoom != nil {
		state.Position.Zoom += req.Translation.Zoom.X
	}

	// Clamp values to valid ranges (simplified)
	const maxPan = 180 // PTZ pan range
	const maxTilt = 90 // PTZ tilt range
	state.Position.Pan = clamp(state.Position.Pan, -maxPan, maxPan)
	state.Position.Tilt = clamp(state.Position.Tilt, -maxTilt, maxTilt)
	state.Position.Zoom = clamp(state.Position.Zoom, 0, 1)

	state.Moving = true
	state.LastUpdate = time.Now()

	// Simulate movement completion
	go func() {
		time.Sleep(500 * time.Millisecond) //nolint:mnd // PTZ movement delay
		ptzMutex.Lock()
		state.Moving = false
		state.PanMoving = false
		state.TiltMoving = false
		state.ZoomMoving = false
		ptzMutex.Unlock()
	}()

	return &RelativeMoveResponse{}, nil
}

// HandleStop handles Stop request.
func (s *Server) HandleStop(body interface{}) (interface{}, error) {
	var req StopRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get PTZ state
	ptzMutex.Lock()
	defer ptzMutex.Unlock()

	state, ok := s.ptzState[req.ProfileToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Stop movement
	if req.PanTilt {
		state.PanMoving = false
		state.TiltMoving = false
	}
	if req.Zoom {
		state.ZoomMoving = false
	}
	if !req.PanTilt && !req.Zoom {
		// Stop all if neither specified
		state.PanMoving = false
		state.TiltMoving = false
		state.ZoomMoving = false
	}
	state.Moving = state.PanMoving || state.TiltMoving || state.ZoomMoving
	state.LastUpdate = time.Now()

	return &StopResponse{}, nil
}

// HandleGetStatus handles GetStatus request.
func (s *Server) HandleGetStatus(body interface{}) (interface{}, error) {
	var req GetStatusRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get PTZ state
	ptzMutex.RLock()
	defer ptzMutex.RUnlock()

	state, ok := s.ptzState[req.ProfileToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Build status response
	status := &PTZStatus{
		Position: PTZVector{
			PanTilt: &Vector2D{
				X:     state.Position.Pan,
				Y:     state.Position.Tilt,
				Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace",
			},
			Zoom: &Vector1D{
				X:     state.Position.Zoom,
				Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace",
			},
		},
		MoveStatus: PTZMoveStatus{
			PanTilt: getMoveStatusString(state.PanMoving || state.TiltMoving),
			Zoom:    getMoveStatusString(state.ZoomMoving),
		},
		UTCTime: time.Now().UTC().Format(time.RFC3339),
	}

	return &GetStatusResponse{
		PTZStatus: status,
	}, nil
}

// HandleGetPresets handles GetPresets request.
func (s *Server) HandleGetPresets(body interface{}) (interface{}, error) {
	var req GetPresetsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Find the profile configuration
	var profileCfg *ProfileConfig
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Token == req.ProfileToken {
			profileCfg = &s.config.Profiles[i]

			break
		}
	}

	if profileCfg == nil || profileCfg.PTZ == nil {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Build presets response
	presets := make([]PTZPreset, len(profileCfg.PTZ.Presets))
	for i, preset := range profileCfg.PTZ.Presets {
		presets[i] = PTZPreset{
			Token: preset.Token,
			Name:  preset.Name,
			PTZPosition: &PTZVector{
				PanTilt: &Vector2D{
					X: preset.Position.Pan,
					Y: preset.Position.Tilt,
				},
				Zoom: &Vector1D{
					X: preset.Position.Zoom,
				},
			},
		}
	}

	return &GetPresetsResponse{
		Preset: presets,
	}, nil
}

// HandleGotoPreset handles GotoPreset request.
func (s *Server) HandleGotoPreset(body interface{}) (interface{}, error) {
	var req GotoPresetRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Find the profile configuration
	var profileCfg *ProfileConfig
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Token == req.ProfileToken {
			profileCfg = &s.config.Profiles[i]

			break
		}
	}

	if profileCfg == nil || profileCfg.PTZ == nil {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Find the preset
	var presetPos *PTZPosition
	for _, preset := range profileCfg.PTZ.Presets {
		if preset.Token == req.PresetToken {
			presetPos = &preset.Position

			break
		}
	}

	if presetPos == nil {
		return nil, fmt.Errorf("%w: %s", ErrPresetNotFound, req.PresetToken)
	}

	// Get PTZ state and move to preset
	ptzMutex.Lock()
	defer ptzMutex.Unlock()

	state := s.ptzState[req.ProfileToken]
	state.Position = *presetPos
	state.Moving = true
	state.PanMoving = true
	state.TiltMoving = true
	state.ZoomMoving = true
	state.LastUpdate = time.Now()

	// Simulate movement completion
	go func() {
		time.Sleep(1 * time.Second)
		ptzMutex.Lock()
		state.Moving = false
		state.PanMoving = false
		state.TiltMoving = false
		state.ZoomMoving = false
		ptzMutex.Unlock()
	}()

	return &GotoPresetResponse{}, nil
}

// Helper functions

func getMoveStatusString(moving bool) string {
	if moving {
		return "MOVING"
	}

	return "IDLE"
}

func clamp(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}

	return value
}
