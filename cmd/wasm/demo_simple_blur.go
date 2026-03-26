// Based on the original AGG examples: simple_blur.cpp.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
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
	drawSimpleBlurLionOutline(ras, sl, renSolid, img.Width(), img.Height())

	// 4. Snapshot the scene before the ellipse outline is drawn so the blur
	//    samples the clean lion pixels.
	bgImg := agg.NewImage(make([]uint8, len(img.Data)), img.Width(), img.Height(), img.Stride())
	copy(bgImg.Data, img.Data)

	// 5. Draw ellipse outline over the lion.
	rx, ry := 100.0, 100.0
	ellipse := shapes.NewEllipseWithParams(simpleBlurCX, simpleBlurCY, rx, ry, 100, false)
	ellConv := &ellipseConvVS{ell: ellipse}
	stroke := conv.NewConvStroke(ellConv)
	stroke.SetWidth(2.0)
	strokeVS := conv.NewRasterizerVertexSourceAdapter(stroke)
	ras.Reset()
	ras.AddPath(strokeVS, 0)
	renSolid.SetColor(color.RGBA8[color.Linear]{R: 0, G: 51, B: 0, A: 255})
	renscan.RenderScanlines(ras, sl, renSolid)

	// 6. Apply 3×3 box-blur inside the ellipse using the pre-outline snapshot.
	applyBlurInsideEllipseSimple(img, bgImg, simpleBlurCX, simpleBlurCY, rx, ry)
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

// drawSimpleBlurLionOutline renders the lion paths as outlines into the right half of the canvas.
func drawSimpleBlurLionOutline(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineP8,
	renSolid *renscan.RendererScanlineAASolid[*renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	w, h int,
) {
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
	stroke := conv.NewConvStroke(transVS)
	stroke.SetWidth(1.0)
	strokeVS := conv.NewRasterizerVertexSourceAdapter(stroke)

	colors := lionSimpleBlurColorView{data: lionData}
	renscan.RenderAllPaths(ras, sl, renSolid, strokeVS, colors, colors, lionData.NPaths)
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

// applyBlurInsideEllipseSimple performs a 3×3 box-blur on dst for all pixels
// inside the ellipse defined by (cx, cy, rx, ry), sampling from src.
// It uses img.Stride() rather than a hardcoded stride to support any buffer layout.
func applyBlurInsideEllipseSimple(dst, src *agg.Image, cx, cy, rx, ry float64) {
	w, h := dst.Width(), dst.Height()
	dstData := dst.Data
	srcData := src.Data
	dstStride := dst.Stride()
	srcStride := src.Stride()
	dstBase := 0
	srcBase := 0
	if dstStride < 0 {
		dstBase = (h - 1) * -dstStride
	}
	if srcStride < 0 {
		srcBase = (h - 1) * -srcStride
	}

	rx2 := rx * rx
	ry2 := ry * ry

	for y := 0; y < h; y++ {
		dy := float64(y) - cy
		dy2 := dy * dy
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			if dx*dx/rx2+dy2/ry2 > 1.0 {
				continue // outside ellipse
			}
			if x == 0 || x == w-1 || y == 0 || y == h-1 {
				continue // skip border pixels
			}
			var r, g, b, a uint32
			for iy := -1; iy <= 1; iy++ {
				rowOff := srcBase + (y+iy)*srcStride
				for ix := -1; ix <= 1; ix++ {
					idx := rowOff + (x+ix)*4
					r += uint32(srcData[idx])
					g += uint32(srcData[idx+1])
					b += uint32(srcData[idx+2])
					a += uint32(srcData[idx+3])
				}
			}
			dstIdx := dstBase + y*dstStride + x*4
			dstData[dstIdx] = uint8(r / 9)
			dstData[dstIdx+1] = uint8(g / 9)
			dstData[dstIdx+2] = uint8(b / 9)
			dstData[dstIdx+3] = uint8(a / 9)
		}
	}
}
