// Port of AGG C++ lion_outline.cpp example.
//
// Renders the lion vector art as stroked outlines rather than filled
// polygons, demonstrating the outline/stroke rendering mode.
// The static output uses the default transform: no rotation, scale=1,
// no shear, centred on the canvas.
package main

import (
	"math"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

const (
	loWidth  = 800
	loHeight = 600

	// Default values matching the WASM demo initial state.
	outlineWidth = 1.0
	lionAngle    = 0.0
	lionScale    = 1.0
	lionSkewX    = 0.0
	lionSkewY    = 0.0

	// Lion bounding box centre, matching the original parse_lion data.
	// Computed from bbox 7..557 x 8..520.
	lionBaseDX = (557.0 - 7.0) * 0.5 // ≈ 275
	lionBaseDY = (520.0 - 8.0) * 0.5 // ≈ 256
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()

	// Set up the affine transform matching the C++ matrix composition:
	//   translate(-baseDX, -baseDY)  → centre lion on origin
	//   scale(scale)
	//   rotate(angle + π)            → +π corrects y-down orientation
	//   skew(skewX/1000, skewY/1000)
	//   translate(w/2, h/2)          → centre on canvas
	a.ResetTransformations()
	a.Translate(-lionBaseDX, -lionBaseDY)
	a.Scale(lionScale, lionScale)
	a.Rotate(lionAngle + math.Pi)
	a.Skew(lionSkewX/1000.0, lionSkewY/1000.0)
	a.Translate(float64(loWidth)/2, float64(loHeight)/2)

	a.LineWidth(outlineWidth)
	a.NoFill()

	for _, lp := range liondemo.Parse() {
		a.LineColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		a.ResetPath()

		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				a.MoveTo(x, y)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				a.LineTo(x, y)
			}
		}
		a.ClosePolygon()
		a.DrawPath(agg.StrokeOnly)
	}

	a.ResetTransformations()
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Lion Outline",
		Width:  loWidth,
		Height: loHeight,
	}, &demo{})
}
