// Package main ports AGG's circles.cpp demo as closely as the Go framework
// allows. The demo keeps its own point cloud and controls, updates on idle, and
// renders through the low-level rasterizer/pixfmt path instead of Agg2D.
package main

import (
	"fmt"
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	scalectrl "github.com/MeKo-Christian/agg_go/internal/ctrl/scale"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/curves"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	isc "github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	startWidth     = 400
	startHeight    = 400
	defaultNPoints = 10000
)

var (
	splineRX = []float64{0.000000, 0.200000, 0.400000, 0.910484, 0.957258, 1.000000}
	splineRY = []float64{1.000000, 0.800000, 0.600000, 0.066667, 0.169697, 0.600000}
	splineGX = []float64{0.000000, 0.292244, 0.485655, 0.564859, 0.795607, 1.000000}
	splineGY = []float64{0.000000, 0.607260, 0.964065, 0.892558, 0.435571, 0.000000}
	splineBX = []float64{0.000000, 0.055045, 0.143034, 0.433082, 0.764859, 1.000000}
	splineBY = []float64{0.385480, 0.128493, 0.021416, 0.271507, 0.713974, 1.000000}
)

type scatterPoint struct {
	x, y, z float64
	r, g, b float64
}

// clibcRand implements glibc's rand() with the default seed=1 state.
// This reproduces the same sequence as C rand() with no srand() call.
type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRand() *clibcRand {
	return &clibcRand{
		state: [31]int32{
			-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
			1627687941, -179304937, -2073333483, 1780058412, -1989503057,
			-615974602, 344556628, 939512070, -1249116260, 1507946756,
			-812545463, 154635395, 1388815473, -1926676823, 525320961,
			-1009028674, 968117788, -123449607, 1284210865, 435012392,
			-2017506339, -911064859, -370259173, 1132637927, 1398500161, -205601318,
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
	}
	r.rptr++
	if r.rptr >= 31 {
		r.rptr = 0
	}
	return result
}

func (r *clibcRand) rand15() uint32 {
	return uint32(r.next()) & 0x7FFF
}

func randomDbl(rng *clibcRand, start, end float64) float64 {
	r := rng.rand15()
	return float64(r)*(end-start)/32768.0 + start
}

func clampToU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func toRGBA8(c icol.RGBA) icol.RGBA8[icol.Linear] {
	return icol.RGBA8[icol.Linear]{
		R: clampToU8(c.R),
		G: clampToU8(c.G),
		B: clampToU8(c.B),
		A: clampToU8(c.A),
	}
}

func generatePoints(w, h float64, spR, spG, spB *curves.BSpline, rng *clibcRand) []scatterPoint {
	rx, ry := w/3.5, h/3.5
	pts := make([]scatterPoint, defaultNPoints)
	for i := range pts {
		z := randomDbl(rng, 0, 1)
		x := math.Cos(z*2*math.Pi) * rx
		y := math.Sin(z*2*math.Pi) * ry
		dist := randomDbl(rng, 0, rx/2)
		angle := randomDbl(rng, 0, math.Pi*2)
		pts[i] = scatterPoint{
			x: w/2 + x + math.Cos(angle)*dist,
			y: h/2 + y + math.Sin(angle)*dist,
			z: z,
			r: spR.Get(z) * 0.8,
			g: spG.Get(z) * 0.8,
			b: spB.Get(z) * 0.8,
		}
	}
	return pts
}

// simpleVS adapts any Rewind(uint)/Vertex() source to the rasterizer interface.
type simpleVS interface {
	Rewind(uint)
	Vertex() (float64, float64, basics.PathCommand)
}

type vsAdapter struct{ src simpleVS }

func (a *vsAdapter) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *vsAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint) { ev.e.Rewind(uint32(id)) }
func (ev *ellipseVS) Vertex() (float64, float64, basics.PathCommand) {
	var x, y float64
	cmd := ev.e.Vertex(&x, &y)
	return x, y, cmd
}

type gsvOutlineVS struct{ o *gsv.GSVTextOutline }

func (g *gsvOutlineVS) Rewind(id uint) { g.o.Rewind(id) }
func (g *gsvOutlineVS) Vertex() (float64, float64, basics.PathCommand) {
	return g.o.Vertex()
}

type bgr24Renderer struct {
	pf *pixfmt.PixFmtBGR24
}

func (r *bgr24Renderer) BlendSolidHspan(x, y, length int, c icol.RGBA8[icol.Linear], covers []basics.Int8u) {
	r.pf.BlendSolidHspan(
		x, y, length,
		icol.RGB8[icol.Linear]{R: c.R, G: c.G, B: c.B},
		c.A,
		covers,
	)
}

func (r *bgr24Renderer) BlendHline(x, y, x2 int, c icol.RGBA8[icol.Linear], cover basics.Int8u) {
	r.pf.BlendHline(
		x, y, x2,
		icol.RGB8[icol.Linear]{R: c.R, G: c.G, B: c.B},
		c.A,
		cover,
	)
}

func (r *bgr24Renderer) BlendColorHspan(x, y, length int, colors []icol.RGBA8[icol.Linear], covers []basics.Int8u, cover basics.Int8u) {
	if length <= 0 || len(colors) == 0 {
		return
	}
	for i := 0; i < length && i < len(colors); i++ {
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr == 0 {
			continue
		}
		r.pf.BlendPixel(x+i, y, icol.RGB8[icol.Linear]{R: colors[i].R, G: colors[i].G, B: colors[i].B}, colors[i].A, cvr)
	}
}

type demo struct {
	rng         *clibcRand
	splineR     *curves.BSpline
	splineG     *curves.BSpline
	splineB     *curves.BSpline
	points      []scatterPoint
	scaleCtrl   *scalectrl.ScaleCtrl
	selCtrl     *sliderctrl.SliderCtrl
	sizeCtrl    *sliderctrl.SliderCtrl
	initialized bool
	paused      bool
	nDrawn      int
}

func newDemo() *demo {
	d := &demo{
		rng: newClibcRand(),
		scaleCtrl: scalectrl.NewScaleCtrl(
			5, 5, startWidth-5, 12, false,
		),
		selCtrl:  sliderctrl.NewSliderCtrl(5, 20, startWidth-5, 27, false),
		sizeCtrl: sliderctrl.NewSliderCtrl(5, 35, startWidth-5, 42, false),
	}
	d.selCtrl.SetLabel("Selectivity=%.2f")
	d.sizeCtrl.SetLabel("Size=%.2f")
	d.prepareState()
	return d
}

func (d *demo) OnInit() {
	d.prepareState()
}

func (d *demo) IsAnimated() bool {
	return !d.paused
}

func (d *demo) OnIdle() {
	if d.paused {
		return
	}
	d.advancePoints()
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	d.prepareState()
	fx, fy := float64(x), float64(y)
	changed := false

	if btn.Left {
		if d.scaleCtrl.OnMouseButtonDown(fx, fy) {
			changed = true
		}
		if d.selCtrl.OnMouseButtonDown(fx, fy) {
			changed = true
		}
		if d.sizeCtrl.OnMouseButtonDown(fx, fy) {
			changed = true
		}
		d.generatePoints(startWidth, startHeight)
		changed = true
	}

	if btn.Right {
		d.paused = !d.paused
	}

	return changed
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	d.prepareState()
	fx, fy := float64(x), float64(y)
	pressed := btn.Left
	changed := d.scaleCtrl.OnMouseMove(fx, fy, pressed)

	if d.selCtrl.OnMouseMove(fx, fy, pressed) {
		changed = true
	}
	if d.sizeCtrl.OnMouseMove(fx, fy, pressed) {
		changed = true
	}

	return changed
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	d.prepareState()
	fx, fy := float64(x), float64(y)
	changed := d.scaleCtrl.OnMouseButtonUp(fx, fy)

	if d.selCtrl.OnMouseButtonUp(fx, fy) {
		changed = true
	}
	if d.sizeCtrl.OnMouseButtonUp(fx, fy) {
		changed = true
	}

	_ = btn
	return changed
}

func (d *demo) prepareState() {
	if d.splineR == nil {
		d.splineR = curves.NewBSplineFromPoints(splineRX, splineRY)
		d.splineG = curves.NewBSplineFromPoints(splineGX, splineGY)
		d.splineB = curves.NewBSplineFromPoints(splineBX, splineBY)
	}
	if d.rng == nil {
		d.rng = newClibcRand()
	}
	if d.scaleCtrl == nil {
		d.scaleCtrl = scalectrl.NewScaleCtrl(5, 5, startWidth-5, 12, false)
	}
	if d.selCtrl == nil {
		d.selCtrl = sliderctrl.NewSliderCtrl(5, 20, startWidth-5, 27, false)
		d.selCtrl.SetLabel("Selectivity=%.2f")
	}
	if d.sizeCtrl == nil {
		d.sizeCtrl = sliderctrl.NewSliderCtrl(5, 35, startWidth-5, 42, false)
		d.sizeCtrl.SetLabel("Size=%.2f")
	}

	if !d.initialized {
		d.points = generatePoints(startWidth, startHeight, d.splineR, d.splineG, d.splineB, d.rng)
		d.initialized = true
	}
}

func (d *demo) generatePoints(w, h float64) {
	d.points = generatePoints(w, h, d.splineR, d.splineG, d.splineB, d.rng)
}

func (d *demo) advancePoints() {
	if len(d.points) == 0 {
		return
	}
	sel := d.selCtrl.Value()
	for i := range d.points {
		d.points[i].x += randomDbl(d.rng, 0, sel) - sel*0.5
		d.points[i].y += randomDbl(d.rng, 0, sel) - sel*0.5
		d.points[i].z += randomDbl(d.rng, 0, sel*0.01) - sel*0.005
		if d.points[i].z < 0 {
			d.points[i].z = 0
		}
		if d.points[i].z > 1 {
			d.points[i].z = 1
		}
	}
}

func (d *demo) Render(img *agg.Image) {
	if img == nil {
		return
	}
	d.prepareState()

	workBuf := make([]uint8, img.Width()*img.Height()*3)
	rbuf := buffer.NewRenderingBufferU8WithData(workBuf, img.Width(), img.Height(), img.Width()*3)
	pf := pixfmt.NewPixFmtBGR24(rbuf)
	renBase := &bgr24Renderer{pf: pf}
	pf.Clear(icol.RGB8[icol.Linear]{R: 255, G: 255, B: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := isc.NewScanlineP8()
	ellipse := shapes.NewEllipse()

	nDrawn := 0
	v1, v2 := d.scaleCtrl.Value1(), d.scaleCtrl.Value2()
	sel := d.selCtrl.Value()
	size := d.sizeCtrl.Value()

	for _, pt := range d.points {
		alpha := 1.0
		if pt.z < v1 {
			alpha = 1.0 - (v1-pt.z)*sel*100.0
		} else if pt.z > v2 {
			alpha = 1.0 - (pt.z-v2)*sel*100.0
		}
		if alpha > 1 {
			alpha = 1
		}
		if alpha < 0 {
			alpha = 0
		}
		if alpha <= 0 {
			continue
		}

		ellipse.Init(pt.x, pt.y, size*5.0, size*5.0, 8, false)
		ras.Reset()
		ras.AddPath(&vsAdapter{src: &ellipseVS{e: ellipse}}, 0)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, icol.RGBA8[icol.Linear]{
			R: clampToU8(pt.r),
			G: clampToU8(pt.g),
			B: clampToU8(pt.b),
			A: clampToU8(alpha),
		})
		nDrawn++
	}

	renderCtrl(ras, sl, renBase, d.scaleCtrl)
	renderCtrl(ras, sl, renBase, d.selCtrl)
	renderCtrl(ras, sl, renBase, d.sizeCtrl)

	txt := gsv.NewGSVText()
	txt.SetSize(15.0, 0)
	txt.SetFlip(false)
	txt.SetStartPoint(10.0, float64(startHeight)-20.0)
	txt.SetText(fmt.Sprintf("%08d", nDrawn))

	outline := gsv.NewGSVTextOutline(txt)
	ras.Reset()
	ras.AddPath(&vsAdapter{src: &gsvOutlineVS{o: outline}}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, renBase, icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: 255})

	copyBGR24ToRGBA32(workBuf, img.Data, img.Width(), img.Height())
	d.nDrawn = nDrawn
}

func copyBGR24ToRGBA32(src, dst []uint8, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	srcStride := width * 3
	dstStride := width * 4
	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * srcStride
		dstOff := y * dstStride
		for x := 0; x < width; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			if s+2 >= len(src) || d+3 >= len(dst) {
				return
			}
			dst[d+0] = src[s+2]
			dst[d+1] = src[s+1]
			dst[d+2] = src[s+0]
			dst[d+3] = 255
		}
	}
}

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func renderCtrl(
	ras *rasType,
	sl *isc.ScanlineP8,
	renBase *bgr24Renderer,
	ctrl ctrlbase.Ctrl[icol.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&vsAdapter{src: ctrl}, uint32(pathID))
		renscan.RenderScanlinesAASolid(ras, sl, renBase, toRGBA8(ctrl.Color(pathID)))
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Circles",
		Width:  startWidth,
		Height: startHeight,
	}, &demo{})
}
