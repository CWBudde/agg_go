package main

import (
	"testing"

	icol "github.com/cwbudde/agg_go/internal/color"
)

func TestRunnerConfigMatchesCPPPipeline(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("pattern_resample must run with FlipY=true to match C++ platform_support")
	}
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("pattern_resample renders in linear space (AGG_BGR24 color_type) and must encode linear->sRGB on save like the C++ platform")
	}
}

func TestControlColorsArePlainQuantized(t *testing.T) {
	// C++ render_ctrl with a linear color_type quantizes rgba floats with a
	// plain *255, no colorspace conversion. The quad ghost uses alpha 0.1.
	got := ctrlColor(icol.NewRGBA(0, 0.3, 0.5, 0.1))
	want := icol.RGBA8[icol.Linear]{R: 0, G: 77, B: 128, A: 26}
	if got != want {
		t.Fatalf("control color = rgba(%d,%d,%d,%d), want rgba(%d,%d,%d,%d)",
			got.R, got.G, got.B, got.A, want.R, want.G, want.B, want.A)
	}
}

func TestControlsMatchPatternResampleSourceDefaults(t *testing.T) {
	d := newDemo()
	if got := d.transType.TextHeight(); got != 7.0 {
		t.Fatalf("pattern_resample rbox text height = %v, want C++ text_size 7.0", got)
	}
	if got := d.transType.CurItem(); got != 4 {
		t.Fatalf("pattern_resample rbox current item = %v, want C++ default 4", got)
	}
	if got := d.gamma.Value(); got != 2.0 {
		t.Fatalf("pattern_resample gamma default = %v, want C++ default 2.0", got)
	}
	if got := d.blur.Value(); got != 1.0 {
		t.Fatalf("pattern_resample blur default = %v, want C++ default 1.0", got)
	}
}

func TestGammaLUTMatchesAGGGammaLut(t *testing.T) {
	d := newDemo()
	// agg::gamma_lut<int8u,int8u,8,8>(2.0): dir[i]=uround(pow(i/255,2)*255),
	// inv[i]=uround(pow(i/255,0.5)*255).
	if got := d.gammaLut.Dir(128); got != 64 {
		t.Fatalf("dir(128) = %d, want 64", got)
	}
	if got := d.gammaLut.Inv(64); got != 128 {
		t.Fatalf("inv(64) = %d, want 128", got)
	}
	if got := d.gammaLut.Dir(255); got != 255 {
		t.Fatalf("dir(255) = %d, want 255", got)
	}
}
