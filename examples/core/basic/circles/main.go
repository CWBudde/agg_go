package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

// demo demonstrates circle rendering, mirroring the original AGG 2.6 circles.cpp example.
type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	a := ctx.GetAgg2D()

	// Clear background to white
	a.ClearAll(agg.Color{R: 255, G: 255, B: 255, A: 255})

	// Test 1: Basic solid circles
	a.FillColor(agg.Color{R: 255, G: 0, B: 0, A: 255}) // Red
	a.FillCircle(50, 50, 30)

	a.FillColor(agg.Color{R: 0, G: 255, B: 0, A: 255}) // Green
	a.FillCircle(120, 50, 25)

	a.FillColor(agg.Color{R: 0, G: 0, B: 255, A: 255}) // Blue
	a.FillCircle(200, 50, 20)

	// Test 2: Outlined circles
	a.LineColor(agg.Color{R: 128, G: 128, B: 128, A: 255}) // Gray
	a.LineWidth(2.0)
	a.DrawCircle(50, 120, 30)
	a.DrawCircle(120, 120, 25)
	a.DrawCircle(200, 120, 20)

	// Test 3: Overlapping circles with alpha
	a.FillColor(agg.Color{R: 255, G: 0, B: 0, A: 128}) // Semi-transparent red
	a.FillCircle(80, 150, 25)

	a.FillColor(agg.Color{R: 0, G: 255, B: 0, A: 128}) // Semi-transparent green
	a.FillCircle(100, 150, 25)

	a.FillColor(agg.Color{R: 0, G: 0, B: 255, A: 128}) // Semi-transparent blue
	a.FillCircle(90, 170, 25)
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Circles", Width: 320, Height: 200}, &demo{})
}
