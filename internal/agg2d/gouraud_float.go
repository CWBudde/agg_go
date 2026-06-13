// Package agg2d float Gouraud shading (L5/breadth). Float twin of the 8-bit
// Agg2D.GouraudTriangle (agg2d.go): renders a smoothly color-interpolated
// triangle through the float span_gouraud generator (span.SpanGouraudRGBA128).
// The triangle geometry / dilation / vertex-source plumbing is color-agnostic
// and reused from the shared SpanGouraud base; only the per-pixel color
// interpolation and blend are float (RGBA32).
package agg2d

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/span"
)

// GouraudTriangle renders a Gouraud-shaded (smoothly color-interpolated)
// triangle. c1/c2/c3 are the vertex colors; d is the dilation distance used for
// AGG's numerically stable beveled-polygon rasterization. Mirrors the 8-bit
// Agg2D.GouraudTriangle, using the float span_gouraud generator and the active
// (blend-mode-aware) float base renderer.
func (a *Agg2DFloat) GouraudTriangle(x1, y1, x2, y2, x3, y3 float64, c1, c2, c3 Color, d float64) {
	baseRen := a.currentRenderer()
	if baseRen == nil || a.rasterizer == nil || a.scanline == nil {
		return
	}
	a.rasterizer.Reset()

	gc1 := a.applyMasterAlpha(colorToRGBA32(c1))
	gc2 := a.applyMasterAlpha(colorToRGBA32(c2))
	gc3 := a.applyMasterAlpha(colorToRGBA32(c3))

	spanGen := span.NewSpanGouraudRGBA128WithTriangle(gc1, gc2, gc3, x1, y1, x2, y2, x3, y3, d)

	ren := &gouraudRendererFloat{
		ren:   baseRen.rendererBase(),
		span:  spanGen,
		alloc: span.NewSpanAllocator[color.RGBA32[color.Linear]](),
	}

	adapter := &gouraudRasAdapterFloat{sg: spanGen}
	a.rasterizer.AddPath(adapter, 0)
	a.scanlineRender(ren)
}

// gouraudRendererFloat blends float Gouraud spans into the float base renderer.
// It satisfies renscan.RendererInterface[color.RGBA32[color.Linear]].
type gouraudRendererFloat struct {
	ren   *renderer.RendererBase[renderer.PixelFormat[color.RGBA32[color.Linear]], color.RGBA32[color.Linear]]
	span  *span.SpanGouraudRGBA128
	alloc *span.SpanAllocator[color.RGBA32[color.Linear]]
}

func (r *gouraudRendererFloat) Prepare() { r.span.Prepare() }

func (r *gouraudRendererFloat) SetColor(c color.RGBA32[color.Linear]) {}

func (r *gouraudRendererFloat) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	it := sl.BeginIterator()
	for {
		spanData := it.GetSpan()
		x := spanData.X
		length := spanData.Len

		colors := r.alloc.Allocate(length)
		r.span.Generate(colors, x, y, uint(length))

		r.ren.BlendColorHspan(x, y, length, colors, spanData.Covers, 255)

		if !it.Next() {
			break
		}
	}
}

// gouraudRasAdapterFloat adapts the float span generator's vertex source to the
// rasterizer's path interface. Mirrors the 8-bit gouraudRasAdapter.
type gouraudRasAdapterFloat struct {
	sg interface {
		Rewind(uint)
		Vertex() (float64, float64, basics.PathCommand)
	}
}

func (a *gouraudRasAdapterFloat) Rewind(pathID uint32) {
	a.sg.Rewind(uint(pathID))
}

func (a *gouraudRasAdapterFloat) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.sg.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}
