// Package font provides font rendering capabilities for the AGG graphics library.
package font

import (
	"unsafe"

	"agg_go/internal/basics"
	"agg_go/internal/path"
)

// Block allocator parameters, matching C++ implementation.
const blockSize = 16384 - 16

// FontEngine defines the interface that font engines must implement
// to work with the cache manager.
type FontEngine interface {
	// Font identification and management
	FontSignature() string
	ChangeStamp() int

	// Glyph preparation and access
	PrepareGlyph(glyphCode uint) bool
	GlyphIndex() uint
	DataSize() uint
	DataType() GlyphDataType
	Bounds() basics.Rect[int]
	AdvanceX() float64
	AdvanceY() float64
	WriteGlyphTo(data []byte)
	AddKerning(first, second uint) (dx, dy float64)

	// Path access for outline rendering
	PathAdaptor() *path.PathStorageStl
}

// FontCache stores glyphs for a single font with efficient two-level indexing.
// This mirrors the C++ font_cache class structure.
type FontCache struct {
	allocator     *blockAllocator
	fontSignature string
	glyphs        [256]*[256]*GlyphCache // Two-level array: [MSB][LSB]
}

// blockAllocator provides efficient memory allocation for font data.
type blockAllocator struct {
	blocks       [][]byte
	currentBlock int
	currentPos   int
}

// newBlockAllocator creates a new block allocator.
func newBlockAllocator() *blockAllocator {
	alloc := &blockAllocator{}
	alloc.addBlock()
	return alloc
}

// addBlock adds a new memory block to the allocator.
func (ba *blockAllocator) addBlock() {
	ba.blocks = append(ba.blocks, make([]byte, blockSize))
	ba.currentBlock = len(ba.blocks) - 1
	ba.currentPos = 0
}

// allocate allocates memory from the current block or creates a new block if needed.
func (ba *blockAllocator) allocate(size int) []byte {
	// Align to pointer size
	alignedSize := (size + int(unsafe.Sizeof(uintptr(0))) - 1) & ^(int(unsafe.Sizeof(uintptr(0))) - 1)

	if ba.currentPos+alignedSize > blockSize {
		ba.addBlock()
	}

	result := ba.blocks[ba.currentBlock][ba.currentPos : ba.currentPos+size]
	ba.currentPos += alignedSize
	return result
}

// NewFontCache creates a new font cache.
func NewFontCache() *FontCache {
	return &FontCache{
		allocator: newBlockAllocator(),
	}
}

// SetSignature sets the font signature for this cache.
func (fc *FontCache) SetSignature(fontSignature string) {
	fc.fontSignature = fontSignature
	// Clear existing glyphs when signature changes
	fc.glyphs = [256]*[256]*GlyphCache{}
}

// FontIs checks if this cache belongs to the specified font.
func (fc *FontCache) FontIs(fontSignature string) bool {
	return fc.fontSignature == fontSignature
}

// FindGlyph finds a cached glyph by character code.
func (fc *FontCache) FindGlyph(glyphCode uint) *GlyphCache {
	msb := (glyphCode >> 8) & 0xFF
	lsb := glyphCode & 0xFF

	if fc.glyphs[msb] != nil {
		return fc.glyphs[msb][lsb]
	}
	return nil
}

// CacheGlyph stores a glyph in the cache.
func (fc *FontCache) CacheGlyph(glyphCode, glyphIndex uint, dataSize uint,
	dataType GlyphDataType, bounds basics.Rect[int], advanceX, advanceY float64,
) *GlyphCache {
	msb := (glyphCode >> 8) & 0xFF
	lsb := glyphCode & 0xFF

	// Allocate LSB array if needed
	if fc.glyphs[msb] == nil {
		// Allocate space for the array and zero it
		arraySize := 256 * int(unsafe.Sizeof((*GlyphCache)(nil)))
		arrayData := fc.allocator.allocate(arraySize)
		fc.glyphs[msb] = (*[256]*GlyphCache)(unsafe.Pointer(&arrayData[0]))
		// Zero the array
		for i := 0; i < 256; i++ {
			fc.glyphs[msb][i] = nil
		}
	}

	// Allocate and initialize the glyph cache entry
	glyphSize := int(unsafe.Sizeof(GlyphCache{}))
	glyphData := fc.allocator.allocate(glyphSize)
	glyph := (*GlyphCache)(unsafe.Pointer(&glyphData[0]))

	*glyph = GlyphCache{
		GlyphIndex: glyphIndex,
		DataSize:   dataSize,
		DataType:   dataType,
		Bounds:     bounds,
		AdvanceX:   advanceX,
		AdvanceY:   advanceY,
	}

	// Allocate data buffer if needed
	if dataSize > 0 {
		glyph.Data = fc.allocator.allocate(int(dataSize))
	}

	fc.glyphs[msb][lsb] = glyph
	return glyph
}

// FontCacheManager manages multiple font caches and coordinates with font engines.
// This is the main interface for font rendering, similar to the C++ font_cache_manager template.
type FontCacheManager struct {
	fontEngine   FontEngine
	fontCaches   []*FontCache
	currentCache *FontCache
	maxFonts     int
	pathAdaptor  *path.PathStorageStl
	gray8Adaptor *SerializedScanlinesAdaptorAA
	monoAdaptor  *SerializedScanlinesAdaptorBin
	lastError    error
}

// NewFontCacheManager creates a new font cache manager with the given font engine.
func NewFontCacheManager(fontEngine FontEngine, maxFonts int) *FontCacheManager {
	if maxFonts <= 0 {
		maxFonts = 32 // Default value from C++ implementation
	}

	return &FontCacheManager{
		fontEngine:  fontEngine,
		fontCaches:  make([]*FontCache, 0, maxFonts),
		maxFonts:    maxFonts,
		pathAdaptor: path.NewPathStorageStl(),
	}
}

// findFontCache finds or creates a cache for the current font signature.
func (fcm *FontCacheManager) findFontCache() *FontCache {
	signature := fcm.fontEngine.FontSignature()

	// Look for existing cache
	for _, cache := range fcm.fontCaches {
		if cache.FontIs(signature) {
			return cache
		}
	}

	// Create new cache
	newCache := NewFontCache()
	newCache.SetSignature(signature)

	// Add to cache list (with LRU eviction if needed)
	if len(fcm.fontCaches) >= fcm.maxFonts {
		// Remove oldest cache (simple FIFO for now)
		fcm.fontCaches = fcm.fontCaches[1:]
	}

	fcm.fontCaches = append(fcm.fontCaches, newCache)
	return newCache
}

// Glyph retrieves a glyph from the cache, loading it from the font engine if necessary.
func (fcm *FontCacheManager) Glyph(charCode uint) *GlyphCache {
	fcm.currentCache = fcm.findFontCache()

	// Look for cached glyph
	if glyph := fcm.currentCache.FindGlyph(charCode); glyph != nil {
		return glyph
	}

	// Load glyph from font engine
	if !fcm.fontEngine.PrepareGlyph(charCode) {
		return nil
	}

	// Cache the glyph with typed values
	glyph := fcm.currentCache.CacheGlyph(
		charCode,
		fcm.fontEngine.GlyphIndex(),
		fcm.fontEngine.DataSize(),
		fcm.fontEngine.DataType(),
		fcm.fontEngine.Bounds(),
		fcm.fontEngine.AdvanceX(),
		fcm.fontEngine.AdvanceY(),
	)

	// Write glyph data
	if glyph.DataSize > 0 {
		fcm.fontEngine.WriteGlyphTo(glyph.Data)
	}

	return glyph
}

// AddKerning adds kerning adjustment between two characters.
func (fcm *FontCacheManager) AddKerning(x, y *float64, first, second uint) {
	dx, dy := fcm.fontEngine.AddKerning(first, second)
	*x += dx
	*y += dy
}

// PathAdaptor returns the path adaptor for vector font rendering.
func (fcm *FontCacheManager) PathAdaptor() *path.PathStorageStl {
	return fcm.pathAdaptor
}

// InitEmbeddedAdaptors initializes the glyph adaptors for rendering at the specified position.
func (fcm *FontCacheManager) InitEmbeddedAdaptors(glyph *GlyphCache, x, y float64) {
	switch glyph.DataType {
	case GlyphDataGray8:
		fcm.gray8Adaptor = NewSerializedScanlinesAdaptorAA(glyph.Data, glyph.Bounds)
	case GlyphDataMono:
		fcm.monoAdaptor = NewSerializedScanlinesAdaptorBin(glyph.Data, glyph.Bounds)
	case GlyphDataOutline:
		// Path data would be reconstructed here for vector fonts
		// This requires integration with the path system
	}
}

// Gray8Adaptor returns the gray8 scanline adaptor.
func (fcm *FontCacheManager) Gray8Adaptor() *SerializedScanlinesAdaptorAA {
	return fcm.gray8Adaptor
}

// MonoAdaptor returns the mono scanline adaptor.
func (fcm *FontCacheManager) MonoAdaptor() *SerializedScanlinesAdaptorBin {
	return fcm.monoAdaptor
}
