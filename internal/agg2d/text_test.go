package agg2d

import (
	"os"
	"testing"
)

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

	// Test with actual font loading when FreeType is available
	testWithFreeType(t, agg2d)
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

// testWithFreeType tests font loading when FreeType is available.
// This function is called from TestTextWidth to add comprehensive font tests.
func testWithFreeType(t *testing.T, agg2d *Agg2D) {
	fontPath := findSystemFont()
	if fontPath == "" {
		t.Skip("No system font found for FreeType testing")
		return
	}

	t.Logf("Testing with font: %s", fontPath)

	// Test basic font loading
	err := agg2d.Font(fontPath, 16.0, false, false, RasterFontCache, 0.0)
	if err != nil {
		t.Logf("FreeType not available or font load failed: %v", err)
		t.Skip("Skipping FreeType-dependent tests")
		return
	}

	// Test text width calculation with actual font
	width := agg2d.TextWidth("Hello World")
	if width <= 0.0 {
		t.Errorf("Expected positive text width with loaded font, got %v", width)
	}

	// Test with different text strings
	testStrings := []string{
		"A",
		"Hello",
		"The quick brown fox",
		"",
	}

	for _, str := range testStrings {
		width := agg2d.TextWidth(str)
		if str == "" && width != 0.0 {
			t.Errorf("Expected zero width for empty string, got %v", width)
		} else if str != "" && width <= 0.0 {
			t.Errorf("Expected positive width for '%s', got %v", str, width)
		}
	}
}

// TestFontLoadingWithFreeType tests various font loading scenarios.
func TestFontLoadingWithFreeType(t *testing.T) {
	fontPath := findSystemFont()
	if fontPath == "" {
		t.Skip("No system font found for FreeType testing")
		return
	}

	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	tests := []struct {
		name      string
		height    float64
		bold      bool
		italic    bool
		cacheType FontCacheType
		angle     float64
	}{
		{"Basic", 16.0, false, false, RasterFontCache, 0.0},
		{"Bold", 16.0, true, false, RasterFontCache, 0.0},
		{"Italic", 16.0, false, true, RasterFontCache, 0.0},
		{"BoldItalic", 16.0, true, true, RasterFontCache, 0.0},
		{"LargeSize", 32.0, false, false, RasterFontCache, 0.0},
		{"SmallSize", 8.0, false, false, RasterFontCache, 0.0},
		{"VectorCache", 16.0, false, false, VectorFontCache, 0.0},
		{"WithAngle", 16.0, false, false, RasterFontCache, 45.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agg2d.Font(fontPath, tt.height, tt.bold, tt.italic, tt.cacheType, tt.angle)
			if err != nil {
				t.Logf("Font loading failed for %s: %v", tt.name, err)
				t.Skip("Skipping test due to font loading failure")
				return
			}

			// Verify font height was set
			if agg2d.FontHeight() != tt.height {
				t.Errorf("Expected font height %v, got %v", tt.height, agg2d.FontHeight())
			}

			// Test text rendering doesn't crash
			agg2d.Text(100, 100, "Test Text", false, 0, 0)

			// Test text width calculation
			width := agg2d.TextWidth("Test")
			// Note: VectorFontCache may return 0 width as it's not designed for text measurement
			if tt.cacheType != VectorFontCache && width <= 0.0 {
				t.Errorf("Expected positive text width for cache type %v, got %v", tt.cacheType, width)
			}
		})
	}
}

// TestTextRenderingWithFreeType tests text rendering with actual fonts.
func TestTextRenderingWithFreeType(t *testing.T) {
	fontPath := findSystemFont()
	if fontPath == "" {
		t.Skip("No system font found for FreeType testing")
		return
	}

	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)
	agg2d.ClearAll(White)
	agg2d.FillColor(Black)

	err := agg2d.Font(fontPath, 16.0, false, false, RasterFontCache, 0.0)
	if err != nil {
		t.Skip("FreeType not available or font load failed")
		return
	}

	// Test rendering with different alignments
	alignments := []struct {
		name   string
		alignX TextAlignment
		alignY TextAlignment
	}{
		{"LeftTop", AlignLeft, AlignTop},
		{"CenterCenter", AlignCenter, AlignCenter},
		{"RightBottom", AlignRight, AlignBottom},
	}

	for _, alignment := range alignments {
		t.Run(alignment.name, func(t *testing.T) {
			agg2d.TextAlignment(alignment.alignX, alignment.alignY)
			agg2d.Text(400, 300, "Test Text with "+alignment.name, false, 0, 0)
			// Should not crash - this is the main test
		})
	}

	// Test with Unicode characters
	agg2d.Text(100, 400, "Unicode: 世界 🌍", false, 0, 0)

	// Test with different font heights
	heights := []float64{8.0, 12.0, 16.0, 24.0, 32.0}
	for i, height := range heights {
		err := agg2d.Font(fontPath, height, false, false, RasterFontCache, 0.0)
		if err != nil {
			continue
		}
		agg2d.Text(100, 100+float64(i*40), "Height test", false, 0, 0)
	}
}

// BenchmarkTextRenderWithFreeType benchmarks text rendering with actual fonts.
func BenchmarkTextRenderWithFreeType(b *testing.B) {
	fontPath := findSystemFont()
	if fontPath == "" {
		b.Skip("No system font found for FreeType testing")
		return
	}

	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)
	agg2d.ClearAll(White)
	agg2d.FillColor(Black)

	err := agg2d.Font(fontPath, 16.0, false, false, RasterFontCache, 0.0)
	if err != nil {
		b.Skip("FreeType not available or font load failed")
		return
	}

	testText := "Benchmark Text with FreeType"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.Text(float64(i%800), float64((i/800)%600), testText, false, 0, 0)
	}
}

// BenchmarkTextWidthWithFreeType benchmarks text width calculation with actual fonts.
func BenchmarkTextWidthWithFreeType(b *testing.B) {
	fontPath := findSystemFont()
	if fontPath == "" {
		b.Skip("No system font found for FreeType testing")
		return
	}

	agg2d := NewAgg2D()
	buf := make([]byte, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	err := agg2d.Font(fontPath, 16.0, false, false, RasterFontCache, 0.0)
	if err != nil {
		b.Skip("FreeType not available or font load failed")
		return
	}

	testText := "The quick brown fox jumps over the lazy dog"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.TextWidth(testText)
	}
}

// findSystemFont attempts to locate a system font for testing.
// This function tries common font paths on Linux, macOS, and Windows.
func findSystemFont() string {
	// Common system font paths
	fontPaths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",                 // Linux (Debian/Ubuntu)
		"/usr/share/fonts/TTF/DejaVuSans.ttf",                             // Linux (Arch)
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf", // Linux (RedHat/Fedora)
		"/usr/share/fonts/truetype/ubuntu-font-family/Ubuntu-R.ttf",       // Linux (Ubuntu)
		"/usr/share/fonts/truetype/lato/Lato-Regular.ttf",                 // Linux (some distros)
		"/System/Library/Fonts/Arial.ttf",                                 // macOS
		"/System/Library/Fonts/Helvetica.ttc",                             // macOS
		"/Library/Fonts/Arial.ttf",                                        // macOS (user)
		"C:\\Windows\\Fonts\\arial.ttf",                                   // Windows
		"C:\\Windows\\Fonts\\calibri.ttf",                                 // Windows
	}

	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			return fontPath
		}
	}

	return ""
}
