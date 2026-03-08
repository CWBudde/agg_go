package main

import "agg_go/internal/demo/interactivepolygon"

var interactivePolygonState *interactivepolygon.State

func ensureInteractivePolygonState() {
	if interactivePolygonState == nil {
		interactivePolygonState = interactivepolygon.NewState(float64(width), float64(height))
	}
}

func handleInteractivePolygonMouseDown(x, y float64) bool {
	ensureInteractivePolygonState()
	return interactivePolygonState.MouseDown(x, y)
}

func handleInteractivePolygonMouseMove(x, y float64) bool {
	ensureInteractivePolygonState()
	return interactivePolygonState.MouseMove(x, y, true)
}

func handleInteractivePolygonMouseUp() bool {
	ensureInteractivePolygonState()
	return interactivePolygonState.MouseUp(0, 0)
}

func drawInteractivePolygonDemo() {
	ensureInteractivePolygonState()
	interactivePolygonState.Draw(ctx)
}
