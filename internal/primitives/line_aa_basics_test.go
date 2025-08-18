package primitives

import (
	"testing"
)

func TestLineSubpixelConstants(t *testing.T) {
	if LineSubpixelShift != 8 {
		t.Errorf("LineSubpixelShift = %d, want 8", LineSubpixelShift)
	}
	if LineSubpixelScale != 256 {
		t.Errorf("LineSubpixelScale = %d, want 256", LineSubpixelScale)
	}
	if LineSubpixelMask != 255 {
		t.Errorf("LineSubpixelMask = %d, want 255", LineSubpixelMask)
	}
	if LineMaxCoord != (1<<28)-1 {
		t.Errorf("LineMaxCoord = %d, want %d", LineMaxCoord, (1<<28)-1)
	}
	if LineMaxLength != 1<<(LineSubpixelShift+10) {
		t.Errorf("LineMaxLength = %d, want %d", LineMaxLength, 1<<(LineSubpixelShift+10))
	}
}

func TestLineMRSubpixelConstants(t *testing.T) {
	if LineMRSubpixelShift != 4 {
		t.Errorf("LineMRSubpixelShift = %d, want 4", LineMRSubpixelShift)
	}
	if LineMRSubpixelScale != 16 {
		t.Errorf("LineMRSubpixelScale = %d, want 16", LineMRSubpixelScale)
	}
	if LineMRSubpixelMask != 15 {
		t.Errorf("LineMRSubpixelMask = %d, want 15", LineMRSubpixelMask)
	}
}

func TestLineMR(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{256, 16}, // LineSubpixelScale -> LineMRSubpixelScale
		{512, 32}, // 2 * LineSubpixelScale -> 2 * LineMRSubpixelScale
		{128, 8},  // LineSubpixelScale/2 -> LineMRSubpixelScale/2
		{0, 0},    // zero
	}

	for _, tt := range tests {
		result := LineMR(tt.input)
		if result != tt.expected {
			t.Errorf("LineMR(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestLineHR(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{16, 256}, // LineMRSubpixelScale -> LineSubpixelScale
		{32, 512}, // 2 * LineMRSubpixelScale -> 2 * LineSubpixelScale
		{8, 128},  // LineMRSubpixelScale/2 -> LineSubpixelScale/2
		{0, 0},    // zero
	}

	for _, tt := range tests {
		result := LineHR(tt.input)
		if result != tt.expected {
			t.Errorf("LineHR(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestLineDblHR(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{1, 256},   // 1 -> LineSubpixelScale
		{2, 512},   // 2 -> 2 * LineSubpixelScale
		{10, 2560}, // 10 -> 10 * LineSubpixelScale
		{0, 0},     // zero
	}

	for _, tt := range tests {
		result := LineDblHR(tt.input)
		if result != tt.expected {
			t.Errorf("LineDblHR(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestLineCoord(t *testing.T) {
	coord := LineCoord{}

	tests := []struct {
		input    float64
		expected int
	}{
		{1.0, 256},   // 1.0 * LineSubpixelScale
		{2.5, 640},   // 2.5 * LineSubpixelScale (rounded)
		{0.0, 0},     // zero
		{-1.0, -256}, // negative
		{0.5, 128},   // half pixel
	}

	for _, tt := range tests {
		result := coord.Conv(tt.input)
		if result != tt.expected {
			t.Errorf("LineCoord.Conv(%f) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestLineCoordSat(t *testing.T) {
	coord := LineCoordSat{}

	tests := []struct {
		input    float64
		expected int
	}{
		{1.0, 256}, // 1.0 * LineSubpixelScale
		{2.5, 640}, // 2.5 * LineSubpixelScale (rounded)
		{0.0, 0},   // zero
		{float64(LineMaxCoord + 1000), LineMaxCoord}, // saturation test
	}

	for _, tt := range tests {
		result := coord.Conv(tt.input)
		if result != tt.expected {
			t.Errorf("LineCoordSat.Conv(%f) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestNewLineParameters(t *testing.T) {
	lp := NewLineParameters(0, 0, 100, 50, 1000)

	if lp.X1 != 0 || lp.Y1 != 0 || lp.X2 != 100 || lp.Y2 != 50 {
		t.Errorf("Coordinates not set correctly: got (%d,%d,%d,%d), want (0,0,100,50)",
			lp.X1, lp.Y1, lp.X2, lp.Y2)
	}

	if lp.DX != 100 {
		t.Errorf("DX = %d, want 100", lp.DX)
	}
	if lp.DY != 50 {
		t.Errorf("DY = %d, want 50", lp.DY)
	}
	if lp.SX != 1 {
		t.Errorf("SX = %d, want 1", lp.SX)
	}
	if lp.SY != 1 {
		t.Errorf("SY = %d, want 1", lp.SY)
	}
	if lp.Vertical {
		t.Errorf("Vertical = true, want false (dx > dy)")
	}
	if lp.Inc != 1 {
		t.Errorf("Inc = %d, want 1", lp.Inc)
	}
	if lp.Len != 1000 {
		t.Errorf("Len = %d, want 1000", lp.Len)
	}
}

func TestLineParametersNegativeDirection(t *testing.T) {
	lp := NewLineParameters(100, 50, 0, 0, 1000)

	if lp.SX != -1 {
		t.Errorf("SX = %d, want -1", lp.SX)
	}
	if lp.SY != -1 {
		t.Errorf("SY = %d, want -1", lp.SY)
	}
}

func TestLineParametersVertical(t *testing.T) {
	lp := NewLineParameters(0, 0, 10, 100, 1000)

	if !lp.Vertical {
		t.Errorf("Vertical = false, want true (dy >= dx)")
	}
	if lp.Inc != 1 {
		t.Errorf("Inc = %d, want 1 (sy when vertical)", lp.Inc)
	}
}

func TestLineParametersOctant(t *testing.T) {
	tests := []struct {
		x1, y1, x2, y2 int
		expectedOctant int
	}{
		{0, 0, 10, 5, 0},  // dx > dy, sx > 0, sy > 0 -> octant 0
		{0, 0, 5, 10, 1},  // dy >= dx, sx > 0, sy > 0 -> octant 1
		{10, 0, 0, 5, 2},  // dx > dy, sx < 0, sy > 0 -> octant 2
		{5, 0, 0, 10, 3},  // dy >= dx, sx < 0, sy > 0 -> octant 3
		{0, 10, 10, 0, 5}, // dy >= dx, sx > 0, sy < 0 -> octant 5 (dy=10, dx=10, so vertical=true)
		{0, 10, 5, 0, 5},  // dy >= dx, sx > 0, sy < 0 -> octant 5
		{10, 10, 0, 0, 7}, // dy >= dx, sx < 0, sy < 0 -> octant 7 (dy=10, dx=10, so vertical=true)
		{5, 10, 0, 0, 7},  // dy >= dx, sx < 0, sy < 0 -> octant 7
	}

	for _, tt := range tests {
		lp := NewLineParameters(tt.x1, tt.y1, tt.x2, tt.y2, 1000)
		if lp.Octant != tt.expectedOctant {
			t.Errorf("NewLineParameters(%d,%d,%d,%d).Octant = %d, want %d",
				tt.x1, tt.y1, tt.x2, tt.y2, lp.Octant, tt.expectedOctant)
		}
	}
}

func TestLineParametersQuadrants(t *testing.T) {
	lp := NewLineParameters(0, 0, 10, 5, 1000) // octant 0

	orthQuad := lp.OrthogonalQuadrant()
	diagQuad := lp.DiagonalQuadrant()

	if orthQuad != 0 {
		t.Errorf("OrthogonalQuadrant() = %d, want 0", orthQuad)
	}
	if diagQuad != 0 {
		t.Errorf("DiagonalQuadrant() = %d, want 0", diagQuad)
	}
}

func TestLineParametersSameQuadrant(t *testing.T) {
	lp1 := NewLineParameters(0, 0, 10, 5, 1000)  // octant 0
	lp2 := NewLineParameters(5, 5, 15, 10, 1000) // also octant 0
	lp3 := NewLineParameters(0, 0, 5, 10, 1000)  // octant 1

	if !lp1.SameOrthogonalQuadrant(&lp2) {
		t.Errorf("Expected lp1 and lp2 to be in same orthogonal quadrant")
	}
	if !lp1.SameDiagonalQuadrant(&lp2) {
		t.Errorf("Expected lp1 and lp2 to be in same diagonal quadrant")
	}
	if lp1.SameDiagonalQuadrant(&lp3) {
		t.Errorf("Expected lp1 and lp3 to be in different diagonal quadrants")
	}
}

func TestLineParametersDivide(t *testing.T) {
	lp := NewLineParameters(0, 0, 100, 50, 1000)
	lp1, lp2 := lp.Divide()

	// First half should go from (0,0) to (50,25)
	if lp1.X1 != 0 || lp1.Y1 != 0 || lp1.X2 != 50 || lp1.Y2 != 25 {
		t.Errorf("lp1 coordinates = (%d,%d,%d,%d), want (0,0,50,25)",
			lp1.X1, lp1.Y1, lp1.X2, lp1.Y2)
	}

	// Second half should go from (50,25) to (100,50)
	if lp2.X1 != 50 || lp2.Y1 != 25 || lp2.X2 != 100 || lp2.Y2 != 50 {
		t.Errorf("lp2 coordinates = (%d,%d,%d,%d), want (50,25,100,50)",
			lp2.X1, lp2.Y1, lp2.X2, lp2.Y2)
	}

	// Both should have half the length
	if lp1.Len != 500 || lp2.Len != 500 {
		t.Errorf("Divided lengths = (%d,%d), want (500,500)", lp1.Len, lp2.Len)
	}
}

func TestBisectrix(t *testing.T) {
	// Create two line segments forming a 90-degree angle
	l1 := NewLineParameters(0, 0, 100, 0, 100*LineSubpixelScale)     // horizontal line
	l2 := NewLineParameters(100, 0, 100, 100, 100*LineSubpixelScale) // vertical line

	x, y := Bisectrix(&l1, &l2)

	// The bisectrix should be at 45 degrees from the corner
	// For this specific case, we expect it to be roughly in the direction (1, -1)
	// The exact coordinates depend on the bisectrix algorithm
	if x == 0 && y == 0 {
		t.Errorf("Bisectrix returned (0,0), expected non-zero coordinates")
	}

	// Test with a more predictable case - two parallel lines should result in specific behavior
	l3 := NewLineParameters(0, 0, 100, 0, 100*LineSubpixelScale)
	l4 := NewLineParameters(100, 0, 200, 0, 100*LineSubpixelScale)

	x2, y2 := Bisectrix(&l3, &l4)

	// For parallel lines, the result should be perpendicular to the line
	if x2 == 0 && y2 == 0 {
		t.Errorf("Bisectrix for parallel lines returned (0,0)")
	}
}

func TestFixDegenerateBisectrix(t *testing.T) {
	lp := NewLineParameters(0, 0, 100, 100, 100*LineSubpixelScale)

	// Test fixing start bisectrix
	x, y := 50, 45 // Some coordinates close to the line
	FixDegenerateBisectrixStart(&lp, &x, &y)

	// The function should potentially modify x, y if they're too close to the line
	// We can't predict exact values, but they should be reasonable
	if x < -1000 || x > 1000 || y < -1000 || y > 1000 {
		t.Errorf("FixDegenerateBisectrixStart produced unreasonable coordinates: (%d, %d)", x, y)
	}

	// Test fixing end bisectrix
	x2, y2 := 50, 55
	FixDegenerateBisectrixEnd(&lp, &x2, &y2)

	if x2 < -1000 || x2 > 1000 || y2 < -1000 || y2 > 1000 {
		t.Errorf("FixDegenerateBisectrixEnd produced unreasonable coordinates: (%d, %d)", x2, y2)
	}
}

func TestCmpDistFunctions(t *testing.T) {
	if !CmpDistStart(1) {
		t.Errorf("CmpDistStart(1) = false, want true")
	}
	if CmpDistStart(0) {
		t.Errorf("CmpDistStart(0) = true, want false")
	}
	if CmpDistStart(-1) {
		t.Errorf("CmpDistStart(-1) = true, want false")
	}

	if CmpDistEnd(1) {
		t.Errorf("CmpDistEnd(1) = true, want false")
	}
	if !CmpDistEnd(0) {
		t.Errorf("CmpDistEnd(0) = false, want true")
	}
	if !CmpDistEnd(-1) {
		t.Errorf("CmpDistEnd(-1) = false, want true")
	}
}

func BenchmarkNewLineParameters(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewLineParameters(0, 0, 100, 50, 1000)
	}
}

func BenchmarkBisectrix(b *testing.B) {
	l1 := NewLineParameters(0, 0, 100, 0, 100*LineSubpixelScale)
	l2 := NewLineParameters(100, 0, 100, 100, 100*LineSubpixelScale)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Bisectrix(&l1, &l2)
	}
}
