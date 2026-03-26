// Based on the original AGG example: image_alpha.cpp
// Demonstrates using brightness as an alpha channel: a large ellipse is filled
// with a rotated image where the alpha value of each pixel is derived from the
// pixel's luminance via a configurable lookup table.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/spline"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageassets"
	"github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// --- Demo state ---

var (
	imgAlphaImage *agg.Image

	// Background ellipses (randomised once)
	imgAlphaEllipses []imgAlphaEllipse

	// Brightness-to-alpha LUT (256*3 entries): alpha = f(r+g+b)
	imgAlphaLUT  [256 * 3]uint8
	imgAlphaCtrl *spline.SplineCtrl[color.RGBA8[color.Linear]]

	// Reusable components
	imgAlphaRbuf        *buffer.RenderingBufferU8
	imgAlphaPixFmt      *pixfmt.PixFmtRGBA32[color.Linear]
	imgAlphaRenBase     *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]
	imgAlphaAlloc       *span.SpanAllocator[color.RGBA8[color.Linear]]
	imgAlphaRas         *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	imgAlphaSl          *scanline.ScanlineU8
	imgAlphaPath        *path.PathStorageStl
	imgAlphaInitialized bool
)

type imgAlphaEllipse struct {
	x, y, rx, ry float64
	r, g, b, a   uint8
}

// imgAlphaSpanGen wraps a bilinear clip span generator and applies brightness→alpha.
type imgAlphaSpanGen struct {
	inner *span.SpanImageFilterRGBABilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
	lut   *[256 * 3]uint8
}

func (g *imgAlphaSpanGen) Prepare() {}
func (g *imgAlphaSpanGen) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	g.inner.Generate(colors[:length], x, y)
	// Apply brightness → alpha from LUT (same as C++ span_conv_brightness_alpha)
	for i := 0; i < length; i++ {
		c := &colors[i]
		sum := int(c.R) + int(c.G) + int(c.B) // 0..765
		lutIdx := sum * (256 * 3) / (3 * 256)
		if lutIdx >= 256*3 {
			lutIdx = 256*3 - 1
		}
		c.A = g.lut[lutIdx]
	}
}

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRandSeed(seed int32) *clibcRand {
	_ = seed
	return &clibcRand{
		state: [31]int32{
			-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
			1627687941, -179304937, -2073333483, 1780058412, -1989503057,
			-615974602, 344556628, 939512070, -1249116260, 1507946756,
			-812545463, 154635395, 1388815473, -1926676823, 525320961,
			-1009028674, 968117788, -123449607, 1284210865, 435012392,
			-2017506339, -911064859, -370259173, 1132637927, 1398500161, -205601318,
		},
		fptr: 3,
		rptr: 0,
	}
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := int32(uint32(r.state[r.fptr]) >> 1)
	r.fptr++
	if r.fptr >= len(r.state) {
		r.fptr = 0
	}
	r.rptr++
	if r.rptr >= len(r.state) {
		r.rptr = 0
	}
	return result
}

func (r *clibcRand) randN(n int) int {
	return int(r.next()) % n
}

func initImgAlphaDemo() {
	if imgAlphaInitialized {
		return
	}
	imgAlphaRbuf = buffer.NewRenderingBufferU8()
	imgAlphaPixFmt = pixfmt.NewPixFmtRGBA32Linear(imgAlphaRbuf)
	imgAlphaRenBase = renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](imgAlphaPixFmt)
	imgAlphaAlloc = span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	imgAlphaRas = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	imgAlphaSl = scanline.NewScanlineU8()
	imgAlphaPath = path.NewPathStorageStl()
	imgAlphaCtrl = spline.NewSplineCtrl[color.RGBA8[color.Linear]](2, 2, 200, 30, 6, false)
	imgAlphaCtrl.SetBackgroundColor(color.NewRGBA8[color.Linear](255, 255, 230, 255))
	imgAlphaCtrl.SetBorderColor(color.NewRGBA8[color.Linear](0, 0, 0, 255))
	imgAlphaCtrl.SetCurveColor(color.NewRGBA8[color.Linear](0, 0, 0, 255))
	imgAlphaCtrl.SetInactivePointColor(color.NewRGBA8[color.Linear](0, 0, 0, 255))
	imgAlphaCtrl.SetActivePointColor(color.NewRGBA8[color.Linear](255, 0, 0, 255))

	// Build background ellipses using the C rand() sequence to match AGG.
	rng := newClibcRandSeed(1)
	imgAlphaEllipses = make([]imgAlphaEllipse, 50)
	for i := range imgAlphaEllipses {
		imgAlphaEllipses[i] = imgAlphaEllipse{
			x:  float64(rng.randN(width)),
			y:  float64(rng.randN(height)),
			rx: float64(rng.randN(60) + 10),
			ry: float64(rng.randN(60) + 10),
			r:  uint8(rng.randN(256)),
			g:  uint8(rng.randN(256)),
			b:  uint8(rng.randN(256)),
			a:  uint8(rng.randN(256)),
		}
	}

	// Default brightness→alpha LUT: same as C++ defaults (control points 1,1,1,0.5,0.5,1).
	imgAlphaCtrl.SetValue(0, 1.0)
	imgAlphaCtrl.SetValue(1, 1.0)
	imgAlphaCtrl.SetValue(2, 1.0)
	imgAlphaCtrl.SetValue(3, 0.5)
	imgAlphaCtrl.SetValue(4, 0.5)
	imgAlphaCtrl.SetValue(5, 1.0)
	for i := range imgAlphaLUT {
		t := float64(i) / float64(len(imgAlphaLUT))
		imgAlphaLUT[i] = uint8(imgAlphaCtrl.Value(t)*255.0 + 0.5)
	}

	imgAlphaInitialized = true
}

func drawImageAlphaDemo() {
	initImgAlphaDemo()

	if imgAlphaImage == nil {
		if src, err := imageassets.Spheres(); err == nil && src != nil {
			imgAlphaImage = src
		} else {
			imgAlphaImage = createSpheresImage(400, 400)
		}
	}

	imgW := float64(imgAlphaImage.Width())
	imgH := float64(imgAlphaImage.Height())

	// Attach rendering target
	img := ctx.GetImage()
	imgAlphaRbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	// Render background ellipses using the low-level pipeline.
	ell := shapes.NewEllipse()
	ellAdapter := &ellipseVS{ell: ell}
	for _, e := range imgAlphaEllipses {
		c := color.RGBA8[color.Linear]{R: e.r, G: e.g, B: e.b, A: e.a}
		ell.Init(e.x, e.y, e.rx, e.ry, 32, false)
		imgAlphaRas.Reset()
		imgAlphaRas.AddPath(ellAdapter, 0)
		renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](imgAlphaRas, imgAlphaSl, imgAlphaRenBase, c)
	}

	// Image transform: 10° rotation around image center, then placed at screen center
	cx := float64(width) * 0.5
	cy := float64(height) * 0.5
	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-imgW/2, -imgH/2)
	imgMtx.Rotate(10.0 * math.Pi / 180.0)
	imgMtx.Translate(cx, cy)
	imgMtx.Invert()

	// Same transform for the polygon (not inverted)
	polyMtx := transform.NewTransAffine()
	polyMtx.Translate(-imgW/2, -imgH/2)
	polyMtx.Rotate(10.0 * math.Pi / 180.0)
	polyMtx.Translate(cx, cy)

	// Span interpolator
	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)

	// Image source
	imgRbuf := buffer.NewRenderingBufferU8()
	// Match the AGG reference: the source texture is sampled bottom-up.
	imgRbuf.Attach(imgAlphaImage.Data, imgAlphaImage.Width(), imgAlphaImage.Height(), -imgAlphaImage.Stride())
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{0, 0, 0, 0})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	bgColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0}
	innerSG := span.NewSpanImageFilterRGBABilinearClipWithParams(src, bgColor, interp)
	sg := &imgAlphaSpanGen{inner: innerSG, lut: &imgAlphaLUT}

	// Large ellipse clipped to screen, rotated (polygon transform)
	r := imgW * 0.9
	if imgH*0.9 < r {
		r = imgH * 0.9
	}

	imgAlphaPath.RemoveAll()
	numPoints := 200

	for i := range numPoints {
		a := 2.0 * math.Pi * float64(i) / float64(numPoints)
		px := imgW*0.5 + r*0.5*math.Cos(a)
		py := imgH*0.5 + r*0.5*math.Sin(a)
		polyMtx.Transform(&px, &py)
		if i == 0 {
			imgAlphaPath.MoveTo(px, py)
		} else {
			imgAlphaPath.LineTo(px, py)
		}
	}
	imgAlphaPath.ClosePolygon(basics.PathFlagsNone)

	imgAlphaRas.Reset()
	imgAlphaRas.ClipBox(0, 0, float64(width), float64(height))
	imgAlphaRas.AddPath(&pathSourceAdapter{ps: imgAlphaPath}, 0)

	if imgAlphaRas.RewindScanlines() {
		imgAlphaSl.Reset(imgAlphaRas.MinX(), imgAlphaRas.MaxX())
		for imgAlphaRas.SweepScanline(imgAlphaSl) {
			y := imgAlphaSl.Y()
			for _, spanData := range imgAlphaSl.Spans() {
				if spanData.Len > 0 {
					colors := imgAlphaAlloc.Allocate(int(spanData.Len))
					sg.Generate(colors, int(spanData.X), y, int(spanData.Len))
					imgAlphaRenBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}
}
