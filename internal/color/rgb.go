// Package color provides RGB color types for AGG.
// This implements RGB colors without alpha channel (24-bit RGB).
package color

// RGB represents a floating-point RGB color (base type, no alpha)
type RGB struct {
	R, G, B float64
}

// NewRGB creates a new RGB color
func NewRGB(r, g, b float64) RGB {
	return RGB{R: r, G: g, B: b}
}

// Clear sets the color to black
func (c *RGB) Clear() {
	c.R, c.G, c.B = 0, 0, 0
}

// Scale multiplies by a scalar
func (c RGB) Scale(k float64) RGB {
	return RGB{
		R: c.R * k,
		G: c.G * k,
		B: c.B * k,
	}
}

// Add adds another RGB color
func (c RGB) Add(c2 RGB) RGB {
	return RGB{
		R: c.R + c2.R,
		G: c.G + c2.G,
		B: c.B + c2.B,
	}
}

// Subtract subtracts another RGB color
func (c RGB) Subtract(c2 RGB) RGB {
	return RGB{
		R: c.R - c2.R,
		G: c.G - c2.G,
		B: c.B - c2.B,
	}
}

// Multiply multiplies by another RGB color
func (c RGB) Multiply(c2 RGB) RGB {
	return RGB{
		R: c.R * c2.R,
		G: c.G * c2.G,
		B: c.B * c2.B,
	}
}

// Gradient performs linear interpolation between two RGB colors
func (c RGB) Gradient(c2 RGB, k float64) RGB {
	return RGB{
		R: c.R + k*(c2.R-c.R),
		G: c.G + k*(c2.G-c.G),
		B: c.B + k*(c2.B-c.B),
	}
}

// ToRGBA converts to RGBA with alpha = 1.0
func (c RGB) ToRGBA() RGBA {
	return RGBA{R: c.R, G: c.G, B: c.B, A: 1.0}
}
