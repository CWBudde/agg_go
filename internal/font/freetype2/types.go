//go:build freetype

// Package freetype2 provides enhanced FreeType2 font engine integration for AGG graphics library.
// This is the v2 implementation with advanced multi-face support and improved memory management.
// It corresponds to AGG's fman (font manager) namespace functionality.
package freetype2

import (
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	"agg_go/internal/fonts"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt/gamma"
	"agg_go/internal/scanline"
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
	Gray8Adaptor *scanline.SerializedScanlinesAdaptorAA[uint8]
	MonoAdaptor  *scanline.SerializedScanlinesAdaptorBin

	// Scanline storage types
	ScanlinesAA  *scanline.ScanlineStorageAA[uint8]
	ScanlinesBin *scanline.ScanlineStorageBin
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

// PathStorageInterface defines the interface for path storage used in font rendering.
// This provides a common interface for both int16 and int32 path storage variants.
type PathStorageInterface[T any] interface {
	RemoveAll()
	MoveTo(x, y T)
	LineTo(x, y T)
	Curve3(xCtrl, yCtrl, xTo, yTo T)
	Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo T)
	ClosePolygon()
}

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
	curves16      *conv.ConvCurveInteger[int16] // ConvCurve wrapper for int16 paths
	curves32      *conv.ConvCurveInteger[int32] // ConvCurve wrapper for int32 paths

	// Scanline components
	scanlineU8   *scanline.ScanlineU8
	scanlineBin  *scanline.ScanlineBin
	scanlinesAA  *scanline.ScanlineStorageAA[uint8]
	scanlinesBin *scanline.ScanlineStorageBin

	// Gamma correction
	gammaFunc gamma.GammaFunction // Gamma correction function
}

// NewFontEngineBase creates a new base font engine with the specified configuration.
func NewFontEngineBase(flag32 bool, maxFaces uint32) *FontEngineBase {
	if maxFaces == 0 {
		maxFaces = 32 // Default from AGG implementation
	}

	// Create path storage instances
	pathStorage16 := path.NewPathStorageInteger[int16]()
	pathStorage32 := path.NewPathStorageInteger[int32]()

	// Create curve converters with default approximation scale (matching AGG default)
	curves16 := conv.NewConvCurveInteger(pathStorage16)
	curves32 := conv.NewConvCurveInteger(pathStorage32)
	curves16.SetApproximationScale(4.0)
	curves32.SetApproximationScale(4.0)

	return &FontEngineBase{
		flag32:        flag32,
		maxFaces:      maxFaces,
		pathStorage16: pathStorage16,
		pathStorage32: pathStorage32,
		curves16:      curves16,
		curves32:      curves32,
		scanlineU8:    scanline.NewScanlineU8(),
		scanlineBin:   scanline.NewScanlineBin(),
		scanlinesAA:   scanline.NewScanlineStorageAA[uint8](),
		scanlinesBin:  scanline.NewScanlineStorageBin(),
		gammaFunc:     gamma.NewGammaNone(), // Default to no gamma correction
	}
}

// SetGamma sets the gamma correction for the rasterizer.
func (feb *FontEngineBase) SetGamma(gammaValue float64) {
	if gammaValue <= 0 {
		// Invalid gamma, use no correction
		feb.gammaFunc = gamma.NewGammaNone()
	} else if gammaValue == 1.0 {
		// No correction needed
		feb.gammaFunc = gamma.NewGammaNone()
	} else {
		// Use power gamma correction
		feb.gammaFunc = gamma.NewGammaPower(gammaValue)
	}
	// Note: The gamma function will be applied during rasterization
	// This corresponds to AGG's template<class GammaF> void gamma(const GammaF& f)
}

// GetGammaFunc returns the current gamma correction function.
func (feb *FontEngineBase) GetGammaFunc() gamma.GammaFunction {
	return feb.gammaFunc
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
