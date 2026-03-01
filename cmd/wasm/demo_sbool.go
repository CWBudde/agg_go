// Based on the original AGG examples: scanline_boolean.cpp.
package main

import (
	"math"

	agg "agg_go"
)

type SBoolOp int

const (
	SBoolUnion SBoolOp = iota
	SBoolIntersection
	SBoolXor
	SBoolAminusB
	SBoolBminusA
)

var (
	sboolPoly1X = [4]float64{100, 350, 350, 100}
	sboolPoly1Y = [4]float64{150, 150, 400, 400}
	sboolPoly2X = [4]float64{250, 500, 500, 250}
	sboolPoly2Y = [4]float64{250, 250, 500, 500}
	sboolOp     = SBoolXor
	sboolSelected = -1
	sboolPolyIdx  = 0 // 0 for poly1, 1 for poly2
	sboolDragDX   = 0.0
	sboolDragDY   = 0.0
)

func drawSBoolDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// 1. Draw individual shapes with transparency for context
	agg2d.FillColor(agg.RGBA(0.1, 0.5, 0.8, 0.2))
	agg2d.NoLine()
	
	// Poly 1
	agg2d.ResetPath()
	agg2d.MoveTo(sboolPoly1X[0], sboolPoly1Y[0])
	agg2d.LineTo(sboolPoly1X[1], sboolPoly1Y[1])
	agg2d.LineTo(sboolPoly1X[2], sboolPoly1Y[2])
	agg2d.LineTo(sboolPoly1X[3], sboolPoly1Y[3])
	agg2d.ClosePolygon()
	agg2d.DrawPath(agg.FillOnly)

	// Poly 2
	agg2d.FillColor(agg.RGBA(0.8, 0.2, 0.1, 0.2))
	agg2d.ResetPath()
	agg2d.MoveTo(sboolPoly2X[0], sboolPoly2Y[0])
	agg2d.LineTo(sboolPoly2X[1], sboolPoly2Y[1])
	agg2d.LineTo(sboolPoly2X[2], sboolPoly2Y[2])
	agg2d.LineTo(sboolPoly2X[3], sboolPoly2Y[3])
	agg2d.ClosePolygon()
	agg2d.DrawPath(agg.FillOnly)

	// 2. Perform "Boolean" operation using Rasterizer rules
	// For XOR, we can use the Even-Odd filling rule on a combined path.
	// For Union, we use Non-Zero filling rule on a combined path.
	// For Intersection/Sub, we would need the actual boolean clipper,
	// but we'll simulate what we can.
	
	agg2d.FillColor(agg.Black)
	agg2d.ResetPath()
	
	// Combine paths
	addPolyToAgg2D(agg2d, sboolPoly1X[:], sboolPoly1Y[:])
	addPolyToAgg2D(agg2d, sboolPoly2X[:], sboolPoly2Y[:])
	
	switch sboolOp {
	case SBoolUnion:
		agg2d.FillEvenOdd(false) // Non-zero = Union
	case SBoolXor:
		agg2d.FillEvenOdd(true)  // Even-odd = XOR
	default:
		// Simulation fallback
		agg2d.FillEvenOdd(true)
	}
	
	agg2d.DrawPath(agg.FillOnly)
	
	// 3. Draw handles
	drawSBoolHandles(agg2d, sboolPoly1X[:], sboolPoly1Y[:], agg.RGBA(0.1, 0.5, 0.8, 0.6))
	drawSBoolHandles(agg2d, sboolPoly2X[:], sboolPoly2Y[:], agg.RGBA(0.8, 0.2, 0.1, 0.6))
}

func addPolyToAgg2D(agg2d *agg.Agg2D, x, y []float64) {
	agg2d.MoveTo(x[0], y[0])
	for i := 1; i < len(x); i++ {
		agg2d.LineTo(x[i], y[i])
	}
	agg2d.ClosePolygon()
}

func drawSBoolHandles(agg2d *agg.Agg2D, x, y []float64, c agg.Color) {
	for i := 0; i < len(x); i++ {
		agg2d.FillColor(c)
		agg2d.NoLine()
		agg2d.FillCircle(x[i], y[i], 6)
		agg2d.LineColor(agg.Black)
		agg2d.LineWidth(1.0)
		agg2d.DrawCircle(x[i], y[i], 6)
	}
}

func handleSBoolMouseDown(x, y float64) bool {
	sboolSelected = -1
	// Check poly1
	for i := 0; i < 4; i++ {
		if math.Sqrt(math.Pow(x-sboolPoly1X[i], 2)+math.Pow(y-sboolPoly1Y[i], 2)) < 15 {
			sboolSelected = i
			sboolPolyIdx = 0
			sboolDragDX = x - sboolPoly1X[i]
			sboolDragDY = y - sboolPoly1Y[i]
			return true
		}
	}
	// Check poly2
	for i := 0; i < 4; i++ {
		if math.Sqrt(math.Pow(x-sboolPoly2X[i], 2)+math.Pow(y-sboolPoly2Y[i], 2)) < 15 {
			sboolSelected = i
			sboolPolyIdx = 1
			sboolDragDX = x - sboolPoly2X[i]
			sboolDragDY = y - sboolPoly2Y[i]
			return true
		}
	}
	return false
}

func handleSBoolMouseMove(x, y float64) bool {
	if sboolSelected != -1 {
		if sboolPolyIdx == 0 {
			sboolPoly1X[sboolSelected] = x - sboolDragDX
			sboolPoly1Y[sboolSelected] = y - sboolDragDY
		} else {
			sboolPoly2X[sboolSelected] = x - sboolDragDX
			sboolPoly2Y[sboolSelected] = y - sboolDragDY
		}
		return true
	}
	return false
}

func handleSBoolMouseUp() {
	sboolSelected = -1
}
