// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

import (
	"agg_go/internal/basics"
)

// RGBA8 represents an 8-bit RGBA color with colorspace
type RGBA8[CS ColorSpace] struct {
	R, G, B, A basics.Int8u
}

// NewRGBA8 creates a new 8-bit RGBA color
func NewRGBA8[CS ColorSpace](r, g, b, a basics.Int8u) RGBA8[CS] {
	return RGBA8[CS]{R: r, G: g, B: b, A: a}
}

// Convert converts between colorspaces for RGBA8
// This method only makes sense when converting between different colorspaces
func (c RGBA8[CS]) Convert() RGBA8[CS] {
	// Same colorspace, no conversion needed
	return c
}

// ConvertToLinear converts RGBA8[SRGB] to RGBA8[Linear]
func ConvertToLinear(c RGBA8[SRGB]) RGBA8[Linear] {
	return ConvertRGBA8SRGBToLinear(c)
}

// ConvertToSRGB converts RGBA8[Linear] to RGBA8[SRGB]
func ConvertToSRGBFromLinear(c RGBA8[Linear]) RGBA8[SRGB] {
	return ConvertRGBA8LinearToSRGB(c)
}

// ConvertToRGBA converts to floating-point RGBA
func (c RGBA8[CS]) ConvertToRGBA() RGBA {
	const scale = 1.0 / 255.0
	return RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: float64(c.A) * scale,
	}
}

// ConvertFromRGBA converts from floating-point RGBA
func ConvertFromRGBA[CS ColorSpace](c RGBA) RGBA8[CS] {
	return RGBA8[CS]{
		R: basics.Int8u(c.R*255 + 0.5),
		G: basics.Int8u(c.G*255 + 0.5),
		B: basics.Int8u(c.B*255 + 0.5),
		A: basics.Int8u(c.A*255 + 0.5),
	}
}

// Premultiply premultiplies the color by alpha
func (c *RGBA8[CS]) Premultiply() {
	if c.A < RGBA8BaseMask {
		if c.A == 0 {
			c.R, c.G, c.B = 0, 0, 0
		} else {
			c.R = RGBA8Multiply(c.R, c.A)
			c.G = RGBA8Multiply(c.G, c.A)
			c.B = RGBA8Multiply(c.B, c.A)
		}
	}
}

// Demultiply demultiplies the color by alpha
func (c *RGBA8[CS]) Demultiply() {
	if c.A < RGBA8BaseMask {
		if c.A == 0 {
			c.R, c.G, c.B = 0, 0, 0
		} else {
			// Use accurate division for demultiplication
			c.R = basics.Int8u((uint32(c.R)*RGBA8BaseMask + uint32(c.A)/2) / uint32(c.A))
			c.G = basics.Int8u((uint32(c.G)*RGBA8BaseMask + uint32(c.A)/2) / uint32(c.A))
			c.B = basics.Int8u((uint32(c.B)*RGBA8BaseMask + uint32(c.A)/2) / uint32(c.A))
		}
	}
}

// Gradient performs linear interpolation between two 8-bit colors
func (c RGBA8[CS]) Gradient(c2 RGBA8[CS], k basics.Int8u) RGBA8[CS] {
	return RGBA8[CS]{
		R: RGBA8Lerp(c.R, c2.R, k),
		G: RGBA8Lerp(c.G, c2.G, k),
		B: RGBA8Lerp(c.B, c2.B, k),
		A: RGBA8Lerp(c.A, c2.A, k),
	}
}

// Clear sets the color to transparent black
func (c *RGBA8[CS]) Clear() {
	c.R, c.G, c.B, c.A = 0, 0, 0, 0
}

// Transparent sets the color to transparent with the same RGB
func (c *RGBA8[CS]) Transparent() {
	c.A = 0
}

// IsTransparent returns true if the color is fully transparent
func (c RGBA8[CS]) IsTransparent() bool {
	return c.A == 0
}

// IsOpaque returns true if the color is fully opaque
func (c RGBA8[CS]) IsOpaque() bool {
	return c.A == 255
}

// Opacity sets the alpha channel (0.0 to 1.0)
func (c *RGBA8[CS]) Opacity(a float64) {
	if a < 0 {
		c.A = 0
	} else if a > 1 {
		c.A = 255
	} else {
		c.A = basics.Int8u(a*255 + 0.5)
	}
}

// GetOpacity returns the alpha as a float64 (0.0 to 1.0)
func (c RGBA8[CS]) GetOpacity() float64 {
	return float64(c.A) / 255.0
}

// Add adds another RGBA8 color
func (c RGBA8[CS]) Add(c2 RGBA8[CS]) RGBA8[CS] {
	return RGBA8[CS]{
		R: basics.Int8u(minUint32(uint32(c.R)+uint32(c2.R), 255)),
		G: basics.Int8u(minUint32(uint32(c.G)+uint32(c2.G), 255)),
		B: basics.Int8u(minUint32(uint32(c.B)+uint32(c2.B), 255)),
		A: basics.Int8u(minUint32(uint32(c.A)+uint32(c2.A), 255)),
	}
}

// AddWithCover adds another RGBA8 color with coverage, matching C++ AGG's add(color, cover) method
func (c *RGBA8[CS]) AddWithCover(c2 RGBA8[CS], cover basics.Int8u) {
	if cover == basics.CoverMask {
		if c2.A == 255 { // base_mask
			*c = c2
			return
		} else {
			cr := uint32(c.R) + uint32(c2.R)
			cg := uint32(c.G) + uint32(c2.G)
			cb := uint32(c.B) + uint32(c2.B)
			ca := uint32(c.A) + uint32(c2.A)
			c.R = basics.Int8u(minUint32(cr, 255))
			c.G = basics.Int8u(minUint32(cg, 255))
			c.B = basics.Int8u(minUint32(cb, 255))
			c.A = basics.Int8u(minUint32(ca, 255))
		}
	} else {
		cr := uint32(c.R) + uint32(RGBA8MultCover(c2.R, cover))
		cg := uint32(c.G) + uint32(RGBA8MultCover(c2.G, cover))
		cb := uint32(c.B) + uint32(RGBA8MultCover(c2.B, cover))
		ca := uint32(c.A) + uint32(RGBA8MultCover(c2.A, cover))
		c.R = basics.Int8u(minUint32(cr, 255))
		c.G = basics.Int8u(minUint32(cg, 255))
		c.B = basics.Int8u(minUint32(cb, 255))
		c.A = basics.Int8u(minUint32(ca, 255))
	}
}

// Scale multiplies the color by a scalar value
func (c RGBA8[CS]) Scale(k float64) RGBA8[CS] {
	return RGBA8[CS]{
		R: basics.Int8u(minFloat64(float64(c.R)*k+0.5, 255)),
		G: basics.Int8u(minFloat64(float64(c.G)*k+0.5, 255)),
		B: basics.Int8u(minFloat64(float64(c.B)*k+0.5, 255)),
		A: basics.Int8u(minFloat64(float64(c.A)*k+0.5, 255)),
	}
}

// Subtract subtracts another RGBA8 color
func (c RGBA8[CS]) Subtract(c2 RGBA8[CS]) RGBA8[CS] {
	return RGBA8[CS]{
		R: basics.Int8u(maxInt32(int32(c.R)-int32(c2.R), 0)),
		G: basics.Int8u(maxInt32(int32(c.G)-int32(c2.G), 0)),
		B: basics.Int8u(maxInt32(int32(c.B)-int32(c2.B), 0)),
		A: basics.Int8u(maxInt32(int32(c.A)-int32(c2.A), 0)),
	}
}

// Multiply multiplies by another RGBA8 color
func (c RGBA8[CS]) Multiply(c2 RGBA8[CS]) RGBA8[CS] {
	return RGBA8[CS]{
		R: RGBA8Multiply(c.R, c2.R),
		G: RGBA8Multiply(c.G, c2.G),
		B: RGBA8Multiply(c.B, c2.B),
		A: RGBA8Multiply(c.A, c2.A),
	}
}

// AddAssign adds another color (modifies receiver)
func (c *RGBA8[CS]) AddAssign(c2 RGBA8[CS]) *RGBA8[CS] {
	c.R = basics.Int8u(minUint32(uint32(c.R)+uint32(c2.R), 255))
	c.G = basics.Int8u(minUint32(uint32(c.G)+uint32(c2.G), 255))
	c.B = basics.Int8u(minUint32(uint32(c.B)+uint32(c2.B), 255))
	c.A = basics.Int8u(minUint32(uint32(c.A)+uint32(c2.A), 255))
	return c
}

// MultiplyAssign multiplies by a scalar (modifies receiver)
func (c *RGBA8[CS]) MultiplyAssign(k float64) *RGBA8[CS] {
	c.R = basics.Int8u(minFloat64(float64(c.R)*k+0.5, 255))
	c.G = basics.Int8u(minFloat64(float64(c.G)*k+0.5, 255))
	c.B = basics.Int8u(minFloat64(float64(c.B)*k+0.5, 255))
	c.A = basics.Int8u(minFloat64(float64(c.A)*k+0.5, 255))
	return c
}

// Constants for RGBA8 fixed-point arithmetic
const (
	RGBA8BaseMask  = 255
	RGBA8BaseShift = 8
	RGBA8BaseMSB   = 1 << (RGBA8BaseShift - 1)
)

// RGBA8Multiply performs fixed-point multiplication for 8-bit values
func RGBA8Multiply(a, b basics.Int8u) basics.Int8u {
	t := uint32(a)*uint32(b) + RGBA8BaseMSB
	return basics.Int8u(((t >> RGBA8BaseShift) + t) >> RGBA8BaseShift)
}

// RGBA8Lerp performs linear interpolation between two values
// Matches AGG's lerp implementation exactly
func RGBA8Lerp(p, q, a basics.Int8u) basics.Int8u {
	// AGG implementation: int t = (q - p) * a + base_MSB - (p > q);
	// return value_type(p + (((t >> base_shift) + t) >> base_shift));
	var greater int32 = 0
	if p > q {
		greater = 1
	}
	t := (int32(q)-int32(p))*int32(a) + RGBA8BaseMSB - greater
	return basics.Int8u(int32(p) + (((t >> RGBA8BaseShift) + t) >> RGBA8BaseShift))
}

// RGBA8Prelerp performs premultiplied linear interpolation
func RGBA8Prelerp(p, q, a basics.Int8u) basics.Int8u {
	return p + q - RGBA8Multiply(p, a)
}

// RGBA8MultCover multiplies a color component by coverage
func RGBA8MultCover(c, cover basics.Int8u) basics.Int8u {
	return RGBA8Multiply(c, cover)
}
