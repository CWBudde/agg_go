package main

import (
	"fmt"
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
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

var multiClipN = 3.0

// multiClipOutlineBaseAdapter adapts RendererMClip for outline rendering.
type multiClipOutlineBaseAdapter struct {
	renBase *renderer.RendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]
}

func (a *multiClipOutlineBaseAdapter) Width() int  { return a.renBase.Width() }
func (a *multiClipOutlineBaseAdapter) Height() int { return a.renBase.Height() }
func (a *multiClipOutlineBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidHspan(x, y, length, c, convCovers)
}

func (a *multiClipOutlineBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidVspan(x, y, length, c, convCovers)
}

// multiClipOutlineAAAdapter wraps RendererOutlineAA for the rasterizer.
type multiClipOutlineAAAdapter struct {
	ren *outline.RendererOutlineAA[*multiClipOutlineBaseAdapter, color.RGBA8[color.Linear]]
}

func (a *multiClipOutlineAAAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *multiClipOutlineAAAdapter) Color(c color.RGBA8[color.Linear]) { a.ren.Color(c) }

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *multiClipOutlineAAAdapter) Line0(lp primitives.LineParameters) {
	a.ren.Line0(&lp)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *multiClipOutlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *multiClipOutlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *multiClipOutlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *multiClipOutlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int) {
	a.ren.Pie(x, y, x1, y1, x2, y2)
}

func (a *multiClipOutlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}

// multiClipEllipseVS adapts shapes.Ellipse as a vertex source.
type multiClipEllipseVS struct {
	e *shapes.Ellipse
}

func (ev *multiClipEllipseVS) Rewind(id uint32) { ev.e.Rewind(id) }
func (ev *multiClipEllipseVS) Vertex(x, y *float64) uint32 {
	return uint32(ev.e.Vertex(x, y))
}

func drawMultiClipDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	// Attach rendering buffer using proper stride.
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())

	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	mclip := renderer.NewRendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)
	mainRb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Build the same affine transform as the standalone example.
	// Lion bounding box is approximately 0..238 x 0..379.
	baseDx := (238 - 0) / 2.0
	baseDy := (379 - 0) / 2.0

	mtx := transform.NewTransAffine()
	mtx.Translate(-baseDx, -baseDy)
	mtx.Scale(amLionScale)
	mtx.Rotate(amLionAngle + math.Pi)
	mtx.Multiply(transform.NewTransAffineSkewing(amLionSkewX/1000.0, amLionSkewY/1000.0))
	mtx.Translate(float64(w)/2, float64(h)/2)

	// Clear background to white.
	mclip.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// Set up N×N inset clip boxes.
	mclip.ResetClipping(false) // false = no visible regions; AddClipBox enables them
	n := int(multiClipN)
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			x1 := int(float64(w) * float64(x) / float64(n))
			y1 := int(float64(h) * float64(y) / float64(n))
			x2 := int(float64(w) * float64(x+1) / float64(n))
			y2 := int(float64(h) * float64(y+1) / float64(n))
			mclip.AddClipBox(x1+5, y1+5, x2-5, y2-5)
		}
	}

	// 1. Render the lion.
	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer[*renderer.RendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](mclip)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, lionData, lionData, lionData.NPaths)

	// 2. Render random Bresenham lines and markers (deterministic, glibc seed 1).
	rng := newMultiClipRandSeed1()
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
			m.Coord(float64(rng.randN(w))), m.Coord(float64(rng.randN(h))),
			m.Coord(float64(rng.randN(w))), m.Coord(float64(rng.randN(h))),
			true,
		)
		m.Marker(rng.randN(w), rng.randN(h), rng.randN(10)+5, markers.MarkerType(rng.randN(int(markers.EndOfMarkers))))
	}

	// 3. Render random anti-aliased lines.
	profile := outline.NewLineProfileAA()
	profile.Width(5.0)

	outAdapt := &multiClipOutlineBaseAdapter{renBase: mclip}
	renOutline := outline.NewRendererOutlineAA[*multiClipOutlineBaseAdapter, color.RGBA8[color.Linear]](outAdapt, profile)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*multiClipOutlineAAAdapter, color.RGBA8[color.Linear]](&multiClipOutlineAAAdapter{ren: renOutline})
	rasOutline.SetRoundCap(true)

	for i := 0; i < 50; i++ {
		renOutline.Color(color.RGBA8[color.Linear]{
			R: uint8(rng.randAnd(0x7F)),
			G: uint8(rng.randAnd(0x7F)),
			B: uint8(rng.randAnd(0x7F)),
			A: uint8(rng.randAnd(0x7F) + 127),
		})
		rasOutline.MoveToD(float64(rng.randN(w)), float64(rng.randN(h)))
		rasOutline.LineToD(float64(rng.randN(w)), float64(rng.randN(h)))
		rasOutline.Render(false)
	}

	// 4. Render random circles with radial gradient.
	sa := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	for i := 0; i < 50; i++ {
		cx := float64(rng.randN(w))
		cy := float64(rng.randN(h))
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
		ras.AddPath(&multiClipEllipseVS{e: ell}, 0)
		renscan.RenderScanlinesAA(ras, sl, mclip, sa, sg)
	}

	// 5. Reset clipping to full buffer for any overlay (slider was here in standalone).
	mclip.ResetClipping(true)
	_ = mainRb // mainRb available for ctrl rendering if needed

	// Apply linear→sRGB encoding.
	applyLinearToSRGB(img)

	logStatus(fmt.Sprintf("Multi-Clip Demo: N=%.0f", multiClipN))
}

func setMultiClipN(n float64) {
	multiClipN = n
}

func handleMultiClipMouseDown(x, y float64) bool {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	dx := x - float64(w)/2
	dy := y - float64(h)/2
	amLionAngle = math.Atan2(dy, dx)
	amLionScale = math.Sqrt(dy*dy+dx*dx) / 100.0
	return true
}

// ---------------------------------------------------------------------------
// glibc rand() with default seed (no srand call = seed 1).
// State pre-computed from glibc srand(1) initialization + 310 warmup cycles.
// Mirrors the clibcRand in examples/core/basic/multi_clip/main.go.
// ---------------------------------------------------------------------------

type multiClipRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newMultiClipRandSeed1() *multiClipRand {
	return &multiClipRand{
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

func (r *multiClipRand) next() int32 {
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
func (r *multiClipRand) randN(n int) int { return int(r.next()) % n }

// randAnd returns rand() & mask, matching C++ rand() & mask.
func (r *multiClipRand) randAnd(mask int) int { return int(r.next()) & mask }
