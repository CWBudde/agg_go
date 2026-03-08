// Port of AGG C++ alpha_mask2.cpp – alpha-masked lion with affine-transformed mask.
//
// Like alpha_mask but the ellipses are drawn with an affine transform applied
// so the mask rotates independently from the lion. Default: 10 ellipses, scale=1.
package main

import (
	"math/rand"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
	"agg_go/internal/shapes"
	"agg_go/internal/transform"
)

const (
	width  = 800
	height = 600

	numEllipses = 10
)

type rasterizerAdapter2 struct {
	ras interface {
		RewindScanlines() bool
		SweepScanline(sl rasterizer.ScanlineInterface) bool
		MinX() int
		MaxX() int
	}
}

func (r *rasterizerAdapter2) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *rasterizerAdapter2) MinX() int             { return r.ras.MinX() }
func (r *rasterizerAdapter2) MaxX() int             { return r.ras.MaxX() }

type rasScanlineAdapterP8v2 struct{ sl *scanline.ScanlineP8 }

func (a *rasScanlineAdapterP8v2) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapterP8v2) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapterP8v2) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapterP8v2) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapterP8v2) NumSpans() int  { return a.sl.NumSpans() }

func (r *rasterizerAdapter2) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*scanlineWrapperP8v2); ok {
		return r.ras.SweepScanline(&rasScanlineAdapterP8v2{sl: w.sl})
	}
	return false
}

type scanlineWrapperP8v2 struct{ sl *scanline.ScanlineP8 }

func (w *scanlineWrapperP8v2) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapperP8v2) Y() int               { return w.sl.Y() }
func (w *scanlineWrapperP8v2) NumSpans() int        { return w.sl.NumSpans() }

type spanIterP8v2 struct {
	spans []scanline.SpanP8
	idx   int
}

func (it *spanIterP8v2) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIterP8v2) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *scanlineWrapperP8v2) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIterP8v2{nil, 0}
	}
	return &spanIterP8v2{spans, 0}
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	img := ctx.GetImage()
	agg2d := ctx.GetAgg2D()
	ras := agg2d.GetInternalRasterizer()
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8v2{sl: sl}
	rasAdapter := &rasterizerAdapter2{ras: ras}

	// Generate alpha mask.
	maskData := make([]uint8, width*height)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	maskPixf := pixfmt.NewPixFmtGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	rng := rand.New(rand.NewSource(1432))
	// Affine transform for the mask ellipses (slight rotation).
	maskMtx := transform.NewTransAffine()
	maskMtx.Rotate(0.3)
	maskMtx.Translate(float64(width)/2, float64(height)/2)

	for i := 0; i < numEllipses; i++ {
		cx := float64(rng.Intn(width)) - float64(width)/2
		cy := float64(rng.Intn(height)) - float64(height)/2
		rx := float64(rng.Intn(100) + 20)
		ry := float64(rng.Intn(100) + 20)

		// Transform centre.
		tx, ty := cx, cy
		maskMtx.Transform(&tx, &ty)

		ell := shapes.NewEllipseWithParams(tx, ty, rx, ry, 100, false)
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

	// Checkered background.
	agg2d.ResetTransformations()
	for y := 0; y < height; y += 8 {
		for x := ((y >> 3) & 1) << 3; x < width; x += 16 {
			agg2d.FillColor(agg.NewColor(0xdf, 0xdf, 0xdf, 0xff))
			agg2d.Rectangle(float64(x), float64(y), float64(x+7), float64(y+7))
		}
	}
	agg2d.DrawPath(agg.FillOnly)

	// Render lion through mask.
	mainBuf := buffer.NewRenderingBufferWithData[uint8](img.Data, width, height, width*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainBuf)
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(mainPixf, mask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	lionPaths := liondemo.Parse()
	agg2d.ResetTransformations()
	agg2d.Translate(-250, -250)
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
	demorunner.Run(demorunner.Config{Title: "Alpha Mask 2", Width: width, Height: height}, &demo{})
}
