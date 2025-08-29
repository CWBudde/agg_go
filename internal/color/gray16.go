package color

import (
	"agg_go/internal/basics"
)

// Gray16 represents a 16-bit grayscale color
type Gray16[CS ColorSpace] struct {
	V basics.Int16u
	A basics.Int16u
}

// Gray16 constants
const (
	Gray16BaseShift = 16
	Gray16BaseScale = 1 << Gray16BaseShift
	Gray16BaseMask  = Gray16BaseScale - 1
)

// NewGray16 creates a new 16-bit grayscale color
func NewGray16[CS ColorSpace](v basics.Int16u) Gray16[CS] {
	return Gray16[CS]{V: v, A: Gray16BaseMask}
}

// NewGray16WithAlpha creates a new 16-bit grayscale color with alpha
func NewGray16WithAlpha[CS ColorSpace](v, a basics.Int16u) Gray16[CS] {
	return Gray16[CS]{V: v, A: a}
}

// ConvertToRGBA converts to floating-point RGBA
func (g Gray16[CS]) ConvertToRGBA() RGBA {
	v := float64(g.V) / 65535.0
	a := float64(g.A) / 65535.0
	return RGBA{R: v, G: v, B: v, A: a}
}

// ConvertToRGBA16 converts to 16-bit RGBA
func (g Gray16[CS]) ConvertToRGBA16() RGBA16[CS] {
	return RGBA16[CS]{R: g.V, G: g.V, B: g.V, A: g.A}
}

// ConvertFromRGBA converts from floating-point RGBA
func ConvertGray16FromRGBA[CS ColorSpace](c RGBA) Gray16[CS] {
	lum := LuminanceFromRGBA(c)
	return Gray16[CS]{
		V: basics.Int16u(lum*65535 + 0.5),
		A: basics.Int16u(c.A*65535 + 0.5),
	}
}

// Common 16-bit grayscale types
type (
	Gray16Linear = Gray16[Linear]
	Gray16SRGB   = Gray16[SRGB]
)

// Clear sets the color to transparent black
func (g *Gray16[CS]) Clear() {
	g.V, g.A = 0, 0
}

// Transparent sets the color to transparent with the same luminance
func (g *Gray16[CS]) Transparent() {
	g.A = 0
}

// IsTransparent returns true if the color is fully transparent
func (g Gray16[CS]) IsTransparent() bool {
	return g.A == 0
}

// IsOpaque returns true if the color is fully opaque
func (g Gray16[CS]) IsOpaque() bool {
	return g.A == Gray16BaseMask
}
