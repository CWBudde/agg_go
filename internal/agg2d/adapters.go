// Package agg2d adapters for AGG2D high-level interface.
// This file contains adapter types that bridge different interfaces.
package agg2d

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/path"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
)

// pathVertexSourceAdapter adapts path.PathStorageStl to vertex source interface
type pathVertexSourceAdapter struct {
	path *path.PathStorageStl
}

func (p *pathVertexSourceAdapter) Rewind(pathID uint) {
	p.path.Rewind(pathID)
}

// Simplified stub - proper implementation needs interface matching
func (p *pathVertexSourceAdapter) Vertex() (float64, float64, basics.PathCommand) {
	// This is a simplified stub
	return 0, 0, 0 // Using 0 as stop command for now
}

// pixFmtAdapter adapts pixfmt to renderer interfaces
type pixFmtAdapter[C any] struct {
	pf renderer.PixelFormat[C]
}

func (p *pixFmtAdapter[C]) Width() int  { return p.pf.Width() }
func (p *pixFmtAdapter[C]) Height() int { return p.pf.Height() }

// RowPtr method doesn't exist in PixelFormat interface, simplified for now
func (p *pixFmtAdapter[C]) RowPtr(y int) []uint8 { return nil }

func (p *pixFmtAdapter[C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
	p.pf.BlendPixel(x, y, c, cover)
}

func (p *pixFmtAdapter[C]) CopyPixel(x, y int, c C) {
	p.pf.CopyPixel(x, y, c)
}

func (p *pixFmtAdapter[C]) BlendHline(x, y, length int, c C, cover basics.Int8u) {
	p.pf.BlendHline(x, y, length, c, cover)
}

// baseRendererAdapter adapts renderer base functionality
type baseRendererAdapter[C any] struct {
	pf renderer.PixelFormat[C]
}

func (b *baseRendererAdapter[C]) Width() int  { return b.pf.Width() }
func (b *baseRendererAdapter[C]) Height() int { return b.pf.Height() }

func (b *baseRendererAdapter[C]) Clear(c C) {
	b.pf.Clear(c)
}

func (b *baseRendererAdapter[C]) CopyPixel(x, y int, c C) {
	b.pf.CopyPixel(x, y, c)
}

func (b *baseRendererAdapter[C]) BlendPixel(x, y int, c C, cover basics.Int8u) {
	b.pf.BlendPixel(x, y, c, cover)
}

func (b *baseRendererAdapter[C]) BlendHline(x, y, length int, c C, cover basics.Int8u) {
	b.pf.BlendHline(x, y, length, c, cover)
}

func (b *baseRendererAdapter[C]) BlendColorHspan(x, y, length int, colors []C, covers []basics.Int8u, cover basics.Int8u) {
	b.pf.BlendColorHspan(x, y, length, colors, covers, cover)
}

func (b *baseRendererAdapter[C]) BlendSolidHspan(x, y, length int, c C, covers []basics.Int8u) {
	b.pf.BlendSolidHspan(x, y, length, c, covers)
}

// BlendFrom blends from another pixel format using the rendering pipeline
func (b *baseRendererAdapter[C]) BlendFrom(src renderer.PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int, cover basics.Int8u) {
	// Create a renderer base to access the BlendFrom functionality
	rendererBase := renderer.NewRendererBaseWithPixfmt(b.pf)
	rendererBase.BlendFrom(src, rectSrcPtr, dx, dy, cover)
}

// CopyFrom copies from another pixel format using the rendering pipeline
func (b *baseRendererAdapter[C]) CopyFrom(src renderer.PixelFormat[C], rectSrcPtr *basics.RectI, dx, dy int) {
	// Create a renderer base to access the CopyFrom functionality
	rendererBase := renderer.NewRendererBaseWithPixfmt(b.pf)
	rendererBase.CopyFrom(src, rectSrcPtr, dx, dy)
}

// scanlineWrapper adapts internal/scanline.ScanlineU8 to renderer/scanline.ScanlineInterface
type scanlineWrapper struct{ sl *scanline.ScanlineU8 }

// Reset implements ResettableScanline
func (w *scanlineWrapper) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapper) Y() int               { return w.sl.Y() }
func (w *scanlineWrapper) NumSpans() int        { return w.sl.NumSpans() }

// spanIter implements renderer/scanline.ScanlineIterator over our scanline spans
type spanIter struct {
	spans []scanline.Span
	idx   int
}

func (it *spanIter) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIter) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *scanlineWrapper) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIter{spans: nil, idx: 0}
	}
	return &spanIter{spans: spans, idx: 0}
}

// rasterizerAdapter adapts internal rasterizer to renderer/scanline.RasterizerInterface
type rasterizerAdapter struct {
	ras *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl]
}

func (r rasterizerAdapter) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r rasterizerAdapter) MinX() int             { return r.ras.MinX() }
func (r rasterizerAdapter) MaxX() int             { return r.ras.MaxX() }

// rasScanlineAdapter adapts scanline.ScanlineU8 to rasterizer.ScanlineInterface
type rasScanlineAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

func (r rasterizerAdapter) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*scanlineWrapper); ok {
		return r.ras.SweepScanline(&rasScanlineAdapter{sl: w.sl})
	}
	return false
}

// imagePixelFormat adapts an Image to the renderer.PixelFormat interface
// This allows images to be used with the rendering pipeline's BlendFrom and CopyFrom methods
type imagePixelFormat struct {
	img *Image
}

func newImagePixelFormat(img *Image) *imagePixelFormat {
	return &imagePixelFormat{img: img}
}

func (ipf *imagePixelFormat) Width() int {
	return ipf.img.Width()
}

func (ipf *imagePixelFormat) Height() int {
	return ipf.img.Height()
}

func (ipf *imagePixelFormat) PixWidth() int {
	return 4 // 4 bytes per pixel for RGBA
}

// Pixel returns a pixel as RGBA8 color
func (ipf *imagePixelFormat) Pixel(x, y int) color.RGBA8[color.Linear] {
	pixel := ipf.img.GetPixel(x, y)
	return color.NewRGBA8[color.Linear](pixel[0], pixel[1], pixel[2], pixel[3])
}

// CopyPixel - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	// Not implemented for source images
}

// BlendPixel - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	// Not implemented for source images
}

// CopyHline - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	// Not implemented for source images
}

// BlendHline - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	// Not implemented for source images
}

// CopyVline - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	// Not implemented for source images
}

// BlendVline - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	// Not implemented for source images
}

// CopyColorHspan - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) CopyColorHspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	// Not implemented for source images
}

// CopyColorVspan - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) CopyColorVspan(x, y, length int, colors []color.RGBA8[color.Linear]) {
	// Not implemented for source images
}

// BlendColorHspan - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendColorHspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	// Not implemented for source images
}

// BlendColorVspan - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendColorVspan(x, y, length int, colors []color.RGBA8[color.Linear], covers []basics.Int8u, cover basics.Int8u) {
	// Not implemented for source images
}

// Clear - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) Clear(c color.RGBA8[color.Linear]) {
	// Not implemented for source images
}

// CopyBar - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear]) {
	// Not implemented for source images
}

// BlendBar - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	// Not implemented for source images
}

// BlendSolidHspan - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	// Not implemented for source images
}

// BlendSolidVspan - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	// Not implemented for source images
}

// Fill - not needed for source image operations, but required by interface
func (ipf *imagePixelFormat) Fill(c color.RGBA8[color.Linear]) {
	// Not implemented for source images
}
