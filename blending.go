package agg

import (
	ia "github.com/MeKo-Christian/agg_go/internal/agg2d"
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

// SetCompositeOperation is an alias for SetBlendMode.
func (ctx *Context) SetCompositeOperation(operation BlendMode) { ctx.SetBlendMode(operation) }

// GetCompositeOperation is an alias for GetBlendMode.
func (ctx *Context) GetCompositeOperation() BlendMode { return ctx.GetBlendMode() }

// SetMasterAlpha sets the context-wide alpha multiplier applied during rendering.
func (ctx *Context) SetMasterAlpha(alpha float64) { ctx.agg2d.impl.SetMasterAlpha(alpha) }

// GetMasterAlpha returns the context-wide alpha multiplier.
func (ctx *Context) GetMasterAlpha() float64 { return ctx.agg2d.impl.GetMasterAlpha() }

// SetBlendNormal selects the standard source-over blend mode.
func (ctx *Context) SetBlendNormal() { ctx.SetBlendMode(BlendSrcOver) }

// SetBlendMultiply selects multiply blending.
func (ctx *Context) SetBlendMultiply() { ctx.SetBlendMode(BlendMultiply) }

// SetBlendScreen selects screen blending.
func (ctx *Context) SetBlendScreen() { ctx.SetBlendMode(BlendScreen) }

// SetBlendOverlay selects overlay blending.
func (ctx *Context) SetBlendOverlay() { ctx.SetBlendMode(BlendOverlay) }

// SetBlendDarken selects darken blending.
func (ctx *Context) SetBlendDarken() { ctx.SetBlendMode(BlendDarken) }

// SetBlendLighten selects lighten blending.
func (ctx *Context) SetBlendLighten() { ctx.SetBlendMode(BlendLighten) }

// SetBlendDifference selects difference blending.
func (ctx *Context) SetBlendDifference() { ctx.SetBlendMode(BlendDifference) }

// SetBlendExclusion selects exclusion blending.
func (ctx *Context) SetBlendExclusion() { ctx.SetBlendMode(BlendExclusion) }

// SetBlendClear selects the Porter-Duff clear operator.
func (ctx *Context) SetBlendClear() { ctx.SetBlendMode(BlendClear) }

// SetBlendSrc selects the Porter-Duff src operator.
func (ctx *Context) SetBlendSrc() { ctx.SetBlendMode(BlendSrc) }

// SetBlendDst selects the Porter-Duff dst operator.
func (ctx *Context) SetBlendDst() { ctx.SetBlendMode(BlendDst) }

// SetBlendSrcIn selects the Porter-Duff src-in operator.
func (ctx *Context) SetBlendSrcIn() { ctx.SetBlendMode(BlendSrcIn) }

// SetBlendDstIn selects the Porter-Duff dst-in operator.
func (ctx *Context) SetBlendDstIn() { ctx.SetBlendMode(BlendDstIn) }

// SetBlendSrcOut selects the Porter-Duff src-out operator.
func (ctx *Context) SetBlendSrcOut() { ctx.SetBlendMode(BlendSrcOut) }

// SetBlendDstOut selects the Porter-Duff dst-out operator.
func (ctx *Context) SetBlendDstOut() { ctx.SetBlendMode(BlendDstOut) }

// SetBlendXor selects the Porter-Duff xor operator.
func (ctx *Context) SetBlendXor() { ctx.SetBlendMode(BlendXor) }

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
