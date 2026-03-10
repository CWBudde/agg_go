// Package main ports AGG's blend_color.cpp demo.
//
// The demo renders a letter "a" glyph shadow with perspective distortion,
// blurred and composited using either a single color or a gradient LUT.
// Drag the four corner handles to reshape the shadow perspective.
//
// Keys:
//
//	m - toggle between Single Color and Color LUT method
//	+ - increase blur radius
//	- - decrease blur radius
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	blendcolor "github.com/MeKo-Christian/agg_go/internal/demo/blendcolor"
)

type demo struct {
	method   int
	radius   float64
	quad     [8]float64
	selected int
	dragDX   float64
	dragDY   float64
}

func newDemo() *demo {
	return &demo{
		method:   1, // Color LUT
		radius:   15.0,
		selected: -1,
	}
}

func (d *demo) Render(ctx *agg.Context) {
	result := blendcolor.Draw(ctx, &blendcolor.Config{
		Method: d.method,
		Radius: d.radius,
		Quad:   d.quad,
	})
	d.quad = result.Quad
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	d.selected = -1
	for i := 0; i < 4; i++ {
		dx := float64(x) - d.quad[i*2]
		dy := float64(y) - d.quad[i*2+1]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			d.selected = i
			d.dragDX = dx
			d.dragDY = dy
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	if d.selected < 0 {
		return false
	}
	d.quad[d.selected*2] = float64(x) - d.dragDX
	d.quad[d.selected*2+1] = float64(y) - d.dragDY
	return true
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	d.selected = -1
	return false
}

func (d *demo) OnKey(key rune) bool {
	switch key {
	case 'm', 'M':
		d.method = 1 - d.method
		return true
	case '+', '=':
		if d.radius < 40 {
			d.radius += 0.5
			return true
		}
	case '-', '_':
		if d.radius > 0 {
			d.radius -= 0.5
			return true
		}
	}
	return false
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "AGG Blend Color",
		Width:  440,
		Height: 330,
	}, newDemo())
}
