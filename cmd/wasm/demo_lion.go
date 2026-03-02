// Based on the original AGG examples: lion.cpp and parse_lion.cpp.
package main

import (
	agg "agg_go"
	"agg_go/internal/basics"
	liondemo "agg_go/internal/demo/lion"
)

type LionPath = liondemo.Path

func drawLionDemo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ResetPath()

	scale := 1.2
	offsetX, offsetY := 250.0, 100.0

	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		agg2d.NoLine()

		agg2d.ResetPath()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			tx, ty := x*scale+offsetX, y*scale+offsetY
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(tx, ty)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
		agg2d.ResetPath()
	}
}
