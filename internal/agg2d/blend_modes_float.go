// Package agg2d float blend modes. Float twin of blend_modes.go: maps the
// public AGG2D BlendMode onto the float composite pixfmts (PixFmtCompositeRGBA128
// / ...Pre) and exposes the blend-mode/image-blend state setters for Agg2DFloat.
package agg2d

// SetBlendMode sets the general blending mode and reconfigures the composite
// pixfmts. Mirrors Agg2D.SetBlendMode.
func (a *Agg2DFloat) SetBlendMode(mode BlendMode) {
	a.blendMode = mode
	a.updateBlendMode()
}

// GetBlendMode returns the current general blending mode.
func (a *Agg2DFloat) GetBlendMode() BlendMode {
	return a.blendMode
}

// SetImageBlendMode sets the image blending mode. Mirrors Agg2D.SetImageBlendMode.
func (a *Agg2DFloat) SetImageBlendMode(mode BlendMode) {
	a.imageBlendMode = mode
}

// GetImageBlendMode returns the current image blending mode.
func (a *Agg2DFloat) GetImageBlendMode() BlendMode {
	return a.imageBlendMode
}

// SetImageBlendColor sets the image blend color.
func (a *Agg2DFloat) SetImageBlendColor(c Color) {
	a.imageBlendColor = c
}

// SetImageBlendColorRGBA sets the image blend color from RGBA components.
func (a *Agg2DFloat) SetImageBlendColorRGBA(r, g, b, alpha uint8) {
	a.imageBlendColor = Color{r, g, b, alpha}
}

// GetImageBlendColor returns the current image blend color.
func (a *Agg2DFloat) GetImageBlendColor() Color {
	return a.imageBlendColor
}

// updateBlendMode applies the current blend mode to the float composite pixfmts.
// The Comp pixfmt drives solid/gradient fills and strokes; the CompPre pixfmt
// drives premultiplied image transfers.
func (a *Agg2DFloat) updateBlendMode() {
	if a.pixfmtComp == nil {
		return
	}
	compOp := blendModeToCompOp(a.blendMode)
	a.pixfmtComp.SetCompOp(compOp)
	if a.pixfmtCompPre != nil {
		a.pixfmtCompPre.SetCompOp(compOp)
	}
}
