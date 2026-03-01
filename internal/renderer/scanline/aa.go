// Package scanline provides anti-aliased span-based renderer implementation.
package scanline

// RendererScanlineAA is a scanline renderer for anti-aliased rendering with span generation.
// This corresponds to AGG's renderer_scanline_aa<BaseRenderer, SpanAllocator, SpanGenerator> template class.
type RendererScanlineAA[BR BaseRendererInterface[C], SA SpanAllocatorInterface[C], SG SpanGeneratorInterface[C], C any] struct {
	baseRenderer  BR // The base renderer
	spanAllocator SA // Span allocator for color arrays
	spanGenerator SG // Span generator for creating colors
}

// NewRendererScanlineAA creates a new anti-aliased span-based renderer.
func NewRendererScanlineAA[BR BaseRendererInterface[C], SA SpanAllocatorInterface[C], SG SpanGeneratorInterface[C], C any]() *RendererScanlineAA[BR, SA, SG, C] {
	return &RendererScanlineAA[BR, SA, SG, C]{}
}

// NewRendererScanlineAAWithComponents creates a new renderer with all components.
func NewRendererScanlineAAWithComponents[BR BaseRendererInterface[C], SA SpanAllocatorInterface[C], SG SpanGeneratorInterface[C], C any](ren BR, alloc SA, spanGen SG) *RendererScanlineAA[BR, SA, SG, C] {
	return &RendererScanlineAA[BR, SA, SG, C]{
		baseRenderer:  ren,
		spanAllocator: alloc,
		spanGenerator: spanGen,
	}
}

// SetColor sets the color for the renderer (if supported by span generator).
func (r *RendererScanlineAA[BR, SA, SG, C]) SetColor(c C) {
	// Some span generators might support setting a base color.
	// We check if the span generator implements ColorSetter.
	if cs, ok := any(r.spanGenerator).(ColorSetter[C]); ok {
		cs.SetColor(c)
	}
}

// Attach attaches all components to this scanline renderer.
func (r *RendererScanlineAA[BR, SA, SG, C]) Attach(ren BR, alloc SA, spanGen SG) {
	r.baseRenderer = ren
	r.spanAllocator = alloc
	r.spanGenerator = spanGen
}

// AttachBaseRenderer attaches only the base renderer.
func (r *RendererScanlineAA[BR, SA, SG, C]) AttachBaseRenderer(ren BR) {
	r.baseRenderer = ren
}

// AttachSpanAllocator attaches only the span allocator.
func (r *RendererScanlineAA[BR, SA, SG, C]) AttachSpanAllocator(alloc SA) {
	r.spanAllocator = alloc
}

// AttachSpanGenerator attaches only the span generator.
func (r *RendererScanlineAA[BR, SA, SG, C]) AttachSpanGenerator(spanGen SG) {
	r.spanGenerator = spanGen
}

// Prepare is called before rendering begins.
// This will prepare the span generator.
func (r *RendererScanlineAA[BR, SA, SG, C]) Prepare() {
	r.spanGenerator.Prepare()
}

// Render renders a single scanline using span generation.
func (r *RendererScanlineAA[BR, SA, SG, C]) Render(sl ScanlineInterface) {
	RenderScanlineAA(sl, r.baseRenderer, r.spanAllocator, r.spanGenerator)
}

// BaseRenderer returns the underlying base renderer.
func (r *RendererScanlineAA[BR, SA, SG, C]) BaseRenderer() BR {
	return r.baseRenderer
}

// SpanAllocator returns the span allocator.
func (r *RendererScanlineAA[BR, SA, SG, C]) SpanAllocator() SA {
	return r.spanAllocator
}

// SpanGenerator returns the span generator.
func (r *RendererScanlineAA[BR, SA, SG, C]) SpanGenerator() SG {
	return r.spanGenerator
}
