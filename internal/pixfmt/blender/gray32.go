package blender

import (
	"agg_go/internal/color"
)

// Gray32Blender blends 32-bit float grayscale pixels in color space S.
// (No channel order concept for single-channel grayscale.)
type Gray32Blender[S color.Space] interface {
	// BlendPix blends a source value v with alpha a into *dst with coverage cover.
	// Semantics depend on the concrete blender:
	//  - BlenderGray32[S]     : non-premultiplied source -> non-premultiplied dst
	//  - BlenderGray32Pre[S]  : premultiplied source     -> premultiplied dst
	BlendPix(dst *float32, v, a, cover float32)
}

////////////////////////////////////////////////////////////////////////////////
// Non-premultiplied grayscale -> Non-premultiplied destination
////////////////////////////////////////////////////////////////////////////////

type BlenderGray32[S color.Space] struct{}

// BlendPix: plain alpha compositing with coverage (a' = a*cover)
func (BlenderGray32[S]) BlendPix(dst *float32, v, a, cover float32) {
	if a <= 0 || cover <= 0 {
		return
	}
	BlendGray32Alpha(dst, v, a*cover)
}

// BlendGray32Alpha performs straight (non-premultiplied) lerp: dst = dst + (v - dst) * a
func BlendGray32Alpha(dst *float32, v, a float32) {
	*dst = Gray32Lerp(*dst, v, a)
}

////////////////////////////////////////////////////////////////////////////////
// Premultiplied grayscale -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

type BlenderGray32Pre[S color.Space] struct{}

// BlendPix: premultiplied compositing; coverage scales both v and a
func (BlenderGray32Pre[S]) BlendPix(dst *float32, v, a, cover float32) {
	if a <= 0 || cover <= 0 {
		return
	}
	BlendGray32PreAlpha(dst, v*cover, a*cover)
}

// BlendGray32PreAlpha performs premultiplied blend: dst = dst + v - dst*a
func BlendGray32PreAlpha(dst *float32, v, a float32) {
	*dst = Gray32Prelerp(*dst, v, a)
}

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases (match your RGBA naming)
////////////////////////////////////////////////////////////////////////////////

type (
	BlenderGray32Linear    = BlenderGray32[color.Linear]
	BlenderGray32SRGB      = BlenderGray32[color.SRGB]
	BlenderGray32PreLinear = BlenderGray32Pre[color.Linear]
	BlenderGray32PreSRGB   = BlenderGray32Pre[color.SRGB]
)

////////////////////////////////////////////////////////////////////////////////
// Arithmetic helpers (unchanged behavior)
////////////////////////////////////////////////////////////////////////////////

// Gray32Lerp: straight alpha interpolation
func Gray32Lerp(p, q, a float32) float32 { return p + (q-p)*a }

// Gray32Prelerp: premultiplied interpolation
func Gray32Prelerp(p, q, a float32) float32 { return p + q - p*a }

////////////////////////////////////////////////////////////////////////////////
// Blending helpers for spans and single pixels
////////////////////////////////////////////////////////////////////////////////

// BlendGray32Pixel blends a single non-premultiplied grayscale pixel.
func BlendGray32Pixel[S color.Space](dst *float32, src color.Gray32[S], cover float32, b Gray32Blender[S]) {
	if src.A > 0 && cover > 0 {
		b.BlendPix(dst, src.V, src.A, cover)
	}
}

// BlendGray32PixelPre blends a single premultiplied grayscale pixel.
// (Signature kept for clarity; you can also just use BlendGray32Pixel with a Pre blender.)
func BlendGray32PixelPre[S color.Space](dst *float32, src color.Gray32[S], cover float32, b BlenderGray32Pre[S]) {
	if src.A > 0 && cover > 0 {
		b.BlendPix(dst, src.V, src.A, cover)
	}
}

// CopyGray32Pixel copies one grayscale value (no blending).
func CopyGray32Pixel[S color.Space](dst *float32, src color.Gray32[S]) {
	*dst = src.V
}

// BlendGray32Hline blends a horizontal line with optional per-pixel coverage.
// Works for both plain and premultiplied variants via the interface.
func BlendGray32Hline[S color.Space](dst []float32, x, n int, src color.Gray32[S], covers []float32, b Gray32Blender[S]) {
	if n <= 0 || src.A <= 0 {
		return
	}
	if covers == nil {
		// Uniform full cover
		for i := 0; i < n; i++ {
			b.BlendPix(&dst[x+i], src.V, src.A, 1.0)
		}
		return
	}
	// Variable coverage
	for i := 0; i < n; i++ {
		cv := covers[i]
		if cv > 0 {
			b.BlendPix(&dst[x+i], src.V, src.A, cv)
		}
	}
}

// CopyGray32Hline copies a horizontal line without blending.
func CopyGray32Hline[S color.Space](dst []float32, x, n int, src color.Gray32[S]) {
	for i := 0; i < n; i++ {
		dst[x+i] = src.V
	}
}

// FillGray32Span is an alias for CopyGray32Hline (semantic sugar).
func FillGray32Span[S color.Space](dst []float32, x, n int, src color.Gray32[S]) {
	CopyGray32Hline[S](dst, x, n, src)
}
