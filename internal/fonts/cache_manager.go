// Package fonts provides font management and caching functionality for AGG.
// This package implements efficient glyph caching for high-performance text rendering.
package fonts

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// GlyphDataType represents the type of glyph data stored in cache.
// This corresponds to AGG's glyph_data_type enum.
type GlyphDataType uint32

const (
	GlyphDataInvalid GlyphDataType = iota // Invalid/empty glyph data
	GlyphDataMono                         // Monochrome bitmap glyph
	GlyphDataGray8                        // 8-bit grayscale glyph
	GlyphDataOutline                      // Vector outline glyph
)

// String returns string representation of glyph data type.
func (gdt GlyphDataType) String() string {
	switch gdt {
	case GlyphDataInvalid:
		return "invalid"
	case GlyphDataMono:
		return "mono"
	case GlyphDataGray8:
		return "gray8"
	case GlyphDataOutline:
		return "outline"
	default:
		return "unknown"
	}
}

// GlyphCache stores cached glyph data including bitmap, metrics, and positioning.
// This corresponds to AGG's glyph_cache struct.
type GlyphCache struct {
	GlyphIndex uint32           // Glyph index in font
	Data       []byte           // Glyph bitmap or outline data
	DataSize   uint32           // Size of data in bytes
	DataType   GlyphDataType    // Type of glyph data
	Bounds     basics.Rect[int] // Glyph bounding rectangle
	AdvanceX   float64          // Horizontal advance
	AdvanceY   float64          // Vertical advance
}

// FontCache manages cached glyphs for a single font.
// Uses a two-level sparse array for efficient glyph lookup by character code.
// This corresponds to AGG's font_cache class.
type FontCache struct {
	allocator     *array.BlockAllocator  // Memory allocator for glyph data
	glyphs        [256]*[256]*GlyphCache // Two-level glyph lookup table
	fontSignature string                 // Font identifier string
}

const (
	// DefaultBlockSize is the default block size for font cache allocator.
	// This matches AGG's block_size constant (16384-16).
	DefaultBlockSize = 16384 - 16
)

// NewFontCache creates a new font cache with default block size.
func NewFontCache() *FontCache {
	return &FontCache{
		allocator: array.NewBlockAllocator(DefaultBlockSize),
	}
}

// NewFontCacheWithBlockSize creates a new font cache with custom block size.
func NewFontCacheWithBlockSize(blockSize int) *FontCache {
	return &FontCache{
		allocator: array.NewBlockAllocator(blockSize),
	}
}

// SetSignature sets the font signature and clears all cached glyphs.
func (fc *FontCache) SetSignature(fontSignature string) {
	// Copy signature to allocator memory for consistency with C++ version
	sigBytes := fc.allocator.AllocateBytes(len(fontSignature) + 1)
	copy(sigBytes, fontSignature)
	sigBytes[len(fontSignature)] = 0
	fc.fontSignature = string(sigBytes[:len(fontSignature)])

	// Clear glyph cache
	fc.glyphs = [256]*[256]*GlyphCache{}
}

// FontIs checks if the current font matches the given signature.
func (fc *FontCache) FontIs(fontSignature string) bool {
	return fc.fontSignature == fontSignature
}

// FindGlyph finds a cached glyph by character code.
// Returns nil if glyph is not cached.
func (fc *FontCache) FindGlyph(glyphCode uint32) *GlyphCache {
	msb := (glyphCode >> 8) & 0xFF
	if fc.glyphs[msb] != nil {
		lsb := glyphCode & 0xFF
		return fc.glyphs[msb][lsb]
	}
	return nil
}

// CacheGlyph caches a new glyph with the given parameters.
// Returns the cached glyph or nil if glyph already exists.
func (fc *FontCache) CacheGlyph(
	glyphCode uint32,
	glyphIndex uint32,
	dataSize uint32,
	dataType GlyphDataType,
	bounds basics.Rect[int],
	advanceX, advanceY float64,
) *GlyphCache {
	msb := (glyphCode >> 8) & 0xFF
	lsb := glyphCode & 0xFF

	// Allocate second-level array if needed
	if fc.glyphs[msb] == nil {
		fc.glyphs[msb] = array.AllocateType[[256]*GlyphCache](fc.allocator)
		if fc.glyphs[msb] == nil {
			return nil
		}
		// Initialize all pointers to nil
		*fc.glyphs[msb] = [256]*GlyphCache{}
	}

	// Check if glyph already exists
	if fc.glyphs[msb][lsb] != nil {
		return nil // Already exists, do not overwrite
	}

	// Allocate glyph cache entry
	glyph := array.AllocateType[GlyphCache](fc.allocator)
	if glyph == nil {
		return nil
	}

	// Allocate glyph data
	data := fc.allocator.AllocateBytes(int(dataSize))
	if data == nil {
		return nil
	}

	// Initialize glyph cache
	glyph.GlyphIndex = glyphIndex
	glyph.Data = data
	glyph.DataSize = dataSize
	glyph.DataType = dataType
	glyph.Bounds = bounds
	glyph.AdvanceX = advanceX
	glyph.AdvanceY = advanceY

	// Store in cache
	fc.glyphs[msb][lsb] = glyph
	return glyph
}

// Reset clears all allocated memory and resets the cache.
func (fc *FontCache) Reset() {
	fc.allocator.RemoveAll()
	fc.glyphs = [256]*[256]*GlyphCache{}
	fc.fontSignature = ""
}

// FontCachePool manages multiple font caches with LRU eviction.
// This corresponds to AGG's font_cache_pool class.
type FontCachePool struct {
	fonts    []*FontCache // Array of font caches
	maxFonts uint32       // Maximum number of fonts to cache
	numFonts uint32       // Current number of cached fonts
	curFont  *FontCache   // Currently active font cache
}

// NewFontCachePool creates a new font cache pool with default capacity.
func NewFontCachePool() *FontCachePool {
	return NewFontCachePoolWithCapacity(32)
}

// NewFontCachePoolWithCapacity creates a new font cache pool with specified capacity.
func NewFontCachePoolWithCapacity(maxFonts uint32) *FontCachePool {
	return &FontCachePool{
		fonts:    make([]*FontCache, maxFonts),
		maxFonts: maxFonts,
		numFonts: 0,
		curFont:  nil,
	}
}

// SetFont sets the current font by signature, creating or reusing cache as needed.
func (fcp *FontCachePool) SetFont(fontSignature string, resetCache bool) {
	idx := fcp.findFont(fontSignature)
	if idx >= 0 {
		// Font exists
		if resetCache {
			// Reset existing cache
			fcp.fonts[idx].Reset()
			fcp.fonts[idx].SetSignature(fontSignature)
		}
		fcp.curFont = fcp.fonts[idx]
	} else {
		// Font doesn't exist, create new cache
		if fcp.numFonts >= fcp.maxFonts {
			// Pool is full, remove oldest font (LRU)
			fcp.fonts[0].Reset()
			// Shift fonts down
			copy(fcp.fonts[0:], fcp.fonts[1:fcp.maxFonts])
			fcp.numFonts = fcp.maxFonts - 1
		}

		// Create new font cache
		newFont := NewFontCache()
		newFont.SetSignature(fontSignature)
		fcp.fonts[fcp.numFonts] = newFont
		fcp.curFont = newFont
		fcp.numFonts++
	}
}

// Font returns the current font cache.
func (fcp *FontCachePool) Font() *FontCache {
	return fcp.curFont
}

// FindGlyph finds a glyph in the current font cache.
func (fcp *FontCachePool) FindGlyph(glyphCode uint32) *GlyphCache {
	if fcp.curFont != nil {
		return fcp.curFont.FindGlyph(glyphCode)
	}
	return nil
}

// CacheGlyph caches a glyph in the current font cache.
func (fcp *FontCachePool) CacheGlyph(
	glyphCode uint32,
	glyphIndex uint32,
	dataSize uint32,
	dataType GlyphDataType,
	bounds basics.Rect[int],
	advanceX, advanceY float64,
) *GlyphCache {
	if fcp.curFont != nil {
		return fcp.curFont.CacheGlyph(
			glyphCode, glyphIndex, dataSize, dataType,
			bounds, advanceX, advanceY,
		)
	}
	return nil
}

// findFont finds a font by signature, returning its index or -1 if not found.
func (fcp *FontCachePool) findFont(fontSignature string) int {
	for i := uint32(0); i < fcp.numFonts; i++ {
		if fcp.fonts[i].FontIs(fontSignature) {
			return int(i)
		}
	}
	return -1
}

// Reset clears all font caches and resets the pool.
func (fcp *FontCachePool) Reset() {
	for i := uint32(0); i < fcp.numFonts; i++ {
		if fcp.fonts[i] != nil {
			fcp.fonts[i].Reset()
		}
	}
	fcp.numFonts = 0
	fcp.curFont = nil
}

// GlyphRendering represents different glyph rendering modes.
// This corresponds to AGG's glyph_rendering enum.
type GlyphRendering uint32

const (
	GlyphRenNativeMono  GlyphRendering = iota // Native monochrome rendering
	GlyphRenNativeGray8                       // Native 8-bit grayscale rendering
	GlyphRenOutline                           // Vector outline rendering
	GlyphRenAggMono                           // AGG monochrome rendering
	GlyphRenAggGray8                          // AGG 8-bit grayscale rendering
)

// String returns string representation of glyph rendering mode.
func (gr GlyphRendering) String() string {
	switch gr {
	case GlyphRenNativeMono:
		return "native_mono"
	case GlyphRenNativeGray8:
		return "native_gray8"
	case GlyphRenOutline:
		return "outline"
	case GlyphRenAggMono:
		return "agg_mono"
	case GlyphRenAggGray8:
		return "agg_gray8"
	default:
		return "unknown"
	}
}

// FontEngine defines the interface that font engines must implement
// to work with FontCacheManager.
type FontEngine interface {
	// FontSignature returns a unique identifier for the current font
	FontSignature() string

	// ChangeStamp returns a value that changes when font parameters change
	ChangeStamp() int

	// PrepareGlyph prepares a glyph for rendering and returns true if successful
	PrepareGlyph(glyphCode uint32) bool

	// GlyphIndex returns the glyph index of the last prepared glyph
	GlyphIndex() uint32

	// DataSize returns the data size of the last prepared glyph
	DataSize() uint32

	// DataType returns the data type of the last prepared glyph
	DataType() GlyphDataType

	// Bounds returns the bounds of the last prepared glyph
	Bounds() basics.Rect[int]

	// AdvanceX returns the horizontal advance of the last prepared glyph
	AdvanceX() float64

	// AdvanceY returns the vertical advance of the last prepared glyph
	AdvanceY() float64

	// WriteGlyphTo writes the glyph data to the provided buffer
	WriteGlyphTo(data []byte)

	// AddKerning adds kerning adjustment between two glyphs
	AddKerning(first, second uint32, x, y *float64) bool
}

// FontCacheManager manages font caching and glyph rendering.
// This is the Go equivalent of AGG's font_cache_manager template class.
type FontCacheManager[T FontEngine] struct {
	fonts       *FontCachePool // Font cache pool
	engine      T              // Font engine
	changeStamp int            // Last known change stamp
	prevGlyph   *GlyphCache    // Previous glyph for kerning
	lastGlyph   *GlyphCache    // Last processed glyph
}

// NewFontCacheManager creates a new font cache manager with the given engine.
func NewFontCacheManager[T FontEngine](engine T) *FontCacheManager[T] {
	return NewFontCacheManagerWithCapacity(engine, 32)
}

// NewFontCacheManagerWithCapacity creates a new font cache manager with custom capacity.
func NewFontCacheManagerWithCapacity[T FontEngine](engine T, maxFonts uint32) *FontCacheManager[T] {
	return &FontCacheManager[T]{
		fonts:       NewFontCachePoolWithCapacity(maxFonts),
		engine:      engine,
		changeStamp: -1,
		prevGlyph:   nil,
		lastGlyph:   nil,
	}
}

// ResetLastGlyph clears the previous and last glyph references.
func (fcm *FontCacheManager[T]) ResetLastGlyph() {
	fcm.prevGlyph = nil
	fcm.lastGlyph = nil
}

// Glyph retrieves or caches a glyph by character code.
func (fcm *FontCacheManager[T]) Glyph(glyphCode uint32) *GlyphCache {
	fcm.synchronize()

	gl := fcm.fonts.FindGlyph(glyphCode)
	if gl != nil {
		// Found in cache
		fcm.prevGlyph = fcm.lastGlyph
		fcm.lastGlyph = gl
		return gl
	}

	// Not in cache, prepare and cache new glyph
	if fcm.engine.PrepareGlyph(glyphCode) {
		fcm.prevGlyph = fcm.lastGlyph
		fcm.lastGlyph = fcm.fonts.CacheGlyph(
			glyphCode,
			fcm.engine.GlyphIndex(),
			fcm.engine.DataSize(),
			fcm.engine.DataType(),
			fcm.engine.Bounds(),
			fcm.engine.AdvanceX(),
			fcm.engine.AdvanceY(),
		)
		if fcm.lastGlyph != nil {
			fcm.engine.WriteGlyphTo(fcm.lastGlyph.Data)
		}
		return fcm.lastGlyph
	}

	return nil
}

// PrevGlyph returns the previous glyph for kerning calculations.
func (fcm *FontCacheManager[T]) PrevGlyph() *GlyphCache {
	return fcm.prevGlyph
}

// LastGlyph returns the last processed glyph.
func (fcm *FontCacheManager[T]) LastGlyph() *GlyphCache {
	return fcm.lastGlyph
}

// AddKerning adds kerning adjustment between the previous and last glyphs.
func (fcm *FontCacheManager[T]) AddKerning(x, y *float64) bool {
	if fcm.prevGlyph != nil && fcm.lastGlyph != nil {
		return fcm.engine.AddKerning(
			fcm.prevGlyph.GlyphIndex,
			fcm.lastGlyph.GlyphIndex,
			x, y,
		)
	}
	return false
}

// Precache caches a range of glyphs for improved performance.
func (fcm *FontCacheManager[T]) Precache(from, to uint32) {
	for glyphCode := from; glyphCode <= to; glyphCode++ {
		fcm.Glyph(glyphCode)
	}
}

// ResetCache resets the current font cache.
func (fcm *FontCacheManager[T]) ResetCache() {
	fcm.fonts.SetFont(fcm.engine.FontSignature(), true)
	fcm.changeStamp = fcm.engine.ChangeStamp()
	fcm.prevGlyph = nil
	fcm.lastGlyph = nil
}

// synchronize ensures the cache is synchronized with the font engine.
func (fcm *FontCacheManager[T]) synchronize() {
	if fcm.changeStamp != fcm.engine.ChangeStamp() {
		// Change stamp changed - reset cache to ensure consistency
		fcm.fonts.SetFont(fcm.engine.FontSignature(), true)
		fcm.changeStamp = fcm.engine.ChangeStamp()
		fcm.prevGlyph = nil
		fcm.lastGlyph = nil
	}
}

// Stats returns cache statistics for debugging and optimization.
type CacheStats struct {
	NumFonts    uint32 // Number of cached fonts
	MaxFonts    uint32 // Maximum font capacity
	CurrentFont string // Current font signature
}

// Stats returns cache statistics.
func (fcm *FontCacheManager[T]) Stats() CacheStats {
	currentFont := ""
	if fcm.fonts.curFont != nil {
		currentFont = fcm.fonts.curFont.fontSignature
	}

	return CacheStats{
		NumFonts:    fcm.fonts.numFonts,
		MaxFonts:    fcm.fonts.maxFonts,
		CurrentFont: currentFont,
	}
}
