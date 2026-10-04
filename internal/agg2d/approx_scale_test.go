package agg2d

import (
	"math"
	"testing"
)

// approximation scale = worldToScreen(1.0) * g_approxScale, updated only by
// the C++ transform setters scale/affine/parallelogram/viewport/
// transformations (agg2d.cpp:345-419).
func TestApproximationScaleFollowsCppSetters(t *testing.T) {
	type twin interface {
		Scale(sx, sy float64)
		Rotate(a float64)
		Skew(sx, sy float64)
		Translate(x, y float64)
		ResetTransformations()
		PushTransform()
		PopTransform() bool
		DrawPath(flag DrawPathFlag)
		ResetPath()
		MoveTo(x, y float64)
		LineTo(x, y float64)
	}
	scales := func(g twin) (float64, float64) {
		switch v := g.(type) {
		case *Agg2D:
			return v.convCurve.ApproximationScale(), v.convStroke.ApproximationScale()
		case *Agg2DFloat:
			return v.convCurve.ApproximationScale(), v.convStroke.ApproximationScale()
		}
		panic("unreachable")
	}
	newTwins := func() map[string]twin {
		g := NewAgg2D()
		g.Attach(make([]uint8, 16*16*4), 16, 16, 16*4)
		f := NewAgg2DFloat()
		f.Attach(make([]float32, 16*16*4), 16, 16, 16*4*4)
		return map[string]twin{"8bit": g, "float": f}
	}
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-12 }
	want := func(t *testing.T, name string, g twin, curve, stroke float64) {
		t.Helper()
		if c, s := scales(g); !near(c, curve) || !near(s, stroke) {
			t.Fatalf("%s: approximation scales = %v/%v, want %v/%v", name, c, s, curve, stroke)
		}
	}
	w2s := func(s float64) float64 { return math.Sqrt(2*s*s) * 0.7071068 }

	for name, g := range newTwins() {
		t.Run(name, func(t *testing.T) {
			want(t, "default", g, 1.0, 1.0)

			g.ResetPath()
			g.MoveTo(1, 1)
			g.LineTo(5, 5)
			g.DrawPath(FillAndStroke)
			want(t, "after DrawPath", g, 1.0, 1.0)

			g.Rotate(0.5)
			g.Skew(0.3, 0.1)
			g.Translate(2, 3)
			want(t, "rotate/skew/translate", g, 1.0, 1.0)

			g.ResetTransformations()
			g.Scale(3, 3)
			s := w2s(3) * ApproxScale
			want(t, "scale", g, s, s)

			g.PushTransform()
			g.Scale(2, 2)
			g.PopTransform()
			want(t, "push/scale/pop restores", g, s, s)

			g.ResetTransformations()
			want(t, "resetTransformations keeps", g, s, s)
		})
	}
}

func TestApproximationScaleUserValueSurvivesDrawPath(t *testing.T) {
	g := NewAgg2D()
	g.Attach(make([]uint8, 16*16*4), 16, 16, 16*4)
	g.ApproximationScale(7.5)
	g.ResetPath()
	g.MoveTo(1, 1)
	g.CubicCurveTo(4, 0, 8, 12, 14, 14)
	g.DrawPath(FillAndStroke)
	if got := g.GetApproximationScale(); got != 7.5 {
		t.Fatalf("approximation scale after DrawPath = %v, want 7.5", got)
	}
	g.AddDash(2, 2)
	if got := g.convCurve.ApproximationScale(); got != 7.5 {
		t.Fatalf("curve approximation scale after AddDash = %v, want 7.5", got)
	}
}

// Master alpha must be applied exactly once (rasterizer gamma) on the float
// twin as well.
func TestFloatMasterAlphaAppliedOnce(t *testing.T) {
	const w, h = 16, 16
	buf := make([]float32, w*h*4)
	f := NewAgg2DFloat()
	f.Attach(buf, w, h, w*4*4)
	f.ClearAll(Color{255, 255, 255, 255})
	f.SetMasterAlpha(0.25)
	f.NoLine()
	f.FillColor(Color{255, 0, 0, 255})
	f.ResetPath()
	f.MoveTo(2, 2)
	f.LineTo(14, 2)
	f.LineTo(14, 14)
	f.LineTo(2, 14)
	f.ClosePolygon()
	f.DrawPath(FillOnly)

	// cover = uround(0.25 * 255) = 64 -> green = 1 - 64/255.
	g := buf[(8*w+8)*4+1]
	wantG := float32(1 - 64.0/255.0)
	if math.Abs(float64(g-wantG)) > 1e-4 {
		t.Fatalf("float master alpha 0.25: green = %v, want %v (double application gives %v)",
			g, wantG, 1-16.0/255.0)
	}
}
