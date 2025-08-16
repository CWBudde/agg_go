package scanline

import (
	"agg_go/internal/basics"
	"testing"
)

// Mock implementations for testing

type MockScanline struct {
	y        int
	numSpans int
	spans    []SpanData
	spanIdx  int
}

func (m *MockScanline) Y() int                  { return m.y }
func (m *MockScanline) NumSpans() int           { return m.numSpans }
func (m *MockScanline) Begin() ScanlineIterator { m.spanIdx = 0; return m }

func (m *MockScanline) GetSpan() SpanData {
	if m.spanIdx < len(m.spans) {
		return m.spans[m.spanIdx]
	}
	return SpanData{}
}

func (m *MockScanline) Next() bool {
	m.spanIdx++
	return m.spanIdx < len(m.spans)
}

func (m *MockScanline) Reset(minX, maxX int) {
	// Mock reset implementation
}

type MockBaseRenderer struct {
	solidHspanCalls []SolidHspanCall
	hlineCalls      []HlineCall
	colorHspanCalls []ColorHspanCall
}

type SolidHspanCall struct {
	X, Y, Len int
	Color     interface{}
	Covers    []basics.Int8u
}

type HlineCall struct {
	X, Y, X2 int
	Color    interface{}
	Cover    basics.Int8u
}

type ColorHspanCall struct {
	X, Y, Len int
	Colors    []interface{}
	Covers    []basics.Int8u
	Cover     basics.Int8u
}

func (m *MockBaseRenderer) BlendSolidHspan(x, y, len int, color interface{}, covers []basics.Int8u) {
	m.solidHspanCalls = append(m.solidHspanCalls, SolidHspanCall{x, y, len, color, covers})
}

func (m *MockBaseRenderer) BlendHline(x, y, x2 int, color interface{}, cover basics.Int8u) {
	m.hlineCalls = append(m.hlineCalls, HlineCall{x, y, x2, color, cover})
}

func (m *MockBaseRenderer) BlendColorHspan(x, y, len int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
	m.colorHspanCalls = append(m.colorHspanCalls, ColorHspanCall{x, y, len, colors, covers, cover})
}

type MockRasterizer struct {
	minX, maxX   int
	scanlines    []*MockScanline
	currentIndex int
}

func (m *MockRasterizer) RewindScanlines() bool {
	m.currentIndex = 0
	return len(m.scanlines) > 0
}

func (m *MockRasterizer) SweepScanline(sl ScanlineInterface) bool {
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

func (m *MockRasterizer) MinX() int { return m.minX }
func (m *MockRasterizer) MaxX() int { return m.maxX }

type MockSpanAllocator struct {
	buffer []interface{}
}

func (m *MockSpanAllocator) Allocate(length int) []interface{} {
	if cap(m.buffer) < length {
		m.buffer = make([]interface{}, length)
	} else {
		m.buffer = m.buffer[:length]
	}
	return m.buffer
}

type MockSpanGenerator struct {
	prepareCalled bool
	color         interface{}
}

func (m *MockSpanGenerator) Prepare() {
	m.prepareCalled = true
}

func (m *MockSpanGenerator) Generate(colors []interface{}, x, y, len int) {
	for i := 0; i < len; i++ {
		colors[i] = m.color
	}
}

func TestRenderScanlineAASolid(t *testing.T) {
	renderer := &MockBaseRenderer{}
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
		renderer = &MockBaseRenderer{}
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
	renderer := &MockBaseRenderer{}
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

	rasterizer := &MockRasterizer{
		minX:      0,
		maxX:      10,
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
	renderer := &MockBaseRenderer{}
	allocator := &MockSpanAllocator{}
	generator := &MockSpanGenerator{color: "green"}

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
