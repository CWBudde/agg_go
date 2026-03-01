package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
)

type point struct {
	x float64
	y float64
}

func main() {
	ctx := agg.NewContext(800, 600)
	ctx.Clear(agg.White)
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	points := []point{
		{70, 420},
		{150, 170},
		{260, 340},
		{360, 110},
		{470, 310},
		{590, 160},
		{710, 410},
	}

	agg2d.NoFill()
	agg2d.LineColor(agg.NewColor(180, 180, 180, 255))
	agg2d.LineWidth(1.0)
	agg2d.AddDash(5, 5)
	agg2d.ResetPath()
	agg2d.MoveTo(points[0].x, points[0].y)
	for i := 1; i < len(points); i++ {
		agg2d.LineTo(points[i].x, points[i].y)
	}
	agg2d.DrawPath(agg.StrokeOnly)
	agg2d.RemoveAllDashes()

	agg2d.NoFill()
	agg2d.LineColor(agg.NewColor(0, 150, 255, 255))
	agg2d.LineWidth(3.0)
	agg2d.ResetPath()
	agg2d.MoveTo(points[0].x, points[0].y)
	for i := 0; i < len(points)-1; i++ {
		p0 := points[max(0, i-1)]
		p1 := points[i]
		p2 := points[i+1]
		p3 := points[min(len(points)-1, i+2)]
		cp1x := p1.x + (p2.x-p0.x)/6.0
		cp1y := p1.y + (p2.y-p0.y)/6.0
		cp2x := p2.x - (p3.x-p1.x)/6.0
		cp2y := p2.y - (p3.y-p1.y)/6.0
		agg2d.CubicCurveTo(cp1x, cp1y, cp2x, cp2y, p2.x, p2.y)
	}
	agg2d.DrawPath(agg.StrokeOnly)

	agg2d.FillColor(agg.Red)
	agg2d.NoLine()
	for _, p := range points {
		agg2d.FillCircle(p.x, p.y, 5)
	}

	const filename = "bspline.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
