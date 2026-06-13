package agg2d

import (
	"math"
	"testing"
)

// viewportScene is the method subset needed to exercise the viewport +
// coordinate-mapping methods across both pipelines. All of these are pure
// affine math (or clip-box membership) and are color-agnostic; the float twin
// must produce bit-identical scalar results to the 8-bit oracle since both
// operate on the same world transform / clip box.
type viewportScene interface {
	ClearAll(Color)
	FillColor(Color)
	ResetPath()
	MoveTo(x, y float64)
	LineTo(x, y float64)
	ClosePolygon()
	DrawPath(DrawPathFlag)
	Viewport(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption)
	WorldToScreenDistance(worldDistance float64) float64
	ScreenToWorldDistance(screenDistance float64) (float64, bool)
	AlignPoint(x, y *float64)
	InBox(worldX, worldY float64) bool
	AffineImageResamplePolicy(policy AffineImageResamplePolicy)
	GetAffineImageResamplePolicy() AffineImageResamplePolicy
}

var (
	_ viewportScene = (*Agg2D)(nil)
	_ viewportScene = (*Agg2DFloat)(nil)
)

// TestParityViewportRender verifies that Viewport mutates the world transform
// identically: a unit-square path drawn through a viewport mapping lands on the
// same pixels in both pipelines.
func TestParityViewportRender(t *testing.T) {
	const w, h = 60, 60
	scene := func(s viewportScene) {
		s.ClearAll(NewColor(255, 255, 255, 255))
		// Map world [0,10]x[0,10] onto screen [5,5]-[55,55] (anisotropic).
		s.Viewport(0, 0, 10, 10, 5, 5, 55, 55, Anisotropic)
		s.FillColor(NewColor(200, 40, 40, 255))
		s.ResetPath()
		s.MoveTo(2, 2)
		s.LineTo(8, 2)
		s.LineTo(8, 8)
		s.LineTo(2, 8)
		s.ClosePolygon()
		s.DrawPath(FillOnly)
	}

	a8 := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a8.Attach(buf, w, h, w*4)
	scene(a8)

	af := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	af.AttachImageFloat(img)
	scene(af)

	maxDiff, ink := 0, 0
	for y := range h {
		for x := range w {
			c8 := pixel8(buf, w, x, y)
			cf := pixelFloatAsU8(img, x, y)
			if d := maxChanDiff(c8, cf); d > maxDiff {
				maxDiff = d
			}
			if c8[0] < 250 || c8[1] < 250 || c8[2] < 250 {
				ink++
			}
		}
	}
	if ink < 50 {
		t.Fatalf("Viewport: 8-bit oracle drew too little ink: %d pixels", ink)
	}
	if maxDiff > 2 {
		t.Errorf("Viewport: float vs 8-bit max channel diff = %d (tol 2)", maxDiff)
	}
}

// setupViewportState applies an identical transform + clip on a viewportScene so
// the coordinate-mapping accessors can be compared.
func setupViewportState(s viewportScene) {
	s.Viewport(0, 0, 100, 100, 0, 0, 50, 50, XMidYMid)
}

func TestParityViewportCoordinateAccessors(t *testing.T) {
	a8 := NewAgg2D()
	buf := make([]uint8, 50*50*4)
	a8.Attach(buf, 50, 50, 50*4)
	setupViewportState(a8)

	af := NewAgg2DFloat()
	img := NewImageFloatEmpty(50, 50)
	af.AttachImageFloat(img)
	setupViewportState(af)

	// WorldToScreenDistance.
	for _, d := range []float64{0, 1, 7.5, 100} {
		got8 := a8.WorldToScreenDistance(d)
		gotF := af.WorldToScreenDistance(d)
		if math.Abs(got8-gotF) > 1e-9 {
			t.Errorf("WorldToScreenDistance(%v): float=%v 8bit=%v", d, gotF, got8)
		}
	}

	// ScreenToWorldDistance.
	for _, d := range []float64{0, 1, 25, 50} {
		got8, ok8 := a8.ScreenToWorldDistance(d)
		gotF, okF := af.ScreenToWorldDistance(d)
		if ok8 != okF || math.Abs(got8-gotF) > 1e-9 {
			t.Errorf("ScreenToWorldDistance(%v): float=(%v,%v) 8bit=(%v,%v)", d, gotF, okF, got8, ok8)
		}
	}

	// AlignPoint.
	for _, p := range [][2]float64{{0, 0}, {3.3, 7.7}, {49, 49}} {
		x8, y8 := p[0], p[1]
		a8.AlignPoint(&x8, &y8)
		xF, yF := p[0], p[1]
		af.AlignPoint(&xF, &yF)
		if math.Abs(x8-xF) > 1e-9 || math.Abs(y8-yF) > 1e-9 {
			t.Errorf("AlignPoint(%v,%v): float=(%v,%v) 8bit=(%v,%v)", p[0], p[1], xF, yF, x8, y8)
		}
	}

	// InBox.
	for _, p := range [][2]float64{{0, 0}, {50, 50}, {99, 99}, {-5, 5}, {100, 100}} {
		if got8, gotF := a8.InBox(p[0], p[1]), af.InBox(p[0], p[1]); got8 != gotF {
			t.Errorf("InBox(%v,%v): float=%v 8bit=%v", p[0], p[1], gotF, got8)
		}
	}

	// AffineImageResamplePolicy round-trip.
	af.AffineImageResamplePolicy(AffineImageResamplePreferFiltered)
	if af.GetAffineImageResamplePolicy() != AffineImageResamplePreferFiltered {
		t.Errorf("AffineImageResamplePolicy round-trip: got %v", af.GetAffineImageResamplePolicy())
	}
	if af.GetAffineImageResamplePolicy() != a8.GetAffineImageResamplePolicy() {
		// a8 still has the default; set it the same to compare the getter wiring.
		a8.AffineImageResamplePolicy(AffineImageResamplePreferFiltered)
		if af.GetAffineImageResamplePolicy() != a8.GetAffineImageResamplePolicy() {
			t.Errorf("GetAffineImageResamplePolicy: float=%v 8bit=%v",
				af.GetAffineImageResamplePolicy(), a8.GetAffineImageResamplePolicy())
		}
	}
}
