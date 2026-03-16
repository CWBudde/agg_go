package span

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// SpanInterpolatorInterface is the common contract for AGG span interpolators.
// Generators use it to walk one destination span while sampling in transformed
// source-space coordinates.
type SpanInterpolatorInterface interface {
	Begin(x, y float64, length int)
	Resynchronize(xe, ye float64, length int)
	Coordinates() (x, y int)
	Next()
	SubpixelShift() int
}

// SpanInterpolatorLinear is the Go equivalent of AGG's
// span_interpolator_linear. It transforms the span start and end once, then
// uses DDA interpolation to step between them at fixed-point precision.
type SpanInterpolatorLinear[T TransformerInterface] struct {
	transformer   T
	subpixelShift int
	subpixelScale int
	liX           Dda2LineInterpolator
	liY           Dda2LineInterpolator
}

// TransformerInterface is the minimal transform contract required by the linear
// interpolators.
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
	dda := &Dda2LineInterpolator{}
	dda.Init(y1, y2, count)
	return dda
}

// Init reinitializes the interpolator with a new range.
func (d *Dda2LineInterpolator) Init(y1, y2, count int) {
	if count <= 0 {
		count = 1
	}

	dy := y2 - y1
	d.cnt = count
	d.lft = dy / count
	d.rem = dy % count
	d.mod = dy % count
	d.y = y1

	// Forward adjustment logic from C++ AGG
	if d.mod <= 0 {
		d.mod += count
		d.rem += count
		d.lft--
	}
	d.mod -= count
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

// NewSpanInterpolatorLinear creates a linear span interpolator with the
// requested fixed-point precision.
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

// NewSpanInterpolatorLinearDefault creates the standard affine interpolator used
// by AGG image-filter and gradient paths.
func NewSpanInterpolatorLinearDefault(transformer *transform.TransAffine) *SpanInterpolatorLinear[*transform.TransAffine] {
	return NewSpanInterpolatorLinear(transformer, 8)
}

// Begin starts interpolation for a destination span beginning at x,y.
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
	s.liX.Init(x1, x2, length)
	s.liY.Init(y1, y2, length)
}

// Resynchronize retargets the current DDA state toward a new end point.
func (s *SpanInterpolatorLinear[T]) Resynchronize(xe, ye float64, length int) {
	// Transform end point
	s.transformer.Transform(&xe, &ye)

	// Create new interpolators from current position to end point
	s.liX.Init(s.liX.Y(), basics.IRound(xe*float64(s.subpixelScale)), length)
	s.liY.Init(s.liY.Y(), basics.IRound(ye*float64(s.subpixelScale)), length)
}

// Coordinates returns the current source-space coordinate in subpixel units.
func (s *SpanInterpolatorLinear[T]) Coordinates() (x, y int) {
	return s.liX.Y(), s.liY.Y()
}

// Next advances to the next pixel in the span.
func (s *SpanInterpolatorLinear[T]) Next() {
	s.liX.Inc()
	s.liY.Inc()
}

// SubpixelShift returns the fixed-point precision used by Coordinates.
func (s *SpanInterpolatorLinear[T]) SubpixelShift() int {
	return s.subpixelShift
}

// Transformer returns the current transformer.
func (s *SpanInterpolatorLinear[T]) Transformer() T {
	return s.transformer
}

// SetTransformer replaces the current transformer.
func (s *SpanInterpolatorLinear[T]) SetTransformer(transformer T) {
	s.transformer = transformer
}

// SpanInterpolatorLinearSubdiv is the Go equivalent of AGG's
// span_interpolator_linear_subdiv. It periodically resynchronizes long spans so
// transform error does not accumulate across the whole run.
type SpanInterpolatorLinearSubdiv[T TransformerInterface] struct {
	transformer   T
	subpixelShift int
	subpixelScale int
	subdivShift   int // Subdivision shift (power of 2)
	subdivSize    int // Subdivision size (1 << subdivShift)
	subdivMask    int // Subdivision mask (subdivSize - 1)
	liX           Dda2LineInterpolator
	liY           Dda2LineInterpolator
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
	s.liX.Init(x1, x2, subdivLength)
	s.liY.Init(y1, y2, subdivLength)
}

// Resynchronize adjusts the interpolation to end at the specified coordinates.
// This creates a new subdivision from current position to the end point.
func (s *SpanInterpolatorLinearSubdiv[T]) Resynchronize(xe, ye float64, length int) {
	// Transform end point
	s.transformer.Transform(&xe, &ye)

	// Create new interpolators from current position to end point
	s.liX.Init(s.liX.Y(), basics.IRound(xe*float64(s.subpixelScale)), length)
	s.liY.Init(s.liY.Y(), basics.IRound(ye*float64(s.subpixelScale)), length)
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

		// AGG parity: transform only (src_x/subpixel + len, src_y) once.
		tx := float64(s.srcX)/float64(s.subpixelScale) + float64(subdivLength)
		ty := s.srcY
		s.transformer.Transform(&tx, &ty)
		s.liX.Init(s.liX.Y(), basics.IRound(tx*float64(s.subpixelScale)), subdivLength)
		s.liY.Init(s.liY.Y(), basics.IRound(ty*float64(s.subpixelScale)), subdivLength)

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
