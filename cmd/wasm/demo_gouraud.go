// Based on the original AGG examples: gouraud.cpp.
package main

import (
	"fmt"
	"math"

	agg "agg_go"
)

var (
	gouraudX        = [3]float64{100, 500, 300}
	gouraudY        = [3]float64{100, 150, 500}
	gouraudDilation = 0.5
	gouraudSelected = -1
	gouraudDragDX   = 0.0
	gouraudDragDY   = 0.0
)

func drawGouraudDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	logStatus(fmt.Sprintf("Gouraud Dilation: %.2f", gouraudDilation))

	// Subdivision into 6 triangles as in original gouraud.cpp
	xc := (gouraudX[0] + gouraudX[1] + gouraudX[2]) / 3.0
	yc := (gouraudY[0] + gouraudY[1] + gouraudY[2]) / 3.0

	x1 := (gouraudX[1]+gouraudX[0])*0.5 - (xc - (gouraudX[1]+gouraudX[0])*0.5)
	y1 := (gouraudY[1]+gouraudY[0])*0.5 - (yc - (gouraudY[1]+gouraudY[0])*0.5)

	x2 := (gouraudX[2]+gouraudX[1])*0.5 - (xc - (gouraudX[2]+gouraudX[1])*0.5)
	y2 := (gouraudY[2]+gouraudY[1])*0.5 - (yc - (gouraudY[2]+gouraudY[1])*0.5)

	x3 := (gouraudX[0]+gouraudX[2])*0.5 - (xc - (gouraudX[0]+gouraudX[2])*0.5)
	y3 := (gouraudY[0]+gouraudY[2])*0.5 - (yc - (gouraudY[0]+gouraudY[2])*0.5)

	cRed := agg.NewColor(255, 0, 0, 255)
	cGreen := agg.NewColor(0, 255, 0, 255)
	cBlue := agg.NewColor(0, 0, 255, 255)
	cWhite := agg.White
	cBlack := agg.Black

	// First three triangles (center-based)
	agg2d.GouraudTriangle(gouraudX[0], gouraudY[0], gouraudX[1], gouraudY[1], xc, yc, cRed, cGreen, cWhite, gouraudDilation)
	agg2d.GouraudTriangle(gouraudX[1], gouraudY[1], gouraudX[2], gouraudY[2], xc, yc, cGreen, cBlue, cWhite, gouraudDilation)
	agg2d.GouraudTriangle(gouraudX[2], gouraudY[2], gouraudX[0], gouraudY[0], xc, yc, cBlue, cRed, cWhite, gouraudDilation)

	// Next three triangles (edge-based)
	agg2d.GouraudTriangle(gouraudX[0], gouraudY[0], gouraudX[1], gouraudY[1], x1, y1, cRed, cGreen, cBlack, gouraudDilation)
	agg2d.GouraudTriangle(gouraudX[1], gouraudY[1], gouraudX[2], gouraudY[2], x2, y2, cGreen, cBlue, cBlack, gouraudDilation)
	agg2d.GouraudTriangle(gouraudX[2], gouraudY[2], gouraudX[0], gouraudY[0], x3, y3, cBlue, cRed, cBlack, gouraudDilation)

	// Draw interactive handles
	for i := 0; i < 3; i++ {
		agg2d.FillColor(agg.NewColor(200, 50, 20, 150))
		agg2d.NoLine()
		agg2d.FillCircle(gouraudX[i], gouraudY[i], 8)
		agg2d.LineColor(agg.Black)
		agg2d.LineWidth(1.0)
		agg2d.DrawCircle(gouraudX[i], gouraudY[i], 8)
	}
}

func handleGouraudMouseDown(x, y float64) bool {
	gouraudSelected = -1
	for i := 0; i < 3; i++ {
		dist := math.Sqrt(math.Pow(x-gouraudX[i], 2) + math.Pow(y-gouraudY[i], 2))
		if dist < 20 {
			gouraudSelected = i
			gouraudDragDX = x - gouraudX[i]
			gouraudDragDY = y - gouraudY[i]
			return true
		}
	}
	return false
}

func handleGouraudMouseMove(x, y float64) bool {
	if gouraudSelected != -1 {
		gouraudX[gouraudSelected] = x - gouraudDragDX
		gouraudY[gouraudSelected] = y - gouraudDragDY
		return true
	}
	return false
}

func handleGouraudMouseUp() {
	gouraudSelected = -1
}
