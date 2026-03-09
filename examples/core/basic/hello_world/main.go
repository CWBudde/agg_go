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

	// Set drawing color to red
	ctx.SetColor(agg.Red)

	// Draw a simple rectangle
	ctx.DrawRectangle(100, 100, 200, 150)
	ctx.Fill()

	// Draw a circle
	ctx.SetColor(agg.RGB(0, 0.8, 0)) // Green
	ctx.DrawCircle(400, 300, 80)
	ctx.Fill()
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Hello World", Width: 800, Height: 600}, &demo{})
}
