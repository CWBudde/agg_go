package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
)

// TestTransCurve2RendersTextNotLion is a bounded render smoke test for the
// double-path demo. It guards the faithfulness fix: trans_curve2 must render the
// GSV text paragraph warped between the two rails (matching C++ and the web
// demo), not the lion. It checks for black text glyphs and the orange guide
// rails over a white background.
//
// C++ reference: ../agg-2.6/agg-src/examples/trans_curve2.cpp.
func TestTransCurve2RendersTextNotLion(t *testing.T) {
	img := agg.NewImage(make([]uint8, width*height*4), width, height, width*4)

	(&demo{}).Render(img)

	goImg := img.ToGoImage()
	if goImg == nil {
		t.Fatal("expected rendered image")
	}

	var blackPx, guidePx, whitePx, coloredPx int
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
			default:
				// Saturated mid-tone fills would indicate the old lion render.
				if r > 60 && g > 60 && b < 60 {
					coloredPx++
				}
			}
		}
	}

	if blackPx == 0 {
		t.Error("expected black GSV text glyph pixels between the rails, got none")
	}
	if guidePx == 0 {
		t.Error("expected orange guide-rail pixels (rgba 170,50,20), got none")
	}
	if whitePx == 0 {
		t.Error("expected white background pixels, got none (frame fully filled?)")
	}
	// The text scene is overwhelmingly black-on-white; the removed lion render
	// filled large areas with saturated yellow/brown body color. A small count
	// is fine (kerning AA), but the lion would dominate.
	if coloredPx > blackPx {
		t.Errorf("scene looks like the lion fill, not text: coloredPx=%d > blackPx=%d", coloredPx, blackPx)
	}
}
