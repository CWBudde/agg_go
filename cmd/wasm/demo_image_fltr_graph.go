package main

import (
	"github.com/MeKo-Christian/agg_go/internal/demo/imagefltrgraph"
)

var imageFltrGraphState = imagefltrgraph.DefaultState()

func setImageFltrGraphRadius(v float64) {
	imageFltrGraphState.Radius = v
	imageFltrGraphState.Clamp()
}

func setImageFltrGraphMask(mask uint32) {
	imageFltrGraphState.SetMask(mask)
}

func drawImageFltrGraphDemo() {
	imagefltrgraph.Draw(ctx, imageFltrGraphState)
}
