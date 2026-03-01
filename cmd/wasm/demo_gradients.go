package main

import (
	agg "agg_go"
)

func drawGradientsDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Linear Gradient Rectangle
	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(1.0)
	agg2d.FillLinearGradient(100, 100, 300, 300, agg.Red, agg.Blue, 1.0)
	ctx.FillRectangle(100, 100, 200, 200)
	ctx.DrawRectangle(100, 100, 200, 200)

	// 2. Radial Gradient Circle
	agg2d.FillRadialGradient(500, 200, 100, agg.Yellow, agg.Transparent, 1.0)
	ctx.FillCircle(500, 200, 100)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(500, 200, 100)

	// 3. Radial Gradient with Multi-Stop (3 colors)
	agg2d.FillRadialGradientMultiStop(400, 450, 120, agg.Green, agg.White, agg.Red)
	ctx.FillCircle(400, 450, 120)
}
