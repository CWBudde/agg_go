// Simplified rounded rectangle example demonstrating AGG's rendering pipeline
// This is a non-interactive version of the original AGG rounded_rect.cpp demo
package main

import (
	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	// Clear background to white
	ctx.Clear(agg.White)

	// Demo 1: Blue filled rounded rectangle
	ctx.SetColor(agg.Blue)
	ctx.FillRoundedRectangle(120, 100, 160, 100, 30)

	// Demo 2: Red outlined rounded rectangle
	ctx.SetColor(agg.Red)
	ctx.DrawRoundedRectangle(370, 100, 160, 100, 30)

	// Demo 3: Green filled rounded rectangle with different proportions
	ctx.SetColor(agg.Green)
	ctx.FillRoundedRectangle(120, 260, 160, 120, 40)

	// Demo 4: Purple outlined rounded rectangle
	ctx.SetColor(agg.RGB(0.6, 0.4, 0.8)) // Purple
	ctx.DrawRoundedRectangle(370, 260, 160, 120, 40)

	// Demo 5: Orange very rounded rectangle (pill shape)
	ctx.SetColor(agg.RGB(1.0, 0.6, 0.0)) // Orange
	ctx.FillRoundedRectangle(220, 400, 200, 70, 35)
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Rounded Rectangle Demo", Width: 640, Height: 480}, &demo{})
}
