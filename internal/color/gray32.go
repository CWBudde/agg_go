package color

// Gray32 represents a 32-bit floating-point grayscale color
type Gray32[CS ColorSpace] struct {
	V float32
	A float32
}

// NewGray32 creates a new 32-bit grayscale color
func NewGray32[CS ColorSpace](v float32) Gray32[CS] {
	return Gray32[CS]{V: v, A: 1.0}
}

// NewGray32WithAlpha creates a new 32-bit grayscale color with alpha
func NewGray32WithAlpha[CS ColorSpace](v, a float32) Gray32[CS] {
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
func ConvertGray32FromRGBA[CS ColorSpace](c RGBA) Gray32[CS] {
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
