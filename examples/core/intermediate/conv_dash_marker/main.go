package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
)

func main() {
	ctx := agg.NewContext(800, 600)
	ctx.Clear(agg.White)
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	patterns := [][]float64{
		{20, 5},
		{10, 3, 3, 3},
		{2, 4},
		{18, 6, 3, 6},
	}
	for i, pattern := range patterns {
		y := 70.0 + float64(i)*42.0
		agg2d.LineColor(agg.NewColor(100, 100, 100, 255))
		agg2d.NoFill()
		agg2d.LineWidth(2.0)
		agg2d.RemoveAllDashes()
		for j := 0; j < len(pattern)-1; j += 2 {
			agg2d.AddDash(pattern[j], pattern[j+1])
		}
		agg2d.Line(50, y, 750, y)
	}

	points := [][2]float64{{157, 160}, {469, 270}, {243, 410}, {285, 280}}

	agg2d.FillColor(agg.NewColor(200, 150, 50, 100))
	agg2d.NoLine()
	agg2d.ResetPath()
	for i, p := range points {
		if i == 0 {
			agg2d.MoveTo(p[0], p[1])
		} else {
			agg2d.LineTo(p[0], p[1])
		}
	}
	agg2d.ClosePolygon()
	agg2d.DrawPath(agg.FillOnly)

	agg2d.NoFill()
	agg2d.LineColor(agg.Black)
	agg2d.LineWidth(3.0)
	agg2d.RemoveAllDashes()
	agg2d.AddDash(20, 5)
	agg2d.AddDash(5, 5)
	agg2d.ResetPath()
	for i, p := range points {
		if i == 0 {
			agg2d.MoveTo(p[0], p[1])
		} else {
			agg2d.LineTo(p[0], p[1])
		}
	}
	agg2d.ClosePolygon()
	agg2d.DrawPath(agg.StrokeOnly)

	for _, p := range points {
		agg2d.FillColor(agg.NewColor(200, 50, 20, 150))
		agg2d.NoLine()
		agg2d.FillCircle(p[0], p[1], 6)
		agg2d.LineColor(agg.Black)
		agg2d.LineWidth(1.0)
		agg2d.DrawCircle(p[0], p[1], 6)
	}

	const filename = "conv_dash_marker.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
