// Package main demonstrates the basic usage of the AGG Go library.
// This example creates a simple image with a colored background.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	// Clear the background to a light blue color
	ctx.Clear(agg.RGB(0.7, 0.8, 1.0))

	// Immediate-mode convenience helpers render right away.
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(100, 100, 200, 150)

	ctx.SetColor(agg.RGB(0, 0.8, 0)) // Green
	ctx.DrawCircle(400, 300, 80)

	// Explicit path mode is still available for custom shapes.
	ctx.SetColor(agg.Black)
	ctx.BeginPath()
	ctx.MoveTo(560, 180)
	ctx.LineTo(700, 340)
	ctx.LineTo(520, 340)
	ctx.ClosePath()
	ctx.Fill()
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Hello World", Width: 800, Height: 600}, &demo{})
}
