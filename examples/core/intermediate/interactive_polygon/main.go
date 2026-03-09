// Package main ports AGG's interactive_polygon.cpp helper as a standalone interactive demo.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/interactivepolygon"
)

type demo struct {
	state *interactivepolygon.State
}

func (d *demo) Render(ctx *agg.Context) {
	d.state.Draw(ctx)
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	return d.state.MouseDown(float64(x), float64(y))
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	return d.state.MouseMove(float64(x), float64(y), btn.Left)
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	return d.state.MouseUp(float64(x), float64(y))
}

func main() {
	const (
		w = 640
		h = 480
	)

	d := &demo{state: interactivepolygon.NewState(w, h)}
	demorunner.Run(demorunner.Config{
		Title:  "Interactive Polygon",
		Width:  w,
		Height: h,
	}, d)
}
