// Based on the original AGG examples: bezier_div.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/curves"
)

var (
	bezierDivPoints    = [8]float64{170, 424, 13, 87, 488, 423, 26, 333}
	bezierDivSelected  = -1
	bezierDivApproximationScale = 1.0
	bezierDivAngleTolerance = 15.0 // degrees
)

func drawBezierDivDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	x1, y1 := bezierDivPoints[0], bezierDivPoints[1]
	x2, y2 := bezierDivPoints[2], bezierDivPoints[3]
	x3, y3 := bezierDivPoints[4], bezierDivPoints[5]
	x4, y4 := bezierDivPoints[6], bezierDivPoints[7]

	// 1. Draw Subdivision Curve (Green)
	curveDiv := curves.NewCurve4Div()
	curveDiv.SetApproximationScale(bezierDivApproximationScale)
	curveDiv.SetAngleTolerance(bezierDivAngleTolerance * math.Pi / 180.0)
	curveDiv.Init(x1, y1, x2, y2, x3, y3, x4, y4)

	agg2d.NoFill()
	agg2d.LineColor(agg.NewColor(0, 150, 0, 150))
	agg2d.LineWidth(2.0)
	
	agg2d.ResetPath()
	for {
		vx, vy, cmd := curveDiv.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			agg2d.MoveTo(vx, vy)
		} else {
			agg2d.LineTo(vx, vy)
		}
	}
	agg2d.DrawPath(agg.StrokeOnly)

	// 2. Draw Incremental Curve (Red, shifted)
	curveInc := curves.NewCurve4Inc()
	curveInc.SetApproximationScale(bezierDivApproximationScale)
	curveInc.Init(x1+50, y1, x2+50, y2, x3+50, y3, x4+50, y4)

	agg2d.LineColor(agg.NewColor(200, 0, 0, 150))
	agg2d.ResetPath()
	first := true
	for {
		vx, vy, cmd := curveInc.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if first {
			agg2d.MoveTo(vx, vy)
			first = false
		} else {
			agg2d.LineTo(vx, vy)
		}
	}
	agg2d.DrawPath(agg.StrokeOnly)

	// 3. Draw control polygon (dashed)
	agg2d.LineColor(agg.NewColor(0, 0, 0, 100))
	agg2d.LineWidth(1.0)
	agg2d.ResetPath()
	agg2d.MoveTo(x1, y1)
	agg2d.LineTo(x2, y2)
	agg2d.LineTo(x3, y3)
	agg2d.LineTo(x4, y4)
	agg2d.DrawPath(agg.StrokeOnly)

	// 4. Draw handles
	for i := 0; i < 4; i++ {
		drawHandle(bezierDivPoints[i*2], bezierDivPoints[i*2+1])
	}
}

func handleBezierDivMouseDown(x, y float64) bool {
	bezierDivSelected = -1
	for i := 0; i < 4; i++ {
		dist := math.Sqrt(math.Pow(x-bezierDivPoints[i*2], 2) + math.Pow(y-bezierDivPoints[i*2+1], 2))
		if dist < 10 {
			bezierDivSelected = i
			return true
		}
	}
	return false
}

func handleBezierDivMouseMove(x, y float64) bool {
	if bezierDivSelected != -1 {
		bezierDivPoints[bezierDivSelected*2] = x
		bezierDivPoints[bezierDivSelected*2+1] = y
		return true
	}
	return false
}

func handleBezierDivMouseUp() {
	bezierDivSelected = -1
}
