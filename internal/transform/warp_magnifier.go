// Package transform provides warp magnifier transformation functionality for AGG.
// This implements a port of AGG's trans_warp_magnifier class.
package transform

import (
	"math"
)

// TransWarpMagnifier implements a lens-like magnification transformation that creates
// a circular area of magnification with smooth falloff outside the radius.
// This is useful for creating magnifying glass effects in graphics applications.
type TransWarpMagnifier struct {
	xc     float64 // center x coordinate
	yc     float64 // center y coordinate
	magn   float64 // magnification factor
	radius float64 // radius of magnified area
}

// NewTransWarpMagnifier creates a new warp magnifier transformation with default values.
// The magnifier is initially centered at origin (0,0) with magnification 1.0 and radius 1.0.
func NewTransWarpMagnifier() *TransWarpMagnifier {
	return &TransWarpMagnifier{
		xc:     0.0,
		yc:     0.0,
		magn:   1.0,
		radius: 1.0,
	}
}

// NewTransWarpMagnifierWithParams creates a new warp magnifier with specified parameters.
func NewTransWarpMagnifierWithParams(xc, yc, magnification, radius float64) *TransWarpMagnifier {
	return &TransWarpMagnifier{
		xc:     xc,
		yc:     yc,
		magn:   magnification,
		radius: radius,
	}
}

// SetCenter sets the center point of the magnification effect.
func (m *TransWarpMagnifier) SetCenter(x, y float64) {
	m.xc = x
	m.yc = y
}

// SetMagnification sets the magnification factor.
// Values > 1.0 magnify (zoom in), values < 1.0 shrink (zoom out).
func (m *TransWarpMagnifier) SetMagnification(magnification float64) {
	m.magn = magnification
}

// SetRadius sets the radius of the magnified area.
func (m *TransWarpMagnifier) SetRadius(radius float64) {
	m.radius = radius
}

// Center returns the center coordinates of the magnification effect.
func (m *TransWarpMagnifier) Center() (x, y float64) {
	return m.xc, m.yc
}

// CenterX returns the x coordinate of the center.
func (m *TransWarpMagnifier) CenterX() float64 {
	return m.xc
}

// CenterY returns the y coordinate of the center.
func (m *TransWarpMagnifier) CenterY() float64 {
	return m.yc
}

// Magnification returns the magnification factor.
func (m *TransWarpMagnifier) Magnification() float64 {
	return m.magn
}

// Radius returns the radius of the magnified area.
func (m *TransWarpMagnifier) Radius() float64 {
	return m.radius
}

// Transform applies the warp magnifier transformation to the given coordinates.
// Points inside the radius are magnified by the magnification factor.
// Points outside the radius experience a smooth falloff effect.
func (m *TransWarpMagnifier) Transform(x, y *float64) {
	dx := *x - m.xc
	dy := *y - m.yc
	r := math.Sqrt(dx*dx + dy*dy)

	if r < m.radius {
		// Inside the magnification radius: apply direct magnification
		*x = m.xc + dx*m.magn
		*y = m.yc + dy*m.magn
		return
	}

	// Outside the radius: apply smooth transition
	// The formula creates a continuous transformation that gradually
	// decreases the magnification effect with distance from center
	mult := (r + m.radius*(m.magn-1.0)) / r
	*x = m.xc + dx*mult
	*y = m.yc + dy*mult
}

// InverseTransform applies the inverse warp magnifier transformation.
// This reverses the magnification effect, allowing for round-trip transformations.
func (m *TransWarpMagnifier) InverseTransform(x, y *float64) {
	dx := *x - m.xc
	dy := *y - m.yc
	r := math.Sqrt(dx*dx + dy*dy)

	if r < m.radius*m.magn {
		// Inside the magnified radius: reverse the magnification
		*x = m.xc + dx/m.magn
		*y = m.yc + dy/m.magn
	} else {
		// Outside the magnified radius: reverse the smooth transition
		rnew := r - m.radius*(m.magn-1.0)
		if r > 0 {
			*x = m.xc + rnew*dx/r
			*y = m.yc + rnew*dy/r
		}
	}
}
