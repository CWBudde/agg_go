package color

import (
	"math"

	"agg_go/internal/basics"
)

// sRGB conversion functions

// ConvertFromSRGB converts from sRGB to linear RGB
func ConvertFromSRGB(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// ConvertToSRGB converts from linear RGB to sRGB
func ConvertToSRGB(v float64) float64 {
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

// Convert RGBA8 between colorspaces using optimized lookup tables
func ConvertRGBA8LinearToSRGB(c RGBA8[Linear]) RGBA8[SRGB] {
	return RGBA8[SRGB]{
		R: SRGBConv8BitRGBToSRGB(c.R),
		G: SRGBConv8BitRGBToSRGB(c.G),
		B: SRGBConv8BitRGBToSRGB(c.B),
		A: SRGBConv8BitAlphaToSRGB(c.A),
	}
}

func ConvertRGBA8SRGBToLinear(c RGBA8[SRGB]) RGBA8[Linear] {
	return RGBA8[Linear]{
		R: SRGBConv8BitRGBFromSRGB(c.R),
		G: SRGBConv8BitRGBFromSRGB(c.G),
		B: SRGBConv8BitRGBFromSRGB(c.B),
		A: SRGBConv8BitAlphaFromSRGB(c.A),
	}
}

// Convert Gray8 between colorspaces
func ConvertGray8LinearToSRGB(g Gray8[Linear]) Gray8[SRGB] {
	return Gray8[SRGB]{
		V: basics.Int8u(ConvertToSRGB(float64(g.V)/255.0)*255 + 0.5),
		A: g.A,
	}
}

func ConvertGray8SRGBToLinear(g Gray8[SRGB]) Gray8[Linear] {
	return Gray8[Linear]{
		V: basics.Int8u(ConvertFromSRGB(float64(g.V)/255.0)*255 + 0.5),
		A: g.A,
	}
}

// Gray8 colorspace conversion functions
func ConvertGray8FromSRGBToLinear(g Gray8[SRGB]) Gray8[Linear] {
	return Gray8[Linear]{
		V: SRGBConv8BitRGBFromSRGB(g.V),
		A: SRGBConv8BitAlphaFromSRGB(g.A),
	}
}

func ConvertGray8FromLinearToSRGB(g Gray8[Linear]) Gray8[SRGB] {
	return Gray8[SRGB]{
		V: SRGBConv8BitRGBToSRGB(g.V),
		A: SRGBConv8BitAlphaToSRGB(g.A),
	}
}

// Helper functions for creating colors with conversion
func MakeRGBA8[CS any](r, g, b, a basics.Int8u) RGBA8[CS] {
	return RGBA8[CS]{R: r, G: g, B: b, A: a}
}

func MakeSRGBA8(r, g, b, a basics.Int8u) RGBA8[SRGB] {
	return RGBA8[SRGB]{R: r, G: g, B: b, A: a}
}

func MakeRGBA16[CS any](r, g, b, a basics.Int16u) RGBA16[CS] {
	return RGBA16[CS]{R: r, G: g, B: b, A: a}
}

func MakeRGBA32[CS any](r, g, b, a float32) RGBA32[CS] {
	return RGBA32[CS]{R: r, G: g, B: b, A: a}
}

// RGB conversion helpers (without alpha)
func RGBConvRGBA8[CS any](r, g, b basics.Int8u) RGBA8[CS] {
	return RGBA8[CS]{R: r, G: g, B: b, A: 255}
}

func RGBConvRGBA16[CS any](r, g, b basics.Int16u) RGBA16[CS] {
	return RGBA16[CS]{R: r, G: g, B: b, A: 65535}
}

// sRGB conversion lookup tables for performance
// These are computed once and cached for 8-bit conversions
var (
	srgbToLinearTable    [256]basics.Int8u
	linearToSRGBTable    [256]basics.Int8u
	srgbToLinearF32Table [256]float32
	linearToSRGBF32Table [256]float32
	tablesInitialized    bool
)

// initSRGBTables initializes the lookup tables
func initSRGBTables() {
	if tablesInitialized {
		return
	}

	for i := 0; i < 256; i++ {
		v := float64(i) / 255.0

		// sRGB to linear
		linear := ConvertFromSRGB(v)
		srgbToLinearTable[i] = basics.Int8u(linear*255 + 0.5)
		srgbToLinearF32Table[i] = float32(linear)

		// Linear to sRGB
		srgb := ConvertToSRGB(v)
		linearToSRGBTable[i] = basics.Int8u(srgb*255 + 0.5)
		linearToSRGBF32Table[i] = float32(srgb)
	}

	tablesInitialized = true
}

// SRGBConv provides generic conversion utilities based on C++ AGG sRGB_conv template
type SRGBConv[T any] struct{}

// SRGBConv8BitRGBFromSRGB converts 8-bit sRGB to linear using lookup table
func SRGBConv8BitRGBFromSRGB(v basics.Int8u) basics.Int8u {
	initSRGBTables()
	return srgbToLinearTable[v]
}

// SRGBConv8BitRGBToSRGB converts 8-bit linear to sRGB using lookup table
func SRGBConv8BitRGBToSRGB(v basics.Int8u) basics.Int8u {
	initSRGBTables()
	return linearToSRGBTable[v]
}

// SRGBConvF32RGBFromSRGB converts using float32 lookup table
func SRGBConvF32RGBFromSRGB(v basics.Int8u) float32 {
	initSRGBTables()
	return srgbToLinearF32Table[v]
}

// SRGBConvF32RGBToSRGB converts from float32 to sRGB using computation
func SRGBConvF32RGBToSRGB(v float32) basics.Int8u {
	return basics.Int8u(ConvertToSRGB(float64(v))*255 + 0.5)
}

// Alpha conversion functions (alpha doesn't undergo gamma correction)
func SRGBConv8BitAlphaFromSRGB(a basics.Int8u) basics.Int8u {
	return a
}

func SRGBConv8BitAlphaToSRGB(a basics.Int8u) basics.Int8u {
	return a
}

func SRGBConvF32AlphaFromSRGB(a basics.Int8u) float32 {
	return float32(a) / 255.0
}

func SRGBConvF32AlphaToSRGB(a float32) basics.Int8u {
	return basics.Int8u(a*255 + 0.5)
}

// Gray8 make functions for different colorspaces
func MakeRGBA8FromGray8Linear[CS any](g Gray8[Linear]) RGBA8[Linear] {
	return RGBA8[Linear]{R: g.V, G: g.V, B: g.V, A: g.A}
}

func MakeRGBA8FromGray8SRGB[CS any](g Gray8[SRGB]) RGBA8[SRGB] {
	return RGBA8[SRGB]{R: g.V, G: g.V, B: g.V, A: g.A}
}

func MakeSRGBA8FromGray8Linear[CS any](g Gray8[Linear]) RGBA8[SRGB] {
	// Convert from linear to sRGB
	return RGBA8[SRGB]{
		R: SRGBConv8BitRGBToSRGB(g.V),
		G: SRGBConv8BitRGBToSRGB(g.V),
		B: SRGBConv8BitRGBToSRGB(g.V),
		A: SRGBConv8BitAlphaToSRGB(g.A),
	}
}

func MakeRGBA8FromGray8SRGB_ToLinear[CS any](g Gray8[SRGB]) RGBA8[Linear] {
	// Convert from sRGB to linear
	return RGBA8[Linear]{
		R: SRGBConv8BitRGBFromSRGB(g.V),
		G: SRGBConv8BitRGBFromSRGB(g.V),
		B: SRGBConv8BitRGBFromSRGB(g.V),
		A: SRGBConv8BitAlphaFromSRGB(g.A),
	}
}

// RGBA8Pre creates a premultiplied RGBA8 color (renamed to avoid conflict)
func RGBA8Pre[CS any](r, g, b, a float64) RGBA8[CS] {
	c := RGBA8[CS]{
		R: basics.Int8u(r*a*255 + 0.5),
		G: basics.Int8u(g*a*255 + 0.5),
		B: basics.Int8u(b*a*255 + 0.5),
		A: basics.Int8u(a*255 + 0.5),
	}
	return c
}

// Generic conversion functions to match C++ AGG template pattern
// These follow the static convert() methods pattern from C++ AGG

// ConvertRGBA8Types converts between RGBA8 colorspaces
func ConvertRGBA8Types[CS1, CS2 any](dst *RGBA8[CS2], src RGBA8[CS1]) {
	// This is where we would dispatch based on colorspace types
	// For now, we implement the most common conversions
	switch any(dst).(type) {
	case *RGBA8[SRGB]:
		if linear, ok := any(src).(RGBA8[Linear]); ok {
			*dst = any(ConvertRGBA8LinearToSRGB(linear)).(RGBA8[CS2])
		}
	case *RGBA8[Linear]:
		if srgb, ok := any(src).(RGBA8[SRGB]); ok {
			*dst = any(ConvertRGBA8SRGBToLinear(srgb)).(RGBA8[CS2])
		}
	default:
		// Same colorspace or unsupported conversion
		*dst = RGBA8[CS2](src)
	}
}

// ConvertRGBAToRGBA8 converts from float64 RGBA to typed RGBA8
func ConvertRGBAToRGBA8[CS any](dst *RGBA8[CS], src RGBA) {
	switch any(dst).(type) {
	case *RGBA8[Linear]:
		*dst = any(RGBA8[Linear]{
			R: basics.Int8u(src.R*255 + 0.5),
			G: basics.Int8u(src.G*255 + 0.5),
			B: basics.Int8u(src.B*255 + 0.5),
			A: basics.Int8u(src.A*255 + 0.5),
		}).(RGBA8[CS])
	case *RGBA8[SRGB]:
		*dst = any(RGBA8[SRGB]{
			R: SRGBConvF32RGBToSRGB(float32(src.R)),
			G: SRGBConvF32RGBToSRGB(float32(src.G)),
			B: SRGBConvF32RGBToSRGB(float32(src.B)),
			A: SRGBConvF32AlphaToSRGB(float32(src.A)),
		}).(RGBA8[CS])
	}
}

// ConvertRGBA8ToRGBA converts from typed RGBA8 to float64 RGBA
func ConvertRGBA8ToRGBA[CS any](dst *RGBA, src RGBA8[CS]) {
	switch any(src).(type) {
	case RGBA8[Linear]:
		linear := any(src).(RGBA8[Linear])
		dst.R = float64(linear.R) / 255.0
		dst.G = float64(linear.G) / 255.0
		dst.B = float64(linear.B) / 255.0
		dst.A = float64(linear.A) / 255.0
	case RGBA8[SRGB]:
		srgb := any(src).(RGBA8[SRGB])
		dst.R = float64(SRGBConvF32RGBFromSRGB(srgb.R))
		dst.G = float64(SRGBConvF32RGBFromSRGB(srgb.G))
		dst.B = float64(SRGBConvF32RGBFromSRGB(srgb.B))
		dst.A = float64(SRGBConvF32AlphaFromSRGB(srgb.A))
	}
}
