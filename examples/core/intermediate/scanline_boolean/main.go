package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
)

func main() {
	ctx := agg.NewContext(900, 420)
	ctx.Clear(agg.White)
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	leftX1 := []float64{110, 290, 260, 90}
	leftY1 := []float64{80, 110, 280, 250}
	leftX2 := []float64{170, 350, 310, 140}
	leftY2 := []float64{40, 150, 340, 260}
	drawBooleanPanel(agg2d, leftX1, leftY1, leftX2, leftY2, false)

	rightX1 := []float64{520, 700, 670, 500}
	rightY1 := []float64{80, 110, 280, 250}
	rightX2 := []float64{580, 760, 720, 550}
	rightY2 := []float64{40, 150, 340, 260}
	drawBooleanPanel(agg2d, rightX1, rightY1, rightX2, rightY2, true)

	const filename = "scanline_boolean.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}

func drawBooleanPanel(agg2d *agg.Agg2D, x1, y1, x2, y2 []float64, evenOdd bool) {
	agg2d.FillColor(agg.RGBA(0.1, 0.5, 0.8, 0.18))
	agg2d.NoLine()
	addPolyToAgg2D(agg2d, x1, y1)
	agg2d.DrawPath(agg.FillOnly)

	agg2d.FillColor(agg.RGBA(0.8, 0.2, 0.1, 0.18))
	addPolyToAgg2D(agg2d, x2, y2)
	agg2d.DrawPath(agg.FillOnly)

	agg2d.FillColor(agg.Black)
	agg2d.ResetPath()
	addPolyToAgg2D(agg2d, x1, y1)
	addPolyToAgg2D(agg2d, x2, y2)
	agg2d.FillEvenOdd(evenOdd)
	agg2d.DrawPath(agg.FillOnly)

	drawHandles(agg2d, x1, y1, agg.RGBA(0.1, 0.5, 0.8, 0.6))
	drawHandles(agg2d, x2, y2, agg.RGBA(0.8, 0.2, 0.1, 0.6))
	agg2d.FillEvenOdd(false)
}

func addPolyToAgg2D(agg2d *agg.Agg2D, x, y []float64) {
	agg2d.MoveTo(x[0], y[0])
	for i := 1; i < len(x); i++ {
		agg2d.LineTo(x[i], y[i])
	}
	agg2d.ClosePolygon()
}

func drawHandles(agg2d *agg.Agg2D, x, y []float64, c agg.Color) {
	for i := range x {
		agg2d.FillColor(c)
		agg2d.NoLine()
		agg2d.FillCircle(x[i], y[i], 6)
		agg2d.LineColor(agg.Black)
		agg2d.LineWidth(1.0)
		agg2d.DrawCircle(x[i], y[i], 6)
	}
}
