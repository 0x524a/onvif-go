package onvif

import "testing"

func TestConvertToPTZVectorXML(t *testing.T) {
	if got := convertToPTZVectorXML(nil); got != nil {
		t.Errorf("convertToPTZVectorXML(nil) = %+v, want nil", got)
	}

	v := &PTZVector{
		PanTilt: &Vector2D{X: 0.1, Y: -0.2, Space: "PanTiltSpace"},
		Zoom:    &Vector1D{X: 0.5, Space: "ZoomSpace"},
	}
	got := convertToPTZVectorXML(v)
	if got == nil {
		t.Fatal("convertToPTZVectorXML() = nil, want non-nil")
	}
	if got.PanTilt == nil || got.PanTilt.X != 0.1 || got.PanTilt.Y != -0.2 || got.PanTilt.Space != "PanTiltSpace" {
		t.Errorf("PanTilt = %+v, want {0.1 -0.2 PanTiltSpace}", got.PanTilt)
	}
	if got.Zoom == nil || got.Zoom.X != 0.5 || got.Zoom.Space != "ZoomSpace" {
		t.Errorf("Zoom = %+v, want {0.5 ZoomSpace}", got.Zoom)
	}

	partial := convertToPTZVectorXML(&PTZVector{PanTilt: &Vector2D{X: 1, Y: 2}})
	if partial.PanTilt == nil {
		t.Error("PanTilt = nil, want non-nil")
	}
	if partial.Zoom != nil {
		t.Errorf("Zoom = %+v, want nil when source Zoom is nil", partial.Zoom)
	}
}

func TestConvertToPTZSpeedXML(t *testing.T) {
	if got := convertToPTZSpeedXML(nil); got != nil {
		t.Errorf("convertToPTZSpeedXML(nil) = %+v, want nil", got)
	}

	s := &PTZSpeed{
		PanTilt: &Vector2D{X: 0.3, Y: 0.4, Space: "PanTiltSpeedSpace"},
		Zoom:    &Vector1D{X: 0.6, Space: "ZoomSpeedSpace"},
	}
	got := convertToPTZSpeedXML(s)
	if got == nil {
		t.Fatal("convertToPTZSpeedXML() = nil, want non-nil")
	}
	if got.PanTilt == nil || got.PanTilt.X != 0.3 || got.PanTilt.Y != 0.4 || got.PanTilt.Space != "PanTiltSpeedSpace" {
		t.Errorf("PanTilt = %+v, want {0.3 0.4 PanTiltSpeedSpace}", got.PanTilt)
	}
	if got.Zoom == nil || got.Zoom.X != 0.6 || got.Zoom.Space != "ZoomSpeedSpace" {
		t.Errorf("Zoom = %+v, want {0.6 ZoomSpeedSpace}", got.Zoom)
	}

	partial := convertToPTZSpeedXML(&PTZSpeed{Zoom: &Vector1D{X: 9}})
	if partial.Zoom == nil {
		t.Error("Zoom = nil, want non-nil")
	}
	if partial.PanTilt != nil {
		t.Errorf("PanTilt = %+v, want nil when source PanTilt is nil", partial.PanTilt)
	}
}
