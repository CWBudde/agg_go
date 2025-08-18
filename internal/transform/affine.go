// Package transform provides affine transformation functionality for AGG.
// This implements a port of AGG's trans_affine class.
package transform

import (
	"math"

	"agg_go/internal/basics"
)

// AffineEpsilon is the default epsilon for affine transformation comparisons
const AffineEpsilon = 1e-14

// TransAffine represents a 2x3 affine transformation matrix.
// The matrix components are:
//
//	sx  shx tx
//	shy sy  ty
//	0   0   1
//
// Where:
//
//	sx, sy   = scaling factors
//	shx, shy = shearing factors
//	tx, ty   = translation values
type TransAffine struct {
	SX, SHY, SHX, SY, TX, TY float64
}

// NewTransAffine creates a new identity transformation matrix.
func NewTransAffine() *TransAffine {
	return &TransAffine{
		SX: 1.0, SHY: 0.0, SHX: 0.0,
		SY: 1.0, TX: 0.0, TY: 0.0,
	}
}

// NewTransAffineFromValues creates a transformation matrix with specific values.
func NewTransAffineFromValues(sx, shy, shx, sy, tx, ty float64) *TransAffine {
	return &TransAffine{
		SX: sx, SHY: shy, SHX: shx,
		SY: sy, TX: tx, TY: ty,
	}
}

// NewTransAffineFromArray creates a transformation matrix from an array of 6 values.
func NewTransAffineFromArray(m [6]float64) *TransAffine {
	return &TransAffine{
		SX: m[0], SHY: m[1], SHX: m[2],
		SY: m[3], TX: m[4], TY: m[5],
	}
}

// NewTransAffineRectToParl creates a transformation from rectangle to parallelogram.
func NewTransAffineRectToParl(x1, y1, x2, y2 float64, parl [6]float64) *TransAffine {
	t := NewTransAffine()
	t.RectToParl(x1, y1, x2, y2, parl)
	return t
}

// NewTransAffineParlToRect creates a transformation from parallelogram to rectangle.
func NewTransAffineParlToRect(parl [6]float64, x1, y1, x2, y2 float64) *TransAffine {
	t := NewTransAffine()
	t.ParlToRect(parl, x1, y1, x2, y2)
	return t
}

// NewTransAffineParlToParl creates a transformation from one parallelogram to another.
func NewTransAffineParlToParl(src, dst [6]float64) *TransAffine {
	t := NewTransAffine()
	t.ParlToParl(src, dst)
	return t
}

// Reset resets the matrix to identity.
func (t *TransAffine) Reset() *TransAffine {
	t.SX = 1.0
	t.SY = 1.0
	t.SHY = 0.0
	t.SHX = 0.0
	t.TX = 0.0
	t.TY = 0.0
	return t
}

// Translate adds translation to the current matrix.
func (t *TransAffine) Translate(x, y float64) *TransAffine {
	t.TX += x
	t.TY += y
	return t
}

// Rotate applies rotation to the current matrix.
func (t *TransAffine) Rotate(angle float64) *TransAffine {
	ca := math.Cos(angle)
	sa := math.Sin(angle)

	t0 := t.SX*ca - t.SHY*sa
	t2 := t.SHX*ca - t.SY*sa
	t4 := t.TX*ca - t.TY*sa

	t.SHY = t.SX*sa + t.SHY*ca
	t.SY = t.SHX*sa + t.SY*ca
	t.TY = t.TX*sa + t.TY*ca

	t.SX = t0
	t.SHX = t2
	t.TX = t4

	return t
}

// Scale applies uniform scaling to the current matrix.
func (t *TransAffine) Scale(s float64) *TransAffine {
	t.SX *= s
	t.SHX *= s
	t.TX *= s
	t.SHY *= s
	t.SY *= s
	t.TY *= s
	return t
}

// ScaleXY applies non-uniform scaling to the current matrix.
func (t *TransAffine) ScaleXY(sx, sy float64) *TransAffine {
	t.SX *= sx
	t.SHX *= sx
	t.TX *= sx
	t.SHY *= sy
	t.SY *= sy
	t.TY *= sy
	return t
}

// Multiply multiplies this matrix by another matrix.
func (t *TransAffine) Multiply(m *TransAffine) *TransAffine {
	t0 := t.SX*m.SX + t.SHY*m.SHX
	t2 := t.SHX*m.SX + t.SY*m.SHX
	t4 := t.TX*m.SX + t.TY*m.SHX + m.TX

	t.SHY = t.SX*m.SHY + t.SHY*m.SY
	t.SY = t.SHX*m.SHY + t.SY*m.SY
	t.TY = t.TX*m.SHY + t.TY*m.SY + m.TY

	t.SX = t0
	t.SHX = t2
	t.TX = t4

	return t
}

// Premultiply multiplies another matrix by this matrix.
func (t *TransAffine) Premultiply(m *TransAffine) *TransAffine {
	temp := *m
	temp.Multiply(t)
	*t = temp
	return t
}

// MultiplyInv multiplies this matrix by the inverse of another matrix.
func (t *TransAffine) MultiplyInv(m *TransAffine) *TransAffine {
	temp := *m
	temp.Invert()
	return t.Multiply(&temp)
}

// PremultiplyInv multiplies the inverse of another matrix by this matrix.
func (t *TransAffine) PremultiplyInv(m *TransAffine) *TransAffine {
	temp := *m
	temp.Invert()
	temp.Multiply(t)
	*t = temp
	return t
}

// Invert inverts the matrix. Returns false if matrix is singular.
func (t *TransAffine) Invert() *TransAffine {
	d := t.DeterminantReciprocal()

	t0 := t.SY * d
	t.SY = t.SX * d
	t.SHY = -t.SHY * d
	t.SHX = -t.SHX * d

	t4 := -t.TX*t0 - t.TY*t.SHX
	t.TY = -t.TX*t.SHY - t.TY*t.SY

	t.SX = t0
	t.TX = t4

	return t
}

// FlipX applies mirroring around the X axis.
func (t *TransAffine) FlipX() *TransAffine {
	t.SX = -t.SX
	t.SHY = -t.SHY
	t.TX = -t.TX
	return t
}

// FlipY applies mirroring around the Y axis.
func (t *TransAffine) FlipY() *TransAffine {
	t.SHX = -t.SHX
	t.SY = -t.SY
	t.TY = -t.TY
	return t
}

// StoreTo stores the matrix to an array of 6 values.
func (t *TransAffine) StoreTo(m []float64) {
	m[0] = t.SX
	m[1] = t.SHY
	m[2] = t.SHX
	m[3] = t.SY
	m[4] = t.TX
	m[5] = t.TY
}

// LoadFrom loads the matrix from an array of 6 values.
func (t *TransAffine) LoadFrom(m []float64) *TransAffine {
	t.SX = m[0]
	t.SHY = m[1]
	t.SHX = m[2]
	t.SY = m[3]
	t.TX = m[4]
	t.TY = m[5]
	return t
}

// Transform applies the transformation to a point.
func (t *TransAffine) Transform(x, y *float64) {
	tmp := *x
	*x = tmp*t.SX + *y*t.SHX + t.TX
	*y = tmp*t.SHY + *y*t.SY + t.TY
}

// Transform2x2 applies only the 2x2 part of the transformation (no translation).
func (t *TransAffine) Transform2x2(x, y *float64) {
	tmp := *x
	*x = tmp*t.SX + *y*t.SHX
	*y = tmp*t.SHY + *y*t.SY
}

// InverseTransform applies the inverse transformation to a point.
func (t *TransAffine) InverseTransform(x, y *float64) {
	d := t.DeterminantReciprocal()
	a := (*x - t.TX) * d
	b := (*y - t.TY) * d
	*x = a*t.SY - b*t.SHX
	*y = b*t.SX - a*t.SHY
}

// Determinant calculates the determinant of the matrix.
func (t *TransAffine) Determinant() float64 {
	return t.SX*t.SY - t.SHY*t.SHX
}

// DeterminantReciprocal calculates the reciprocal of the determinant.
func (t *TransAffine) DeterminantReciprocal() float64 {
	return 1.0 / (t.SX*t.SY - t.SHY*t.SHX)
}

// Scale returns the average scale factor.
func (t *TransAffine) GetScale() float64 {
	x := 0.707106781*t.SX + 0.707106781*t.SHX
	y := 0.707106781*t.SHY + 0.707106781*t.SY
	return math.Sqrt(x*x + y*y)
}

// IsValid checks if the matrix is not degenerate.
func (t *TransAffine) IsValid(epsilon float64) bool {
	return math.Abs(t.SX) > epsilon && math.Abs(t.SY) > epsilon
}

// IsIdentity checks if the matrix is an identity matrix.
func (t *TransAffine) IsIdentity(epsilon float64) bool {
	return basics.IsEqualEps(t.SX, 1.0, epsilon) &&
		basics.IsEqualEps(t.SHY, 0.0, epsilon) &&
		basics.IsEqualEps(t.SHX, 0.0, epsilon) &&
		basics.IsEqualEps(t.SY, 1.0, epsilon) &&
		basics.IsEqualEps(t.TX, 0.0, epsilon) &&
		basics.IsEqualEps(t.TY, 0.0, epsilon)
}

// IsEqual checks if two matrices are equal within epsilon.
func (t *TransAffine) IsEqual(m *TransAffine, epsilon float64) bool {
	return basics.IsEqualEps(t.SX, m.SX, epsilon) &&
		basics.IsEqualEps(t.SHY, m.SHY, epsilon) &&
		basics.IsEqualEps(t.SHX, m.SHX, epsilon) &&
		basics.IsEqualEps(t.SY, m.SY, epsilon) &&
		basics.IsEqualEps(t.TX, m.TX, epsilon) &&
		basics.IsEqualEps(t.TY, m.TY, epsilon)
}

// GetRotation extracts the rotation angle from the matrix.
func (t *TransAffine) GetRotation() float64 {
	x1, y1 := 0.0, 0.0
	x2, y2 := 1.0, 0.0
	t.Transform(&x1, &y1)
	t.Transform(&x2, &y2)
	return math.Atan2(y2-y1, x2-x1)
}

// GetTranslation extracts the translation values from the matrix.
func (t *TransAffine) GetTranslation() (float64, float64) {
	return t.TX, t.TY
}

// GetScaling extracts the scaling factors from the matrix.
func (t *TransAffine) GetScaling() (float64, float64) {
	x1, y1 := 0.0, 0.0
	x2, y2 := 1.0, 1.0
	temp := *t
	temp.Multiply(NewTransAffineRotation(-t.GetRotation()))
	temp.Transform(&x1, &y1)
	temp.Transform(&x2, &y2)
	return x2 - x1, y2 - y1
}

// GetScalingAbs extracts absolute scaling factors (used for image resampling).
func (t *TransAffine) GetScalingAbs() (float64, float64) {
	x := math.Sqrt(t.SX*t.SX + t.SHX*t.SHX)
	y := math.Sqrt(t.SHY*t.SHY + t.SY*t.SY)
	return x, y
}

// ParlToParl transforms from one parallelogram to another.
func (t *TransAffine) ParlToParl(src, dst [6]float64) *TransAffine {
	t.SX = src[2] - src[0]
	t.SHY = src[3] - src[1]
	t.SHX = src[4] - src[0]
	t.SY = src[5] - src[1]
	t.TX = src[0]
	t.TY = src[1]
	t.Invert()
	t.Multiply(NewTransAffineFromValues(
		dst[2]-dst[0], dst[3]-dst[1],
		dst[4]-dst[0], dst[5]-dst[1],
		dst[0], dst[1]))
	return t
}

// RectToParl transforms from rectangle to parallelogram.
func (t *TransAffine) RectToParl(x1, y1, x2, y2 float64, parl [6]float64) *TransAffine {
	src := [6]float64{x1, y1, x2, y1, x2, y2}
	return t.ParlToParl(src, parl)
}

// ParlToRect transforms from parallelogram to rectangle.
func (t *TransAffine) ParlToRect(parl [6]float64, x1, y1, x2, y2 float64) *TransAffine {
	dst := [6]float64{x1, y1, x2, y1, x2, y2}
	return t.ParlToParl(parl, dst)
}

// Copy creates a copy of the transformation matrix.
func (t *TransAffine) Copy() *TransAffine {
	return &TransAffine{
		SX: t.SX, SHY: t.SHY, SHX: t.SHX,
		SY: t.SY, TX: t.TX, TY: t.TY,
	}
}

// Equal checks equality with default epsilon.
func (t *TransAffine) Equal(m *TransAffine) bool {
	return t.IsEqual(m, AffineEpsilon)
}

// NotEqual checks inequality with default epsilon.
func (t *TransAffine) NotEqual(m *TransAffine) bool {
	return !t.IsEqual(m, AffineEpsilon)
}

// MultiplyBy creates a new matrix that is the result of multiplying this matrix by another.
func (t *TransAffine) MultiplyBy(m *TransAffine) *TransAffine {
	result := t.Copy()
	return result.Multiply(m)
}

// DivideBy creates a new matrix that is the result of multiplying this matrix by the inverse of another.
func (t *TransAffine) DivideBy(m *TransAffine) *TransAffine {
	result := t.Copy()
	return result.MultiplyInv(m)
}

// Inverse creates a new matrix that is the inverse of this matrix.
func (t *TransAffine) Inverse() *TransAffine {
	result := t.Copy()
	return result.Invert()
}
