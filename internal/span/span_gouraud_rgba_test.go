package span

import (
	"testing"
)

func TestRGBAColorCreation(t *testing.T) {
	c := RGBAColor{R: 255, G: 128, B: 64, A: 200}
	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 200 {
		t.Errorf("RGBAColor not created correctly: got R=%d G=%d B=%d A=%d", c.R, c.G, c.B, c.A)
	}
}

func TestSpanGouraudRGBACreation(t *testing.T) {
	sg := NewSpanGouraudRGBA()
	if sg == nil {
		t.Fatal("NewSpanGouraudRGBA returned nil")
	}
}

func TestSpanGouraudRGBAWithTriangle(t *testing.T) {
	c1 := RGBAColor{R: 255, G: 0, B: 0, A: 255} // Red
	c2 := RGBAColor{R: 0, G: 255, B: 0, A: 255} // Green
	c3 := RGBAColor{R: 0, G: 0, B: 255, A: 255} // Blue

	sg := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	if sg == nil {
		t.Fatal("NewSpanGouraudRGBAWithTriangle returned nil")
	}

	coord := sg.Coord()
	if coord[0].Color.R != 255 || coord[0].Color.G != 0 {
		t.Errorf("First vertex color not set correctly")
	}
	if coord[1].Color.G != 255 || coord[1].Color.R != 0 {
		t.Errorf("Second vertex color not set correctly")
	}
	if coord[2].Color.B != 255 || coord[2].Color.R != 0 {
		t.Errorf("Third vertex color not set correctly")
	}
}

func TestRGBACalcInit(t *testing.T) {
	c1 := CoordType[RGBAColor]{
		X: 0, Y: 0,
		Color: RGBAColor{R: 100, G: 150, B: 200, A: 255},
	}
	c2 := CoordType[RGBAColor]{
		X: 100, Y: 100,
		Color: RGBAColor{R: 200, G: 50, B: 100, A: 128},
	}

	var calc RGBACalc
	calc.init(c1, c2)

	// Check initial values
	if calc.r1 != 100 || calc.g1 != 150 || calc.b1 != 200 || calc.a1 != 255 {
		t.Errorf("Initial RGBA values not set correctly: R=%d G=%d B=%d A=%d",
			calc.r1, calc.g1, calc.b1, calc.a1)
	}

	// Check deltas
	if calc.dr != 100 || calc.dg != -100 || calc.db != -100 || calc.da != -127 {
		t.Errorf("RGBA deltas not calculated correctly: dR=%d dG=%d dB=%d dA=%d",
			calc.dr, calc.dg, calc.db, calc.da)
	}
}

func TestRGBACalcCalc(t *testing.T) {
	c1 := CoordType[RGBAColor]{
		X: 0, Y: 0,
		Color: RGBAColor{R: 0, G: 0, B: 0, A: 0},
	}
	c2 := CoordType[RGBAColor]{
		X: 100, Y: 100,
		Color: RGBAColor{R: 255, G: 255, B: 255, A: 255},
	}

	var calc RGBACalc
	calc.init(c1, c2)

	// Test interpolation at middle
	calc.calc(50.0)

	// Should be approximately halfway between colors
	tolerance := 5 // Allow some rounding error
	if absRGBA(calc.r-127) > tolerance || absRGBA(calc.g-127) > tolerance ||
		absRGBA(calc.b-127) > tolerance || absRGBA(calc.a-127) > tolerance {
		t.Errorf("Midpoint interpolation incorrect: R=%d G=%d B=%d A=%d",
			calc.r, calc.g, calc.b, calc.a)
	}
}

func TestSpanGouraudRGBAPrepare(t *testing.T) {
	c1 := RGBAColor{R: 255, G: 0, B: 0, A: 255}
	c2 := RGBAColor{R: 0, G: 255, B: 0, A: 255}
	c3 := RGBAColor{R: 0, G: 0, B: 255, A: 255}

	sg := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 25, 50, 100, 0)
	sg.Prepare()

	// After prepare, y2 should be set to the middle vertex Y coordinate
	if sg.y2 != 25 {
		t.Errorf("y2 should be 25 after Prepare(), got %d", sg.y2)
	}

	// swap flag should be determined
	_ = sg.swap // Just verify it's accessible
}

func TestSpanGouraudRGBAGenerate(t *testing.T) {
	// Create a simple triangle with gradient from red to green to blue
	c1 := RGBAColor{R: 255, G: 0, B: 0, A: 255} // Red at (0,0)
	c2 := RGBAColor{R: 0, G: 255, B: 0, A: 255} // Green at (100,0)
	c3 := RGBAColor{R: 0, G: 0, B: 255, A: 255} // Blue at (50,50)

	sg := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 50, 0)
	sg.Prepare()

	// Generate a span at y=25 (middle height)
	span := make([]RGBAColor, 10)
	sg.Generate(span, 25, 25, 10)

	// Verify that colors are interpolated (not all the same)
	allSame := true
	first := span[0]
	for i := 1; i < len(span); i++ {
		if span[i].R != first.R || span[i].G != first.G ||
			span[i].B != first.B || span[i].A != first.A {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Generated span should have color variation")
	}

	// Verify colors are in valid range
	for i, color := range span {
		if color.R < 0 || color.R > 255 || color.G < 0 || color.G > 255 ||
			color.B < 0 || color.B > 255 || color.A < 0 || color.A > 255 {
			t.Errorf("Span[%d] has invalid color values: R=%d G=%d B=%d A=%d",
				i, color.R, color.G, color.B, color.A)
		}
	}
}

func TestSpanGouraudRGBAGenerateEdgeCases(t *testing.T) {
	c1 := RGBAColor{R: 0, G: 0, B: 0, A: 255}
	c2 := RGBAColor{R: 255, G: 255, B: 255, A: 255}
	c3 := RGBAColor{R: 128, G: 128, B: 128, A: 255}

	sg := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
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

			span := make([]RGBAColor, tt.length)

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

func TestSpanGouraudRGBAOverflowHandling(t *testing.T) {
	// Create colors that might cause overflow during interpolation
	c1 := RGBAColor{R: 0, G: 0, B: 0, A: 0}
	c2 := RGBAColor{R: 300, G: 300, B: 300, A: 300} // Over 255
	c3 := RGBAColor{R: -50, G: -50, B: -50, A: -50} // Negative

	sg := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	sg.Prepare()

	span := make([]RGBAColor, 10)
	sg.Generate(span, 25, 25, 10)

	// Verify clamping works
	for i, color := range span {
		if color.R < 0 || color.R > 255 || color.G < 0 || color.G > 255 ||
			color.B < 0 || color.B > 255 || color.A < 0 || color.A > 255 {
			t.Errorf("Overflow not handled correctly at span[%d]: R=%d G=%d B=%d A=%d",
				i, color.R, color.G, color.B, color.A)
		}
	}
}

func TestSpanGouraudRGBATriangleOrientation(t *testing.T) {
	c1 := RGBAColor{R: 255, G: 0, B: 0, A: 255}
	c2 := RGBAColor{R: 0, G: 255, B: 0, A: 255}
	c3 := RGBAColor{R: 0, G: 0, B: 255, A: 255}

	// Test clockwise triangle when arranged: (0,0) bottom, (100,50) middle, (50,100) top
	// This should have clockwise orientation after arrangement
	sgCW := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 50, 50, 100, 0)
	sgCW.Prepare()

	// Test counter-clockwise triangle when arranged: (0,0) bottom, (0,50) middle, (50,100) top
	// This should have counter-clockwise orientation after arrangement
	sgCCW := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 0, 50, 50, 100, 0)
	sgCCW.Prepare()

	// The swap flags should be different
	if sgCW.swap == sgCCW.swap {
		t.Errorf("Triangle orientation should affect swap flag: CW_swap=%t CCW_swap=%t", sgCW.swap, sgCCW.swap)
	}
}

func absRGBA(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkSpanGouraudRGBAPrepare(b *testing.B) {
	c1 := RGBAColor{R: 255, G: 0, B: 0, A: 255}
	c2 := RGBAColor{R: 0, G: 255, B: 0, A: 255}
	c3 := RGBAColor{R: 0, G: 0, B: 255, A: 255}

	sg := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.Prepare()
	}
}

func BenchmarkSpanGouraudRGBAGenerate(b *testing.B) {
	c1 := RGBAColor{R: 255, G: 0, B: 0, A: 255}
	c2 := RGBAColor{R: 0, G: 255, B: 0, A: 255}
	c3 := RGBAColor{R: 0, G: 0, B: 255, A: 255}

	sg := NewSpanGouraudRGBAWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	sg.Prepare()

	span := make([]RGBAColor, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.Generate(span, 0, 25, 100)
	}
}
