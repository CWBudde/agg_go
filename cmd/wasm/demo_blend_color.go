package main

import (
	"math"

	blendcolor "github.com/MeKo-Christian/agg_go/internal/demo/blendcolor"
)

var (
	blendColorMethod       = 1    // 0: Single Color, 1: Color LUT
	blendColorRadius       = 15.0 // Blur radius
	blendColorQuad         [8]float64
	blendColorShapeBounds  [4]float64
	blendColorSelectedNode = -1
	blendColorDragDX       = 0.0
	blendColorDragDY       = 0.0
)

func drawBlendColorDemo() {
	result := blendcolor.Draw(ctx, blendcolor.Config{
		Method: blendColorMethod,
		Radius: blendColorRadius,
		Quad:   blendColorQuad,
	})
	blendColorQuad = result.Quad
	blendColorShapeBounds = result.ShapeBounds
}

// Mouse handlers — dragging 4 quad corners (8 values: x0,y0, x1,y1, x2,y2, x3,y3)
func handleBlendColorMouseDown(x, y float64) bool {
	blendColorSelectedNode = -1
	for i := 0; i < 4; i++ {
		dx := x - blendColorQuad[i*2]
		dy := y - blendColorQuad[i*2+1]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			blendColorSelectedNode = i
			blendColorDragDX = dx
			blendColorDragDY = dy
			return true
		}
	}
	return false
}

func handleBlendColorMouseMove(x, y float64) bool {
	if blendColorSelectedNode < 0 {
		return false
	}
	blendColorQuad[blendColorSelectedNode*2] = x - blendColorDragDX
	blendColorQuad[blendColorSelectedNode*2+1] = y - blendColorDragDY
	return true
}

func handleBlendColorMouseUp() {
	blendColorSelectedNode = -1
}
