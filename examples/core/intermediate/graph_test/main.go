// Package main ports AGG's graph_test.cpp demo.
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/demo/graphtest"
)

type demo struct {
	graph *graphtest.Graph
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	graphtest.Draw(ctx, d.graph, graphtest.Config{
		Mode:         0,
		Width:        2.0,
		Translucent:  false,
		DrawNodes:    true,
		DrawEdges:    true,
		ShowControls: true,
	})
}

func main() {
	d := &demo{graph: graphtest.NewGraph(200, 100)}
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Graph Test",
		Width:                 700,
		Height:                530,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, d)
}
