// Port of AGG C++ alpha_mask.cpp – alpha-masked lion rendering.
//
// Generates a grayscale alpha mask from random ellipses, then renders the
// lion through it so only the mask's bright regions show the lion colours.
// Default: scale=1.0, angle=0.0 (no rotation).
package main

import (
	"math/rand"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	width  = 512
	height = 400
)

// Adapters to bridge internal rasterizer/scanline to renderer/scanline interfaces.

type rasterizerAdapter struct {
	ras interface {
		RewindScanlines() bool
		SweepScanline(sl rasterizer.ScanlineInterface) bool
		MinX() int
		MaxX() int
	}
}

func (r *rasterizerAdapter) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasterizerAdapter) MinX() int             { return r.ras.MinX() }
func (r *rasterizerAdapter) MaxX() int             { return r.ras.MaxX() }

type rasScanlineAdapterP8 struct{ sl *scanline.ScanlineP8 }

func (a *rasScanlineAdapterP8) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapterP8) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapterP8) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapterP8) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapterP8) NumSpans() int  { return a.sl.NumSpans() }

func (r *rasterizerAdapter) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*scanlineWrapperP8); ok {
		return r.ras.SweepScanline(&rasScanlineAdapterP8{sl: w.sl})
	}
	return false
}

type scanlineWrapperP8 struct{ sl *scanline.ScanlineP8 }

func (w *scanlineWrapperP8) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapperP8) Y() int               { return w.sl.Y() }
func (w *scanlineWrapperP8) NumSpans() int        { return w.sl.NumSpans() }

type spanIterP8 struct {
	spans []scanline.SpanP8
	idx   int
}

func (it *spanIterP8) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIterP8) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *scanlineWrapperP8) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIterP8{nil, 0}
	}
	return &spanIterP8{spans, 0}
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	img := ctx.GetImage()

	// --- Generate grayscale alpha mask ---
	maskData := make([]uint8, width*height)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	agg2d := ctx.GetAgg2D()
	ras := agg2d.GetInternalRasterizer()
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8{sl: sl}
	rasAdapter := &rasterizerAdapter{ras: ras}

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10; i++ {
		cx := float64(rng.Intn(width))
		cy := float64(rng.Intn(height))
		rx := float64(rng.Intn(100) + 20)
		ry := float64(rng.Intn(100) + 20)

		ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
		ras.Reset()
		ell.Rewind(0)
		for {
			var x, y float64
			cmd := ell.Vertex(&x, &y)
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		c := uint8(rng.Intn(200) + 55)
		gray := color.Gray8[color.Linear]{V: c, A: 255}
		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, maskRb, gray)
	}

	mask := pixfmt.NewAlphaMaskU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})

	// --- Draw checkered background (shows through where mask is transparent) ---
	for y := 0; y < height; y += 8 {
		for x := ((y >> 3) & 1) << 3; x < width; x += 16 {
			agg2d.FillColor(agg.NewColor(0xdf, 0xdf, 0xdf, 0xff))
			agg2d.Rectangle(float64(x), float64(y), float64(x+7), float64(y+7))
		}
	}
	agg2d.DrawPath(agg.FillOnly)

	// --- Render lion through alpha mask ---
	mainBuf := buffer.NewRenderingBufferWithData[uint8](img.Data, width, height, width*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainBuf)
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixf, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	lionPaths := liondemo.Parse()

	agg2d.ResetTransformations()
	agg2d.Translate(-250, -250)
	agg2d.Scale(1.0, 1.0)
	agg2d.Rotate(basics.Pi)
	agg2d.Translate(float64(width)/2, float64(height)/2)

	mtx := agg2d.GetTransformations()

	for _, lp := range lionPaths {
		c := color.RGBA8[color.Linear]{R: lp.Color[0], G: lp.Color[1], B: lp.Color[2], A: 255}

		ras.Reset()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			tx, ty := mtx.Transform(x, y)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			}
		}
		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rbAMask, c)
	}
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Alpha Mask", Width: width, Height: height}, &demo{})
}
