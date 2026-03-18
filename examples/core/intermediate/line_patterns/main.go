// Package main ports AGG's line_patterns.cpp demo (image-patterned Bezier curves).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/linepatterns"
)

const (
	lpScaleX = 1.0
	lpStartX = 0.0
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	linepatterns.Draw(ctx.GetImage(), lpScaleX, lpStartX)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Line Patterns",
		Width:  500,
		Height: 450,
	}, &demo{})
}
