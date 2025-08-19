// Package span provides span interpolation functionality for AGG.
// This implements a port of AGG's span interpolator classes.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// SpanInterpolatorInterface defines the interface for span interpolators.
// This is used by span generators that need coordinate transformation during iteration.
type SpanInterpolatorInterface interface {
	// Begin starts interpolation for a span of pixels
	Begin(x, y float64, length int)

	// Resynchronize adjusts the interpolation to end at the specified coordinates
	Resynchronize(xe, ye float64, length int)

	// Coordinates returns the current transformed coordinates
	Coordinates() (x, y int)

	// Next advances to the next pixel (equivalent to operator++)
	Next()

	// SubpixelShift returns the subpixel precision shift
	SubpixelShift() int
}

// SpanInterpolatorLinear implements linear span interpolation with affine transformation.
// This is a port of AGG's span_interpolator_linear template class.
type SpanInterpolatorLinear[T TransformerInterface] struct {
	transformer   T
	subpixelShift int
	subpixelScale int
	liX           *Dda2LineInterpolator
	liY           *Dda2LineInterpolator
}

// TransformerInterface defines the interface for coordinate transformers.
type TransformerInterface interface {
	Transform(x, y *float64)
}

// Dda2LineInterpolator implements a simple DDA line interpolator for span interpolation.
// This is adapted from the existing implementation in primitives package.
type Dda2LineInterpolator struct {
	cnt int // count remaining
	lft int // left (whole part)
	rem int // remainder
	mod int // modulo
	y   int // current value
}

// NewDda2LineInterpolator creates a new DDA2 line interpolator.
func NewDda2LineInterpolator(y1, y2, count int) *Dda2LineInterpolator {
	if count <= 0 {
		count = 1
	}

	dy := y2 - y1
	return &Dda2LineInterpolator{
		cnt: count,
		lft: dy / count,
		rem: dy % count,
		mod: dy % count,
		y:   y1,
	}
}

// Y returns the current interpolated value.
func (d *Dda2LineInterpolator) Y() int {
	return d.y
}

// Inc advances the interpolator to the next position (equivalent to operator++).
func (d *Dda2LineInterpolator) Inc() {
	d.rem += d.mod
	if d.rem < 0 {
		d.rem += d.cnt
		d.y += d.lft - 1
	} else {
		d.y += d.lft
		if d.rem >= d.cnt {
			d.rem -= d.cnt
			d.y++
		}
	}
}

// NewSpanInterpolatorLinear creates a new linear span interpolator.
func NewSpanInterpolatorLinear[T TransformerInterface](transformer T, subpixelShift int) *SpanInterpolatorLinear[T] {
	if subpixelShift == 0 {
		subpixelShift = 8 // Default subpixel shift from AGG
	}

	return &SpanInterpolatorLinear[T]{
		transformer:   transformer,
		subpixelShift: subpixelShift,
		subpixelScale: 1 << subpixelShift,
	}
}

// NewSpanInterpolatorLinearDefault creates a linear span interpolator with TransAffine.
func NewSpanInterpolatorLinearDefault(transformer *transform.TransAffine) *SpanInterpolatorLinear[*transform.TransAffine] {
	return NewSpanInterpolatorLinear(transformer, 8)
}

// Begin starts interpolation for a span of pixels from (x,y) with specified length.
func (s *SpanInterpolatorLinear[T]) Begin(x, y float64, length int) {
	// Transform start point
	tx, ty := x, y
	s.transformer.Transform(&tx, &ty)
	x1 := basics.IRound(tx * float64(s.subpixelScale))
	y1 := basics.IRound(ty * float64(s.subpixelScale))

	// Transform end point
	tx, ty = x+float64(length), y
	s.transformer.Transform(&tx, &ty)
	x2 := basics.IRound(tx * float64(s.subpixelScale))
	y2 := basics.IRound(ty * float64(s.subpixelScale))

	// Create DDA interpolators for X and Y coordinates
	s.liX = NewDda2LineInterpolator(x1, x2, length)
	s.liY = NewDda2LineInterpolator(y1, y2, length)
}

// Resynchronize adjusts the interpolation to end at the specified coordinates.
func (s *SpanInterpolatorLinear[T]) Resynchronize(xe, ye float64, length int) {
	// Transform end point
	s.transformer.Transform(&xe, &ye)

	// Create new interpolators from current position to end point
	s.liX = NewDda2LineInterpolator(s.liX.Y(), basics.IRound(xe*float64(s.subpixelScale)), length)
	s.liY = NewDda2LineInterpolator(s.liY.Y(), basics.IRound(ye*float64(s.subpixelScale)), length)
}

// Coordinates returns the current transformed coordinates.
func (s *SpanInterpolatorLinear[T]) Coordinates() (x, y int) {
	return s.liX.Y(), s.liY.Y()
}

// Next advances the interpolator to the next pixel position.
func (s *SpanInterpolatorLinear[T]) Next() {
	s.liX.Inc()
	s.liY.Inc()
}

// SubpixelShift returns the subpixel precision shift value.
func (s *SpanInterpolatorLinear[T]) SubpixelShift() int {
	return s.subpixelShift
}

// Transformer returns the current transformer.
func (s *SpanInterpolatorLinear[T]) Transformer() T {
	return s.transformer
}

// SetTransformer sets a new transformer.
func (s *SpanInterpolatorLinear[T]) SetTransformer(transformer T) {
	s.transformer = transformer
}
