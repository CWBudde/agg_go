package primitives

import (
	"math"
)

// LineBresenhamInterpolator implements Bresenham line interpolation.
// This is a port of AGG's line_bresenham_interpolator class.
type LineBresenhamInterpolator struct {
	// Subpixel scale constants
	SubpixelShift int
	SubpixelScale int
	SubpixelMask  int

	x1Lr         int                   // x1 low resolution
	y1Lr         int                   // y1 low resolution
	x2Lr         int                   // x2 low resolution
	y2Lr         int                   // y2 low resolution
	ver          bool                  // vertical line flag
	length       int                   // line length
	inc          int                   // increment direction
	interpolator *Dda2LineInterpolator // DDA interpolator
}

// Dda2LineInterpolator implements a simple DDA line interpolator.
// This is a port of AGG's dda2_line_interpolator class.
type Dda2LineInterpolator struct {
	cnt int // count
	lft int // left part
	rem int // remainder
	mod int // modulo
	y   int // current y
}

// NewDda2LineInterpolator creates a new DDA2 line interpolator.
func NewDda2LineInterpolator(y1, y2, count int) *Dda2LineInterpolator {
	if count <= 0 {
		count = 1
	}

	return &Dda2LineInterpolator{
		cnt: count,
		lft: (y2 - y1) / count,
		rem: (y2 - y1) % count,
		mod: (y2 - y1) % count,
		y:   y1,
	}
}

// Inc increments the interpolator (equivalent to operator++).
func (d *Dda2LineInterpolator) Inc() {
	d.mod += d.rem
	d.y += d.lft
	if d.mod >= d.cnt {
		d.mod -= d.cnt
		d.y++
	}
}

// Y returns the current y value.
func (d *Dda2LineInterpolator) Y() int {
	return d.y
}

// AdjustForward adjusts the interpolator forward (used in anti-aliased line rendering).
func (d *Dda2LineInterpolator) AdjustForward() {
	d.mod -= d.cnt
}

// NewLineBresenhamInterpolator creates a new line Bresenham interpolator.
func NewLineBresenhamInterpolator(x1, y1, x2, y2 int) *LineBresenhamInterpolator {
	li := &LineBresenhamInterpolator{
		SubpixelShift: 8,
		SubpixelScale: 1 << 8,
		SubpixelMask:  (1 << 8) - 1,
	}

	li.x1Lr = li.LineLr(x1)
	li.y1Lr = li.LineLr(y1)
	li.x2Lr = li.LineLr(x2)
	li.y2Lr = li.LineLr(y2)

	li.ver = int(math.Abs(float64(li.x2Lr-li.x1Lr))) < int(math.Abs(float64(li.y2Lr-li.y1Lr)))

	if li.ver {
		li.length = int(math.Abs(float64(li.y2Lr - li.y1Lr)))
	} else {
		li.length = int(math.Abs(float64(li.x2Lr - li.x1Lr)))
	}

	if li.ver {
		if y2 > y1 {
			li.inc = 1
		} else {
			li.inc = -1
		}
	} else {
		if x2 > x1 {
			li.inc = 1
		} else {
			li.inc = -1
		}
	}

	if li.ver {
		li.interpolator = NewDda2LineInterpolator(x1, x2, li.length)
	} else {
		li.interpolator = NewDda2LineInterpolator(y1, y2, li.length)
	}

	return li
}

// LineLr converts a coordinate to low resolution.
func (li *LineBresenhamInterpolator) LineLr(v int) int {
	return v >> li.SubpixelShift
}

// IsVer returns true if this is a vertical line.
func (li *LineBresenhamInterpolator) IsVer() bool {
	return li.ver
}

// Len returns the length of the line.
func (li *LineBresenhamInterpolator) Len() int {
	return li.length
}

// HStep performs a horizontal step.
func (li *LineBresenhamInterpolator) HStep() {
	li.interpolator.Inc()
	li.x1Lr += li.inc
}

// VStep performs a vertical step.
func (li *LineBresenhamInterpolator) VStep() {
	li.interpolator.Inc()
	li.y1Lr += li.inc
}

// X1 returns the current x1 coordinate.
func (li *LineBresenhamInterpolator) X1() int {
	return li.x1Lr
}

// Y1 returns the current y1 coordinate.
func (li *LineBresenhamInterpolator) Y1() int {
	return li.y1Lr
}

// X2 returns the current x2 coordinate.
func (li *LineBresenhamInterpolator) X2() int {
	return li.LineLr(li.interpolator.Y())
}

// Y2 returns the current y2 coordinate.
func (li *LineBresenhamInterpolator) Y2() int {
	return li.LineLr(li.interpolator.Y())
}
