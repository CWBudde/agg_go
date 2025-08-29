package color

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// --- helpers ---------------------------------------------------------------

func expLumLinear8(r, g, b uint8) basics.Int8u {
	return basics.Int8u((uint32(r)*bt709R + uint32(g)*bt709G + uint32(b)*bt709B) >> 8)
}

func almostEqualFloat(a, b, tol float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return math.Abs(a-b) <= tol
}

func almostEqualU8(a, b basics.Int8u, tol uint8) bool {
	da := int(a)
	db := int(b)
	return absInt(da-db) <= int(tol)
}

// --- tests -----------------------------------------------------------------

func TestLuminanceFromRGBA8Linear_KnownValues(t *testing.T) {
	cases := []struct {
		name    string
		r, g, b uint8
	}{
		{"black", 0, 0, 0},
		{"white", 255, 255, 255},
		{"red", 255, 0, 0},
		{"green", 0, 255, 0},
		{"blue", 0, 0, 255},
		{"random-ish", 12, 34, 56},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LuminanceFromRGBA8Linear(RGBA8[Linear]{
				R: basics.Int8u(tc.r),
				G: basics.Int8u(tc.g),
				B: basics.Int8u(tc.b),
				A: 255,
			})
			want := expLumLinear8(tc.r, tc.g, tc.b)
			if got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		})
	}
}

func TestLuminanceFromRGBA8SRGB_UsesLUTThen709(t *testing.T) {
	// We validate the integer pipeline by running the same math against
	// the LUT-linearized channels (no magic numbers beyond the shared weights).
	cases := []struct {
		name    string
		r, g, b uint8
	}{
		{"black", 0, 0, 0},
		{"white", 255, 255, 255},
		{"mid-gray", 128, 128, 128},
		{"quarter-gray", 64, 64, 64},
		{"three-quarter-gray", 192, 192, 192},
		{"sr", 255, 32, 32},
		{"sg", 32, 255, 32},
		{"sb", 32, 32, 255},
		{"rgb-mix", 12, 200, 150},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LuminanceFromRGBA8SRGB(RGBA8[SRGB]{
				R: basics.Int8u(tc.r),
				G: basics.Int8u(tc.g),
				B: basics.Int8u(tc.b),
				A: 255,
			})

			// Reference expectation: linearize via the same LUT, then apply integer BT.709.
			rl := srgb8ToLinear8(basics.Int8u(tc.r))
			gl := srgb8ToLinear8(basics.Int8u(tc.g))
			bl := srgb8ToLinear8(basics.Int8u(tc.b))
			want := expLumLinear8(uint8(rl), uint8(gl), uint8(bl))

			if got != want {
				t.Fatalf("got %d, want %d (rl=%d gl=%d bl=%d)", got, want, rl, gl, bl)
			}
		})
	}
}

func TestLuminanceFromRGBA_FloatKnownValues(t *testing.T) {
	cases := []struct {
		name    string
		r, g, b float64
		want    float64
	}{
		{"black", 0, 0, 0, 0},
		{"white", 1, 1, 1, 1},
		{"red", 1, 0, 0, 0.2126},
		{"green", 0, 1, 0, 0.7152},
		{"blue", 0, 0, 1, 0.0722},
		{"mid-gray", 0.5, 0.5, 0.5, 0.5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LuminanceFromRGBA(RGBA{
				R: tc.r, G: tc.g, B: tc.b, A: 1,
			})
			if !almostEqualFloat(got, tc.want, 1e-9) {
				t.Fatalf("got %.10f, want %.10f", got, tc.want)
			}
		})
	}
}

func TestConsistency_Linear8_vs_Float(t *testing.T) {
	// For a few representative colors, the 8-bit integer path and the
	// float path (with normalized linear inputs) should agree within ~1 U8.
	cases := []struct {
		name    string
		r, g, b uint8
	}{
		{"black", 0, 0, 0},
		{"white", 255, 255, 255},
		{"red", 255, 0, 0},
		{"green", 0, 255, 0},
		{"blue", 0, 0, 255},
		{"mid-gray", 128, 128, 128},
		{"random-ish", 23, 45, 210},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l8 := LuminanceFromRGBA8Linear(RGBA8[Linear]{
				R: basics.Int8u(tc.r),
				G: basics.Int8u(tc.g),
				B: basics.Int8u(tc.b),
				A: 255,
			})

			lf := LuminanceFromRGBA(RGBA{
				R: float64(tc.r) / 255.0,
				G: float64(tc.g) / 255.0,
				B: float64(tc.b) / 255.0,
				A: 1.0,
			})
			// Scale float 0..1 back to 0..255 and round.
			lfU8 := basics.Int8u(math.Round(lf * 255.0))

			// Allow small mismatch due to integer weights (sum 257)>>8 vs exact 0.2126/0.7152/0.0722
			if !almostEqualU8(l8, lfU8, 1) {
				t.Fatalf("linear8=%d, float≈%d differ by >1", l8, lfU8)
			}
		})
	}
}
