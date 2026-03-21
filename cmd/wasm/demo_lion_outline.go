// Port of AGG C++ lion_outline.cpp – interactive lion with outline rendering.
//
// The lion vector art is rendered as stroked outlines rather than filled
// polygons. Left-drag rotates and scales; right-drag applies shear.
// A slider controls outline width.
//
// Note on coordinate systems: AGG's original uses flip_y=true (y-up rendering).
// In Go's y-down canvas, rotate(angle+Pi)+flip_y is replaced by
// Scale(-1,1)+Rotate(angle). Centering uses the actual bounding-box centre.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

// --- State ---

var (
	lionOutlineWidth = 1.0

	lionAngle = 0.0
	lionScale = 1.0
	lionSkewX = 0.0
	lionSkewY = 0.0

	lionDragging      = false
	lionRightDragging = false
)

// --- Drawing ---

func drawLionOutlineDemo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	a := ctx.GetAgg2D()

	x1, y1, x2, y2 := getLionBoundingRect(lionPaths)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	a.ResetTransformations()
	a.Translate(-cx, -cy)
	a.Scale(lionScale, lionScale)
	a.Scale(-1, 1)
	a.Rotate(lionAngle)
	a.Skew(lionSkewX/1000.0, lionSkewY/1000.0)
	a.Translate(float64(width)*0.5, float64(height)*0.5)

	a.LineWidth(lionOutlineWidth)
	a.NoFill()

	for _, lp := range lionPaths {
		a.LineColor(agg.NewColor(lp.Color.R, lp.Color.G, lp.Color.B, 255))
		a.ResetPath()

		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				a.MoveTo(x, y)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				a.LineTo(x, y)
			}
		}
		a.ClosePolygon()
		a.DrawPath(agg.StrokeOnly)
	}

	a.ResetTransformations()
}

// --- Mouse handlers ---

func handleLionOutlineMouseDown(x, y float64, right bool) bool {
	if right {
		lionRightDragging = true
		lionSkewX = x
		lionSkewY = y
	} else {
		lionDragging = true
		applyLionTransform(x, y)
	}
	return true
}

func handleLionOutlineMouseMove(x, y float64, right bool) bool {
	if right && lionRightDragging {
		lionSkewX = x
		lionSkewY = y
		return true
	}
	if lionDragging {
		applyLionTransform(x, y)
		return true
	}
	return false
}

func handleLionOutlineMouseUp() {
	lionDragging = false
	lionRightDragging = false
}

func applyLionTransform(x, y float64) {
	cx := float64(width) * 0.5
	cy := float64(height) * 0.5
	dx := x - cx
	dy := y - cy
	lionAngle = math.Atan2(dy, dx)
	lionScale = math.Sqrt(dx*dx+dy*dy) / 100.0
	if lionScale < 0.01 {
		lionScale = 0.01
	}
}
