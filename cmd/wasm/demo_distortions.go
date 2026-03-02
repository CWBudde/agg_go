// Based on the original AGG examples: distortions.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/image"
	"agg_go/internal/path"
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

type imagePixFmt struct {
	rbuf *buffer.RenderingBufferU8
}

func (p *imagePixFmt) Width() int { return p.rbuf.Width() }
func (p *imagePixFmt) Height() int { return p.rbuf.Height() }
func (p *imagePixFmt) PixWidth() int { return 4 }
func (p *imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	row := buffer.RowU8(p.rbuf, y)
	return row[x*4:]
}

type distortionsSource struct {
	*image.ImageAccessorClip[imagePixFmt]
}

func (s *distortionsSource) ColorType() string { return "RGBA8" }
func (s *distortionsSource) OrderType() color.ColorOrder { return color.OrderRGBA }

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
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)

	// Image matrices
	imgW, imgH := float64(distortionsImage.Width()), float64(distortionsImage.Height())
	
	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-imgW/2, -imgH/2)
	imgMtx.Rotate(distortionsAngle * math.Pi / 180.0)
	imgMtx.Scale(distortionsScale)
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
	li := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	interpolator := span.NewSpanInterpolatorAdaptor[*span.SpanInterpolatorLinear[*transform.TransAffine], span.Distortion](li, dist)

	// Image span generator
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(distortionsImage.Data, distortionsImage.Width(), distortionsImage.Height(), distortionsImage.Width()*4)
	ipf := &imagePixFmt{rbuf: imgRbuf}
	
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	
	// Accessor
	accessor := image.NewImageAccessorClip(ipf, []basics.Int8u{255, 255, 255, 255})
	source := &distortionsSource{accessor}
	
	// Span generator - using bilinear clip
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(source, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}, interpolator)

	// Rasterizer
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, 
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Draw an ellipse with distorted image fill
	r := imgW
	if imgH < r { r = imgH }
	
	p := path.NewPathStorageStl()
	
	// Basic circle path for the ellipse
	numPoints := 100
	for i := 0; i < numPoints; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(numPoints)
		x := imgW/2 + (r/2-20)*math.Cos(angle)
		y := imgH/2 + (r/2-20)*math.Sin(angle)
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.ClosePolygon(basics.PathFlagsNone)
	
	// Manual rendering
	psAdapter := &pathSourceAdapter{ps: p}
	ras.AddPath(psAdapter, 0)
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
