package main

import (
	"testing"

	icol "github.com/cwbudde/agg_go/internal/color"
)

func TestRunnerConfigKeepsPatternResamplePixelsRaw(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("pattern_resample must run with FlipY=true to match C++ platform_support")
	}
	if cfg.EncodeLinearRGBToSRGB {
		t.Fatal("pattern_resample must not post-encode the full framebuffer; gamma handling is part of the demo")
	}
	if cfg.DisableLinearRGBToSRGB {
		t.Fatal("pattern_resample should leave the default output path untouched instead of using the explicit opt-out flag")
	}
}

func TestControlColorsAreEncodedForDisplay(t *testing.T) {
	got := toDisplayAggColor(icol.NewRGBA(0.8, 0, 0, 0.6))
	want := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{
		R: 204,
		G: 0,
		B: 0,
		A: 153,
	})
	if got.R != want.R || got.G != want.G || got.B != want.B || got.A != want.A {
		t.Fatalf("display control color = rgba(%d,%d,%d,%d), want rgba(%d,%d,%d,%d)",
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
