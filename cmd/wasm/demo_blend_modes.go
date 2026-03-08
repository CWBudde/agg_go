// Blend mode gallery demo (separate from compositing.cpp direct port).
package main

import (
	agg "agg_go"
)

func drawBlendModesDemo() {
	ctx.Clear(agg.RGBA(0.9, 0.9, 0.9, 1.0))

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	modes := []struct {
		name string
		mode agg.BlendMode
	}{
		{"Alpha", agg.BlendAlpha},
		{"Multiply", agg.BlendMultiply},
		{"Screen", agg.BlendScreen},
		{"Overlay", agg.BlendOverlay},
		{"Darken", agg.BlendDarken},
		{"Lighten", agg.BlendLighten},
		{"ColorDodge", agg.BlendColorDodge},
		{"ColorBurn", agg.BlendColorBurn},
		{"HardLight", agg.BlendHardLight},
		{"SoftLight", agg.BlendSoftLight},
		{"Difference", agg.BlendDifference},
		{"Exclusion", agg.BlendExclusion},
	}

	const cellW, cellH = 200.0, 150.0
	const cols = 4

	for i, m := range modes {
		x := float64(i%cols) * cellW
		y := float64(i/cols) * cellH

		agg2d.BlendMode(agg.BlendAlpha)
		agg2d.FillColor(agg.RGBA(1.0, 1.0, 1.0, 0.6))
		agg2d.NoLine()
		agg2d.ResetPath()
		agg2d.MoveTo(x+4, y+4)
		agg2d.LineTo(x+cellW-4, y+4)
		agg2d.LineTo(x+cellW-4, y+cellH-4)
		agg2d.LineTo(x+4, y+cellH-4)
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)

		agg2d.FillColor(agg.NewColor(255, 0, 0, 180))
		agg2d.FillCircle(x+70, y+60, 40)
		agg2d.FillColor(agg.NewColor(0, 255, 0, 180))
		agg2d.FillCircle(x+110, y+60, 40)

		agg2d.BlendMode(m.mode)
		agg2d.FillColor(agg.NewColor(0, 0, 255, 200))
		agg2d.FillCircle(x+90, y+100, 40)

		_ = m.name
	}

	agg2d.BlendMode(agg.BlendAlpha)
}
