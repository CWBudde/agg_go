// Package main ports AGG's scanline_boolean2.cpp demo (URL/control-ready variant).
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/internal/demo/scanlineboolean2"
)

func main() {
	const (
		w       = 655
		h       = 520
		outFile = "scanline_boolean2.png"
	)
	ctx := agg.NewContext(w, h)
	scanlineboolean2.Draw(ctx, scanlineboolean2.Config{
		Mode:      3,
		FillRule:  1,
		Operation: 2,
		CenterX:   float64(w) / 2,
		CenterY:   float64(h) / 2,
	})
	if err := ctx.GetImage().SaveToPNG(outFile); err != nil {
		fmt.Printf("error writing %s: %v\n", outFile, err)
		return
	}
	fmt.Printf("wrote %s (%dx%d)\n", outFile, w, h)
}
