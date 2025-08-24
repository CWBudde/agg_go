package fonts

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
)

// MockLoadedFace implements LoadedFace for testing
type MockLoadedFace struct {
	height          float64
	width           float64
	ascent          float64
	descent         float64
	ascentB         float64
	descentB        float64
	glyphData       map[uint32]mockFaceGlyphData
	currentInstance mockInstance
}

type mockInstance struct {
	height    float64
	width     float64
	hinting   bool
	rendering FmanGlyphRendering
}

type mockFaceGlyphData struct {
	index    uint32
	data     []byte
	dataType FmanGlyphDataType
	bounds   basics.Rect[int]
	advanceX float64
	advanceY float64
}

func NewMockLoadedFace() *MockLoadedFace {
	return &MockLoadedFace{
		height:    16.0,
		width:     8.0,
		ascent:    12.0,
		descent:   4.0,
		ascentB:   13.0,
		descentB:  3.0,
		glyphData: make(map[uint32]mockFaceGlyphData),
	}
}

func (m *MockLoadedFace) AddGlyph(code uint32, index uint32, dataType FmanGlyphDataType,
	bounds basics.Rect[int], advanceX, advanceY float64, data []byte,
) {
	m.glyphData[code] = mockFaceGlyphData{
		index:    index,
		data:     data,
		dataType: dataType,
		bounds:   bounds,
		advanceX: advanceX,
		advanceY: advanceY,
	}
}

func (m *MockLoadedFace) Height() float64   { return m.height }
func (m *MockLoadedFace) Width() float64    { return m.width }
func (m *MockLoadedFace) Ascent() float64   { return m.ascent }
func (m *MockLoadedFace) Descent() float64  { return m.descent }
func (m *MockLoadedFace) AscentB() float64  { return m.ascentB }
func (m *MockLoadedFace) DescentB() float64 { return m.descentB }

func (m *MockLoadedFace) SelectInstance(height, width float64, hinting bool, rendering FmanGlyphRendering) {
	m.currentInstance = mockInstance{
		height:    height,
		width:     width,
		hinting:   hinting,
		rendering: rendering,
	}
}

func (m *MockLoadedFace) PrepareGlyph(cp uint32) (PreparedGlyph, bool) {
	if glyph, exists := m.glyphData[cp]; exists {
		return PreparedGlyph{
			GlyphCode:  cp,
			GlyphIndex: glyph.index,
			DataSize:   uint32(len(glyph.data)),
			DataType:   glyph.dataType,
			Bounds:     glyph.bounds,
			AdvanceX:   glyph.advanceX,
			AdvanceY:   glyph.advanceY,
		}, true
	}
	return PreparedGlyph{}, false
}

func (m *MockLoadedFace) WriteGlyphTo(prepared *PreparedGlyph, data []byte) {
	if glyph, exists := m.glyphData[prepared.GlyphCode]; exists {
		if len(data) >= len(glyph.data) {
			copy(data, glyph.data)
		}
	}
}

func (m *MockLoadedFace) AddKerning(first, second uint32, x, y *float64) bool {
	// Simple mock kerning: add 1.5 to x for any pair
	if first != 0 && second != 0 {
		*x += 1.5
		return true
	}
	return false
}

// MockFontEngine2 implements FontEngine2 for testing
type MockFontEngine2 struct {
	pathAdaptor   *MockPathAdaptor
	gray8Adaptor  *MockGray8Adaptor
	gray8Scanline *MockGray8Scanline
	monoAdaptor   *MockMonoAdaptor
	monoScanline  *MockMonoScanline
}

// Separate mock types for each interface
type MockPathAdaptor struct {
	data        []byte
	x, y        float64
	scale       float64
	initialized bool
}

type MockGray8Adaptor struct {
	data        []byte
	dataSize    uint32
	x, y        float64
	initialized bool
	bounds      basics.Rect[int]
	numSpans    uint
}

type MockMonoAdaptor struct {
	data        []byte
	dataSize    uint32
	x, y        float64
	initialized bool
	bounds      basics.Rect[int]
	numSpans    uint
}

type MockGray8Scanline struct {
	initialized bool
	minX, maxX  int
	y           int
	numSpans    uint
}

type MockMonoScanline struct {
	initialized bool
	minX, maxX  int
	y           int
	numSpans    uint
}

// MockSpanIterator implements both Gray8SpanIterator and MonoSpanIterator
type MockSpanIterator struct {
	valid  bool
	x      int
	length int
	covers []uint8
}

func (msi *MockSpanIterator) Next()           { msi.valid = false }
func (msi *MockSpanIterator) IsValid() bool   { return msi.valid }
func (msi *MockSpanIterator) X() int          { return msi.x }
func (msi *MockSpanIterator) Len() int        { return msi.length }
func (msi *MockSpanIterator) Covers() []uint8 { return msi.covers }

// PathAdaptorType methods for MockPathAdaptor
func (mpa *MockPathAdaptor) Init(data []byte, dx, dy, scale float64, coordShift int) {
	mpa.data = data
	mpa.x = dx
	mpa.y = dy
	mpa.scale = scale
	mpa.initialized = true
}

func (mpa *MockPathAdaptor) InitWithScale(data []byte, dataSize uint32, x, y, scale float64) {
	mpa.data = data
	mpa.x = x
	mpa.y = y
	mpa.scale = scale
	mpa.initialized = true
}

func (mpa *MockPathAdaptor) Rewind(pathID uint) {}

func (mpa *MockPathAdaptor) Vertex(x, y *float64) uint {
	*x = mpa.x
	*y = mpa.y
	return 0 // PathCmdStop
}

// Gray8AdaptorType methods for MockGray8Adaptor
func (mga *MockGray8Adaptor) InitGlyph(data []byte, dataSize uint32, x, y float64) {
	mga.data = data
	mga.dataSize = dataSize
	mga.x = x
	mga.y = y
	mga.initialized = true
}

func (mga *MockGray8Adaptor) Bounds() basics.Rect[int] { return mga.bounds }
func (mga *MockGray8Adaptor) Rewind(pathID uint)       {}
func (mga *MockGray8Adaptor) SweepScanline() bool      { return false }
func (mga *MockGray8Adaptor) NumSpans() uint           { return mga.numSpans }

func (mga *MockGray8Adaptor) Begin() Gray8SpanIterator {
	return &MockSpanIterator{valid: true, x: int(mga.x), length: 10, covers: []uint8{255}}
}

// MonoAdaptorType methods for MockMonoAdaptor
func (mma *MockMonoAdaptor) InitGlyph(data []byte, dataSize uint32, x, y float64) {
	mma.data = data
	mma.dataSize = dataSize
	mma.x = x
	mma.y = y
	mma.initialized = true
}

func (mma *MockMonoAdaptor) Bounds() basics.Rect[int] { return mma.bounds }
func (mma *MockMonoAdaptor) Rewind(pathID uint)       {}
func (mma *MockMonoAdaptor) SweepScanline() bool      { return false }
func (mma *MockMonoAdaptor) NumSpans() uint           { return mma.numSpans }

func (mma *MockMonoAdaptor) Begin() MonoSpanIterator {
	return &MockSpanIterator{valid: true, x: int(mma.x), length: 10}
}

// Gray8ScanlineType methods for MockGray8Scanline
func (mgs *MockGray8Scanline) Reset(minX, maxX int) {
	mgs.minX = minX
	mgs.maxX = maxX
	mgs.initialized = true
}

func (mgs *MockGray8Scanline) Y() int         { return mgs.y }
func (mgs *MockGray8Scanline) NumSpans() uint { return mgs.numSpans }

func (mgs *MockGray8Scanline) Begin() Gray8SpanIterator {
	return &MockSpanIterator{valid: true, x: mgs.minX, length: 10, covers: []uint8{255}}
}

// MonoScanlineType methods for MockMonoScanline
func (mms *MockMonoScanline) Reset(minX, maxX int) {
	mms.minX = minX
	mms.maxX = maxX
	mms.initialized = true
}

func (mms *MockMonoScanline) Y() int         { return mms.y }
func (mms *MockMonoScanline) NumSpans() uint { return mms.numSpans }

func (mms *MockMonoScanline) Begin() MonoSpanIterator {
	return &MockSpanIterator{valid: true, x: mms.minX, length: 10}
}

func NewMockFontEngine2() *MockFontEngine2 {
	return &MockFontEngine2{
		pathAdaptor:   &MockPathAdaptor{},
		gray8Adaptor:  &MockGray8Adaptor{},
		gray8Scanline: &MockGray8Scanline{},
		monoAdaptor:   &MockMonoAdaptor{},
		monoScanline:  &MockMonoScanline{},
	}
}

func (m *MockFontEngine2) PathAdaptor() PathAdaptorType     { return m.pathAdaptor }
func (m *MockFontEngine2) Gray8Adaptor() Gray8AdaptorType   { return m.gray8Adaptor }
func (m *MockFontEngine2) Gray8Scanline() Gray8ScanlineType { return m.gray8Scanline }
func (m *MockFontEngine2) MonoAdaptor() MonoAdaptorType     { return m.monoAdaptor }
func (m *MockFontEngine2) MonoScanline() MonoScanlineType   { return m.monoScanline }

func TestFmanGlyphDataType_String(t *testing.T) {
	tests := []struct {
		gdt      FmanGlyphDataType
		expected string
	}{
		{FmanGlyphDataInvalid, "invalid"},
		{FmanGlyphDataMono, "mono"},
		{FmanGlyphDataGray8, "gray8"},
		{FmanGlyphDataOutline, "outline"},
		{FmanGlyphDataType(999), "unknown"},
	}

	for _, test := range tests {
		result := test.gdt.String()
		if result != test.expected {
			t.Errorf("FmanGlyphDataType(%d).String() = %q, expected %q", test.gdt, result, test.expected)
		}
	}
}

func TestFmanGlyphRendering_String(t *testing.T) {
	tests := []struct {
		gr       FmanGlyphRendering
		expected string
	}{
		{FmanGlyphRenNativeMono, "native_mono"},
		{FmanGlyphRenNativeGray8, "native_gray8"},
		{FmanGlyphRenOutline, "outline"},
		{FmanGlyphRenAggMono, "agg_mono"},
		{FmanGlyphRenAggGray8, "agg_gray8"},
		{FmanGlyphRendering(999), "unknown"},
	}

	for _, test := range tests {
		result := test.gr.String()
		if result != test.expected {
			t.Errorf("FmanGlyphRendering(%d).String() = %q, expected %q", test.gr, result, test.expected)
		}
	}
}

func TestFmanCachedGlyphs_NewAndBasics(t *testing.T) {
	cg := NewFmanCachedGlyphs()
	if cg == nil {
		t.Fatal("NewFmanCachedGlyphs() returned nil")
	}

	cg2 := NewFmanCachedGlyphsWithBlockSize(8192)
	if cg2 == nil {
		t.Fatal("NewFmanCachedGlyphsWithBlockSize() returned nil")
	}
}

func TestFmanCachedGlyphs_CacheAndFind(t *testing.T) {
	cg := NewFmanCachedGlyphs()

	// Test finding non-existent glyph
	if glyph := cg.FindGlyph(999); glyph != nil {
		t.Error("FindGlyph returned non-nil for non-existent glyph")
	}

	// Cache a glyph
	glyphCode := uint32(65) // 'A'
	glyphIndex := uint32(42)
	dataSize := uint32(64)
	bounds := basics.Rect[int]{X1: 10, Y1: 20, X2: 30, Y2: 40}
	advanceX := 12.5
	advanceY := 0.0

	// Create a mock cached font
	mockFace := NewMockLoadedFace()
	cachedFont := NewFmanCachedFont(mockFace, 16.0, 8.0, false, FmanGlyphRenAggGray8)

	glyph := cg.CacheGlyph(cachedFont, glyphCode, glyphIndex, dataSize, FmanGlyphDataGray8, bounds, advanceX, advanceY)
	if glyph == nil {
		t.Fatal("CacheGlyph returned nil")
	}

	// Verify glyph properties
	if glyph.CachedFont != cachedFont {
		t.Errorf("CachedFont = %v, expected %v", glyph.CachedFont, cachedFont)
	}
	if glyph.GlyphCode != glyphCode {
		t.Errorf("GlyphCode = %d, expected %d", glyph.GlyphCode, glyphCode)
	}
	if glyph.GlyphIndex != glyphIndex {
		t.Errorf("GlyphIndex = %d, expected %d", glyph.GlyphIndex, glyphIndex)
	}
	if glyph.DataSize != dataSize {
		t.Errorf("DataSize = %d, expected %d", glyph.DataSize, dataSize)
	}
	if glyph.DataType != FmanGlyphDataGray8 {
		t.Errorf("DataType = %v, expected %v", glyph.DataType, FmanGlyphDataGray8)
	}
	if glyph.Bounds != bounds {
		t.Errorf("Bounds = %v, expected %v", glyph.Bounds, bounds)
	}
	if glyph.AdvanceX != advanceX {
		t.Errorf("AdvanceX = %f, expected %f", glyph.AdvanceX, advanceX)
	}
	if glyph.AdvanceY != advanceY {
		t.Errorf("AdvanceY = %f, expected %f", glyph.AdvanceY, advanceY)
	}
	if len(glyph.Data) != int(dataSize) {
		t.Errorf("Data length = %d, expected %d", len(glyph.Data), dataSize)
	}

	// Find the cached glyph
	foundGlyph := cg.FindGlyph(glyphCode)
	if foundGlyph == nil {
		t.Error("FindGlyph returned nil for cached glyph")
	}
	if foundGlyph != glyph {
		t.Error("FindGlyph returned different glyph than cached")
	}
}

func TestFmanCachedGlyphs_DuplicateGlyph(t *testing.T) {
	cg := NewFmanCachedGlyphs()

	glyphCode := uint32(65)
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 10}

	// Create mock cached fonts
	mockFace1 := NewMockLoadedFace()
	font1 := NewFmanCachedFont(mockFace1, 16.0, 8.0, false, FmanGlyphRenAggMono)
	mockFace2 := NewMockLoadedFace()
	font2 := NewFmanCachedFont(mockFace2, 18.0, 9.0, true, FmanGlyphRenAggGray8)

	// Cache first glyph
	glyph1 := cg.CacheGlyph(font1, glyphCode, 1, 32, FmanGlyphDataMono, bounds, 10.0, 0.0)
	if glyph1 == nil {
		t.Fatal("Failed to cache first glyph")
	}

	// Try to cache duplicate - should return nil
	glyph2 := cg.CacheGlyph(font2, glyphCode, 2, 64, FmanGlyphDataGray8, bounds, 20.0, 5.0)
	if glyph2 != nil {
		t.Error("CacheGlyph should return nil for duplicate glyph")
	}

	// Original glyph should be unchanged
	foundGlyph := cg.FindGlyph(glyphCode)
	if foundGlyph != glyph1 {
		t.Error("Original glyph was modified")
	}
	if foundGlyph.GlyphIndex != 1 {
		t.Error("Original glyph index was changed")
	}
}

func TestFmanCachedGlyphs_MultiLevelLookup(t *testing.T) {
	cg := NewFmanCachedGlyphs()

	// Test glyphs in different MSB ranges
	testCodes := []uint32{
		0x0041, // ASCII 'A' (MSB=0, LSB=65)
		0x00C1, // Latin 'Á' (MSB=0, LSB=193)
		0x0141, // Latin 'Ł' (MSB=1, LSB=65)
		0xFF41, // Unicode (MSB=255, LSB=65)
	}

	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 10}

	// Cache glyphs in different ranges
	var cachedGlyphs []*FmanCachedGlyph
	for i, code := range testCodes {
		mockFace := NewMockLoadedFace()
		font := NewFmanCachedFont(mockFace, 16.0, 8.0, false, FmanGlyphRenAggMono)
		glyph := cg.CacheGlyph(font, code, uint32(i+1), 32, FmanGlyphDataMono, bounds, 10.0, 0.0)
		if glyph == nil {
			t.Fatalf("Failed to cache glyph 0x%04X", code)
		}
		cachedGlyphs = append(cachedGlyphs, glyph)
	}

	// Verify all glyphs can be found
	for i, code := range testCodes {
		foundGlyph := cg.FindGlyph(code)
		if foundGlyph == nil {
			t.Errorf("Failed to find glyph 0x%04X", code)
			continue
		}
		if foundGlyph != cachedGlyphs[i] {
			t.Errorf("Found wrong glyph for code 0x%04X", code)
		}
		if foundGlyph.GlyphIndex != uint32(i+1) {
			t.Errorf("Wrong glyph index for code 0x%04X: got %d, expected %d",
				code, foundGlyph.GlyphIndex, i+1)
		}
	}
}

func TestFmanCachedFont_NewAndBasics(t *testing.T) {
	face := NewMockLoadedFace()

	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)
	if cf == nil {
		t.Fatal("NewFmanCachedFont returned nil")
	}

	// Test cached metrics
	if cf.Height() != 16.0 {
		t.Errorf("Height() = %f, expected 16.0", cf.Height())
	}
	if cf.Width() != 8.0 {
		t.Errorf("Width() = %f, expected 8.0", cf.Width())
	}
	if cf.Ascent() != 12.0 {
		t.Errorf("Ascent() = %f, expected 12.0", cf.Ascent())
	}
	if cf.Descent() != 4.0 {
		t.Errorf("Descent() = %f, expected 4.0", cf.Descent())
	}
	if cf.AscentB() != 13.0 {
		t.Errorf("AscentB() = %f, expected 13.0", cf.AscentB())
	}
	if cf.DescentB() != 3.0 {
		t.Errorf("DescentB() = %f, expected 3.0", cf.DescentB())
	}
}

func TestFmanCachedFont_GetGlyph(t *testing.T) {
	face := NewMockLoadedFace()

	// Add some test glyphs to face
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	face.AddGlyph(65, 1, FmanGlyphDataGray8, bounds, 12.0, 0.0, []byte{0xFF, 0x80, 0x40, 0x20})
	face.AddGlyph(66, 2, FmanGlyphDataMono, bounds, 10.0, 0.0, []byte{0xFF, 0xF0})

	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)

	// Get first glyph - should cache it
	glyph1 := cf.GetGlyph(65)
	if glyph1 == nil {
		t.Fatal("Failed to get glyph 65")
	}
	if glyph1.GlyphCode != 65 {
		t.Errorf("Glyph 65 code = %d, expected 65", glyph1.GlyphCode)
	}
	if glyph1.GlyphIndex != 1 {
		t.Errorf("Glyph 65 index = %d, expected 1", glyph1.GlyphIndex)
	}
	if glyph1.DataType != FmanGlyphDataGray8 {
		t.Errorf("Glyph 65 type = %v, expected %v", glyph1.DataType, FmanGlyphDataGray8)
	}
	if glyph1.CachedFont != cf {
		t.Error("Glyph 65 CachedFont should reference the cached font")
	}

	// Get second glyph
	glyph2 := cf.GetGlyph(66)
	if glyph2 == nil {
		t.Fatal("Failed to get glyph 66")
	}
	if glyph2.GlyphIndex != 2 {
		t.Errorf("Glyph 66 index = %d, expected 2", glyph2.GlyphIndex)
	}

	// Get first glyph again - should come from cache
	glyph1Again := cf.GetGlyph(65)
	if glyph1Again != glyph1 {
		t.Error("Second request for glyph 65 returned different instance")
	}
}

func TestFmanCachedFont_NonExistentGlyph(t *testing.T) {
	face := NewMockLoadedFace()
	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)

	// Try to get non-existent glyph
	glyph := cf.GetGlyph(999)
	if glyph != nil {
		t.Error("GetGlyph(999) should return nil for non-existent glyph")
	}
}

func TestFmanCachedFont_Kerning(t *testing.T) {
	face := NewMockLoadedFace()

	// Add test glyphs
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	face.AddGlyph(65, 1, FmanGlyphDataMono, bounds, 10.0, 0.0, []byte{0xFF})
	face.AddGlyph(66, 2, FmanGlyphDataMono, bounds, 10.0, 0.0, []byte{0xF0})

	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)

	// Get glyphs
	glyph1 := cf.GetGlyph(65)
	glyph2 := cf.GetGlyph(66)

	// Test kerning
	x, y := 0.0, 0.0
	hasKerning := cf.AddKerning(glyph1, glyph2, &x, &y)
	if !hasKerning {
		t.Error("AddKerning should return true for valid glyph pair")
	}
	if x != 1.5 {
		t.Errorf("Kerning x = %f, expected 1.5", x)
	}

	// Test kerning with nil glyph
	x, y = 0.0, 0.0
	hasKerning = cf.AddKerning(nil, glyph2, &x, &y)
	if hasKerning {
		t.Error("AddKerning should return false with nil glyph")
	}
}

func TestFmanFontCacheManager2_NewAndBasics(t *testing.T) {
	engine := NewMockFontEngine2()
	fcm := NewFmanFontCacheManager2(engine)
	if fcm == nil {
		t.Fatal("NewFmanFontCacheManager2 returned nil")
	}

	if fcm.Engine() != engine {
		t.Error("Engine() should return the original engine")
	}

	// Test adaptor access
	if fcm.PathAdaptor() != engine.PathAdaptor() {
		t.Error("PathAdaptor() should return engine's path adaptor")
	}
	if fcm.Gray8Adaptor() != engine.Gray8Adaptor() {
		t.Error("Gray8Adaptor() should return engine's gray8 adaptor")
	}
	if fcm.Gray8Scanline() != engine.Gray8Scanline() {
		t.Error("Gray8Scanline() should return engine's gray8 scanline")
	}
	if fcm.MonoAdaptor() != engine.MonoAdaptor() {
		t.Error("MonoAdaptor() should return engine's mono adaptor")
	}
	if fcm.MonoScanline() != engine.MonoScanline() {
		t.Error("MonoScanline() should return engine's mono scanline")
	}
}

func TestFmanFontCacheManager2_InitEmbeddedAdaptors(t *testing.T) {
	engine := NewMockFontEngine2()
	fcm := NewFmanFontCacheManager2(engine)

	// Test with nil glyph
	fcm.InitEmbeddedAdaptors(nil, 10.0, 20.0, 1.5)
	// Should not crash

	// Create test glyphs for each data type
	testGlyphs := []*FmanCachedGlyph{
		{
			DataType: FmanGlyphDataMono,
			Data:     []byte{0xFF, 0x80},
			DataSize: 2,
		},
		{
			DataType: FmanGlyphDataGray8,
			Data:     []byte{0xFF, 0x80, 0x40, 0x20},
			DataSize: 4,
		},
		{
			DataType: FmanGlyphDataOutline,
			Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			DataSize: 6,
		},
		{
			DataType: FmanGlyphDataInvalid,
			Data:     []byte{},
			DataSize: 0,
		},
	}

	x, y, scale := 10.0, 20.0, 1.5

	for i, glyph := range testGlyphs {
		fcm.InitEmbeddedAdaptors(glyph, x, y, scale)

		switch glyph.DataType {
		case FmanGlyphDataMono:
			monoAdaptor := engine.MonoAdaptor().(*MockMonoAdaptor)
			if !monoAdaptor.initialized {
				t.Errorf("Test %d: Mono adaptor should be initialized", i)
			}
			if monoAdaptor.x != x || monoAdaptor.y != y {
				t.Errorf("Test %d: Mono adaptor coordinates wrong", i)
			}

		case FmanGlyphDataGray8:
			gray8Adaptor := engine.Gray8Adaptor().(*MockGray8Adaptor)
			if !gray8Adaptor.initialized {
				t.Errorf("Test %d: Gray8 adaptor should be initialized", i)
			}
			if gray8Adaptor.x != x || gray8Adaptor.y != y {
				t.Errorf("Test %d: Gray8 adaptor coordinates wrong", i)
			}

		case FmanGlyphDataOutline:
			pathAdaptor := engine.PathAdaptor().(*MockPathAdaptor)
			if !pathAdaptor.initialized {
				t.Errorf("Test %d: Path adaptor should be initialized", i)
			}
			if pathAdaptor.x != x || pathAdaptor.y != y || pathAdaptor.scale != scale {
				t.Errorf("Test %d: Path adaptor parameters wrong", i)
			}

		case FmanGlyphDataInvalid:
			// Should not initialize any adaptor
		}

		// Reset adaptors for next test
		engine.monoAdaptor.initialized = false
		engine.gray8Adaptor.initialized = false
		engine.pathAdaptor.initialized = false
	}
}

func TestGetFontMetrics(t *testing.T) {
	face := NewMockLoadedFace()
	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)

	metrics := GetFontMetrics(cf)

	if metrics.Height != 16.0 {
		t.Errorf("Height = %f, expected 16.0", metrics.Height)
	}
	if metrics.Width != 8.0 {
		t.Errorf("Width = %f, expected 8.0", metrics.Width)
	}
	if metrics.Ascent != 12.0 {
		t.Errorf("Ascent = %f, expected 12.0", metrics.Ascent)
	}
	if metrics.Descent != 4.0 {
		t.Errorf("Descent = %f, expected 4.0", metrics.Descent)
	}
	if metrics.AscentB != 13.0 {
		t.Errorf("AscentB = %f, expected 13.0", metrics.AscentB)
	}
	if metrics.DescentB != 3.0 {
		t.Errorf("DescentB = %f, expected 3.0", metrics.DescentB)
	}
}

func TestGetCacheStats(t *testing.T) {
	face := NewMockLoadedFace()

	// Add test glyphs
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	face.AddGlyph(65, 1, FmanGlyphDataMono, bounds, 10.0, 0.0, []byte{0xFF})
	face.AddGlyph(66, 2, FmanGlyphDataGray8, bounds, 10.0, 0.0, []byte{0xFF, 0x80})
	face.AddGlyph(67, 3, FmanGlyphDataOutline, bounds, 10.0, 0.0, []byte{0xFF, 0x80, 0x40})

	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)

	// Get initial stats (no glyphs cached yet)
	stats := GetCacheStats(cf)
	if stats.CachedGlyphs != 0 {
		t.Errorf("Initial cached glyphs = %d, expected 0", stats.CachedGlyphs)
	}

	// Cache some glyphs
	cf.GetGlyph(65)
	cf.GetGlyph(66)
	cf.GetGlyph(67)

	// Get stats after caching
	stats = GetCacheStats(cf)
	if stats.CachedGlyphs != 3 {
		t.Errorf("Cached glyphs after caching = %d, expected 3", stats.CachedGlyphs)
	}
	if stats.AllocatedBlocks == 0 {
		t.Error("Should have at least one allocated block")
	}
	if stats.TotalMemory == 0 {
		t.Error("Should have allocated some memory")
	}
}

func BenchmarkFmanCachedFont_GetGlyph(b *testing.B) {
	face := NewMockLoadedFace()

	// Add many glyphs
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	for i := uint32(0); i < 1000; i++ {
		data := make([]byte, 32)
		for j := range data {
			data[j] = byte(i + uint32(j))
		}
		face.AddGlyph(i, i, FmanGlyphDataGray8, bounds, 10.0, 0.0, data)
	}

	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Access random glyph
		glyphCode := uint32(i % 1000)
		glyph := cf.GetGlyph(glyphCode)
		if glyph == nil {
			b.Fatalf("Failed to get glyph %d", glyphCode)
		}
	}
}

func BenchmarkFmanCachedGlyphs_FindGlyph(b *testing.B) {
	cg := NewFmanCachedGlyphs()

	// Cache many glyphs
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 10}
	for i := uint32(0); i < 1000; i++ {
		cg.CacheGlyph(nil, i, i, 32, FmanGlyphDataMono, bounds, 10.0, 0.0)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		glyphCode := uint32(i % 1000)
		glyph := cg.FindGlyph(glyphCode)
		if glyph == nil {
			b.Fatalf("Failed to find glyph %d", glyphCode)
		}
	}
}

func ExampleFmanCachedFont() {
	// Create a mock font face
	face := NewMockLoadedFace()

	// Add some glyphs to the face
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 12, Y2: 16}
	face.AddGlyph(65, 1, FmanGlyphDataGray8, bounds, 12.0, 0.0, []byte{0xFF, 0x80, 0x40, 0x20}) // 'A'
	face.AddGlyph(66, 2, FmanGlyphDataGray8, bounds, 11.0, 0.0, []byte{0xFF, 0xF0, 0x80, 0x40}) // 'B'

	// Create cached font
	cf := NewFmanCachedFont(face, 16.0, 8.0, true, FmanGlyphRenNativeGray8)

	// Get font metrics
	metrics := GetFontMetrics(cf)
	fmt.Printf("Font metrics: height=%.1f, width=%.1f, ascent=%.1f\n",
		metrics.Height, metrics.Width, metrics.Ascent)

	// Get glyphs (will be cached automatically)
	glyphA := cf.GetGlyph(65)
	glyphB := cf.GetGlyph(66)

	fmt.Printf("Glyph A: code=%d, index=%d, advance=%.1f, type=%s\n",
		glyphA.GlyphCode, glyphA.GlyphIndex, glyphA.AdvanceX, glyphA.DataType)
	fmt.Printf("Glyph B: code=%d, index=%d, advance=%.1f, type=%s\n",
		glyphB.GlyphCode, glyphB.GlyphIndex, glyphB.AdvanceX, glyphB.DataType)

	// Check kerning between A and B
	x, y := 0.0, 0.0
	if cf.AddKerning(glyphA, glyphB, &x, &y) {
		fmt.Printf("Kerning adjustment: x=%.1f, y=%.1f\n", x, y)
	}

	// Get cache statistics
	stats := GetCacheStats(cf)
	fmt.Printf("Cache stats: %d glyphs cached, %d blocks allocated\n",
		stats.CachedGlyphs, stats.AllocatedBlocks)

	// Output:
	// Font metrics: height=16.0, width=8.0, ascent=12.0
	// Glyph A: code=65, index=1, advance=12.0, type=gray8
	// Glyph B: code=66, index=2, advance=11.0, type=gray8
	// Kerning adjustment: x=1.5, y=0.0
	// Cache stats: 2 glyphs cached, 1 blocks allocated
}

func ExampleFmanFontCacheManager2() {
	// Create a mock font engine
	engine := NewMockFontEngine2()

	// Create enhanced font cache manager
	fcm := NewFmanFontCacheManager2(engine)

	// Create a test glyph for adaptor initialization
	testGlyph := &FmanCachedGlyph{
		DataType: FmanGlyphDataGray8,
		Data:     []byte{0xFF, 0x80, 0x40, 0x20},
		DataSize: 4,
	}

	// Initialize embedded adaptors for the glyph
	fcm.InitEmbeddedAdaptors(testGlyph, 100.0, 200.0, 1.0)

	// Access adaptors
	gray8Adaptor := fcm.Gray8Adaptor().(*MockGray8Adaptor)
	fmt.Printf("Gray8 adaptor initialized: %t\n", gray8Adaptor.initialized)
	fmt.Printf("Gray8 adaptor position: x=%.1f, y=%.1f\n", gray8Adaptor.x, gray8Adaptor.y)
	fmt.Printf("Gray8 adaptor data size: %d bytes\n", gray8Adaptor.dataSize)

	// Output:
	// Gray8 adaptor initialized: true
	// Gray8 adaptor position: x=100.0, y=200.0
	// Gray8 adaptor data size: 4 bytes
}
