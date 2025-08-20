package fonts

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
)

// MockFontEngine implements FontEngine for testing
type MockFontEngine struct {
	signature    string
	stamp        int
	glyphData    map[uint32]mockGlyphData
	prepared     *mockGlyphData
	preparedCode uint32
}

type mockGlyphData struct {
	index    uint32
	data     []byte
	dataType GlyphDataType
	bounds   basics.Rect[int]
	advanceX float64
	advanceY float64
}

func NewMockFontEngine(signature string) *MockFontEngine {
	return &MockFontEngine{
		signature: signature,
		stamp:     1,
		glyphData: make(map[uint32]mockGlyphData),
	}
}

func (m *MockFontEngine) AddGlyph(code uint32, index uint32, dataType GlyphDataType,
	bounds basics.Rect[int], advanceX, advanceY float64, data []byte) {
	m.glyphData[code] = mockGlyphData{
		index:    index,
		data:     data,
		dataType: dataType,
		bounds:   bounds,
		advanceX: advanceX,
		advanceY: advanceY,
	}
}

func (m *MockFontEngine) FontSignature() string    { return m.signature }
func (m *MockFontEngine) ChangeStamp() int         { return m.stamp }
func (m *MockFontEngine) SetChangeStamp(stamp int) { m.stamp = stamp }

func (m *MockFontEngine) PrepareGlyph(glyphCode uint32) bool {
	if glyph, exists := m.glyphData[glyphCode]; exists {
		m.prepared = &glyph
		m.preparedCode = glyphCode
		return true
	}
	m.prepared = nil
	return false
}

func (m *MockFontEngine) GlyphIndex() uint32 {
	if m.prepared != nil {
		return m.prepared.index
	}
	return 0
}

func (m *MockFontEngine) DataSize() uint32 {
	if m.prepared != nil {
		return uint32(len(m.prepared.data))
	}
	return 0
}

func (m *MockFontEngine) DataType() GlyphDataType {
	if m.prepared != nil {
		return m.prepared.dataType
	}
	return GlyphDataInvalid
}

func (m *MockFontEngine) Bounds() basics.Rect[int] {
	if m.prepared != nil {
		return m.prepared.bounds
	}
	return basics.Rect[int]{}
}

func (m *MockFontEngine) AdvanceX() float64 {
	if m.prepared != nil {
		return m.prepared.advanceX
	}
	return 0
}

func (m *MockFontEngine) AdvanceY() float64 {
	if m.prepared != nil {
		return m.prepared.advanceY
	}
	return 0
}

func (m *MockFontEngine) WriteGlyphTo(data []byte) {
	if m.prepared != nil && len(data) >= len(m.prepared.data) {
		copy(data, m.prepared.data)
	}
}

func (m *MockFontEngine) AddKerning(first, second uint32, x, y *float64) bool {
	// Simple mock kerning: add 1.0 to x for any pair
	if first != 0 && second != 0 {
		*x += 1.0
		return true
	}
	return false
}

func TestGlyphDataType_String(t *testing.T) {
	tests := []struct {
		gdt      GlyphDataType
		expected string
	}{
		{GlyphDataInvalid, "invalid"},
		{GlyphDataMono, "mono"},
		{GlyphDataGray8, "gray8"},
		{GlyphDataOutline, "outline"},
		{GlyphDataType(999), "unknown"},
	}

	for _, test := range tests {
		result := test.gdt.String()
		if result != test.expected {
			t.Errorf("GlyphDataType(%d).String() = %q, expected %q", test.gdt, result, test.expected)
		}
	}
}

func TestGlyphRendering_String(t *testing.T) {
	tests := []struct {
		gr       GlyphRendering
		expected string
	}{
		{GlyphRenNativeMono, "native_mono"},
		{GlyphRenNativeGray8, "native_gray8"},
		{GlyphRenOutline, "outline"},
		{GlyphRenAggMono, "agg_mono"},
		{GlyphRenAggGray8, "agg_gray8"},
		{GlyphRendering(999), "unknown"},
	}

	for _, test := range tests {
		result := test.gr.String()
		if result != test.expected {
			t.Errorf("GlyphRendering(%d).String() = %q, expected %q", test.gr, result, test.expected)
		}
	}
}

func TestFontCache_NewAndBasics(t *testing.T) {
	fc := NewFontCache()
	if fc == nil {
		t.Fatal("NewFontCache() returned nil")
	}

	fc2 := NewFontCacheWithBlockSize(8192)
	if fc2 == nil {
		t.Fatal("NewFontCacheWithBlockSize() returned nil")
	}
}

func TestFontCache_Signature(t *testing.T) {
	fc := NewFontCache()

	signature := "Arial-12pt-bold"
	fc.SetSignature(signature)

	if !fc.FontIs(signature) {
		t.Errorf("FontIs(%q) returned false after SetSignature", signature)
	}

	if fc.FontIs("Different-font") {
		t.Errorf("FontIs returned true for different signature")
	}
}

func TestFontCache_GlyphCache(t *testing.T) {
	fc := NewFontCache()
	fc.SetSignature("test-font")

	// Test caching a glyph
	glyphCode := uint32(65) // 'A'
	glyphIndex := uint32(42)
	dataSize := uint32(64)
	bounds := basics.Rect[int]{X1: 10, Y1: 20, X2: 30, Y2: 40}
	advanceX := 12.5
	advanceY := 0.0

	glyph := fc.CacheGlyph(glyphCode, glyphIndex, dataSize, GlyphDataGray8, bounds, advanceX, advanceY)
	if glyph == nil {
		t.Fatal("CacheGlyph returned nil")
	}

	// Verify glyph properties
	if glyph.GlyphIndex != glyphIndex {
		t.Errorf("GlyphIndex = %d, expected %d", glyph.GlyphIndex, glyphIndex)
	}
	if glyph.DataSize != dataSize {
		t.Errorf("DataSize = %d, expected %d", glyph.DataSize, dataSize)
	}
	if glyph.DataType != GlyphDataGray8 {
		t.Errorf("DataType = %v, expected %v", glyph.DataType, GlyphDataGray8)
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
}

func TestFontCache_FindGlyph(t *testing.T) {
	fc := NewFontCache()
	fc.SetSignature("test-font")

	// Test finding non-existent glyph
	if glyph := fc.FindGlyph(999); glyph != nil {
		t.Error("FindGlyph returned non-nil for non-existent glyph")
	}

	// Cache a glyph
	glyphCode := uint32(65)
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 10}
	cachedGlyph := fc.CacheGlyph(glyphCode, 1, 32, GlyphDataMono, bounds, 10.0, 0.0)
	if cachedGlyph == nil {
		t.Fatal("Failed to cache glyph")
	}

	// Find the cached glyph
	foundGlyph := fc.FindGlyph(glyphCode)
	if foundGlyph == nil {
		t.Error("FindGlyph returned nil for cached glyph")
	}
	if foundGlyph != cachedGlyph {
		t.Error("FindGlyph returned different glyph than cached")
	}
}

func TestFontCache_DuplicateGlyph(t *testing.T) {
	fc := NewFontCache()
	fc.SetSignature("test-font")

	glyphCode := uint32(65)
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 10}

	// Cache first glyph
	glyph1 := fc.CacheGlyph(glyphCode, 1, 32, GlyphDataMono, bounds, 10.0, 0.0)
	if glyph1 == nil {
		t.Fatal("Failed to cache first glyph")
	}

	// Try to cache duplicate - should return nil
	glyph2 := fc.CacheGlyph(glyphCode, 2, 64, GlyphDataGray8, bounds, 20.0, 5.0)
	if glyph2 != nil {
		t.Error("CacheGlyph should return nil for duplicate glyph")
	}

	// Original glyph should be unchanged
	foundGlyph := fc.FindGlyph(glyphCode)
	if foundGlyph != glyph1 {
		t.Error("Original glyph was modified")
	}
	if foundGlyph.GlyphIndex != 1 {
		t.Error("Original glyph index was changed")
	}
}

func TestFontCache_MultiLevelLookup(t *testing.T) {
	fc := NewFontCache()
	fc.SetSignature("test-font")

	// Test glyphs in different MSB ranges
	testCodes := []uint32{
		0x0041, // ASCII 'A' (MSB=0, LSB=65)
		0x00C1, // Latin 'Á' (MSB=0, LSB=193)
		0x0141, // Latin 'Ł' (MSB=1, LSB=65)
		0xFF41, // Unicode (MSB=255, LSB=65)
	}

	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 10}

	// Cache glyphs in different ranges
	var cachedGlyphs []*GlyphCache
	for i, code := range testCodes {
		glyph := fc.CacheGlyph(code, uint32(i+1), 32, GlyphDataMono, bounds, 10.0, 0.0)
		if glyph == nil {
			t.Fatalf("Failed to cache glyph 0x%04X", code)
		}
		cachedGlyphs = append(cachedGlyphs, glyph)
	}

	// Verify all glyphs can be found
	for i, code := range testCodes {
		foundGlyph := fc.FindGlyph(code)
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

func TestFontCachePool_NewAndBasics(t *testing.T) {
	fcp := NewFontCachePool()
	if fcp == nil {
		t.Fatal("NewFontCachePool() returned nil")
	}

	fcp2 := NewFontCachePoolWithCapacity(16)
	if fcp2 == nil {
		t.Fatal("NewFontCachePoolWithCapacity() returned nil")
	}
	if fcp2.maxFonts != 16 {
		t.Errorf("maxFonts = %d, expected 16", fcp2.maxFonts)
	}
}

func TestFontCachePool_SetFont(t *testing.T) {
	fcp := NewFontCachePoolWithCapacity(3)

	// Set first font
	fcp.SetFont("Arial-12", false)
	if fcp.Font() == nil {
		t.Error("Font() returned nil after SetFont")
	}
	if !fcp.Font().FontIs("Arial-12") {
		t.Error("Current font has wrong signature")
	}
	if fcp.numFonts != 1 {
		t.Errorf("numFonts = %d, expected 1", fcp.numFonts)
	}

	// Set second font
	fcp.SetFont("Times-14", false)
	if !fcp.Font().FontIs("Times-14") {
		t.Error("Current font not switched to Times-14")
	}
	if fcp.numFonts != 2 {
		t.Errorf("numFonts = %d, expected 2", fcp.numFonts)
	}

	// Switch back to first font
	fcp.SetFont("Arial-12", false)
	if !fcp.Font().FontIs("Arial-12") {
		t.Error("Failed to switch back to Arial-12")
	}
	if fcp.numFonts != 2 {
		t.Errorf("numFonts = %d, expected 2", fcp.numFonts)
	}
}

func TestFontCachePool_LRUEviction(t *testing.T) {
	fcp := NewFontCachePoolWithCapacity(2)

	// Fill pool to capacity
	fcp.SetFont("Font1", false)
	fcp.SetFont("Font2", false)
	if fcp.numFonts != 2 {
		t.Errorf("numFonts = %d, expected 2", fcp.numFonts)
	}

	// Add third font - should evict oldest (Font1)
	fcp.SetFont("Font3", false)
	if fcp.numFonts != 2 {
		t.Errorf("numFonts = %d, expected 2 after eviction", fcp.numFonts)
	}

	// Font1 should be evicted, Font2 and Font3 should remain
	if fcp.findFont("Font1") >= 0 {
		t.Error("Font1 should have been evicted")
	}
	if fcp.findFont("Font2") < 0 {
		t.Error("Font2 should still be cached")
	}
	if fcp.findFont("Font3") < 0 {
		t.Error("Font3 should be cached")
	}
}

func TestFontCachePool_ResetCache(t *testing.T) {
	fcp := NewFontCachePool()

	// Add font and cache some glyphs
	fcp.SetFont("TestFont", false)
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 10}
	glyph := fcp.CacheGlyph(65, 1, 32, GlyphDataMono, bounds, 10.0, 0.0)
	if glyph == nil {
		t.Fatal("Failed to cache glyph")
	}

	// Verify glyph is cached
	if foundGlyph := fcp.FindGlyph(65); foundGlyph == nil {
		t.Error("Glyph not found after caching")
	}

	// Reset cache
	fcp.SetFont("TestFont", true)

	// Glyph should be gone
	if foundGlyph := fcp.FindGlyph(65); foundGlyph != nil {
		t.Error("Glyph still found after cache reset")
	}
}

func TestFontCacheManager_NewAndBasics(t *testing.T) {
	engine := NewMockFontEngine("test-font")
	fcm := NewFontCacheManager(engine)
	if fcm == nil {
		t.Fatal("NewFontCacheManager returned nil")
	}

	fcm2 := NewFontCacheManagerWithCapacity(engine, 16)
	if fcm2 == nil {
		t.Fatal("NewFontCacheManagerWithCapacity returned nil")
	}

	stats := fcm.Stats()
	if stats.MaxFonts != 32 {
		t.Errorf("Default max fonts = %d, expected 32", stats.MaxFonts)
	}

	stats2 := fcm2.Stats()
	if stats2.MaxFonts != 16 {
		t.Errorf("Custom max fonts = %d, expected 16", stats2.MaxFonts)
	}
}

func TestFontCacheManager_GlyphCaching(t *testing.T) {
	engine := NewMockFontEngine("test-font")

	// Add some test glyphs to engine
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	engine.AddGlyph(65, 1, GlyphDataGray8, bounds, 12.0, 0.0, []byte{0xFF, 0x80, 0x40, 0x20})
	engine.AddGlyph(66, 2, GlyphDataMono, bounds, 10.0, 0.0, []byte{0xFF, 0xF0})

	fcm := NewFontCacheManager(engine)

	// Get first glyph - should cache it
	glyph1 := fcm.Glyph(65)
	if glyph1 == nil {
		t.Fatal("Failed to get glyph 65")
	}
	if glyph1.GlyphIndex != 1 {
		t.Errorf("Glyph 65 index = %d, expected 1", glyph1.GlyphIndex)
	}
	if glyph1.DataType != GlyphDataGray8 {
		t.Errorf("Glyph 65 type = %v, expected %v", glyph1.DataType, GlyphDataGray8)
	}

	// Get second glyph
	glyph2 := fcm.Glyph(66)
	if glyph2 == nil {
		t.Fatal("Failed to get glyph 66")
	}
	if glyph2.GlyphIndex != 2 {
		t.Errorf("Glyph 66 index = %d, expected 2", glyph2.GlyphIndex)
	}

	// Get first glyph again - should come from cache
	glyph1Again := fcm.Glyph(65)
	if glyph1Again != glyph1 {
		t.Error("Second request for glyph 65 returned different instance")
	}

	// Check last/prev glyph tracking
	if fcm.LastGlyph() != glyph1 {
		t.Error("LastGlyph() should return glyph 65")
	}
	if fcm.PrevGlyph() != glyph2 {
		t.Error("PrevGlyph() should return glyph 66")
	}
}

func TestFontCacheManager_NonExistentGlyph(t *testing.T) {
	engine := NewMockFontEngine("test-font")
	fcm := NewFontCacheManager(engine)

	// Try to get non-existent glyph
	glyph := fcm.Glyph(999)
	if glyph != nil {
		t.Error("Glyph(999) should return nil for non-existent glyph")
	}
}

func TestFontCacheManager_ChangeStampSync(t *testing.T) {
	engine := NewMockFontEngine("test-font")

	// Add test glyph
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	engine.AddGlyph(65, 1, GlyphDataMono, bounds, 10.0, 0.0, []byte{0xFF})

	fcm := NewFontCacheManager(engine)

	// Cache a glyph
	glyph1 := fcm.Glyph(65)
	if glyph1 == nil {
		t.Fatal("Failed to cache glyph")
	}

	// Change engine stamp (simulates font parameter change)
	engine.SetChangeStamp(2)

	// Add new glyph data with same code but different properties
	engine.AddGlyph(65, 99, GlyphDataGray8, bounds, 20.0, 5.0, []byte{0x80, 0x40})

	// Get glyph again - should re-cache due to stamp change
	glyph2 := fcm.Glyph(65)
	if glyph2 == nil {
		t.Fatal("Failed to get glyph after stamp change")
	}
	if glyph2 == glyph1 {
		t.Error("Should have re-cached glyph after stamp change")
	}
	if glyph2.GlyphIndex != 99 {
		t.Errorf("New glyph index = %d, expected 99", glyph2.GlyphIndex)
	}
}

func TestFontCacheManager_Kerning(t *testing.T) {
	engine := NewMockFontEngine("test-font")

	// Add test glyphs
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	engine.AddGlyph(65, 1, GlyphDataMono, bounds, 10.0, 0.0, []byte{0xFF})
	engine.AddGlyph(66, 2, GlyphDataMono, bounds, 10.0, 0.0, []byte{0xF0})

	fcm := NewFontCacheManager(engine)

	// Get first glyph
	fcm.Glyph(65)

	// Get second glyph
	fcm.Glyph(66)

	// Test kerning
	x, y := 0.0, 0.0
	hasKerning := fcm.AddKerning(&x, &y)
	if !hasKerning {
		t.Error("AddKerning should return true for valid glyph pair")
	}
	if x != 1.0 {
		t.Errorf("Kerning x = %f, expected 1.0", x)
	}

	// Test kerning with no previous glyph
	fcm.ResetLastGlyph()
	x, y = 0.0, 0.0
	hasKerning = fcm.AddKerning(&x, &y)
	if hasKerning {
		t.Error("AddKerning should return false with no previous glyph")
	}
}

func TestFontCacheManager_Precache(t *testing.T) {
	engine := NewMockFontEngine("test-font")

	// Add range of glyphs
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	for i := uint32(65); i <= 70; i++ { // A-F
		engine.AddGlyph(i, i-64, GlyphDataMono, bounds, 10.0, 0.0, []byte{byte(i)})
	}

	fcm := NewFontCacheManager(engine)

	// Precache range
	fcm.Precache(65, 68) // A-D

	// Verify cached glyphs are available
	for i := uint32(65); i <= 68; i++ {
		glyph := fcm.Glyph(i)
		if glyph == nil {
			t.Errorf("Precached glyph %d not found", i)
		}
	}
}

func TestFontCacheManager_ResetCache(t *testing.T) {
	engine := NewMockFontEngine("test-font")

	// Add test glyph
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	engine.AddGlyph(65, 1, GlyphDataMono, bounds, 10.0, 0.0, []byte{0xFF})

	fcm := NewFontCacheManager(engine)

	// Cache glyph
	glyph1 := fcm.Glyph(65)
	if glyph1 == nil {
		t.Fatal("Failed to cache glyph")
	}

	// Reset cache
	fcm.ResetCache()

	// Glyph should be re-cached on next access
	glyph2 := fcm.Glyph(65)
	if glyph2 == nil {
		t.Fatal("Failed to get glyph after reset")
	}
	if glyph2 == glyph1 {
		t.Error("Should have re-cached glyph after reset")
	}

	// Last/prev glyphs should be reset
	if fcm.PrevGlyph() != nil {
		t.Error("PrevGlyph should be nil after reset")
	}
}

func BenchmarkFontCacheManager_GlyphAccess(b *testing.B) {
	engine := NewMockFontEngine("bench-font")

	// Add many glyphs
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	for i := uint32(0); i < 1000; i++ {
		data := make([]byte, 64)
		for j := range data {
			data[j] = byte(i + uint32(j))
		}
		engine.AddGlyph(i, i, GlyphDataGray8, bounds, 10.0, 0.0, data)
	}

	fcm := NewFontCacheManager(engine)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Access random glyph
		glyphCode := uint32(i % 1000)
		glyph := fcm.Glyph(glyphCode)
		if glyph == nil {
			b.Fatalf("Failed to get glyph %d", glyphCode)
		}
	}
}

func BenchmarkFontCacheManager_SequentialAccess(b *testing.B) {
	engine := NewMockFontEngine("bench-font")

	// Add ASCII range
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 10, Y2: 15}
	for i := uint32(32); i < 127; i++ {
		data := []byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}
		engine.AddGlyph(i, i, GlyphDataMono, bounds, 8.0, 0.0, data)
	}

	fcm := NewFontCacheManager(engine)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Sequential access pattern (like rendering text)
		for j := uint32(32); j < 127; j++ {
			glyph := fcm.Glyph(j)
			if glyph == nil {
				b.Fatalf("Failed to get glyph %d", j)
			}
		}
	}
}

func ExampleFontCacheManager() {
	// Create a mock font engine
	engine := NewMockFontEngine("Arial-12pt")

	// Add some glyphs to the engine
	bounds := basics.Rect[int]{X1: 0, Y1: 0, X2: 12, Y2: 16}
	engine.AddGlyph(65, 1, GlyphDataGray8, bounds, 12.0, 0.0, []byte{0xFF, 0x80, 0x40, 0x20}) // 'A'
	engine.AddGlyph(66, 2, GlyphDataGray8, bounds, 11.0, 0.0, []byte{0xFF, 0xF0, 0x80, 0x40}) // 'B'

	// Create font cache manager
	fcm := NewFontCacheManager(engine)

	// Get glyphs (will be cached automatically)
	glyphA := fcm.Glyph(65)
	glyphB := fcm.Glyph(66)

	fmt.Printf("Glyph A: index=%d, advance=%.1f, type=%s\n",
		glyphA.GlyphIndex, glyphA.AdvanceX, glyphA.DataType)
	fmt.Printf("Glyph B: index=%d, advance=%.1f, type=%s\n",
		glyphB.GlyphIndex, glyphB.AdvanceX, glyphB.DataType)

	// Check kerning between A and B
	x, y := 0.0, 0.0
	if fcm.AddKerning(&x, &y) {
		fmt.Printf("Kerning adjustment: x=%.1f, y=%.1f\n", x, y)
	}

	// Output:
	// Glyph A: index=1, advance=12.0, type=gray8
	// Glyph B: index=2, advance=11.0, type=gray8
	// Kerning adjustment: x=1.0, y=0.0
}
