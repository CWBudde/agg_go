package agg2d

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

// makeSourceImage8 builds an opaque 8-bit source image with a smooth two-axis
// color ramp so bilinear/affine sampling produces interpolated values.
func makeSourceImage8(w, h int) *Image {
	data := make([]uint8, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := (y*w + x) * 4
			data[o+0] = uint8(20 + x*200/maxInt(1, w-1))
			data[o+1] = uint8(30 + y*200/maxInt(1, h-1))
			data[o+2] = uint8(128)
			data[o+3] = 255
		}
	}
	return NewImage(data, w, h, w*4)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// transformParity renders the same image transform through the 8-bit and float
// pipelines (opaque source over opaque-white background, so premultiply is a
// no-op and blending is a straight replace) and compares interior pixels.
func transformParity(t *testing.T, w, h int, run8 func(*Agg2D, *Image), runF func(*Agg2DFloat, *ImageFloat), samples [][2]int, tol int) {
	t.Helper()

	srcW, srcH := 8, 8
	src8 := makeSourceImage8(srcW, srcH)
	srcF := NewImageFloatFromImage8(src8)

	a8 := NewAgg2D()
	dst8 := make([]uint8, w*h*4)
	a8.Attach(dst8, w, h, w*4)
	a8.ClearAll(NewColor(255, 255, 255, 255))
	run8(a8, src8)

	aF := NewAgg2DFloat()
	imgF := NewImageFloatEmpty(w, h)
	aF.AttachImageFloat(imgF)
	aF.ClearAll(NewColor(255, 255, 255, 255))
	runF(aF, srcF)

	for _, p := range samples {
		c8 := pixel8(dst8, w, p[0], p[1])
		cf := pixelFloatAsU8(imgF, p[0], p[1])
		if d := maxChanDiff(c8, cf); d > tol {
			t.Errorf("transform mismatch at (%d,%d): 8bit=%v float=%v maxdiff=%d (tol=%d)", p[0], p[1], c8, cf, d, tol)
		}
	}
}

// TestTransformImageFloatAffineScaleParity scales the source up to a destination
// rectangle (the default bilinear filter path) and checks parity with 8-bit.
func TestTransformImageFloatAffineScaleParity(t *testing.T) {
	const w, h = 48, 48
	transformParity(
		t, w, h,
		func(a *Agg2D, src *Image) {
			if err := a.TransformImageSimple(src, 8, 8, 40, 40); err != nil {
				t.Fatalf("8bit TransformImageSimple: %v", err)
			}
		},
		func(a *Agg2DFloat, src *ImageFloat) {
			if err := a.TransformImageFloatSimple(src, 8, 8, 40, 40); err != nil {
				t.Fatalf("float TransformImageFloatSimple: %v", err)
			}
		},
		[][2]int{{16, 16}, {24, 24}, {32, 32}, {20, 30}, {30, 18}},
		3,
	)
}

// TestTransformImageFloatParallelogramParity maps the source to a sheared
// parallelogram and checks parity with the 8-bit path.
func TestTransformImageFloatParallelogramParity(t *testing.T) {
	const w, h = 64, 64
	parl := []float64{12, 12, 52, 18, 46, 50}
	transformParity(
		t, w, h,
		func(a *Agg2D, src *Image) {
			if err := a.TransformImageParallelogramSimple(src, parl); err != nil {
				t.Fatalf("8bit TransformImageParallelogramSimple: %v", err)
			}
		},
		func(a *Agg2DFloat, src *ImageFloat) {
			if err := a.TransformImageFloatParallelogramSimple(src, parl); err != nil {
				t.Fatalf("float TransformImageFloatParallelogramSimple: %v", err)
			}
		},
		[][2]int{{30, 25}, {35, 30}, {28, 35}, {40, 28}},
		4,
	)
}

// TestTransformImageFloatQuadParity maps the source to a perspective quadrangle
// and checks parity with the 8-bit perspective path.
func TestTransformImageFloatQuadParity(t *testing.T) {
	const w, h = 64, 64
	quad := [8]float64{10, 12, 54, 8, 50, 56, 14, 48}
	transformParity(
		t, w, h,
		func(a *Agg2D, src *Image) {
			if err := a.TransformImageQuadSimple(src, quad); err != nil {
				t.Fatalf("8bit TransformImageQuadSimple: %v", err)
			}
		},
		func(a *Agg2DFloat, src *ImageFloat) {
			if err := a.TransformImageFloatQuadSimple(src, quad); err != nil {
				t.Fatalf("float TransformImageFloatQuadSimple: %v", err)
			}
		},
		[][2]int{{30, 28}, {32, 32}, {28, 36}, {38, 30}},
		4,
	)
}

// TestTransformImageFloatInvalidArgs checks bounds/argument validation.
func TestTransformImageFloatInvalidArgs(t *testing.T) {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(16, 16)
	a.AttachImageFloat(img)
	src := NewImageFloatEmpty(8, 8)

	if err := a.TransformImageFloat(nil, 0, 0, 8, 8, 0, 0, 8, 8); err == nil {
		t.Error("expected error for nil image")
	}
	if err := a.TransformImageFloat(src, -1, 0, 8, 8, 0, 0, 8, 8); err == nil {
		t.Error("expected error for out-of-bounds source rect")
	}
	if err := a.TransformImageFloatParallelogram(src, 0, 0, 8, 8, []float64{0, 0, 1, 1}); err == nil {
		t.Error("expected error for short parallelogram")
	}
}

// TestSpanImageFilterRGBA32BilinearNoBias verifies the float bilinear filter
// returns the true weighted average without the integer rounding bias that the
// 8-bit template carries (a uniform source must come back unchanged, not +0.5).
func TestSpanImageFilterRGBA32BilinearNoBias(t *testing.T) {
	// 4x4 uniform mid-gray opaque source.
	const n = 4
	src := NewImageFloatEmpty(n, n)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			src.SetPixel(x, y, color.RGBA32[color.Linear]{R: 0.5, G: 0.25, B: 0.75, A: 1})
		}
	}
	source := newImagePixelFormatFloat(src)

	// Identity affine interpolator: destination pixel (x,y) maps to source (x,y).
	id := transform.NewTransAffine()
	interp := span.NewSpanInterpolatorLinearDefault(id)
	filter := span.NewSpanImageFilterRGBA32BilinearWithParams[*imagePixelFormatFloat, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interp)

	out := make([]color.RGBA32[color.Linear], 2)
	filter.Generate(out, 1, 1)
	got := out[0]
	const eps = 1e-3
	if absF(got.R-0.5) > eps || absF(got.G-0.25) > eps || absF(got.B-0.75) > eps || absF(got.A-1) > eps {
		t.Errorf("bilinear on uniform source returned %v, want ~{0.5,0.25,0.75,1} (bias leak?)", got)
	}
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
