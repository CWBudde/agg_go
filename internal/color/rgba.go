// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

import (
	"math"
)

// ColorOrder defines component ordering for different pixel formats
type ColorOrder struct {
	R, G, B, A int
}

var (
	OrderRGB  = ColorOrder{R: 0, G: 1, B: 2, A: 3}
	OrderBGR  = ColorOrder{R: 2, G: 1, B: 0, A: 3}
	OrderRGBA = ColorOrder{R: 0, G: 1, B: 2, A: 3}
	OrderARGB = ColorOrder{R: 1, G: 2, B: 3, A: 0}
	OrderABGR = ColorOrder{R: 3, G: 2, B: 1, A: 0}
	OrderBGRA = ColorOrder{R: 2, G: 1, B: 0, A: 3}
)

// RGBA represents a floating-point RGBA color (base type)
type RGBA struct {
	R, G, B, A float64
}

// NewRGBA creates a new RGBA color
func NewRGBA(r, g, b, a float64) RGBA {
	return RGBA{R: r, G: g, B: b, A: a}
}

// Clear sets the color to transparent black
func (c *RGBA) Clear() *RGBA {
	c.R, c.G, c.B, c.A = 0, 0, 0, 0
	return c
}

// Transparent sets the color to transparent with the same RGB
func (c *RGBA) Transparent() *RGBA {
	c.A = 0
	return c
}

// Opacity sets the alpha channel
func (c *RGBA) Opacity(a float64) *RGBA {
	if a < 0 {
		c.A = 0
	} else if a > 1 {
		c.A = 1
	} else {
		c.A = a
	}
	return c
}

// GetOpacity returns the current alpha value
func (c RGBA) GetOpacity() float64 {
	return c.A
}

// Premultiply premultiplies the color by alpha
func (c *RGBA) Premultiply() *RGBA {
	if c.A <= 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A < 1 {
		c.R *= c.A
		c.G *= c.A
		c.B *= c.A
	}
	return c
}

// PremultiplyAlpha premultiplies the color by a specific alpha value
func (c *RGBA) PremultiplyAlpha(a float64) *RGBA {
	if c.A <= 0 || a <= 0 {
		c.R, c.G, c.B, c.A = 0, 0, 0, 0
	} else {
		a /= c.A
		c.R *= a
		c.G *= a
		c.B *= a
		c.A = a
	}
	return c
}

// Demultiply demultiplies the color by alpha
func (c *RGBA) Demultiply() *RGBA {
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

// Gradient performs linear interpolation between two colors
func (c RGBA) Gradient(c2 RGBA, k float64) RGBA {
	return RGBA{
		R: c.R + k*(c2.R-c.R),
		G: c.G + k*(c2.G-c.G),
		B: c.B + k*(c2.B-c.B),
		A: c.A + k*(c2.A-c.A),
	}
}

// Add adds another color
func (c RGBA) Add(c2 RGBA) RGBA {
	return RGBA{
		R: c.R + c2.R,
		G: c.G + c2.G,
		B: c.B + c2.B,
		A: c.A + c2.A,
	}
}

// Subtract subtracts another color
func (c RGBA) Subtract(c2 RGBA) RGBA {
	return RGBA{
		R: c.R - c2.R,
		G: c.G - c2.G,
		B: c.B - c2.B,
		A: c.A - c2.A,
	}
}

// Multiply multiplies by another color
func (c RGBA) Multiply(c2 RGBA) RGBA {
	return RGBA{
		R: c.R * c2.R,
		G: c.G * c2.G,
		B: c.B * c2.B,
		A: c.A * c2.A,
	}
}

// Scale multiplies by a scalar
func (c RGBA) Scale(k float64) RGBA {
	return RGBA{
		R: c.R * k,
		G: c.G * k,
		B: c.B * k,
		A: c.A * k,
	}
}

// AddAssign adds another color (equivalent to C++ +=)
func (c *RGBA) AddAssign(c2 RGBA) *RGBA {
	c.R += c2.R
	c.G += c2.G
	c.B += c2.B
	c.A += c2.A
	return c
}

// MultiplyAssign multiplies by a scalar (equivalent to C++ *=)
func (c *RGBA) MultiplyAssign(k float64) *RGBA {
	c.R *= k
	c.G *= k
	c.B *= k
	c.A *= k
	return c
}

// NoColor returns transparent black
func NoColor() RGBA {
	return RGBA{0, 0, 0, 0}
}

// FromWavelength creates an RGB color from a wavelength (380-780 nm)
// This follows the C++ AGG implementation exactly
func FromWavelength(wl, gamma float64) RGBA {
	t := RGBA{0.0, 0.0, 0.0, 1.0}

	if wl >= 380.0 && wl <= 440.0 {
		t.R = -1.0 * (wl - 440.0) / (440.0 - 380.0)
		t.B = 1.0
	} else if wl >= 440.0 && wl <= 490.0 {
		t.G = (wl - 440.0) / (490.0 - 440.0)
		t.B = 1.0
	} else if wl >= 490.0 && wl <= 510.0 {
		t.G = 1.0
		t.B = -1.0 * (wl - 510.0) / (510.0 - 490.0)
	} else if wl >= 510.0 && wl <= 580.0 {
		t.R = (wl - 510.0) / (580.0 - 510.0)
		t.G = 1.0
	} else if wl >= 580.0 && wl <= 645.0 {
		t.R = 1.0
		t.G = -1.0 * (wl - 645.0) / (645.0 - 580.0)
	} else if wl >= 645.0 && wl <= 780.0 {
		t.R = 1.0
	}

	s := 1.0
	if wl > 700.0 {
		s = 0.3 + 0.7*(780.0-wl)/(780.0-700.0)
	} else if wl < 420.0 {
		s = 0.3 + 0.7*(wl-380.0)/(420.0-380.0)
	}

	t.R = math.Pow(t.R*s, gamma)
	t.G = math.Pow(t.G*s, gamma)
	t.B = math.Pow(t.B*s, gamma)
	return t
}

// NewRGBAFromWavelength creates an RGBA color from wavelength (constructor equivalent)
func NewRGBAFromWavelength(wl, gamma float64) RGBA {
	return FromWavelength(wl, gamma)
}

// RGBAPre creates a premultiplied RGBA color
func RGBAPre(r, g, b, a float64) RGBA {
	c := NewRGBA(r, g, b, a)
	c.Premultiply()
	return c
}
