// Package transform provides specialized transformation constructors.
// This implements specialized transformation types from AGG.
package transform

import (
	"math"
)

// NewTransAffineRotation creates a pure rotation transformation matrix.
func NewTransAffineRotation(angle float64) *TransAffine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return NewTransAffineFromValues(cos, sin, -sin, cos, 0.0, 0.0)
}

// NewTransAffineScaling creates a uniform scaling transformation matrix.
func NewTransAffineScaling(scale float64) *TransAffine {
	return NewTransAffineFromValues(scale, 0.0, 0.0, scale, 0.0, 0.0)
}

// NewTransAffineScalingXY creates a non-uniform scaling transformation matrix.
func NewTransAffineScalingXY(sx, sy float64) *TransAffine {
	return NewTransAffineFromValues(sx, 0.0, 0.0, sy, 0.0, 0.0)
}

// NewTransAffineTranslation creates a pure translation transformation matrix.
func NewTransAffineTranslation(tx, ty float64) *TransAffine {
	return NewTransAffineFromValues(1.0, 0.0, 0.0, 1.0, tx, ty)
}

// NewTransAffineSkewing creates a skewing (shear) transformation matrix.
func NewTransAffineSkewing(sx, sy float64) *TransAffine {
	return NewTransAffineFromValues(1.0, math.Tan(sy), math.Tan(sx), 1.0, 0.0, 0.0)
}

// NewTransAffineLineSegment creates a transformation that aligns the coordinate
// system with a line segment, scaling the distance to fit.
func NewTransAffineLineSegment(x1, y1, x2, y2, dist float64) *TransAffine {
	dx := x2 - x1
	dy := y2 - y1

	result := NewTransAffine()

	if dist > 0.0 {
		scale := math.Sqrt(dx*dx+dy*dy) / dist
		result.Multiply(NewTransAffineScaling(scale))
	}

	result.Multiply(NewTransAffineRotation(math.Atan2(dy, dx)))
	result.Multiply(NewTransAffineTranslation(x1, y1))

	return result
}

// NewTransAffineReflectionUnit creates a reflection matrix across a line
// through the origin containing the unit vector (ux, uy).
func NewTransAffineReflectionUnit(ux, uy float64) *TransAffine {
	return NewTransAffineFromValues(
		2.0*ux*ux-1.0,
		2.0*ux*uy,
		2.0*ux*uy,
		2.0*uy*uy-1.0,
		0.0, 0.0)
}

// NewTransAffineReflection creates a reflection matrix across a line
// through the origin at angle a (in radians).
func NewTransAffineReflection(angle float64) *TransAffine {
	ux := math.Cos(angle)
	uy := math.Sin(angle)
	return NewTransAffineReflectionUnit(ux, uy)
}

// NewTransAffineReflectionXY creates a reflection matrix across a line
// through the origin containing the (non-unit) vector (x, y).
func NewTransAffineReflectionXY(x, y float64) *TransAffine {
	length := math.Sqrt(x*x + y*y)
	if length > 0.0 {
		return NewTransAffineReflectionUnit(x/length, y/length)
	}
	return NewTransAffine() // Identity if zero vector
}

// RotateAround creates a transformation that rotates around a specific point.
func (t *TransAffine) RotateAround(angle, cx, cy float64) *TransAffine {
	t.Translate(-cx, -cy)
	t.Rotate(angle)
	t.Translate(cx, cy)
	return t
}

// ScaleAround creates a transformation that scales around a specific point.
func (t *TransAffine) ScaleAround(scale, cx, cy float64) *TransAffine {
	t.Translate(-cx, -cy)
	t.Scale(scale)
	t.Translate(cx, cy)
	return t
}

// ScaleAroundXY creates a transformation that scales around a specific point with different X/Y factors.
func (t *TransAffine) ScaleAroundXY(sx, sy, cx, cy float64) *TransAffine {
	t.Translate(-cx, -cy)
	t.ScaleXY(sx, sy)
	t.Translate(cx, cy)
	return t
}

// NewTransAffineRotateAround creates a new transformation that rotates around a specific point.
func NewTransAffineRotateAround(angle, cx, cy float64) *TransAffine {
	result := NewTransAffine()
	return result.RotateAround(angle, cx, cy)
}

// NewTransAffineScaleAround creates a new transformation that scales around a specific point.
func NewTransAffineScaleAround(scale, cx, cy float64) *TransAffine {
	result := NewTransAffine()
	return result.ScaleAround(scale, cx, cy)
}

// NewTransAffineScaleAroundXY creates a new transformation that scales around a specific point with different X/Y factors.
func NewTransAffineScaleAroundXY(sx, sy, cx, cy float64) *TransAffine {
	result := NewTransAffine()
	return result.ScaleAroundXY(sx, sy, cx, cy)
}
