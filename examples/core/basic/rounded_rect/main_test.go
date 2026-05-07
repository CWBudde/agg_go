package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	agg "github.com/cwbudde/agg_go"
)

func TestRoundedRectDemoRendersControlsInCXXFlipYPosition(t *testing.T) {
	stride := -demoWidth * 4
	img := agg.NewImage(make([]uint8, demoWidth*demoHeight*4), demoWidth, demoHeight, stride)

	newDemo().Render(img)
	got := img.ToGoImage()
	if got == nil {
		t.Fatal("ToGoImage returned nil")
	}

	if countNonWhite(got, image.Rect(0, demoHeight-60, demoWidth, demoHeight)) < 100 {
		t.Fatalf("expected controls in bottom band for flip_y=true output")
	}
}

func TestRoundedRectDemoSavedPNGUsesCXXControlGray(t *testing.T) {
	t.Chdir(t.TempDir())
	main()

	f, err := os.Open(filepath.Join(".", "rounded_rectangle.png"))
	if err != nil {
		t.Fatalf("open output PNG: %v", err)
	}
	defer f.Close()

	got, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output PNG: %v", err)
	}

	if countNeutralGray(got, image.Rect(0, demoHeight-60, demoWidth, demoHeight), 110, 145) < 50 {
		t.Fatalf("expected C++ srgba8(127,127,127) control gray in saved PNG")
	}
}

func countNonWhite(img image.Image, rect image.Rectangle) int {
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a != 0 && (r < 0xff00 || g < 0xff00 || b < 0xff00) {
				count++
			}
		}
	}
	return count
}

func countNeutralGray(img image.Image, rect image.Rectangle, minValue, maxValue uint8) int {
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r16, g16, b16, a16 := img.At(x, y).RGBA()
			if a16 == 0 {
				continue
			}
			r := uint8(r16 >> 8)
			g := uint8(g16 >> 8)
			b := uint8(b16 >> 8)
			if r >= minValue && r <= maxValue && g >= minValue && g <= maxValue && b >= minValue && b <= maxValue {
				count++
			}
		}
	}
	return count
}
