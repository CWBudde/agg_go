package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
)

var (
	compAlphaSrc = 0.75
	compAlphaDst = 1.0
	compOp       = blender.CompOpSrcOver
)

func drawCompositingDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	// Attach main rendering buffer using img.Stride().
	mainRbuf := buffer.NewRenderingBufferWithData[uint8](img.Data, w, h, img.Stride())
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainRbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)

	// Clear to white.
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// Draw checkered background.
	// 0xDF sRGB → linear for the grey squares.
	greyLinear := color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{R: 0xDF, G: 0xDF, B: 0xDF, A: 0xFF})
	checkRas := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	checkSl := scanline.NewScanlineU8()
	for y := 0; y < h; y += 8 {
		xStart := ((y >> 3) & 1) << 3
		for x := xStart; x < w; x += 16 {
			fx, fy := float64(x), float64(y)
			checkRas.Reset()
			checkRas.AddVertex(fx, fy, uint32(basics.PathCmdMoveTo))
			checkRas.AddVertex(fx+7, fy, uint32(basics.PathCmdLineTo))
			checkRas.AddVertex(fx+7, fy+7, uint32(basics.PathCmdLineTo))
			checkRas.AddVertex(fx, fy+7, uint32(basics.PathCmdLineTo))
			renscan.RenderScanlinesAASolid(checkRas, checkSl, mainRb, greyLinear)
		}
	}

	// Temporary buffer for compositing (transparent black initially).
	tempBuf := make([]uint8, w*h*4)
	tempRbuf := buffer.NewRenderingBufferWithData[uint8](tempBuf, w, h, w*4)

	// Use RGBA32 (non-premultiplied) for compositing.
	pixf := pixfmt.NewPixFmtRGBA32[color.Linear](tempRbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pixf)
	rb.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0})

	// Draw destination circle.
	drawCircleComp(rb,
		color.RGBA8[color.Linear]{R: 0xFD, G: 0xF0, B: 0x6F, A: uint8(compAlphaDst * 255)},
		color.RGBA8[color.Linear]{R: 0xFE, G: 0x9F, B: 0x34, A: uint8(compAlphaDst * 255)},
		70*3, 100+24*3, 37*3, 100+79*3)

	// Draw source shape with selected compositing op.
	compBlender := blender.NewCompositeBlender[color.Linear, order.RGBA](compOp)
	compPixf := pixfmt.NewPixFmtAlphaBlendRGBA[color.Linear, blender.CompositeBlender[color.Linear, order.RGBA]](tempRbuf, compBlender)
	compRb := renderer.NewRendererBaseWithPixfmt(compPixf)

	drawSourceShapeComp(compRb,
		color.RGBA8[color.Linear]{R: 0x7F, G: 0xC1, B: 0xFF, A: uint8(compAlphaSrc * 255)},
		color.RGBA8[color.Linear]{R: 0x05, G: 0x00, B: 0x5F, A: uint8(compAlphaSrc * 255)},
		300+50, 100+24*3, 107+50, 100+79*3)

	// Blend temp buffer onto the main buffer.
	mainRb.BlendFrom(pixf, nil, 0, 0, 255)

	// Gamma-encode linear→sRGB.
	applyLinearToSRGB(img)

	logStatus(fmt.Sprintf("Compositing Demo: Op=%d, AlphaSrc=%.2f, AlphaDst=%.2f", compOp, compAlphaSrc, compAlphaDst))
}

func drawCircleComp(rb renscan.BaseRendererInterface[color.RGBA8[color.Linear]], c1, c2 color.RGBA8[color.Linear], x1, y1, x2, y2 float64) {
	r := math.Hypot(x2-x1, y2-y1) / 2
	cx, cy := (x1+x2)/2, (y1+y2)/2

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip())

	sl := scanline.NewScanlineU8()

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
	renscan.RenderScanlinesAASolid(ras, sl, rb, color.RGBA8[color.Linear]{R: 153, G: 153, B: 153, A: uint8(0.7 * float64(c1.A))})

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

	renscan.RenderScanlinesAASolid(ras, sl, rb, c1)
}

func drawSourceShapeComp(rb renscan.BaseRendererInterface[color.RGBA8[color.Linear]], c1, c2 color.RGBA8[color.Linear], x1, y1, x2, y2 float64) {
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip())

	sl := scanline.NewScanlineU8()

	// Just use a rectangle for now since we don't have a path helper here
	ras.AddVertex(x1, y1, uint32(basics.PathCmdMoveTo))
	ras.AddVertex(x2, y1, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x2, y2, uint32(basics.PathCmdLineTo))
	ras.AddVertex(x1, y2, uint32(basics.PathCmdLineTo))

	renscan.RenderScanlinesAASolid(ras, sl, rb, c1)
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
