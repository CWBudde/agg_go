// Package main ports AGG's gpc_test.cpp demo.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/gpctest"
)

type demo struct {
	cx float64
	cy float64
}

func (d *demo) Render(ctx *agg.Context) {
	gpctest.Draw(ctx, gpctest.Config{
		Scene:     3,
		Operation: 2,
		CenterX:   d.cx,
		CenterY:   d.cy,
	})
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	d.cx, d.cy = float64(x), float64(y)
	return true
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	d.cx, d.cy = float64(x), float64(y)
	return true
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	return false
}

func main() {
	d := &demo{cx: math.NaN(), cy: math.NaN()}
	demorunner.Run(demorunner.Config{
		Title:  "GPC Test",
		Width:  640,
		Height: 520,
	}, d)
}
