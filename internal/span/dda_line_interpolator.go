package span

// DDALineInterpolator implements a DDA (Digital Differential Analyzer) line interpolator
// specifically for gradient color interpolation. This provides precise fixed-point
// arithmetic for smooth color transitions.
//
// This is equivalent to AGG's dda_line_interpolator<FractionShift> template class
// used in color_interpolator specializations.
type DDALineInterpolator struct {
	y             int // current value
	inc           int // increment per step
	dy            int // accumulated fractional part
	fractionShift int // fraction shift for fixed-point precision
}

// NewDDALineInterpolator creates a new DDA line interpolator for color channel interpolation.
// Parameters:
//   - y1: starting value (0-255 for 8-bit colors)
//   - y2: ending value (0-255 for 8-bit colors)
//   - count: number of interpolation steps
//   - fractionShift: precision shift amount (typically 14 for color interpolation)
func NewDDALineInterpolator(y1, y2 int, count uint, fractionShift int) *DDALineInterpolator {
	if count == 0 {
		count = 1
	}

	return &DDALineInterpolator{
		y:             y1,
		inc:           ((y2 - y1) << fractionShift) / int(count),
		dy:            0,
		fractionShift: fractionShift,
	}
}

// Inc increments the interpolator to the next step.
func (d *DDALineInterpolator) Inc() {
	d.dy += d.inc
}

// Dec decrements the interpolator to the previous step.
func (d *DDALineInterpolator) Dec() {
	d.dy -= d.inc
}

// Add advances the interpolator by n steps.
func (d *DDALineInterpolator) Add(n uint) {
	d.dy += d.inc * int(n)
}

// Sub moves the interpolator back by n steps.
func (d *DDALineInterpolator) Sub(n uint) {
	d.dy -= d.inc * int(n)
}

// Y returns the current interpolated value.
func (d *DDALineInterpolator) Y() int {
	return d.y + (d.dy >> d.fractionShift)
}

// DY returns the raw fractional accumulator value.
func (d *DDALineInterpolator) DY() int {
	return d.dy
}
