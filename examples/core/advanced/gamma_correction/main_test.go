package main

import (
	"testing"

	icolor "github.com/cwbudde/agg_go/internal/color"
)

func TestSRGBA8DecodesToRendererLinearColor(t *testing.T) {
	got := srgba8(80, 127, 80, 255)
	want := icolor.ConvertRGBA8SRGBToLinear(icolor.RGBA8[icolor.SRGB]{
		R: 80,
		G: 127,
		B: 80,
		A: 255,
	})
	if got != want {
		t.Fatalf("srgba8(80,127,80,255) = %+v, want C++ srgba8 converted to renderer rgba8 %+v", got, want)
	}
}

func TestRGBA8KeepsCopyBarColorsRawLinear(t *testing.T) {
	got := rgba8(128, 128, 128, 255)
	want := icolor.RGBA8[icolor.Linear]{R: 128, G: 128, B: 128, A: 255}
	if got != want {
		t.Fatalf("rgba8(128,128,128,255) = %+v, want raw linear copy_bar color %+v", got, want)
	}
}

func TestRunnerConfigEncodesLinearOutputToSRGB(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("gamma_correction renders a linear AGG_BGR24-style buffer and must encode output to sRGB")
	}
	if cfg.DisableLinearRGBToSRGB {
		t.Fatal("gamma_correction must not disable linear-to-sRGB output encoding")
	}
}
