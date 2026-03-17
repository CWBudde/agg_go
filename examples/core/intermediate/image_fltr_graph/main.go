// Package main ports AGG's image_fltr_graph.cpp demo.
//
// It compares interpolation filter shapes by plotting:
// - raw filter function (red),
// - unnormalized discrete sum response (green),
// - normalized LUT weights (blue).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/imagefltrgraph"
)

type demo struct {
	state imagefltrgraph.State
}

func newDemo() *demo {
	st := imagefltrgraph.DefaultState()
	// Enable a few filters by default to make the comparison visible.
	st.Enabled[0] = true  // bilinear
	st.Enabled[1] = true  // bicubic
	st.Enabled[14] = true // lanczos
	return &demo{state: st}
}

func (d *demo) Render(ctx *agg.Context) {
	imagefltrgraph.Draw(ctx, d.state)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Image Filter Graph",
		Width:  400,
		Height: 320,
	}, newDemo())
}
