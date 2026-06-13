package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/demo/timing"
)

func TestTimingTextDisabledOmitsAlphaMaskLabels(t *testing.T) {
	t.Setenv(timing.TextEnv, "0")

	img := agg.NewImage(make([]uint8, frameWidth*frameHeight*4), frameWidth, frameHeight, frameWidth*4)
	newDemo().Render(img)

	if dark := countDarkPixels(img, 245, 480, 470, 518); dark > 0 {
		t.Fatalf("expected no dark timing-label pixels with %s=0, got %d", timing.TextEnv, dark)
	}
}

func countDarkPixels(img *agg.Image, x1, y1, x2, y2 int) int {
	count := 0
	for y := y1; y < y2; y++ {
		row := y * img.Stride()
		for x := x1; x < x2; x++ {
			i := row + x*4
			r, g, b, a := img.Data[i], img.Data[i+1], img.Data[i+2], img.Data[i+3]
			if a > 240 && r < 16 && g < 16 && b < 16 {
				count++
			}
		}
	}
	return count
}
