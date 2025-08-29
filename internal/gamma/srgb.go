package gamma

import (
	"agg_go/internal/basics"
)

// ---- sRGB lookup tables (AGG sRGB_lut) ------------------------------------

// Float32 variant: dir/thresholds are float32; inverse uses unrolled binary search.
type SRGBLUTFloat struct {
	dir [256]float32
	inv [256]float32
}

func NewSRGBLUTFloat() *SRGBLUTFloat {
	l := &SRGBLUTFloat{}
	l.dir[0] = 0
	l.inv[0] = 0
	for i := 1; i <= 255; i++ {
		xf := float64(i) / 255.0
		// dir: sRGB -> linear
		l.dir[i] = float32(SRGBToLinear(xf))
		// inv table holds thresholds in linear domain at mid-code points
		l.inv[i] = float32(SRGBToLinear((float64(i) - 0.5) / 255.0))
	}
	return l
}

func (l *SRGBLUTFloat) Dir(v basics.Int8u) float32 {
	return l.dir[v]
}

func (l *SRGBLUTFloat) Inv(v float32) basics.Int8u {
	x := basics.Int8u(0)
	if v > l.inv[128] {
		x = 128
	}
	if v > l.inv[x+64] {
		x += 64
	}
	if v > l.inv[x+32] {
		x += 32
	}
	if v > l.inv[x+16] {
		x += 16
	}
	if v > l.inv[x+8] {
		x += 8
	}
	if v > l.inv[x+4] {
		x += 4
	}
	if v > l.inv[x+2] {
		x += 2
	}
	if v > l.inv[x+1] {
		x += 1
	}
	return x
}

// 16-bit variant: dir/thresholds are 16-bit; inverse uses unrolled binary search.
type SRGBLUT16 struct {
	dir [256]basics.Int16u
	inv [256]basics.Int16u
}

func NewSRGBLUT16() *SRGBLUT16 {
	l := &SRGBLUT16{}
	l.dir[0] = 0
	l.inv[0] = 0
	for i := 1; i <= 255; i++ {
		xf := float64(i) / 255.0
		l.dir[i] = basics.Int16u(65535.0*SRGBToLinear(xf) + 0.5)
		l.inv[i] = basics.Int16u(65535.0*SRGBToLinear((float64(i)-0.5)/255.0) + 0.5)
	}
	return l
}

func (l *SRGBLUT16) Dir(v basics.Int8u) basics.Int16u {
	return l.dir[v]
}

func (l *SRGBLUT16) Inv(v basics.Int16u) basics.Int8u {
	x := basics.Int8u(0)
	if v > l.inv[128] {
		x = 128
	}
	if v > l.inv[x+64] {
		x += 64
	}
	if v > l.inv[x+32] {
		x += 32
	}
	if v > l.inv[x+16] {
		x += 16
	}
	if v > l.inv[x+8] {
		x += 8
	}
	if v > l.inv[x+4] {
		x += 4
	}
	if v > l.inv[x+2] {
		x += 2
	}
	if v > l.inv[x+1] {
		x += 1
	}
	return x
}

// 8-bit variant: both directions are simple lookups (AGG specialization).
type SRGBLUT8 struct {
	dir [256]basics.Int8u
	inv [256]basics.Int8u
}

func NewSRGBLUT8() *SRGBLUT8 {
	l := &SRGBLUT8{}
	l.dir[0] = 0
	l.inv[0] = 0
	for i := 1; i <= 255; i++ {
		xf := float64(i) / 255.0
		l.dir[i] = basics.Int8u(255.0*SRGBToLinear(xf) + 0.5) // sRGB -> linear 8
		l.inv[i] = basics.Int8u(255.0*LinearToSRGB(xf) + 0.5) // linear 8 -> sRGB
	}
	return l
}

func (l *SRGBLUT8) Dir(v basics.Int8u) basics.Int8u { return l.dir[v] }
func (l *SRGBLUT8) Inv(v basics.Int8u) basics.Int8u { return l.inv[v] }

// ---- sRGB_conv<T> wrappers (AGG sRGB_conv) --------------------------------

// Singleton LUTs (match AGG’s static template member).
var (
	srgbLUTFloat = NewSRGBLUTFloat()
	srgbLUT16    = NewSRGBLUT16()
	srgbLUT8     = NewSRGBLUT8()
)

type SRGBConvFloat struct{}

func (SRGBConvFloat) RGBFromSRGB(x basics.Int8u) float32 { return srgbLUTFloat.Dir(x) }
func (SRGBConvFloat) RGBToSRGB(x float32) basics.Int8u   { return srgbLUTFloat.Inv(x) }
func (SRGBConvFloat) AlphaFromSRGB(x basics.Int8u) float32 {
	const y = 1.0 / 255.0
	return float32(float64(x) * y)
}

func (SRGBConvFloat) AlphaToSRGB(x float32) basics.Int8u {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 255
	}
	return basics.Int8u(0.5 + float64(x)*255.0)
}

type SRGBConv16 struct{}

func (SRGBConv16) RGBFromSRGB(x basics.Int8u) basics.Int16u { return srgbLUT16.Dir(x) }
func (SRGBConv16) RGBToSRGB(x basics.Int16u) basics.Int8u   { return srgbLUT16.Inv(x) }
func (SRGBConv16) AlphaFromSRGB(x basics.Int8u) basics.Int16u {
	return (basics.Int16u(x) << 8) | basics.Int16u(x)
}
func (SRGBConv16) AlphaToSRGB(x basics.Int16u) basics.Int8u { return basics.Int8u(x >> 8) }

type SRGBConv8 struct{}

func (SRGBConv8) RGBFromSRGB(x basics.Int8u) basics.Int8u   { return srgbLUT8.Dir(x) }
func (SRGBConv8) RGBToSRGB(x basics.Int8u) basics.Int8u     { return srgbLUT8.Inv(x) }
func (SRGBConv8) AlphaFromSRGB(x basics.Int8u) basics.Int8u { return x }
func (SRGBConv8) AlphaToSRGB(x basics.Int8u) basics.Int8u   { return x }
