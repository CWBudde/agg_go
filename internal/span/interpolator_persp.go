// Package span provides perspective span interpolation functionality for AGG.
// This implements a port of AGG's span_interpolator_persp_exact and span_interpolator_persp_lerp classes.
package span

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// SpanInterpolatorPerspectiveExact implements exact perspective span interpolation.
// This is a port of AGG's span_interpolator_persp_exact template class.
// It provides precise perspective transformations using the full perspective iterator
// and calculates local scale factors for accurate texture mapping.
type SpanInterpolatorPerspectiveExact struct {
	transformDir  *transform.TransPerspective
	transformInv  *transform.TransPerspective
	iterator      *transform.PerspectiveIteratorX
	scaleX        *Dda2LineInterpolator
	scaleY        *Dda2LineInterpolator
	subpixelShift int
	subpixelScale int
}

// NewSpanInterpolatorPerspectiveExact creates a new exact perspective span interpolator.
func NewSpanInterpolatorPerspectiveExact(subpixelShift int) *SpanInterpolatorPerspectiveExact {
	if subpixelShift == 0 {
		subpixelShift = 8 // Default from AGG
	}

	return &SpanInterpolatorPerspectiveExact{
		transformDir:  transform.NewTransPerspective(),
		transformInv:  transform.NewTransPerspective(),
		subpixelShift: subpixelShift,
		subpixelScale: 1 << subpixelShift,
	}
}

// NewSpanInterpolatorPerspectiveExactQuadToQuad creates an interpolator with quad-to-quad transformation.
func NewSpanInterpolatorPerspectiveExactQuadToQuad(src, dst [8]float64, subpixelShift int) *SpanInterpolatorPerspectiveExact {
	s := NewSpanInterpolatorPerspectiveExact(subpixelShift)
	s.QuadToQuad(src, dst)
	return s
}

// NewSpanInterpolatorPerspectiveExactRectToQuad creates an interpolator with rect-to-quad transformation.
func NewSpanInterpolatorPerspectiveExactRectToQuad(x1, y1, x2, y2 float64, quad [8]float64, subpixelShift int) *SpanInterpolatorPerspectiveExact {
	s := NewSpanInterpolatorPerspectiveExact(subpixelShift)
	s.RectToQuad(x1, y1, x2, y2, quad)
	return s
}

// NewSpanInterpolatorPerspectiveExactQuadToRect creates an interpolator with quad-to-rect transformation.
func NewSpanInterpolatorPerspectiveExactQuadToRect(quad [8]float64, x1, y1, x2, y2 float64, subpixelShift int) *SpanInterpolatorPerspectiveExact {
	s := NewSpanInterpolatorPerspectiveExact(subpixelShift)
	s.QuadToRect(quad, x1, y1, x2, y2)
	return s
}

// QuadToQuad sets the transformation using two arbitrary quadrangles.
func (s *SpanInterpolatorPerspectiveExact) QuadToQuad(src, dst [8]float64) {
	s.transformDir.QuadToQuad(src, dst)
	s.transformInv.QuadToQuad(dst, src)
}

// RectToQuad sets the direct transformation (rectangle → quadrangle).
func (s *SpanInterpolatorPerspectiveExact) RectToQuad(x1, y1, x2, y2 float64, quad [8]float64) {
	src := [8]float64{x1, y1, x2, y1, x2, y2, x1, y2}
	s.QuadToQuad(src, quad)
}

// QuadToRect sets the reverse transformation (quadrangle → rectangle).
func (s *SpanInterpolatorPerspectiveExact) QuadToRect(quad [8]float64, x1, y1, x2, y2 float64) {
	dst := [8]float64{x1, y1, x2, y1, x2, y2, x1, y2}
	s.QuadToQuad(quad, dst)
}

// IsValid checks if the equations were solved successfully.
func (s *SpanInterpolatorPerspectiveExact) IsValid() bool {
	return s.transformDir.IsValid(1e-10)
}

// Begin starts interpolation for a span of pixels.
func (s *SpanInterpolatorPerspectiveExact) Begin(x, y float64, length int) {
	s.iterator = s.transformDir.Begin(x, y, 1.0)
	xt := s.iterator.X
	yt := s.iterator.Y

	delta := 1.0 / float64(s.subpixelScale)

	// Calculate scale by X at start point
	dx := xt + delta
	dy := yt
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sx1 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate scale by Y at start point
	dx = xt
	dy = yt + delta
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sy1 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate transformed coordinates at end point
	x += float64(length)
	xt = x
	yt = y
	s.transformDir.Transform(&xt, &yt)

	// Calculate scale by X at end point
	dx = xt + delta
	dy = yt
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sx2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate scale by Y at end point
	dx = xt
	dy = yt + delta
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sy2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Initialize scale interpolators
	s.scaleX = NewDda2LineInterpolator(sx1, sx2, length)
	s.scaleY = NewDda2LineInterpolator(sy1, sy2, length)
}

// Resynchronize adjusts the interpolation to end at specified coordinates.
func (s *SpanInterpolatorPerspectiveExact) Resynchronize(xe, ye float64, length int) {
	// Assume x1,y1 are equal to the ones at the previous end point
	sx1 := s.scaleX.Y()
	sy1 := s.scaleY.Y()

	// Calculate transformed coordinates at x2,y2
	xt := xe
	yt := ye
	s.transformDir.Transform(&xt, &yt)

	delta := 1.0 / float64(s.subpixelScale)

	// Calculate scale by X at x2,y2
	dx := xt + delta
	dy := yt
	s.transformInv.Transform(&dx, &dy)
	dx -= xe
	dy -= ye
	sx2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate scale by Y at x2,y2
	dx = xt
	dy = yt + delta
	s.transformInv.Transform(&dx, &dy)
	dx -= xe
	dy -= ye
	sy2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Initialize the interpolators
	s.scaleX = NewDda2LineInterpolator(sx1, sx2, length)
	s.scaleY = NewDda2LineInterpolator(sy1, sy2, length)
}

// Next advances the interpolator to the next position.
func (s *SpanInterpolatorPerspectiveExact) Next() {
	s.iterator.Next()
	s.scaleX.Inc()
	s.scaleY.Inc()
}

// Coordinates returns the current transformed coordinates.
func (s *SpanInterpolatorPerspectiveExact) Coordinates() (x, y int) {
	return basics.IRound(s.iterator.X * float64(s.subpixelScale)),
		basics.IRound(s.iterator.Y * float64(s.subpixelScale))
}

// LocalScale returns the local scale factors for texture mapping.
func (s *SpanInterpolatorPerspectiveExact) LocalScale() (x, y int) {
	return s.scaleX.Y(), s.scaleY.Y()
}

// Transform applies the transformation to a point.
func (s *SpanInterpolatorPerspectiveExact) Transform(x, y *float64) {
	s.transformDir.Transform(x, y)
}

// SubpixelShift returns the subpixel precision shift value.
func (s *SpanInterpolatorPerspectiveExact) SubpixelShift() int {
	return s.subpixelShift
}

// SpanInterpolatorPerspectiveLerp implements perspective span interpolation with linear approximation.
// This is a port of AGG's span_interpolator_persp_lerp template class.
// It provides faster perspective transformations by interpolating transformed coordinates directly
// rather than using the exact perspective iterator.
type SpanInterpolatorPerspectiveLerp struct {
	transformDir  *transform.TransPerspective
	transformInv  *transform.TransPerspective
	coordX        *Dda2LineInterpolator
	coordY        *Dda2LineInterpolator
	scaleX        *Dda2LineInterpolator
	scaleY        *Dda2LineInterpolator
	subpixelShift int
	subpixelScale int
}

// NewSpanInterpolatorPerspectiveLerp creates a new linear approximation perspective span interpolator.
func NewSpanInterpolatorPerspectiveLerp(subpixelShift int) *SpanInterpolatorPerspectiveLerp {
	if subpixelShift == 0 {
		subpixelShift = 8 // Default from AGG
	}

	return &SpanInterpolatorPerspectiveLerp{
		transformDir:  transform.NewTransPerspective(),
		transformInv:  transform.NewTransPerspective(),
		subpixelShift: subpixelShift,
		subpixelScale: 1 << subpixelShift,
	}
}

// NewSpanInterpolatorPerspectiveLerpQuadToQuad creates an interpolator with quad-to-quad transformation.
func NewSpanInterpolatorPerspectiveLerpQuadToQuad(src, dst [8]float64, subpixelShift int) *SpanInterpolatorPerspectiveLerp {
	s := NewSpanInterpolatorPerspectiveLerp(subpixelShift)
	s.QuadToQuad(src, dst)
	return s
}

// NewSpanInterpolatorPerspectiveLerpRectToQuad creates an interpolator with rect-to-quad transformation.
func NewSpanInterpolatorPerspectiveLerpRectToQuad(x1, y1, x2, y2 float64, quad [8]float64, subpixelShift int) *SpanInterpolatorPerspectiveLerp {
	s := NewSpanInterpolatorPerspectiveLerp(subpixelShift)
	s.RectToQuad(x1, y1, x2, y2, quad)
	return s
}

// NewSpanInterpolatorPerspectiveLerpQuadToRect creates an interpolator with quad-to-rect transformation.
func NewSpanInterpolatorPerspectiveLerpQuadToRect(quad [8]float64, x1, y1, x2, y2 float64, subpixelShift int) *SpanInterpolatorPerspectiveLerp {
	s := NewSpanInterpolatorPerspectiveLerp(subpixelShift)
	s.QuadToRect(quad, x1, y1, x2, y2)
	return s
}

// QuadToQuad sets the transformation using two arbitrary quadrangles.
func (s *SpanInterpolatorPerspectiveLerp) QuadToQuad(src, dst [8]float64) {
	s.transformDir.QuadToQuad(src, dst)
	s.transformInv.QuadToQuad(dst, src)
}

// RectToQuad sets the direct transformation (rectangle → quadrangle).
func (s *SpanInterpolatorPerspectiveLerp) RectToQuad(x1, y1, x2, y2 float64, quad [8]float64) {
	src := [8]float64{x1, y1, x2, y1, x2, y2, x1, y2}
	s.QuadToQuad(src, quad)
}

// QuadToRect sets the reverse transformation (quadrangle → rectangle).
func (s *SpanInterpolatorPerspectiveLerp) QuadToRect(quad [8]float64, x1, y1, x2, y2 float64) {
	dst := [8]float64{x1, y1, x2, y1, x2, y2, x1, y2}
	s.QuadToQuad(quad, dst)
}

// IsValid checks if the equations were solved successfully.
func (s *SpanInterpolatorPerspectiveLerp) IsValid() bool {
	return s.transformDir.IsValid(1e-10)
}

// Begin starts interpolation for a span of pixels.
func (s *SpanInterpolatorPerspectiveLerp) Begin(x, y float64, length int) {
	// Calculate transformed coordinates at x1,y1
	xt := x
	yt := y
	s.transformDir.Transform(&xt, &yt)
	x1 := basics.IRound(xt * float64(s.subpixelScale))
	y1 := basics.IRound(yt * float64(s.subpixelScale))

	delta := 1.0 / float64(s.subpixelScale)

	// Calculate scale by X at x1,y1
	dx := xt + delta
	dy := yt
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sx1 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate scale by Y at x1,y1
	dx = xt
	dy = yt + delta
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sy1 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate transformed coordinates at x2,y2
	x += float64(length)
	xt = x
	yt = y
	s.transformDir.Transform(&xt, &yt)
	x2 := basics.IRound(xt * float64(s.subpixelScale))
	y2 := basics.IRound(yt * float64(s.subpixelScale))

	// Calculate scale by X at x2,y2
	dx = xt + delta
	dy = yt
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sx2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate scale by Y at x2,y2
	dx = xt
	dy = yt + delta
	s.transformInv.Transform(&dx, &dy)
	dx -= x
	dy -= y
	sy2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Initialize the interpolators
	s.coordX = NewDda2LineInterpolator(x1, x2, length)
	s.coordY = NewDda2LineInterpolator(y1, y2, length)
	s.scaleX = NewDda2LineInterpolator(sx1, sx2, length)
	s.scaleY = NewDda2LineInterpolator(sy1, sy2, length)
}

// Resynchronize adjusts the interpolation to end at specified coordinates.
func (s *SpanInterpolatorPerspectiveLerp) Resynchronize(xe, ye float64, length int) {
	// Assume x1,y1 are equal to the ones at the previous end point
	x1 := s.coordX.Y()
	y1 := s.coordY.Y()
	sx1 := s.scaleX.Y()
	sy1 := s.scaleY.Y()

	// Calculate transformed coordinates at x2,y2
	xt := xe
	yt := ye
	s.transformDir.Transform(&xt, &yt)
	x2 := basics.IRound(xt * float64(s.subpixelScale))
	y2 := basics.IRound(yt * float64(s.subpixelScale))

	delta := 1.0 / float64(s.subpixelScale)

	// Calculate scale by X at x2,y2
	dx := xt + delta
	dy := yt
	s.transformInv.Transform(&dx, &dy)
	dx -= xe
	dy -= ye
	sx2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Calculate scale by Y at x2,y2
	dx = xt
	dy = yt + delta
	s.transformInv.Transform(&dx, &dy)
	dx -= xe
	dy -= ye
	sy2 := int(basics.URound(float64(s.subpixelScale)/math.Sqrt(dx*dx+dy*dy))) >> s.subpixelShift

	// Initialize the interpolators
	s.coordX = NewDda2LineInterpolator(x1, x2, length)
	s.coordY = NewDda2LineInterpolator(y1, y2, length)
	s.scaleX = NewDda2LineInterpolator(sx1, sx2, length)
	s.scaleY = NewDda2LineInterpolator(sy1, sy2, length)
}

// Next advances the interpolator to the next position.
func (s *SpanInterpolatorPerspectiveLerp) Next() {
	s.coordX.Inc()
	s.coordY.Inc()
	s.scaleX.Inc()
	s.scaleY.Inc()
}

// Coordinates returns the current transformed coordinates.
func (s *SpanInterpolatorPerspectiveLerp) Coordinates() (x, y int) {
	return s.coordX.Y(), s.coordY.Y()
}

// LocalScale returns the local scale factors for texture mapping.
func (s *SpanInterpolatorPerspectiveLerp) LocalScale() (x, y int) {
	return s.scaleX.Y(), s.scaleY.Y()
}

// Transform applies the transformation to a point.
func (s *SpanInterpolatorPerspectiveLerp) Transform(x, y *float64) {
	s.transformDir.Transform(x, y)
}

// SubpixelShift returns the subpixel precision shift value.
func (s *SpanInterpolatorPerspectiveLerp) SubpixelShift() int {
	return s.subpixelShift
}
