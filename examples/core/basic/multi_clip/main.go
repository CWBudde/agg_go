package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/renderer/markers"
	"github.com/MeKo-Christian/agg_go/internal/renderer/outline"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

type demo struct {
	ld     liondemo.LionData
	baseDx float64
	baseDy float64
	angle  float64
	scale  float64
	skewX  float64
	skewY  float64
	w, h   int

	numCb *sliderctrl.SliderCtrl
}

func newDemo() *demo {
	ld := liondemo.Parse()

	// C++: m_num_cb(5, 5, 150, 12, !flip_y)  => !true = false
	numCb := sliderctrl.NewSliderCtrl(5, 5, 150, 12, false)
	numCb.SetRange(2, 10)
	numCb.SetValue(6.0)
	numCb.SetLabel("N=%.2f")

	return &demo{
		ld: ld,
		// hardcoded values for liondemo to avoid re-parsing for bounds
		baseDx: (238 - 0) / 2.0, // lion is approx 0 to 238
		baseDy: (379 - 0) / 2.0, // lion is approx 0 to 379
		scale:  1.0,
		numCb:  numCb,
	}
}

type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVS) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type ellipseVS struct {
	e *shapes.Ellipse
}

func (ev *ellipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *ellipseVS) Vertex(x, y *float64) uint32 {
	return uint32(ev.e.Vertex(x, y))
}

type outlineBaseAdapter struct {
	renBase *renderer.RendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]
}

func (a *outlineBaseAdapter) Width() int  { return a.renBase.Width() }
func (a *outlineBaseAdapter) Height() int { return a.renBase.Height() }
func (a *outlineBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidHspan(x, y, length, c, convCovers)
}
func (a *outlineBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidVspan(x, y, length, c, convCovers)
}

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

func clampU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255.0 + 0.5)
}

func renderCtrl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVS{ctrl: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, color.RGBA8[color.Linear]{
			R: clampU8(c.R),
			G: clampU8(c.G),
			B: clampU8(c.B),
			A: clampU8(c.A),
		})
	}
}

func (d *demo) Render(img *agg.Image) {
	d.w = img.Width()
	d.h = img.Height()

	mainBuf := buffer.NewRenderingBufferU8WithData(img.Data, d.w, d.h, img.Stride())
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainBuf)
	mclip := renderer.NewRendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](mainPixf)
	mainRb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](mainPixf)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	mtx := transform.NewTransAffine()
	mtx.Translate(-d.baseDx, -d.baseDy)
	mtx.Scale(d.scale)
	mtx.Rotate(d.angle + math.Pi)
	mtx.Multiply(transform.NewTransAffineSkewing(d.skewX/1000.0, d.skewY/1000.0))
	mtx.Translate(float64(d.w)/2, float64(d.h)/2)

	mclip.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	mclip.ResetClipping(false) // "false" means "no visible regions"
	n := int(d.numCb.Value())
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			x1 := int(float64(d.w) * float64(x) / float64(n))
			y1 := int(float64(d.h) * float64(y) / float64(n))
			x2 := int(float64(d.w) * float64(x+1) / float64(n))
			y2 := int(float64(d.h) * float64(y+1) / float64(n))
			mclip.AddClipBox(x1+5, y1+5, x2-5, y2-5)
		}
	}

	// 1. Render the lion
	pathVS := path.NewPathStorageStlVertexSourceAdapter(d.ld.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer[*renderer.RendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](mclip)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, &d.ld, &d.ld, d.ld.NPaths)

	// 2. Render random Bresenham lines and markers
	rng := newClibcRandSeed1()
	m := markers.NewRendererMarkers[*renderer.RendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](mclip)
	for i := 0; i < 50; i++ {
		m.LineColor(color.RGBA8[color.Linear]{
			R: uint8(rng.randAnd(0x7F)),
			G: uint8(rng.randAnd(0x7F)),
			B: uint8(rng.randAnd(0x7F)),
			A: uint8(rng.randAnd(0x7F) + 127),
		})
		m.FillColor(color.RGBA8[color.Linear]{
			R: uint8(rng.randAnd(0x7F)),
			G: uint8(rng.randAnd(0x7F)),
			B: uint8(rng.randAnd(0x7F)),
			A: uint8(rng.randAnd(0x7F) + 127),
		})

		m.Line(
			m.Coord(float64(rng.randN(d.w))), m.Coord(float64(rng.randN(d.h))),
			m.Coord(float64(rng.randN(d.w))), m.Coord(float64(rng.randN(d.h))),
			true,
		)

		m.Marker(rng.randN(d.w), rng.randN(d.h), rng.randN(10)+5, markers.MarkerType(rng.randN(int(markers.EndOfMarkers))))
	}

	// 3. Render random anti-aliased lines
	profile := outline.NewLineProfileAA()
	profile.Width(5.0)

	outAdapt := &outlineBaseAdapter{renBase: mclip}
	renOutline := outline.NewRendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]](outAdapt, profile)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter, color.RGBA8[color.Linear]](&outlineAAAdapter{ren: renOutline})
	rasOutline.SetRoundCap(true)

	for i := 0; i < 50; i++ {
		renOutline.Color(color.RGBA8[color.Linear]{
			R: uint8(rng.randAnd(0x7F)),
			G: uint8(rng.randAnd(0x7F)),
			B: uint8(rng.randAnd(0x7F)),
			A: uint8(rng.randAnd(0x7F) + 127),
		})
		rasOutline.MoveToD(float64(rng.randN(d.w)), float64(rng.randN(d.h)))
		rasOutline.LineToD(float64(rng.randN(d.w)), float64(rng.randN(d.h)))
		rasOutline.Render(false)
	}

	// 4. Render random circles with gradient
	sa := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	for i := 0; i < 50; i++ {
		cx := float64(rng.randN(d.w))
		cy := float64(rng.randN(d.h))
		radius := float64(rng.randN(10) + 5)

		grm := transform.NewTransAffine()
		grm.Scale(radius / 10.0)
		grm.Translate(cx, cy)
		grm.Invert()

		inter := span.NewSpanInterpolatorLinearDefault(grm)

		c1 := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 0}
		c2 := color.RGBA8[color.Linear]{
			R: uint8(rng.randAnd(0x7F)),
			G: uint8(rng.randAnd(0x7F)),
			B: uint8(rng.randAnd(0x7F)),
			A: 255,
		}

		sg := span.NewRadialGradientRGBA8(inter, c1, c2, 0, 10, 256)

		ell := shapes.NewEllipseWithParams(cx, cy, radius, radius, 32, false)

		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		renscan.RenderScanlinesAA(ras, sl, mclip, sa, sg)
	}

	// 5. Render slider
	mclip.ResetClipping(true) // "true" means "all rendering buffer is visible".
	renderCtrl(ras, sl, mainRb, d.numCb)
}

func (d *demo) handleTransform(x, y int) {
	fx := float64(x) - float64(d.w)/2
	fy := float64(y) - float64(d.h)/2
	d.angle = math.Atan2(fy, fx)
	d.scale = math.Sqrt(fy*fy+fx*fx) / 100.0
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if btn.Left {
		if d.numCb.OnMouseButtonDown(fx, fy) {
			return true
		}
		d.handleTransform(x, y)
		return true
	}
	if btn.Right {
		d.skewX = float64(x)
		d.skewY = float64(y)
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.numCb.OnMouseMove(fx, fy, btn.Left) {
		return true
	}
	if btn.Left {
		d.handleTransform(x, y)
		return true
	}
	if btn.Right {
		d.skewX = float64(x)
		d.skewY = float64(y)
		return true
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	return d.numCb.OnMouseButtonUp(fx, fy)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG Example. Clipping to multiple rectangle regions",
		Width:  512,
		Height: 400,
		FlipY:  true,
	}, newDemo())
}

// ---------------------------------------------------------------------------
// glibc rand() with default seed (no srand call = seed 1).
// State pre-computed from glibc srand(1) initialization + 310 warmup cycles.
// ---------------------------------------------------------------------------

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRandSeed1() *clibcRand {
	return &clibcRand{
		state: [31]int32{
			-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
			1627687941, -179304937, -2073915851, 19113796, -73392711,
			864575501, 1954350912, 1853386453, 108502596, 1770989849,
			1140076113, 2120506151, 1431634354, 1162235973, 1961623253,
			1362719266, 2132549216, 1961162464, -225679901, 196417531,
			1647413401, 1435272633, 1081395475, 411831818, 52187654,
			-499380962,
		},
		fptr: 3,
		rptr: 0,
	}
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := int32(uint32(r.state[r.fptr]) >> 1)
	r.fptr++
	if r.fptr >= 31 {
		r.fptr = 0
		r.rptr++
	} else {
		r.rptr++
		if r.rptr >= 31 {
			r.rptr = 0
		}
	}
	return result
}

// randN returns rand() % n, matching C++ rand() % n.
func (r *clibcRand) randN(n int) int { return int(r.next()) % n }

// randAnd returns rand() & mask, matching C++ rand() & mask.
func (r *clibcRand) randAnd(mask int) int { return int(r.next()) & mask }

// Port of AGG C++ multi_clip.cpp – multi-clip region rendering.
//
// Renders the lion and other primitives through a grid of N×N inset clip rectangles.
