// Port of AGG C++ bezier_div.cpp – Bezier curve subdivision accuracy demo.
//
// Shows a cubic Bezier curve rendered as a wide stroked shape together with
// the subdivision points. Default values from the WASM demo are used as
// constants; interactive sliders belong in the platform (SDL2/X11) variant.
//
// Default: subdivision mode, control points (170,424)(13,87)(488,423)(26,333),
// angle tolerance=15°, approx scale=1.0, stroke width=50.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/curves"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	width  = 655
	height = 520

	cx1, cy1 = 170.0, 424.0
	cx2, cy2 = 13.0, 87.0
	cx3, cy3 = 488.0, 423.0
	cx4, cy4 = 26.0, 333.0

	defaultAngleTol    = 15.0 // degrees
	defaultApproxScale = 1.0
	defaultCuspLimit   = 0.0 // degrees
	defaultWidth       = 50.0
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
// Vertex-source adapters
// ---------------------------------------------------------------------------

// convVS adapts conv.VertexSource to rasterizer.VertexSource.
type convVS struct{ src conv.VertexSource }

func (a *convVS) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ellipseVS adapts shapes.Ellipse to rasterizer.VertexSource.
type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	var vx, vy float64
	cmd := ev.e.Vertex(&vx, &vy)
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](workRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)

	// Light cream background.
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	ras := newRasterizer()
	sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}

	angleTol := defaultAngleTol * math.Pi / 180.0
	cuspLimit := defaultCuspLimit * math.Pi / 180.0

	// Build the curve using subdivision.
	curve := curves.NewCurve4Div()
	curve.SetApproximationScale(defaultApproxScale)
	curve.SetAngleTolerance(angleTol)
	curve.SetCuspLimit(cuspLimit)
	curve.Init(cx1, cy1, cx2, cy2, cx3, cy3, cx4, cy4)

	// Collect subdivision points.
	curvePath := path.NewPathStorageStl()
	curve.Rewind(0)
	for {
		x, y, cmd := curve.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			curvePath.MoveTo(x, y)
		} else if basics.IsVertex(cmd) {
			curvePath.LineTo(x, y)
		}
	}

	// Wide stroke from the curve.
	curveAdapter := path.NewPathStorageStlVertexSourceAdapter(curvePath)
	stroke := conv.NewConvStroke(curveAdapter)
	stroke.SetWidth(defaultWidth)
	stroke.SetLineJoin(basics.MiterJoin)
	stroke.SetLineCap(basics.ButtCap)

	// Fill the wide stroke (semi-transparent green).
	ras.Reset()
	ras.AddPath(&convVS{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 128, B: 0, A: 128})

	// Outline of the wide stroke (stroke of a stroke).
	stroke2 := conv.NewConvStroke(stroke)
	stroke2.SetWidth(1.5)
	ras.Reset()
	ras.AddPath(&convVS{src: stroke2}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128})

	// Subdivision points as small dots.
	dotColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 153}
	curvePath.Rewind(0)
	for {
		x, y, cmd := curvePath.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsVertex(basics.PathCommand(cmd)) {
			dot := shapes.NewEllipseWithParams(x, y, 1.5, 1.5, 8, false)
			ras.Reset()
			ras.AddPath(&ellipseVS{e: dot}, 0)
			renscan.RenderScanlinesAASolid(ras, sl, mainRb, dotColor)
		}
	}

	// Control polygon.
	ctrlPs := path.NewPathStorageStl()
	ctrlPs.MoveTo(cx1, cy1)
	ctrlPs.LineTo(cx2, cy2)
	ctrlPs.LineTo(cx3, cy3)
	ctrlPs.LineTo(cx4, cy4)
	ctrlStroke := conv.NewConvStroke(path.NewPathStorageStlVertexSourceAdapter(ctrlPs))
	ctrlStroke.SetWidth(1.0)
	ras.Reset()
	ras.AddPath(&convVS{src: ctrlStroke}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, mainRb,
		color.RGBA8[color.Linear]{R: 80, G: 80, B: 200, A: 180})

	// Control point dots.
	ctrlPtColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 200, A: 200}
	for _, pt := range [][2]float64{{cx1, cy1}, {cx2, cy2}, {cx3, cy3}, {cx4, cy4}} {
		dot := shapes.NewEllipseWithParams(pt[0], pt[1], 5, 5, 20, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: dot}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, mainRb, ctrlPtColor)
	}

	// Copy with y-flip (C++ uses flip_y=true).
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
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Bezier Div",
		Width:  width,
		Height: height,
	}, &demo{})
}
