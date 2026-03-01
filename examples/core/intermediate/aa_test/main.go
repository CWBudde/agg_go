package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
)

const (
	width  = 800
	height = 600
)

func main() {
	ctx := agg.NewContext(width, height)
	ctx.Clear(agg.RGB(0.08, 0.08, 0.08))

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.NoFill()
	agg2d.LineColor(agg.RGBA(1.0, 1.0, 1.0, 0.2))
	agg2d.LineWidth(1.0)

	cx, cy := 170.0, 170.0
	for i := 1; i <= 40; i++ {
		angle := float64(i) * agg.Pi / 20.0
		x2 := cx + 120.0*math.Cos(angle)
		y2 := cy + 120.0*math.Sin(angle)
		agg2d.RemoveAllDashes()
		if i%4 == 0 {
			agg2d.AddDash(float64(i), float64(i))
		}
		agg2d.Line(cx, cy, x2, y2)
	}
	agg2d.RemoveAllDashes()

	for i := 0; i < 12; i++ {
		x := 100.0 + float64(i%4)*85.0
		y := 320.0 + float64(i/4)*80.0
		r := 3.5 + float64(i)*2.4
		agg2d.FillColor(agg.White)
		agg2d.NoLine()
		agg2d.FillCircle(x, y, r)
	}

	lwX1, lwY1 := 420.0, 90.0
	for i := 1; i <= 12; i++ {
		y := lwY1 + float64(i-1)*18.0
		x2 := 720.0 - float64(i)*6.0
		c2 := agg.RGBA(float64(i%2), float64(i%3)*0.5, float64(i%5)*0.25, 1.0)
		agg2d.FillLinearGradient(lwX1, y, x2, y, agg.White, c2, 1.0)
		agg2d.LineWidth(float64(i))
		agg2d.LineColor(agg.Black)
		agg2d.Line(lwX1, y, x2, y)
	}

	for i := 0; i < 6; i++ {
		x1 := 430.0 + float64(i%3)*110.0
		y1 := 360.0 + float64(i/3)*120.0
		x2 := x1 + 75.0
		y2 := y1 + 10.0
		x3 := x1 + 30.0
		y3 := y1 + 72.0
		c2 := agg.RGBA(float64(i%2), float64(i%3)*0.5, float64(i%5)*0.25, 1.0)
		agg2d.FillLinearGradient(x1, y1, x2, y2, agg.White, c2, 1.0)
		agg2d.NoLine()
		agg2d.ResetPath()
		agg2d.MoveTo(x1, y1)
		agg2d.LineTo(x2, y2)
		agg2d.LineTo(x3, y3)
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}

	const filename = "aa_test.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
