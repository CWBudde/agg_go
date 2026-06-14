package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/demo/molview"
)

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

	// molview.Draw was written for a y-up coordinate system, matching the
	// standalone example which uses FlipY: true.  The WASM canvas is y-down,
	// so we attach the canvas buffer with a negative stride (bottom-up) to give
	// the renderer the y-up frame it expects.  The resulting pixel data in
	// canvasBuf ends up in the correct top-down screen order for putImageData.
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()
	flippedImg := agg.NewImage(img.Data, w, h, -img.Stride())
	flippedCtx := agg.NewContextForImage(flippedImg)
	molview.Draw(flippedCtx, molViewState)

	applyLinearToSRGB(img)
}

// molViewFlipY converts a canvas Y coordinate (y-down, origin at top) to the
// y-up coordinate system that molview.Draw uses.
func molViewFlipY(y float64) float64 {
	return float64(ctx.GetImage().Height()) - 1.0 - y
}

func handleMolViewMouseDown(x, y float64, right bool) bool {
	molview.BeginDrag(&molViewState, &molViewDrag, x, molViewFlipY(y), right)
	return true
}

func handleMolViewMouseMove(x, y float64, right bool) bool {
	return molview.UpdateDrag(&molViewState, &molViewDrag, x, molViewFlipY(y), right)
}

func handleMolViewMouseUp() {
	molview.EndDrag(&molViewDrag)
}
