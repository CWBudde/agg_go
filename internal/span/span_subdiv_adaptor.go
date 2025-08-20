package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// LocalScaler defines the interface for interpolators that support local scaling.
// This interface replaces duck-typing patterns for accessing local scale factors.
type LocalScaler interface {
	// LocalScale returns the local scale factors
	LocalScale() (x, y int)
}

// SubdivInterpolator represents an interpolator that supports resynchronization,
// which is required for subdivision adaptor functionality.
type SubdivInterpolator interface {
	// Begin starts interpolation for a span of pixels
	Begin(x, y float64, length int)

	// Next advances to the next pixel
	Next()

	// Resynchronize adjusts the interpolation to end at the specified coordinates
	Resynchronize(xe, ye float64, length int)

	// Coordinates returns the current transformed coordinates
	Coordinates() (x, y int)
}

// TransformerAccessor represents an interpolator that provides access to its transformer.
type TransformerAccessor[T transform.Transformer] interface {
	Transformer() T
	SetTransformer(transformer T)
}

// SpanSubdivAdaptor is a subdivision adaptor for span interpolators.
// This is a port of AGG's span_subdiv_adaptor template class.
// It subdivides long spans into smaller chunks and resynchronizes the interpolator
// periodically to prevent accumulation of rounding errors and improve performance.
type SpanSubdivAdaptor[I SubdivInterpolator] struct {
	interpolator  I
	subdivShift   int     // Subdivision shift (power of 2)
	subdivSize    int     // Subdivision size (1 << subdivShift)
	subdivMask    int     // Subdivision mask (subdivSize - 1)
	subpixelShift int     // Subpixel shift for coordinate precision
	subpixelScale int     // Subpixel scale (1 << subpixelShift)
	pos           int     // Current position within subdivision
	srcX          int     // Source X coordinate in subpixel units
	srcY          float64 // Source Y coordinate
	len           int     // Remaining length of span
}

// NewSpanSubdivAdaptor creates a new subdivision adaptor with default settings.
// Default subdivision shift is 4 (16 pixels), default subpixel shift is 8.
func NewSpanSubdivAdaptor[I SubdivInterpolator](interpolator I) *SpanSubdivAdaptor[I] {
	return NewSpanSubdivAdaptorWithShifts(interpolator, 4, defaultSubpixelShift)
}

// NewSpanSubdivAdaptorWithShift creates a new subdivision adaptor with custom subdivision shift.
// subdivision shift determines the chunk size: chunk_size = 1 << subdivShift
func NewSpanSubdivAdaptorWithShift[I SubdivInterpolator](interpolator I, subdivShift int) *SpanSubdivAdaptor[I] {
	return NewSpanSubdivAdaptorWithShifts(interpolator, subdivShift, defaultSubpixelShift)
}

// NewSpanSubdivAdaptorWithShifts creates a new subdivision adaptor with custom subdivision and subpixel shifts.
func NewSpanSubdivAdaptorWithShifts[I SubdivInterpolator](interpolator I, subdivShift, subpixelShift int) *SpanSubdivAdaptor[I] {
	subpixelScale := 1 << subpixelShift
	return &SpanSubdivAdaptor[I]{
		interpolator:  interpolator,
		subdivShift:   subdivShift,
		subdivSize:    1 << subdivShift,
		subdivMask:    (1 << subdivShift) - 1,
		subpixelShift: subpixelShift,
		subpixelScale: subpixelScale,
	}
}

// NewSpanSubdivAdaptorAtPoint creates a new subdivision adaptor and starts at the given point.
func NewSpanSubdivAdaptorAtPoint[I SubdivInterpolator](interpolator I, x, y float64, length int) *SpanSubdivAdaptor[I] {
	adaptor := NewSpanSubdivAdaptor(interpolator)
	adaptor.Begin(x, y, length)
	return adaptor
}

// Interpolator returns the current wrapped interpolator.
func (s *SpanSubdivAdaptor[I]) Interpolator() I {
	return s.interpolator
}

// SetInterpolator sets a new wrapped interpolator.
func (s *SpanSubdivAdaptor[I]) SetInterpolator(interpolator I) {
	s.interpolator = interpolator
}

// Transformer returns the transformer from the wrapped interpolator if it supports the interface.
func (s *SpanSubdivAdaptor[I]) Transformer() transform.Transformer {
	if transformerGetter, ok := any(s.interpolator).(transform.TransformerGetter); ok {
		return transformerGetter.Transformer()
	}
	return nil
}

// SetTransformer sets the transformer on the wrapped interpolator if it supports the interface.
func (s *SpanSubdivAdaptor[I]) SetTransformer(transformer transform.Transformer) {
	if transformerSetter, ok := any(s.interpolator).(transform.TransformerSetter); ok {
		transformerSetter.SetTransformer(transformer)
	}
}

// SubdivShift returns the current subdivision shift value.
func (s *SpanSubdivAdaptor[I]) SubdivShift() int {
	return s.subdivShift
}

// SetSubdivShift sets a new subdivision shift value and recalculates size and mask.
func (s *SpanSubdivAdaptor[I]) SetSubdivShift(shift int) {
	s.subdivShift = shift
	s.subdivSize = 1 << shift
	s.subdivMask = s.subdivSize - 1
}

// SubpixelShift returns the subpixel shift value used for coordinate precision.
func (s *SpanSubdivAdaptor[I]) SubpixelShift() int {
	return s.subpixelShift
}

// SubpixelScale returns the subpixel scale value (1 << SubpixelShift).
func (s *SpanSubdivAdaptor[I]) SubpixelScale() int {
	return s.subpixelScale
}

// Begin starts interpolation at the given coordinates for the specified length.
func (s *SpanSubdivAdaptor[I]) Begin(x, y float64, length int) {
	s.pos = 1
	s.srcX = basics.IRound(x*float64(s.subpixelScale)) + s.subpixelScale
	s.srcY = y
	s.len = length

	// Limit initial interpolation length to subdivision size
	interpLength := length
	if interpLength > s.subdivSize {
		interpLength = s.subdivSize
	}

	s.interpolator.Begin(x, y, interpLength)
}

// Next advances to the next pixel along the span with automatic subdivision resynchronization.
func (s *SpanSubdivAdaptor[I]) Next() {
	s.interpolator.Next()

	if s.pos >= s.subdivSize {
		// We've reached the subdivision boundary, resynchronize
		length := s.len
		if length > s.subdivSize {
			length = s.subdivSize
		}

		// Calculate the new starting point for resynchronization
		newX := float64(s.srcX)/float64(s.subpixelScale) + float64(length)
		s.interpolator.Resynchronize(newX, s.srcY, length)
		s.pos = 0
	}

	s.srcX += s.subpixelScale
	s.pos++
	s.len--
}

// Coordinates returns the current transformed coordinates from the wrapped interpolator.
func (s *SpanSubdivAdaptor[I]) Coordinates() (x, y int) {
	return s.interpolator.Coordinates()
}

// LocalScale returns the local scale factors from the wrapped interpolator if supported.
// Returns (1, 1) if the wrapped interpolator doesn't support LocalScale.
func (s *SpanSubdivAdaptor[I]) LocalScale() (x, y int) {
	if localScaler, ok := any(s.interpolator).(LocalScaler); ok {
		return localScaler.LocalScale()
	}
	// Return default scale if not supported
	return 1, 1
}

// Resynchronize adjusts the interpolation to end at the specified coordinates.
// This delegates to the wrapped interpolator.
func (s *SpanSubdivAdaptor[I]) Resynchronize(xe, ye float64, length int) {
	s.interpolator.Resynchronize(xe, ye, length)
}
