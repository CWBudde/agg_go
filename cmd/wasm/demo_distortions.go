// Based on the original AGG examples: distortions.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// --- Distortion implementations ---

type distortionBase struct {
	cx, cy    float64
	period    float64
	amplitude float64
	phase     float64
}

type distortionWave struct {
	distortionBase
}

func (d *distortionWave) Calculate(x, y *int) {
	xd := float64(*x)/float64(basics.PolySubpixelScale) - d.cx
	yd := float64(*y)/float64(basics.PolySubpixelScale) - d.cy
	dist := math.Sqrt(xd*xd + yd*yd)
	if dist > 1 {
		a := math.Cos(dist/(16.0*d.period)-d.phase)*(1.0/(d.amplitude*dist)) + 1.0
		*x = int((xd*a + d.cx) * float64(basics.PolySubpixelScale))
		*y = int((yd*a + d.cy) * float64(basics.PolySubpixelScale))
	}
}

type distortionSwirl struct {
	distortionBase
}

func (d *distortionSwirl) Calculate(x, y *int) {
	xd := float64(*x)/float64(basics.PolySubpixelScale) - d.cx
	yd := float64(*y)/float64(basics.PolySubpixelScale) - d.cy
	a := (100.0 - math.Sqrt(xd*xd+yd*yd)) / 100.0 * (0.1 / -d.amplitude)
	sa := math.Sin(a - d.phase/25.0)
	ca := math.Cos(a - d.phase/25.0)
	*x = int((xd*ca - yd*sa + d.cx) * float64(basics.PolySubpixelScale))
	*y = int((xd*sa + yd*ca + d.cy) * float64(basics.PolySubpixelScale))
}

// --- Demo state ---

var (
	distortionsCenterX   = 400.0
	distortionsCenterY   = 300.0
	distortionsPhase     = 0.0
	distortionsAngle     = 20.0
	distortionsScale     = 1.0
	distortionsAmplitude = 10.0
	distortionsPeriod    = 1.0
	distortionsType      = 0 // 0: Wave, 1: Swirl
	distortionsImage     *agg.Image
)

func initDistortionsDemo() {
	if distortionsImage == nil {
		distortionsImage = createTestImage(width/2, height/2)
	}
}

func drawDistortionsDemo() {
	initDistortionsDemo()

	// Update phase for animation
	distortionsPhase += 15.0 * math.Pi / 180.0
	if distortionsPhase > math.Pi*200.0 {
		distortionsPhase -= math.Pi * 200.0
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)

	// Image matrices
	imgW, imgH := float64(distortionsImage.Width()), float64(distortionsImage.Height())
	
	srcMtx := transform.NewTransAffine()
	srcMtx.Translate(-imgW/2, -imgH/2)
	srcMtx.Rotate(distortionsAngle * math.Pi / 180.0)
	srcMtx.Translate(imgW/2+10, imgH/2+50)

	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-imgW/2, -imgH/2)
	imgMtx.Rotate(distortionsAngle * math.Pi / 180.0)
	imgMtx.Scale(distortionsScale, distortionsScale)
	imgMtx.Translate(imgW/2+10, imgH/2+50)
	imgMtx.Invert()

	// Distortion
	var dist span.Distortion
	db := distortionBase{
		period:    distortionsPeriod,
		amplitude: 1.0 / distortionsAmplitude,
		phase:     distortionsPhase,
	}
	
	cx, cy := distortionsCenterX, distortionsCenterY
	imgMtx.Transform(&cx, &cy)
	db.cx, db.cy = cx, cy

	if distortionsType == 0 {
		dist = &distortionWave{db}
	} else {
		dist = &distortionSwirl{db}
	}

	// Interpolator
	li := span.NewSpanInterpolatorLinearWithTransformer(imgMtx)
	interpolator := span.NewSpanInterpolatorAdaptor(li, dist)

	// Image span generator
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(distortionsImage.Data, distortionsImage.Width(), distortionsImage.Height(), distortionsImage.Stride())
	imgPixFmt := pixfmt.NewPixFmtRGBA32PreLinear(imgRbuf)
	
	// We use SpanImageFilterRGBA to draw the distorted image
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	
	// Accessor
	accessor := span.NewImageAccessorClip[pixfmt.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](imgPixFmt, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})
	
	// Span generator
	filter := span.NewImageFilterBilinearRGBA8()
	sg := span.NewSpanImageFilterRGBA[span.ImageAccessorInterface[color.RGBA8[color.Linear]], *span.SpanInterpolatorAdaptor[*span.SpanInterpolatorLinear, span.Distortion]](accessor, interpolator, filter)

	// Rasterizer
	ras := rasterizer.NewRasterizerScanlineAA()
	sl := scanline.NewScanlineU8()

	// Draw an ellipse with distorted image fill
	r := imgW
	if imgH < r { r = imgH }
	agg2d.ResetPath()
	agg2d.Ellipse(imgW/2, imgH/2, r/2-20, r/2-20)
	
	// Manual rendering
	ras.AddPath(agg2d.GetInternalPath(), 0)
	renscan.RenderScanlinesAA(ras, sl, renBase, alloc, sg)

	// Draw interactive handle
	drawHandle(distortionsCenterX, distortionsCenterY)
}

func handleDistortionsMouseDown(x, y float64) bool {
	distortionsCenterX = x
	distortionsCenterY = y
	return true
}

func handleDistortionsMouseMove(x, y float64) bool {
	distortionsCenterX = x
	distortionsCenterY = y
	return true
}

func handleDistortionsMouseUp() {}
