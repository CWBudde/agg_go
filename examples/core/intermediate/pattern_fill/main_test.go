package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestPatternFillConfigEncodesLinearOutput(t *testing.T) {
	cfg := demoConfig()

	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatalf("EncodeLinearRGBToSRGB = false, want true for AGG_BGRA32 linear output")
	}
}

func TestGeneratePatternConvertsSRGBMotifColorsToLinear(t *testing.T) {
	pf := generatePattern(int(defaultPatternSize), defaultPatternAngle, defaultPatternAlpha)
	fill := color.ConvertRGBA8SRGBToLinear(color.NewRGBA8[color.SRGB](110, 130, 50, 255))
	stroke := color.ConvertRGBA8SRGBToLinear(color.NewRGBA8[color.SRGB](0, 50, 80, 255))

	if !containsRGBA(pf, fill) {
		t.Fatalf("generated pattern does not contain linearized fill color %+v", fill)
	}
	if !containsRGBA(pf, stroke) {
		t.Fatalf("generated pattern does not contain linearized stroke color %+v", stroke)
	}
}

func containsRGBA(pf patternPixFmt, c color.RGBA8[color.Linear]) bool {
	for y := 0; y < pf.h; y++ {
		for x := 0; x < pf.w; x++ {
			off := y*pf.stride + x*4
			if pf.data[off+0] == c.R &&
				pf.data[off+1] == c.G &&
				pf.data[off+2] == c.B &&
				pf.data[off+3] == c.A {
				return true
			}
		}
	}
	return false
}
