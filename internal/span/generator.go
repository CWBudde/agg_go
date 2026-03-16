package span

// SpanColorType is the color constraint used by generic span allocators and
// generators.
type SpanColorType interface {
	any
}

// SpanGenerator is the common contract shared by AGG-style span generators.
// Prepare lets generators cache state before rasterization starts, and Generate
// expands one horizontal run into concrete colors.
type SpanGenerator[C SpanColorType] interface {
	Prepare()
	Generate(colors []C, x, y, len int)
}

// SolidSpanGenerator is the trivial span generator that emits the same color for
// every pixel in the span.
type SolidSpanGenerator[C SpanColorType] struct {
	color C
}

// NewSolidSpanGenerator creates a solid-color span generator.
func NewSolidSpanGenerator[C SpanColorType](color C) *SolidSpanGenerator[C] {
	return &SolidSpanGenerator[C]{
		color: color,
	}
}

// SetColor replaces the emitted solid color.
func (sg *SolidSpanGenerator[C]) SetColor(color C) {
	sg.color = color
}

// Color returns the current solid color.
func (sg *SolidSpanGenerator[C]) Color() C {
	return sg.color
}

// Prepare is a no-op for solid spans.
func (sg *SolidSpanGenerator[C]) Prepare() {
}

// Generate fills the requested run with the solid color.
func (sg *SolidSpanGenerator[C]) Generate(colors []C, x, y, length int) {
	for i := 0; i < length; i++ {
		colors[i] = sg.color
	}
}
