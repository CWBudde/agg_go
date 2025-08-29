package color

import (
	"testing"

	"agg_go/internal/basics"
)

// Helpers
func absI32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// One “8-bit replicated” LSB on 16-bit is 257 (0x0101)
const oneReplicated8to16 = 257

// ----- Constructors, flags, and basic methods -----

func TestGray16_ConstructorsAndFlags(t *testing.T) {
	g := NewGray16[Linear](32768)
	if g.V != 32768 {
		t.Errorf("NewGray16: V=%d, want 32768", g.V)
	}
	if g.A != Gray16BaseMask {
		t.Errorf("NewGray16: A=%d, want %d", g.A, Gray16BaseMask)
	}
	if g.IsTransparent() {
		t.Errorf("NewGray16: should not be transparent when A=%d", g.A)
	}
	if !g.IsOpaque() {
		t.Errorf("NewGray16: should be opaque when A=%d", g.A)
	}

	g2 := NewGray16WithAlpha[Linear](12345, 54321)
	if g2.V != 12345 || g2.A != 54321 {
		t.Errorf("NewGray16WithAlpha: got (V=%d,A=%d), want (12345,54321)", g2.V, g2.A)
	}
	if g2.IsTransparent() {
		t.Errorf("NewGray16WithAlpha: should not be transparent when A=%d", g2.A)
	}
	if g2.IsOpaque() {
		t.Errorf("NewGray16WithAlpha: should not be opaque when A=%d", g2.A)
	}

	g2.Transparent()
	if g2.A != 0 || g2.V != 12345 {
		t.Errorf("Transparent(): got (V=%d,A=%d), want (12345,0)", g2.V, g2.A)
	}
	if !g2.IsTransparent() {
		t.Errorf("IsTransparent() should be true after Transparent()")
	}

	g2.Clear()
	if g2.V != 0 || g2.A != 0 {
		t.Errorf("Clear(): got (V=%d,A=%d), want (0,0)", g2.V, g2.A)
	}
}

func TestGray16_BoundaryValues(t *testing.T) {
	z := NewGray16WithAlpha[Linear](0, 0)
	if !z.IsTransparent() {
		t.Errorf("A=0 must be transparent")
	}
	if z.IsOpaque() {
		t.Errorf("A=0 must not be opaque")
	}
	f := NewGray16WithAlpha[Linear](Gray16BaseMask, Gray16BaseMask)
	if f.IsTransparent() {
		t.Errorf("A=full must not be transparent")
	}
	if !f.IsOpaque() {
		t.Errorf("A=full must be opaque")
	}
}

// ----- Conversion to float and back (linear path) -----

func TestGray16_ConvertToRGBA_And_Back(t *testing.T) {
	// Converting Gray16 -> RGBA -> Gray16 should recover within 1 LSB (rounding)
	cases := []Gray16[Linear]{
		NewGray16WithAlpha[Linear](0, 0),
		NewGray16WithAlpha[Linear](1, 1),
		NewGray16WithAlpha[Linear](12345, 23456),
		NewGray16WithAlpha[Linear](32768, 40000),
		NewGray16WithAlpha[Linear](65535, 65535),
	}
	for _, g := range cases {
		r := g.ConvertToRGBA()                   // linear floats
		back := ConvertGray16FromRGBA[Linear](r) // uses BT.709 luminance on linear floats

		if d := absI32(int32(back.V) - int32(g.V)); d > 1 {
			t.Fatalf("Gray16->RGBA->Gray16 drift >1 LSB for V: orig=%d back=%d (d=%d)", g.V, back.V, d)
		}
		if d := absI32(int32(back.A) - int32(g.A)); d > 1 {
			t.Fatalf("Gray16->RGBA->Gray16 drift >1 LSB for A: orig=%d back=%d (d=%d)", g.A, back.A, d)
		}
	}
}

// ----- Luminance (BT.709) correctness on linear floats -----

func TestGray16_FromLinearRGBA_BT709(t *testing.T) {
	// These are linear RGBA samples; expected V = round((0.2126R + 0.7152G + 0.0722B) * 65535)
	type S struct{ R, G, B, A float64 }
	samples := []S{
		{0, 0, 0, 1},
		{1, 1, 1, 1},
		{0.5, 0.5, 0.5, 0.75},
		{0.2, 0.7, 0.1, 0.25},
		{0.9, 0.1, 0.3, 0.6},
	}

	for _, s := range samples {
		lum := 0.2126*s.R + 0.7152*s.G + 0.0722*s.B
		expV := basics.Int16u(lum*65535 + 0.5)
		expA := basics.Int16u(s.A*65535 + 0.5)

		got := ConvertGray16FromRGBA[Linear](RGBA(s))
		if d := absI32(int32(got.V) - int32(expV)); d > 1 {
			t.Fatalf("BT.709 V mismatch: exp=%d got=%d (d=%d) for sample=%+v", expV, got.V, d, s)
		}
		if d := absI32(int32(got.A) - int32(expA)); d > 1 {
			t.Fatalf("Alpha mismatch: exp=%d got=%d (d=%d) for sample=%+v", expA, got.A, d, s)
		}
	}
}

// ----- RGBA16 conversion (integer triplet) -----

func TestGray16_ConvertToRGBA16(t *testing.T) {
	g := NewGray16WithAlpha[Linear](12345, 54321)
	r16 := g.ConvertToRGBA16()
	if r16.R != g.V || r16.G != g.V || r16.B != g.V || r16.A != g.A {
		t.Fatalf("ConvertToRGBA16 mismatch: got %+v want V=%d A=%d", r16, g.V, g.A)
	}
}

// ----- sRGB <-> Linear conversions on Gray16 (via 8-bit tables) -----

func TestGray16_SRGB_Linear_Conversions(t *testing.T) {
	// sRGB middle gray: 186 in 8-bit encodes ~0.5 linear.
	// As 16-bit replicated: 0xBABA (186<<8 | 186) = 47802.
	srgbMid16 := basics.Int16u(0xBABA)
	linHalf16 := basics.Int16u(0x8080) // 0.5 replicated to 16-bit is ~0x8080 (= 32896)

	// SRGB -> Linear should land near 0x8080 (within one replicated 8->16 LSB)
	gS := Gray16[SRGB]{V: srgbMid16, A: Gray16BaseMask}
	gL := ConvertGray16SRGBToLinear(gS)
	if d := absI32(int32(gL.V) - int32(linHalf16)); d > oneReplicated8to16 {
		t.Fatalf("sRGB->Linear V too far: got=0x%04X want≈0x%04X (|d|=%d)", gL.V, linHalf16, d)
	}
	if gL.A != gS.A {
		t.Fatalf("sRGB->Linear alpha changed: got=%d want=%d", gL.A, gS.A)
	}

	// Linear -> SRGB should return near 0xBABA
	gL2 := Gray16[Linear]{V: linHalf16, A: Gray16BaseMask}
	gS2 := ConvertGray16LinearToSRGB(gL2)
	if d := absI32(int32(gS2.V) - int32(srgbMid16)); d > oneReplicated8to16 {
		t.Fatalf("Linear->sRGB V too far: got=0x%04X want≈0x%04X (|d|=%d)", gS2.V, srgbMid16, d)
	}
	if gS2.A != gL2.A {
		t.Fatalf("Linear->sRGB alpha changed: got=%d want=%d", gS2.A, gL2.A)
	}

	// Round-trip tolerance: Linear -> sRGB -> Linear within ~1 replicated step
	rt := ConvertGray16SRGBToLinear(ConvertGray16LinearToSRGB(gL2))
	if d := absI32(int32(rt.V) - int32(gL2.V)); d > oneReplicated8to16 {
		t.Fatalf("Linear->sRGB->Linear drift too large: orig=0x%04X back=0x%04X (|d|=%d)", gL2.V, rt.V, d)
	}
	if rt.A != gL2.A {
		t.Fatalf("Round-trip alpha changed: got=%d want=%d", rt.A, gL2.A)
	}
}

// ----- Interop with Gray8 paths (sanity) -----

func TestGray16_WithGray8Interop(t *testing.T) {
	// Take a Gray8 linear mid gray (128), expand to 16-bit (0x8080),
	// then to sRGB16 and back, ensuring drift stays small.
	g8 := NewGray8WithAlpha[Linear](128, 200)

	// Expand to 16-bit by replicate (like your Gray16 constructors would do)
	g16 := NewGray16WithAlpha[Linear](basics.Int16u(g8.V)<<8|basics.Int16u(g8.V), basics.Int16u(g8.A)<<8|basics.Int16u(g8.A))

	// Convert to sRGB and back (approximate via 8-bit LUT)
	r := ConvertGray16SRGBToLinear(ConvertGray16LinearToSRGB(g16))

	// Allow one replicated LSB of drift in V & A
	if d := absI32(int32(r.V) - int32(g16.V)); d > oneReplicated8to16 {
		t.Fatalf("Gray8->Gray16 interop V drift too large: orig=0x%04X back=0x%04X (|d|=%d)", g16.V, r.V, d)
	}
	if d := absI32(int32(r.A) - int32(g16.A)); d > oneReplicated8to16 {
		t.Fatalf("Gray8->Gray16 interop A drift too large: orig=0x%04X back=0x%04X (|d|=%d)", g16.A, r.A, d)
	}
}
