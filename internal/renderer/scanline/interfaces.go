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

// Typed interfaces for better type safety and performance
// These interfaces use concrete color types instead of interface{}

// BaseRendererInterface defines the interface for base renderers.
// This corresponds to the BaseRenderer template parameter in AGG's functions.
type BaseRendererInterface[C any] interface {
	// BlendSolidHspan blends a horizontal span with solid color and coverage array
	BlendSolidHspan(x, y, len int, color C, covers []basics.Int8u)

	// BlendHline blends a horizontal line with solid color and single coverage
	BlendHline(x, y, x2 int, color C, cover basics.Int8u)

	// BlendColorHspan blends a horizontal span with color array and coverage
	BlendColorHspan(x, y, len int, colors []C, covers []basics.Int8u, cover basics.Int8u)
}

// SpanAllocatorInterface defines the interface for span allocators.
// This corresponds to the SpanAllocator template parameter in AGG's functions.
type SpanAllocatorInterface[C any] interface {
	// Allocate allocates an array of colors for the given length
	// Returns a slice that can hold 'len' color values
	Allocate(len int) []C
}

// SpanGeneratorInterface defines the interface for span generators.
// This corresponds to the SpanGenerator template parameter in AGG's functions.
type SpanGeneratorInterface[C any] interface {
	// Prepare is called before rendering begins
	Prepare()

	// Generate fills the colors array with generated colors for the given span
	Generate(colors []C, x, y, len int)
}

// StyleHandlerInterface defines the interface for style handlers in compound rendering.
type StyleHandlerInterface[C any] interface {
	// IsSolid returns true if the style is a solid color
	IsSolid(style int) bool

	// Color returns the color for a solid style
	Color(style int) C

	// GenerateSpan generates colors for a span with the given style
	GenerateSpan(colors []C, x, y, len, style int)
}

// ColorSetter defines the interface for objects that can have their color set.
// This interface is used by renderers and other objects that need to maintain a current color.
type ColorSetter[C any] interface {
	// SetColor sets the current color for the object
	SetColor(color C)
}

// RendererInterface defines the interface for scanline renderers.
// This provides a common interface for all scanline renderer implementations.
type RendererInterface[C any] interface {
	ColorSetter[C]

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

// ResettableScanline defines the interface for scanlines that can be reset.
// This interface is implemented by scanline types that support resetting
// their bounds and internal state for reuse across multiple rendering passes.
type ResettableScanline interface {
	ScanlineInterface

	// Reset resets the scanline for the given horizontal bounds.
	// This prepares the scanline for a new rendering pass within the specified X range.
	Reset(minX, maxX int)
}

// Resettable defines the interface for objects that can be reset to their initial state.
// This interface is used for objects that support resetting without parameters.
type Resettable interface {
	// Reset resets the object to its initial state
	Reset()
}

// Compile-time interface checks
// These ensure that expected types implement the required interfaces at compile time.
// If a type doesn't implement an interface, the compilation will fail with a clear error.

// Ensure common scanline types implement ResettableScanline
// These checks should be added in the scanline package implementations to avoid import cycles:
// var _ renderer.ResettableScanline = (*ScanlineU8)(nil)
// var _ renderer.ResettableScanline = (*ScanlineP8)(nil)
// var _ renderer.ResettableScanline = (*ScanlineBin)(nil)
// var _ renderer.ResettableScanline = (*Scanline32U8)(nil)
