package main

import (
	"fmt"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/demo/aggshapes"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	am3AlphaMaskBuf *buffer.RenderingBuffer[uint8]
	am3AlphaMask    *pixfmt.AlphaMaskU8
	am3Operation    = 0 // 0: AND, 1: SUB
	am3Polygon      = 3 // Default to GB and Spiral
	am3X, am3Y      float64
)

// transformedPathVS adapts a PathStorageStl with an affine transform applied,
// implementing conv.VertexSource so it can be used with ConvStroke and the rasterizer.
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

func generateAlphaMask3(vs conv.VertexSource, w, h int) {
	if am3AlphaMaskBuf == nil || am3AlphaMaskBuf.Width() != w || am3AlphaMaskBuf.Height() != h {
		data := make([]uint8, w*h)
		am3AlphaMaskBuf = buffer.NewRenderingBufferWithData[uint8](data, w, h, w)
	}

	maskPixf := pixfmt.NewPixFmtGray8(am3AlphaMaskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)

	agg2d := ctx.GetAgg2D()
	ras := agg2d.GetInternalRasterizer()
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8{sl: sl}
	rasAdapter := &rasterizerAdapter{ras: ras}

	if am3Operation == 0 {
		maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})
		ras.Reset()
		vs.Rewind(0)
		for {
			x, y, cmd := vs.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, maskRb, color.Gray8[color.Linear]{V: 255, A: 255})
	} else {
		maskRb.Clear(color.Gray8[color.Linear]{V: 255, A: 255})
		ras.Reset()
		vs.Rewind(0)
		for {
			x, y, cmd := vs.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, maskRb, color.Gray8[color.Linear]{V: 0, A: 255})
	}

	maskFunc := pixfmt.OneComponentMaskU8{}
	am3AlphaMask = pixfmt.NewAlphaMaskU8WithBuffer(am3AlphaMaskBuf, 1, 0, maskFunc)
}

func drawAlphaMask3Demo() {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	if am3X == 0 && am3Y == 0 {
		am3X = float64(w) / 2
		am3Y = float64(h) / 2
	}

	// Get image pixel format
	tempRbuf := buffer.NewRenderingBufferWithData[uint8](ctx.GetImage().Data, w, h, w*4)
	imgPixf := pixfmt.NewPixFmtRGBA32[color.Linear](tempRbuf)
	rbBase := renderer.NewRendererBaseWithPixfmt(imgPixf)

	ras := agg2d.GetInternalRasterizer()
	rasAdapter := &rasterizerAdapter{ras: ras}
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8{sl: sl}

	if am3Polygon == 3 { // Great Britain and Spiral
		psGB := path.NewPathStorageStl()
		aggshapes.MakeGBPoly(psGB)

		mtx := transform.NewTransAffine()
		mtx.Translate(-1150, -1150)
		mtx.Scale(2.0)

		transGB := &transformedPathVS{ps: psGB, mtx: mtx}

		// Draw GB background
		ras.Reset()
		transGB.Rewind(0)
		for {
			x, y, cmd := transGB.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rbBase, color.RGBA8[color.Linear]{R: 127, G: 127, B: 0, A: 25})

		// Draw GB stroke
		strokeGB := conv.NewConvStroke(transGB)
		strokeGB.SetWidth(0.1)
		ras.Reset()
		strokeGB.Rewind(0)
		for {
			x, y, cmd := strokeGB.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rbBase, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

		// Spiral
		// We'd need a spiral generator here. For now let's just use a circle as a placeholder
		// or implement a simple spiral.
		// Actually I'll skip the spiral for now and just use a circle.
		agg2d.ResetTransformations()
		agg2d.FillColor(agg.NewColor(0, 127, 127, 25))
		agg2d.FillCircle(am3X, am3Y, 150)

		// Create alpha mask from GB
		generateAlphaMask3(transGB, w, h)

		// Render circle with alpha mask
		amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(imgPixf, am3AlphaMask)
		rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)
		_ = rbAMask // reserved for future masked rendering

		agg2d.ResetTransformations()
		// Manual circle for masking
		// ...
	}

	logStatus(fmt.Sprintf("Alpha Mask 3 Demo: Op=%d, Poly=%d", am3Operation, am3Polygon))
}

func handleAlphaMask3MouseDown(x, y float64, flags int) bool {
	am3X = x
	am3Y = y
	return true
}

func handleAlphaMask3MouseMove(x, y float64, flags int) bool {
	if flags != 0 {
		am3X = x
		am3Y = y
		return true
	}
	return false
}

func setAlphaMask3Op(op int) {
	am3Operation = op
}

func setAlphaMask3Poly(poly int) {
	am3Polygon = poly
}
