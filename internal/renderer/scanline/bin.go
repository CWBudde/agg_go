// Package scanline provides binary span-based renderer implementation.
package scanline

// RendererScanlineBin is a scanline renderer for binary (non-anti-aliased) rendering with span generation.
// This corresponds to AGG's renderer_scanline_bin<BaseRenderer, SpanAllocator, SpanGenerator> template class.
type RendererScanlineBin[BR BaseRendererInterface[C], SA SpanAllocatorInterface[C], SG SpanGeneratorInterface[C], C any] struct {
	baseRenderer  BR // The base renderer
	spanAllocator SA // Span allocator for color arrays
	spanGenerator SG // Span generator for creating colors
}

// NewRendererScanlineBin creates a new binary span-based renderer.
func NewRendererScanlineBin[BR BaseRendererInterface[C], SA SpanAllocatorInterface[C], SG SpanGeneratorInterface[C], C any]() *RendererScanlineBin[BR, SA, SG, C] {
	return &RendererScanlineBin[BR, SA, SG, C]{}
}

// NewRendererScanlineBinWithComponents creates a new renderer with all components.
func NewRendererScanlineBinWithComponents[BR BaseRendererInterface[C], SA SpanAllocatorInterface[C], SG SpanGeneratorInterface[C], C any](ren BR, alloc SA, spanGen SG) *RendererScanlineBin[BR, SA, SG, C] {
	return &RendererScanlineBin[BR, SA, SG, C]{
		baseRenderer:  ren,
		spanAllocator: alloc,
		spanGenerator: spanGen,
	}
}

// Attach attaches all components to this scanline renderer.
func (r *RendererScanlineBin[BR, SA, SG, C]) Attach(ren BR, alloc SA, spanGen SG) {
	r.baseRenderer = ren
	r.spanAllocator = alloc
	r.spanGenerator = spanGen
}

// AttachBaseRenderer attaches only the base renderer.
func (r *RendererScanlineBin[BR, SA, SG, C]) AttachBaseRenderer(ren BR) {
	r.baseRenderer = ren
}

// AttachSpanAllocator attaches only the span allocator.
func (r *RendererScanlineBin[BR, SA, SG, C]) AttachSpanAllocator(alloc SA) {
	r.spanAllocator = alloc
}

// AttachSpanGenerator attaches only the span generator.
func (r *RendererScanlineBin[BR, SA, SG, C]) AttachSpanGenerator(spanGen SG) {
	r.spanGenerator = spanGen
}

// Prepare is called before rendering begins.
// This will prepare the span generator.
func (r *RendererScanlineBin[BR, SA, SG, C]) Prepare() {
	r.spanGenerator.Prepare()
}

// Render renders a single scanline using span generation.
func (r *RendererScanlineBin[BR, SA, SG, C]) Render(sl ScanlineInterface) {
	RenderScanlineBin(sl, r.baseRenderer, r.spanAllocator, r.spanGenerator)
}

// BaseRenderer returns the underlying base renderer.
func (r *RendererScanlineBin[BR, SA, SG, C]) BaseRenderer() BR {
	return r.baseRenderer
}

// SpanAllocator returns the span allocator.
func (r *RendererScanlineBin[BR, SA, SG, C]) SpanAllocator() SA {
	return r.spanAllocator
}

// SpanGenerator returns the span generator.
func (r *RendererScanlineBin[BR, SA, SG, C]) SpanGenerator() SG {
	return r.spanGenerator
}
