// Package main ports AGG's circles.cpp demo — a scatter-plot visualisation.
//
// Generates 10 000 random scatter points parametrised by z ∈ [0,1].
// Each point lies on a ring of radius W/3.5 at angle z·2π, perturbed by a
// random offset.  Colour comes from three B-splines evaluated at z.
// A scale-ctrl selects the visible z-range (default 0.3…0.7); a selectivity
// slider controls fade-out speed (default 0.5); a size slider sets the circle
// radius (default 0.5 → 2.5 px).  The drawn count is shown top-left.
package main

import (
	"fmt"
	"math"
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/curves"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	scalectrl "github.com/MeKo-Christian/agg_go/internal/ctrl/scale"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	startWidth     = 400
	startHeight    = 400
	defaultNPoints = 10000
)

var splineRX = []float64{0.000000, 0.200000, 0.400000, 0.910484, 0.957258, 1.000000}
var splineRY = []float64{1.000000, 0.800000, 0.600000, 0.066667, 0.169697, 0.600000}
var splineGX = []float64{0.000000, 0.292244, 0.485655, 0.564859, 0.795607, 1.000000}
var splineGY = []float64{0.000000, 0.607260, 0.964065, 0.892558, 0.435571, 0.000000}
var splineBX = []float64{0.000000, 0.055045, 0.143034, 0.433082, 0.764859, 1.000000}
var splineBY = []float64{0.385480, 0.128493, 0.021416, 0.271507, 0.713974, 1.000000}

type scatterPoint struct {
	x, y, z    float64
	r, g, b float64
}

// clibcRand implements glibc's rand() with the default seed=1 state.
// This replicates the exact sequence produced by C's rand() with no srand() call
// (POSIX default seed=1), enabling pixel-perfect parity with the C++ reference.
type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRand() *clibcRand {
	// State computed by glibc srand(1): Park-Miller LCG init, then 310 warm-up cycles.
	// Verified to produce the same sequence as C rand() with no srand() call.
	r := &clibcRand{
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
	return r
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := (int32)(uint32(r.state[r.fptr]) >> 1)
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

func generatePoints(w, h float64, spR, spG, spB *curves.BSpline) []scatterPoint {
	rng := newClibcRand()
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

// simpleVS adapts any Rewind(uint)/Vertex() source to the rasterizer's uint32 interface.
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

// ellipseVS wraps shapes.Ellipse.
type ellipseVS struct{ e *shapes.Ellipse }

func (ev *ellipseVS) Rewind(id uint) { ev.e.Rewind(uint32(id)) }
func (ev *ellipseVS) Vertex() (float64, float64, basics.PathCommand) {
	var x, y float64
	cmd := ev.e.Vertex(&x, &y)
	return x, y, cmd
}

// gsvOutlineVS wraps GSVTextOutline.
type gsvOutlineVS struct{ o *gsv.GSVTextOutline }

func (g *gsvOutlineVS) Rewind(id uint) { g.o.Rewind(id) }
func (g *gsvOutlineVS) Vertex() (float64, float64, basics.PathCommand) {
	return g.o.Vertex()
}

func toAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		if v <= 0 { return 0 }
		if v >= 1 { return 255 }
		return uint8(v*255 + 0.5)
	}
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&vsAdapter{src: c}, uint32(i))
		a.RenderRasterizerWithColor(toAggColor(c.Color(i)))
	}
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	w := float64(ctx.GetImage().Width())
	h := float64(ctx.GetImage().Height())

	spR := curves.NewBSplineFromPoints(splineRX, splineRY)
	spG := curves.NewBSplineFromPoints(splineGX, splineGY)
	spB := curves.NewBSplineFromPoints(splineBX, splineBY)

	pts := generatePoints(w, h, spR, spG, spB)

	// Controls with C++ defaults: scale=[0.3,0.7], sliders=0.5
	// flipY=false matches C++ "!flip_y" (flip_y=true window, so !flip_y=false for ctrls)
	scaleZ := scalectrl.NewScaleCtrl(5, 5, w-5, 12, false)
	sliderSel := sliderctrl.NewSliderCtrl(5, 20, w-5, 27, false)
	sliderSel.SetLabel("Selectivity=%.2f")
	sliderSize := sliderctrl.NewSliderCtrl(5, 35, w-5, 42, false)
	sliderSize.SetLabel("Size=%.2f")

	v1, v2 := scaleZ.Value1(), scaleZ.Value2()
	sel := sliderSel.Value()
	size := sliderSize.Value()

	ctx.Clear(agg.White)

	ras := a.GetInternalRasterizer()
	e := shapes.NewEllipse()
	nDrawn := 0

	for i := range pts {
		z := pts[i].z
		alpha := 1.0
		if z < v1 {
			alpha = 1.0 - (v1-z)*sel*100.0
		} else if z > v2 {
			alpha = 1.0 - (z-v2)*sel*100.0
		}
		if alpha > 1 { alpha = 1 }
		if alpha < 0 { alpha = 0 }
		if alpha <= 0 { continue }

		r8 := uint8(pts[i].r * 255)
		g8 := uint8(pts[i].g * 255)
		b8 := uint8(pts[i].b * 255)
		a8 := uint8(alpha * 255)

		radius := size * 5.0
		e.Init(pts[i].x, pts[i].y, radius, radius, 8, false)
		ras.Reset()
		ras.AddPath(&vsAdapter{src: &ellipseVS{e: e}}, 0)
		a.RenderRasterizerWithColor(agg.NewColor(r8, g8, b8, a8))
		nDrawn++
	}

	// Render controls.
	renderCtrl(a, scaleZ)
	renderCtrl(a, sliderSel)
	renderCtrl(a, sliderSize)

	// Draw drawn-count text (top-left, height=15, like C++ gsv_text+outline).
	txt := gsv.NewGSVText()
	txt.SetSize(15.0, 0)
	txt.SetFlip(false)
	txt.SetStartPoint(10.0, h-20.0)
	txt.SetText(fmt.Sprintf("%08d", nDrawn))

	outline := gsv.NewGSVTextOutline(txt)
	ras.Reset()
	ras.AddPath(&vsAdapter{src: &gsvOutlineVS{o: outline}}, 0)
	a.RenderRasterizerWithColor(agg.NewColor(0, 0, 0, 255))
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Circles",
		Width:  startWidth,
		Height: startHeight,
	}, &demo{})
}
