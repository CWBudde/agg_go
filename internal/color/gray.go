package color

import (
	"agg_go/internal/basics"
)

// Gray8 represents an 8-bit grayscale color with colorspace
type Gray8[CS any] struct {
	V basics.Int8u // Value (luminance)
	A basics.Int8u // Alpha channel
}

// Gray8 constants
const (
	Gray8BaseShift = 8
	Gray8BaseScale = 1 << Gray8BaseShift
	Gray8BaseMask  = Gray8BaseScale - 1
	Gray8BaseMSB   = 1 << (Gray8BaseShift - 1)
)

// NewGray8 creates a new 8-bit grayscale color
func NewGray8[CS any](v basics.Int8u) Gray8[CS] {
	return Gray8[CS]{V: v, A: Gray8BaseMask}
}

// NewGray8WithAlpha creates a new 8-bit grayscale color with alpha
func NewGray8WithAlpha[CS any](v, a basics.Int8u) Gray8[CS] {
	return Gray8[CS]{V: v, A: a}
}

// Luminance calculates luminance from RGBA using ITU-R BT.709
func LuminanceFromRGBA(c RGBA) float64 {
	return 0.299*c.R + 0.587*c.G + 0.114*c.B
}

// LuminanceFromRGBA8 calculates luminance from 8-bit RGBA (optimized)
func LuminanceFromRGBA8[CS any](c RGBA8[CS]) basics.Int8u {
	// Using integer arithmetic: 0.299*77 + 0.587*150 + 0.114*29 ≈ 256
	return basics.Int8u((uint32(c.R)*77 + uint32(c.G)*150 + uint32(c.B)*29) >> 8)
}

// ConvertFromRGBA converts from floating-point RGBA to grayscale
func ConvertGray8FromRGBA[CS any](c RGBA) Gray8[CS] {
	lum := LuminanceFromRGBA(c)
	return Gray8[CS]{
		V: basics.Int8u(lum*255 + 0.5),
		A: basics.Int8u(c.A*255 + 0.5),
	}
}

// ConvertFromRGBA8 converts from 8-bit RGBA to grayscale
func ConvertGray8FromRGBA8[CS any](c RGBA8[CS]) Gray8[CS] {
	return Gray8[CS]{
		V: LuminanceFromRGBA8(c),
		A: c.A,
	}
}

// ConvertToRGBA converts grayscale to floating-point RGBA
func (g Gray8[CS]) ConvertToRGBA() RGBA {
	v := float64(g.V) / 255.0
	a := float64(g.A) / 255.0
	return RGBA{R: v, G: v, B: v, A: a}
}

// ConvertToRGBA8 converts grayscale to 8-bit RGBA
func (g Gray8[CS]) ConvertToRGBA8() RGBA8[CS] {
	return RGBA8[CS]{R: g.V, G: g.V, B: g.V, A: g.A}
}

// Common grayscale types
type (
	Gray8Linear = Gray8[Linear]
	Gray8SRGB   = Gray8[SRGB]
	SGray8      = Gray8[SRGB] // Alias for backwards compatibility
)

// Clear sets the color to transparent black
func (g *Gray8[CS]) Clear() {
	g.V, g.A = 0, 0
}

// Transparent sets the color to transparent with the same luminance
func (g *Gray8[CS]) Transparent() {
	g.A = 0
}

// Opacity sets the alpha channel (0.0 to 1.0)
func (g *Gray8[CS]) Opacity(a float64) {
	if a < 0 {
		g.A = 0
	} else if a > 1 {
		g.A = Gray8BaseMask
	} else {
		g.A = basics.Int8u(a*float64(Gray8BaseMask) + 0.5)
	}
}

// GetOpacity returns the alpha as a float64 (0.0 to 1.0)
func (g Gray8[CS]) GetOpacity() float64 {
	return float64(g.A) / float64(Gray8BaseMask)
}

// IsTransparent returns true if the color is fully transparent
func (g Gray8[CS]) IsTransparent() bool {
	return g.A == 0
}

// IsOpaque returns true if the color is fully opaque
func (g Gray8[CS]) IsOpaque() bool {
	return g.A == Gray8BaseMask
}

// Multiply performs fixed-point multiplication (exact over int8u)
func Gray8Multiply(a, b basics.Int8u) basics.Int8u {
	t := uint32(a)*uint32(b) + Gray8BaseMSB
	return basics.Int8u(((t >> Gray8BaseShift) + t) >> Gray8BaseShift)
}

// Lerp performs linear interpolation
func Gray8Lerp(p, q, a basics.Int8u) basics.Int8u {
	var t int32
	if p > q {
		t = int32(q-p)*int32(a) + Gray8BaseMSB - 1
	} else {
		t = int32(q-p)*int32(a) + Gray8BaseMSB
	}
	return basics.Int8u(int32(p) + (((t >> Gray8BaseShift) + t) >> Gray8BaseShift))
}

// Prelerp performs premultiplied linear interpolation
func Gray8Prelerp(p, q, a basics.Int8u) basics.Int8u {
	return p + q - Gray8Multiply(p, a)
}

// Premultiply premultiplies the color by alpha
func (g *Gray8[CS]) Premultiply() {
	if g.A < Gray8BaseMask {
		if g.A == 0 {
			g.V = 0
		} else {
			g.V = Gray8Multiply(g.V, g.A)
		}
	}
}

// Demultiply demultiplies the color by alpha
func (g *Gray8[CS]) Demultiply() {
	if g.A < Gray8BaseMask {
		if g.A == 0 {
			g.V = 0
		} else {
			v := (uint32(g.V) * Gray8BaseMask) / uint32(g.A)
			if v > Gray8BaseMask {
				g.V = Gray8BaseMask
			} else {
				g.V = basics.Int8u(v)
			}
		}
	}
}

// Gradient performs linear interpolation between two colors
func (g Gray8[CS]) Gradient(c2 Gray8[CS], k float64) Gray8[CS] {
	ik := basics.Int8u(k*float64(Gray8BaseScale) + 0.5)
	return Gray8[CS]{
		V: Gray8Lerp(g.V, c2.V, ik),
		A: Gray8Lerp(g.A, c2.A, ik),
	}
}

// Add blends another color using normal blending
func (g *Gray8[CS]) Add(c Gray8[CS], cover basics.Int8u) {
	if cover == 255 { // cover_mask
		if c.A == Gray8BaseMask {
			*g = c
			return
		} else {
			cv := uint32(g.V) + uint32(c.V)
			ca := uint32(g.A) + uint32(c.A)
			if cv > Gray8BaseMask {
				g.V = Gray8BaseMask
			} else {
				g.V = basics.Int8u(cv)
			}
			if ca > Gray8BaseMask {
				g.A = Gray8BaseMask
			} else {
				g.A = basics.Int8u(ca)
			}
		}
	} else {
		cv := uint32(g.V) + uint32(Gray8Multiply(c.V, cover))
		ca := uint32(g.A) + uint32(Gray8Multiply(c.A, cover))
		if cv > Gray8BaseMask {
			g.V = Gray8BaseMask
		} else {
			g.V = basics.Int8u(cv)
		}
		if ca > Gray8BaseMask {
			g.A = Gray8BaseMask
		} else {
			g.A = basics.Int8u(ca)
		}
	}
}

// EmptyValue returns the empty (transparent) value
func Gray8EmptyValue() basics.Int8u {
	return 0
}

// FullValue returns the full (opaque) value
func Gray8FullValue() basics.Int8u {
	return Gray8BaseMask
}

// Gray16 represents a 16-bit grayscale color
type Gray16[CS any] struct {
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
func NewGray16[CS any](v basics.Int16u) Gray16[CS] {
	return Gray16[CS]{V: v, A: Gray16BaseMask}
}

// NewGray16WithAlpha creates a new 16-bit grayscale color with alpha
func NewGray16WithAlpha[CS any](v, a basics.Int16u) Gray16[CS] {
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
func ConvertGray16FromRGBA[CS any](c RGBA) Gray16[CS] {
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

// Gray32 represents a 32-bit floating-point grayscale color
type Gray32[CS any] struct {
	V float32
	A float32
}

// NewGray32 creates a new 32-bit grayscale color
func NewGray32[CS any](v float32) Gray32[CS] {
	return Gray32[CS]{V: v, A: 1.0}
}

// NewGray32WithAlpha creates a new 32-bit grayscale color with alpha
func NewGray32WithAlpha[CS any](v, a float32) Gray32[CS] {
	return Gray32[CS]{V: v, A: a}
}

// ConvertToRGBA converts to floating-point RGBA
func (g Gray32[CS]) ConvertToRGBA() RGBA {
	v := float64(g.V)
	a := float64(g.A)
	return RGBA{R: v, G: v, B: v, A: a}
}

// Clear sets the color to transparent black
func (g *Gray32[CS]) Clear() {
	g.V, g.A = 0.0, 0.0
}

// Transparent sets the color to transparent with the same luminance
func (g *Gray32[CS]) Transparent() {
	g.A = 0.0
}

// IsTransparent returns true if the color is fully transparent
func (g Gray32[CS]) IsTransparent() bool {
	return g.A == 0.0
}

// IsOpaque returns true if the color is fully opaque
func (g Gray32[CS]) IsOpaque() bool {
	return g.A == 1.0
}

// ConvertFromRGBA converts from floating-point RGBA
func ConvertGray32FromRGBA[CS any](c RGBA) Gray32[CS] {
	lum := LuminanceFromRGBA(c)
	return Gray32[CS]{
		V: float32(lum),
		A: float32(c.A),
	}
}

// Common 32-bit grayscale types
type (
	Gray32Linear = Gray32[Linear]
	Gray32SRGB   = Gray32[SRGB]
)
