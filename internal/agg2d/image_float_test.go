package agg2d

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func filledFloatImage(w, h int, c color.RGBA32[color.Linear]) *ImageFloat {
	img := NewImageFloatEmpty(w, h)
	for y := range h {
		for x := range w {
			img.SetPixel(x, y, c)
		}
	}
	return img
}

func TestAgg2DFloatCopyImage(t *testing.T) {
	a, dst := setupFloatTarget(12, 12)
	src := filledFloatImage(4, 4, color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 1.0))

	a.CopyImageFloat(src, 3, 3)

	in := dst.GetPixel(4, 4)
	if !approxF(in.R, 0.2) || !approxF(in.G, 0.4) || !approxF(in.B, 0.6) || !approxF(in.A, 1.0) {
		t.Fatalf("copied pixel(4,4) = %+v, want {0.2,0.4,0.6,1}", in)
	}
	out := dst.GetPixel(0, 0)
	if out.A != 0 {
		t.Fatalf("outside-copy pixel(0,0) alpha = %v, want 0", out.A)
	}
}

func TestAgg2DFloatBlendImageOpaqueOverTransparent(t *testing.T) {
	a, dst := setupFloatTarget(12, 12)
	src := filledFloatImage(4, 4, color.NewRGBA32[color.Linear](1.0, 0.0, 0.0, 1.0))

	a.BlendImageFloat(src, 3, 3, 255)

	in := dst.GetPixel(5, 5)
	if !approxF(in.R, 1.0) || in.A <= 0 {
		t.Fatalf("blended pixel(5,5) = %+v, want opaque red", in)
	}
}
