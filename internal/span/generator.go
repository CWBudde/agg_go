// Package span provides span generation functionality for AGG rendering.
package span

// SpanGenerator provides the interface for generating colors across a span.
// This is the base interface that specific span generators should implement.
type SpanGenerator interface {
	// Prepare is called before rendering begins
	Prepare()

	// Generate fills the colors array with generated colors for the given span
	Generate(colors []interface{}, x, y, len int)
}

// SolidSpanGenerator generates solid colors for spans.
// This is the simplest span generator that fills all pixels with the same color.
type SolidSpanGenerator struct {
	color interface{} // The solid color to generate
}

// NewSolidSpanGenerator creates a new solid span generator.
func NewSolidSpanGenerator(color interface{}) *SolidSpanGenerator {
	return &SolidSpanGenerator{
		color: color,
	}
}

// SetColor sets the solid color for this generator.
func (sg *SolidSpanGenerator) SetColor(color interface{}) {
	sg.color = color
}

// Color returns the current solid color.
func (sg *SolidSpanGenerator) Color() interface{} {
	return sg.color
}

// Prepare is called before rendering begins.
// For solid color generation, no preparation is needed.
func (sg *SolidSpanGenerator) Prepare() {
	// Nothing to prepare for solid colors
}

// Generate fills the colors array with the solid color.
func (sg *SolidSpanGenerator) Generate(colors []interface{}, x, y, len int) {
	// Fill all positions with the solid color
	for i := 0; i < len; i++ {
		colors[i] = sg.color
	}
}
