package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces (Gray 16-bit)
////////////////////////////////////////////////////////////////////////////////

// Gray16Blender blends 16-bit grayscale pixels in color space S.
// (Grayscale has no channel order; we only carry S.)
type Gray16Blender[S color.Space] interface {
	// v = source value (plain or premultiplied depending on impl)
	// a = source alpha (plain or premultiplied depending on impl)
	// cover scales contribution in [0..65535]
	BlendPix(dst *basics.Int16u, v, a, cover basics.Int16u)
}

////////////////////////////////////////////////////////////////////////////////
// Non-premultiplied grayscale -> Non-premultiplied destination
////////////////////////////////////////////////////////////////////////////////

type BlenderGray16[S color.Space] struct{}

// BlendPix: straight (non-premultiplied) alpha compositing with coverage.
// a' = a * cover (fixed-point), then dst = lerp(dst, v, a').
func (BlenderGray16[S]) BlendPix(dst *basics.Int16u, v, a, cover basics.Int16u) {
	if a == 0 || cover == 0 {
		return
	}
	BlendGray16Alpha(dst, v, Gray16Multiply(a, cover))
}

// BlendGray16Alpha performs straight alpha blend: dst = dst + (v - dst) * a
func BlendGray16Alpha(dst *basics.Int16u, v, a basics.Int16u) {
	*dst = Gray16Lerp(*dst, v, a)
}

////////////////////////////////////////////////////////////////////////////////
// Premultiplied grayscale -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

type BlenderGray16Pre[S color.Space] struct{}

// BlendPix: premultiplied compositing; coverage scales both v and a.
// dst = prelerp(dst, v*cover, a*cover)
func (BlenderGray16Pre[S]) BlendPix(dst *basics.Int16u, v, a, cover basics.Int16u) {
	if a == 0 || cover == 0 {
		return
	}
	BlendGray16PreAlpha(dst, Gray16Multiply(v, cover), Gray16Multiply(a, cover))
}

// BlendGray16PreAlpha performs premultiplied blend: dst = dst + v - dst*a
func BlendGray16PreAlpha(dst *basics.Int16u, v, a basics.Int16u) {
	*dst = Gray16Prelerp(*dst, v, a)
}

////////////////////////////////////////////////////////////////////////////////
/* Convenience aliases (concrete, non-generic) */
////////////////////////////////////////////////////////////////////////////////

type (
	BlenderGray16Linear    = BlenderGray16[color.Linear]
	BlenderGray16SRGB      = BlenderGray16[color.SRGB]
	BlenderGray16PreLinear = BlenderGray16Pre[color.Linear]
	BlenderGray16PreSRGB   = BlenderGray16Pre[color.SRGB]
)

////////////////////////////////////////////////////////////////////////////////
// Fixed-point arithmetic helpers (unchanged semantics)
////////////////////////////////////////////////////////////////////////////////

// Gray16Multiply performs fixed-point multiplication for 16-bit values.
// Same “(x*65535 + 0x8000) / 65535” rounded behavior as your original.
func Gray16Multiply(a, b basics.Int16u) basics.Int16u {
	t := uint64(a)*uint64(b) + 0x8000
	return basics.Int16u(((t >> 16) + t) >> 16)
}

// Gray16Lerp: straight interpolation p + (q - p)*a with rounding parity guard
func Gray16Lerp(p, q, a basics.Int16u) basics.Int16u {
	var t int64
	if p > q {
		t = int64(q-p)*int64(a) + 0x8000 - 1
	} else {
		t = int64(q-p)*int64(a) + 0x8000
	}
	return basics.Int16u(int64(p) + (((t >> 16) + t) >> 16))
}

// Gray16Prelerp: premultiplied interpolation p + q - p*a
func Gray16Prelerp(p, q, a basics.Int16u) basics.Int16u {
	return p + q - Gray16Multiply(p, a)
}

////////////////////////////////////////////////////////////////////////////////
// Helpers for single pixels and spans (generic over S color.Space)
////////////////////////////////////////////////////////////////////////////////

// BlendGray16Pixel blends one non-premultiplied grayscale pixel.
func BlendGray16Pixel[S color.Space](dst *basics.Int16u, src color.Gray16[S], cover basics.Int16u, b Gray16Blender[S]) {
	if src.A != 0 && cover != 0 {
		b.BlendPix(dst, src.V, src.A, cover)
	}
}

// BlendGray16PixelPre blends one premultiplied grayscale pixel.
// (Kept for API clarity; you can also call BlendGray16Pixel with a Pre blender.)
func BlendGray16PixelPre[S color.Space](dst *basics.Int16u, src color.Gray16[S], cover basics.Int16u, b BlenderGray16Pre[S]) {
	if src.A != 0 && cover != 0 {
		b.BlendPix(dst, src.V, src.A, cover)
	}
}

func (BlenderGray16[S]) SetPlain(dst *basics.Int16u, v, a basics.Int16u) {
	*dst = v
}

func (BlenderGray16[S]) GetPlain(src *basics.Int16u) (v, a basics.Int16u) {
	return *src, 0xFFFF
}

func (BlenderGray16Pre[S]) SetPlain(dst *basics.Int16u, v, a basics.Int16u) {
	*dst = Gray16Multiply(v, a)
}

func (BlenderGray16Pre[S]) GetPlain(src *basics.Int16u) (v, a basics.Int16u) {
	return *src, 0xFFFF
}

// CopyGray16Pixel copies a grayscale value (no blending).
func CopyGray16Pixel[S color.Space](dst *basics.Int16u, src color.Gray16[S]) {
	*dst = src.V
}

// BlendGray16Hline blends a horizontal run with optional per-pixel coverage.
// Works with both plain and premultiplied variants via the interface.
func BlendGray16Hline[S color.Space](dst []basics.Int16u, x, n int, src color.Gray16[S], covers []basics.Int16u, b Gray16Blender[S]) {
	if n <= 0 || src.A == 0 {
		return
	}
	if covers == nil {
		// Uniform full cover
		for i := 0; i < n; i++ {
			b.BlendPix(&dst[x+i], src.V, src.A, 0xFFFF)
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

// BlendGray16HlinePre blends a premultiplied run (explicit helper).
func BlendGray16HlinePre[S color.Space](dst []basics.Int16u, x, n int, src color.Gray16[S], covers []basics.Int16u, b BlenderGray16Pre[S]) {
	if n <= 0 || src.A == 0 {
		return
	}
	if covers == nil {
		for i := 0; i < n; i++ {
			b.BlendPix(&dst[x+i], src.V, src.A, 0xFFFF)
		}
		return
	}
	for i := 0; i < n; i++ {
		if cv := covers[i]; cv != 0 {
			b.BlendPix(&dst[x+i], src.V, src.A, cv)
		}
	}
}

// CopyGray16Hline copies a horizontal run (no blending).
func CopyGray16Hline[S color.Space](dst []basics.Int16u, x, n int, src color.Gray16[S]) {
	for i := 0; i < n; i++ {
		dst[x+i] = src.V
	}
}

// FillGray16Span is an alias for CopyGray16Hline (semantic sugar).
func FillGray16Span[S color.Space](dst []basics.Int16u, x, n int, src color.Gray16[S]) {
	CopyGray16Hline[S](dst, x, n, src)
}
