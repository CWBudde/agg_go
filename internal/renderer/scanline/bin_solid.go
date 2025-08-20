// Package scanline provides binary solid color renderer implementation.
package scanline

// RendererScanlineBinSolid is a scanline renderer for binary (non-anti-aliased) rendering with solid colors.
// This corresponds to AGG's renderer_scanline_bin_solid<BaseRenderer> template class.
type RendererScanlineBinSolid[BR BaseRendererInterface] struct {
	baseRenderer BR          // The base renderer
	color        interface{} // Current solid color
}

// NewRendererScanlineBinSolid creates a new binary solid color renderer.
func NewRendererScanlineBinSolid[BR BaseRendererInterface]() *RendererScanlineBinSolid[BR] {
	return &RendererScanlineBinSolid[BR]{}
}

// NewRendererScanlineBinSolidWithRenderer creates a new renderer with the given base renderer.
func NewRendererScanlineBinSolidWithRenderer[BR BaseRendererInterface](ren BR) *RendererScanlineBinSolid[BR] {
	return &RendererScanlineBinSolid[BR]{
		baseRenderer: ren,
	}
}

// Attach attaches a base renderer to this scanline renderer.
func (r *RendererScanlineBinSolid[BR]) Attach(ren BR) {
	r.baseRenderer = ren
}

// SetColor sets the solid color for rendering.
func (r *RendererScanlineBinSolid[BR]) SetColor(color interface{}) {
	r.color = color
}

// Color returns the current solid color.
func (r *RendererScanlineBinSolid[BR]) Color() interface{} {
	return r.color
}

// Prepare is called before rendering begins.
// For solid color rendering, no preparation is needed.
func (r *RendererScanlineBinSolid[BR]) Prepare() {
	// Nothing to prepare for solid color rendering
}

// Render renders a single scanline using the solid color.
func (r *RendererScanlineBinSolid[BR]) Render(sl ScanlineInterface) {
	RenderScanlineBinSolid(sl, r.baseRenderer, r.color)
}

// BaseRenderer returns the underlying base renderer.
func (r *RendererScanlineBinSolid[BR]) BaseRenderer() BR {
	return r.baseRenderer
}
