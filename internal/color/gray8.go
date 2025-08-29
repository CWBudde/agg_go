package color

import (
	"agg_go/internal/basics"
)

// Gray8 represents an 8-bit grayscale color with colorspace
type Gray8[CS ColorSpace] struct {
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
func NewGray8[CS ColorSpace](v basics.Int8u) Gray8[CS] {
	return Gray8[CS]{V: v, A: Gray8BaseMask}
}

// NewGray8WithAlpha creates a new 8-bit grayscale color with alpha
func NewGray8WithAlpha[CS ColorSpace](v, a basics.Int8u) Gray8[CS] {
	return Gray8[CS]{V: v, A: a}
}

// ConvertFromRGBA converts from floating-point RGBA to grayscale
func ConvertGray8FromRGBA[CS ColorSpace](c RGBA) Gray8[CS] {
	lum := LuminanceFromRGBA(c)
	return Gray8[CS]{
		V: basics.Int8u(lum*255 + 0.5),
		A: basics.Int8u(c.A*255 + 0.5),
	}
}

// ConvertFromRGBA8 converts from 8-bit RGBA to grayscale
func ConvertGray8FromRGBA8[CS ColorSpace](c RGBA8[CS]) Gray8[CS] {
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

// AddWithCover adds another Gray8 color with coverage for compound rendering compatibility
func (g *Gray8[CS]) AddWithCover(c Gray8[CS], cover basics.Int8u) {
	g.Add(c, cover)
}

// EmptyValue returns the empty (transparent) value
func Gray8EmptyValue() basics.Int8u {
	return 0
}

// FullValue returns the full (opaque) value
func Gray8FullValue() basics.Int8u {
	return Gray8BaseMask
}
