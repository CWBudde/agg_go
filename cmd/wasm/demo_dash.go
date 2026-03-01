package main

import (
	agg "agg_go"
)

func drawDashDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	agg2d.LineColor(agg.Black)
	agg2d.NoFill()

	patterns := [][]float64{
		{10, 5},
		{20, 10, 5, 5},
		{1, 1},
		{5, 10},
		{20, 2},
	}

	for i, p := range patterns {
		y := 50.0 + float64(i)*50.0
		agg2d.LineWidth(float64(i+1) * 2.0)

		agg2d.RemoveAllDashes()
		for j := 0; j < len(p); j += 2 {
			agg2d.AddDash(p[j], p[j+1])
		}

		agg2d.Line(50, y, 750, y)
	}

	// Complex path with dashes
	agg2d.LineWidth(3.0)
	agg2d.LineColor(agg.NewColor(0, 100, 200, 255))
	agg2d.RemoveAllDashes()
	agg2d.AddDash(15, 10)

	agg2d.ResetPath()
	agg2d.AddEllipse(400, 400, 150, 100, agg.CCW)
	agg2d.DrawPath(agg.StrokeOnly)
}
