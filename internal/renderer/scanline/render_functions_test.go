package scanline

import (
	"testing"

	"agg_go/internal/basics"
)

// MockRasterizerWithScanlines extends MockRasterizer for render_functions testing
type MockRasterizerWithScanlines struct {
	MockRasterizer
	scanlines    []*MockScanline
	currentIndex int
}

func (m *MockRasterizerWithScanlines) RewindScanlines() bool {
	m.currentIndex = 0
	return len(m.scanlines) > 0
}

func (m *MockRasterizerWithScanlines) SweepScanline(sl ScanlineInterface) bool {
	if m.currentIndex >= len(m.scanlines) {
		return false
	}

	// Copy current scanline data to the provided scanline
	current := m.scanlines[m.currentIndex]
	if mockSl, ok := sl.(*MockScanline); ok {
		mockSl.y = current.y
		mockSl.numSpans = current.numSpans
		mockSl.spans = current.spans
	}

	m.currentIndex++
	return true
}

// MockSpanAllocatorWithBuffer extends MockSpanAllocator for render_functions testing
type MockSpanAllocatorWithBuffer[C any] struct {
	MockSpanAllocator[C]
	buffer []C
}

func (m *MockSpanAllocatorWithBuffer[C]) Allocate(length int) []C {
	if cap(m.buffer) < length {
		m.buffer = make([]C, length)
	} else {
		m.buffer = m.buffer[:length]
	}
	m.allocations = append(m.allocations, m.buffer)
	return m.buffer
}

// MockSpanGeneratorWithColor extends MockSpanGenerator for render_functions testing
type MockSpanGeneratorWithColor[C any] struct {
	MockSpanGenerator[C]
	color C
}

func (m *MockSpanGeneratorWithColor[C]) Generate(colors []C, x, y, len int) {
	m.generateCalls = append(m.generateCalls, GenerateCall[C]{
		X: x, Y: y, Len: len, Colors: colors,
	})
	for i := 0; i < len; i++ {
		colors[i] = m.color
	}
}

func TestRenderScanlineAASolid(t *testing.T) {
	renderer := &MockBaseRenderer[string]{}
	color := "red"

	t.Run("positive length span", func(t *testing.T) {
		scanline := &MockScanline{
			y:        10,
			numSpans: 1,
			spans: []SpanData{
				{X: 5, Len: 10, Covers: []basics.Int8u{255, 128, 64}},
			},
		}

		RenderScanlineAASolid(scanline, renderer, color)

		if len(renderer.solidHspanCalls) != 1 {
			t.Errorf("Expected 1 solid hspan call, got %d", len(renderer.solidHspanCalls))
		}

		call := renderer.solidHspanCalls[0]
		if call.X != 5 || call.Y != 10 || call.Len != 10 || call.Color != color {
			t.Errorf("Unexpected solid hspan call: %+v", call)
		}
	})

	t.Run("negative length span", func(t *testing.T) {
		renderer = &MockBaseRenderer[string]{}
		scanline := &MockScanline{
			y:        15,
			numSpans: 1,
			spans: []SpanData{
				{X: 10, Len: -5, Covers: []basics.Int8u{200}},
			},
		}

		RenderScanlineAASolid(scanline, renderer, color)

		if len(renderer.hlineCalls) != 1 {
			t.Errorf("Expected 1 hline call, got %d", len(renderer.hlineCalls))
		}

		call := renderer.hlineCalls[0]
		expectedEndX := 10 - (-5) - 1 // 10 + 5 - 1 = 14
		if call.X != 10 || call.Y != 15 || call.X2 != expectedEndX || call.Color != color || call.Cover != 200 {
			t.Errorf("Unexpected hline call: %+v", call)
		}
	})
}

func TestRenderScanlinesAASolid(t *testing.T) {
	renderer := &MockBaseRenderer[string]{}
	color := "blue"

	scanline1 := &MockScanline{
		y:        0,
		numSpans: 1,
		spans:    []SpanData{{X: 0, Len: 5, Covers: []basics.Int8u{255, 255, 255, 255, 255}}},
	}

	scanline2 := &MockScanline{
		y:        1,
		numSpans: 1,
		spans:    []SpanData{{X: 2, Len: 3, Covers: []basics.Int8u{128, 128, 128}}},
	}

	rasterizer := &MockRasterizerWithScanlines{
		MockRasterizer: MockRasterizer{
			minX: 0,
			maxX: 10,
		},
		scanlines: []*MockScanline{scanline1, scanline2},
	}

	scanline := &MockScanline{}

	RenderScanlinesAASolid(rasterizer, scanline, renderer, color)

	if len(renderer.solidHspanCalls) != 2 {
		t.Errorf("Expected 2 solid hspan calls, got %d", len(renderer.solidHspanCalls))
	}

	// Check first scanline
	call1 := renderer.solidHspanCalls[0]
	if call1.X != 0 || call1.Y != 0 || call1.Len != 5 || call1.Color != color {
		t.Errorf("Unexpected first solid hspan call: %+v", call1)
	}

	// Check second scanline
	call2 := renderer.solidHspanCalls[1]
	if call2.X != 2 || call2.Y != 1 || call2.Len != 3 || call2.Color != color {
		t.Errorf("Unexpected second solid hspan call: %+v", call2)
	}
}

func TestRenderScanlineAA(t *testing.T) {
	renderer := &MockBaseRenderer[string]{}
	allocator := &MockSpanAllocatorWithBuffer[string]{}
	generator := &MockSpanGeneratorWithColor[string]{color: "green"}

	scanline := &MockScanline{
		y:        20,
		numSpans: 1,
		spans: []SpanData{
			{X: 15, Len: 8, Covers: []basics.Int8u{255, 200, 150, 100, 50, 25, 12, 6}},
		},
	}

	RenderScanlineAA(scanline, renderer, allocator, generator)

	if len(renderer.colorHspanCalls) != 1 {
		t.Errorf("Expected 1 color hspan call, got %d", len(renderer.colorHspanCalls))
	}

	call := renderer.colorHspanCalls[0]
	if call.X != 15 || call.Y != 20 || call.Len != 8 {
		t.Errorf("Unexpected color hspan call: %+v", call)
	}

	// Check that colors were generated
	for i, color := range call.Colors {
		if color != "green" {
			t.Errorf("Expected color 'green' at index %d, got %v", i, color)
		}
	}
}

func TestRenderScanlineBinSolid(t *testing.T) {
	tests := []struct {
		name          string
		spans         []SpanData
		expectedCalls []HlineCall[string]
	}{
		{
			name: "positive_length_span",
			spans: []SpanData{
				{X: 10, Len: 5, Covers: nil}, // Binary spans don't use covers
			},
			expectedCalls: []HlineCall[string]{
				{X: 10, Y: 5, X2: 14, Color: "red", Cover: basics.CoverFull}, // X + Len - 1 = 10 + 5 - 1 = 14
			},
		},
		{
			name: "negative_length_span",
			spans: []SpanData{
				{X: 20, Len: -8, Covers: nil}, // Negative length
			},
			expectedCalls: []HlineCall[string]{
				{X: 20, Y: 5, X2: 27, Color: "red", Cover: basics.CoverFull}, // X - Len - 1 = 20 - (-8) - 1 = 20 + 8 - 1 = 27
			},
		},
		{
			name: "multiple_spans_mixed",
			spans: []SpanData{
				{X: 5, Len: 3, Covers: nil},   // Positive: endX = 5 + 3 - 1 = 7
				{X: 15, Len: -4, Covers: nil}, // Negative: endX = 15 - (-4) - 1 = 18
			},
			expectedCalls: []HlineCall[string]{
				{X: 5, Y: 5, X2: 7, Color: "red", Cover: basics.CoverFull},
				{X: 15, Y: 5, X2: 18, Color: "red", Cover: basics.CoverFull},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := &MockBaseRenderer[string]{}
			scanline := &MockScanline{
				y:        5,
				numSpans: len(tt.spans),
				spans:    tt.spans,
			}

			RenderScanlineBinSolid(scanline, renderer, "red")

			if len(renderer.hlineCalls) != len(tt.expectedCalls) {
				t.Errorf("Expected %d hline calls, got %d", len(tt.expectedCalls), len(renderer.hlineCalls))
			}

			for i, expected := range tt.expectedCalls {
				if i >= len(renderer.hlineCalls) {
					t.Errorf("Missing hline call %d", i)
					continue
				}
				actual := renderer.hlineCalls[i]
				if actual != expected {
					t.Errorf("Hline call %d: expected %+v, got %+v", i, expected, actual)
				}
			}
		})
	}
}

func TestRenderScanlinesBinSolid(t *testing.T) {
	renderer := &MockBaseRenderer[string]{}
	color := "blue"

	// Create test scanlines with negative lengths to verify the fix
	scanline1 := &MockScanline{
		y:        0,
		numSpans: 1,
		spans: []SpanData{
			{X: 10, Len: -5, Covers: nil}, // Should result in endX = 10 - (-5) - 1 = 14
		},
	}

	scanline2 := &MockScanline{
		y:        1,
		numSpans: 1,
		spans: []SpanData{
			{X: 20, Len: 3, Covers: nil}, // Should result in endX = 20 + 3 - 1 = 22
		},
	}

	rasterizer := &MockRasterizerWithScanlines{
		MockRasterizer: MockRasterizer{
			minX: 0,
			maxX: 30,
		},
		scanlines: []*MockScanline{scanline1, scanline2},
	}

	scanline := &MockScanline{}

	RenderScanlinesBinSolid(rasterizer, scanline, renderer, color)

	if len(renderer.hlineCalls) != 2 {
		t.Errorf("Expected 2 hline calls, got %d", len(renderer.hlineCalls))
	}

	// Check first scanline (negative length)
	call1 := renderer.hlineCalls[0]
	expected1 := HlineCall[string]{X: 10, Y: 0, X2: 14, Color: color, Cover: basics.CoverFull}
	if call1 != expected1 {
		t.Errorf("First hline call: expected %+v, got %+v", expected1, call1)
	}

	// Check second scanline (positive length)
	call2 := renderer.hlineCalls[1]
	expected2 := HlineCall[string]{X: 20, Y: 1, X2: 22, Color: color, Cover: basics.CoverFull}
	if call2 != expected2 {
		t.Errorf("Second hline call: expected %+v, got %+v", expected2, call2)
	}
}
