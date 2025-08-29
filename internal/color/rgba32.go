// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

import (
	"agg_go/internal/basics"
)

// RGBA32 represents a 32-bit floating-point RGBA color
type RGBA32[CS ColorSpace] struct {
	R, G, B, A float32
}

// NewRGBA32 creates a new 32-bit RGBA color
func NewRGBA32[CS ColorSpace](r, g, b, a float32) RGBA32[CS] {
	return RGBA32[CS]{R: r, G: g, B: b, A: a}
}

// ConvertToRGBA converts to floating-point RGBA
func (c RGBA32[CS]) ConvertToRGBA() RGBA {
	return RGBA{
		R: float64(c.R),
		G: float64(c.G),
		B: float64(c.B),
		A: float64(c.A),
	}
}

// ConvertFromRGBA converts from floating-point RGBA
func ConvertFromRGBA32[CS ColorSpace](c RGBA) RGBA32[CS] {
	return RGBA32[CS]{
		R: float32(c.R),
		G: float32(c.G),
		B: float32(c.B),
		A: float32(c.A),
	}
}

// Premultiply premultiplies the color by alpha
func (c *RGBA32[CS]) Premultiply() *RGBA32[CS] {
	if c.A <= 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A < 1 {
		c.R *= c.A
		c.G *= c.A
		c.B *= c.A
	}
	return c
}

// Demultiply demultiplies the color by alpha
func (c *RGBA32[CS]) Demultiply() *RGBA32[CS] {
	if c.A == 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A > 0 && c.A < 1 {
		inv := 1.0 / c.A
		c.R *= inv
		c.G *= inv
		c.B *= inv
	}
	return c
}

// Gradient performs linear interpolation between two 32-bit colors
func (c RGBA32[CS]) Gradient(c2 RGBA32[CS], k float32) RGBA32[CS] {
	return RGBA32[CS]{
		R: c.R + k*(c2.R-c.R),
		G: c.G + k*(c2.G-c.G),
		B: c.B + k*(c2.B-c.B),
		A: c.A + k*(c2.A-c.A),
	}
}

// Clear sets the color to transparent black
func (c *RGBA32[CS]) Clear() *RGBA32[CS] {
	c.R, c.G, c.B, c.A = 0, 0, 0, 0
	return c
}

// Transparent sets the color to transparent with the same RGB
func (c *RGBA32[CS]) Transparent() *RGBA32[CS] {
	c.A = 0
	return c
}

// IsTransparent returns true if the color is fully transparent
func (c RGBA32[CS]) IsTransparent() bool {
	return c.A == 0
}

// IsOpaque returns true if the color is fully opaque
func (c RGBA32[CS]) IsOpaque() bool {
	return c.A == 1.0
}

// Opacity sets the alpha channel (0.0 to 1.0)
func (c *RGBA32[CS]) Opacity(a float64) *RGBA32[CS] {
	if a < 0 {
		c.A = 0
	} else if a > 1 {
		c.A = 1
	} else {
		c.A = float32(a)
	}
	return c
}

// GetOpacity returns the alpha as a float64 (0.0 to 1.0)
func (c RGBA32[CS]) GetOpacity() float64 {
	return float64(c.A)
}

// Add adds another RGBA32 color
func (c RGBA32[CS]) Add(c2 RGBA32[CS]) RGBA32[CS] {
	return RGBA32[CS]{
		R: c.R + c2.R,
		G: c.G + c2.G,
		B: c.B + c2.B,
		A: c.A + c2.A,
	}
}

// AddWithCover adds another RGBA32 color with coverage, matching C++ AGG's add(color, cover) method
func (c *RGBA32[CS]) AddWithCover(c2 RGBA32[CS], cover basics.Int8u) {
	coverFloat := float32(cover) / 255.0
	if cover == basics.CoverMask {
		if c2.A == 1.0 { // base_mask for float32 (1.0)
			*c = c2
			return
		} else {
			c.R = minFloat32(c.R+c2.R, 1.0)
			c.G = minFloat32(c.G+c2.G, 1.0)
			c.B = minFloat32(c.B+c2.B, 1.0)
			c.A = minFloat32(c.A+c2.A, 1.0)
		}
	} else {
		c.R = minFloat32(c.R+c2.R*coverFloat, 1.0)
		c.G = minFloat32(c.G+c2.G*coverFloat, 1.0)
		c.B = minFloat32(c.B+c2.B*coverFloat, 1.0)
		c.A = minFloat32(c.A+c2.A*coverFloat, 1.0)
	}
}

// Scale multiplies the color by a scalar value
func (c RGBA32[CS]) Scale(k float32) RGBA32[CS] {
	return RGBA32[CS]{
		R: c.R * k,
		G: c.G * k,
		B: c.B * k,
		A: c.A * k,
	}
}

// Common 32-bit color types
type (
	RGBA32Linear = RGBA32[Linear]
	RGBA32SRGB   = RGBA32[SRGB]
)
