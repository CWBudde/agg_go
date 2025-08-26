// Package agg provides blending and compositing functionality for 2D graphics.
// This file exposes blend modes and alpha/composite controls, backed by internal/agg2d.
package agg

import (
	ia "agg_go/internal/agg2d"
)

// Public blend mode type (aliases internal type so values match).
type BlendMode = ia.BlendMode

// Blend mode constants (re-exported from internal).
const (
	BlendAlpha      = ia.BlendAlpha
	BlendClear      = ia.BlendClear
	BlendSrc        = ia.BlendSrc
	BlendDst        = ia.BlendDst
	BlendSrcOver    = ia.BlendSrcOver
	BlendDstOver    = ia.BlendDstOver
	BlendSrcIn      = ia.BlendSrcIn
	BlendDstIn      = ia.BlendDstIn
	BlendSrcOut     = ia.BlendSrcOut
	BlendDstOut     = ia.BlendDstOut
	BlendSrcAtop    = ia.BlendSrcAtop
	BlendDstAtop    = ia.BlendDstAtop
	BlendXor        = ia.BlendXor
	BlendAdd        = ia.BlendAdd
	BlendMultiply   = ia.BlendMultiply
	BlendScreen     = ia.BlendScreen
	BlendOverlay    = ia.BlendOverlay
	BlendDarken     = ia.BlendDarken
	BlendLighten    = ia.BlendLighten
	BlendColorDodge = ia.BlendColorDodge
	BlendColorBurn  = ia.BlendColorBurn
	BlendHardLight  = ia.BlendHardLight
	BlendSoftLight  = ia.BlendSoftLight
	BlendDifference = ia.BlendDifference
	BlendExclusion  = ia.BlendExclusion
)

// Blend mode operations

// SetBlendMode sets the blending mode for subsequent drawing operations.
func (ctx *Context) SetBlendMode(mode BlendMode) { ctx.agg2d.impl.SetBlendMode(mode) }

// GetBlendMode returns the current blending mode.
func (ctx *Context) GetBlendMode() BlendMode { return ctx.agg2d.impl.GetBlendMode() }

// Alpha operations

// SetGlobalAlpha sets the global alpha by updating current fill/stroke colors.
// This does not modify master alpha; use SetMasterAlpha for that.
func (ctx *Context) SetGlobalAlpha(alpha float64) {
	if alpha < 0.0 {
		alpha = 0.0
	} else if alpha > 1.0 {
		alpha = 1.0
	}
	a8 := uint8(alpha * 255.0)

	// Update internal colors and reapply
	fc := ctx.agg2d.impl.GetFillColor()
	lc := ctx.agg2d.impl.GetLineColor()
	fc[3] = a8
	lc[3] = a8
	ctx.agg2d.impl.FillColor(fc)
	ctx.agg2d.impl.LineColor(lc)
}

// GetGlobalAlpha returns the current alpha from the fill color.
func (ctx *Context) GetGlobalAlpha() float64 {
	fc := ctx.agg2d.impl.GetFillColor()
	return float64(fc[3]) / 255.0
}

// Compositing operations (aliases)
func (ctx *Context) SetCompositeOperation(operation BlendMode) { ctx.SetBlendMode(operation) }
func (ctx *Context) GetCompositeOperation() BlendMode          { return ctx.GetBlendMode() }

// Master alpha (overall context multiplier)
func (ctx *Context) SetMasterAlpha(alpha float64) { ctx.agg2d.impl.SetMasterAlpha(alpha) }
func (ctx *Context) GetMasterAlpha() float64      { return ctx.agg2d.impl.GetMasterAlpha() }

// Convenience blend mode setters
func (ctx *Context) SetBlendNormal()     { ctx.SetBlendMode(BlendSrcOver) }
func (ctx *Context) SetBlendMultiply()   { ctx.SetBlendMode(BlendMultiply) }
func (ctx *Context) SetBlendScreen()     { ctx.SetBlendMode(BlendScreen) }
func (ctx *Context) SetBlendOverlay()    { ctx.SetBlendMode(BlendOverlay) }
func (ctx *Context) SetBlendDarken()     { ctx.SetBlendMode(BlendDarken) }
func (ctx *Context) SetBlendLighten()    { ctx.SetBlendMode(BlendLighten) }
func (ctx *Context) SetBlendDifference() { ctx.SetBlendMode(BlendDifference) }
func (ctx *Context) SetBlendExclusion()  { ctx.SetBlendMode(BlendExclusion) }

// Porter-Duff helpers
func (ctx *Context) SetBlendClear()  { ctx.SetBlendMode(BlendClear) }
func (ctx *Context) SetBlendSrc()    { ctx.SetBlendMode(BlendSrc) }
func (ctx *Context) SetBlendDst()    { ctx.SetBlendMode(BlendDst) }
func (ctx *Context) SetBlendSrcIn()  { ctx.SetBlendMode(BlendSrcIn) }
func (ctx *Context) SetBlendDstIn()  { ctx.SetBlendMode(BlendDstIn) }
func (ctx *Context) SetBlendSrcOut() { ctx.SetBlendMode(BlendSrcOut) }
func (ctx *Context) SetBlendDstOut() { ctx.SetBlendMode(BlendDstOut) }
func (ctx *Context) SetBlendXor()    { ctx.SetBlendMode(BlendXor) }

// Utilities

// BlendModeToString converts a blend mode to a human-readable string.
func BlendModeToString(mode BlendMode) string { return ia.BlendModeString(mode) }

// StringToBlendMode converts a string to a blend mode (best-effort).
func StringToBlendMode(s string) BlendMode {
	switch s {
	case "alpha":
		return BlendAlpha
	case "clear":
		return BlendClear
	case "src":
		return BlendSrc
	case "dst":
		return BlendDst
	case "src-over":
		return BlendSrcOver
	case "dst-over":
		return BlendDstOver
	case "src-in":
		return BlendSrcIn
	case "dst-in":
		return BlendDstIn
	case "src-out":
		return BlendSrcOut
	case "dst-out":
		return BlendDstOut
	case "src-atop":
		return BlendSrcAtop
	case "dst-atop":
		return BlendDstAtop
	case "xor":
		return BlendXor
	case "add", "plus":
		return BlendAdd
	case "multiply":
		return BlendMultiply
	case "screen":
		return BlendScreen
	case "overlay":
		return BlendOverlay
	case "darken":
		return BlendDarken
	case "lighten":
		return BlendLighten
	case "color-dodge":
		return BlendColorDodge
	case "color-burn":
		return BlendColorBurn
	case "hard-light":
		return BlendHardLight
	case "soft-light":
		return BlendSoftLight
	case "difference":
		return BlendDifference
	case "exclusion":
		return BlendExclusion
	default:
		return BlendSrcOver
	}
}
