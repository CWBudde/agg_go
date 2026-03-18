// Package main ports AGG's pattern_perspective.cpp demo.
package main

import (
	"flag"
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/demo/patternperspective"
)

type demo struct {
	mode    int
	quad    [4][2]float64
	dragIdx int
}

const handleRadius = 8.0

func (d *demo) Render(ctx *agg.Context) {
	patternperspective.Draw(ctx, patternperspective.Config{
		Mode: d.mode,
		Quad: d.quad,
	})
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	for i, pt := range d.quad {
		dx := fx - pt[0]
		dy := fy - pt[1]
		if math.Sqrt(dx*dx+dy*dy) <= handleRadius {
			d.dragIdx = i
			return true
		}
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	if d.dragIdx >= 0 {
		d.dragIdx = -1
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	if d.dragIdx < 0 || !btn.Left {
		return false
	}
	d.quad[d.dragIdx] = [2]float64{float64(x), float64(y)}
	return true
}

func main() {
	mode := flag.Int("mode", 2, "0=Affine, 1=Bilinear, 2=Perspective")
	flag.Parse()

	d := &demo{
		mode:    *mode,
		quad:    [4][2]float64{{200, 100}, {600, 100}, {600, 500}, {200, 500}},
		dragIdx: -1,
	}

	demorunner.Run(demorunner.Config{
		Title:  "Pattern Perspective",
		Width:  600,
		Height: 600,
	}, d)
}
