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
		// For now, use fixed RGB order - we can make this configurable later
		dst[0] = color.RGBA8Lerp(dst[0], r, blendAlpha)
		dst[1] = color.RGBA8Lerp(dst[1], g, blendAlpha)
		dst[2] = color.RGBA8Lerp(dst[2], b, blendAlpha)
	}
}

// BlendPixSimple blends RGB without coverage
func (bl BlenderRGB[CS, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, alpha basics.Int8u) {
	if alpha > 0 {
		// For now, use fixed RGB order - we can make this configurable later
		dst[0] = color.RGBA8Lerp(dst[0], r, alpha)
		dst[1] = color.RGBA8Lerp(dst[1], g, alpha)
		dst[2] = color.RGBA8Lerp(dst[2], b, alpha)
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

	// For RGB format, we use premultiplied blending but don't store alpha
	dst[0] = color.RGBA8Prelerp(dst[0], cr, ca)
	dst[1] = color.RGBA8Prelerp(dst[1], cg, ca)
	dst[2] = color.RGBA8Prelerp(dst[2], cb, ca)
}

// BlendPixSimple blends premultiplied RGB without coverage
func (bl BlenderRGBPre[CS, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, alpha basics.Int8u) {
	dst[0] = color.RGBA8Prelerp(dst[0], r, alpha)
	dst[1] = color.RGBA8Prelerp(dst[1], g, alpha)
	dst[2] = color.RGBA8Prelerp(dst[2], b, alpha)
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
