// Package color provides RGB color types for AGG.
// This implements RGB colors without alpha channel (24-bit RGB).
package color

import (
	"agg_go/internal/basics"
)

// RGB16 represents a 16-bit RGB color
type RGB16[CS Space] struct {
	R, G, B basics.Int16u
}

// NewRGB16 creates a new 16-bit RGB color
func NewRGB16[CS Space](r, g, b basics.Int16u) RGB16[CS] {
	return RGB16[CS]{R: r, G: g, B: b}
}

// Convert converts between colorspaces for RGB16
// This is a self-converting method that doesn't change colorspace but
// can be used as a base for colorspace-specific conversions
func (c RGB16[CS]) Convert() RGB16[CS] {
	return c
}

// ConvertToRGB converts to floating-point RGB
func (c RGB16[CS]) ConvertToRGB() RGB {
	const scale = 1.0 / 65535.0
	return RGB{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
	}
}

// ConvertToRGBA converts to floating-point RGBA with alpha = 1.0
func (c RGB16[CS]) ConvertToRGBA() RGBA {
	const scale = 1.0 / 65535.0
	return RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: 1.0,
	}
}

// ConvertFromRGB converts from floating-point RGB
func ConvertFromRGB16[CS Space](c RGB) RGB16[CS] {
	return RGB16[CS]{
		R: basics.Int16u(c.R*65535 + 0.5),
		G: basics.Int16u(c.G*65535 + 0.5),
		B: basics.Int16u(c.B*65535 + 0.5),
	}
}

// ConvertRGBAToRGB16 converts from RGBA (ignores alpha)
func ConvertRGBAToRGB16[CS Space](c RGBA) RGB16[CS] {
	return RGB16[CS]{
		R: basics.Int16u(c.R*65535 + 0.5),
		G: basics.Int16u(c.G*65535 + 0.5),
		B: basics.Int16u(c.B*65535 + 0.5),
	}
}

// ToRGBA16 converts to RGBA16 with alpha = 65535
func (c RGB16[CS]) ToRGBA16() RGBA16[CS] {
	return RGBA16[CS]{R: c.R, G: c.G, B: c.B, A: 65535}
}

// Gradient performs linear interpolation between two 16-bit RGB colors
func (c RGB16[CS]) Gradient(c2 RGB16[CS], k basics.Int16u) RGB16[CS] {
	return RGB16[CS]{
		R: RGBA16Lerp(c.R, c2.R, k),
		G: RGBA16Lerp(c.G, c2.G, k),
		B: RGBA16Lerp(c.B, c2.B, k),
	}
}

// Clear sets the color to black
func (c *RGB16[CS]) Clear() {
	c.R, c.G, c.B = 0, 0, 0
}

// Add adds another RGB16 color
func (c RGB16[CS]) Add(c2 RGB16[CS]) RGB16[CS] {
	return RGB16[CS]{
		R: basics.Int16u(minUint32(uint32(c.R)+uint32(c2.R), 65535)),
		G: basics.Int16u(minUint32(uint32(c.G)+uint32(c2.G), 65535)),
		B: basics.Int16u(minUint32(uint32(c.B)+uint32(c2.B), 65535)),
	}
}

// Scale multiplies the RGB color by a scalar value
func (c RGB16[CS]) Scale(k float64) RGB16[CS] {
	return RGB16[CS]{
		R: basics.Int16u(minFloat64(float64(c.R)*k+0.5, 65535)),
		G: basics.Int16u(minFloat64(float64(c.G)*k+0.5, 65535)),
		B: basics.Int16u(minFloat64(float64(c.B)*k+0.5, 65535)),
	}
}

// IsBlack returns true if the color is black
func (c RGB16[CS]) IsBlack() bool {
	return c.R == 0 && c.G == 0 && c.B == 0
}

// IsWhite returns true if the color is white
func (c RGB16[CS]) IsWhite() bool {
	return c.R == 65535 && c.G == 65535 && c.B == 65535
}

// Luminance calculates the ITU-R BT.709 luminance
func (c RGB16[CS]) Luminance() basics.Int16u {
	// ITU-R BT.709: Y = 0.2126*R + 0.7152*G + 0.0722*B
	// Using fixed-point arithmetic for performance, scaled for 16-bit
	return basics.Int16u((uint32(c.R)*13933 + uint32(c.G)*46871 + uint32(c.B)*4731) >> 16)
}

// Common 16-bit RGB color types
type (
	RGB16Linear = RGB16[Linear]
	RGB16SRGB   = RGB16[SRGB]
)
