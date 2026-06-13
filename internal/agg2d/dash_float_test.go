package agg2d

import "testing"

// dashScene is the method subset needed to exercise dashed-stroke rendering
// across both the 8-bit and float pipelines. Dash-phase positioning, segment
// lengths, and the solid-fallback branch (NumDashes()==0) must match.
type dashScene interface {
	ClearAll(Color)
	LineColor(Color)
	LineWidth(float64)
	LineCap(LineCap)
	ResetPath()
	MoveTo(x, y float64)
	LineTo(x, y float64)
	AddDash(dashLen, gapLen float64)
	RemoveAllDashes()
	DashStart(offset float64)
	GetDashStart() float64
	NoDashes()
	DrawPath(DrawPathFlag)
}

var (
	_ dashScene = (*Agg2D)(nil)
	_ dashScene = (*Agg2DFloat)(nil)
)

func renderDash8(w, h int, scene func(dashScene)) []uint8 {
	a := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a.Attach(buf, w, h, w*4)
	scene(a)
	return buf
}

func renderDashFloat(w, h int, scene func(dashScene)) *ImageFloat {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	a.AttachImageFloat(img)
	scene(a)
	return img
}

func assertDashParity(t *testing.T, name string, w, h, tol int, scene func(dashScene)) {
	t.Helper()
	buf := renderDash8(w, h, scene)
	img := renderDashFloat(w, h, scene)

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

func dashSceneSetup(s dashScene) {
	s.ClearAll(NewColor(255, 255, 255, 255))
	s.LineColor(NewColor(20, 60, 180, 255))
	s.LineWidth(3)
	s.LineCap(CapButt)
	s.ResetPath()
}

func TestParityDashSimpleLine(t *testing.T) {
	assertDashParity(t, "DashSimpleLine", 120, 40, 2, func(s dashScene) {
		dashSceneSetup(s)
		s.AddDash(10, 6)
		s.MoveTo(10, 20)
		s.LineTo(110, 20)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityDashMultiPattern(t *testing.T) {
	assertDashParity(t, "DashMultiPattern", 120, 60, 2, func(s dashScene) {
		dashSceneSetup(s)
		s.AddDash(12, 4)
		s.AddDash(4, 4)
		s.MoveTo(10, 30)
		s.LineTo(110, 30)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityDashStartOffset(t *testing.T) {
	assertDashParity(t, "DashStartOffset", 120, 40, 2, func(s dashScene) {
		dashSceneSetup(s)
		s.AddDash(10, 6)
		s.DashStart(5)
		s.MoveTo(10, 20)
		s.LineTo(110, 20)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityDashPolyline(t *testing.T) {
	assertDashParity(t, "DashPolyline", 120, 120, 2, func(s dashScene) {
		dashSceneSetup(s)
		s.AddDash(8, 5)
		s.MoveTo(15, 15)
		s.LineTo(105, 15)
		s.LineTo(105, 105)
		s.LineTo(15, 105)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityDashRemoveAll(t *testing.T) {
	// After RemoveAllDashes the stroke must render solid (NumDashes()==0 branch).
	assertDashParity(t, "DashRemoveAll", 120, 40, 2, func(s dashScene) {
		dashSceneSetup(s)
		s.AddDash(10, 6)
		s.RemoveAllDashes()
		s.MoveTo(10, 20)
		s.LineTo(110, 20)
		s.DrawPath(StrokeOnly)
	})
}

func TestParityDashNoDashes(t *testing.T) {
	// NoDashes is the AGG-style alias; same solid result as RemoveAllDashes.
	assertDashParity(t, "DashNoDashes", 120, 40, 2, func(s dashScene) {
		dashSceneSetup(s)
		s.AddDash(10, 6)
		s.NoDashes()
		s.MoveTo(10, 20)
		s.LineTo(110, 20)
		s.DrawPath(StrokeOnly)
	})
}

func TestAgg2DFloatDashStartRoundTrip(t *testing.T) {
	a := NewAgg2DFloat()
	if got := a.GetDashStart(); got != 0.0 {
		t.Fatalf("default GetDashStart = %v, want 0", got)
	}
	a.AddDash(10, 6)
	a.DashStart(7.5)
	if got := a.GetDashStart(); got != 7.5 {
		t.Errorf("GetDashStart after DashStart(7.5) = %v, want 7.5", got)
	}
}
