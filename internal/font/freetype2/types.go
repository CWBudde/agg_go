//go:build freetype

// Package freetype2 provides enhanced FreeType2 font engine integration for AGG graphics library.
// This is the v2 implementation with advanced multi-face support and improved memory management.
// It corresponds to AGG's fman (font manager) namespace functionality.
package freetype2

import (
	"agg_go/internal/basics"
	"agg_go/internal/fonts"
	"agg_go/internal/path"
	"agg_go/internal/transform"
)

// GlyphRendering defines the rendering mode for glyphs in the v2 engine.
// This corresponds to AGG's fman::glyph_rendering enum.
type GlyphRendering int

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

// PreparedGlyph contains prepared glyph information for rendering.
// This corresponds to AGG's fman::prepared_glyph struct.
type PreparedGlyph struct {
	GlyphCode  uint32                  // Character code
	GlyphIndex uint32                  // Glyph index in font
	DataSize   uint32                  // Size of glyph data
	DataType   fonts.FmanGlyphDataType // Type of glyph data
	Bounds     basics.Rect[int]        // Glyph bounding rectangle
	AdvanceX   float64                 // Horizontal advance
	AdvanceY   float64                 // Vertical advance
}

// FontEngineAdaptorTypes defines type aliases for different path storage precisions.
// This corresponds to AGG's serialized path adaptors.
type FontEngineAdaptorTypes struct {
	// For Int16 variant (compact, for glyphs < 200px height)
	PathAdaptorInt16 *path.SerializedIntegerPathAdaptor[int16]

	// For Int32 variant (for very large glyphs)
	PathAdaptorInt32 *path.SerializedIntegerPathAdaptor[int32]

	// Shared scanline adaptors
	Gray8Adaptor interface{} // Will be *scanline.SerializedScanlinesAdaptorAA when available
	MonoAdaptor  interface{} // Will be *scanline.SerializedScanlinesAdaptorBin when available

	// Scanline storage types
	ScanlinesAA  interface{} // Will be *scanline.ScanlineStorageAA8 when available
	ScanlinesBin interface{} // Will be *scanline.ScanlineStorageBin when available
}

// FontEngineInterface defines the interface that all FreeType2 engines must implement.
// This provides a common interface for both Int16 and Int32 variants.
type FontEngineInterface interface {
	// Font management
	LoadFace(buffer []byte, bytes uint) (LoadedFaceInterface, error)
	LoadFaceFile(fileName string) (LoadedFaceInterface, error)
	UnloadFace(face LoadedFaceInterface) error

	// Gamma correction
	SetGamma(gamma float64)

	// Engine information
	Is32Bit() bool
	LastError() error

	// Resource cleanup
	Close() error
}

// LoadedFaceInterface defines the interface for managing individual font faces.
// This corresponds to AGG's fman::loaded_face class functionality.
type LoadedFaceInterface interface {
	// Face information
	NumFaces() uint32
	Name() string
	Resolution() uint32
	Height() float64
	Width() float64
	Ascent() float64
	Descent() float64
	AscentB() float64
	DescentB() float64

	// Rendering configuration
	Hinting() bool
	FlipY() bool
	Transform() *transform.TransAffine
	CharMap() CharEncoding

	// Instance selection and configuration
	SelectInstance(height, width float64, hinting bool, rendering GlyphRendering)
	CapableRendering(rendering GlyphRendering) GlyphRendering
	SetHinting(hinting bool)
	SetFlipY(flipY bool)
	SetTransform(affine *transform.TransAffine)
	SetCharMap(encoding CharEncoding) error

	// Glyph operations
	PrepareGlyph(glyphCode uint32) (*PreparedGlyph, bool)
	AddKerning(first, second uint32) (dx, dy float64)
	WriteGlyphTo(prepared *PreparedGlyph, data []byte)

	// Resource cleanup
	Close() error
}

// CharEncoding represents FreeType character encoding types.
type CharEncoding int

const (
	EncodingNone CharEncoding = iota
	EncodingMS
	EncodingUnicode
	EncodingSymbol
	EncodingAdobeLatin1
	EncodingAdobeCustom
	EncodingAdobeExpert
)

// FontEngineBase provides common functionality for both Int16 and Int32 variants.
// This corresponds to AGG's fman::font_engine_freetype_base class.
type FontEngineBase struct {
	// Configuration
	flag32             bool
	lastError          error
	libraryInitialized bool
	maxFaces           uint32

	// Storage for different precision levels
	pathStorage16 *path.PathStorageInteger[int16]
	pathStorage32 *path.PathStorageInteger[int32]
	curves16      *path.PathStorageInteger[int16] // TODO: Add conv_curve wrapper
	curves32      *path.PathStorageInteger[int32] // TODO: Add conv_curve wrapper

	// Scanline components
	scanlineU8   interface{} // Will be *scanline.ScanlineU8 when available
	scanlineBin  interface{} // Will be *scanline.ScanlineBin when available
	scanlinesAA  interface{} // Will be *scanline.ScanlineStorageAA8 when available
	scanlinesBin interface{} // Will be *scanline.ScanlineStorageBin when available

	// Rasterizer
	rasterizer interface{} // Will be *rasterizer.RasterizerScanlineAA when available
}

// NewFontEngineBase creates a new base font engine with the specified configuration.
func NewFontEngineBase(flag32 bool, maxFaces uint32) *FontEngineBase {
	if maxFaces == 0 {
		maxFaces = 32 // Default from AGG implementation
	}

	return &FontEngineBase{
		flag32:        flag32,
		maxFaces:      maxFaces,
		pathStorage16: path.NewPathStorageInteger[int16](),
		pathStorage32: path.NewPathStorageInteger[int32](),
		scanlineU8:    nil, // TODO: Initialize when scanline types are available
		scanlineBin:   nil, // TODO: Initialize when scanline types are available
		scanlinesAA:   nil, // TODO: Initialize when scanline types are available
		scanlinesBin:  nil, // TODO: Initialize when scanline types are available
		rasterizer:    nil, // TODO: Initialize when rasterizer is available
	}
}

// SetGamma sets the gamma correction for the rasterizer.
func (feb *FontEngineBase) SetGamma(gamma float64) {
	// TODO: Implement gamma function and apply to rasterizer
	// This would correspond to AGG's template<class GammaF> void gamma(const GammaF& f)
}

// Is32Bit returns whether this engine uses 32-bit precision.
func (feb *FontEngineBase) Is32Bit() bool {
	return feb.flag32
}

// LastError returns the last error that occurred.
func (feb *FontEngineBase) LastError() error {
	return feb.lastError
}

// Close cleans up resources used by the base engine.
func (feb *FontEngineBase) Close() error {
	// Clean up any base resources
	feb.libraryInitialized = false
	feb.lastError = nil
	return nil
}
