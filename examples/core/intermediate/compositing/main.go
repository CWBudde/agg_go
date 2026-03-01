package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
)

func main() {
	ctx := agg.NewContext(960, 700)
	ctx.Clear(agg.RGB(0.97, 0.97, 0.96))
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
		{"Difference", agg.BlendDifference},
		{"Exclusion", agg.BlendExclusion},
	}

	for i, m := range modes {
		col := i % 4
		row := i / 4
		x := 70.0 + float64(col)*210.0
		y := 90.0 + float64(row)*280.0

		agg2d.BlendMode(agg.BlendAlpha)
		agg2d.FillColor(agg.NewColor(255, 0, 0, 180))
		agg2d.FillCircle(x+55, y+65, 46)
		agg2d.FillColor(agg.NewColor(0, 255, 0, 180))
		agg2d.FillCircle(x+110, y+65, 46)

		agg2d.BlendMode(m.mode)
		agg2d.FillColor(agg.NewColor(0, 0, 255, 255))
		agg2d.FillCircle(x+82, y+108, 46)
	}

	const filename = "compositing.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
