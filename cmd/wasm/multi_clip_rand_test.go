package main

import "testing"

func TestNewMultiClipRandMatchesAlphaMask2Seed1State(t *testing.T) {
	t.Parallel()

	want := [31]int32{
		-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
		1627687941, -179304937, -2073333483, 1780058412, -1989503057,
		-615974602, 344556628, 939512070, -1249116260, 1507946756,
		-812545463, 154635395, 1388815473, -1926676823, 525320961,
		-1009028674, 968117788, -123449607, 1284210865, 435012392,
		-2017506339, -911064859, -370259173, 1132637927, 1398500161, -205601318,
	}

	rng := newMultiClipRand(1)
	if rng.fptr != 3 || rng.rptr != 0 {
		t.Fatalf("pointers = (%d,%d), want (3,0)", rng.fptr, rng.rptr)
	}
	if rng.state != want {
		t.Fatalf("state = %v, want %v", rng.state, want)
	}
}
