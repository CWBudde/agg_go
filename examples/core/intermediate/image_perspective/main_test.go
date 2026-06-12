package main

import "testing"

func TestRunnerConfigMatchesCPPPipeline(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("image_perspective must run with FlipY=true to match C++ platform_support")
	}
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("image_perspective renders in linear space (AGG_BGRA32 color_type) and must encode linear->sRGB on save like the C++ platform")
	}
}

func TestRboxTextMatchesOriginalDefaults(t *testing.T) {
	d := newDemo()
	if got := d.transType.TextHeight(); got != 9.0 {
		t.Fatalf("image_perspective rbox text height = %v, want original AGG default 9.0", got)
	}
	if got := d.transType.TextThickness(); got != 1.5 {
		t.Fatalf("image_perspective rbox text thickness = %v, want original AGG default 1.5", got)
	}
}
