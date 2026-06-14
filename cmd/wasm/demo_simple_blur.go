// Based on the original AGG examples: simple_blur.cpp.
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	liondemo "github.com/cwbudde/agg_go/internal/demo/lion"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	outline "github.com/cwbudde/agg_go/internal/renderer/outline"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

var (
	simpleBlurCX = 400.0
	simpleBlurCY = 300.0
)

func drawSimpleBlurDemo() {
	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt(pf)

	// 1. Clear background to white.
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(ren)

	// 2. Draw lion fill (left half of canvas).
	drawSimpleBlurLionFill(ras, sl, renSolid, img.Width(), img.Height())

	// 3. Draw lion outline (right half of canvas).
	drawSimpleBlurLionOutline(img, img.Width(), img.Height())

	// 4. Match C++ simple_blur.cpp: stroke the ellipse, then stroke that stroke.
	rx, ry := 100.0, 100.0
	ellipse := shapes.NewEllipseWithParams(simpleBlurCX, simpleBlurCY, rx, ry, 100, false)
	ellStroke1 := conv.NewConvStroke(&ellipseConvVS{ell: ellipse})
	ellStroke1.SetWidth(6.0)
	ellStroke2 := conv.NewConvStroke(ellStroke1)
	ellStroke2.SetWidth(2.0)
	ras.Reset()
	ras.AddPath(conv.NewRasterizerVertexSourceAdapter(ellStroke2), 0)
	renSolid.SetColor(color.RGBA8[color.Linear]{R: 0, G: 51, B: 0, A: 255})
	renscan.RenderScanlines(ras, sl, renSolid)

	// 5. Snapshot after the boundary is drawn, like C++ copy_window_to_img(0).
	bgImg := agg.NewImage(make([]uint8, len(img.Data)), img.Width(), img.Height(), img.Stride())
	copy(bgImg.Data, img.Data)

	// 6. Apply 3x3 box-blur inside the ellipse through AA rasterized coverage.
	applyBlurInsideEllipseSimple(img, bgImg, simpleBlurCX, simpleBlurCY, rx, ry)

	applyPremulLinearToSRGB(img)
}

// drawSimpleBlurLionFill renders the lion fill paths into the left half of the canvas.
func drawSimpleBlurLionFill(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineP8,
	renSolid *renscan.RendererScanlineAASolid[*renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	w, h int,
) {
	x1, y1, x2, y2 := getLionBoundingRect(lionData)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	// In C++ the transform includes rotation(pi) which flips both X and Y.
	// Combined with the FlipY=true rendering buffer the Y-flip cancels, leaving
	// only an X-mirror.  The WASM canvas uses Y-down (no FlipY), so we replace
	// the 180° rotation with an X-flip-only scale to get the same visual result.
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-cx, -cy))
	mtx.Multiply(transform.NewTransAffineScalingXY(-1, 1))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(w)*0.25, float64(h)*0.5))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)

	colors := lionSimpleBlurColorView{data: lionData}
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, colors, colors, lionData.NPaths)
}

// drawSimpleBlurLionOutline renders the lion paths as anti-aliased outlines into
// the right half of the canvas. It mirrors the C++ pipeline faithfully:
// rasterizer_outline_aa + line_profile_aa with round caps (not conv_stroke).
func drawSimpleBlurLionOutline(img *agg.Image, w, h int) {
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)

	x1, y1, x2, y2 := getLionBoundingRect(lionData)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	// Same X-flip-only approach as drawSimpleBlurLionFill (see comment there).
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-cx, -cy))
	mtx.Multiply(transform.NewTransAffineScalingXY(-1, 1))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(w)*0.75, float64(h)*0.5))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	outlineVS := conv.NewRasterizerVertexSourceAdapter(transVS)

	// C++: line_profile_aa profile; profile.width(1.0);
	profile := outline.NewLineProfileAA()
	profile.Width(1.0)

	outlineBase := &loOutlineBaseAdapter{rb: rb}
	renOutline := outline.NewRendererOutlineAA[*loOutlineBaseAdapter, color.RGBA8[color.Linear]](outlineBase, profile)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*loOutlineAAAdapter, color.RGBA8[color.Linear]](&loOutlineAAAdapter{ren: renOutline})
	rasOutline.SetRoundCap(true) // C++: ras.round_cap(true)

	rasOutline.RenderAllPaths(outlineVS, loLionColorView{data: lionData}, lionData, lionData.NPaths)
}

// lionSimpleBlurColorView provides per-path fill colors for RenderAllPaths.
type lionSimpleBlurColorView struct {
	data *liondemo.LionData
}

func (v lionSimpleBlurColorView) GetColor(index int) color.RGBA8[color.Linear] {
	return v.data.Colors[index]
}

func (v lionSimpleBlurColorView) GetPathID(index int) uint32 {
	return uint32(v.data.PathIdx[index])
}

// applyBlurInsideEllipseSimple performs the C++ span_simple_blur_rgb24 operation
// through an anti-aliased ellipse rasterizer, sampling from src.
func applyBlurInsideEllipseSimple(dst, src *agg.Image, cx, cy, rx, ry float64) {
	dstRbuf := buffer.NewRenderingBufferU8()
	dstRbuf.Attach(dst.Data, dst.Width(), dst.Height(), dst.Stride())
	srcRbuf := buffer.NewRenderingBufferU8()
	srcRbuf.Attach(src.Data, src.Width(), src.Height(), src.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(dstRbuf)
	ren := renderer.NewRendererBaseWithPixfmt(pf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	gen := &simpleBlurSpanGenerator{src: srcRbuf}

	ellipse := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
	ras := newSimpleBlurRasterizer()
	ras.AddPath(&ellipseVS{ell: ellipse}, 0)
	renscan.RenderScanlinesAA(ras, scanline.NewScanlineU8(), ren, alloc, gen)
}

type simpleBlurRasterizer = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newSimpleBlurRasterizer() *simpleBlurRasterizer {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
}

type simpleBlurSpanGenerator struct {
	src *buffer.RenderingBufferU8
}

func (g *simpleBlurSpanGenerator) Prepare() {}

func (g *simpleBlurSpanGenerator) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	w, h := g.src.Width(), g.src.Height()
	if y < 1 || y >= h-1 {
		return
	}

	for i := 0; i < length; i++ {
		if x > 0 && x < w-1 {
			var r, gg, b uint32
			for iy := -1; iy <= 1; iy++ {
				row := g.src.Row(y + iy)
				off := (x - 1) * 4
				for ix := 0; ix < 3; ix++ {
					r += uint32(row[off])
					gg += uint32(row[off+1])
					b += uint32(row[off+2])
					off += 4
				}
			}
			colors[i] = color.RGBA8[color.Linear]{R: uint8(r / 9), G: uint8(gg / 9), B: uint8(b / 9), A: 255}
		} else {
			colors[i] = color.RGBA8[color.Linear]{A: 255}
		}
		x++
	}
}
