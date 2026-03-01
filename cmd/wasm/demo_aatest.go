// Based on the original AGG examples: aa_test.cpp.
package main

import (
	"math"

	agg "agg_go"
)

func drawAATestDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Radial Line Test
	cx, cy := float64(width)/2.0, float64(height)/2.0
	agg2d.NoFill()
	agg2d.LineColor(agg.RGBA(1.0, 1.0, 1.0, 0.2))
	agg2d.LineWidth(1.0)

	radius := math.Min(cx, cy)
	for i := 180; i > 0; i-- {
		n := 2.0 * math.Pi * float64(i) / 180.0
		x2 := cx + radius*math.Sin(n)
		y2 := cy + radius*math.Cos(n)

		agg2d.RemoveAllDashes()
		if i < 90 {
			agg2d.AddDash(float64(i), float64(i))
		}
		agg2d.Line(cx, cy, x2, y2)
	}
	agg2d.NoDashes()

	// 2. Integral and Fractional Point/Ellipse Test
	for i := 1; i <= 20; i++ {
		agg2d.FillColor(agg.White)
		agg2d.NoLine()

		// Integral point sizes 1..20
		x := 20.0 + float64(i*(i+1)) + 0.5
		y := 20.5
		r := float64(i) / 2.0
		agg2d.FillCircle(x, y, r)

		// Fractional point sizes 0..2
		x2 := 18.0 + float64(i)*4.0 + 0.5
		y2 := 33.0 + 0.5
		r2 := float64(i) / 20.0
		agg2d.FillCircle(x2, y2, r2)

		// Fractional point positioning
		x3 := 18.0 + float64(i)*4.0 + float64(i-1)/10.0 + 0.5
		y3 := 27.0 + float64(i-1)/10.0 + 0.5
		r3 := 0.5
		agg2d.FillCircle(x3, y3, r3)

		// Integral line widths 1..20 with gradients
		lwX1 := 20.0 + float64(i*(i+1))
		lwY1 := 40.5
		lwX2 := 20.0 + float64(i*(i+1)) + float64(i-1)*4.0
		lwY2 := 100.5
		
		c2 := agg.RGBA(float64(i%2), float64(i%3)*0.5, float64(i%5)*0.25, 1.0)
		agg2d.FillLinearGradient(lwX1, lwY1, lwX2, lwY2, agg.White, c2, 1.0)
		agg2d.LineWidth(float64(i))
		agg2d.LineColor(agg.Black) // Use Solid for line color if no line gradient set
		// Original uses dash_gradient.draw which renders a stroked line
		// In Agg2D we can set LineLinearGradient if we wanted gradient strokes
		agg2d.Line(lwX1, lwY1, lwX2, lwY2)
	}

	// 3. Triangles with Gradients
	for i := 1; i <= 13; i++ {
		x1 := float64(width) - 150.0
		y1 := float64(height) - 20.0 - float64(i)*(float64(i)+1.5)
		x2 := float64(width) - 20.0
		y2 := float64(height) - 20.0 - float64(i)*(float64(i)+1.0)
		x3 := x2
		y3 := float64(height) - 20.0 - float64(i)*(float64(i)+2.0)

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
}
