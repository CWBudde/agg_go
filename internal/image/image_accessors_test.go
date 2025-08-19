package image

import (
	"testing"

	"agg_go/internal/basics"
)

// MockPixelFormat implements PixelFormat for testing
type MockPixelFormat struct {
	width    int
	height   int
	pixWidth int
	data     []basics.Int8u
}

func NewMockPixelFormat(width, height, pixWidth int) *MockPixelFormat {
	return &MockPixelFormat{
		width:    width,
		height:   height,
		pixWidth: pixWidth,
		data:     make([]basics.Int8u, width*height*pixWidth),
	}
}

func (m *MockPixelFormat) Width() int {
	return m.width
}

func (m *MockPixelFormat) Height() int {
	return m.height
}

func (m *MockPixelFormat) PixWidth() int {
	return m.pixWidth
}

func (m *MockPixelFormat) PixPtr(x, y int) []basics.Int8u {
	if x < 0 || y < 0 || x >= m.width || y >= m.height {
		// Return empty slice for out-of-bounds (will panic if accessed)
		return nil
	}
	offset := (y*m.width + x) * m.pixWidth
	return m.data[offset:]
}

// SetPixel sets a pixel value for testing
func (m *MockPixelFormat) SetPixel(x, y int, values []basics.Int8u) {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		offset := (y*m.width + x) * m.pixWidth
		copy(m.data[offset:offset+m.pixWidth], values)
	}
}

// GetPixel gets a pixel value for testing
func (m *MockPixelFormat) GetPixel(x, y int) []basics.Int8u {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		offset := (y*m.width + x) * m.pixWidth
		result := make([]basics.Int8u, m.pixWidth)
		copy(result, m.data[offset:offset+m.pixWidth])
		return result
	}
	return nil
}

func TestImageAccessorClip(t *testing.T) {
	// Create a 4x4 pixel format with 3 bytes per pixel (RGB)
	pixfmt := NewMockPixelFormat(4, 4, 3)

	// Set some test pixels
	pixfmt.SetPixel(1, 1, []basics.Int8u{255, 0, 0}) // Red
	pixfmt.SetPixel(2, 2, []basics.Int8u{0, 255, 0}) // Green
	pixfmt.SetPixel(3, 3, []basics.Int8u{0, 0, 255}) // Blue

	// Background color: White
	backgroundColor := []basics.Int8u{255, 255, 255}
	accessor := NewImageAccessorClip(&pixfmt, backgroundColor)

	t.Run("in bounds access", func(t *testing.T) {
		span := accessor.Span(1, 1, 1)
		if span[0] != 255 || span[1] != 0 || span[2] != 0 {
			t.Errorf("Expected red pixel (255,0,0), got (%d,%d,%d)", span[0], span[1], span[2])
		}
	})

	t.Run("out of bounds access", func(t *testing.T) {
		span := accessor.Span(-1, -1, 1)
		if span[0] != 255 || span[1] != 255 || span[2] != 255 {
			t.Errorf("Expected background color (255,255,255), got (%d,%d,%d)", span[0], span[1], span[2])
		}
	})

	t.Run("span extending out of bounds", func(t *testing.T) {
		// With width=4, x=3, length=2 means span goes from 3 to 5, but max valid is 4
		span := accessor.Span(3, 0, 2) // x+length = 3+2 = 5 > width=4
		if span[0] != 255 || span[1] != 255 || span[2] != 255 {
			t.Errorf("Expected background color for out-of-bounds span, got (%d,%d,%d)", span[0], span[1], span[2])
		}
	})

	t.Run("NextX navigation", func(t *testing.T) {
		accessor.Span(1, 1, 2) // Start at red pixel
		next := accessor.NextX()

		// Should be at (2,1) which is unset (default 0,0,0)
		if next[0] != 0 || next[1] != 0 || next[2] != 0 {
			t.Errorf("Expected black pixel (0,0,0), got (%d,%d,%d)", next[0], next[1], next[2])
		}
	})

	t.Run("NextY navigation", func(t *testing.T) {
		accessor.Span(2, 2, 1) // Start at green pixel
		next := accessor.NextY()

		// Should be at (2,3) which is unset (default 0,0,0)
		if next[0] != 0 || next[1] != 0 || next[2] != 0 {
			t.Errorf("Expected black pixel (0,0,0), got (%d,%d,%d)", next[0], next[1], next[2])
		}
	})
}

func TestImageAccessorNoClip(t *testing.T) {
	// Create a 3x3 pixel format with 2 bytes per pixel
	pixfmt := NewMockPixelFormat(3, 3, 2)

	// Set test pixel
	pixfmt.SetPixel(1, 1, []basics.Int8u{100, 200})

	accessor := NewImageAccessorNoClip(&pixfmt)

	t.Run("basic access", func(t *testing.T) {
		span := accessor.Span(1, 1, 1)
		if span[0] != 100 || span[1] != 200 {
			t.Errorf("Expected (100,200), got (%d,%d)", span[0], span[1])
		}
	})

	t.Run("NextX navigation", func(t *testing.T) {
		accessor.Span(0, 0, 2)
		next := accessor.NextX()

		// Should be at (1,0)
		if len(next) < 2 {
			t.Errorf("NextX should return valid pixel data")
		}
	})

	t.Run("NextY navigation", func(t *testing.T) {
		accessor.Span(1, 0, 1)
		next := accessor.NextY()

		// Should be at (1,1) which has our test data
		if next[0] != 100 || next[1] != 200 {
			t.Errorf("Expected (100,200) at (1,1), got (%d,%d)", next[0], next[1])
		}
	})
}

func TestImageAccessorClone(t *testing.T) {
	// Create a 3x3 pixel format with 1 byte per pixel
	pixfmt := NewMockPixelFormat(3, 3, 1)

	// Set corner and edge pixels
	pixfmt.SetPixel(0, 0, []basics.Int8u{10})  // Top-left
	pixfmt.SetPixel(2, 0, []basics.Int8u{20})  // Top-right
	pixfmt.SetPixel(0, 2, []basics.Int8u{30})  // Bottom-left
	pixfmt.SetPixel(2, 2, []basics.Int8u{40})  // Bottom-right
	pixfmt.SetPixel(1, 1, []basics.Int8u{100}) // Center

	accessor := NewImageAccessorClone(&pixfmt)

	t.Run("in bounds access", func(t *testing.T) {
		span := accessor.Span(1, 1, 1)
		if span[0] != 100 {
			t.Errorf("Expected center pixel value 100, got %d", span[0])
		}
	})

	t.Run("clamp negative x", func(t *testing.T) {
		span := accessor.Span(-5, 0, 1)
		if span[0] != 10 { // Should clamp to (0,0)
			t.Errorf("Expected top-left pixel value 10, got %d", span[0])
		}
	})

	t.Run("clamp negative y", func(t *testing.T) {
		span := accessor.Span(2, -3, 1)
		if span[0] != 20 { // Should clamp to (2,0)
			t.Errorf("Expected top-right pixel value 20, got %d", span[0])
		}
	})

	t.Run("clamp large x", func(t *testing.T) {
		span := accessor.Span(10, 2, 1)
		if span[0] != 40 { // Should clamp to (2,2)
			t.Errorf("Expected bottom-right pixel value 40, got %d", span[0])
		}
	})

	t.Run("clamp large y", func(t *testing.T) {
		span := accessor.Span(0, 10, 1)
		if span[0] != 30 { // Should clamp to (0,2)
			t.Errorf("Expected bottom-left pixel value 30, got %d", span[0])
		}
	})

	t.Run("NextX with clamping", func(t *testing.T) {
		accessor.Span(2, 1, 1)   // Start at right edge (2,1)
		next := accessor.NextX() // x becomes 3, should clamp to 2

		// Since (2,1) hasn't been set explicitly, it should be 0
		if next[0] != 0 {
			t.Errorf("Expected clamped pixel value 0, got %d", next[0])
		}
	})
}

func TestImageAccessorWrap(t *testing.T) {
	// Create a 2x2 pixel format with 1 byte per pixel
	pixfmt := NewMockPixelFormat(2, 2, 1)

	// Set all pixels to unique values
	pixfmt.SetPixel(0, 0, []basics.Int8u{1})
	pixfmt.SetPixel(1, 0, []basics.Int8u{2})
	pixfmt.SetPixel(0, 1, []basics.Int8u{3})
	pixfmt.SetPixel(1, 1, []basics.Int8u{4})

	wrapX := NewWrapModeRepeat(2)
	wrapY := NewWrapModeRepeat(2)
	accessor := NewImageAccessorWrap(&pixfmt, wrapX, wrapY)

	t.Run("normal access", func(t *testing.T) {
		span := accessor.Span(0, 0, 1)
		if span[0] != 1 {
			t.Errorf("Expected pixel value 1, got %d", span[0])
		}
	})

	t.Run("wrap x coordinate", func(t *testing.T) {
		span := accessor.Span(2, 0, 1) // x=2 should wrap to x=0
		if span[0] != 1 {
			t.Errorf("Expected wrapped pixel value 1, got %d", span[0])
		}
	})

	t.Run("wrap y coordinate", func(t *testing.T) {
		span := accessor.Span(1, 2, 1) // y=2 should wrap to y=0
		if span[0] != 2 {
			t.Errorf("Expected wrapped pixel value 2, got %d", span[0])
		}
	})

	t.Run("wrap both coordinates", func(t *testing.T) {
		span := accessor.Span(3, 3, 1) // Both should wrap to (1,1)
		if span[0] != 4 {
			t.Errorf("Expected wrapped pixel value 4, got %d", span[0])
		}
	})

	t.Run("NextX wrapping", func(t *testing.T) {
		accessor.Span(1, 0, 1)   // Start at right edge
		next := accessor.NextX() // Should wrap to left edge
		if next[0] != 1 {        // Should be at (0,0)
			t.Errorf("Expected wrapped NextX value 1, got %d", next[0])
		}
	})

	t.Run("NextY wrapping", func(t *testing.T) {
		accessor.Span(0, 1, 1)   // Start at bottom edge
		next := accessor.NextY() // Should wrap to top edge
		if next[0] != 1 {        // Should be at (0,0)
			t.Errorf("Expected wrapped NextY value 1, got %d", next[0])
		}
	})
}

func TestImageAccessorWrap_WithReflect(t *testing.T) {
	// Test wrapping with reflect mode
	pixfmt := NewMockPixelFormat(3, 3, 1)

	// Set distinctive pattern
	pixfmt.SetPixel(0, 0, []basics.Int8u{10})
	pixfmt.SetPixel(1, 0, []basics.Int8u{20})
	pixfmt.SetPixel(2, 0, []basics.Int8u{30})

	wrapX := NewWrapModeReflect(3)
	wrapY := NewWrapModeReflect(3)
	accessor := NewImageAccessorWrap(&pixfmt, wrapX, wrapY)

	t.Run("reflect access", func(t *testing.T) {
		span := accessor.Span(4, 0, 1) // Should reflect to x=1
		if span[0] != 20 {
			t.Errorf("Expected reflected pixel value 20, got %d", span[0])
		}
	})
}

// Benchmark tests for performance comparison
func BenchmarkImageAccessorClip(b *testing.B) {
	pixfmt := NewMockPixelFormat(100, 100, 3)
	backgroundColor := []basics.Int8u{255, 255, 255}
	accessor := NewImageAccessorClip(&pixfmt, backgroundColor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessor.Span(i%100, (i/100)%100, 1)
		accessor.NextX()
	}
}

func BenchmarkImageAccessorNoClip(b *testing.B) {
	pixfmt := NewMockPixelFormat(100, 100, 3)
	accessor := NewImageAccessorNoClip(&pixfmt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessor.Span(i%100, (i/100)%100, 1)
		accessor.NextX()
	}
}

func BenchmarkImageAccessorClone(b *testing.B) {
	pixfmt := NewMockPixelFormat(100, 100, 3)
	accessor := NewImageAccessorClone(&pixfmt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessor.Span(i%100, (i/100)%100, 1)
		accessor.NextX()
	}
}

func BenchmarkImageAccessorWrap(b *testing.B) {
	pixfmt := NewMockPixelFormat(100, 100, 3)
	wrapX := NewWrapModeRepeat(100)
	wrapY := NewWrapModeRepeat(100)
	accessor := NewImageAccessorWrap(&pixfmt, wrapX, wrapY)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessor.Span(i%100, (i/100)%100, 1)
		accessor.NextX()
	}
}
