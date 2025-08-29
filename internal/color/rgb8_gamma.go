package color

// Apply 8-bit gamma (RGB only) to an RGB8 pixel in-place.
func ApplyGammaDir8RGB[CS Space, LUT lut8Like](px *RGB8[CS], lut LUT) {
	px.R = lut.Dir(px.R)
	px.G = lut.Dir(px.G)
	px.B = lut.Dir(px.B)
}

func ApplyGammaInv8RGB[CS Space, LUT lut8Like](px *RGB8[CS], lut LUT) {
	px.R = lut.Inv(px.R)
	px.G = lut.Inv(px.G)
	px.B = lut.Inv(px.B)
}

// Helper methods for RGB8
func (c *RGB8[CS]) ApplyGammaDir(lut lut8Like) { ApplyGammaDir8RGB(c, lut) }
func (c *RGB8[CS]) ApplyGammaInv(lut lut8Like) { ApplyGammaInv8RGB(c, lut) }