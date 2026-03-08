// Package main ports AGG's image_perspective.cpp demo.
package main

import (
	"flag"
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/demo/imageperspective"
)

func main() {
	var (
		mode = flag.Int("mode", 2, "0=AffineParallelogram, 1=Bilinear, 2=Perspective")
		out  = flag.String("out", "image_perspective.png", "output PNG path")
	)
	flag.Parse()

	ctx := agg.NewContext(800, 600)
	imageperspective.Draw(ctx, imageperspective.Config{
		Mode: *mode,
		Quad: [4][2]float64{{100, 100}, {700, 100}, {700, 500}, {100, 500}},
	})

	if err := renderutil.SavePNG(ctx.GetImage(), *out); err != nil {
		panic(err)
	}
	fmt.Println("saved", *out)
}
