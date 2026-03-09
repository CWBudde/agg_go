package main

import (
	"math"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	agg2d := ctx.GetAgg2D()
	width, height := 400, 300

	agg2d.LineColor(agg.Color{R: 0, G: 0, B: 0, A: 255})
	agg2d.LineWidth(1.0)

	centerX, centerY := float64(width/4), float64(height/4)
	radius := 50.0

	for i := 0; i < 16; i++ {
		angle := float64(i) * math.Pi / 8.0
		endX := centerX + radius*math.Cos(angle)
		endY := centerY + radius*math.Sin(angle)
		agg2d.Line(centerX, centerY, endX, endY)
	}

	startX := float64(width * 3 / 4)
	for i := 0; i < 4; i++ {
		strokeWidth := 0.5 + float64(i)*0.5
		circleRadius := 15.0 + float64(i)*5.0
		circleY := float64(50 + i*40)
		agg2d.LineWidth(strokeWidth)
		agg2d.DrawCircle(startX, circleY, circleRadius)
	}

	agg2d.LineWidth(1.0)
	agg2d.Line(50, float64(height*2/3), 150, float64(height*2/3)+100)
	agg2d.Line(200, float64(height*2/3), 350, float64(height*2/3)+30)

	agg2d.FillColor(agg.Color{R: 128, G: 128, B: 255, A: 180})
	agg2d.ResetPath()
	agg2d.MoveTo(50, float64(height-50))
	agg2d.LineTo(100, float64(height-100))
	agg2d.LineTo(100, float64(height-50))
	agg2d.ClosePolygon()
	agg2d.DrawPath(agg.FillOnly)

	agg2d.ResetPath()
	agg2d.AddEllipse(200, float64(height-75), 40, 25, agg.CCW)
	agg2d.DrawPath(agg.FillOnly)

	agg2d.LineColor(agg.Color{R: 255, G: 0, B: 0, A: 255})
	agg2d.LineWidth(1.0)
	for i := 0; i < 5; i++ {
		offset := float64(i) * 0.25
		x := 300 + offset
		agg2d.Line(x, 200, x, 250)
	}
}

func main() {
	demorunner.Run(demorunner.Config{Title: "AA Demo", Width: 400, Height: 300}, &demo{})
}
