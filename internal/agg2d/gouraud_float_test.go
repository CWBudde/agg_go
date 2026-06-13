package agg2d

import "testing"

// TestAgg2DFloatGouraudTriangleSolid renders a single-color Gouraud triangle and
// verifies the interior is painted with that color.
func TestAgg2DFloatGouraudTriangleSolid(t *testing.T) {
	a, dst := setupFloatTarget(40, 40)
	red := NewColor(255, 0, 0, 255)

	a.GouraudTriangle(4, 4, 36, 4, 20, 36, red, red, red, 0)

	if in := dst.GetPixel(20, 18); !approxF(in.R, 1.0) || in.A <= 0 {
		t.Fatalf("Gouraud interior pixel(20,18) = %+v, want opaque red", in)
	}
	if out := dst.GetPixel(1, 38); out.A != 0 {
		t.Fatalf("outside-triangle pixel(1,38) alpha = %v, want 0", out.A)
	}
}

// TestAgg2DFloatGouraudTriangleInterpolates renders a three-color triangle and
// verifies the vertex regions take on their respective colors (smooth shading).
func TestAgg2DFloatGouraudTriangleInterpolates(t *testing.T) {
	a, dst := setupFloatTarget(60, 60)
	red := NewColor(255, 0, 0, 255)
	green := NewColor(0, 255, 0, 255)
	blue := NewColor(0, 0, 255, 255)

	// v1 top-left (red), v2 top-right (green), v3 bottom-center (blue).
	a.GouraudTriangle(6, 6, 54, 6, 30, 54, red, green, blue, 0)

	// Near v1: red should dominate.
	nearRed := dst.GetPixel(10, 9)
	if !(nearRed.R > nearRed.G && nearRed.R > nearRed.B) {
		t.Fatalf("near-v1 pixel should be red-dominant: %+v", nearRed)
	}
	// Near v2: green should dominate.
	nearGreen := dst.GetPixel(50, 9)
	if !(nearGreen.G > nearGreen.R && nearGreen.G > nearGreen.B) {
		t.Fatalf("near-v2 pixel should be green-dominant: %+v", nearGreen)
	}
	// Near v3: blue should dominate.
	nearBlue := dst.GetPixel(30, 50)
	if !(nearBlue.B > nearBlue.R && nearBlue.B > nearBlue.G) {
		t.Fatalf("near-v3 pixel should be blue-dominant: %+v", nearBlue)
	}
}
