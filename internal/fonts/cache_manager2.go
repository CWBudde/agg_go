// Package fonts provides enhanced font management and caching functionality for AGG.
// This package implements the advanced font cache manager (version 2) with improved
// cache management and font face handling.
package fonts

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// V2 namespace equivalent - fman (font manager) from AGG
// All types and constants are prefixed with Fman to match the C++ namespace

// FmanGlyphDataType represents the type of glyph data in the enhanced cache system.
// This corresponds to AGG's fman::glyph_data_type enum.
type FmanGlyphDataType uint32

const (
	FmanGlyphDataInvalid FmanGlyphDataType = iota // Invalid/empty glyph data
	FmanGlyphDataMono                             // Monochrome bitmap glyph
	FmanGlyphDataGray8                            // 8-bit grayscale glyph
	FmanGlyphDataOutline                          // Vector outline glyph
)

// String returns string representation of glyph data type.
func (gdt FmanGlyphDataType) String() string {
	switch gdt {
	case FmanGlyphDataInvalid:
		return "invalid"
	case FmanGlyphDataMono:
		return "mono"
	case FmanGlyphDataGray8:
		return "gray8"
	case FmanGlyphDataOutline:
		return "outline"
	default:
		return "unknown"
	}
}

// FmanGlyphRendering represents different glyph rendering modes in v2.
// This corresponds to AGG's fman::glyph_rendering enum.
type FmanGlyphRendering uint32

const (
	FmanGlyphRenNativeMono  FmanGlyphRendering = iota // Native monochrome rendering
	FmanGlyphRenNativeGray8                           // Native 8-bit grayscale rendering
	FmanGlyphRenOutline                               // Vector outline rendering
	FmanGlyphRenAggMono                               // AGG monochrome rendering
	FmanGlyphRenAggGray8                              // AGG 8-bit grayscale rendering
)

// String returns string representation of glyph rendering mode.
func (gr FmanGlyphRendering) String() string {
	switch gr {
	case FmanGlyphRenNativeMono:
		return "native_mono"
	case FmanGlyphRenNativeGray8:
		return "native_gray8"
	case FmanGlyphRenOutline:
		return "outline"
	case FmanGlyphRenAggMono:
		return "agg_mono"
	case FmanGlyphRenAggGray8:
		return "agg_gray8"
	default:
		return "unknown"
	}
}

// FmanCachedGlyph stores cached glyph data with enhanced metadata including font reference.
// This corresponds to AGG's fman::cached_glyph struct.
type FmanCachedGlyph struct {
	CachedFont *FmanCachedFont   // Reference to the cached font
	GlyphCode  uint32            // Character code
	GlyphIndex uint32            // Glyph index in font
	Data       []byte            // Glyph bitmap or outline data
	DataSize   uint32            // Size of data in bytes
	DataType   FmanGlyphDataType // Type of glyph data
	Bounds     basics.Rect[int]  // Glyph bounding rectangle
	AdvanceX   float64           // Horizontal advance
	AdvanceY   float64           // Vertical advance
}

// FmanCachedGlyphs manages a collection of cached glyphs with optimized storage.
// This corresponds to AGG's fman::cached_glyphs class.
type FmanCachedGlyphs struct {
	allocator *array.BlockAllocator       // Memory allocator for glyph data
	glyphs    [256]*[256]*FmanCachedGlyph // Two-level glyph lookup table
}

const (
	// FmanDefaultBlockSize is the default block size for the enhanced cache allocator.
	// This matches AGG's fman block_size constant (16384-16).
	FmanDefaultBlockSize = 16384 - 16
)

// NewFmanCachedGlyphs creates a new enhanced glyph cache.
func NewFmanCachedGlyphs() *FmanCachedGlyphs {
	return &FmanCachedGlyphs{
		allocator: array.NewBlockAllocator(FmanDefaultBlockSize),
		glyphs:    [256]*[256]*FmanCachedGlyph{},
	}
}

// NewFmanCachedGlyphsWithBlockSize creates a new enhanced glyph cache with custom block size.
func NewFmanCachedGlyphsWithBlockSize(blockSize int) *FmanCachedGlyphs {
	return &FmanCachedGlyphs{
		allocator: array.NewBlockAllocator(blockSize),
		glyphs:    [256]*[256]*FmanCachedGlyph{},
	}
}

// FindGlyph finds a cached glyph by character code.
// Returns nil if glyph is not cached.
func (cg *FmanCachedGlyphs) FindGlyph(glyphCode uint32) *FmanCachedGlyph {
	msb := (glyphCode >> 8) & 0xFF
	if cg.glyphs[msb] != nil {
		lsb := glyphCode & 0xFF
		return cg.glyphs[msb][lsb]
	}
	return nil
}

// CacheGlyph caches a new glyph with the given parameters.
// Returns the cached glyph or nil if glyph already exists.
func (cg *FmanCachedGlyphs) CacheGlyph(
	cachedFont *FmanCachedFont,
	glyphCode uint32,
	glyphIndex uint32,
	dataSize uint32,
	dataType FmanGlyphDataType,
	bounds basics.Rect[int],
	advanceX, advanceY float64,
) *FmanCachedGlyph {
	msb := (glyphCode >> 8) & 0xFF
	lsb := glyphCode & 0xFF

	// Allocate second-level array if needed
	if cg.glyphs[msb] == nil {
		cg.glyphs[msb] = array.AllocateType[[256]*FmanCachedGlyph](cg.allocator)
		if cg.glyphs[msb] == nil {
			return nil
		}
		// Initialize all pointers to nil
		*cg.glyphs[msb] = [256]*FmanCachedGlyph{}
	}

	// Check if glyph already exists
	if cg.glyphs[msb][lsb] != nil {
		return nil // Already exists, do not overwrite
	}

	// Allocate glyph cache entry
	glyph := array.AllocateType[FmanCachedGlyph](cg.allocator)
	if glyph == nil {
		return nil
	}

	// Allocate glyph data
	data := cg.allocator.AllocateBytes(int(dataSize))
	if data == nil {
		return nil
	}

	// Initialize glyph cache
	glyph.CachedFont = cachedFont
	glyph.GlyphCode = glyphCode
	glyph.GlyphIndex = glyphIndex
	glyph.Data = data
	glyph.DataSize = dataSize
	glyph.DataType = dataType
	glyph.Bounds = bounds
	glyph.AdvanceX = advanceX
	glyph.AdvanceY = advanceY

	// Store in cache
	cg.glyphs[msb][lsb] = glyph
	return glyph
}

// Reset clears all allocated memory and resets the cache.
func (cg *FmanCachedGlyphs) Reset() {
	cg.allocator.RemoveAll()
	cg.glyphs = [256]*[256]*FmanCachedGlyph{}
}

// LoadedFace defines the interface for loaded font faces that work with the enhanced cache manager.
// This corresponds to the concept of FontEngine::loaded_face in AGG.
type LoadedFace interface {
	// Height returns the font height
	Height() float64

	// Width returns the font width
	Width() float64

	// Ascent returns the font ascent
	Ascent() float64

	// Descent returns the font descent
	Descent() float64

	// AscentB returns the bold font ascent
	AscentB() float64

	// DescentB returns the bold font descent
	DescentB() float64

	// SelectInstance selects the font instance with given parameters
	SelectInstance(height, width float64, hinting bool, rendering FmanGlyphRendering)

	// PrepareGlyph prepares a glyph for rendering and returns preparation data
	PrepareGlyph(cp uint32) (PreparedGlyph, bool)

	// WriteGlyphTo writes the prepared glyph data to the provided buffer
	WriteGlyphTo(prepared *PreparedGlyph, data []byte)

	// AddKerning adds kerning adjustment between two glyphs
	AddKerning(first, second uint32, x, y *float64) bool
}

// PreparedGlyph contains the data for a prepared glyph ready for caching.
// This corresponds to AGG's FontEngine::prepared_glyph concept.
type PreparedGlyph struct {
	GlyphCode  uint32            // Character code
	GlyphIndex uint32            // Glyph index in font
	DataSize   uint32            // Size of glyph data
	DataType   FmanGlyphDataType // Type of glyph data
	Bounds     basics.Rect[int]  // Glyph bounding rectangle
	AdvanceX   float64           // Horizontal advance
	AdvanceY   float64           // Vertical advance
}

// FmanCachedFont encapsulates a cached font with its face, metrics, and glyph cache.
// This corresponds to AGG's fman::font_cache_manager::cached_font struct.
type FmanCachedFont struct {
	face         LoadedFace         // Font face interface
	height       float64            // Font height
	width        float64            // Font width
	hinting      bool               // Hinting enabled
	rendering    FmanGlyphRendering // Rendering mode
	faceHeight   float64            // Cached face height
	faceWidth    float64            // Cached face width
	faceAscent   float64            // Cached face ascent
	faceDescent  float64            // Cached face descent
	faceAscentB  float64            // Cached face bold ascent
	faceDescentB float64            // Cached face bold descent
	glyphs       *FmanCachedGlyphs  // Glyph cache
}

// NewFmanCachedFont creates a new cached font with the given parameters.
func NewFmanCachedFont(
	face LoadedFace,
	height, width float64,
	hinting bool,
	rendering FmanGlyphRendering,
) *FmanCachedFont {
	cf := &FmanCachedFont{
		face:      face,
		height:    height,
		width:     width,
		hinting:   hinting,
		rendering: rendering,
		glyphs:    NewFmanCachedGlyphs(),
	}

	// Select the face and cache metrics
	cf.selectFace()
	cf.faceHeight = cf.face.Height()
	cf.faceWidth = cf.face.Width()
	cf.faceAscent = cf.face.Ascent()
	cf.faceDescent = cf.face.Descent()
	cf.faceAscentB = cf.face.AscentB()
	cf.faceDescentB = cf.face.DescentB()

	return cf
}

// Height returns the cached font height.
func (cf *FmanCachedFont) Height() float64 {
	return cf.faceHeight
}

// Width returns the cached font width.
func (cf *FmanCachedFont) Width() float64 {
	return cf.faceWidth
}

// Ascent returns the cached font ascent.
func (cf *FmanCachedFont) Ascent() float64 {
	return cf.faceAscent
}

// Descent returns the cached font descent.
func (cf *FmanCachedFont) Descent() float64 {
	return cf.faceDescent
}

// AscentB returns the cached font bold ascent.
func (cf *FmanCachedFont) AscentB() float64 {
	return cf.faceAscentB
}

// DescentB returns the cached font bold descent.
func (cf *FmanCachedFont) DescentB() float64 {
	return cf.faceDescentB
}

// AddKerning adds kerning adjustment between two cached glyphs.
func (cf *FmanCachedFont) AddKerning(first, second *FmanCachedGlyph, x, y *float64) bool {
	if first == nil || second == nil {
		return false
	}
	cf.selectFace()
	return cf.face.AddKerning(first.GlyphIndex, second.GlyphIndex, x, y)
}

// selectFace selects the font face with current parameters.
func (cf *FmanCachedFont) selectFace() {
	cf.face.SelectInstance(cf.height, cf.width, cf.hinting, cf.rendering)
}

// GetGlyph retrieves or caches a glyph by character code.
func (cf *FmanCachedFont) GetGlyph(cp uint32) *FmanCachedGlyph {
	glyph := cf.glyphs.FindGlyph(cp)
	if glyph == nil {
		// Glyph not in cache, prepare and cache it
		cf.selectFace()
		prepared, success := cf.face.PrepareGlyph(cp)
		if success {
			glyph = cf.glyphs.CacheGlyph(
				cf, // Reference to this cached font
				prepared.GlyphCode,
				prepared.GlyphIndex,
				prepared.DataSize,
				prepared.DataType,
				prepared.Bounds,
				prepared.AdvanceX,
				prepared.AdvanceY,
			)
			if glyph != nil {
				cf.face.WriteGlyphTo(&prepared, glyph.Data)
			}
		}
	}
	return glyph
}

// Reset clears the glyph cache for this font.
func (cf *FmanCachedFont) Reset() {
	cf.glyphs.Reset()
}

// FontEngine2 defines the interface for enhanced font engines working with v2 cache manager.
// This is a simplified interface compared to v1, focusing on adaptor management.
type FontEngine2 interface {
	// PathAdaptor returns the path adaptor for outline glyphs
	PathAdaptor() PathAdaptorType

	// Gray8Adaptor returns the gray8 adaptor for grayscale glyphs
	Gray8Adaptor() Gray8AdaptorType

	// Gray8Scanline returns the gray8 scanline for grayscale rendering
	Gray8Scanline() Gray8ScanlineType

	// MonoAdaptor returns the mono adaptor for monochrome glyphs
	MonoAdaptor() MonoAdaptorType

	// MonoScanline returns the mono scanline for monochrome rendering
	MonoScanline() MonoScanlineType
}

// FmanFontCacheManager2 is the enhanced font cache manager (version 2).
// This corresponds to AGG's fman::font_cache_manager template class.
// Unlike v1, this focuses primarily on adaptor management rather than cache management.
type FmanFontCacheManager2[T FontEngine2] struct {
	engine T // Font engine

	// Adaptors for different glyph types (properly typed)
	pathAdaptor   PathAdaptorType   // Path adaptor from engine
	gray8Adaptor  Gray8AdaptorType  // Gray8 adaptor from engine
	gray8Scanline Gray8ScanlineType // Gray8 scanline from engine
	monoAdaptor   MonoAdaptorType   // Mono adaptor from engine
	monoScanline  MonoScanlineType  // Mono scanline from engine
}

// NewFmanFontCacheManager2 creates a new enhanced font cache manager.
func NewFmanFontCacheManager2[T FontEngine2](engine T) *FmanFontCacheManager2[T] {
	return &FmanFontCacheManager2[T]{
		engine:        engine,
		pathAdaptor:   engine.PathAdaptor(),
		gray8Adaptor:  engine.Gray8Adaptor(),
		gray8Scanline: engine.Gray8Scanline(),
		monoAdaptor:   engine.MonoAdaptor(),
		monoScanline:  engine.MonoScanline(),
	}
}

// InitEmbeddedAdaptors initializes the embedded adaptors for a cached glyph.
// This corresponds to AGG's init_embedded_adaptors method.
func (fcm *FmanFontCacheManager2[T]) InitEmbeddedAdaptors(
	gl *FmanCachedGlyph,
	x, y float64,
	scale float64,
) {
	if gl == nil {
		return
	}

	switch gl.DataType {
	case FmanGlyphDataMono:
		// Initialize mono adaptor - now type-safe
		fcm.monoAdaptor.InitGlyph(gl.Data, gl.DataSize, x, y)

	case FmanGlyphDataGray8:
		// Initialize gray8 adaptor - now type-safe
		fcm.gray8Adaptor.InitGlyph(gl.Data, gl.DataSize, x, y)

	case FmanGlyphDataOutline:
		// Initialize path adaptor with scale - now type-safe
		fcm.pathAdaptor.InitWithScale(gl.Data, gl.DataSize, x, y, scale)

	default:
		// Invalid or unknown data type
		return
	}
}

// PathAdaptor returns the path adaptor for outline glyphs.
func (fcm *FmanFontCacheManager2[T]) PathAdaptor() PathAdaptorType {
	return fcm.pathAdaptor
}

// Gray8Adaptor returns the gray8 adaptor for grayscale glyphs.
func (fcm *FmanFontCacheManager2[T]) Gray8Adaptor() Gray8AdaptorType {
	return fcm.gray8Adaptor
}

// Gray8Scanline returns the gray8 scanline for grayscale rendering.
func (fcm *FmanFontCacheManager2[T]) Gray8Scanline() Gray8ScanlineType {
	return fcm.gray8Scanline
}

// MonoAdaptor returns the mono adaptor for monochrome glyphs.
func (fcm *FmanFontCacheManager2[T]) MonoAdaptor() MonoAdaptorType {
	return fcm.monoAdaptor
}

// MonoScanline returns the mono scanline for monochrome rendering.
func (fcm *FmanFontCacheManager2[T]) MonoScanline() MonoScanlineType {
	return fcm.monoScanline
}

// Engine returns the underlying font engine.
func (fcm *FmanFontCacheManager2[T]) Engine() T {
	return fcm.engine
}

// FontMetrics provides detailed font metrics information.
type FontMetrics struct {
	Height   float64 // Font height
	Width    float64 // Font width
	Ascent   float64 // Font ascent
	Descent  float64 // Font descent
	AscentB  float64 // Bold font ascent
	DescentB float64 // Bold font descent
}

// GetFontMetrics extracts font metrics from a cached font.
func GetFontMetrics(cf *FmanCachedFont) FontMetrics {
	return FontMetrics{
		Height:   cf.Height(),
		Width:    cf.Width(),
		Ascent:   cf.Ascent(),
		Descent:  cf.Descent(),
		AscentB:  cf.AscentB(),
		DescentB: cf.DescentB(),
	}
}

// CacheStats2 provides enhanced cache statistics for debugging and optimization.
type CacheStats2 struct {
	AllocatedBlocks int // Number of allocated memory blocks
	TotalMemory     int // Total allocated memory in bytes
	UsedMemory      int // Used memory in bytes
	CachedGlyphs    int // Number of cached glyphs (approximate)
}

// GetCacheStats returns cache statistics for a cached font.
func GetCacheStats(cf *FmanCachedFont) CacheStats2 {
	stats := cf.glyphs.allocator.Stats()

	// Count cached glyphs (approximate)
	cachedGlyphs := 0
	for i := 0; i < 256; i++ {
		if cf.glyphs.glyphs[i] != nil {
			for j := 0; j < 256; j++ {
				if cf.glyphs.glyphs[i][j] != nil {
					cachedGlyphs++
				}
			}
		}
	}

	return CacheStats2{
		AllocatedBlocks: stats.NumBlocks,
		TotalMemory:     stats.TotalAllocated,
		UsedMemory:      stats.TotalUsed,
		CachedGlyphs:    cachedGlyphs,
	}
}
