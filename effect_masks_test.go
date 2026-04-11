package agg

import "testing"

func TestAlphaMaskFromRGBAAndRenderMaskedSolidRGBA(t *testing.T) {
	mask := AlphaMaskFromRGBA([]byte{
		10, 20, 30, 64,
		40, 50, 60, 128,
	}, 2, 1)

	if got := mask.At(0, 0); got != 64 {
		t.Fatalf("mask alpha at (0,0) = %d, want 64", got)
	}
	if got := mask.At(1, 0); got != 128 {
		t.Fatalf("mask alpha at (1,0) = %d, want 128", got)
	}

	surface := RenderMaskedSolidRGBA(mask, NewColor(10, 20, 30, 200))
	if got := surface[0:4]; got[0] != 10 || got[1] != 20 || got[2] != 30 || got[3] != 50 {
		t.Fatalf("first pixel = %v, want [10 20 30 50]", got)
	}
	if got := surface[4:8]; got[0] != 10 || got[1] != 20 || got[2] != 30 || got[3] != 100 {
		t.Fatalf("second pixel = %v, want [10 20 30 100]", got)
	}
}

func TestRenderMaskedLinearGradientRGBA_UsesDirectionAndMaskAlpha(t *testing.T) {
	mask := AlphaMask{
		Width:  5,
		Height: 1,
		Pix:    []uint8{255, 255, 128, 255, 255},
	}

	surface := RenderMaskedLinearGradientRGBA(mask, LinearGradientFill{
		Start: NewColorRGB(32, 64, 255),
		End:   NewColorRGB(255, 196, 32),
		Angle: 0,
		Scale: 1,
	})

	left := effectRGBAAt(surface, 5, 0, 0)
	center := effectRGBAAt(surface, 5, 2, 0)
	right := effectRGBAAt(surface, 5, 4, 0)

	if left[2] <= left[0] {
		t.Fatalf("left pixel = %v, want blue-dominant gradient start", left)
	}
	if right[0] <= right[2] {
		t.Fatalf("right pixel = %v, want warm gradient end", right)
	}
	if center[3] != 128 {
		t.Fatalf("center alpha = %d, want masked alpha 128", center[3])
	}
}

func TestRenderMaskedCheckerPatternRGBA_RepeatsTile(t *testing.T) {
	mask := AlphaMask{
		Width:  8,
		Height: 8,
		Pix:    filledMask(8, 8, 255),
	}

	surface := RenderMaskedCheckerPatternRGBA(mask, CheckerPatternFill{
		First:  NewColorRGB(32, 160, 255),
		Second: NewColorRGB(255, 224, 96),
		Scale:  1,
	})

	if got := effectRGBAAt(surface, 8, 0, 0); got != ([4]uint8{32, 160, 255, 255}) {
		t.Fatalf("pixel (0,0) = %v, want first checker color", got)
	}
	if got := effectRGBAAt(surface, 8, 5, 0); got != ([4]uint8{255, 224, 96, 255}) {
		t.Fatalf("pixel (5,0) = %v, want second checker color", got)
	}
	if got := effectRGBAAt(surface, 8, 5, 5); got != ([4]uint8{32, 160, 255, 255}) {
		t.Fatalf("pixel (5,5) = %v, want repeated checker tile", got)
	}
}

func TestAlphaMaskOps_FilterAndCompose(t *testing.T) {
	base := AlphaMask{
		Width:  3,
		Height: 3,
		Pix: []uint8{
			0, 0, 0,
			0, 255, 0,
			0, 0, 0,
		},
	}

	shifted := base.Shifted(1, 0)
	if got := shifted.At(2, 1); got != 255 {
		t.Fatalf("shifted alpha at (2,1) = %d, want 255", got)
	}

	dilated := base.Dilated(1)
	if got := dilated.At(0, 0); got != 255 {
		t.Fatalf("dilated alpha at (0,0) = %d, want 255", got)
	}

	subtracted := dilated.Subtract(base)
	if got := subtracted.At(1, 1); got != 0 {
		t.Fatalf("subtracted center alpha = %d, want 0", got)
	}
	if got := subtracted.At(0, 1); got != 255 {
		t.Fatalf("subtracted edge alpha = %d, want 255", got)
	}

	blurred := base.Blurred(1)
	if got := blurred.At(1, 1); got == 0 || got == 255 {
		t.Fatalf("blurred center alpha = %d, want softened value", got)
	}
	if got := blurred.At(0, 1); got == 0 {
		t.Fatalf("blurred neighbor alpha = %d, want blur spread", got)
	}

	diff := base.AbsDiff(shifted)
	if got := diff.At(1, 1); got != 255 {
		t.Fatalf("abs diff alpha at (1,1) = %d, want 255", got)
	}
	if got := diff.At(2, 1); got != 255 {
		t.Fatalf("abs diff alpha at (2,1) = %d, want 255", got)
	}
}

func TestAlphaMaskIntersect_MultipliesPartialCoverage(t *testing.T) {
	left := AlphaMask{
		Width:  1,
		Height: 1,
		Pix:    []uint8{128},
	}
	right := AlphaMask{
		Width:  1,
		Height: 1,
		Pix:    []uint8{128},
	}

	intersected := left.Intersect(right)
	if got := intersected.At(0, 0); got != 64 {
		t.Fatalf("intersected alpha = %d, want multiplicative coverage 64", got)
	}
}

func effectRGBAAt(surface []byte, width, x, y int) [4]uint8 {
	offset := (y*width + x) * 4
	return [4]uint8{
		surface[offset],
		surface[offset+1],
		surface[offset+2],
		surface[offset+3],
	}
}

func filledMask(width, height int, alpha uint8) []uint8 {
	mask := make([]uint8, width*height)
	for index := range mask {
		mask[index] = alpha
	}
	return mask
}
