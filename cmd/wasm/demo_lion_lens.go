// Based on the original AGG examples: lion_lens.cpp.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	lionLensScale        = 3.0
	lionLensRadius       = 70.0
	lionLensX, lionLensY float64
	lionLensInitialized  bool

)

func initLionLensDemo() {
	if lionLensInitialized {
		return
	}

	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	lionLensX = float64(width) * 0.5
	lionLensY = float64(height) * 0.5

	lionLensInitialized = true
}

func drawLionLensDemo() {
	initLionLensDemo()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	// Set up the lens
	lens := transform.NewTransWarpMagnifier()
	lens.SetCenter(lionLensX, lionLensY)
	lens.SetMagnification(lionLensScale)
	lens.SetRadius(lionLensRadius / lionLensScale)

	// Set up the base transformation for the lion
	g_x1, g_y1, g_x2, g_y2 := getLionBoundingRect(lionData)
	base_dx := (g_x2 - g_x1) * 0.5
	base_dy := (g_y2 - g_y1) * 0.5

	mtx := transform.NewTransAffine()
	mtx.Translate(-base_dx, -base_dy)
	// Go has no flip_y rendering; ScaleXY(-1,1) mirrors X to reproduce
	// the same visual as C++ rotate(Pi) + flip_y=true.
	mtx.ScaleXY(-1, 1)
	mtx.Translate(float64(width)*0.5, float64(height)*0.5)

	for i := 0; i < lionData.NPaths; i++ {
		agg2d.ResetPath()

		lionData.Path.Rewind(lionData.PathIdx[i])
		for {
			x, y, cmd := lionData.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			// Apply base transform then lens.
			fx, fy := x, y
			mtx.Transform(&fx, &fy)
			lens.Transform(&fx, &fy)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(fx, fy)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(fx, fy)
			}
		}

		agg2d.FillColor(agg.NewColor(lionData.Colors[i].R, lionData.Colors[i].G, lionData.Colors[i].B, 255))
		agg2d.NoLine()
		agg2d.DrawPath(agg.FillOnly)
	}
}

func getLionBoundingRect(ld *liondemo.LionData) (x1, y1, x2, y2 float64) {
	x1, y1, x2, y2 = 1e100, 1e100, -1e100, -1e100
	first := true
	for idx := uint(0); idx < ld.Path.TotalVertices(); idx++ {
		x, y, cmd := ld.Path.Vertex(idx)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			if first {
				x1, y1, x2, y2 = x, y, x, y
				first = false
			} else {
				if x < x1 {
					x1 = x
				}
				if y < y1 {
					y1 = y
				}
				if x > x2 {
					x2 = x
				}
				if y > y2 {
					y2 = y
				}
			}
		}
	}
	return
}


func setLionLensScale(v float64)  { lionLensScale = v }
func setLionLensRadius(v float64) { lionLensRadius = v }

func handleLionLensMouseDown(x, y float64) bool {
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseMove(x, y float64) bool {
	lionLensX = x
	lionLensY = y
	return true
}

func handleLionLensMouseUp() {}
