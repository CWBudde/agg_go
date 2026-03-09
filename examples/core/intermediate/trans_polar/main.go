// Port of AGG C++ trans_polar.cpp – polar coordinate transformation.
//
// Renders the lion vector art through a custom polar-coordinate transform,
// wrapping the lion into a ring/spiral shape on the canvas.
// Default: baseY=120, spiral=0 (pure ring, no spiral).
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

const (
	width  = 800
	height = 600

	polarBaseY  = 120.0
	polarSpiral = 0.0
)

// transPolar applies a polar warp: x-axis maps to angle, y-axis to radius.
type transPolar struct {
	baseAngle      float64
	baseScale      float64
	baseX, baseY   float64
	transX, transY float64
	spiral         float64
}

func (p *transPolar) Transform(x, y *float64) {
	x1 := (*x + p.baseX) * p.baseAngle
	y1 := (*y+p.baseY)*p.baseScale + (*x * p.spiral)
	*x = math.Cos(x1)*y1 + p.transX
	*y = math.Sin(x1)*y1 + p.transY
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	lionPaths := liondemo.Parse()

	// Find bounding box.
	lx1, ly1, lx2, ly2 := 1e9, 1e9, -1e9, -1e9
	for _, lp := range lionPaths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if x < lx1 {
				lx1 = x
			}
			if x > lx2 {
				lx2 = x
			}
			if y < ly1 {
				ly1 = y
			}
			if y > ly2 {
				ly2 = y
			}
		}
	}

	lionW := lx2 - lx1

	trans := &transPolar{
		baseAngle: 2.0 * math.Pi / lionW,
		baseScale: 1.0,
		baseX:     -lx1,
		baseY:     polarBaseY,
		transX:    float64(width) * 0.5,
		transY:    float64(height) * 0.5,
		spiral:    polarSpiral,
	}

	for _, lp := range lionPaths {
		a.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 200))
		a.NoLine()
		a.ResetPath()

		lp.Path.Rewind(0)
		first := true
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			tx, ty := x, y
			trans.Transform(&tx, &ty)
			if first || basics.IsMoveTo(basics.PathCommand(cmd)) {
				a.MoveTo(tx, ty)
				first = false
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				a.LineTo(tx, ty)
			}
		}
		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Trans Polar",
		Width:  width,
		Height: height,
	}, &demo{})
}
