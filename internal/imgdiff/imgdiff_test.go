package imgdiff

import (
	"image"
	"image/color"
	"testing"
)

func TestAnalyzeReportsRMSEAndDifferentPixels(t *testing.T) {
	ref := image.NewRGBA(image.Rect(0, 0, 2, 1))
	gen := image.NewRGBA(image.Rect(0, 0, 2, 1))
	ref.SetRGBA(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	ref.SetRGBA(1, 0, color.RGBA{R: 100, G: 110, B: 120, A: 255})
	gen.SetRGBA(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	gen.SetRGBA(1, 0, color.RGBA{R: 90, G: 110, B: 140, A: 255})

	stats, err := Analyze(ref, gen)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if stats.DifferentPixels != 1 {
		t.Fatalf("DifferentPixels = %d, want 1", stats.DifferentPixels)
	}
	if stats.MaxDiff != 20 {
		t.Fatalf("MaxDiff = %d, want 20", stats.MaxDiff)
	}
	if stats.TotalPixels != 2 {
		t.Fatalf("TotalPixels = %d, want 2", stats.TotalPixels)
	}
	if stats.RMSE < 9.12 || stats.RMSE > 9.13 {
		t.Fatalf("RMSE = %.6f, want about 9.1287", stats.RMSE)
	}
	if got, want := stats.Ratio(), 0.5; got != want {
		t.Fatalf("Ratio = %v, want %v", got, want)
	}
}

func TestAmplifyUsesDimReferenceForIdenticalAndAmplifiesDifferences(t *testing.T) {
	ref := image.NewRGBA(image.Rect(0, 0, 2, 1))
	gen := image.NewRGBA(image.Rect(0, 0, 2, 1))
	ref.SetRGBA(0, 0, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	gen.SetRGBA(0, 0, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	ref.SetRGBA(1, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	gen.SetRGBA(1, 0, color.RGBA{R: 13, G: 28, B: 10, A: 255})

	diff, err := Amplify(ref, gen, 10)
	if err != nil {
		t.Fatalf("Amplify returned error: %v", err)
	}

	if got, want := diff.RGBAAt(0, 0), (color.RGBA{R: 45, G: 45, B: 45, A: 255}); got != want {
		t.Fatalf("identical pixel = %#v, want %#v", got, want)
	}
	if got, want := diff.RGBAAt(1, 0), (color.RGBA{R: 30, G: 80, B: 200, A: 255}); got != want {
		t.Fatalf("different pixel = %#v, want %#v", got, want)
	}
}

func TestAnalyzeRejectsDimensionMismatch(t *testing.T) {
	ref := image.NewRGBA(image.Rect(0, 0, 1, 1))
	gen := image.NewRGBA(image.Rect(0, 0, 2, 1))

	if _, err := Analyze(ref, gen); err == nil {
		t.Fatal("Analyze accepted images with different dimensions")
	}
	if _, err := Amplify(ref, gen, 8); err == nil {
		t.Fatal("Amplify accepted images with different dimensions")
	}
	if _, err := Amplify(ref, ref, 0); err == nil {
		t.Fatal("Amplify accepted non-positive factor")
	}
}
