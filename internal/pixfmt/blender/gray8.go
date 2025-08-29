package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

////////////////////////////////////////////////////////////////////////////////
// Gray Blender interface
////////////////////////////////////////////////////////////////////////////////

// GrayBlender blends 8-bit grayscale pixels in color space S.
// GrayBlender defines the minimal interface used by grayscale pixfmt implementations.
// The blender handles color space interpretation.
type GrayBlender[S color.Space] interface {
	// GetPlain reads a pixel and returns plain grayscale value and alpha
	// interpreted according to color space S
	GetPlain(px *basics.Int8u) (v, a basics.Int8u)

	// SetPlain writes plain grayscale value to a pixel with alpha
	SetPlain(px *basics.Int8u, v, a basics.Int8u)

	// BlendPix blends plain grayscale source into the pixel with given alpha and coverage
	// v is interpreted according to S
	BlendPix(px *basics.Int8u, v, a, cover basics.Int8u)
}

////////////////////////////////////////////////////////////////////////////////
// Non-premultiplied grayscale -> Non-premultiplied destination
////////////////////////////////////////////////////////////////////////////////

type BlenderGray8[S color.Space] struct{}

// BlendPix: straight (non-premultiplied) alpha with coverage.
// a' = a * cover (8-bit), then dst = lerp(dst, v, a').
func (BlenderGray8[S]) BlendPix(dst *basics.Int8u, v, a, cover basics.Int8u) {
	if a == 0 || cover == 0 {
		return
	}
	BlendGrayAlpha(dst, v, color.Gray8Multiply(a, cover))
}

// BlendGrayAlpha: dst = dst + (v - dst) * a
func BlendGrayAlpha(dst *basics.Int8u, v, a basics.Int8u) {
	*dst = color.Gray8Lerp(*dst, v, a)
}

////////////////////////////////////////////////////////////////////////////////
// Premultiplied grayscale -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

type BlenderGray8Pre[S color.Space] struct{}

// BlendPix: premultiplied compositing; coverage scales both v and a.
// dst = prelerp(dst, v*cover, a*cover)
func (BlenderGray8Pre[S]) BlendPix(dst *basics.Int8u, v, a, cover basics.Int8u) {
	if a == 0 || cover == 0 {
		return
	}
	BlendGrayPreAlpha(dst, color.Gray8Multiply(v, cover), color.Gray8Multiply(a, cover))
}

// BlendGrayPreAlpha: dst = dst + v - dst*a
func BlendGrayPreAlpha(dst *basics.Int8u, v, a basics.Int8u) {
	*dst = color.Gray8Prelerp(*dst, v, a)
}

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases (concrete, non-generic)
////////////////////////////////////////////////////////////////////////////////

type (
	// Standard grayscale blenders
	BlenderGray8SRGB    = BlenderGray8[color.SRGB]
	BlenderGray8PreSRGB = BlenderGray8Pre[color.SRGB]

	// Linear space variants for high-quality rendering
	BlenderGray8Linear    = BlenderGray8[color.Linear]
	BlenderGray8PreLinear = BlenderGray8Pre[color.Linear]
)

////////////////////////////////////////////////////////////////////////////////
// Platform and usage-specific aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// Common single-channel formats
	BlenderMonochrome    = BlenderGray8[color.SRGB]
	BlenderMonochromePre = BlenderGray8Pre[color.SRGB]

	// Alpha channel blending (grayscale used for alpha)
	BlenderAlpha8    = BlenderGray8[color.Linear]
	BlenderAlpha8Pre = BlenderGray8Pre[color.Linear]

	// Mask blending (for compositing operations)
	BlenderMask8    = BlenderGray8[color.Linear]
	BlenderMask8Pre = BlenderGray8Pre[color.Linear]
)

////////////////////////////////////////////////////////////////////////////////
// Helpers for single pixels and spans (generic over S)
////////////////////////////////////////////////////////////////////////////////

// BlendGrayPixel blends one non-premultiplied grayscale pixel.
func BlendGrayPixel[S color.Space](dst *basics.Int8u, src color.Gray8[S], cover basics.Int8u, b GrayBlender[S]) {
	if src.A != 0 && cover != 0 {
		b.BlendPix(dst, src.V, src.A, cover)
	}
}

// BlendGrayPixelPre blends one premultiplied grayscale pixel.
// (Kept for API clarity; you can also call BlendGrayPixel with a Pre blender.)
func BlendGrayPixelPre[S color.Space](dst *basics.Int8u, src color.Gray8[S], cover basics.Int8u, b BlenderGray8Pre[S]) {
	if src.A != 0 && cover != 0 {
		b.BlendPix(dst, src.V, src.A, cover)
	}
}

func (BlenderGray8[S]) SetPlain(dst *basics.Int8u, v, a basics.Int8u) {
	*dst = v
}

func (BlenderGray8[S]) GetPlain(src *basics.Int8u) (v, a basics.Int8u) {
	return *src, 255
}

func (BlenderGray8Pre[S]) SetPlain(dst *basics.Int8u, v, a basics.Int8u) {
	*dst = color.Gray8Multiply(v, a)
}

func (BlenderGray8Pre[S]) GetPlain(src *basics.Int8u) (v, a basics.Int8u) {
	return *src, 255
}

// CopyGrayPixel copies a grayscale value (no blending).
func CopyGrayPixel[S color.Space](dst *basics.Int8u, src color.Gray8[S]) {
	*dst = src.V
}

// BlendGrayHline blends a horizontal run with optional per-pixel coverage.
// Works with both plain and premultiplied variants via the interface.
func BlendGrayHline[S color.Space](dst []basics.Int8u, x, n int, src color.Gray8[S], covers []basics.Int8u, b GrayBlender[S]) {
	if n <= 0 || src.A == 0 {
		return
	}
	if covers == nil {
		// Uniform full cover
		for i := 0; i < n; i++ {
			b.BlendPix(&dst[x+i], src.V, src.A, 255)
		}
		return
	}
	// Variable coverage
	for i := 0; i < n; i++ {
		if cv := covers[i]; cv != 0 {
			b.BlendPix(&dst[x+i], src.V, src.A, cv)
		}
	}
}

// BlendGrayHlinePre blends a premultiplied run (explicit helper).
func BlendGrayHlinePre[S color.Space](dst []basics.Int8u, x, n int, src color.Gray8[S], covers []basics.Int8u, b BlenderGray8Pre[S]) {
	if n <= 0 || src.A == 0 {
		return
	}
	if covers == nil {
		for i := 0; i < n; i++ {
			b.BlendPix(&dst[x+i], src.V, src.A, 255)
		}
		return
	}
	for i := 0; i < n; i++ {
		if cv := covers[i]; cv != 0 {
			b.BlendPix(&dst[x+i], src.V, src.A, cv)
		}
	}
}

// CopyGrayHline copies a horizontal run (no blending).
func CopyGrayHline[S color.Space](dst []basics.Int8u, x, n int, src color.Gray8[S]) {
	for i := 0; i < n; i++ {
		dst[x+i] = src.V
	}
}

// FillGraySpan is an alias for CopyGrayHline (semantic sugar).
func FillGraySpan[S color.Space](dst []basics.Int8u, x, n int, src color.Gray8[S]) {
	CopyGrayHline[S](dst, x, n, src)
}
