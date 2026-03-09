package main

import "github.com/MeKo-Christian/agg_go/internal/demo/graphtest"

var graphTestGraph = graphtest.NewGraph(200, 100)

func drawGraphTestDemo() {
	graphtest.Draw(ctx, graphTestGraph, graphtest.Config{
		Mode:        1,
		Width:       2.0,
		Translucent: true,
		DrawNodes:   true,
		DrawEdges:   true,
	})
}
