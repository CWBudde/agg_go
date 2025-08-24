package gamma

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
)

// ApplyGammaDirRGBA applies direct gamma correction to RGBA pixels
type ApplyGammaDirRGBA[C any] struct {
	gamma GammaLut
}

// NewApplyGammaDirRGBA creates a new direct gamma applicator for RGBA
func NewApplyGammaDirRGBA[C any](gamma GammaLut) *ApplyGammaDirRGBA[C] {
	return &ApplyGammaDirRGBA[C]{gamma: gamma}
}

// Apply applies direct gamma correction to an RGBA pixel (RGB components only)
func (a *ApplyGammaDirRGBA[C]) Apply(p []basics.Int8u) {
	if len(p) >= 4 {
		// Apply gamma to RGB, leave A unchanged
		p[0] = a.gamma.Dir(p[0]) // R
		p[1] = a.gamma.Dir(p[1]) // G
		p[2] = a.gamma.Dir(p[2]) // B
		// p[3] (Alpha) remains unchanged
	}
}

// ApplyGammaInvRGBA applies inverse gamma correction to RGBA pixels
type ApplyGammaInvRGBA[C any] struct {
	gamma GammaLut
}

// NewApplyGammaInvRGBA creates a new inverse gamma applicator for RGBA
func NewApplyGammaInvRGBA[C any](gamma GammaLut) *ApplyGammaInvRGBA[C] {
	return &ApplyGammaInvRGBA[C]{gamma: gamma}
}

// Apply applies inverse gamma correction to an RGBA pixel (RGB components only)
func (a *ApplyGammaInvRGBA[C]) Apply(p []basics.Int8u) {
	if len(p) >= 4 {
		// Apply inverse gamma to RGB, leave A unchanged
		p[0] = a.gamma.Inv(p[0]) // R
		p[1] = a.gamma.Inv(p[1]) // G
		p[2] = a.gamma.Inv(p[2]) // B
		// p[3] (Alpha) remains unchanged
	}
}

// PixFmtRGBAGamma wraps an RGBA pixel format with gamma correction
type PixFmtRGBAGamma[PF interface {
	Width() int
	Height() int
	PixWidth() int
}, G any] struct {
	pixfmt PF
	gamma  G
}

// NewPixFmtRGBAGamma creates a new gamma-corrected RGBA pixel format
func NewPixFmtRGBAGamma[PF interface {
	Width() int
	Height() int
	PixWidth() int
}, G any](pixfmt PF, gamma G) *PixFmtRGBAGamma[PF, G] {
	return &PixFmtRGBAGamma[PF, G]{
		pixfmt: pixfmt,
		gamma:  gamma,
	}
}

// Width returns the buffer width
func (pf *PixFmtRGBAGamma[PF, G]) Width() int {
	return pf.pixfmt.Width()
}

// Height returns the buffer height
func (pf *PixFmtRGBAGamma[PF, G]) Height() int {
	return pf.pixfmt.Height()
}

// PixWidth returns bytes per pixel
func (pf *PixFmtRGBAGamma[PF, G]) PixWidth() int {
	return pf.pixfmt.PixWidth()
}

// RGBA Gamma correction pixel format types
type (
	PixFmtRGBA32Gamma  = PixFmtRGBAGamma[*pixfmt.PixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear], *SimpleGammaLut]
	PixFmtRGBA32Linear = PixFmtRGBAGamma[*pixfmt.PixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear], *LinearGammaLut]
)

// Constructor functions for gamma-corrected RGBA formats
func NewPixFmtRGBA32Gamma(pf *pixfmt.PixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear], gamma float64) *PixFmtRGBA32Gamma {
	return NewPixFmtRGBAGamma[*pixfmt.PixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear]](pf, NewSimpleGammaLut(gamma))
}

func NewPixFmtRGBA32Linear(pf *pixfmt.PixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear]) *PixFmtRGBA32Linear {
	return NewPixFmtRGBAGamma[*pixfmt.PixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear]](pf, NewLinearGammaLut())
}

// RGBA multiplier for premultiplication/demultiplication with different component orders
type RGBAMultiplier[O any] struct{}

// Premultiply premultiplies RGBA pixel components by alpha
func (m RGBAMultiplier[O]) Premultiply(p []basics.Int8u) {
	if len(p) >= 4 {
		order := blender.GetColorOrder[O]()
		a := p[order.A]
		if a < 255 {
			if a == 0 {
				p[order.R] = 0
				p[order.G] = 0
				p[order.B] = 0
			} else {
				p[order.R] = basics.Int8u((uint32(p[order.R]) * uint32(a)) / 255)
				p[order.G] = basics.Int8u((uint32(p[order.G]) * uint32(a)) / 255)
				p[order.B] = basics.Int8u((uint32(p[order.B]) * uint32(a)) / 255)
			}
		}
	}
}

// Demultiply demultiplies RGBA pixel components by alpha
func (m RGBAMultiplier[O]) Demultiply(p []basics.Int8u) {
	if len(p) >= 4 {
		order := blender.GetColorOrder[O]()
		a := p[order.A]
		if a < 255 && a > 0 {
			p[order.R] = basics.Int8u((uint32(p[order.R])*255 + uint32(a)/2) / uint32(a))
			p[order.G] = basics.Int8u((uint32(p[order.G])*255 + uint32(a)/2) / uint32(a))
			p[order.B] = basics.Int8u((uint32(p[order.B])*255 + uint32(a)/2) / uint32(a))
		}
	}
}

// Concrete multiplier types for different color orders
type (
	RGBAMultiplierRGBA = RGBAMultiplier[blender.RGBAOrder]
	RGBAMultiplierARGB = RGBAMultiplier[blender.ARGBOrder]
	RGBAMultiplierBGRA = RGBAMultiplier[blender.BGRAOrder]
	RGBAMultiplierABGR = RGBAMultiplier[blender.ABGROrder]
)
