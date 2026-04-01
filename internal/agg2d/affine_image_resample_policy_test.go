package agg2d

import (
	"bytes"
	"testing"

	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

func makeAffinePolicyFixture() []uint8 {
	buf := make([]uint8, 4*4*4)
	colors := [][4]uint8{
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
		{255, 255, 0, 255},
	}
	for y := range 4 {
		for x := range 4 {
			c := colors[(x+y)%len(colors)]
			i := (y*4 + x) * 4
			buf[i+0] = c[0]
			buf[i+1] = c[1]
			buf[i+2] = c[2]
			buf[i+3] = c[3]
		}
	}
	return buf
}

func renderAffinePolicyOutput(t *testing.T, filter ImageFilter, policy AffineImageResamplePolicy) []uint8 {
	t.Helper()
	buf := make([]uint8, 8*8*4)
	agg2d := NewAgg2D()
	agg2d.Attach(buf, 8, 8, 8*4)
	agg2d.ImageFilter(filter)
	agg2d.ImageResample(NoResample)
	agg2d.AffineImageResamplePolicy(policy)
	img := NewImage(makeAffinePolicyFixture(), 4, 4, 4*4)
	if err := agg2d.TransformImageSimple(img, 0.35, 0.4, 6.75, 5.8); err != nil {
		t.Fatalf("TransformImageSimple failed: %v", err)
	}
	return buf
}

func TestAffineImageResamplePolicyState(t *testing.T) {
	ctx := NewAgg2D()
	if got := ctx.GetAffineImageResamplePolicy(); got != AffineImageResampleAgg2D {
		t.Fatalf("default affine image resample policy = %v, want %v", got, AffineImageResampleAgg2D)
	}
	ctx.AffineImageResamplePolicy(AffineImageResamplePreferFiltered)
	if got := ctx.GetAffineImageResamplePolicy(); got != AffineImageResamplePreferFiltered {
		t.Fatalf("affine image resample policy = %v, want %v", got, AffineImageResamplePreferFiltered)
	}

	buf := make([]uint8, 8*8*4)
	ctx.Attach(buf, 8, 8, 8*4)
	if got := ctx.GetAffineImageResamplePolicy(); got != AffineImageResampleAgg2D {
		t.Fatalf("Attach reset affine image resample policy = %v, want %v", got, AffineImageResampleAgg2D)
	}
}

func TestAffineImageResamplePolicyDispatch(t *testing.T) {
	agg2d := NewAgg2D()
	imgBuf := make([]uint8, 4*4*4)
	img := NewImage(imgBuf, 4, 4, 4*4)
	source := newImagePixelFormat(img)
	identityInterpolator := span.NewSpanInterpolatorLinearDefault(transform.NewTransAffine())

	agg2d.ImageFilter(Bilinear)
	agg2d.ImageResample(NoResample)
	agg2d.AffineImageResamplePolicy(AffineImageResampleAgg2D)
	gen := agg2d.newImageFilterGenerator(source, identityInterpolator)
	if _, ok := gen.(*span.SpanImageFilterRGBABilinear[*imagePixelFormat, *span.SpanInterpolatorLinear[*transform.TransAffine]]); !ok {
		t.Fatalf("default affine generator = %T, want bilinear generator", gen)
	}

	agg2d.AffineImageResamplePolicy(AffineImageResamplePreferFiltered)
	gen = agg2d.newImageFilterGenerator(source, identityInterpolator)
	if _, ok := gen.(*span.SpanImageResampleRGBAAffine[*imagePixelFormat]); !ok {
		t.Fatalf("opt-in affine generator = %T, want affine resampler", gen)
	}

	agg2d.ImageFilter(NoFilter)
	gen = agg2d.newImageFilterGenerator(source, identityInterpolator)
	if _, ok := gen.(*span.SpanImageFilterRGBANN[*imagePixelFormat, *span.SpanInterpolatorLinear[*transform.TransAffine]]); !ok {
		t.Fatalf("nearest affine generator = %T, want nearest-neighbor generator", gen)
	}
}

func TestAffineImageResamplePolicyAffectsFilteredOutput(t *testing.T) {
	defaultBilinear := renderAffinePolicyOutput(t, Bilinear, AffineImageResampleAgg2D)
	optInBilinear := renderAffinePolicyOutput(t, Bilinear, AffineImageResamplePreferFiltered)
	if bytes.Equal(defaultBilinear, optInBilinear) {
		t.Fatal("expected opt-in affine policy to change bilinear filtered output")
	}

	optInBicubic := renderAffinePolicyOutput(t, Bicubic, AffineImageResamplePreferFiltered)
	if bytes.Equal(optInBilinear, optInBicubic) {
		t.Fatal("expected opt-in affine bilinear and bicubic outputs to diverge")
	}
}
