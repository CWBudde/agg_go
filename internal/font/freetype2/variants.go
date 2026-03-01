//go:build freetype

// Package freetype2 provides precision-specific font engine variants.
// This implements the Int16 and Int32 variants corresponding to AGG's
// font_engine_freetype_int16 and font_engine_freetype_int32 classes.
package freetype2

import (
	"unsafe"

	"agg_go/internal/path"
	"agg_go/internal/scanline"
)

// FontEngineInt16 uses 16-bit precision (10.6 format) for vector cache.
// The vector cache is compact, but integer overflow can occur for glyphs
// with height more than 200 pixels.
// This corresponds to AGG's fman::font_engine_freetype_int16 class.
type FontEngineInt16 struct {
	*FontEngine
}

// Type aliases for Int16 variant to match AGG's typedef structure.
type PathAdaptorInt16Type = *path.SerializedIntegerPathAdaptor[int16]

// NewFontEngineInt16 creates a new 16-bit precision FreeType font engine.
// This engine is optimized for smaller glyphs (< 200px height) and uses less memory.
func NewFontEngineInt16(maxFaces uint32, ftMemory unsafe.Pointer) (*FontEngineInt16, error) {
	baseEngine, err := NewFontEngine(false, maxFaces, ftMemory) // flag32 = false for 16-bit
	if err != nil {
		return nil, err
	}

	return &FontEngineInt16{
		FontEngine: baseEngine,
	}, nil
}

// PathAdaptor returns the 16-bit path adaptor for this engine.
func (fe16 *FontEngineInt16) PathAdaptor() PathAdaptorInt16Type {
	// Create a new serialized path adaptor for int16
	return path.NewSerializedIntegerPathAdaptor[int16]()
}

// PathAdaptorInterface returns the path adaptor through the common interface.
func (fe16 *FontEngineInt16) PathAdaptorInterface() path.IntegerPathAdaptor {
	return fe16.PathAdaptor()
}

// Gray8SerializedAdaptor returns the serialized AA adaptor used by CacheManager2.
func (fe16 *FontEngineInt16) Gray8SerializedAdaptor() *scanline.SerializedScanlinesAdaptorAA[uint8] {
	return scanline.NewSerializedScanlinesAdaptorAAEmpty[uint8]()
}

// MonoSerializedAdaptor returns the serialized binary adaptor used by CacheManager2.
func (fe16 *FontEngineInt16) MonoSerializedAdaptor() *scanline.SerializedScanlinesAdaptorBin {
	return scanline.NewSerializedScanlinesAdaptorBin()
}

// FontEngineInt32 uses 32-bit precision (26.6 format) for vector cache.
// The vector cache is twice as large as the Int16 variant, but it allows
// rendering of very large glyphs without integer overflow.
// This corresponds to AGG's fman::font_engine_freetype_int32 class.
type FontEngineInt32 struct {
	*FontEngine
}

// Type aliases for Int32 variant to match AGG's typedef structure.
type PathAdaptorInt32Type = *path.SerializedIntegerPathAdaptor[int32]

// NewFontEngineInt32 creates a new 32-bit precision FreeType font engine.
// This engine can handle very large glyphs without overflow but uses more memory.
func NewFontEngineInt32(maxFaces uint32, ftMemory unsafe.Pointer) (*FontEngineInt32, error) {
	baseEngine, err := NewFontEngine(true, maxFaces, ftMemory) // flag32 = true for 32-bit
	if err != nil {
		return nil, err
	}

	return &FontEngineInt32{
		FontEngine: baseEngine,
	}, nil
}

// PathAdaptor returns the 32-bit path adaptor for this engine.
func (fe32 *FontEngineInt32) PathAdaptor() PathAdaptorInt32Type {
	// Create a new serialized path adaptor for int32
	return path.NewSerializedIntegerPathAdaptor[int32]()
}

// PathAdaptorInterface returns the path adaptor through the common interface.
func (fe32 *FontEngineInt32) PathAdaptorInterface() path.IntegerPathAdaptor {
	return fe32.PathAdaptor()
}

// Gray8SerializedAdaptor returns the serialized AA adaptor used by CacheManager2.
func (fe32 *FontEngineInt32) Gray8SerializedAdaptor() *scanline.SerializedScanlinesAdaptorAA[uint8] {
	return scanline.NewSerializedScanlinesAdaptorAAEmpty[uint8]()
}

// MonoSerializedAdaptor returns the serialized binary adaptor used by CacheManager2.
func (fe32 *FontEngineInt32) MonoSerializedAdaptor() *scanline.SerializedScanlinesAdaptorBin {
	return scanline.NewSerializedScanlinesAdaptorBin()
}

// Convenience constructors that match AGG's typical usage patterns

// NewFontEngineInt16Default creates a 16-bit engine with default parameters.
func NewFontEngineInt16Default() (*FontEngineInt16, error) {
	return NewFontEngineInt16(32, nil) // 32 faces, no custom memory
}

// NewFontEngineInt32Default creates a 32-bit engine with default parameters.
func NewFontEngineInt32Default() (*FontEngineInt32, error) {
	return NewFontEngineInt32(32, nil) // 32 faces, no custom memory
}

// FontEngineInterface implementations for both variants

// LoadFace forwards to the base engine's LoadFace method.
func (fe16 *FontEngineInt16) LoadFace(buffer []byte, bytes uint) (LoadedFaceInterface, error) {
	return fe16.FontEngine.LoadFace(buffer, bytes)
}

// LoadFaceFile forwards to the base engine's LoadFaceFile method.
func (fe16 *FontEngineInt16) LoadFaceFile(fileName string) (LoadedFaceInterface, error) {
	return fe16.FontEngine.LoadFaceFile(fileName)
}

// UnloadFace forwards to the base engine's UnloadFace method.
func (fe16 *FontEngineInt16) UnloadFace(face LoadedFaceInterface) error {
	return fe16.FontEngine.UnloadFace(face)
}

// SetGamma forwards to the base engine's SetGamma method.
func (fe16 *FontEngineInt16) SetGamma(gamma float64) {
	fe16.FontEngine.SetGamma(gamma)
}

// Is32Bit returns false for the 16-bit variant.
func (fe16 *FontEngineInt16) Is32Bit() bool {
	return false
}

// LastError forwards to the base engine's LastError method.
func (fe16 *FontEngineInt16) LastError() error {
	return fe16.FontEngine.LastError()
}

// Close forwards to the base engine's Close method.
func (fe16 *FontEngineInt16) Close() error {
	return fe16.FontEngine.Close()
}

// Int32 variant interface implementations

// LoadFace forwards to the base engine's LoadFace method.
func (fe32 *FontEngineInt32) LoadFace(buffer []byte, bytes uint) (LoadedFaceInterface, error) {
	return fe32.FontEngine.LoadFace(buffer, bytes)
}

// LoadFaceFile forwards to the base engine's LoadFaceFile method.
func (fe32 *FontEngineInt32) LoadFaceFile(fileName string) (LoadedFaceInterface, error) {
	return fe32.FontEngine.LoadFaceFile(fileName)
}

// UnloadFace forwards to the base engine's UnloadFace method.
func (fe32 *FontEngineInt32) UnloadFace(face LoadedFaceInterface) error {
	return fe32.FontEngine.UnloadFace(face)
}

// SetGamma forwards to the base engine's SetGamma method.
func (fe32 *FontEngineInt32) SetGamma(gamma float64) {
	fe32.FontEngine.SetGamma(gamma)
}

// Is32Bit returns true for the 32-bit variant.
func (fe32 *FontEngineInt32) Is32Bit() bool {
	return true
}

// LastError forwards to the base engine's LastError method.
func (fe32 *FontEngineInt32) LastError() error {
	return fe32.FontEngine.LastError()
}

// Close forwards to the base engine's Close method.
func (fe32 *FontEngineInt32) Close() error {
	return fe32.FontEngine.Close()
}

// Package-local engine selection helpers layered above AGG's concrete
// int16/int32 types. Production code should prefer the explicit constructors.

// recommendedEngineForGlyphSize returns a Go-side heuristic for choosing the
// int16 vs int32 engine based on glyph size.
func recommendedEngineForGlyphSize(maxGlyphHeight float64) string {
	if maxGlyphHeight <= 200.0 {
		return "int16" // Compact storage for smaller glyphs
	}
	return "int32" // Safe for very large glyphs
}

// createRecommendedEngine creates an engine using the Go-side size heuristic.
// AGG exposes the concrete int16/int32 engine types directly instead.
func createRecommendedEngine(maxGlyphHeight float64, maxFaces uint32) (FontEngineInterface, error) {
	if recommendedEngineForGlyphSize(maxGlyphHeight) == "int16" {
		return NewFontEngineInt16(maxFaces, nil)
	}
	return NewFontEngineInt32(maxFaces, nil)
}
