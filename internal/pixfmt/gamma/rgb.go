package gamma

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
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
		order := blender.GetRGBColorOrder[O]()
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
		order := blender.GetRGBColorOrder[O]()
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
	PixFmtRGB24Gamma       = PixFmtRGBGamma[*pixfmt.PixFmtRGB24, *SimpleGammaLut]
	PixFmtRGB24GammaLinear = PixFmtRGBGamma[*pixfmt.PixFmtRGB24, *LinearGammaLut]
	PixFmtBGR24Gamma       = PixFmtRGBGamma[*pixfmt.PixFmtBGR24, *SimpleGammaLut]
	PixFmtBGR24GammaLinear = PixFmtRGBGamma[*pixfmt.PixFmtBGR24, *LinearGammaLut]

	PixFmtSRGB24Gamma       = PixFmtRGBGamma[*pixfmt.PixFmtSRGB24, *SimpleGammaLut]
	PixFmtSRGB24GammaLinear = PixFmtRGBGamma[*pixfmt.PixFmtSRGB24, *LinearGammaLut]
	PixFmtSBGR24Gamma       = PixFmtRGBGamma[*pixfmt.PixFmtSBGR24, *SimpleGammaLut]
	PixFmtSBGR24GammaLinear = PixFmtRGBGamma[*pixfmt.PixFmtSBGR24, *LinearGammaLut]

	// RGB48 gamma variants
	PixFmtRGB48Gamma       = PixFmtRGBGamma[pixfmt.PixFmtRGB48Linear, *SimpleGammaLut]
	PixFmtRGB48GammaLinear = PixFmtRGBGamma[pixfmt.PixFmtRGB48Linear, *LinearGammaLut]
	PixFmtBGR48Gamma       = PixFmtRGBGamma[pixfmt.PixFmtBGR48Linear, *SimpleGammaLut]
	PixFmtBGR48GammaLinear = PixFmtRGBGamma[pixfmt.PixFmtBGR48Linear, *LinearGammaLut]
)

// Constructor functions for gamma-corrected RGB24 formats
func NewPixFmtRGB24Gamma[PF any](pixfmt PF, gamma float64) *PixFmtRGBGamma[PF, *SimpleGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtRGB24GammaLinear[PF any](pixfmt PF) *PixFmtRGBGamma[PF, *LinearGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewLinearGammaLut())
}

func NewPixFmtBGR24Gamma[PF any](pixfmt PF, gamma float64) *PixFmtRGBGamma[PF, *SimpleGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtBGR24GammaLinear[PF any](pixfmt PF) *PixFmtRGBGamma[PF, *LinearGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewLinearGammaLut())
}

func NewPixFmtSRGB24Gamma[PF any](pixfmt PF, gamma float64) *PixFmtRGBGamma[PF, *SimpleGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtSRGB24GammaLinear[PF any](pixfmt PF) *PixFmtRGBGamma[PF, *LinearGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewLinearGammaLut())
}

func NewPixFmtSBGR24Gamma[PF any](pixfmt PF, gamma float64) *PixFmtRGBGamma[PF, *SimpleGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtSBGR24GammaLinear[PF any](pixfmt PF) *PixFmtRGBGamma[PF, *LinearGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewLinearGammaLut())
}

// Constructor functions for gamma-corrected RGB48 formats
func NewPixFmtRGB48Gamma[PF any](pixfmt PF, gamma float64) *PixFmtRGBGamma[PF, *SimpleGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtRGB48GammaLinear[PF any](pixfmt PF) *PixFmtRGBGamma[PF, *LinearGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewLinearGammaLut())
}

func NewPixFmtBGR48Gamma[PF any](pixfmt PF, gamma float64) *PixFmtRGBGamma[PF, *SimpleGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewSimpleGammaLut(gamma))
}

func NewPixFmtBGR48GammaLinear[PF any](pixfmt PF) *PixFmtRGBGamma[PF, *LinearGammaLut] {
	return NewPixFmtRGBGamma[PF](pixfmt, NewLinearGammaLut())
}
