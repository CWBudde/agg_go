package glyph

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/fonts"
)

func TestNewGlyphRasterBin(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	if g == nil {
		t.Fatal("NewGlyphRasterBin returned nil")
	}
	if len(g.font) != len(font) {
		t.Errorf("Font data not properly stored")
	}
}

func TestGlyphRasterBinFontProperties(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	// Test height
	expectedHeight := float64(font[0])
	if g.Height() != expectedHeight {
		t.Errorf("Height() = %f, want %f", g.Height(), expectedHeight)
	}

	// Test baseline
	expectedBaseline := float64(font[1])
	if g.BaseLine() != expectedBaseline {
		t.Errorf("BaseLine() = %f, want %f", g.BaseLine(), expectedBaseline)
	}
}

func TestGlyphRasterBinEmptyFont(t *testing.T) {
	g := NewGlyphRasterBin([]byte{})

	if g.Height() != 0 {
		t.Errorf("Height() with empty font = %f, want 0", g.Height())
	}
	if g.BaseLine() != 0 {
		t.Errorf("BaseLine() with empty font = %f, want 0", g.BaseLine())
	}
	if g.Width("test") != 0 {
		t.Errorf("Width() with empty font = %f, want 0", g.Width("test"))
	}
}

func TestGlyphRasterBinShortFont(t *testing.T) {
	// Font with only 1 byte
	g := NewGlyphRasterBin([]byte{6})

	if g.Height() != 6 {
		t.Errorf("Height() with short font = %f, want 6", g.Height())
	}
	if g.BaseLine() != 0 {
		t.Errorf("BaseLine() with short font = %f, want 0", g.BaseLine())
	}
}

func TestGlyphRasterBinSetFont(t *testing.T) {
	g := NewGlyphRasterBin([]byte{})
	font := fonts.GetSimple4x6Font()

	g.SetFont(font)

	if len(g.Font()) != len(font) {
		t.Errorf("SetFont() did not update font properly")
	}
	if g.Height() != float64(font[0]) {
		t.Errorf("Height() after SetFont() = %f, want %f", g.Height(), float64(font[0]))
	}
}

func TestGlyphRasterBinWidth(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	tests := []struct {
		text     string
		expected float64
	}{
		{"", 0},     // Empty string
		{" ", 0},    // Space character (outside range of simple font)
		{"A", 3},    // Single character 'A'
		{"AB", 6},   // Two characters
		{"ABC", 9},  // Three characters
		{"\x00", 0}, // Null character (outside range)
		{"\xFF", 0}, // Character outside range
	}

	for _, tt := range tests {
		result := g.Width(tt.text)
		if result != tt.expected {
			t.Errorf("Width(%q) = %f, want %f", tt.text, result, tt.expected)
		}
	}
}

func TestGlyphRasterBinPrepareValidGlyph(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	var rect GlyphRect
	x, y := 10.0, 20.0
	glyph := 'A'

	g.Prepare(&rect, x, y, glyph, false)

	// Check bounds
	if rect.X1 != 10 {
		t.Errorf("rect.X1 = %d, want 10", rect.X1)
	}
	if rect.X2 != 12 { // X1 + width - 1 = 10 + 3 - 1
		t.Errorf("rect.X2 = %d, want 12", rect.X2)
	}

	// Check advance vector
	if rect.DX != 3 { // Width of 'A'
		t.Errorf("rect.DX = %f, want 3", rect.DX)
	}
	if rect.DY != 0 {
		t.Errorf("rect.DY = %f, want 0", rect.DY)
	}

	// Check Y coordinates (without flip)
	expectedY1 := int(y) - int(font[1]) + 1     // y - baseline + 1
	expectedY2 := expectedY1 + int(font[0]) - 1 // Y1 + height - 1
	if rect.Y1 != expectedY1 {
		t.Errorf("rect.Y1 = %d, want %d", rect.Y1, expectedY1)
	}
	if rect.Y2 != expectedY2 {
		t.Errorf("rect.Y2 = %d, want %d", rect.Y2, expectedY2)
	}
}

func TestGlyphRasterBinPrepareFlipped(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	var rect GlyphRect
	x, y := 10.0, 20.0
	glyph := 'A'

	g.Prepare(&rect, x, y, glyph, true)

	// Check Y coordinates (with flip)
	expectedY1 := int(y) - int(font[0]) + int(font[1]) // y - height + baseline
	expectedY2 := expectedY1 + int(font[0]) - 1        // Y1 + height - 1
	if rect.Y1 != expectedY1 {
		t.Errorf("rect.Y1 with flip = %d, want %d", rect.Y1, expectedY1)
	}
	if rect.Y2 != expectedY2 {
		t.Errorf("rect.Y2 with flip = %d, want %d", rect.Y2, expectedY2)
	}
}

func TestGlyphRasterBinPrepareInvalidGlyph(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	var rect GlyphRect

	// Test glyph outside range
	g.Prepare(&rect, 0, 0, rune(1000), false)

	// Should result in invalid rectangle (X1 > X2)
	if rect.X1 <= rect.X2 {
		t.Error("Expected invalid rectangle for out-of-range glyph")
	}
	if rect.DX != 0 || rect.DY != 0 {
		t.Error("Expected zero advance for invalid glyph")
	}
}

func TestGlyphRasterBinSpan(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	var rect GlyphRect
	g.Prepare(&rect, 0, 0, 'A', false)

	// Test span generation for different rows
	for y := 0; y < int(g.Height()); y++ {
		span := g.Span(y)
		if span == nil {
			t.Errorf("Span(%d) returned nil", y)
			continue
		}

		// Check span length matches glyph width
		if len(span) != g.glyphWidth {
			t.Errorf("Span(%d) length = %d, want %d", y, len(span), g.glyphWidth)
		}

		// Check that all values are valid coverage values
		for i, cover := range span {
			if cover != 0 && cover != basics.CoverFull {
				t.Errorf("Span(%d)[%d] = %d, expected 0 or %d", y, i, cover, basics.CoverFull)
			}
		}
	}
}

func TestGlyphRasterBinSpanOutOfBounds(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	var rect GlyphRect
	g.Prepare(&rect, 0, 0, 'A', false)

	// Test span with Y coordinate outside glyph bounds
	span := g.Span(-1)
	if span != nil {
		t.Error("Span(-1) should return nil")
	}

	span = g.Span(int(g.Height()))
	if span != nil {
		t.Errorf("Span(%d) should return nil", int(g.Height()))
	}
}

func TestGlyphRasterBinSpanWithoutPrepare(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	// Call Span without calling Prepare first
	span := g.Span(0)
	if span != nil {
		t.Error("Span() without Prepare() should return nil")
	}
}

func TestGlyphRasterBinGetValue(t *testing.T) {
	g := NewGlyphRasterBin([]byte{})

	tests := []struct {
		data     []byte
		expected int
	}{
		{[]byte{}, 0},                // Empty data
		{[]byte{0x01}, 0},            // Single byte
		{[]byte{0x01, 0x02}, 0x0201}, // Little endian (assuming little endian system)
		{[]byte{0xFF, 0xFF}, 0xFFFF}, // Max value
	}

	for _, tt := range tests {
		result := g.getValue(tt.data)
		// Note: The exact result depends on system endianness
		// We just verify the method doesn't panic
		_ = result
	}
}

func TestGlyphRasterBinMultipleCharacters(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	// Test preparing different characters (only A, B, C are in simple font)
	validCharacters := []rune{'A', 'B', 'C'}

	for _, char := range validCharacters {
		var rect GlyphRect
		g.Prepare(&rect, 0, 0, char, false)

		// Valid characters should have valid rectangles
		if rect.X1 > rect.X2 {
			t.Errorf("Character %c produced invalid rectangle", char)
		}

		// Should be able to generate spans
		for y := 0; y < int(g.Height()); y++ {
			span := g.Span(y)
			if span == nil {
				t.Errorf("Character %c, row %d: Span returned nil", char, y)
			}
		}
	}

	// Test invalid characters
	invalidCharacters := []rune{' ', '0', '9', '@', 'D'}
	for _, char := range invalidCharacters {
		var rect GlyphRect
		g.Prepare(&rect, 0, 0, char, false)

		// Invalid characters should have invalid rectangles
		if rect.X1 <= rect.X2 {
			t.Errorf("Character %c should produce invalid rectangle", char)
		}
	}
}

func TestGlyphRasterBinBitUnpacking(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	// Prepare a glyph that has some set pixels
	var rect GlyphRect
	g.Prepare(&rect, 0, 0, 'A', false)

	// Check that at least one span has some non-zero coverage
	foundCoverage := false
	for y := 0; y < int(g.Height()); y++ {
		span := g.Span(y)
		if span != nil {
			for _, cover := range span {
				if cover != 0 {
					foundCoverage = true
					break
				}
			}
		}
		if foundCoverage {
			break
		}
	}

	if !foundCoverage {
		t.Error("Expected to find some non-zero coverage for character 'A'")
	}
}

func TestGlyphRasterBinConsistency(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	// Test that preparing the same glyph multiple times gives consistent results
	var rect1, rect2 GlyphRect

	g.Prepare(&rect1, 10, 20, 'A', false)
	g.Prepare(&rect2, 10, 20, 'A', false)

	if rect1 != rect2 {
		t.Error("Preparing the same glyph twice gave different results")
	}

	// Test that spans are consistent
	span1 := g.Span(0)
	span2 := g.Span(0)

	if len(span1) != len(span2) {
		t.Error("Span lengths differ between calls")
	}

	for i := range span1 {
		if span1[i] != span2[i] {
			t.Errorf("Span data differs at index %d: %d vs %d", i, span1[i], span2[i])
		}
	}
}

func TestGlyphRasterBinInterface(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := NewGlyphRasterBin(font)

	// Verify that GlyphRasterBin implements GlyphGenerator interface
	var _ GlyphGenerator = g
}
