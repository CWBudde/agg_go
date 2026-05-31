package agg2d

import (
	"testing"
)

// setupFloatTarget makes a float context rendering into a fresh w x h image.
func setupFloatTarget(w, h int) (*Agg2DFloat, *ImageFloat) {
	a := newAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	a.AttachImageFloat(img)
	a.ClearAll(NewColor(0, 0, 0, 0))
	return a, img
}

func TestAgg2DFloatSolidFill(t *testing.T) {
	a, img := setupFloatTarget(20, 20)

	a.FillColor(NewColor(255, 0, 0, 255))
	a.ResetPath()
	a.MoveTo(2, 2)
	a.LineTo(18, 2)
	a.LineTo(18, 18)
	a.LineTo(2, 18)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	center := img.GetPixel(10, 10)
	if !approxF(center.R, 1.0) || !approxF(center.G, 0.0) || !approxF(center.B, 0.0) || !approxF(center.A, 1.0) {
		t.Fatalf("center pixel = %+v, want opaque red {1,0,0,1}", center)
	}
	corner := img.GetPixel(0, 0)
	if corner.A != 0 {
		t.Fatalf("corner pixel alpha = %v, want 0 (untouched)", corner.A)
	}
}

func TestAgg2DFloatSolidStroke(t *testing.T) {
	a, img := setupFloatTarget(20, 20)

	a.LineColor(NewColor(0, 255, 0, 255))
	a.LineWidth(4)
	a.ResetPath()
	a.MoveTo(2, 10)
	a.LineTo(18, 10)
	a.DrawPath(StrokeOnly)

	onLine := img.GetPixel(10, 10)
	if onLine.A <= 0 {
		t.Fatalf("on-line pixel alpha = %v, want > 0", onLine.A)
	}
	if onLine.G <= 0.5 {
		t.Fatalf("on-line pixel green = %v, want dominant green", onLine.G)
	}
	off := img.GetPixel(10, 2)
	if off.A != 0 {
		t.Fatalf("off-line pixel alpha = %v, want 0", off.A)
	}
}

func TestAgg2DFloatRectangleConvenience(t *testing.T) {
	a, img := setupFloatTarget(20, 20)
	a.FillColor(NewColor(0, 0, 255, 255))
	a.LineColor(NewColor(0, 0, 255, 255))
	a.LineWidth(1)
	a.Rectangle(3, 3, 17, 17)

	center := img.GetPixel(10, 10)
	if !approxF(center.B, 1.0) || center.A <= 0 {
		t.Fatalf("center pixel = %+v, want blue interior", center)
	}
}

func TestAgg2DFloatLinearGradientFill(t *testing.T) {
	a, img := setupFloatTarget(40, 10)

	a.FillLinearGradient(0, 0, 40, 0, NewColor(255, 0, 0, 255), NewColor(0, 0, 255, 255), 1.0)
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(40, 0)
	a.LineTo(40, 10)
	a.LineTo(0, 10)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	left := img.GetPixel(3, 5)
	right := img.GetPixel(36, 5)
	// Left end should be more red, right end more blue.
	if left.R <= right.R {
		t.Errorf("expected left redder than right: left.R=%v right.R=%v", left.R, right.R)
	}
	if right.B <= left.B {
		t.Errorf("expected right bluer than left: left.B=%v right.B=%v", left.B, right.B)
	}
}

func TestAgg2DFloatRadialGradientFill(t *testing.T) {
	a, img := setupFloatTarget(30, 30)

	a.FillRadialGradient(15, 15, 14, NewColor(255, 255, 255, 255), NewColor(0, 0, 0, 255), 1.0)
	a.ResetPath()
	a.MoveTo(1, 1)
	a.LineTo(29, 1)
	a.LineTo(29, 29)
	a.LineTo(1, 29)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	center := img.GetPixel(15, 15)
	edge := img.GetPixel(2, 15)
	// Centre is white (bright), edge is dark.
	if center.R <= edge.R {
		t.Errorf("expected centre brighter than edge: center.R=%v edge.R=%v", center.R, edge.R)
	}
	if center.A <= 0 {
		t.Errorf("centre alpha = %v, want > 0", center.A)
	}
}
