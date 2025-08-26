// Package agg provides transformation functionality for 2D graphics.
// This file contains transformation matrices and operations, wired to internal/agg2d.
package agg

import (
	"math"

	ia "agg_go/internal/agg2d"
	"agg_go/internal/transform"
)

// Re-export internal types and options for public API.
type (
	// Local public Transformations type; we convert to/from internal as needed.
	Transformations struct{ AffineMatrix [6]float64 }
	ViewportOption  = ia.ViewportOption
)

// Viewport alignment/scaling options (aliases to internal constants).
const (
	Anisotropic ViewportOption = ia.Anisotropic
	XMinYMin    ViewportOption = ia.XMinYMin
	XMinYMid    ViewportOption = ia.XMinYMid
	XMinYMax    ViewportOption = ia.XMinYMax
	XMidYMin    ViewportOption = ia.XMidYMin
	XMidYMid    ViewportOption = ia.XMidYMid
	XMidYMax    ViewportOption = ia.XMidYMax
	XMaxYMin    ViewportOption = ia.XMaxYMin
	XMaxYMid    ViewportOption = ia.XMaxYMid
	XMaxYMax    ViewportOption = ia.XMaxYMax
)

// NewTransformationsFromValues creates a transformation from matrix values.
func NewTransformationsFromValues(sx, shy, shx, sy, tx, ty float64) *Transformations {
	return &Transformations{AffineMatrix: [6]float64{sx, shy, shx, sy, tx, ty}}
}

// Context transformation methods

// GetTransform returns the current transformation matrix from the context.
func (ctx *Context) GetTransform() *Transformations {
	return fromInternalTransformations(ctx.agg2d.impl.GetTransformations())
}

// SetTransform sets the transformation matrix on the context.
func (ctx *Context) SetTransform(tr *Transformations) {
	ctx.agg2d.impl.SetTransformations(toInternalTransformations(tr))
}

// ResetTransform resets the transformation matrix to identity.
func (ctx *Context) ResetTransform() {
	ctx.agg2d.ResetTransformations()
}

// Transform multiplies the current transformation matrix with another.
func (ctx *Context) Transform(tr *Transformations) {
	ctx.agg2d.impl.AffineFromMatrix(toInternalTransformations(tr))
}

// Translate applies a translation transformation.
func (ctx *Context) Translate(tx, ty float64) {
	ctx.agg2d.impl.Translate(tx, ty)
}

// Rotate applies a rotation transformation (angle in radians).
func (ctx *Context) Rotate(angle float64) {
	ctx.agg2d.impl.Rotate(angle)
}

// RotateDegrees applies a rotation transformation (angle in degrees).
func (ctx *Context) RotateDegrees(degrees float64) {
	ctx.agg2d.impl.Rotate(degrees * math.Pi / 180.0)
}

// Scale applies a scaling transformation.
func (ctx *Context) Scale(sx, sy float64) {
	ctx.agg2d.impl.Scale(sx, sy)
}

// ScaleUniform applies uniform scaling (same factor for both axes).
func (ctx *Context) ScaleUniform(s float64) {
	ctx.agg2d.impl.UniformScale(s)
}

// Skew applies a skewing transformation (angles in radians).
func (ctx *Context) Skew(sx, sy float64) {
	ctx.agg2d.impl.Skew(sx, sy)
}

// SkewDegrees applies a skewing transformation (angles in degrees).
func (ctx *Context) SkewDegrees(sx, sy float64) {
	ctx.agg2d.impl.Skew(sx*math.Pi/180.0, sy*math.Pi/180.0)
}

// Advanced transformation operations

// PushTransform saves the current transformation on the stack.
func (ctx *Context) PushTransform() {
	ctx.agg2d.impl.PushTransform()
}

// PopTransform restores the last saved transformation from the stack.
func (ctx *Context) PopTransform() bool {
	return ctx.agg2d.impl.PopTransform()
}

// Coordinate conversions

// WorldToScreen converts world coordinates to screen coordinates.
func (ctx *Context) WorldToScreen(wx, wy float64) (sx, sy float64) {
	sx, sy = wx, wy
	ctx.agg2d.impl.WorldToScreen(&sx, &sy)
	return sx, sy
}

// ScreenToWorld converts screen coordinates to world coordinates.
func (ctx *Context) ScreenToWorld(sx, sy float64) (wx, wy float64) {
	wx, wy = sx, sy
	ctx.agg2d.impl.ScreenToWorld(&wx, &wy)
	return wx, wy
}

// WorldToScreenDistance converts a world distance to screen distance.
func (ctx *Context) WorldToScreenDistance(wd float64) float64 {
	return ctx.agg2d.impl.WorldToScreenDistance(wd)
}

// ScreenToWorldDistance converts a screen distance to world distance.
func (ctx *Context) ScreenToWorldDistance(sd float64) (float64, bool) {
	return ctx.agg2d.impl.ScreenToWorldDistance(sd)
}

// Viewport operations

// Viewport sets up a viewport transformation.
func (ctx *Context) Viewport(worldX1, worldY1, worldX2, worldY2 float64,
	screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption,
) {
	ctx.agg2d.impl.Viewport(worldX1, worldY1, worldX2, worldY2,
		screenX1, screenY1, screenX2, screenY2, opt)
}

// ViewportFitPage sets up viewport to fit the entire page.
func (ctx *Context) ViewportFitPage(worldX1, worldY1, worldX2, worldY2 float64) {
	ctx.agg2d.impl.Viewport(worldX1, worldY1, worldX2, worldY2,
		0, 0, float64(ctx.width), float64(ctx.height), XMidYMid)
}

// ViewportCenter centers the world coordinates in the screen.
func (ctx *Context) ViewportCenter(worldX1, worldY1, worldX2, worldY2 float64) {
	ctx.agg2d.impl.Viewport(worldX1, worldY1, worldX2, worldY2,
		0, 0, float64(ctx.width), float64(ctx.height), XMidYMid)
}

// Matrix operations on Transformations

// Multiply multiplies this transformation by another.
func (t *Transformations) Multiply(other *Transformations) {
	// Create TransAffine objects for multiplication
	t1 := transform.NewTransAffineFromValues(
		t.AffineMatrix[0], t.AffineMatrix[1], t.AffineMatrix[2],
		t.AffineMatrix[3], t.AffineMatrix[4], t.AffineMatrix[5])
	t2 := transform.NewTransAffineFromValues(
		other.AffineMatrix[0], other.AffineMatrix[1], other.AffineMatrix[2],
		other.AffineMatrix[3], other.AffineMatrix[4], other.AffineMatrix[5])

	t1.Multiply(t2)

	t.AffineMatrix[0] = t1.SX
	t.AffineMatrix[1] = t1.SHY
	t.AffineMatrix[2] = t1.SHX
	t.AffineMatrix[3] = t1.SY
	t.AffineMatrix[4] = t1.TX
	t.AffineMatrix[5] = t1.TY
}

// Invert inverts the transformation matrix.
func (t *Transformations) Invert() bool {
	affine := transform.NewTransAffineFromValues(
		t.AffineMatrix[0], t.AffineMatrix[1], t.AffineMatrix[2],
		t.AffineMatrix[3], t.AffineMatrix[4], t.AffineMatrix[5])

	// Check if matrix is invertible (determinant != 0)
	det := affine.SX*affine.SY - affine.SHX*affine.SHY
	if transformAbs(det) < 1e-10 {
		return false
	}

	affine.Invert()

	t.AffineMatrix[0] = affine.SX
	t.AffineMatrix[1] = affine.SHY
	t.AffineMatrix[2] = affine.SHX
	t.AffineMatrix[3] = affine.SY
	t.AffineMatrix[4] = affine.TX
	t.AffineMatrix[5] = affine.TY

	return true
}

// Transform applies the transformation to a point.
func (t *Transformations) Transform(x, y float64) (float64, float64) {
	affine := transform.NewTransAffineFromValues(
		t.AffineMatrix[0], t.AffineMatrix[1], t.AffineMatrix[2],
		t.AffineMatrix[3], t.AffineMatrix[4], t.AffineMatrix[5])

	px, py := x, y
	affine.Transform(&px, &py)
	return px, py
}

// IsIdentity checks if this is an identity transformation.
func (t *Transformations) IsIdentity() bool {
	const epsilon = 1e-10
	return transformAbs(t.AffineMatrix[0]-1.0) < epsilon &&
		transformAbs(t.AffineMatrix[1]) < epsilon &&
		transformAbs(t.AffineMatrix[2]) < epsilon &&
		transformAbs(t.AffineMatrix[3]-1.0) < epsilon &&
		transformAbs(t.AffineMatrix[4]) < epsilon &&
		transformAbs(t.AffineMatrix[5]) < epsilon
}

// Reset resets to identity transformation.
func (t *Transformations) Reset() {
	t.AffineMatrix = [6]float64{1.0, 0.0, 0.0, 1.0, 0.0, 0.0}
}

// Clone creates a copy of this transformation.
func (t *Transformations) Clone() *Transformations {
	return &Transformations{AffineMatrix: t.AffineMatrix}
}

// Static transformation constructors

// Translation creates a translation transformation.
func Translation(tx, ty float64) *Transformations {
	return &Transformations{AffineMatrix: [6]float64{1.0, 0.0, 0.0, 1.0, tx, ty}}
}

// Rotation creates a rotation transformation (angle in radians).
func Rotation(angle float64) *Transformations {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return &Transformations{AffineMatrix: [6]float64{cos, sin, -sin, cos, 0.0, 0.0}}
}

// RotationDegrees creates a rotation transformation (angle in degrees).
func RotationDegrees(degrees float64) *Transformations {
	return Rotation(degrees * math.Pi / 180.0)
}

// Scaling creates a scaling transformation.
func Scaling(sx, sy float64) *Transformations {
	return &Transformations{AffineMatrix: [6]float64{sx, 0.0, 0.0, sy, 0.0, 0.0}}
}

// UniformScaling creates a uniform scaling transformation.
func UniformScaling(s float64) *Transformations {
	return Scaling(s, s)
}

// Skewing creates a skewing transformation (angles in radians).
func Skewing(sx, sy float64) *Transformations {
	return &Transformations{AffineMatrix: [6]float64{1.0, math.Tan(sy), math.Tan(sx), 1.0, 0.0, 0.0}}
}

// Helper function for transformation calculations
func transformAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Internal conversion helpers
func toInternalTransformations(t *Transformations) *ia.Transformations {
	if t == nil {
		return nil
	}
	return &ia.Transformations{AffineMatrix: t.AffineMatrix}
}

func fromInternalTransformations(it *ia.Transformations) *Transformations {
	if it == nil {
		return nil
	}
	return &Transformations{AffineMatrix: it.AffineMatrix}
}
