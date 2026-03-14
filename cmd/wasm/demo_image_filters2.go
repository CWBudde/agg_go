package main

import "github.com/MeKo-Christian/agg_go/internal/demo/imagefilters2"

var imageFilters2State = imagefilters2.DefaultState()

func setImageFilters2Filter(v int) {
	imageFilters2State.FilterIdx = v
	imageFilters2State.Clamp()
}

func setImageFilters2Radius(v float64) {
	imageFilters2State.Radius = v
	imageFilters2State.Clamp()
}

func setImageFilters2Normalize(v bool) {
	imageFilters2State.Normalize = v
}

func drawImageFilters2Demo() {
	imagefilters2.Draw(ctx, imageFilters2State)
}
