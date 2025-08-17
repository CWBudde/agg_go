// Package color provides RGB color types for AGG.
// This implements RGB colors without alpha channel (24-bit RGB).
package color

import (
	"agg_go/internal/basics"
)

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

// RGB8 represents an 8-bit RGB color with colorspace (24-bit, 3 bytes)
type RGB8[CS any] struct {
	R, G, B basics.Int8u
}

// NewRGB8 creates a new 8-bit RGB color
func NewRGB8[CS any](r, g, b basics.Int8u) RGB8[CS] {
	return RGB8[CS]{R: r, G: g, B: b}
}

// Convert converts between colorspaces for RGB8
func (c RGB8[CS]) Convert() RGB8[CS] {
	// For now, just return the same color
	// TODO: Implement proper colorspace conversion
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
func ConvertFromRGB[CS any](c RGB) RGB8[CS] {
	return RGB8[CS]{
		R: basics.Int8u(c.R*255 + 0.5),
		G: basics.Int8u(c.G*255 + 0.5),
		B: basics.Int8u(c.B*255 + 0.5),
	}
}

// ConvertRGBAToRGB8 converts from RGBA (ignores alpha)
func ConvertRGBAToRGB8[CS any](c RGBA) RGB8[CS] {
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

// Common RGB8 color types
type (
	RGB8Linear = RGB8[Linear]
	RGB8SRGB   = RGB8[SRGB]
	SRGB8      = RGB8[SRGB] // Alias for backwards compatibility
)

// RGB16 represents a 16-bit RGB color
type RGB16[CS any] struct {
	R, G, B basics.Int16u
}

// NewRGB16 creates a new 16-bit RGB color
func NewRGB16[CS any](r, g, b basics.Int16u) RGB16[CS] {
	return RGB16[CS]{R: r, G: g, B: b}
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

// Common 16-bit RGB color types
type (
	RGB16Linear = RGB16[Linear]
	RGB16SRGB   = RGB16[SRGB]
)

// RGB32 represents a 32-bit floating-point RGB color
type RGB32[CS any] struct {
	R, G, B float32
}

// NewRGB32 creates a new 32-bit RGB color
func NewRGB32[CS any](r, g, b float32) RGB32[CS] {
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
