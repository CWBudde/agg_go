package color

import (
	"math"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

func almostEqualRGBAComponent(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}

func assertRGBAEqual(t *testing.T, got, want RGBA) {
	t.Helper()
	if !almostEqualRGBAComponent(got.R, want.R) ||
		!almostEqualRGBAComponent(got.G, want.G) ||
		!almostEqualRGBAComponent(got.B, want.B) ||
		!almostEqualRGBAComponent(got.A, want.A) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestRGBAFloatOperations(t *testing.T) {
	c := NewRGBA(0.25, 0.5, 0.75, 0.5)
	if c != (RGBA{R: 0.25, G: 0.5, B: 0.75, A: 0.5}) {
		t.Fatalf("NewRGBA() = %+v", c)
	}

	if got := c.Gradient(RGBA{R: 0.75, G: 0.0, B: 0.25, A: 1.0}, 0.5); got != (RGBA{R: 0.5, G: 0.25, B: 0.5, A: 0.75}) {
		t.Fatalf("Gradient() = %+v", got)
	}

	assertRGBAEqual(t, c.Add(RGBA{R: 0.1, G: 0.2, B: 0.3, A: 0.4}), RGBA{R: 0.35, G: 0.7, B: 1.05, A: 0.9})
	assertRGBAEqual(t, c.Subtract(RGBA{R: 0.1, G: 0.2, B: 0.3, A: 0.4}), RGBA{R: 0.15, G: 0.3, B: 0.45, A: 0.1})
	assertRGBAEqual(t, c.Multiply(RGBA{R: 0.4, G: 0.2, B: 0.5, A: 0.5}), RGBA{R: 0.1, G: 0.1, B: 0.375, A: 0.25})
	assertRGBAEqual(t, c.Scale(2), RGBA{R: 0.5, G: 1.0, B: 1.5, A: 1.0})
}

func TestRGBAFloatMutationHelpers(t *testing.T) {
	c := NewRGBA(0.3, 0.4, 0.5, 0.6)
	c.AddAssign(RGBA{R: 0.1, G: 0.2, B: 0.3, A: 0.4}).MultiplyAssign(0.5)
	assertRGBAEqual(t, c, RGBA{R: 0.2, G: 0.3, B: 0.4, A: 0.5})

	c.Clear()
	if c != NoColor() {
		t.Fatalf("Clear() = %+v", c)
	}

	c = NewRGBA(0.2, 0.3, 0.4, 0.5)
	c.Transparent()
	if c != (RGBA{R: 0.2, G: 0.3, B: 0.4, A: 0.0}) {
		t.Fatalf("Transparent() = %+v", c)
	}

	c.Opacity(-1)
	if c.GetOpacity() != 0 {
		t.Fatalf("Opacity(-1) = %v", c.GetOpacity())
	}
	c.Opacity(2)
	if c.GetOpacity() != 1 {
		t.Fatalf("Opacity(2) = %v", c.GetOpacity())
	}
	c.Opacity(0.25)
	if c.GetOpacity() != 0.25 {
		t.Fatalf("Opacity(0.25) = %v", c.GetOpacity())
	}
}

func TestRGBAPremultiplyAndDemultiply(t *testing.T) {
	c := NewRGBA(0.8, 0.4, 0.2, 0.5)
	c.Premultiply()
	if c != (RGBA{R: 0.4, G: 0.2, B: 0.1, A: 0.5}) {
		t.Fatalf("Premultiply() = %+v", c)
	}
	c.Demultiply()
	if !almostEqualRGBAComponent(c.R, 0.8) || !almostEqualRGBAComponent(c.G, 0.4) || !almostEqualRGBAComponent(c.B, 0.2) || !almostEqualRGBAComponent(c.A, 0.5) {
		t.Fatalf("Demultiply() = %+v", c)
	}

	zero := NewRGBA(0.8, 0.4, 0.2, 0)
	zero.Premultiply()
	if zero != (RGBA{A: 0}) {
		t.Fatalf("Premultiply zero alpha = %+v", zero)
	}
	zero = NewRGBA(0.8, 0.4, 0.2, 0)
	zero.Demultiply()
	if zero != (RGBA{A: 0}) {
		t.Fatalf("Demultiply zero alpha = %+v", zero)
	}
}

func TestRGBAPremultiplyAlphaAndHelpers(t *testing.T) {
	c := NewRGBA(0.8, 0.4, 0.2, 0.5)
	c.PremultiplyAlpha(0.25)
	if !almostEqualRGBAComponent(c.R, 0.4) || !almostEqualRGBAComponent(c.G, 0.2) || !almostEqualRGBAComponent(c.B, 0.1) || !almostEqualRGBAComponent(c.A, 0.5) {
		t.Fatalf("PremultiplyAlpha() = %+v", c)
	}

	zero := NewRGBA(0.8, 0.4, 0.2, 0.5)
	zero.PremultiplyAlpha(0)
	if zero != (RGBA{A: 0}) {
		t.Fatalf("PremultiplyAlpha(0) = %+v", zero)
	}

	pre := RGBAPre(0.8, 0.4, 0.2, 0.5)
	if !almostEqualRGBAComponent(pre.R, 0.4) || !almostEqualRGBAComponent(pre.G, 0.2) || !almostEqualRGBAComponent(pre.B, 0.1) || !almostEqualRGBAComponent(pre.A, 0.5) {
		t.Fatalf("RGBAPre() = %+v", pre)
	}
}

func TestRGBAWavelengthAndConversionHelpers(t *testing.T) {
	red := FromWavelength(700, 1.0)
	if red.R <= 0 || red.G != 0 || red.B != 0 || red.A != 1 {
		t.Fatalf("FromWavelength(700) = %+v", red)
	}

	outside := NewRGBAFromWavelength(800, 1.0)
	if outside.R != 0 || outside.G != 0 || outside.B != 0 || outside.A != 1 {
		t.Fatalf("NewRGBAFromWavelength(800) = %+v", outside)
	}

	g := Gray8[SRGB]{V: 123, A: 231}
	if got := ConvertGray8FromSRGBToLinear(g); got != ConvertGray8SRGBToLinear(g) {
		t.Fatalf("ConvertGray8FromSRGBToLinear() = %+v, want %+v", got, ConvertGray8SRGBToLinear(g))
	}

	gl := Gray8[Linear]{V: 45, A: 67}
	if got := ConvertGray8FromLinearToSRGB(gl); got != ConvertGray8LinearToSRGB(gl) {
		t.Fatalf("ConvertGray8FromLinearToSRGB() = %+v, want %+v", got, ConvertGray8LinearToSRGB(gl))
	}

	if got := MakeRGBA8[Linear](1, 2, 3, 4); got != (RGBA8[Linear]{R: 1, G: 2, B: 3, A: 4}) {
		t.Fatalf("MakeRGBA8() = %+v", got)
	}
	if got := MakeSRGBA8(5, 6, 7, 8); got != (RGBA8[SRGB]{R: 5, G: 6, B: 7, A: 8}) {
		t.Fatalf("MakeSRGBA8() = %+v", got)
	}
	if got := MakeRGBA16[Linear](9, 10, 11, 12); got != (RGBA16[Linear]{R: 9, G: 10, B: 11, A: 12}) {
		t.Fatalf("MakeRGBA16() = %+v", got)
	}
	if got := MakeRGBA32[Linear](0.1, 0.2, 0.3, 0.4); got != (RGBA32[Linear]{R: 0.1, G: 0.2, B: 0.3, A: 0.4}) {
		t.Fatalf("MakeRGBA32() = %+v", got)
	}
	if got := RGBA8Pre[Linear](1.0, 0.5, 0.25, 0.5); got != (RGBA8[Linear]{R: 128, G: 64, B: 32, A: 128}) {
		t.Fatalf("RGBA8Pre() = %+v", got)
	}

	if got := MakeRGBA8FromGray8Linear[Linear](Gray8[Linear]{V: 12, A: 34}); got != (RGBA8[Linear]{R: 12, G: 12, B: 12, A: 34}) {
		t.Fatalf("MakeRGBA8FromGray8Linear() = %+v", got)
	}
	if got := MakeRGBA8FromGray8SRGB[SRGB](Gray8[SRGB]{V: 56, A: 78}); got != (RGBA8[SRGB]{R: 56, G: 56, B: 56, A: 78}) {
		t.Fatalf("MakeRGBA8FromGray8SRGB() = %+v", got)
	}
	if got := MakeSRGBA8FromGray8Linear[SRGB](Gray8[Linear]{V: 128, A: 99}); got.A != 99 || got.R != got.G || got.G != got.B {
		t.Fatalf("MakeSRGBA8FromGray8Linear() = %+v", got)
	}
	if got := MakeRGBA8FromGray8SRGB_ToLinear[Linear](Gray8[SRGB]{V: 128, A: 77}); got.A != 77 || got.R != got.G || got.G != got.B {
		t.Fatalf("MakeRGBA8FromGray8SRGB_ToLinear() = %+v", got)
	}

	linearRGB := RGB8[Linear]{R: 10, G: 20, B: 30}
	if got := RGB8ToSRGB(linearRGB); got != ConvertRGB8LinearToSRGB(linearRGB) {
		t.Fatalf("RGB8ToSRGB() = %+v, want %+v", got, ConvertRGB8LinearToSRGB(linearRGB))
	}
	srgbRGB := RGB8[SRGB]{R: 40, G: 50, B: 60}
	if got := RGB8ToLinear(srgbRGB); got != ConvertRGB8SRGBToLinear(srgbRGB) {
		t.Fatalf("RGB8ToLinear() = %+v, want %+v", got, ConvertRGB8SRGBToLinear(srgbRGB))
	}

	if got := linearF32ToSrgb8(0.5); got != basics.Int8u(ConvertToSRGB(0.5)*255+0.5) {
		t.Fatalf("linearF32ToSrgb8() = %d", got)
	}
}
