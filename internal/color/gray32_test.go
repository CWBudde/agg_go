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
		{0, 0}, {1, 0}, {0, 1}, {1, 1},
		{0.8, 0.5}, {0.123, 0.01}, {0.75, 0.99},
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
	if !feq(Gray32Prelerp(0.4, 0.2, 0.0), 0.4, epsTight) {
		t.Fatalf("Prelerp a=0")
	}
	if !feq(Gray32Prelerp(0.4, 0.2, 1.0), 0.2, epsTight) {
		t.Fatalf("Prelerp a=1")
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

// -------- helpers --------

func feq(a, b, eps float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

const (
	epsTight = 1e-6
	epsLoose = 1e-4
)
