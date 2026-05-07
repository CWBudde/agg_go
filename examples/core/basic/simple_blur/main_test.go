package main

import "testing"

func TestLionBaseOffsetMatchesCPPParseLion(t *testing.T) {
	ld := liondata()
	x1, y1, x2, y2 := getLionBoundingRect(ld)

	baseDX, baseDY := getLionBaseOffset(ld)

	if baseDX != (x2-x1)*0.5 {
		t.Fatalf("baseDX = %v, want %v", baseDX, (x2-x1)*0.5)
	}
	if baseDY != (y2-y1)*0.5 {
		t.Fatalf("baseDY = %v, want %v", baseDY, (y2-y1)*0.5)
	}

	centerY := (y1 + y2) * 0.5
	if baseDY == centerY {
		t.Fatalf("test is not guarding the C++ offset difference: baseDY=%v centerY=%v", baseDY, centerY)
	}
}
