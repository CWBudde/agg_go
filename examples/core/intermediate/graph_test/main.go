// Package main ports AGG's graph_test.cpp demo.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/graphtest"
)

type demo struct {
	graph *graphtest.Graph
}

func (d *demo) Render(ctx *agg.Context) {
	graphtest.Draw(ctx, d.graph, graphtest.Config{
		Mode:        1,
		Width:       2.0,
		Translucent: true,
		DrawNodes:   true,
		DrawEdges:   true,
	})
}

func main() {
	d := &demo{graph: graphtest.NewGraph(200, 100)}
	demorunner.Run(demorunner.Config{
		Title:  "Graph Test",
		Width:  700,
		Height: 530,
	}, d)
}
