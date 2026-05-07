package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestWASMGammaCorrectionSRGBA8DecodesToLinear(t *testing.T) {
	got := gammaCorrectionSRGBA8(80, 127, 80, 255)
	want := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{
		R: 80,
		G: 127,
		B: 80,
		A: 255,
	})
	if got != want {
		t.Fatalf("gammaCorrectionSRGBA8(80,127,80,255) = %+v, want C++ srgba8 converted to renderer rgba8 %+v", got, want)
	}
}
