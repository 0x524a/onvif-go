package onvif

import (
	"encoding/xml"
	"testing"
)

// A range with one bound missing is absent, not one whose missing bound is 0:
// <Brightness><Max>100</Max></Brightness> must not decode as 0..100.
func TestImagingRangeMissingABoundIsNil(t *testing.T) {
	cases := map[string]*imagingRangeXML{
		"nil":      nil,
		"empty":    {},
		"max only": {Max: ptzBoundXML{v: 100, set: true}},
		"min only": {Min: ptzBoundXML{v: 1, set: true}},
	}
	for name, r := range cases {
		if got := r.toFloatRange(); got != nil {
			t.Errorf("%s: toFloatRange() = %+v, want nil", name, got)
		}
	}
	both := &imagingRangeXML{Min: ptzBoundXML{v: 0, set: true}, Max: ptzBoundXML{v: 100, set: true}}
	if got := both.toFloatRange(); got == nil || got.Min != 0 || got.Max != 100 {
		t.Errorf("both bounds: toFloatRange() = %+v, want {0 100}", got)
	}
}

func TestImagingRangeDecodesHalfRangeAsNil(t *testing.T) {
	var x struct {
		B *imagingRangeXML `xml:"Brightness"`
	}
	if err := xml.Unmarshal([]byte(`<R><Brightness><Max>100</Max></Brightness></R>`), &x); err != nil {
		t.Fatal(err)
	}
	if got := x.B.toFloatRange(); got != nil {
		t.Errorf("half range decoded as %+v, want nil", got)
	}
	if err := xml.Unmarshal([]byte(`<R><Brightness><Min>0</Min><Max>100</Max></Brightness></R>`), &x); err != nil {
		t.Fatal(err)
	}
	if got := x.B.toFloatRange(); got == nil || got.Max != 100 {
		t.Errorf("full range decoded as %+v, want {0 100}", got)
	}
}
