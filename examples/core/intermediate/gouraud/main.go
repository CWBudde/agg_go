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

	x := [3]float64{100, 500, 300}
	y := [3]float64{100, 150, 500}
	dilation := 0.5

	xc := (x[0] + x[1] + x[2]) / 3.0
	yc := (y[0] + y[1] + y[2]) / 3.0
	x1 := (x[1]+x[0])*0.5 - (xc - (x[1]+x[0])*0.5)
	y1 := (y[1]+y[0])*0.5 - (yc - (y[1]+y[0])*0.5)
	x2 := (x[2]+x[1])*0.5 - (xc - (x[2]+x[1])*0.5)
	y2 := (y[2]+y[1])*0.5 - (yc - (y[2]+y[1])*0.5)
	x3 := (x[0]+x[2])*0.5 - (xc - (x[0]+x[2])*0.5)
	y3 := (y[0]+y[2])*0.5 - (yc - (y[0]+y[2])*0.5)

	cRed := agg.NewColor(255, 0, 0, 255)
	cGreen := agg.NewColor(0, 255, 0, 255)
	cBlue := agg.NewColor(0, 0, 255, 255)
	cWhite := agg.White
	cBlack := agg.Black

	agg2d.GouraudTriangle(x[0], y[0], x[1], y[1], xc, yc, cRed, cGreen, cWhite, dilation)
	agg2d.GouraudTriangle(x[1], y[1], x[2], y[2], xc, yc, cGreen, cBlue, cWhite, dilation)
	agg2d.GouraudTriangle(x[2], y[2], x[0], y[0], xc, yc, cBlue, cRed, cWhite, dilation)
	agg2d.GouraudTriangle(x[0], y[0], x[1], y[1], x1, y1, cRed, cGreen, cBlack, dilation)
	agg2d.GouraudTriangle(x[1], y[1], x[2], y[2], x2, y2, cGreen, cBlue, cBlack, dilation)
	agg2d.GouraudTriangle(x[2], y[2], x[0], y[0], x3, y3, cBlue, cRed, cBlack, dilation)

	for i := range x {
		agg2d.FillColor(agg.NewColor(200, 50, 20, 150))
		agg2d.NoLine()
		agg2d.FillCircle(x[i], y[i], 8)
		agg2d.LineColor(agg.Black)
		agg2d.LineWidth(1.0)
		agg2d.DrawCircle(x[i], y[i], 8)
	}

	const filename = "gouraud.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
