// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

import (
	"agg_go/internal/basics"
)

// Constants for RGBA16 fixed-point arithmetic
const (
	RGBA16BaseMask  = 65535
	RGBA16BaseShift = 16
	RGBA16BaseMSB   = 1 << (RGBA16BaseShift - 1)
)

// RGBA16Multiply performs fixed-point multiplication for 16-bit values
func RGBA16Multiply(a, b basics.Int16u) basics.Int16u {
	t := uint32(a)*uint32(b) + RGBA16BaseMSB
	return basics.Int16u(((t >> RGBA16BaseShift) + t) >> RGBA16BaseShift)
}

// RGBA16Lerp performs linear interpolation between two 16-bit values
func RGBA16Lerp(p, q, a basics.Int16u) basics.Int16u {
	diff := int64(q) - int64(p)
	result := int64(p) + (diff*int64(a)+32767)/65535
	if result < 0 {
		return 0
	}
	if result > 65535 {
		return 65535
	}
	return basics.Int16u(result)
}

// RGBA16Prelerp performs premultiplied linear interpolation for 16-bit values
func RGBA16Prelerp(p, q, a basics.Int16u) basics.Int16u {
	diff := int64(q) - int64(p)
	result := int64(p) + (diff*int64(a)+32767)/65535
	if result < 0 {
		return 0
	}
	if result > 65535 {
		return 65535
	}
	return basics.Int16u(result)
}

// RGBA16MultCover multiplies a 16-bit color component by coverage
func RGBA16MultCover(c, cover basics.Int16u) basics.Int16u {
	return RGBA16Multiply(c, cover)
}

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

// Common color types
type (
	RGBA8Linear = RGBA8[Linear]
	RGBA8SRGB   = RGBA8[SRGB]
	SRGBA8      = RGBA8[SRGB] // Alias for backwards compatibility
)

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
		R: basics.Int16u(c.R*65535 + 0.5),
		G: basics.Int16u(c.G*65535 + 0.5),
		B: basics.Int16u(c.B*65535 + 0.5),
		A: basics.Int16u(c.A*65535 + 0.5),
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
			c.R = basics.Int16u((uint32(c.R)*65535 + uint32(c.A)/2) / uint32(c.A))
			c.G = basics.Int16u((uint32(c.G)*65535 + uint32(c.A)/2) / uint32(c.A))
			c.B = basics.Int16u((uint32(c.B)*65535 + uint32(c.A)/2) / uint32(c.A))
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

// Opacity sets the alpha channel (0.0 to 1.0)
func (c *RGBA16[CS]) Opacity(a float64) *RGBA16[CS] {
	if a < 0 {
		c.A = 0
	} else if a > 1 {
		c.A = 65535
	} else {
		c.A = basics.Int16u(a*65535 + 0.5)
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

// Common 16-bit color types
type (
	RGBA16Linear = RGBA16[Linear]
	RGBA16SRGB   = RGBA16[SRGB]
)
