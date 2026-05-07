package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
)

func TestConvDashMarkerControlsRespectFlipYOutput(t *testing.T) {
	stride := -frameWidth * 4
	img := agg.NewImage(make([]uint8, frameWidth*frameHeight*4), frameWidth, frameHeight, stride)

	newDemo().Render(img)

	goImg := img.ToGoImage()
	if goImg == nil {
		t.Fatal("expected rendered image")
	}

	var redPixels, ySum int
	for y := 240; y < 320; y++ {
		for x := 10; x < 36; x++ {
			idx := goImg.PixOffset(x, y)
			r := goImg.Pix[idx]
			g := goImg.Pix[idx+1]
			b := goImg.Pix[idx+2]
			if r > 50 && g < 35 && b < 35 {
				redPixels++
				ySum += y
			}
		}
	}

	if redPixels == 0 {
		t.Fatal("expected active radio marker pixels")
	}
	centroidY := float64(ySum) / float64(redPixels)
	if centroidY < 298 {
		t.Fatalf("active radio marker is vertically flipped: centroidY=%.1f, want bottom control item", centroidY)
	}
}
