// Package main ports AGG's line_patterns.cpp demo (image-patterned Bezier curves).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/linepatterns"
)

const (
	lpScaleX = 1.0
	lpStartX = 0.0
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	linepatterns.Draw(img, lpScaleX, lpStartX)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Line Patterns",
		Width:  500,
		Height: 450,
	}, &demo{})
}
