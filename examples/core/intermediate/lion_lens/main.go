// Port of AGG C++ lion_lens.cpp – lion with warp-magnifier lens effect.
//
// Renders the lion vector art with a warp-magnifier lens applied at the
// canvas centre (default position). In an interactive version the lens
// position would follow the mouse.
package main

import (
	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	llWidth  = 800
	llHeight = 600

	// Default lens parameters.
	lensScale  = 5.0
	lensRadius = 70.0
	lensX      = float64(llWidth) * 0.5
	lensY      = float64(llHeight) * 0.5
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	lionPaths := liondemo.Parse()

	// Compute lion bounding box.
	bx1, by1, bx2, by2 := 1e9, 1e9, -1e9, -1e9
	for _, lp := range lionPaths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsVertex(basics.PathCommand(cmd)) {
				if x < bx1 {
					bx1 = x
				}
				if y < by1 {
					by1 = y
				}
				if x > bx2 {
					bx2 = x
				}
				if y > by2 {
					by2 = y
				}
			}
		}
	}

	baseDX := (bx2 - bx1) * 0.5
	baseDY := (by2 - by1) * 0.5

	// Base affine: centre lion on canvas.
	mtx := transform.NewTransAffine()
	mtx.Translate(-baseDX, -baseDY)
	mtx.ScaleXY(-1, 1) // mirror X (matches flip_y + rotate(Pi) in the original)
	mtx.Translate(float64(llWidth)*0.5, float64(llHeight)*0.5)

	// Warp magnifier lens.
	lens := transform.NewTransWarpMagnifier()
	lens.SetCenter(lensX, lensY)
	lens.SetMagnification(lensScale)
	lens.SetRadius(lensRadius / lensScale)

	for _, lp := range lionPaths {
		a.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		a.NoLine()
		a.ResetPath()

		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			// Apply base transform then lens.
			mtx.Transform(&x, &y)
			lens.Transform(&x, &y)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				a.MoveTo(x, y)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				a.LineTo(x, y)
			}
		}
		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)
	}

	// Draw lens circle outline.
	a.NoFill()
	a.LineColor(agg.NewColor(80, 80, 80, 180))
	a.LineWidth(1.0)
	a.DrawCircle(lensX, lensY, lensRadius)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Lion Lens",
		Width:  llWidth,
		Height: llHeight,
	}, &demo{})
}
