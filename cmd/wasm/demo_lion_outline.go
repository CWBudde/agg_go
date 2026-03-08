// Port of AGG C++ lion_outline.cpp – interactive lion with outline rendering.
//
// The lion vector art is rendered as stroked outlines rather than filled
// polygons. Left-drag rotates and scales; right-drag applies shear.
// A slider controls outline width.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	liondemo "agg_go/internal/demo/lion"
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

// Lion bounding box centre, matching the original parse_lion data.
// Computed once from the lion path data (approx. bbox 7..557 × 8..520).
const (
	lionBaseDX = (557.0 - 7.0) * 0.5 // ≈ 275
	lionBaseDY = (520.0 - 8.0) * 0.5 // ≈ 256
)

// --- Drawing ---

func drawLionOutlineDemo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	a := ctx.GetAgg2D()

	// Set up the affine transform matching the C++ matrix composition:
	//   translate(-baseDX, -baseDY)  → centre lion on origin
	//   scale(scale)
	//   rotate(angle + π)            → +π corrects y-down orientation
	//   skew(skewX/1000, skewY/1000)
	//   translate(w/2, h/2)          → centre on canvas
	a.ResetTransformations()
	a.Translate(-lionBaseDX, -lionBaseDY)
	a.Scale(lionScale, lionScale)
	a.Rotate(lionAngle + math.Pi)
	a.Skew(lionSkewX/1000.0, lionSkewY/1000.0)
	a.Translate(float64(width)/2, float64(height)/2)

	a.LineWidth(lionOutlineWidth)
	a.NoFill()

	for _, lp := range lionPaths {
		a.LineColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
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
