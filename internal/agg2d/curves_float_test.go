package agg2d

import "testing"

// curveScene is the method subset needed to exercise the curve-command and
// relative path-building methods across both pipelines. The smooth-curve
// reflection math depends on the shared lastCtrl/hasLastCtrl state, so the
// 8-bit and float paths must track control points identically.
type curveScene interface {
	ClearAll(Color)
	LineColor(Color)
	LineWidth(float64)
	ResetPath()
	MoveTo(x, y float64)
	HorLineRel(dx float64)
	VerLineRel(dy float64)
	QuadricCurveTo(xCtrl, yCtrl, xTo, yTo float64)
	QuadricCurveRel(dxCtrl, dyCtrl, dxTo, dyTo float64)
	QuadricCurveToSmooth(xTo, yTo float64)
	QuadricCurveRelSmooth(dxTo, dyTo float64)
	CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64)
	CubicCurveRel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo float64)
	CubicCurveToSmooth(xCtrl2, yCtrl2, xTo, yTo float64)
	CubicCurveRelSmooth(dxCtrl2, dyCtrl2, dxTo, dyTo float64)
	DrawPath(DrawPathFlag)
}

var (
	_ curveScene = (*Agg2D)(nil)
	_ curveScene = (*Agg2DFloat)(nil)
)

func renderCurve8(w, h int, scene func(curveScene)) []uint8 {
	a := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a.Attach(buf, w, h, w*4)
	scene(a)
	return buf
}

func renderCurveFloat(w, h int, scene func(curveScene)) *ImageFloat {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	a.AttachImageFloat(img)
	scene(a)
	return img
}

func assertCurveParity(t *testing.T, name string, w, h, tol int, scene func(curveScene)) {
	t.Helper()
	buf := renderCurve8(w, h, scene)
	img := renderCurveFloat(w, h, scene)

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
	if ink < 20 {
		t.Fatalf("%s: 8-bit oracle drew too little ink: %d pixels", name, ink)
	}
	if maxDiff > tol {
		t.Errorf("%s: float vs 8-bit max channel diff = %d (tol %d)", name, maxDiff, tol)
	}
}

func curveSceneSetup(s curveScene) {
	s.ClearAll(NewColor(255, 255, 255, 255))
	s.LineColor(NewColor(20, 60, 180, 255))
	s.LineWidth(2)
	s.ResetPath()
}

func TestParityCurveQuadricRel(t *testing.T) {
	assertCurveParity(t, "QuadricCurveRel", 90, 90, 2, func(s curveScene) {
		curveSceneSetup(s)
		s.MoveTo(10, 45)
		s.QuadricCurveRel(20, -35, 40, 0)
		s.QuadricCurveRel(20, 35, 40, 0)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityCurveQuadricSmooth(t *testing.T) {
	assertCurveParity(t, "QuadricCurveToSmooth", 90, 90, 2, func(s curveScene) {
		curveSceneSetup(s)
		s.MoveTo(10, 45)
		s.QuadricCurveTo(30, 8, 45, 45)
		s.QuadricCurveToSmooth(80, 45) // reflects the previous control point
		s.DrawPath(StrokeOnly)
	})
}

func TestParityCurveQuadricRelSmooth(t *testing.T) {
	assertCurveParity(t, "QuadricCurveRelSmooth", 90, 90, 2, func(s curveScene) {
		curveSceneSetup(s)
		s.MoveTo(10, 45)
		s.QuadricCurveTo(30, 8, 45, 45)
		s.QuadricCurveRelSmooth(35, 0)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityCurveCubicRel(t *testing.T) {
	assertCurveParity(t, "CubicCurveRel", 100, 90, 2, func(s curveScene) {
		curveSceneSetup(s)
		s.MoveTo(10, 45)
		s.CubicCurveRel(15, -40, 45, 40, 60, 0)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityCurveCubicSmooth(t *testing.T) {
	assertCurveParity(t, "CubicCurveToSmooth", 110, 90, 2, func(s curveScene) {
		curveSceneSetup(s)
		s.MoveTo(10, 45)
		s.CubicCurveTo(20, 5, 40, 5, 50, 45)
		s.CubicCurveToSmooth(80, 85, 95, 45) // first ctrl reflected
		s.DrawPath(StrokeOnly)
	})
}

func TestParityCurveCubicRelSmooth(t *testing.T) {
	assertCurveParity(t, "CubicCurveRelSmooth", 110, 90, 2, func(s curveScene) {
		curveSceneSetup(s)
		s.MoveTo(10, 45)
		s.CubicCurveTo(20, 5, 40, 5, 50, 45)
		s.CubicCurveRelSmooth(30, 40, 45, 0)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityCurveHorVerLineRel(t *testing.T) {
	assertCurveParity(t, "HorVerLineRel", 90, 90, 2, func(s curveScene) {
		curveSceneSetup(s)
		s.MoveTo(15, 15)
		s.HorLineRel(55)
		s.VerLineRel(55)
		s.HorLineRel(-55)
		s.VerLineRel(-55)
		s.DrawPath(StrokeOnly)
	})
}
