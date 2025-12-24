package color

import (
	"testing"
)

// -------- constructors, flags, basic methods --------

func TestGray32_ConstructorsAndFlags(t *testing.T) {
	g := NewGray32[Linear](0.5)
	if !feq(g.V, 0.5, epsTight) {
		t.Fatalf("NewGray32 V=%f, want 0.5", g.V)
	}
	if !feq(g.A, 1.0, epsTight) {
		t.Fatalf("NewGray32 A=%f, want 1.0", g.A)
	}
	if g.IsTransparent() {
		t.Fatalf("IsTransparent should be false (A=1)")
	}
	if !g.IsOpaque() {
		t.Fatalf("IsOpaque should be true (A=1)")
	}

	g2 := NewGray32WithAlpha[Linear](0.25, 0.75)
	if !feq(g2.V, 0.25, epsTight) || !feq(g2.A, 0.75, epsTight) {
		t.Fatalf("NewGray32WithAlpha got (V=%f,A=%f), want (0.25,0.75)", g2.V, g2.A)
	}
	if g2.IsTransparent() || g2.IsOpaque() {
		t.Fatalf("Flags wrong for A=0.75")
	}

	g2.Transparent()
	if !feq(g2.A, 0.0, epsTight) || !feq(g2.V, 0.25, epsTight) {
		t.Fatalf("Transparent() got (V=%f,A=%f), want (0.25,0.0)", g2.V, g2.A)
	}
	if !g2.IsTransparent() {
		t.Fatalf("IsTransparent should be true after Transparent()")
	}

	g2.Clear()
	if !feq(g2.V, 0.0, epsTight) || !feq(g2.A, 0.0, epsTight) {
		t.Fatalf("Clear() got (V=%f,A=%f), want (0,0)", g2.V, g2.A)
	}
}

func TestGray32_ConvertToRGBA(t *testing.T) {
	g := NewGray32WithAlpha[Linear](0.3, 0.8)
	r := g.ConvertToRGBA()
	if !feq(float32(r.R), 0.3, epsTight) ||
		!feq(float32(r.G), 0.3, epsTight) ||
		!feq(float32(r.B), 0.3, epsTight) ||
		!feq(float32(r.A), 0.8, epsTight) {
		t.Fatalf("ConvertToRGBA mismatch: %+v", r)
	}
}

// -------- opacity clamping (if you added Opacity/GetOpacity) --------

func TestGray32_OpacityClamp(t *testing.T) {
	g := NewGray32WithAlpha[Linear](0.1, 0.2)
	g.Opacity(-0.5)
	if !feq(g.A, 0.0, epsTight) {
		t.Fatalf("Opacity(<0) should clamp to 0, got %f", g.A)
	}
	g.Opacity(1.5)
	if !feq(g.A, 1.0, epsTight) {
		t.Fatalf("Opacity(>1) should clamp to 1, got %f", g.A)
	}
	g.Opacity(0.6)
	if !feq(g.GetOpacity(), 0.6, epsTight) {
		t.Fatalf("GetOpacity mismatch, got %f want 0.6", g.GetOpacity())
	}
}

// -------- premultiply/demultiply round-trip --------

func TestGray32_PremultiplyDemultiply_RoundTrip(t *testing.T) {
	cases := []struct {
		v, a float32
	}{
		{0, 0},
		{1, 0},
		{0, 1},
		{1, 1},
		{0.8, 0.5},
		{0.123, 0.01},
		{0.75, 0.99},
	}
	for _, c := range cases {
		g := NewGray32WithAlpha[Linear](c.v, c.a)
		orig := g
		g.Premultiply()
		g.Demultiply()
		if c.a == 0 {
			if !feq(g.V, 0, epsTight) {
				t.Fatalf("A=0 should force V=0 after demultiply, got %f", g.V)
			}
			continue
		}
		if !feq(g.V, orig.V, epsLoose) {
			t.Fatalf("round-trip V drift: orig=%f back=%f (A=%f)", orig.V, g.V, c.a)
		}
		if !feq(g.A, orig.A, epsTight) {
			t.Fatalf("alpha changed: orig=%f back=%f", orig.A, g.A)
		}
	}
}

// -------- interpolation --------

func TestGray32_LerpAndPrelerp(t *testing.T) {
	// Endpoints
	if !feq(Gray32Lerp(0.1, 0.9, 0.0), 0.1, epsTight) {
		t.Fatalf("Lerp a=0 should return p")
	}
	if !feq(Gray32Lerp(0.1, 0.9, 1.0), 0.9, epsTight) {
		t.Fatalf("Lerp a=1 should return q")
	}
	// Midpoint
	m := Gray32Lerp(0.0, 1.0, 0.5)
	if !feq(m, 0.5, epsTight) {
		t.Fatalf("Lerp midpoint got %f want 0.5", m)
	}

	// Prelerp extremes
	// Formula: (1-a)*p + q  which equals  p + q - p*a
	// When a=0: p + q
	// When a=1: q
	if !feq(Gray32Prelerp(0.4, 0.2, 0.0), 0.6, epsTight) {
		t.Fatalf("Prelerp a=0 should return p+q = 0.6")
	}
	if !feq(Gray32Prelerp(0.4, 0.2, 1.0), 0.2, epsTight) {
		t.Fatalf("Prelerp a=1 should return q")
	}
	// Prelerp mid
	pm := Gray32Prelerp(0.4, 0.2, 0.5)
	want := (1-float32(0.5))*0.4 + 0.2
	if !feq(pm, want, epsTight) {
		t.Fatalf("Prelerp mid got %f want %f", pm, want)
	}
}

func TestGray32_Gradient(t *testing.T) {
	g1 := NewGray32WithAlpha[Linear](0.1, 0.2)
	g2 := NewGray32WithAlpha[Linear](0.9, 0.8)

	if r := g1.Gradient(g2, 0.0); !feq(r.V, g1.V, epsTight) || !feq(r.A, g1.A, epsTight) {
		t.Fatalf("Gradient k=0 returns first failed: %+v", r)
	}
	if r := g1.Gradient(g2, 1.0); !feq(r.V, g2.V, epsTight) || !feq(r.A, g2.A, epsTight) {
		t.Fatalf("Gradient k=1 returns second failed: %+v", r)
	}
	r := g1.Gradient(g2, 0.5)
	if !feq(r.V, 0.5, epsLoose) || !feq(r.A, 0.5, epsLoose) {
		t.Fatalf("Gradient k=0.5 got (V=%f,A=%f) want (~0.5,~0.5)", r.V, r.A)
	}
}

// -------- add/blend with coverage --------

func TestGray32_Add_PartialCover_And_Clamp(t *testing.T) {
	// partial cover
	g := NewGray32WithAlpha[Linear](0.4, 0.4)
	c := NewGray32WithAlpha[Linear](0.8, 0.8)
	g.Add(c, 128) // ~50% cover
	if !(g.V > 0.4 && g.A > 0.4) {
		t.Fatalf("Add with partial cover should increase components: got (V=%f,A=%f)", g.V, g.A)
	}

	// full cover + opaque → replace
	g = NewGray32WithAlpha[Linear](0.2, 0.2)
	c = NewGray32WithAlpha[Linear](0.7, 1.0)
	g.Add(c, 255)
	if !feq(g.V, 0.7, epsTight) || !feq(g.A, 1.0, epsTight) {
		t.Fatalf("Add full cover with opaque should replace: got (V=%f,A=%f)", g.V, g.A)
	}

	// clamping
	g = NewGray32WithAlpha[Linear](0.9, 0.9)
	c = NewGray32WithAlpha[Linear](0.9, 0.9)
	g.Add(c, 255)
	if !feq(g.V, 1.0, epsTight) || !feq(g.A, 1.0, epsTight) {
		t.Fatalf("Add should clamp to 1: got (V=%f,A=%f)", g.V, g.A)
	}
}

// -------- luminance / conversion paths --------

func TestGray32_FromRGBA_Linear_BT709(t *testing.T) {
	// RGBA is linear in this codebase
	type S struct{ R, G, B, A float64 }
	samples := []S{
		{0, 0, 0, 1},
		{1, 1, 1, 1},
		{0.5, 0.5, 0.5, 0.75},
		{0.2, 0.7, 0.1, 0.25},
		{0.9, 0.1, 0.3, 0.6},
	}
	for _, s := range samples {
		lum := float32(0.2126*s.R + 0.7152*s.G + 0.0722*s.B)
		expA := float32(s.A)

		got := ConvertGray32FromRGBA[Linear](RGBA(s))
		if !feq(got.V, lum, epsLoose) {
			t.Fatalf("BT.709 V mismatch: exp=%f got=%f for sample=%+v", lum, got.V, s)
		}
		if !feq(got.A, expA, epsTight) {
			t.Fatalf("Alpha mismatch: exp=%f got=%f for sample=%+v", expA, got.A, s)
		}
	}
}

func TestGray32_ColorspaceConversions(t *testing.T) {
	// Start with linear ~0.5
	gl := NewGray32WithAlpha[Linear](0.5, 0.8)
	gs := ConvertGray32LinearToSRGB(gl)
	if !(gs.V > 0.70 && gs.V < 0.76) || !feq(gs.A, gl.A, epsTight) {
		t.Fatalf("Linear->sRGB unexpected: %+v", gs)
	}
	gl2 := ConvertGray32SRGBToLinear(gs)
	if !feq(gl2.V, gl.V, epsLoose) || !feq(gl2.A, gl.A, epsTight) {
		t.Fatalf("sRGB->Linear roundtrip drift: got %+v want %+v", gl2, gl)
	}
}

// -------- Additional comprehensive tests (like Gray8) --------

func TestGray32_BasicFunctionality(t *testing.T) {
	// Test basic constructor
	g := NewGray32[Linear](0.5)
	if !feq(g.V, 0.5, epsTight) {
		t.Errorf("Expected V=0.5, got V=%f", g.V)
	}
	if !feq(g.A, 1.0, epsTight) {
		t.Errorf("Expected A=1.0, got A=%f", g.A)
	}

	// Test constructor with alpha
	g2 := NewGray32WithAlpha[Linear](0.25, 0.75)
	if !feq(g2.V, 0.25, epsTight) || !feq(g2.A, 0.75, epsTight) {
		t.Errorf("Expected V=0.25, A=0.75, got V=%f, A=%f", g2.V, g2.A)
	}
}

func TestGray32_ArithmeticOperations(t *testing.T) {
	// Test Multiply
	result := Gray32Multiply(0.5, 0.5)
	expected := float32(0.25)
	if !feq(result, expected, epsTight) {
		t.Errorf("Gray32Multiply(0.5, 0.5) expected %f, got %f", expected, result)
	}

	// Test Lerp
	result = Gray32Lerp(0.0, 1.0, 0.5) // 50% interpolation
	if !feq(result, 0.5, epsTight) {
		t.Errorf("Gray32Lerp(0.0, 1.0, 0.5) expected 0.5, got %f", result)
	}

	// Test Prelerp
	result = Gray32Prelerp(0.4, 0.2, 0.5)
	// (1-a)*p + q = (1-0.5)*0.4 + 0.2 = 0.2 + 0.2 = 0.4
	expected = 0.4
	if !feq(result, expected, epsTight) {
		t.Errorf("Gray32Prelerp(0.4, 0.2, 0.5) expected %f, got %f", expected, result)
	}

	// Test Demultiply
	result = Gray32Demultiply(0.25, 0.5)
	expected = 0.5 // 0.25 / 0.5
	if !feq(result, expected, epsTight) {
		t.Errorf("Gray32Demultiply(0.25, 0.5) expected %f, got %f", expected, result)
	}

	// Test Demultiply edge case (divide by zero)
	result = Gray32Demultiply(0.25, 0.0)
	if !feq(result, 0.0, epsTight) {
		t.Errorf("Gray32Demultiply(0.25, 0.0) expected 0.0, got %f", result)
	}
}

func TestGray32_PremultiplyDemultiply(t *testing.T) {
	g := NewGray32WithAlpha[Linear](0.8, 0.5)
	originalV := g.V

	g.Premultiply()

	// V should be reduced by alpha
	expectedV := originalV * 0.5
	if !feq(g.V, expectedV, epsTight) {
		t.Errorf("Premultiply() expected V=%f, got V=%f", expectedV, g.V)
	}

	// Alpha should remain unchanged
	if !feq(g.A, 0.5, epsTight) {
		t.Errorf("Premultiply() should not change alpha: got A=%f", g.A)
	}

	// Test demultiply
	g.Demultiply()

	// V should be restored (approximately)
	if !feq(g.V, originalV, epsLoose) {
		t.Errorf("Demultiply() should restore V: original=%f, got=%f", originalV, g.V)
	}
}

func TestGray32_ConversionTests(t *testing.T) {
	// Test conversion from RGBA
	rgba := NewRGBA(0.5, 0.5, 0.5, 0.8)
	gray := ConvertGray32FromRGBA[Linear](rgba)

	// Should be approximately 50% gray with 80% alpha
	if !feq(gray.V, 0.5, epsLoose) {
		t.Errorf("ConvertGray32FromRGBA V expected ~0.5, got %f", gray.V)
	}
	if !feq(gray.A, 0.8, epsTight) {
		t.Errorf("ConvertGray32FromRGBA A expected 0.8, got %f", gray.A)
	}

	// Test conversion to RGBA
	rgba2 := gray.ConvertToRGBA()

	tolerance := float64(0.001)
	if absFloat64Gray32(rgba2.R-float64(gray.V)) > tolerance {
		t.Errorf("ConvertToRGBA R expected %f, got %f", gray.V, rgba2.R)
	}
	if absFloat64Gray32(rgba2.G-float64(gray.V)) > tolerance {
		t.Errorf("ConvertToRGBA G expected %f, got %f", gray.V, rgba2.G)
	}
	if absFloat64Gray32(rgba2.B-float64(gray.V)) > tolerance {
		t.Errorf("ConvertToRGBA B expected %f, got %f", gray.V, rgba2.B)
	}
	if absFloat64Gray32(rgba2.A-float64(gray.A)) > tolerance {
		t.Errorf("ConvertToRGBA A expected %f, got %f", gray.A, rgba2.A)
	}
}

func TestGray32_Constants(t *testing.T) {
	if !feq(Gray32EmptyValue(), 0.0, epsTight) {
		t.Errorf("Gray32EmptyValue() expected 0.0, got %f", Gray32EmptyValue())
	}

	if !feq(Gray32FullValue(), 1.0, epsTight) {
		t.Errorf("Gray32FullValue() expected 1.0, got %f", Gray32FullValue())
	}
}

func TestGray32_EdgeCases(t *testing.T) {
	// Test with extreme values
	g := NewGray32WithAlpha[Linear](1.5, -0.5) // Values outside [0,1]
	g.Opacity(g.A)                             // This should clamp

	if g.A < 0.0 || g.A > 1.0 {
		t.Errorf("Alpha should be clamped to [0,1], got %f", g.A)
	}

	// Test clamping in Add operation
	g = NewGray32WithAlpha[Linear](0.9, 0.9)
	c := NewGray32WithAlpha[Linear](0.9, 0.9)
	g.Add(c, 255)
	if !feq(g.V, 1.0, epsTight) || !feq(g.A, 1.0, epsTight) {
		t.Errorf("Add should clamp to 1.0, got V=%f A=%f", g.V, g.A)
	}
}

func TestGray32_MultiplyProperties(t *testing.T) {
	testValues := []float32{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, a := range testValues {
		// Test multiply by 0
		if !feq(Gray32Multiply(a, 0.0), 0.0, epsTight) {
			t.Fatalf("a*0 != 0 for a=%f", a)
		}
		// Test multiply by 1
		if !feq(Gray32Multiply(a, 1.0), a, epsTight) {
			t.Fatalf("a*1 != a for a=%f", a)
		}

		for _, b := range testValues {
			// Test commutativity
			if !feq(Gray32Multiply(a, b), Gray32Multiply(b, a), epsTight) {
				t.Fatalf("commutativity broken: a=%f b=%f", a, b)
			}
		}
	}
}

func TestGray32_LerpProperties(t *testing.T) {
	// Test endpoints
	if !feq(Gray32Lerp(0.1, 0.9, 0.0), 0.1, epsTight) {
		t.Fatal("a=0 should return p")
	}
	if !feq(Gray32Lerp(0.1, 0.9, 1.0), 0.9, epsTight) {
		t.Fatal("a=1 should return q")
	}

	// Test midpoint
	mid := Gray32Lerp(0.0, 1.0, 0.5)
	if !feq(mid, 0.5, epsTight) {
		t.Fatalf("50%% interpolation expected 0.5, got %f", mid)
	}

	// Test reverse interpolation (p > q)
	rev := Gray32Lerp(0.8, 0.2, 0.5)
	if !feq(rev, 0.5, epsTight) {
		t.Fatalf("Reverse interpolation expected 0.5, got %f", rev)
	}
}

func TestGray32_PrelerpProperties(t *testing.T) {
	// Test extremes
	// Prelerp formula: (1-a)*p + q = p + q - p*a
	// When a=0: p + q (not p!)
	// When a=1: q
	if !feq(Gray32Prelerp(0.4, 0.2, 0.0), 0.6, epsTight) {
		t.Fatal("a=0 should return p+q = 0.6")
	}
	if !feq(Gray32Prelerp(0.4, 0.2, 1.0), 0.2, epsTight) {
		t.Fatal("a=1 should return q")
	}

	// Test midpoint with reference calculation
	mid := Gray32Prelerp(0.4, 0.2, 0.5)
	want := (1-float32(0.5))*0.4 + 0.2 // = 0.2 + 0.2 = 0.4
	if !feq(mid, want, epsTight) {
		t.Fatalf("Prelerp mid mismatch: got %f want %f", mid, want)
	}
}

func TestGray32_GradientEdgeCases(t *testing.T) {
	g1 := NewGray32WithAlpha[Linear](0.1, 0.2)
	g2 := NewGray32WithAlpha[Linear](0.9, 0.8)

	// Test k=0 returns first color
	result := g1.Gradient(g2, 0.0)
	if !feq(result.V, g1.V, epsTight) || !feq(result.A, g1.A, epsTight) {
		t.Fatalf("k=0 should return first color")
	}

	// Test k=1 returns second color
	result = g1.Gradient(g2, 1.0)
	if !feq(result.V, g2.V, epsTight) || !feq(result.A, g2.A, epsTight) {
		t.Fatalf("k=1 should return second color")
	}

	// Test k values outside [0,1] get clamped
	result = g1.Gradient(g2, -0.5) // Should be clamped to 0
	if !feq(result.V, g1.V, epsTight) || !feq(result.A, g1.A, epsTight) {
		t.Fatalf("k<0 should be clamped to 0")
	}

	result = g1.Gradient(g2, 1.5) // Should be clamped to 1
	if !feq(result.V, g2.V, epsTight) || !feq(result.A, g2.A, epsTight) {
		t.Fatalf("k>1 should be clamped to 1")
	}
}

func TestGray32_AddWithPartialCoverage(t *testing.T) {
	// Test partial coverage behavior
	g := NewGray32WithAlpha[Linear](0.4, 0.4)
	c := NewGray32WithAlpha[Linear](0.8, 0.8)
	g.Add(c, 128) // ~50% coverage (128/255 = 0.5020)

	// Add formula: g.V += c.V * cover
	// Expected: 0.4 + 0.8 * 0.5020 = 0.4 + 0.4016 ≈ 0.8016
	// Components should increase from original 0.4
	// Result can be >= 0.8 because we're adding, not blending with alpha
	if g.V <= 0.4 || g.A <= 0.4 {
		t.Fatalf("Partial coverage should increase components: got V=%f A=%f", g.V, g.A)
	}

	// Test zero coverage
	g = NewGray32WithAlpha[Linear](0.5, 0.5)
	original := g
	g.Add(c, 0)
	if !feq(g.V, original.V, epsTight) || !feq(g.A, original.A, epsTight) {
		t.Fatalf("Zero coverage should not change color: got V=%f A=%f, want V=%f A=%f",
			g.V, g.A, original.V, original.A)
	}
}

func TestGray32_ColorSpaceConversions(t *testing.T) {
	// Test sRGB to Linear conversion
	gs := NewGray32WithAlpha[SRGB](0.73, 0.8) // ~0.5 linear
	gl := ConvertGray32SRGBToLinear(gs)

	// Linear value should be significantly different from sRGB
	if absFloat32(gl.V-gs.V) < 0.1 {
		t.Fatalf("sRGB->Linear conversion should change value significantly: sRGB=%f Linear=%f", gs.V, gl.V)
	}
	if !feq(gl.A, gs.A, epsTight) {
		t.Fatalf("Alpha should remain unchanged in colorspace conversion")
	}

	// Test round-trip conversion
	gs2 := ConvertGray32LinearToSRGB(gl)
	if !feq(gs2.V, gs.V, epsLoose) {
		t.Fatalf("Round-trip conversion drift: original=%f back=%f", gs.V, gs2.V)
	}
}

func TestGray32_LuminanceConsistency(t *testing.T) {
	// Test that luminance calculation is consistent
	samples := []RGBA{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 0.5, G: 0.5, B: 0.5, A: 1},
		{R: 0.2, G: 0.7, B: 0.1, A: 1},
		{R: 1.0, G: 0.0, B: 0.0, A: 0.5}, // Pure red
		{R: 0.0, G: 1.0, B: 0.0, A: 0.5}, // Pure green
		{R: 0.0, G: 0.0, B: 1.0, A: 0.5}, // Pure blue
	}

	for _, sample := range samples {
		gray := ConvertGray32FromRGBA[Linear](sample)

		// Calculate expected luminance using BT.709 coefficients
		expectedLum := float32(0.2126*sample.R + 0.7152*sample.G + 0.0722*sample.B)

		if !feq(gray.V, expectedLum, epsLoose) {
			t.Errorf("Luminance mismatch for %+v: expected %f, got %f", sample, expectedLum, gray.V)
		}
		if !feq(gray.A, float32(sample.A), epsTight) {
			t.Errorf("Alpha mismatch for %+v: expected %f, got %f", sample, sample.A, gray.A)
		}
	}
}

func TestGray32_PremultiplyEdgeCases(t *testing.T) {
	// Test with zero alpha
	g := NewGray32WithAlpha[Linear](0.8, 0.0)
	g.Premultiply()
	if !feq(g.V, 0.0, epsTight) {
		t.Errorf("Premultiply with A=0 should set V=0, got V=%f", g.V)
	}

	// Test with alpha = 1 (should not change V)
	g = NewGray32WithAlpha[Linear](0.8, 1.0)
	originalV := g.V
	g.Premultiply()
	if !feq(g.V, originalV, epsTight) {
		t.Errorf("Premultiply with A=1 should not change V: original=%f, got=%f", originalV, g.V)
	}

	// Test with negative alpha
	g = NewGray32WithAlpha[Linear](0.8, -0.1)
	g.Premultiply()
	if !feq(g.V, 0.0, epsTight) {
		t.Errorf("Premultiply with negative A should set V=0, got V=%f", g.V)
	}
}

func TestGray32_DemultiplyEdgeCases(t *testing.T) {
	// Test with zero alpha
	g := NewGray32WithAlpha[Linear](0.5, 0.0)
	g.Demultiply()
	if !feq(g.V, 0.0, epsTight) {
		t.Errorf("Demultiply with A=0 should set V=0, got V=%f", g.V)
	}

	// Test with alpha = 1 (should not change V)
	g = NewGray32WithAlpha[Linear](0.8, 1.0)
	originalV := g.V
	g.Demultiply()
	if !feq(g.V, originalV, epsTight) {
		t.Errorf("Demultiply with A=1 should not change V: original=%f, got=%f", originalV, g.V)
	}

	// Test with negative alpha
	g = NewGray32WithAlpha[Linear](0.8, -0.1)
	g.Demultiply()
	if !feq(g.V, 0.0, epsTight) {
		t.Errorf("Demultiply with negative A should set V=0, got V=%f", g.V)
	}
}

func TestGray32_AddWithCoverMethod(t *testing.T) {
	g := NewGray32WithAlpha[Linear](0.3, 0.3)
	c := NewGray32WithAlpha[Linear](0.6, 0.6)

	// Test AddWithCover (should behave same as Add)
	g1 := g
	g2 := g

	g1.Add(c, 128)
	g2.AddWithCover(c, 128)

	if !feq(g1.V, g2.V, epsTight) || !feq(g1.A, g2.A, epsTight) {
		t.Errorf("AddWithCover should behave same as Add: Add=%+v AddWithCover=%+v", g1, g2)
	}
}

// -------- helpers --------

func feq(a, b, eps float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

func absFloat32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func absFloat64Gray32(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

const (
	epsTight = 1e-6
	epsLoose = 1e-4
)
