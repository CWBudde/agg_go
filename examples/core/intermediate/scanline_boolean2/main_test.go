package main

import "testing"

func TestRunnerConfigMatchesCPPFlipYAndSRGBExport(t *testing.T) {
	cfg := runnerConfig()
	if !cfg.FlipY {
		t.Fatal("scanline_boolean2 must run with FlipY=true to match C++ platform_support")
	}
	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatal("scanline_boolean2 must encode linear framebuffer output to sRGB for visual parity")
	}
}
