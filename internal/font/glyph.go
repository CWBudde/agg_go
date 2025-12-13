// Package font provides font rendering capabilities for the AGG graphics library.
// This is a Go port of the AGG font system, including glyph management and rendering.
package font

import (
	"agg_go/internal/basics"
)

// GlyphDataType defines the type of glyph data stored in a cache entry.
type GlyphDataType int

const (
	GlyphDataInvalid GlyphDataType = iota // Invalid/empty glyph data
	GlyphDataMono                         // 1-bit monochrome glyph data
	GlyphDataGray8                        // 8-bit anti-aliased glyph data
	GlyphDataOutline                      // Vector outline glyph data
)

// GlyphCache represents a cached glyph with its rendering data and metrics.
// This mirrors the C++ glyph_cache struct from agg_font_cache_manager.h.
type GlyphCache struct {
	GlyphIndex uint             // Font-specific glyph index
	Data       []byte           // Serialized glyph data (scanlines or outline)
	DataSize   uint             // Size of the glyph data
	DataType   GlyphDataType    // Type of data stored
	Bounds     basics.Rect[int] // Bounding rectangle of the glyph
	AdvanceX   float64          // Horizontal advance for glyph positioning
	AdvanceY   float64          // Vertical advance for glyph positioning
}

// GlyphRenderingType defines how glyphs should be rendered.
type GlyphRenderingType int

const (
	GlyphRenderingNative  GlyphRenderingType = iota // Use font's native rendering
	GlyphRenderingOutline                           // Render as vector outline
	GlyphRenderingAAGray8                           // Anti-aliased gray8 rendering
	GlyphRenderingAAMono                            // Anti-aliased mono rendering
	GlyphRenderingMono                              // 1-bit mono rendering
)

// FontMetrics contains standard font metrics.
type FontMetrics struct {
	Height    float64 // Font height in points
	Ascender  float64 // Maximum ascender
	Descender float64 // Maximum descender (typically negative)
	LineGap   float64 // Recommended line spacing gap
}

// SerializedScanlinesAdaptorAA and SerializedScanlinesAdaptorBin both implement
// the font.SerializedScanlinesAdaptor interface, providing a unified way to access
// serialized scanline data for glyph rendering. Both types expose the same interface
// methods (Bounds() and Data()), eliminating the need for interface{} type assertions
// in rendering code.

// SerializedScanlinesAdaptorAA provides access to anti-aliased scanline data.
// This adapts serialized AA scanline data for rendering.
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

// SerializedScanlinesAdaptorBin provides access to binary scanline data.
// This adapts serialized binary scanline data for rendering.
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
