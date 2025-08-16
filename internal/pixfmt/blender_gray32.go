package pixfmt

import (
	"agg_go/internal/color"
)

// Gray32Blender provides the interface for 32-bit float grayscale blending operations
type Gray32Blender interface {
	BlendPix(dst *float32, cv, alpha, cover float32)
}

// BlenderGray32 implements non-premultiplied blending for 32-bit float grayscale colors
type BlenderGray32[C any] struct{}

// BlendPix blends a pixel using non-premultiplied alpha compositing
func (b BlenderGray32[C]) BlendPix(dst *float32, cv, alpha, cover float32) {
	b.BlendPixAlpha(dst, cv, alpha*cover)
}

// BlendPixAlpha blends a pixel with the given alpha value
func (b BlenderGray32[C]) BlendPixAlpha(dst *float32, cv, alpha float32) {
	*dst = Gray32Lerp(*dst, cv, alpha)
}

// BlenderGray32Pre implements premultiplied blending for 32-bit float grayscale colors
type BlenderGray32Pre[C any] struct{}

// BlendPix blends a pixel using premultiplied alpha compositing
func (b BlenderGray32Pre[C]) BlendPix(dst *float32, cv, alpha, cover float32) {
	b.BlendPixAlpha(dst, cv*cover, alpha*cover)
}

// BlendPixAlpha blends a pixel with premultiplied values
func (b BlenderGray32Pre[C]) BlendPixAlpha(dst *float32, cv, alpha float32) {
	*dst = Gray32Prelerp(*dst, cv, alpha)
}

// Concrete blender types for common 32-bit float grayscale formats
type (
	BlenderGray32Linear    = BlenderGray32[color.Gray32Linear]
	BlenderGray32SRGB      = BlenderGray32[color.Gray32SRGB]
	BlenderGray32PreLinear = BlenderGray32Pre[color.Gray32Linear]
	BlenderGray32PreSRGB   = BlenderGray32Pre[color.Gray32SRGB]
)

// Gray32 arithmetic operations

// Gray32Lerp performs linear interpolation for 32-bit float values
func Gray32Lerp(p, q, a float32) float32 {
	return p + (q-p)*a
}

// Gray32Prelerp performs premultiplied linear interpolation for 32-bit float values
func Gray32Prelerp(p, q, a float32) float32 {
	return p + q - p*a
}

// Helper functions for blending operations

// BlendGray32Pixel blends a single 32-bit float grayscale pixel
func BlendGray32Pixel[C any](dst *float32, src color.Gray32[C], cover float32, blender BlenderGray32[C]) {
	if src.A > 0.0 {
		blender.BlendPix(dst, src.V, src.A, cover)
	}
}

// BlendGray32PixelPre blends a single premultiplied 32-bit float grayscale pixel
func BlendGray32PixelPre[C any](dst *float32, src color.Gray32[C], cover float32, blender BlenderGray32Pre[C]) {
	if src.A > 0.0 {
		blender.BlendPix(dst, src.V, src.A, cover)
	}
}

// CopyGray32Pixel copies a 32-bit float grayscale pixel (no blending)
func CopyGray32Pixel[C any](dst *float32, src color.Gray32[C]) {
	*dst = src.V
}

// BlendGray32Hline blends a horizontal line of 32-bit float grayscale pixels
func BlendGray32Hline[C any](dst []float32, x, len int, src color.Gray32[C], covers []float32, blender BlenderGray32[C]) {
	if src.A == 0.0 {
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
			if covers[i] > 0.0 {
				blender.BlendPix(&dst[x+i], src.V, src.A, covers[i])
			}
		}
	}
}

// BlendGray32HlinePre blends a horizontal line of premultiplied 32-bit float grayscale pixels
func BlendGray32HlinePre[C any](dst []float32, x, len int, src color.Gray32[C], covers []float32, blender BlenderGray32Pre[C]) {
	if src.A == 0.0 {
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
			if covers[i] > 0.0 {
				blender.BlendPix(&dst[x+i], src.V, src.A, covers[i])
			}
		}
	}
}

// CopyGray32Hline copies a horizontal line of 32-bit float grayscale pixels
func CopyGray32Hline[C any](dst []float32, x, len int, src color.Gray32[C]) {
	for i := 0; i < len; i++ {
		dst[x+i] = src.V
	}
}

// FillGray32Span fills a span with a solid 32-bit float grayscale color
func FillGray32Span[C any](dst []float32, x, len int, src color.Gray32[C]) {
	for i := 0; i < len; i++ {
		dst[x+i] = src.V
	}
}
