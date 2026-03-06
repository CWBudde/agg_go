package primitives

import (
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────
// LineDBLHR / LineCoordSatConv (line_aa_basics.go)
// ──────────────────────────────────────────────────────────────────────────

func TestLineDBLHR(t *testing.T) {
	tests := []struct {
		in, want int
	}{
		{0, 0},
		{1, 256},
		{2, 512},
		{10, 2560},
	}
	for _, tt := range tests {
		if got := LineDBLHR(tt.in); got != tt.want {
			t.Errorf("LineDBLHR(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestLineCoordSatConv(t *testing.T) {
	// Normal values
	if got := LineCoordSatConv(1.0); got != 256 {
		t.Errorf("LineCoordSatConv(1.0) = %d, want 256", got)
	}
	if got := LineCoordSatConv(0.0); got != 0 {
		t.Errorf("LineCoordSatConv(0.0) = %d, want 0", got)
	}
	// Saturation: very large value should be clamped to LineMaxCoord
	huge := float64(LineMaxCoord+1000) / float64(LineSubpixelScale)
	if got := LineCoordSatConv(huge); got != LineMaxCoord {
		t.Errorf("LineCoordSatConv(huge) = %d, want %d", got, LineMaxCoord)
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Dda2LineInterpolator
// ──────────────────────────────────────────────────────────────────────────

func TestDda2LineInterpolatorInc(t *testing.T) {
	d := NewDda2LineInterpolator(0, 10, 5)
	if d.Y() != 0 {
		t.Fatalf("initial Y = %d, want 0", d.Y())
	}
	d.Inc()
	if d.Y() != 2 {
		t.Errorf("after one Inc Y = %d, want 2", d.Y())
	}
}

func TestDda2LineInterpolatorDec(t *testing.T) {
	d := NewDda2LineInterpolator(0, 10, 5)
	// Walk forward then back
	d.Inc()
	d.Inc()
	d.Dec()
	if d.Y() != 2 {
		t.Errorf("after Inc*2 then Dec, Y = %d, want 2", d.Y())
	}
}

func TestDda2LineInterpolatorDecInc(t *testing.T) {
	d := NewDda2LineInterpolator(0, 10, 5)
	d.Inc()
	y1 := d.Y()
	d.DecInc() // should behave like Dec
	if d.Y() != 0 {
		t.Errorf("DecInc after Inc: Y = %d, want 0 (undoes the Inc)", d.Y())
	}
	_ = y1
}

func TestDda2LineInterpolatorAdjustForward(t *testing.T) {
	d := NewDda2LineInterpolator(0, 10, 5)
	// AdjustForward shifts the modulus down by cnt, must not panic
	d.AdjustForward()
	_ = d.Y()
}

func TestDda2LineInterpolatorZeroCount(t *testing.T) {
	// count=0 must be guarded (constructor clamps to 1)
	d := NewDda2LineInterpolator(0, 10, 0)
	_ = d.Y()
}

// ──────────────────────────────────────────────────────────────────────────
// EllipseBresenhamInterpolator
// ──────────────────────────────────────────────────────────────────────────

func TestEllipseBresenhamBasic(t *testing.T) {
	e := NewEllipseBresenhamInterpolator(5, 3)
	if e == nil {
		t.Fatal("NewEllipseBresenhamInterpolator returned nil")
	}
	// Initial dx/dy should be 0
	if e.Dx() != 0 {
		t.Errorf("initial Dx = %d, want 0", e.Dx())
	}
	if e.Dy() != 0 {
		t.Errorf("initial Dy = %d, want 0", e.Dy())
	}
}

func TestEllipseBresenhamInc(t *testing.T) {
	e := NewEllipseBresenhamInterpolator(5, 3)
	// After several increments dx and dy should be 0 or 1 only
	for i := 0; i < 20; i++ {
		e.Inc()
		dx := e.Dx()
		dy := e.Dy()
		if dx < 0 || dx > 1 {
			t.Errorf("step %d: Dx = %d, want 0 or 1", i, dx)
		}
		if dy < 0 || dy > 1 {
			t.Errorf("step %d: Dy = %d, want 0 or 1", i, dy)
		}
	}
}

func TestEllipseBresenhamCircle(t *testing.T) {
	// For a circle (rx==ry) each step should advance by 1 in at least one direction
	e := NewEllipseBresenhamInterpolator(4, 4)
	for i := 0; i < 12; i++ {
		e.Inc()
		if e.Dx() == 0 && e.Dy() == 0 {
			t.Errorf("step %d: both Dx and Dy are 0 unexpectedly", i)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────
// LineBresenhamInterpolator
// ──────────────────────────────────────────────────────────────────────────

func TestLineBresenhamInterpolatorHorizontal(t *testing.T) {
	// Horizontal line in subpixel coords
	scale := 256
	li := NewLineBresenhamInterpolator(0, 0, 10*scale, 0)

	if li.IsVer() {
		t.Error("horizontal line should not be vertical")
	}
	if li.Len() != 10 {
		t.Errorf("Len = %d, want 10", li.Len())
	}

	// LineLr converts subpixel → pixel
	if got := li.LineLr(256); got != 1 {
		t.Errorf("LineLr(256) = %d, want 1", got)
	}

	x0 := li.X1()
	_ = li.Y1()

	// Step forward via HStep
	li.HStep()
	if li.X1() <= x0 {
		t.Errorf("X1 after HStep = %d, should be > %d", li.X1(), x0)
	}
}

func TestLineBresenhamInterpolatorVertical(t *testing.T) {
	scale := 256
	li := NewLineBresenhamInterpolator(0, 0, 0, 10*scale)

	if !li.IsVer() {
		t.Error("vertical line should report IsVer() = true")
	}
	if li.Len() != 10 {
		t.Errorf("Len = %d, want 10", li.Len())
	}

	y0 := li.Y1()

	// Step forward via VStep
	li.VStep()
	if li.Y1() <= y0 {
		t.Errorf("Y1 after VStep = %d, should be > %d", li.Y1(), y0)
	}
}

func TestLineBresenhamInterpolatorX2Y2(t *testing.T) {
	scale := 256
	li := NewLineBresenhamInterpolator(0, 0, 6*scale, 3*scale)

	// X2/Y2 come from the DDA interpolator – just verify they don't panic and return ints
	_ = li.X2()
	_ = li.Y2()
}

func TestLineBresenhamInterpolatorDiagonal(t *testing.T) {
	scale := 256
	li := NewLineBresenhamInterpolator(0, 0, 4*scale, 4*scale)
	// dy == dx → vertical
	if !li.IsVer() {
		t.Log("diagonal line classified as horizontal (dy==dx tie-break)")
	}
	// Still able to step
	for i := 0; i < li.Len(); i++ {
		if li.IsVer() {
			li.VStep()
		} else {
			li.HStep()
		}
	}
}
