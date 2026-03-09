package main

import (
	"github.com/MeKo-Christian/agg_go/internal/demo/imageperspective"
)

var (
	imagePerspectiveType = 2
	imagePerspectiveNode = -1
	imagePerspectiveQuad = [4][2]float64{{100, 100}, {700, 100}, {700, 500}, {100, 500}}
)

func handleImagePerspectiveMouseDown(x, y float64) bool {
	return handleQuadMouseDown(x, y, &imagePerspectiveQuad, &imagePerspectiveNode)
}

func handleImagePerspectiveMouseMove(x, y float64) bool {
	return handleQuadMouseMove(x, y, &imagePerspectiveQuad, &imagePerspectiveNode)
}

func handleImagePerspectiveMouseUp() {
	handleQuadMouseUp(&imagePerspectiveNode)
}

func setImagePerspectiveType(v int) {
	if v < 0 {
		v = 0
	}
	if v > 2 {
		v = 2
	}
	imagePerspectiveType = v
}

func setImagePerspectiveQuad(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	imagePerspectiveQuad[0][0], imagePerspectiveQuad[0][1] = x0, y0
	imagePerspectiveQuad[1][0], imagePerspectiveQuad[1][1] = x1, y1
	imagePerspectiveQuad[2][0], imagePerspectiveQuad[2][1] = x2, y2
	imagePerspectiveQuad[3][0], imagePerspectiveQuad[3][1] = x3, y3
}

func drawImagePerspectiveDemo() {
	imageperspective.Draw(ctx, imageperspective.Config{
		Mode: imagePerspectiveType,
		Quad: imagePerspectiveQuad,
	})
}
