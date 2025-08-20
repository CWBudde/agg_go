package span

// Distortion represents a coordinate distortion transformation.
// This interface matches the C++ distortion concept used with span_interpolator_adaptor.
type Distortion interface {
	// Calculate applies the distortion to the given coordinates.
	// The coordinates are passed by pointer and modified in place.
	// Coordinates are typically in subpixel units.
	Calculate(x, y *int)
}

// SpanInterpolatorAdaptor is a wrapper around a base interpolator that applies
// coordinate distortion. This is a port of AGG's span_interpolator_adaptor template class.
// The adaptor delegates all interpolation to the base interpolator and then applies
// distortion to the final coordinates.
type SpanInterpolatorAdaptor[I SpanInterpolatorInterface, D Distortion] struct {
	base       I
	distortion D
}

// NewSpanInterpolatorAdaptor creates a new interpolator adaptor.
func NewSpanInterpolatorAdaptor[I SpanInterpolatorInterface, D Distortion](base I, distortion D) *SpanInterpolatorAdaptor[I, D] {
	return &SpanInterpolatorAdaptor[I, D]{
		base:       base,
		distortion: distortion,
	}
}

// Base returns the underlying base interpolator.
func (s *SpanInterpolatorAdaptor[I, D]) Base() I {
	return s.base
}

// SetBase sets a new base interpolator.
func (s *SpanInterpolatorAdaptor[I, D]) SetBase(base I) {
	s.base = base
}

// Distortion returns the current distortion.
func (s *SpanInterpolatorAdaptor[I, D]) Distortion() D {
	return s.distortion
}

// SetDistortion sets a new distortion.
func (s *SpanInterpolatorAdaptor[I, D]) SetDistortion(distortion D) {
	s.distortion = distortion
}

// Begin starts interpolation at the given coordinates for the specified length.
// This delegates to the base interpolator.
func (s *SpanInterpolatorAdaptor[I, D]) Begin(x, y float64, length int) {
	s.base.Begin(x, y, length)
}

// Next advances to the next pixel along the span.
// This delegates to the base interpolator.
func (s *SpanInterpolatorAdaptor[I, D]) Next() {
	s.base.Next()
}

// Coordinates returns the current transformed coordinates with distortion applied.
// This gets coordinates from the base interpolator and then applies the distortion.
func (s *SpanInterpolatorAdaptor[I, D]) Coordinates() (x, y int) {
	x, y = s.base.Coordinates()
	s.distortion.Calculate(&x, &y)
	return x, y
}

// Resynchronize adjusts the interpolation to end at the specified coordinates.
// This delegates to the base interpolator.
func (s *SpanInterpolatorAdaptor[I, D]) Resynchronize(xe, ye float64, length int) {
	s.base.Resynchronize(xe, ye, length)
}

// SubpixelShift returns the subpixel precision shift from the base interpolator.
// This delegates to the base interpolator.
func (s *SpanInterpolatorAdaptor[I, D]) SubpixelShift() int {
	return s.base.SubpixelShift()
}
