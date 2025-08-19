// Package transform provides perspective transformation functionality for AGG.
// This implements a port of AGG's trans_perspective class.
package transform

import (
	"math"

	"agg_go/internal/basics"
)

// TransPerspective represents a 3x3 perspective transformation matrix.
// The matrix components are:
//
//	sx  shy w0
//	shx sy  w1
//	tx  ty  w2
//
// Where the matrix transforms homogeneous coordinates [x, y, 1] to [x', y', w'],
// and the final 2D coordinates are [x'/w', y'/w'].
type TransPerspective struct {
	SX, SHY, W0 float64 // First row
	SHX, SY, W1 float64 // Second row
	TX, TY, W2  float64 // Third row
}

// NewTransPerspective creates a new identity perspective transformation matrix.
func NewTransPerspective() *TransPerspective {
	return &TransPerspective{
		SX: 1.0, SHY: 0.0, W0: 0.0,
		SHX: 0.0, SY: 1.0, W1: 0.0,
		TX: 0.0, TY: 0.0, W2: 1.0,
	}
}

// NewTransPerspectiveFromValues creates a perspective transformation matrix with specific values.
func NewTransPerspectiveFromValues(sx, shy, w0, shx, sy, w1, tx, ty, w2 float64) *TransPerspective {
	return &TransPerspective{
		SX: sx, SHY: shy, W0: w0,
		SHX: shx, SY: sy, W1: w1,
		TX: tx, TY: ty, W2: w2,
	}
}

// NewTransPerspectiveFromArray creates a perspective transformation matrix from an array of 9 values.
func NewTransPerspectiveFromArray(m [9]float64) *TransPerspective {
	return &TransPerspective{
		SX: m[0], SHY: m[1], W0: m[2],
		SHX: m[3], SY: m[4], W1: m[5],
		TX: m[6], TY: m[7], W2: m[8],
	}
}

// NewTransPerspectiveFromAffine creates a perspective transformation from an affine transformation.
func NewTransPerspectiveFromAffine(a *TransAffine) *TransPerspective {
	return &TransPerspective{
		SX: a.SX, SHY: a.SHY, W0: 0.0,
		SHX: a.SHX, SY: a.SY, W1: 0.0,
		TX: a.TX, TY: a.TY, W2: 1.0,
	}
}

// NewTransPerspectiveRectToQuad creates a transformation from rectangle to quadrilateral.
func NewTransPerspectiveRectToQuad(x1, y1, x2, y2 float64, quad [8]float64) *TransPerspective {
	t := NewTransPerspective()
	t.RectToQuad(x1, y1, x2, y2, quad)
	return t
}

// NewTransPerspectiveQuadToRect creates a transformation from quadrilateral to rectangle.
func NewTransPerspectiveQuadToRect(quad [8]float64, x1, y1, x2, y2 float64) *TransPerspective {
	t := NewTransPerspective()
	t.QuadToRect(quad, x1, y1, x2, y2)
	return t
}

// NewTransPerspectiveQuadToQuad creates a transformation from one quadrilateral to another.
func NewTransPerspectiveQuadToQuad(src, dst [8]float64) *TransPerspective {
	t := NewTransPerspective()
	t.QuadToQuad(src, dst)
	return t
}

// Reset resets the matrix to identity.
func (t *TransPerspective) Reset() *TransPerspective {
	t.SX = 1.0
	t.SHY = 0.0
	t.W0 = 0.0
	t.SHX = 0.0
	t.SY = 1.0
	t.W1 = 0.0
	t.TX = 0.0
	t.TY = 0.0
	t.W2 = 1.0
	return t
}

// SquareToQuad maps the unit square (0,0,1,1) to the quadrilateral.
// The quadrilateral is represented as [x0,y0, x1,y1, x2,y2, x3,y3].
func (t *TransPerspective) SquareToQuad(q [8]float64) bool {
	dx := q[0] - q[2] + q[4] - q[6]
	dy := q[1] - q[3] + q[5] - q[7]

	if dx == 0.0 && dy == 0.0 {
		// Affine case (parallelogram)
		t.SX = q[2] - q[0]
		t.SHY = q[3] - q[1]
		t.W0 = 0.0
		t.SHX = q[4] - q[2]
		t.SY = q[5] - q[3]
		t.W1 = 0.0
		t.TX = q[0]
		t.TY = q[1]
		t.W2 = 1.0
	} else {
		dx1 := q[2] - q[4]
		dy1 := q[3] - q[5]
		dx2 := q[6] - q[4]
		dy2 := q[7] - q[5]
		den := dx1*dy2 - dx2*dy1

		if den == 0.0 {
			// Singular case
			t.SX = 0.0
			t.SHY = 0.0
			t.W0 = 0.0
			t.SHX = 0.0
			t.SY = 0.0
			t.W1 = 0.0
			t.TX = 0.0
			t.TY = 0.0
			t.W2 = 0.0
			return false
		}

		// General case
		u := (dx*dy2 - dy*dx2) / den
		v := (dy*dx1 - dx*dy1) / den
		t.SX = q[2] - q[0] + u*q[2]
		t.SHY = q[3] - q[1] + u*q[3]
		t.W0 = u
		t.SHX = q[6] - q[0] + v*q[6]
		t.SY = q[7] - q[1] + v*q[7]
		t.W1 = v
		t.TX = q[0]
		t.TY = q[1]
		t.W2 = 1.0
	}
	return true
}

// QuadToSquare maps the quadrilateral to the unit square (0,0,1,1).
// The quadrilateral is represented as [x0,y0, x1,y1, x2,y2, x3,y3].
func (t *TransPerspective) QuadToSquare(q [8]float64) bool {
	if !t.SquareToQuad(q) {
		return false
	}
	return t.Invert()
}

// QuadToQuad maps one quadrilateral to another.
// Both quadrilaterals are represented as [x0,y0, x1,y1, x2,y2, x3,y3].
func (t *TransPerspective) QuadToQuad(src, dst [8]float64) bool {
	p := NewTransPerspective()
	if !t.QuadToSquare(src) {
		return false
	}
	if !p.SquareToQuad(dst) {
		return false
	}
	t.Multiply(p)
	return true
}

// RectToQuad maps rectangle to quadrilateral.
// The quadrilateral is represented as [x0,y0, x1,y1, x2,y2, x3,y3].
func (t *TransPerspective) RectToQuad(x1, y1, x2, y2 float64, q [8]float64) bool {
	r := [8]float64{x1, y1, x2, y1, x2, y2, x1, y2}
	return t.QuadToQuad(r, q)
}

// QuadToRect maps quadrilateral to rectangle.
// The quadrilateral is represented as [x0,y0, x1,y1, x2,y2, x3,y3].
func (t *TransPerspective) QuadToRect(q [8]float64, x1, y1, x2, y2 float64) bool {
	r := [8]float64{x1, y1, x2, y1, x2, y2, x1, y2}
	return t.QuadToQuad(q, r)
}

// Invert inverts the matrix. Returns false in degenerate case.
func (t *TransPerspective) Invert() bool {
	d0 := t.SY*t.W2 - t.W1*t.TY
	d1 := t.W0*t.TY - t.SHY*t.W2
	d2 := t.SHY*t.W1 - t.W0*t.SY
	d := t.SX*d0 + t.SHX*d1 + t.TX*d2

	if d == 0.0 {
		t.SX = 0.0
		t.SHY = 0.0
		t.W0 = 0.0
		t.SHX = 0.0
		t.SY = 0.0
		t.W1 = 0.0
		t.TX = 0.0
		t.TY = 0.0
		t.W2 = 0.0
		return false
	}

	d = 1.0 / d
	a := *t

	t.SX = d * d0
	t.SHY = d * d1
	t.W0 = d * d2
	t.SHX = d * (a.W1*a.TX - a.SHX*a.W2)
	t.SY = d * (a.SX*a.W2 - a.W0*a.TX)
	t.W1 = d * (a.W0*a.SHX - a.SX*a.W1)
	t.TX = d * (a.SHX*a.TY - a.SY*a.TX)
	t.TY = d * (a.SHY*a.TX - a.SX*a.TY)
	t.W2 = d * (a.SX*a.SY - a.SHY*a.SHX)
	return true
}

// Multiply multiplies the matrix by another perspective matrix.
func (t *TransPerspective) Multiply(a *TransPerspective) *TransPerspective {
	b := *t
	t.SX = a.SX*b.SX + a.SHX*b.SHY + a.TX*b.W0
	t.SHX = a.SX*b.SHX + a.SHX*b.SY + a.TX*b.W1
	t.TX = a.SX*b.TX + a.SHX*b.TY + a.TX*b.W2
	t.SHY = a.SHY*b.SX + a.SY*b.SHY + a.TY*b.W0
	t.SY = a.SHY*b.SHX + a.SY*b.SY + a.TY*b.W1
	t.TY = a.SHY*b.TX + a.SY*b.TY + a.TY*b.W2
	t.W0 = a.W0*b.SX + a.W1*b.SHY + a.W2*b.W0
	t.W1 = a.W0*b.SHX + a.W1*b.SY + a.W2*b.W1
	t.W2 = a.W0*b.TX + a.W1*b.TY + a.W2*b.W2
	return t
}

// MultiplyAffine multiplies the matrix by an affine matrix.
func (t *TransPerspective) MultiplyAffine(a *TransAffine) *TransPerspective {
	b := *t
	t.SX = a.SX*b.SX + a.SHX*b.SHY + a.TX*b.W0
	t.SHX = a.SX*b.SHX + a.SHX*b.SY + a.TX*b.W1
	t.TX = a.SX*b.TX + a.SHX*b.TY + a.TX*b.W2
	t.SHY = a.SHY*b.SX + a.SY*b.SHY + a.TY*b.W0
	t.SY = a.SHY*b.SHX + a.SY*b.SY + a.TY*b.W1
	t.TY = a.SHY*b.TX + a.SY*b.TY + a.TY*b.W2
	return t
}

// Premultiply multiplies "b" by "this" and assigns the result to "this".
func (t *TransPerspective) Premultiply(b *TransPerspective) *TransPerspective {
	a := *t
	t.SX = a.SX*b.SX + a.SHX*b.SHY + a.TX*b.W0
	t.SHX = a.SX*b.SHX + a.SHX*b.SY + a.TX*b.W1
	t.TX = a.SX*b.TX + a.SHX*b.TY + a.TX*b.W2
	t.SHY = a.SHY*b.SX + a.SY*b.SHY + a.TY*b.W0
	t.SY = a.SHY*b.SHX + a.SY*b.SY + a.TY*b.W1
	t.TY = a.SHY*b.TX + a.SY*b.TY + a.TY*b.W2
	t.W0 = a.W0*b.SX + a.W1*b.SHY + a.W2*b.W0
	t.W1 = a.W0*b.SHX + a.W1*b.SY + a.W2*b.W1
	t.W2 = a.W0*b.TX + a.W1*b.TY + a.W2*b.W2
	return t
}

// PremultiplyAffine multiplies "b" by "this" and assigns the result to "this".
func (t *TransPerspective) PremultiplyAffine(b *TransAffine) *TransPerspective {
	a := *t
	t.SX = a.SX*b.SX + a.SHX*b.SHY
	t.SHX = a.SX*b.SHX + a.SHX*b.SY
	t.TX = a.SX*b.TX + a.SHX*b.TY + a.TX
	t.SHY = a.SHY*b.SX + a.SY*b.SHY
	t.SY = a.SHY*b.SHX + a.SY*b.SY
	t.TY = a.SHY*b.TX + a.SY*b.TY + a.TY
	t.W0 = a.W0*b.SX + a.W1*b.SHY
	t.W1 = a.W0*b.SHX + a.W1*b.SY
	t.W2 = a.W0*b.TX + a.W1*b.TY + a.W2
	return t
}

// MultiplyInv multiplies the matrix by the inverse of another perspective matrix.
func (t *TransPerspective) MultiplyInv(m *TransPerspective) *TransPerspective {
	temp := *m
	temp.Invert()
	return t.Multiply(&temp)
}

// MultiplyInvAffine multiplies the matrix by the inverse of an affine matrix.
func (t *TransPerspective) MultiplyInvAffine(m *TransAffine) *TransPerspective {
	temp := *m
	temp.Invert()
	return t.MultiplyAffine(&temp)
}

// PremultiplyInv multiplies the inverse of "m" by "this" and assigns the result to "this".
func (t *TransPerspective) PremultiplyInv(m *TransPerspective) *TransPerspective {
	temp := *m
	temp.Invert()
	*t = temp
	return t.Multiply(m)
}

// PremultiplyInvAffine multiplies the inverse of "m" by "this" and assigns the result to "this".
func (t *TransPerspective) PremultiplyInvAffine(m *TransAffine) *TransPerspective {
	temp := NewTransPerspectiveFromAffine(m)
	temp.Invert()
	*t = *temp.Multiply(t)
	return t
}

// Translate adds translation to the current matrix.
func (t *TransPerspective) Translate(x, y float64) *TransPerspective {
	t.TX += x
	t.TY += y
	return t
}

// Rotate applies rotation to the current matrix.
func (t *TransPerspective) Rotate(angle float64) *TransPerspective {
	return t.MultiplyAffine(NewTransAffineRotation(angle))
}

// Scale applies uniform scaling to the current matrix.
func (t *TransPerspective) Scale(s float64) *TransPerspective {
	return t.MultiplyAffine(NewTransAffineScaling(s))
}

// ScaleXY applies non-uniform scaling to the current matrix.
func (t *TransPerspective) ScaleXY(sx, sy float64) *TransPerspective {
	return t.MultiplyAffine(NewTransAffineScalingXY(sx, sy))
}

// Transform applies the perspective transformation to a point.
func (t *TransPerspective) Transform(x, y *float64) {
	px := *x
	py := *y
	m := 1.0 / (px*t.W0 + py*t.W1 + t.W2)
	*x = m * (px*t.SX + py*t.SHX + t.TX)
	*y = m * (px*t.SHY + py*t.SY + t.TY)
}

// TransformAffine applies only the affine part of the transformation to a point.
func (t *TransPerspective) TransformAffine(x, y *float64) {
	tmp := *x
	*x = tmp*t.SX + *y*t.SHX + t.TX
	*y = tmp*t.SHY + *y*t.SY + t.TY
}

// Transform2x2 applies only the 2x2 matrix part (no translation) to a point.
func (t *TransPerspective) Transform2x2(x, y *float64) {
	tmp := *x
	*x = tmp*t.SX + *y*t.SHX
	*y = tmp*t.SHY + *y*t.SY
}

// InverseTransform applies the inverse transformation to a point.
// It works slow because it explicitly inverts the matrix on every call.
// For massive operations it's better to invert() the matrix and then use direct transformations.
func (t *TransPerspective) InverseTransform(x, y *float64) {
	temp := *t
	if temp.Invert() {
		temp.Transform(x, y)
	}
}

// StoreTo stores the matrix to an array of 9 values.
func (t *TransPerspective) StoreTo(m []float64) {
	m[0] = t.SX
	m[1] = t.SHY
	m[2] = t.W0
	m[3] = t.SHX
	m[4] = t.SY
	m[5] = t.W1
	m[6] = t.TX
	m[7] = t.TY
	m[8] = t.W2
}

// LoadFrom loads the matrix from an array of 9 values.
func (t *TransPerspective) LoadFrom(m []float64) *TransPerspective {
	t.SX = m[0]
	t.SHY = m[1]
	t.W0 = m[2]
	t.SHX = m[3]
	t.SY = m[4]
	t.W1 = m[5]
	t.TX = m[6]
	t.TY = m[7]
	t.W2 = m[8]
	return t
}

// FromAffine converts an affine transformation to perspective.
func (t *TransPerspective) FromAffine(a *TransAffine) *TransPerspective {
	t.SX = a.SX
	t.SHY = a.SHY
	t.W0 = 0.0
	t.SHX = a.SHX
	t.SY = a.SY
	t.W1 = 0.0
	t.TX = a.TX
	t.TY = a.TY
	t.W2 = 1.0
	return t
}

// Determinant calculates the determinant of the transformation matrix.
func (t *TransPerspective) Determinant() float64 {
	return t.SX*(t.SY*t.W2-t.TY*t.W1) +
		t.SHX*(t.TY*t.W0-t.SHY*t.W2) +
		t.TX*(t.SHY*t.W1-t.SY*t.W0)
}

// DeterminantReciprocal calculates the reciprocal of the determinant.
func (t *TransPerspective) DeterminantReciprocal() float64 {
	return 1.0 / t.Determinant()
}

// IsValid checks if the matrix is valid (non-degenerate) with the given epsilon.
func (t *TransPerspective) IsValid(epsilon float64) bool {
	return math.Abs(t.SX) > epsilon &&
		math.Abs(t.SY) > epsilon &&
		math.Abs(t.W2) > epsilon
}

// IsIdentity checks if the matrix is an identity matrix with the given epsilon.
func (t *TransPerspective) IsIdentity(epsilon float64) bool {
	return basics.IsEqualEps(t.SX, 1.0, epsilon) &&
		basics.IsEqualEps(t.SHY, 0.0, epsilon) &&
		basics.IsEqualEps(t.W0, 0.0, epsilon) &&
		basics.IsEqualEps(t.SHX, 0.0, epsilon) &&
		basics.IsEqualEps(t.SY, 1.0, epsilon) &&
		basics.IsEqualEps(t.W1, 0.0, epsilon) &&
		basics.IsEqualEps(t.TX, 0.0, epsilon) &&
		basics.IsEqualEps(t.TY, 0.0, epsilon) &&
		basics.IsEqualEps(t.W2, 1.0, epsilon)
}

// IsEqual checks if this matrix is equal to another matrix with the given epsilon.
func (t *TransPerspective) IsEqual(m *TransPerspective, epsilon float64) bool {
	return basics.IsEqualEps(t.SX, m.SX, epsilon) &&
		basics.IsEqualEps(t.SHY, m.SHY, epsilon) &&
		basics.IsEqualEps(t.W0, m.W0, epsilon) &&
		basics.IsEqualEps(t.SHX, m.SHX, epsilon) &&
		basics.IsEqualEps(t.SY, m.SY, epsilon) &&
		basics.IsEqualEps(t.W1, m.W1, epsilon) &&
		basics.IsEqualEps(t.TX, m.TX, epsilon) &&
		basics.IsEqualEps(t.TY, m.TY, epsilon) &&
		basics.IsEqualEps(t.W2, m.W2, epsilon)
}

// ScaleFactor determines the major affine scale factor.
// Use with caution considering possible degenerate cases.
func (t *TransPerspective) ScaleFactor() float64 {
	x := 0.707106781*t.SX + 0.707106781*t.SHX
	y := 0.707106781*t.SHY + 0.707106781*t.SY
	return math.Sqrt(x*x + y*y)
}

// Rotation determines the major affine rotation angle.
// Use with caution considering possible degenerate cases.
func (t *TransPerspective) Rotation() float64 {
	x1, y1 := 0.0, 0.0
	x2, y2 := 1.0, 0.0
	t.Transform(&x1, &y1)
	t.Transform(&x2, &y2)
	return math.Atan2(y2-y1, x2-x1)
}

// Translation extracts the translation components.
func (t *TransPerspective) Translation() (float64, float64) {
	return t.TX, t.TY
}

// Scaling determines the major affine scaling factors.
// Use with caution considering possible degenerate cases.
func (t *TransPerspective) Scaling() (float64, float64) {
	x1, y1 := 0.0, 0.0
	x2, y2 := 1.0, 1.0
	temp := *t
	temp.MultiplyAffine(NewTransAffineRotation(-temp.Rotation()))
	temp.Transform(&x1, &y1)
	temp.Transform(&x2, &y2)
	return x2 - x1, y2 - y1
}

// ScalingAbs determines the absolute scaling factors.
func (t *TransPerspective) ScalingAbs() (float64, float64) {
	x := math.Sqrt(t.SX*t.SX + t.SHX*t.SHX)
	y := math.Sqrt(t.SHY*t.SHY + t.SY*t.SY)
	return x, y
}

// PerspectiveIteratorX provides an optimized iterator for horizontal line transformations.
// This is useful for rasterization where you need to transform many points along a horizontal line.
type PerspectiveIteratorX struct {
	den      float64
	denStep  float64
	nomX     float64
	nomXStep float64
	nomY     float64
	nomYStep float64
	X, Y     float64
}

// NewIteratorX creates a new horizontal line iterator.
func (t *TransPerspective) NewIteratorX(px, py, step float64) *PerspectiveIteratorX {
	it := &PerspectiveIteratorX{
		den:      px*t.W0 + py*t.W1 + t.W2,
		denStep:  t.W0 * step,
		nomX:     px*t.SX + py*t.SHX + t.TX,
		nomXStep: step * t.SX,
		nomY:     px*t.SHY + py*t.SY + t.TY,
		nomYStep: step * t.SHY,
	}
	it.X = it.nomX / it.den
	it.Y = it.nomY / it.den
	return it
}

// Next advances the iterator to the next position.
func (it *PerspectiveIteratorX) Next() {
	it.den += it.denStep
	it.nomX += it.nomXStep
	it.nomY += it.nomYStep
	d := 1.0 / it.den
	it.X = it.nomX * d
	it.Y = it.nomY * d
}

// Begin creates a new horizontal line iterator for perspective transformation.
// This method matches the AGG C++ API where trans_perspective has a begin() method.
func (t *TransPerspective) Begin(x, y, step float64) *PerspectiveIteratorX {
	return t.NewIteratorX(x, y, step)
}
