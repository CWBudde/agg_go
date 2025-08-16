// Package scanline provides anti-aliased span-based renderer implementation.
package scanline

// RendererScanlineAA is a scanline renderer for anti-aliased rendering with span generation.
// This corresponds to AGG's renderer_scanline_aa<BaseRenderer, SpanAllocator, SpanGenerator> template class.
type RendererScanlineAA[BR BaseRendererInterface, SA SpanAllocatorInterface, SG SpanGeneratorInterface] struct {
	baseRenderer  BR // The base renderer
	spanAllocator SA // Span allocator for color arrays
	spanGenerator SG // Span generator for creating colors
}

// NewRendererScanlineAA creates a new anti-aliased span-based renderer.
func NewRendererScanlineAA[BR BaseRendererInterface, SA SpanAllocatorInterface, SG SpanGeneratorInterface]() *RendererScanlineAA[BR, SA, SG] {
	return &RendererScanlineAA[BR, SA, SG]{}
}

// NewRendererScanlineAAWithComponents creates a new renderer with all components.
func NewRendererScanlineAAWithComponents[BR BaseRendererInterface, SA SpanAllocatorInterface, SG SpanGeneratorInterface](ren BR, alloc SA, spanGen SG) *RendererScanlineAA[BR, SA, SG] {
	return &RendererScanlineAA[BR, SA, SG]{
		baseRenderer:  ren,
		spanAllocator: alloc,
		spanGenerator: spanGen,
	}
}

// Attach attaches all components to this scanline renderer.
func (r *RendererScanlineAA[BR, SA, SG]) Attach(ren BR, alloc SA, spanGen SG) {
	r.baseRenderer = ren
	r.spanAllocator = alloc
	r.spanGenerator = spanGen
}

// AttachBaseRenderer attaches only the base renderer.
func (r *RendererScanlineAA[BR, SA, SG]) AttachBaseRenderer(ren BR) {
	r.baseRenderer = ren
}

// AttachSpanAllocator attaches only the span allocator.
func (r *RendererScanlineAA[BR, SA, SG]) AttachSpanAllocator(alloc SA) {
	r.spanAllocator = alloc
}

// AttachSpanGenerator attaches only the span generator.
func (r *RendererScanlineAA[BR, SA, SG]) AttachSpanGenerator(spanGen SG) {
	r.spanGenerator = spanGen
}

// Prepare is called before rendering begins.
// This will prepare the span generator.
func (r *RendererScanlineAA[BR, SA, SG]) Prepare() {
	r.spanGenerator.Prepare()
}

// Render renders a single scanline using span generation.
func (r *RendererScanlineAA[BR, SA, SG]) Render(sl ScanlineInterface) {
	RenderScanlineAA(sl, r.baseRenderer, r.spanAllocator, r.spanGenerator)
}

// BaseRenderer returns the underlying base renderer.
func (r *RendererScanlineAA[BR, SA, SG]) BaseRenderer() BR {
	return r.baseRenderer
}

// SpanAllocator returns the span allocator.
func (r *RendererScanlineAA[BR, SA, SG]) SpanAllocator() SA {
	return r.spanAllocator
}

// SpanGenerator returns the span generator.
func (r *RendererScanlineAA[BR, SA, SG]) SpanGenerator() SG {
	return r.spanGenerator
}
