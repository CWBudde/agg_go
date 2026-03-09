package main

import (
	"fmt"
	"math"
	"math/rand"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	am2AlphaMaskBuf *buffer.RenderingBuffer[uint8]
	am2AlphaMask    *pixfmt.AlphaMaskU8
	am2NumEllipses  = 10
	am2LionAngle    = 0.0
	am2LionScale    = 1.0
	am2LionSkewX    = 0.0
	am2LionSkewY    = 0.0
	am2SliderValue  = 10.0
)

func generateAlphaMask2(w, h int) {
	if am2AlphaMaskBuf == nil || am2AlphaMaskBuf.Width() != w || am2AlphaMaskBuf.Height() != h {
		data := make([]uint8, w*h)
		am2AlphaMaskBuf = buffer.NewRenderingBufferWithData[uint8](data, w, h, w)
	}

	// Create a grayscale pixel format for the mask buffer
	maskPixf := pixfmt.NewPixFmtGray8(am2AlphaMaskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)

	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	agg2d := ctx.GetAgg2D()
	ras := agg2d.GetInternalRasterizer()
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8{sl: sl}
	rasAdapter := &rasterizerAdapter{ras: ras}

	rnd := rand.New(rand.NewSource(1432))

	for i := 0; i < am2NumEllipses; i++ {
		cx := float64(rnd.Intn(w))
		cy := float64(rnd.Intn(h))
		rx := float64(rnd.Intn(100) + 20)
		ry := float64(rnd.Intn(100) + 20)

		ras.Reset()
		ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
		ell.Rewind(0)
		for {
			var x, y float64
			cmd := ell.Vertex(&x, &y)
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}

		v := uint8(rnd.Intn(128) + 128)
		a := uint8(rnd.Intn(128) + 128)
		gray := color.Gray8[color.Linear]{V: v, A: a}

		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, maskRb, gray)
	}

	// Create the alpha mask
	maskFunc := pixfmt.OneComponentMaskU8{}
	am2AlphaMask = pixfmt.NewAlphaMaskU8WithBuffer(am2AlphaMaskBuf, 1, 0, maskFunc)
}

func drawAlphaMask2Demo() {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()

	if float64(am2NumEllipses) != am2SliderValue {
		am2NumEllipses = int(am2SliderValue)
		generateAlphaMask2(w, h)
	}

	if am2AlphaMask == nil {
		generateAlphaMask2(w, h)
	}

	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Fill background with white
	agg2d.ClearAll(agg.White)

	// Set up the mask adaptor
	tempRbuf := buffer.NewRenderingBufferWithData[uint8](ctx.GetImage().Data, w, h, w*4)
	imgPixf := pixfmt.NewPixFmtRGBA32[color.Linear](tempRbuf)
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(imgPixf, am2AlphaMask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	ras := agg2d.GetInternalRasterizer()
	rasAdapter := &rasterizerAdapter{ras: ras}
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8{sl: sl}

	// 1. Render the lion
	baseDX, baseDY := 0.0, 0.0
	if len(lionPaths) > 0 {
		x1, y1, x2, y2 := 20.0, 20.0, 480.0, 380.0
		baseDX = (x2 - x1) * 0.5
		baseDY = (y2 - y1) * 0.5
	}

	mtx := transform.NewTransAffine()
	mtx.Translate(-baseDX, -baseDY)
	mtx.Scale(am2LionScale)
	mtx.Rotate(am2LionAngle + basics.Pi)
	mtx.Translate(float64(w)/2, float64(h)/2)

	for _, lp := range lionPaths {
		c := color.RGBA8[color.Linear]{R: lp.Color[0], G: lp.Color[1], B: lp.Color[2], A: 255}

		ras.Reset()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			tx, ty := x, y
			mtx.Transform(&tx, &ty)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			}
		}

		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rbAMask, c)
	}

	logStatus(fmt.Sprintf("Alpha Mask 2 Demo: Ellipses=%d", am2NumEllipses))
}

func handleAlphaMask2MouseDown(x, y float64, flags int) bool {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	dx := x - float64(w)/2
	dy := y - float64(h)/2
	am2LionAngle = math.Atan2(dy, dx)
	am2LionScale = math.Sqrt(dy*dy+dx*dx) / 100.0
	return true
}

func handleAlphaMask2RightMouseDown(x, y float64) bool {
	am2LionSkewX = x
	am2LionSkewY = y
	return true
}

func setAlphaMask2NumEllipses(n float64) {
	am2SliderValue = n
}
