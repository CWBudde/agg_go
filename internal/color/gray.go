package color

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// BT.709 integer weights used by AGG-style 8-bit luminance conversion.
const (
	bt709R = 55
	bt709G = 184
	bt709B = 18
)

// LuminanceFromRGBA8Linear computes grayscale luminance from an already-linear
// RGBA8 color using AGG-compatible BT.709 weights.
func LuminanceFromRGBA8Linear(c RGBA8[Linear]) basics.Int8u {
	return basics.Int8u((uint32(c.R)*bt709R + uint32(c.G)*bt709G + uint32(c.B)*bt709B) >> 8)
}

// LuminanceFromRGBA8SRGB first linearizes RGB channels through the sRGB LUT and
// then applies the same BT.709 luminance weights.
func LuminanceFromRGBA8SRGB(c RGBA8[SRGB]) basics.Int8u {
	r := srgb8ToLinear8(c.R)
	g := srgb8ToLinear8(c.G)
	b := srgb8ToLinear8(c.B)
	return basics.Int8u((uint32(r)*bt709R + uint32(g)*bt709G + uint32(b)*bt709B) >> 8)
}

// LuminanceFromRGBA computes floating-point BT.709 luminance.
func LuminanceFromRGBA(c RGBA) float64 {
	return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
}
