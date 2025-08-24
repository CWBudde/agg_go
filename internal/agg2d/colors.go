// Package agg2d color management for AGG2D high-level interface.
// This file contains additional color-related methods and functionality.
package agg2d

import "agg_go/internal/color"

// Additional predefined colors
var (
	Red   = Color{255, 0, 0, 255}
	Green = Color{0, 255, 0, 255}
	Blue  = Color{0, 0, 255, 255}
)

// NewColorRGB creates a new opaque Color with alpha = 255.
func NewColorRGB(r, g, b uint8) Color {
	return Color{r, g, b, 255}
}

// ToRGBA converts the color to floating-point color.RGBA (0.0-1.0 range).
func (c Color) ToRGBA() color.RGBA {
	const scale = 1.0 / 255.0
	return color.RGBA{
		R: float64(c[0]) * scale,
		G: float64(c[1]) * scale,
		B: float64(c[2]) * scale,
		A: float64(c[3]) * scale,
	}
}

// RGBA returns the RGBA components of the color.
func (c Color) RGBA() (r, g, b, a uint8) {
	return c[0], c[1], c[2], c[3]
}

// R returns the red component.
func (c Color) R() uint8 { return c[0] }

// G returns the green component.
func (c Color) G() uint8 { return c[1] }

// B returns the blue component.
func (c Color) B() uint8 { return c[2] }

// A returns the alpha component.
func (c Color) A() uint8 { return c[3] }

// FillColorRGBA sets the fill color using RGBA components.
func (agg2d *Agg2D) FillColorRGBA(r, g, b, a uint8) {
	agg2d.FillColor(NewColor(r, g, b, a))
}

// LineColorRGBA sets the line color using RGBA components.
func (agg2d *Agg2D) LineColorRGBA(r, g, b, a uint8) {
	agg2d.LineColor(NewColor(r, g, b, a))
}

// GetFillColor returns the current fill color.
func (agg2d *Agg2D) GetFillColor() Color {
	return agg2d.fillColor
}

// GetLineColor returns the current line color.
func (agg2d *Agg2D) GetLineColor() Color {
	return agg2d.lineColor
}

// GetClipBox returns the current clipping rectangle.
func (agg2d *Agg2D) GetClipBox() (x1, y1, x2, y2 float64) {
	return agg2d.clipBox.X1, agg2d.clipBox.Y1, agg2d.clipBox.X2, agg2d.clipBox.Y2
}

// ClearAllRGBA fills the entire buffer with the specified RGBA color.
func (agg2d *Agg2D) ClearAllRGBA(r, g, b, a uint8) {
	agg2d.ClearAll(NewColor(r, g, b, a))
}
