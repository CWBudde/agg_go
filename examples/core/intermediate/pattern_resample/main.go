// Package main ports AGG's pattern_resample.cpp demo.
package main

import (
	"fmt"
	"math"
	"time"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	polygonctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/patternresample"
)

const (
	frameWidth  = 600
	frameHeight = 600
)

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

func toAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}

	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

type demo struct {
	quad      *polygonctrl.PolygonCtrl[icol.RGBA]
	transType *rboxctrl.RboxCtrl[icol.RGBA]
	gamma     *sliderctrl.SliderCtrl
	blur      *sliderctrl.SliderCtrl
	controls  []ctrlbase.Ctrl[icol.RGBA]
}

func newDemo() *demo {
	quad := polygonctrl.NewDefaultPolygonCtrl(4, 5.0)
	quad.SetClose(true)
	quad.SetInPolygonCheck(true)
	quad.SetXn(0, 100)
	quad.SetYn(0, 100)
	quad.SetXn(1, 500)
	quad.SetYn(1, 100)
	quad.SetXn(2, 500)
	quad.SetYn(2, 500)
	quad.SetXn(3, 100)
	quad.SetYn(3, 500)

	transType := rboxctrl.NewDefaultRboxCtrl(400, 500.0, 600, 595.0, true)
	transType.SetTextSize(7, 0)
	transType.AddItem("Affine No Resample")
	transType.AddItem("Affine Resample")
	transType.AddItem("Perspective No Resample LERP")
	transType.AddItem("Perspective No Resample Exact")
	transType.AddItem("Perspective Resample LERP")
	transType.AddItem("Perspective Resample Exact")
	transType.SetCurItem(4)

	gamma := sliderctrl.NewSliderCtrl(5.0, 590.0, 395.0, 595.0, true)
	gamma.SetRange(0.5, 3.0)
	gamma.SetValue(2.0)
	gamma.SetLabel("Gamma=%.3f")

	blur := sliderctrl.NewSliderCtrl(5.0, 575.0, 395.0, 580.0, true)
	blur.SetRange(0.5, 2.0)
	blur.SetValue(1.0)
	blur.SetLabel("Blur=%.3f")

	return &demo{
		quad:      quad,
		transType: transType,
		gamma:     gamma,
		blur:      blur,
		controls:  []ctrlbase.Ctrl[icol.RGBA]{transType, gamma, blur},
	}
}

func (d *demo) quadPoints() [4][2]float64 {
	return [4][2]float64{
		{d.quad.Xn(0), d.quad.Yn(0)},
		{d.quad.Xn(1), d.quad.Yn(1)},
		{d.quad.Xn(2), d.quad.Yn(2)},
		{d.quad.Xn(3), d.quad.Yn(3)},
	}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	elapsed := patternresample.DrawTimed(ctx, patternresample.Config{
		Mode:  d.transType.CurItem(),
		Gamma: d.gamma.Value(),
		Blur:  d.blur.Value(),
		Quad:  d.quadPoints(),
	})

	a := ctx.GetAgg2D()
	a.FontGSV(10)
	a.FillColor(agg.Black)
	a.Text(10, frameHeight-70, fmt.Sprintf("%3.2f ms", float64(elapsed)/float64(time.Millisecond)), false, 0, 0)

	for _, ctrl := range d.controls {
		renderCtrl(a, ctrl)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return d.quad.OnMouseButtonDown(fx, fy)
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls {
		if ctrl.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}
	if d.quad.OnMouseMove(fx, fy, btn.Left) {
		redraw = true
	}

	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	_ = btn
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}
	if d.quad.OnMouseButtonUp(fx, fy) {
		redraw = true
	}

	return redraw
}

func (d *demo) OnKey(key rune) bool {
	if key != ' ' {
		return false
	}

	points := [4][2]float64{
		{d.quad.Xn(0), d.quad.Yn(0)},
		{d.quad.Xn(1), d.quad.Yn(1)},
		{d.quad.Xn(2), d.quad.Yn(2)},
		{d.quad.Xn(3), d.quad.Yn(3)},
	}

	cx := (points[0][0] + points[1][0] + points[2][0] + points[3][0]) / 4.0
	cy := (points[0][1] + points[1][1] + points[2][1] + points[3][1]) / 4.0
	s, c := math.Sincos(math.Pi / 2.0)
	for i := range points {
		dx := points[i][0] - cx
		dy := points[i][1] - cy
		x := cx + dx*c - dy*s
		y := cy + dx*s + dy*c
		d.quad.SetXn(uint(i), x)
		d.quad.SetYn(uint(i), y)
	}

	return true
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Pattern Resample",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
