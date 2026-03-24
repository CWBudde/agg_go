package main

import "github.com/MeKo-Christian/agg_go/internal/demo/graphtest"

var graphTestGraph = graphtest.NewGraph(200, 100)

func drawGraphTestDemo() {
	graphtest.Draw(ctx, graphTestGraph, graphtest.Config{
		Mode:        0,
		Width:       2.0,
		Translucent: false,
		DrawNodes:   true,
		DrawEdges:   true,
	})
}
