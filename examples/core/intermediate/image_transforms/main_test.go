package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
	icol "github.com/cwbudde/agg_go/internal/color"
)

func TestRunnerConfigEncodesLinearFramebuffer(t *testing.T) {
	cfg := runnerConfig(320, 300)
	if !cfg.FlipY {
		t.Fatal("image_transforms must run with FlipY=true to match C++ platform_support")
	}
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("image_transforms must encode the linear framebuffer back to sRGB for display")
	}
	if cfg.DisableLinearRGBToSRGB {
		t.Fatal("image_transforms should leave the default output path untouched instead of using the explicit opt-out flag")
	}
}

func TestLinearizeSRGBImageConvertsRGBAndPreservesAlpha(t *testing.T) {
	src := agg.NewImage([]byte{128, 64, 32, 200}, 1, 1, 4)
	got := linearizeSRGBImage(src)
	if got == nil {
		t.Fatal("linearizeSRGBImage returned nil")
	}

	want := icol.ConvertRGBA8SRGBToLinear(icol.RGBA8[icol.SRGB]{R: 128, G: 64, B: 32, A: 200})
	if got.Data[0] != want.R || got.Data[1] != want.G || got.Data[2] != want.B || got.Data[3] != want.A {
		t.Fatalf("linearized pixel = rgba(%d,%d,%d,%d), want rgba(%d,%d,%d,%d)",
			got.Data[0], got.Data[1], got.Data[2], got.Data[3], want.R, want.G, want.B, want.A)
	}
}
