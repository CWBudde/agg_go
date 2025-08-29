// Package renderer provides specialized renderers for AGG.
// This typed variant of RendererEnlarged works with a concrete color type C
// and a typed base renderer.
package renderer

import (
	"agg_go/internal/basics"
	aggcolor "agg_go/internal/color"
	scanline_renderer "agg_go/internal/renderer/scanline"
)

// BaseRendererTWithCopyBar defines the minimal typed base renderer surface
// needed by the enlarged renderer (CopyBar with typed color).
type BaseRendererTWithCopyBar[C any] interface {
	CopyBar(x1, y1, x2, y2 int, c C)
}

// AlphaModFunc modifies the input color C by the given coverage (0..255),
// typically by scaling its alpha channel.
type AlphaModFunc[C any] func(c C, cover basics.Int8u) C

// RendererEnlargedT magnifies pixels for visualization using a typed base renderer.
type RendererEnlargedT[Ren BaseRendererTWithCopyBar[C], C any] struct {
	baseRenderer Ren
	size         float64
	color        C
	modAlpha     AlphaModFunc[C]
}

// NewRendererEnlargedT creates a new typed enlarged renderer.
// modAlpha can be nil; in that case, coverage is ignored and the color is used as-is.
func NewRendererEnlargedT[Ren BaseRendererTWithCopyBar[C], C any](
	baseRenderer Ren, size float64, modAlpha AlphaModFunc[C],
) *RendererEnlargedT[Ren, C] {
	return &RendererEnlargedT[Ren, C]{
		baseRenderer: baseRenderer,
		size:         size,
		modAlpha:     modAlpha,
	}
}

// Color returns the current color for rendering.
func (r *RendererEnlargedT[Ren, C]) Color() C { return r.color }

// SetColor sets the current color for rendering.
func (r *RendererEnlargedT[Ren, C]) SetColor(color C) { r.color = color }

// Prepare prepares the renderer for rendering (no-op for this renderer).
func (r *RendererEnlargedT[Ren, C]) Prepare() {}

// Render renders a scanline with pixel magnification.
func (r *RendererEnlargedT[Ren, C]) Render(sl scanline_renderer.ScanlineInterface) {
	y := sl.Y()
	numSpans := sl.NumSpans()
	if numSpans == 0 {
		return
	}

	it := sl.Begin()
	for i := 0; i < numSpans; i++ {
		sp := it.GetSpan()
		x := sp.X
		covers := sp.Covers
		n := sp.Len

		for j := 0; j < n; j++ {
			if j < len(covers) {
				cover := covers[j]
				c := r.color
				if r.modAlpha != nil {
					c = r.modAlpha(c, cover)
				}
				cx := x + j
				r.drawMagnifiedPixel(float64(cx), float64(y), c)
			}
		}
		if i < numSpans-1 {
			it.Next()
		}
	}
}

// drawMagnifiedPixel draws a single magnified pixel as a filled rectangle.
func (r *RendererEnlargedT[Ren, C]) drawMagnifiedPixel(x, y float64, c C) {
	magX1 := int(x * r.size)
	magY1 := int(y * r.size)
	magX2 := int((x + 1) * r.size)
	magY2 := int((y + 1) * r.size)
	r.baseRenderer.CopyBar(magX1, magY1, magX2-1, magY2-1, c)
}

// RGBA8CoverMod is a helper alpha modifier for agg color.RGBA8[CS] types.
func RGBA8CoverMod[CS ColorSpace](c aggcolor.RGBA8[CS], cover basics.Int8u) aggcolor.RGBA8[CS] {
	a := basics.Int8u((int(cover) * int(c.A)) >> 8)
	return aggcolor.RGBA8[CS]{R: c.R, G: c.G, B: c.B, A: a}
}
