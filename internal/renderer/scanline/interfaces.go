// Package scanline provides scanline-based rendering functionality for AGG.
// This package implements the core rendering functions and classes that take
// scanline data from rasterizers and render it using pixel formats.
package scanline

import (
	"agg_go/internal/basics"
)

// ScanlineInterface defines the interface for scanline containers.
// This corresponds to the scanline template parameter in AGG's renderer functions.
type ScanlineInterface interface {
	// Y returns the current scanline's Y coordinate
	Y() int

	// NumSpans returns the number of spans in this scanline
	NumSpans() int

	// Begin returns an iterator to the first span
	Begin() ScanlineIterator
}

// ScanlineIterator provides iteration over spans in a scanline.
// This corresponds to the const_iterator in AGG's scanline classes.
type ScanlineIterator interface {
	// GetSpan returns the current span data
	GetSpan() SpanData

	// Next advances to the next span and returns true if valid
	Next() bool
}

// SpanData represents a single span within a scanline.
// This corresponds to the span struct in AGG's scanline classes.
type SpanData struct {
	X      int            // Starting X coordinate
	Len    int            // Length (positive) or end coordinate (negative for solid spans)
	Covers []basics.Int8u // Coverage values (may be nil for solid spans)
}

// RasterizerInterface defines the interface for rasterizers that produce scanlines.
// This corresponds to the Rasterizer template parameter in AGG's render functions.
type RasterizerInterface interface {
	// RewindScanlines prepares the rasterizer for scanline sweeping
	// Returns true if there are scanlines to render
	RewindScanlines() bool

	// SweepScanline fills the scanline with the next row of data
	// Returns true if a scanline was filled, false if done
	SweepScanline(sl ScanlineInterface) bool

	// MinX returns the minimum X coordinate of all scanlines
	MinX() int

	// MaxX returns the maximum X coordinate of all scanlines
	MaxX() int
}

// BaseRendererInterface defines the interface for base renderers.
// This corresponds to the BaseRenderer template parameter in AGG's functions.
type BaseRendererInterface interface {
	// Color type - we use interface{} for flexibility
	// In a real implementation, this might be constrained to specific color types

	// BlendSolidHspan blends a horizontal span with solid color and coverage array
	BlendSolidHspan(x, y, len int, color interface{}, covers []basics.Int8u)

	// BlendHline blends a horizontal line with solid color and single coverage
	BlendHline(x, y, x2 int, color interface{}, cover basics.Int8u)

	// BlendColorHspan blends a horizontal span with color array and coverage
	BlendColorHspan(x, y, len int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u)
}

// SpanAllocatorInterface defines the interface for span allocators.
// This corresponds to the SpanAllocator template parameter in AGG's functions.
type SpanAllocatorInterface interface {
	// Allocate allocates an array of colors for the given length
	// Returns a slice that can hold 'len' color values
	Allocate(len int) []interface{}
}

// SpanGeneratorInterface defines the interface for span generators.
// This corresponds to the SpanGenerator template parameter in AGG's functions.
type SpanGeneratorInterface interface {
	// Prepare is called before rendering begins
	Prepare()

	// Generate fills the colors array with generated colors for the given span
	Generate(colors []interface{}, x, y, len int)
}

// RendererInterface defines the interface for scanline renderers.
// This provides a common interface for all scanline renderer implementations.
type RendererInterface interface {
	// Prepare is called before rendering begins
	Prepare()

	// Render renders a single scanline
	Render(sl ScanlineInterface)
}

// CompoundRasterizerInterface extends RasterizerInterface for compound rendering.
// This is used for multi-style rendering with compound rasterizers.
type CompoundRasterizerInterface interface {
	RasterizerInterface

	// SweepStyles returns the number of styles for the current scanline
	SweepStyles() int

	// SweepScanlineWithStyle fills scanline for a specific style
	SweepScanlineWithStyle(sl ScanlineInterface, styleId int) bool

	// Style returns the style ID for the given index
	Style(index int) int

	// ScanlineStart returns the starting X coordinate of current scanline
	ScanlineStart() int

	// ScanlineLength returns the length of current scanline
	ScanlineLength() int

	// AllocateCoverBuffer allocates a cover buffer for compound rendering
	AllocateCoverBuffer(len int) []basics.Int8u
}

// StyleHandlerInterface defines the interface for style handlers in compound rendering.
type StyleHandlerInterface interface {
	// IsSolid returns true if the style is a solid color
	IsSolid(style int) bool

	// Color returns the color for a solid style
	Color(style int) interface{}

	// GenerateSpan generates colors for a span with the given style
	GenerateSpan(colors []interface{}, x, y, len, style int)
}
