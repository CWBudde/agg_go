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
)

var (
	amAlphaMaskBuf *buffer.RenderingBuffer[uint8]
	amAlphaMask    *pixfmt.AlphaMaskU8
	amLionAngle    = 0.0
	amLionScale    = 1.0
	amLionSkewX    = 0.0
	amLionSkewY    = 0.0
)

func generateAlphaMask(w, h int) {
	if amAlphaMaskBuf == nil || amAlphaMaskBuf.Width() != w || amAlphaMaskBuf.Height() != h {
		data := make([]uint8, w*h)
		amAlphaMaskBuf = buffer.NewRenderingBufferWithData[uint8](data, w, h, w)
	}

	// Create a grayscale pixel format for the mask buffer
	maskPixf := pixfmt.NewPixFmtGray8(amAlphaMaskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)

	maskRb.Clear(color.Gray8[color.Linear]{V: 0, A: 255})

	agg2d := ctx.GetAgg2D()
	ras := agg2d.GetInternalRasterizer()
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8{sl: sl}
	rasAdapter := &rasterizerAdapter{ras: ras}

	for i := 0; i < 10; i++ {
		cx := float64(rand.Intn(w))
		cy := float64(rand.Intn(h))
		rx := float64(rand.Intn(100) + 20)
		ry := float64(rand.Intn(100) + 20)

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

		c := uint8(rand.Intn(256))
		opacity := uint8(rand.Intn(256))
		gray := color.Gray8[color.Linear]{V: c, A: opacity}

		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, maskRb, gray)
	}

	// Create the alpha mask
	maskFunc := pixfmt.OneComponentMaskU8{}
	amAlphaMask = pixfmt.NewAlphaMaskU8WithBuffer(amAlphaMaskBuf, 1, 0, maskFunc)
}

func drawAlphaMaskDemo() {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	if amAlphaMask == nil {
		generateAlphaMask(w, h)
	}

	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Fill background with white
	agg2d.ClearAll(agg.White)

	// Draw checkered background
	for y := 0; y < h; y += 8 {
		for x := ((y >> 3) & 1) << 3; x < w; x += 16 {
			agg2d.FillColor(agg.NewColor(0xdf, 0xdf, 0xdf, 0xff))
			agg2d.Rectangle(float64(x), float64(y), float64(x+7), float64(y+7))
		}
	}

	// Setup transformation for the lion
	baseDX, baseDY := 0.0, 0.0
	// Get bounding box for lion
	if len(lionPaths) > 0 {
		x1, y1, x2, y2 := 20.0, 20.0, 480.0, 380.0
		baseDX = (x2 - x1) * 0.5
		baseDY = (y2 - y1) * 0.5
	}

	agg2d.ResetTransformations()
	agg2d.Translate(-baseDX, -baseDY)
	agg2d.Scale(amLionScale, amLionScale)
	agg2d.Rotate(amLionAngle + basics.Pi)
	agg2d.Skew(amLionSkewX/1000.0, amLionSkewY/1000.0)
	agg2d.Translate(float64(w)/2, float64(h)/2)

	// In the Go port, Agg2D doesn't support masking directly yet.
	// We need to use the lower-level API with PixFmtAMaskAdaptor.

	// Get the image pixel format
	tempRbuf := buffer.NewRenderingBufferWithData[uint8](ctx.GetImage().Data, w, h, w*4)
	imgPixf := pixfmt.NewPixFmtRGBA32[color.Linear](tempRbuf)

	// Create alpha mask adaptor
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(imgPixf, amAlphaMask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	ras := agg2d.GetInternalRasterizer()
	rasAdapter := &rasterizerAdapter{ras: ras}
	sl := scanline.NewScanlineP8()
	slAdapter := &scanlineWrapperP8{sl: sl}

	for _, lp := range lionPaths {
		c := color.RGBA8[color.Linear]{R: lp.Color.R, G: lp.Color.G, B: lp.Color.B, A: 255}

		ras.Reset()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			// Apply Agg2D transformation manually since we are using low-level rasterizer
			tx, ty := x, y
			agg2d.GetTransformations().Transform(tx, ty)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			}
		}

		renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rbAMask, c)
	}

	logStatus(fmt.Sprintf("Alpha Mask Demo: Scale=%.2f, Angle=%.2f", amLionScale, amLionAngle))
}

func handleAlphaMaskMouseDown(x, y float64, flags int) bool {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	dx := x - float64(w)/2
	dy := y - float64(h)/2
	amLionAngle = math.Atan2(dy, dx)
	amLionScale = math.Sqrt(dy*dy+dx*dx) / 100.0
	return true
}

func handleAlphaMaskRightMouseDown(x, y float64) bool {
	amLionSkewX = x
	amLionSkewY = y
	return true
}
