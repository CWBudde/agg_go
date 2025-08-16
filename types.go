// Package agg provides a Go port of the Anti-Grain Geometry (AGG) rendering library.
//
// AGG is a high-quality 2D graphics library that provides anti-aliased rendering
// with subpixel accuracy. This Go port maintains the core functionality while
// providing a Go-idiomatic API.
//
// Basic usage:
//
//	ctx := agg.NewContext(800, 600)
//	ctx.SetColor(agg.RGBA{1, 0, 0, 1}) // Red
//	ctx.DrawCircle(400, 300, 100)
//	ctx.Fill()
package agg

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

// Color represents a color that can be used for rendering.
// The library supports both floating-point and integer color formats.
type Color interface {
	// ConvertToRGBA converts the color to floating-point RGBA
	ConvertToRGBA() color.RGBA
}

// RGBA represents a floating-point RGBA color (0.0 to 1.0 range).
type RGBA struct {
	R, G, B, A float64
}

// ConvertToRGBA implements the Color interface
func (c RGBA) ConvertToRGBA() color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// NewRGBA creates a new RGBA color with the specified components.
func NewRGBA(r, g, b, a float64) RGBA {
	return RGBA{R: r, G: g, B: b, A: a}
}

// RGB creates an opaque RGBA color with alpha = 1.0.
func RGB(r, g, b float64) RGBA {
	return RGBA{R: r, G: g, B: b, A: 1.0}
}

// RGBA8 represents an 8-bit RGBA color (0 to 255 range).
type RGBA8 struct {
	R, G, B, A uint8
}

// ConvertToRGBA implements the Color interface
func (c RGBA8) ConvertToRGBA() color.RGBA {
	const scale = 1.0 / 255.0
	return color.RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: float64(c.A) * scale,
	}
}

// NewRGBA8 creates a new 8-bit RGBA color.
func NewRGBA8(r, g, b, a uint8) RGBA8 {
	return RGBA8{R: r, G: g, B: b, A: a}
}

// RGB8 creates an opaque 8-bit RGBA color with alpha = 255.
func RGB8(r, g, b uint8) RGBA8 {
	return RGBA8{R: r, G: g, B: b, A: 255}
}

// Gray represents a grayscale color.
type Gray struct {
	V float64 // Value (luminance)
}

// ConvertToRGBA implements the Color interface
func (g Gray) ConvertToRGBA() color.RGBA {
	return color.RGBA{R: g.V, G: g.V, B: g.V, A: 1.0}
}

// NewGray creates a new grayscale color.
func NewGray(v float64) Gray {
	return Gray{V: v}
}

// Point represents a 2D point with floating-point coordinates.
type Point struct {
	X, Y float64
}

// NewPoint creates a new point.
func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Rect represents a rectangle.
type Rect struct {
	X, Y, W, H float64
}

// NewRect creates a new rectangle.
func NewRect(x, y, w, h float64) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

// Transform represents a 2D affine transformation matrix.
type Transform struct {
	M11, M12, M21, M22, M31, M32 float64
}

// NewTransform creates an identity transform.
func NewTransform() Transform {
	return Transform{M11: 1, M22: 1}
}

// Translate creates a translation transform.
func Translate(dx, dy float64) Transform {
	return Transform{M11: 1, M22: 1, M31: dx, M32: dy}
}

// Scale creates a scaling transform.
func Scale(sx, sy float64) Transform {
	return Transform{M11: sx, M22: sy}
}

// Rotate creates a rotation transform (angle in radians).
func Rotate(angle float64) Transform {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Transform{M11: c, M12: -s, M21: s, M22: c}
}

// Context represents the main rendering context.
// It provides a high-level interface for drawing operations.
type Context struct {
	width, height    int
	currentColor     Color
	currentPath      *Path
	currentTransform Transform
	buffer           *buffer.RenderingBuffer[uint8]
}

// Path represents a vector path that can contain lines, curves, and shapes.
type Path struct {
	// Internal path data will be implemented with internal/curves
	commands []pathCommand
}

type pathCommand struct {
	cmd    basics.PathCommand
	points []Point
}

// Image represents a raster image that can be used as a rendering target.
type Image struct {
	Width, Height int
	Stride        int
	Data          []uint8
}

// Predefined colors for convenience
var (
	Black       = RGB(0, 0, 0)
	White       = RGB(1, 1, 1)
	Red         = RGB(1, 0, 0)
	Green       = RGB(0, 1, 0)
	Blue        = RGB(0, 0, 1)
	Yellow      = RGB(1, 1, 0)
	Cyan        = RGB(0, 1, 1)
	Magenta     = RGB(1, 0, 1)
	Transparent = RGBA{0, 0, 0, 0}
)
