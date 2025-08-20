// Package span provides span generation functionality for AGG rendering.
package span

// SpanColorType defines the constraint for color types used in span operations.
// This ensures type safety while allowing different color representations.
type SpanColorType interface {
	// Any type can be used as a span color type
	// This allows struct types, but also other types for compatibility
	any
}

// SpanGenerator provides the interface for generating colors across a span.
// This is the base interface that specific span generators should implement.
type SpanGenerator[C SpanColorType] interface {
	// Prepare is called before rendering begins
	Prepare()

	// Generate fills the colors array with generated colors for the given span
	Generate(colors []C, x, y, len int)
}

// SolidSpanGenerator generates solid colors for spans.
// This is the simplest span generator that fills all pixels with the same color.
type SolidSpanGenerator[C SpanColorType] struct {
	color C // The solid color to generate
}

// NewSolidSpanGenerator creates a new solid span generator.
func NewSolidSpanGenerator[C SpanColorType](color C) *SolidSpanGenerator[C] {
	return &SolidSpanGenerator[C]{
		color: color,
	}
}

// SetColor sets the solid color for this generator.
func (sg *SolidSpanGenerator[C]) SetColor(color C) {
	sg.color = color
}

// Color returns the current solid color.
func (sg *SolidSpanGenerator[C]) Color() C {
	return sg.color
}

// Prepare is called before rendering begins.
// For solid color generation, no preparation is needed.
func (sg *SolidSpanGenerator[C]) Prepare() {
	// Nothing to prepare for solid colors
}

// Generate fills the colors array with the solid color.
func (sg *SolidSpanGenerator[C]) Generate(colors []C, x, y, len int) {
	// Fill all positions with the solid color
	for i := 0; i < len; i++ {
		colors[i] = sg.color
	}
}
