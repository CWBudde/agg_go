package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// ApplyGammaDirectRGB applies direct gamma correction to RGB pixels
type ApplyGammaDirectRGB[C any, O any] struct {
	gamma GammaLut
}

// NewApplyGammaDirectRGB creates a new direct gamma applicator for RGB
func NewApplyGammaDirectRGB[C any, O any](gamma GammaLut) *ApplyGammaDirectRGB[C, O] {
	return &ApplyGammaDirectRGB[C, O]{gamma: gamma}
}

// Apply applies direct gamma correction to an RGB pixel array
func (a *ApplyGammaDirectRGB[C, O]) Apply(p []basics.Int8u) {
	if len(p) >= 3 {
		order := getRGBColorOrder[O]()
		p[order.R] = a.gamma.Dir(p[order.R])
		p[order.G] = a.gamma.Dir(p[order.G])
		p[order.B] = a.gamma.Dir(p[order.B])
	}
}

// ApplyGammaInverseRGB applies inverse gamma correction to RGB pixels
type ApplyGammaInverseRGB[C any, O any] struct {
	gamma GammaLut
}

// NewApplyGammaInverseRGB creates a new inverse gamma applicator for RGB
func NewApplyGammaInverseRGB[C any, O any](gamma GammaLut) *ApplyGammaInverseRGB[C, O] {
	return &ApplyGammaInverseRGB[C, O]{gamma: gamma}
}

// Apply applies inverse gamma correction to an RGB pixel array
func (a *ApplyGammaInverseRGB[C, O]) Apply(p []basics.Int8u) {
	if len(p) >= 3 {
		order := getRGBColorOrder[O]()
		p[order.R] = a.gamma.Inv(p[order.R])
		p[order.G] = a.gamma.Inv(p[order.G])
		p[order.B] = a.gamma.Inv(p[order.B])
	}
}

// PixFmtRGBGamma wraps an RGB pixel format with gamma correction
type PixFmtRGBGamma[PF any, G any] struct {
	pixfmt PF
	gamma  G
	dirApp ApplyGammaDirectRGB[color.Linear, color.RGB24Order]
	invApp ApplyGammaInverseRGB[color.Linear, color.RGB24Order]
}

// NewPixFmtRGBGamma creates a new gamma-corrected RGB pixel format wrapper
func NewPixFmtRGBGamma[PF any, G any](pixfmt PF, gamma G) *PixFmtRGBGamma[PF, G] {
	var gammaLut GammaLut
	if gl, ok := any(gamma).(GammaLut); ok {
		gammaLut = gl
	}
	return &PixFmtRGBGamma[PF, G]{
		pixfmt: pixfmt,
		gamma:  gamma,
		dirApp: *NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gammaLut),
		invApp: *NewApplyGammaInverseRGB[color.Linear, color.RGB24Order](gammaLut),
	}
}

// Width returns the buffer width
func (pf *PixFmtRGBGamma[PF, G]) Width() int {
	if w, ok := any(pf.pixfmt).(interface{ Width() int }); ok {
		return w.Width()
	}
	return 0
}

// Height returns the buffer height
func (pf *PixFmtRGBGamma[PF, G]) Height() int {
	if h, ok := any(pf.pixfmt).(interface{ Height() int }); ok {
		return h.Height()
	}
	return 0
}

// PixWidth returns bytes per pixel
func (pf *PixFmtRGBGamma[PF, G]) PixWidth() int {
	if pw, ok := any(pf.pixfmt).(interface{ PixWidth() int }); ok {
		return pw.PixWidth()
	}
	return 3 // Default for RGB24
}

// GetPixel returns the pixel at the given coordinates with inverse gamma applied
func (pf *PixFmtRGBGamma[PF, G]) GetPixel(x, y int) color.RGB8[color.Linear] {
	if gp, ok := any(pf.pixfmt).(interface {
		GetPixel(int, int) color.RGB8[color.Linear]
	}); ok {
		pixel := gp.GetPixel(x, y)
		// Apply inverse gamma correction to the retrieved pixel
		rgb := []basics.Int8u{pixel.R, pixel.G, pixel.B}
		pf.invApp.Apply(rgb)
		return color.RGB8[color.Linear]{R: rgb[0], G: rgb[1], B: rgb[2]}
	}
	return color.RGB8[color.Linear]{}
}

// CopyPixel copies a pixel with direct gamma applied
func (pf *PixFmtRGBGamma[PF, G]) CopyPixel(x, y int, c color.RGB8[color.Linear]) {
	if cp, ok := any(pf.pixfmt).(interface {
		CopyPixel(int, int, color.RGB8[color.Linear])
	}); ok {
		// Apply direct gamma correction before storing
		rgb := []basics.Int8u{c.R, c.G, c.B}
		pf.dirApp.Apply(rgb)
		cp.CopyPixel(x, y, color.RGB8[color.Linear]{R: rgb[0], G: rgb[1], B: rgb[2]})
	}
}

// BlendPixel blends a pixel with direct gamma applied
func (pf *PixFmtRGBGamma[PF, G]) BlendPixel(x, y int, c color.RGB8[color.Linear], alpha, cover basics.Int8u) {
	if bp, ok := any(pf.pixfmt).(interface {
		BlendPixel(int, int, color.RGB8[color.Linear], basics.Int8u, basics.Int8u)
	}); ok {
		// Apply direct gamma correction before blending
		rgb := []basics.Int8u{c.R, c.G, c.B}
		pf.dirApp.Apply(rgb)
		bp.BlendPixel(x, y, color.RGB8[color.Linear]{R: rgb[0], G: rgb[1], B: rgb[2]}, alpha, cover)
	}
}

// Clear clears the entire buffer with gamma-corrected color
func (pf *PixFmtRGBGamma[PF, G]) Clear(c color.RGB8[color.Linear]) {
	if cl, ok := any(pf.pixfmt).(interface {
		Clear(color.RGB8[color.Linear])
	}); ok {
		// Apply direct gamma correction before clearing
		rgb := []basics.Int8u{c.R, c.G, c.B}
		pf.dirApp.Apply(rgb)
		cl.Clear(color.RGB8[color.Linear]{R: rgb[0], G: rgb[1], B: rgb[2]})
	}
}

// ApplyGammaDirect applies direct gamma correction to all pixels
func (pf *PixFmtRGBGamma[PF, G]) ApplyGammaDirect() {
	if fe, ok := any(pf.pixfmt).(interface{ ForEachPixel(func([]basics.Int8u)) }); ok {
		fe.ForEachPixel(pf.dirApp.Apply)
	}
}

// ApplyGammaInverse applies inverse gamma correction to all pixels
func (pf *PixFmtRGBGamma[PF, G]) ApplyGammaInverse() {
	if fe, ok := any(pf.pixfmt).(interface{ ForEachPixel(func([]basics.Int8u)) }); ok {
		fe.ForEachPixel(pf.invApp.Apply)
	}
}

// Concrete gamma-corrected RGB pixel format types
type (
	PixFmtRGB24Gamma       = PixFmtRGBGamma[*PixFmtRGB24, *SimpleGammaLut]
	PixFmtRGB24GammaLinear = PixFmtRGBGamma[*PixFmtRGB24, *LinearGammaLut]
	PixFmtBGR24Gamma       = PixFmtRGBGamma[*PixFmtBGR24, *SimpleGammaLut]
	PixFmtBGR24GammaLinear = PixFmtRGBGamma[*PixFmtBGR24, *LinearGammaLut]

	PixFmtSRGB24Gamma       = PixFmtRGBGamma[*PixFmtSRGB24, *SimpleGammaLut]
	PixFmtSRGB24GammaLinear = PixFmtRGBGamma[*PixFmtSRGB24, *LinearGammaLut]
	PixFmtSBGR24Gamma       = PixFmtRGBGamma[*PixFmtSBGR24, *SimpleGammaLut]
	PixFmtSBGR24GammaLinear = PixFmtRGBGamma[*PixFmtSBGR24, *LinearGammaLut]

	// RGB48 gamma variants
	PixFmtRGB48Gamma       = PixFmtRGBGamma[*PixFmtRGB48Linear, *SimpleGammaLut]
	PixFmtRGB48GammaLinear = PixFmtRGBGamma[*PixFmtRGB48Linear, *LinearGammaLut]
	PixFmtBGR48Gamma       = PixFmtRGBGamma[*PixFmtBGR48Linear, *SimpleGammaLut]
	PixFmtBGR48GammaLinear = PixFmtRGBGamma[*PixFmtBGR48Linear, *LinearGammaLut]
)

// Constructor functions for gamma-corrected RGB24 formats
func NewPixFmtRGB24Gamma(pixfmt *PixFmtRGB24, gamma float64) *PixFmtRGB24Gamma {
	return NewPixFmtRGBGamma[*PixFmtRGB24](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtRGB24GammaLinear(pixfmt *PixFmtRGB24) *PixFmtRGB24GammaLinear {
	return NewPixFmtRGBGamma[*PixFmtRGB24](pixfmt, NewLinearGammaLut())
}

func NewPixFmtBGR24Gamma(pixfmt *PixFmtBGR24, gamma float64) *PixFmtBGR24Gamma {
	return NewPixFmtRGBGamma[*PixFmtBGR24](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtBGR24GammaLinear(pixfmt *PixFmtBGR24) *PixFmtBGR24GammaLinear {
	return NewPixFmtRGBGamma[*PixFmtBGR24](pixfmt, NewLinearGammaLut())
}

func NewPixFmtSRGB24Gamma(pixfmt *PixFmtSRGB24, gamma float64) *PixFmtSRGB24Gamma {
	return NewPixFmtRGBGamma[*PixFmtSRGB24](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtSRGB24GammaLinear(pixfmt *PixFmtSRGB24) *PixFmtSRGB24GammaLinear {
	return NewPixFmtRGBGamma[*PixFmtSRGB24](pixfmt, NewLinearGammaLut())
}

func NewPixFmtSBGR24Gamma(pixfmt *PixFmtSBGR24, gamma float64) *PixFmtSBGR24Gamma {
	return NewPixFmtRGBGamma[*PixFmtSBGR24](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtSBGR24GammaLinear(pixfmt *PixFmtSBGR24) *PixFmtSBGR24GammaLinear {
	return NewPixFmtRGBGamma[*PixFmtSBGR24](pixfmt, NewLinearGammaLut())
}

// Constructor functions for gamma-corrected RGB48 formats
func NewPixFmtRGB48Gamma(pixfmt *PixFmtRGB48Linear, gamma float64) *PixFmtRGB48Gamma {
	return NewPixFmtRGBGamma[*PixFmtRGB48Linear](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtRGB48GammaLinear(pixfmt *PixFmtRGB48Linear) *PixFmtRGB48GammaLinear {
	return NewPixFmtRGBGamma[*PixFmtRGB48Linear](pixfmt, NewLinearGammaLut())
}

func NewPixFmtBGR48Gamma(pixfmt *PixFmtBGR48Linear, gamma float64) *PixFmtBGR48Gamma {
	return NewPixFmtRGBGamma[*PixFmtBGR48Linear](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtBGR48GammaLinear(pixfmt *PixFmtBGR48Linear) *PixFmtBGR48GammaLinear {
	return NewPixFmtRGBGamma[*PixFmtBGR48Linear](pixfmt, NewLinearGammaLut())
}
