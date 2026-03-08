// Package main ports AGG's pattern_resample.cpp demo.
package main

import (
	"flag"
	"math"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/demo/patternresample"
)

type demo struct {
	mode    int
	gamma   float64
	blur    float64
	quad    [4][2]float64
	dragIdx int
}

const handleRadius = 8.0

func (d *demo) Render(ctx *agg.Context) {
	patternresample.Draw(ctx, patternresample.Config{
		Mode:  d.mode,
		Gamma: d.gamma,
		Blur:  d.blur,
		Quad:  d.quad,
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
	mode := flag.Int("mode", 4, "0=AffineNoResample,1=AffineResample,2=PerspNoResampleLerp,3=PerspNoResampleExact,4=PerspResampleLerp,5=PerspResampleExact")
	gamma := flag.Float64("gamma", 2.0, "gamma (0.5..3.0)")
	blur := flag.Float64("blur", 1.0, "blur (0.5..2.0)")
	flag.Parse()

	d := &demo{
		mode:    *mode,
		gamma:   *gamma,
		blur:    *blur,
		quad:    [4][2]float64{{200, 100}, {600, 100}, {600, 500}, {200, 500}},
		dragIdx: -1,
	}

	demorunner.Run(demorunner.Config{
		Title:  "Pattern Resample",
		Width:  800,
		Height: 600,
	}, d)
}
