package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"testing"
)

// MockRGBAPatternSource implements RGBASourceInterface for pattern testing
type MockRGBAPatternSource struct {
	width  int
	height int
	data   []basics.Int8u // RGBA data (4 bytes per pixel)
	order  color.ColorOrder
	posX   int
	posY   int
}

// NewMockRGBAPatternSource creates a new mock RGBA pattern source with test data
func NewMockRGBAPatternSource(width, height int, order color.ColorOrder) *MockRGBAPatternSource {
	// Create test data with distinct RGBA values for easy verification
	data := make([]basics.Int8u, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			// Use position-based values for easy identification
			data[idx+order.R] = basics.Int8u((x*10 + y) % 256)         // R component
			data[idx+order.G] = basics.Int8u((x*20 + y*2) % 256)       // G component
			data[idx+order.B] = basics.Int8u((x*30 + y*3) % 256)       // B component
			data[idx+order.A] = basics.Int8u((x*40 + y*4 + 100) % 256) // A component (offset by 100 for distinction)
		}
	}
	return &MockRGBAPatternSource{
		width:  width,
		height: height,
		data:   data,
		order:  order,
	}
}

func (m *MockRGBAPatternSource) Width() int                  { return m.width }
func (m *MockRGBAPatternSource) Height() int                 { return m.height }
func (m *MockRGBAPatternSource) ColorType() string           { return "RGBA8" }
func (m *MockRGBAPatternSource) OrderType() color.ColorOrder { return m.order }

func (m *MockRGBAPatternSource) Span(x, y, length int) []basics.Int8u {
	m.posX = x
	m.posY = y

	// Bounds check
	if y >= m.height || x >= m.width {
		return []basics.Int8u{}
	}

	result := make([]basics.Int8u, 0, length*4)
	for i := 0; i < length && (x+i) < m.width; i++ {
		idx := (y*m.width + (x + i)) * 4
		if idx+3 < len(m.data) {
			result = append(result, m.data[idx:idx+4]...)
		}
	}
	return result
}

func (m *MockRGBAPatternSource) NextX() []basics.Int8u {
	m.posX++
	if m.posX >= m.width || m.posY >= m.height {
		return []basics.Int8u{}
	}
	idx := (m.posY*m.width + m.posX) * 4
	if idx+3 < len(m.data) {
		return m.data[idx : idx+4]
	}
	return []basics.Int8u{}
}

func (m *MockRGBAPatternSource) NextY() []basics.Int8u {
	m.posY++
	if m.posY >= m.height {
		return []basics.Int8u{}
	}
	idx := (m.posY*m.width + m.posX) * 4
	if idx+3 < len(m.data) {
		return m.data[idx : idx+4]
	}
	return []basics.Int8u{}
}

func (m *MockRGBAPatternSource) RowPtr(y int) []basics.Int8u {
	if y >= m.height {
		return []basics.Int8u{}
	}
	start := y * m.width * 4
	end := (y + 1) * m.width * 4
	if end > len(m.data) {
		end = len(m.data)
	}
	return m.data[start:end]
}

func TestNewSpanPatternRGBA(t *testing.T) {
	sp := NewSpanPatternRGBA[*MockRGBAPatternSource]()

	if sp.offsetX != 0 || sp.offsetY != 0 {
		t.Errorf("Expected default offsets (0,0), got (%d,%d)", sp.offsetX, sp.offsetY)
	}
	// Check that Alpha() returns 0 as per C++ implementation
	if sp.Alpha() != 0 {
		t.Errorf("Expected Alpha() to return 0, got %d", sp.Alpha())
	}
}

func TestNewSpanPatternRGBAWithParams(t *testing.T) {
	source := NewMockRGBAPatternSource(10, 10, color.OrderRGBA)
	sp := NewSpanPatternRGBAWithParams(source, 5, 3)

	if sp.source != source {
		t.Error("Source not set correctly")
	}
	if sp.offsetX != 5 || sp.offsetY != 3 {
		t.Errorf("Expected offsets (5,3), got (%d,%d)", sp.offsetX, sp.offsetY)
	}
}

func TestSpanPatternRGBAAttachAndSource(t *testing.T) {
	sp := NewSpanPatternRGBA[*MockRGBAPatternSource]()
	source := NewMockRGBAPatternSource(10, 10, color.OrderRGBA)

	sp.Attach(source)

	if sp.Source() != source {
		t.Error("Attach/Source methods not working correctly")
	}
}

func TestSpanPatternRGBAOffsets(t *testing.T) {
	sp := NewSpanPatternRGBA[*MockRGBAPatternSource]()

	sp.SetOffsetX(10)
	sp.SetOffsetY(20)

	if sp.OffsetX() != 10 {
		t.Errorf("Expected X offset 10, got %d", sp.OffsetX())
	}
	if sp.OffsetY() != 20 {
		t.Errorf("Expected Y offset 20, got %d", sp.OffsetY())
	}
}

func TestSpanPatternRGBAAlpha(t *testing.T) {
	sp := NewSpanPatternRGBA[*MockRGBAPatternSource]()

	// SetAlpha should be a no-op for RGBA patterns
	sp.SetAlpha(128)

	// Alpha() should always return 0
	if sp.Alpha() != 0 {
		t.Errorf("Expected Alpha() to return 0, got %d", sp.Alpha())
	}
}

func TestSpanPatternRGBAGenerate(t *testing.T) {
	source := NewMockRGBAPatternSource(5, 5, color.OrderRGBA)
	sp := NewSpanPatternRGBAWithParams(source, 0, 0)

	// Test generating a span
	span := make([]color.RGBA8[color.Linear], 3)
	sp.Generate(span, 0, 0, 3)

	// Verify the generated colors match expected values from mock source
	expected := []color.RGBA8[color.Linear]{
		{R: 0, G: 0, B: 0, A: 100},    // Position (0,0): R=0*10+0=0, G=0*20+0*2=0, B=0*30+0*3=0, A=0*40+0*4+100=100
		{R: 10, G: 20, B: 30, A: 140}, // Position (1,0): R=1*10+0=10, G=1*20+0*2=20, B=1*30+0*3=30, A=1*40+0*4+100=140
		{R: 20, G: 40, B: 60, A: 180}, // Position (2,0): R=2*10+0=20, G=2*20+0*2=40, B=2*30+0*3=60, A=2*40+0*4+100=180
	}

	for i, expectedColor := range expected {
		if span[i].R != expectedColor.R || span[i].G != expectedColor.G ||
			span[i].B != expectedColor.B || span[i].A != expectedColor.A {
			t.Errorf("Span[%d]: expected RGBA(%d,%d,%d,%d), got RGBA(%d,%d,%d,%d)",
				i, expectedColor.R, expectedColor.G, expectedColor.B, expectedColor.A,
				span[i].R, span[i].G, span[i].B, span[i].A)
		}
	}
}

func TestSpanPatternRGBAGenerateWithOffset(t *testing.T) {
	source := NewMockRGBAPatternSource(5, 5, color.OrderRGBA)
	sp := NewSpanPatternRGBAWithParams(source, 2, 1) // Offset by (2, 1)

	// Generate span at position (0, 0), but source should read from (2, 1) due to offset
	span := make([]color.RGBA8[color.Linear], 2)
	sp.Generate(span, 0, 0, 2)

	// Expected values should be from source position (2, 1) and (3, 1)
	expected := []color.RGBA8[color.Linear]{
		{R: 21, G: 42, B: 63, A: 184}, // Position (2,1): R=2*10+1=21, G=2*20+1*2=42, B=2*30+1*3=63, A=2*40+1*4+100=184
		{R: 31, G: 62, B: 93, A: 224}, // Position (3,1): R=3*10+1=31, G=3*20+1*2=62, B=3*30+1*3=93, A=3*40+1*4+100=224
	}

	for i, expectedColor := range expected {
		if span[i].R != expectedColor.R || span[i].G != expectedColor.G ||
			span[i].B != expectedColor.B || span[i].A != expectedColor.A {
			t.Errorf("Span[%d] with offset: expected RGBA(%d,%d,%d,%d), got RGBA(%d,%d,%d,%d)",
				i, expectedColor.R, expectedColor.G, expectedColor.B, expectedColor.A,
				span[i].R, span[i].G, span[i].B, span[i].A)
		}
	}
}

func TestSpanPatternRGBAGenerateWithBGRAOrder(t *testing.T) {
	source := NewMockRGBAPatternSource(3, 3, color.OrderBGRA)
	sp := NewSpanPatternRGBAWithParams(source, 0, 0)

	span := make([]color.RGBA8[color.Linear], 2)
	sp.Generate(span, 0, 0, 2)

	// With BGRA order, components should still come out correctly in RGBA8 output
	expected := []color.RGBA8[color.Linear]{
		{R: 0, G: 0, B: 0, A: 100},    // Position (0,0)
		{R: 10, G: 20, B: 30, A: 140}, // Position (1,0)
	}

	for i, expectedColor := range expected {
		if span[i].R != expectedColor.R || span[i].G != expectedColor.G ||
			span[i].B != expectedColor.B || span[i].A != expectedColor.A {
			t.Errorf("Span[%d] with BGRA order: expected RGBA(%d,%d,%d,%d), got RGBA(%d,%d,%d,%d)",
				i, expectedColor.R, expectedColor.G, expectedColor.B, expectedColor.A,
				span[i].R, span[i].G, span[i].B, span[i].A)
		}
	}
}

func TestSpanPatternRGBAGenerateWithARGBOrder(t *testing.T) {
	source := NewMockRGBAPatternSource(3, 3, color.OrderARGB)
	sp := NewSpanPatternRGBAWithParams(source, 0, 0)

	span := make([]color.RGBA8[color.Linear], 2)
	sp.Generate(span, 0, 0, 2)

	// With ARGB order, components should still come out correctly in RGBA8 output
	expected := []color.RGBA8[color.Linear]{
		{R: 0, G: 0, B: 0, A: 100},    // Position (0,0)
		{R: 10, G: 20, B: 30, A: 140}, // Position (1,0)
	}

	for i, expectedColor := range expected {
		if span[i].R != expectedColor.R || span[i].G != expectedColor.G ||
			span[i].B != expectedColor.B || span[i].A != expectedColor.A {
			t.Errorf("Span[%d] with ARGB order: expected RGBA(%d,%d,%d,%d), got RGBA(%d,%d,%d,%d)",
				i, expectedColor.R, expectedColor.G, expectedColor.B, expectedColor.A,
				span[i].R, span[i].G, span[i].B, span[i].A)
		}
	}
}

func TestSpanPatternRGBAGenerateEmptySpan(t *testing.T) {
	source := NewMockRGBAPatternSource(5, 5, color.OrderRGBA)
	sp := NewSpanPatternRGBAWithParams(source, 0, 0)

	// Test with empty span
	span := make([]color.RGBA8[color.Linear], 0)
	sp.Generate(span, 0, 0, 0)

	// Should not panic and span should remain empty
	if len(span) != 0 {
		t.Error("Empty span generation should not modify span")
	}
}

func TestSpanPatternRGBAGenerateOutOfBounds(t *testing.T) {
	source := NewMockRGBAPatternSource(2, 2, color.OrderRGBA)
	sp := NewSpanPatternRGBAWithParams(source, 0, 0)

	// Request span longer than source width
	span := make([]color.RGBA8[color.Linear], 5)
	sp.Generate(span, 0, 0, 5)

	// First two pixels should have valid data, rest should fall back to NextX behavior
	// This tests the bounds handling in the Generate method
	if span[0].R != 0 || span[0].G != 0 || span[0].B != 0 || span[0].A != 100 {
		t.Errorf("First pixel should be (0,0,0,100), got (%d,%d,%d,%d)",
			span[0].R, span[0].G, span[0].B, span[0].A)
	}
	if span[1].R != 10 || span[1].G != 20 || span[1].B != 30 || span[1].A != 140 {
		t.Errorf("Second pixel should be (10,20,30,140), got (%d,%d,%d,%d)",
			span[1].R, span[1].G, span[1].B, span[1].A)
	}
}

func TestSpanPatternRGBAPrepare(t *testing.T) {
	sp := NewSpanPatternRGBA[*MockRGBAPatternSource]()

	// Prepare should not panic and should be a no-op
	sp.Prepare()

	// Test passes if no panic occurs
}

func TestSpanPatternRGBAAlphaCompatibility(t *testing.T) {
	source := NewMockRGBAPatternSource(3, 3, color.OrderRGBA)
	sp := NewSpanPatternRGBAWithParams(source, 0, 0)

	// Test that alpha channel is preserved from source data
	span := make([]color.RGBA8[color.Linear], 3)
	sp.Generate(span, 0, 0, 3)

	// Verify alpha values are different (proving they come from source, not a global value)
	if span[0].A == span[1].A || span[1].A == span[2].A {
		t.Error("Alpha values should be different, proving they come from source data")
	}

	// Verify alpha values match expected pattern
	if span[0].A != 100 || span[1].A != 140 || span[2].A != 180 {
		t.Errorf("Alpha values don't match expected pattern: got %d, %d, %d",
			span[0].A, span[1].A, span[2].A)
	}
}

func BenchmarkSpanPatternRGBAGenerate(b *testing.B) {
	source := NewMockRGBAPatternSource(1000, 1000, color.OrderRGBA)
	sp := NewSpanPatternRGBAWithParams(source, 0, 0)
	span := make([]color.RGBA8[color.Linear], 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sp.Generate(span, i%900, (i/900)%900, 100)
	}
}
