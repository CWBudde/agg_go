// Port of the AGG C++ example simple_blur.cpp.
//
// Renders the AGG lion, draws the double-stroked blur boundary, then applies
// a simple 3x3 box-blur inside the ellipse — demonstrating basic pixel-level
// post-processing on a rendered scene.
package main

import (
	"math"
	"sync"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	liondemo "github.com/cwbudde/agg_go/internal/demo/lion"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

var (
	lionOnce sync.Once
	lionData *liondemo.LionData
)

type demo struct {
	cx, cy float64
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Draw the lion fill on the left, then the colored outline pass on the right.
	drawLionFill(agg2d, img.Width(), img.Height())
	drawLionOutline(agg2d, img.Width(), img.Height())

	rx, ry := 100.0, 100.0

	// Match C++ simple_blur.cpp: stroke the ellipse, then stroke that stroke.
	drawBlurBoundary(img, d.cx, d.cy, rx, ry)

	// Snapshot after the boundary is drawn. The C++ demo calls
	// copy_window_to_img(0) after rendering ell_stroke2, so the inner ring
	// participates in the blurred span instead of being overwritten.
	bgImg := agg.NewImage(make([]uint8, len(img.Data)), img.Width(), img.Height(), img.Stride())
	copy(bgImg.Data, img.Data)

	// Apply 3x3 box-blur inside the ellipse through the rasterizer, preserving
	// the original anti-aliased ellipse coverage.
	applyBlurInsideEllipse(img, bgImg, d.cx, d.cy, rx, ry)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if btn.Left {
		d.cx = float64(x)
		d.cy = float64(y)
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if btn.Left {
		d.cx = float64(x)
		d.cy = float64(y)
		return true
	}
	return false
}

func (d *demo) OnMouseUp(_, _ int, _ lowlevelrunner.Buttons) bool { return false }

func drawLionFill(agg2d *agg.Agg2D, width, height int) {
	ld := liondata()
	x1, y1, x2, y2 := getLionBoundingRect(ld)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	m := transform.NewTransAffine()
	m.Translate(-cx, -cy)
	m.Scale(1.0)
	m.Rotate(math.Pi)
	m.Translate(float64(width)*0.25, float64(height)*0.5)

	agg2d.NoLine()
	for i := 0; i < ld.NPaths; i++ {
		agg2d.FillColor(agg.NewColor(ld.Colors[i].R, ld.Colors[i].G, ld.Colors[i].B, 255))
		agg2d.ResetPath()
		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			m.Transform(&x, &y)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(x, y)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(x, y)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}
}

func drawLionOutline(agg2d *agg.Agg2D, width, height int) {
	ld := liondata()
	x1, y1, x2, y2 := getLionBoundingRect(ld)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	m := transform.NewTransAffine()
	m.Translate(-cx, -cy)
	m.Scale(1.0)
	m.Rotate(math.Pi)
	m.Translate(float64(width)*0.75, float64(height)*0.5)

	agg2d.NoFill()
	agg2d.LineWidth(1.0)
	for i := 0; i < ld.NPaths; i++ {
		agg2d.LineColor(agg.NewColor(ld.Colors[i].R, ld.Colors[i].G, ld.Colors[i].B, 255))
		agg2d.ResetPath()
		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			m.Transform(&x, &y)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(x, y)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(x, y)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.StrokeOnly)
	}
}

func drawBlurBoundary(img *agg.Image, cx, cy, rx, ry float64) {
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32PlainLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt(pf)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(ren)
	renSolid.SetColor(color.RGBA8[color.Linear]{R: 0, G: 51, B: 0, A: 255})

	ellipse := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
	stroke1 := conv.NewConvStroke(&ellipseConvVS{ell: ellipse})
	stroke1.SetWidth(6.0)
	stroke2 := conv.NewConvStroke(stroke1)
	stroke2.SetWidth(2.0)

	ras := newSimpleBlurRasterizer()
	ras.AddPath(conv.NewRasterizerVertexSourceAdapter(stroke2), 0)
	renscan.RenderScanlines(ras, scanline.NewScanlineP8(), renSolid)
}

// applyBlurInsideEllipse performs the C++ span_simple_blur_rgb24 operation
// through an anti-aliased ellipse rasterizer, sampling from src.
func applyBlurInsideEllipse(dst, src *agg.Image, cx, cy, rx, ry float64) {
	dstRbuf := buffer.NewRenderingBufferU8()
	dstRbuf.Attach(dst.Data, dst.Width(), dst.Height(), dst.Stride())
	srcRbuf := buffer.NewRenderingBufferU8()
	srcRbuf.Attach(src.Data, src.Width(), src.Height(), src.Stride())

	pf := pixfmt.NewPixFmtRGBA32PlainLinear(dstRbuf)
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

type ellipseVS struct{ ell *shapes.Ellipse }

func (v *ellipseVS) Rewind(pathID uint32)        { v.ell.Rewind(pathID) }
func (v *ellipseVS) Vertex(x, y *float64) uint32 { return uint32(v.ell.Vertex(x, y)) }

type ellipseConvVS struct{ ell *shapes.Ellipse }

func (v *ellipseConvVS) Rewind(pathID uint) { v.ell.Rewind(uint32(pathID)) }
func (v *ellipseConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = v.ell.Vertex(&x, &y)
	return x, y, cmd
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

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Simple Blur",
		Width:                 512,
		Height:                400,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, &demo{
		cx: 100,
		cy: 102,
	})
}

func liondata() *liondemo.LionData {
	lionOnce.Do(func() {
		ld := liondemo.Parse()
		lionData = &ld
	})
	return lionData
}

func getLionBoundingRect(ld *liondemo.LionData) (x1, y1, x2, y2 float64) {
	x1, y1, x2, y2 = math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for idx := uint(0); idx < ld.Path.TotalVertices(); idx++ {
		x, y, cmd := ld.Path.Vertex(idx)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			if x < x1 {
				x1 = x
			}
			if y < y1 {
				y1 = y
			}
			if x > x2 {
				x2 = x
			}
			if y > y2 {
				y2 = y
			}
		}
	}
	return
}
