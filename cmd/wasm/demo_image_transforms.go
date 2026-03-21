// Based on the original AGG example: image_transforms.cpp
// Demonstrates different combinations of polygon and image affine transforms
// used to fill a star polygon with a transformed image.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageassets"
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
	imgTransPolygonAngle = 0.0 // degrees, -180..180
	imgTransPolygonScale = 1.0 // 0.1..5.0
	imgTransImageAngle   = 0.0 // degrees, -180..180
	imgTransImageScale   = 1.0 // 0.1..5.0
	imgTransExample      = 0   // 0..6
	imgTransImageCX      = 0.0 // image center (set on init)
	imgTransImageCY      = 0.0
	imgTransPolygonCX    = 0.0 // polygon center (set on init)
	imgTransPolygonCY    = 0.0
	imgTransImageCenterX = 0.0
	imgTransImageCenterY = 0.0
	imgTransImage        *agg.Image

	// Reusable components
	imgTransRbuf        *buffer.RenderingBufferU8
	imgTransPixFmt      *pixfmt.PixFmtRGBA32Pre[color.Linear]
	imgTransRenBase     *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]]
	imgTransAlloc       *span.SpanAllocator[color.RGBA8[color.Linear]]
	imgTransRas         *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	imgTransSl          *scanline.ScanlineU8
	imgTransPath        *path.PathStorageStl
	imgTransInitialized bool
)

type imgTransSpanGenAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
}

func (a *imgTransSpanGenAdapter) Prepare() {}
func (a *imgTransSpanGenAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	a.sg.Generate(colors[:length], x, y)
}

func initImgTransDemo() {
	if imgTransInitialized {
		return
	}
	imgTransRbuf = buffer.NewRenderingBufferU8()
	imgTransPixFmt = pixfmt.NewPixFmtRGBA32PreLinear(imgTransRbuf)
	imgTransRenBase = renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](imgTransPixFmt)
	imgTransAlloc = span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	imgTransRas = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	imgTransSl = scanline.NewScanlineU8()
	imgTransPath = path.NewPathStorageStl()

	imgTransPolygonCX = float64(width) * 0.5
	imgTransPolygonCY = float64(height) * 0.5
	imgTransImageCX = imgTransPolygonCX
	imgTransImageCY = imgTransPolygonCY
	imgTransInitialized = true
}

func drawImgTransStar(cx, cy, frameW, frameH float64) {
	r := frameW
	if frameH < r {
		r = frameH
	}
	r1 := r/3 - 8.0
	r2 := r1 / 1.45
	nr := 14

	const twoPi = 2.0 * math.Pi

	imgTransPath.RemoveAll()
	for i := range nr {
		a := twoPi*float64(i)/float64(nr) - math.Pi*0.5
		dx := math.Cos(a)
		dy := math.Sin(a)
		if i&1 != 0 {
			imgTransPath.LineTo(cx+dx*r1, cy+dy*r1)
		} else {
			if i > 0 {
				imgTransPath.LineTo(cx+dx*r2, cy+dy*r2)
			} else {
				imgTransPath.MoveTo(cx+dx*r2, cy+dy*r2)
			}
		}
	}
	imgTransPath.ClosePolygon(basics.PathFlagsClose)
}

func drawImageTransformsDemo() {
	initImgTransDemo()

	if imgTransImage == nil {
		if src, err := imageassets.Spheres(); err == nil && src != nil {
			imgTransImage = src
		} else {
			imgTransImage = createSpheresImage(400, 300)
		}
		imgTransImageCenterX = float64(imgTransImage.Width()) * 0.5
		imgTransImageCenterY = float64(imgTransImage.Height()) * 0.5
	}
	frameW := float64(imgTransImage.Width())
	frameH := float64(imgTransImage.Height())
	frameOffX := (float64(width) - frameW) * 0.5
	frameOffY := (float64(height) - frameH) * 0.5

	// Attach rendering target
	img := ctx.GetImage()
	imgTransRbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	imgTransRenBase.Attach(imgTransPixFmt)
	ctx.GetAgg2D().ClearAll(agg.White)

	polyAngleRad := imgTransPolygonAngle * math.Pi / 180.0
	imgAngleRad := imgTransImageAngle * math.Pi / 180.0
	if !imgTransIsFinitePositive(imgTransPolygonScale) {
		imgTransPolygonScale = 1.0
	}
	if !imgTransIsFinitePositive(imgTransImageScale) {
		imgTransImageScale = 1.0
	}

	// Build polygon transform
	polyMtx := transform.NewTransAffine()
	polyMtx.Translate(-imgTransPolygonCX, -imgTransPolygonCY)
	polyMtx.Rotate(polyAngleRad)
	polyMtx.Scale(imgTransPolygonScale)
	polyMtx.Translate(imgTransPolygonCX, imgTransPolygonCY)

	// Build image matrix based on example mode
	imageMtx := transform.NewTransAffine()
	switch imgTransExample {
	case 0:
		// Identity — image stays fixed
	case 1:
		imageMtx.Translate(-imgTransImageCenterX, -imgTransImageCenterY)
		imageMtx.Rotate(polyAngleRad)
		imageMtx.Scale(imgTransPolygonScale)
		imageMtx.Translate(imgTransPolygonCX, imgTransPolygonCY)
		imageMtx.Invert()
	case 2:
		imageMtx.Translate(-imgTransImageCenterX, -imgTransImageCenterY)
		imageMtx.Rotate(imgAngleRad)
		imageMtx.Scale(imgTransImageScale)
		imageMtx.Translate(imgTransImageCX, imgTransImageCY)
		imageMtx.Invert()
	case 3:
		imageMtx.Translate(-imgTransImageCenterX, -imgTransImageCenterY)
		imageMtx.Rotate(imgAngleRad)
		imageMtx.Scale(imgTransImageScale)
		imageMtx.Translate(imgTransPolygonCX, imgTransPolygonCY)
		imageMtx.Invert()
	case 4:
		imageMtx.Translate(-imgTransImageCX, -imgTransImageCY)
		imageMtx.Rotate(polyAngleRad)
		imageMtx.Scale(imgTransPolygonScale)
		imageMtx.Translate(imgTransPolygonCX, imgTransPolygonCY)
		imageMtx.Invert()
	case 5:
		imageMtx.Translate(-imgTransImageCenterX, -imgTransImageCenterY)
		imageMtx.Rotate(imgAngleRad)
		imageMtx.Rotate(polyAngleRad)
		imageMtx.Scale(imgTransImageScale)
		imageMtx.Scale(imgTransPolygonScale)
		imageMtx.Translate(imgTransImageCX, imgTransImageCY)
		imageMtx.Invert()
	case 6:
		imageMtx.Translate(-imgTransImageCX, -imgTransImageCY)
		imageMtx.Rotate(imgAngleRad)
		imageMtx.Scale(imgTransImageScale)
		imageMtx.Translate(imgTransImageCX, imgTransImageCY)
		imageMtx.Invert()
	}
	// The original demo window equals source image size. In web canvases larger
	// than the source, map screen coordinates back into that centered frame.
	imageMtx.Translate(-frameOffX, -frameOffY)

	// Span interpolator
	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imageMtx, 8)

	// Image source
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(imgTransImage.Data, imgTransImage.Width(), imgTransImage.Height(), imgTransImage.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{255, 255, 255, 255})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	bgColor := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(src, bgColor, interp)
	adapterSG := &imgTransSpanGenAdapter{sg: sg}

	// Build star polygon with polygon transform applied
	drawImgTransStar(imgTransPolygonCX, imgTransPolygonCY, frameW, frameH)

	// Apply polygon matrix to path vertices manually
	transformed := path.NewPathStorageStl()
	imgTransPath.Rewind(0)
	for {
		vx, vy, rawCmd := imgTransPath.NextVertex()
		cmd := basics.PathCommand(rawCmd)
		if basics.IsStop(cmd) {
			break
		}
		polyMtx.Transform(&vx, &vy)
		switch {
		case basics.IsMoveTo(cmd):
			transformed.MoveTo(vx, vy)
		case basics.IsLineTo(cmd):
			transformed.LineTo(vx, vy)
		case basics.IsEndPoly(cmd):
			transformed.ClosePolygon(basics.PathFlagsClose)
		}
	}

	// Render star with image fill
	imgTransRas.Reset()
	imgTransRas.ClipBox(0, 0, float64(width), float64(height))
	imgTransRas.AddPath(&pathSourceAdapter{ps: transformed}, 0)

	if imgTransRas.RewindScanlines() {
		imgTransSl.Reset(imgTransRas.MinX(), imgTransRas.MaxX())
		for imgTransRas.SweepScanline(imgTransSl) {
			y := imgTransSl.Y()
			for _, spanData := range imgTransSl.Spans() {
				if spanData.Len > 0 {
					colors := imgTransAlloc.Allocate(int(spanData.Len))
					adapterSG.Generate(colors, int(spanData.X), y, int(spanData.Len))
					imgTransRenBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}

	// Draw image center handle (interactive point)
	drawHandle(imgTransImageCX, imgTransImageCY)
}

func imgTransIsFinitePositive(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0.0
}

func handleImgTransMouseDown(x, y float64) bool {
	dist := math.Sqrt((x-imgTransImageCX)*(x-imgTransImageCX) + (y-imgTransImageCY)*(y-imgTransImageCY))
	if dist < 7.0 {
		imgTransImageCX = x
		imgTransImageCY = y
		return true
	}
	return false
}

func handleImgTransMouseMove(x, y float64) bool {
	imgTransImageCX = x
	imgTransImageCY = y
	return true
}

func setImgTransPolygonAngle(v float64) { imgTransPolygonAngle = v }
func setImgTransPolygonScale(v float64) { imgTransPolygonScale = v }
func setImgTransImageAngle(v float64)   { imgTransImageAngle = v }
func setImgTransImageScale(v float64)   { imgTransImageScale = v }
func setImgTransExample(v int)          { imgTransExample = v }
