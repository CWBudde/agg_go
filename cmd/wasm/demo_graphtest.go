package main

import "github.com/cwbudde/agg_go/internal/demo/graphtest"

var graphTestGraph = graphtest.NewGraph(200, 100)

var (
	graphTestMode        = 1
	graphTestWidth       = 2.0
	graphTestTranslucent = false
	graphTestDrawNodes   = true
	graphTestDrawEdges   = true
)

func setGraphTestMode(v int)         { graphTestMode = v }
func setGraphTestWidth(v float64)    { graphTestWidth = v }
func setGraphTestTranslucent(v bool) { graphTestTranslucent = v }
func setGraphTestDrawNodes(v bool)   { graphTestDrawNodes = v }
func setGraphTestDrawEdges(v bool)   { graphTestDrawEdges = v }

func drawGraphTestDemo() {
	graphtest.Draw(ctx, graphTestGraph, graphtest.Config{
		Mode:        graphTestMode,
		Width:       graphTestWidth,
		Translucent: graphTestTranslucent,
		DrawNodes:   graphTestDrawNodes,
		DrawEdges:   graphTestDrawEdges,
	})
}
