package server

import (
	"encoding/xml"
	"fmt"
	"sync"
)

// Imaging service SOAP message types

// GetImagingSettingsRequest represents GetImagingSettings request.
type GetImagingSettingsRequest struct {
	XMLName          xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl GetImagingSettings"`
	VideoSourceToken string   `xml:"VideoSourceToken"`
}

// GetImagingSettingsResponse represents GetImagingSettings response.
type GetImagingSettingsResponse struct {
	XMLName         xml.Name         `xml:"http://www.onvif.org/ver20/imaging/wsdl GetImagingSettingsResponse"`
	ImagingSettings *ImagingSettings `xml:"ImagingSettings"`
}

// ImagingSettings represents imaging settings.
type ImagingSettings struct {
	BacklightCompensation *BacklightCompensationSettings `xml:"BacklightCompensation,omitempty"`
	Brightness            *float64                       `xml:"Brightness,omitempty"`
	ColorSaturation       *float64                       `xml:"ColorSaturation,omitempty"`
	Contrast              *float64                       `xml:"Contrast,omitempty"`
	Exposure              *ExposureSettings20            `xml:"Exposure,omitempty"`
	Focus                 *FocusConfiguration20          `xml:"Focus,omitempty"`
	IrCutFilter           *string                        `xml:"IrCutFilter,omitempty"`
	Sharpness             *float64                       `xml:"Sharpness,omitempty"`
	WideDynamicRange      *WideDynamicRangeSettings      `xml:"WideDynamicRange,omitempty"`
	WhiteBalance          *WhiteBalanceSettings20        `xml:"WhiteBalance,omitempty"`
}

// BacklightCompensationSettings represents backlight compensation settings.
type BacklightCompensationSettings struct {
	Mode  string   `xml:"Mode"`
	Level *float64 `xml:"Level,omitempty"`
}

// ExposureSettings20 represents exposure settings for ONVIF 2.0.
type ExposureSettings20 struct {
	Mode            string     `xml:"Mode"`
	Priority        *string    `xml:"Priority,omitempty"`
	Window          *Rectangle `xml:"Window,omitempty"`
	MinExposureTime *float64   `xml:"MinExposureTime,omitempty"`
	MaxExposureTime *float64   `xml:"MaxExposureTime,omitempty"`
	MinGain         *float64   `xml:"MinGain,omitempty"`
	MaxGain         *float64   `xml:"MaxGain,omitempty"`
	MinIris         *float64   `xml:"MinIris,omitempty"`
	MaxIris         *float64   `xml:"MaxIris,omitempty"`
	ExposureTime    *float64   `xml:"ExposureTime,omitempty"`
	Gain            *float64   `xml:"Gain,omitempty"`
	Iris            *float64   `xml:"Iris,omitempty"`
}

// FocusConfiguration20 represents focus configuration for ONVIF 2.0.
type FocusConfiguration20 struct {
	AutoFocusMode string   `xml:"AutoFocusMode"`
	DefaultSpeed  *float64 `xml:"DefaultSpeed,omitempty"`
	NearLimit     *float64 `xml:"NearLimit,omitempty"`
	FarLimit      *float64 `xml:"FarLimit,omitempty"`
}

// WideDynamicRangeSettings represents WDR settings.
type WideDynamicRangeSettings struct {
	Mode  string   `xml:"Mode"`
	Level *float64 `xml:"Level,omitempty"`
}

// WhiteBalanceSettings20 represents white balance settings for ONVIF 2.0.
type WhiteBalanceSettings20 struct {
	Mode   string   `xml:"Mode"`
	CrGain *float64 `xml:"CrGain,omitempty"`
	CbGain *float64 `xml:"CbGain,omitempty"`
}

// Rectangle represents a rectangle.
type Rectangle struct {
	Bottom float64 `xml:"bottom,attr"`
	Top    float64 `xml:"top,attr"`
	Right  float64 `xml:"right,attr"`
	Left   float64 `xml:"left,attr"`
}

// SetImagingSettingsRequest represents SetImagingSettings request.
type SetImagingSettingsRequest struct {
	XMLName          xml.Name         `xml:"http://www.onvif.org/ver20/imaging/wsdl SetImagingSettings"`
	VideoSourceToken string           `xml:"VideoSourceToken"`
	ImagingSettings  *ImagingSettings `xml:"ImagingSettings"`
	ForcePersistence bool             `xml:"ForcePersistence,omitempty"`
}

// SetImagingSettingsResponse represents SetImagingSettings response.
type SetImagingSettingsResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl SetImagingSettingsResponse"`
}

// GetOptionsRequest represents GetOptions request.
type GetOptionsRequest struct {
	XMLName          xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl GetOptions"`
	VideoSourceToken string   `xml:"VideoSourceToken"`
}

// GetOptionsResponse represents GetOptions response.
type GetOptionsResponse struct {
	XMLName        xml.Name        `xml:"http://www.onvif.org/ver20/imaging/wsdl GetOptionsResponse"`
	ImagingOptions *ImagingOptions `xml:"ImagingOptions"`
}

// ImagingOptions represents imaging options/capabilities.
type ImagingOptions struct {
	BacklightCompensation *BacklightCompensationOptions `xml:"BacklightCompensation,omitempty"`
	Brightness            *FloatRange                   `xml:"Brightness,omitempty"`
	ColorSaturation       *FloatRange                   `xml:"ColorSaturation,omitempty"`
	Contrast              *FloatRange                   `xml:"Contrast,omitempty"`
	Exposure              *ExposureOptions              `xml:"Exposure,omitempty"`
	Focus                 *FocusOptions                 `xml:"Focus,omitempty"`
	IrCutFilterModes      []string                      `xml:"IrCutFilterModes,omitempty"`
	Sharpness             *FloatRange                   `xml:"Sharpness,omitempty"`
	WideDynamicRange      *WideDynamicRangeOptions      `xml:"WideDynamicRange,omitempty"`
	WhiteBalance          *WhiteBalanceOptions          `xml:"WhiteBalance,omitempty"`
}

// BacklightCompensationOptions represents backlight compensation options.
type BacklightCompensationOptions struct {
	Mode  []string    `xml:"Mode"`
	Level *FloatRange `xml:"Level,omitempty"`
}

// ExposureOptions represents exposure options.
type ExposureOptions struct {
	Mode            []string    `xml:"Mode"`
	Priority        []string    `xml:"Priority,omitempty"`
	MinExposureTime *FloatRange `xml:"MinExposureTime,omitempty"`
	MaxExposureTime *FloatRange `xml:"MaxExposureTime,omitempty"`
	MinGain         *FloatRange `xml:"MinGain,omitempty"`
	MaxGain         *FloatRange `xml:"MaxGain,omitempty"`
	MinIris         *FloatRange `xml:"MinIris,omitempty"`
	MaxIris         *FloatRange `xml:"MaxIris,omitempty"`
	ExposureTime    *FloatRange `xml:"ExposureTime,omitempty"`
	Gain            *FloatRange `xml:"Gain,omitempty"`
	Iris            *FloatRange `xml:"Iris,omitempty"`
}

// FocusOptions represents focus options.
type FocusOptions struct {
	AutoFocusModes []string    `xml:"AutoFocusModes"`
	DefaultSpeed   *FloatRange `xml:"DefaultSpeed,omitempty"`
	NearLimit      *FloatRange `xml:"NearLimit,omitempty"`
	FarLimit       *FloatRange `xml:"FarLimit,omitempty"`
}

// WideDynamicRangeOptions represents WDR options.
type WideDynamicRangeOptions struct {
	Mode  []string    `xml:"Mode"`
	Level *FloatRange `xml:"Level,omitempty"`
}

// WhiteBalanceOptions represents white balance options.
type WhiteBalanceOptions struct {
	Mode   []string    `xml:"Mode"`
	YrGain *FloatRange `xml:"YrGain,omitempty"`
	YbGain *FloatRange `xml:"YbGain,omitempty"`
}

// MoveRequest represents Move (focus) request.
type MoveRequest struct {
	XMLName          xml.Name   `xml:"http://www.onvif.org/ver20/imaging/wsdl Move"`
	VideoSourceToken string     `xml:"VideoSourceToken"`
	Focus            *FocusMove `xml:"Focus"`
}

// FocusMove represents focus move parameters.
type FocusMove struct {
	Absolute   *AbsoluteFocus   `xml:"Absolute,omitempty"`
	Relative   *RelativeFocus   `xml:"Relative,omitempty"`
	Continuous *ContinuousFocus `xml:"Continuous,omitempty"`
}

// AbsoluteFocus represents absolute focus.
type AbsoluteFocus struct {
	Position float64  `xml:"Position"`
	Speed    *float64 `xml:"Speed,omitempty"`
}

// RelativeFocus represents relative focus.
type RelativeFocus struct {
	Distance float64  `xml:"Distance"`
	Speed    *float64 `xml:"Speed,omitempty"`
}

// ContinuousFocus represents continuous focus.
type ContinuousFocus struct {
	Speed float64 `xml:"Speed"`
}

// MoveResponse represents Move response.
type MoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl MoveResponse"`
}

// Imaging service handlers

var imagingMutex sync.RWMutex

// HandleGetImagingSettings handles GetImagingSettings request.
func (s *Server) HandleGetImagingSettings(body interface{}) (interface{}, error) {
	var req GetImagingSettingsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get imaging state
	imagingMutex.RLock()
	defer imagingMutex.RUnlock()

	state, ok := s.imagingState[req.VideoSourceToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrVideoSourceNotFound, req.VideoSourceToken)
	}

	// Build imaging settings response
	settings := &ImagingSettings{
		Brightness:      &state.Brightness,
		ColorSaturation: &state.Saturation,
		Contrast:        &state.Contrast,
		Sharpness:       &state.Sharpness,
		IrCutFilter:     &state.IrCutFilter,
		BacklightCompensation: &BacklightCompensationSettings{
			Mode:  state.BacklightComp.Mode,
			Level: &state.BacklightComp.Level,
		},
		Exposure: &ExposureSettings20{
			Mode:            state.Exposure.Mode,
			Priority:        &state.Exposure.Priority,
			MinExposureTime: &state.Exposure.MinExposure,
			MaxExposureTime: &state.Exposure.MaxExposure,
			MinGain:         &state.Exposure.MinGain,
			MaxGain:         &state.Exposure.MaxGain,
			ExposureTime:    &state.Exposure.ExposureTime,
			Gain:            &state.Exposure.Gain,
		},
		Focus: &FocusConfiguration20{
			AutoFocusMode: state.Focus.AutoFocusMode,
			DefaultSpeed:  &state.Focus.DefaultSpeed,
			NearLimit:     &state.Focus.NearLimit,
			FarLimit:      &state.Focus.FarLimit,
		},
		WideDynamicRange: &WideDynamicRangeSettings{
			Mode:  state.WideDynamicRange.Mode,
			Level: &state.WideDynamicRange.Level,
		},
		WhiteBalance: &WhiteBalanceSettings20{
			Mode:   state.WhiteBalance.Mode,
			CrGain: &state.WhiteBalance.CrGain,
			CbGain: &state.WhiteBalance.CbGain,
		},
	}

	return &GetImagingSettingsResponse{
		ImagingSettings: settings,
	}, nil
}

// HandleSetImagingSettings handles SetImagingSettings request.
//
//nolint:gocyclo // SetImagingSettings has high complexity due to multiple validation and update paths
func (s *Server) HandleSetImagingSettings(body interface{}) (interface{}, error) {
	var req SetImagingSettingsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get imaging state
	imagingMutex.Lock()
	defer imagingMutex.Unlock()

	state, ok := s.imagingState[req.VideoSourceToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrVideoSourceNotFound, req.VideoSourceToken)
	}

	// Update settings
	settings := req.ImagingSettings
	if settings == nil {
		// Return success if no settings to update
		return &SetImagingSettingsResponse{}, nil
	}
	if settings.Brightness != nil {
		state.Brightness = *settings.Brightness
	}
	if settings.ColorSaturation != nil {
		state.Saturation = *settings.ColorSaturation
	}
	if settings.Contrast != nil {
		state.Contrast = *settings.Contrast
	}
	if settings.Sharpness != nil {
		state.Sharpness = *settings.Sharpness
	}
	if settings.IrCutFilter != nil {
		state.IrCutFilter = *settings.IrCutFilter
	}
	if settings.BacklightCompensation != nil {
		state.BacklightComp.Mode = settings.BacklightCompensation.Mode
		if settings.BacklightCompensation.Level != nil {
			state.BacklightComp.Level = *settings.BacklightCompensation.Level
		}
	}
	if settings.Exposure != nil {
		state.Exposure.Mode = settings.Exposure.Mode
		if settings.Exposure.Priority != nil {
			state.Exposure.Priority = *settings.Exposure.Priority
		}
		if settings.Exposure.ExposureTime != nil {
			state.Exposure.ExposureTime = *settings.Exposure.ExposureTime
		}
		if settings.Exposure.Gain != nil {
			state.Exposure.Gain = *settings.Exposure.Gain
		}
	}
	if settings.Focus != nil {
		state.Focus.AutoFocusMode = settings.Focus.AutoFocusMode
	}
	if settings.WideDynamicRange != nil {
		state.WideDynamicRange.Mode = settings.WideDynamicRange.Mode
		if settings.WideDynamicRange.Level != nil {
			state.WideDynamicRange.Level = *settings.WideDynamicRange.Level
		}
	}
	if settings.WhiteBalance != nil {
		state.WhiteBalance.Mode = settings.WhiteBalance.Mode
		if settings.WhiteBalance.CrGain != nil {
			state.WhiteBalance.CrGain = *settings.WhiteBalance.CrGain
		}
		if settings.WhiteBalance.CbGain != nil {
			state.WhiteBalance.CbGain = *settings.WhiteBalance.CbGain
		}
	}

	return &SetImagingSettingsResponse{}, nil
}

// HandleGetOptions handles GetOptions request.
func (s *Server) HandleGetOptions(body interface{}) (interface{}, error) {
	// Return available imaging options/capabilities
	const maxImagingValue = 100   // Maximum imaging parameter value
	const maxExposureTime = 10000 // Maximum exposure time in microseconds
	options := &ImagingOptions{
		Brightness:       &FloatRange{Min: 0, Max: maxImagingValue},
		ColorSaturation:  &FloatRange{Min: 0, Max: maxImagingValue},
		Contrast:         &FloatRange{Min: 0, Max: maxImagingValue},
		Sharpness:        &FloatRange{Min: 0, Max: maxImagingValue},
		IrCutFilterModes: []string{"ON", "OFF", "AUTO"},
		BacklightCompensation: &BacklightCompensationOptions{
			Mode:  []string{"OFF", "ON"},
			Level: &FloatRange{Min: 0, Max: maxImagingValue},
		},
		Exposure: &ExposureOptions{
			Mode:            []string{"AUTO", "MANUAL"},
			Priority:        []string{"LowNoise", "FrameRate"},
			MinExposureTime: &FloatRange{Min: 1, Max: maxExposureTime},
			MaxExposureTime: &FloatRange{Min: 1, Max: maxExposureTime},
			MinGain:         &FloatRange{Min: 0, Max: maxImagingValue},
			MaxGain:         &FloatRange{Min: 0, Max: maxImagingValue},
			ExposureTime:    &FloatRange{Min: 1, Max: maxExposureTime},
			Gain:            &FloatRange{Min: 0, Max: maxImagingValue},
		},
		Focus: &FocusOptions{
			AutoFocusModes: []string{"AUTO", "MANUAL"},
			DefaultSpeed:   &FloatRange{Min: 0, Max: 1},
			NearLimit:      &FloatRange{Min: 0, Max: 1},
			FarLimit:       &FloatRange{Min: 0, Max: 1},
		},
		WideDynamicRange: &WideDynamicRangeOptions{
			Mode:  []string{"OFF", "ON"},
			Level: &FloatRange{Min: 0, Max: 100}, //nolint:mnd // Imaging parameter range
		},
		WhiteBalance: &WhiteBalanceOptions{
			Mode:   []string{"AUTO", "MANUAL"},
			YrGain: &FloatRange{Min: 0, Max: 255}, //nolint:mnd // White balance gain range
			YbGain: &FloatRange{Min: 0, Max: 255}, //nolint:mnd // White balance gain range
		},
	}

	return &GetOptionsResponse{
		ImagingOptions: options,
	}, nil
}

// HandleMove handles Move (focus) request.
func (s *Server) HandleMove(body interface{}) (interface{}, error) {
	var req MoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get imaging state
	imagingMutex.Lock()
	defer imagingMutex.Unlock()

	state, ok := s.imagingState[req.VideoSourceToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrVideoSourceNotFound, req.VideoSourceToken)
	}

	// Process focus move
	if req.Focus != nil {
		if req.Focus.Absolute != nil {
			state.Focus.CurrentPos = req.Focus.Absolute.Position
		} else if req.Focus.Relative != nil {
			state.Focus.CurrentPos += req.Focus.Relative.Distance
			// Clamp to valid range
			if state.Focus.CurrentPos < 0 {
				state.Focus.CurrentPos = 0
			} else if state.Focus.CurrentPos > 1 {
				state.Focus.CurrentPos = 1
			}
		}
		// Continuous focus would start a background task
	}

	return &MoveResponse{}, nil
}
