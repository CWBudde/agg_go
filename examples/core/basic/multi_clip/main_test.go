package main

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

func TestNewClibcRandMatchesAlphaMask2Seed1State(t *testing.T) {
	t.Parallel()

	want := [31]int32{
		-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
		1627687941, -179304937, -2073333483, 1780058412, -1989503057,
		-615974602, 344556628, 939512070, -1249116260, 1507946756,
		-812545463, 154635395, 1388815473, -1926676823, 525320961,
		-1009028674, 968117788, -123449607, 1284210865, 435012392,
		-2017506339, -911064859, -370259173, 1132637927, 1398500161, -205601318,
	}

	rng := newClibcRand(1)
	if rng.fptr != 3 || rng.rptr != 0 {
		t.Fatalf("pointers = (%d,%d), want (3,0)", rng.fptr, rng.rptr)
	}
	if rng.state != want {
		t.Fatalf("state = %v, want %v", rng.state, want)
	}
}

func TestRGBARandRTLMatchesCppSrgba8Conversion(t *testing.T) {
	rng := newClibcRand(1)
	got := rgbaRandRTL(rng, 0x7F)

	expectedRng := newClibcRand(1)
	raw := color.RGBA8[color.SRGB]{
		A: uint8(expectedRng.randAnd(0x7F) + 0x7F),
		B: uint8(expectedRng.randAnd(0x7F)),
		G: uint8(expectedRng.randAnd(0x7F)),
		R: uint8(expectedRng.randAnd(0x7F)),
	}
	want := color.ConvertRGBA8SRGBToLinear(raw)

	if got != want {
		t.Fatalf("rgbaRandRTL = %+v, want C++ srgba8 decoded to linear %+v from raw %+v", got, want, raw)
	}
}

func TestRGBARandRGBRTLMatchesCppSrgba8Conversion(t *testing.T) {
	rng := newClibcRand(1)
	got := rgbaRandRGBRTL(rng, 255)

	expectedRng := newClibcRand(1)
	raw := color.RGBA8[color.SRGB]{
		B: uint8(expectedRng.randAnd(0x7F)),
		G: uint8(expectedRng.randAnd(0x7F)),
		R: uint8(expectedRng.randAnd(0x7F)),
		A: 255,
	}
	want := color.ConvertRGBA8SRGBToLinear(raw)

	if got != want {
		t.Fatalf("rgbaRandRGBRTL = %+v, want C++ srgba8 decoded to linear %+v from raw %+v", got, want, raw)
	}
}

func TestNewDemoUsesCppLionBoundingRect(t *testing.T) {
	d := newDemo()

	if d.baseDx != 119.0 {
		t.Fatalf("baseDx = %v, want C++ bounding_rect delta 119", d.baseDx)
	}
	if d.baseDy != 188.5 {
		t.Fatalf("baseDy = %v, want C++ bounding_rect delta 188.5", d.baseDy)
	}
}
