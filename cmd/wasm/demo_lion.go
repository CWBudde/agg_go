// Port of AGG C++ lion.cpp – classic lion demo with alpha, rotate/scale, skew.
//
// Left-drag rotates and scales; right-drag applies shear.
// An alpha slider controls global opacity of all paths.
//
// Note on coordinate systems: AGG's original example uses flip_y=true (y-up
// rendering). In Go's y-down canvas, rotate(angle+Pi)+flip_y is replaced by
// Scale(-1,1)+Rotate(angle), which produces the same visual result.
// Centering uses the actual bounding-box centre, not just the half-extents.
package main

import (
	"math"

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
	lionFillAlpha         = 1.0
	lionFillAngle         = 0.0
	lionFillScale         = 1.0
	lionFillSkewX         = 0.0
	lionFillSkewY         = 0.0
	lionFillDragging      = false
	lionFillRightDragging = false
)

// lionAlphaColorView wraps LionData and overrides the alpha channel.
type lionAlphaColorView struct {
	data  *liondemo.LionData
	alpha uint8
}

func (v lionAlphaColorView) GetColor(index int) color.RGBA8[color.Linear] {
	c := v.data.Colors[index]
	c.A = v.alpha
	return c
}

func (v lionAlphaColorView) GetPathID(index int) uint32 {
	return uint32(v.data.PathIdx[index])
}

func drawLionDemo() {
	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt(pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	// Compute the true bounding-box centre so the lion rotates around its own
	// centre and is centred on the canvas in the default (angle=0) state.
	x1, y1, x2, y2 := getLionBoundingRect(lionData)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	// Transform chain (mirrors the C++ compose order):
	//   translate(-cx, -cy)   – move lion centre to origin
	//   scale(s, s)           – uniform scale
	//   scale(-1, 1)          – x-mirror: equivalent to C++ rotate(Pi)+flip_y
	//   rotate(angle)         – interactive rotation
	//   skew(sx/1000, sy/1000)
	//   translate(W/2, H/2)   – move to canvas centre
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-cx, -cy))
	mtx.Multiply(transform.NewTransAffineScalingXY(lionFillScale, lionFillScale))
	mtx.Multiply(transform.NewTransAffineScalingXY(-1, 1))
	mtx.Multiply(transform.NewTransAffineRotation(lionFillAngle))
	mtx.Multiply(transform.NewTransAffineSkewing(lionFillSkewX/1000.0, lionFillSkewY/1000.0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(img.Width())*0.5, float64(img.Height())*0.5))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(ren)

	colors := lionAlphaColorView{
		data:  lionData,
		alpha: uint8(lionFillAlpha * 255),
	}
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, colors, colors, lionData.NPaths)

	applyLinearToSRGB(img)
}

// --- Mouse handlers ---

func handleLionMouseDown(x, y float64, right bool) bool {
	if right {
		lionFillRightDragging = true
		lionFillSkewX = x
		lionFillSkewY = y
	} else {
		lionFillDragging = true
		applyLionFillTransform(x, y)
	}
	return true
}

func handleLionMouseMove(x, y float64, right bool) bool {
	if right && lionFillRightDragging {
		lionFillSkewX = x
		lionFillSkewY = y
		return true
	}
	if lionFillDragging {
		applyLionFillTransform(x, y)
		return true
	}
	return false
}

func handleLionMouseUp() {
	lionFillDragging = false
	lionFillRightDragging = false
}

func applyLionFillTransform(x, y float64) {
	dx := x - float64(width)*0.5
	dy := y - float64(height)*0.5
	lionFillAngle = math.Atan2(dy, dx)
	lionFillScale = math.Sqrt(dx*dx+dy*dy) / 100.0
	if lionFillScale < 0.01 {
		lionFillScale = 0.01
	}
}
