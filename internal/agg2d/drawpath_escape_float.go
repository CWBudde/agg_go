// Package agg2d float DrawPath escape-hatch methods (L5/breadth). Float twins of
// the 8-bit Agg2D advanced-rendering escape hatches (agg2d.go / rendering.go):
// direct access to the rasterizer plus custom solid / span-generator scanline
// rendering. RenderRasterizerWithColor lives in rendering_float.go; this file
// adds GetInternalRasterizer, ScanlineRender, and RenderScanlinesAAWithSpanGen.
//
// The rasterizer and scanline are color-agnostic and shared verbatim with the
// 8-bit twin; only the renderer/span-generator color type differs (RGBA32).
package agg2d

import (
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
)

// GetInternalRasterizer returns the underlying rasterizer for advanced usage
// (e.g. adding a custom vertex source then rendering via RenderRasterizerWithColor
// or ScanlineRender). Mirrors the 8-bit GetInternalRasterizer.
func (a *Agg2DFloat) GetInternalRasterizer() *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip] {
	return a.rasterizer
}

// ScanlineRender renders the given rasterizer data using a caller-supplied
// renderer, without touching the float context's own rasterizer state. Mirrors
// the 8-bit ScanlineRender.
func (a *Agg2DFloat) ScanlineRender(ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip], renderer renscan.RendererInterface[color.RGBA32[color.Linear]]) {
	if !ras.RewindScanlines() {
		return
	}

	sl := a.scanline
	sl.Reset(ras.MinX(), ras.MaxX())
	renderer.Prepare()

	for ras.SweepScanline(sl) {
		renderer.Render(sl)
	}
}

// RenderScanlinesAAWithSpanGen renders the rasterizer using a caller-supplied
// span generator, enabling advanced effects such as combining color gradients
// with alpha gradients. Mirrors the 8-bit RenderScanlinesAAWithSpanGen.
func (a *Agg2DFloat) RenderScanlinesAAWithSpanGen(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	spanGen renscan.SpanGeneratorInterface[color.RGBA32[color.Linear]],
) {
	renderer := a.currentRenderer()
	if renderer == nil || a.spanAllocator == nil {
		return
	}
	renscan.RenderScanlinesAA(ras, a.scanline, renderer, a.spanAllocator, spanGen)
}
