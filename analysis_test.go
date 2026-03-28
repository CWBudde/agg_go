package agg

import (
	"math"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/color"
)

// ---------------------------------------------------------------------------
// Mock pixel formats for testing — mirrors how C++ AGG tests instantiate
// concrete pixfmt types for algorithm testing.
// ---------------------------------------------------------------------------

// mockRGBA8Image is a simple in-memory pixel format for testing, analogous to
// constructing a pixfmt_alpha_blend_rgba over a rendering_buffer in C++ tests.
type mockRGBA8Image struct {
	pixels []color.RGBA8[color.Linear]
	w, h   int
}

func newMockRGBA8Image(w, h int) *mockRGBA8Image {
	return &mockRGBA8Image{pixels: make([]color.RGBA8[color.Linear], w*h), w: w, h: h}
}

func (m *mockRGBA8Image) Width() int  { return m.w }
func (m *mockRGBA8Image) Height() int { return m.h }

func (m *mockRGBA8Image) Pixel(x, y int) color.RGBA8[color.Linear] {
	return m.pixels[y*m.w+x]
}

func (m *mockRGBA8Image) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	copy(m.pixels[y*m.w+x:], colors[:length])
}

func (m *mockRGBA8Image) set(x, y int, c color.RGBA8[color.Linear]) {
	m.pixels[y*m.w+x] = c
}

// ---------------------------------------------------------------------------
// SobelGradient tests — exercised through the generic interface
// ---------------------------------------------------------------------------

// TestSobelGradientFlat verifies that a uniform (flat) image produces
// zero gradient everywhere.
func TestSobelGradientFlat(t *testing.T) {
	img := newMockRGBA8Image(4, 4)
	c := color.RGBA8[color.Linear]{R: 128, G: 128, B: 128, A: 255}
	for i := range img.pixels {
		img.pixels[i] = c
	}
	grad := SobelGradient(img, LuminanceRGBA8Linear())
	for i, v := range grad {
		if v > 1e-6 {
			t.Fatalf("grad[%d] = %v on flat image, want ~0", i, v)
		}
	}
}

// TestSobelGradientEdge verifies that a sharp vertical edge produces
// non-zero gradient in the transition columns and zero in the flat halves.
func TestSobelGradientEdge(t *testing.T) {
	const w, h = 6, 1
	img := newMockRGBA8Image(w, h)
	for x := range w {
		val := byte(0)
		if x >= w/2 {
			val = 255
		}
		img.set(x, 0, color.RGBA8[color.Linear]{R: val, G: val, B: val, A: 255})
	}
	grad := SobelGradient(img, LuminanceRGBA8Linear())

	maxGrad := float32(0)
	for _, v := range grad {
		if v > maxGrad {
			maxGrad = v
		}
	}
	if maxGrad < 0.5 {
		t.Fatalf("expected high gradient near the edge, max = %v", maxGrad)
	}
	for i, v := range grad {
		if v < 0 || v > 1 {
			t.Fatalf("grad[%d] = %v out of [0,1] range", i, v)
		}
	}
}


// ---------------------------------------------------------------------------
// StackBlur tests — exercised through the PixelReadWriter interface
// ---------------------------------------------------------------------------

// TestStackBlurPreservesOpaque verifies that blurring a solid-colour image
// leaves the colour unchanged (all-identical neighbours average to themselves).
func TestStackBlurPreservesOpaque(t *testing.T) {
	img := newMockRGBA8Image(8, 8)
	want := color.RGBA8[color.Linear]{R: 200, G: 100, B: 50, A: 255}
	for i := range img.pixels {
		img.pixels[i] = want
	}

	sb := NewStackBlur()
	sb.Blur(img, 2)

	for i, got := range img.pixels {
		if got != want {
			t.Fatalf("pixel[%d] = %v, want %v after solid-colour blur", i, got, want)
		}
	}
}


// TestStackBlurSpreads verifies that a single bright pixel gets spread
// across its neighbours.
func TestStackBlurSpreads(t *testing.T) {
	img := newMockRGBA8Image(5, 5)
	img.set(2, 2, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255})

	sb := NewStackBlur()
	sb.Blur(img, 1)

	centre := img.Pixel(2, 2)
	if centre.R == 255 {
		t.Fatal("centre should have been blurred down from 255")
	}
	if centre.R == 0 {
		t.Fatal("centre should still have some red after blur")
	}
}

// ---------------------------------------------------------------------------
// Polyline simplification tests (unchanged from geometry.go)
// ---------------------------------------------------------------------------

func TestSimplifyPolylineCollinear(t *testing.T) {
	pts := []Point{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}}
	got := SimplifyPolyline(pts, 0.5)
	if len(got) != 2 {
		t.Fatalf("SimplifyPolyline collinear: got %d points, want 2", len(got))
	}
}

func TestSimplifyPolylineKink(t *testing.T) {
	pts := []Point{{0, 0}, {5, 5}, {10, 0}}
	got := SimplifyPolyline(pts, 0.5)
	if len(got) < 3 {
		t.Fatalf("SimplifyPolyline kink: got %d points, want >= 3", len(got))
	}
}

func TestSimplifyPolylineTrivial(t *testing.T) {
	pts1 := []Point{{0, 0}}
	if got := SimplifyPolyline(pts1, 1.0); len(got) != 1 {
		t.Fatalf("single-point: got %d, want 1", len(got))
	}
	pts2 := []Point{{0, 0}, {1, 1}}
	if got := SimplifyPolyline(pts2, 1.0); len(got) != 2 {
		t.Fatalf("two-point: got %d, want 2", len(got))
	}
}

func TestPointLineDistanceOrthogonal(t *testing.T) {
	p := Point{0, 3}
	a := Point{-5, 0}
	b := Point{5, 0}
	d := pointLineDistance(p, a, b)
	if math.Abs(d-3.0) > 1e-9 {
		t.Fatalf("pointLineDistance = %v, want 3.0", d)
	}
}
