package agg

import (
	"testing"
)

// Mock font for testing when FreeType is not available
type mockFontEngine struct {
	signature   string
	changeStamp int
	height      float64
	width       float64
	hinting     bool
	flipY       bool
}

func (m *mockFontEngine) FontSignature() string { return m.signature }
func (m *mockFontEngine) ChangeStamp() int      { return m.changeStamp }

func (m *mockFontEngine) PrepareGlyph(glyphCode uint) bool { return true }
func (m *mockFontEngine) GlyphIndex() uint                 { return 0 }
func (m *mockFontEngine) DataSize() uint                   { return 64 }
func (m *mockFontEngine) DataType() interface{}            { return nil }
func (m *mockFontEngine) Bounds() interface{}              { return nil }
func (m *mockFontEngine) AdvanceX() float64                { return 12.0 }
func (m *mockFontEngine) AdvanceY() float64                { return 0.0 }
func (m *mockFontEngine) WriteGlyphTo(data []byte)         {}
func (m *mockFontEngine) AddKerning(first, second uint) (dx, dy float64) {
	return 1.0, 0.0
}
func (m *mockFontEngine) PathAdaptor() interface{} { return nil }

// TestTextAlignmentText tests the text alignment functionality for text rendering.
func TestTextAlignmentText(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	tests := []struct {
		name   string
		alignX TextAlignment
		alignY TextAlignment
	}{
		{"LeftBottom", AlignLeft, AlignBottom},
		{"CenterCenter", AlignCenter, AlignCenter},
		{"RightTop", AlignRight, AlignTop},
		{"LeftTop", AlignLeft, AlignTop},
		{"RightBottom", AlignRight, AlignBottom},
		{"CenterTop", AlignCenter, AlignTop},
		{"CenterBottom", AlignCenter, AlignBottom},
		{"LeftCenter", AlignLeft, AlignCenter},
		{"RightCenter", AlignRight, AlignCenter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg2d.TextAlignment(tt.alignX, tt.alignY)

			if agg2d.textAlignX != tt.alignX {
				t.Errorf("Expected X alignment %v, got %v", tt.alignX, agg2d.textAlignX)
			}
			if agg2d.textAlignY != tt.alignY {
				t.Errorf("Expected Y alignment %v, got %v", tt.alignY, agg2d.textAlignY)
			}
		})
	}
}

// TestTextHints tests the text hinting functionality.
func TestTextHints(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	// Test default value (check actual default in initialization)
	defaultHints := agg2d.GetTextHints()
	t.Logf("Default text hints: %t", defaultHints)

	// Test setting to true
	agg2d.TextHints(true)
	if !agg2d.GetTextHints() {
		t.Error("Expected text hints to be true after setting to true")
	}

	// Test setting to false
	agg2d.TextHints(false)
	if agg2d.GetTextHints() {
		t.Error("Expected text hints to be false after setting to false")
	}
}

// TestFontHeight tests the font height functionality.
func TestFontHeight(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	heights := []float64{12.0, 16.0, 24.0, 32.0, 48.0}

	for _, height := range heights {
		t.Run("Height"+string(rune(int(height))), func(t *testing.T) {
			agg2d.fontHeight = height
			if agg2d.FontHeight() != height {
				t.Errorf("Expected font height %v, got %v", height, agg2d.FontHeight())
			}
		})
	}
}

// TestTextWidth tests the text width calculation.
// Note: This test uses a mock implementation since FreeType may not be available.
func TestTextWidth(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	// Test with no font loaded
	width := agg2d.TextWidth("Hello World")
	if width != 0.0 {
		t.Errorf("Expected text width to be 0 with no font loaded, got %v", width)
	}

	// TODO: Add tests with actual font loading when FreeType is available
}

// TestTextRendering tests basic text rendering functionality.
func TestTextRendering(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	// Clear with white background
	agg2d.ClearAll(White)

	// Set text color to black
	agg2d.FillColor(Black)

	// Test rendering text without crashing
	// This should work even without a font loaded (just return early)
	agg2d.Text(100, 100, "Hello World", false, 0, 0)
	agg2d.Text(200, 200, "Test Text", true, 5, -2)

	// Test with different alignments
	agg2d.TextAlignment(AlignCenter, AlignCenter)
	agg2d.Text(400, 300, "Centered", false, 0, 0)

	agg2d.TextAlignment(AlignRight, AlignTop)
	agg2d.Text(700, 100, "Right Top", false, 0, 0)
}

// TestFontCacheTypes tests different font cache types.
func TestFontCacheTypes(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	cacheTypes := []FontCacheType{
		RasterFontCache,
		VectorFontCache,
	}

	for _, cacheType := range cacheTypes {
		t.Run("CacheType"+string(rune(int(cacheType))), func(t *testing.T) {
			agg2d.fontCacheType = cacheType
			// Test that the cache type is stored correctly
			if agg2d.fontCacheType != cacheType {
				t.Errorf("Expected cache type %v, got %v", cacheType, agg2d.fontCacheType)
			}
		})
	}
}

// TestFlipText tests the text flipping functionality.
func TestFlipText(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	// Test FlipText without crashing (should handle nil font engine gracefully)
	agg2d.FlipText(true)
	agg2d.FlipText(false)
}

// TestTextWithRotation tests text rendering with rotation angle.
func TestTextWithRotation(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	// Test different rotation angles
	angles := []float64{0.0, 45.0, 90.0, 180.0, 270.0, 360.0}

	for _, angle := range angles {
		t.Run("Angle"+string(rune(int(angle))), func(t *testing.T) {
			agg2d.textAngle = angle * Pi / 180.0 // Convert to radians
			// Test that text rendering doesn't crash with rotation
			agg2d.Text(400, 300, "Rotated Text", false, 0, 0)
		})
	}
}

// TestTextPositioning tests text positioning with different parameters.
func TestTextPositioning(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	positions := []struct {
		x, y     float64
		roundOff bool
		dx, dy   float64
	}{
		{100, 100, false, 0, 0},
		{200.5, 150.7, true, 0, 0},  // Test rounding
		{300, 200, false, 10, -5},   // Test with offset
		{400.3, 250.8, true, -3, 7}, // Test rounding with offset
	}

	for i, pos := range positions {
		t.Run("Position"+string(rune('0'+i)), func(t *testing.T) {
			// Should not crash regardless of parameters
			agg2d.Text(pos.x, pos.y, "Test", pos.roundOff, pos.dx, pos.dy)
		})
	}
}

// TestTextWithEmptyString tests text rendering with edge cases.
func TestTextWithEmptyString(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	// Test with empty string - should return early without error
	agg2d.Text(100, 100, "", false, 0, 0)

	// Test with single character
	agg2d.Text(200, 200, "A", false, 0, 0)

	// Test with Unicode characters
	agg2d.Text(300, 300, "Hello 世界 🌍", false, 0, 0)
}

// TestMultipleTextRenders tests rendering multiple text strings.
func TestMultipleTextRenders(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	// Clear background
	agg2d.ClearAll(White)
	agg2d.FillColor(Black)

	// Render multiple text strings with different settings
	texts := []struct {
		x, y   float64
		text   string
		alignX TextAlignment
		alignY TextAlignment
	}{
		{100, 100, "Top Left", AlignLeft, AlignTop},
		{400, 100, "Top Center", AlignCenter, AlignTop},
		{700, 100, "Top Right", AlignRight, AlignTop},
		{100, 300, "Middle Left", AlignLeft, AlignCenter},
		{400, 300, "Center", AlignCenter, AlignCenter},
		{700, 300, "Middle Right", AlignRight, AlignCenter},
		{100, 500, "Bottom Left", AlignLeft, AlignBottom},
		{400, 500, "Bottom Center", AlignCenter, AlignBottom},
		{700, 500, "Bottom Right", AlignRight, AlignBottom},
	}

	for _, text := range texts {
		agg2d.TextAlignment(text.alignX, text.alignY)
		agg2d.Text(text.x, text.y, text.text, false, 0, 0)
	}
}

// BenchmarkTextWidth benchmarks the text width calculation.
func BenchmarkTextWidth(b *testing.B) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	testText := "The quick brown fox jumps over the lazy dog"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.TextWidth(testText)
	}
}

// BenchmarkTextRender benchmarks text rendering.
func BenchmarkTextRender(b *testing.B) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)
	agg2d.ClearAll(White)
	agg2d.FillColor(Black)

	testText := "Benchmark Text"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.Text(float64(i%800), float64((i/800)%600), testText, false, 0, 0)
	}
}
