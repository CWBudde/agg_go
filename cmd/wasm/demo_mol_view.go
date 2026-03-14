package main

import "github.com/MeKo-Christian/agg_go/internal/demo/molview"

var (
	molViewState = molview.DefaultState()
	molViewDrag  molview.DragState
)

func setMolViewMolecule(v int) {
	molViewState.MoleculeIdx = v
	molViewState.Clamp()
}

func setMolViewThickness(v float64) {
	molViewState.Thickness = v
	molViewState.Clamp()
}

func setMolViewTextSize(v float64) {
	molViewState.TextSize = v
	molViewState.Clamp()
}

func setMolViewAutoRotate(v bool) {
	molViewState.AutoRotate = v
}

func drawMolViewDemo() {
	molViewState.Advance()
	molview.Draw(ctx, molViewState)
}

func handleMolViewMouseDown(x, y float64, right bool) bool {
	molview.BeginDrag(&molViewState, &molViewDrag, x, y, right)
	return true
}

func handleMolViewMouseMove(x, y float64, right bool) bool {
	return molview.UpdateDrag(&molViewState, &molViewDrag, x, y, right)
}

func handleMolViewMouseUp() {
	molview.EndDrag(&molViewDrag)
}
