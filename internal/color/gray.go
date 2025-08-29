package color

import (
	"agg_go/internal/basics"
)

// BT.709 integer weights (AGG-compatible; 55+184+18 = 257, >>8)
const (
	bt709R = 55
	bt709G = 184
	bt709B = 18
)

// Linear 8-bit path: inputs are already linear
func LuminanceFromRGBA8Linear(c RGBA8[Linear]) basics.Int8u {
	return basics.Int8u((uint32(c.R)*bt709R + uint32(c.G)*bt709G + uint32(c.B)*bt709B) >> 8)
}

// sRGB 8-bit path: linearize first via your LUTs, then apply BT.709
func LuminanceFromRGBA8SRGB(c RGBA8[SRGB]) basics.Int8u {
	r := srgb8ToLinear8(c.R)
	g := srgb8ToLinear8(c.G)
	b := srgb8ToLinear8(c.B)
	return basics.Int8u((uint32(r)*bt709R + uint32(g)*bt709G + uint32(b)*bt709B) >> 8)
}

func LuminanceFromRGBA(c RGBA) float64 {
	return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
}
