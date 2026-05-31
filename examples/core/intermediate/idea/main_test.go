package main

import (
	"image"
	"testing"

	agg "github.com/cwbudde/agg_go"
)

func TestIdeaDemoRendersControlsWithCXXTextOrientation(t *testing.T) {
	const (
		width  = 250
		height = 280
	)
	img := agg.NewImage(make([]uint8, width*height*4), width, height, width*4)

	newDemo().Render(img)
	got := img.ToGoImage()
	if got == nil {
		t.Fatal("ToGoImage returned nil")
	}

	upper := countDarkPixels(got, image.Rect(20, 4, width-5, 10))
	lower := countDarkPixels(got, image.Rect(20, 10, width-5, 16))
	if lower <= upper {
		t.Fatalf("control label pixels look vertically flipped: upper=%d lower=%d", upper, lower)
	}
}

func TestIdeaConfigEncodesLinearOutput(t *testing.T) {
	cfg := demoConfig()

	if !cfg.EncodeLinearRGBToSRGB {
		t.Fatalf("EncodeLinearRGBToSRGB = false, want true for AGG_BGR24 linear output")
	}
}

func TestControlColorsStayLinearBeforeOutputEncoding(t *testing.T) {
	got := clampUnitToU8(0.3)
	if got != 77 {
		t.Fatalf("clampUnitToU8(0.3) = %d, want 77", got)
	}
}

func countDarkPixels(img image.Image, rect image.Rectangle) int {
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r16, g16, b16, a16 := img.At(x, y).RGBA()
			if a16 == 0 {
				continue
			}
			if uint8(r16>>8) < 100 && uint8(g16>>8) < 100 && uint8(b16>>8) < 100 {
				count++
			}
		}
	}
	return count
}
