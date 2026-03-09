// Package main ports AGG's scanline_boolean2.cpp demo (URL/control-ready variant).
package main

import (
	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/scanlineboolean2"
)

const (
	w = 655
	h = 520
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	scanlineboolean2.Draw(ctx, scanlineboolean2.Config{
		Mode:      3,
		FillRule:  1,
		Operation: 2,
		CenterX:   float64(w) / 2,
		CenterY:   float64(h) / 2,
	})
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Scanline Boolean 2",
		Width:  w,
		Height: h,
	}, &demo{})
}
