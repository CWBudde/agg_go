// Package span provides RGBA pattern span generation functionality for AGG.
// This implements a port of AGG's agg_span_pattern_rgba.h functionality.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// RGBASourceInterface defines the interface for RGBA image sources used in pattern spans.
// This extends the basic SourceInterface with RGBA-specific methods.
type RGBASourceInterface interface {
	SourceInterface
	// ColorType returns the RGBA color type identifier
	ColorType() string
	// OrderType returns the color component ordering (RGBA, BGRA, ARGB, etc.)
	OrderType() color.ColorOrder
	// Span returns RGBA pixel data starting at (x, y) with given length
	// Returns raw byte data where each RGBA pixel is 4 consecutive bytes
	Span(x, y, length int) []basics.Int8u
	// NextX advances to the next pixel in current span and returns its RGBA data
	NextX() []basics.Int8u
	// NextY advances to the next row at original x position
	NextY() []basics.Int8u
	// RowPtr returns pointer to row data starting at specified row
	RowPtr(y int) []basics.Int8u
}

// SpanPatternRGBA generates spans from an RGBA source with offset support.
// This is a port of AGG's span_pattern_rgba template class.
// It provides pattern-based rendering where a source image is used as a repeating pattern
// with configurable offset. Unlike the RGB version, this doesn't have a separate alpha
// parameter since alpha is part of the RGBA data.
type SpanPatternRGBA[Source RGBASourceInterface] struct {
	source  Source
	offsetX uint
	offsetY uint
}

// NewSpanPatternRGBA creates a new RGBA pattern span generator.
func NewSpanPatternRGBA[Source RGBASourceInterface]() *SpanPatternRGBA[Source] {
	return &SpanPatternRGBA[Source]{}
}

// NewSpanPatternRGBAWithParams creates a new RGBA pattern span generator with parameters.
func NewSpanPatternRGBAWithParams[Source RGBASourceInterface](
	src Source,
	offsetX, offsetY uint,
) *SpanPatternRGBA[Source] {
	return &SpanPatternRGBA[Source]{
		source:  src,
		offsetX: offsetX,
		offsetY: offsetY,
	}
}

// Attach attaches a source to the pattern generator.
func (sp *SpanPatternRGBA[Source]) Attach(src Source) {
	sp.source = src
}

// Source returns the current source.
func (sp *SpanPatternRGBA[Source]) Source() Source {
	return sp.source
}

// SetOffsetX sets the X offset for the pattern.
func (sp *SpanPatternRGBA[Source]) SetOffsetX(offset uint) {
	sp.offsetX = offset
}

// SetOffsetY sets the Y offset for the pattern.
func (sp *SpanPatternRGBA[Source]) SetOffsetY(offset uint) {
	sp.offsetY = offset
}

// OffsetX returns the current X offset.
func (sp *SpanPatternRGBA[Source]) OffsetX() uint {
	return sp.offsetX
}

// OffsetY returns the current Y offset.
func (sp *SpanPatternRGBA[Source]) OffsetY() uint {
	return sp.offsetY
}

// SetAlpha sets the alpha value. For RGBA patterns, this is a no-op since
// alpha comes from the source data itself. This method exists for compatibility
// with the C++ AGG interface.
func (sp *SpanPatternRGBA[Source]) SetAlpha(alpha basics.Int8u) {
	// No-op: RGBA patterns use alpha from source data, not a global alpha
}

// Alpha returns the alpha value. For RGBA patterns, this always returns 0
// since alpha comes from the source data, not a global value.
// This method exists for compatibility with the C++ AGG interface.
func (sp *SpanPatternRGBA[Source]) Alpha() basics.Int8u {
	return 0 // Always 0 as per C++ implementation
}

// Prepare prepares the span generator for rendering.
// For pattern generators, this is typically a no-op as no setup is required.
func (sp *SpanPatternRGBA[Source]) Prepare() {
	// No preparation needed for pattern spans
}

// Generate generates a span of RGBA colors with applied pattern offset.
// The generated span contains RGBA values from the source with the configured
// offset applied. This follows the C++ AGG implementation where RGBA values
// are copied directly from source, including the alpha channel.
func (sp *SpanPatternRGBA[Source]) Generate(span []color.RGBA8[color.Linear], x, y int, length uint) {
	if length == 0 {
		return
	}

	// Apply the pattern offset to the coordinates (like C++: x += m_offset_x; y += m_offset_y;)
	sourceX := x + int(sp.offsetX)
	sourceY := y + int(sp.offsetY)

	// Get the first pixel's data from source (like C++: const value_type* p = ...)
	sourceData := sp.source.Span(sourceX, sourceY, int(length))
	order := sp.source.OrderType()

	// Generate the output span following C++ do-while loop structure
	for i := uint(0); i < length; i++ {
		var r, g, b, a basics.Int8u

		// For the first pixel or when we have enough data in sourceData
		if i == 0 && len(sourceData) >= 4 {
			// Extract RGBA components according to the color order (like C++: span->r = p[order_type::R])
			r = sourceData[order.R]
			g = sourceData[order.G]
			b = sourceData[order.B]
			a = sourceData[order.A]
		} else {
			// Get next pixel data (like C++: p = (const value_type*)m_src->next_x())
			nextData := sp.source.NextX()
			if len(nextData) >= 4 {
				r = nextData[order.R]
				g = nextData[order.G]
				b = nextData[order.B]
				a = nextData[order.A]
			} else {
				// Fallback to black if no data available
				r, g, b, a = 0, 0, 0, 0
			}
		}

		// Create the output color with RGBA values from source
		span[i] = color.RGBA8[color.Linear]{
			R: r,
			G: g,
			B: b,
			A: a,
		}
	}
}
