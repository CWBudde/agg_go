package gsv

import (
	"agg_go/internal/basics"
	"agg_go/internal/transform"
	"math"
	"testing"
)

func TestNewGSVText(t *testing.T) {
	gsv := NewGSVText()

	if gsv == nil {
		t.Fatal("NewGSVText returned nil")
	}

	if gsv.width != 10.0 {
		t.Errorf("Expected default width 10.0, got %f", gsv.width)
	}

	if gsv.height != 0.0 {
		t.Errorf("Expected default height 0.0, got %f", gsv.height)
	}

	if gsv.status != StatusInitial {
		t.Errorf("Expected initial status %d, got %d", StatusInitial, gsv.status)
	}

	if len(gsv.font) == 0 {
		t.Error("Expected default font to be loaded")
	}
}

func TestGSVTextSetFont(t *testing.T) {
	gsv := NewGSVText()

	// Test setting custom font
	customFont := []byte{0x01, 0x02, 0x03, 0x04}
	gsv.SetFont(customFont)

	if len(gsv.font) != len(customFont) {
		t.Errorf("Expected font length %d, got %d", len(customFont), len(gsv.font))
	}

	// Test setting nil (should use default)
	gsv.SetFont(nil)

	if len(gsv.font) == 0 {
		t.Error("Setting nil font should use default font")
	}
}

func TestGSVTextConfiguration(t *testing.T) {
	gsv := NewGSVText()

	// Test size setting
	gsv.SetSize(20.0, 15.0)
	if gsv.height != 20.0 {
		t.Errorf("Expected height 20.0, got %f", gsv.height)
	}
	if gsv.width != 15.0 {
		t.Errorf("Expected width 15.0, got %f", gsv.width)
	}

	// Test spacing
	gsv.SetSpace(2.5)
	if gsv.space != 2.5 {
		t.Errorf("Expected space 2.5, got %f", gsv.space)
	}

	gsv.SetLineSpace(5.0)
	if gsv.lineSpace != 5.0 {
		t.Errorf("Expected line space 5.0, got %f", gsv.lineSpace)
	}

	// Test start point
	gsv.SetStartPoint(100.0, 200.0)
	if gsv.x != 100.0 {
		t.Errorf("Expected x 100.0, got %f", gsv.x)
	}
	if gsv.y != 200.0 {
		t.Errorf("Expected y 200.0, got %f", gsv.y)
	}
	if gsv.startX != 100.0 {
		t.Errorf("Expected startX 100.0, got %f", gsv.startX)
	}

	// Test flip
	gsv.SetFlip(true)
	if !gsv.flip {
		t.Error("Expected flip to be true")
	}

	// Test text
	gsv.SetText("Hello")
	if gsv.text != "Hello" {
		t.Errorf("Expected text 'Hello', got '%s'", gsv.text)
	}
}

func TestGSVTextRewind(t *testing.T) {
	gsv := NewGSVText()
	gsv.SetText("A")
	gsv.SetSize(16.0, 0.0) // Use default width
	gsv.SetStartPoint(10.0, 20.0)

	gsv.Rewind(0)

	if gsv.status != StatusInitial {
		t.Errorf("Expected status %d after rewind, got %d", StatusInitial, gsv.status)
	}

	if gsv.curChar != 0 {
		t.Errorf("Expected curChar 0 after rewind, got %d", gsv.curChar)
	}

	// Check that font parsing was initialized
	if len(gsv.indices) == 0 {
		t.Error("Expected indices to be initialized after rewind")
	}

	if len(gsv.glyphs) == 0 {
		t.Error("Expected glyphs to be initialized after rewind")
	}
}

func TestGSVTextBasicVertexGeneration(t *testing.T) {
	gsv := NewGSVText()
	gsv.SetText("A")
	gsv.SetSize(16.0, 0.0)
	gsv.SetStartPoint(0.0, 0.0)

	gsv.Rewind(0)

	// First vertex should be a move_to
	x, y, cmd := gsv.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be MoveTo (%d), got %d", basics.PathCmdMoveTo, cmd)
	}

	if x != 0.0 || y != 0.0 {
		t.Errorf("Expected first vertex at (0, 0), got (%f, %f)", x, y)
	}

	// Should generate more vertices for the 'A' character
	vertexCount := 1
	for {
		_, _, cmd := gsv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
		if vertexCount > 100 { // Prevent infinite loop
			break
		}
	}

	if vertexCount <= 1 {
		t.Error("Expected multiple vertices for character 'A'")
	}
}

func TestGSVTextMultipleCharacters(t *testing.T) {
	gsv := NewGSVText()
	gsv.SetText("AB")
	gsv.SetSize(16.0, 0.0)
	gsv.SetStartPoint(0.0, 0.0)
	gsv.SetSpace(2.0) // Add some character spacing

	gsv.Rewind(0)

	moveToCount := 0
	lineToCount := 0

	for {
		_, _, cmd := gsv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		} else if cmd == basics.PathCmdMoveTo {
			moveToCount++
		} else if cmd == basics.PathCmdLineTo {
			lineToCount++
		}
	}

	// Should have at least 2 move_to commands (one for each character)
	if moveToCount < 2 {
		t.Errorf("Expected at least 2 MoveTo commands, got %d", moveToCount)
	}

	// Should have some line_to commands
	if lineToCount == 0 {
		t.Error("Expected some LineTo commands")
	}
}

func TestGSVTextNewlineHandling(t *testing.T) {
	gsv := NewGSVText()
	gsv.SetText("A\nB")
	gsv.SetSize(16.0, 0.0)
	gsv.SetStartPoint(10.0, 20.0)
	gsv.SetLineSpace(4.0)

	gsv.Rewind(0)

	// Process vertices and track position changes
	positions := []struct{ x, y float64 }{}

	for {
		x, y, cmd := gsv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdMoveTo {
			positions = append(positions, struct{ x, y float64 }{x, y})
		}
	}

	if len(positions) < 2 {
		t.Fatal("Expected at least 2 MoveTo positions for A\\nB")
	}

	// First character should start at (10, 20)
	if positions[0].x != 10.0 || positions[0].y != 20.0 {
		t.Errorf("Expected first character at (10, 20), got (%f, %f)",
			positions[0].x, positions[0].y)
	}

	// Second character should be on next line - X coordinate may vary due to character spacing
	if math.Abs(positions[1].x-10.0) > 10.0 { // Allow some tolerance for character spacing
		t.Errorf("Expected second character X near 10.0, got %f", positions[1].x)
	}

	// Y should have changed (increased or decreased depending on flip) by roughly height + line spacing
	yDiff := math.Abs(positions[1].y - positions[0].y)
	expectedYDiff := 16.0 + 4.0              // height + line spacing
	if math.Abs(yDiff-expectedYDiff) > 5.0 { // Allow some tolerance
		t.Errorf("Expected Y difference around %f, got %f (positions: %f -> %f)",
			expectedYDiff, yDiff, positions[0].y, positions[1].y)
	}
}

func TestGSVTextWidth(t *testing.T) {
	gsv := NewGSVText()
	gsv.SetSize(16.0, 0.0)

	// Empty text should have zero width
	gsv.SetText("")
	width := gsv.TextWidth()
	if width != 0.0 {
		t.Errorf("Expected zero width for empty text, got %f", width)
	}

	// Single character should have some width
	gsv.SetText("A")
	width1 := gsv.TextWidth()
	if width1 <= 0.0 {
		t.Errorf("Expected positive width for single character, got %f", width1)
	}

	// Multiple characters should have greater width
	gsv.SetText("ABC")
	width3 := gsv.TextWidth()
	if width3 <= width1 {
		t.Errorf("Expected width of 'ABC' (%f) to be greater than 'A' (%f)", width3, width1)
	}

	// With spacing
	gsv.SetSpace(5.0)
	widthWithSpacing := gsv.TextWidth()
	if widthWithSpacing <= width3 {
		t.Errorf("Expected width with spacing (%f) to be greater than without (%f)",
			widthWithSpacing, width3)
	}
}

func TestGSVTextOutline(t *testing.T) {
	text := NewGSVText()
	text.SetText("A")
	text.SetSize(16.0, 0.0)

	outline := NewGSVTextOutline(text)

	if outline.text != text {
		t.Error("Expected outline to reference the provided text")
	}

	if outline.Width() != 1.0 {
		t.Errorf("Expected default width 1.0, got %f", outline.Width())
	}

	// Test width setting
	outline.SetWidth(2.5)
	if outline.Width() != 2.5 {
		t.Errorf("Expected width 2.5, got %f", outline.Width())
	}

	// Test transform
	trans := transform.NewTransAffineTranslation(10.0, 20.0)
	outline.SetTransform(trans)

	if outline.Transform() != trans {
		t.Error("Expected transform to be set")
	}
}

func TestGSVTextOutlineVertexGeneration(t *testing.T) {
	text := NewGSVText()
	text.SetText("A")
	text.SetSize(16.0, 0.0)
	text.SetStartPoint(0.0, 0.0)

	outline := NewGSVTextOutline(text)
	outline.SetWidth(2.0)

	outline.Rewind(0)

	vertexCount := 0
	for {
		_, _, cmd := outline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
		if vertexCount > 1000 { // Prevent infinite loop
			break
		}
	}

	// Stroked text should generate many more vertices than filled text
	if vertexCount <= 10 {
		t.Errorf("Expected many vertices for stroked text, got %d", vertexCount)
	}
}

func TestGSVTextOutlineWithTransform(t *testing.T) {
	text := NewGSVText()
	text.SetText("A")
	text.SetSize(16.0, 0.0)
	text.SetStartPoint(0.0, 0.0)

	// Create transform that translates by (10, 20)
	trans := transform.NewTransAffineTranslation(10.0, 20.0)
	outline := NewGSVTextOutlineWithTransform(text, trans)

	outline.Rewind(0)

	// First vertex should be transformed
	x, y, cmd := outline.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected MoveTo command, got %d", cmd)
	}

	// Due to stroked outline, exact coordinates are hard to predict,
	// but they should be offset by the transformation
	if x < 5.0 || y < 15.0 { // Rough bounds check
		t.Errorf("Expected transformed coordinates, got (%f, %f)", x, y)
	}
}

func TestGSVTextSimple(t *testing.T) {
	simple := NewGSVTextSimple()

	if simple == nil {
		t.Fatal("NewGSVTextSimple returned nil")
	}

	// Test configuration methods
	simple.SetText("Hello")
	simple.SetPosition(10.0, 20.0)
	simple.SetSize(16.0, 0.0)
	simple.SetStrokeWidth(1.5)
	simple.SetSpacing(1.0, 2.0)

	// Test filled vs stroked
	simple.SetFilled(true)
	simple.Rewind(0)

	filledVertexCount := 0
	for {
		_, _, cmd := simple.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		filledVertexCount++
		if filledVertexCount > 1000 {
			break
		}
	}

	simple.SetFilled(false)
	simple.Rewind(0)

	strokedVertexCount := 0
	for {
		_, _, cmd := simple.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		strokedVertexCount++
		if strokedVertexCount > 1000 {
			break
		}
	}

	// Stroked text should generate more vertices than filled text
	if strokedVertexCount <= filledVertexCount {
		t.Errorf("Expected stroked text (%d vertices) to have more than filled text (%d vertices)",
			strokedVertexCount, filledVertexCount)
	}

	// Test width estimation
	width := simple.EstimateWidth()
	if width <= 0.0 {
		t.Errorf("Expected positive estimated width, got %f", width)
	}
}

func TestGSVTextFontDataIntegrity(t *testing.T) {
	// Test that the embedded font data has the expected structure
	font := GSVDefaultFont

	if len(font) < 100 {
		t.Fatalf("Font data seems too small: %d bytes", len(font))
	}

	// Check that we can read basic header values
	gsv := NewGSVText()

	// Should not panic when reading font data
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Font data reading caused panic: %v", r)
		}
	}()

	gsv.SetSize(16.0, 0.0)
	gsv.SetText("ABC123!@#")
	gsv.Rewind(0)

	// Should generate vertices without error
	vertexCount := 0
	for {
		_, _, cmd := gsv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
		if vertexCount > 10000 { // Prevent runaway
			break
		}
	}

	if vertexCount == 0 {
		t.Error("Font data produced no vertices")
	}
}

func TestEndianness(t *testing.T) {
	// Test that endianness detection works
	bigEndian := isBigEndian()

	// This test mainly ensures the function doesn't panic
	// The actual value depends on the system architecture
	_ = bigEndian // Prevent unused variable warning
}

func BenchmarkGSVTextVertexGeneration(b *testing.B) {
	gsv := NewGSVText()
	gsv.SetText("The quick brown fox jumps over the lazy dog")
	gsv.SetSize(16.0, 0.0)
	gsv.SetStartPoint(0.0, 0.0)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gsv.Rewind(0)
		for {
			_, _, cmd := gsv.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkGSVTextOutlineVertexGeneration(b *testing.B) {
	text := NewGSVText()
	text.SetText("Hello World")
	text.SetSize(16.0, 0.0)
	text.SetStartPoint(0.0, 0.0)

	outline := NewGSVTextOutline(text)
	outline.SetWidth(1.0)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outline.Rewind(0)
		for {
			_, _, cmd := outline.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
