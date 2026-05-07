package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestWASMGradientContourLUTConvertsSRGBStopsToLinear(t *testing.T) {
	lut := buildGradientContourLUT(2)
	if len(lut) == 0 {
		t.Fatal("empty LUT")
	}

	want := color.ConvertToLinear(color.RGBA8[color.SRGB]{
		R: 178,
		G: 34,
		B: 34,
		A: 255,
	})
	if lut[0] != want {
		t.Fatalf("first LUT color = %+v, want srgba8 stop converted to linear rgba8 %+v", lut[0], want)
	}
}
