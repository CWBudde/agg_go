package font

import (
	"github.com/cwbudde/agg_go/internal/basics"
	isc "github.com/cwbudde/agg_go/internal/scanline"
)

type serializedScanlineTarget struct {
	sl isc.Scanline
}

func (t serializedScanlineTarget) Y() int {
	return t.sl.Y()
}

func (t serializedScanlineTarget) NumSpans() int {
	return t.sl.NumSpans()
}

func (t serializedScanlineTarget) Begin() isc.ScanlineIterator {
	return t.sl.BeginIterator()
}

func (t serializedScanlineTarget) ResetSpans() {
	t.sl.ResetSpans()
}

func (t serializedScanlineTarget) AddSpan(x, length int, cover basics.Int8u) {
	t.sl.AddSpan(x, length, uint(cover))
}

func (t serializedScanlineTarget) AddCells(x, length int, covers []basics.Int8u) {
	for i := 0; i < length && i < len(covers); i++ {
		t.sl.AddCell(x+i, uint(covers[i]))
	}
}

func (t serializedScanlineTarget) Finalize(y int) {
	t.sl.Finalize(y)
}

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

// PositionedGlyph stores the glyph index and placement returned by a shaped
// text layout run. Advance and offset values are expressed in the same units as
// GlyphCache advances for the active font engine configuration.
type PositionedGlyph struct {
	GlyphIndex uint
	XAdvance   float64
	YAdvance   float64
	XOffset    float64
	YOffset    float64
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
	bounds basics.Rect[int]
	inner  *isc.SerializedScanlinesAdaptorAA[basics.Int8u]
}

// NewSerializedScanlinesAdaptorAA creates a new AA scanline adaptor.
func NewSerializedScanlinesAdaptorAA(data []byte, bounds basics.Rect[int]) *SerializedScanlinesAdaptorAA {
	adaptor := &SerializedScanlinesAdaptorAA{
		data:   data,
		bounds: bounds,
	}
	adaptor.Init(data, bounds, 0, 0)
	return adaptor
}

// Init reinitializes the adaptor with serialized scanline data and placement offsets.
func (s *SerializedScanlinesAdaptorAA) Init(data []byte, bounds basics.Rect[int], dx, dy float64) {
	s.data = data
	s.bounds = bounds
	if s.inner == nil {
		s.inner = isc.NewSerializedScanlinesAdaptorAA[basics.Int8u](data, len(data), dx, dy)
		return
	}
	s.inner.Init(data, len(data), dx, dy)
}

// Bounds returns the bounding rectangle of the scanlines.
func (s *SerializedScanlinesAdaptorAA) Bounds() basics.Rect[int] {
	return s.bounds
}

// Data returns the serialized scanline data.
func (s *SerializedScanlinesAdaptorAA) Data() []byte {
	return s.data
}

// RewindScanlines prepares the adaptor for scanline iteration.
func (s *SerializedScanlinesAdaptorAA) RewindScanlines() bool {
	return s.inner != nil && s.inner.RewindScanlines()
}

// MinX returns the minimum X coordinate of the serialized scanlines.
func (s *SerializedScanlinesAdaptorAA) MinX() int {
	if s.inner == nil {
		return 0
	}
	return s.inner.MinX()
}

// MaxX returns the maximum X coordinate of the serialized scanlines.
func (s *SerializedScanlinesAdaptorAA) MaxX() int {
	if s.inner == nil {
		return 0
	}
	return s.inner.MaxX()
}

// SweepScanline decodes one serialized scanline into sl.
func (s *SerializedScanlinesAdaptorAA) SweepScanline(sl isc.Scanline) bool {
	return s.inner != nil && s.inner.SweepScanline(serializedScanlineTarget{sl: sl})
}

// SerializedScanlinesAdaptorBin adapts serialized 1-bit glyph scanlines to the
// same read-only interface as the AA adaptor.
type SerializedScanlinesAdaptorBin struct {
	data   []byte
	bounds basics.Rect[int]
	inner  *isc.SerializedScanlinesAdaptorBin
}

// NewSerializedScanlinesAdaptorBin creates a new binary scanline adaptor.
func NewSerializedScanlinesAdaptorBin(data []byte, bounds basics.Rect[int]) *SerializedScanlinesAdaptorBin {
	adaptor := &SerializedScanlinesAdaptorBin{
		data:   data,
		bounds: bounds,
	}
	adaptor.Init(data, bounds, 0, 0)
	return adaptor
}

// Init reinitializes the adaptor with serialized scanline data and placement offsets.
func (s *SerializedScanlinesAdaptorBin) Init(data []byte, bounds basics.Rect[int], dx, dy float64) {
	s.data = data
	s.bounds = bounds
	if s.inner == nil {
		s.inner = isc.NewSerializedScanlinesAdaptorBin()
	}
	s.inner.Init(data, dx, dy)
}

// Bounds returns the bounding rectangle of the scanlines.
func (s *SerializedScanlinesAdaptorBin) Bounds() basics.Rect[int] {
	return s.bounds
}

// Data returns the serialized scanline data.
func (s *SerializedScanlinesAdaptorBin) Data() []byte {
	return s.data
}

// RewindScanlines prepares the adaptor for scanline iteration.
func (s *SerializedScanlinesAdaptorBin) RewindScanlines() bool {
	return s.inner != nil && s.inner.RewindScanlines()
}

// MinX returns the minimum X coordinate of the serialized scanlines.
func (s *SerializedScanlinesAdaptorBin) MinX() int {
	if s.inner == nil {
		return 0
	}
	return s.inner.MinX()
}

// MaxX returns the maximum X coordinate of the serialized scanlines.
func (s *SerializedScanlinesAdaptorBin) MaxX() int {
	if s.inner == nil {
		return 0
	}
	return s.inner.MaxX()
}

// SweepScanline decodes one serialized scanline into sl.
func (s *SerializedScanlinesAdaptorBin) SweepScanline(sl isc.Scanline) bool {
	return s.inner != nil && s.inner.SweepScanline(serializedScanlineTarget{sl: sl})
}
