package agg2d

import "testing"

func TestAgg2DFloatTranslate(t *testing.T) {
	a, img := setupFloatTarget(24, 24)
	a.FillColor(NewColor(255, 0, 0, 255))
	a.Translate(10, 10)
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(6, 0)
	a.LineTo(6, 6)
	a.LineTo(0, 6)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	// World (0..6) maps to screen (10..16).
	if c := img.GetPixel(13, 13); !approxF(c.R, 1.0) || c.A <= 0 {
		t.Fatalf("translated fill missing at (13,13): %+v", c)
	}
	if c := img.GetPixel(3, 3); c.A != 0 {
		t.Fatalf("original (untranslated) location should be empty at (3,3): %+v", c)
	}
}

func TestAgg2DFloatScale(t *testing.T) {
	a, img := setupFloatTarget(40, 40)
	a.FillColor(NewColor(0, 0, 255, 255))
	a.Scale(2, 2)
	a.ResetPath()
	a.MoveTo(2, 2)
	a.LineTo(8, 2)
	a.LineTo(8, 8)
	a.LineTo(2, 8)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	// World (2..8) scaled x2 -> screen (4..16).
	if c := img.GetPixel(10, 10); !approxF(c.B, 1.0) || c.A <= 0 {
		t.Fatalf("scaled fill missing at (10,10): %+v", c)
	}
}

func TestAgg2DFloatNoFill(t *testing.T) {
	a, img := setupFloatTarget(24, 24)
	a.LineColor(NewColor(0, 255, 0, 255))
	a.LineWidth(2)
	a.NoFill()
	a.Rectangle(4, 4, 20, 20) // FillAndStroke, but fill is transparent

	if c := img.GetPixel(12, 12); c.A != 0 {
		t.Fatalf("NoFill interior should be empty at (12,12): %+v", c)
	}
	// Border still stroked green.
	if c := img.GetPixel(4, 12); c.A <= 0 || c.G <= 0.5 {
		t.Fatalf("stroke border missing/green at (4,12): %+v", c)
	}
}

func TestAgg2DFloatFillEvenOddDonut(t *testing.T) {
	a, img := setupFloatTarget(30, 30)
	a.FillColor(NewColor(255, 0, 0, 255))
	if a.GetFillEvenOdd() {
		t.Fatal("default fill rule should be non-zero (GetFillEvenOdd=false)")
	}
	a.FillEvenOdd(true)
	if !a.GetFillEvenOdd() {
		t.Fatal("GetFillEvenOdd should be true after FillEvenOdd(true)")
	}

	a.ResetPath()
	// Outer square.
	a.MoveTo(4, 4)
	a.LineTo(26, 4)
	a.LineTo(26, 26)
	a.LineTo(4, 26)
	a.ClosePolygon()
	// Inner square, same winding -> even-odd creates a hole.
	a.MoveTo(11, 11)
	a.LineTo(19, 11)
	a.LineTo(19, 19)
	a.LineTo(11, 19)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	if c := img.GetPixel(15, 15); c.A != 0 {
		t.Fatalf("even-odd hole expected empty at center (15,15): %+v", c)
	}
	if c := img.GetPixel(7, 15); c.A <= 0 {
		t.Fatalf("even-odd ring expected filled at (7,15): %+v", c)
	}
}
