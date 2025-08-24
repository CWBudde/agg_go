// Package scanline provides binary solid color renderer implementation.
package scanline

// RendererScanlineBinSolid is a scanline renderer for binary (non-anti-aliased) rendering with solid colors.
// This corresponds to AGG's renderer_scanline_bin_solid<BaseRenderer> template class.
type RendererScanlineBinSolid[BR BaseRendererInterface[C], C any] struct {
	baseRenderer BR // The base renderer
	color        C  // Current solid color
}

// NewRendererScanlineBinSolid creates a new binary solid color renderer.
func NewRendererScanlineBinSolid[BR BaseRendererInterface[C], C any]() *RendererScanlineBinSolid[BR, C] {
	var zero C
	return &RendererScanlineBinSolid[BR, C]{
		color: zero,
	}
}

// NewRendererScanlineBinSolidWithRenderer creates a new renderer with the given base renderer.
func NewRendererScanlineBinSolidWithRenderer[BR BaseRendererInterface[C], C any](ren BR) *RendererScanlineBinSolid[BR, C] {
	var zero C
	return &RendererScanlineBinSolid[BR, C]{
		baseRenderer: ren,
		color:        zero,
	}
}

// NewRendererScanlineBinSolidWithColor creates a new renderer with the given base renderer and color.
func NewRendererScanlineBinSolidWithColor[BR BaseRendererInterface[C], C any](ren BR, color C) *RendererScanlineBinSolid[BR, C] {
	return &RendererScanlineBinSolid[BR, C]{
		baseRenderer: ren,
		color:        color,
	}
}

// Attach attaches a base renderer to this scanline renderer.
func (r *RendererScanlineBinSolid[BR, C]) Attach(ren BR) {
	r.baseRenderer = ren
}

// SetColor sets the solid color for rendering.
func (r *RendererScanlineBinSolid[BR, C]) SetColor(color C) {
	r.color = color
}

// Color returns the current solid color.
func (r *RendererScanlineBinSolid[BR, C]) Color() C {
	return r.color
}

// Prepare is called before rendering begins.
// For solid color rendering, no preparation is needed.
func (r *RendererScanlineBinSolid[BR, C]) Prepare() {
	// Nothing to prepare for solid color rendering
}

// Render renders a single scanline using the solid color.
func (r *RendererScanlineBinSolid[BR, C]) Render(sl ScanlineInterface) {
	RenderScanlineBinSolid(sl, r.baseRenderer, r.color)
}

// BaseRenderer returns the underlying base renderer.
func (r *RendererScanlineBinSolid[BR, C]) BaseRenderer() BR {
	return r.baseRenderer
}
