// Port of AGG C++ image_transforms.cpp – image fill inside a star polygon.
//
// Renders a star polygon filled with an affine-transformed image (spheres).
// The polygon and image transforms are independent. Default: no rotation,
// scale=1.0, example=0 (polygon fixed, image fixed).
package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
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

const (
	canvasW = 800
	canvasH = 600
)

// imagePixFmt wraps a RenderingBufferU8 and implements image.PixelFormat.
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

type imageClipSource struct {
	accessor *image.ImageAccessorClip[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *imageClipSource) Width() int                  { return s.ipf.Width() }
func (s *imageClipSource) Height() int                 { return s.ipf.Height() }
func (s *imageClipSource) ColorType() string           { return "RGBA8" }
func (s *imageClipSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *imageClipSource) Span(x, y, l int) []basics.Int8u {
	return s.accessor.Span(x, y, l)
}
func (s *imageClipSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *imageClipSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *imageClipSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

type spanGenAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
}

func (a *spanGenAdapter) Prepare() {}
func (a *spanGenAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	a.sg.Generate(colors[:length], x, y)
}

type rasScanlineAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rasScanlineAdapter) ResetSpans()                { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, c uint32)    { a.sl.AddCell(x, uint(c)) }
func (a *rasScanlineAdapter) AddSpan(x, l int, c uint32) { a.sl.AddSpan(x, l, uint(c)) }
func (a *rasScanlineAdapter) Finalize(y int)             { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int              { return a.sl.NumSpans() }

type pathSourceAdapter struct{ ps *path.PathStorageStl }

func (a *pathSourceAdapter) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

func createSpheresIT(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)
	imgCtx.SetColor(agg.RGBA(0.05, 0.05, 0.12, 1.0))
	imgCtx.FillRectangle(0, 0, float64(w), float64(h))
	type sphere struct{ x, y, r, r0, g0, b0 float64 }
	spheres := []sphere{
		{float64(w) * 0.25, float64(h) * 0.30, float64(w) * 0.18, 0.9, 0.2, 0.1},
		{float64(w) * 0.65, float64(h) * 0.30, float64(w) * 0.15, 0.1, 0.4, 0.9},
		{float64(w) * 0.45, float64(h) * 0.70, float64(w) * 0.20, 0.1, 0.8, 0.3},
	}
	for _, sp := range spheres {
		imgCtx.SetColor(agg.RGBA(sp.r0, sp.g0, sp.b0, 0.9))
		imgCtx.FillCircle(sp.x, sp.y, sp.r)
		imgCtx.SetColor(agg.RGBA(1.0, 1.0, 1.0, 0.6))
		imgCtx.FillCircle(sp.x-sp.r*0.30, sp.y-sp.r*0.30, sp.r*0.30)
	}
	return img
}

func buildStarIT(cx, cy float64, w, h int) *path.PathStorageStl {
	r := float64(w)
	if float64(h) < r {
		r = float64(h)
	}
	r1 := r/3 - 8
	r2 := r1 / 1.45
	nr := 14

	ps := path.NewPathStorageStl()
	for i := 0; i < nr; i++ {
		a := math.Pi*2.0*float64(i)/float64(nr) - math.Pi/2.0
		dx := math.Cos(a)
		dy := math.Sin(a)
		if i&1 != 0 {
			ps.LineTo(cx+dx*r1, cy+dy*r1)
		} else {
			if i == 0 {
				ps.MoveTo(cx+dx*r2, cy+dy*r2)
			} else {
				ps.LineTo(cx+dx*r2, cy+dy*r2)
			}
		}
	}
	ps.ClosePolygon(basics.PathFlagsNone)
	return ps
}

func main() {
	srcImg := createSpheresIT(400, 400)
	imgW := float64(srcImg.Width())
	imgH := float64(srcImg.Height())

	ctx := agg.NewContext(canvasW, canvasH)
	ctx.Clear(agg.RGBA(0.1, 0.1, 0.1, 1.0))

	dstImg := ctx.GetImage()
	dstRbuf := buffer.NewRenderingBufferWithData[uint8](dstImg.Data, dstImg.Width(), dstImg.Height(), dstImg.Width()*4)
	dstPixf := pixfmt.NewPixFmtRGBA32Pre[color.Linear](dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	cx, cy := float64(canvasW)/2, float64(canvasH)/2

	// Build affine transform: image → polygon center, then invert for sampling.
	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-imgW/2, -imgH/2)
	imgMtx.Rotate(0)
	imgMtx.Translate(cx, cy)
	imgMtx.Invert()

	// Image source.
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(srcImg.Data, srcImg.Width(), srcImg.Height(), srcImg.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{0, 0, 0, 0})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	bgRGBA := color.RGBA8[color.Linear]{}
	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(src, bgRGBA, interp)
	sgAdp := &spanGenAdapter{sg: sg}

	// Rasterize the star polygon.
	ps := buildStarIT(cx, cy, canvasW, canvasH)
	ras.Reset()
	ras.ClipBox(0, 0, float64(canvasW), float64(canvasH))
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					colors := alloc.Allocate(int(spanData.Len))
					sgAdp.Generate(colors, int(spanData.X), y, int(spanData.Len))
					renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}

	const filename = "image_transforms.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
