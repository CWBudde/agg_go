package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/color"
)

func TestDistortionsMatchesCPPReference(t *testing.T) {
	img := agg.NewImage(make([]uint8, canvasW*canvasH*4), canvasW, canvasH, -canvasW*4)
	(&demo{}).Render(img)

	got := encodeLinearToSRGBForTest(img.ToGoImage())
	want := loadPNGForTest(t, filepath.Join("..", "..", "..", "..", "tests", "visual", "reference", "cpp", "examples", "distortions.png"))

	if got.Bounds().Size() != want.Bounds().Size() {
		t.Fatalf("dimension mismatch: got=%v want=%v", got.Bounds().Size(), want.Bounds().Size())
	}

	diffPixels, totalPixels, maxDiff := compareRGBWithTolerance(got, want, 10)
	maxAllowed := totalPixels / 100
	if diffPixels > maxAllowed {
		t.Fatalf("image differs from C++ reference: different_pixels=%d/%d max_allowed=%d max_diff=%d",
			diffPixels, totalPixels, maxAllowed, maxDiff)
	}
}

func TestBuildGradientColorsConvertsSRGBTableToLinear(t *testing.T) {
	grad := buildGradientColors()
	index := 80
	got := grad.ColorAt(index)
	want := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{
		R: g_gradient_colors[index*4+0],
		G: g_gradient_colors[index*4+1],
		B: g_gradient_colors[index*4+2],
		A: g_gradient_colors[index*4+3],
	})
	if got != want {
		t.Fatalf("gradient[%d]=%+v, want sRGB-decoded %+v", index, got, want)
	}
}

func encodeLinearToSRGBForTest(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	for i := 0; i+3 < len(dst.Pix); i += 4 {
		c := color.ConvertToSRGBFromLinear(color.RGBA8[color.Linear]{
			R: dst.Pix[i],
			G: dst.Pix[i+1],
			B: dst.Pix[i+2],
			A: dst.Pix[i+3],
		})
		dst.Pix[i] = c.R
		dst.Pix[i+1] = c.G
		dst.Pix[i+2] = c.B
		dst.Pix[i+3] = c.A
	}
	return dst
}

func loadPNGForTest(t *testing.T, path string) *image.RGBA {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rgba.Set(x, y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return rgba
}

func compareRGBWithTolerance(got, want *image.RGBA, tolerance uint8) (diffPixels, totalPixels, maxDiff int) {
	bounds := got.Bounds()
	totalPixels = bounds.Dx() * bounds.Dy()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			gi := got.PixOffset(x, y)
			wi := want.PixOffset(x, y)
			pixelDifferent := false
			for c := 0; c < 3; c++ {
				diff := absDiffUint8(got.Pix[gi+c], want.Pix[wi+c])
				if int(diff) > maxDiff {
					maxDiff = int(diff)
				}
				if diff > tolerance {
					pixelDifferent = true
				}
			}
			if pixelDifferent {
				diffPixels++
			}
		}
	}
	return diffPixels, totalPixels, maxDiff
}

func absDiffUint8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
