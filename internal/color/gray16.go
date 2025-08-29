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

func NewGray16Linear(v basics.Int16u) Gray16[Linear] {
	return NewGray16[Linear](v)
}

func NewGray16LinearA(v, a basics.Int16u) Gray16[Linear] {
	return NewGray16WithAlpha[Linear](v, a)
}

func NewGray16SRGB(v basics.Int16u) Gray16[SRGB] {
	return NewGray16[SRGB](v)
}

func NewGray16SRGBA(v, a basics.Int16u) Gray16[SRGB] {
	return NewGray16WithAlpha[SRGB](v, a)
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

func (g *Gray16[CS]) Opacity(a float32) {
	g.A = basics.Int16u(clamp01f32(a) * 65535.0)
}

func (g Gray16[CS]) GetOpacity() float32 {
	return float32(g.A) / 65535.0
}

// Premultiply premultiplies the color by alpha
func (g *Gray16[CS]) Premultiply() {
	if g.A < Gray16BaseMask {
		if g.A == 0 {
			g.V = 0
		} else {
			g.V = Gray16Multiply(g.V, g.A)
		}
	}
}

// Demultiply demultiplies the color by alpha
func (g *Gray16[CS]) Demultiply() {
	if g.A < Gray16BaseMask {
		if g.A == 0 {
			g.V = 0
		} else {
			v := (uint32(g.V) * Gray16BaseMask) / uint32(g.A)
			if v > Gray16BaseMask {
				g.V = Gray16BaseMask
			} else {
				g.V = basics.Int16u(v)
			}
		}
	}
}

// Gray16 constants
const Gray16BaseMSB = 1 << (Gray16BaseShift - 1)

// Multiply performs fixed-point multiplication (exact over int16u)
func Gray16Multiply(a, b basics.Int16u) basics.Int16u {
	t := uint32(a)*uint32(b) + Gray16BaseMSB
	return basics.Int16u(((t >> Gray16BaseShift) + t) >> Gray16BaseShift)
}

// Lerp performs linear interpolation
func Gray16Lerp(p, q, a basics.Int16u) basics.Int16u {
	var t int64
	if p > q {
		t = int64(q-p)*int64(a) + Gray16BaseMSB - 1
	} else {
		t = int64(q-p)*int64(a) + Gray16BaseMSB
	}
	return basics.Int16u(int64(p) + (((t >> Gray16BaseShift) + t) >> Gray16BaseShift))
}

// Prelerp performs premultiplied linear interpolation
func Gray16Prelerp(p, q, a basics.Int16u) basics.Int16u {
	return p + q - Gray16Multiply(p, a)
}

// Gradient performs linear interpolation between two colors
func (g Gray16[CS]) Gradient(c2 Gray16[CS], k float64) Gray16[CS] {
	ik := basics.Int16u(k*float64(Gray16BaseScale) + 0.5)
	return Gray16[CS]{
		V: Gray16Lerp(g.V, c2.V, ik),
		A: Gray16Lerp(g.A, c2.A, ik),
	}
}

// Add blends another color using normal blending
func (g *Gray16[CS]) Add(c Gray16[CS], cover basics.Int8u) {
	if cover == 255 { // cover_mask
		if c.A == Gray16BaseMask {
			*g = c
			return
		} else {
			cv := uint32(g.V) + uint32(c.V)
			ca := uint32(g.A) + uint32(c.A)
			if cv > Gray16BaseMask {
				g.V = Gray16BaseMask
			} else {
				g.V = basics.Int16u(cv)
			}
			if ca > Gray16BaseMask {
				g.A = Gray16BaseMask
			} else {
				g.A = basics.Int16u(ca)
			}
		}
	} else {
		// Scale 8-bit cover to 16-bit
		cover16 := basics.Int16u(cover)
		cover16 = (cover16 << 8) | cover16
		cv := uint32(g.V) + uint32(Gray16Multiply(c.V, cover16))
		ca := uint32(g.A) + uint32(Gray16Multiply(c.A, cover16))
		if cv > Gray16BaseMask {
			g.V = Gray16BaseMask
		} else {
			g.V = basics.Int16u(cv)
		}
		if ca > Gray16BaseMask {
			g.A = Gray16BaseMask
		} else {
			g.A = basics.Int16u(ca)
		}
	}
}

// AddWithCover adds another Gray16 color with coverage for compound rendering compatibility
func (g *Gray16[CS]) AddWithCover(c Gray16[CS], cover basics.Int8u) {
	g.Add(c, cover)
}

// EmptyValue returns the empty (transparent) value
func Gray16EmptyValue() basics.Int16u {
	return 0
}

// FullValue returns the full (opaque) value
func Gray16FullValue() basics.Int16u {
	return Gray16BaseMask
}
