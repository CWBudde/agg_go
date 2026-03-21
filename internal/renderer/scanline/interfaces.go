package scanline

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	sl "github.com/MeKo-Christian/agg_go/internal/scanline"
)

// ScanlineInterface is the unified scanline contract. Concrete scanline types
// (ScanlineP8, ScanlineU8, ScanlineBin, etc.) satisfy this interface directly,
// eliminating the need for adapter boilerplate between rasterizer and renderer.
type ScanlineInterface = sl.Scanline

// ScanlineIterator iterates over the spans stored in one scanline.
type ScanlineIterator = sl.ScanlineIterator

// SpanData stores one contiguous coverage run within a scanline.
type SpanData = sl.SpanInfo

// RasterizerInterface is the sweep contract expected by AGG-style render helpers.
// The rasterizer fills a Scanline during SweepScanline; the renderer reads it back.
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

// BaseRendererInterface is the clipped pixel-write contract scanline renderers need.
type BaseRendererInterface[C any] interface {
	// BlendSolidHspan blends a horizontal span with solid color and coverage array
	BlendSolidHspan(x, y, len int, color C, covers []basics.Int8u)

	// BlendHline blends a horizontal line with solid color and single coverage
	BlendHline(x, y, x2 int, color C, cover basics.Int8u)

	// BlendColorHspan blends a horizontal span with color array and coverage
	BlendColorHspan(x, y, len int, colors []C, covers []basics.Int8u, cover basics.Int8u)
}

// SpanAllocatorInterface allocates temporary color buffers for generated spans.
type SpanAllocatorInterface[C any] interface {
	// Allocate allocates an array of colors for the given length
	// Returns a slice that can hold 'len' color values
	Allocate(len int) []C
}

// SpanGeneratorInterface generates colors for a requested span.
type SpanGeneratorInterface[C any] interface {
	// Prepare is called before rendering begins
	Prepare()

	// Generate fills the colors array with generated colors for the given span
	Generate(colors []C, x, y, len int)
}

// StyleHandlerInterface resolves solid colors or generated spans for compound styles.
type StyleHandlerInterface[C any] interface {
	// IsSolid returns true if the style is a solid color
	IsSolid(style int) bool

	// Color returns the color for a solid style
	Color(style int) C

	// GenerateSpan generates colors for a span with the given style
	GenerateSpan(colors []C, x, y, len, style int)
}

// ColorSetter is implemented by renderers that expose a mutable current color.
type ColorSetter[C any] interface {
	// SetColor sets the current color for the object
	SetColor(color C)
}

// RendererInterface is the common prepare/render contract shared by scanline renderers.
type RendererInterface[C any] interface {
	ColorSetter[C]

	// Prepare is called before rendering begins
	Prepare()

	// Render renders a single scanline
	Render(sl ScanlineInterface)
}

// CompoundRasterizerInterface extends RasterizerInterface with style-aware scanline sweeping.
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

// Resettable is the minimal reusable-state contract used by helper routines.
type Resettable interface {
	// Reset resets the object to its initial state
	Reset()
}
