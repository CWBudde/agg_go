package span

import (
	"testing"

	"agg_go/internal/transform"
)

// MockSubdivInterpolator is a test interpolator that implements SubdivInterpolator
type MockSubdivInterpolator struct {
	x, y           int
	startX, startY float64
	currentX       float64
	length         int
	resyncCount    int
	nextCount      int
	beginCount     int
}

func NewMockSubdivInterpolator() *MockSubdivInterpolator {
	return &MockSubdivInterpolator{}
}

func (m *MockSubdivInterpolator) Begin(x, y float64, length int) {
	m.startX = x
	m.startY = y
	m.currentX = x
	m.length = length
	m.x = int(x * 256) // Convert to subpixel units
	m.y = int(y * 256)
	m.beginCount++
}

func (m *MockSubdivInterpolator) Next() {
	m.currentX += 1.0
	m.x = int(m.currentX * 256) // Convert to subpixel units
	m.nextCount++
}

func (m *MockSubdivInterpolator) Resynchronize(xe, ye float64, length int) {
	m.currentX = xe
	m.x = int(xe * 256) // Convert to subpixel units
	m.y = int(ye * 256)
	m.resyncCount++
}

func (m *MockSubdivInterpolator) Coordinates() (x, y int) {
	return m.x, m.y
}

// LocalScale method for testing LocalScale support
func (m *MockSubdivInterpolator) LocalScale() (x, y int) {
	// Return fixed scale for testing
	return 256, 256
}

func TestSpanSubdivAdaptor_Construction(t *testing.T) {
	t.Run("DefaultConstruction", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		if adaptor.SubdivShift() != 4 {
			t.Errorf("Default subdivision shift: got %d, want 4", adaptor.SubdivShift())
		}

		if adaptor.SubpixelShift() != 8 {
			t.Errorf("Default subpixel shift: got %d, want 8", adaptor.SubpixelShift())
		}

		if adaptor.SubpixelScale() != 256 {
			t.Errorf("Subpixel scale: got %d, want 256", adaptor.SubpixelScale())
		}

		if adaptor.Interpolator() != mock {
			t.Error("Interpolator not properly stored")
		}
	})

	t.Run("CustomSubdivShift", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptorWithShift(mock, 6)

		if adaptor.SubdivShift() != 6 {
			t.Errorf("Custom subdivision shift: got %d, want 6", adaptor.SubdivShift())
		}

		// Subdivision size should be 2^6 = 64
		expectedSize := 1 << 6
		if adaptor.subdivSize != expectedSize {
			t.Errorf("Subdivision size: got %d, want %d", adaptor.subdivSize, expectedSize)
		}
	})

	t.Run("CustomBothShifts", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptorWithShifts(mock, 5, 10)

		if adaptor.SubdivShift() != 5 {
			t.Errorf("Custom subdivision shift: got %d, want 5", adaptor.SubdivShift())
		}

		if adaptor.SubpixelShift() != 10 {
			t.Errorf("Custom subpixel shift: got %d, want 10", adaptor.SubpixelShift())
		}

		if adaptor.SubpixelScale() != 1024 {
			t.Errorf("Subpixel scale: got %d, want 1024", adaptor.SubpixelScale())
		}
	})

	t.Run("AtPointConstruction", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		_ = NewSpanSubdivAdaptorAtPoint(mock, 10.5, 20.5, 50)

		if mock.beginCount != 1 {
			t.Errorf("Begin should be called once: got %d calls", mock.beginCount)
		}

		if mock.startX != 10.5 || mock.startY != 20.5 {
			t.Errorf("Begin coordinates: got (%.1f, %.1f), want (10.5, 20.5)", mock.startX, mock.startY)
		}

		if mock.length != 16 { // Limited by default subdivision size
			t.Errorf("Begin length: got %d, want 16", mock.length)
		}
	})
}

func TestSpanSubdivAdaptor_BasicFunctionality(t *testing.T) {
	t.Run("BeginAndCoordinates", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		adaptor.Begin(5.0, 10.0, 32)

		// Check that Begin was called on wrapped interpolator
		if mock.beginCount != 1 {
			t.Errorf("Begin calls: got %d, want 1", mock.beginCount)
		}

		// Check that initial coordinates match
		x, y := adaptor.Coordinates()
		expectedX := int(5.0 * 256)
		expectedY := int(10.0 * 256)
		if x != expectedX || y != expectedY {
			t.Errorf("Initial coordinates: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("NextAdvancement", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		adaptor.Begin(0.0, 0.0, 10)

		// Advance a few steps within subdivision
		for i := 0; i < 5; i++ {
			adaptor.Next()
		}

		if mock.nextCount != 5 {
			t.Errorf("Next calls: got %d, want 5", mock.nextCount)
		}

		// Should not have resynchronized yet
		if mock.resyncCount != 0 {
			t.Errorf("Unexpected resynchronization: got %d calls", mock.resyncCount)
		}
	})

	t.Run("LocalScale", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		// Mock does implement LocalScale, should return (256, 256)
		sx, sy := adaptor.LocalScale()
		if sx != 256 || sy != 256 {
			t.Errorf("Local scale: got (%d, %d), want (256, 256)", sx, sy)
		}
	})
}

func TestSpanSubdivAdaptor_Subdivision(t *testing.T) {
	t.Run("SubdivisionBoundary", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		// Use subdivision shift of 2 (4 pixels) for easier testing
		adaptor := NewSpanSubdivAdaptorWithShift(mock, 2)

		adaptor.Begin(0.0, 0.0, 20) // Long span that will be subdivided

		// Step to subdivision boundary (4 steps with shift=2)
		for i := 0; i < 4; i++ {
			adaptor.Next()
		}

		// At this point, we should have triggered resynchronization
		if mock.resyncCount != 1 {
			t.Errorf("Resynchronization count: got %d, want 1", mock.resyncCount)
		}

		// Continue stepping
		for i := 0; i < 4; i++ {
			adaptor.Next()
		}

		// Should have resynchronized again
		if mock.resyncCount != 2 {
			t.Errorf("Resynchronization count after second boundary: got %d, want 2", mock.resyncCount)
		}
	})

	t.Run("ShortSpan", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock) // Default shift=4 (16 pixels)

		adaptor.Begin(0.0, 0.0, 5) // Short span, less than subdivision size

		// Step through entire span
		for i := 0; i < 5; i++ {
			adaptor.Next()
		}

		// Should not have triggered resynchronization for short span
		if mock.resyncCount != 0 {
			t.Errorf("Unexpected resynchronization for short span: got %d calls", mock.resyncCount)
		}
	})

	t.Run("ExactSubdivisionMultiple", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptorWithShift(mock, 2) // 4 pixels per subdivision

		adaptor.Begin(0.0, 0.0, 8) // Exactly 2 subdivisions

		// Step through first subdivision
		for i := 0; i < 4; i++ {
			adaptor.Next()
		}

		if mock.resyncCount != 1 {
			t.Errorf("Resynchronization after first subdivision: got %d, want 1", mock.resyncCount)
		}

		// Step through second subdivision
		for i := 0; i < 4; i++ {
			adaptor.Next()
		}

		if mock.resyncCount != 2 {
			t.Errorf("Resynchronization after second subdivision: got %d, want 2", mock.resyncCount)
		}
	})
}

func TestSpanSubdivAdaptor_AccessorMethods(t *testing.T) {
	t.Run("InterpolatorAccessors", func(t *testing.T) {
		mock1 := NewMockSubdivInterpolator()
		mock2 := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock1)

		if adaptor.Interpolator() != mock1 {
			t.Error("Initial interpolator not correct")
		}

		adaptor.SetInterpolator(mock2)
		if adaptor.Interpolator() != mock2 {
			t.Error("Updated interpolator not correct")
		}
	})

	t.Run("SubdivShiftAccessors", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		adaptor.SetSubdivShift(6)
		if adaptor.SubdivShift() != 6 {
			t.Errorf("Updated subdivision shift: got %d, want 6", adaptor.SubdivShift())
		}

		// Check that internal values were updated
		expectedSize := 1 << 6
		if adaptor.subdivSize != expectedSize {
			t.Errorf("Subdivision size after update: got %d, want %d", adaptor.subdivSize, expectedSize)
		}

		expectedMask := expectedSize - 1
		if adaptor.subdivMask != expectedMask {
			t.Errorf("Subdivision mask after update: got %d, want %d", adaptor.subdivMask, expectedMask)
		}
	})

	t.Run("TransformerAccessors", func(t *testing.T) {
		// Test with a mock that doesn't support transformer interface
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		// Should return nil for mock that doesn't support transformer
		transformer := adaptor.Transformer()
		if transformer != nil {
			t.Error("Expected nil transformer for mock interpolator")
		}

		// SetTransformer should be a no-op for mock that doesn't support it
		trans := transform.NewTransAffine()
		adaptor.SetTransformer(trans) // Should not panic
	})
}

func TestSpanSubdivAdaptor_EdgeCases(t *testing.T) {
	t.Run("ZeroLength", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		// Should not panic with zero length
		adaptor.Begin(0.0, 0.0, 0)
		if mock.length != 0 {
			t.Errorf("Zero length not preserved: got %d", mock.length)
		}
	})

	t.Run("SinglePixel", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		adaptor.Begin(5.0, 10.0, 1)
		adaptor.Next()

		// Should not resynchronize for single pixel
		if mock.resyncCount != 0 {
			t.Errorf("Unexpected resynchronization for single pixel: got %d", mock.resyncCount)
		}
	})

	t.Run("VeryLongSpan", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptorWithShift(mock, 2) // 4 pixels per subdivision

		adaptor.Begin(0.0, 0.0, 100) // Very long span

		// Step through multiple subdivisions
		for i := 0; i < 20; i++ {
			adaptor.Next()
		}

		// Should have resynchronized multiple times (every 4 pixels)
		expectedResyncs := 20 / 4 // 5 resynchronizations
		if mock.resyncCount != expectedResyncs {
			t.Errorf("Resynchronization count for long span: got %d, want %d", mock.resyncCount, expectedResyncs)
		}
	})

	t.Run("Resynchronize", func(t *testing.T) {
		mock := NewMockSubdivInterpolator()
		adaptor := NewSpanSubdivAdaptor(mock)

		adaptor.Begin(0.0, 0.0, 10)

		// Direct resynchronize should delegate to wrapped interpolator
		adaptor.Resynchronize(50.0, 25.0, 5)

		if mock.resyncCount != 1 {
			t.Errorf("Resynchronize call count: got %d, want 1", mock.resyncCount)
		}
	})
}

// MockTransformerInterpolator extends MockSubdivInterpolator with transformer support
type MockTransformerInterpolator struct {
	*MockSubdivInterpolator
	transformer transform.Transformer
}

func NewMockTransformerInterpolator() *MockTransformerInterpolator {
	return &MockTransformerInterpolator{
		MockSubdivInterpolator: NewMockSubdivInterpolator(),
		transformer:            transform.NewTransAffine(),
	}
}

func (m *MockTransformerInterpolator) Transformer() transform.Transformer {
	return m.transformer
}

func (m *MockTransformerInterpolator) SetTransformer(transformer transform.Transformer) {
	m.transformer = transformer
}

func TestSpanSubdivAdaptor_WithTransformer(t *testing.T) {
	t.Run("TransformerSupport", func(t *testing.T) {
		mock := NewMockTransformerInterpolator()
		adaptor := NewSpanSubdivAdaptor[*MockTransformerInterpolator](mock)

		// Should return the transformer
		transformer := adaptor.Transformer()
		if transformer != mock.transformer {
			t.Error("Transformer not returned correctly")
		}

		// Should be able to set new transformer
		newTrans := transform.NewTransAffine()
		newTrans.ScaleXY(2.0, 2.0)
		adaptor.SetTransformer(newTrans)

		if mock.transformer != newTrans {
			t.Error("Transformer not set correctly")
		}
	})
}

func TestSpanSubdivAdaptor_Integration(t *testing.T) {
	t.Run("WithLinearInterpolator", func(t *testing.T) {
		// Test with actual linear interpolator
		trans := transform.NewTransAffine()
		trans.ScaleXY(2.0, 2.0)
		linear := NewSpanInterpolatorLinearDefault(trans)

		adaptor := NewSpanSubdivAdaptorWithShift(linear, 3) // 8 pixels per subdivision

		adaptor.Begin(1.0, 1.0, 20)

		// Should work without panicking
		x1, y1 := adaptor.Coordinates()

		// Step through subdivision boundary
		for i := 0; i < 10; i++ {
			adaptor.Next()
		}

		x2, y2 := adaptor.Coordinates()

		// Coordinates should have advanced
		if x2 <= x1 {
			t.Errorf("X coordinate should have advanced: x1=%d, x2=%d", x1, x2)
		}

		// Y coordinate should be similar (horizontal scan)
		if absInt(y2-y1) > 1000 { // Allow some tolerance for rounding
			t.Errorf("Y coordinate changed too much: y1=%d, y2=%d", y1, y2)
		}
	})

	t.Run("PerformanceComparison", func(t *testing.T) {
		// This test verifies that subdivision adaptor produces similar results
		// to direct interpolation (basic functionality test)
		trans := transform.NewTransAffine()
		trans.ScaleXY(1.5, 1.5)

		// Direct interpolation
		direct := NewSpanInterpolatorLinearDefault(trans)
		direct.Begin(0.0, 0.0, 50)

		// Subdivision adaptor
		adapted := NewSpanInterpolatorLinearDefault(trans)
		adaptor := NewSpanSubdivAdaptorWithShift(adapted, 4) // 16 pixels
		adaptor.Begin(0.0, 0.0, 50)

		// Compare coordinates at various points
		for i := 0; i < 20; i++ {
			x1, y1 := direct.Coordinates()
			x2, y2 := adaptor.Coordinates()

			// Allow some tolerance due to resynchronization
			if absInt(x2-x1) > 100 || absInt(y2-y1) > 100 {
				t.Errorf("Step %d: coordinates differ too much: direct=(%d,%d), adapted=(%d,%d)", i, x1, y1, x2, y2)
			}

			direct.Next()
			adaptor.Next()
		}
	})
}
