// Package primitives provides efficient algorithms for primitive shape drawing.
// This package implements Bresenham-style algorithms for fast integer-based
// drawing operations that don't require anti-aliasing.
package primitives

// EllipseBresenhamInterpolator implements a simple Bresenham interpolator for ellipses.
// This is a direct port of AGG's ellipse_bresenham_interpolator class.
// It uses integer arithmetic to efficiently step around an ellipse perimeter.
type EllipseBresenhamInterpolator struct {
	rx2    int // rx * rx
	ry2    int // ry * ry
	twoRx2 int // rx2 << 1
	twoRy2 int // ry2 << 1
	dx     int // current dx step
	dy     int // current dy step
	incX   int // x increment value
	incY   int // y increment value
	curF   int // current function value
}

// NewEllipseBresenhamInterpolator creates a new ellipse interpolator for the given radii.
func NewEllipseBresenhamInterpolator(rx, ry int) *EllipseBresenhamInterpolator {
	rx2 := rx * rx
	ry2 := ry * ry

	return &EllipseBresenhamInterpolator{
		rx2:    rx2,
		ry2:    ry2,
		twoRx2: rx2 << 1,
		twoRy2: ry2 << 1,
		dx:     0,
		dy:     0,
		incX:   0,
		incY:   -ry * (rx2 << 1),
		curF:   0,
	}
}

// Dx returns the current x step.
func (e *EllipseBresenhamInterpolator) Dx() int {
	return e.dx
}

// Dy returns the current y step.
func (e *EllipseBresenhamInterpolator) Dy() int {
	return e.dy
}

// Inc increments the interpolator to the next point on the ellipse.
// This is equivalent to the operator++ in the C++ version.
func (e *EllipseBresenhamInterpolator) Inc() {
	var mx, my, mxy, minM int
	var fx, fy, fxy int

	fx = e.curF + e.incX + e.ry2
	mx = fx
	if mx < 0 {
		mx = -mx
	}

	fy = e.curF + e.incY + e.rx2
	my = fy
	if my < 0 {
		my = -my
	}

	fxy = e.curF + e.incX + e.ry2 + e.incY + e.rx2
	mxy = fxy
	if mxy < 0 {
		mxy = -mxy
	}

	minM = mx
	flag := true

	if minM > my {
		minM = my
		flag = false
	}

	e.dx = 0
	e.dy = 0

	if minM > mxy {
		e.incX += e.twoRy2
		e.incY += e.twoRx2
		e.curF = fxy
		e.dx = 1
		e.dy = 1
		return
	}

	if flag {
		e.incX += e.twoRy2
		e.curF = fx
		e.dx = 1
		return
	}

	e.incY += e.twoRx2
	e.curF = fy
	e.dy = 1
}
