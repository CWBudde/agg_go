package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestWASMGradientContourLUTConvertsSRGBStopsToLinear(t *testing.T) {
	lut := buildGradientContourLUT(2)
	if len(lut) == 0 {
		t.Fatal("empty LUT")
	}

	want := color.ConvertToLinear(color.RGBA8[color.SRGB]{
		R: 178,
		G: 34,
		B: 34,
		A: 255,
	})
	if lut[0] != want {
		t.Fatalf("first LUT color = %+v, want srgba8 stop converted to linear rgba8 %+v", lut[0], want)
	}
}

func TestWASMGradientContourOutputEncodesLinearBufferForDisplay(t *testing.T) {
	oldWidth, oldHeight := width, height
	oldCtx, oldCanvasBuf := ctx, canvasBuf
	oldPolygon := gradientsContourPolygon
	oldGradient := gradientsContourGradient
	oldReflect := gradientsContourReflect
	oldC1, oldC2 := gradientsContourC1, gradientsContourC2
	oldD1, oldD2 := gradientsContourD1, gradientsContourD2
	oldColors := gradientsContourColors
	defer func() {
		width, height = oldWidth, oldHeight
		ctx, canvasBuf = oldCtx, oldCanvasBuf
		gradientsContourPolygon = oldPolygon
		gradientsContourGradient = oldGradient
		gradientsContourReflect = oldReflect
		gradientsContourC1, gradientsContourC2 = oldC1, oldC2
		gradientsContourD1, gradientsContourD2 = oldD1, oldD2
		gradientsContourColors = oldColors
	}()

	width, height = 800, 600
	ctx = agg.NewContext(width, height)
	canvasBuf = ctx.GetImage().Data
	gradientsContourPolygon = 0
	gradientsContourGradient = 1
	gradientsContourReflect = true
	gradientsContourC1, gradientsContourC2 = 0, 512
	gradientsContourD1, gradientsContourD2 = 0, 100
	gradientsContourColors = 2

	drawGradientsContourDemo()

	var foundDisplayStop, foundLinearStop bool
	data := ctx.GetImage().Data
	for i := 0; i+3 < len(data); i += 4 {
		r, g, b := data[i], data[i+1], data[i+2]
		if r == 178 && g == 34 && b == 34 {
			foundDisplayStop = true
		}
		if r == 114 && g == 4 && b == 4 {
			foundLinearStop = true
		}
	}
	if !foundDisplayStop {
		t.Fatalf("rendered output did not contain display-encoded firebrick stop; found raw linear stop: %v", foundLinearStop)
	}
}
