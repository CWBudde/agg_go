package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces
////////////////////////////////////////////////////////////////////////////////

// GrayBlender blends 8-bit grayscale pixels in color space S.
type GrayBlender[S color.Space] interface {
	// v = source value (plain or premultiplied depending on impl)
	// a = source alpha (plain or premultiplied depending on impl)
	// cover scales contribution in [0..255]
	BlendPix(dst *basics.Int8u, v, a, cover basics.Int8u)
}

////////////////////////////////////////////////////////////////////////////////
// Non-premultiplied grayscale -> Non-premultiplied destination
////////////////////////////////////////////////////////////////////////////////

type BlenderGray[S color.Space] struct{}

// BlendPix: straight (non-premultiplied) alpha with coverage.
// a' = a * cover (8-bit), then dst = lerp(dst, v, a').
func (BlenderGray[S]) BlendPix(dst *basics.Int8u, v, a, cover basics.Int8u) {
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

type BlenderGrayPre[S color.Space] struct{}

// BlendPix: premultiplied compositing; coverage scales both v and a.
// dst = prelerp(dst, v*cover, a*cover)
func (BlenderGrayPre[S]) BlendPix(dst *basics.Int8u, v, a, cover basics.Int8u) {
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
	BlenderGray8        = BlenderGray[color.Linear]
	BlenderGray8SRGB    = BlenderGray[color.SRGB]
	BlenderGray8Pre     = BlenderGrayPre[color.Linear]
	BlenderGray8PreSRGB = BlenderGrayPre[color.SRGB]
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
func BlendGrayPixelPre[S color.Space](dst *basics.Int8u, src color.Gray8[S], cover basics.Int8u, b BlenderGrayPre[S]) {
	if src.A != 0 && cover != 0 {
		b.BlendPix(dst, src.V, src.A, cover)
	}
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
func BlendGrayHlinePre[S color.Space](dst []basics.Int8u, x, n int, src color.Gray8[S], covers []basics.Int8u, b BlenderGrayPre[S]) {
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
