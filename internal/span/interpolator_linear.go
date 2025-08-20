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

// Dda2LineInterpolator implements DDA2 line interpolator matching AGG's dda2_line_interpolator.
// This is a direct port of the C++ implementation from agg_dda_line.h.
type Dda2LineInterpolator struct {
	cnt int // count
	lft int // left (integer part of increment)
	rem int // remainder
	mod int // modulo (tracking fractional accumulation)
	y   int // current value
}

// NewDda2LineInterpolator creates a new DDA2 line interpolator.
// This implements the forward-adjusted line constructor from C++ AGG.
func NewDda2LineInterpolator(y1, y2, count int) *Dda2LineInterpolator {
	if count <= 0 {
		count = 1
	}

	dy := y2 - y1
	dda := &Dda2LineInterpolator{
		cnt: count,
		lft: dy / count,
		rem: dy % count,
		mod: dy % count,
		y:   y1,
	}

	// Forward adjustment logic from C++ AGG
	if dda.mod <= 0 {
		dda.mod += count
		dda.rem += count
		dda.lft--
	}
	dda.mod -= count

	return dda
}

// Y returns the current interpolated value.
func (d *Dda2LineInterpolator) Y() int {
	return d.y
}

// Inc advances the interpolator to the next position (equivalent to C++ operator++).
// This implements the exact algorithm from C++ AGG dda2_line_interpolator.
func (d *Dda2LineInterpolator) Inc() {
	d.mod += d.rem
	d.y += d.lft
	if d.mod > 0 {
		d.mod -= d.cnt
		d.y++
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

// SpanInterpolatorLinearSubdiv implements subdivided linear span interpolation.
// This is a port of AGG's span_interpolator_linear_subdiv template class.
// It subdivides long spans and resynchronizes periodically to prevent error accumulation.
type SpanInterpolatorLinearSubdiv[T TransformerInterface] struct {
	transformer   T
	subpixelShift int
	subpixelScale int
	subdivShift   int // Subdivision shift (power of 2)
	subdivSize    int // Subdivision size (1 << subdivShift)
	subdivMask    int // Subdivision mask (subdivSize - 1)
	liX           *Dda2LineInterpolator
	liY           *Dda2LineInterpolator
	srcX          int     // Source X in subpixel units
	srcY          float64 // Source Y coordinate
	pos           int     // Current position within subdivision
	length        int     // Remaining span length
}

// NewSpanInterpolatorLinearSubdiv creates a new subdivided linear span interpolator.
func NewSpanInterpolatorLinearSubdiv[T TransformerInterface](transformer T, subpixelShift, subdivShift int) *SpanInterpolatorLinearSubdiv[T] {
	if subpixelShift == 0 {
		subpixelShift = 8 // Default subpixel shift from AGG
	}
	if subdivShift == 0 {
		subdivShift = 4 // Default subdivision shift from AGG
	}

	return &SpanInterpolatorLinearSubdiv[T]{
		transformer:   transformer,
		subpixelShift: subpixelShift,
		subpixelScale: 1 << subpixelShift,
		subdivShift:   subdivShift,
		subdivSize:    1 << subdivShift,
		subdivMask:    (1 << subdivShift) - 1,
	}
}

// NewSpanInterpolatorLinearSubdivDefault creates a subdivided interpolator with TransAffine and default parameters.
func NewSpanInterpolatorLinearSubdivDefault(transformer *transform.TransAffine) *SpanInterpolatorLinearSubdiv[*transform.TransAffine] {
	return NewSpanInterpolatorLinearSubdiv(transformer, 8, 4)
}

// Begin starts interpolation for a span of pixels from (x,y) with specified length.
func (s *SpanInterpolatorLinearSubdiv[T]) Begin(x, y float64, length int) {
	s.pos = 1
	s.srcX = basics.IRound(x*float64(s.subpixelScale)) + s.subpixelScale
	s.srcY = y
	s.length = length

	// Limit subdivision length
	subdivLength := length
	if subdivLength > s.subdivSize {
		subdivLength = s.subdivSize
	}

	// Transform start point
	tx, ty := x, y
	s.transformer.Transform(&tx, &ty)
	x1 := basics.IRound(tx * float64(s.subpixelScale))
	y1 := basics.IRound(ty * float64(s.subpixelScale))

	// Transform subdivision end point
	tx, ty = x+float64(subdivLength), y
	s.transformer.Transform(&tx, &ty)
	x2 := basics.IRound(tx * float64(s.subpixelScale))
	y2 := basics.IRound(ty * float64(s.subpixelScale))

	// Create DDA interpolators for the subdivision
	s.liX = NewDda2LineInterpolator(x1, x2, subdivLength)
	s.liY = NewDda2LineInterpolator(y1, y2, subdivLength)
}

// Resynchronize adjusts the interpolation to end at the specified coordinates.
// This creates a new subdivision from current position to the end point.
func (s *SpanInterpolatorLinearSubdiv[T]) Resynchronize(xe, ye float64, length int) {
	// Transform end point
	s.transformer.Transform(&xe, &ye)

	// Create new interpolators from current position to end point
	s.liX = NewDda2LineInterpolator(s.liX.Y(), basics.IRound(xe*float64(s.subpixelScale)), length)
	s.liY = NewDda2LineInterpolator(s.liY.Y(), basics.IRound(ye*float64(s.subpixelScale)), length)
	s.length = length
	s.pos = 1
}

// Coordinates returns the current transformed coordinates.
func (s *SpanInterpolatorLinearSubdiv[T]) Coordinates() (x, y int) {
	return s.liX.Y(), s.liY.Y()
}

// Next advances the interpolator to the next pixel position.
// Implements subdivision logic with periodic resynchronization.
func (s *SpanInterpolatorLinearSubdiv[T]) Next() {
	s.liX.Inc()
	s.liY.Inc()

	// Check if we need to resynchronize (reached subdivision boundary)
	if s.pos >= s.subdivSize {
		// Calculate remaining subdivision length
		subdivLength := s.length
		if subdivLength > s.subdivSize {
			subdivLength = s.subdivSize
		}

		// Calculate next subdivision start point
		tx := float64(s.srcX) / float64(s.subpixelScale)
		ty := s.srcY
		s.transformer.Transform(&tx, &ty)

		// Calculate subdivision end point
		endX := tx + float64(subdivLength)
		endY := ty
		s.transformer.Transform(&endX, &endY)

		// Create new DDA interpolators for this subdivision
		s.liX = NewDda2LineInterpolator(s.liX.Y(), basics.IRound(endX*float64(s.subpixelScale)), subdivLength)
		s.liY = NewDda2LineInterpolator(s.liY.Y(), basics.IRound(endY*float64(s.subpixelScale)), subdivLength)

		s.pos = 0
	}

	s.srcX += s.subpixelScale
	s.pos++
	s.length--
}

// SubpixelShift returns the subpixel precision shift value.
func (s *SpanInterpolatorLinearSubdiv[T]) SubpixelShift() int {
	return s.subpixelShift
}

// SubdivShift returns the subdivision shift value.
func (s *SpanInterpolatorLinearSubdiv[T]) SubdivShift() int {
	return s.subdivShift
}

// SetSubdivShift sets a new subdivision shift value.
func (s *SpanInterpolatorLinearSubdiv[T]) SetSubdivShift(shift int) {
	s.subdivShift = shift
	s.subdivSize = 1 << shift
	s.subdivMask = (1 << shift) - 1
}

// Transformer returns the current transformer.
func (s *SpanInterpolatorLinearSubdiv[T]) Transformer() T {
	return s.transformer
}

// SetTransformer sets a new transformer.
func (s *SpanInterpolatorLinearSubdiv[T]) SetTransformer(transformer T) {
	s.transformer = transformer
}
