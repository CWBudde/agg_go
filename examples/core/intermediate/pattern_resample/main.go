// Package main ports AGG's pattern_resample.cpp demo.
package main

import (
	"flag"
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/demo/patternresample"
)

func main() {
	var (
		mode  = flag.Int("mode", 4, "0=AffineNoResample,1=AffineResample,2=PerspNoResampleLerp,3=PerspNoResampleExact,4=PerspResampleLerp,5=PerspResampleExact")
		gamma = flag.Float64("gamma", 2.0, "gamma (0.5..3.0)")
		blur  = flag.Float64("blur", 1.0, "blur (0.5..2.0)")
		out   = flag.String("out", "pattern_resample.png", "output PNG path")
	)
	flag.Parse()

	ctx := agg.NewContext(800, 600)
	patternresample.Draw(ctx, patternresample.Config{
		Mode:  *mode,
		Gamma: *gamma,
		Blur:  *blur,
		Quad:  [4][2]float64{{200, 100}, {600, 100}, {600, 500}, {200, 500}},
	})

	if err := renderutil.SavePNG(ctx.GetImage(), *out); err != nil {
		panic(err)
	}
	fmt.Println("saved", *out)
}
