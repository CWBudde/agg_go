package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// RGBABlender represents the interface for RGBA pixel blending operations
type RGBABlender interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u)
}

// RGBABlenderSimple represents simplified blending without coverage
type RGBABlenderSimple interface {
	BlendPix(dst []basics.Int8u, r, g, b, a basics.Int8u)
}

// BlenderRGBA implements standard RGBA blending (non-premultiplied source into premultiplied buffer)
type BlenderRGBA[CS any, O any] struct{}

// BlendPix blends a non-premultiplied RGBA pixel into a premultiplied buffer
func (bl BlenderRGBA[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	alpha := color.RGBA8MultCover(a, cover)
	if alpha > 0 {
		order := getColorOrder[O]()
		dst[order.R] = color.RGBA8Lerp(dst[order.R], r, alpha)
		dst[order.G] = color.RGBA8Lerp(dst[order.G], g, alpha)
		dst[order.B] = color.RGBA8Lerp(dst[order.B], b, alpha)
		dst[order.A] = color.RGBA8Prelerp(dst[order.A], alpha, alpha)
	}
}

// BlendPixSimple blends without coverage
func (bl BlenderRGBA[CS, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	if a > 0 {
		order := getColorOrder[O]()
		dst[order.R] = color.RGBA8Lerp(dst[order.R], r, a)
		dst[order.G] = color.RGBA8Lerp(dst[order.G], g, a)
		dst[order.B] = color.RGBA8Lerp(dst[order.B], b, a)
		dst[order.A] = color.RGBA8Prelerp(dst[order.A], a, a)
	}
}

// BlenderRGBAPre implements premultiplied RGBA blending (premultiplied source into premultiplied buffer)
type BlenderRGBAPre[CS any, O any] struct{}

// BlendPix blends a premultiplied RGBA pixel into a premultiplied buffer
func (bl BlenderRGBAPre[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	order := getColorOrder[O]()
	cr := color.RGBA8MultCover(r, cover)
	cg := color.RGBA8MultCover(g, cover)
	cb := color.RGBA8MultCover(b, cover)
	ca := color.RGBA8MultCover(a, cover)

	dst[order.R] = color.RGBA8Prelerp(dst[order.R], cr, ca)
	dst[order.G] = color.RGBA8Prelerp(dst[order.G], cg, ca)
	dst[order.B] = color.RGBA8Prelerp(dst[order.B], cb, ca)
	dst[order.A] = color.RGBA8Prelerp(dst[order.A], ca, ca)
}

// BlendPixSimple blends without coverage
func (bl BlenderRGBAPre[CS, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	order := getColorOrder[O]()
	dst[order.R] = color.RGBA8Prelerp(dst[order.R], r, a)
	dst[order.G] = color.RGBA8Prelerp(dst[order.G], g, a)
	dst[order.B] = color.RGBA8Prelerp(dst[order.B], b, a)
	dst[order.A] = color.RGBA8Prelerp(dst[order.A], a, a)
}

// BlenderRGBAPlain implements plain RGBA blending (non-premultiplied source into non-premultiplied buffer)
type BlenderRGBAPlain[CS any, O any] struct{}

// BlendPix blends a non-premultiplied RGBA pixel into a non-premultiplied buffer
func (bl BlenderRGBAPlain[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	alpha := color.RGBA8MultCover(a, cover)
	if alpha > 0 {
		order := getColorOrder[O]()
		da := dst[order.A]

		// Premultiply destination for calculation
		dr := color.RGBA8Multiply(dst[order.R], da)
		dg := color.RGBA8Multiply(dst[order.G], da)
		db := color.RGBA8Multiply(dst[order.B], da)

		// Blend in premultiplied space
		dst[order.R] = color.RGBA8Lerp(dr, r, alpha)
		dst[order.G] = color.RGBA8Lerp(dg, g, alpha)
		dst[order.B] = color.RGBA8Lerp(db, b, alpha)
		dst[order.A] = color.RGBA8Prelerp(da, alpha, alpha)

		// Demultiply result back to plain space
		if dst[order.A] > 0 {
			inv := basics.Int8u((uint32(dst[order.R])*255 + uint32(dst[order.A])/2) / uint32(dst[order.A]))
			dst[order.R] = inv
			inv = basics.Int8u((uint32(dst[order.G])*255 + uint32(dst[order.A])/2) / uint32(dst[order.A]))
			dst[order.G] = inv
			inv = basics.Int8u((uint32(dst[order.B])*255 + uint32(dst[order.A])/2) / uint32(dst[order.A]))
			dst[order.B] = inv
		}
	}
}

// Helper function to get color order based on type parameter
func getColorOrder[O any]() color.ColorOrder {
	var order color.ColorOrder
	switch any(*new(O)).(type) {
	case RGBAOrder:
		order = color.OrderRGBA
	case ARGBOrder:
		order = color.OrderARGB
	case BGRAOrder:
		order = color.OrderBGRA
	case ABGROrder:
		order = color.OrderABGR
	default:
		// Default to RGBA order
		order = color.OrderRGBA
	}
	return order
}

// Concrete blender types for convenience
type (
	BlenderRGBA8        = BlenderRGBA[color.Linear, RGBAOrder]
	BlenderRGBA8SRGB    = BlenderRGBA[color.SRGB, RGBAOrder]
	BlenderRGBA8Pre     = BlenderRGBAPre[color.Linear, RGBAOrder]
	BlenderRGBA8PreSRGB = BlenderRGBAPre[color.SRGB, RGBAOrder]
	BlenderRGBA8Plain   = BlenderRGBAPlain[color.Linear, RGBAOrder]

	BlenderARGB8    = BlenderRGBA[color.Linear, ARGBOrder]
	BlenderARGB8Pre = BlenderRGBAPre[color.Linear, ARGBOrder]
	BlenderBGRA8    = BlenderRGBA[color.Linear, BGRAOrder]
	BlenderBGRA8Pre = BlenderRGBAPre[color.Linear, BGRAOrder]
	BlenderABGR8    = BlenderRGBA[color.Linear, ABGROrder]
	BlenderABGR8Pre = BlenderRGBAPre[color.Linear, ABGROrder]
)

// Color order type markers
type (
	RGBAOrder struct{}
	ARGBOrder struct{}
	BGRAOrder struct{}
	ABGROrder struct{}
)

// Helper functions for RGBA pixel operations

// BlendRGBAPixel blends a single RGBA pixel
func BlendRGBAPixel[B RGBABlender](dst []basics.Int8u, src color.RGBA8[color.Linear], cover basics.Int8u, blender B) {
	if !src.IsTransparent() {
		blender.BlendPix(dst, src.R, src.G, src.B, src.A, cover)
	}
}

// CopyRGBAPixel copies a single RGBA pixel
func CopyRGBAPixel[O any](dst []basics.Int8u, src color.RGBA8[color.Linear]) {
	order := getColorOrder[O]()
	dst[order.R] = src.R
	dst[order.G] = src.G
	dst[order.B] = src.B
	dst[order.A] = src.A
}

// BlendRGBAHline blends a horizontal line of RGBA pixels
func BlendRGBAHline[B RGBABlender](dst []basics.Int8u, x, length int, src color.RGBA8[color.Linear], covers []basics.Int8u, blender B) {
	if !src.IsTransparent() {
		pixStep := 4
		dstPtr := x * pixStep

		if covers == nil {
			// Uniform coverage
			for i := 0; i < length; i++ {
				blender.BlendPix(dst[dstPtr:], src.R, src.G, src.B, src.A, 255)
				dstPtr += pixStep
			}
		} else {
			// Varying coverage
			for i := 0; i < length; i++ {
				if covers[i] > 0 {
					blender.BlendPix(dst[dstPtr:], src.R, src.G, src.B, src.A, covers[i])
				}
				dstPtr += pixStep
			}
		}
	}
}

// CopyRGBAHline copies a horizontal line of RGBA pixels
func CopyRGBAHline[O any](dst []basics.Int8u, x, length int, src color.RGBA8[color.Linear]) {
	order := getColorOrder[O]()
	pixStep := 4
	dstPtr := x * pixStep

	for i := 0; i < length; i++ {
		dst[dstPtr+order.R] = src.R
		dst[dstPtr+order.G] = src.G
		dst[dstPtr+order.B] = src.B
		dst[dstPtr+order.A] = src.A
		dstPtr += pixStep
	}
}

// FillRGBASpan fills a span with a solid RGBA color
func FillRGBASpan[O any](dst []basics.Int8u, x, length int, src color.RGBA8[color.Linear]) {
	CopyRGBAHline[O](dst, x, length, src)
}
