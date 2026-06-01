package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestRunnerConfigKeepsImagePixelsRaw(t *testing.T) {
	cfg := runnerConfig(&demo{w: 340, h: 360})
	if !cfg.FlipY {
		t.Fatal("image1 must run with FlipY=true to match C++ platform_support")
	}
	if cfg.EncodeLinearRGBToSRGB {
		t.Fatal("image1 must not post-encode the full framebuffer; the sampled source image is already in display byte space")
	}
	if cfg.DisableLinearRGBToSRGB {
		t.Fatal("image1 should leave the default output path untouched instead of using the explicit opt-out flag")
	}
}

func TestClipBackgroundMatchesDisplayEncodedCPPColor(t *testing.T) {
	clip := displayPremulOverWhite(rgba8Pre(0, 0.4, 0, 0.5))
	got := color.RGBA8[color.Linear]{
		R: color.RGBA8Prelerp(255, clip.R, clip.A),
		G: color.RGBA8Prelerp(255, clip.G, clip.A),
		B: color.RGBA8Prelerp(255, clip.B, clip.A),
		A: 255,
	}
	want := color.ConvertToSRGBFromLinear(color.RGBA8[color.Linear]{
		R: color.RGBA8Prelerp(255, rgba8Pre(0, 0.4, 0, 0.5).R, clip.A),
		G: color.RGBA8Prelerp(255, rgba8Pre(0, 0.4, 0, 0.5).G, clip.A),
		B: color.RGBA8Prelerp(255, rgba8Pre(0, 0.4, 0, 0.5).B, clip.A),
		A: 255,
	})
	wantLinear := color.RGBA8[color.Linear](want)

	if got != wantLinear {
		t.Fatalf("clip background over white = %+v, want display-encoded %+v", got, want)
	}
}
