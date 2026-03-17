// Package main ports AGG's idea.cpp demo.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	icolor "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/idea"
)

// ctrlIface is the minimal vertex-source interface for rendering controls.
type ctrlIface interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) icolor.RGBA
}

type ctrlPathAdapter struct{ ctrl ctrlIface }

func (a *ctrlPathAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }
func (a *ctrlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func linearToSRGB(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	var s float64
	if v <= 0.0031308 {
		s = 12.92 * v
	} else {
		s = 1.055*math.Pow(v, 1.0/2.4) - 0.055
	}
	return uint8(s*255 + 0.5)
}

func renderCtrl(ag *agg.Agg2D, ctrl ctrlIface) {
	ras := ag.GetInternalRasterizer()
	adapter := &ctrlPathAdapter{ctrl: ctrl}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		c := ctrl.Color(pathID)
		var a uint8
		if c.A <= 0 {
			a = 0
		} else if c.A >= 1 {
			a = 255
		} else {
			a = uint8(c.A*255 + 0.5)
		}
		ag.RenderRasterizerWithColor(agg.NewColor(linearToSRGB(c.R), linearToSRGB(c.G), linearToSRGB(c.B), a))
	}
}

type demo struct {
	state idea.State
}

func newDemo() *demo {
	return &demo{state: idea.DefaultState()}
}

func (d *demo) Render(ctx *agg.Context) {
	idea.Draw(ctx, d.state)

	// Render UI controls on top (matching C++ on_draw render_ctrl calls).
	a := ctx.GetAgg2D()
	rotateCb := checkbox.NewDefaultCheckboxCtrl(10, 3, "Rotate", false)
	rotateCb.SetChecked(d.state.Rotate)
	rotateCb.SetTextSize(7.0, 0)
	evenOddCb := checkbox.NewDefaultCheckboxCtrl(60, 3, "Even-Odd", false)
	evenOddCb.SetChecked(d.state.EvenOdd)
	evenOddCb.SetTextSize(7.0, 0)
	draftCb := checkbox.NewDefaultCheckboxCtrl(130, 3, "Draft", false)
	draftCb.SetChecked(d.state.Draft)
	draftCb.SetTextSize(7.0, 0)
	roundoffCb := checkbox.NewDefaultCheckboxCtrl(175, 3, "Roundoff", false)
	roundoffCb.SetChecked(d.state.Roundoff)
	roundoffCb.SetTextSize(7.0, 0)
	angleSlider := slider.NewSliderCtrl(10, 21, 240, 27, false)
	angleSlider.SetLabel("Step=%4.3f degree")
	angleSlider.SetValue(d.state.AngleDelta)
	for _, c := range []ctrlIface{rotateCb, evenOddCb, draftCb, roundoffCb, angleSlider} {
		renderCtrl(a, c)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Idea",
		Width:  250,
		Height: 280,
	}, newDemo())
}
