package main

import (
	"testing"

	icol "github.com/cwbudde/agg_go/internal/color"
)

func TestRunnerConfigMatchesCPPPipeline(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("pattern_perspective must run with FlipY=true to match C++ platform_support")
	}
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("pattern_perspective renders in linear space (AGG_BGR24 color_type) and must encode linear->sRGB on save like the C++ platform")
	}
}

func TestControlColorsArePlainQuantized(t *testing.T) {
	// C++ render_ctrl with a linear color_type quantizes rgba floats with a
	// plain *255, no colorspace conversion.
	got := ctrlColor(icol.NewRGBA(0, 0.3, 0.5, 0.6))
	want := icol.RGBA8[icol.Linear]{R: 0, G: 77, B: 128, A: 153}
	if got != want {
		t.Fatalf("control color = rgba(%d,%d,%d,%d), want rgba(%d,%d,%d,%d)",
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
