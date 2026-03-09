// Based on the original AGG examples: trans_polar.cpp.
package main

import (
	"math"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

type transPolar struct {
	baseAngle      float64
	baseScale      float64
	baseX, baseY   float64
	transX, transY float64
	spiral         float64
}

func (p *transPolar) transform(x, y *float64) {
	x1 := (*x + p.baseX) * p.baseAngle
	y1 := (*y+p.baseY)*p.baseScale + (*x * p.spiral)
	*x = math.Cos(x1)*y1 + p.transX
	*y = math.Sin(x1)*y1 + p.transY
}

func (p *transPolar) Transform(x, y *float64) {
	p.transform(x, y)
}

var (
	polarBaseY  = 120.0
	polarSpiral = 0.0
)

func drawTransPolarDemo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Setup polar transformer
	trans := &transPolar{
		baseAngle: 2.0 * math.Pi / 600.0, // spread 600 units over 2PI
		baseScale: 1.0,
		baseX:     0.0,
		baseY:     polarBaseY,
		transX:    float64(width) * 0.5,
		transY:    float64(height) * 0.5,
		spiral:    polarSpiral,
	}

	// We'll transform the lion
	// Find bounding box of the lion
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
	// Scale lion to fit the "circle"
	// We want it to be roughly 600 units wide in logical space
	scaleX := 600.0 / lionW

	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		agg2d.NoLine()

		agg2d.ResetPath()
		lp.Path.Rewind(0)

		// Use a segmentator to ensure the lion curves nicely
		// Actually for a simple demo we can just transform vertices,
		// but segmentator would be better if we had long straight lines.
		// Lion has many small segments already.

		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			// Normalize and scale lion
			tx := (x-lx1)*scaleX - 300.0 // Center it horizontally
			ty := (y - (ly1+ly2)*0.5)

			// Transform to polar
			trans.Transform(&tx, &ty)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(tx, ty)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}
}

func handleTransPolarMouseDown(x, y float64) bool {
	polarBaseY = y - float64(height)*0.5 + 120.0
	return true
}

func handleTransPolarMouseMove(x, y float64) bool {
	polarBaseY = y - float64(height)*0.5 + 120.0
	polarSpiral = (x - float64(width)*0.5) / 1000.0
	return true
}

func handleTransPolarMouseUp() {}
