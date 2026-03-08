// Package main ports AGG's pattern_perspective.cpp demo.
package main

import (
	"flag"
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/demo/patternperspective"
)

func main() {
	var (
		mode = flag.Int("mode", 2, "0=Affine, 1=Bilinear, 2=Perspective")
		out  = flag.String("out", "pattern_perspective.png", "output PNG path")
	)
	flag.Parse()

	ctx := agg.NewContext(800, 600)
	patternperspective.Draw(ctx, patternperspective.Config{
		Mode: *mode,
		Quad: [4][2]float64{{200, 100}, {600, 100}, {600, 500}, {200, 500}},
	})

	if err := renderutil.SavePNG(ctx.GetImage(), *out); err != nil {
		panic(err)
	}
	fmt.Println("saved", *out)
}
