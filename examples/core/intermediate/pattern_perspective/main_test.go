package main

import (
	"testing"

	icol "github.com/cwbudde/agg_go/internal/color"
)

func TestRunnerConfigKeepsPatternPixelsRaw(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("pattern_perspective must run with FlipY=true to match C++ platform_support")
	}
	if cfg.EncodeLinearRGBToSRGB {
		t.Fatal("pattern_perspective must not post-encode the full framebuffer; the sampled pattern image is already in display byte space")
	}
	if cfg.DisableLinearRGBToSRGB {
		t.Fatal("pattern_perspective should leave the default output path untouched instead of using the explicit opt-out flag")
	}
}

func TestControlColorsAreEncodedForDisplay(t *testing.T) {
	got := toDisplayAggColor(icol.NewRGBA(0, 0.3, 0.5, 0.6))
	want := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{
		R: 0,
		G: 77,
		B: 128,
		A: 153,
	})
	if got.R != want.R || got.G != want.G || got.B != want.B || got.A != want.A {
		t.Fatalf("display control color = rgba(%d,%d,%d,%d), want rgba(%d,%d,%d,%d)",
			got.R, got.G, got.B, got.A, want.R, want.G, want.B, want.A)
	}
}

func TestRboxTextMatchesPatternPerspectiveSource(t *testing.T) {
	d := newDemo()
	if got := d.transType.TextHeight(); got != 8.0 {
		t.Fatalf("pattern_perspective rbox text height = %v, want C++ text_size 8.0", got)
	}
	if got := d.transType.TextThickness(); got != 1.0 {
		t.Fatalf("pattern_perspective rbox text thickness = %v, want C++ text_thickness 1.0", got)
	}
}
