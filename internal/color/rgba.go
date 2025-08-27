// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

import (
	"math"

	"agg_go/internal/basics"
)

// ColorOrder defines component ordering for different pixel formats
type ColorOrder struct {
	R, G, B, A int
}

var (
	OrderRGB  = ColorOrder{R: 0, G: 1, B: 2, A: 3}
	OrderBGR  = ColorOrder{R: 2, G: 1, B: 0, A: 3}
	OrderRGBA = ColorOrder{R: 0, G: 1, B: 2, A: 3}
	OrderARGB = ColorOrder{R: 1, G: 2, B: 3, A: 0}
	OrderABGR = ColorOrder{R: 3, G: 2, B: 1, A: 0}
	OrderBGRA = ColorOrder{R: 2, G: 1, B: 0, A: 3}
)

// GammaLUT represents a gamma lookup table interface
// This is defined here to avoid circular imports with pixfmt
type GammaLUT interface {
	Dir(v basics.Int8u) basics.Int8u // Direct gamma correction
	Inv(v basics.Int8u) basics.Int8u // Inverse gamma correction
}

// GammaFunc represents a gamma correction function interface for floating-point values
type GammaFunc interface {
	DirFloat(v float32) float32 // Direct gamma correction for float
	InvFloat(v float32) float32 // Inverse gamma correction for float
}

// ColorBlender defines the interface for colors that can be blended with cover values
// for compound rendering operations. This matches the C++ AGG add(color, cover) method.
type ColorBlender interface {
	AddWithCover(c ColorBlender, cover basics.Int8u)
}

// RGBA represents a floating-point RGBA color (base type)
type RGBA struct {
	R, G, B, A float64
}

// NewRGBA creates a new RGBA color
func NewRGBA(r, g, b, a float64) RGBA {
	return RGBA{R: r, G: g, B: b, A: a}
}

// Clear sets the color to transparent black
func (c *RGBA) Clear() *RGBA {
	c.R, c.G, c.B, c.A = 0, 0, 0, 0
	return c
}

// Transparent sets the color to transparent with the same RGB
func (c *RGBA) Transparent() *RGBA {
	c.A = 0
	return c
}

// Opacity sets the alpha channel
func (c *RGBA) Opacity(a float64) *RGBA {
	if a < 0 {
		c.A = 0
	} else if a > 1 {
		c.A = 1
	} else {
		c.A = a
	}
	return c
}

// GetOpacity returns the current alpha value
func (c RGBA) GetOpacity() float64 {
	return c.A
}

// Premultiply premultiplies the color by alpha
func (c *RGBA) Premultiply() *RGBA {
	if c.A <= 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A < 1 {
		c.R *= c.A
		c.G *= c.A
		c.B *= c.A
	}
	return c
}

// PremultiplyAlpha premultiplies the color by a specific alpha value
func (c *RGBA) PremultiplyAlpha(a float64) *RGBA {
	if c.A <= 0 || a <= 0 {
		c.R, c.G, c.B, c.A = 0, 0, 0, 0
	} else {
		a /= c.A
		c.R *= a
		c.G *= a
		c.B *= a
		c.A = a
	}
	return c
}

// Demultiply demultiplies the color by alpha
func (c *RGBA) Demultiply() *RGBA {
	if c.A == 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A > 0 && c.A < 1 {
		inv := 1.0 / c.A
		c.R *= inv
		c.G *= inv
		c.B *= inv
	}
	return c
}

// Gradient performs linear interpolation between two colors
func (c RGBA) Gradient(c2 RGBA, k float64) RGBA {
	return RGBA{
		R: c.R + k*(c2.R-c.R),
		G: c.G + k*(c2.G-c.G),
		B: c.B + k*(c2.B-c.B),
		A: c.A + k*(c2.A-c.A),
	}
}

// Add adds another color
func (c RGBA) Add(c2 RGBA) RGBA {
	return RGBA{
		R: c.R + c2.R,
		G: c.G + c2.G,
		B: c.B + c2.B,
		A: c.A + c2.A,
	}
}

// Subtract subtracts another color
func (c RGBA) Subtract(c2 RGBA) RGBA {
	return RGBA{
		R: c.R - c2.R,
		G: c.G - c2.G,
		B: c.B - c2.B,
		A: c.A - c2.A,
	}
}

// Multiply multiplies by another color
func (c RGBA) Multiply(c2 RGBA) RGBA {
	return RGBA{
		R: c.R * c2.R,
		G: c.G * c2.G,
		B: c.B * c2.B,
		A: c.A * c2.A,
	}
}

// Scale multiplies by a scalar
func (c RGBA) Scale(k float64) RGBA {
	return RGBA{
		R: c.R * k,
		G: c.G * k,
		B: c.B * k,
		A: c.A * k,
	}
}

// AddAssign adds another color (equivalent to C++ +=)
func (c *RGBA) AddAssign(c2 RGBA) *RGBA {
	c.R += c2.R
	c.G += c2.G
	c.B += c2.B
	c.A += c2.A
	return c
}

// MultiplyAssign multiplies by a scalar (equivalent to C++ *=)
func (c *RGBA) MultiplyAssign(k float64) *RGBA {
	c.R *= k
	c.G *= k
	c.B *= k
	c.A *= k
	return c
}

// NoColor returns transparent black
func NoColor() RGBA {
	return RGBA{0, 0, 0, 0}
}

// FromWavelength creates an RGB color from a wavelength (380-780 nm)
// This follows the C++ AGG implementation exactly
func FromWavelength(wl, gamma float64) RGBA {
	t := RGBA{0.0, 0.0, 0.0, 1.0}

	if wl >= 380.0 && wl <= 440.0 {
		t.R = -1.0 * (wl - 440.0) / (440.0 - 380.0)
		t.B = 1.0
	} else if wl >= 440.0 && wl <= 490.0 {
		t.G = (wl - 440.0) / (490.0 - 440.0)
		t.B = 1.0
	} else if wl >= 490.0 && wl <= 510.0 {
		t.G = 1.0
		t.B = -1.0 * (wl - 510.0) / (510.0 - 490.0)
	} else if wl >= 510.0 && wl <= 580.0 {
		t.R = (wl - 510.0) / (580.0 - 510.0)
		t.G = 1.0
	} else if wl >= 580.0 && wl <= 645.0 {
		t.R = 1.0
		t.G = -1.0 * (wl - 645.0) / (645.0 - 580.0)
	} else if wl >= 645.0 && wl <= 780.0 {
		t.R = 1.0
	}

	s := 1.0
	if wl > 700.0 {
		s = 0.3 + 0.7*(780.0-wl)/(780.0-700.0)
	} else if wl < 420.0 {
		s = 0.3 + 0.7*(wl-380.0)/(420.0-380.0)
	}

	t.R = math.Pow(t.R*s, gamma)
	t.G = math.Pow(t.G*s, gamma)
	t.B = math.Pow(t.B*s, gamma)
	return t
}

// NewRGBAFromWavelength creates an RGBA color from wavelength (constructor equivalent)
func NewRGBAFromWavelength(wl, gamma float64) RGBA {
	return FromWavelength(wl, gamma)
}

// RGBAPre creates a premultiplied RGBA color
func RGBAPre(r, g, b, a float64) RGBA {
	c := NewRGBA(r, g, b, a)
	c.Premultiply()
	return c
}

// RGBA8 represents an 8-bit RGBA color with colorspace
type RGBA8[CS any] struct {
	R, G, B, A basics.Int8u
}

// NewRGBA8 creates a new 8-bit RGBA color
func NewRGBA8[CS any](r, g, b, a basics.Int8u) RGBA8[CS] {
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
func ConvertFromRGBA[CS any](c RGBA) RGBA8[CS] {
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

// ApplyGammaDir applies direct gamma correction to RGB components (alpha unchanged)
func (c *RGBA8[CS]) ApplyGammaDir(gamma GammaLUT) {
	c.R = gamma.Dir(c.R)
	c.G = gamma.Dir(c.G)
	c.B = gamma.Dir(c.B)
	// Alpha component is not affected by gamma correction
}

// ApplyGammaInv applies inverse gamma correction to RGB components (alpha unchanged)
func (c *RGBA8[CS]) ApplyGammaInv(gamma GammaLUT) {
	c.R = gamma.Inv(c.R)
	c.G = gamma.Inv(c.G)
	c.B = gamma.Inv(c.B)
	// Alpha component is not affected by gamma correction
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
type RGBA16[CS any] struct {
	R, G, B, A basics.Int16u
}

// NewRGBA16 creates a new 16-bit RGBA color
func NewRGBA16[CS any](r, g, b, a basics.Int16u) RGBA16[CS] {
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
func ConvertFromRGBA16[CS any](c RGBA) RGBA16[CS] {
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

// ApplyGammaDir applies direct gamma correction to RGB components (alpha unchanged)
// Converts 16-bit values to 8-bit for gamma lookup, then scales back up
func (c *RGBA16[CS]) ApplyGammaDir(gamma GammaLUT) {
	// Convert to 8-bit for gamma lookup
	r8 := basics.Int8u(c.R >> 8)
	g8 := basics.Int8u(c.G >> 8)
	b8 := basics.Int8u(c.B >> 8)

	// Apply gamma correction
	r8Corrected := gamma.Dir(r8)
	g8Corrected := gamma.Dir(g8)
	b8Corrected := gamma.Dir(b8)

	// Scale back to 16-bit (duplicate the byte: 0xAB -> 0xABAB)
	c.R = basics.Int16u(r8Corrected)<<8 | basics.Int16u(r8Corrected)
	c.G = basics.Int16u(g8Corrected)<<8 | basics.Int16u(g8Corrected)
	c.B = basics.Int16u(b8Corrected)<<8 | basics.Int16u(b8Corrected)
	// Alpha component is not affected by gamma correction
}

// ApplyGammaInv applies inverse gamma correction to RGB components (alpha unchanged)
// Converts 16-bit values to 8-bit for gamma lookup, then scales back up
func (c *RGBA16[CS]) ApplyGammaInv(gamma GammaLUT) {
	// Convert to 8-bit for gamma lookup
	r8 := basics.Int8u(c.R >> 8)
	g8 := basics.Int8u(c.G >> 8)
	b8 := basics.Int8u(c.B >> 8)

	// Apply inverse gamma correction
	r8Corrected := gamma.Inv(r8)
	g8Corrected := gamma.Inv(g8)
	b8Corrected := gamma.Inv(b8)

	// Scale back to 16-bit (duplicate the byte: 0xAB -> 0xABAB)
	c.R = basics.Int16u(r8Corrected)<<8 | basics.Int16u(r8Corrected)
	c.G = basics.Int16u(g8Corrected)<<8 | basics.Int16u(g8Corrected)
	c.B = basics.Int16u(b8Corrected)<<8 | basics.Int16u(b8Corrected)
	// Alpha component is not affected by gamma correction
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

// RGBA32 represents a 32-bit floating-point RGBA color
type RGBA32[CS any] struct {
	R, G, B, A float32
}

// NewRGBA32 creates a new 32-bit RGBA color
func NewRGBA32[CS any](r, g, b, a float32) RGBA32[CS] {
	return RGBA32[CS]{R: r, G: g, B: b, A: a}
}

// ConvertToRGBA converts to floating-point RGBA
func (c RGBA32[CS]) ConvertToRGBA() RGBA {
	return RGBA{
		R: float64(c.R),
		G: float64(c.G),
		B: float64(c.B),
		A: float64(c.A),
	}
}

// ConvertFromRGBA converts from floating-point RGBA
func ConvertFromRGBA32[CS any](c RGBA) RGBA32[CS] {
	return RGBA32[CS]{
		R: float32(c.R),
		G: float32(c.G),
		B: float32(c.B),
		A: float32(c.A),
	}
}

// Premultiply premultiplies the color by alpha
func (c *RGBA32[CS]) Premultiply() *RGBA32[CS] {
	if c.A <= 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A < 1 {
		c.R *= c.A
		c.G *= c.A
		c.B *= c.A
	}
	return c
}

// Demultiply demultiplies the color by alpha
func (c *RGBA32[CS]) Demultiply() *RGBA32[CS] {
	if c.A == 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A > 0 && c.A < 1 {
		inv := 1.0 / c.A
		c.R *= inv
		c.G *= inv
		c.B *= inv
	}
	return c
}

// Gradient performs linear interpolation between two 32-bit colors
func (c RGBA32[CS]) Gradient(c2 RGBA32[CS], k float32) RGBA32[CS] {
	return RGBA32[CS]{
		R: c.R + k*(c2.R-c.R),
		G: c.G + k*(c2.G-c.G),
		B: c.B + k*(c2.B-c.B),
		A: c.A + k*(c2.A-c.A),
	}
}

// Clear sets the color to transparent black
func (c *RGBA32[CS]) Clear() *RGBA32[CS] {
	c.R, c.G, c.B, c.A = 0, 0, 0, 0
	return c
}

// Transparent sets the color to transparent with the same RGB
func (c *RGBA32[CS]) Transparent() *RGBA32[CS] {
	c.A = 0
	return c
}

// IsTransparent returns true if the color is fully transparent
func (c RGBA32[CS]) IsTransparent() bool {
	return c.A == 0
}

// IsOpaque returns true if the color is fully opaque
func (c RGBA32[CS]) IsOpaque() bool {
	return c.A == 1.0
}

// Opacity sets the alpha channel (0.0 to 1.0)
func (c *RGBA32[CS]) Opacity(a float64) *RGBA32[CS] {
	if a < 0 {
		c.A = 0
	} else if a > 1 {
		c.A = 1
	} else {
		c.A = float32(a)
	}
	return c
}

// GetOpacity returns the alpha as a float64 (0.0 to 1.0)
func (c RGBA32[CS]) GetOpacity() float64 {
	return float64(c.A)
}

// ApplyGammaDir applies direct gamma correction to RGB components (alpha unchanged)
func (c *RGBA32[CS]) ApplyGammaDir(gamma GammaFunc) {
	c.R = gamma.DirFloat(c.R)
	c.G = gamma.DirFloat(c.G)
	c.B = gamma.DirFloat(c.B)
	// Alpha component is not affected by gamma correction
}

// ApplyGammaInv applies inverse gamma correction to RGB components (alpha unchanged)
func (c *RGBA32[CS]) ApplyGammaInv(gamma GammaFunc) {
	c.R = gamma.InvFloat(c.R)
	c.G = gamma.InvFloat(c.G)
	c.B = gamma.InvFloat(c.B)
	// Alpha component is not affected by gamma correction
}

// Add adds another RGBA32 color
func (c RGBA32[CS]) Add(c2 RGBA32[CS]) RGBA32[CS] {
	return RGBA32[CS]{
		R: c.R + c2.R,
		G: c.G + c2.G,
		B: c.B + c2.B,
		A: c.A + c2.A,
	}
}

// AddWithCover adds another RGBA32 color with coverage, matching C++ AGG's add(color, cover) method
func (c *RGBA32[CS]) AddWithCover(c2 RGBA32[CS], cover basics.Int8u) {
	coverFloat := float32(cover) / 255.0
	if cover == basics.CoverMask {
		if c2.A == 1.0 { // base_mask for float32 (1.0)
			*c = c2
			return
		} else {
			c.R = minFloat32(c.R+c2.R, 1.0)
			c.G = minFloat32(c.G+c2.G, 1.0)
			c.B = minFloat32(c.B+c2.B, 1.0)
			c.A = minFloat32(c.A+c2.A, 1.0)
		}
	} else {
		c.R = minFloat32(c.R+c2.R*coverFloat, 1.0)
		c.G = minFloat32(c.G+c2.G*coverFloat, 1.0)
		c.B = minFloat32(c.B+c2.B*coverFloat, 1.0)
		c.A = minFloat32(c.A+c2.A*coverFloat, 1.0)
	}
}

// Scale multiplies the color by a scalar value
func (c RGBA32[CS]) Scale(k float32) RGBA32[CS] {
	return RGBA32[CS]{
		R: c.R * k,
		G: c.G * k,
		B: c.B * k,
		A: c.A * k,
	}
}

// Common 32-bit color types
type (
	RGBA32Linear = RGBA32[Linear]
	RGBA32SRGB   = RGBA32[SRGB]
)
