package color

import "math"

// Gray32 represents a 32-bit floating-point grayscale color
type Gray32[CS Space] struct {
	V float32
	A float32
}

// NewGray32 creates a new 32-bit grayscale color
func NewGray32[CS Space](v float32) Gray32[CS] {
	return Gray32[CS]{V: v, A: 1.0}
}

// NewGray32WithAlpha creates a new 32-bit grayscale color with alpha
func NewGray32WithAlpha[CS Space](v, a float32) Gray32[CS] {
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

func (g *Gray32[CS]) Opacity(a float32) {
	g.A = clamp01f32(a)
}

func (g Gray32[CS]) GetOpacity() float32 { return clamp01f32(g.A) }

func (g *Gray32[CS]) Premultiply() {
	if g.A <= 0 {
		g.V = 0
	} else if g.A < 1 {
		g.V *= g.A
	}
}

func (g *Gray32[CS]) Demultiply() {
	if g.A <= 0 {
		g.V = 0
	} else if g.A < 1 {
		g.V /= g.A
	}
}

func Gray32Multiply(a, b float32) float32 { return a * b }

func Gray32Demultiply(a, b float32) float32 {
	if b <= 0 {
		return 0
	}
	return a / b
}

// Interpolate p→q by a (assuming q is premultiplied by a)
func Gray32Prelerp(p, q, a float32) float32 {
	// more accurate than p + q - p*a in float32:
	return (1-a)*p + q
}

// Interpolate p→q by a (0..1)
func Gray32Lerp(p, q, a float32) float32 {
	return (1-a)*p + a*q
}

// Gradient on Gray32 (value + alpha)
func (g Gray32[CS]) Gradient(c2 Gray32[CS], k float32) Gray32[CS] {
	k = clamp01f32(k)
	return Gray32[CS]{
		V: Gray32Lerp(g.V, c2.V, k),
		A: Gray32Lerp(g.A, c2.A, k),
	}
}

// ===== Add / blend with coverage =====
// cover is 0..255 (like AGG); we scale to 0..1 and reuse multiply.

func (g *Gray32[CS]) Add(c Gray32[CS], coverU8 uint8) {
	cover := float32(coverU8) / 255.0
	if coverU8 == 255 {
		if c.A >= 1 {
			*g = c
			return
		}
		g.V += c.V
		g.A += c.A
	} else {
		g.V += Gray32Multiply(c.V, cover)
		g.A += Gray32Multiply(c.A, cover)
	}
	// clamp to [0,1]
	g.V = clamp01f32(g.V)
	g.A = clamp01f32(g.A)
}

func (g *Gray32[CS]) AddWithCover(c Gray32[CS], cover uint8) { g.Add(c, cover) }

// EmptyValue returns the empty (transparent) value
func Gray32EmptyValue() float32 {
	return 0.0
}

// FullValue returns the full (opaque) value
func Gray32FullValue() float32 {
	return 1.0
}

// ===== Constants & small helpers =====

func clamp01f32(x float32) float32 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// ConvertFromRGBA converts from floating-point RGBA
func ConvertGray32FromRGBA[CS Space](c RGBA) Gray32[CS] {
	lum := LuminanceFromRGBA(c)
	return Gray32[CS]{
		V: float32(lum),
		A: float32(c.A),
	}
}

func (g *Gray32[CS]) Sanitize() {
	// eliminate NaNs and clamp
	if math.IsNaN(float64(g.V)) {
		g.V = 0
	}
	if math.IsNaN(float64(g.A)) {
		g.A = 0
	}
	g.V = clamp01f32(g.V)
	g.A = clamp01f32(g.A)
}

// Apply 32-bit gamma (V channel only) to a Gray32 pixel in-place.
func ApplyGammaDir32Gray[CS Space, G lut32Like](px *Gray32[CS], g G) {
	px.V = g.DirFloat(px.V)
}

func ApplyGammaInv32Gray[CS Space, G lut32Like](px *Gray32[CS], g G) {
	px.V = g.InvFloat(px.V)
}

// Helper methods for Gray32
func (g *Gray32[CS]) ApplyGammaDir(gamma lut32Like) { ApplyGammaDir32Gray(g, gamma) }
func (g *Gray32[CS]) ApplyGammaInv(gamma lut32Like) { ApplyGammaInv32Gray(g, gamma) }

// Common 32-bit grayscale types
type (
	Gray32Linear = Gray32[Linear]
	Gray32SRGB   = Gray32[SRGB]
)
