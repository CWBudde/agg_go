// Port of AGG C++ alpha_mask3.cpp – alpha-mask as polygon clipper.
//
// Renders case 3 (default): Great Britain polygon as alpha mask, spiral stroke
// as content rendered through the mask.  The C++ original has radio buttons to
// select different polygon/operation combos — only the default is ported here.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/demo/aggshapes"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 640
	frameHeight = 520
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

func (r *rasterizerAdaptor) AddPath(vs rasterizer.VertexSource, pathID uint32) {
	r.ras.AddPath(vs, pathID)
}

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
func (w *scanlineWrapper) NumSpans() int        { return w.sl.NumSpans() }

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
// Spiral vertex source – matches C++ spiral class from alpha_mask3.cpp
// ---------------------------------------------------------------------------

type spiral struct {
	x, y         float64
	r1, r2       float64
	step         float64
	startAngle   float64
	angle, currR float64
	da, dr       float64
	start        bool
}

func newSpiral(x, y, r1, r2, step, startAngle float64) *spiral {
	return &spiral{
		x:          x,
		y:          y,
		r1:         r1,
		r2:         r2,
		step:       step,
		startAngle: startAngle,
		da:         4.0 * basics.Deg2Rad,
		dr:         step / 90.0,
	}
}

func (s *spiral) Rewind(_ uint) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

func (s *spiral) Vertex() (float64, float64, basics.PathCommand) {
	if s.currR > s.r2 {
		return 0, 0, basics.PathCmdStop
	}
	x := s.x + math.Cos(s.angle)*s.currR
	y := s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da
	if s.start {
		s.start = false
		return x, y, basics.PathCmdMoveTo
	}
	return x, y, basics.PathCmdLineTo
}

// ---------------------------------------------------------------------------
// Adapter: conv.VertexSource wrapping a path + affine transform
// ---------------------------------------------------------------------------

type transformedPathVS struct {
	ps  *path.PathStorageStl
	mtx *transform.TransAffine
}

func (t *transformedPathVS) Rewind(id uint) { t.ps.Rewind(id) }
func (t *transformedPathVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := t.ps.NextVertex()
	t.mtx.Transform(&x, &y)
	return x, y, basics.PathCommand(cmd)
}

// rasterVS wraps a conv.VertexSource as a rasterizer.VertexSource.
type rasterVS struct{ src conv.VertexSource }

func (a *rasterVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *rasterVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	mx, my float64 // spiral centre
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)

	ras := newRasterizer()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}

	// White background.
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// --- Case 3: Great Britain and Spiral ---

	// Build transformed GB polygon.
	psGB := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(psGB)
	gbMtx := transform.NewTransAffine()
	gbMtx.Translate(-1150, -1150)
	gbMtx.Scale(2.0)
	transGB := &transformedPathVS{ps: psGB, mtx: gbMtx}

	// Draw GB fill (faint background).
	ras.Reset()
	ras.AddPath(&rasterVS{src: transGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 127, G: 127, B: 0, A: 25})

	// Draw GB stroke.
	strokeGB := conv.NewConvStroke(transGB)
	strokeGB.SetWidth(0.1)
	ras.Reset()
	ras.AddPath(&rasterVS{src: strokeGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Draw spiral stroke (faint preview).
	sp := newSpiral(d.mx, d.my, 10, 150, 30, 0.0)
	strokeSp := conv.NewConvStroke(sp)
	strokeSp.SetWidth(15.0)
	ras.Reset()
	ras.AddPath(&rasterVS{src: strokeSp}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 127, B: 127, A: 25})

	// --- Generate alpha mask from GB polygon ---
	maskData := make([]uint8, w*h)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, w, h, w)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	// AND operation: clear black, render white.
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	ras.Reset()
	ras.AddPath(&rasterVS{src: transGB}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, maskRb,
		color.Gray8[color.Linear]{V: 255, A: 255})

	mask := pixfmt.NewAlphaMaskU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})

	// --- Render spiral stroke through mask ---
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixf, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	// Re-rasterize spiral stroke for masked rendering.
	sp2 := newSpiral(d.mx, d.my, 10, 150, 30, 0.0)
	strokeSp2 := conv.NewConvStroke(sp2)
	strokeSp2.SetWidth(15.0)
	ras.Reset()
	ras.AddPath(&rasterVS{src: strokeSp2}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rbAMask,
		color.RGBA8[color.Linear]{R: 127, G: 0, B: 0, A: 127})

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
		mx: float64(frameWidth) / 2,
		my: float64(frameHeight) / 2,
	}
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Alpha Mask3",
		Width:  frameWidth,
		Height: frameHeight,
	}, d)
}
