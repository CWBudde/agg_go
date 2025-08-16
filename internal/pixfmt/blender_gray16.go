package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// Gray16Blender provides the interface for 16-bit grayscale blending operations
type Gray16Blender interface {
	BlendPix(dst *basics.Int16u, cv, alpha, cover basics.Int16u)
}

// BlenderGray16 implements non-premultiplied blending for 16-bit grayscale colors
type BlenderGray16[C any] struct{}

// BlendPix blends a pixel using non-premultiplied alpha compositing
func (b BlenderGray16[C]) BlendPix(dst *basics.Int16u, cv, alpha, cover basics.Int16u) {
	b.BlendPixAlpha(dst, cv, Gray16Multiply(alpha, cover))
}

// BlendPixAlpha blends a pixel with the given alpha value
func (b BlenderGray16[C]) BlendPixAlpha(dst *basics.Int16u, cv, alpha basics.Int16u) {
	*dst = Gray16Lerp(*dst, cv, alpha)
}

// BlenderGray16Pre implements premultiplied blending for 16-bit grayscale colors
type BlenderGray16Pre[C any] struct{}

// BlendPix blends a pixel using premultiplied alpha compositing
func (b BlenderGray16Pre[C]) BlendPix(dst *basics.Int16u, cv, alpha, cover basics.Int16u) {
	b.BlendPixAlpha(dst, Gray16Multiply(cv, cover), Gray16Multiply(alpha, cover))
}

// BlendPixAlpha blends a pixel with premultiplied values
func (b BlenderGray16Pre[C]) BlendPixAlpha(dst *basics.Int16u, cv, alpha basics.Int16u) {
	*dst = Gray16Prelerp(*dst, cv, alpha)
}

// Concrete blender types for common 16-bit grayscale formats
type (
	BlenderGray16Linear    = BlenderGray16[color.Gray16Linear]
	BlenderGray16SRGB      = BlenderGray16[color.Gray16SRGB]
	BlenderGray16PreLinear = BlenderGray16Pre[color.Gray16Linear]
	BlenderGray16PreSRGB   = BlenderGray16Pre[color.Gray16SRGB]
)

// Gray16 arithmetic operations

// Gray16Multiply performs fixed-point multiplication for 16-bit values
func Gray16Multiply(a, b basics.Int16u) basics.Int16u {
	t := uint64(a)*uint64(b) + 0x8000
	return basics.Int16u(((t >> 16) + t) >> 16)
}

// Gray16Lerp performs linear interpolation for 16-bit values
func Gray16Lerp(p, q, a basics.Int16u) basics.Int16u {
	var t int64
	if p > q {
		t = int64(q-p)*int64(a) + 0x8000 - 1
	} else {
		t = int64(q-p)*int64(a) + 0x8000
	}
	return basics.Int16u(int64(p) + (((t >> 16) + t) >> 16))
}

// Gray16Prelerp performs premultiplied linear interpolation for 16-bit values
func Gray16Prelerp(p, q, a basics.Int16u) basics.Int16u {
	return p + q - Gray16Multiply(p, a)
}

// Helper functions for blending operations

// BlendGray16Pixel blends a single 16-bit grayscale pixel
func BlendGray16Pixel[C any](dst *basics.Int16u, src color.Gray16[C], cover basics.Int16u, blender BlenderGray16[C]) {
	if src.A > 0 {
		blender.BlendPix(dst, src.V, src.A, cover)
	}
}

// BlendGray16PixelPre blends a single premultiplied 16-bit grayscale pixel
func BlendGray16PixelPre[C any](dst *basics.Int16u, src color.Gray16[C], cover basics.Int16u, blender BlenderGray16Pre[C]) {
	if src.A > 0 {
		blender.BlendPix(dst, src.V, src.A, cover)
	}
}

// CopyGray16Pixel copies a 16-bit grayscale pixel (no blending)
func CopyGray16Pixel[C any](dst *basics.Int16u, src color.Gray16[C]) {
	*dst = src.V
}

// BlendGray16Hline blends a horizontal line of 16-bit grayscale pixels
func BlendGray16Hline[C any](dst []basics.Int16u, x, len int, src color.Gray16[C], covers []basics.Int16u, blender BlenderGray16[C]) {
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

// BlendGray16HlinePre blends a horizontal line of premultiplied 16-bit grayscale pixels
func BlendGray16HlinePre[C any](dst []basics.Int16u, x, len int, src color.Gray16[C], covers []basics.Int16u, blender BlenderGray16Pre[C]) {
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

// CopyGray16Hline copies a horizontal line of 16-bit grayscale pixels
func CopyGray16Hline[C any](dst []basics.Int16u, x, len int, src color.Gray16[C]) {
	for i := 0; i < len; i++ {
		dst[x+i] = src.V
	}
}

// FillGray16Span fills a span with a solid 16-bit grayscale color
func FillGray16Span[C any](dst []basics.Int16u, x, len int, src color.Gray16[C]) {
	for i := 0; i < len; i++ {
		dst[x+i] = src.V
	}
}
