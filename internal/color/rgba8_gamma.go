// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

// Apply 8-bit gamma (RGB only) to an RGBA8 pixel in-place.
func ApplyGammaDir8[CS Space, LUT lut8Like](px *RGBA8[CS], lut LUT) {
	px.R = lut.Dir(px.R)
	px.G = lut.Dir(px.G)
	px.B = lut.Dir(px.B)
}

func ApplyGammaInv8[CS Space, LUT lut8Like](px *RGBA8[CS], lut LUT) {
	px.R = lut.Inv(px.R)
	px.G = lut.Inv(px.G)
	px.B = lut.Inv(px.B)
}

// Helper for method receivers:
func (c *RGBA8[CS]) ApplyGammaDir(lut lut8Like) { ApplyGammaDir8(c, lut) }
func (c *RGBA8[CS]) ApplyGammaInv(lut lut8Like) { ApplyGammaInv8(c, lut) }
