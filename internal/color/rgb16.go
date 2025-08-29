// Package color provides RGB color types for AGG.
// This implements RGB colors without alpha channel (24-bit RGB).
package color

import (
	"agg_go/internal/basics"
)

// RGB16 represents a 16-bit RGB color
type RGB16[CS ColorSpace] struct {
	R, G, B basics.Int16u
}

// NewRGB16 creates a new 16-bit RGB color
func NewRGB16[CS ColorSpace](r, g, b basics.Int16u) RGB16[CS] {
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
