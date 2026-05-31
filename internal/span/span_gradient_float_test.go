package span

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/transform"
)

// buildRGBA32LUT makes a 256-entry black->white linear ramp LUT.
func buildRGBA32LUT() []color.RGBA32[color.Linear] {
	lut := make([]color.RGBA32[color.Linear], 256)
	for i := range lut {
		v := float32(i) / 255.0
		lut[i] = color.RGBA32[color.Linear]{R: v, G: v, B: v, A: 1.0}
	}
	return lut
}

func TestGradientPrebuiltColorRGBA32(t *testing.T) {
	lut := buildRGBA32LUT()
	cf := NewGradientPrebuiltColorRGBA32[color.Linear](lut)
	if cf.Size() != 256 {
		t.Fatalf("Size() = %d, want 256", cf.Size())
	}
	c := cf.ColorAt(128)
	if c.R != float32(128)/255.0 || c.A != 1.0 {
		t.Fatalf("ColorAt(128) = %+v, unexpected", c)
	}
}

func TestNewLinearGradientFromLUT32GeneratesColors(t *testing.T) {
	trans := transform.NewTransAffine()
	interp := NewSpanInterpolatorLinearDefault(trans)
	lut := buildRGBA32LUT()

	gen := NewLinearGradientFromLUT32(interp, lut, 0.0, 100.0)
	gen.Prepare()

	out := make([]color.RGBA32[color.Linear], 10)
	gen.Generate(out, 0, 0, 10)

	// Colors should vary along the span (not all identical).
	allSame := true
	for _, c := range out {
		if c.R != out[0].R {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("linear gradient produced a constant span, expected variation")
	}
}

func TestNewRadialGradientFromLUT32GeneratesColors(t *testing.T) {
	trans := transform.NewTransAffine()
	interp := NewSpanInterpolatorLinearDefault(trans)
	lut := buildRGBA32LUT()

	gen := NewRadialGradientFromLUT32(interp, lut, 0.0, 50.0)
	gen.Prepare()

	out := make([]color.RGBA32[color.Linear], 5)
	gen.Generate(out, 10, 10, 5)

	// Alpha from the LUT (1.0) must be preserved.
	for i, c := range out {
		if c.A != 1.0 {
			t.Errorf("out[%d].A = %v, want 1.0", i, c.A)
		}
	}
}
