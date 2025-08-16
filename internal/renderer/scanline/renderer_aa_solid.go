// Package scanline provides anti-aliased solid color renderer implementation.
package scanline

// RendererScanlineAASolid is a scanline renderer for anti-aliased rendering with solid colors.
// This corresponds to AGG's renderer_scanline_aa_solid<BaseRenderer> template class.
type RendererScanlineAASolid[BR BaseRendererInterface] struct {
	baseRenderer BR          // The base renderer
	color        interface{} // Current solid color
}

// NewRendererScanlineAASolid creates a new anti-aliased solid color renderer.
func NewRendererScanlineAASolid[BR BaseRendererInterface]() *RendererScanlineAASolid[BR] {
	return &RendererScanlineAASolid[BR]{}
}

// NewRendererScanlineAASolidWithRenderer creates a new renderer with the given base renderer.
func NewRendererScanlineAASolidWithRenderer[BR BaseRendererInterface](ren BR) *RendererScanlineAASolid[BR] {
	return &RendererScanlineAASolid[BR]{
		baseRenderer: ren,
	}
}

// Attach attaches a base renderer to this scanline renderer.
func (r *RendererScanlineAASolid[BR]) Attach(ren BR) {
	r.baseRenderer = ren
}

// SetColor sets the solid color for rendering.
func (r *RendererScanlineAASolid[BR]) SetColor(color interface{}) {
	r.color = color
}

// Color returns the current solid color.
func (r *RendererScanlineAASolid[BR]) Color() interface{} {
	return r.color
}

// Prepare is called before rendering begins.
// For solid color rendering, no preparation is needed.
func (r *RendererScanlineAASolid[BR]) Prepare() {
	// Nothing to prepare for solid color rendering
}

// Render renders a single scanline using the solid color.
func (r *RendererScanlineAASolid[BR]) Render(sl ScanlineInterface) {
	RenderScanlineAASolid(sl, r.baseRenderer, r.color)
}

// BaseRenderer returns the underlying base renderer.
func (r *RendererScanlineAASolid[BR]) BaseRenderer() BR {
	return r.baseRenderer
}
