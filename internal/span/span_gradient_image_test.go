package span

import (
	"testing"

	"agg_go/internal/color"
)

func TestOneColorFunction(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		ocf := NewOneColorFunction[color.RGBA8[color.SRGB]]()

		// Should always return size 1
		if ocf.Size() != 1 {
			t.Errorf("Size(): got %d, want 1", ocf.Size())
		}

		// Initial color should be zero value
		initialColor := ocf.ColorAt(0)
		if initialColor.R != 0 || initialColor.G != 0 || initialColor.B != 0 || initialColor.A != 0 {
			t.Errorf("Initial color not zero: got %+v", initialColor)
		}

		// Modify color and verify
		colorPtr := ocf.Color()
		*colorPtr = color.RGBA8[color.SRGB]{R: 255, G: 128, B: 64, A: 200}

		modified := ocf.ColorAt(0)
		if modified.R != 255 || modified.G != 128 || modified.B != 64 || modified.A != 200 {
			t.Errorf("Modified color incorrect: got %+v, want {255 128 64 200}", modified)
		}

		// ColorAt should ignore index
		sameColor := ocf.ColorAt(999)
		if sameColor != modified {
			t.Errorf("ColorAt(999) should equal ColorAt(0)")
		}
	})
}

func TestGradientImage_ImageCreate(t *testing.T) {
	t.Run("Basic creation", func(t *testing.T) {
		gi := NewGradientImageRGBA8()

		// Create a 4x3 image
		buffer := gi.ImageCreate(4, 3)
		if buffer == nil {
			t.Fatal("ImageCreate returned nil buffer")
		}

		// Verify dimensions
		if gi.ImageWidth() != 4 {
			t.Errorf("ImageWidth(): got %d, want 4", gi.ImageWidth())
		}
		if gi.ImageHeight() != 3 {
			t.Errorf("ImageHeight(): got %d, want 3", gi.ImageHeight())
		}
		if gi.ImageStride() != 16 { // 4 pixels * 4 bytes
			t.Errorf("ImageStride(): got %d, want 16", gi.ImageStride())
		}

		// Buffer should be cleared to zero
		if len(buffer) != 12 { // 4*3 = 12 pixels
			t.Errorf("Buffer length: got %d, want 12", len(buffer))
		}

		for i, pixel := range buffer {
			if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 || pixel.A != 0 {
				t.Errorf("Pixel %d not cleared: %+v", i, pixel)
			}
		}
	})

	t.Run("Reallocation behavior", func(t *testing.T) {
		gi := NewGradientImageRGBA8()

		// Create small image first
		buffer1 := gi.ImageCreate(2, 2)
		if len(buffer1) != 4 {
			t.Errorf("First allocation: got %d pixels, want 4", len(buffer1))
		}

		// Create larger image - should reallocate
		buffer2 := gi.ImageCreate(5, 4)
		if len(buffer2) != 20 {
			t.Errorf("Second allocation: got %d pixels, want 20", len(buffer2))
		}

		// Create smaller image - should reuse buffer
		gi.ImageCreate(3, 2)
		if gi.ImageWidth() != 3 || gi.ImageHeight() != 2 {
			t.Errorf("Third allocation dimensions: got %dx%d, want 3x2", gi.ImageWidth(), gi.ImageHeight())
		}
		// Allocated size should still be 5x4
		if gi.allocWidth != 5 || gi.allocHeight != 4 {
			t.Errorf("Allocated size changed: got %dx%d, want 5x4", gi.allocWidth, gi.allocHeight)
		}
	})
}

func TestGradientImage_Calculate(t *testing.T) {
	t.Run("No buffer", func(t *testing.T) {
		gi := NewGradientImageRGBA8()
		// Should handle no buffer gracefully
		result := gi.Calculate(0, 0, 100)
		if result != 0 {
			t.Errorf("Calculate with no buffer: got %d, want 0", result)
		}

		// Color should be set to transparent black
		color := gi.ColorFunction().ColorAt(0)
		if color.R != 0 || color.G != 0 || color.B != 0 || color.A != 0 {
			t.Errorf("Color not cleared: %+v", color)
		}
	})

	t.Run("Basic sampling", func(t *testing.T) {
		gi := NewGradientImageRGBA8()
		// Create 3x2 image
		buffer := gi.ImageCreate(3, 2)
		if buffer == nil {
			t.Fatal("Failed to create image")
		}

		// Set up test pattern
		// Row 0: [Red, Green, Blue]
		// Row 1: [Yellow, Cyan, Magenta]
		buffer[0] = color.RGBA8[color.SRGB]{R: 255, G: 0, B: 0, A: 255}   // Red
		buffer[1] = color.RGBA8[color.SRGB]{R: 0, G: 255, B: 0, A: 255}   // Green
		buffer[2] = color.RGBA8[color.SRGB]{R: 0, G: 0, B: 255, A: 255}   // Blue
		buffer[3] = color.RGBA8[color.SRGB]{R: 255, G: 255, B: 0, A: 255} // Yellow
		buffer[4] = color.RGBA8[color.SRGB]{R: 0, G: 255, B: 255, A: 255} // Cyan
		buffer[5] = color.RGBA8[color.SRGB]{R: 255, G: 0, B: 255, A: 255} // Magenta

		// Test sampling at pixel coordinates (need to convert to subpixel)
		testCases := []struct {
			x, y     int // pixel coordinates
			expected color.RGBA8[color.SRGB]
			name     string
		}{
			{0, 0, buffer[0], "Red at (0,0)"},
			{1, 0, buffer[1], "Green at (1,0)"},
			{2, 0, buffer[2], "Blue at (2,0)"},
			{0, 1, buffer[3], "Yellow at (0,1)"},
			{1, 1, buffer[4], "Cyan at (1,1)"},
			{2, 1, buffer[5], "Magenta at (2,1)"},
		}

		for _, tc := range testCases {
			// Convert to subpixel coordinates
			subX := tc.x << GradientSubpixelShift
			subY := tc.y << GradientSubpixelShift

			gi.Calculate(subX, subY, 100)
			sampledColor := gi.ColorFunction().ColorAt(0)

			if sampledColor != tc.expected {
				t.Errorf("%s: got %+v, want %+v", tc.name, sampledColor, tc.expected)
			}
		}
	})

	t.Run("Coordinate wrapping", func(t *testing.T) {
		gi := NewGradientImageRGBA8()
		// Create 2x2 image
		buffer := gi.ImageCreate(2, 2)
		if buffer == nil {
			t.Fatal("Failed to create image")
		}

		// Set up simple pattern
		buffer[0] = color.RGBA8[color.SRGB]{R: 100, G: 0, B: 0, A: 255} // (0,0)
		buffer[1] = color.RGBA8[color.SRGB]{R: 200, G: 0, B: 0, A: 255} // (1,0)
		buffer[2] = color.RGBA8[color.SRGB]{R: 0, G: 100, B: 0, A: 255} // (0,1)
		buffer[3] = color.RGBA8[color.SRGB]{R: 0, G: 200, B: 0, A: 255} // (1,1)

		// Test wrapping behavior
		testCases := []struct {
			x, y     int // pixel coordinates (can be negative or >= dimensions)
			expected color.RGBA8[color.SRGB]
			name     string
		}{
			// Positive wrapping
			{2, 0, buffer[0], "x=2 wraps to x=0"}, // 2 % 2 = 0, y=0 -> buffer[0*2+0] = buffer[0]
			{0, 2, buffer[0], "y=2 wraps to y=0"}, // x=0, 2 % 2 = 0 -> buffer[0*2+0] = buffer[0]
			{3, 1, buffer[3], "x=3 wraps to x=1"}, // 3 % 2 = 1, y=1 -> buffer[1*2+1] = buffer[3]
			{1, 3, buffer[3], "y=3 wraps to y=1"}, // x=1, 3 % 2 = 1 -> buffer[1*2+1] = buffer[3]

			// Negative wrapping
			{-1, 0, buffer[1], "x=-1 wraps to x=1"},         // -1 % 2 + 2 = 1, y=0 -> buffer[0*2+1] = buffer[1]
			{0, -1, buffer[2], "y=-1 wraps to y=1"},         // x=0, -1 % 2 + 2 = 1 -> buffer[1*2+0] = buffer[2]
			{-2, -2, buffer[0], "x=-2,y=-2 wraps to (0,0)"}, // both wrap to 0 -> buffer[0*2+0] = buffer[0]
		}

		for _, tc := range testCases {
			// Convert to subpixel coordinates
			subX := tc.x << GradientSubpixelShift
			subY := tc.y << GradientSubpixelShift

			gi.Calculate(subX, subY, 100)
			sampledColor := gi.ColorFunction().ColorAt(0)

			if sampledColor != tc.expected {
				// Debug info for failing tests
				px := tc.x % gi.ImageWidth()
				if px < 0 {
					px += gi.ImageWidth()
				}
				py := tc.y % gi.ImageHeight()
				if py < 0 {
					py += gi.ImageHeight()
				}
				expectedIndex := py*gi.AllocWidth() + px
				t.Errorf("%s: got %+v, want %+v (coords %d,%d -> %d,%d, index %d, buffer len %d)",
					tc.name, sampledColor, tc.expected, tc.x, tc.y, px, py, expectedIndex, len(buffer))
			}
		}
	})

	t.Run("Subpixel precision handling", func(t *testing.T) {
		gi := NewGradientImageRGBA8()
		// Create 2x2 image
		buffer := gi.ImageCreate(2, 2)
		if buffer == nil {
			t.Fatal("Failed to create image")
		}

		// Set distinct colors
		buffer[0] = color.RGBA8[color.SRGB]{R: 255, G: 0, B: 0, A: 255} // (0,0) Red
		buffer[1] = color.RGBA8[color.SRGB]{R: 0, G: 255, B: 0, A: 255} // (1,0) Green

		// Test that subpixel coordinates in the same pixel return same color
		baseX := 1 << GradientSubpixelShift // pixel 1 in subpixel coords
		baseY := 0 << GradientSubpixelShift // pixel 0 in subpixel coords

		// All these should map to pixel (1,0)
		subpixelOffsets := []int{0, 1, 7, 15} // within same pixel
		for _, offset := range subpixelOffsets {
			gi.Calculate(baseX+offset, baseY, 100)
			sampledColor := gi.ColorFunction().ColorAt(0)

			expected := buffer[1] // Green at (1,0)
			if sampledColor != expected {
				t.Errorf("Subpixel offset %d: got %+v, want %+v", offset, sampledColor, expected)
			}
		}
	})
}

func TestGradientImage_Integration(t *testing.T) {
	t.Run("ColorFunction integration", func(t *testing.T) {
		gi := NewGradientImageRGBA8()
		colorFunc := gi.ColorFunction()

		if colorFunc == nil {
			t.Fatal("ColorFunction returned nil")
		}

		if colorFunc.Size() != 1 {
			t.Errorf("ColorFunction size: got %d, want 1", colorFunc.Size())
		}

		// Create image and sample a color
		buffer := gi.ImageCreate(1, 1)
		buffer[0] = color.RGBA8[color.SRGB]{R: 123, G: 45, B: 67, A: 89}

		gi.Calculate(0, 0, 100)

		// Verify color function returns the sampled color
		sampledColor := colorFunc.ColorAt(0)
		expected := color.RGBA8[color.SRGB]{R: 123, G: 45, B: 67, A: 89}
		if sampledColor != expected {
			t.Errorf("Integration test: got %+v, want %+v", sampledColor, expected)
		}
	})

	t.Run("Gradient function interface compliance", func(t *testing.T) {
		gi := NewGradientImageRGBA8()

		// Create a simple 2x2 gradient pattern
		buffer := gi.ImageCreate(2, 2)
		buffer[0] = color.RGBA8[color.SRGB]{R: 255, G: 0, B: 0, A: 255}     // Red
		buffer[1] = color.RGBA8[color.SRGB]{R: 0, G: 255, B: 0, A: 255}     // Green
		buffer[2] = color.RGBA8[color.SRGB]{R: 0, G: 0, B: 255, A: 255}     // Blue
		buffer[3] = color.RGBA8[color.SRGB]{R: 255, G: 255, B: 255, A: 255} // White

		// Test that it works as a GradientFunction interface
		var gf GradientFunction = gi

		// Sample colors at different positions
		coords := []struct{ x, y int }{
			{0, 0}, {1, 0}, {0, 1}, {1, 1},
		}

		expectedColors := []color.RGBA8[color.SRGB]{
			{R: 255, G: 0, B: 0, A: 255},     // Red
			{R: 0, G: 255, B: 0, A: 255},     // Green
			{R: 0, G: 0, B: 255, A: 255},     // Blue
			{R: 255, G: 255, B: 255, A: 255}, // White
		}

		for i, coord := range coords {
			// Convert to subpixel coordinates
			subX := coord.x << GradientSubpixelShift
			subY := coord.y << GradientSubpixelShift

			// Call through interface
			result := gf.Calculate(subX, subY, 1000)

			// Should always return 0 for image gradients
			if result != 0 {
				t.Errorf("GradientFunction.Calculate returned %d, expected 0", result)
			}

			// Check that color was sampled correctly
			sampledColor := gi.ColorFunction().ColorAt(0)
			expected := expectedColors[i]
			if sampledColor != expected {
				t.Errorf("Position (%d,%d): got %+v, want %+v", coord.x, coord.y, sampledColor, expected)
			}
		}
	})
}

func BenchmarkGradientImage_Calculate(b *testing.B) {
	gi := NewGradientImageRGBA8()

	// Create moderate-sized image
	buffer := gi.ImageCreate(256, 256)

	// Fill with test pattern
	for i := range buffer {
		buffer[i] = color.RGBA8[color.SRGB]{
			R: uint8(i & 0xFF),
			G: uint8((i >> 8) & 0xFF),
			B: uint8((i >> 16) & 0xFF),
			A: 255,
		}
	}

	b.ResetTimer()

	// Benchmark sampling at various coordinates
	for i := 0; i < b.N; i++ {
		x := (i * 17) % (256 << GradientSubpixelShift) // Pseudo-random coords
		y := (i * 31) % (256 << GradientSubpixelShift)
		gi.Calculate(x, y, 1000)
	}
}

func BenchmarkGradientImage_ImageCreate(b *testing.B) {
	gi := NewGradientImageRGBA8()

	b.ResetTimer()

	// Benchmark image creation with various sizes
	sizes := []struct{ w, h int }{
		{64, 64},
		{128, 128},
		{256, 256},
		{512, 256},
	}

	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		gi.ImageCreate(size.w, size.h)
	}
}
