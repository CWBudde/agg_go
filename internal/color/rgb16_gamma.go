package color

import (
	"agg_go/internal/basics"
)

func ApplyGammaDir16RGB[CS Space, LUT lut16Like](px *RGB16[CS], lut LUT) {
	px.R = lut.Dir(basics.Int8u(px.R >> 8))
	px.G = lut.Dir(basics.Int8u(px.G >> 8))
	px.B = lut.Dir(basics.Int8u(px.B >> 8))
}

func ApplyGammaInv16RGB[CS Space, LUT lut16Like](px *RGB16[CS], lut LUT) {
	r8 := lut.Inv(px.R)
	g8 := lut.Inv(px.G)
	b8 := lut.Inv(px.B)
	px.R = basics.Int16u(r8)<<8 | basics.Int16u(r8)
	px.G = basics.Int16u(g8)<<8 | basics.Int16u(g8)
	px.B = basics.Int16u(b8)<<8 | basics.Int16u(b8)
}

func ApplyGammaDir16RGBUsing8[CS Space, LUT lut8Like](px *RGB16[CS], lut LUT) {
	r8 := basics.Int8u(px.R >> 8)
	g8 := basics.Int8u(px.G >> 8)
	b8 := basics.Int8u(px.B >> 8)
	r8 = lut.Dir(r8)
	g8 = lut.Dir(g8)
	b8 = lut.Dir(b8)
	px.R = basics.Int16u(r8)<<8 | basics.Int16u(r8)
	px.G = basics.Int16u(g8)<<8 | basics.Int16u(g8)
	px.B = basics.Int16u(b8)<<8 | basics.Int16u(b8)
}

func ApplyGammaInv16RGBUsing8[CS Space, LUT lut8Like](px *RGB16[CS], lut LUT) {
	r8 := basics.Int8u(px.R >> 8)
	g8 := basics.Int8u(px.G >> 8)
	b8 := basics.Int8u(px.B >> 8)
	r8 = lut.Inv(r8)
	g8 = lut.Inv(g8)
	b8 = lut.Inv(b8)
	px.R = basics.Int16u(r8)<<8 | basics.Int16u(r8)
	px.G = basics.Int16u(g8)<<8 | basics.Int16u(g8)
	px.B = basics.Int16u(b8)<<8 | basics.Int16u(b8)
}

// Helper methods for RGB16
func (c *RGB16[CS]) ApplyGammaDir(lut lut16Like) { ApplyGammaDir16RGB(c, lut) }
func (c *RGB16[CS]) ApplyGammaInv(lut lut16Like) { ApplyGammaInv16RGB(c, lut) }

func (c *RGB16[CS]) ApplyGammaDirUsing8(lut lut8Like) { ApplyGammaDir16RGBUsing8(c, lut) }
func (c *RGB16[CS]) ApplyGammaInvUsing8(lut lut8Like) { ApplyGammaInv16RGBUsing8(c, lut) }