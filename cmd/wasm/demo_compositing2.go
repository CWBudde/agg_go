package main

import (
	"fmt"

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
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

var (
	comp2AlphaSrc = 1.0
	comp2AlphaDst = 1.0
	comp2Op       = blender.CompOpSrcOver
)

func generateColorRamp(c []color.RGBA8[color.Linear], c1, c2, c3, c4 color.RGBA8[color.Linear]) {
	for i := 0; i < 85; i++ {
		c[i] = c1.Gradient(c2, basics.Int8u(float64(i)/85.0*255))
	}
	for i := 85; i < 170; i++ {
		c[i] = c2.Gradient(c3, basics.Int8u(float64(i-85)/85.0*255))
	}
	for i := 170; i < 256; i++ {
		c[i] = c3.Gradient(c4, basics.Int8u(float64(i-170)/85.0*255))
	}
}

type comp2RasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newComp2Rasterizer() *comp2RasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

func radialShape2(rb renscan.BaseRendererInterface[color.RGBA8[color.Linear]], colors []color.RGBA8[color.Linear], x1, y1, x2, y2 float64) {
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5
	r := 0.5 * mathMin(x2-x1, y2-y1)

	mtx := transform.NewTransAffine()
	mtx.Scale(r / 100.0)
	mtx.Translate(cx, cy)
	mtx.Invert()

	inter := span.NewSpanInterpolatorLinearDefault(mtx)
	grad := span.GradientRadial{}

	// Create a color function that uses the ramp
	colorFunc := &colorRampFunc{ramp: colors}

	sg := span.NewSpanGradient[color.RGBA8[color.Linear]](inter, grad, colorFunc, 0, 100)
	sa := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := newComp2Rasterizer()
	sl := scanline.NewScanlineU8()

	ras.Reset()
	ell := shapes.NewEllipseWithParams(cx, cy, r, r, 100, false)
	ell.Rewind(0)
	for {
		var x, y float64
		cmd := ell.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		ras.AddVertex(x, y, uint32(cmd))
	}

	renscan.RenderScanlinesAA(ras, sl, rb, sa, sg)
}

type colorRampFunc struct {
	ramp []color.RGBA8[color.Linear]
}

func (f *colorRampFunc) Size() int { return len(f.ramp) }
func (f *colorRampFunc) ColorAt(v int) color.RGBA8[color.Linear] {
	if v < 0 {
		v = 0
	}
	if v >= len(f.ramp) {
		v = len(f.ramp) - 1
	}
	return f.ramp[v]
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func drawCompositing2Demo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())

	// Clear to white
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(mainPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ramp1 := make([]color.RGBA8[color.Linear], 256)
	ramp2 := make([]color.RGBA8[color.Linear], 256)

	aDst := uint8(comp2AlphaDst * 255)
	aSrc := uint8(comp2AlphaSrc * 255)

	generateColorRamp(ramp1,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: aDst},
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: aDst},
		color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: aDst},
		color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 0})

	generateColorRamp(ramp2,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: aSrc},
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: aSrc},
		color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: aSrc},
		color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 0})

	// Temporary buffer for compositing
	tempBuf := make([]uint8, w*h*4)
	tempRbuf := buffer.NewRenderingBufferWithData[uint8](tempBuf, w, h, w*4)

	// Draw destination
	pixf1 := pixfmt.NewPixFmtRGBA32[color.Linear](tempRbuf)
	rb1 := renderer.NewRendererBaseWithPixfmt(pixf1)
	rb1.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0})

	// Difference mode for destination background as in C++ example
	compBlenderDiff := blender.NewCompositeBlender[color.Linear, order.RGBA](blender.CompOpDifference)
	pixfDiff := pixfmt.NewPixFmtAlphaBlendRGBA[color.Linear, blender.CompositeBlender[color.Linear, order.RGBA]](tempRbuf, compBlenderDiff)
	rbDiff := renderer.NewRendererBaseWithPixfmt(pixfDiff)
	radialShape2(rbDiff, ramp1, 50, 50, 50+320, 50+320)

	// Draw source with selected comp op
	compBlender := blender.NewCompositeBlender[color.Linear, order.RGBA](comp2Op)
	compPixf := pixfmt.NewPixFmtAlphaBlendRGBA[color.Linear, blender.CompositeBlender[color.Linear, order.RGBA]](tempRbuf, compBlender)
	compRb := renderer.NewRendererBaseWithPixfmt(compPixf)

	cx, cy := 50.0, 50.0
	radialShape2(compRb, ramp2, cx+120-70, cy+120-70, cx+120+70, cy+120+70)
	radialShape2(compRb, ramp2, cx+200-70, cy+120-70, cx+200+70, cy+120+70)
	radialShape2(compRb, ramp2, cx+120-70, cy+200-70, cx+120+70, cy+200+70)
	radialShape2(compRb, ramp2, cx+200-70, cy+200-70, cx+200+70, cy+200+70)

	// Blend composited result back to main buffer
	mainRb.BlendFrom(pixf1, nil, 0, 0, 255)

	applyLinearToSRGB(img)

	logStatus(fmt.Sprintf("Compositing 2 Demo: Op=%d, AlphaSrc=%.2f, AlphaDst=%.2f", comp2Op, comp2AlphaSrc, comp2AlphaDst))
}

func setComp2Op(op int) {
	comp2Op = blender.CompOp(op)
}

func setComp2AlphaSrc(a float64) {
	comp2AlphaSrc = a
}

func setComp2AlphaDst(a float64) {
	comp2AlphaDst = a
}
