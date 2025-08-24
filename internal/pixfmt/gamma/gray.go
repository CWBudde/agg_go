package gamma

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
)

// GammaLut represents a gamma lookup table interface
type GammaLut interface {
	Dir(v basics.Int8u) basics.Int8u // Direct gamma correction
	Inv(v basics.Int8u) basics.Int8u // Inverse gamma correction
}

// ApplyGammaDirGray applies direct gamma correction to grayscale pixels
type ApplyGammaDirGray[C any] struct {
	gamma GammaLut
}

// NewApplyGammaDirGray creates a new direct gamma applicator
func NewApplyGammaDirGray[C any](gamma GammaLut) *ApplyGammaDirGray[C] {
	return &ApplyGammaDirGray[C]{gamma: gamma}
}

// Apply applies direct gamma correction to a pixel
func (a *ApplyGammaDirGray[C]) Apply(p *basics.Int8u) {
	*p = a.gamma.Dir(*p)
}

// ApplyGammaInvGray applies inverse gamma correction to grayscale pixels
type ApplyGammaInvGray[C any] struct {
	gamma GammaLut
}

// NewApplyGammaInvGray creates a new inverse gamma applicator
func NewApplyGammaInvGray[C any](gamma GammaLut) *ApplyGammaInvGray[C] {
	return &ApplyGammaInvGray[C]{gamma: gamma}
}

// Apply applies inverse gamma correction to a pixel
func (a *ApplyGammaInvGray[C]) Apply(p *basics.Int8u) {
	*p = a.gamma.Inv(*p)
}

// SimpleGammaLut provides a simple gamma lookup table implementation
type SimpleGammaLut struct {
	dirTable [256]basics.Int8u
	invTable [256]basics.Int8u
}

// NewSimpleGammaLut creates a new simple gamma lookup table
func NewSimpleGammaLut(gamma float64) *SimpleGammaLut {
	lut := &SimpleGammaLut{}
	lut.buildTables(gamma)
	return lut
}

// buildTables builds the gamma correction lookup tables
func (lut *SimpleGammaLut) buildTables(gamma float64) {
	if gamma <= 0 {
		gamma = 1.0
	}

	invGamma := 1.0 / gamma

	for i := 0; i < 256; i++ {
		// Normalize to 0-1 range
		v := float64(i) / 255.0

		// Apply gamma correction using proper power function
		corrected := math.Pow(v, gamma)

		// For inverse gamma, apply the inverse power
		invCorrected := math.Pow(v, invGamma)

		// Clamp and convert back to 8-bit
		if corrected > 1.0 {
			corrected = 1.0
		}
		if corrected < 0.0 {
			corrected = 0.0
		}

		if invCorrected > 1.0 {
			invCorrected = 1.0
		}
		if invCorrected < 0.0 {
			invCorrected = 0.0
		}

		lut.dirTable[i] = basics.Int8u(corrected*255.0 + 0.5)
		lut.invTable[i] = basics.Int8u(invCorrected*255.0 + 0.5)
	}
}

// Dir returns the direct gamma corrected value
func (lut *SimpleGammaLut) Dir(v basics.Int8u) basics.Int8u {
	return lut.dirTable[v]
}

// Inv returns the inverse gamma corrected value
func (lut *SimpleGammaLut) Inv(v basics.Int8u) basics.Int8u {
	return lut.invTable[v]
}

// LinearGammaLut provides a linear (no-op) gamma correction
type LinearGammaLut struct{}

// NewLinearGammaLut creates a new linear gamma lookup table
func NewLinearGammaLut() *LinearGammaLut {
	return &LinearGammaLut{}
}

// Dir returns the value unchanged (linear)
func (lut *LinearGammaLut) Dir(v basics.Int8u) basics.Int8u {
	return v
}

// Inv returns the value unchanged (linear)
func (lut *LinearGammaLut) Inv(v basics.Int8u) basics.Int8u {
	return v
}

// PixFmtGrayGamma wraps a grayscale pixel format with gamma correction
type PixFmtGrayGamma[PF interface {
	Width() int
	Height() int
	PixWidth() int
}, G any] struct {
	pixfmt PF
	gamma  G
}

// NewPixFmtGrayGamma creates a new gamma-corrected grayscale pixel format
func NewPixFmtGrayGamma[PF interface {
	Width() int
	Height() int
	PixWidth() int
}, G any](pixfmt PF, gamma G) *PixFmtGrayGamma[PF, G] {
	return &PixFmtGrayGamma[PF, G]{
		pixfmt: pixfmt,
		gamma:  gamma,
	}
}

// Width returns the buffer width
func (pf *PixFmtGrayGamma[PF, G]) Width() int {
	return pf.pixfmt.Width()
}

// Height returns the buffer height
func (pf *PixFmtGrayGamma[PF, G]) Height() int {
	return pf.pixfmt.Height()
}

// PixWidth returns bytes per pixel
func (pf *PixFmtGrayGamma[PF, G]) PixWidth() int {
	return pf.pixfmt.PixWidth()
}

// Gamma correction pixel format types
type (
	PixFmtGray8Gamma  = PixFmtGrayGamma[*pixfmt.PixFmtAlphaBlendGray[blender.BlenderGray8, any], *SimpleGammaLut]
	PixFmtGray8Linear = PixFmtGrayGamma[*pixfmt.PixFmtAlphaBlendGray[blender.BlenderGray8, any], *LinearGammaLut]
)

// Constructor functions for gamma-corrected formats
func NewPixFmtGray8Gamma[PF interface {
	Width() int
	Height() int
	PixWidth() int
}](pixfmt PF, gamma float64) *PixFmtGrayGamma[PF, *SimpleGammaLut] {
	return NewPixFmtGrayGamma[PF](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtGray8Linear[PF interface {
	Width() int
	Height() int
	PixWidth() int
}](pixfmt PF) *PixFmtGrayGamma[PF, *LinearGammaLut] {
	return NewPixFmtGrayGamma[PF](pixfmt, NewLinearGammaLut())
}
