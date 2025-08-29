package color

// Zero-overhead, monomorphized helpers (no interface at call-site):
func ApplyGammaDir32[CS Space, G lut32Like](px *RGBA32[CS], g G) {
	px.R = g.DirFloat(px.R)
	px.G = g.DirFloat(px.G)
	px.B = g.DirFloat(px.B)
}

func ApplyGammaInv32[CS Space, G lut32Like](px *RGBA32[CS], g G) {
	px.R = g.InvFloat(px.R)
	px.G = g.InvFloat(px.G)
	px.B = g.InvFloat(px.B)
}

// Helper for method receivers:
func (c *RGBA32[CS]) ApplyGammaDir(lut lut32Like) { ApplyGammaDir32(c, lut) }
func (c *RGBA32[CS]) ApplyGammaInv(lut lut32Like) { ApplyGammaInv32(c, lut) }
