package color

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
)

// Golden parity tests for the fixed-point helpers of the 8- and 16-bit colour
// types. The expected values come from the original AGG 2.6 C++ code via
// testdata/fixedpoint_oracle.cpp (see that file for regeneration). Every loop
// below mirrors the oracle loop that produced the corresponding hash.

type fnv64 uint64

func newFNV64() fnv64 { return 14695981039346656037 }

func (h *fnv64) b(v uint8) {
	*h ^= fnv64(v)
	*h *= 1099511628211
}

func (h *fnv64) u8(v basics.Int8u) { h.b(v) }

func (h *fnv64) u16(v basics.Int16u) {
	h.b(uint8(v))
	h.b(uint8(v >> 8))
}

type lcg64 uint64

func (r *lcg64) next16() basics.Int16u {
	*r = *r*6364136223846793005 + 1442695040888963407
	return basics.Int16u(uint64(*r) >> 48)
}

var e16 = []basics.Int16u{
	0, 1, 2, 127, 128, 255, 256, 257, 32767,
	32768, 32769, 65279, 65280, 65534, 65535,
}

const (
	lerpSamples16   = 1 << 20
	fromDoubleSteps = 1 << 20
	gradientSteps   = 4096
)

func sample16() []basics.Int16u {
	s := append([]basics.Int16u(nil), e16...)
	r := lcg64(1)
	for len(s) < 64 {
		s = append(s, r.next16())
	}
	return s
}

func triples16(f func(p, q, a basics.Int16u)) {
	for _, p := range e16 {
		for _, q := range e16 {
			for _, a := range e16 {
				f(p, q, a)
			}
		}
	}
	r := lcg64(2)
	for n := 0; n < lerpSamples16; n++ {
		p := r.next16()
		q := r.next16()
		a := r.next16()
		f(p, q, a)
	}
}

func checkHash(t *testing.T, name string, got fnv64, want uint64) {
	t.Helper()
	if uint64(got) != want {
		t.Errorf("%s: hash 0x%016x, C++ golden 0x%016x", name, uint64(got), want)
	}
}

func cover8to16(c basics.Int8u) basics.Int16u {
	return basics.Int16u(c)<<8 | basics.Int16u(c)
}

func TestFixedPointParity8Bit(t *testing.T) {
	m8, gm8, d8, sc8 := newFNV64(), newFNV64(), newFNV64(), newFNV64()
	l8, gl8, pl8, gpl8 := newFNV64(), newFNV64(), newFNV64(), newFNV64()
	md8, gmd8, lum8 := newFNV64(), newFNV64(), newFNV64()
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			ua, ub := basics.Int8u(a), basics.Int8u(b)
			m8.u8(RGBA8Multiply(ua, ub))
			gm8.u8(Gray8Multiply(ua, ub))
			d8.u8(RGBA8Demultiply(ua, ub))
			sc8.u8(RGBA8ScaleCover(ua, ub))

			c := NewRGBA8[Linear](ua, ua, ua, ub)
			c.Demultiply()
			md8.u8(c.R)
			g := NewGray8WithAlpha[Linear](ua, ub)
			g.Demultiply()
			gmd8.u8(g.V)

			for x := 0; x < 256; x++ {
				ux := basics.Int8u(x)
				l8.u8(RGBA8Lerp(ua, ub, ux))
				gl8.u8(Gray8Lerp(ua, ub, ux))
				pl8.u8(RGBA8Prelerp(ua, ub, ux))
				gpl8.u8(Gray8Prelerp(ua, ub, ux))
				lum8.u8(LuminanceFromRGBA8Linear(NewRGBA8[Linear](ua, ub, ux, 255)))
			}
		}
	}
	checkHash(t, "rgba8::multiply", m8, goldenHashRGBA8Multiply)
	checkHash(t, "gray8::multiply", gm8, goldenHashGray8Multiply)
	checkHash(t, "rgba8::demultiply", d8, goldenHashRGBA8Demultiply)
	checkHash(t, "rgba8::scale_cover", sc8, goldenHashRGBA8ScaleCover)
	checkHash(t, "rgba8::lerp", l8, goldenHashRGBA8Lerp)
	checkHash(t, "gray8::lerp", gl8, goldenHashGray8Lerp)
	checkHash(t, "rgba8::prelerp", pl8, goldenHashRGBA8Prelerp)
	checkHash(t, "gray8::prelerp", gpl8, goldenHashGray8Prelerp)
	checkHash(t, "rgba8::demultiply()", md8, goldenHashRGBA8MemberDemultiply)
	checkHash(t, "gray8::demultiply()", gmd8, goldenHashGray8MemberDemultiply)
	checkHash(t, "gray8::luminance(rgba8)", lum8, goldenHashGray8LuminanceRGBA8)

	// mult_cover(a, b) == multiply(a, b) for 8-bit types.
	mc8 := newFNV64()
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			mc8.u8(RGBA8MultCover(basics.Int8u(a), basics.Int8u(b)))
		}
	}
	checkHash(t, "rgba8::mult_cover", mc8, goldenHashRGBA8Multiply)

	// RGB8.Luminance shares gray8::luminance(rgba8) weights.
	rl8 := newFNV64()
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			for x := 0; x < 256; x++ {
				rl8.u8(NewRGB8[Linear](basics.Int8u(a), basics.Int8u(b), basics.Int8u(x)).Luminance())
			}
		}
	}
	checkHash(t, "RGB8.Luminance", rl8, goldenHashGray8LuminanceRGBA8)

	fd8, gfd8 := newFNV64(), newFNV64()
	for i := 0; i <= fromDoubleSteps; i++ {
		d := float64(i) / float64(fromDoubleSteps)
		fd8.u8(RGBA8FromDouble(d))
		gfd8.u8(ConvertGray8FromRGBA[Linear](RGBA{A: d}).A)
	}
	checkHash(t, "rgba8::from_double", fd8, goldenHashRGBA8FromDouble)
	checkHash(t, "gray8(rgba).a", gfd8, goldenHashGray8FromDouble)

	type gp struct{ v1, a1, v2, a2 basics.Int8u }
	pairs := []gp{
		{0, 0, 255, 255},
		{255, 255, 0, 0},
		{10, 20, 200, 250},
		{200, 250, 10, 20},
		{128, 64, 127, 192},
		{0, 255, 255, 0},
		{77, 77, 77, 77},
		{1, 254, 254, 1},
	}
	gg8 := newFNV64()
	for _, p := range pairs {
		c1 := NewGray8WithAlpha[Linear](p.v1, p.a1)
		c2 := NewGray8WithAlpha[Linear](p.v2, p.a2)
		for i := 0; i <= gradientSteps; i++ {
			r := c1.Gradient(c2, float64(i)/float64(gradientSteps))
			gg8.u8(r.V)
			gg8.u8(r.A)
		}
	}
	checkHash(t, "gray8::gradient", gg8, goldenHashGray8Gradient)
}

func TestFixedPointParity16Bit(t *testing.T) {
	s16 := sample16()

	m16, gm16, dd16, dw16 := newFNV64(), newFNV64(), newFNV64(), newFNV64()
	md16, gmd16 := newFNV64(), newFNV64()
	mc16, gmc16, sc16 := newFNV64(), newFNV64(), newFNV64()
	for ai := 0; ai < 65536; ai++ {
		a := basics.Int16u(ai)
		for _, b := range s16 {
			m16.u16(RGBA16Multiply(a, b))
			gm16.u16(Gray16Multiply(a, b))
			w := RGBA16Demultiply(a, b)
			if uint64(a)*65535+uint64(b>>1) <= 1<<31-1 {
				dd16.u16(w)
			}
			dw16.u16(w)

			c := NewRGBA16[Linear](b, b, b, a)
			c.Demultiply()
			md16.u16(c.R)
			g := NewGray16WithAlpha[Linear](b, a)
			g.Demultiply()
			gmd16.u16(g.V)
		}
		for cv := 0; cv < 256; cv++ {
			c8 := basics.Int8u(cv)
			mc16.u16(RGBA16MultCover(a, cover8to16(c8)))
			gmc16.u16(Gray16Multiply(a, cover8to16(c8)))
			sc16.u8(RGBA16ScaleCover(c8, a))
		}
	}
	checkHash(t, "rgba16::multiply", m16, goldenHashRGBA16Multiply)
	checkHash(t, "gray16::multiply", gm16, goldenHashGray16Multiply)
	checkHash(t, "rgba16::demultiply (defined region)", dd16, goldenHashRGBA16DemultiplyDefined)
	checkHash(t, "rgba16::demultiply (overflow-free)", dw16, goldenHashRGBA16DemultiplyWide)
	checkHash(t, "rgba16::demultiply()", md16, goldenHashRGBA16MemberDemultiply)
	checkHash(t, "gray16::demultiply()", gmd16, goldenHashGray16MemberDemultiply)
	checkHash(t, "rgba16::mult_cover", mc16, goldenHashRGBA16MultCover)
	checkHash(t, "gray16::mult_cover", gmc16, goldenHashGray16MultCover)
	checkHash(t, "rgba16::scale_cover", sc16, goldenHashRGBA16ScaleCover)

	l16, gl16, pl16, gpl16 := newFNV64(), newFNV64(), newFNV64(), newFNV64()
	triples16(func(p, q, a basics.Int16u) {
		l16.u16(RGBA16Lerp(p, q, a))
		gl16.u16(Gray16Lerp(p, q, a))
		pl16.u16(RGBA16Prelerp(p, q, a))
		gpl16.u16(Gray16Prelerp(p, q, a))
	})
	checkHash(t, "rgba16::lerp", l16, goldenHashRGBA16Lerp)
	checkHash(t, "gray16::lerp", gl16, goldenHashGray16Lerp)
	checkHash(t, "rgba16::prelerp", pl16, goldenHashRGBA16Prelerp)
	checkHash(t, "gray16::prelerp", gpl16, goldenHashGray16Prelerp)

	lum16 := newFNV64()
	for _, r := range s16 {
		for _, g := range s16 {
			for _, b := range s16 {
				lum16.u16(NewRGB16[Linear](r, g, b).Luminance())
			}
		}
	}
	checkHash(t, "RGB16.Luminance vs gray16::luminance(rgba16)", lum16, goldenHashGray16LuminanceRGBA16)

	fd16, gfd16 := newFNV64(), newFNV64()
	for i := 0; i <= fromDoubleSteps; i++ {
		d := float64(i) / float64(fromDoubleSteps)
		fd16.u16(RGBA16FromDouble(d))
		gfd16.u16(ConvertGray16FromRGBA[Linear](RGBA{A: d}).A)
	}
	checkHash(t, "rgba16::from_double", fd16, goldenHashRGBA16FromDouble)
	checkHash(t, "gray16(rgba).a", gfd16, goldenHashGray16FromDouble)

	type gp struct{ v1, a1, v2, a2 basics.Int16u }
	pairs := []gp{
		{0, 0, 65535, 65535},
		{65535, 65535, 0, 0},
		{1000, 2000, 60000, 65000},
		{60000, 65000, 1000, 2000},
		{32768, 16384, 32767, 49152},
		{0, 65535, 65535, 0},
		{12345, 12345, 12345, 12345},
		{1, 65534, 65534, 1},
	}
	gg16 := newFNV64()
	for _, p := range pairs {
		c1 := NewGray16WithAlpha[Linear](p.v1, p.a1)
		c2 := NewGray16WithAlpha[Linear](p.v2, p.a2)
		for i := 0; i <= gradientSteps; i++ {
			r := c1.Gradient(c2, float64(i)/float64(gradientSteps))
			gg16.u16(r.V)
			gg16.u16(r.A)
		}
	}
	checkHash(t, "gray16::gradient", gg16, goldenHashGray16Gradient)
}

func TestFixedPointSpotValues(t *testing.T) {
	for _, s := range goldenRGBA16LerpSpots {
		p, q, a := basics.Int16u(s[0]), basics.Int16u(s[1]), basics.Int16u(s[2])
		if got := RGBA16Lerp(p, q, a); got != basics.Int16u(s[3]) {
			t.Errorf("RGBA16Lerp(%d,%d,%d) = %d, C++ %d", p, q, a, got, s[3])
		}
		if got := Gray16Lerp(p, q, a); got != basics.Int16u(s[3]) {
			t.Errorf("Gray16Lerp(%d,%d,%d) = %d, C++ %d", p, q, a, got, s[3])
		}
		if got := RGBA16Prelerp(p, q, a); got != basics.Int16u(s[4]) {
			t.Errorf("RGBA16Prelerp(%d,%d,%d) = %d, C++ %d", p, q, a, got, s[4])
		}
	}
	for _, s := range goldenRGBA16MulSpots {
		a, b := basics.Int16u(s[0]), basics.Int16u(s[1])
		if got := RGBA16Multiply(a, b); got != basics.Int16u(s[2]) {
			t.Errorf("RGBA16Multiply(%d,%d) = %d, C++ %d", a, b, got, s[2])
		}
		if got := RGBA16Demultiply(a, b); got != basics.Int16u(s[3]) {
			t.Errorf("RGBA16Demultiply(%d,%d) = %d, C++ %d", a, b, got, s[3])
		}
	}
	for _, s := range goldenRGBA16CoverSpots {
		v, cv := basics.Int16u(s[0]), basics.Int8u(s[1])
		if got := RGBA16MultCover(v, cover8to16(cv)); got != basics.Int16u(s[2]) {
			t.Errorf("RGBA16MultCover(%d, cover %d) = %d, C++ %d", v, cv, got, s[2])
		}
		if got := RGBA16ScaleCover(cv, v); got != basics.Int8u(s[3]) {
			t.Errorf("RGBA16ScaleCover(%d, %d) = %d, C++ %d", cv, v, got, s[3])
		}
	}
	for _, s := range goldenRGBA8LerpSpots {
		p, q, a := basics.Int8u(s[0]), basics.Int8u(s[1]), basics.Int8u(s[2])
		if got := RGBA8Lerp(p, q, a); got != s[3] {
			t.Errorf("RGBA8Lerp(%d,%d,%d) = %d, C++ %d", p, q, a, got, s[3])
		}
		if got := RGBA8Prelerp(p, q, a); got != s[4] {
			t.Errorf("RGBA8Prelerp(%d,%d,%d) = %d, C++ %d", p, q, a, got, s[4])
		}
	}
	for _, s := range goldenGrayGradientSpots {
		k := float64(s[0]) / 1e6
		g8 := NewGray8WithAlpha[Linear](10, 20).Gradient(NewGray8WithAlpha[Linear](200, 250), k)
		if uint32(g8.V) != s[1] || uint32(g8.A) != s[2] {
			t.Errorf("Gray8.Gradient(k=%v) = (%d,%d), C++ (%d,%d)", k, g8.V, g8.A, s[1], s[2])
		}
		g16 := NewGray16WithAlpha[Linear](1000, 2000).Gradient(NewGray16WithAlpha[Linear](60000, 65000), k)
		if uint32(g16.V) != s[3] || uint32(g16.A) != s[4] {
			t.Errorf("Gray16.Gradient(k=%v) = (%d,%d), C++ (%d,%d)", k, g16.V, g16.A, s[3], s[4])
		}
	}
}

func TestFixedPointAddParity(t *testing.T) {
	add8, gadd8 := newFNV64(), newFNV64()
	for x := 0; x < 256; x += 5 {
		for y := 0; y < 256; y += 5 {
			for m := 0; m < 2; m++ {
				ux, uy := basics.Int8u(x), basics.Int8u(y)
				ca := uy
				if m == 1 {
					ca = 255
				}
				for cv := 0; cv < 256; cv++ {
					d := NewRGBA8[Linear](ux, ux, ux, ux)
					d.AddWithCover(NewRGBA8[Linear](uy, uy, uy, ca), basics.Int8u(cv))
					add8.u8(d.R)
					add8.u8(d.A)
					g := NewGray8WithAlpha[Linear](ux, ux)
					g.Add(NewGray8WithAlpha[Linear](uy, ca), basics.Int8u(cv))
					gadd8.u8(g.V)
					gadd8.u8(g.A)
				}
			}
		}
	}
	checkHash(t, "rgba8::add", add8, goldenHashRGBA8Add)
	checkHash(t, "gray8::add", gadd8, goldenHashGray8Add)

	s16 := sample16()
	add16, gadd16 := newFNV64(), newFNV64()
	for _, x := range s16 {
		for _, y := range s16 {
			for m := 0; m < 2; m++ {
				ca := y
				if m == 1 {
					ca = 65535
				}
				for cv := 0; cv < 256; cv++ {
					d := NewRGBA16[Linear](x, x, x, x)
					d.AddWithCover(NewRGBA16[Linear](y, y, y, ca), basics.Int8u(cv))
					add16.u16(d.R)
					add16.u16(d.A)
					g := NewGray16WithAlpha[Linear](x, x)
					g.Add(NewGray16WithAlpha[Linear](y, ca), basics.Int8u(cv))
					gadd16.u16(g.V)
					gadd16.u16(g.A)
				}
			}
		}
	}
	checkHash(t, "rgba16::add", add16, goldenHashRGBA16Add)
	checkHash(t, "gray16::add", gadd16, goldenHashGray16Add)
}

func TestFixedPointTrivialHelpers(t *testing.T) {
	if RGBA8EmptyValue() != 0 || RGBA8FullValue() != 255 || Gray8EmptyValue() != 0 || Gray8FullValue() != 255 {
		t.Error("8-bit empty/full values differ from C++ (0 / base_mask)")
	}
	if RGBA16EmptyValue() != 0 || RGBA16FullValue() != 65535 || Gray16EmptyValue() != 0 || Gray16FullValue() != 65535 {
		t.Error("16-bit empty/full values differ from C++ (0 / base_mask)")
	}
	for i := 0; i < 256; i++ {
		if RGBA8Invert(basics.Int8u(i)) != basics.Int8u(255-i) {
			t.Fatalf("RGBA8Invert(%d) != %d", i, 255-i)
		}
	}
	for i := 0; i < 65536; i++ {
		if RGBA16Invert(basics.Int16u(i)) != basics.Int16u(65535-i) {
			t.Fatalf("RGBA16Invert(%d) != %d", i, 65535-i)
		}
	}
}
