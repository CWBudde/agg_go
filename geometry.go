// Package agg provides geometric types and operations for 2D graphics.
// This file contains rectangles, points, and geometric utilities.
package agg

import "agg_go/internal/transform"

// Rect represents an integer rectangle.
type Rect struct {
	X1, Y1, X2, Y2 int
}

// NewRect creates a new rectangle with the specified coordinates.
func NewRect(x1, y1, x2, y2 int) Rect {
	return Rect{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

// Width returns the width of the rectangle.
func (r Rect) Width() int {
	return r.X2 - r.X1
}

// Height returns the height of the rectangle.
func (r Rect) Height() int {
	return r.Y2 - r.Y1
}

// Affine represents a 2D affine transformation matrix.
// This matches AGG's trans_affine structure.
type Affine struct {
	matrix [6]float64 // [sx, shy, shx, sy, tx, ty]
}

// NewAffine creates an identity affine transformation.
func NewAffine() *Affine {
	a := &Affine{}
	a.Reset()
	return a
}

// Reset sets the matrix to identity.
func (a *Affine) Reset() {
	a.matrix[0] = 1.0 // sx
	a.matrix[1] = 0.0 // shy
	a.matrix[2] = 0.0 // shx
	a.matrix[3] = 1.0 // sy
	a.matrix[4] = 0.0 // tx
	a.matrix[5] = 0.0 // ty
}

// Transform applies the affine transformation to a point.
func (a *Affine) Transform(x, y float64) (float64, float64) {
	tx := a.matrix[0]*x + a.matrix[2]*y + a.matrix[4]
	ty := a.matrix[1]*x + a.matrix[3]*y + a.matrix[5]
	return tx, ty
}

// Point represents a 2D point with floating-point coordinates.
type Point struct {
	X, Y float64
}

// NewPoint creates a new point with the specified coordinates.
func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Distance calculates the distance between two points.
func (p Point) Distance(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return geomSqrt(dx*dx + dy*dy)
}

// PointI represents a 2D point with integer coordinates.
type PointI struct {
	X, Y int
}

// NewPointI creates a new integer point.
func NewPointI(x, y int) PointI {
	return PointI{X: x, Y: y}
}

// ToFloat converts integer point to floating-point point.
func (p PointI) ToFloat() Point {
	return Point{X: float64(p.X), Y: float64(p.Y)}
}

// Size represents dimensions with floating-point values.
type Size struct {
	Width, Height float64
}

// NewSize creates a new size with the specified dimensions.
func NewSize(width, height float64) Size {
	return Size{Width: width, Height: height}
}

// SizeI represents dimensions with integer values.
type SizeI struct {
	Width, Height int
}

// NewSizeI creates a new integer size.
func NewSizeI(width, height int) SizeI {
	return SizeI{Width: width, Height: height}
}

// ToFloat converts integer size to floating-point size.
func (s SizeI) ToFloat() Size {
	return Size{Width: float64(s.Width), Height: float64(s.Height)}
}

// Geometric utility functions

// Distance calculates the distance between two points.
func Distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return geomSqrt(dx*dx + dy*dy)
}

// Angle calculates the angle between two points (in radians).
func Angle(x1, y1, x2, y2 float64) float64 {
	return geomAtan2(y2-y1, x2-x1)
}

// Helper functions (prefixed to avoid conflicts)
func geomSqrt(x float64) float64 {
	// Simple square root approximation - in practice would use math.Sqrt
	if x <= 0 {
		return 0
	}

	// Newton's method approximation
	guess := x / 2
	for i := 0; i < 10; i++ {
		better := (guess + x/guess) / 2
		if geomAbs(better-guess) < 0.00001 {
			break
		}
		guess = better
	}
	return guess
}

func geomAtan2(y, x float64) float64 {
	// Simple atan2 approximation - in practice would use math.Atan2
	if x > 0 {
		return geomAtan(y / x)
	}
	if x < 0 && y >= 0 {
		return geomAtan(y/x) + 3.14159265359
	}
	if x < 0 && y < 0 {
		return geomAtan(y/x) - 3.14159265359
	}
	if x == 0 && y > 0 {
		return 3.14159265359 / 2
	}
	if x == 0 && y < 0 {
		return -3.14159265359 / 2
	}
	return 0 // x == 0 && y == 0
}

func geomAtan(x float64) float64 {
	// Simple atan approximation using Taylor series
	// For more accuracy, would use math.Atan
	if geomAbs(x) > 1 {
		return 3.14159265359/2 - geomAtan(1/x)
	}

	// Taylor series: atan(x) = x - x^3/3 + x^5/5 - x^7/7 + ...
	result := x
	term := x
	for i := 1; i < 10; i++ {
		term *= -x * x
		result += term / float64(2*i+1)
	}
	return result
}

func geomAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Conversion to internal transform types

// ToTransAffine converts Affine to internal transform.TransAffine.
func (a *Affine) ToTransAffine() *transform.TransAffine {
	return transform.NewTransAffineFromValues(
		a.matrix[0], a.matrix[1], a.matrix[2],
		a.matrix[3], a.matrix[4], a.matrix[5])
}

// FromTransAffine creates an Affine from internal transform.TransAffine.
func FromTransAffine(t *transform.TransAffine) *Affine {
	a := &Affine{}
	a.matrix[0] = t.SX
	a.matrix[1] = t.SHY
	a.matrix[2] = t.SHX
	a.matrix[3] = t.SY
	a.matrix[4] = t.TX
	a.matrix[5] = t.TY
	return a
}
