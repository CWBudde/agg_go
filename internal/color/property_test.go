package color

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// unitFloat is a float64 in [0, 1] used to test color conversion functions
// that expect inputs in that range.
type unitFloat struct{ V float64 }

func (unitFloat) Generate(r *rand.Rand, _ int) reflect.Value {
	return reflect.ValueOf(unitFloat{r.Float64()})
}

// Property: ConvertToSRGB(ConvertFromSRGB(v)) ≈ v for all v in [0, 1].
// i.e. sRGB→linear→sRGB is a round-trip.
func TestPropertySRGBScalarRoundTrip(t *testing.T) {
	err := quick.Check(func(f unitFloat) bool {
		v := f.V
		got := ConvertToSRGB(ConvertFromSRGB(v))
		return math.Abs(got-v) < 1e-10
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: ConvertFromSRGB(ConvertToSRGB(v)) ≈ v for all v in [0, 1].
// i.e. linear→sRGB→linear is a round-trip.
func TestPropertyLinearScalarRoundTrip(t *testing.T) {
	err := quick.Check(func(f unitFloat) bool {
		v := f.V
		got := ConvertFromSRGB(ConvertToSRGB(v))
		return math.Abs(got-v) < 1e-10
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: ConvertFromSRGB is monotonically non-decreasing on [0, 1].
func TestPropertyConvertFromSRGBMonotonicity(t *testing.T) {
	err := quick.Check(func(a, b unitFloat) bool {
		la := ConvertFromSRGB(a.V)
		lb := ConvertFromSRGB(b.V)
		if a.V <= b.V {
			return la <= lb+1e-15
		}
		return la >= lb-1e-15
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: ConvertToSRGB is monotonically non-decreasing on [0, 1].
func TestPropertyConvertToSRGBMonotonicity(t *testing.T) {
	err := quick.Check(func(a, b unitFloat) bool {
		sa := ConvertToSRGB(a.V)
		sb := ConvertToSRGB(b.V)
		if a.V <= b.V {
			return sa <= sb+1e-15
		}
		return sa >= sb-1e-15
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: RGBA8 Gradient(c2, 0) == c1 (k=0 selects the receiver).
func TestPropertyGradientAtZeroIsReceiver(t *testing.T) {
	err := quick.Check(func(r1, g1, b1, a1, r2, g2, b2, a2 uint8) bool {
		c1 := NewRGBA8[Linear](basics.Int8u(r1), basics.Int8u(g1), basics.Int8u(b1), basics.Int8u(a1))
		c2 := NewRGBA8[Linear](basics.Int8u(r2), basics.Int8u(g2), basics.Int8u(b2), basics.Int8u(a2))
		got := c1.Gradient(c2, 0)
		return got.R == c1.R && got.G == c1.G && got.B == c1.B && got.A == c1.A
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: RGBA8 Gradient(c2, 255) == c2 (k=255 selects the argument).
func TestPropertyGradientAtMaxIsArgument(t *testing.T) {
	err := quick.Check(func(r1, g1, b1, a1, r2, g2, b2, a2 uint8) bool {
		c1 := NewRGBA8[Linear](basics.Int8u(r1), basics.Int8u(g1), basics.Int8u(b1), basics.Int8u(a1))
		c2 := NewRGBA8[Linear](basics.Int8u(r2), basics.Int8u(g2), basics.Int8u(b2), basics.Int8u(a2))
		got := c1.Gradient(c2, 255)
		return got.R == c2.R && got.G == c2.G && got.B == c2.B && got.A == c2.A
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: RGBA8 sRGB→linear conversion is monotonically non-decreasing per channel.
// Higher sRGB input must produce higher or equal linear output. The LUT is built from
// the same ConvertFromSRGB formula used by the scalar tests.
func TestPropertyRGBA8SRGBToLinearMonotonicity(t *testing.T) {
	err := quick.Check(func(r1, r2 uint8) bool {
		c1 := NewRGBA8[SRGB](basics.Int8u(r1), 0, 0, 255)
		c2 := NewRGBA8[SRGB](basics.Int8u(r2), 0, 0, 255)
		lin1 := ConvertToLinear(c1)
		lin2 := ConvertToLinear(c2)
		if r1 <= r2 {
			return lin1.R <= lin2.R
		}
		return lin1.R >= lin2.R
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: RGBA8 linear→sRGB conversion is monotonically non-decreasing per channel.
func TestPropertyRGBA8LinearToSRGBMonotonicity(t *testing.T) {
	err := quick.Check(func(r1, r2 uint8) bool {
		c1 := NewRGBA8[Linear](basics.Int8u(r1), 0, 0, 255)
		c2 := NewRGBA8[Linear](basics.Int8u(r2), 0, 0, 255)
		s1 := ConvertToSRGBFromLinear(c1)
		s2 := ConvertToSRGBFromLinear(c2)
		if r1 <= r2 {
			return s1.R <= s2.R
		}
		return s1.R >= s2.R
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: RGBA8 alpha channel is always preserved unchanged during colorspace conversion
// (AGG does not apply gamma to the alpha channel).
func TestPropertyAlphaPreservedInColorspaceConversion(t *testing.T) {
	err := quick.Check(func(r, g, b, a uint8) bool {
		srgb := NewRGBA8[SRGB](basics.Int8u(r), basics.Int8u(g), basics.Int8u(b), basics.Int8u(a))
		lin := ConvertToLinear(srgb)
		back := ConvertToSRGBFromLinear(lin)
		return lin.A == srgb.A && back.A == srgb.A
	}, nil)
	if err != nil {
		t.Error(err)
	}
}
