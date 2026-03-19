// Port of AGG C++ aa_demo.cpp – anti-aliasing demonstration.
//
// Renders a triangle using the enlarged-pixel technique: each logical pixel
// in the rasterized triangle is drawn as a large square coloured by its AA
// coverage value, making the anti-aliasing algorithm visible.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"

	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	frameWidth  = 600
	frameHeight = 400
)

// ---------------------------------------------------------------------------
// Rasterizer / scanline adapters
// ---------------------------------------------------------------------------

type rasterizerAdaptor struct {
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	sl  rasScanlineAdaptor
}

func newRasterizer() *rasterizerAdaptor {
	return &rasterizerAdaptor{
		ras: rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
			rasterizer.RasConvInt{},
			rasterizer.NewRasterizerSlNoClip(),
		),
		sl: rasScanlineAdaptor{sl: scanline.NewScanlineP8()},
	}
}

func (r *rasterizerAdaptor) Reset()                { r.ras.Reset() }
func (r *rasterizerAdaptor) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasterizerAdaptor) MinX() int             { return r.ras.MinX() }
func (r *rasterizerAdaptor) MaxX() int             { return r.ras.MaxX() }

func (r *rasterizerAdaptor) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*scanlineWrapper); ok {
		r.sl.sl = w.sl
		return r.ras.SweepScanline(&r.sl)
	}
	return false
}

type rasScanlineAdaptor struct{ sl *scanline.ScanlineP8 }

func (a *rasScanlineAdaptor) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdaptor) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdaptor) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdaptor) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdaptor) NumSpans() int  { return a.sl.NumSpans() }

type scanlineWrapper struct{ sl *scanline.ScanlineP8 }

func (w *scanlineWrapper) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapper) Y() int               { return w.sl.Y() }
func (w *scanlineWrapper) NumSpans() int         { return w.sl.NumSpans() }

func (w *scanlineWrapper) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIter{nil, 0}
	}
	return &spanIter{spans, 0}
}

type spanIter struct {
	spans []scanline.SpanP8
	idx   int
}

func (it *spanIter) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIter) Next() bool { it.idx++; return it.idx < len(it.spans) }

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

type pathSourceAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathSourceAdapter) Rewind(pathID uint32) { a.ps.Rewind(uint(pathID)) }
func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
}

type convVS struct{ src conv.VertexSource }

func (a *convVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// pathStlVS wraps PathStorageStl as a conv.VertexSource.
type pathStlVS struct{ ps *path.PathStorageStl }

func (a *pathStlVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathStlVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := a.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

// ---------------------------------------------------------------------------
// rendererEnlarged – C++ renderer_enlarged: draws each scanline pixel as
// a large square via its own rasterizer, matching the C++ original exactly.
// ---------------------------------------------------------------------------

type rendererEnlarged struct {
	ras   *rasterizerAdaptor
	sl    *scanlineWrapper
	renRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]
	size  float64
	col   color.RGBA8[color.Linear]
}

func newRendererEnlarged(
	renRb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	size float64,
) *rendererEnlarged {
	return &rendererEnlarged{
		ras:   newRasterizer(),
		sl:    &scanlineWrapper{sl: scanline.NewScanlineP8()},
		renRb: renRb,
		size:  size,
	}
}

func (r *rendererEnlarged) Prepare() {}

func (r *rendererEnlarged) SetColor(c color.RGBA8[color.Linear]) { r.col = c }

func (r *rendererEnlarged) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	it := sl.Begin()
	for i, n := 0, sl.NumSpans(); i < n; i++ {
		span := it.GetSpan()
		x := span.X
		numPix := span.Len
		covers := span.Covers
		solid := numPix < 0
		if solid {
			numPix = -numPix
		}
		for j := 0; j < numPix; j++ {
			cover := covers[0]
			if !solid {
				cover = covers[j]
			}
			a := (uint16(cover) * uint16(r.col.A)) >> 8
			r.drawSquare(float64(x+j), float64(y),
				color.RGBA8[color.Linear]{R: r.col.R, G: r.col.G, B: r.col.B, A: uint8(a)})
		}
		if i < n-1 {
			it.Next()
		}
	}
}

func (r *rendererEnlarged) drawSquare(x, y float64, c color.RGBA8[color.Linear]) {
	r.ras.ras.Reset()
	r.ras.ras.MoveToD(x*r.size, y*r.size)
	r.ras.ras.LineToD(x*r.size+r.size, y*r.size)
	r.ras.ras.LineToD(x*r.size+r.size, y*r.size+r.size)
	r.ras.ras.LineToD(x*r.size, y*r.size+r.size)
	renscan.RenderScanlinesAASolid(r.ras, r.sl, r.renRb, c)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	x, y [3]float64
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	sizeMul := 32.0 // C++ default slider value

	ras := newRasterizer()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}

	// 1. Enlarged-pixel rendering.
	renEnlarged := newRendererEnlarged(mainRb, sizeMul)
	renEnlarged.SetColor(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras.ras.Reset()
	ras.ras.MoveToD(d.x[0]/sizeMul, d.y[0]/sizeMul)
	ras.ras.LineToD(d.x[1]/sizeMul, d.y[1]/sizeMul)
	ras.ras.LineToD(d.x[2]/sizeMul, d.y[2]/sizeMul)
	renscan.RenderScanlines(ras, sl, renEnlarged)

	// 2. Actual-size solid black fill at scaled coordinates.
	ras.ras.Reset()
	ras.ras.MoveToD(d.x[0]/sizeMul, d.y[0]/sizeMul)
	ras.ras.LineToD(d.x[1]/sizeMul, d.y[1]/sizeMul)
	ras.ras.LineToD(d.x[2]/sizeMul, d.y[2]/sizeMul)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// 3. Full-scale triangle outline in teal via conv_stroke.
	teal := color.RGBA8[color.Linear]{R: 0, G: 150, B: 160, A: 200}
	edges := [3][2]int{{0, 1}, {1, 2}, {2, 0}}
	for _, e := range edges {
		ps := path.NewPathStorageStl()
		ps.MoveTo(d.x[e[0]], d.y[e[0]])
		ps.LineTo(d.x[e[1]], d.y[e[1]])
		stroke := conv.NewConvStroke(&pathStlVS{ps: ps})
		stroke.SetWidth(2.0)
		ras.Reset()
		ras.ras.AddPath(&convVS{src: stroke}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, teal)
	}

	// Copy with y-flip.
	copyFlipY(workBuf, img.Data, w, h)
}

func copyFlipY(src, dst []uint8, width, height int) {
	stride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func main() {
	d := &demo{
		x: [3]float64{57, 369, 143},
		y: [3]float64{100, 170, 310},
	}
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AA Demo",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
