package span

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// MockGrayPatternSource implements GraySourceInterface for testing span patterns.
type MockGrayPatternSource struct {
	width   int
	height  int
	data    []basics.Int8u
	curX    int
	curY    int
	spanPtr int // Current position in span
}

// NewMockGrayPatternSource creates a new mock gray source for pattern testing.
func NewMockGrayPatternSource(width, height int) *MockGrayPatternSource {
	data := make([]basics.Int8u, width*height)
	// Fill with a simple gradient pattern for testing
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a checkerboard-like pattern
			value := basics.Int8u((x + y) % 256)
			data[y*width+x] = value
		}
	}

	return &MockGrayPatternSource{
		width:  width,
		height: height,
		data:   data,
	}
}

// Width returns the source width.
func (m *MockGrayPatternSource) Width() int {
	return m.width
}

// Height returns the source height.
func (m *MockGrayPatternSource) Height() int {
	return m.height
}

// ColorType returns the color type string.
func (m *MockGrayPatternSource) ColorType() string {
	return "gray8"
}

// Span returns pixel data starting at (x, y) with given length.
func (m *MockGrayPatternSource) Span(x, y, length int) []basics.Int8u {
	// Handle coordinates that might be outside bounds (pattern behavior)
	safeX := x % m.width
	if safeX < 0 {
		safeX += m.width
	}
	safeY := y % m.height
	if safeY < 0 {
		safeY += m.height
	}

	m.curX = safeX
	m.curY = safeY
	m.spanPtr = 0

	result := make([]basics.Int8u, length)
	for i := 0; i < length; i++ {
		// Wrap around horizontally within the source bounds
		pixelX := (safeX + i) % m.width
		if pixelX >= 0 && pixelX < m.width && safeY >= 0 && safeY < m.height {
			result[i] = m.data[safeY*m.width+pixelX]
		} else {
			result[i] = 0
		}
	}

	return result
}

// NextX advances to the next pixel in current span.
func (m *MockGrayPatternSource) NextX() []basics.Int8u {
	m.curX++
	if m.curX >= m.width {
		m.curX = 0
	}

	if m.curY >= 0 && m.curY < m.height && m.curX >= 0 && m.curX < m.width {
		return []basics.Int8u{m.data[m.curY*m.width+m.curX]}
	}
	return []basics.Int8u{0}
}

// NextY advances to the next row at original x position.
func (m *MockGrayPatternSource) NextY() []basics.Int8u {
	m.curY++
	if m.curY >= m.height {
		m.curY = 0
	}

	if m.curY >= 0 && m.curY < m.height && m.curX >= 0 && m.curX < m.width {
		return []basics.Int8u{m.data[m.curY*m.width+m.curX]}
	}
	return []basics.Int8u{0}
}

// RowPtr returns pointer to row data.
func (m *MockGrayPatternSource) RowPtr(y int) []basics.Int8u {
	safeY := y % m.height
	if safeY < 0 {
		safeY += m.height
	}

	if safeY >= 0 && safeY < m.height {
		start := safeY * m.width
		return m.data[start : start+m.width]
	}
	return make([]basics.Int8u, m.width)
}

func TestSpanPatternGray_NewSpanPatternGray(t *testing.T) {
	sp := NewSpanPatternGray[*MockGrayPatternSource]()

	if sp == nil {
		t.Fatal("NewSpanPatternGray should return a non-nil pointer")
	}

	// Check default alpha value
	if sp.Alpha() != 255 {
		t.Errorf("Expected default alpha 255, got %d", sp.Alpha())
	}

	// Check default offsets
	if sp.OffsetX() != 0 {
		t.Errorf("Expected default offsetX 0, got %d", sp.OffsetX())
	}
	if sp.OffsetY() != 0 {
		t.Errorf("Expected default offsetY 0, got %d", sp.OffsetY())
	}
}

func TestSpanPatternGray_NewSpanPatternGrayWithParams(t *testing.T) {
	source := NewMockGrayPatternSource(8, 8)
	sp := NewSpanPatternGrayWithParams(source, 10, 20)

	if sp == nil {
		t.Fatal("NewSpanPatternGrayWithParams should return a non-nil pointer")
	}

	// Check parameters were set correctly
	if sp.OffsetX() != 10 {
		t.Errorf("Expected offsetX 10, got %d", sp.OffsetX())
	}
	if sp.OffsetY() != 20 {
		t.Errorf("Expected offsetY 20, got %d", sp.OffsetY())
	}
	if sp.Alpha() != 255 {
		t.Errorf("Expected alpha 255, got %d", sp.Alpha())
	}

	// Verify source was attached
	attachedSource := sp.Source()
	if attachedSource.Width() != 8 || attachedSource.Height() != 8 {
		t.Errorf("Source not properly attached")
	}
}

func TestSpanPatternGray_AttachAndSource(t *testing.T) {
	sp := NewSpanPatternGray[*MockGrayPatternSource]()
	source := NewMockGrayPatternSource(16, 16)

	sp.Attach(source)

	attachedSource := sp.Source()
	if attachedSource.Width() != 16 || attachedSource.Height() != 16 {
		t.Errorf("Expected attached source with dimensions 16x16, got %dx%d",
			attachedSource.Width(), attachedSource.Height())
	}
}

func TestSpanPatternGray_OffsetAccessors(t *testing.T) {
	sp := NewSpanPatternGray[*MockGrayPatternSource]()

	// Test X offset
	sp.SetOffsetX(42)
	if sp.OffsetX() != 42 {
		t.Errorf("Expected offsetX 42, got %d", sp.OffsetX())
	}

	// Test Y offset
	sp.SetOffsetY(84)
	if sp.OffsetY() != 84 {
		t.Errorf("Expected offsetY 84, got %d", sp.OffsetY())
	}
}

func TestSpanPatternGray_AlphaAccessors(t *testing.T) {
	sp := NewSpanPatternGray[*MockGrayPatternSource]()

	// Test alpha setting
	sp.SetAlpha(128)
	if sp.Alpha() != 128 {
		t.Errorf("Expected alpha 128, got %d", sp.Alpha())
	}

	// Test alpha boundary values
	sp.SetAlpha(0)
	if sp.Alpha() != 0 {
		t.Errorf("Expected alpha 0, got %d", sp.Alpha())
	}

	sp.SetAlpha(255)
	if sp.Alpha() != 255 {
		t.Errorf("Expected alpha 255, got %d", sp.Alpha())
	}
}

func TestSpanPatternGray_Prepare(t *testing.T) {
	sp := NewSpanPatternGray[*MockGrayPatternSource]()

	// Prepare should not panic and should be a no-op
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Prepare() should not panic, but it did: %v", r)
		}
	}()

	sp.Prepare()
}

func TestSpanPatternGray_Generate(t *testing.T) {
	source := NewMockGrayPatternSource(4, 4)
	sp := NewSpanPatternGrayWithParams(source, 0, 0)
	sp.SetAlpha(200)

	// Test basic generation
	span := make([]color.Gray8[color.Linear], 4)
	sp.Generate(span, 0, 0, 4)

	// Verify the span was generated
	for i := 0; i < 4; i++ {
		if span[i].A != 200 {
			t.Errorf("Expected alpha 200 at position %d, got %d", i, span[i].A)
		}
		// The grayscale value should come from the source pattern
		// For our test pattern: value = (x + y) % 256
		expectedValue := basics.Int8u((i + 0) % 256) // y=0
		if span[i].V != expectedValue {
			t.Errorf("Expected gray value %d at position %d, got %d",
				expectedValue, i, span[i].V)
		}
	}
}

func TestSpanPatternGray_GenerateWithOffset(t *testing.T) {
	source := NewMockGrayPatternSource(8, 8)
	sp := NewSpanPatternGrayWithParams(source, 2, 3) // Offset by (2, 3)
	sp.SetAlpha(255)

	span := make([]color.Gray8[color.Linear], 4)
	sp.Generate(span, 1, 1, 4) // Start at (1, 1), will read from (3, 4) in source

	// Verify the offset was applied
	for i := 0; i < 4; i++ {
		if span[i].A != 255 {
			t.Errorf("Expected alpha 255 at position %d, got %d", i, span[i].A)
		}
		// For our test pattern: value = (x + y) % 256
		// With offset (2, 3) and start (1, 1), we read from ((1+2)+i, 1+3) = (3+i, 4)
		expectedValue := basics.Int8u(((3 + i) + 4) % 256)
		if span[i].V != expectedValue {
			t.Errorf("Expected gray value %d at position %d, got %d",
				expectedValue, i, span[i].V)
		}
	}
}

func TestSpanPatternGray_GenerateWithZeroAlpha(t *testing.T) {
	source := NewMockGrayPatternSource(4, 4)
	sp := NewSpanPatternGrayWithParams(source, 0, 0)
	sp.SetAlpha(0) // Fully transparent

	span := make([]color.Gray8[color.Linear], 3)
	sp.Generate(span, 0, 0, 3)

	// All pixels should have alpha 0
	for i := 0; i < 3; i++ {
		if span[i].A != 0 {
			t.Errorf("Expected alpha 0 at position %d, got %d", i, span[i].A)
		}
		// Gray values should still come from source
		expectedValue := basics.Int8u((i + 0) % 256)
		if span[i].V != expectedValue {
			t.Errorf("Expected gray value %d at position %d, got %d",
				expectedValue, i, span[i].V)
		}
	}
}

func TestSpanPatternGray_GenerateEmptySpan(t *testing.T) {
	source := NewMockGrayPatternSource(4, 4)
	sp := NewSpanPatternGrayWithParams(source, 0, 0)

	// Test with zero length
	span := make([]color.Gray8[color.Linear], 0)
	sp.Generate(span, 0, 0, 0)

	// Should not panic and span should remain empty
	if len(span) != 0 {
		t.Errorf("Expected empty span to remain empty")
	}
}

func TestSpanPatternGray_GenerateLargeSpan(t *testing.T) {
	source := NewMockGrayPatternSource(4, 4)
	sp := NewSpanPatternGrayWithParams(source, 0, 0)
	sp.SetAlpha(100)

	// Test with span longer than source width (should wrap/repeat)
	span := make([]color.Gray8[color.Linear], 10)
	sp.Generate(span, 0, 0, 10)

	// Verify all pixels have correct alpha
	for i := 0; i < 10; i++ {
		if span[i].A != 100 {
			t.Errorf("Expected alpha 100 at position %d, got %d", i, span[i].A)
		}
	}

	// First 4 values should match pattern, then it should wrap
	for i := 0; i < 4; i++ {
		expectedValue := basics.Int8u((i + 0) % 256)
		if span[i].V != expectedValue {
			t.Errorf("Expected gray value %d at position %d, got %d",
				expectedValue, i, span[i].V)
		}
	}
}

// Benchmark tests
func BenchmarkSpanPatternGray_Generate(b *testing.B) {
	source := NewMockGrayPatternSource(256, 256)
	sp := NewSpanPatternGrayWithParams(source, 0, 0)
	span := make([]color.Gray8[color.Linear], 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sp.Generate(span, 0, 0, 256)
	}
}

func BenchmarkSpanPatternGray_GenerateWithOffset(b *testing.B) {
	source := NewMockGrayPatternSource(256, 256)
	sp := NewSpanPatternGrayWithParams(source, 100, 150)
	span := make([]color.Gray8[color.Linear], 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sp.Generate(span, 50, 75, 256)
	}
}
