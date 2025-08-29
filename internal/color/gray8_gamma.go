package color

// Apply 8-bit gamma (V channel only) to a Gray8 pixel in-place.
func ApplyGammaDir8Gray[CS ColorSpace, LUT lut8Like](px *Gray8[CS], lut LUT) {
	px.V = lut.Dir(px.V)
}

func ApplyGammaInv8Gray[CS ColorSpace, LUT lut8Like](px *Gray8[CS], lut LUT) {
	px.V = lut.Inv(px.V)
}

// Helper methods for Gray8
func (g *Gray8[CS]) ApplyGammaDir(lut lut8Like) { ApplyGammaDir8Gray(g, lut) }
func (g *Gray8[CS]) ApplyGammaInv(lut lut8Like) { ApplyGammaInv8Gray(g, lut) }
