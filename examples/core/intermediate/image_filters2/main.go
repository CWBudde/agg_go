// Package main ports AGG's image_filters2.cpp demo.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/imagefilters2"
)

type demo struct {
	state imagefilters2.State
}

func newDemo() *demo {
	return &demo{state: imagefilters2.DefaultState()}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	imagefilters2.Draw(ctx, d.state)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Filters 2",
		Width:  500,
		Height: 340,
	}, newDemo())
}
