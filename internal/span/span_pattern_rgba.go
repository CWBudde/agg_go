package span

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
)

// RGBASourceInterface is the source contract for AGG-style RGBA pattern spans.
// It exposes row/span stepping methods so the generator can mirror AGG's
// pointer-based next_x/next_y traversal without copying the whole source row.
type RGBASourceInterface interface {
	SourceInterface
	ColorType() string
	OrderType() color.ColorOrder
	Span(x, y, length int) []basics.Int8u
	NextX() []basics.Int8u
	NextY() []basics.Int8u
	RowPtr(y int) []basics.Int8u
}

// SpanPatternRGBA is the Go equivalent of AGG's span_pattern_rgba template. It
// reads source pixels with an x/y offset and emits them directly into the
// destination span, including per-pixel alpha.
type SpanPatternRGBA[Source RGBASourceInterface] struct {
	source  Source
	offsetX uint
	offsetY uint
}

// NewSpanPatternRGBA creates an unattached RGBA pattern generator.
func NewSpanPatternRGBA[Source RGBASourceInterface]() *SpanPatternRGBA[Source] {
	return &SpanPatternRGBA[Source]{}
}

// NewSpanPatternRGBAWithParams creates an attached RGBA pattern generator.
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

// Attach replaces the pattern source.
func (sp *SpanPatternRGBA[Source]) Attach(src Source) {
	sp.source = src
}

// Source returns the current source.
func (sp *SpanPatternRGBA[Source]) Source() Source {
	return sp.source
}

// SetOffsetX sets the x offset applied before sampling.
func (sp *SpanPatternRGBA[Source]) SetOffsetX(offset uint) {
	sp.offsetX = offset
}

// SetOffsetY sets the y offset applied before sampling.
func (sp *SpanPatternRGBA[Source]) SetOffsetY(offset uint) {
	sp.offsetY = offset
}

// OffsetX returns the current x offset.
func (sp *SpanPatternRGBA[Source]) OffsetX() uint {
	return sp.offsetX
}

// OffsetY returns the current y offset.
func (sp *SpanPatternRGBA[Source]) OffsetY() uint {
	return sp.offsetY
}

// SetAlpha exists for API compatibility with the RGB pattern family, but it is
// a no-op because RGBA patterns already carry alpha in the source pixels.
func (sp *SpanPatternRGBA[Source]) SetAlpha(alpha basics.Int8u) {
}

// Alpha returns 0, matching AGG's convention that RGBA pattern generators do
// not carry a separate global alpha value.
func (sp *SpanPatternRGBA[Source]) Alpha() basics.Int8u {
	return 0
}

// Prepare is a no-op for pattern generators.
func (sp *SpanPatternRGBA[Source]) Prepare() {
}

// Generate copies one run of source pixels into span after applying the stored
// offset, following the same first-pixel/next_x stepping structure as AGG's
// span_pattern_rgba.
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
