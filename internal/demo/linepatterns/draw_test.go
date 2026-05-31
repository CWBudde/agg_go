package linepatterns

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestImagePatternSourceConvertsSRGBPatternPixelToLinear(t *testing.T) {
	src := imagePatternSource{img: PatternImage{
		Width:  1,
		Height: 1,
		Pixels: []uint32{0x666666},
	}}

	got := src.Pixel(0, 0)
	want := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{
		R: 102,
		G: 102,
		B: 102,
		A: BrightnessToAlpha(34 + 34 + 34),
	}).ConvertToRGBA()

	if got != want {
		t.Fatalf("Pixel(0,0) = %+v, want %+v", got, want)
	}
}
