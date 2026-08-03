package agg

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/agg_go/internal/simd"
)

func TestCompositeImageStraightSourceOver(t *testing.T) {
	dst := NewImage([]byte{0, 0, 255, 255}, 1, 1, 4)
	src := NewImage([]byte{255, 0, 0, 128}, 1, 1, 4)

	err := CompositeImage(dst, src, Rect{X2: 1, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode: BlendSrcOver,
		Opacity:   1,
		AlphaMode: AlphaStraight,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{128, 0, 127, 255}; !reflect.DeepEqual(dst.Data, want) {
		t.Fatalf("composite = %v, want %v", dst.Data, want)
	}
}

func TestCompositeImageSolidRowSIMDMatchesScalar(t *testing.T) {
	features := simd.DetectFeatures()
	if !features.HasAVX2 && !features.HasSSE2 {
		t.Skip("no supported SIMD source-over tier")
	}
	t.Cleanup(simd.ResetDetection)
	const width = 64
	sourcePixels := make([]byte, width*4)
	destinationPixels := make([]byte, width*4)
	for index := range width {
		copy(sourcePixels[index*4:index*4+4], []byte{180, 70, 25, 149})
		destinationPixels[index*4] = byte(index * 3)
		destinationPixels[index*4+1] = byte(220 - index*2)
		destinationPixels[index*4+2] = byte(30 + index)
		destinationPixels[index*4+3] = byte(80 + index*2)
	}
	render := func(forced simd.Features) []byte {
		simd.SetForcedFeatures(forced)
		result := append([]byte(nil), destinationPixels...)
		if err := CompositeImage(
			NewImage(result, width, 1, width*4),
			NewImage(sourcePixels, width, 1, width*4),
			Rect{X2: width, Y2: 1}, PointI{},
			CompositeOptions{BlendMode: BlendSrcOver, Opacity: 1, AlphaMode: AlphaStraight},
		); err != nil {
			t.Fatal(err)
		}
		return result
	}
	scalar := render(simd.Features{ForceGeneric: true})
	accelerated := render(features)
	if !reflect.DeepEqual(accelerated, scalar) {
		t.Fatal("CompositeImage SIMD solid-row path differs from scalar")
	}
}

func TestCompositeImagePremultiplied(t *testing.T) {
	dst := NewImage(make([]byte, 4), 1, 1, 4)
	src := NewImage([]byte{64, 32, 16, 128}, 1, 1, 4)

	err := CompositeImage(dst, src, Rect{X2: 1, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode: BlendSrcOver,
		Opacity:   1,
		AlphaMode: AlphaPremultiplied,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dst.Data, src.Data) {
		t.Fatalf("premultiplied composite = %v, want %v", dst.Data, src.Data)
	}
}

func TestCompositeImagePorterDuffZeroAlphaSource(t *testing.T) {
	tests := []struct {
		name string
		mode BlendMode
		want []byte
	}{
		{name: "clear", mode: BlendClear, want: []byte{0, 0, 0, 0}},
		{name: "source", mode: BlendSrc, want: []byte{0, 0, 0, 0}},
		{name: "source-in", mode: BlendSrcIn, want: []byte{0, 0, 0, 0}},
		{name: "destination-in", mode: BlendDstIn, want: []byte{0, 0, 0, 0}},
		{name: "source-out", mode: BlendSrcOut, want: []byte{0, 0, 0, 0}},
		{name: "destination-atop", mode: BlendDstAtop, want: []byte{0, 0, 0, 0}},
		{name: "source-over", mode: BlendSrcOver, want: []byte{100, 150, 200, 128}},
		{name: "destination-out", mode: BlendDstOut, want: []byte{100, 150, 200, 128}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destination := []byte{100, 150, 200, 128}
			err := CompositeImage(
				NewImage(destination, 1, 1, 4),
				NewImage([]byte{220, 80, 30, 0}, 1, 1, 4),
				Rect{X2: 1, Y2: 1}, PointI{},
				CompositeOptions{BlendMode: test.mode, Opacity: 1, AlphaMode: AlphaStraight},
			)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(destination, test.want) {
				t.Fatalf("zero-alpha result = %v, want %v", destination, test.want)
			}
		})
	}
}

func TestCompositeImageZeroCoverageDoesNotApplyPorterDuffOperator(t *testing.T) {
	destination := []byte{100, 150, 200, 128}
	mask := AlphaMask{Width: 1, Height: 1, Pix: []byte{0}}
	err := CompositeImage(
		NewImage(destination, 1, 1, 4),
		NewImage([]byte{220, 80, 30, 0}, 1, 1, 4),
		Rect{X2: 1, Y2: 1}, PointI{},
		CompositeOptions{BlendMode: BlendClear, Opacity: 1, AlphaMode: AlphaStraight, Mask: &mask},
	)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{100, 150, 200, 128}; !reflect.DeepEqual(destination, want) {
		t.Fatalf("zero-coverage result = %v, want %v", destination, want)
	}
}

func TestCompositeImagePremultipliedZeroAlphaSourceAppliesOperator(t *testing.T) {
	destination := []byte{50, 75, 100, 128}
	if err := CompositeImage(
		NewImage(destination, 1, 1, 4),
		NewImage([]byte{0, 0, 0, 0}, 1, 1, 4),
		Rect{X2: 1, Y2: 1}, PointI{},
		CompositeOptions{BlendMode: BlendDstIn, Opacity: 1, AlphaMode: AlphaPremultiplied},
	); err != nil {
		t.Fatal(err)
	}
	if want := []byte{0, 0, 0, 0}; !reflect.DeepEqual(destination, want) {
		t.Fatalf("premultiplied zero-alpha result = %v, want %v", destination, want)
	}
}

func TestCompositeImageKeepsOpacityInFloat64(t *testing.T) {
	dst := NewImage(make([]byte, 4), 1, 1, 4)
	src := NewImage([]byte{255, 0, 0, 2}, 1, 1, 4)

	err := CompositeImage(dst, src, Rect{X2: 1, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode: BlendSrcOver,
		Opacity:   0.504,
		AlphaMode: AlphaStraight,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Quantizing opacity to a cover of 129 before applying source alpha would
	// produce alpha 2. Retaining 0.504 through the equation produces alpha 1.
	if got, want := dst.Data[3], byte(1); got != want {
		t.Fatalf("alpha = %d, want %d", got, want)
	}
}

func TestCompositeImageDissolveUsesDestinationSeedAndEffectiveAlpha(t *testing.T) {
	dst := NewImage([]byte{
		0, 0, 255, 255,
		0, 0, 255, 255,
	}, 2, 1, 8)
	src := NewImage([]byte{
		255, 0, 0, 128,
		255, 0, 0, 128,
	}, 2, 1, 8)
	var coordinates []PointI
	seed := func(x, y int) uint32 {
		coordinates = append(coordinates, PointI{X: x, Y: y})
		if x == 0 {
			return 0
		}
		return ^uint32(0)
	}

	err := CompositeImage(dst, src, Rect{X2: 2, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode:    BlendDissolve,
		Opacity:      0.75,
		AlphaMode:    AlphaStraight,
		DissolveSeed: seed,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		255, 0, 0, 255,
		0, 0, 255, 255,
	}
	if !reflect.DeepEqual(dst.Data, want) {
		t.Fatalf("dissolve composite = %v, want %v", dst.Data, want)
	}
	if wantCoordinates := []PointI{{X: 0, Y: 0}, {X: 1, Y: 0}}; !reflect.DeepEqual(coordinates, wantCoordinates) {
		t.Fatalf("seed coordinates = %v, want %v", coordinates, wantCoordinates)
	}
}

func TestCompositeImageDissolvePremultipliedAcceptedPixelIsOpaque(t *testing.T) {
	dst := NewImage(make([]byte, 4), 1, 1, 4)
	src := NewImage([]byte{64, 32, 16, 128}, 1, 1, 4)

	err := CompositeImage(dst, src, Rect{X2: 1, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode:    BlendDissolve,
		Opacity:      1,
		AlphaMode:    AlphaPremultiplied,
		DissolveSeed: func(_, _ int) uint32 { return 0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{128, 64, 32, 255}; !reflect.DeepEqual(dst.Data, want) {
		t.Fatalf("premultiplied dissolve = %v, want %v", dst.Data, want)
	}
}

func TestCompositeImageDissolveClipMatchesFullRender(t *testing.T) {
	const width = 32
	srcPixels := make([]byte, width*4)
	for x := 0; x < width; x++ {
		offset := x * 4
		srcPixels[offset], srcPixels[offset+1], srcPixels[offset+2], srcPixels[offset+3] = 220, 90, 30, 117
	}
	src := NewImage(srcPixels, width, 1, width*4)
	full := NewImage(make([]byte, width*4), width, 1, width*4)
	clipped := NewImage(make([]byte, width*4), width, 1, width*4)
	opts := CompositeOptions{BlendMode: BlendDissolve, Opacity: 0.63, AlphaMode: AlphaStraight}
	if err := CompositeImage(full, src, Rect{X2: width, Y2: 1}, PointI{}, opts); err != nil {
		t.Fatal(err)
	}
	clip := Rect{X1: 7, Y1: 0, X2: 25, Y2: 1}
	opts.Clip = &clip
	if err := CompositeImage(clipped, src, Rect{X2: width, Y2: 1}, PointI{}, opts); err != nil {
		t.Fatal(err)
	}
	if got, want := clipped.Data[clip.X1*4:clip.X2*4], full.Data[clip.X1*4:clip.X2*4]; !reflect.DeepEqual(got, want) {
		t.Fatalf("clipped dissolve = %v, want full-render subset %v", got, want)
	}
}

func TestCompositeImageSupportsEveryBlendMode(t *testing.T) {
	modes := []BlendMode{
		BlendAlpha, BlendClear, BlendSrc, BlendDst, BlendSrcOver, BlendDstOver,
		BlendSrcIn, BlendDstIn, BlendSrcOut, BlendDstOut, BlendSrcAtop,
		BlendDstAtop, BlendXor, BlendAdd, BlendMultiply, BlendScreen,
		BlendOverlay, BlendDarken, BlendLighten, BlendColorDodge, BlendColorBurn,
		BlendHardLight, BlendSoftLight, BlendDifference, BlendExclusion,
		BlendDissolve, BlendLinearBurn, BlendDarkerColor, BlendLinearDodge,
		BlendLighterColor, BlendVividLight, BlendLinearLight, BlendPinLight,
		BlendHardMix, BlendSubtract, BlendDivide, BlendHue, BlendSaturation,
		BlendColor, BlendLuminosity, BlendColorBurnPhotoshop, BlendSoftLightPhotoshop,
	}
	for _, mode := range modes {
		t.Run(BlendModeToString(mode), func(t *testing.T) {
			dst := NewImage([]byte{30, 60, 90, 180}, 1, 1, 4)
			src := NewImage([]byte{200, 120, 40, 160}, 1, 1, 4)
			err := CompositeImage(dst, src, Rect{X2: 1, Y2: 1}, PointI{}, CompositeOptions{
				BlendMode: mode,
				Opacity:   0.7,
				AlphaMode: AlphaStraight,
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCompositeImagePhotoshopSoftLightIsAdditiveCompatibilityMode(t *testing.T) {
	destination := []byte{64, 128, 192, 173}
	source := []byte{192, 96, 32, 137}
	if err := CompositeImage(
		NewImage(destination, 1, 1, 4),
		NewImage(source, 1, 1, 4),
		Rect{X2: 1, Y2: 1},
		PointI{},
		CompositeOptions{BlendMode: BlendSoftLightPhotoshop, Opacity: 0.63, AlphaMode: AlphaStraight},
	); err != nil {
		t.Fatal(err)
	}
	if want := [4]byte{91, 119, 160, 201}; [4]byte(destination) != want {
		t.Fatalf("Photoshop Soft Light = %v, want %v", destination, want)
	}
}

func TestCompositeImageMaskOriginOpacityAndClip(t *testing.T) {
	dstPixels := []byte{
		0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0, 255,
	}
	srcPixels := []byte{
		255, 0, 0, 255, 255, 0, 0, 255, 255, 0, 0, 255, 255, 0, 0, 255,
	}
	dst := NewImage(dstPixels, 4, 1, 16)
	src := NewImage(srcPixels, 4, 1, 16)
	mask := AlphaMask{Width: 2, Height: 1, Pix: []byte{128, 255}}
	clip := Rect{X1: 1, Y1: 0, X2: 3, Y2: 1}

	err := CompositeImage(dst, src, Rect{X2: 4, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode:  BlendSrcOver,
		Opacity:    0.5,
		AlphaMode:  AlphaStraight,
		Mask:       &mask,
		MaskOrigin: PointI{X: 1},
		Clip:       &clip,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		0, 0, 0, 255, 64, 0, 0, 255, 128, 0, 0, 255, 0, 0, 0, 255,
	}
	if !reflect.DeepEqual(dst.Data, want) {
		t.Fatalf("masked composite = %v, want %v", dst.Data, want)
	}
}

func TestCompositeImageClipsSourceAndPreservesMapping(t *testing.T) {
	dst := NewImage(make([]byte, 4*3), 3, 1, 12)
	src := NewImage([]byte{10, 0, 0, 255, 20, 0, 0, 255}, 2, 1, 8)

	err := CompositeImage(dst, src, Rect{X1: -1, Y1: 0, X2: 2, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode: BlendSrc,
		Opacity:   1,
		AlphaMode: AlphaStraight,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0, 0, 0, 0, 10, 0, 0, 255, 20, 0, 0, 255}
	if !reflect.DeepEqual(dst.Data, want) {
		t.Fatalf("clipped composite = %v, want %v", dst.Data, want)
	}
}

func TestCompositeImageNegativeStride(t *testing.T) {
	// Negative stride presents the second physical row as logical row zero.
	src := NewImage([]byte{
		0, 255, 0, 255,
		255, 0, 0, 255,
	}, 1, 2, -4)
	dst := NewImage(make([]byte, 8), 1, 2, 4)

	err := CompositeImage(dst, src, Rect{X2: 1, Y2: 2}, PointI{}, CompositeOptions{
		BlendMode: BlendSrc,
		Opacity:   1,
		AlphaMode: AlphaStraight,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{255, 0, 0, 255, 0, 255, 0, 255}
	if !reflect.DeepEqual(dst.Data, want) {
		t.Fatalf("negative-stride composite = %v, want %v", dst.Data, want)
	}
}

func TestCompositeImageNegativeDestinationStrideAndPadding(t *testing.T) {
	src := NewImage([]byte{
		255, 0, 0, 255,
		0, 255, 0, 255,
	}, 1, 2, 4)
	dstPixels := []byte{
		0, 0, 0, 0, 0xee, 0xee, 0xee, 0xee,
		0, 0, 0, 0,
	}
	dst := NewImage(dstPixels, 1, 2, -8)

	err := CompositeImage(dst, src, Rect{X2: 1, Y2: 2}, PointI{}, CompositeOptions{
		BlendMode: BlendSrc,
		Opacity:   1,
		AlphaMode: AlphaStraight,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		0, 255, 0, 255, 0xee, 0xee, 0xee, 0xee,
		255, 0, 0, 255,
	}
	if !reflect.DeepEqual(dst.Data, want) {
		t.Fatalf("negative-stride destination = %v, want %v", dst.Data, want)
	}
}

func TestCompositeImageOverlappingStorageSnapshotsSource(t *testing.T) {
	pixels := []byte{
		255, 0, 0, 255,
		0, 255, 0, 255,
		0, 0, 255, 255,
	}
	image := NewImage(pixels, 1, 3, 4)

	err := CompositeImage(image, image, Rect{X2: 1, Y2: 2}, PointI{Y: 1}, CompositeOptions{
		BlendMode: BlendSrc,
		Opacity:   1,
		AlphaMode: AlphaStraight,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		255, 0, 0, 255,
		255, 0, 0, 255,
		0, 255, 0, 255,
	}
	if !reflect.DeepEqual(image.Data, want) {
		t.Fatalf("overlapping composite = %v, want %v", image.Data, want)
	}
}

func TestCompositeImageAllocationContract(t *testing.T) {
	const width = 32
	src := NewImage(make([]byte, width*4), width, 1, width*4)
	dst := NewImage(make([]byte, width*4), width, 1, width*4)
	bounds := Rect{X2: width, Y2: 1}
	opts := CompositeOptions{BlendMode: BlendSrcOver, Opacity: 0.75, AlphaMode: AlphaStraight}
	var compositeErr error
	if got := testing.AllocsPerRun(100, func() {
		compositeErr = CompositeImage(dst, src, bounds, PointI{}, opts)
	}); got != 0 {
		t.Fatalf("non-alias allocations = %g, want 0", got)
	}
	if compositeErr != nil {
		t.Fatal(compositeErr)
	}

	aliased := NewImage(make([]byte, width*4), width, 1, width*4)
	if got := testing.AllocsPerRun(100, func() {
		compositeErr = CompositeImage(aliased, aliased, bounds, PointI{}, opts)
	}); got != 1 {
		t.Fatalf("alias allocations = %g, want one clipped-rectangle snapshot", got)
	}
	if compositeErr != nil {
		t.Fatal(compositeErr)
	}
}

func TestCompositeImageValidation(t *testing.T) {
	valid := NewImage(make([]byte, 4), 1, 1, 4)
	tests := []struct {
		name string
		dst  *Image
		src  *Image
		opts CompositeOptions
		want string
	}{
		{name: "nil destination", src: valid, opts: CompositeOptions{Opacity: 1}, want: "destination image is nil"},
		{name: "nil source", dst: valid, opts: CompositeOptions{Opacity: 1}, want: "source image is nil"},
		{name: "short buffer", dst: valid, src: NewImage(make([]byte, 3), 1, 1, 4), opts: CompositeOptions{Opacity: 1}, want: "need at least 4"},
		{name: "short stride", dst: valid, src: NewImage(make([]byte, 4), 1, 1, 3), opts: CompositeOptions{Opacity: 1}, want: "smaller than RGBA row size"},
		{name: "alpha mode", dst: valid, src: valid, opts: CompositeOptions{Opacity: 1, AlphaMode: 3}, want: "invalid alpha mode"},
		{name: "opacity", dst: valid, src: valid, opts: CompositeOptions{Opacity: 2}, want: "opacity must be finite"},
		{name: "mask", dst: valid, src: valid, opts: CompositeOptions{Opacity: 1, Mask: &AlphaMask{Width: 2, Height: 1, Pix: []byte{1}}}, want: "need at least 2"},
		{name: "blend mode", dst: valid, src: valid, opts: CompositeOptions{BlendMode: 999, Opacity: 1}, want: "unsupported blend mode"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CompositeImage(test.dst, test.src, Rect{X2: 1, Y2: 1}, PointI{}, test.opts)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestDefaultDissolveSeedStable(t *testing.T) {
	tests := []struct {
		x, y int
		want uint32
	}{
		{0, 0, 0x00000000},
		{1, 0, 0x10b0f93f},
		{-1, 7, 0x2a8c22e7},
	}
	for _, test := range tests {
		if got := DefaultDissolveSeed(test.x, test.y); got != test.want {
			t.Fatalf("DefaultDissolveSeed(%d, %d) = %#08x, want %#08x", test.x, test.y, got, test.want)
		}
	}
}

func ExampleCompositeImage() {
	dst := NewImage([]byte{0, 0, 255, 255}, 1, 1, 4)
	src := NewImage([]byte{255, 0, 0, 128}, 1, 1, 4)
	err := CompositeImage(dst, src, Rect{X2: 1, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode: BlendSrcOver,
		Opacity:   1,
		AlphaMode: AlphaStraight,
	})
	fmt.Printf("%v %v\n", dst.Data, err)
	// Output: [128 0 127 255] <nil>
}

func BenchmarkCompositeImage(b *testing.B) {
	for _, width := range []int{256, 4096} {
		b.Run(fmt.Sprintf("width=%d", width), func(b *testing.B) {
			dst := NewImage(make([]byte, width*4), width, 1, width*4)
			srcPixels := make([]byte, width*4)
			for index := 0; index < width; index++ {
				offset := index * 4
				srcPixels[offset], srcPixels[offset+1], srcPixels[offset+2], srcPixels[offset+3] = 210, 80, 30, 160
			}
			src := NewImage(srcPixels, width, 1, width*4)
			opts := CompositeOptions{BlendMode: BlendSrcOver, Opacity: 0.73, AlphaMode: AlphaStraight}
			bounds := Rect{X2: width, Y2: 1}
			b.ReportAllocs()
			b.SetBytes(int64(width * 4))
			for b.Loop() {
				if err := CompositeImage(dst, src, bounds, PointI{}, opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
