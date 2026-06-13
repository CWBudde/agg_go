package agg2d

import "testing"

// gradVarScene is the method subset needed to exercise the gradient-variant
// setters and accessors across both pipelines. The gradient LUT, span pipeline,
// and world-radial setup are color-agnostic; the float builder interpolates
// stops in RGBA32 space but must produce the same on-screen result as 8-bit.
type gradVarScene interface {
	ClearAll(Color)
	FillColor(Color)
	LineColor(Color)
	LineWidth(float64)
	ResetPath()
	MoveTo(x, y float64)
	LineTo(x, y float64)
	ClosePolygon()
	FillRadialGradientStops(x, y, r float64, stops []ColorStop)
	FillRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color)
	LineRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color)
	FillRadialGradientPos(x, y, r float64)
	LineRadialGradientPos(x, y, r float64)
	DrawPath(DrawPathFlag)
}

var (
	_ gradVarScene = (*Agg2D)(nil)
	_ gradVarScene = (*Agg2DFloat)(nil)
)

func renderGradVar8(w, h int, scene func(gradVarScene)) []uint8 {
	a := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a.Attach(buf, w, h, w*4)
	scene(a)
	return buf
}

func renderGradVarFloat(w, h int, scene func(gradVarScene)) *ImageFloat {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	a.AttachImageFloat(img)
	scene(a)
	return img
}

func assertGradVarParity(t *testing.T, name string, w, h, tol int, scene func(gradVarScene)) {
	t.Helper()
	buf := renderGradVar8(w, h, scene)
	img := renderGradVarFloat(w, h, scene)

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
		t.Fatalf("%s: 8-bit oracle drew too little ink: %d pixels", name, ink)
	}
	if maxDiff > tol {
		t.Errorf("%s: float vs 8-bit max channel diff = %d (tol %d)", name, maxDiff, tol)
	}
}

func gradVarRect(s gradVarScene, x1, y1, x2, y2 float64) {
	s.ResetPath()
	s.MoveTo(x1, y1)
	s.LineTo(x2, y1)
	s.LineTo(x2, y2)
	s.LineTo(x1, y2)
	s.ClosePolygon()
}

func TestParityGradRadialStops(t *testing.T) {
	assertGradVarParity(t, "FillRadialGradientStops", 60, 60, 3, func(s gradVarScene) {
		s.ClearAll(NewColor(255, 255, 255, 255))
		s.FillRadialGradientStops(30, 30, 28, []ColorStop{
			{Position: 0.0, Color: NewColor(255, 0, 0, 255)},
			{Position: 0.4, Color: NewColor(0, 255, 0, 255)},
			{Position: 0.75, Color: NewColor(0, 0, 255, 255)},
			{Position: 1.0, Color: NewColor(0, 0, 0, 255)},
		})
		gradVarRect(s, 2, 2, 58, 58)
		s.DrawPath(FillOnly)
	})
}

func TestParityGradFillRadialMultiStop(t *testing.T) {
	assertGradVarParity(t, "FillRadialGradientMultiStop", 60, 60, 3, func(s gradVarScene) {
		s.ClearAll(NewColor(255, 255, 255, 255))
		s.FillRadialGradientMultiStop(30, 30, 28,
			NewColor(240, 200, 30, 255), NewColor(200, 40, 40, 255), NewColor(20, 40, 160, 255))
		gradVarRect(s, 2, 2, 58, 58)
		s.DrawPath(FillOnly)
	})
}

func TestParityGradLineRadialMultiStop(t *testing.T) {
	assertGradVarParity(t, "LineRadialGradientMultiStop", 60, 60, 3, func(s gradVarScene) {
		s.ClearAll(NewColor(255, 255, 255, 255))
		s.LineWidth(6)
		s.LineRadialGradientMultiStop(30, 30, 28,
			NewColor(240, 200, 30, 255), NewColor(200, 40, 40, 255), NewColor(20, 40, 160, 255))
		gradVarRect(s, 6, 6, 54, 54)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityGradFillRadialPos(t *testing.T) {
	// FillRadialGradientPos repositions an existing gradient without touching the
	// color ramp. Set up colors at one centre, then move the centre.
	assertGradVarParity(t, "FillRadialGradientPos", 60, 60, 3, func(s gradVarScene) {
		s.ClearAll(NewColor(255, 255, 255, 255))
		s.FillRadialGradientMultiStop(10, 10, 28,
			NewColor(240, 200, 30, 255), NewColor(200, 40, 40, 255), NewColor(20, 40, 160, 255))
		s.FillRadialGradientPos(30, 30, 28) // reposition only
		gradVarRect(s, 2, 2, 58, 58)
		s.DrawPath(FillOnly)
	})
}

func TestParityGradLineRadialPos(t *testing.T) {
	assertGradVarParity(t, "LineRadialGradientPos", 60, 60, 3, func(s gradVarScene) {
		s.ClearAll(NewColor(255, 255, 255, 255))
		s.LineWidth(6)
		s.LineRadialGradientMultiStop(10, 10, 28,
			NewColor(240, 200, 30, 255), NewColor(200, 40, 40, 255), NewColor(20, 40, 160, 255))
		s.LineRadialGradientPos(30, 30, 28) // reposition only
		gradVarRect(s, 6, 6, 54, 54)
		s.DrawPath(StrokeOnly)
	})
}

// TestAgg2DFloatGradientAccessors verifies the D1/D2/flag readbacks match the
// 8-bit oracle after the same setup.
func TestAgg2DFloatGradientAccessors(t *testing.T) {
	a8 := NewAgg2D()
	af := NewAgg2DFloat()
	buf := make([]uint8, 60*60*4)
	a8.Attach(buf, 60, 60, 60*4)
	img := NewImageFloatEmpty(60, 60)
	af.AttachImageFloat(img)

	for _, a := range []gradVarScene{a8, af} {
		a.FillRadialGradientMultiStop(30, 30, 28,
			NewColor(255, 0, 0, 255), NewColor(0, 255, 0, 255), NewColor(0, 0, 255, 255))
		a.LineRadialGradientMultiStop(10, 10, 12,
			NewColor(0, 0, 0, 255), NewColor(128, 128, 128, 255), NewColor(255, 255, 255, 255))
	}

	if af.FillGradientFlag() != a8.FillGradientFlag() {
		t.Errorf("FillGradientFlag float=%v 8bit=%v", af.FillGradientFlag(), a8.FillGradientFlag())
	}
	if af.LineGradientFlag() != a8.LineGradientFlag() {
		t.Errorf("LineGradientFlag float=%v 8bit=%v", af.LineGradientFlag(), a8.LineGradientFlag())
	}
	if af.FillGradientD1() != a8.FillGradientD1() || af.FillGradientD2() != a8.FillGradientD2() {
		t.Errorf("FillGradient D1/D2 float=(%v,%v) 8bit=(%v,%v)",
			af.FillGradientD1(), af.FillGradientD2(), a8.FillGradientD1(), a8.FillGradientD2())
	}
	if af.LineGradientD1() != a8.LineGradientD1() || af.LineGradientD2() != a8.LineGradientD2() {
		t.Errorf("LineGradient D1/D2 float=(%v,%v) 8bit=(%v,%v)",
			af.LineGradientD1(), af.LineGradientD2(), a8.LineGradientD1(), a8.LineGradientD2())
	}
}
