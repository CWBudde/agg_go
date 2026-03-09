// Port of AGG C++ conv_stroke.cpp – "Line Join" interactive demo.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
)

// --- State ---

var (
	strokePts = [3][2]float64{
		{157, 160},
		{469, 270},
		{243, 410},
	}
	strokeJoin       = 0 // 0=miter, 1=round, 2=bevel
	strokeCap        = 0 // 0=butt, 1=square, 2=round
	strokeWidth      = 20.0
	strokeMiterLimit = 4.0

	strokeSelected = -1
	strokeDragDX   = 0.0
	strokeDragDY   = 0.0
)

var (
	strokeJoins = []agg.LineJoin{agg.JoinMiter, agg.JoinRound, agg.JoinBevel}
	strokeCaps  = []agg.LineCap{agg.CapButt, agg.CapSquare, agg.CapRound}
)

// --- Drawing ---

func drawConvStrokeDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	x := [3]float64{strokePts[0][0], strokePts[1][0], strokePts[2][0]}
	y := [3]float64{strokePts[0][1], strokePts[1][1], strokePts[2][1]}

	buildPaths := func() {
		// Open zigzag: pt0 → midpt01 → pt1 → pt2 → pt2 (duplicate for stability).
		a.MoveTo(x[0], y[0])
		a.LineTo((x[0]+x[1])/2, (y[0]+y[1])/2)
		a.LineTo(x[1], y[1])
		a.LineTo(x[2], y[2])
		a.LineTo(x[2], y[2])

		// Closed triangle from midpoints.
		a.MoveTo((x[0]+x[1])/2, (y[0]+y[1])/2)
		a.LineTo((x[1]+x[2])/2, (y[1]+y[2])/2)
		a.LineTo((x[2]+x[0])/2, (y[2]+y[0])/2)
		a.ClosePolygon()
	}

	join := strokeJoins[strokeJoin]
	lineCap := strokeCaps[strokeCap]

	// (1) Wide stroked path with selected join/cap.
	a.ResetPath()
	buildPaths()
	a.LineJoin(join)
	a.LineCap(lineCap)
	a.MiterLimit(strokeMiterLimit)
	a.LineWidth(strokeWidth)
	a.LineColor(agg.NewColor(204, 178, 153, 255))
	a.NoFill()
	a.DrawPath(agg.StrokeOnly)

	// (2) Thin outline of the raw path in black.
	a.ResetPath()
	buildPaths()
	a.LineJoin(agg.JoinMiter)
	a.LineCap(agg.CapButt)
	a.LineWidth(1.5)
	a.LineColor(agg.Black)
	a.DrawPath(agg.StrokeOnly)

	// (3) Dashed thin overlay on the wide stroke (matching the C++ poly2).
	a.ResetPath()
	buildPaths()
	a.LineJoin(join)
	a.LineCap(lineCap)
	a.LineWidth(strokeWidth / 5.0)
	a.RemoveAllDashes()
	a.AddDash(20, strokeWidth/2.5)
	a.LineColor(agg.NewColor(0, 0, 77, 255))
	a.DrawPath(agg.StrokeOnly)
	a.RemoveAllDashes()

	// (4) Semi-transparent fill of the raw path.
	a.ResetPath()
	buildPaths()
	a.FillColor(agg.NewColor(0, 0, 0, 51))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	// Interactive handles.
	for i := 0; i < 3; i++ {
		a.FillColor(agg.NewColor(200, 50, 20, 180))
		a.NoLine()
		a.FillCircle(x[i], y[i], 7)
		a.LineColor(agg.Black)
		a.LineWidth(1.0)
		a.DrawCircle(x[i], y[i], 7)
	}
}

// --- Mouse handlers ---

func handleConvStrokeMouseDown(x, y float64) bool {
	strokeSelected = -1
	for i := 0; i < 3; i++ {
		dx := x - strokePts[i][0]
		dy := y - strokePts[i][1]
		if math.Sqrt(dx*dx+dy*dy) < 15 {
			strokeSelected = i
			strokeDragDX = dx
			strokeDragDY = dy
			return true
		}
	}
	return false
}

func handleConvStrokeMouseMove(x, y float64) bool {
	if strokeSelected < 0 {
		return false
	}
	strokePts[strokeSelected][0] = x - strokeDragDX
	strokePts[strokeSelected][1] = y - strokeDragDY
	return true
}

func handleConvStrokeMouseUp() {
	strokeSelected = -1
}
