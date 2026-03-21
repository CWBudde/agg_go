// Based on the original AGG examples: distortions.cpp.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
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
		// C++ parity: a = cos(...)*(amplitude/dist) + 1, with amplitude already inverted at setup.
		a := math.Cos(dist/(16.0*d.period)-d.phase)*(d.amplitude/dist) + 1.0
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

func (p imagePixFmt) Width() int    { return p.rbuf.Width() }
func (p imagePixFmt) Height() int   { return p.rbuf.Height() }
func (p imagePixFmt) PixWidth() int { return 4 }
func (p imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	row := buffer.RowU8(p.rbuf, y)
	return row[x*4:]
}

type distortionsSource struct {
	accessor *image.ImageAccessorClip[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *distortionsSource) Width() int                  { return s.ipf.Width() }
func (s *distortionsSource) Height() int                 { return s.ipf.Height() }
func (s *distortionsSource) ColorType() string           { return "RGBA8" }
func (s *distortionsSource) OrderType() color.ColorOrder { return color.OrderRGBA }

// Delegate SpanInterpolatorInterface methods to accessor
func (s *distortionsSource) Span(x, y, length int) []basics.Int8u {
	return s.accessor.Span(x, y, length)
}

func (s *distortionsSource) NextX() []basics.Int8u {
	return s.accessor.NextX()
}

func (s *distortionsSource) NextY() []basics.Int8u {
	return s.accessor.NextY()
}

func (s *distortionsSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

// spanGeneratorAdapter bridges signature mismatch
type spanGeneratorAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*distortionsSource, *span.SpanInterpolatorAdaptor[*span.SpanInterpolatorLinear[*transform.TransAffine], span.Distortion]]
}

func (a *spanGeneratorAdapter) Prepare() {}

func (a *spanGeneratorAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	a.sg.Generate(colors[:length], x, y)
}

// --- Demo state ---

var (
	distortionsCenterX   = math.NaN()
	distortionsCenterY   = math.NaN()
	distortionsPhase     = 0.0
	distortionsAngle     = 20.0
	distortionsScale     = 1.0
	distortionsAmplitude = 10.0
	distortionsPeriod    = 1.0
	distortionsType      = 0 // 0: Wave, 1: Swirl
	distortionsImageType = 0 // 0: spheres, 1: test-grid
	distortionsImage     *agg.Image

	// Reusable components
	distortionsRbuf        *buffer.RenderingBufferU8
	distortionsPixFmt      *pixfmt.PixFmtRGBA32Pre[color.Linear]
	distortionsRenBase     *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]]
	distortionsAlloc       *span.SpanAllocator[color.RGBA8[color.Linear]]
	distortionsRas         *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	distortionsSl          *scanline.ScanlineU8
	distortionsPath        *path.PathStorageStl
	distortionsInitialized bool
)

func initDistortionsDemo() {
	if distortionsInitialized {
		return
	}

	if distortionsImage == nil {
		distortionsImage = createDistortionsSourceImage(distortionsImageType)
	}

	distortionsRbuf = buffer.NewRenderingBufferU8()
	distortionsPixFmt = pixfmt.NewPixFmtRGBA32PreLinear(distortionsRbuf)
	distortionsRenBase = renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](distortionsPixFmt)
	distortionsAlloc = span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	distortionsRas = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	distortionsSl = scanline.NewScanlineU8()
	distortionsPath = path.NewPathStorageStl()

	distortionsInitialized = true
}

func createDistortionsSourceImage(imageType int) *agg.Image {
	switch imageType {
	case 1:
		return createTestImage(width/2, height/2)
	default:
		// Original AGG demo uses "spheres" image; procedural spheres gives much closer visual parity.
		return createSpheresImage(width/2, height/2)
	}
}

func setDistortionsImageType(t int) {
	if t < 0 || t > 1 || distortionsImageType == t {
		return
	}
	distortionsImageType = t
	distortionsImage = createDistortionsSourceImage(distortionsImageType)
	// Reinitialize default center for the new source dimensions until user drags again.
	distortionsCenterX = math.NaN()
	distortionsCenterY = math.NaN()
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
	distortionsRbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	distortionsRenBase.Attach(distortionsPixFmt)

	// Image matrices
	imgW, imgH := float64(distortionsImage.Width()), float64(distortionsImage.Height())
	if math.IsNaN(distortionsCenterX) || math.IsNaN(distortionsCenterY) {
		// Match original on_init default center: image center plus demo offset.
		distortionsCenterX = imgW/2 + 10
		distortionsCenterY = imgH/2 + 50
	}

	imgMtx := transform.NewTransAffine()
	srcMtx := transform.NewTransAffine()
	srcMtx.Translate(-imgW/2, -imgH/2)
	srcMtx.Rotate(distortionsAngle * math.Pi / 180.0)
	srcMtx.Translate(imgW/2+10, imgH/2+50)

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
	ipf := imagePixFmt{rbuf: imgRbuf}

	// Accessor
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{255, 255, 255, 255})
	source := &distortionsSource{accessor: accessor, ipf: &ipf}

	// Span generator - using bilinear clip
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(source, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}, interpolator)
	adapterSG := &spanGeneratorAdapter{sg: sg}

	// Draw an ellipse with distorted image fill
	r := imgW
	if imgH < r {
		r = imgH
	}

	distortionsPath.RemoveAll()
	numPoints := 100
	for i := 0; i < numPoints; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(numPoints)
		x := imgW/2 + (r/2-20)*math.Cos(angle)
		y := imgH/2 + (r/2-20)*math.Sin(angle)
		srcMtx.Transform(&x, &y)
		if i == 0 {
			distortionsPath.MoveTo(x, y)
		} else {
			distortionsPath.LineTo(x, y)
		}
	}
	distortionsPath.ClosePolygon(basics.PathFlagsClose)

	// Manual rendering loop
	psAdapter := &pathSourceAdapter{ps: distortionsPath}
	distortionsRas.Reset()
	distortionsRas.AddPath(psAdapter, 0)

	if distortionsRas.RewindScanlines() {
		distortionsSl.Reset(distortionsRas.MinX(), distortionsRas.MaxX())
		for distortionsRas.SweepScanline(distortionsSl) {
			y := distortionsSl.Y()
			for _, spanData := range distortionsSl.Spans() {
				if spanData.Len > 0 {
					colors := distortionsAlloc.Allocate(int(spanData.Len))
					adapterSG.Generate(colors, int(spanData.X), y, int(spanData.Len))
					distortionsRenBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}

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
