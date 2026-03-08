// Package main ports AGG's line_patterns.cpp demo (image-patterned Bezier curves).
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/internal/demo/linepatterns"
)

func main() {
	const (
		w       = 500
		h       = 450
		scaleX  = 1.0
		startX  = 0.0
		outFile = "line_patterns.png"
	)

	ctx := agg.NewContext(w, h)
	linepatterns.Draw(ctx.GetImage(), scaleX, startX)

	if err := ctx.GetImage().SaveToPNG(outFile); err != nil {
		fmt.Printf("error writing %s: %v\n", outFile, err)
		return
	}
	fmt.Printf("wrote %s (%dx%d)\n", outFile, w, h)
}
