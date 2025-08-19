package span

import (
	"testing"
)

func TestGrayColorCreation(t *testing.T) {
	c := GrayColor{V: 128, A: 200}
	if c.V != 128 || c.A != 200 {
		t.Errorf("GrayColor not created correctly: got V=%d A=%d", c.V, c.A)
	}
}

func TestSpanGouraudGrayCreation(t *testing.T) {
	sg := NewSpanGouraudGray()
	if sg == nil {
		t.Fatal("NewSpanGouraudGray returned nil")
	}
}

func TestSpanGouraudGrayWithTriangle(t *testing.T) {
	c1 := GrayColor{V: 255, A: 255} // White
	c2 := GrayColor{V: 128, A: 255} // Gray
	c3 := GrayColor{V: 0, A: 255}   // Black

	sg := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	if sg == nil {
		t.Fatal("NewSpanGouraudGrayWithTriangle returned nil")
	}

	coord := sg.Coord()
	if coord[0].Color.V != 255 {
		t.Errorf("First vertex color not set correctly: expected V=255, got V=%d", coord[0].Color.V)
	}
	if coord[1].Color.V != 128 {
		t.Errorf("Second vertex color not set correctly: expected V=128, got V=%d", coord[1].Color.V)
	}
	if coord[2].Color.V != 0 {
		t.Errorf("Third vertex color not set correctly: expected V=0, got V=%d", coord[2].Color.V)
	}
}

func TestGrayCalcInit(t *testing.T) {
	c1 := CoordType[GrayColor]{
		X: 0, Y: 0,
		Color: GrayColor{V: 100, A: 255},
	}
	c2 := CoordType[GrayColor]{
		X: 100, Y: 100,
		Color: GrayColor{V: 200, A: 128},
	}

	var calc GrayCalc
	calc.init(c1, c2)

	// Check initial values
	if calc.v1 != 100 || calc.a1 != 255 {
		t.Errorf("Initial values not set correctly: V=%d A=%d", calc.v1, calc.a1)
	}

	// Check deltas
	if calc.dv != 100 || calc.da != -127 {
		t.Errorf("Deltas not calculated correctly: dV=%d dA=%d", calc.dv, calc.da)
	}
}

func TestGrayCalcCalc(t *testing.T) {
	c1 := CoordType[GrayColor]{
		X: 0, Y: 0,
		Color: GrayColor{V: 0, A: 0},
	}
	c2 := CoordType[GrayColor]{
		X: 100, Y: 100,
		Color: GrayColor{V: 255, A: 255},
	}

	var calc GrayCalc
	calc.init(c1, c2)

	// Test interpolation at middle
	calc.calc(50.0)

	// Should be approximately halfway between colors
	tolerance := 5 // Allow some rounding error
	if absGray(calc.v-127) > tolerance || absGray(calc.a-127) > tolerance {
		t.Errorf("Midpoint interpolation incorrect: V=%d A=%d", calc.v, calc.a)
	}
}

func TestGrayCalcEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		c1, c2    CoordType[GrayColor]
		testY     float64
		expectedV int
		expectedA int
		tolerance int
	}{
		{
			name:      "Start point",
			c1:        CoordType[GrayColor]{X: 0, Y: 0, Color: GrayColor{V: 50, A: 100}},
			c2:        CoordType[GrayColor]{X: 100, Y: 100, Color: GrayColor{V: 200, A: 255}},
			testY:     0.0,
			expectedV: 50,
			expectedA: 100,
			tolerance: 2,
		},
		{
			name:      "End point",
			c1:        CoordType[GrayColor]{X: 0, Y: 0, Color: GrayColor{V: 50, A: 100}},
			c2:        CoordType[GrayColor]{X: 100, Y: 100, Color: GrayColor{V: 200, A: 255}},
			testY:     100.0,
			expectedV: 200,
			expectedA: 255,
			tolerance: 2,
		},
		{
			name:      "Beyond end (clamped)",
			c1:        CoordType[GrayColor]{X: 0, Y: 0, Color: GrayColor{V: 50, A: 100}},
			c2:        CoordType[GrayColor]{X: 100, Y: 100, Color: GrayColor{V: 200, A: 255}},
			testY:     200.0,
			expectedV: 200,
			expectedA: 255,
			tolerance: 2,
		},
		{
			name:      "Before start (clamped)",
			c1:        CoordType[GrayColor]{X: 0, Y: 0, Color: GrayColor{V: 50, A: 100}},
			c2:        CoordType[GrayColor]{X: 100, Y: 100, Color: GrayColor{V: 200, A: 255}},
			testY:     -50.0,
			expectedV: 50,
			expectedA: 100,
			tolerance: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calc GrayCalc
			calc.init(tt.c1, tt.c2)
			calc.calc(tt.testY)

			if absGray(calc.v-tt.expectedV) > tt.tolerance {
				t.Errorf("Expected V=%d ±%d, got V=%d", tt.expectedV, tt.tolerance, calc.v)
			}
			if absGray(calc.a-tt.expectedA) > tt.tolerance {
				t.Errorf("Expected A=%d ±%d, got A=%d", tt.expectedA, tt.tolerance, calc.a)
			}
		})
	}
}

func TestSpanGouraudGrayPrepare(t *testing.T) {
	c1 := GrayColor{V: 255, A: 255}
	c2 := GrayColor{V: 128, A: 255}
	c3 := GrayColor{V: 0, A: 255}

	sg := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 25, 50, 100, 0)
	sg.Prepare()

	// After prepare, y2 should be set to the middle vertex Y coordinate
	if sg.y2 != 25 {
		t.Errorf("y2 should be 25 after Prepare(), got %d", sg.y2)
	}

	// swap flag should be determined
	_ = sg.swap // Just verify it's accessible
}

func TestSpanGouraudGrayGenerate(t *testing.T) {
	// Create a simple triangle with gradient from white to gray to black
	c1 := GrayColor{V: 255, A: 255} // White at (0,0)
	c2 := GrayColor{V: 128, A: 255} // Gray at (100,0)
	c3 := GrayColor{V: 0, A: 255}   // Black at (50,50)

	sg := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 50, 0)
	sg.Prepare()

	// Generate a span at y=25 (middle height)
	span := make([]GrayColor, 10)
	sg.Generate(span, 25, 25, 10)

	// Verify that colors are interpolated (not all the same)
	allSame := true
	first := span[0]
	for i := 1; i < len(span); i++ {
		if span[i].V != first.V || span[i].A != first.A {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Generated span should have color variation")
	}

	// Verify colors are in valid range
	for i, color := range span {
		if color.V < 0 || color.V > 255 || color.A < 0 || color.A > 255 {
			t.Errorf("Span[%d] has invalid color values: V=%d A=%d", i, color.V, color.A)
		}
	}
}

func TestSpanGouraudGrayGenerateEdgeCases(t *testing.T) {
	c1 := GrayColor{V: 0, A: 255}
	c2 := GrayColor{V: 255, A: 255}
	c3 := GrayColor{V: 128, A: 255}

	sg := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	sg.Prepare()

	tests := []struct {
		name   string
		x, y   int
		length uint
	}{
		{"Single pixel", 50, 25, 1},
		{"Small span", 40, 30, 5},
		{"Long span", 0, 10, 100},
		{"Zero length", 50, 25, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.length == 0 {
				return // Skip zero length test
			}

			span := make([]GrayColor, tt.length)

			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Generate panicked: %v", r)
				}
			}()

			sg.Generate(span, tt.x, tt.y, tt.length)
		})
	}
}

func TestSpanGouraudGrayOverflowHandling(t *testing.T) {
	// Create colors that might cause overflow during interpolation
	c1 := GrayColor{V: 0, A: 0}
	c2 := GrayColor{V: 300, A: 300} // Over 255
	c3 := GrayColor{V: -50, A: -50} // Negative

	sg := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	sg.Prepare()

	span := make([]GrayColor, 10)
	sg.Generate(span, 25, 25, 10)

	// Verify clamping works
	for i, color := range span {
		if color.V < 0 || color.V > 255 || color.A < 0 || color.A > 255 {
			t.Errorf("Overflow not handled correctly at span[%d]: V=%d A=%d", i, color.V, color.A)
		}
	}
}

func TestSpanGouraudGrayTriangleOrientation(t *testing.T) {
	c1 := GrayColor{V: 255, A: 255}
	c2 := GrayColor{V: 128, A: 255}
	c3 := GrayColor{V: 0, A: 255}

	// Test clockwise triangle when arranged: (0,0) bottom, (100,50) middle, (50,100) top
	sgCW := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 50, 50, 100, 0)
	sgCW.Prepare()

	// Test counter-clockwise triangle when arranged: (0,0) bottom, (0,50) middle, (50,100) top
	sgCCW := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 0, 50, 50, 100, 0)
	sgCCW.Prepare()

	// The swap flags should be different
	if sgCW.swap == sgCCW.swap {
		t.Errorf("Triangle orientation should affect swap flag: CW_swap=%t CCW_swap=%t", sgCW.swap, sgCCW.swap)
	}
}

func TestSpanGouraudGrayVersusRGBA(t *testing.T) {
	// Compare grayscale vs RGBA with same values for consistency
	grayC1 := GrayColor{V: 100, A: 255}
	grayC2 := GrayColor{V: 200, A: 255}
	grayC3 := GrayColor{V: 50, A: 255}

	rgbaC1 := RGBAColor{R: 100, G: 100, B: 100, A: 255}
	rgbaC2 := RGBAColor{R: 200, G: 200, B: 200, A: 255}
	rgbaC3 := RGBAColor{R: 50, G: 50, B: 50, A: 255}

	sgGray := NewSpanGouraudGrayWithTriangle(grayC1, grayC2, grayC3, 0, 0, 100, 0, 50, 50, 0)
	sgRGBA := NewSpanGouraudRGBAWithTriangle(rgbaC1, rgbaC2, rgbaC3, 0, 0, 100, 0, 50, 50, 0)

	sgGray.Prepare()
	sgRGBA.Prepare()

	spanGray := make([]GrayColor, 10)
	spanRGBA := make([]RGBAColor, 10)

	sgGray.Generate(spanGray, 25, 25, 10)
	sgRGBA.Generate(spanRGBA, 25, 25, 10)

	// Compare results - they should be similar (allowing for rounding differences)
	for i := 0; i < len(spanGray); i++ {
		tolerance := 5

		if absGray(spanGray[i].V-spanRGBA[i].R) > tolerance ||
			absGray(spanGray[i].V-spanRGBA[i].G) > tolerance ||
			absGray(spanGray[i].V-spanRGBA[i].B) > tolerance ||
			absGray(spanGray[i].A-spanRGBA[i].A) > tolerance {
			t.Errorf("Gray vs RGBA mismatch at index %d: Gray(V=%d,A=%d) vs RGBA(R=%d,G=%d,B=%d,A=%d)",
				i, spanGray[i].V, spanGray[i].A, spanRGBA[i].R, spanRGBA[i].G, spanRGBA[i].B, spanRGBA[i].A)
		}
	}
}

func absGray(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkSpanGouraudGrayPrepare(b *testing.B) {
	c1 := GrayColor{V: 255, A: 255}
	c2 := GrayColor{V: 128, A: 255}
	c3 := GrayColor{V: 0, A: 255}

	sg := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.Prepare()
	}
}

func BenchmarkSpanGouraudGrayGenerate(b *testing.B) {
	c1 := GrayColor{V: 255, A: 255}
	c2 := GrayColor{V: 128, A: 255}
	c3 := GrayColor{V: 0, A: 255}

	sg := NewSpanGouraudGrayWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	sg.Prepare()

	span := make([]GrayColor, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.Generate(span, 0, 25, 100)
	}
}

func BenchmarkSpanGouraudGrayVsRGBA(b *testing.B) {
	grayC1 := GrayColor{V: 255, A: 255}
	grayC2 := GrayColor{V: 128, A: 255}
	grayC3 := GrayColor{V: 0, A: 255}

	rgbaC1 := RGBAColor{R: 255, G: 255, B: 255, A: 255}
	rgbaC2 := RGBAColor{R: 128, G: 128, B: 128, A: 255}
	rgbaC3 := RGBAColor{R: 0, G: 0, B: 0, A: 255}

	sgGray := NewSpanGouraudGrayWithTriangle(grayC1, grayC2, grayC3, 0, 0, 100, 0, 50, 100, 0)
	sgRGBA := NewSpanGouraudRGBAWithTriangle(rgbaC1, rgbaC2, rgbaC3, 0, 0, 100, 0, 50, 100, 0)

	sgGray.Prepare()
	sgRGBA.Prepare()

	spanGray := make([]GrayColor, 100)
	spanRGBA := make([]RGBAColor, 100)

	b.Run("Gray", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sgGray.Generate(spanGray, 0, 25, 100)
		}
	})

	b.Run("RGBA", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sgRGBA.Generate(spanRGBA, 0, 25, 100)
		}
	})
}
