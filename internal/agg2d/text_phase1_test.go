package agg2d

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/font"
	"agg_go/internal/path"
)

type mockOutlineGlyph struct {
	glyphIndex uint
	advanceX   float64
	bounds     basics.Rect[int]
	buildPath  func(ps *path.PathStorageStl)
}

type mockTextFontEngine struct {
	glyphs       map[uint]mockOutlineGlyph
	kerning      map[[2]uint]float64
	current      mockOutlineGlyph
	currentValid bool
	pathStorage  *path.PathStorageStl
	lastKerning  [][2]uint
}

func newMockTextFontEngine() *mockTextFontEngine {
	return &mockTextFontEngine{
		glyphs:      make(map[uint]mockOutlineGlyph),
		kerning:     make(map[[2]uint]float64),
		pathStorage: path.NewPathStorageStl(),
	}
}

func (m *mockTextFontEngine) FontSignature() string { return "mock-text-font" }
func (m *mockTextFontEngine) ChangeStamp() int      { return 0 }

func (m *mockTextFontEngine) PrepareGlyph(glyphCode uint) bool {
	g, ok := m.glyphs[glyphCode]
	if !ok {
		m.currentValid = false
		return false
	}
	m.current = g
	m.currentValid = true
	m.pathStorage.RemoveAll()
	if g.buildPath != nil {
		g.buildPath(m.pathStorage)
	}
	return true
}

func (m *mockTextFontEngine) GlyphIndex() uint {
	if !m.currentValid {
		return 0
	}
	return m.current.glyphIndex
}

func (m *mockTextFontEngine) DataSize() uint {
	return 0
}

func (m *mockTextFontEngine) DataType() font.GlyphDataType {
	return font.GlyphDataOutline
}

func (m *mockTextFontEngine) Bounds() basics.Rect[int] {
	if !m.currentValid {
		return basics.Rect[int]{}
	}
	return m.current.bounds
}

func (m *mockTextFontEngine) AdvanceX() float64 {
	if !m.currentValid {
		return 0
	}
	return m.current.advanceX
}

func (m *mockTextFontEngine) AdvanceY() float64 {
	return 0
}

func (m *mockTextFontEngine) WriteGlyphTo(_ []byte) {}

func (m *mockTextFontEngine) AddKerning(first, second uint) (dx, dy float64) {
	m.lastKerning = append(m.lastKerning, [2]uint{first, second})
	if dx, ok := m.kerning[[2]uint{first, second}]; ok {
		return dx, 0
	}
	return 0, 0
}

func (m *mockTextFontEngine) PathAdaptor() *path.PathStorageStl {
	return m.pathStorage
}

func TestTextWidthUsesGlyphIndexKerning(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{
		glyphIndex: 100,
		advanceX:   10,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 2},
	}
	engine.glyphs[uint('V')] = mockOutlineGlyph{
		glyphIndex: 200,
		advanceX:   10,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 2},
	}
	engine.kerning[[2]uint{100, 200}] = -3

	agg2d := NewAgg2D()
	buf := make([]byte, 32*16*4)
	agg2d.Attach(buf, 32, 16, 32*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)

	got := agg2d.TextWidth("AV")
	if got != 17 {
		t.Fatalf("TextWidth(AV)=%v, want 17", got)
	}

	if len(engine.lastKerning) != 1 {
		t.Fatalf("expected one kerning call, got %d", len(engine.lastKerning))
	}
	if engine.lastKerning[0] != [2]uint{100, 200} {
		t.Fatalf("expected kerning pair [100 200], got %v", engine.lastKerning[0])
	}
}

func TestVectorTextUsesGlyphTranslation(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{
		glyphIndex: 100,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 2},
		buildPath: func(ps *path.PathStorageStl) {
			ps.MoveTo(0, 0)
			ps.LineTo(2, 0)
			ps.LineTo(2, 2)
			ps.LineTo(0, 2)
			ps.ClosePolygon(basics.PathFlagsNone)
		},
	}

	agg2d := NewAgg2D()
	width, height := 32, 16
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)
	agg2d.FillColor(Color{255, 0, 0, 255})
	agg2d.NoLine()
	agg2d.Text(8, 6, "A", false, 0, 0)

	_, _, _, aOrigin := pixelAt(buf, width, 0, 0)
	if aOrigin != 0 {
		t.Fatalf("expected origin to stay transparent, got alpha=%d", aOrigin)
	}

	_, _, _, aGlyph := pixelAt(buf, width, 8, 6)
	if aGlyph == 0 {
		t.Fatalf("expected translated glyph coverage at (8,6)")
	}
}

func TestVectorGlyphCacheHitRefreshesEngineOutlineState(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{
		glyphIndex: 100,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 2},
		buildPath: func(ps *path.PathStorageStl) {
			ps.MoveTo(0, 0)
			ps.LineTo(2, 0)
			ps.LineTo(2, 2)
			ps.LineTo(0, 2)
			ps.ClosePolygon(basics.PathFlagsNone)
		},
	}
	engine.glyphs[uint('V')] = mockOutlineGlyph{
		glyphIndex: 200,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 4, X2: 2, Y2: 6},
		buildPath: func(ps *path.PathStorageStl) {
			ps.MoveTo(0, 4)
			ps.LineTo(2, 4)
			ps.LineTo(2, 6)
			ps.LineTo(0, 6)
			ps.ClosePolygon(basics.PathFlagsNone)
		},
	}

	fcm := font.NewFontCacheManager(engine, 32)

	if glyph := fcm.Glyph(uint('A')); glyph == nil {
		t.Fatal("expected glyph A to load")
	}
	if engine.GlyphIndex() != 100 {
		t.Fatalf("expected engine state at glyph index 100 after A, got %d", engine.GlyphIndex())
	}

	if glyph := fcm.Glyph(uint('V')); glyph == nil {
		t.Fatal("expected glyph V to load")
	}
	if engine.GlyphIndex() != 200 {
		t.Fatalf("expected engine state at glyph index 200 after V, got %d", engine.GlyphIndex())
	}

	// A is now a cache hit; engine state must still be refreshed to A for outline adaptor parity.
	if glyph := fcm.Glyph(uint('A')); glyph == nil {
		t.Fatal("expected cached glyph A to load")
	}
	if engine.GlyphIndex() != 100 {
		t.Fatalf("expected engine state refreshed to glyph index 100 on cache hit, got %d", engine.GlyphIndex())
	}
}
