// Port of AGG C++ lion.cpp – classic lion demo with alpha, rotate/scale, skew.
//
// Left-drag rotates and scales; right-drag applies shear.
// An alpha slider controls global opacity of all paths.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	liondemo "agg_go/internal/demo/lion"
)

type LionPath = liondemo.Path

// Lion bounding box centre derived from the parse_lion data (bbox ≈ 7..557 × 8..520).
const (
	lionFillBaseDX = (557.0 - 7.0) * 0.5 // ≈ 275
	lionFillBaseDY = (520.0 - 8.0) * 0.5 // ≈ 256
)

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

	// Set up the affine transform matching the C++ matrix composition:
	//   translate(-baseDX, -baseDY)  → centre lion on origin
	//   scale(scale)
	//   rotate(angle + π)            → +π corrects y-down orientation
	//   skew(skewX/1000, skewY/1000)
	//   translate(w/2, h/2)          → centre on canvas
	a.ResetTransformations()
	a.Translate(-lionFillBaseDX, -lionFillBaseDY)
	a.Scale(lionFillScale, lionFillScale)
	a.Rotate(lionFillAngle + math.Pi)
	a.Skew(lionFillSkewX/1000.0, lionFillSkewY/1000.0)
	a.Translate(float64(width)/2, float64(height)/2)

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
