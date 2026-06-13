// Port of the AGG C++ example simple_blur.cpp.
//
// Renders the AGG lion twice — a scanline-AA fill on the left and an
// anti-aliased outline (rasterizer_outline_aa + line_profile_aa) on the right —
// draws the double-stroked blur boundary, then applies a simple 3x3 box-blur
// inside the ellipse.
//
// The rendering mirrors the C++ pipeline faithfully: render_all_paths for the
// fill, and rasterizer_outline_aa with round caps for the outline. (The earlier
// version drew the outline with Agg2D StrokeOnly / conv_stroke, a different
// algorithm that produced visibly different sub-pixel edge coverage.)
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
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/primitives"
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
	lionOnce sync.Once
	lionData *liondemo.LionData
)

type renBaseType = renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]

type demo struct {
	cx, cy float64
}

type lionColorView struct {
	data *liondemo.LionData
}

func (v lionColorView) GetColor(index int) color.RGBA8[color.Linear] {
	return v.data.Colors[index]
}

func (d *demo) Render(img *agg.Image) {
	// Fill on the left, anti-aliased outline on the right (matches C++).
	// drawLionFill clears the window to white first.
	drawLionFill(img)
	drawLionOutline(img)

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

// lionTransform builds the affine used for both lion copies. tx is the final
// X translation (initial_width/4 for the fill, initial_width*3/4 for the
// outline — i.e. fill position + initial_width/2 as in the C++ source).
func lionTransform(ld *liondemo.LionData, tx, ty float64) *transform.TransAffine {
	baseDX, baseDY := getLionBaseOffset(ld)
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-baseDX, -baseDY))
	mtx.Multiply(transform.NewTransAffineScaling(1.0))
	mtx.Multiply(transform.NewTransAffineRotation(math.Pi))
	mtx.Multiply(transform.NewTransAffineTranslation(tx, ty))
	return mtx
}

func drawLionFill(img *agg.Image) {
	ld := liondata()

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newSimpleBlurRasterizer()
	sl := scanline.NewScanlineP8()

	mtx := lionTransform(ld, float64(img.Width())*0.25, float64(img.Height())*0.5)

	pathVS := path.NewPathStorageStlVertexSourceAdapter(ld.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rb)

	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, lionColorView{data: ld}, ld, ld.NPaths)
}

func drawLionOutline(img *agg.Image) {
	ld := liondata()

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)

	mtx := lionTransform(ld, float64(img.Width())*0.75, float64(img.Height())*0.5)

	pathVS := path.NewPathStorageStlVertexSourceAdapter(ld.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)

	// C++: line_profile_aa profile; profile.width(1.0);
	profile := outline.NewLineProfileAA()
	profile.Width(1.0)

	outlineBase := &outlineBaseAdapter{rb: rb}
	renOutline := outline.NewRendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]](outlineBase, profile)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter, color.RGBA8[color.Linear]](&outlineAAAdapter{ren: renOutline})
	rasOutline.SetRoundCap(true) // C++: ras.round_cap(true)

	outlineVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	rasOutline.RenderAllPaths(outlineVS, lionColorView{data: ld}, ld, ld.NPaths)
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

// outlineBaseAdapter adapts the renderer base to the outline-AA renderer's
// span-blending contract.
type outlineBaseAdapter struct {
	rb *renBaseType
}

func (a *outlineBaseAdapter) Width() int  { return a.rb.Width() }
func (a *outlineBaseAdapter) Height() int { return a.rb.Height() }

func (a *outlineBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.rb.BlendSolidHspan(x, y, length, c, convCovers)
}

func (a *outlineBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.rb.BlendSolidVspan(x, y, length, c, convCovers)
}

// outlineAAAdapter bridges the concrete RendererOutlineAA to the line-drawing
// interface expected by RasterizerOutlineAA.
type outlineAAAdapter struct {
	ren *outline.RendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]]
}

func (a *outlineAAAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *outlineAAAdapter) Color(c color.RGBA8[color.Linear]) { a.ren.Color(c) }

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line0(lp primitives.LineParameters) {
	a.ren.Line0(&lp)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *outlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int) { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *outlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
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

func getLionBaseOffset(ld *liondemo.LionData) (baseDX, baseDY float64) {
	x1, y1, x2, y2 := getLionBoundingRect(ld)
	return (x2 - x1) * 0.5, (y2 - y1) * 0.5
}
