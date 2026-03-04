package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
	"agg_go/internal/shapes"
)

var (
	compAlphaSrc = 0.75
	compAlphaDst = 1.0
	compOp       = blender.CompOpSrcOver
	compImage    *agg.Image
)

func drawCompositingDemo() {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()

	if compImage == nil {
		compImage = createTestImage(200, 200)
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	// Draw checkered background
	for y := 0; y < h; y += 8 {
		for x := ((y >> 3) & 1) << 3; x < w; x += 16 {
			agg2d.FillColor(agg.NewColor(0xdf, 0xdf, 0xdf, 0xff))
			agg2d.Rectangle(float64(x), float64(y), float64(x+7), float64(y+7))
		}
	}

	// 1. Draw Destination Shape (Yellow circle)
	// We draw it directly to the context using normal alpha blending first
	agg2d.ResetTransformations()

	// Create a temporary buffer for compositing
	tempBuf := make([]uint8, w*h*4)
	tempRbuf := buffer.NewRenderingBufferWithData[uint8](tempBuf, w, h, w*4)

	// We need to use RGBA32 (premultiplied) for compositing
	pixf := pixfmt.NewPixFmtRGBA32[color.Linear](tempRbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pixf)
	rb.Clear(color.RGBA8[color.Linear]{0, 0, 0, 0})

	// Draw destination image from the test image
	srcPixf := pixfmt.NewPixFmtRGBA32[color.Linear](buffer.NewRenderingBufferWithData[uint8](compImage.Data, 200, 200, 200*4))
	rb.BlendFrom(srcPixf, &basics.RectI{X1: 0, Y1: 0, X2: 200, Y2: 200}, 0, 250, basics.Int8u(compAlphaDst*255))

	// Draw destination circle
	drawCircleComp(rb,
		color.RGBA8[color.Linear]{0xFD, 0xF0, 0x6F, uint8(compAlphaDst * 255)},
		color.RGBA8[color.Linear]{0xFE, 0x9F, 0x34, uint8(compAlphaDst * 255)},
		70*3, 100+24*3, 37*3, 100+79*3)

	// 2. Draw Source Shape (Blue rounded rect) with Compositing Op
	// Create a custom blender for the compositing op
	compBlender := blender.NewCompositeBlender[color.Linear, order.RGBA](compOp)
	compPixf := pixfmt.NewPixFmtAlphaBlendRGBA[color.Linear, blender.CompositeBlender[color.Linear, order.RGBA]](tempRbuf, compBlender)
	compRb := renderer.NewRendererBaseWithPixfmt(compPixf)

	drawSourceShapeComp(compRb,
		color.RGBA8[color.Linear]{0x7F, 0xC1, 0xFF, uint8(compAlphaSrc * 255)},
		color.RGBA8[color.Linear]{0x05, 0x00, 0x5F, uint8(compAlphaSrc * 255)},
		300+50, 100+24*3, 107+50, 100+79*3)

	// Final step: blend the temp buffer back to the main context
	mainRbuf := buffer.NewRenderingBufferWithData[uint8](ctx.GetImage().Data, w, h, w*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.BlendFrom(pixf, nil, 0, 0, 255)

	logStatus(fmt.Sprintf("Compositing Demo: Op=%d, AlphaSrc=%.2f, AlphaDst=%.2f", compOp, compAlphaSrc, compAlphaDst))
}

func drawCircleComp(rb renscan.BaseRendererInterface[color.RGBA8[color.Linear]], c1, c2 color.RGBA8[color.Linear], x1, y1, x2, y2 float64) {
	r := math.Hypot(x2-x1, y2-y1) / 2
	cx, cy := (x1+x2)/2, (y1+y2)/2

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip())
	rasAdapter := &rasterizerAdapter{ras: ras}
	sl := scanline.NewScanlineU8()
	slAdapter := &scanlineWrapperU8{sl: sl}

	// Shadow
	circle := shapes.NewEllipseWithParams(cx+5, cy-3, r, r, 0, false)

	circle.Rewind(0)
	for {
		var x, y float64
		cmd := circle.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		ras.AddVertex(x, y, uint32(cmd))
	}
	renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rb, color.RGBA8[color.Linear]{153, 153, 153, uint8(0.7 * float64(c1.A))})

	ras.Reset()
	circle.Init(cx, cy, r, r, 0, false)
	circle.Rewind(0)
	for {
		var x, y float64
		cmd := circle.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		ras.AddVertex(x, y, uint32(cmd))
	}

	renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rb, c1)
}

func drawSourceShapeComp(rb renscan.BaseRendererInterface[color.RGBA8[color.Linear]], c1, c2 color.RGBA8[color.Linear], x1, y1, x2, y2 float64) {
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip())
	rasAdapter := &rasterizerAdapter{ras: ras}
	sl := scanline.NewScanlineU8()
	slAdapter := &scanlineWrapperU8{sl: sl}

	// Just use a rectangle for now since we don't have a path helper here
	ras.AddVertex(x1, y1, uint32(basics.PathCmdMoveTo))
	ras.AddVertex(x2, y1, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x2, y2, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x1, y2, uint32(basics.PathCmdLineTo))

	renscan.RenderScanlinesAASolid(rasAdapter, slAdapter, rb, c1)
}

func setCompOp(op int) {
	compOp = blender.CompOp(op)
}

func setCompAlphaSrc(a float64) {
	compAlphaSrc = a
}

func setCompAlphaDst(a float64) {
	compAlphaDst = a
}
