package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestRunnerConfigUsesLinearPipeline(t *testing.T) {
	cfg := runnerConfig(&demo{w: 340, h: 360})
	if !cfg.FlipY {
		t.Fatal("image1 must run with FlipY=true to match C++ platform_support")
	}
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("image1 renders in linear space (sRGB->linear decoded source); the framebuffer must be sRGB-encoded at save")
	}
}

// TestClipBackgroundMatchesCPPColor verifies the clip/outside color matches
// C++ agg::rgba_pre(0, 0.4, 0, 0.5): premultiplied linear bytes (0, 51, 0, 128).
func TestClipBackgroundMatchesCPPColor(t *testing.T) {
	got := rgba8Pre(0, 0.4, 0, 0.5)
	want := color.RGBA8[color.Linear]{R: 0, G: 51, B: 0, A: 128}
	if got != want {
		t.Fatalf("clip background = %+v, want %+v", got, want)
	}
}
