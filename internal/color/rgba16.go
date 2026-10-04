// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

import (
	"github.com/cwbudde/agg_go/internal/basics"
)

// Constants for RGBA16 fixed-point arithmetic
const (
	RGBA16BaseMask  = 65535
	RGBA16BaseShift = 16
	RGBA16BaseMSB   = 1 << (RGBA16BaseShift - 1)
)

// RGBA16Multiply performs fixed-point multiplication for 16-bit values.
// Matches C++ rgba16::multiply (exact over int16u).
func RGBA16Multiply(a, b basics.Int16u) basics.Int16u {
	t := uint32(a)*uint32(b) + RGBA16BaseMSB
	return basics.Int16u(((t >> RGBA16BaseShift) + t) >> RGBA16BaseShift)
}

// RGBA16Lerp interpolates p to q by a. Matches C++ rgba16::lerp:
//
//	int t = (q - p) * a + base_MSB - (p > q);
//	return value_type(p + (((t >> base_shift) + t) >> base_shift));
//
// The product is evaluated in 32-bit int exactly as in C++, so it wraps for
// |(q-p)*a| >= 2^31 (e.g. lerp(0, 65535, 65535) == 65534). Every C++ build
// (-O0 and -O2, x86-64 and arm64) produces this result; it is kept for parity.
func RGBA16Lerp(p, q, a basics.Int16u) basics.Int16u {
	var gt int32
	if p > q {
		gt = 1
	}
	t := (int32(q)-int32(p))*int32(a) + RGBA16BaseMSB - gt
	return basics.Int16u(int32(p) + (((t >> RGBA16BaseShift) + t) >> RGBA16BaseShift))
}

// RGBA16Prelerp interpolates p to q by a, assuming q is premultiplied by a.
// Matches C++ rgba16::prelerp: p + q - multiply(p, a) (modulo 2^16).
func RGBA16Prelerp(p, q, a basics.Int16u) basics.Int16u {
	return p + q - RGBA16Multiply(p, a)
}

// RGBA16MultCover multiplies a 16-bit color component by a coverage value
// that has already been expanded to 16 bits ((cover8 << 8) | cover8).
// With an expanded cover this equals C++ rgba16::mult_cover(c, cover8),
// which is multiply(c, (cover8 << 8) | cover8).
func RGBA16MultCover(c, cover basics.Int16u) basics.Int16u {
	return RGBA16Multiply(c, cover)
}

// RGBA16ScaleCover scales an 8-bit coverage by a 16-bit value.
// Matches C++ rgba16::scale_cover: multiply((a << 8) | a, b) >> 8.
func RGBA16ScaleCover(cover basics.Int8u, b basics.Int16u) basics.Int8u {
	c16 := basics.Int16u(cover)<<8 | basics.Int16u(cover)
	return basics.Int8u(RGBA16Multiply(c16, b) >> 8)
}

// RGBA16Demultiply divides a by b in fixed point (static C++
// rgba16::demultiply). C++ evaluates a*base_mask in signed int, which
// overflows for a > 32768 (undefined behaviour; -O0 and -O2 builds disagree).
// Go uses the overflow-free unsigned reading of the same formula, which is
// what optimised C++ builds produce. See docs/AGG_DELTAS.md.
func RGBA16Demultiply(a, b basics.Int16u) basics.Int16u {
	switch {
	case a == 0 || b == 0:
		return 0
	case a >= b:
		return RGBA16BaseMask
	default:
		return basics.Int16u((uint32(a)*RGBA16BaseMask + uint32(b>>1)) / uint32(b))
	}
}

// RGBA16FromDouble converts a [0,1] value to 16-bit fixed point.
// Matches C++ rgba16::from_double: value_type(uround(a * base_mask)).
func RGBA16FromDouble(a float64) basics.Int16u {
	return basics.Int16u(basics.URound(a * RGBA16BaseMask))
}

// RGBA16Invert returns base_mask - x (C++ rgba16::invert).
func RGBA16Invert(x basics.Int16u) basics.Int16u {
	return RGBA16BaseMask - x
}

// RGBA16EmptyValue returns the empty value (C++ rgba16::empty_value).
func RGBA16EmptyValue() basics.Int16u { return 0 }

// RGBA16FullValue returns the full value (C++ rgba16::full_value).
func RGBA16FullValue() basics.Int16u { return RGBA16BaseMask }

// Helper functions for min operations
func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// RGBA16 represents a 16-bit RGBA color
type RGBA16[CS Space] struct {
	R, G, B, A basics.Int16u
}

// NewRGBA16 creates a new 16-bit RGBA color
func NewRGBA16[CS Space](r, g, b, a basics.Int16u) RGBA16[CS] {
	return RGBA16[CS]{R: r, G: g, B: b, A: a}
}

// ConvertToRGBA converts to floating-point RGBA
func (c RGBA16[CS]) ConvertToRGBA() RGBA {
	const scale = 1.0 / 65535.0
	return RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: float64(c.A) * scale,
	}
}

// ConvertFromRGBA converts from floating-point RGBA
func ConvertFromRGBA16[CS Space](c RGBA) RGBA16[CS] {
	return RGBA16[CS]{
		R: RGBA16FromDouble(c.R),
		G: RGBA16FromDouble(c.G),
		B: RGBA16FromDouble(c.B),
		A: RGBA16FromDouble(c.A),
	}
}

// Premultiply premultiplies the color by alpha
func (c *RGBA16[CS]) Premultiply() *RGBA16[CS] {
	if c.A < 65535 {
		if c.A == 0 {
			c.R, c.G, c.B = 0, 0, 0
		} else {
			c.R = RGBA16Multiply(c.R, c.A)
			c.G = RGBA16Multiply(c.G, c.A)
			c.B = RGBA16Multiply(c.B, c.A)
		}
	}
	return c
}

// Demultiply demultiplies the color by alpha
func (c *RGBA16[CS]) Demultiply() *RGBA16[CS] {
	if c.A < 65535 {
		if c.A == 0 {
			c.R, c.G, c.B = 0, 0, 0
		} else {
			// C++ rgba16::demultiply: truncating division, clamped to base_mask.
			a := uint32(c.A)
			c.R = basics.Int16u(minUint32(uint32(c.R)*RGBA16BaseMask/a, RGBA16BaseMask))
			c.G = basics.Int16u(minUint32(uint32(c.G)*RGBA16BaseMask/a, RGBA16BaseMask))
			c.B = basics.Int16u(minUint32(uint32(c.B)*RGBA16BaseMask/a, RGBA16BaseMask))
		}
	}
	return c
}

// Gradient performs linear interpolation between two 16-bit colors
func (c RGBA16[CS]) Gradient(c2 RGBA16[CS], k basics.Int16u) RGBA16[CS] {
	return RGBA16[CS]{
		R: RGBA16Lerp(c.R, c2.R, k),
		G: RGBA16Lerp(c.G, c2.G, k),
		B: RGBA16Lerp(c.B, c2.B, k),
		A: RGBA16Lerp(c.A, c2.A, k),
	}
}

// Clear sets the color to transparent black
func (c *RGBA16[CS]) Clear() *RGBA16[CS] {
	c.R, c.G, c.B, c.A = 0, 0, 0, 0
	return c
}

// Transparent sets the color to transparent with the same RGB
func (c *RGBA16[CS]) Transparent() *RGBA16[CS] {
	c.A = 0
	return c
}

// IsTransparent returns true if the color is fully transparent
func (c RGBA16[CS]) IsTransparent() bool {
	return c.A == 0
}

// IsOpaque returns true if the color is fully opaque
func (c RGBA16[CS]) IsOpaque() bool {
	return c.A == 65535
}

// Opacity sets the alpha channel (0.0 to 1.0). In-range values use
// from_double rounding like C++. Out-of-range values saturate to 0 / 65535;
// C++ rgba16::opacity has no else-branches and stores uround(a*65535) even
// for a < 0 or a > 1 (see docs/AGG_DELTAS.md).
func (c *RGBA16[CS]) Opacity(a float64) *RGBA16[CS] {
	switch {
	case a < 0:
		c.A = 0
	case a > 1:
		c.A = 65535
	default:
		c.A = RGBA16FromDouble(a)
	}
	return c
}

// GetOpacity returns the alpha as a float64 (0.0 to 1.0)
func (c RGBA16[CS]) GetOpacity() float64 {
	return float64(c.A) / 65535.0
}

// Add adds another RGBA16 color
func (c RGBA16[CS]) Add(c2 RGBA16[CS]) RGBA16[CS] {
	return RGBA16[CS]{
		R: basics.Int16u(minUint32(uint32(c.R)+uint32(c2.R), 65535)),
		G: basics.Int16u(minUint32(uint32(c.G)+uint32(c2.G), 65535)),
		B: basics.Int16u(minUint32(uint32(c.B)+uint32(c2.B), 65535)),
		A: basics.Int16u(minUint32(uint32(c.A)+uint32(c2.A), 65535)),
	}
}

// AddWithCover adds another RGBA16 color with coverage, matching C++ AGG's add(color, cover) method
func (c *RGBA16[CS]) AddWithCover(c2 RGBA16[CS], cover basics.Int8u) {
	cover16 := basics.Int16u(cover)<<8 | basics.Int16u(cover) // Convert 8-bit cover to 16-bit
	if cover == basics.CoverMask {
		if c2.A == 65535 { // base_mask for 16-bit
			*c = c2
			return
		} else {
			cr := uint32(c.R) + uint32(c2.R)
			cg := uint32(c.G) + uint32(c2.G)
			cb := uint32(c.B) + uint32(c2.B)
			ca := uint32(c.A) + uint32(c2.A)
			c.R = basics.Int16u(minUint32(cr, 65535))
			c.G = basics.Int16u(minUint32(cg, 65535))
			c.B = basics.Int16u(minUint32(cb, 65535))
			c.A = basics.Int16u(minUint32(ca, 65535))
		}
	} else {
		cr := uint32(c.R) + uint32(RGBA16MultCover(c2.R, cover16))
		cg := uint32(c.G) + uint32(RGBA16MultCover(c2.G, cover16))
		cb := uint32(c.B) + uint32(RGBA16MultCover(c2.B, cover16))
		ca := uint32(c.A) + uint32(RGBA16MultCover(c2.A, cover16))
		c.R = basics.Int16u(minUint32(cr, 65535))
		c.G = basics.Int16u(minUint32(cg, 65535))
		c.B = basics.Int16u(minUint32(cb, 65535))
		c.A = basics.Int16u(minUint32(ca, 65535))
	}
}

func ApplyGammaDir16[CS Space, LUT lut16Like](px *RGBA16[CS], lut LUT) {
	px.R = lut.Dir(basics.Int8u(px.R >> 8))
	px.G = lut.Dir(basics.Int8u(px.G >> 8))
	px.B = lut.Dir(basics.Int8u(px.B >> 8))
}

func ApplyGammaInv16[CS Space, LUT lut16Like](px *RGBA16[CS], lut LUT) {
	r8 := lut.Inv(px.R)
	g8 := lut.Inv(px.G)
	b8 := lut.Inv(px.B)
	px.R = basics.Int16u(r8)<<8 | basics.Int16u(r8)
	px.G = basics.Int16u(g8)<<8 | basics.Int16u(g8)
	px.B = basics.Int16u(b8)<<8 | basics.Int16u(b8)
}

func ApplyGammaDir16Using8[CS Space, LUT lut8Like](px *RGBA16[CS], lut LUT) {
	r8 := basics.Int8u(px.R >> 8)
	g8 := basics.Int8u(px.G >> 8)
	b8 := basics.Int8u(px.B >> 8)
	r8 = lut.Dir(r8)
	g8 = lut.Dir(g8)
	b8 = lut.Dir(b8)
	px.R = basics.Int16u(r8)<<8 | basics.Int16u(r8)
	px.G = basics.Int16u(g8)<<8 | basics.Int16u(g8)
	px.B = basics.Int16u(b8)<<8 | basics.Int16u(b8)
}

func ApplyGammaInv16Using8[CS Space, LUT lut8Like](px *RGBA16[CS], lut LUT) {
	r8 := basics.Int8u(px.R >> 8)
	g8 := basics.Int8u(px.G >> 8)
	b8 := basics.Int8u(px.B >> 8)
	r8 = lut.Inv(r8)
	g8 = lut.Inv(g8)
	b8 = lut.Inv(b8)
	px.R = basics.Int16u(r8)<<8 | basics.Int16u(r8)
	px.G = basics.Int16u(g8)<<8 | basics.Int16u(g8)
	px.B = basics.Int16u(b8)<<8 | basics.Int16u(b8)
}

// Helper for method receivers:
func (c *RGBA16[CS]) ApplyGammaDir(lut lut16Like) { ApplyGammaDir16(c, lut) }
func (c *RGBA16[CS]) ApplyGammaInv(lut lut16Like) { ApplyGammaInv16(c, lut) }

func (c *RGBA16[CS]) ApplyGammaDirUsing8(lut lut8Like) { ApplyGammaDir16Using8(c, lut) }
func (c *RGBA16[CS]) ApplyGammaInvUsing8(lut lut8Like) { ApplyGammaInv16Using8(c, lut) }

// Common 16-bit color types
type (
	RGBA16Linear = RGBA16[Linear]
	RGBA16SRGB   = RGBA16[SRGB]
)
