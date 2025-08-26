// Package agg provides stroke attribute functionality for 2D graphics.
// This file exposes stroke width, caps/joins, dashes, and related utilities
// by wiring the public Context API to internal/agg2d.
package agg

import (
	ia "agg_go/internal/agg2d"
)

// Stroke properties

// SetStrokeWidth sets the width of strokes (alias for LineWidth on Context).
func (ctx *Context) SetStrokeWidth(width float64) { ctx.agg2d.LineWidth(width) }

// GetLineWidth returns the current line width.
func (ctx *Context) GetLineWidth() float64 { return ctx.agg2d.impl.GetLineWidth() }

// SetLineCap sets the line cap style.
func (ctx *Context) SetLineCap(cap LineCap) { ctx.agg2d.LineCap(cap) }

// GetLineCap returns the current line cap style.
func (ctx *Context) GetLineCap() LineCap { return LineCap(ctx.agg2d.impl.GetLineCap()) }

// SetLineJoin sets the line join style.
func (ctx *Context) SetLineJoin(join LineJoin) { ctx.agg2d.LineJoin(join) }

// GetLineJoin returns the current line join style.
func (ctx *Context) GetLineJoin() LineJoin { return LineJoin(ctx.agg2d.impl.GetLineJoin()) }

// Miter limit controls

// SetMiterLimit sets the miter limit for line joins.
func (ctx *Context) SetMiterLimit(limit float64) { ctx.agg2d.impl.MiterLimit(limit) }

// GetMiterLimit returns the current miter limit.
func (ctx *Context) GetMiterLimit() float64 { return ctx.agg2d.impl.GetMiterLimit() }

// SetMiterLimitAngle sets the miter limit from an angle in radians.
func (ctx *Context) SetMiterLimitAngle(angle float64) { ctx.agg2d.impl.MiterLimitTheta(angle) }

// SetMiterLimitDegrees sets the miter limit from an angle in degrees.
func (ctx *Context) SetMiterLimitDegrees(degrees float64) {
	ctx.agg2d.impl.MiterLimitTheta(degrees * 3.14159265359 / 180.0)
}

// Inner miter controls (for inner corners)

// SetInnerMiterLimit sets the inner miter limit.
func (ctx *Context) SetInnerMiterLimit(limit float64) { ctx.agg2d.impl.InnerMiterLimit(limit) }

// GetInnerMiterLimit returns the current inner miter limit.
func (ctx *Context) GetInnerMiterLimit() float64 { return ctx.agg2d.impl.GetInnerMiterLimit() }

// Dash patterns

// AddDash adds a dash pattern to the stroke.
func (ctx *Context) AddDash(dashLength, gapLength float64) {
	ctx.agg2d.impl.AddDash(dashLength, gapLength)
}

// SetDashPattern sets a complete dash pattern.
func (ctx *Context) SetDashPattern(pattern []float64) {
	ctx.ClearDashes()
	for i := 0; i < len(pattern)-1; i += 2 {
		dashLen := pattern[i]
		gapLen := dashLen
		if i+1 < len(pattern) {
			gapLen = pattern[i+1]
		}
		ctx.agg2d.impl.AddDash(dashLen, gapLen)
	}
}

// ClearDashes removes all dash patterns, returning to solid lines.
func (ctx *Context) ClearDashes() { ctx.agg2d.impl.RemoveAllDashes() }

// SetDashOffset sets the starting offset for dash patterns.
func (ctx *Context) SetDashOffset(offset float64) { ctx.agg2d.impl.DashStart(offset) }

// GetDashOffset returns the current dash offset.
func (ctx *Context) GetDashOffset() float64 { return ctx.agg2d.impl.GetDashStart() }

// Path shortening

// SetPathShorten sets the path shortening distance.
func (ctx *Context) SetPathShorten(distance float64) { ctx.agg2d.impl.Shorten(distance) }

// GetPathShorten returns the current path shortening distance.
func (ctx *Context) GetPathShorten() float64 { return ctx.agg2d.impl.GetShorten() }

// Approximation scale (affects curve quality)

// SetApproximationScale sets the approximation scale for curves.
func (ctx *Context) SetApproximationScale(scale float64) { ctx.agg2d.impl.ApproximationScale(scale) }

// GetApproximationScale returns the current approximation scale.
func (ctx *Context) GetApproximationScale() float64 { return ctx.agg2d.impl.GetApproximationScale() }

// Convenience methods for common stroke styles

// SetStrokeStyle sets multiple stroke properties at once.
func (ctx *Context) SetStrokeStyle(width float64, cap LineCap, join LineJoin) {
	ctx.SetLineWidth(width)
	ctx.SetLineCap(cap)
	ctx.SetLineJoin(join)
}

// SetDashedStroke sets up a basic dashed line pattern.
func (ctx *Context) SetDashedStroke(width, dashLength, gapLength float64) {
	ctx.SetLineWidth(width)
	ctx.ClearDashes()
	ctx.AddDash(dashLength, gapLength)
}

// SetDottedStroke sets up a dotted line pattern.
func (ctx *Context) SetDottedStroke(width, dotSize, spacing float64) {
	ctx.SetLineWidth(width)
	ctx.SetLineCap(CapRound)
	ctx.ClearDashes()
	ctx.AddDash(dotSize, spacing)
}

// SetSolidStroke sets up a solid line with specified width.
func (ctx *Context) SetSolidStroke(width float64) {
	ctx.SetLineWidth(width)
	ctx.ClearDashes()
}

// Predefined dash patterns

func (ctx *Context) SetDashPatternSmall()  { ctx.SetDashPattern([]float64{3, 3}) }
func (ctx *Context) SetDashPatternMedium() { ctx.SetDashPattern([]float64{6, 6}) }
func (ctx *Context) SetDashPatternLarge()  { ctx.SetDashPattern([]float64{12, 12}) }

// SetDashPatternDotted sets a dotted pattern.
func (ctx *Context) SetDashPatternDotted() {
	ctx.SetLineCap(CapRound)
	ctx.SetDashPattern([]float64{1, 3})
}

// SetDashPatternDashDot sets a dash-dot pattern.
func (ctx *Context) SetDashPatternDashDot() { ctx.SetDashPattern([]float64{8, 3, 1, 3}) }

// SetDashPatternDashDotDot sets a dash-dot-dot pattern.
func (ctx *Context) SetDashPatternDashDotDot() { ctx.SetDashPattern([]float64{8, 3, 1, 3, 1, 3}) }

// Stroke attributes bundling

// ContextStrokeAttributes represents a complete set of stroke attributes for Context.
type ContextStrokeAttributes struct {
	Width           float64
	Cap             LineCap
	Join            LineJoin
	MiterLimit      float64
	InnerMiterLimit float64
	DashPattern     []float64
	DashOffset      float64
	PathShorten     float64
	Approximation   float64
}

// GetContextStrokeAttributes returns the current stroke attributes.
func (ctx *Context) GetContextStrokeAttributes() ContextStrokeAttributes {
	attrs := ctx.agg2d.impl.GetStrokeAttributes()
	return ContextStrokeAttributes{
		Width:           attrs.Width,
		Cap:             LineCap(attrs.Cap),
		Join:            LineJoin(attrs.Join),
		MiterLimit:      attrs.MiterLimit,
		InnerMiterLimit: attrs.InnerMiterLimit,
		DashPattern:     attrs.DashPattern,
		DashOffset:      attrs.DashOffset,
		PathShorten:     attrs.PathShorten,
		Approximation:   attrs.ApproximationScale,
	}
}

// SetContextStrokeAttributes sets multiple stroke attributes at once.
func (ctx *Context) SetContextStrokeAttributes(attrs ContextStrokeAttributes) {
	iaAttrs := ia.StrokeAttributes{
		Width:              attrs.Width,
		MiterLimit:         attrs.MiterLimit,
		InnerMiterLimit:    attrs.InnerMiterLimit,
		Cap:                ia.LineCap(attrs.Cap),
		Join:               ia.LineJoin(attrs.Join),
		DashStart:          attrs.DashOffset,
		DashPattern:        attrs.DashPattern,
		DashOffset:         attrs.DashOffset,
		PathShorten:        attrs.PathShorten,
		Shorten:            attrs.PathShorten,
		ApproximationScale: attrs.Approximation,
	}
	ctx.agg2d.impl.SetStrokeAttributes(iaAttrs)
}

// Save/Restore helpers for stroke attributes
func (ctx *Context) SaveContextStrokeAttributes() ContextStrokeAttributes {
	return ctx.GetContextStrokeAttributes()
}

func (ctx *Context) RestoreContextStrokeAttributes(attrs ContextStrokeAttributes) {
	ctx.SetContextStrokeAttributes(attrs)
}

// Reset stroke attributes to defaults
func (ctx *Context) ResetStrokeAttributes() {
	ctx.SetLineWidth(1.0)
	ctx.SetLineCap(CapButt)
	ctx.SetLineJoin(JoinMiter)
	ctx.SetMiterLimit(4.0)
	ctx.SetInnerMiterLimit(1.01)
	ctx.ClearDashes()
	ctx.SetDashOffset(0.0)
	ctx.SetPathShorten(0.0)
	ctx.SetApproximationScale(1.0)
}

// Specialized stroke effects

// SetHairlineStroke sets an ultra-thin stroke (often rendered as 1px width).
func (ctx *Context) SetHairlineStroke() { ctx.SetLineWidth(0.0) }

// SetThickStroke sets a thick stroke with round caps and joins.
func (ctx *Context) SetThickStroke(width float64) {
	ctx.SetLineWidth(width)
	ctx.SetLineCap(CapRound)
	ctx.SetLineJoin(JoinRound)
}

// SetPenStroke mimics a pen-like stroke.
func (ctx *Context) SetPenStroke(width float64) {
	ctx.SetLineWidth(width)
	ctx.SetLineCap(CapRound)
	ctx.SetLineJoin(JoinRound)
	ctx.SetMiterLimit(2.0)
}

// SetTechnicalDrawingStroke sets attributes suitable for technical drawings.
func (ctx *Context) SetTechnicalDrawingStroke(width float64) {
	ctx.SetLineWidth(width)
	ctx.SetLineCap(CapButt)
	ctx.SetLineJoin(JoinMiter)
	ctx.SetMiterLimit(10.0)
}
