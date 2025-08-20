// Package span provides grayscale pattern span generation functionality for AGG.
// This implements a port of AGG's agg_span_pattern_gray.h functionality.
package span

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// SpanPatternGray generates spans from a grayscale source with offset and alpha support.
// This is a port of AGG's span_pattern_gray template class.
// It provides pattern-based rendering where a source image is used as a repeating pattern
// with configurable offset and global alpha blending.
type SpanPatternGray[Source GraySourceInterface] struct {
	source  Source
	offsetX uint
	offsetY uint
	alpha   int // Global alpha value applied to all generated pixels
}

// NewSpanPatternGray creates a new grayscale pattern span generator.
func NewSpanPatternGray[Source GraySourceInterface]() *SpanPatternGray[Source] {
	return &SpanPatternGray[Source]{
		alpha: 255, // Full opacity by default (base mask)
	}
}

// NewSpanPatternGrayWithParams creates a new grayscale pattern span generator with parameters.
func NewSpanPatternGrayWithParams[Source GraySourceInterface](
	src Source,
	offsetX, offsetY uint,
) *SpanPatternGray[Source] {
	return &SpanPatternGray[Source]{
		source:  src,
		offsetX: offsetX,
		offsetY: offsetY,
		alpha:   255, // Full opacity by default (base mask)
	}
}

// Attach attaches a source to the pattern generator.
func (sp *SpanPatternGray[Source]) Attach(src Source) {
	sp.source = src
}

// Source returns the current source.
func (sp *SpanPatternGray[Source]) Source() Source {
	return sp.source
}

// SetOffsetX sets the X offset for the pattern.
func (sp *SpanPatternGray[Source]) SetOffsetX(offset uint) {
	sp.offsetX = offset
}

// SetOffsetY sets the Y offset for the pattern.
func (sp *SpanPatternGray[Source]) SetOffsetY(offset uint) {
	sp.offsetY = offset
}

// OffsetX returns the current X offset.
func (sp *SpanPatternGray[Source]) OffsetX() uint {
	return sp.offsetX
}

// OffsetY returns the current Y offset.
func (sp *SpanPatternGray[Source]) OffsetY() uint {
	return sp.offsetY
}

// SetAlpha sets the global alpha value applied to all generated pixels.
// The alpha value should be in the range [0, 255] where 0 is fully transparent
// and 255 is fully opaque.
func (sp *SpanPatternGray[Source]) SetAlpha(alpha int) {
	sp.alpha = alpha
}

// Alpha returns the current global alpha value.
func (sp *SpanPatternGray[Source]) Alpha() int {
	return sp.alpha
}

// Prepare prepares the span generator for rendering.
// For pattern generators, this is typically a no-op as no setup is required.
func (sp *SpanPatternGray[Source]) Prepare() {
	// No preparation needed for pattern spans
}

// Generate generates a span of grayscale colors with applied pattern offset and alpha.
// The generated span contains grayscale values from the source with the configured
// offset applied and the global alpha value set for each pixel.
func (sp *SpanPatternGray[Source]) Generate(span []color.Gray8[color.Linear], x, y int, length uint) {
	// Apply the pattern offset to the coordinates
	sourceX := x + int(sp.offsetX)
	sourceY := y + int(sp.offsetY)

	// Get the source span data starting at the offset coordinates
	sourceData := sp.source.Span(sourceX, sourceY, int(length))

	// Generate the output span
	for i := uint(0); i < length; i++ {
		// Get the grayscale value from source
		var grayValue basics.Int8u
		if len(sourceData) > int(i) {
			grayValue = sourceData[i]
		} else {
			// If we run out of source data, get next pixel
			nextData := sp.source.NextX()
			if len(nextData) > 0 {
				grayValue = nextData[0]
			} else {
				grayValue = 0
			}
		}

		// Create the output color with the source grayscale value and global alpha
		span[i] = color.Gray8[color.Linear]{
			V: grayValue,
			A: basics.Int8u(sp.alpha),
		}
	}
}
