// Package scanline provides anti-aliased solid color renderer implementation.
package scanline

// RendererScanlineAASolid is a type-safe scanline renderer for anti-aliased rendering with solid colors.
// This corresponds to AGG's renderer_scanline_aa_solid<BaseRenderer> template class.
type RendererScanlineAASolid[BR BaseRendererInterface[C], C any] struct {
	baseRenderer BR // The base renderer
	color        C  // Current solid color
}

// NewRendererScanlineAASolid creates a new anti-aliased solid color renderer.
func NewRendererScanlineAASolid[BR BaseRendererInterface[C], C any]() *RendererScanlineAASolid[BR, C] {
	var zero C
	return &RendererScanlineAASolid[BR, C]{
		color: zero,
	}
}

// NewRendererScanlineAASolidWithRenderer creates a new renderer with the given base renderer.
func NewRendererScanlineAASolidWithRenderer[BR BaseRendererInterface[C], C any](ren BR) *RendererScanlineAASolid[BR, C] {
	var zero C
	return &RendererScanlineAASolid[BR, C]{
		baseRenderer: ren,
		color:        zero,
	}
}

// NewRendererScanlineAASolidWithColor creates a new renderer with the given base renderer and color.
func NewRendererScanlineAASolidWithColor[BR BaseRendererInterface[C], C any](ren BR, color C) *RendererScanlineAASolid[BR, C] {
	return &RendererScanlineAASolid[BR, C]{
		baseRenderer: ren,
		color:        color,
	}
}

// Attach attaches a base renderer to this scanline renderer.
func (r *RendererScanlineAASolid[BR, C]) Attach(ren BR) {
	r.baseRenderer = ren
}

// SetColor sets the solid color for rendering.
func (r *RendererScanlineAASolid[BR, C]) SetColor(color C) {
	r.color = color
}

// Color returns the current solid color.
func (r *RendererScanlineAASolid[BR, C]) Color() C {
	return r.color
}

// Prepare is called before rendering begins.
// For solid color rendering, no preparation is needed.
func (r *RendererScanlineAASolid[BR, C]) Prepare() {
	// Nothing to prepare for solid color rendering
}

// Render renders a single scanline using the solid color.
func (r *RendererScanlineAASolid[BR, C]) Render(sl ScanlineInterface) {
	RenderScanlineAASolid(sl, r.baseRenderer, r.color)
}

// BaseRenderer returns the underlying base renderer.
func (r *RendererScanlineAASolid[BR, C]) BaseRenderer() BR {
	return r.baseRenderer
}

// Ensure RendererScanlineAASolid implements RendererInterface
var _ RendererInterface[any] = (*RendererScanlineAASolid[BaseRendererInterface[any], any])(nil)
