package blender

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

func TestExtendedBlendModePhotoshopGoldens(t *testing.T) {
	tests := []struct {
		name string
		op   CompOp
		want [4]basics.Int8u
	}{
		{"linear-burn", CompOpLinearBurn, [4]basics.Int8u{1, 0, 0, 255}},
		{"darker-color", CompOpDarkerColor, [4]basics.Int8u{64, 128, 192, 255}},
		{"lighter-color", CompOpLighterColor, [4]basics.Int8u{192, 96, 32, 255}},
		{"vivid-light", CompOpVividLight, [4]basics.Int8u{130, 86, 4, 255}},
		{"linear-light", CompOpLinearLight, [4]basics.Int8u{193, 65, 1, 255}},
		{"pin-light", CompOpPinLight, [4]basics.Int8u{129, 128, 64, 255}},
		{"hard-mix", CompOpHardMix, [4]basics.Int8u{255, 0, 0, 255}},
		{"subtract", CompOpSubtract, [4]basics.Int8u{0, 32, 160, 255}},
		{"divide", CompOpDivide, [4]basics.Int8u{85, 255, 255, 255}},
		{"hue", CompOpHue, [4]basics.Int8u{175, 98, 47, 255}},
		{"saturation", CompOpSaturation, [4]basics.Int8u{51, 131, 211, 255}},
		{"color", CompOpColor, [4]basics.Int8u{190, 94, 30, 255}},
		{"luminosity", CompOpLuminosity, [4]basics.Int8u{66, 130, 194, 255}},
		{"photoshop-color-burn", CompOpColorBurnPhotoshop, [4]basics.Int8u{1, 0, 0, 255}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := []basics.Int8u{64, 128, 192, 255}
			NewCompositeBlenderPlain[color.Linear, order.RGBA](tt.op).
				BlendPix(dst, 192, 96, 32, 255, 255)
			if got := [4]basics.Int8u(dst); got != tt.want {
				t.Fatalf("pixel = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacyCompositeOperatorValuesAreStable(t *testing.T) {
	legacy := []CompOp{
		CompOpClear, CompOpSrc, CompOpDst, CompOpSrcOver, CompOpDstOver,
		CompOpSrcIn, CompOpDstIn, CompOpSrcOut, CompOpDstOut, CompOpSrcAtop,
		CompOpDstAtop, CompOpXor, CompOpPlus, CompOpMultiply, CompOpScreen,
		CompOpOverlay, CompOpDarken, CompOpLighten, CompOpColorDodge,
		CompOpColorBurn, CompOpHardLight, CompOpSoftLight, CompOpDifference,
		CompOpExclusion,
	}
	for want, got := range legacy {
		if got != CompOp(want) {
			t.Fatalf("legacy operator %d changed to %d", want, got)
		}
	}
}

func TestLegacyColorBurnOutputIsPinned(t *testing.T) {
	dst := []basics.Int8u{29, 91, 173, 147}
	NewCompositeBlenderPlain[color.Linear, order.RGBA](CompOpColorBurn).
		BlendPix(dst, 201, 77, 11, 113, 149)
	if want := []basics.Int8u{50, 69, 109, 175}; !equalPixel(dst, want) {
		t.Fatalf("legacy Color Burn = %v, want %v", dst, want)
	}
}

func TestDissolveUsesFullSeedAndUnquantizedAlpha(t *testing.T) {
	alpha := 0.5
	threshold := uint32(1 << 31)
	if !DissolveAccept(alpha, threshold-1) {
		t.Fatal("seed immediately below threshold should be accepted")
	}
	if DissolveAccept(alpha, threshold) {
		t.Fatal("seed at threshold should be rejected")
	}
	if !DissolveAccept(alpha+1e-9, threshold) {
		t.Fatal("sub-8-bit alpha precision should affect the threshold")
	}
	if DissolveSeed(17, -4) == DissolveSeed(18, -4) {
		t.Fatal("adjacent coordinates unexpectedly share a seed")
	}
}

func TestBlendPixFloatDoesNotQuantizeCoverage(t *testing.T) {
	bl := NewCompositeBlenderPlain[color.Linear, order.RGBA](CompOpSrcOver)
	low := []basics.Int8u{0, 0, 0, 255}
	high := append([]basics.Int8u(nil), low...)
	bl.BlendPixFloat(low, 255, 255, 255, 255, 128.1/255)
	bl.BlendPixFloat(high, 255, 255, 255, 255, 128.9/255)
	if equalPixel(low, high) {
		t.Fatalf("float coverages collapsed to one 8-bit cover: %v", low)
	}
}

func equalPixel(a, b []basics.Int8u) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func BenchmarkExtendedCompositeSpan(b *testing.B) {
	for _, size := range []int{256, 4096} {
		b.Run(string(rune(size)), func(b *testing.B) {
			dst := make([]basics.Int8u, size*4)
			bl := NewCompositeBlenderPlain[color.Linear, order.RGBA](CompOpHue)
			b.SetBytes(int64(len(dst)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bl.BlendSolidSpanStraight(dst, 192, 96, 32, 173, nil, size)
			}
		})
	}
}
