// Based on the original AGG examples: perspective.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/transform"
)

var (
	perspectiveQuad                                                            = [8]float64{100, 100, 500, 100, 500, 500, 100, 500}
	perspectiveSelectedNode                                                    = -1
	perspectiveType                                                            = 0 // 0: Bilinear, 1: Perspective
	perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2 float64
	perspectiveInitialized                                                     = false
)

func initPerspectiveDemo() {
	if lionPaths == nil {
		lionPaths = liondemo.Parse()
	}

	if perspectiveInitialized {
		return
	}

	// Find bounding box of the lion
	x1, y1, x2, y2 := 1e9, 1e9, -1e9, -1e9
	for _, lp := range lionPaths {
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if x < x1 {
				x1 = x
			}
			if x > x2 {
				x2 = x
			}
			if y < y1 {
				y1 = y
			}
			if y > y2 {
				y2 = y
			}
		}
	}
	perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2 = x1, y1, x2, y2

	// Initialize quad to center the lion
	cx, cy := float64(width)/2, float64(height)/2
	w, h := (x2 - x1), (y2 - y1)
	perspectiveQuad[0], perspectiveQuad[1] = cx-w/2, cy-h/2
	perspectiveQuad[2], perspectiveQuad[3] = cx+w/2, cy-h/2
	perspectiveQuad[4], perspectiveQuad[5] = cx+w/2, cy+h/2
	perspectiveQuad[6], perspectiveQuad[7] = cx-w/2, cy+h/2

	perspectiveInitialized = true
}

func drawPerspectiveDemo() {
	initPerspectiveDemo()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	var tr transform.Transformer
	if perspectiveType == 0 {
		tr = transform.NewTransBilinearRectToQuad(perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2, perspectiveQuad)
	} else {
		tr = transform.NewTransPerspectiveRectToQuad(perspectiveLionX1, perspectiveLionY1, perspectiveLionX2, perspectiveLionY2, perspectiveQuad)
	}

	// Render transformed lion
	for _, lp := range lionPaths {
		agg2d.FillColor(agg.NewColor(lp.Color[0], lp.Color[1], lp.Color[2], 255))
		agg2d.NoLine()

		agg2d.ResetPath()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			tx, ty := x, y
			tr.Transform(&tx, &ty)

			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(tx, ty)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}

	// Draw the quad tool
	ctx.SetColor(agg.RGBA(0, 0.3, 0.5, 0.6))
	ctx.SetLineWidth(2.0)
	ctx.BeginPath()
	ctx.MoveTo(perspectiveQuad[0], perspectiveQuad[1])
	ctx.LineTo(perspectiveQuad[2], perspectiveQuad[3])
	ctx.LineTo(perspectiveQuad[4], perspectiveQuad[5])
	ctx.LineTo(perspectiveQuad[6], perspectiveQuad[7])
	ctx.ClosePath()
	ctx.Stroke()

	// Draw handles
	for i := 0; i < 4; i++ {
		drawHandle(perspectiveQuad[i*2], perspectiveQuad[i*2+1])
	}
}

func handlePerspectiveMouseDown(x, y float64) bool {
	perspectiveSelectedNode = -1
	for i := 0; i < 4; i++ {
		dist := math.Sqrt(math.Pow(x-perspectiveQuad[i*2], 2) + math.Pow(y-perspectiveQuad[i*2+1], 2))
		if dist < 10 {
			perspectiveSelectedNode = i
			return true
		}
	}
	return false
}

func handlePerspectiveMouseMove(x, y float64) bool {
	if perspectiveSelectedNode != -1 {
		perspectiveQuad[perspectiveSelectedNode*2] = x
		perspectiveQuad[perspectiveSelectedNode*2+1] = y
		return true
	}
	return false
}

func handlePerspectiveMouseUp() {
	perspectiveSelectedNode = -1
}

func setPerspectiveType(t int) {
	perspectiveType = t
}
