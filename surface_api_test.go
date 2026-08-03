package agg

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestCanonicalImageFilterConstantsPreserveLegacyAliases(t *testing.T) {
	if ImageFilterMitchell != 14 || ImageFilterSinc != 15 || ImageFilterLanczos != 16 {
		t.Fatalf("canonical filters = %d/%d/%d, want 14/15/16", ImageFilterMitchell, ImageFilterSinc, ImageFilterLanczos)
	}
	// Historical constants are deliberately frozen even though their names were
	// mapped to the wrong internal values.
	if FilterMitchell != 11 || FilterSinc != 14 || FilterLanczos != 15 {
		t.Fatalf("legacy filters changed: %d/%d/%d", FilterMitchell, FilterSinc, FilterLanczos)
	}
}

func TestGradientLUTAndSurfaceEndpoints(t *testing.T) {
	lut, err := NewGradientLUT([]GradientStop{
		{Position: 0, Color: NewColor(255, 0, 0, 255)},
		{Position: 1, Color: NewColor(0, 0, 255, 128)},
	}, 256)
	if err != nil {
		t.Fatal(err)
	}
	if got := lut.At(0); got != NewColor(255, 0, 0, 255) {
		t.Fatalf("first LUT color = %#v", got)
	}
	if got := lut.At(255); got != NewColor(0, 0, 255, 128) {
		t.Fatalf("last LUT color = %#v", got)
	}
	image := CreateImage(3, 1)
	if err := RenderGradient(image, lut, GradientRenderOptions{
		Shape: GradientShapeLinear,
		Start: Point{X: 0, Y: 0},
		End:   Point{X: 2, Y: 0},
	}); err != nil {
		t.Fatal(err)
	}
	if image.Data[0] < 250 || image.Data[2] > 5 {
		t.Fatalf("linear start pixel = %v", image.Data[:4])
	}
	if image.Data[8] > 5 || image.Data[10] < 250 {
		t.Fatalf("linear end pixel = %v", image.Data[8:12])
	}
}

func TestRenderGradientReflectedUsesEditorPeriod(t *testing.T) {
	pixels := make([]byte, 5*4)
	image := NewImage(pixels, 5, 1, 5*4)
	lut, err := NewGradientLUT([]GradientStop{
		{Position: 0, Color: NewColor(255, 0, 0, 255)},
		{Position: 1, Color: NewColor(0, 0, 255, 255)},
	}, 256)
	if err != nil {
		t.Fatal(err)
	}
	if err := RenderGradient(image, lut, GradientRenderOptions{
		Shape: GradientShapeReflected,
		Start: Point{X: 0, Y: 0},
		End:   Point{X: 4, Y: 0},
	}); err != nil {
		t.Fatal(err)
	}
	if got := pixels[:4]; got[0] != 0 || got[2] != 255 {
		t.Fatalf("start = %v, want end color", got)
	}
	if got := pixels[2*4 : 3*4]; got[0] != 255 || got[2] != 0 {
		t.Fatalf("midpoint = %v, want start color", got)
	}
	if got := pixels[4*4 : 5*4]; got[0] != 0 || got[2] != 255 {
		t.Fatalf("period boundary = %v, want end color", got)
	}
}

func TestRenderGradientPublicShapesProduceSpans(t *testing.T) {
	lut, err := NewGradientLUT([]GradientStop{
		{Position: 0, Color: NewColor(255, 20, 10, 255)},
		{Position: 1, Color: NewColor(0, 40, 250, 96)},
	}, 256)
	if err != nil {
		t.Fatal(err)
	}
	for _, shape := range []GradientShape{
		GradientShapeLinear,
		GradientShapeRadial,
		GradientShapeAngular,
		GradientShapeDiamond,
		GradientShapeReflected,
	} {
		t.Run(string(rune('0'+shape)), func(t *testing.T) {
			image := CreateImage(7, 7)
			if err := RenderGradient(image, lut, GradientRenderOptions{
				Shape: shape,
				Start: Point{X: 3, Y: 3},
				End:   Point{X: 6, Y: 3},
			}); err != nil {
				t.Fatal(err)
			}
			first := [4]byte(image.Data[:4])
			var varied bool
			for offset := 4; offset < len(image.Data); offset += 4 {
				if [4]byte(image.Data[offset:offset+4]) != first {
					varied = true
					break
				}
			}
			if !varied {
				t.Fatalf("shape %d produced one flat color", shape)
			}
		})
	}
}

func TestRenderGradientReverseDitherClipAndNegativeStride(t *testing.T) {
	lut, err := NewGradientLUT([]GradientStop{
		{Position: 0, Color: NewColor(255, 0, 0, 255)},
		{Position: 1, Color: NewColor(0, 0, 255, 64)},
	}, 256)
	if err != nil {
		t.Fatal(err)
	}
	normal := CreateImage(3, 1)
	reversed := CreateImage(3, 1)
	opts := GradientRenderOptions{Shape: GradientShapeLinear, Start: Point{}, End: Point{X: 2}}
	if err := RenderGradient(normal, lut, opts); err != nil {
		t.Fatal(err)
	}
	opts.Reverse = true
	if err := RenderGradient(reversed, lut, opts); err != nil {
		t.Fatal(err)
	}
	if [4]byte(normal.Data[:4]) != [4]byte(reversed.Data[8:12]) || [4]byte(normal.Data[8:12]) != [4]byte(reversed.Data[:4]) {
		t.Fatalf("reverse endpoints did not swap: normal=%v reverse=%v", normal.Data, reversed.Data)
	}

	makeClipped := func(dither bool) []byte {
		pixels := make([]byte, 3*2*4)
		for index := range pixels {
			pixels[index] = 0xaa
		}
		image := NewImage(pixels, 3, 2, -12)
		clip := Rect{X1: 1, Y1: 0, X2: 2, Y2: 2}
		if err := RenderGradient(image, lut, GradientRenderOptions{
			Shape:  GradientShapeLinear,
			Start:  Point{},
			End:    Point{X: 2},
			Clip:   &clip,
			Dither: dither,
		}); err != nil {
			t.Fatal(err)
		}
		for y := 0; y < 2; y++ {
			row := image.renBuf.Row(y)
			if !reflect.DeepEqual(row[:4], []byte{0xaa, 0xaa, 0xaa, 0xaa}) || !reflect.DeepEqual(row[8:12], []byte{0xaa, 0xaa, 0xaa, 0xaa}) {
				t.Fatalf("clip modified row %d outside x=1: %v", y, row)
			}
		}
		return pixels
	}
	withoutDither := makeClipped(false)
	withDither := makeClipped(true)
	if !reflect.DeepEqual(withDither, makeClipped(true)) {
		t.Fatal("dither output is not deterministic")
	}
	for _, alphaOffset := range []int{7, 19} {
		if withDither[alphaOffset] != withoutDither[alphaOffset] {
			t.Fatalf("dither changed alpha at byte %d: %d != %d", alphaOffset, withDither[alphaOffset], withoutDither[alphaOffset])
		}
	}
}

func TestRenderGradientValidation(t *testing.T) {
	validLUT, err := NewGradientLUT([]GradientStop{{Color: White}, {Position: 1, Color: Black}}, 2)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		lut  *GradientLUT
		opts GradientRenderOptions
		want string
	}{
		{name: "nil lut", opts: GradientRenderOptions{}, want: "nil or too small"},
		{name: "shape", lut: validLUT, opts: GradientRenderOptions{Shape: 99}, want: "invalid gradient shape"},
		{name: "endpoint", lut: validLUT, opts: GradientRenderOptions{End: Point{X: math.Inf(1)}}, want: "endpoints must be finite"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := RenderGradient(CreateImage(1, 1), test.lut, test.opts)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestDrawImageAffineIdentity(t *testing.T) {
	sourcePixels := []byte{
		255, 0, 0, 255, 0, 255, 0, 128,
		0, 0, 255, 64, 255, 255, 255, 255,
	}
	source := NewImage(append([]byte(nil), sourcePixels...), 2, 2, 8)
	destination := CreateImage(2, 2)
	identity := NewTransformationsFromValues(1, 0, 0, 1, 0, 0)
	if err := DrawImageAffine(destination, source, Rect{X1: 0, Y1: 0, X2: 2, Y2: 2}, identity, ImageTransformOptions{
		Filter:           ImageFilterNoFilter,
		Resample:         NoResample,
		EdgeMode:         ImageEdgeClamp,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrcOver,
		Opacity:          1,
	}); err != nil {
		t.Fatal(err)
	}
	if destination.Data[0] < 250 || destination.Data[3] < 250 {
		t.Fatalf("identity first pixel = %v", destination.Data[:4])
	}
}
