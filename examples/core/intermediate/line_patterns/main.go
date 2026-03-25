// Package main ports AGG's line_patterns.cpp demo (image-patterned Bezier curves).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	bezierctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/bezier"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/linepatterns"
)

type ctrlIface interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) color.RGBA
}

type ctrlVSAdaptor struct{ c ctrlIface }

func (a *ctrlVSAdaptor) Rewind(id uint32) { a.c.Rewind(uint(id)) }

func (a *ctrlVSAdaptor) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.c.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func renderCtrl(ag *agg.Agg2D, ctrl ctrlIface) {
	ras := ag.GetInternalRasterizer()
	vs := &ctrlVSAdaptor{c: ctrl}
	for id := uint(0); id < ctrl.NumPaths(); id++ {
		ras.Reset()
		ras.AddPath(vs, uint32(id))
		c := ctrl.Color(id)
		ag.RenderRasterizerWithColor(agg.NewColor(
			clampU8(c.R),
			clampU8(c.G),
			clampU8(c.B),
			clampU8(c.A),
		))
	}
}

func clampU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

type demo struct {
	curves   []*bezierctrl.BezierCtrl[color.RGBA]
	scaleX   *sliderctrl.SliderCtrl
	startX   *sliderctrl.SliderCtrl
	allCtrls []ctrlbase.Ctrl[color.RGBA]
}

func newDemo() *demo {
	ctrlColor := color.NewRGBA(0.0, 0.3, 0.5, 0.3)
	defaultCurves := linepatterns.DefaultCurves()
	curves := make([]*bezierctrl.BezierCtrl[color.RGBA], len(defaultCurves))
	allCtrls := make([]ctrlbase.Ctrl[color.RGBA], 0, len(defaultCurves)+2)
	for i, curve := range defaultCurves {
		ctrl := bezierctrl.NewBezierCtrl[color.RGBA](ctrlColor)
		ctrl.SetLineColor(ctrlColor)
		ctrl.SetCurve(curve.X1, curve.Y1, curve.X2, curve.Y2, curve.X3, curve.Y3, curve.X4, curve.Y4)
		curves[i] = ctrl
		allCtrls = append(allCtrls, ctrl)
	}

	scaleX := sliderctrl.NewSliderCtrl(5.0, 5.0, 240.0, 12.0, false)
	scaleX.SetLabel("Scale X=%.2f")
	scaleX.SetRange(0.2, 3.0)
	scaleX.SetValue(1.0)
	allCtrls = append(allCtrls, scaleX)

	startX := sliderctrl.NewSliderCtrl(250.0, 5.0, 495.0, 12.0, false)
	startX.SetLabel("Start X=%.2f")
	startX.SetRange(0.0, 10.0)
	startX.SetValue(0.0)
	allCtrls = append(allCtrls, startX)

	return &demo{
		curves:   curves,
		scaleX:   scaleX,
		startX:   startX,
		allCtrls: allCtrls,
	}
}

func (d *demo) Render(img *agg.Image) {
	curves := make([]linepatterns.Curve, len(d.curves))
	for i, ctrl := range d.curves {
		curves[i] = linepatterns.Curve{
			X1: ctrl.X1(), Y1: ctrl.Y1(),
			X2: ctrl.X2(), Y2: ctrl.Y2(),
			X3: ctrl.X3(), Y3: ctrl.Y3(),
			X4: ctrl.X4(), Y4: ctrl.Y4(),
		}
	}
	linepatterns.DrawCurves(img, d.scaleX.Value(), d.startX.Value(), curves)

	ctx := agg.NewContextForImage(img)
	ag := ctx.GetAgg2D()
	ag.ResetTransformations()
	for _, ctrl := range d.allCtrls {
		renderCtrl(ag, ctrl)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.allCtrls {
		if ctrl.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.allCtrls {
		if ctrl.OnMouseMove(fx, fy, btn.Left) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, _ lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.allCtrls {
		if ctrl.OnMouseButtonUp(fx, fy) {
			return true
		}
	}
	return false
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Line Patterns",
		Width:                 500,
		Height:                450,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
