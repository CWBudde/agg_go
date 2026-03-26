// Based on the original AGG examples: lion_lens.cpp.
//
// Renders the lion vector art with a warp-magnifier lens effect.
// Left-click / left-drag moves the lens center.
//
// Note on coordinate systems: the current canvas is y-down. The base
// transform mirrors the lion in X (ScaleXY(-1,1)) to reproduce the same
// visual as C++ rotate(Pi)+flip_y=true.
package main

import (
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
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	lionLensScale        = 3.0
	lionLensRadius       = 70.0
	lionLensX, lionLensY float64
	lionLensInitialized  bool
)

// lionLensTransformer applies an affine transform followed by a warp-magnifier
// lens transform. This composes two sequential transforms for conv.ConvTransform.
type lionLensTransformer struct {
	mtx  *transform.TransAffine
	lens *transform.TransWarpMagnifier
}

func (t *lionLensTransformer) Transform(x, y *float64) {
	t.mtx.Transform(x, y)
	t.lens.Transform(x, y)
}

func initLionLensDemo() {
	if lionLensInitialized {
		return
	}

	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	lionLensX = float64(width) * 0.5
	lionLensY = float64(height) * 0.5

	lionLensInitialized = true
}

func drawLionLensDemo() {
	initLionLensDemo()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Set up the lens.
	lens := transform.NewTransWarpMagnifier()
	lens.SetCenter(lionLensX, lionLensY)
	lens.SetMagnification(lionLensScale)
	lens.SetRadius(lionLensRadius / lionLensScale)

	// Set up the base transformation for the lion.
	g_x1, g_y1, g_x2, g_y2 := getLionBoundingRect(lionData)
	base_dx := (g_x2 - g_x1) * 0.5
	base_dy := (g_y2 - g_y1) * 0.5

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-base_dx, -base_dy))
	// ScaleXY(-1, 1) mirrors X to reproduce the same visual as C++
	// rotate(Pi) + flip_y=true.
	mtx.Multiply(transform.NewTransAffineScalingXY(-1, 1))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(img.Width())*0.5, float64(img.Height())*0.5))

	combined := &lionLensTransformer{mtx: mtx, lens: lens}

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, combined)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(ren)

	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, lionData, lionData, lionData.NPaths)
}

func setLionLensScale(v float64)  { lionLensScale = v }
func setLionLensRadius(v float64) { lionLensRadius = v }

func handleLionLensMouseDown(x, y float64) bool {
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseMove(x, y float64) bool {
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseUp() {}
