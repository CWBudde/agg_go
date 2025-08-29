package color

import (
	"math"
	"sync"

	"agg_go/internal/basics"
)

//
// sRGB <-> Linear scalar helpers (double precision)
//

const (
	srgbBreak   = 0.04045           // sRGB-domain breakpoint
	linearBreak = srgbBreak / 12.92 // exact corresponding linear breakpoint
	invGamma    = 1.0 / 2.4
)

func ConvertFromSRGB(v float64) float64 {
	if v <= srgbBreak {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func ConvertToSRGB(v float64) float64 {
	if v <= linearBreak {
		return v * 12.92
	}
	return 1.055*math.Pow(v, invGamma) - 0.055
}

//
// Small, cache-once lookup tables for 8-bit sRGB <-> linear
//

var (
	onceTables      sync.Once
	srgbToLinearU8  [256]basics.Int8u
	linearToSrgbU8  [256]basics.Int8u
	srgbToLinearF32 [256]float32
	linearToSrgbF32 [256]float32
)

func initSRGBTables() {
	onceTables.Do(func() {
		for i := range 256 {
			v := float64(i) / 255.0

			lin := ConvertFromSRGB(v)
			srgbToLinearU8[i] = basics.Int8u(lin*255 + 0.5)
			srgbToLinearF32[i] = float32(lin)

			s := ConvertToSRGB(v)
			linearToSrgbU8[i] = basics.Int8u(s*255 + 0.5)
			linearToSrgbF32[i] = float32(s)
		}
	})
}

// Fast 8-bit helpers
func srgb8ToLinear8(v basics.Int8u) basics.Int8u { initSRGBTables(); return srgbToLinearU8[v] }
func linear8ToSrgb8(v basics.Int8u) basics.Int8u { initSRGBTables(); return linearToSrgbU8[v] }

// Fast float helpers (indexed from U8 when going sRGB->linear; computed for linear->sRGB)
func srgb8ToLinearF32(v basics.Int8u) float32 { initSRGBTables(); return srgbToLinearF32[v] }

func linearF32ToSrgb8(v float32) basics.Int8u {
	return basics.Int8u(ConvertToSRGB(float64(v))*255 + 0.5)
}

// Alpha passthrough (no gamma on alpha in AGG)
func alphaU8FromSRGB(a basics.Int8u) basics.Int8u { return a }
func alphaU8ToSRGB(a basics.Int8u) basics.Int8u   { return a }
func alphaF32FromSRGB(a basics.Int8u) float32     { return float32(a) / 255.0 }
func alphaF32ToSRGB(a float32) basics.Int8u       { return basics.Int8u(a*255 + 0.5) }

//
// RGBA8 <-> RGBA8 colorspace conversions
//

func ConvertRGBA8LinearToSRGB(c RGBA8[Linear]) RGBA8[SRGB] {
	return RGBA8[SRGB]{
		R: linear8ToSrgb8(c.R),
		G: linear8ToSrgb8(c.G),
		B: linear8ToSrgb8(c.B),
		A: alphaU8ToSRGB(c.A),
	}
}

func ConvertRGBA8SRGBToLinear(c RGBA8[SRGB]) RGBA8[Linear] {
	return RGBA8[Linear]{
		R: srgb8ToLinear8(c.R),
		G: srgb8ToLinear8(c.G),
		B: srgb8ToLinear8(c.B),
		A: alphaU8FromSRGB(c.A),
	}
}

//
// RGB8 (no alpha)
//

func ConvertRGB8LinearToSRGB(c RGB8[Linear]) RGB8[SRGB] {
	return RGB8[SRGB]{
		R: linear8ToSrgb8(c.R),
		G: linear8ToSrgb8(c.G),
		B: linear8ToSrgb8(c.B),
	}
}

func ConvertRGB8SRGBToLinear(c RGB8[SRGB]) RGB8[Linear] {
	return RGB8[Linear]{
		R: srgb8ToLinear8(c.R),
		G: srgb8ToLinear8(c.G),
		B: srgb8ToLinear8(c.B),
	}
}

//
// Gray8
// Prefer table-based conversions to avoid per-pixel pow.
//

func ConvertGray8LinearToSRGB(g Gray8[Linear]) Gray8[SRGB] {
	return Gray8[SRGB]{V: linear8ToSrgb8(g.V), A: alphaU8ToSRGB(g.A)}
}

func ConvertGray8SRGBToLinear(g Gray8[SRGB]) Gray8[Linear] {
	return Gray8[Linear]{V: srgb8ToLinear8(g.V), A: alphaU8FromSRGB(g.A)}
}

// (Aliases kept for compatibility)
func ConvertGray8FromSRGBToLinear(g Gray8[SRGB]) Gray8[Linear] { return ConvertGray8SRGBToLinear(g) }
func ConvertGray8FromLinearToSRGB(g Gray8[Linear]) Gray8[SRGB] { return ConvertGray8LinearToSRGB(g) }

//
// Gray16
// We approximate via 8-bit table (>>8 then replicate) to keep tables small.
// If you need exactness, you can add 65536-entry tables, but that’s quite large.
//

func ConvertGray16LinearToSRGB(g Gray16[Linear]) Gray16[SRGB] {
	v8 := basics.Int8u(g.V >> 8)
	a8 := basics.Int8u(g.A >> 8)
	V := basics.Int16u(linear8ToSrgb8(v8))
	A := basics.Int16u(alphaU8ToSRGB(a8))
	return Gray16[SRGB]{
		V: (V << 8) | V,
		A: (A << 8) | A,
	}
}

func ConvertGray16SRGBToLinear(g Gray16[SRGB]) Gray16[Linear] {
	v8 := basics.Int8u(g.V >> 8)
	a8 := basics.Int8u(g.A >> 8)
	V := basics.Int16u(srgb8ToLinear8(v8))
	A := basics.Int16u(alphaU8FromSRGB(a8))
	return Gray16[Linear]{
		V: (V << 8) | V,
		A: (A << 8) | A,
	}
}

//
// Float32 (RGBA32 & Gray32)
//

func ConvertRGBA32LinearToSRGB(c RGBA32[Linear]) RGBA32[SRGB] {
	return RGBA32[SRGB]{
		R: float32(ConvertToSRGB(float64(c.R))),
		G: float32(ConvertToSRGB(float64(c.G))),
		B: float32(ConvertToSRGB(float64(c.B))),
		A: c.A, // alpha unchanged
	}
}

func ConvertRGBA32SRGBToLinear(c RGBA32[SRGB]) RGBA32[Linear] {
	return RGBA32[Linear]{
		R: float32(ConvertFromSRGB(float64(c.R))),
		G: float32(ConvertFromSRGB(float64(c.G))),
		B: float32(ConvertFromSRGB(float64(c.B))),
		A: c.A, // alpha unchanged
	}
}

func ConvertGray32LinearToSRGB(g Gray32[Linear]) Gray32[SRGB] {
	return Gray32[SRGB]{V: float32(ConvertToSRGB(float64(g.V))), A: g.A}
}

func ConvertGray32SRGBToLinear(g Gray32[SRGB]) Gray32[Linear] {
	return Gray32[Linear]{V: float32(ConvertFromSRGB(float64(g.V))), A: g.A}
}

//
// Simple “Make” helpers (unchanged)
//

func MakeRGBA8[CS Space](r, g, b, a basics.Int8u) RGBA8[CS] {
	return RGBA8[CS]{R: r, G: g, B: b, A: a}
}
func MakeSRGBA8(r, g, b, a basics.Int8u) RGBA8[SRGB] { return RGBA8[SRGB]{R: r, G: g, B: b, A: a} }
func MakeRGBA16[CS Space](r, g, b, a basics.Int16u) RGBA16[CS] {
	return RGBA16[CS]{R: r, G: g, B: b, A: a}
}

func MakeRGBA32[CS Space](r, g, b, a float32) RGBA32[CS] {
	return RGBA32[CS]{R: r, G: g, B: b, A: a}
}

// Convenience: build premultiplied RGBA8 directly (kept from your version)
func RGBA8Pre[CS Space](r, g, b, a float64) RGBA8[CS] {
	return RGBA8[CS]{
		R: basics.Int8u(r*a*255 + 0.5),
		G: basics.Int8u(g*a*255 + 0.5),
		B: basics.Int8u(b*a*255 + 0.5),
		A: basics.Int8u(a*255 + 0.5),
	}
}

//
// Gray8 -> RGBA8 “make” helpers (kept)
//

func MakeRGBA8FromGray8Linear[CS Space](g Gray8[Linear]) RGBA8[Linear] {
	return RGBA8[Linear]{R: g.V, G: g.V, B: g.V, A: g.A}
}

func MakeRGBA8FromGray8SRGB[CS Space](g Gray8[SRGB]) RGBA8[SRGB] {
	return RGBA8[SRGB]{R: g.V, G: g.V, B: g.V, A: g.A}
}

func MakeSRGBA8FromGray8Linear[CS Space](g Gray8[Linear]) RGBA8[SRGB] {
	return RGBA8[SRGB]{
		R: linear8ToSrgb8(g.V),
		G: linear8ToSrgb8(g.V),
		B: linear8ToSrgb8(g.V),
		A: alphaU8ToSRGB(g.A),
	}
}

func MakeRGBA8FromGray8SRGB_ToLinear[CS Space](g Gray8[SRGB]) RGBA8[Linear] {
	return RGBA8[Linear]{
		R: srgb8ToLinear8(g.V),
		G: srgb8ToLinear8(g.V),
		B: srgb8ToLinear8(g.V),
		A: alphaU8FromSRGB(g.A),
	}
}

func RGB8ToSRGB(src RGB8[Linear]) RGB8[SRGB]   { return ConvertRGB8LinearToSRGB(src) }
func RGB8ToLinear(src RGB8[SRGB]) RGB8[Linear] { return ConvertRGB8SRGBToLinear(src) }
