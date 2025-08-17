package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// RGBBlender represents the interface for RGB pixel blending operations (no alpha)
type RGBBlender interface {
	BlendPix(dst []basics.Int8u, r, g, b, alpha, cover basics.Int8u)
}

// RGBBlenderSimple represents simplified RGB blending without coverage
type RGBBlenderSimple interface {
	BlendPix(dst []basics.Int8u, r, g, b, alpha basics.Int8u)
}

// BlenderRGB implements standard RGB blending (non-premultiplied source into RGB buffer)
// For RGB formats, alpha is used as opacity but not stored in the destination
type BlenderRGB[CS any, O any] struct{}

// BlendPix blends an RGB pixel with alpha into an RGB buffer
func (bl BlenderRGB[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		order := getRGBColorOrder[O]()
		dst[order.R] = color.RGBA8Lerp(dst[order.R], r, blendAlpha)
		dst[order.G] = color.RGBA8Lerp(dst[order.G], g, blendAlpha)
		dst[order.B] = color.RGBA8Lerp(dst[order.B], b, blendAlpha)
	}
}

// BlendPixSimple blends RGB without coverage
func (bl BlenderRGB[CS, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, alpha basics.Int8u) {
	if alpha > 0 {
		order := getRGBColorOrder[O]()
		dst[order.R] = color.RGBA8Lerp(dst[order.R], r, alpha)
		dst[order.G] = color.RGBA8Lerp(dst[order.G], g, alpha)
		dst[order.B] = color.RGBA8Lerp(dst[order.B], b, alpha)
	}
}

// BlenderRGBPre implements premultiplied RGB blending
type BlenderRGBPre[CS any, O any] struct{}

// BlendPix blends a premultiplied RGB pixel into an RGB buffer
func (bl BlenderRGBPre[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, alpha, cover basics.Int8u) {
	cr := color.RGBA8MultCover(r, cover)
	cg := color.RGBA8MultCover(g, cover)
	cb := color.RGBA8MultCover(b, cover)
	ca := color.RGBA8MultCover(alpha, cover)

	order := getRGBColorOrder[O]()
	dst[order.R] = color.RGBA8Prelerp(dst[order.R], cr, ca)
	dst[order.G] = color.RGBA8Prelerp(dst[order.G], cg, ca)
	dst[order.B] = color.RGBA8Prelerp(dst[order.B], cb, ca)
}

// BlendPixSimple blends premultiplied RGB without coverage
func (bl BlenderRGBPre[CS, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, alpha basics.Int8u) {
	order := getRGBColorOrder[O]()
	dst[order.R] = color.RGBA8Prelerp(dst[order.R], r, alpha)
	dst[order.G] = color.RGBA8Prelerp(dst[order.G], g, alpha)
	dst[order.B] = color.RGBA8Prelerp(dst[order.B], b, alpha)
}

// BlenderRGBGamma implements gamma-corrected RGB blending
type BlenderRGBGamma[CS any, O any, G any] struct {
	gamma G
}

// NewBlenderRGBGamma creates a new gamma blender
func NewBlenderRGBGamma[CS any, O any, G any](gamma G) BlenderRGBGamma[CS, O, G] {
	return BlenderRGBGamma[CS, O, G]{gamma: gamma}
}

// BlendPix blends RGB with gamma correction
func (bl BlenderRGBGamma[CS, O, G]) BlendPix(dst []basics.Int8u, r, g, b, alpha, cover basics.Int8u) {
	blendAlpha := color.RGBA8MultCover(alpha, cover)
	if blendAlpha > 0 {
		order := getRGBColorOrder[O]()

		// Apply gamma correction if gamma interface is implemented
		if gamma, ok := interface{}(bl.gamma).(GammaCorrector); ok {
			dr := gamma.Dir(dst[order.R])
			dg := gamma.Dir(dst[order.G])
			db := gamma.Dir(dst[order.B])

			sr := gamma.Dir(r)
			sg := gamma.Dir(g)
			sb := gamma.Dir(b)

			dst[order.R] = gamma.Inv(color.RGBA8Lerp(dr, sr, blendAlpha))
			dst[order.G] = gamma.Inv(color.RGBA8Lerp(dg, sg, blendAlpha))
			dst[order.B] = gamma.Inv(color.RGBA8Lerp(db, sb, blendAlpha))
		} else {
			// Fallback to regular blending
			dst[order.R] = color.RGBA8Lerp(dst[order.R], r, blendAlpha)
			dst[order.G] = color.RGBA8Lerp(dst[order.G], g, blendAlpha)
			dst[order.B] = color.RGBA8Lerp(dst[order.B], b, blendAlpha)
		}
	}
}

// GammaCorrector interface for gamma correction
type GammaCorrector interface {
	Dir(v basics.Int8u) basics.Int8u // Apply gamma correction
	Inv(v basics.Int8u) basics.Int8u // Apply inverse gamma correction
}

// Helper function to get RGB color order based on type parameter
func getRGBColorOrder[O any]() color.ColorOrder {
	var order color.ColorOrder
	switch any(*new(O)).(type) {
	case color.RGB24Order:
		order = color.OrderRGB24
	case color.BGR24Order:
		order = color.OrderBGR24
	default:
		// Default to RGB order
		order = color.OrderRGB24
	}
	return order
}

// Concrete blender types for convenience
type (
	BlenderRGB24        = BlenderRGB[color.Linear, color.RGB24Order]
	BlenderRGB24SRGB    = BlenderRGB[color.SRGB, color.RGB24Order]
	BlenderRGB24Pre     = BlenderRGBPre[color.Linear, color.RGB24Order]
	BlenderRGB24PreSRGB = BlenderRGBPre[color.SRGB, color.RGB24Order]

	BlenderBGR24        = BlenderRGB[color.Linear, color.BGR24Order]
	BlenderBGR24SRGB    = BlenderRGB[color.SRGB, color.BGR24Order]
	BlenderBGR24Pre     = BlenderRGBPre[color.Linear, color.BGR24Order]
	BlenderBGR24PreSRGB = BlenderRGBPre[color.SRGB, color.BGR24Order]
)

// Helper functions for RGB pixel operations

// BlendRGBPixel blends a single RGB pixel
func BlendRGBPixel[B RGBBlender](dst []basics.Int8u, src color.RGB8[color.Linear], alpha, cover basics.Int8u, blender B) {
	blender.BlendPix(dst, src.R, src.G, src.B, alpha, cover)
}

// CopyRGBPixel copies a single RGB pixel
func CopyRGBPixel[O any](dst []basics.Int8u, src color.RGB8[color.Linear]) {
	order := getRGBColorOrder[O]()
	dst[order.R] = src.R
	dst[order.G] = src.G
	dst[order.B] = src.B
}

// BlendRGBHline blends a horizontal line of RGB pixels
func BlendRGBHline[B RGBBlender](dst []basics.Int8u, x, length int, src color.RGB8[color.Linear], alpha basics.Int8u, covers []basics.Int8u, blender B) {
	pixStep := 3 // RGB24 = 3 bytes per pixel
	dstPtr := x * pixStep

	if covers == nil {
		// Uniform coverage
		for i := 0; i < length; i++ {
			blender.BlendPix(dst[dstPtr:], src.R, src.G, src.B, alpha, 255)
			dstPtr += pixStep
		}
	} else {
		// Varying coverage
		for i := 0; i < length; i++ {
			if covers[i] > 0 {
				blender.BlendPix(dst[dstPtr:], src.R, src.G, src.B, alpha, covers[i])
			}
			dstPtr += pixStep
		}
	}
}

// CopyRGBHline copies a horizontal line of RGB pixels
func CopyRGBHline[O any](dst []basics.Int8u, x, length int, src color.RGB8[color.Linear]) {
	order := getRGBColorOrder[O]()
	pixStep := 3 // RGB24 = 3 bytes per pixel
	dstPtr := x * pixStep

	for i := 0; i < length; i++ {
		dst[dstPtr+order.R] = src.R
		dst[dstPtr+order.G] = src.G
		dst[dstPtr+order.B] = src.B
		dstPtr += pixStep
	}
}

// FillRGBSpan fills a span with a solid RGB color
func FillRGBSpan[O any](dst []basics.Int8u, x, length int, src color.RGB8[color.Linear]) {
	CopyRGBHline[O](dst, x, length, src)
}

// ConvertRGBAToRGB converts RGBA color to RGB for blending (ignores alpha)
func ConvertRGBAToRGB[CS any](rgba color.RGBA8[CS]) color.RGB8[CS] {
	return color.RGB8[CS]{R: rgba.R, G: rgba.G, B: rgba.B}
}

// ConvertRGBToRGBA converts RGB color to RGBA with full opacity
func ConvertRGBToRGBA[CS any](rgb color.RGB8[CS]) color.RGBA8[CS] {
	return color.RGBA8[CS]{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

//==============================================================================
// RGB48 (16-bit per channel) Blenders
//==============================================================================

// RGB48Blender represents the interface for 16-bit RGB pixel blending operations
type RGB48Blender interface {
	BlendPix(dst []basics.Int16u, r, g, b, alpha, cover basics.Int16u)
}

// BlenderRGB48 implements standard RGB48 blending (16-bit per channel)
type BlenderRGB48[CS any, O any] struct{}

// BlendPix blends an RGB48 pixel with alpha into an RGB48 buffer
func (bl BlenderRGB48[CS, O]) BlendPix(dst []basics.Int16u, r, g, b, alpha, cover basics.Int16u) {
	blendAlpha := color.RGB16MultCover(alpha, cover)
	if blendAlpha > 0 {
		order := getRGBColorOrder[O]()
		dst[order.R] = color.RGB16Lerp(dst[order.R], r, blendAlpha)
		dst[order.G] = color.RGB16Lerp(dst[order.G], g, blendAlpha)
		dst[order.B] = color.RGB16Lerp(dst[order.B], b, blendAlpha)
	}
}

// BlenderRGB48Pre implements premultiplied RGB48 blending
type BlenderRGB48Pre[CS any, O any] struct{}

// BlendPix blends a premultiplied RGB48 pixel into an RGB48 buffer
func (bl BlenderRGB48Pre[CS, O]) BlendPix(dst []basics.Int16u, r, g, b, alpha, cover basics.Int16u) {
	cr := color.RGB16MultCover(r, cover)
	cg := color.RGB16MultCover(g, cover)
	cb := color.RGB16MultCover(b, cover)
	ca := color.RGB16MultCover(alpha, cover)

	order := getRGBColorOrder[O]()
	dst[order.R] = color.RGB16Prelerp(dst[order.R], cr, ca)
	dst[order.G] = color.RGB16Prelerp(dst[order.G], cg, ca)
	dst[order.B] = color.RGB16Prelerp(dst[order.B], cb, ca)
}

// BlenderRGB48Gamma implements gamma-corrected RGB48 blending
type BlenderRGB48Gamma[CS any, O any, G any] struct {
	gamma G
}

// NewBlenderRGB48Gamma creates a new gamma blender for 16-bit
func NewBlenderRGB48Gamma[CS any, O any, G any](gamma G) BlenderRGB48Gamma[CS, O, G] {
	return BlenderRGB48Gamma[CS, O, G]{gamma: gamma}
}

// BlendPix blends RGB48 with gamma correction
func (bl BlenderRGB48Gamma[CS, O, G]) BlendPix(dst []basics.Int16u, r, g, b, alpha, cover basics.Int16u) {
	blendAlpha := color.RGB16MultCover(alpha, cover)
	if blendAlpha > 0 {
		order := getRGBColorOrder[O]()

		// Apply gamma correction if gamma interface is implemented
		if gamma, ok := interface{}(bl.gamma).(Gamma16Corrector); ok {
			dr := gamma.Dir(dst[order.R])
			dg := gamma.Dir(dst[order.G])
			db := gamma.Dir(dst[order.B])

			sr := gamma.Dir(r)
			sg := gamma.Dir(g)
			sb := gamma.Dir(b)

			dst[order.R] = gamma.Inv(color.RGB16Lerp(dr, sr, blendAlpha))
			dst[order.G] = gamma.Inv(color.RGB16Lerp(dg, sg, blendAlpha))
			dst[order.B] = gamma.Inv(color.RGB16Lerp(db, sb, blendAlpha))
		} else {
			// Fallback to regular blending
			dst[order.R] = color.RGB16Lerp(dst[order.R], r, blendAlpha)
			dst[order.G] = color.RGB16Lerp(dst[order.G], g, blendAlpha)
			dst[order.B] = color.RGB16Lerp(dst[order.B], b, blendAlpha)
		}
	}
}

// Gamma16Corrector interface for 16-bit gamma correction
type Gamma16Corrector interface {
	Dir(v basics.Int16u) basics.Int16u // Apply gamma correction
	Inv(v basics.Int16u) basics.Int16u // Apply inverse gamma correction
}

// Concrete RGB48 blender types
type (
	BlenderRGB48Linear    = BlenderRGB48[color.Linear, color.RGB24Order]
	BlenderRGB48SRGB      = BlenderRGB48[color.SRGB, color.RGB24Order]
	BlenderRGB48PreLinear = BlenderRGB48Pre[color.Linear, color.RGB24Order]
	BlenderRGB48PreSRGB   = BlenderRGB48Pre[color.SRGB, color.RGB24Order]

	BlenderBGR48Linear    = BlenderRGB48[color.Linear, color.BGR24Order]
	BlenderBGR48SRGB      = BlenderRGB48[color.SRGB, color.BGR24Order]
	BlenderBGR48PreLinear = BlenderRGB48Pre[color.Linear, color.BGR24Order]
	BlenderBGR48PreSRGB   = BlenderRGB48Pre[color.SRGB, color.BGR24Order]
)

//==============================================================================
// RGB96 (32-bit float per channel) Blenders
//==============================================================================

// RGB96Blender represents the interface for 32-bit float RGB pixel blending operations
type RGB96Blender interface {
	BlendPix(dst []float32, r, g, b, alpha, cover float32)
}

// BlenderRGB96 implements standard RGB96 blending (32-bit float per channel)
type BlenderRGB96[CS any, O any] struct{}

// BlendPix blends an RGB96 pixel with alpha into an RGB96 buffer
func (bl BlenderRGB96[CS, O]) BlendPix(dst []float32, r, g, b, alpha, cover float32) {
	blendAlpha := alpha * cover
	if blendAlpha > 0 {
		order := getRGBColorOrder[O]()
		invAlpha := 1.0 - blendAlpha
		dst[order.R] = dst[order.R]*invAlpha + r*blendAlpha
		dst[order.G] = dst[order.G]*invAlpha + g*blendAlpha
		dst[order.B] = dst[order.B]*invAlpha + b*blendAlpha
	}
}

// BlenderRGB96Pre implements premultiplied RGB96 blending
type BlenderRGB96Pre[CS any, O any] struct{}

// BlendPix blends a premultiplied RGB96 pixel into an RGB96 buffer
func (bl BlenderRGB96Pre[CS, O]) BlendPix(dst []float32, r, g, b, alpha, cover float32) {
	cr := r * cover
	cg := g * cover
	cb := b * cover
	ca := alpha * cover

	order := getRGBColorOrder[O]()
	invAlpha := 1.0 - ca
	dst[order.R] = dst[order.R]*invAlpha + cr
	dst[order.G] = dst[order.G]*invAlpha + cg
	dst[order.B] = dst[order.B]*invAlpha + cb
}

// BlenderRGB96Gamma implements gamma-corrected RGB96 blending
type BlenderRGB96Gamma[CS any, O any, G any] struct {
	gamma G
}

// NewBlenderRGB96Gamma creates a new gamma blender for 32-bit float
func NewBlenderRGB96Gamma[CS any, O any, G any](gamma G) BlenderRGB96Gamma[CS, O, G] {
	return BlenderRGB96Gamma[CS, O, G]{gamma: gamma}
}

// BlendPix blends RGB96 with gamma correction
func (bl BlenderRGB96Gamma[CS, O, G]) BlendPix(dst []float32, r, g, b, alpha, cover float32) {
	blendAlpha := alpha * cover
	if blendAlpha > 0 {
		order := getRGBColorOrder[O]()

		// Apply gamma correction if gamma interface is implemented
		if gamma, ok := interface{}(bl.gamma).(Gamma32Corrector); ok {
			dr := gamma.Dir(dst[order.R])
			dg := gamma.Dir(dst[order.G])
			db := gamma.Dir(dst[order.B])

			sr := gamma.Dir(r)
			sg := gamma.Dir(g)
			sb := gamma.Dir(b)

			invAlpha := 1.0 - blendAlpha
			dst[order.R] = gamma.Inv(dr*invAlpha + sr*blendAlpha)
			dst[order.G] = gamma.Inv(dg*invAlpha + sg*blendAlpha)
			dst[order.B] = gamma.Inv(db*invAlpha + sb*blendAlpha)
		} else {
			// Fallback to regular blending
			invAlpha := 1.0 - blendAlpha
			dst[order.R] = dst[order.R]*invAlpha + r*blendAlpha
			dst[order.G] = dst[order.G]*invAlpha + g*blendAlpha
			dst[order.B] = dst[order.B]*invAlpha + b*blendAlpha
		}
	}
}

// Gamma32Corrector interface for 32-bit float gamma correction
type Gamma32Corrector interface {
	Dir(v float32) float32 // Apply gamma correction
	Inv(v float32) float32 // Apply inverse gamma correction
}

// Concrete RGB96 blender types
type (
	BlenderRGB96Linear    = BlenderRGB96[color.Linear, color.RGB24Order]
	BlenderRGB96SRGB      = BlenderRGB96[color.SRGB, color.RGB24Order]
	BlenderRGB96PreLinear = BlenderRGB96Pre[color.Linear, color.RGB24Order]
	BlenderRGB96PreSRGB   = BlenderRGB96Pre[color.SRGB, color.RGB24Order]

	BlenderBGR96Linear    = BlenderRGB96[color.Linear, color.BGR24Order]
	BlenderBGR96SRGB      = BlenderRGB96[color.SRGB, color.BGR24Order]
	BlenderBGR96PreLinear = BlenderRGB96Pre[color.Linear, color.BGR24Order]
	BlenderBGR96PreSRGB   = BlenderRGB96Pre[color.SRGB, color.BGR24Order]
)
