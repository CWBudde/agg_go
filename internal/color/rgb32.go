// Package color provides RGB color types for AGG.
// This implements RGB colors without alpha channel (24-bit RGB).
package color

import (
	"agg_go/internal/basics"
)

// RGB32 represents a 32-bit floating-point RGB color
type RGB32[CS Space] struct {
	R, G, B float32
}

// NewRGB32 creates a new 32-bit RGB color
func NewRGB32[CS Space](r, g, b float32) RGB32[CS] {
	return RGB32[CS]{R: r, G: g, B: b}
}

// ConvertToRGB converts to floating-point RGB
func (c RGB32[CS]) ConvertToRGB() RGB {
	return RGB{
		R: float64(c.R),
		G: float64(c.G),
		B: float64(c.B),
	}
}

// ConvertToRGBA converts to floating-point RGBA with alpha = 1.0
func (c RGB32[CS]) ConvertToRGBA() RGBA {
	return RGBA{
		R: float64(c.R),
		G: float64(c.G),
		B: float64(c.B),
		A: 1.0,
	}
}

// Convert converts between colorspaces for RGB32
// This is a self-converting method that doesn't change colorspace but
// can be used as a base for colorspace-specific conversions
func (c RGB32[CS]) Convert() RGB32[CS] {
	return c
}

// ConvertFromRGB converts from floating-point RGB
func ConvertFromRGB32[CS Space](c RGB) RGB32[CS] {
	return RGB32[CS]{
		R: float32(c.R),
		G: float32(c.G),
		B: float32(c.B),
	}
}

// ConvertRGBAToRGB32 converts from RGBA (ignores alpha)
func ConvertRGBAToRGB32[CS Space](c RGBA) RGB32[CS] {
	return RGB32[CS]{
		R: float32(c.R),
		G: float32(c.G),
		B: float32(c.B),
	}
}

// ToRGBA32 converts to RGBA32 with alpha = 1.0
func (c RGB32[CS]) ToRGBA32() RGBA32[CS] {
	return RGBA32[CS]{R: c.R, G: c.G, B: c.B, A: 1.0}
}

// Gradient performs linear interpolation between two 32-bit RGB colors
func (c RGB32[CS]) Gradient(c2 RGB32[CS], k float32) RGB32[CS] {
	return RGB32[CS]{
		R: c.R + (c2.R-c.R)*k,
		G: c.G + (c2.G-c.G)*k,
		B: c.B + (c2.B-c.B)*k,
	}
}

// Clear sets the color to black
func (c *RGB32[CS]) Clear() {
	c.R, c.G, c.B = 0.0, 0.0, 0.0
}

// Add adds another RGB32 color
func (c RGB32[CS]) Add(c2 RGB32[CS]) RGB32[CS] {
	return RGB32[CS]{
		R: c.R + c2.R,
		G: c.G + c2.G,
		B: c.B + c2.B,
	}
}

// Scale multiplies the RGB color by a scalar value
func (c RGB32[CS]) Scale(k float32) RGB32[CS] {
	return RGB32[CS]{
		R: c.R * k,
		G: c.G * k,
		B: c.B * k,
	}
}

// IsBlack returns true if the color is black (within floating-point epsilon)
func (c RGB32[CS]) IsBlack() bool {
	const epsilon = 1e-6
	return c.R < epsilon && c.G < epsilon && c.B < epsilon
}

// IsWhite returns true if the color is white (within floating-point epsilon)
func (c RGB32[CS]) IsWhite() bool {
	const epsilon = 1e-5
	return (c.R > 1.0-epsilon) && (c.G > 1.0-epsilon) && (c.B > 1.0-epsilon)
}

// Luminance calculates the ITU-R BT.709 luminance
func (c RGB32[CS]) Luminance() float32 {
	// ITU-R BT.709: Y = 0.2126*R + 0.7152*G + 0.0722*B
	return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
}

// Common 32-bit RGB color types
type (
	RGB32Linear = RGB32[Linear]
	RGB32SRGB   = RGB32[SRGB]
)

// RGB24Order and BGR24Order type markers
type (
	RGB24Order struct{}
	BGR24Order struct{}
)

// RGB color order structures for 24-bit formats
var (
	OrderRGB24 = ColorOrder{R: 0, G: 1, B: 2, A: -1} // A = -1 indicates no alpha
	OrderBGR24 = ColorOrder{R: 2, G: 1, B: 0, A: -1}
)

// Helper functions for RGB24 operations

// RGB8Multiply24 performs fixed-point multiplication for 8-bit RGB values
func RGB8Multiply24(a, b basics.Int8u) basics.Int8u {
	return RGBA8Multiply(a, b) // Reuse RGBA function
}

// RGB8Lerp24 performs linear interpolation between two RGB values
func RGB8Lerp24(p, q, a basics.Int8u) basics.Int8u {
	return RGBA8Lerp(p, q, a) // Reuse RGBA function
}

// RGB8MultCover24 multiplies an RGB component by coverage
func RGB8MultCover24(c, cover basics.Int8u) basics.Int8u {
	return RGBA8MultCover(c, cover) // Reuse RGBA function
}

// RGB16 manipulation functions for 16-bit RGB operations

// RGB16Lerp performs linear interpolation between two 16-bit RGB values
func RGB16Lerp(p, q, a basics.Int16u) basics.Int16u {
	return RGBA16Lerp(p, q, a) // Reuse RGBA function
}

// RGB16Prelerp performs premultiplied linear interpolation for 16-bit RGB
func RGB16Prelerp(p, q, a basics.Int16u) basics.Int16u {
	return RGBA16Prelerp(p, q, a) // Reuse RGBA function
}

// RGB16MultCover multiplies a 16-bit RGB component by coverage
func RGB16MultCover(c, cover basics.Int16u) basics.Int16u {
	return RGBA16MultCover(c, cover) // Reuse RGBA function
}

// Common RGB constants
var (
	RGB8Black   = RGB8Linear{R: 0, G: 0, B: 0}
	RGB8White   = RGB8Linear{R: 255, G: 255, B: 255}
	RGB8Red     = RGB8Linear{R: 255, G: 0, B: 0}
	RGB8Green   = RGB8Linear{R: 0, G: 255, B: 0}
	RGB8Blue    = RGB8Linear{R: 0, G: 0, B: 255}
	RGB8Cyan    = RGB8Linear{R: 0, G: 255, B: 255}
	RGB8Magenta = RGB8Linear{R: 255, G: 0, B: 255}
	RGB8Yellow  = RGB8Linear{R: 255, G: 255, B: 0}
)
