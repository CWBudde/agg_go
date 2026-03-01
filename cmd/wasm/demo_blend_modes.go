package main

import (
	agg "agg_go"
)

func drawBlendModesDemo() {
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

	cellW, cellH := 200.0, 150.0
	cols := 4

	for i, m := range modes {
		x := float64(i%cols) * cellW
		y := float64(i/cols) * cellH

		// Draw background circles
		agg2d.BlendMode(agg.BlendAlpha)
		agg2d.FillColor(agg.NewColor(255, 0, 0, 180))
		agg2d.FillCircle(x+70, y+60, 40)
		
		agg2d.FillColor(agg.NewColor(0, 255, 0, 180))
		agg2d.FillCircle(x+110, y+60, 40)

		// Draw overlapping circle with the specific blend mode
		agg2d.BlendMode(m.mode)
		agg2d.FillColor(agg.NewColor(0, 0, 255, 255))
		agg2d.FillCircle(x+90, y+90, 40)

		// Reset blend mode for text
		agg2d.BlendMode(agg.BlendAlpha)
		// We'll skip text for now until we ensure font loading in WASM is robust,
		// or use a simple line-based labeling if needed.
	}
}
