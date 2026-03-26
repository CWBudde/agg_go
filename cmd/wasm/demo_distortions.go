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
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
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

// --- Gradient color table (matches C++ g_gradient_colors, 256 RGBA entries) ---

//nolint:gochecknoglobals
var distortionsGradColorTable = [256 * 4]uint8{
	255, 255, 255, 255, 255, 255, 254, 255, 255, 255, 254, 255, 255, 255, 254, 255,
	255, 255, 253, 255, 255, 255, 253, 255, 255, 255, 252, 255, 255, 255, 251, 255,
	255, 255, 250, 255, 255, 255, 248, 255, 255, 255, 246, 255, 255, 255, 244, 255,
	255, 255, 241, 255, 255, 255, 238, 255, 255, 255, 235, 255, 255, 255, 231, 255,
	255, 255, 227, 255, 255, 255, 222, 255, 255, 255, 217, 255, 255, 255, 211, 255,
	255, 255, 206, 255, 255, 255, 200, 255, 255, 254, 194, 255, 255, 253, 188, 255,
	255, 252, 182, 255, 255, 250, 176, 255, 255, 249, 170, 255, 255, 247, 164, 255,
	255, 246, 158, 255, 255, 244, 152, 255, 254, 242, 146, 255, 254, 240, 141, 255,
	254, 238, 136, 255, 254, 236, 131, 255, 253, 234, 126, 255, 253, 232, 121, 255,
	253, 229, 116, 255, 252, 227, 112, 255, 252, 224, 108, 255, 251, 222, 104, 255,
	251, 219, 100, 255, 251, 216, 96, 255, 250, 214, 93, 255, 250, 211, 89, 255,
	249, 208, 86, 255, 249, 205, 83, 255, 248, 202, 80, 255, 247, 199, 77, 255,
	247, 196, 74, 255, 246, 193, 72, 255, 246, 190, 69, 255, 245, 187, 67, 255,
	244, 183, 64, 255, 244, 180, 62, 255, 243, 177, 60, 255, 242, 174, 58, 255,
	242, 170, 56, 255, 241, 167, 54, 255, 240, 164, 52, 255, 239, 161, 51, 255,
	239, 157, 49, 255, 238, 154, 47, 255, 237, 151, 46, 255, 236, 147, 44, 255,
	235, 144, 43, 255, 235, 141, 41, 255, 234, 138, 40, 255, 233, 134, 39, 255,
	232, 131, 37, 255, 231, 128, 36, 255, 230, 125, 35, 255, 229, 122, 34, 255,
	228, 119, 33, 255, 227, 116, 31, 255, 226, 113, 30, 255, 225, 110, 29, 255,
	224, 107, 28, 255, 223, 104, 27, 255, 222, 101, 26, 255, 221, 99, 25, 255,
	220, 96, 24, 255, 219, 93, 23, 255, 218, 91, 22, 255, 217, 88, 21, 255,
	216, 86, 20, 255, 215, 83, 19, 255, 214, 81, 18, 255, 213, 79, 17, 255,
	212, 77, 17, 255, 211, 74, 16, 255, 210, 72, 15, 255, 209, 70, 14, 255,
	207, 68, 13, 255, 206, 66, 13, 255, 205, 64, 12, 255, 204, 62, 11, 255,
	203, 60, 10, 255, 202, 58, 10, 255, 201, 56, 9, 255, 199, 55, 9, 255,
	198, 53, 8, 255, 197, 51, 7, 255, 196, 50, 7, 255, 195, 48, 6, 255,
	193, 46, 6, 255, 192, 45, 5, 255, 191, 43, 5, 255, 190, 42, 4, 255,
	188, 41, 4, 255, 187, 39, 3, 255, 186, 38, 3, 255, 185, 37, 2, 255,
	183, 35, 2, 255, 182, 34, 1, 255, 181, 33, 1, 255, 179, 32, 1, 255,
	178, 30, 0, 255, 177, 29, 0, 255, 175, 28, 0, 255, 174, 27, 0, 255,
	173, 26, 0, 255, 171, 25, 0, 255, 170, 24, 0, 255, 168, 23, 0, 255,
	167, 22, 0, 255, 165, 21, 0, 255, 164, 21, 0, 255, 163, 20, 0, 255,
	161, 19, 0, 255, 160, 18, 0, 255, 158, 17, 0, 255, 156, 17, 0, 255,
	155, 16, 0, 255, 153, 15, 0, 255, 152, 14, 0, 255, 150, 14, 0, 255,
	149, 13, 0, 255, 147, 12, 0, 255, 145, 12, 0, 255, 144, 11, 0, 255,
	142, 11, 0, 255, 140, 10, 0, 255, 139, 10, 0, 255, 137, 9, 0, 255,
	135, 9, 0, 255, 134, 8, 0, 255, 132, 8, 0, 255, 130, 7, 0, 255,
	128, 7, 0, 255, 126, 6, 0, 255, 125, 6, 0, 255, 123, 5, 0, 255,
	121, 5, 0, 255, 119, 4, 0, 255, 117, 4, 0, 255, 115, 4, 0, 255,
	113, 3, 0, 255, 111, 3, 0, 255, 109, 2, 0, 255, 107, 2, 0, 255,
	105, 2, 0, 255, 103, 1, 0, 255, 101, 1, 0, 255, 99, 1, 0, 255,
	97, 0, 0, 255, 95, 0, 0, 255, 93, 0, 0, 255, 91, 0, 0, 255,
	90, 0, 0, 255, 88, 0, 0, 255, 86, 0, 0, 255, 84, 0, 0, 255,
	82, 0, 0, 255, 80, 0, 0, 255, 78, 0, 0, 255, 77, 0, 0, 255,
	75, 0, 0, 255, 73, 0, 0, 255, 72, 0, 0, 255, 70, 0, 0, 255,
	68, 0, 0, 255, 67, 0, 0, 255, 65, 0, 0, 255, 64, 0, 0, 255,
	63, 0, 0, 255, 61, 0, 0, 255, 60, 0, 0, 255, 59, 0, 0, 255,
	58, 0, 0, 255, 57, 0, 0, 255, 56, 0, 0, 255, 55, 0, 0, 255,
	54, 0, 0, 255, 53, 0, 0, 255, 53, 0, 0, 255, 52, 0, 0, 255,
	52, 0, 0, 255, 51, 0, 0, 255, 51, 0, 0, 255, 51, 0, 0, 255,
	50, 0, 0, 255, 50, 0, 0, 255, 51, 0, 0, 255, 51, 0, 0, 255,
	51, 0, 0, 255, 51, 0, 0, 255, 52, 0, 0, 255, 52, 0, 0, 255,
	53, 0, 0, 255, 54, 1, 0, 255, 55, 2, 0, 255, 56, 3, 0, 255,
	57, 4, 0, 255, 58, 5, 0, 255, 59, 6, 0, 255, 60, 7, 0, 255,
	62, 8, 0, 255, 63, 9, 0, 255, 64, 11, 0, 255, 66, 12, 0, 255,
	68, 13, 0, 255, 69, 14, 0, 255, 71, 16, 0, 255, 73, 17, 0, 255,
	75, 18, 0, 255, 77, 20, 0, 255, 79, 21, 0, 255, 81, 23, 0, 255,
	83, 24, 0, 255, 85, 26, 0, 255, 87, 28, 0, 255, 90, 29, 0, 255,
	92, 31, 0, 255, 94, 33, 0, 255, 97, 34, 0, 255, 99, 36, 0, 255,
	102, 38, 0, 255, 104, 40, 0, 255, 107, 41, 0, 255, 109, 43, 0, 255,
	112, 45, 0, 255, 115, 47, 0, 255, 117, 49, 0, 255, 120, 51, 0, 255,
	123, 52, 0, 255, 126, 54, 0, 255, 128, 56, 0, 255, 131, 58, 0, 255,
	134, 60, 0, 255, 137, 62, 0, 255, 140, 64, 0, 255, 143, 66, 0, 255,
	145, 68, 0, 255, 148, 70, 0, 255, 151, 72, 0, 255, 154, 74, 0, 255,
}

// distortionsGradColorFunc is a pre-built 256-entry gradient color LUT.
type distortionsGradColorFunc struct {
	colors [256]color.RGBA8[color.Linear]
}

func (g *distortionsGradColorFunc) Size() int { return 256 }
func (g *distortionsGradColorFunc) ColorAt(i int) color.RGBA8[color.Linear] {
	return g.colors[i]
}

var distortionsGradColors *distortionsGradColorFunc //nolint:gochecknoglobals

func buildDistortionsGradColors() *distortionsGradColorFunc {
	if distortionsGradColors != nil {
		return distortionsGradColors
	}
	f := &distortionsGradColorFunc{}
	for i := range 256 {
		f.colors[i] = color.RGBA8[color.Linear]{
			R: distortionsGradColorTable[i*4+0],
			G: distortionsGradColorTable[i*4+1],
			B: distortionsGradColorTable[i*4+2],
			A: distortionsGradColorTable[i*4+3],
		}
	}
	distortionsGradColors = f
	return f
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
	distortionsSlP8        *scanline.ScanlineP8
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
	distortionsSlP8 = scanline.NewScanlineP8()
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

	// Draw ellipse outline (Pass 2): same ellipse shifted right by imgW-imgW/10.
	// Matches C++: src_mtx *= trans_affine_translation(img_width - img_width/10, 0)
	shiftX := imgW - imgW/10
	distortionsPath.RemoveAll()
	for i := 0; i < numPoints; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(numPoints)
		x := imgW/2 + (r/2-20)*math.Cos(angle)
		y := imgH/2 + (r/2-20)*math.Sin(angle)
		srcMtx.Transform(&x, &y)
		x += shiftX
		if i == 0 {
			distortionsPath.MoveTo(x, y)
		} else {
			distortionsPath.LineTo(x, y)
		}
	}
	distortionsPath.ClosePolygon(basics.PathFlagsClose)

	outlineAdapter := &pathSourceAdapter{ps: distortionsPath}
	distortionsRas.Reset()
	distortionsRas.AddPath(outlineAdapter, 0)
	renscan.RenderScanlinesAASolid(distortionsRas, distortionsSlP8, distortionsRenBase, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Pass 3: Gradient circle with distortion.
	// C++: builds gr1_mtx for the ellipse path and gr2_mtx (inverted) for the
	// interpolator, shifts the distortion center right by shiftX, then renders
	// the same ellipse filled with span_gradient<gradient_circle, color_array>.
	gr1Mtx := transform.NewTransAffine()
	gr1Mtx.Translate(-imgW/2, -imgH/2)
	gr1Mtx.Scale(0.8)
	gr1Mtx.Rotate(distortionsAngle * math.Pi / 180.0)
	gr1Mtx.Translate(shiftX+imgW/2+10, imgH/2+50)

	gr2Mtx := transform.NewTransAffine()
	gr2Mtx.Rotate(distortionsAngle * math.Pi / 180.0)
	gr2Mtx.Scale(distortionsScale)
	gr2Mtx.Translate(shiftX+imgW/2+10+50, imgH/2+100)
	gr2Mtx.Invert()

	cx2, cy2 := distortionsCenterX+shiftX, distortionsCenterY
	gr2Mtx.Transform(&cx2, &cy2)

	db2 := distortionBase{
		period:    distortionsPeriod,
		amplitude: 1.0 / distortionsAmplitude,
		phase:     distortionsPhase,
		cx:        cx2,
		cy:        cy2,
	}
	var dist2 span.Distortion
	if distortionsType == 0 {
		dist2 = &distortionWave{db2}
	} else {
		dist2 = &distortionSwirl{db2}
	}

	li2 := span.NewSpanInterpolatorLinear[*transform.TransAffine](gr2Mtx, 8)
	interpolator.SetBase(li2)
	interpolator.SetDistortion(dist2)

	gradColors := buildDistortionsGradColors()
	gradSpan := span.NewSpanGradient(interpolator, span.GradientRadial{}, gradColors, 0.0, 180.0)

	distortionsPath.RemoveAll()
	for i := 0; i < numPoints; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(numPoints)
		x := imgW/2 + (r/2-20)*math.Cos(angle)
		y := imgH/2 + (r/2-20)*math.Sin(angle)
		gr1Mtx.Transform(&x, &y)
		if i == 0 {
			distortionsPath.MoveTo(x, y)
		} else {
			distortionsPath.LineTo(x, y)
		}
	}
	distortionsPath.ClosePolygon(basics.PathFlagsClose)

	gr1PathAdapter := &pathSourceAdapter{ps: distortionsPath}
	distortionsRas.Reset()
	distortionsRas.AddPath(gr1PathAdapter, 0)
	renscan.RenderScanlinesAA(distortionsRas, distortionsSl, distortionsRenBase, distortionsAlloc, gradSpan)

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
