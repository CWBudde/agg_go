// Package agg provides gradient functionality for 2D graphics.
// This file contains gradient types and operations.
package agg

import (
	"math"
)

// Gradient represents gradient types.
type Gradient int

const (
	Solid  Gradient = iota // Solid color
	Linear                 // Linear gradient
	Radial                 // Radial gradient
)

// Gradient types for different gradient modes
type GradientType int

const (
	SolidGradient  GradientType = 0
	LinearGradient GradientType = 1
	RadialGradient GradientType = 2
)

// GradientStop represents a color stop in a gradient
type GradientStop struct {
	Position float64 // Position along gradient (0.0 to 1.0)
	Color    Color   // Color at this position
}

// LinearGradientSpec defines a linear gradient
type LinearGradientSpec struct {
	X1, Y1  float64        // Starting point
	X2, Y2  float64        // Ending point
	Stops   []GradientStop // Color stops
	Profile float64        // Gradient profile (sharpness)
}

// RadialGradientSpec defines a radial gradient
type RadialGradientSpec struct {
	CX, CY  float64        // Center point
	Radius  float64        // Radius
	Stops   []GradientStop // Color stops
	Profile float64        // Gradient profile (sharpness)
}

// Context gradient methods

// SetLinearGradient sets a linear gradient for fill operations.
func (ctx *Context) SetLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color) {
	ctx.agg2d.FillLinearGradient(x1, y1, x2, y2, c1, c2, 1.0)
}

// SetLinearGradientWithProfile sets a linear gradient with custom profile.
func (ctx *Context) SetLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	ctx.agg2d.FillLinearGradient(x1, y1, x2, y2, c1, c2, profile)
}

// SetRadialGradient sets a radial gradient for fill operations.
func (ctx *Context) SetRadialGradient(cx, cy, radius float64, c1, c2 Color) {
	ctx.agg2d.FillRadialGradient(cx, cy, radius, c1, c2, 1.0)
}

// SetRadialGradientWithProfile sets a radial gradient with custom profile.
func (ctx *Context) SetRadialGradientWithProfile(cx, cy, radius float64, c1, c2 Color, profile float64) {
	ctx.agg2d.FillRadialGradient(cx, cy, radius, c1, c2, profile)
}

// SetRadialGradientMultiStop sets a radial gradient with three color stops.
func (ctx *Context) SetRadialGradientMultiStop(cx, cy, radius float64, c1, c2, c3 Color) {
	ctx.agg2d.FillRadialGradientMultiStop(cx, cy, radius, c1, c2, c3)
}

// SetStrokeLinearGradient sets a linear gradient for stroke operations.
func (ctx *Context) SetStrokeLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color) {
	ctx.agg2d.LineLinearGradient(x1, y1, x2, y2, c1, c2, 1.0)
}

// SetStrokeLinearGradientWithProfile sets a linear gradient for strokes with custom profile.
func (ctx *Context) SetStrokeLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	ctx.agg2d.LineLinearGradient(x1, y1, x2, y2, c1, c2, profile)
}

// SetStrokeRadialGradient sets a radial gradient for stroke operations.
func (ctx *Context) SetStrokeRadialGradient(cx, cy, radius float64, c1, c2 Color) {
	ctx.agg2d.LineRadialGradient(cx, cy, radius, c1, c2, 1.0)
}

// SetStrokeRadialGradientWithProfile sets a radial gradient for strokes with custom profile.
func (ctx *Context) SetStrokeRadialGradientWithProfile(cx, cy, radius float64, c1, c2 Color, profile float64) {
	ctx.agg2d.LineRadialGradient(cx, cy, radius, c1, c2, profile)
}

// SetStrokeRadialGradientMultiStop sets a radial gradient with three color stops for strokes.
func (ctx *Context) SetStrokeRadialGradientMultiStop(cx, cy, radius float64, c1, c2, c3 Color) {
	ctx.agg2d.LineRadialGradientMultiStop(cx, cy, radius, c1, c2, c3)
}

// Gradient utility functions

// CreateLinearGradientSpec creates a linear gradient specification.
func CreateLinearGradientSpec(x1, y1, x2, y2 float64) *LinearGradientSpec {
	return &LinearGradientSpec{
		X1: x1, Y1: y1,
		X2: x2, Y2: y2,
		Stops:   make([]GradientStop, 0),
		Profile: 1.0,
	}
}

// AddStop adds a color stop to a linear gradient.
func (lg *LinearGradientSpec) AddStop(position float64, color Color) {
	// Clamp position to valid range
	if position < 0.0 {
		position = 0.0
	} else if position > 1.0 {
		position = 1.0
	}

	lg.Stops = append(lg.Stops, GradientStop{
		Position: position,
		Color:    color,
	})
}

// SetProfile sets the gradient profile (sharpness).
func (lg *LinearGradientSpec) SetProfile(profile float64) {
	if profile < 0.0 {
		profile = 0.0
	} else if profile > 2.0 {
		profile = 2.0
	}
	lg.Profile = profile
}

// CreateRadialGradientSpec creates a radial gradient specification.
func CreateRadialGradientSpec(cx, cy, radius float64) *RadialGradientSpec {
	return &RadialGradientSpec{
		CX: cx, CY: cy,
		Radius:  radius,
		Stops:   make([]GradientStop, 0),
		Profile: 1.0,
	}
}

// AddStop adds a color stop to a radial gradient.
func (rg *RadialGradientSpec) AddStop(position float64, color Color) {
	// Clamp position to valid range
	if position < 0.0 {
		position = 0.0
	} else if position > 1.0 {
		position = 1.0
	}

	rg.Stops = append(rg.Stops, GradientStop{
		Position: position,
		Color:    color,
	})
}

// SetProfile sets the gradient profile (sharpness).
func (rg *RadialGradientSpec) SetProfile(profile float64) {
	if profile < 0.0 {
		profile = 0.0
	} else if profile > 2.0 {
		profile = 2.0
	}
	rg.Profile = profile
}

// Gradient query methods

// GetFillGradientType returns the current fill gradient type.
func (ctx *Context) GetFillGradientType() GradientType {
	return GradientType(ctx.agg2d.FillGradientFlag())
}

// GetStrokeGradientType returns the current stroke gradient type.
func (ctx *Context) GetStrokeGradientType() GradientType {
	return GradientType(ctx.agg2d.LineGradientFlag())
}

// GetFillGradientBounds returns the bounds of the current fill gradient.
func (ctx *Context) GetFillGradientBounds() (d1, d2 float64) {
	return ctx.agg2d.FillGradientD1(), ctx.agg2d.FillGradientD2()
}

// GetStrokeGradientBounds returns the bounds of the current stroke gradient.
func (ctx *Context) GetStrokeGradientBounds() (d1, d2 float64) {
	return ctx.agg2d.LineGradientD1(), ctx.agg2d.LineGradientD2()
}

// Convenience gradient constructors

// SimpleLinearGradient creates a simple two-color linear gradient from top to bottom.
func SimpleLinearGradient(topColor, bottomColor Color, height float64) *LinearGradientSpec {
	lg := CreateLinearGradientSpec(0, 0, 0, height)
	lg.AddStop(0.0, topColor)
	lg.AddStop(1.0, bottomColor)
	return lg
}

// SimpleRadialGradient creates a simple two-color radial gradient.
func SimpleRadialGradient(centerColor, edgeColor Color, cx, cy, radius float64) *RadialGradientSpec {
	rg := CreateRadialGradientSpec(cx, cy, radius)
	rg.AddStop(0.0, centerColor)
	rg.AddStop(1.0, edgeColor)
	return rg
}

// ThreeColorLinearGradient creates a three-color linear gradient.
func ThreeColorLinearGradient(startColor, middleColor, endColor Color, x1, y1, x2, y2 float64) *LinearGradientSpec {
	lg := CreateLinearGradientSpec(x1, y1, x2, y2)
	lg.AddStop(0.0, startColor)
	lg.AddStop(0.5, middleColor)
	lg.AddStop(1.0, endColor)
	return lg
}

// ThreeColorRadialGradient creates a three-color radial gradient.
func ThreeColorRadialGradient(innerColor, middleColor, outerColor Color, cx, cy, radius float64) *RadialGradientSpec {
	rg := CreateRadialGradientSpec(cx, cy, radius)
	rg.AddStop(0.0, innerColor)
	rg.AddStop(0.5, middleColor)
	rg.AddStop(1.0, outerColor)
	return rg
}

// Helper function to interpolate between colors based on position
func interpolateColor(c1, c2 Color, position float64) Color {
	return c1.Gradient(c2, position)
}

// Helper function to calculate gradient angle
func calculateGradientAngle(x1, y1, x2, y2 float64) float64 {
	return math.Atan2(y2-y1, x2-x1)
}

// Helper function to calculate gradient distance
func calculateGradientDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
