package span

import (
	"agg_go/internal/basics"
)

// GouraudDDAInterpolator implements a DDA (Digital Differential Analyzer) line interpolator.
// This is a specialized interpolator for Gouraud shading with configurable fraction precision.
// It's equivalent to AGG's dda_line_interpolator template class.
type GouraudDDAInterpolator struct {
	y             int // current y value
	inc           int // increment per step
	dy            int // accumulated fractional part
	fractionShift int // fraction shift amount
}

// NewGouraudDDAInterpolator creates a new DDA line interpolator.
// Parameters:
//   - y1: starting y value
//   - y2: ending y value
//   - count: number of steps
//   - fractionShift: precision shift amount
func NewGouraudDDAInterpolator(y1, y2 int, count uint, fractionShift int) *GouraudDDAInterpolator {
	if count == 0 {
		count = 1
	}

	inc := ((y2 - y1) << fractionShift) / int(count)
	// Clamp inc to prevent insane values that might lead to numeric instability
	limit := 1000000 // reasonable limit for 14-bit fraction
	if inc > limit {
		inc = limit
	} else if inc < -limit {
		inc = -limit
	}

	return &GouraudDDAInterpolator{
		y:             y1,
		inc:           inc,
		dy:            0,
		fractionShift: fractionShift,
	}
}

// Inc increments the interpolator (equivalent to operator++).
func (d *GouraudDDAInterpolator) Inc() {
	d.dy += d.inc
}

// Dec decrements the interpolator (equivalent to operator--).
func (d *GouraudDDAInterpolator) Dec() {
	d.dy -= d.inc
}

// Add adds n steps to the interpolator (equivalent to operator+=).
func (d *GouraudDDAInterpolator) Add(n uint) {
	d.dy += d.inc * int(n)
}

// Sub subtracts n steps from the interpolator (equivalent to operator-=).
func (d *GouraudDDAInterpolator) Sub(n uint) {
	d.dy -= d.inc * int(n)
}

// Y returns the current y value with fraction shift applied.
func (d *GouraudDDAInterpolator) Y() int {
	return d.y + (d.dy >> d.fractionShift)
}

// DY returns the raw fractional accumulator value.
func (d *GouraudDDAInterpolator) DY() int {
	return d.dy
}

// GouraudDda2Interpolator implements a more precise DDA line interpolator.
// This is equivalent to AGG's dda2_line_interpolator class and is used
// for high-precision interpolation in span generation.
type GouraudDda2Interpolator struct {
	cnt int // count (divisor)
	lft int // left part (quotient)
	rem int // remainder
	mod int // modulo accumulator
	y   int // current y value
}

// GouraudSaveData represents the saved state of a GouraudDda2Interpolator.
type GouraudSaveData struct {
	Mod int
	Y   int
}

// NewGouraudDda2Interpolator creates a new DDA2 line interpolator with forward adjustment.
func NewGouraudDda2Interpolator(y1, y2, count int) *GouraudDda2Interpolator {
	if count <= 0 {
		count = 1
	}

	d := &GouraudDda2Interpolator{
		cnt: count,
		lft: (y2 - y1) / count,
		rem: (y2 - y1) % count,
		y:   y1,
	}

	d.mod = d.rem

	if d.mod <= 0 {
		d.mod += count
		d.rem += count
		d.lft--
	}
	d.mod -= count

	return d
}

// NewDda2LineInterpolatorBackward creates a new DDA2 line interpolator with backward adjustment.
func NewGouraudDda2InterpolatorBackward(y1, y2, count int) *GouraudDda2Interpolator {
	if count <= 0 {
		count = 1
	}

	d := &GouraudDda2Interpolator{
		cnt: count,
		lft: (y2 - y1) / count,
		rem: (y2 - y1) % count,
		y:   y1,
	}

	d.mod = d.rem

	if d.mod <= 0 {
		d.mod += count
		d.rem += count
		d.lft--
	}

	return d
}

// NewDda2LineInterpolatorSimple creates a new DDA2 line interpolator for simple cases.
func NewGouraudDda2InterpolatorSimple(y, count int) *GouraudDda2Interpolator {
	if count <= 0 {
		count = 1
	}

	d := &GouraudDda2Interpolator{
		cnt: count,
		lft: y / count,
		rem: y % count,
		y:   0,
	}

	d.mod = d.rem

	if d.mod <= 0 {
		d.mod += count
		d.rem += count
		d.lft--
	}

	return d
}

// Save saves the current state of the interpolator.
func (d *GouraudDda2Interpolator) Save() GouraudSaveData {
	return GouraudSaveData{
		Mod: d.mod,
		Y:   d.y,
	}
}

// Load restores the state of the interpolator.
func (d *GouraudDda2Interpolator) Load(data GouraudSaveData) {
	d.mod = data.Mod
	d.y = data.Y
}

// Inc increments the interpolator (equivalent to operator++).
func (d *GouraudDda2Interpolator) Inc() {
	d.mod += d.rem
	d.y += d.lft
	if d.mod > 0 {
		d.mod -= d.cnt
		d.y++
	}
}

// Dec decrements the interpolator (equivalent to operator--).
func (d *GouraudDda2Interpolator) Dec() {
	if d.mod <= d.rem {
		d.mod += d.cnt
		d.y--
	}
	d.mod -= d.rem
	d.y -= d.lft
}

// Y returns the current y value.
func (d *GouraudDda2Interpolator) Y() int {
	return d.y
}

// Mod returns the current modulo value.
func (d *GouraudDda2Interpolator) Mod() int {
	return d.mod
}

// Rem returns the remainder value.
func (d *GouraudDda2Interpolator) Rem() int {
	return d.rem
}

// FastRound provides fast integer rounding for performance-critical paths.
func FastRound(v float64) int {
	return basics.IRound(v)
}
