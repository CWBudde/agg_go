package main

import "testing"

func TestRunnerConfigKeepsImagePixelsRaw(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("image_perspective must run with FlipY=true to match C++ platform_support")
	}
	if cfg.EncodeLinearRGBToSRGB {
		t.Fatal("image_perspective must not post-encode the full framebuffer; the sampled source image is already in display byte space")
	}
	if cfg.DisableLinearRGBToSRGB {
		t.Fatal("image_perspective should leave the default output path untouched instead of using the explicit opt-out flag")
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
