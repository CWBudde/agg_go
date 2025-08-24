package scanline

import (
	"agg_go/internal/basics"
)

// Shared mock types for all scanline tests

// MockScanline implements ScanlineInterface for testing
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

// MockResettableScanline extends MockScanline with reset tracking
type MockResettableScanline struct {
	MockScanline
	ResetCalled bool
	ResetMinX   int
	ResetMaxX   int
}

func (m *MockResettableScanline) Reset(minX, maxX int) {
	m.ResetCalled = true
	m.ResetMinX = minX
	m.ResetMaxX = maxX
}

// MockBaseRenderer implements BaseRendererInterface for testing
type MockBaseRenderer[C any] struct {
	solidHspanCalls []SolidHspanCall[C]
	hlineCalls      []HlineCall[C]
	colorHspanCalls []ColorHspanCall[C]
}

type SolidHspanCall[C any] struct {
	X, Y, Len int
	Color     C
	Covers    []basics.Int8u
}

type HlineCall[C any] struct {
	X, Y, X2 int
	Color    C
	Cover    basics.Int8u
}

type ColorHspanCall[C any] struct {
	X, Y, Len int
	Colors    []C
	Covers    []basics.Int8u
	Cover     basics.Int8u
}

func (m *MockBaseRenderer[C]) BlendSolidHspan(x, y, len int, color C, covers []basics.Int8u) {
	m.solidHspanCalls = append(m.solidHspanCalls, SolidHspanCall[C]{
		X: x, Y: y, Len: len, Color: color, Covers: covers,
	})
}

func (m *MockBaseRenderer[C]) BlendHline(x, y, x2 int, color C, cover basics.Int8u) {
	m.hlineCalls = append(m.hlineCalls, HlineCall[C]{
		X: x, Y: y, X2: x2, Color: color, Cover: cover,
	})
}

func (m *MockBaseRenderer[C]) BlendColorHspan(x, y, len int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	m.colorHspanCalls = append(m.colorHspanCalls, ColorHspanCall[C]{
		X: x, Y: y, Len: len, Colors: colors, Covers: covers, Cover: cover,
	})
}

// MockSpanAllocator for testing span-based renderers
type MockSpanAllocator[C any] struct {
	allocations [][]C
}

func (m *MockSpanAllocator[C]) Allocate(length int) []C {
	colors := make([]C, length)
	m.allocations = append(m.allocations, colors)
	return colors
}

// MockSpanGenerator for testing span-based renderers
type MockSpanGenerator[C any] struct {
	prepareCalled bool
	generateCalls []GenerateCall[C]
}

type GenerateCall[C any] struct {
	X, Y, Len int
	Colors    []C
}

func (m *MockSpanGenerator[C]) Prepare() {
	m.prepareCalled = true
}

func (m *MockSpanGenerator[C]) Generate(colors []C, x, y, len int) {
	m.generateCalls = append(m.generateCalls, GenerateCall[C]{
		X: x, Y: y, Len: len, Colors: colors,
	})
	// Fill with some test data
	for i := 0; i < len; i++ {
		var testColor C
		colors[i] = testColor
	}
}

// MockRasterizer for testing helper functions
type MockRasterizer struct {
	rewindResult   bool
	sweepResults   []bool
	sweepCallCount int
	minX, maxX     int
	resetCalled    bool
}

func (m *MockRasterizer) RewindScanlines() bool {
	return m.rewindResult
}

func (m *MockRasterizer) SweepScanline(sl ScanlineInterface) bool {
	if m.sweepCallCount < len(m.sweepResults) {
		result := m.sweepResults[m.sweepCallCount]
		m.sweepCallCount++
		return result
	}
	return false
}

func (m *MockRasterizer) MinX() int { return m.minX }
func (m *MockRasterizer) MaxX() int { return m.maxX }
func (m *MockRasterizer) Reset()    { m.resetCalled = true }

// MockRenderer for testing helper functions
type MockRenderer[C any] struct {
	prepareCalled bool
	renderCalls   []ScanlineInterface
	color         C
}

func (m *MockRenderer[C]) Prepare() {
	m.prepareCalled = true
}

func (m *MockRenderer[C]) Render(sl ScanlineInterface) {
	m.renderCalls = append(m.renderCalls, sl)
}

func (m *MockRenderer[C]) SetColor(color C) {
	m.color = color
}

// MockPathColorStorage for testing RenderAllPaths
type MockPathColorStorage[C any] struct {
	colors     []C
	defaultVal C
}

func (m *MockPathColorStorage[C]) GetColor(index int) C {
	if index >= 0 && index < len(m.colors) {
		return m.colors[index]
	}
	return m.defaultVal
}

// MockPathIdStorage for testing RenderAllPaths
type MockPathIdStorage struct {
	pathIds []int
}

func (m *MockPathIdStorage) GetPathId(index int) int {
	if index >= 0 && index < len(m.pathIds) {
		return m.pathIds[index]
	}
	return 0
}

// MockVertexSource for testing RenderAllPaths
type MockVertexSource struct{}

// MockCompoundRasterizer for testing compound rendering
type MockCompoundRasterizer struct {
	MockRasterizer
	sweepStylesResults []int
	sweepStylesIndex   int
	styleResults       [][]bool // [styleIndex][sweepIndex]
	styleSweepCounts   []int    // separate sweep count per style
	styles             []int
}

func (m *MockCompoundRasterizer) SweepStyles() int {
	if m.sweepStylesIndex < len(m.sweepStylesResults) {
		result := m.sweepStylesResults[m.sweepStylesIndex]
		m.sweepStylesIndex++
		return result
	}
	return 0
}

func (m *MockCompoundRasterizer) SweepScanlineWithStyle(sl ScanlineInterface, styleId int) bool {
	if styleId == -1 {
		// Binary scanline sweep - use regular sweep count
		return m.SweepScanline(sl)
	}
	if styleId >= 0 && styleId < len(m.styleResults) {
		// Ensure styleSweepCounts is initialized
		if len(m.styleSweepCounts) <= styleId {
			m.styleSweepCounts = make([]int, styleId+1)
		}

		results := m.styleResults[styleId]
		if m.styleSweepCounts[styleId] < len(results) {
			result := results[m.styleSweepCounts[styleId]]
			m.styleSweepCounts[styleId]++

			// Copy scanline data for AA scanline
			if mockSl, ok := sl.(*MockScanline); ok && len(m.sweepResults) > 0 {
				mockSl.y = 8
				mockSl.numSpans = 1
				mockSl.spans = []SpanData{{X: 5, Len: 2, Covers: []basics.Int8u{255, 128}}}
			}

			return result
		}
	}
	return false
}

func (m *MockCompoundRasterizer) Style(index int) int {
	if index >= 0 && index < len(m.styles) {
		return m.styles[index]
	}
	return 0
}

func (m *MockCompoundRasterizer) ScanlineStart() int  { return m.minX }
func (m *MockCompoundRasterizer) ScanlineLength() int { return m.maxX - m.minX + 1 }
func (m *MockCompoundRasterizer) AllocateCoverBuffer(len int) []basics.Int8u {
	return make([]basics.Int8u, len)
}

// MockStyleHandler for testing compound rendering
type MockStyleHandler[C any] struct {
	solidFlags    []bool
	colors        []C
	generateCalls []GenerateSpanCall[C]
}

type GenerateSpanCall[C any] struct {
	Colors    []C
	X, Y, Len int
	Style     int
}

func (m *MockStyleHandler[C]) IsSolid(style int) bool {
	// Style can be an ID like 200, not necessarily an index
	// For testing, map common style IDs to flags
	switch style {
	case 100: // from single solid test
		return true
	case 200: // from single generated test
		return false
	case 300, 301: // from multiple styles test
		return true
	default:
		if style >= 0 && style < len(m.solidFlags) {
			return m.solidFlags[style]
		}
		return true
	}
}

func (m *MockStyleHandler[C]) Color(style int) C {
	// Map style IDs to colors for testing
	switch style {
	case 100:
		if len(m.colors) > 0 {
			return m.colors[0]
		}
	case 200:
		if len(m.colors) > 1 {
			return m.colors[1]
		}
	case 300:
		if len(m.colors) > 2 {
			return m.colors[2]
		}
	case 301:
		if len(m.colors) > 3 {
			return m.colors[3]
		}
	default:
		if style >= 0 && style < len(m.colors) {
			return m.colors[style]
		}
	}
	var zero C
	return zero
}

func (m *MockStyleHandler[C]) GenerateSpan(colors []C, x, y, len, style int) {
	m.generateCalls = append(m.generateCalls, GenerateSpanCall[C]{
		Colors: colors,
		X:      x, Y: y, Len: len,
		Style: style,
	})
}
