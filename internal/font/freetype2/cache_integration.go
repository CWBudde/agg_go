//go:build freetype

// Package freetype2 provides integration with the enhanced cache manager v2.
// This bridges the new FreeType2 engine with the fman cache system.
package freetype2

import (
	"fmt"

	"agg_go/internal/fonts"
	"agg_go/internal/path"
)

// CacheManager2 integrates the FreeType2 engine with the separate fman cache path.
// It mirrors the adaptor-facing behavior of AGG's fman::font_cache_manager
// from agg_font_cache_manager2.h, while remaining distinct from Agg2D's v1 stack.
//
// The glyph cache itself is delegated to internal/fonts.FmanCachedFont so this
// wrapper stays focused on package-boundary adaptation and current-face context.
type CacheManager2 struct {
	fontEngine        FontEngineInterface
	currentFont       LoadedFaceInterface     // Reference to current font context
	currentCachedFont *fonts.FmanCachedFont   // Per-face-instance cache, mirroring AGG's cached_font more closely
	pathAdaptor       path.IntegerPathAdaptor // Either *path.SerializedIntegerPathAdaptor[int16] or *path.SerializedIntegerPathAdaptor[int32]
	gray8Adaptor      *Gray8AdaptorWrapper
	monoAdaptor       *MonoAdaptorWrapper
	lastError         error
	cacheContext      cacheManagerFontContext
}

type cacheManagerFontContext struct {
	face       LoadedFaceInterface
	resolution uint32
	height     float64
	width      float64
	hinting    bool
	flipY      bool
	rendering  GlyphRendering
	charMap    CharEncoding
	transform  [6]float64
}

// NewCacheManager2 creates a new cache manager with the specified font engine.
func NewCacheManager2(fontEngine FontEngineInterface) *CacheManager2 {
	cm := &CacheManager2{
		fontEngine: fontEngine,
	}

	// Initialize adaptors based on engine type
	cm.initializeAdaptors()

	return cm
}

// initializeAdaptors sets up the appropriate adaptors based on the engine type.
func (cm *CacheManager2) initializeAdaptors() {
	if cm.fontEngine == nil {
		return
	}

	cm.pathAdaptor = cm.fontEngine.PathAdaptorInterface()

	adaptorTypes := cm.fontEngine.AdaptorTypes()
	if adaptorTypes != nil {
		cm.gray8Adaptor = NewGray8AdaptorWrapper(adaptorTypes.Gray8Adaptor)
		cm.monoAdaptor = NewMonoAdaptorWrapper(adaptorTypes.MonoAdaptor)
	}
}

// Glyph retrieves a glyph from the cache, loading it if necessary.
// This is part of the port's standalone fman cache wrapper; the surrounding
// cache ownership is Go-managed rather than a direct one-to-one AGG object.
func (cm *CacheManager2) Glyph(charCode uint32) *fonts.FmanCachedGlyph {
	if cm.currentFont != nil {
		cm.ensureCachedFontForCurrentContext()
	}

	if cm.currentFont == nil {
		cm.lastError = fmt.Errorf("no font loaded for glyph preparation")
		return nil
	}
	if cm.currentCachedFont == nil {
		cm.lastError = fmt.Errorf("no cached font context available for glyph preparation")
		return nil
	}

	cachedGlyph := cm.currentCachedFont.GetGlyph(charCode)
	if cachedGlyph == nil {
		cm.lastError = fmt.Errorf("glyph not found for character code %d", charCode)
		return nil
	}
	return cachedGlyph
}

// PathAdaptor returns the path adaptor for vector font rendering.
func (cm *CacheManager2) PathAdaptor() path.IntegerPathAdaptor {
	return cm.pathAdaptor
}

// Gray8Adaptor returns the gray8 scanline adaptor.
func (cm *CacheManager2) Gray8Adaptor() *Gray8AdaptorWrapper {
	return cm.gray8Adaptor
}

// MonoAdaptor returns the mono scanline adaptor.
func (cm *CacheManager2) MonoAdaptor() *MonoAdaptorWrapper {
	return cm.monoAdaptor
}

// InitEmbeddedAdaptors initializes the embedded adaptors for a specific glyph.
// This corresponds to AGG's init_embedded_adaptors method.
func (cm *CacheManager2) InitEmbeddedAdaptors(glyph *fonts.FmanCachedGlyph, x, y float64) {
	if glyph == nil {
		return
	}

	switch glyph.DataType {
	case fonts.FmanGlyphDataGray8:
		// Initialize gray8 adaptor with glyph data
		if cm.gray8Adaptor != nil && len(glyph.Data) > 0 {
			cm.gray8Adaptor.InitGlyph(glyph.Data, glyph.DataSize, x, y)
		}

	case fonts.FmanGlyphDataMono:
		// Initialize mono adaptor with glyph data
		if cm.monoAdaptor != nil && len(glyph.Data) > 0 {
			cm.monoAdaptor.InitGlyph(glyph.Data, glyph.DataSize, x, y)
		}

	case fonts.FmanGlyphDataOutline:
		// For outline glyphs, the path data should be available in the path adaptor
		// This is handled by the font engine's outline decomposition
		break
	}
}

// AddKerning adds kerning adjustment between two characters.
// This uses the font engine's kerning support.
func (cm *CacheManager2) AddKerning(x, y *float64, first, second uint32) {
	// This requires access to a loaded face
	// In practice, this would be called on the currently active font face
	if cm.currentFont == nil {
		cm.lastError = fmt.Errorf("no font loaded for kerning calculation")
		return
	}

	// Get kerning adjustment from the loaded face
	dx, dy := cm.currentFont.AddKerning(first, second)

	// Apply the kerning adjustment
	if x != nil {
		*x += dx
	}
	if y != nil {
		*y += dy
	}
}

// SetCurrentFont sets the current font face for glyph preparation and kerning.
func (cm *CacheManager2) SetCurrentFont(face LoadedFaceInterface) {
	cm.currentFont = face
	cm.currentCachedFont = nil
	cm.cacheContext = cacheManagerFontContext{}
}

// GetCurrentFont returns the current font face.
func (cm *CacheManager2) GetCurrentFont() LoadedFaceInterface {
	return cm.currentFont
}

// LastError returns the last error that occurred.
func (cm *CacheManager2) LastError() error {
	return cm.lastError
}

// Close cleans up resources used by the cache manager.
func (cm *CacheManager2) Close() error {
	// Clean up adaptors
	cm.pathAdaptor = nil
	cm.gray8Adaptor = nil
	cm.monoAdaptor = nil

	// Clean up cached glyphs
	if cm.currentCachedFont != nil {
		cm.currentCachedFont.Reset()
		cm.currentCachedFont = nil
	}
	cm.currentFont = nil
	cm.cacheContext = cacheManagerFontContext{}

	return nil
}

func (cm *CacheManager2) ensureCachedFontForCurrentContext() {
	if cm.currentFont == nil {
		return
	}

	next := cacheManagerFontContext{
		face:       cm.currentFont,
		resolution: cm.currentFont.Resolution(),
		height:     cm.currentFont.Height(),
		width:      cm.currentFont.Width(),
		hinting:    cm.currentFont.Hinting(),
		flipY:      cm.currentFont.FlipY(),
		rendering:  cm.currentFont.Rendering(),
		charMap:    cm.currentFont.CharMap(),
	}
	if affine := cm.currentFont.Transform(); affine != nil {
		next.transform = affine.ToArray()
	}

	if cm.cacheContext != next || cm.currentCachedFont == nil {
		if cm.currentCachedFont != nil {
			cm.currentCachedFont.Reset()
		}
		cm.currentCachedFont = fonts.NewFmanCachedFont(
			freetypeLoadedFaceAdapter{face: cm.currentFont},
			cm.currentFont.Height(),
			cm.currentFont.Width(),
			cm.currentFont.Hinting(),
			toFmanGlyphRendering(cm.currentFont.Rendering()),
		)
	}
	cm.cacheContext = next
}

type freetypeLoadedFaceAdapter struct {
	face LoadedFaceInterface
}

func (a freetypeLoadedFaceAdapter) Height() float64  { return a.face.Height() }
func (a freetypeLoadedFaceAdapter) Width() float64   { return a.face.Width() }
func (a freetypeLoadedFaceAdapter) Ascent() float64  { return a.face.Ascent() }
func (a freetypeLoadedFaceAdapter) Descent() float64 { return a.face.Descent() }
func (a freetypeLoadedFaceAdapter) AscentB() float64 { return a.face.AscentB() }
func (a freetypeLoadedFaceAdapter) DescentB() float64 {
	return a.face.DescentB()
}
func (a freetypeLoadedFaceAdapter) SelectInstance(height, width float64, hinting bool, rendering fonts.FmanGlyphRendering) {
	a.face.SelectInstance(height, width, hinting, fromFmanGlyphRendering(rendering))
}
func (a freetypeLoadedFaceAdapter) PrepareGlyph(cp uint32) (fonts.PreparedGlyph, bool) {
	prepared, ok := a.face.PrepareGlyph(cp)
	if !ok || prepared == nil {
		return fonts.PreparedGlyph{}, false
	}
	return fonts.PreparedGlyph{
		GlyphCode:  prepared.GlyphCode,
		GlyphIndex: prepared.GlyphIndex,
		DataSize:   prepared.DataSize,
		DataType:   prepared.DataType,
		Bounds:     prepared.Bounds,
		AdvanceX:   prepared.AdvanceX,
		AdvanceY:   prepared.AdvanceY,
	}, true
}
func (a freetypeLoadedFaceAdapter) WriteGlyphTo(prepared *fonts.PreparedGlyph, data []byte) {
	if prepared == nil {
		return
	}
	a.face.WriteGlyphTo(&PreparedGlyph{
		GlyphCode:  prepared.GlyphCode,
		GlyphIndex: prepared.GlyphIndex,
		DataSize:   prepared.DataSize,
		DataType:   prepared.DataType,
		Bounds:     prepared.Bounds,
		AdvanceX:   prepared.AdvanceX,
		AdvanceY:   prepared.AdvanceY,
	}, data)
}
func (a freetypeLoadedFaceAdapter) AddKerning(first, second uint32, x, y *float64) bool {
	dx, dy := a.face.AddKerning(first, second)
	if x != nil {
		*x += dx
	}
	if y != nil {
		*y += dy
	}
	return true
}

func toFmanGlyphRendering(rendering GlyphRendering) fonts.FmanGlyphRendering {
	switch rendering {
	case GlyphRenNativeMono:
		return fonts.FmanGlyphRenNativeMono
	case GlyphRenNativeGray8:
		return fonts.FmanGlyphRenNativeGray8
	case GlyphRenOutline:
		return fonts.FmanGlyphRenOutline
	case GlyphRenAggMono:
		return fonts.FmanGlyphRenAggMono
	case GlyphRenAggGray8:
		return fonts.FmanGlyphRenAggGray8
	default:
		return fonts.FmanGlyphRenNativeGray8
	}
}

func fromFmanGlyphRendering(rendering fonts.FmanGlyphRendering) GlyphRendering {
	switch rendering {
	case fonts.FmanGlyphRenNativeMono:
		return GlyphRenNativeMono
	case fonts.FmanGlyphRenNativeGray8:
		return GlyphRenNativeGray8
	case fonts.FmanGlyphRenOutline:
		return GlyphRenOutline
	case fonts.FmanGlyphRenAggMono:
		return GlyphRenAggMono
	case fonts.FmanGlyphRenAggGray8:
		return GlyphRenAggGray8
	default:
		return GlyphRenNativeGray8
	}
}

// FontManager is a Go convenience wrapper around the lower-level FreeType2/fman
// pieces. It does not correspond to a direct AGG type and remains an explicit
// port delta above AGG's concrete engine types.
type FontManager struct {
	engines       map[string]FontEngineInterface
	cacheManager  *CacheManager2
	currentFace   LoadedFaceInterface
	defaultEngine string
}

// NewFontManager creates a convenience wrapper with both 16-bit and 32-bit
// engines pre-initialized. AGG exposes the concrete engines separately.
func NewFontManager() (*FontManager, error) {
	// Create both engine types
	engine16, err := NewFontEngineInt16Default()
	if err != nil {
		return nil, err
	}

	engine32, err := NewFontEngineInt32Default()
	if err != nil {
		engine16.Close()
		return nil, err
	}

	fm := &FontManager{
		engines: map[string]FontEngineInterface{
			"int16": engine16,
			"int32": engine32,
		},
		defaultEngine: "int16", // Start with compact 16-bit engine
	}

	// Initialize cache manager with default engine
	fm.cacheManager = NewCacheManager2(fm.engines[fm.defaultEngine])

	return fm, nil
}

// LoadFont loads a font file through the selected engine.
func (fm *FontManager) LoadFont(fileName string, preferredEngine string) (LoadedFaceInterface, error) {
	// Select the appropriate engine
	engineKey := preferredEngine
	if engineKey == "" {
		engineKey = fm.defaultEngine
	}

	engine, exists := fm.engines[engineKey]
	if !exists {
		engineKey = fm.defaultEngine
		engine = fm.engines[engineKey]
	}

	// Load the font face
	loadedFace, err := engine.LoadFaceFile(fileName)
	if err != nil {
		return nil, err
	}

	// Update cache manager if we switched engines
	if engineKey != fm.defaultEngine {
		if fm.cacheManager != nil {
			_ = fm.cacheManager.Close()
		}
		fm.cacheManager = NewCacheManager2(engine)
	}
	fm.currentFace = loadedFace
	fm.cacheManager.SetCurrentFont(loadedFace)

	return loadedFace, nil
}

// LoadFontFromMemory loads a font from memory through the selected engine.
func (fm *FontManager) LoadFontFromMemory(buffer []byte, preferredEngine string) (LoadedFaceInterface, error) {
	// Select the appropriate engine
	engineKey := preferredEngine
	if engineKey == "" {
		engineKey = fm.defaultEngine
	}

	engine, exists := fm.engines[engineKey]
	if !exists {
		engineKey = fm.defaultEngine
		engine = fm.engines[engineKey]
	}

	// Load the font face
	loadedFace, err := engine.LoadFace(buffer, uint(len(buffer)))
	if err != nil {
		return nil, err
	}

	// Update cache manager if we switched engines
	if engineKey != fm.defaultEngine {
		if fm.cacheManager != nil {
			_ = fm.cacheManager.Close()
		}
		fm.cacheManager = NewCacheManager2(engine)
	}
	fm.currentFace = loadedFace
	fm.cacheManager.SetCurrentFont(loadedFace)

	return loadedFace, nil
}

// GetCacheManager returns the current Go-managed cache wrapper.
func (fm *FontManager) GetCacheManager() *CacheManager2 {
	return fm.cacheManager
}

// CurrentFace returns the currently selected loaded face for this convenience wrapper.
func (fm *FontManager) CurrentFace() LoadedFaceInterface {
	return fm.currentFace
}

// SwitchEngine switches to a different engine type and updates the cache manager.
// This is a Go convenience operation, not a direct AGG API method.
func (fm *FontManager) SwitchEngine(engineType string) error {
	if _, exists := fm.engines[engineType]; !exists {
		return fmt.Errorf("unknown engine type: %s", engineType)
	}

	if engineType != fm.defaultEngine {
		if fm.cacheManager != nil {
			_ = fm.cacheManager.Close()
		}
		fm.defaultEngine = engineType
		fm.cacheManager = NewCacheManager2(fm.engines[engineType])
		fm.currentFace = nil
	}

	return nil
}

// Close cleans up all resources used by the font manager.
func (fm *FontManager) Close() error {
	// Close cache manager
	if fm.cacheManager != nil {
		fm.cacheManager.Close()
		fm.cacheManager = nil
	}
	fm.currentFace = nil

	// Close all engines
	for _, engine := range fm.engines {
		if engine != nil {
			engine.Close()
		}
	}
	fm.engines = nil

	return nil
}
