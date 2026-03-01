package agg

import "testing"

func TestAgg2DPublicWrappers(t *testing.T) {
	a := NewAgg2D()
	buf := make([]uint8, 16*16*4)
	a.Attach(buf, 16, 16, 16*4)

	x1, y1, x2, y2 := 1.0, 2.0, 10.0, 12.0
	a.ClipBox(x1, y1, x2, y2)
	gx1, gy1, gx2, gy2 := a.GetClipBox()
	if gx1 != x1 || gy1 != y1 || gx2 != x2 || gy2 != y2 {
		t.Fatalf("GetClipBox() = (%v, %v, %v, %v), want (%v, %v, %v, %v)", gx1, gy1, gx2, gy2, x1, y1, x2, y2)
	}

	tr := Translation(3, 4)
	a.SetTransformations(tr)
	got := a.GetTransformations()
	if got == nil || got.AffineMatrix != tr.AffineMatrix {
		t.Fatalf("GetTransformations() = %#v, want %#v", got, tr)
	}

	a.PushTransform()
	a.Translate(5, 6)
	if !a.PopTransform() {
		t.Fatal("PopTransform() = false, want true")
	}
	got = a.GetTransformations()
	if got == nil || got.AffineMatrix != tr.AffineMatrix {
		t.Fatalf("transform after PopTransform() = %#v, want %#v", got, tr)
	}

	a.LineCap(CapSquare)
	if a.GetLineCap() != CapSquare {
		t.Fatalf("GetLineCap() = %v, want %v", a.GetLineCap(), CapSquare)
	}
	a.LineJoin(JoinBevel)
	if a.GetLineJoin() != JoinBevel {
		t.Fatalf("GetLineJoin() = %v, want %v", a.GetLineJoin(), JoinBevel)
	}

	a.ImageBlendMode(BlendMultiply)
	if a.GetImageBlendMode() != BlendMultiply {
		t.Fatalf("GetImageBlendMode() = %v, want %v", a.GetImageBlendMode(), BlendMultiply)
	}
	blendColor := Color{R: 10, G: 20, B: 30, A: 40}
	a.ImageBlendColor(blendColor)
	if gotColor := a.GetImageBlendColor(); gotColor != blendColor {
		t.Fatalf("GetImageBlendColor() = %#v, want %#v", gotColor, blendColor)
	}

	a.AntiAliasGamma(2.0)
	if gotGamma := a.GetAntiAliasGamma(); gotGamma != 2.0 {
		t.Fatalf("GetAntiAliasGamma() = %v, want 2.0", gotGamma)
	}

	a.MoveTo(1, 1)
	a.MoveRel(1, 0)
	a.LineTo(4, 4)
	a.HorLineTo(6)
	a.VerLineTo(7)
	a.ArcTo(2, 2, 0.5, false, true, 8, 8)
	a.QuadricCurveTo(8, 9, 10, 11)
	a.QuadricCurveRel(1, 1, 2, 2)
	a.QuadricCurveToSmooth(12, 13)
	a.QuadricCurveRelSmooth(1, 1)
	a.CubicCurveTo(1, 2, 3, 4, 5, 6)
	a.CubicCurveRel(1, 1, 2, 2, 3, 3)
	a.CubicCurveToSmooth(7, 8, 9, 10)
	a.CubicCurveRelSmooth(1, 2, 3, 4)
	a.DrawPathNoTransform(StrokeOnly)

	a.Triangle(1, 1, 5, 1, 3, 4)
	a.RoundedRectXY(1, 1, 8, 8, 2, 3)
	a.RoundedRectVariableRadii(1, 1, 8, 8, 2, 2, 3, 3)
	a.Arc(5, 5, 3, 2, 0, 1.5)
	a.Star(5, 5, 2, 4, 0.3, 5)
	a.Curve(0, 0, 2, 3, 4, 5)
	a.Curve4(0, 0, 2, 3, 4, 5, 6, 7)
	a.Polygon([]float64{1, 1, 4, 1, 4, 4}, 3)
	a.Polyline([]float64{1, 1, 2, 2, 3, 3}, 3)
}
