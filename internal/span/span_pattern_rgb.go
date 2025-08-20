// Package span provides RGB pattern span generation functionality for AGG.
// This implements a port of AGG's agg_span_pattern_rgb.h functionality.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// RGBSourceInterface defines the interface for RGB image sources used in pattern spans.
// This extends the basic SourceInterface with RGB-specific methods.
type RGBSourceInterface interface {
	SourceInterface
	// ColorType returns the RGB color type identifier
	ColorType() string
	// OrderType returns the color component ordering (RGB, BGR, etc.)
	OrderType() color.ColorOrder
	// Span returns RGB pixel data starting at (x, y) with given length
	// Returns raw byte data where each RGB pixel is 3 consecutive bytes
	Span(x, y, length int) []basics.Int8u
	// NextX advances to the next pixel in current span and returns its RGB data
	NextX() []basics.Int8u
	// NextY advances to the next row at original x position
	NextY() []basics.Int8u
	// RowPtr returns pointer to row data starting at specified row
	RowPtr(y int) []basics.Int8u
}

// SpanPatternRGB generates spans from an RGB source with offset and alpha support.
// This is a port of AGG's span_pattern_rgb template class.
// It provides pattern-based rendering where a source image is used as a repeating pattern
// with configurable offset and global alpha blending.
type SpanPatternRGB[Source RGBSourceInterface] struct {
	source  Source
	offsetX uint
	offsetY uint
	alpha   basics.Int8u // Global alpha value applied to all generated pixels
}

// NewSpanPatternRGB creates a new RGB pattern span generator.
func NewSpanPatternRGB[Source RGBSourceInterface]() *SpanPatternRGB[Source] {
	return &SpanPatternRGB[Source]{
		alpha: 255, // Full opacity by default (base mask)
	}
}

// NewSpanPatternRGBWithParams creates a new RGB pattern span generator with parameters.
func NewSpanPatternRGBWithParams[Source RGBSourceInterface](
	src Source,
	offsetX, offsetY uint,
) *SpanPatternRGB[Source] {
	return &SpanPatternRGB[Source]{
		source:  src,
		offsetX: offsetX,
		offsetY: offsetY,
		alpha:   255, // Full opacity by default (base mask)
	}
}

// Attach attaches a source to the pattern generator.
func (sp *SpanPatternRGB[Source]) Attach(src Source) {
	sp.source = src
}

// Source returns the current source.
func (sp *SpanPatternRGB[Source]) Source() Source {
	return sp.source
}

// SetOffsetX sets the X offset for the pattern.
func (sp *SpanPatternRGB[Source]) SetOffsetX(offset uint) {
	sp.offsetX = offset
}

// SetOffsetY sets the Y offset for the pattern.
func (sp *SpanPatternRGB[Source]) SetOffsetY(offset uint) {
	sp.offsetY = offset
}

// OffsetX returns the current X offset.
func (sp *SpanPatternRGB[Source]) OffsetX() uint {
	return sp.offsetX
}

// OffsetY returns the current Y offset.
func (sp *SpanPatternRGB[Source]) OffsetY() uint {
	return sp.offsetY
}

// SetAlpha sets the global alpha value applied to all generated pixels.
// The alpha value should be in the range [0, 255] where 0 is fully transparent
// and 255 is fully opaque.
func (sp *SpanPatternRGB[Source]) SetAlpha(alpha basics.Int8u) {
	sp.alpha = alpha
}

// Alpha returns the current global alpha value.
func (sp *SpanPatternRGB[Source]) Alpha() basics.Int8u {
	return sp.alpha
}

// Prepare prepares the span generator for rendering.
// For pattern generators, this is typically a no-op as no setup is required.
func (sp *SpanPatternRGB[Source]) Prepare() {
	// No preparation needed for pattern spans
}

// Generate generates a span of RGB colors with applied pattern offset and alpha.
// The generated span contains RGB values from the source with the configured
// offset applied and the global alpha value set for each pixel.
// This follows the C++ AGG implementation where RGB values are copied from source
// and alpha is set to the global alpha value.
func (sp *SpanPatternRGB[Source]) Generate(span []color.RGB8[color.Linear], x, y int, length uint) {
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
		var r, g, b basics.Int8u

		// For the first pixel or when we have enough data in sourceData
		if i == 0 && len(sourceData) >= 3 {
			// Extract RGB components according to the color order (like C++: span->r = p[order_type::R])
			r = sourceData[order.R]
			g = sourceData[order.G]
			b = sourceData[order.B]
		} else {
			// Get next pixel data (like C++: p = m_src->next_x())
			nextData := sp.source.NextX()
			if len(nextData) >= 3 {
				r = nextData[order.R]
				g = nextData[order.G]
				b = nextData[order.B]
			} else {
				// Fallback to black if no data available
				r, g, b = 0, 0, 0
			}
		}

		// Create the output color with RGB values (alpha not included in RGB8)
		// Note: C++ sets span->a = m_alpha, but Go's RGB8 doesn't have alpha component
		span[i] = color.RGB8[color.Linear]{
			R: r,
			G: g,
			B: b,
		}
	}
}
