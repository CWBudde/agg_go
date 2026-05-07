package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
)

func TestBezierDivRendersStatsTextOverlay(t *testing.T) {
	img := agg.NewImage(make([]uint8, width*height*4), width, height, width*4)

	newDemo().Render(img)

	goImg := img.ToGoImage()
	if goImg == nil {
		t.Fatal("expected rendered image")
	}

	var darkPixels int
	for y := 418; y < 454; y++ {
		for x := 35; x < 220; x++ {
			idx := goImg.PixOffset(x, y)
			r := goImg.Pix[idx]
			g := goImg.Pix[idx+1]
			b := goImg.Pix[idx+2]
			if r < 80 && g < 80 && b < 80 {
				darkPixels++
			}
		}
	}

	if darkPixels < 200 {
		t.Fatalf("expected black GSV stats text in overlay area, got %d dark pixels", darkPixels)
	}
}
