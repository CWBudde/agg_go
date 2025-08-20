package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

const (
	defaultSubpixelShift = 8
	defaultSubpixelScale = 1 << defaultSubpixelShift
)

// SpanInterpolatorTrans is a horizontal span interpolator that works with arbitrary transformers.
// This is a port of AGG's span_interpolator_trans template class.
// The efficiency highly depends on the operations done in the transformer.
type SpanInterpolatorTrans[T transform.Transformer] struct {
	transformer   T
	subpixelShift int
	subpixelScale int
	x             float64 // Current original x coordinate
	y             float64 // Current original y coordinate
	ix            int     // Current transformed x coordinate (in subpixel units)
	iy            int     // Current transformed y coordinate (in subpixel units)
}

// NewSpanInterpolatorTrans creates a new transformer-based span interpolator.
func NewSpanInterpolatorTrans[T transform.Transformer](transformer T) *SpanInterpolatorTrans[T] {
	return &SpanInterpolatorTrans[T]{
		transformer:   transformer,
		subpixelShift: defaultSubpixelShift,
		subpixelScale: defaultSubpixelScale,
	}
}

// NewSpanInterpolatorTransWithShift creates a new transformer-based span interpolator with custom subpixel shift.
func NewSpanInterpolatorTransWithShift[T transform.Transformer](transformer T, subpixelShift int) *SpanInterpolatorTrans[T] {
	subpixelScale := 1 << subpixelShift
	return &SpanInterpolatorTrans[T]{
		transformer:   transformer,
		subpixelShift: subpixelShift,
		subpixelScale: subpixelScale,
	}
}

// NewSpanInterpolatorTransAtPoint creates a new transformer-based span interpolator and starts at the given point.
func NewSpanInterpolatorTransAtPoint[T transform.Transformer](transformer T, x, y float64, length int) *SpanInterpolatorTrans[T] {
	interp := NewSpanInterpolatorTrans(transformer)
	interp.Begin(x, y, length)
	return interp
}

// Transformer returns the current transformer.
func (s *SpanInterpolatorTrans[T]) Transformer() T {
	return s.transformer
}

// SetTransformer sets a new transformer.
func (s *SpanInterpolatorTrans[T]) SetTransformer(transformer T) {
	s.transformer = transformer
}

// Begin starts interpolation at the given coordinates for the specified length.
// The length parameter is ignored as this interpolator transforms each pixel individually.
func (s *SpanInterpolatorTrans[T]) Begin(x, y float64, length int) {
	s.x = x
	s.y = y

	// Transform the starting coordinates
	tx, ty := x, y
	s.transformer.Transform(&tx, &ty)

	// Convert to subpixel coordinates
	s.ix = basics.IRound(tx * float64(s.subpixelScale))
	s.iy = basics.IRound(ty * float64(s.subpixelScale))
}

// Next advances to the next pixel along the span (equivalent to C++ operator++).
func (s *SpanInterpolatorTrans[T]) Next() {
	s.x += 1.0

	// Transform the new coordinates
	tx, ty := s.x, s.y
	s.transformer.Transform(&tx, &ty)

	// Convert to subpixel coordinates
	s.ix = basics.IRound(tx * float64(s.subpixelScale))
	s.iy = basics.IRound(ty * float64(s.subpixelScale))
}

// Coordinates returns the current transformed coordinates in subpixel units.
func (s *SpanInterpolatorTrans[T]) Coordinates() (x, y int) {
	return s.ix, s.iy
}

// SubpixelShift returns the subpixel shift value used for coordinate precision.
func (s *SpanInterpolatorTrans[T]) SubpixelShift() int {
	return s.subpixelShift
}

// SubpixelScale returns the subpixel scale value (1 << SubpixelShift).
func (s *SpanInterpolatorTrans[T]) SubpixelScale() int {
	return s.subpixelScale
}

// Resynchronize is not supported by this interpolator type.
// This is a no-op to maintain interface compatibility.
func (s *SpanInterpolatorTrans[T]) Resynchronize(xe, ye float64, length int) {
	// This interpolator doesn't support resynchronization as it transforms
	// each pixel individually. This method is included for interface compatibility.
}
