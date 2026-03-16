package font

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// GlyphDataType identifies which serialized representation a cached glyph uses.
type GlyphDataType int

const (
	GlyphDataInvalid GlyphDataType = iota // Invalid/empty glyph data
	GlyphDataMono                         // 1-bit monochrome glyph data
	GlyphDataGray8                        // 8-bit anti-aliased glyph data
	GlyphDataOutline                      // Vector outline glyph data
)

// GlyphCache stores the cached metrics and serialized glyph payload for one
// glyph, mirroring AGG's glyph_cache structure.
type GlyphCache struct {
	GlyphIndex uint             // Font-specific glyph index
	Data       []byte           // Serialized glyph data (scanlines or outline)
	DataSize   uint             // Size of the glyph data
	DataType   GlyphDataType    // Type of data stored
	Bounds     basics.Rect[int] // Bounding rectangle of the glyph
	AdvanceX   float64          // Horizontal advance for glyph positioning
	AdvanceY   float64          // Vertical advance for glyph positioning
}

// GlyphRenderingType selects the raster or outline form requested from a font
// engine.
type GlyphRenderingType int

const (
	GlyphRenderingNative  GlyphRenderingType = iota // Use font's native rendering
	GlyphRenderingOutline                           // Render as vector outline
	GlyphRenderingAAGray8                           // Anti-aliased gray8 rendering
	GlyphRenderingAAMono                            // Anti-aliased mono rendering
	GlyphRenderingMono                              // 1-bit mono rendering
)

// FontMetrics stores the line metrics reported by a font face.
type FontMetrics struct {
	Height    float64 // Font height in points
	Ascender  float64 // Maximum ascender
	Descender float64 // Maximum descender (typically negative)
	LineGap   float64 // Recommended line spacing gap
}

// SerializedScanlinesAdaptorAA adapts serialized anti-aliased glyph scanlines to
// the minimal read-only interface used by text renderers.
type SerializedScanlinesAdaptorAA struct {
	data   []byte
	size   int
	bounds basics.Rect[int]
}

// NewSerializedScanlinesAdaptorAA creates a new AA scanline adaptor.
func NewSerializedScanlinesAdaptorAA(data []byte, bounds basics.Rect[int]) *SerializedScanlinesAdaptorAA {
	return &SerializedScanlinesAdaptorAA{
		data:   data,
		size:   len(data),
		bounds: bounds,
	}
}

// Bounds returns the bounding rectangle of the scanlines.
func (s *SerializedScanlinesAdaptorAA) Bounds() basics.Rect[int] {
	return s.bounds
}

// Data returns the serialized scanline data.
func (s *SerializedScanlinesAdaptorAA) Data() []byte {
	return s.data
}

// SerializedScanlinesAdaptorBin adapts serialized 1-bit glyph scanlines to the
// same read-only interface as the AA adaptor.
type SerializedScanlinesAdaptorBin struct {
	data   []byte
	size   int
	bounds basics.Rect[int]
}

// NewSerializedScanlinesAdaptorBin creates a new binary scanline adaptor.
func NewSerializedScanlinesAdaptorBin(data []byte, bounds basics.Rect[int]) *SerializedScanlinesAdaptorBin {
	return &SerializedScanlinesAdaptorBin{
		data:   data,
		size:   len(data),
		bounds: bounds,
	}
}

// Bounds returns the bounding rectangle of the scanlines.
func (s *SerializedScanlinesAdaptorBin) Bounds() basics.Rect[int] {
	return s.bounds
}

// Data returns the serialized scanline data.
func (s *SerializedScanlinesAdaptorBin) Data() []byte {
	return s.data
}
