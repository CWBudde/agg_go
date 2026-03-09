// Port of AGG C++ bspline.cpp – B-Spline Interpolation interactive demo.
package main

import (
	"math"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
)

// --- State ---

var (
	bsplinePts = [6][2]float64{} // 6 draggable control points (initialized lazily)

	bsplineClosed    = false
	bsplineNumPoints = 20.0 // interpolation quality (1–40); same as m_num_points default

	bsplineSelected = -1
	bsplineDragDX   = 0.0
	bsplineDragDY   = 0.0
)

// bsplineInit seeds the control points matching the C++ on_init() (flip_y=false).
func bsplineInit() {
	w := float64(width)
	h := float64(height)
	bsplinePts[0] = [2]float64{100, 100}
	bsplinePts[1] = [2]float64{w - 100, 100}
	bsplinePts[2] = [2]float64{w - 100, h - 100}
	bsplinePts[3] = [2]float64{100, h - 100}
	bsplinePts[4] = [2]float64{w / 2, h / 2}
	bsplinePts[5] = [2]float64{w / 2, h / 3}
}

func init() {
	bsplineInit()
}

// --- Drawing ---

func drawBSplineDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Flat coordinate slice required by SimplePolygonVertexSource.
	coords := make([]float64, len(bsplinePts)*2)
	for i, p := range bsplinePts {
		coords[i*2] = p[0]
		coords[i*2+1] = p[1]
	}

	// 1. B-spline curve via conv_bspline (the real thing, not a Bezier hack).
	src := polygon.NewSimplePolygonVertexSource(coords, uint(len(bsplinePts)), false, bsplineClosed)
	bspline := conv.NewConvBSpline(src)
	bspline.SetInterpolationStep(1.0 / bsplineNumPoints)

	a.ResetPath()
	bspline.Rewind(0)
	for {
		x, y, cmd := bspline.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		switch {
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsVertex(cmd):
			a.LineTo(x, y)
		case basics.IsClosed(uint32(cmd)):
			a.ClosePolygon()
		}
	}
	a.LineColor(agg.Black)
	a.LineWidth(2.0)
	a.NoFill()
	a.DrawPath(agg.StrokeOnly)

	// 2. Control polygon – translucent blue, matching C++ rgba(0, 0.3, 0.5, 0.6).
	a.ResetPath()
	a.MoveTo(bsplinePts[0][0], bsplinePts[0][1])
	for i := 1; i < len(bsplinePts); i++ {
		a.LineTo(bsplinePts[i][0], bsplinePts[i][1])
	}
	if bsplineClosed {
		a.ClosePolygon()
	}
	a.LineColor(agg.RGBA(0, 0.3, 0.5, 0.6))
	a.LineWidth(1.0)
	a.NoFill()
	a.DrawPath(agg.StrokeOnly)

	// 3. Draggable handle circles at each control point.
	for _, p := range bsplinePts {
		drawHandle(p[0], p[1])
	}
}

// --- Mouse handlers ---

func handleBSplineMouseDown(x, y float64) bool {
	bsplineSelected = -1
	for i, p := range bsplinePts {
		dx := x - p[0]
		dy := y - p[1]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			bsplineSelected = i
			bsplineDragDX = dx
			bsplineDragDY = dy
			return true
		}
	}
	return false
}

func handleBSplineMouseMove(x, y float64) bool {
	if bsplineSelected < 0 {
		return false
	}
	bsplinePts[bsplineSelected][0] = x - bsplineDragDX
	bsplinePts[bsplineSelected][1] = y - bsplineDragDY
	return true
}

func handleBSplineMouseUp() {
	bsplineSelected = -1
}
