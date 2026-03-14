// Based on the original AGG example: image_alpha.cpp
// Demonstrates using brightness as an alpha channel: a large ellipse is filled
// with a rotated image where the alpha value of each pixel is derived from the
// pixel's luminance via a configurable lookup table.
package main

import (
	"math"
	"math/rand"

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

// --- Demo state ---

var (
	imgAlphaImage *agg.Image

	// Background ellipses (randomised once)
	imgAlphaEllipses []imgAlphaEllipse

	// Brightness-to-alpha LUT (256*3 entries): alpha = f(r+g+b)
	imgAlphaLUT [256 * 3]uint8

	// Reusable components
	imgAlphaRbuf        *buffer.RenderingBufferU8
	imgAlphaPixFmt      *pixfmt.PixFmtRGBA32Pre[color.Linear]
	imgAlphaRenBase     *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]]
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

func initImgAlphaDemo() {
	if imgAlphaInitialized {
		return
	}
	imgAlphaRbuf = buffer.NewRenderingBufferU8()
	imgAlphaPixFmt = pixfmt.NewPixFmtRGBA32PreLinear(imgAlphaRbuf)
	imgAlphaRenBase = renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](imgAlphaPixFmt)
	imgAlphaAlloc = span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	imgAlphaRas = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	imgAlphaSl = scanline.NewScanlineU8()
	imgAlphaPath = path.NewPathStorageStl()

	// Build background ellipses (same seed each run for reproducibility)
	rng := rand.New(rand.NewSource(42))
	imgAlphaEllipses = make([]imgAlphaEllipse, 50)
	for i := range imgAlphaEllipses {
		imgAlphaEllipses[i] = imgAlphaEllipse{
			x:  float64(rng.Intn(width)),
			y:  float64(rng.Intn(height)),
			rx: float64(rng.Intn(60) + 10),
			ry: float64(rng.Intn(60) + 10),
			r:  uint8(rng.Intn(256)),
			g:  uint8(rng.Intn(256)),
			b:  uint8(rng.Intn(256)),
			a:  uint8(rng.Intn(256)),
		}
	}

	// Default brightness→alpha LUT: same as C++ defaults (control points 1,1,1,0.5,0.5,1)
	// Approximation: linear fade in the middle region.
	buildImgAlphaLUT([]float64{1.0, 1.0, 1.0, 0.5, 0.5, 1.0})

	imgAlphaInitialized = true
}

// buildImgAlphaLUT builds the brightness→alpha LUT from 6 control values (spline approximation).
func buildImgAlphaLUT(ctrlValues []float64) {
	n := 256 * 3
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1) // 0..1
		// Simple piecewise linear interpolation of the control points
		nc := len(ctrlValues)
		seg := t * float64(nc-1)
		lo := int(seg)
		if lo >= nc-1 {
			lo = nc - 2
		}
		frac := seg - float64(lo)
		v := ctrlValues[lo]*(1-frac) + ctrlValues[lo+1]*frac
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		imgAlphaLUT[i] = uint8(v * 255)
	}
}

func drawImageAlphaDemo() {
	initImgAlphaDemo()

	if imgAlphaImage == nil {
		imgAlphaImage = createSpheresImage(400, 400)
	}

	imgW := float64(imgAlphaImage.Width())
	imgH := float64(imgAlphaImage.Height())

	// Attach rendering target
	img := ctx.GetImage()
	imgAlphaRbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)

	// Render background ellipses using the public API
	ctx.GetAgg2D().ResetTransformations()
	const oneTwoHundredFiftyFifth = 1.0 / 255.0
	for _, e := range imgAlphaEllipses {
		ctx.SetColor(agg.RGBA(float64(e.r)*oneTwoHundredFiftyFifth, float64(e.g)*oneTwoHundredFiftyFifth, float64(e.b)*oneTwoHundredFiftyFifth, float64(e.a)*oneTwoHundredFiftyFifth))
		ctx.FillEllipse(e.x, e.y, e.rx, e.ry)
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
	imgRbuf.Attach(imgAlphaImage.Data, imgAlphaImage.Width(), imgAlphaImage.Height(), imgAlphaImage.Width()*4)
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
		for imgAlphaRas.SweepScanline(&rasScanlineAdapter{sl: imgAlphaSl}) {
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
