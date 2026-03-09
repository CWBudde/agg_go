// Port of AGG C++ rounded_rect.cpp – interactive rounded rectangle.
//
// Two draggable control points define the opposite corners of a rectangle.
// Sliders control the corner radius and a sub-pixel offset.
// A checkbox toggles white-on-black rendering.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
)

// --- State ---

var (
	rrPts = [2][2]float64{
		{100, 100},
		{500, 350},
	}
	rrRadius   = 25.0
	rrOffset   = 0.0
	rrDarkBg   = false
	rrSelected = -1
	rrDragDX   = 0.0
	rrDragDY   = 0.0
)

// --- Drawing ---

func drawRoundedRectDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Background
	if rrDarkBg {
		ctx.Clear(agg.Black)
	} else {
		ctx.Clear(agg.White)
	}

	// Small circles at the two control points
	fg := agg.NewColor(127, 127, 127, 255)
	a.FillColor(fg)
	a.NoLine()
	for _, pt := range rrPts {
		a.ResetPath()
		a.Ellipse(pt[0], pt[1], 5, 5)
		a.DrawPath(agg.FillOnly)
	}

	// Rounded rectangle with optional sub-pixel offset
	x1 := rrPts[0][0] + rrOffset
	y1 := rrPts[0][1] + rrOffset
	x2 := rrPts[1][0] + rrOffset
	y2 := rrPts[1][1] + rrOffset

	if rrDarkBg {
		a.LineColor(agg.White)
	} else {
		a.LineColor(agg.Black)
	}
	a.NoFill()
	a.LineWidth(1.0)
	a.ResetPath()
	a.RoundedRect(x1, y1, x2, y2, rrRadius)
	a.DrawPath(agg.StrokeOnly)
}

// --- Mouse handlers ---

func handleRoundedRectMouseDown(x, y float64) bool {
	for i, pt := range rrPts {
		dx := x - pt[0]
		dy := y - pt[1]
		if math.Sqrt(dx*dx+dy*dy) < 8.0 {
			rrSelected = i
			rrDragDX = dx
			rrDragDY = dy
			return true
		}
	}
	return false
}

func handleRoundedRectMouseMove(x, y float64) bool {
	if rrSelected < 0 {
		return false
	}
	rrPts[rrSelected][0] = x - rrDragDX
	rrPts[rrSelected][1] = y - rrDragDY
	return true
}

func handleRoundedRectMouseUp() {
	rrSelected = -1
}
