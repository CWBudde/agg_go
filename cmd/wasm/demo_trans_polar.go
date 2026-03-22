// Based on the original AGG examples: trans_polar.cpp.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
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
	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
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

	// Find bounding box of the lion
	lx1, ly1, lx2, ly2 := 1e9, 1e9, -1e9, -1e9
	for idx := uint(0); idx < lionData.Path.TotalVertices(); idx++ {
		x, y, cmd := lionData.Path.Vertex(idx)
		if !basics.IsVertex(basics.PathCommand(cmd)) {
			continue
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

	lionW := lx2 - lx1
	// Scale lion to fit the "circle"
	scaleX := 600.0 / lionW

	for i := 0; i < lionData.NPaths; i++ {
		agg2d.FillColor(agg.NewColor(lionData.Colors[i].R, lionData.Colors[i].G, lionData.Colors[i].B, 255))
		agg2d.NoLine()

		agg2d.ResetPath()
		lionData.Path.Rewind(lionData.PathIdx[i])

		for {
			x, y, cmd := lionData.Path.NextVertex()
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
