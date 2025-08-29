package color

import (
	"agg_go/internal/basics"
)

// Apply 16-bit gamma (V channel only) to a Gray16 pixel in-place.
func ApplyGammaDir16Gray[CS Space, LUT lut16Like](px *Gray16[CS], lut LUT) {
	px.V = lut.Dir(basics.Int8u(px.V >> 8))
}

func ApplyGammaInv16Gray[CS Space, LUT lut16Like](px *Gray16[CS], lut LUT) {
	v8 := lut.Inv(px.V)
	px.V = basics.Int16u(v8)<<8 | basics.Int16u(v8)
}

// Apply gamma using 8-bit LUT (down-sample, apply, up-sample)
func ApplyGammaDir16GrayUsing8[CS Space, LUT lut8Like](px *Gray16[CS], lut LUT) {
	v8 := basics.Int8u(px.V >> 8)
	v8 = lut.Dir(v8)
	px.V = basics.Int16u(v8)<<8 | basics.Int16u(v8)
}

func ApplyGammaInv16GrayUsing8[CS Space, LUT lut8Like](px *Gray16[CS], lut LUT) {
	v8 := basics.Int8u(px.V >> 8)
	v8 = lut.Inv(v8)
	px.V = basics.Int16u(v8)<<8 | basics.Int16u(v8)
}

// Helper methods for Gray16
func (g *Gray16[CS]) ApplyGammaDir(lut lut16Like)      { ApplyGammaDir16Gray(g, lut) }
func (g *Gray16[CS]) ApplyGammaInv(lut lut16Like)      { ApplyGammaInv16Gray(g, lut) }
func (g *Gray16[CS]) ApplyGammaDirUsing8(lut lut8Like) { ApplyGammaDir16GrayUsing8(g, lut) }
func (g *Gray16[CS]) ApplyGammaInvUsing8(lut lut8Like) { ApplyGammaInv16GrayUsing8(g, lut) }
