// Package fonts provides embedded raster font support plus the separate
// fman/font_cache_manager2 interfaces from AGG.
package fonts

import (
	"agg_go/internal/basics"
)

// PathAdaptorType defines the interface for path adaptors used in font rendering.
// This interface is implemented by both int16 and int32 path adaptors.
type PathAdaptorType interface {
	// Init initializes the adaptor with serialized data and transformation parameters
	Init(data []byte, dx, dy, scale float64, coordShift int)

	// InitWithScale initializes the adaptor with data, position, and scale
	InitWithScale(data []byte, dataSize uint32, x, y, scale float64)

	// Rewind resets the adaptor to the beginning of the path
	Rewind(pathID uint)

	// Vertex returns the next vertex in the path
	Vertex(x, y *float64) uint
}

// Gray8AdaptorType defines the interface for 8-bit grayscale scanline adaptors.
// This adaptor provides access to anti-aliased glyph data.
type Gray8AdaptorType interface {
	// InitGlyph initializes the adaptor with glyph data and position
	InitGlyph(data []byte, dataSize uint32, x, y float64)

	// Bounds returns the bounding rectangle of the glyph
	Bounds() basics.Rect[int]

	// Rewind prepares the adaptor for scanline iteration
	Rewind(pathID uint)

	// SweepScanline returns the next scanline
	SweepScanline() bool

	// NumSpans returns the number of spans in current scanline
	NumSpans() uint

	// Begin returns iterator for the first span
	Begin() Gray8SpanIterator
}

// Gray8SpanIterator provides iteration over spans in a gray8 scanline.
type Gray8SpanIterator interface {
	// Next advances to the next span
	Next()

	// IsValid returns true if the iterator is at a valid span
	IsValid() bool

	// X returns the starting X coordinate of current span
	X() int

	// Len returns the length of current span
	Len() int

	// Covers returns the coverage array for current span
	Covers() []uint8
}

// Gray8ScanlineType defines the interface for 8-bit grayscale scanlines.
// This provides the scanline structure used for rendering.
type Gray8ScanlineType interface {
	// Reset prepares the scanline for a new row
	Reset(minX, maxX int)

	// Y returns the Y coordinate of the scanline
	Y() int

	// NumSpans returns the number of spans in the scanline
	NumSpans() uint

	// Begin returns an iterator for the spans
	Begin() Gray8SpanIterator
}

// MonoAdaptorType defines the interface for monochrome scanline adaptors.
// This adaptor provides access to 1-bit monochrome glyph data.
type MonoAdaptorType interface {
	// InitGlyph initializes the adaptor with glyph data and position
	InitGlyph(data []byte, dataSize uint32, x, y float64)

	// Bounds returns the bounding rectangle of the glyph
	Bounds() basics.Rect[int]

	// Rewind prepares the adaptor for scanline iteration
	Rewind(pathID uint)

	// SweepScanline returns the next scanline
	SweepScanline() bool

	// NumSpans returns the number of spans in current scanline
	NumSpans() uint

	// Begin returns iterator for the first span
	Begin() MonoSpanIterator
}

// MonoSpanIterator provides iteration over spans in a mono scanline.
type MonoSpanIterator interface {
	// Next advances to the next span
	Next()

	// IsValid returns true if the iterator is at a valid span
	IsValid() bool

	// X returns the starting X coordinate of current span
	X() int

	// Len returns the length of current span
	Len() int
}

// MonoScanlineType defines the interface for monochrome scanlines.
// This provides the scanline structure used for rendering.
type MonoScanlineType interface {
	// Reset prepares the scanline for a new row
	Reset(minX, maxX int)

	// Y returns the Y coordinate of the scanline
	Y() int

	// NumSpans returns the number of spans in the scanline
	NumSpans() uint

	// Begin returns an iterator for the spans
	Begin() MonoSpanIterator
}
