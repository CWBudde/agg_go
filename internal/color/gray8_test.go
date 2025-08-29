package color

import (
	"testing"

	"agg_go/internal/basics"
)

// Test Gray8 basic functionality
func TestGray8Basic(t *testing.T) {
	// Test constructor
	g := NewGray8[Linear](128)
	if g.V != 128 {
		t.Errorf("Expected V=128, got V=%d", g.V)
	}
	if g.A != Gray8BaseMask {
		t.Errorf("Expected A=%d, got A=%d", Gray8BaseMask, g.A)
	}

	// Test constructor with alpha
	g2 := NewGray8WithAlpha[Linear](64, 200)
	if g2.V != 64 || g2.A != 200 {
		t.Errorf("Expected V=64, A=200, got V=%d, A=%d", g2.V, g2.A)
	}
}

func TestGray8Methods(t *testing.T) {
	g := NewGray8WithAlpha[Linear](128, 200)

	// Test Clear
	g.Clear()
	if g.V != 0 || g.A != 0 {
		t.Errorf("Clear() failed: V=%d, A=%d", g.V, g.A)
	}

	// Test Transparent
	g = NewGray8WithAlpha[Linear](128, 200)
	g.Transparent()
	if g.V != 128 || g.A != 0 {
		t.Errorf("Transparent() failed: V=%d, A=%d", g.V, g.A)
	}

	// Test Opacity
	g = NewGray8WithAlpha[Linear](128, 200)
	g.Opacity(0.5)
	expected := basics.Int8u(0.5*float64(Gray8BaseMask) + 0.5)
	if g.A != expected {
		t.Errorf("Opacity(0.5) failed: expected A=%d, got A=%d", expected, g.A)
	}

	// Test IsTransparent
	g.A = 0
	if !g.IsTransparent() {
		t.Error("IsTransparent() should return true for A=0")
	}

	// Test IsOpaque
	g.A = Gray8BaseMask
	if !g.IsOpaque() {
		t.Error("IsOpaque() should return true for A=255")
	}
}

func TestGray8Arithmetic(t *testing.T) {
	// Test Multiply
	result := Gray8Multiply(128, 128)
	expected := basics.Int8u(64) // 128*128/256 ≈ 64
	if result != expected {
		t.Errorf("Gray8Multiply(128, 128) expected %d, got %d", expected, result)
	}

	// Test Lerp
	result = Gray8Lerp(0, 255, 128) // 50% interpolation
	if result < 126 || result > 129 {
		t.Errorf("Gray8Lerp(0, 255, 128) expected ~127, got %d", result)
	}

	// Test Prelerp
	result = Gray8Prelerp(100, 50, 128)
	// p + q - multiply(p, a) = 100 + 50 - multiply(100, 128)
	mulResult := Gray8Multiply(100, 128)
	expected = 100 + 50 - mulResult
	if result != expected {
		t.Errorf("Gray8Prelerp(100, 50, 128) expected %d, got %d", expected, result)
	}
}

func TestGray8Premultiply(t *testing.T) {
	g := NewGray8WithAlpha[Linear](200, 128)
	originalV := g.V

	g.Premultiply()

	// V should be reduced by alpha
	expectedV := Gray8Multiply(originalV, 128)
	if g.V != expectedV {
		t.Errorf("Premultiply() expected V=%d, got V=%d", expectedV, g.V)
	}

	// Alpha should remain unchanged
	if g.A != 128 {
		t.Errorf("Premultiply() should not change alpha: got A=%d", g.A)
	}
}

func TestGray8Demultiply(t *testing.T) {
	g := NewGray8WithAlpha[Linear](100, 128)
	g.Premultiply()
	premultipliedV := g.V

	g.Demultiply()

	// V should be increased back (approximately)
	if g.V <= premultipliedV {
		t.Errorf("Demultiply() should increase V: premult=%d, demult=%d", premultipliedV, g.V)
	}
}

func TestGray8Gradient(t *testing.T) {
	g1 := NewGray8WithAlpha[Linear](0, 0)
	g2 := NewGray8WithAlpha[Linear](255, 255)

	// 50% interpolation
	result := g1.Gradient(g2, 0.5)

	// Should be approximately halfway
	if result.V < 125 || result.V > 130 {
		t.Errorf("Gradient V expected ~127, got %d", result.V)
	}
	if result.A < 125 || result.A > 130 {
		t.Errorf("Gradient A expected ~127, got %d", result.A)
	}
}

func TestGray8Add(t *testing.T) {
	g := NewGray8WithAlpha[Linear](100, 100)
	c := NewGray8WithAlpha[Linear](50, 50)

	g.Add(c, 255) // Full coverage

	// Values should be added (with clamping)
	expectedV := basics.Int8u(150)
	expectedA := basics.Int8u(150)

	if g.V != expectedV {
		t.Errorf("Add() V expected %d, got %d", expectedV, g.V)
	}
	if g.A != expectedA {
		t.Errorf("Add() A expected %d, got %d", expectedA, g.A)
	}
}

func TestGray8Conversion(t *testing.T) {
	// Test conversion from RGBA
	rgba := NewRGBA(0.5, 0.5, 0.5, 0.8)
	gray := ConvertGray8FromRGBA[Linear](rgba)

	// Should be approximately 50% gray with 80% alpha
	// Allow for rounding differences
	if gray.V < 127 || gray.V > 128 {
		t.Errorf("ConvertGray8FromRGBA V expected ~128, got %d", gray.V)
	}
	if gray.A < 203 || gray.A > 205 {
		t.Errorf("ConvertGray8FromRGBA A expected ~204, got %d", gray.A)
	}

	// Test conversion to RGBA
	rgba2 := gray.ConvertToRGBA()

	expectedR := float64(gray.V) / 255.0
	expectedG := expectedR
	expectedB := expectedR
	expectedAlpha := float64(gray.A) / 255.0

	tolerance := 0.01
	if abs(rgba2.R-expectedR) > tolerance {
		t.Errorf("ConvertToRGBA R expected %.3f, got %.3f", expectedR, rgba2.R)
	}
	if abs(rgba2.G-expectedG) > tolerance {
		t.Errorf("ConvertToRGBA G expected %.3f, got %.3f", expectedG, rgba2.G)
	}
	if abs(rgba2.B-expectedB) > tolerance {
		t.Errorf("ConvertToRGBA B expected %.3f, got %.3f", expectedB, rgba2.B)
	}
	if abs(rgba2.A-expectedAlpha) > tolerance {
		t.Errorf("ConvertToRGBA A expected %.3f, got %.3f", expectedAlpha, rgba2.A)
	}
}

func TestGray8Constants(t *testing.T) {
	if Gray8EmptyValue() != 0 {
		t.Errorf("Gray8EmptyValue() expected 0, got %d", Gray8EmptyValue())
	}

	if Gray8FullValue() != Gray8BaseMask {
		t.Errorf("Gray8FullValue() expected %d, got %d", Gray8BaseMask, Gray8FullValue())
	}
}

func TestGray8_ConvertFromRGBA8_SRGBNeedsLinearization(t *testing.T) {
	// Middle gray in sRGB space: 188 encodes ~0.5 linear
	c := RGBA8[SRGB]{R: 188, G: 188, B: 188, A: 255}
	g := ConvertGray8FromRGBA8[SRGB](c)

	// Expect near 128 (0.5 in linear) — allow 1 LSB
	if g.V < 127 || g.V > 129 {
		t.Fatalf("SRGB input must be linearized before luminance: got %d", g.V)
	}
}

func TestGray8_PremultiplyDemultiply_RoundTrip(t *testing.T) {
	cases := []struct{ v, a basics.Int8u }{
		{0, 0}, {255, 0}, {0, 255}, {255, 255}, {123, 1}, {200, 2}, {200, 128},
	}
	for _, c := range cases {
		g := NewGray8WithAlpha[Linear](c.v, c.a)
		orig := g
		g.Premultiply()
		g.Demultiply()
		if c.a == 0 {
			// value should end up 0
			if g.V != 0 {
				t.Fatalf("A=0 should force V=0 after demultiply, got %d", g.V)
			}
			continue
		}
		// For very small alpha values, precision loss is expected due to quantization
		if c.a <= 2 {
			// With alpha=1 or 2, the multiply operation loses significant precision
			// This is expected behavior and matches AGG's implementation
			continue
		}
		diff := int(g.V) - int(orig.V)
		if diff < -1 || diff > 1 {
			t.Fatalf("round-trip drift >1 LSB for V: orig=%d back=%d (A=%d)", orig.V, g.V, c.a)
		}
		if g.A != orig.A {
			t.Fatalf("alpha changed on round-trip: orig=%d back=%d", orig.A, g.A)
		}
	}
}

func TestGray8Multiply_Properties(t *testing.T) {
	for a := basics.Int8u(0); a < 255; a += 17 {
		if Gray8Multiply(a, 0) != 0 {
			t.Fatalf("a*0 != 0 for a=%d", a)
		}
		if Gray8Multiply(a, 255) != a {
			t.Fatalf("a*255 != a for a=%d", a)
		}

		for b := basics.Int8u(0); b < 255; b += 29 {
			if Gray8Multiply(a, b) != Gray8Multiply(b, a) {
				t.Fatalf("commutativity broken: a=%d b=%d", a, b)
			}
			if b < 255 && Gray8Multiply(a, b) > Gray8Multiply(a, b+1) {
				t.Fatalf("monotonicity broken: a=%d b=%d", a, b)
			}
		}
	}
}

func TestGray8Lerp_Endpoints_And_Branches(t *testing.T) {
	if Gray8Lerp(10, 200, 0) != 10 {
		t.Fatal("a=0 should return p")
	}
	if Gray8Lerp(10, 200, 255) != 200 {
		t.Fatal("a=255 should return q")
	}
	// p>q branch
	r := Gray8Lerp(200, 10, 128)
	if r < 104 || r > 106 {
		t.Fatalf("p>q 50%% expected ~105, got %d", r)
	}
}

func TestGray8Prelerp_Extremes(t *testing.T) {
	// Prelerp formula: p + q - multiply(p, a)
	// When a=0: p + q - multiply(p, 0) = p + q - 0 = p + q
	expected := basics.Int8u(150) // 100 + 50 = 150
	if Gray8Prelerp(100, 50, 0) != expected {
		t.Fatalf("a=0: expected %d, got %d", expected, Gray8Prelerp(100, 50, 0))
	}
	if Gray8Prelerp(100, 50, 255) != 50 {
		t.Fatal("a=255")
	}
	mid := Gray8Prelerp(100, 50, 128)
	// Reference calculation: p + q - multiply(p, a)
	want := 100 + 50 - Gray8Multiply(100, 128)
	if mid != want {
		t.Fatalf("mid mismatch: got %d want %d", mid, want)
	}
}

func TestGray8Add_PartialCover_And_Clamp(t *testing.T) {
	// Partial coverage
	g := NewGray8WithAlpha[Linear](100, 100)
	c := NewGray8WithAlpha[Linear](200, 200)
	g.Add(c, 128) // ~50% contribution from c
	if g.V <= 100 || g.A <= 100 {
		t.Fatalf("partial cover should increase components: got V=%d A=%d", g.V, g.A)
	}

	// Clamping
	g = NewGray8WithAlpha[Linear](250, 250)
	c = NewGray8WithAlpha[Linear](250, 250)
	g.Add(c, 255)
	if g.V != 255 || g.A != 255 {
		t.Fatalf("should clamp to 255, got V=%d A=%d", g.V, g.A)
	}
}

func TestGray8Gradient_Endpoints_And_Rounding(t *testing.T) {
	g1 := NewGray8WithAlpha[Linear](10, 20)
	g2 := NewGray8WithAlpha[Linear](250, 240)

	if r := g1.Gradient(g2, 0.0); r != g1 {
		t.Fatalf("k=0 should return first")
	}
	// k=1.0 should return very close to second (might not be exact due to rounding)
	r := g1.Gradient(g2, 1.0)
	diffV := int(r.V) - int(g2.V)
	diffA := int(r.A) - int(g2.A)
	if diffV < 0 {
		diffV = -diffV
	}
	if diffA < 0 {
		diffA = -diffA
	}
	if diffV > 1 || diffA > 1 {
		t.Fatalf("k=1 should return close to second: got V=%d A=%d, expected V=%d A=%d", r.V, r.A, g2.V, g2.A)
	}

	// Rounding near half
	r = g1.Gradient(g2, 0.50196) // ~128/255
	if r.V < 129 || r.V > 132 {
		t.Fatalf("rounding check V around 0.5: got %d", r.V)
	}
}

func TestGray8_Opacity_Clamping(t *testing.T) {
	g := NewGray8WithAlpha[Linear](100, 123)
	g.Opacity(-0.1)
	if g.A != 0 {
		t.Fatal("opacity <0 must clamp to 0")
	}
	g.Opacity(1.1)
	if g.A != Gray8BaseMask {
		t.Fatal("opacity >1 must clamp to 255")
	}
}

func TestGray8_ConvertToRGBA8_BothColorSpaces(t *testing.T) {
	gl := NewGray8WithAlpha[Linear](128, 200)
	rl := gl.ConvertToRGBA8()
	if rl.R != 128 || rl.G != 128 || rl.B != 128 || rl.A != 200 {
		t.Fatalf("Linear -> RGBA8 mismatch: %+v", rl)
	}

	gs := NewGray8WithAlpha[SRGB](128, 200)
	rs := gs.ConvertToRGBA8()
	if rs.R != 128 || rs.G != 128 || rs.B != 128 || rs.A != 200 {
		t.Fatalf("sRGB -> sRGBA8 mismatch: %+v", rs)
	}
}

func TestGray8_Luminance_FloatVsInt_Consistency(t *testing.T) {
	samples := []RGBA{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 0.5, G: 0.5, B: 0.5, A: 1},
		{R: 0.2, G: 0.7, B: 0.1, A: 1},
	}
	for _, s := range samples {
		gF := ConvertGray8FromRGBA[Linear](s)

		// Build a matching 8-bit linear sample and compare
		to8 := func(x float64) basics.Int8u { return basics.Int8u(x*255 + 0.5) }
		s8 := RGBA8[Linear]{R: to8(s.R), G: to8(s.G), B: to8(s.B), A: to8(s.A)}
		gI := ConvertGray8FromRGBA8[Linear](s8)

		diff := int(gF.V) - int(gI.V)
		if diff < -1 || diff > 1 {
			t.Fatalf("float vs int luminance drift >1 LSB: float=%d int=%d for %+v", gF.V, gI.V, s)
		}
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
