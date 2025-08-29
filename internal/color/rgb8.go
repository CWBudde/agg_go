// Package color provides RGB color types for AGG.
// This implements RGB colors without alpha channel (24-bit RGB).
package color

import (
	"agg_go/internal/basics"
)

// RGB8 represents an 8-bit RGB color with colorspace (24-bit, 3 bytes)
type RGB8[CS Space] struct {
	R, G, B basics.Int8u
}

// NewRGB8 creates a new 8-bit RGB color
func NewRGB8[CS Space](r, g, b basics.Int8u) RGB8[CS] {
	return RGB8[CS]{R: r, G: g, B: b}
}

// Convert converts between colorspaces for RGB8
// This is a self-converting method that doesn't change colorspace but
// can be used as a base for colorspace-specific conversions
func (c RGB8[CS]) Convert() RGB8[CS] {
	return c
}

// ConvertToRGB converts to floating-point RGB
func (c RGB8[CS]) ConvertToRGB() RGB {
	const scale = 1.0 / 255.0
	return RGB{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
	}
}

// ConvertToRGBA converts to floating-point RGBA with alpha = 1.0
func (c RGB8[CS]) ConvertToRGBA() RGBA {
	const scale = 1.0 / 255.0
	return RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: 1.0,
	}
}

// ConvertFromRGB converts from floating-point RGB
func ConvertFromRGB[CS Space](c RGB) RGB8[CS] {
	return RGB8[CS]{
		R: basics.Int8u(c.R*255 + 0.5),
		G: basics.Int8u(c.G*255 + 0.5),
		B: basics.Int8u(c.B*255 + 0.5),
	}
}

// ConvertRGBAToRGB8 converts from RGBA (ignores alpha)
func ConvertRGBAToRGB8[CS Space](c RGBA) RGB8[CS] {
	return RGB8[CS]{
		R: basics.Int8u(c.R*255 + 0.5),
		G: basics.Int8u(c.G*255 + 0.5),
		B: basics.Int8u(c.B*255 + 0.5),
	}
}

// ToRGBA8 converts to RGBA8 with alpha = 255
func (c RGB8[CS]) ToRGBA8() RGBA8[CS] {
	return RGBA8[CS]{R: c.R, G: c.G, B: c.B, A: 255}
}

// Gradient performs linear interpolation between two 8-bit RGB colors
func (c RGB8[CS]) Gradient(c2 RGB8[CS], k basics.Int8u) RGB8[CS] {
	return RGB8[CS]{
		R: RGBA8Lerp(c.R, c2.R, k),
		G: RGBA8Lerp(c.G, c2.G, k),
		B: RGBA8Lerp(c.B, c2.B, k),
	}
}

// Clear sets the color to black
func (c *RGB8[CS]) Clear() {
	c.R, c.G, c.B = 0, 0, 0
}

// Add adds another RGB8 color
func (c RGB8[CS]) Add(c2 RGB8[CS]) RGB8[CS] {
	return RGB8[CS]{
		R: basics.Int8u(minUint32(uint32(c.R)+uint32(c2.R), 255)),
		G: basics.Int8u(minUint32(uint32(c.G)+uint32(c2.G), 255)),
		B: basics.Int8u(minUint32(uint32(c.B)+uint32(c2.B), 255)),
	}
}

// Scale multiplies the RGB color by a scalar value
func (c RGB8[CS]) Scale(k float64) RGB8[CS] {
	return RGB8[CS]{
		R: basics.Int8u(minFloat64(float64(c.R)*k+0.5, 255)),
		G: basics.Int8u(minFloat64(float64(c.G)*k+0.5, 255)),
		B: basics.Int8u(minFloat64(float64(c.B)*k+0.5, 255)),
	}
}

// IsBlack returns true if the color is black
func (c RGB8[CS]) IsBlack() bool {
	return c.R == 0 && c.G == 0 && c.B == 0
}

// IsWhite returns true if the color is white
func (c RGB8[CS]) IsWhite() bool {
	return c.R == 255 && c.G == 255 && c.B == 255
}

// Luminance calculates the ITU-R BT.709 luminance
func (c RGB8[CS]) Luminance() basics.Int8u {
	// ITU-R BT.709: Y = 0.2126*R + 0.7152*G + 0.0722*B
	// Using fixed-point arithmetic for performance
	return basics.Int8u((uint32(c.R)*54 + uint32(c.G)*183 + uint32(c.B)*18) >> 8)
}

// Apply 8-bit gamma (RGB only) to an RGB8 pixel in-place.
func ApplyGammaDir8RGB[CS Space, LUT lut8Like](px *RGB8[CS], lut LUT) {
	px.R = lut.Dir(px.R)
	px.G = lut.Dir(px.G)
	px.B = lut.Dir(px.B)
}

func ApplyGammaInv8RGB[CS Space, LUT lut8Like](px *RGB8[CS], lut LUT) {
	px.R = lut.Inv(px.R)
	px.G = lut.Inv(px.G)
	px.B = lut.Inv(px.B)
}

// Helper methods for RGB8
func (c *RGB8[CS]) ApplyGammaDir(lut lut8Like) { ApplyGammaDir8RGB(c, lut) }
func (c *RGB8[CS]) ApplyGammaInv(lut lut8Like) { ApplyGammaInv8RGB(c, lut) }

// Common RGB8 color types
type (
	RGB8Linear = RGB8[Linear]
	RGB8SRGB   = RGB8[SRGB]
	SRGB8      = RGB8[SRGB] // Alias for backwards compatibility
)
