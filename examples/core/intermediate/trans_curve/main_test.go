package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
)

// TestTransCurveRendersLayeredContent is a bounded render smoke test: it verifies
// the single-path demo produces the expected layered scene — GSV text glyphs
// (black), the orange guide spline (rgba 170,50,20), and a white background —
// rather than a blank or fully-filled frame.
//
// C++ reference: ../agg-2.6/agg-src/examples/trans_curve1.cpp.
func TestTransCurveRendersLayeredContent(t *testing.T) {
	img := agg.NewImage(make([]uint8, width*height*4), width, height, width*4)

	(&demo{}).Render(img)

	goImg := img.ToGoImage()
	if goImg == nil {
		t.Fatal("expected rendered image")
	}

	var blackPx, guidePx, whitePx int
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := goImg.PixOffset(x, y)
			r, g, b := goImg.Pix[idx], goImg.Pix[idx+1], goImg.Pix[idx+2]
			switch {
			case r < 40 && g < 40 && b < 40:
				blackPx++
			case r > 120 && g >= 20 && g < 110 && b < 70:
				guidePx++
			case r > 240 && g > 240 && b > 240:
				whitePx++
			}
		}
	}

	if blackPx == 0 {
		t.Error("expected black GSV text glyph pixels along the curve, got none")
	}
	if guidePx == 0 {
		t.Error("expected orange guide-spline pixels (rgba 170,50,20), got none")
	}
	if whitePx == 0 {
		t.Error("expected white background pixels, got none (frame fully filled?)")
	}
}
