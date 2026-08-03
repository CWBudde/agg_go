package agg

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestDrawImageAffineTransparentEdgeDiffersFromClamp(t *testing.T) {
	source := NewImage([]byte{255, 40, 0, 255}, 1, 1, 4)
	transform := NewTransformationsFromValues(3, 0, 0, 3, 0, 0)
	draw := func(edge ImageEdgeMode) *Image {
		destination := CreateImage(3, 3)
		err := DrawImageAffine(destination, source, Rect{X2: 1, Y2: 1}, transform, ImageTransformOptions{
			Filter:           ImageFilterBilinear,
			EdgeMode:         edge,
			SourceAlpha:      AlphaStraight,
			DestinationAlpha: AlphaStraight,
			BlendMode:        BlendSrcOver,
			Opacity:          1,
		})
		if err != nil {
			t.Fatal(err)
		}
		return destination
	}
	clamped := draw(ImageEdgeClamp)
	transparent := draw(ImageEdgeTransparent)
	if got := affineAlphaSum(clamped); got != 9*255 {
		t.Fatalf("clamped alpha sum = %d, want %d", got, 9*255)
	}
	transparentSum := affineAlphaSum(transparent)
	if transparentSum <= 0 || transparentSum >= affineAlphaSum(clamped) {
		t.Fatalf("transparent alpha sum = %d, clamp = %d", transparentSum, affineAlphaSum(clamped))
	}
}

func TestDrawImageAffineStraightAndPremultipliedSourcesAgree(t *testing.T) {
	straight := NewImage([]byte{
		255, 0, 0, 128,
		0, 255, 0, 0,
	}, 2, 1, 8)
	premultiplied := NewImage([]byte{
		128, 0, 0, 128,
		0, 0, 0, 0,
	}, 2, 1, 8)
	transform := NewTransformationsFromValues(2, 0, 0, 2, 0, 0)
	draw := func(source *Image, alpha AlphaMode) *Image {
		destination := CreateImage(4, 2)
		if err := DrawImageAffine(destination, source, Rect{X2: 2, Y2: 1}, transform, ImageTransformOptions{
			Filter:           ImageFilterBilinear,
			EdgeMode:         ImageEdgeTransparent,
			SourceAlpha:      alpha,
			DestinationAlpha: AlphaStraight,
			BlendMode:        BlendSrcOver,
			Opacity:          1,
		}); err != nil {
			t.Fatal(err)
		}
		return destination
	}
	straightResult := draw(straight, AlphaStraight)
	premultipliedResult := draw(premultiplied, AlphaPremultiplied)
	for index := range straightResult.Data {
		difference := int(straightResult.Data[index]) - int(premultipliedResult.Data[index])
		if difference < -1 || difference > 1 {
			t.Fatalf("byte %d differs: straight=%d premultiplied=%d", index, straightResult.Data[index], premultipliedResult.Data[index])
		}
	}
}

func TestDrawImageAffineStraightAndPremultipliedDestinationsAgree(t *testing.T) {
	source := NewImage([]byte{220, 40, 10, 160}, 1, 1, 4)
	straight := NewImage([]byte{20, 40, 200, 128}, 1, 1, 4)
	premultiplied := NewImage([]byte{10, 20, 100, 128}, 1, 1, 4)
	draw := func(destination *Image, alpha AlphaMode) {
		t.Helper()
		if err := DrawImageAffine(destination, source, Rect{X2: 1, Y2: 1}, NewTransformationsFromValues(1, 0, 0, 1, 0, 0), ImageTransformOptions{
			Filter:           ImageFilterBilinear,
			EdgeMode:         ImageEdgeClamp,
			SourceAlpha:      AlphaStraight,
			DestinationAlpha: alpha,
			BlendMode:        BlendSrcOver,
			Opacity:          1,
		}); err != nil {
			t.Fatal(err)
		}
	}
	draw(straight, AlphaStraight)
	draw(premultiplied, AlphaPremultiplied)
	if err := premultiplied.Demultiply(); err != nil {
		t.Fatal(err)
	}
	for index := range straight.Data {
		difference := int(straight.Data[index]) - int(premultiplied.Data[index])
		if difference < -2 || difference > 2 {
			t.Fatalf("byte %d differs: straight=%d premultiplied=%d", index, straight.Data[index], premultiplied.Data[index])
		}
	}
}

func TestDrawImageAffineOnePixelDestinationAndClip(t *testing.T) {
	source := NewImage([]byte{240, 80, 20, 128}, 1, 1, 4)
	destination := CreateImage(1, 1)
	clip := Rect{X2: 1, Y2: 1}
	if err := DrawImageAffine(destination, source, Rect{X2: 1, Y2: 1}, NewTransformationsFromValues(1, 0, 0, 1, 0, 0), ImageTransformOptions{
		Filter:           ImageFilterBilinear,
		EdgeMode:         ImageEdgeClamp,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrcOver,
		Opacity:          1,
		Clip:             &clip,
	}); err != nil {
		t.Fatal(err)
	}
	if destination.Data[3] == 0 {
		t.Fatalf("1x1 destination was not drawn: %v", destination.Data)
	}
}

func TestDrawImageAffineSampleOffsetDoesNotMoveDestinationPolygon(t *testing.T) {
	source := NewImage([]byte{255, 0, 0, 255, 0, 0, 255, 255}, 2, 1, 8)
	destination := CreateImage(2, 1)
	identity := NewTransformationsFromValues(1, 0, 0, 1, 0, 0)
	if err := DrawImageAffine(destination, source, Rect{X2: 2, Y2: 1}, identity, ImageTransformOptions{
		Filter:           ImageFilterBilinear,
		Resample:         NoResample,
		EdgeMode:         ImageEdgeClamp,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrc,
		Opacity:          1,
		SampleOffset:     Point{X: 0.5, Y: 0.5},
	}); err != nil {
		t.Fatal(err)
	}
	if destination.Data[3] != 255 || destination.Data[7] != 255 {
		t.Fatalf("sample offset changed polygon coverage: %v", destination.Data)
	}
	if first := destination.Data[:4]; first[0] < 120 || first[0] > 135 || first[2] < 120 || first[2] > 135 {
		t.Fatalf("offset first sample = %v, want an even red/blue blend", first)
	}
}

func TestDrawImageAffineClipAndPaddedStride(t *testing.T) {
	const sourceStride = 12
	sourceBytes := make([]byte, sourceStride*2)
	for index := range sourceBytes {
		sourceBytes[index] = 0xa5
	}
	source := NewImage(sourceBytes, 2, 2, sourceStride)
	for y := 0; y < 2; y++ {
		row := source.renBuf.Row(y)
		copy(row[:8], []byte{220, 30, 10, 255, 20, 210, 40, 255})
	}

	const destinationStride = 20
	destinationBytes := make([]byte, destinationStride*3)
	for index := range destinationBytes {
		destinationBytes[index] = 0xa5
	}
	destination := NewImage(destinationBytes, 4, 3, destinationStride)
	for y := 0; y < destination.Height(); y++ {
		row := destination.renBuf.Row(y)
		for x := 0; x < destination.Width(); x++ {
			copy(row[x*4:x*4+4], []byte{5, 10, 15, 255})
		}
	}
	clip := Rect{X1: 1, Y1: 1, X2: 3, Y2: 2}
	if err := DrawImageAffine(destination, source, Rect{X2: 2, Y2: 2}, NewTransformationsFromValues(2, 0, 0, 1.5, 0, 0), ImageTransformOptions{
		Filter:           ImageFilterBilinear,
		EdgeMode:         ImageEdgeTransparent,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrcOver,
		Opacity:          1,
		Clip:             &clip,
	}); err != nil {
		t.Fatal(err)
	}
	for y := 0; y < destination.Height(); y++ {
		row := destination.renBuf.Row(y)
		for x := 0; x < destination.Width(); x++ {
			changed := row[x*4] != 5 || row[x*4+1] != 10 || row[x*4+2] != 15 || row[x*4+3] != 255
			wantChanged := y == 1 && x >= 1 && x < 3
			if changed != wantChanged {
				t.Fatalf("pixel (%d,%d) changed=%t, want %t: %v", x, y, changed, wantChanged, row[x*4:x*4+4])
			}
		}
		for index, value := range row[16:] {
			if value != 0xa5 {
				t.Fatalf("row %d padding byte %d = %#x, want 0xa5", y, index, value)
			}
		}
	}
}

func TestDrawImageAffineNegativeSourceStride(t *testing.T) {
	// With a negative stride, logical row zero is stored last.
	source := NewImage([]byte{
		0, 0, 255, 255,
		255, 0, 0, 255,
	}, 1, 2, -4)
	destination := CreateImage(1, 2)
	if err := DrawImageAffine(destination, source, Rect{X2: 1, Y2: 2}, NewTransformationsFromValues(1, 0, 0, 1, 0, 0), ImageTransformOptions{
		Filter:           ImageFilterNoFilter,
		EdgeMode:         ImageEdgeTransparent,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrcOver,
		Opacity:          1,
	}); err != nil {
		t.Fatal(err)
	}
	if destination.Data[0] < 250 || destination.Data[2] != 0 {
		t.Fatalf("top pixel = %v, want red", destination.Data[:4])
	}
	if destination.Data[4] != 0 || destination.Data[6] < 250 {
		t.Fatalf("bottom pixel = %v, want blue", destination.Data[4:8])
	}
}

func TestImageAlphaConversionRespectsSubimageStride(t *testing.T) {
	buffer := []byte{
		9, 9, 9, 9, 200, 100, 50, 128, 7, 7, 7, 7,
		8, 8, 8, 8, 120, 60, 30, 64, 6, 6, 6, 6,
	}
	image := NewImage(buffer[4:], 1, 2, 12)
	if err := image.Premultiply(); err != nil {
		t.Fatal(err)
	}
	if got := buffer[4:8]; got[0] != 100 || got[1] != 50 || got[2] != 25 || got[3] != 128 {
		t.Fatalf("first converted pixel = %v", got)
	}
	if got := buffer[16:20]; got[0] != 30 || got[1] != 15 || got[2] != 7 || got[3] != 64 {
		t.Fatalf("second converted pixel = %v", got)
	}
	if got := buffer[8:12]; got[0] != 7 || got[1] != 7 || got[2] != 7 || got[3] != 7 {
		t.Fatalf("row padding changed: %v", got)
	}
	if err := image.Demultiply(); err != nil {
		t.Fatal(err)
	}
	if got := buffer[8:12]; got[0] != 7 || got[1] != 7 || got[2] != 7 || got[3] != 7 {
		t.Fatalf("row padding changed during demultiply: %v", got)
	}
}

func TestDrawImageAffineOverlappingStorageSnapshotsSource(t *testing.T) {
	original := []byte{
		255, 0, 0, 255,
		0, 255, 0, 255,
		0, 0, 255, 255,
	}
	transform := NewTransformationsFromValues(1, 0, 0, 1, 0, 1)
	opts := ImageTransformOptions{
		Filter:           ImageFilterNoFilter,
		Resample:         NoResample,
		EdgeMode:         ImageEdgeClamp,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrc,
		Opacity:          1,
	}
	expectedPixels := append([]byte(nil), original...)
	if err := DrawImageAffine(
		NewImage(expectedPixels, 1, 3, 4),
		NewImage(append([]byte(nil), original...), 1, 3, 4),
		Rect{X2: 1, Y2: 2}, transform, opts,
	); err != nil {
		t.Fatal(err)
	}
	actualPixels := append([]byte(nil), original...)
	aliased := NewImage(actualPixels, 1, 3, 4)
	if err := DrawImageAffine(aliased, aliased, Rect{X2: 1, Y2: 2}, transform, opts); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actualPixels, expectedPixels) {
		t.Fatalf("overlapping affine = %v, want snapshot result %v", actualPixels, expectedPixels)
	}
}

func TestDrawImageAffineRejectsInvalidOptionsAndTransforms(t *testing.T) {
	destination := CreateImage(1, 1)
	source := CreateImageFromColor(1, 1, NewColor(255, 0, 0, 255))
	identity := NewTransformationsFromValues(1, 0, 0, 1, 0, 0)
	valid := ImageTransformOptions{
		Filter:           ImageFilterBilinear,
		Resample:         NoResample,
		EdgeMode:         ImageEdgeClamp,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrcOver,
		Opacity:          1,
	}
	tests := []struct {
		name      string
		transform *Transformations
		mutate    func(*ImageTransformOptions)
		want      string
	}{
		{name: "filter", transform: identity, mutate: func(opts *ImageTransformOptions) { opts.Filter = 99 }, want: "image filter"},
		{name: "resample", transform: identity, mutate: func(opts *ImageTransformOptions) { opts.Resample = 99 }, want: "resample mode"},
		{name: "blend", transform: identity, mutate: func(opts *ImageTransformOptions) { opts.BlendMode = 999 }, want: "blend mode"},
		{name: "dissolve", transform: identity, mutate: func(opts *ImageTransformOptions) { opts.BlendMode = BlendDissolve }, want: "dissolve"},
		{name: "non-finite", transform: NewTransformationsFromValues(math.NaN(), 0, 0, 1, 0, 0), want: "non-finite"},
		{name: "singular", transform: NewTransformationsFromValues(1, 0, 1, 0, 0, 0), want: "not invertible"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := valid
			if test.mutate != nil {
				test.mutate(&opts)
			}
			err := DrawImageAffine(destination, source, Rect{X2: 1, Y2: 1}, test.transform, opts)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func ExampleDrawImageAffine() {
	destination := CreateImage(1, 1)
	source := NewImage([]byte{255, 80, 0, 255}, 1, 1, 4)
	identity := NewTransformationsFromValues(1, 0, 0, 1, 0, 0)
	_ = DrawImageAffine(destination, source, Rect{X2: 1, Y2: 1}, identity, ImageTransformOptions{
		Filter:           ImageFilterBilinear,
		EdgeMode:         ImageEdgeClamp,
		SourceAlpha:      AlphaStraight,
		DestinationAlpha: AlphaStraight,
		BlendMode:        BlendSrcOver,
		Opacity:          1,
	})
	fmt.Printf("%v\n", destination.Data)
	// Output: [255 80 0 255]
}

func affineAlphaSum(image *Image) int {
	total := 0
	for y := 0; y < image.Height(); y++ {
		row := image.renBuf.Row(y)
		for x := 0; x < image.Width(); x++ {
			total += int(row[x*4+3])
		}
	}
	return total
}

func BenchmarkDrawImageAffineOpaque(b *testing.B) {
	const size = 512
	source := CreateImageFromColor(size, size, NewColor(210, 90, 30, 255))
	angle := 7 * math.Pi / 180
	cosine, sine := math.Cos(angle), math.Sin(angle)
	tests := []struct {
		name      string
		filter    ImageFilter
		transform *Transformations
	}{
		{"subpixel-bilinear", ImageFilterBilinear, NewTransformationsFromValues(1, 0, 0, 1, 0.25, 0.25)},
		{"rotated-bilinear", ImageFilterBilinear, NewTransformationsFromValues(
			cosine, sine, -sine, cosine,
			size/2-cosine*size/2+sine*size/2,
			size/2-sine*size/2-cosine*size/2,
		)},
		{"10x-nearest", ImageFilterNoFilter, NewTransformationsFromValues(10, 0, 0, 10, 0, 0)},
	}
	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			destination := CreateImage(size, size)
			opts := ImageTransformOptions{
				Filter:           test.filter,
				EdgeMode:         ImageEdgeClamp,
				SourceAlpha:      AlphaStraight,
				DestinationAlpha: AlphaStraight,
				BlendMode:        BlendSrcOver,
				Opacity:          1,
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if err := DrawImageAffine(destination, source, Rect{X2: size, Y2: size}, test.transform, opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
