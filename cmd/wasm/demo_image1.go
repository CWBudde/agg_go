// Based on the original AGG example: image1.cpp
// Demonstrates affine image transformation (rotation + scale) applied as a fill
// inside a large ellipse.
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
	"agg_go/internal/scanline"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// --- Demo state ---

var (
	img1Angle = 0.0 // degrees, -180..180
	img1Scale = 1.0 // 0.1..5.0
	img1Image *agg.Image

	// Reusable components
	img1Rbuf        *buffer.RenderingBufferU8
	img1PixFmt      *pixfmt.PixFmtRGBA32Pre[color.Linear]
	img1RenBase     *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]]
	img1Alloc       *span.SpanAllocator[color.RGBA8[color.Linear]]
	img1Ras         *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	img1Sl          *scanline.ScanlineU8
	img1Path        *path.PathStorageStl
	img1Initialized bool
)

func initImg1Demo() {
	if img1Initialized {
		return
	}
	img1Rbuf = buffer.NewRenderingBufferU8()
	img1PixFmt = pixfmt.NewPixFmtRGBA32PreLinear(img1Rbuf)
	img1RenBase = renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](img1PixFmt)
	img1Alloc = span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	img1Ras = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	img1Sl = scanline.NewScanlineU8()
	img1Path = path.NewPathStorageStl()
	img1Initialized = true
}

// img1SpanGenAdapter bridges the span generator signature for the render loop.
type img1SpanGenAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
}

func (a *img1SpanGenAdapter) Prepare() {}
func (a *img1SpanGenAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	a.sg.Generate(colors[:length], x, y)
}

func drawImage1Demo() {
	initImg1Demo()

	if img1Image == nil {
		img1Image = createSpheresImage(400, 400)
	}

	imgW := float64(img1Image.Width())
	imgH := float64(img1Image.Height())
	cx := float64(width) / 2.0
	cy := float64(height) / 2.0

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Attach rendering target
	img := ctx.GetImage()
	img1Rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	img1RenBase.Attach(img1PixFmt)

	// Image transform: translate to center, rotate, scale, then translate to screen center
	// Then invert so we can map screen -> image coords.
	angleRad := img1Angle * math.Pi / 180.0
	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-imgW/2-10, -imgH/2-20-10)
	imgMtx.Rotate(angleRad)
	imgMtx.Scale(img1Scale)
	imgMtx.Translate(cx, cy+20)
	imgMtx.Invert()

	// Polygon transform for the ellipse (same rotation/scale, not inverted)
	polyMtx := transform.NewTransAffine()
	polyMtx.Translate(-imgW/2+10, -imgH/2+20+10)
	polyMtx.Rotate(angleRad)
	polyMtx.Scale(img1Scale)
	polyMtx.Translate(cx, cy+20)

	// Span interpolator over the image matrix
	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)

	// Build image source
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(img1Image.Data, img1Image.Width(), img1Image.Height(), img1Image.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{0, 100, 0, 128})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	// Span generator
	bgColor := color.RGBA8[color.Linear]{R: 0, G: 100, B: 0, A: 128}
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(src, bgColor, interp)
	adapterSG := &img1SpanGenAdapter{sg: sg}

	// Ellipse path (no transformation - we'll apply polyMtx manually)
	r := imgW
	if imgH-60 < r {
		r = imgH - 60
	}

	img1Path.RemoveAll()
	numPoints := 200
	ellCx := imgW/2.0 + 10
	ellCy := imgH/2.0 + 20 + 10
	ellRx := r/2.0 + 16.0
	ellRy := r/2.0 + 16.0
	for i := 0; i < numPoints; i++ {
		a := 2.0 * math.Pi * float64(i) / float64(numPoints)
		px := ellCx + ellRx*math.Cos(a)
		py := ellCy + ellRy*math.Sin(a)
		// Apply polygon transform
		polyMtx.Transform(&px, &py)
		if i == 0 {
			img1Path.MoveTo(px, py)
		} else {
			img1Path.LineTo(px, py)
		}
	}
	img1Path.ClosePolygon(basics.PathFlagsClose)

	// Render
	img1Ras.Reset()
	img1Ras.ClipBox(0, 0, float64(width), float64(height))
	img1Ras.AddPath(&pathSourceAdapter{ps: img1Path}, 0)

	if img1Ras.RewindScanlines() {
		img1Sl.Reset(img1Ras.MinX(), img1Ras.MaxX())
		for img1Ras.SweepScanline(&rasScanlineAdapter{sl: img1Sl}) {
			y := img1Sl.Y()
			for _, spanData := range img1Sl.Spans() {
				if spanData.Len > 0 {
					colors := img1Alloc.Allocate(int(spanData.Len))
					adapterSG.Generate(colors, int(spanData.X), y, int(spanData.Len))
					img1RenBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}
}
