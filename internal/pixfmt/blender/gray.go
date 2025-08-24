package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// GrayBlender provides the interface for grayscale blending operations
type GrayBlender interface {
	BlendPix(dst *basics.Int8u, cv, alpha, cover basics.Int8u)
}

// BlenderGray implements non-premultiplied blending for grayscale colors
type BlenderGray[C any] struct{}

// BlendPix blends a pixel using non-premultiplied alpha compositing
func (b BlenderGray[C]) BlendPix(dst *basics.Int8u, cv, alpha basics.Int8u, cover basics.Int8u) {
	b.BlendPixAlpha(dst, cv, color.Gray8Multiply(alpha, cover))
}

// BlendPixAlpha blends a pixel with the given alpha value
func (b BlenderGray[C]) BlendPixAlpha(dst *basics.Int8u, cv, alpha basics.Int8u) {
	*dst = color.Gray8Lerp(*dst, cv, alpha)
}

// BlenderGrayPre implements premultiplied blending for grayscale colors
type BlenderGrayPre[C any] struct{}

// BlendPix blends a pixel using premultiplied alpha compositing
func (b BlenderGrayPre[C]) BlendPix(dst *basics.Int8u, cv, alpha basics.Int8u, cover basics.Int8u) {
	b.BlendPixAlpha(dst, color.Gray8Multiply(cv, cover), color.Gray8Multiply(alpha, cover))
}

// BlendPixAlpha blends a pixel with premultiplied values
func (b BlenderGrayPre[C]) BlendPixAlpha(dst *basics.Int8u, cv, alpha basics.Int8u) {
	*dst = color.Gray8Prelerp(*dst, cv, alpha)
}

// Concrete blender types for common grayscale formats
type (
	BlenderGray8        = BlenderGray[color.Gray8Linear]
	BlenderGray8SRGB    = BlenderGray[color.Gray8SRGB]
	BlenderGray8Pre     = BlenderGrayPre[color.Gray8Linear]
	BlenderGray8PreSRGB = BlenderGrayPre[color.Gray8SRGB]
)

// Helper functions for blending operations

// BlendGrayPixel blends a single grayscale pixel
func BlendGrayPixel[C any](dst *basics.Int8u, src color.Gray8[C], cover basics.Int8u, blender BlenderGray[C]) {
	if src.A > 0 {
		blender.BlendPix(dst, src.V, src.A, cover)
	}
}

// BlendGrayPixelPre blends a single premultiplied grayscale pixel
func BlendGrayPixelPre[C any](dst *basics.Int8u, src color.Gray8[C], cover basics.Int8u, blender BlenderGrayPre[C]) {
	if src.A > 0 {
		blender.BlendPix(dst, src.V, src.A, cover)
	}
}

// CopyGrayPixel copies a grayscale pixel (no blending)
func CopyGrayPixel[C any](dst *basics.Int8u, src color.Gray8[C]) {
	*dst = src.V
}

// BlendGrayHline blends a horizontal line of grayscale pixels
func BlendGrayHline[C any](dst []basics.Int8u, x, len int, src color.Gray8[C], covers []basics.Int8u, blender BlenderGray[C]) {
	if src.A == 0 {
		return
	}

	if covers == nil {
		// Solid color - no coverage array
		for i := 0; i < len; i++ {
			blender.BlendPixAlpha(&dst[x+i], src.V, src.A)
		}
	} else {
		// Variable coverage
		for i := 0; i < len; i++ {
			if covers[i] > 0 {
				blender.BlendPix(&dst[x+i], src.V, src.A, covers[i])
			}
		}
	}
}

// BlendGrayHlinePre blends a horizontal line of premultiplied grayscale pixels
func BlendGrayHlinePre[C any](dst []basics.Int8u, x, len int, src color.Gray8[C], covers []basics.Int8u, blender BlenderGrayPre[C]) {
	if src.A == 0 {
		return
	}

	if covers == nil {
		// Solid color - no coverage array
		for i := 0; i < len; i++ {
			blender.BlendPixAlpha(&dst[x+i], src.V, src.A)
		}
	} else {
		// Variable coverage
		for i := 0; i < len; i++ {
			if covers[i] > 0 {
				blender.BlendPix(&dst[x+i], src.V, src.A, covers[i])
			}
		}
	}
}

// CopyGrayHline copies a horizontal line of grayscale pixels
func CopyGrayHline[C any](dst []basics.Int8u, x, len int, src color.Gray8[C]) {
	for i := 0; i < len; i++ {
		dst[x+i] = src.V
	}
}

// FillGraySpan fills a span with a solid grayscale color
func FillGraySpan[C any](dst []basics.Int8u, x, len int, src color.Gray8[C]) {
	for i := 0; i < len; i++ {
		dst[x+i] = src.V
	}
}
