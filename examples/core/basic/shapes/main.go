package main

import (
	agg "agg_go"
	"agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	// Clear with white background
	ctx.Clear(agg.White)

	// Draw a red filled ellipse
	ctx.SetColor(agg.Red)
	ctx.FillEllipse(200, 150, 80, 60)

	// Draw a blue ellipse outline
	ctx.SetColor(agg.Blue)
	ctx.DrawEllipse(200, 150, 100, 80)

	// Draw some smaller ellipses
	ctx.SetColor(agg.Green)
	ctx.FillEllipse(100, 100, 30, 20)

	ctx.SetColor(agg.Yellow)
	ctx.FillEllipse(300, 200, 25, 40)
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Shapes", Width: 400, Height: 300}, &demo{})
}
