// Based on the original AGG examples: gradients.cpp.
// Ported to the low-level rendering pipeline (no Agg2D).
package main

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// Native demo dimensions (matches C++ gradients.cpp platform_support window).
const (
	gradientsNativeW = 512
	gradientsNativeH = 400
	// Centre of the gradient ellipse in the native y-up coordinate frame.
	gradNativeCX = 350.0
	gradNativeCY = 280.0
)

// gradEllipseConvVS adapts shapes.Ellipse to the conv.VertexSource interface.
type gradEllipseConvVS struct{ e *shapes.Ellipse }

func (v *gradEllipseConvVS) Rewind(pathID uint) { v.e.Rewind(uint32(pathID)) }
func (v *gradEllipseConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = v.e.Vertex(&x, &y)
	return x, y, cmd
}

// gradRasConvVS adapts conv.VertexSource to the rasterizer VertexSource interface.
type gradRasConvVS struct{ vs conv.VertexSource }

func (v *gradRasConvVS) Rewind(id uint32) { v.vs.Rewind(uint(id)) }
func (v *gradRasConvVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.vs.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// gradColorFunc is a 256-entry colour lookup table that implements span.ColorFunction.
type gradColorFunc struct {
	colors [256]color.RGBA8[color.Linear]
}

func (f *gradColorFunc) Size() int                               { return 256 }
func (f *gradColorFunc) ColorAt(i int) color.RGBA8[color.Linear] { return f.colors[i] }

// buildDefaultGradientLUT builds the default colour profile used by the C++ demo.
// The spline controls default to a linear 1→0 ramp for R, G, B and constant 1
// for A. The gamma-profile control defaults to an identity mapping (linear).
// This reproduces the out-of-the-box appearance without interactive widgets.
func buildDefaultGradientLUT() *gradColorFunc {
	f := &gradColorFunc{}
	for i := 0; i < 256; i++ {
		// C++ spline defaults: y = 1 − x for R, G, B; y = 1 for A.
		// With the identity gamma profile, colors[i] is sampled at i/255.
		// R = G = B = 1 − i/255, A = 1.
		v := uint8(255 - i)
		f.colors[i] = color.RGBA8[color.Linear]{R: v, G: v, B: v, A: 255}
	}
	return f
}

func drawGradientsDemo() {
	img := ctx.GetImage()

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	// Canvas is 800×600; native demo is 512×400.
	// The standalone uses FlipY=true (y=0 at bottom).
	// Map native (x,y) with y-up to canvas y-down:
	//   canvasX = nativeX + offsetX
	//   canvasY = (nativeH - nativeY) + offsetY
	offsetX := float64(width-gradientsNativeW) / 2  // 144
	offsetY := float64(height-gradientsNativeH) / 2 // 100

	// Gradient centre in canvas (y-down) coordinates.
	cx := gradNativeCX + offsetX
	cy := float64(gradientsNativeH)-gradNativeCY + offsetY // (400-280)+100 = 220

	colorFunc := buildDefaultGradientLUT()

	// Gradient transform: maps canvas pixels back to gradient space (circle of
	// radius 150 centred at the gradient centre).
	mtxG := transform.NewTransAffine()
	mtxG.Multiply(transform.NewTransAffineTranslation(cx, cy))
	mtxG.Invert()

	// Shape transform: translate ellipse (centred at origin) to the canvas position.
	mtxShape := transform.NewTransAffine()
	mtxShape.Multiply(transform.NewTransAffineTranslation(cx, cy))

	// Build the ellipse (110 px radius, matching C++ demo).
	ellipse := shapes.NewEllipseWithParams(0, 0, 110, 110, 64, false)
	ellipsePath := conv.NewConvTransform(&gradEllipseConvVS{e: ellipse}, mtxShape)

	interpolator := span.NewSpanInterpolatorLinearDefault(mtxG)
	spanGen := span.NewSpanGradient(interpolator, span.GradientRadial{}, colorFunc, 0, 150)

	ras.Reset()
	ras.AddPath(&gradRasConvVS{vs: ellipsePath}, 0)
	renscan.RenderScanlinesAA(ras, sl, ren, alloc, spanGen)
}
