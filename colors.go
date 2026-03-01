// Package agg provides color types and color management functionality.
// This file contains all color-related types and operations.
package agg

import (
	"agg_go/internal/color"
)

// Color represents an SRGBA8 color (0-255 range) matching AGG's srgba8.
// This is the primary color type used throughout the AGG2D interface.
type Color struct {
	R, G, B, A uint8
}

// NewColor creates a new Color with the specified RGBA components.
func NewColor(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// NewColorRGBA8 creates a new Color from internal RGBA8.
func NewColorRGBA8(c color.RGBA8[color.Linear]) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: c.A}
}

// NewColorRGB creates a new opaque Color with alpha = 255.
func NewColorRGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: 255}
}

// ConvertToRGBA converts the color to floating-point RGBA (0.0-1.0 range).
func (c Color) ConvertToRGBA() color.RGBA {
	const scale = 1.0 / 255.0
	return color.RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: float64(c.A) * scale,
	}
}

// ConvertToInternalRGBA converts the public Color to internal color.RGBA.
func (c Color) ConvertToInternalRGBA() color.RGBA {
	const scale = 1.0 / 255.0
	return color.RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: float64(c.A) * scale,
	}
}

// Gradient performs linear interpolation between this color and another color.
// Parameter k should be in the range [0.0, 1.0] where:
// - k = 0.0 returns this color (c)
// - k = 1.0 returns the target color (c2)
// - k = 0.5 returns the midpoint between both colors
// This matches the C++ AGG srgba8::gradient method.
func (c Color) Gradient(c2 Color, k float64) Color {
	if k <= 0.0 {
		return c
	}
	if k >= 1.0 {
		return c2
	}

	// Linear interpolation: c + (c2 - c) * k
	return Color{
		R: uint8(float64(c.R) + (float64(c2.R)-float64(c.R))*k),
		G: uint8(float64(c.G) + (float64(c2.G)-float64(c.G))*k),
		B: uint8(float64(c.B) + (float64(c2.B)-float64(c.B))*k),
		A: uint8(float64(c.A) + (float64(c2.A)-float64(c.A))*k),
	}
}

// RGBA8 type for compatibility with examples that might use it
type RGBA8 struct {
	R, G, B, A uint8
}

// NewRGBA8 creates a new RGBA8 color.
func NewRGBA8(r, g, b, a uint8) RGBA8 {
	return RGBA8{R: r, G: g, B: b, A: a}
}

// RGBAColor represents a floating-point RGBA color.
type RGBAColor struct {
	R, G, B, A float64
}

// NewRGBAColor creates a new RGBAColor with floating-point values (0.0-1.0).
func NewRGBAColor(r, g, b, a float64) RGBAColor {
	return RGBAColor{R: r, G: g, B: b, A: a}
}

// ToColor converts RGBAColor to Color (0-255 range).
func (c RGBAColor) ToColor() Color {
	return Color{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}

// Predefined colors for convenience
var (
	Black       = NewColorRGB(0, 0, 0)
	White       = NewColorRGB(255, 255, 255)
	Red         = NewColorRGB(255, 0, 0)
	Green       = NewColorRGB(0, 255, 0)
	Blue        = NewColorRGB(0, 0, 255)
	Yellow      = NewColorRGB(255, 255, 0)
	Cyan        = NewColorRGB(0, 255, 255)
	Magenta     = NewColorRGB(255, 0, 255)
	Transparent = NewColor(0, 0, 0, 0)

	// Additional common colors
	Gray      = NewColorRGB(128, 128, 128)
	LightGray = NewColorRGB(192, 192, 192)
	DarkGray  = NewColorRGB(64, 64, 64)
	Orange    = NewColorRGB(255, 165, 0)
	Purple    = NewColorRGB(128, 0, 128)
	Pink      = NewColorRGB(255, 192, 203)
	Brown     = NewColorRGB(165, 42, 42)
)

// Color convenience constructors

// RGB creates a Color from floating-point RGB values (0.0-1.0 range).
// This is a convenience function for creating colors from normalized values.
func RGB(r, g, b float64) Color {
	return Color{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// RGBA creates a Color from floating-point RGBA values (0.0-1.0 range).
// This is a convenience function for creating colors from normalized values.
func RGBA(r, g, b, a float64) Color {
	return Color{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: uint8(a * 255),
	}
}

// HSV creates a Color from HSV (Hue, Saturation, Value) components.
// h is in degrees (0-360), s and v are in range (0.0-1.0).
func HSV(h, s, v float64) Color {
	// Normalize hue to [0, 360)
	for h < 0 {
		h += 360
	}
	for h >= 360 {
		h -= 360
	}

	// Clamp s and v to [0, 1]
	if s < 0 {
		s = 0
	}
	if s > 1 {
		s = 1
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	c := v * s
	x := c * (1 - colorAbs(colorMod(h/60, 2)-1))
	m := v - c

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return RGB(r+m, g+m, b+m)
}

// Helper functions for HSV (prefixed to avoid conflicts)
func colorAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func colorMod(x, y float64) float64 {
	return x - y*float64(int(x/y))
}

// Lighter returns a lighter version of the color by the given amount (0.0-1.0).
func (c Color) Lighter(amount float64) Color {
	if amount < 0 {
		amount = 0
	}
	if amount > 1 {
		amount = 1
	}

	factor := 1 + amount
	return Color{
		R: uint8(colorMin(float64(c.R)*factor, 255)),
		G: uint8(colorMin(float64(c.G)*factor, 255)),
		B: uint8(colorMin(float64(c.B)*factor, 255)),
		A: c.A,
	}
}

// Darker returns a darker version of the color by the given amount (0.0-1.0).
func (c Color) Darker(amount float64) Color {
	if amount < 0 {
		amount = 0
	}
	if amount > 1 {
		amount = 1
	}

	factor := 1 - amount
	return Color{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

// WithAlpha returns a new color with the specified alpha value.
func (c Color) WithAlpha(alpha uint8) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: alpha}
}

// WithAlphaF returns a new color with the specified alpha value (0.0-1.0).
func (c Color) WithAlphaF(alpha float64) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: uint8(alpha * 255)}
}

func colorMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
