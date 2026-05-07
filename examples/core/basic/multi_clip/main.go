package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	liondemo "github.com/cwbudde/agg_go/internal/demo/lion"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/primitives"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	"github.com/cwbudde/agg_go/internal/renderer/markers"
	"github.com/cwbudde/agg_go/internal/renderer/outline"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
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
	rng := newClibcRand(1)
	m := markers.NewRendererMarkers[*renderer.RendererMClip[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](mclip)
	for i := 0; i < 50; i++ {
		m.LineColor(rgbaRandRTL(rng, 0x7F))
		m.FillColor(rgbaRandRTL(rng, 0x7F))

		y2 := rng.randN(d.h)
		x2 := rng.randN(d.w)
		y1 := rng.randN(d.h)
		x1 := rng.randN(d.w)
		m.Line(
			m.Coord(float64(x1)), m.Coord(float64(y1)),
			m.Coord(float64(x2)), m.Coord(float64(y2)),
			true,
		)

		markerType := markers.MarkerType(rng.randN(int(markers.EndOfMarkers)))
		radius := rng.randN(10) + 5
		y := rng.randN(d.h)
		x := rng.randN(d.w)
		m.Marker(x, y, radius, markerType)
	}

	// 3. Render random anti-aliased lines
	profile := outline.NewLineProfileAA()
	profile.Width(5.0)

	outAdapt := &outlineBaseAdapter{renBase: mclip}
	renOutline := outline.NewRendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]](outAdapt, profile)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter, color.RGBA8[color.Linear]](&outlineAAAdapter{ren: renOutline})
	rasOutline.SetRoundCap(true)

	for i := 0; i < 50; i++ {
		renOutline.Color(rgbaRandRTL(rng, 0x7F))
		y1 := rng.randN(d.h)
		x1 := rng.randN(d.w)
		rasOutline.MoveToD(float64(x1), float64(y1))
		y2 := rng.randN(d.h)
		x2 := rng.randN(d.w)
		rasOutline.LineToD(float64(x2), float64(y2))
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
		c2 := rgbaRandRGBRTL(rng, 255)

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
		Title:                 "AGG Example. Clipping to multiple rectangle regions",
		Width:                 512,
		Height:                400,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}

// ---------------------------------------------------------------------------
// glibc rand()/random() with the same initialization as srand(seed).
// This mirrors internal/demo/alphamask2's parity RNG.
// ---------------------------------------------------------------------------

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRand(seed int32) *clibcRand {
	if seed == 0 {
		seed = 1
	}

	const (
		mod  int64 = 2147483647
		mult int64 = 16807
	)

	var seq [344]int32
	seq[0] = seed
	for i := 1; i < 31; i++ {
		v := mult * int64(seq[i-1]) % mod
		if v < 0 {
			v += mod
		}
		seq[i] = int32(v)
	}
	for i := 31; i < 34; i++ {
		seq[i] = seq[i-31]
	}
	for i := 34; i < len(seq); i++ {
		seq[i] = seq[i-31] + seq[i-3]
	}

	rng := &clibcRand{
		fptr: 3,
		rptr: 0,
	}
	copy(rng.state[:3], seq[341:344])
	copy(rng.state[3:], seq[313:341])
	return rng
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := int32(uint32(r.state[r.fptr]) >> 1)
	r.fptr++
	if r.fptr >= 31 {
		r.fptr = 0
	}
	r.rptr++
	if r.rptr >= 31 {
		r.rptr = 0
	}
	return result
}

// randN returns rand() % n, matching C++ rand() % n.
func (r *clibcRand) randN(n int) int { return int(r.next()) % n }

// randAnd returns rand() & mask, matching C++ rand() & mask.
func (r *clibcRand) randAnd(mask int) int { return int(r.next()) & mask }

func multiClipSrgba8(r, g, b, a uint8) color.RGBA8[color.Linear] {
	return color.ConvertRGBA8SRGBToLinear(color.RGBA8[color.SRGB]{
		R: r, G: g, B: b, A: a,
	})
}

func rgbaRandRTL(rng *clibcRand, alphaBase int) color.RGBA8[color.Linear] {
	a := uint8(rng.randAnd(0x7F) + alphaBase)
	b := uint8(rng.randAnd(0x7F))
	g := uint8(rng.randAnd(0x7F))
	r := uint8(rng.randAnd(0x7F))
	return multiClipSrgba8(r, g, b, a)
}

func rgbaRandRGBRTL(rng *clibcRand, alpha int) color.RGBA8[color.Linear] {
	b := uint8(rng.randAnd(0x7F))
	g := uint8(rng.randAnd(0x7F))
	r := uint8(rng.randAnd(0x7F))
	return multiClipSrgba8(r, g, b, uint8(alpha))
}

// Port of AGG C++ multi_clip.cpp – multi-clip region rendering.
//
// Renders the lion and other primitives through a grid of N×N inset clip rectangles.
