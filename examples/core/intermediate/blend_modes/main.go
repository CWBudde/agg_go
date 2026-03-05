// Port of AGG C++ compositing.cpp – blend mode demonstration.
//
// Shows all major Porter-Duff / SVG blend modes by rendering three overlapping
// circles (red, green, blue) in a 4×3 grid. Each cell uses a different blend
// mode for the third (blue) circle.
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
)

func main() {
	const width, height = 800, 600

	ctx := agg.NewContext(width, height)
	ctx.Clear(agg.RGBA(0.9, 0.9, 0.9, 1.0))

	a := ctx.GetAgg2D()
	a.ResetTransformations()

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

		// Light cell background.
		a.BlendMode(agg.BlendAlpha)
		a.FillColor(agg.RGBA(1.0, 1.0, 1.0, 0.6))
		a.NoLine()
		a.ResetPath()
		a.MoveTo(x+4, y+4)
		a.LineTo(x+cellW-4, y+4)
		a.LineTo(x+cellW-4, y+cellH-4)
		a.LineTo(x+4, y+cellH-4)
		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)

		// Red + green circles with normal blending.
		a.BlendMode(agg.BlendAlpha)
		a.FillColor(agg.NewColor(255, 0, 0, 180))
		a.FillCircle(x+70, y+60, 40)
		a.FillColor(agg.NewColor(0, 255, 0, 180))
		a.FillCircle(x+110, y+60, 40)

		// Blue circle with the demo blend mode.
		a.BlendMode(m.mode)
		a.FillColor(agg.NewColor(0, 0, 255, 200))
		a.FillCircle(x+90, y+100, 40)

		_ = m.name // label would require text rendering
	}

	// Restore.
	a.BlendMode(agg.BlendAlpha)

	const filename = "blend_modes.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
