package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"testing"
)

// MockRGBPatternSource implements RGBSourceInterface for pattern testing
type MockRGBPatternSource struct {
	width  int
	height int
	data   []basics.Int8u // RGB data (3 bytes per pixel)
	order  color.ColorOrder
	posX   int
	posY   int
}

// NewMockRGBPatternSource creates a new mock RGB pattern source with test data
func NewMockRGBPatternSource(width, height int, order color.ColorOrder) *MockRGBPatternSource {
	// Create test data with distinct RGB values for easy verification
	data := make([]basics.Int8u, width*height*3)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 3
			// Use position-based values for easy identification
			data[idx+order.R] = basics.Int8u((x*10 + y) % 256)   // R component
			data[idx+order.G] = basics.Int8u((x*20 + y*2) % 256) // G component
			data[idx+order.B] = basics.Int8u((x*30 + y*3) % 256) // B component
		}
	}
	return &MockRGBPatternSource{
		width:  width,
		height: height,
		data:   data,
		order:  order,
	}
}

func (m *MockRGBPatternSource) Width() int                  { return m.width }
func (m *MockRGBPatternSource) Height() int                 { return m.height }
func (m *MockRGBPatternSource) ColorType() string           { return "RGB8" }
func (m *MockRGBPatternSource) OrderType() color.ColorOrder { return m.order }

func (m *MockRGBPatternSource) Span(x, y, length int) []basics.Int8u {
	m.posX = x
	m.posY = y

	// Bounds check
	if y >= m.height || x >= m.width {
		return []basics.Int8u{}
	}

	result := make([]basics.Int8u, 0, length*3)
	for i := 0; i < length && (x+i) < m.width; i++ {
		idx := (y*m.width + (x + i)) * 3
		if idx+2 < len(m.data) {
			result = append(result, m.data[idx:idx+3]...)
		}
	}
	return result
}

func (m *MockRGBPatternSource) NextX() []basics.Int8u {
	m.posX++
	if m.posX >= m.width || m.posY >= m.height {
		return []basics.Int8u{}
	}
	idx := (m.posY*m.width + m.posX) * 3
	if idx+2 < len(m.data) {
		return m.data[idx : idx+3]
	}
	return []basics.Int8u{}
}

func (m *MockRGBPatternSource) NextY() []basics.Int8u {
	m.posY++
	if m.posY >= m.height {
		return []basics.Int8u{}
	}
	idx := (m.posY*m.width + m.posX) * 3
	if idx+2 < len(m.data) {
		return m.data[idx : idx+3]
	}
	return []basics.Int8u{}
}

func (m *MockRGBPatternSource) RowPtr(y int) []basics.Int8u {
	if y >= m.height {
		return []basics.Int8u{}
	}
	start := y * m.width * 3
	end := (y + 1) * m.width * 3
	if end > len(m.data) {
		end = len(m.data)
	}
	return m.data[start:end]
}

func TestNewSpanPatternRGB(t *testing.T) {
	sp := NewSpanPatternRGB[*MockRGBPatternSource]()

	if sp.alpha != 255 {
		t.Errorf("Expected default alpha 255, got %d", sp.alpha)
	}
	if sp.offsetX != 0 || sp.offsetY != 0 {
		t.Errorf("Expected default offsets (0,0), got (%d,%d)", sp.offsetX, sp.offsetY)
	}
}

func TestNewSpanPatternRGBWithParams(t *testing.T) {
	source := NewMockRGBPatternSource(10, 10, color.OrderRGB)
	sp := NewSpanPatternRGBWithParams(source, 5, 3)

	if sp.source != source {
		t.Error("Source not set correctly")
	}
	if sp.offsetX != 5 || sp.offsetY != 3 {
		t.Errorf("Expected offsets (5,3), got (%d,%d)", sp.offsetX, sp.offsetY)
	}
	if sp.alpha != 255 {
		t.Errorf("Expected default alpha 255, got %d", sp.alpha)
	}
}

func TestSpanPatternRGBAttachAndSource(t *testing.T) {
	sp := NewSpanPatternRGB[*MockRGBPatternSource]()
	source := NewMockRGBPatternSource(10, 10, color.OrderRGB)

	sp.Attach(source)

	if sp.Source() != source {
		t.Error("Attach/Source methods not working correctly")
	}
}

func TestSpanPatternRGBOffsets(t *testing.T) {
	sp := NewSpanPatternRGB[*MockRGBPatternSource]()

	sp.SetOffsetX(10)
	sp.SetOffsetY(20)

	if sp.OffsetX() != 10 {
		t.Errorf("Expected X offset 10, got %d", sp.OffsetX())
	}
	if sp.OffsetY() != 20 {
		t.Errorf("Expected Y offset 20, got %d", sp.OffsetY())
	}
}

func TestSpanPatternRGBAlpha(t *testing.T) {
	sp := NewSpanPatternRGB[*MockRGBPatternSource]()

	sp.SetAlpha(128)

	if sp.Alpha() != 128 {
		t.Errorf("Expected alpha 128, got %d", sp.Alpha())
	}
}

func TestSpanPatternRGBGenerate(t *testing.T) {
	source := NewMockRGBPatternSource(5, 5, color.OrderRGB)
	sp := NewSpanPatternRGBWithParams(source, 0, 0)

	// Test generating a span
	span := make([]color.RGB8[color.Linear], 3)
	sp.Generate(span, 0, 0, 3)

	// Verify the generated colors match expected values from mock source
	expected := []color.RGB8[color.Linear]{
		{R: 0, G: 0, B: 0},    // Position (0,0): R=0*10+0=0, G=0*20+0*2=0, B=0*30+0*3=0
		{R: 10, G: 20, B: 30}, // Position (1,0): R=1*10+0=10, G=1*20+0*2=20, B=1*30+0*3=30
		{R: 20, G: 40, B: 60}, // Position (2,0): R=2*10+0=20, G=2*20+0*2=40, B=2*30+0*3=60
	}

	for i, expectedColor := range expected {
		if span[i].R != expectedColor.R || span[i].G != expectedColor.G || span[i].B != expectedColor.B {
			t.Errorf("Span[%d]: expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
				i, expectedColor.R, expectedColor.G, expectedColor.B,
				span[i].R, span[i].G, span[i].B)
		}
	}
}

func TestSpanPatternRGBGenerateWithOffset(t *testing.T) {
	source := NewMockRGBPatternSource(5, 5, color.OrderRGB)
	sp := NewSpanPatternRGBWithParams(source, 2, 1) // Offset by (2, 1)

	// Generate span at position (0, 0), but source should read from (2, 1) due to offset
	span := make([]color.RGB8[color.Linear], 2)
	sp.Generate(span, 0, 0, 2)

	// Expected values should be from source position (2, 1) and (3, 1)
	expected := []color.RGB8[color.Linear]{
		{R: 21, G: 42, B: 63}, // Position (2,1): R=2*10+1=21, G=2*20+1*2=42, B=2*30+1*3=63
		{R: 31, G: 62, B: 93}, // Position (3,1): R=3*10+1=31, G=3*20+1*2=62, B=3*30+1*3=93
	}

	for i, expectedColor := range expected {
		if span[i].R != expectedColor.R || span[i].G != expectedColor.G || span[i].B != expectedColor.B {
			t.Errorf("Span[%d] with offset: expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
				i, expectedColor.R, expectedColor.G, expectedColor.B,
				span[i].R, span[i].G, span[i].B)
		}
	}
}

func TestSpanPatternRGBGenerateWithBGROrder(t *testing.T) {
	source := NewMockRGBPatternSource(3, 3, color.OrderBGR)
	sp := NewSpanPatternRGBWithParams(source, 0, 0)

	span := make([]color.RGB8[color.Linear], 2)
	sp.Generate(span, 0, 0, 2)

	// With BGR order, the R and B components should be swapped in the source data
	// but our RGB8 output should still have R, G, B in the correct positions
	expected := []color.RGB8[color.Linear]{
		{R: 0, G: 0, B: 0},    // Position (0,0)
		{R: 10, G: 20, B: 30}, // Position (1,0)
	}

	for i, expectedColor := range expected {
		if span[i].R != expectedColor.R || span[i].G != expectedColor.G || span[i].B != expectedColor.B {
			t.Errorf("Span[%d] with BGR order: expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
				i, expectedColor.R, expectedColor.G, expectedColor.B,
				span[i].R, span[i].G, span[i].B)
		}
	}
}

func TestSpanPatternRGBGenerateEmptySpan(t *testing.T) {
	source := NewMockRGBPatternSource(5, 5, color.OrderRGB)
	sp := NewSpanPatternRGBWithParams(source, 0, 0)

	// Test with empty span
	span := make([]color.RGB8[color.Linear], 0)
	sp.Generate(span, 0, 0, 0)

	// Should not panic and span should remain empty
	if len(span) != 0 {
		t.Error("Empty span generation should not modify span")
	}
}

func TestSpanPatternRGBGenerateOutOfBounds(t *testing.T) {
	source := NewMockRGBPatternSource(2, 2, color.OrderRGB)
	sp := NewSpanPatternRGBWithParams(source, 0, 0)

	// Request span longer than source width
	span := make([]color.RGB8[color.Linear], 5)
	sp.Generate(span, 0, 0, 5)

	// First two pixels should have valid data, rest should fall back to NextX behavior
	// This tests the bounds handling in the Generate method
	if span[0].R != 0 || span[0].G != 0 || span[0].B != 0 {
		t.Errorf("First pixel should be (0,0,0), got (%d,%d,%d)", span[0].R, span[0].G, span[0].B)
	}
	if span[1].R != 10 || span[1].G != 20 || span[1].B != 30 {
		t.Errorf("Second pixel should be (10,20,30), got (%d,%d,%d)", span[1].R, span[1].G, span[1].B)
	}
}

func TestSpanPatternRGBPrepare(t *testing.T) {
	sp := NewSpanPatternRGB[*MockRGBPatternSource]()

	// Prepare should not panic and should be a no-op
	sp.Prepare()

	// Test passes if no panic occurs
}

func BenchmarkSpanPatternRGBGenerate(b *testing.B) {
	source := NewMockRGBPatternSource(1000, 1000, color.OrderRGB)
	sp := NewSpanPatternRGBWithParams(source, 0, 0)
	span := make([]color.RGB8[color.Linear], 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sp.Generate(span, i%900, (i/900)%900, 100)
	}
}
