package renderer

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/fonts"
	"agg_go/internal/glyph"
)

// MockBaseRenderer implements BaseRendererInterface for testing
type MockBaseRenderer struct {
	blendHSpanCalls []BlendHSpanCall
	blendVSpanCalls []BlendVSpanCall
}

type BlendHSpanCall struct {
	X, Y, Len int
	Color     interface{}
	Covers    []basics.CoverType
}

type BlendVSpanCall struct {
	X, Y, Len int
	Color     interface{}
	Covers    []basics.CoverType
}

func NewMockBaseRenderer() *MockBaseRenderer {
	return &MockBaseRenderer{}
}

func (m *MockBaseRenderer) BlendSolidHspan(x, y, len int, c interface{}, covers []basics.CoverType) {
	// Make a copy of covers to avoid sharing slice references
	coversCopy := make([]basics.CoverType, len)
	copy(coversCopy, covers)
	m.blendHSpanCalls = append(m.blendHSpanCalls, BlendHSpanCall{
		X: x, Y: y, Len: len, Color: c, Covers: coversCopy,
	})
}

func (m *MockBaseRenderer) BlendSolidVspan(x, y, len int, c interface{}, covers []basics.CoverType) {
	// Make a copy of covers to avoid sharing slice references
	coversCopy := make([]basics.CoverType, len)
	copy(coversCopy, covers)
	m.blendVSpanCalls = append(m.blendVSpanCalls, BlendVSpanCall{
		X: x, Y: y, Len: len, Color: c, Covers: coversCopy,
	})
}

// MockScanlineRenderer implements ScanlineRendererInterface for testing
type MockScanlineRenderer struct {
	prepareCalls int
	renderCalls  []RenderCall
}

type RenderCall struct {
	Y        int
	NumSpans int
	Spans    []SpanData
}

type SpanData struct {
	X      int
	Len    int
	Covers []basics.CoverType
}

func NewMockScanlineRenderer() *MockScanlineRenderer {
	return &MockScanlineRenderer{}
}

func (m *MockScanlineRenderer) Prepare() {
	m.prepareCalls++
}

func (m *MockScanlineRenderer) Render(scanline ScanlineInterface) {
	call := RenderCall{
		Y:        scanline.Y(),
		NumSpans: scanline.NumSpans(),
		Spans:    []SpanData{},
	}

	iter := scanline.Begin()
	for iter.HasNext() {
		span := iter.Next()
		if span != nil {
			coversCopy := make([]basics.CoverType, len(span.Covers))
			copy(coversCopy, span.Covers)
			call.Spans = append(call.Spans, SpanData{
				X: span.X, Len: span.Len, Covers: coversCopy,
			})
		}
	}

	m.renderCalls = append(m.renderCalls, call)
}

func TestGlyphRasterBin(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	// Test font properties
	if g.Height() != 6 {
		t.Errorf("Expected height 6, got %f", g.Height())
	}
	if g.BaseLine() != 5 {
		t.Errorf("Expected baseline 5, got %f", g.BaseLine())
	}

	// Test width calculation
	width := g.Width("A")
	if width != 3 { // Based on our font data, 'A' has width 3
		t.Errorf("Expected width 3 for 'A', got %f", width)
	}

	// Test glyph preparation
	var rect glyph.GlyphRect
	g.Prepare(&rect, 10, 20, 'A', false)

	if rect.X1 != 10 {
		t.Errorf("Expected X1=10, got %d", rect.X1)
	}
	if rect.X2 != 12 { // X1 + width - 1
		t.Errorf("Expected X2=12, got %d", rect.X2)
	}
	if rect.DX != 3 {
		t.Errorf("Expected DX=3, got %f", rect.DX)
	}
	if rect.DY != 0 {
		t.Errorf("Expected DY=0, got %f", rect.DY)
	}

	// Test span data
	span := g.Span(0)
	if span == nil {
		t.Fatal("Expected non-nil span")
	}
	if len(span) != 3 {
		t.Errorf("Expected span length 3, got %d", len(span))
	}
}

func TestRendererRasterHTextSolid(t *testing.T) {
	mockRenderer := NewMockBaseRenderer()
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	renderer := NewRendererRasterHTextSolid[*MockBaseRenderer, *glyph.GlyphRasterBin](mockRenderer, g)

	// Set color
	color := "red"
	renderer.SetColor(color)
	if renderer.Color() != color {
		t.Errorf("Expected color %v, got %v", color, renderer.Color())
	}

	// Render text
	renderer.RenderText(5, 10, "A", false)

	// Check that BlendSolidHspan was called
	if len(mockRenderer.blendHSpanCalls) == 0 {
		t.Fatal("Expected BlendSolidHspan calls")
	}

	// Verify first call
	call := mockRenderer.blendHSpanCalls[0]
	if call.X != 5 {
		t.Errorf("Expected X=5, got %d", call.X)
	}
	if call.Color != color {
		t.Errorf("Expected color=%v, got %v", color, call.Color)
	}
	if call.Len != 3 { // Width of 'A'
		t.Errorf("Expected len=3, got %d", call.Len)
	}
}

func TestRendererRasterVTextSolid(t *testing.T) {
	mockRenderer := NewMockBaseRenderer()
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	renderer := NewRendererRasterVTextSolid[*MockBaseRenderer, *glyph.GlyphRasterBin](mockRenderer, g)

	// Set color
	color := "blue"
	renderer.SetColor(color)
	if renderer.Color() != color {
		t.Errorf("Expected color %v, got %v", color, renderer.Color())
	}

	// Render text
	renderer.RenderText(5, 10, "B", false)

	// Check that BlendSolidVspan was called
	if len(mockRenderer.blendVSpanCalls) == 0 {
		t.Fatal("Expected BlendSolidVspan calls")
	}

	// Verify first call
	call := mockRenderer.blendVSpanCalls[0]
	if call.Color != color {
		t.Errorf("Expected color=%v, got %v", color, call.Color)
	}
	if call.Len != 3 { // Width of 'B'
		t.Errorf("Expected len=3, got %d", call.Len)
	}
}

func TestRendererRasterHText(t *testing.T) {
	mockRenderer := NewMockScanlineRenderer()
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	renderer := NewRendererRasterHText[*MockScanlineRenderer, *glyph.GlyphRasterBin](mockRenderer, g)

	// Render text
	renderer.RenderText(5, 10, "C", false)

	// Check that Prepare was called
	if mockRenderer.prepareCalls == 0 {
		t.Error("Expected Prepare to be called")
	}

	// Check that Render was called
	if len(mockRenderer.renderCalls) == 0 {
		t.Fatal("Expected Render calls")
	}

	// Verify render call structure
	call := mockRenderer.renderCalls[0]
	if call.NumSpans != 1 {
		t.Errorf("Expected 1 span, got %d", call.NumSpans)
	}
	if len(call.Spans) != 1 {
		t.Errorf("Expected 1 span data, got %d", len(call.Spans))
	}

	span := call.Spans[0]
	if span.X != 5 {
		t.Errorf("Expected span X=5, got %d", span.X)
	}
	if span.Len != 3 { // Width of 'C'
		t.Errorf("Expected span len=3, got %d", span.Len)
	}
}

func TestScanlineSingleSpan(t *testing.T) {
	covers := []basics.CoverType{255, 128, 64}
	scanline := NewScanlineSingleSpan(10, 20, 3, covers)

	if scanline.Y() != 20 {
		t.Errorf("Expected Y=20, got %d", scanline.Y())
	}
	if scanline.NumSpans() != 1 {
		t.Errorf("Expected NumSpans=1, got %d", scanline.NumSpans())
	}

	iter := scanline.Begin()
	if !iter.HasNext() {
		t.Fatal("Expected iterator to have next")
	}

	span := iter.Next()
	if span == nil {
		t.Fatal("Expected non-nil span")
	}
	if span.X != 10 {
		t.Errorf("Expected span X=10, got %d", span.X)
	}
	if span.Len != 3 {
		t.Errorf("Expected span len=3, got %d", span.Len)
	}
	if len(span.Covers) != 3 {
		t.Errorf("Expected covers length=3, got %d", len(span.Covers))
	}

	// Check that iterator is exhausted
	if iter.HasNext() {
		t.Error("Expected iterator to be exhausted")
	}
	if iter.Next() != nil {
		t.Error("Expected Next() to return nil after exhaustion")
	}
}

func TestRendererRasterTextWithMultipleCharacters(t *testing.T) {
	mockRenderer := NewMockBaseRenderer()
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	renderer := NewRendererRasterHTextSolid[*MockBaseRenderer, *glyph.GlyphRasterBin](mockRenderer, g)
	renderer.SetColor("green")

	// Render multiple characters
	renderer.RenderText(0, 10, "AB", false)

	// Should have calls for both characters
	if len(mockRenderer.blendHSpanCalls) == 0 {
		t.Fatal("Expected BlendSolidHspan calls for multiple characters")
	}

	// The calls should be for different X positions due to character advance
	// First character at X=0, second character should be at X=3 (width of 'A')
	foundX0 := false
	foundX3 := false

	for _, call := range mockRenderer.blendHSpanCalls {
		if call.X == 0 {
			foundX0 = true
		}
		if call.X == 3 {
			foundX3 = true
		}
	}

	if !foundX0 {
		t.Error("Expected call at X=0 for first character")
	}
	if !foundX3 {
		t.Error("Expected call at X=3 for second character")
	}
}

func TestRendererRasterTextWithFlip(t *testing.T) {
	mockRenderer := NewMockBaseRenderer()
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	renderer := NewRendererRasterHTextSolid[*MockBaseRenderer, *glyph.GlyphRasterBin](mockRenderer, g)
	renderer.SetColor("purple")

	// Render with flip=true
	renderer.RenderText(5, 10, "A", true)

	// Should still have calls
	if len(mockRenderer.blendHSpanCalls) == 0 {
		t.Fatal("Expected BlendSolidHspan calls with flip")
	}

	// With flip, the span calculation should be different
	// This is tested implicitly by ensuring no panics occur
}

func TestGlyphRasterBinInvalidGlyph(t *testing.T) {
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	var rect glyph.GlyphRect

	// Test with glyph outside font range
	g.Prepare(&rect, 0, 0, rune(1000), false)

	// Should result in invalid rectangle
	if rect.X1 <= rect.X2 {
		t.Error("Expected invalid rectangle for out-of-range glyph")
	}
}

func TestGlyphRasterBinEmptyFont(t *testing.T) {
	g := glyph.NewGlyphRasterBin([]byte{})

	if g.Height() != 0 {
		t.Errorf("Expected height 0 for empty font, got %f", g.Height())
	}
	if g.BaseLine() != 0 {
		t.Errorf("Expected baseline 0 for empty font, got %f", g.BaseLine())
	}
	if g.Width("test") != 0 {
		t.Errorf("Expected width 0 for empty font, got %f", g.Width("test"))
	}
}

func TestRendererRasterVTextSolidAttach(t *testing.T) {
	mockRenderer1 := NewMockBaseRenderer()
	mockRenderer2 := NewMockBaseRenderer()
	font := fonts.GetSimple4x6Font()
	g := glyph.NewGlyphRasterBin(font)

	renderer := NewRendererRasterVTextSolid[*MockBaseRenderer, *glyph.GlyphRasterBin](mockRenderer1, g)

	// Test initial renderer
	renderer.SetColor("red")
	renderer.RenderText(0, 10, "A", false)

	if len(mockRenderer1.blendVSpanCalls) == 0 {
		t.Fatal("Expected calls to first renderer")
	}
	if len(mockRenderer2.blendVSpanCalls) > 0 {
		t.Fatal("Expected no calls to second renderer yet")
	}

	// Attach new renderer
	renderer.Attach(mockRenderer2)
	renderer.RenderText(5, 15, "B", false)

	// Should now use second renderer
	if len(mockRenderer2.blendVSpanCalls) == 0 {
		t.Fatal("Expected calls to second renderer after attach")
	}
}
