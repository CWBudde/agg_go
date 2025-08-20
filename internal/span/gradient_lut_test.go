package span

import (
	"fmt"
	"testing"

	"agg_go/internal/color"
)

func TestDDALineInterpolator(t *testing.T) {
	// Test basic interpolation from 0 to 255 over 256 steps
	dda := NewDDALineInterpolator(0, 255, 256, 14)

	// Check initial value
	if dda.Y() != 0 {
		t.Errorf("Expected initial value 0, got %d", dda.Y())
	}

	// Check interpolation at 25% (should be ~63)
	for i := 0; i < 64; i++ {
		dda.Inc()
	}
	val := dda.Y()
	if val < 60 || val > 67 {
		t.Errorf("Expected value around 63 at 25%%, got %d", val)
	}

	// Check interpolation at 50% (should be ~127)
	for i := 0; i < 64; i++ {
		dda.Inc()
	}
	val = dda.Y()
	if val < 124 || val > 131 {
		t.Errorf("Expected value around 127 at 50%%, got %d", val)
	}

	// Test Dec operation
	dda.Dec()
	val2 := dda.Y()
	if val2 >= val {
		t.Errorf("Dec should decrease value: %d >= %d", val2, val)
	}
}

func TestColorInterpolatorRGBA8(t *testing.T) {
	// Test interpolation from black to white
	c1 := color.NewRGBA8[color.Linear](0, 0, 0, 255)
	c2 := color.NewRGBA8[color.Linear](255, 255, 255, 255)

	ci := NewColorInterpolatorRGBA8(c1, c2, 256)

	// Check initial color (black)
	initial := ci.Color()
	if initial.R != 0 || initial.G != 0 || initial.B != 0 || initial.A != 255 {
		t.Errorf("Expected black (0,0,0,255), got (%d,%d,%d,%d)",
			initial.R, initial.G, initial.B, initial.A)
	}

	// Advance to middle (should be gray ~127)
	for i := 0; i < 128; i++ {
		ci.Inc()
	}
	mid := ci.Color()

	// Allow some tolerance for rounding
	if mid.R < 120 || mid.R > 135 ||
		mid.G < 120 || mid.G > 135 ||
		mid.B < 120 || mid.B > 135 {
		t.Errorf("Expected gray around (127,127,127), got (%d,%d,%d,%d)",
			mid.R, mid.G, mid.B, mid.A)
	}
}

func TestColorInterpolatorGray8(t *testing.T) {
	// Test interpolation from black to white
	c1 := color.NewGray8WithAlpha[color.Linear](0, 255)
	c2 := color.NewGray8WithAlpha[color.Linear](255, 255)

	ci := NewColorInterpolatorGray8(c1, c2, 256)

	// Check initial color (black)
	initial := ci.Color()
	if initial.V != 0 || initial.A != 255 {
		t.Errorf("Expected black (0,255), got (%d,%d)", initial.V, initial.A)
	}

	// Advance to middle (should be gray ~127)
	for i := 0; i < 128; i++ {
		ci.Inc()
	}
	mid := ci.Color()

	// Allow some tolerance for rounding
	if mid.V < 120 || mid.V > 135 {
		t.Errorf("Expected gray around 127, got %d", mid.V)
	}
}

func TestColorPoint(t *testing.T) {
	// Test normal color point
	cp := NewColorPoint(0.5, color.NewRGBA8[color.Linear](255, 0, 0, 255))
	if cp.Offset != 0.5 {
		t.Errorf("Expected offset 0.5, got %f", cp.Offset)
	}

	// Test clamping below 0
	cp = NewColorPoint(-0.1, color.NewRGBA8[color.Linear](255, 0, 0, 255))
	if cp.Offset != 0.0 {
		t.Errorf("Expected offset clamped to 0.0, got %f", cp.Offset)
	}

	// Test clamping above 1
	cp = NewColorPoint(1.5, color.NewRGBA8[color.Linear](255, 0, 0, 255))
	if cp.Offset != 1.0 {
		t.Errorf("Expected offset clamped to 1.0, got %f", cp.Offset)
	}
}

func TestGradientLUTBasic(t *testing.T) {
	// Test basic two-color gradient (black to white)
	lut := NewGradientLUT[color.RGBA8[color.Linear], *ColorInterpolatorRGBA8[color.Linear]](256)

	black := color.NewRGBA8[color.Linear](0, 0, 0, 255)
	white := color.NewRGBA8[color.Linear](255, 255, 255, 255)

	lut.AddColor(0.0, black)
	lut.AddColor(1.0, white)

	// Build the LUT
	lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
		return NewColorInterpolatorRGBA8(c1, c2, length)
	})

	// Check size
	if lut.Size() != 256 {
		t.Errorf("Expected size 256, got %d", lut.Size())
	}

	// Check first color (black)
	first := lut.At(0)
	if first.R != 0 || first.G != 0 || first.B != 0 {
		t.Errorf("Expected first color to be black, got (%d,%d,%d,%d)",
			first.R, first.G, first.B, first.A)
	}

	// Check last color (white)
	last := lut.At(255)
	if last.R != 255 || last.G != 255 || last.B != 255 {
		t.Errorf("Expected last color to be white, got (%d,%d,%d,%d)",
			last.R, last.G, last.B, last.A)
	}

	// Check middle color (should be gray)
	mid := lut.At(128)
	if mid.R < 120 || mid.R > 135 || mid.G < 120 || mid.G > 135 || mid.B < 120 || mid.B > 135 {
		t.Errorf("Expected middle color to be gray, got (%d,%d,%d,%d)",
			mid.R, mid.G, mid.B, mid.A)
	}

	// Test bounds checking
	outOfBounds := lut.At(-1)
	if outOfBounds != first {
		t.Errorf("Expected At(-1) to return first color")
	}

	outOfBounds = lut.At(1000)
	if outOfBounds != last {
		t.Errorf("Expected At(1000) to return last color")
	}
}

func TestGradientLUTMultiStop(t *testing.T) {
	// Test multi-stop gradient: red -> green -> blue
	lut := NewGradientLUT[color.RGBA8[color.Linear], *ColorInterpolatorRGBA8[color.Linear]](256)

	red := color.NewRGBA8[color.Linear](255, 0, 0, 255)
	green := color.NewRGBA8[color.Linear](0, 255, 0, 255)
	blue := color.NewRGBA8[color.Linear](0, 0, 255, 255)

	lut.AddColor(0.0, red)
	lut.AddColor(0.5, green)
	lut.AddColor(1.0, blue)

	// Build the LUT
	lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
		return NewColorInterpolatorRGBA8(c1, c2, length)
	})

	// Check first color (red)
	first := lut.At(0)
	if first.R != 255 || first.G != 0 || first.B != 0 {
		t.Errorf("Expected first color to be red, got (%d,%d,%d,%d)",
			first.R, first.G, first.B, first.A)
	}

	// Check middle color (should be green)
	mid := lut.At(128)
	if mid.G < 200 || mid.R > 55 || mid.B > 55 {
		t.Errorf("Expected middle color to be mostly green, got (%d,%d,%d,%d)",
			mid.R, mid.G, mid.B, mid.A)
	}

	// Check last color (blue)
	last := lut.At(255)
	if last.B != 255 || last.R != 0 || last.G != 0 {
		t.Errorf("Expected last color to be blue, got (%d,%d,%d,%d)",
			last.R, last.G, last.B, last.A)
	}
}

func TestGradientLUTGrayscale(t *testing.T) {
	// Test grayscale gradient
	lut := NewGradientLUT[color.Gray8[color.Linear], *ColorInterpolatorGray8[color.Linear]](256)

	black := color.NewGray8WithAlpha[color.Linear](0, 255)
	white := color.NewGray8WithAlpha[color.Linear](255, 255)

	lut.AddColor(0.0, black)
	lut.AddColor(1.0, white)

	// Build the LUT
	lut.BuildLUT(func(c1, c2 color.Gray8[color.Linear], length uint) *ColorInterpolatorGray8[color.Linear] {
		return NewColorInterpolatorGray8(c1, c2, length)
	})

	// Check first color (black)
	first := lut.At(0)
	if first.V != 0 {
		t.Errorf("Expected first value to be 0, got %d", first.V)
	}

	// Check last color (white)
	last := lut.At(255)
	if last.V != 255 {
		t.Errorf("Expected last value to be 255, got %d", last.V)
	}

	// Check middle color (should be gray)
	mid := lut.At(128)
	if mid.V < 120 || mid.V > 135 {
		t.Errorf("Expected middle value to be around 127, got %d", mid.V)
	}
}

func TestGradientLUTEdgeCases(t *testing.T) {
	lut := NewGradientLUT[color.RGBA8[color.Linear], *ColorInterpolatorRGBA8[color.Linear]](256)

	// Test empty gradient
	lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
		return NewColorInterpolatorRGBA8(c1, c2, length)
	})
	// Should not crash

	// Test single color
	lut.RemoveAll()
	red := color.NewRGBA8[color.Linear](255, 0, 0, 255)
	lut.AddColor(0.5, red)
	lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
		return NewColorInterpolatorRGBA8(c1, c2, length)
	})
	// Should not crash

	// Test duplicate offsets
	lut.RemoveAll()
	red1 := color.NewRGBA8[color.Linear](255, 0, 0, 255)
	red2 := color.NewRGBA8[color.Linear](128, 0, 0, 255)
	blue := color.NewRGBA8[color.Linear](0, 0, 255, 255)

	lut.AddColor(0.0, red1)
	lut.AddColor(0.0, red2) // duplicate offset
	lut.AddColor(1.0, blue)

	lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
		return NewColorInterpolatorRGBA8(c1, c2, length)
	})

	// Should handle duplicates gracefully
	first := lut.At(0)
	if first.B == 255 {
		t.Errorf("Expected duplicate handling to keep one red color, not blue")
	}
}

func TestGradientLUTDifferentSizes(t *testing.T) {
	sizes := []int{64, 128, 256, 512, 1024}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
			lut := NewGradientLUT[color.RGBA8[color.Linear], *ColorInterpolatorRGBA8[color.Linear]](size)

			black := color.NewRGBA8[color.Linear](0, 0, 0, 255)
			white := color.NewRGBA8[color.Linear](255, 255, 255, 255)

			lut.AddColor(0.0, black)
			lut.AddColor(1.0, white)

			lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
				return NewColorInterpolatorRGBA8(c1, c2, length)
			})

			if lut.Size() != size {
				t.Errorf("Expected size %d, got %d", size, lut.Size())
			}

			// Check gradient consistency
			first := lut.At(0)
			last := lut.At(size - 1)

			if first.R != 0 || first.G != 0 || first.B != 0 {
				t.Errorf("First color should be black")
			}

			if last.R != 255 || last.G != 255 || last.B != 255 {
				t.Errorf("Last color should be white")
			}
		})
	}
}

// Benchmark tests
func BenchmarkColorInterpolatorRGBA8(b *testing.B) {
	c1 := color.NewRGBA8[color.Linear](0, 0, 0, 255)
	c2 := color.NewRGBA8[color.Linear](255, 255, 255, 255)
	ci := NewColorInterpolatorRGBA8(c1, c2, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ci.Inc()
		_ = ci.Color()
	}
}

func BenchmarkColorInterpolatorGray8(b *testing.B) {
	c1 := color.NewGray8WithAlpha[color.Linear](0, 255)
	c2 := color.NewGray8WithAlpha[color.Linear](255, 255)
	ci := NewColorInterpolatorGray8(c1, c2, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ci.Inc()
		_ = ci.Color()
	}
}

func BenchmarkGradientLUTBuild(b *testing.B) {
	black := color.NewRGBA8[color.Linear](0, 0, 0, 255)
	white := color.NewRGBA8[color.Linear](255, 255, 255, 255)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lut := NewGradientLUT[color.RGBA8[color.Linear], *ColorInterpolatorRGBA8[color.Linear]](256)
		lut.AddColor(0.0, black)
		lut.AddColor(1.0, white)
		lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
			return NewColorInterpolatorRGBA8(c1, c2, length)
		})
	}
}

func BenchmarkGradientLUTAccess(b *testing.B) {
	lut := NewGradientLUT[color.RGBA8[color.Linear], *ColorInterpolatorRGBA8[color.Linear]](256)

	black := color.NewRGBA8[color.Linear](0, 0, 0, 255)
	white := color.NewRGBA8[color.Linear](255, 255, 255, 255)

	lut.AddColor(0.0, black)
	lut.AddColor(1.0, white)
	lut.BuildLUT(func(c1, c2 color.RGBA8[color.Linear], length uint) *ColorInterpolatorRGBA8[color.Linear] {
		return NewColorInterpolatorRGBA8(c1, c2, length)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lut.At(i % lut.Size())
	}
}
