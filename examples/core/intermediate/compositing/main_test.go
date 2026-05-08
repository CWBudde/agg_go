package main

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/color"
)

func TestRenderFrontRoundedRectMatchesCPPReferenceSample(t *testing.T) {
	img := agg.NewImage(make([]uint8, frameWidth*frameHeight*4), frameWidth, frameHeight, -frameWidth*4)
	newDemo().Render(img)

	goImg := img.ToGoImage()
	if goImg == nil {
		t.Fatalf("ToGoImage returned nil")
	}

	gotLinear := goImg.RGBAAt(250, 145)
	got := color.ConvertToSRGBFromLinear(color.RGBA8[color.Linear]{
		R: gotLinear.R,
		G: gotLinear.G,
		B: gotLinear.B,
		A: gotLinear.A,
	})

	// Sampled from tests/visual/reference/cpp/examples/compositing.png in the
	// front rounded rectangle. The previous Go port rendered the rounded
	// rectangle twice, producing roughly (110,147,191).
	want := color.RGBA8[color.SRGB]{R: 155, G: 166, B: 178, A: 255}
	const tolerance = 4
	if diffU8(got.R, want.R) > tolerance ||
		diffU8(got.G, want.G) > tolerance ||
		diffU8(got.B, want.B) > tolerance ||
		diffU8(got.A, want.A) > tolerance {
		t.Fatalf("sample at (250,145) = RGBA(%d,%d,%d,%d), want near RGBA(%d,%d,%d,%d)",
			got.R, got.G, got.B, got.A,
			want.R, want.G, want.B, want.A)
	}
}

func diffU8(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
