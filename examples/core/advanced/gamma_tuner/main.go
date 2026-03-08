// Port of AGG C++ gamma_tuner.cpp – per-channel gamma tuning with patterns.
//
// Renders horizontal, vertical, and checkered test patterns with per-channel
// gamma correction applied. Default: R=1.0, G=1.0, B=1.0, Gamma=2.2,
// pattern=Checkered (index 2).
package main

import (
	"math"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
)

const (
	width  = 500
	height = 500

	defaultR       = 1.0
	defaultG       = 1.0
	defaultB       = 1.0
	defaultGamma   = 2.2
	defaultPattern = 2 // 0=Horizontal, 1=Vertical, 2=Checkered
	squareSize     = 400
)

func renderCheckered(img *agg.Image, rScale, gScale, bScale, gamma float64) {
	w, h := img.Width(), img.Height()
	data := img.Data
	stride := w * 4
	invG := 1.0 / gamma

	const strips = 5

	offsetY := (h - squareSize) / 2
	offsetX := (w - squareSize) / 2

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*stride + x*4

			// Background: white
			data[idx] = 255
			data[idx+1] = 255
			data[idx+2] = 255
			data[idx+3] = 255

			lx := x - offsetX
			ly := y - offsetY
			if lx < 0 || ly < 0 || lx >= squareSize || ly >= squareSize {
				continue
			}

			// Brightness gradient from top to bottom.
			k := float64(ly) / float64(squareSize-1)

			var r, g, b float64
			switch defaultPattern {
			case 0: // Horizontal stripes
				stripIdx := lx * strips / squareSize
				if stripIdx%2 == 0 {
					r = math.Pow(k, invG) * rScale
					g = math.Pow(k, invG) * gScale
					b = math.Pow(k, invG) * bScale
				} else {
					r, g, b = k, k, k
				}
			case 1: // Vertical stripes
				stripIdx := ly * strips / squareSize
				if stripIdx%2 == 0 {
					r = math.Pow(k, invG) * rScale
					g = math.Pow(k, invG) * gScale
					b = math.Pow(k, invG) * bScale
				} else {
					r, g, b = k, k, k
				}
			default: // Checkered
				cx := lx * strips / squareSize
				cy := ly * strips / squareSize
				if (cx+cy)%2 == 0 {
					r = math.Pow(k, invG) * rScale
					g = math.Pow(k, invG) * gScale
					b = math.Pow(k, invG) * bScale
				} else {
					r, g, b = k, k, k
				}
			}

			clamp := func(v float64) uint8 {
				if v <= 0 {
					return 0
				}
				if v >= 1 {
					return 255
				}
				return uint8(v * 255)
			}
			data[idx] = clamp(r)
			data[idx+1] = clamp(g)
			data[idx+2] = clamp(b)
			data[idx+3] = 255
		}
	}
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	renderCheckered(ctx.GetImage(), defaultR, defaultG, defaultB, defaultGamma)

	// Overlay some labels (borders).
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.NoFill()
	a.LineColor(agg.NewColor(100, 100, 100, 150))
	a.LineWidth(1.0)
	offsetX := float64(width-squareSize) / 2
	offsetY := float64(height-squareSize) / 2
	a.ResetPath()
	a.MoveTo(offsetX, offsetY)
	a.LineTo(offsetX+squareSize, offsetY)
	a.LineTo(offsetX+squareSize, offsetY+squareSize)
	a.LineTo(offsetX, offsetY+squareSize)
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Gamma Tuner",
		Width:  width,
		Height: height,
	}, &demo{})
}
