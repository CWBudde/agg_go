// Port of AGG C++ lion.cpp – classic lion demo with alpha, rotate/scale, skew.
//
// Left-drag rotates and scales; right-drag applies shear.
// An alpha slider controls global opacity of all paths.
//
// Note on coordinate systems: AGG's original example uses flip_y=true (y-up
// rendering). In Go's y-down canvas, rotate(angle+Pi)+flip_y is replaced by
// Scale(-1,1)+Rotate(angle), which produces the same visual result.
// Centering uses the actual bounding-box centre, not just the half-extents.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

type LionPath = liondemo.Path

var (
	lionFillAlpha         = 1.0
	lionFillAngle         = 0.0
	lionFillScale         = 1.0
	lionFillSkewX         = 0.0
	lionFillSkewY         = 0.0
	lionFillDragging      = false
	lionFillRightDragging = false
)

func drawLionDemo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	a := ctx.GetAgg2D()

	// Compute the true bounding-box centre so the lion rotates around its own
	// centre and is centred on the canvas in the default (angle=0) state.
	x1, y1, x2, y2 := getLionBoundingRect(lionPaths)
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5

	// Transform chain (mirrors the C++ compose order):
	//   translate(-cx, -cy)   – move lion centre to origin
	//   scale(s, s)           – uniform scale
	//   scale(-1, 1)          – x-mirror: equivalent to C++ rotate(Pi)+flip_y
	//   rotate(angle)         – interactive rotation
	//   skew(sx/1000, sy/1000)
	//   translate(W/2, H/2)   – move to canvas centre
	a.ResetTransformations()
	a.Translate(-cx, -cy)
	a.Scale(lionFillScale, lionFillScale)
	a.Scale(-1, 1)
	a.Rotate(lionFillAngle)
	a.Skew(lionFillSkewX/1000.0, lionFillSkewY/1000.0)
	a.Translate(float64(width)*0.5, float64(height)*0.5)

	alpha := uint8(lionFillAlpha * 255)
	a.NoLine()

	for _, lp := range lionPaths {
		a.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], alpha))
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
		a.DrawPath(agg.FillOnly)
		a.ResetPath()
	}

	a.ResetTransformations()
}

// --- Mouse handlers ---

func handleLionMouseDown(x, y float64, right bool) bool {
	if right {
		lionFillRightDragging = true
		lionFillSkewX = x
		lionFillSkewY = y
	} else {
		lionFillDragging = true
		applyLionFillTransform(x, y)
	}
	return true
}

func handleLionMouseMove(x, y float64, right bool) bool {
	if right && lionFillRightDragging {
		lionFillSkewX = x
		lionFillSkewY = y
		return true
	}
	if lionFillDragging {
		applyLionFillTransform(x, y)
		return true
	}
	return false
}

func handleLionMouseUp() {
	lionFillDragging = false
	lionFillRightDragging = false
}

func applyLionFillTransform(x, y float64) {
	dx := x - float64(width)*0.5
	dy := y - float64(height)*0.5
	lionFillAngle = math.Atan2(dy, dx)
	lionFillScale = math.Sqrt(dx*dx+dy*dy) / 100.0
	if lionFillScale < 0.01 {
		lionFillScale = 0.01
	}
}
