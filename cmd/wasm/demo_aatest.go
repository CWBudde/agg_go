// Based on the original AGG examples: aa_test.cpp.
package main

import (
	"math"

	agg "agg_go"
)

func drawAATestDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.Black)

	cx, cy := float64(width)*0.5, float64(height)*0.5

	// 1. Radial Line Test
	agg2d.NoFill()
	agg2d.LineColor(agg.RGBA(1.0, 1.0, 1.0, 0.2))

	agg2d.LineWidth(1.0)

	const deg2rad = 2.0 * math.Pi / 180.0

	radius := math.Min(cx, cy)
	for i := 180; i > 0; i-- {
		n := float64(i) * deg2rad
		x2 := cx + radius*math.Sin(n)
		y2 := cy + radius*math.Cos(n)

		agg2d.RemoveAllDashes()
		if i < 90 {
			agg2d.AddDash(float64(i), float64(i))
		}
		agg2d.Line(cx, cy, x2, y2)
	}
	agg2d.NoDashes()

	// 2. Point and Line Tests
	for i := 1; i <= 20; i++ {
		agg2d.FillColor(agg.White)
		agg2d.NoLine()

		// Integral point sizes 1..20
		x := 20.0 + float64(i*(i+1)) + 0.5
		agg2d.FillCircle(x, 20.5, float64(i)*0.5)

		// Fractional point sizes 0..2
		agg2d.FillCircle(18.0+float64(i)*4.0+0.5, 33.5, float64(i)/20.0)

		// Fractional point positioning
		agg2d.FillCircle(18.0+float64(i)*4.0+float64(i-1)/10.0+0.5, 27.0+float64(i-1)/10.0+0.5, 0.5)

		// Integral line widths 1..20 with gradient
		lwX1 := 20.0 + float64(i*(i+1))
		lwY1 := 40.5
		lwX2 := 20.0 + float64(i*(i+1)) + float64(i-1)*4.0
		lwY2 := 100.5
		c2 := agg.RGBA(float64(i%2), float64(i%3)*0.5, float64(i%5)*0.25, 1.0)
		agg2d.LineLinearGradient(lwX1, lwY1, lwX2, lwY2, agg.White, c2, 1.0)
		agg2d.LineWidth(float64(i))
		agg2d.Line(lwX1, lwY1, lwX2, lwY2)

		// Fractional line lengths H (red/blue)
		agg2d.LineLinearGradient(17.5+float64(i)*4, 107, 17.5+float64(i)*4+float64(i)/6.66666667, 107, agg.RGBA(1, 0, 0, 1), agg.RGBA(0, 0, 1, 1), 1.0)
		agg2d.LineWidth(1.0)
		agg2d.Line(17.5+float64(i)*4, 107, 17.5+float64(i)*4+float64(i)/6.66666667, 107)

		// Fractional line lengths V (red/blue)
		agg2d.LineLinearGradient(18+float64(i)*4, 112.5, 18+float64(i)*4, 112.5+float64(i)/6.66666667, agg.RGBA(1, 0, 0, 1), agg.RGBA(0, 0, 1, 1), 1.0)
		agg2d.Line(18+float64(i)*4, 112.5, 18+float64(i)*4, 112.5+float64(i)/6.66666667)

		// Fractional line positioning (red to white)
		y := 120.0 + float64(i-1)*3.1
		agg2d.LineLinearGradient(21.5, y, 52.5, y, agg.RGBA(1, 0, 0, 1), agg.White, 1.0)
		agg2d.Line(21.5, y, 52.5, y)

		// Fractional line width 2..0 (green to white)
		y2 := 118.0 + float64(i)*3.0
		agg2d.LineLinearGradient(52.5, y2, 83.5, y2, agg.RGBA(0, 1, 0, 1), agg.White, 1.0)
		agg2d.LineWidth(2.0 - float64(i-1)/10.0)
		agg2d.Line(52.5, y2, 83.5, y2)

		// Stippled fractional width 2..0 (blue to white)
		y3 := 119.0 + float64(i)*3.0
		agg2d.LineLinearGradient(83.5, y3, 114.5, y3, agg.RGBA(0, 0, 1, 1), agg.White, 1.0)
		agg2d.RemoveAllDashes()
		agg2d.AddDash(3.0, 3.0)
		agg2d.Line(83.5, y3, 114.5, y3)
		agg2d.NoDashes()

		// Integral line width, horizontal (mipmap test)
		agg2d.LineColor(agg.White)
		agg2d.LineWidth(1.0)
		if i <= 10 {
			agg2d.Line(125.5, 119.5+float64(i+2)*float64(i)*0.5, 135.5, 119.5+float64(i+2)*float64(i)*0.5)
		}

		// Fractional line width 0..2, 1px H
		agg2d.LineWidth(float64(i) / 10.0)
		agg2d.Line(17.5+float64(i)*4, 192, 18.5+float64(i)*4, 192)

		// Fractional line positioning, 1px H
		agg2d.LineWidth(1.0)
		agg2d.Line(17.5+float64(i)*4+float64(i-1)/10.0, 186, 18.5+float64(i)*4+float64(i-1)/10.0, 186)
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
