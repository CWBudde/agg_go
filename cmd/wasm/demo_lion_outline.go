// Port of AGG C++ lion_outline.cpp – interactive lion with outline rendering.
//
// The lion vector art is rendered as stroked outlines rather than filled
// polygons. Left-drag rotates and scales; right-drag applies shear.
// A slider controls outline width.
//
// Note on coordinate systems: AGG's original uses flip_y=true (y-up rendering).
// In Go's y-down canvas, rotate(angle+Pi)+flip_y is replaced by
// Scale(-1,1)+Rotate(angle). Centering uses the actual bounding-box centre.
package main

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	liondemo "github.com/cwbudde/agg_go/internal/demo/lion"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/primitives"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	outline "github.com/cwbudde/agg_go/internal/renderer/outline"
	"github.com/cwbudde/agg_go/internal/transform"
)

// --- State ---

var (
	lionOutlineWidth = 1.0

	lionAngle = 0.0
	lionScale = 1.0
	lionSkewX = 0.0
	lionSkewY = 0.0

	lionDragging      = false
	lionRightDragging = false
)

// --- Drawing ---

type loRendererBase = renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]

type loOutlineBaseAdapter struct {
	rb *loRendererBase
}

func (a *loOutlineBaseAdapter) Width() int  { return a.rb.Width() }
func (a *loOutlineBaseAdapter) Height() int { return a.rb.Height() }

func (a *loOutlineBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.rb.BlendSolidHspan(x, y, length, c, convCovers)
}

func (a *loOutlineBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.rb.BlendSolidVspan(x, y, length, c, convCovers)
}

type loOutlineAAAdapter struct {
	ren *outline.RendererOutlineAA[*loOutlineBaseAdapter, color.RGBA8[color.Linear]]
}

func (a *loOutlineAAAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *loOutlineAAAdapter) Color(c color.RGBA8[color.Linear]) { a.ren.Color(c) }

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *loOutlineAAAdapter) Line0(lp primitives.LineParameters) {
	a.ren.Line0(&lp)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *loOutlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *loOutlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *loOutlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *loOutlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int) { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *loOutlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}

type loLionColorView struct {
	data *liondemo.LionData
}

func (v loLionColorView) GetColor(index int) color.RGBA8[color.Linear] {
	return v.data.Colors[index]
}

func drawLionOutlineDemo() {
	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	x1, y1, x2, y2 := getLionBoundingRect(lionData)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-cx, -cy))
	mtx.Multiply(transform.NewTransAffineScaling(lionScale))
	mtx.Multiply(transform.NewTransAffineScalingXY(-1, 1))
	mtx.Multiply(transform.NewTransAffineRotation(lionAngle))
	mtx.Multiply(transform.NewTransAffineSkewing(lionSkewX/1000.0, lionSkewY/1000.0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(width)*0.5, float64(height)*0.5))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	outlineVS := conv.NewRasterizerVertexSourceAdapter(transVS)

	profile := outline.NewLineProfileAA()
	profile.Width(lionOutlineWidth * mtx.GetScale())

	outlineBase := &loOutlineBaseAdapter{rb: rb}
	renOutline := outline.NewRendererOutlineAA[*loOutlineBaseAdapter, color.RGBA8[color.Linear]](outlineBase, profile)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*loOutlineAAAdapter, color.RGBA8[color.Linear]](&loOutlineAAAdapter{ren: renOutline})

	rasOutline.RenderAllPaths(outlineVS, loLionColorView{data: lionData}, lionData, lionData.NPaths)

	applyLinearToSRGB(img)
}

// --- Mouse handlers ---

func handleLionOutlineMouseDown(x, y float64, right bool) bool {
	if right {
		lionRightDragging = true
		lionSkewX = x
		lionSkewY = y
	} else {
		lionDragging = true
		applyLionTransform(x, y)
	}
	return true
}

func handleLionOutlineMouseMove(x, y float64, right bool) bool {
	if right && lionRightDragging {
		lionSkewX = x
		lionSkewY = y
		return true
	}
	if lionDragging {
		applyLionTransform(x, y)
		return true
	}
	return false
}

func handleLionOutlineMouseUp() {
	lionDragging = false
	lionRightDragging = false
}

func applyLionTransform(x, y float64) {
	cx := float64(width) * 0.5
	cy := float64(height) * 0.5
	dx := x - cx
	dy := y - cy
	lionAngle = math.Atan2(dy, dx)
	lionScale = math.Sqrt(dx*dx+dy*dy) / 100.0
	if lionScale < 0.01 {
		lionScale = 0.01
	}
}
