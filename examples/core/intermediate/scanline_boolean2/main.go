// Package main ports AGG's scanline_boolean2.cpp demo (URL/control-ready variant).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/scanlineboolean2"
)

const (
	w = 655
	h = 520
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	scanlineboolean2.Draw(ctx, scanlineboolean2.Config{
		Mode:      3,
		FillRule:  1,
		Operation: 2,
		CenterX:   float64(w) / 2,
		CenterY:   float64(h) / 2,
	})
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Scanline Boolean 2",
		Width:  w,
		Height: h,
	}, &demo{})
}
