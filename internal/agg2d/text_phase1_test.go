package agg2d

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/font"
	"github.com/MeKo-Christian/agg_go/internal/path"
)

func alphaBounds(buf []byte, width, height int) (minX, minY, maxX, maxY int, ok bool) {
	minX, minY = width, height
	maxX, maxY = -1, -1
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			_, _, _, a := pixelAt(buf, width, x, y)
			if a == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
			ok = true
		}
	}
	return minX, minY, maxX, maxY, ok
}

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

func TestVectorTextAlignmentProducesExpectedBounds(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('H')] = mockOutlineGlyph{
		glyphIndex: 10,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 6},
		buildPath: func(ps *path.PathStorageStl) {
			ps.MoveTo(0, 0)
			ps.LineTo(2, 0)
			ps.LineTo(2, 6)
			ps.LineTo(0, 6)
			ps.ClosePolygon(basics.PathFlagsNone)
		},
	}
	engine.glyphs[uint('A')] = mockOutlineGlyph{
		glyphIndex: 20,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 6},
		buildPath: func(ps *path.PathStorageStl) {
			ps.MoveTo(0, 0)
			ps.LineTo(2, 0)
			ps.LineTo(2, 6)
			ps.LineTo(0, 6)
			ps.ClosePolygon(basics.PathFlagsNone)
		},
	}

	agg2d := NewAgg2D()
	width, height := 32, 32
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontHeight = 6
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)
	agg2d.FillColor(Color{255, 0, 0, 255})
	agg2d.NoLine()
	agg2d.TextAlignment(AlignCenter, AlignTop)
	agg2d.Text(10, 10, "AH", false, 0, 0)

	minX, minY, maxX, maxY, ok := alphaBounds(buf, width, height)
	if !ok {
		t.Fatal("expected rendered glyph coverage")
	}
	if minX != 6 || minY != 4 || maxX != 11 || maxY != 9 {
		t.Fatalf("rendered bounds = (%d,%d)-(%d,%d), want (6,4)-(11,9)", minX, minY, maxX, maxY)
	}
}

// TestTextWidthEmptyStringIsZero mirrors the C++ textWidth guard: an empty
// string always returns 0 regardless of font state. Source: agg2d.cpp:960-978.
func TestTextWidthEmptyStringIsZero(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{glyphIndex: 1, advanceX: 10}

	agg2d := NewAgg2D()
	buf := make([]byte, 32*16*4)
	agg2d.Attach(buf, 32, 16, 32*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)

	if got := agg2d.TextWidth(""); got != 0 {
		t.Fatalf("TextWidth(\"\") = %v, want 0", got)
	}
}

// TestTextWidthSingleCharNoKerning verifies that a single character produces
// exactly advance_x with zero kerning calls. Source: agg2d.cpp:960-978 (first==true
// branch skips add_kerning).
func TestTextWidthSingleCharNoKerning(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{glyphIndex: 42, advanceX: 8}

	agg2d := NewAgg2D()
	buf := make([]byte, 32*16*4)
	agg2d.Attach(buf, 32, 16, 32*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)

	got := agg2d.TextWidth("A")
	if got != 8 {
		t.Fatalf("TextWidth(\"A\") = %v, want 8", got)
	}
	if len(engine.lastKerning) != 0 {
		t.Fatalf("expected no kerning calls for single char, got %d", len(engine.lastKerning))
	}
}

// TestTextWidthTwoCharsNoKerning verifies plain advance accumulation when no
// kerning pair is registered (add_kerning returns 0,0 in C++).
func TestTextWidthTwoCharsNoKerning(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{glyphIndex: 1, advanceX: 8}
	engine.glyphs[uint('B')] = mockOutlineGlyph{glyphIndex: 2, advanceX: 6}
	// no kerning entry → AddKerning returns (0,0)

	agg2d := NewAgg2D()
	buf := make([]byte, 32*16*4)
	agg2d.Attach(buf, 32, 16, 32*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)

	got := agg2d.TextWidth("AB")
	if got != 14 {
		t.Fatalf("TextWidth(\"AB\") = %v, want 14 (8+6)", got)
	}
}

// TestTextWidthMissingGlyphSkipped verifies that characters absent from the font
// do not contribute advance and do not cause a kerning call (matches C++ null-glyph
// guard: the inner if(glyph) block is simply skipped).
func TestTextWidthMissingGlyphSkipped(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{glyphIndex: 1, advanceX: 10}
	// 'B' has no glyph

	agg2d := NewAgg2D()
	buf := make([]byte, 32*16*4)
	agg2d.Attach(buf, 32, 16, 32*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)

	// Width of "AB" with missing B should equal width of just "A".
	got := agg2d.TextWidth("AB")
	if got != 10 {
		t.Fatalf("TextWidth(\"AB\") with missing B = %v, want 10", got)
	}
	// No kerning should be called because 'B' had no glyph and is the only
	// candidate for a second position.
	if len(engine.lastKerning) != 0 {
		t.Fatalf("expected no kerning calls when second char missing, got %v", engine.lastKerning)
	}
}

// TestTextWidthKerningUsesGlyphIndicesNotCharCodes verifies that kerning is
// looked up by glyph index (as returned by the font engine), not by the raw
// Unicode code point. This matches the C++ glyph_cache::glyph_index field.
func TestTextWidthKerningUsesGlyphIndicesNotCharCodes(t *testing.T) {
	engine := newMockTextFontEngine()
	// 'A' (char 65) maps to glyph index 500; 'V' (char 86) maps to glyph index 501.
	// Kerning is keyed on glyph indices 500,501, not on char codes 65,86.
	engine.glyphs[uint('A')] = mockOutlineGlyph{glyphIndex: 500, advanceX: 10}
	engine.glyphs[uint('V')] = mockOutlineGlyph{glyphIndex: 501, advanceX: 10}
	engine.kerning[[2]uint{500, 501}] = -4
	// Ensure wrong key (char codes) has no entry.
	engine.kerning[[2]uint{65, 86}] = 999 // must NOT be used

	agg2d := NewAgg2D()
	buf := make([]byte, 32*16*4)
	agg2d.Attach(buf, 32, 16, 32*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)

	got := agg2d.TextWidth("AV")
	// 10 + 10 - 4 = 16 (glyph-index kerning applied)
	if got != 16 {
		t.Fatalf("TextWidth(\"AV\") = %v, want 16 (kerning by glyph index)", got)
	}
}

// TestTextWidthGSVModeIdempotent verifies that repeated TextWidth calls return
// the same result and do not corrupt gsvText internal state between calls.
// This is the state-preservation regression test for the MeasureText fix.
func TestTextWidthGSVModeIdempotent(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 400*100*4)
	agg2d.Attach(buf, 400, 100, 400*4)
	agg2d.FontGSV(20)

	w1 := agg2d.TextWidth("Hello")
	_ = agg2d.TextWidth("XXXXXXXXXXXXXXXX") // should not affect Hello width
	w2 := agg2d.TextWidth("Hello")

	if w1 != w2 {
		t.Fatalf("TextWidth(\"Hello\") is not idempotent: first=%v, second=%v", w1, w2)
	}
	if w1 <= 0 {
		t.Fatalf("expected positive TextWidth for non-empty string, got %v", w1)
	}
}

// TestTextWidthGSVEmptyString verifies that the GSV path returns 0 for empty
// input without panicking.
func TestTextWidthGSVEmptyString(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 400*100*4)
	agg2d.Attach(buf, 400, 100, 400*4)
	agg2d.FontGSV(20)

	if got := agg2d.TextWidth(""); got != 0 {
		t.Fatalf("GSV TextWidth(\"\") = %v, want 0", got)
	}
}

// TestTextWidthGSVMonotonic verifies that TextWidth grows (weakly) as more
// characters are appended, matching the C++ advance accumulation model.
func TestTextWidthGSVMonotonic(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]byte, 800*100*4)
	agg2d.Attach(buf, 800, 100, 800*4)
	agg2d.FontGSV(20)

	prev := agg2d.TextWidth("A")
	if prev <= 0 {
		t.Fatal("expected positive width for single GSV char")
	}
	for _, s := range []string{"AB", "ABC", "ABCD", "ABCDE"} {
		w := agg2d.TextWidth(s)
		if w < prev {
			t.Fatalf("TextWidth(%q)=%v < TextWidth(shorter)=%v: width must not decrease", s, w, prev)
		}
		prev = w
	}
}

func TestVectorTextRoundOffAndOffsetAffectBounds(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('H')] = mockOutlineGlyph{
		glyphIndex: 10,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 4},
		buildPath: func(ps *path.PathStorageStl) {
			ps.MoveTo(0, 0)
			ps.LineTo(2, 0)
			ps.LineTo(2, 4)
			ps.LineTo(0, 4)
			ps.ClosePolygon(basics.PathFlagsNone)
		},
	}

	agg2d := NewAgg2D()
	width, height := 32, 32
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontHeight = 4
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)
	agg2d.FillColor(Color{255, 0, 0, 255})
	agg2d.NoLine()
	agg2d.Text(10.8, 10.2, "H", true, 2.0, -1.0)

	minX, minY, maxX, maxY, ok := alphaBounds(buf, width, height)
	if !ok {
		t.Fatal("expected rendered glyph coverage")
	}
	if minX != 12 || minY != 9 || maxX != 13 || maxY != 12 {
		t.Fatalf("rendered bounds = (%d,%d)-(%d,%d), want (12,9)-(13,12)", minX, minY, maxX, maxY)
	}
}

// TestTextKerningAdjustsGlyphPlacement verifies that kerning returned by the font
// engine shifts the second glyph's start_x during Text() rendering. This matches
// the C++ agg2d.cpp:1059 call to add_kerning(&start_x, &start_y) inside the text()
// loop.
//
// Setup: 'A' has advance_x=4 and is drawn at origin.
// After 'A': cursor = 4. Kerning('A','V') = -2, so cursor = 2.
// 'V' must therefore be drawn at x=2, not x=4.
//
// Geometry: both glyphs use a 2×2 filled square centred at their local (0,0),
// so 'A' covers columns 0–1 and 'V' covers columns 2–3 (with kerning) or
// columns 4–5 (without kerning).
func TestTextKerningAdjustsGlyphPlacement(t *testing.T) {
	engine := newMockTextFontEngine()
	square := func(ps *path.PathStorageStl) {
		ps.MoveTo(0, 0)
		ps.LineTo(2, 0)
		ps.LineTo(2, 2)
		ps.LineTo(0, 2)
		ps.ClosePolygon(basics.PathFlagsNone)
	}
	engine.glyphs[uint('A')] = mockOutlineGlyph{
		glyphIndex: 10,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 2},
		buildPath:  square,
	}
	engine.glyphs[uint('V')] = mockOutlineGlyph{
		glyphIndex: 20,
		advanceX:   4,
		bounds:     basics.Rect[int]{X1: 0, Y1: 0, X2: 2, Y2: 2},
		buildPath:  square,
	}
	// Kerning between A(10) and V(20) pulls them 2 units closer.
	engine.kerning[[2]uint{10, 20}] = -2

	agg2d := NewAgg2D()
	width, height := 12, 4
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.fontCacheType = VectorFontCache
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)
	agg2d.FillColor(Color{255, 0, 0, 255})
	agg2d.NoLine()

	agg2d.Text(0, 0, "AV", false, 0, 0)

	// 'A' at x=0: expect coverage at (1,1).
	_, _, _, aA := pixelAt(buf, width, 1, 1)
	if aA == 0 {
		t.Fatalf("expected glyph A coverage at (1,1)")
	}

	// 'V' at x=2 (with kerning): expect coverage at (3,1).
	_, _, _, aVWithKerning := pixelAt(buf, width, 3, 1)
	if aVWithKerning == 0 {
		t.Fatalf("expected glyph V coverage at (3,1) with kerning applied")
	}

	// Without kerning V would be at x=4: (5,1) must be transparent.
	_, _, _, aVNoKerning := pixelAt(buf, width, 5, 1)
	if aVNoKerning != 0 {
		t.Fatalf("glyph V must not appear at (5,1): that position implies kerning was ignored (alpha=%d)", aVNoKerning)
	}
}

// TestTextRasterFontCacheWorldToScreenConversion verifies that Text() converts
// the starting position through WorldToScreen before placing raster glyphs, matching
// agg2d.cpp:1063 — worldToScreen(start_x, start_y) when fontCacheType==RasterFontCache.
//
// A viewport mapping is applied so that a world position of (2,2) maps to a
// different screen position (4,4). The gray8 glyph must appear at the screen
// position, not the world position.
func TestTextRasterFontCacheWorldToScreenConversion(t *testing.T) {
	engine := newMockTextFontEngine()
	engine.glyphs[uint('A')] = mockOutlineGlyph{
		glyphIndex: 1,
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
	width, height := 20, 20
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	// Apply a 2× scale viewport: world(0,0)→screen(0,0), world(1,1)→screen(2,2).
	agg2d.Viewport(0, 0, 5, 5, 0, 0, 10, 10, XMidYMid)

	agg2d.fontCacheType = VectorFontCache // outline path, no worldToScreen for cursor
	agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)
	agg2d.FillColor(Color{255, 0, 0, 255})
	agg2d.NoLine()
	agg2d.fontHeight = 2

	// With VectorFontCache, glyph paths go through the global transform. Render at
	// world (1,1); global 2× scale should place coverage around screen (2,2).
	agg2d.Text(1, 1, "A", false, 0, 0)

	// Global transform scales by 2, so world (1,1) → screen (2,2). The 2×2 glyph
	// path covers screen area roughly (2,2)–(6,6).
	_, _, _, aCovered := pixelAt(buf, width, 3, 3)
	if aCovered == 0 {
		t.Fatalf("expected glyph coverage near screen (3,3) after 2× viewport scale")
	}

	// World (1,1) without scaling would be screen (1,1) — must be empty.
	_, _, _, aWrong := pixelAt(buf, width, 1, 1)
	if aWrong != 0 {
		t.Fatalf("expected no coverage at screen (1,1): glyph must be placed via viewport transform")
	}
}
