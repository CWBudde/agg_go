package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
	icol "github.com/cwbudde/agg_go/internal/color"
)

func TestSRGBA8ColorStoresLinearValueForEncodedOutput(t *testing.T) {
	got := srgba8Color(127, 127, 127, 255)
	want := icol.ConvertRGBA8SRGBToLinear(icol.RGBA8[icol.SRGB]{
		R: 127,
		G: 127,
		B: 127,
		A: 255,
	})

	if got != agg.NewColor(want.R, want.G, want.B, want.A) {
		t.Fatalf("srgba8Color(127,127,127,255) = rgba(%d,%d,%d,%d), want linear rgba(%d,%d,%d,%d)",
			got.R, got.G, got.B, got.A,
			want.R, want.G, want.B, want.A)
	}
}
