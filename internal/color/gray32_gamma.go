package color

// Apply 32-bit gamma (V channel only) to a Gray32 pixel in-place.
func ApplyGammaDir32Gray[CS Space, G lut32Like](px *Gray32[CS], g G) {
	px.V = g.DirFloat(px.V)
}

func ApplyGammaInv32Gray[CS Space, G lut32Like](px *Gray32[CS], g G) {
	px.V = g.InvFloat(px.V)
}

// Helper methods for Gray32
func (g *Gray32[CS]) ApplyGammaDir(gamma lut32Like) { ApplyGammaDir32Gray(g, gamma) }
func (g *Gray32[CS]) ApplyGammaInv(gamma lut32Like) { ApplyGammaInv32Gray(g, gamma) }
