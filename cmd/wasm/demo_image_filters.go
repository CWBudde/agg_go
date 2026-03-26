// Based on the original AGG example: image_filters.cpp.
package main

import (
	"fmt"
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	imgacc "github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	imgFilterAngle  = 0.0
	imgFilterType   = agg.FilterBilinear
	imgFilterRadius = 4.0
	testImage       *agg.Image
)

// imgFilterSpanGenInterface is the subset of span generators used in this demo.
type imgFilterSpanGenInterface interface {
	Generate(span []color.RGBA8[color.Linear], x, y int)
}

// imgFilterSpanGenAdapter adapts imgFilterSpanGenInterface to the renderer span callback.
type imgFilterSpanGenAdapter struct {
	gen imgFilterSpanGenInterface
}

func (a *imgFilterSpanGenAdapter) Prepare() {}

func (a *imgFilterSpanGenAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	if length <= 0 {
		return
	}
	a.gen.Generate(colors[:length], x, y)
}

// imgFilterBuildLUT builds an ImageFilterLUT from the current imgFilterType / imgFilterRadius.
func imgFilterBuildLUT() *imgacc.ImageFilterLUT {
	normalize := true
	r := imgFilterRadius
	switch imgFilterType {
	case agg.FilterHanning:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.HanningFilter{}, normalize)
	case agg.FilterHermite:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.HermiteFilter{}, normalize)
	case agg.FilterQuadric:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.QuadricFilter{}, normalize)
	case agg.FilterBicubic:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.BicubicFilter{}, normalize)
	case agg.FilterCatrom:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.CatromFilter{}, normalize)
	case agg.FilterSpline16:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.Spline16Filter{}, normalize)
	case agg.FilterSpline36:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.Spline36Filter{}, normalize)
	case agg.FilterBlackman:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.NewBlackmanFilter(r), normalize)
	case agg.FilterHamming:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.HammingFilter{}, normalize)
	case agg.FilterGaussian:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.GaussianFilter{}, normalize)
	case agg.FilterBessel:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.BesselFilter{}, normalize)
	case agg.FilterMitchell:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.NewMitchellFilter(1.0/3.0, 1.0/3.0), normalize)
	case agg.FilterSinc:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.NewSincFilter(r), normalize)
	case agg.FilterLanczos:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.NewLanczosFilter(r), normalize)
	default:
		return imgacc.NewImageFilterLUTWithFilter(imgacc.BilinearFilter{}, normalize)
	}
}

// imgFilterBuildSpanGen builds the span generator for the given source and interpolator.
func imgFilterBuildSpanGen(
	source *imageClipSource,
	interp *span.SpanInterpolatorLinear[*transform.TransAffine],
) imgFilterSpanGenInterface {
	switch imgFilterType {
	case agg.FilterNoFilter:
		return span.NewSpanImageFilterRGBANNWithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](source, interp)
	case agg.FilterBilinear:
		return span.NewSpanImageFilterRGBABilinearClipWithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			color.RGBA8[color.Linear]{},
			interp,
		)
	case agg.FilterHanning, agg.FilterHamming, agg.FilterHermite:
		return span.NewSpanImageFilterRGBA2x2WithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interp,
			imgFilterBuildLUT(),
		)
	default:
		return span.NewSpanImageFilterRGBAWithParams[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source,
			interp,
			imgFilterBuildLUT(),
		)
	}
}

func drawImageFiltersDemo() {
	if testImage == nil {
		testImage = createTestImage(200, 200)
	}

	logStatus(fmt.Sprintf("Filter: %d, Radius: %.2f, Angle: %.1f", imgFilterType, imgFilterRadius, imgFilterAngle))

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Build the source image rendering buffer
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(testImage.Data, testImage.Width(), testImage.Height(), testImage.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := imgacc.NewImageAccessorClip(&ipf, []basics.Int8u{0, 0, 0, 0})
	source := &imageClipSource{accessor: accessor, ipf: &ipf}

	srcW := float64(testImage.Width())
	srcH := float64(testImage.Height())
	angleRad := imgFilterAngle * math.Pi / 180.0

	// renderWithMtx renders the source image through imgMtx (screen→source) clipped to clipPs.
	renderWithMtx := func(imgMtx *transform.TransAffine, clipPs *path.PathStorageStl) {
		interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
		sg := imgFilterBuildSpanGen(source, interp)
		sgAdp := &imgFilterSpanGenAdapter{gen: sg}

		ras.Reset()
		ras.ClipBox(0, 0, float64(img.Width()), float64(img.Height()))
		ras.AddPath(&pathSourceAdapter{ps: clipPs}, 0)

		if ras.RewindScanlines() {
			sl.Reset(ras.MinX(), ras.MaxX())
			for ras.SweepScanline(sl) {
				y := sl.Y()
				for _, spanData := range sl.Spans() {
					if spanData.Len <= 0 {
						continue
					}
					colors := alloc.Allocate(int(spanData.Len))
					sgAdp.Generate(colors, int(spanData.X), y, int(spanData.Len))
					ren.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}

	// Draw original image for reference (no rotation, at top-left region)
	{
		x0, y0 := 50.0, 50.0
		// imgMtx maps screen → source: translate so that screen(x0,y0) = source(0,0)
		imgMtx := transform.NewTransAffine()
		imgMtx.Translate(-x0, -y0)
		clipPs := path.NewPathStorageStl()
		clipPs.MoveTo(x0, y0)
		clipPs.LineTo(x0+srcW, y0)
		clipPs.LineTo(x0+srcW, y0+srcH)
		clipPs.LineTo(x0, y0+srcH)
		clipPs.ClosePolygon(basics.PathFlagsNone)
		renderWithMtx(imgMtx, clipPs)
	}

	// Draw 3 rotated/scaled versions
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	scales := []float64{0.5, 1.0, 2.0}
	for i, s := range scales {
		cx := 350.0 + float64(i)*150.0
		cy := 150.0
		halfW := (srcW * s) * 0.5
		halfH := (srcH * s) * 0.5

		// imgMtx: screen → source
		// Forward mapping: source center → rotate by angle → scale by s → screen center (cx,cy)
		// Inverse: screen → translate(-cx,-cy) → rotate(-angle) → scale(1/s) → translate(srcW/2,srcH/2)
		imgMtx := transform.NewTransAffine()
		imgMtx.Translate(-cx, -cy)
		imgMtx.Rotate(-angleRad)
		imgMtx.Scale(1.0 / s)
		imgMtx.Translate(srcW/2.0, srcH/2.0)

		// Clip polygon: the parallelogram corners
		tlX := cx - halfW*cosA + halfH*sinA
		tlY := cy - halfW*sinA - halfH*cosA
		trX := cx + halfW*cosA + halfH*sinA
		trY := cy + halfW*sinA - halfH*cosA
		brX := cx + halfW*cosA - halfH*sinA
		brY := cy + halfW*sinA + halfH*cosA
		blX := cx - halfW*cosA - halfH*sinA
		blY := cy - halfW*sinA + halfH*cosA

		clipPs := path.NewPathStorageStl()
		clipPs.MoveTo(tlX, tlY)
		clipPs.LineTo(trX, trY)
		clipPs.LineTo(brX, brY)
		clipPs.LineTo(blX, blY)
		clipPs.ClosePolygon(basics.PathFlagsNone)
		renderWithMtx(imgMtx, clipPs)
	}
}

func createTestImage(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)

	imgCtx.Clear(agg.White)

	// Draw a grid
	imgCtx.SetColor(agg.RGBA(0.8, 0.8, 0.8, 1.0))
	for i := 0; i < w; i += 20 {
		imgCtx.DrawLine(float64(i), 0, float64(i), float64(h))
	}
	for i := 0; i < h; i += 20 {
		imgCtx.DrawLine(0, float64(i), float64(w), float64(i))
	}

	// Draw some shapes
	imgCtx.SetColor(agg.Red)
	imgCtx.FillCircle(float64(w)/2, float64(h)/2, float64(w)/4)

	imgCtx.SetColor(agg.Blue)
	imgCtx.SetStrokeWidth(5.0)
	imgCtx.DrawRectangle(10, 10, float64(w-20), float64(h-20))

	// Draw high-frequency pattern (diagonal lines)
	imgCtx.SetColor(agg.Black)
	imgCtx.SetStrokeWidth(1.0)
	for i := -w; i < w; i += 4 {
		imgCtx.DrawLine(float64(i), 0, float64(i+w), float64(h))
	}

	return img
}
