package main

import (
	"github.com/cwbudde/agg_go/internal/demo/imageresample"
)

var (
	imageResampleType = 4   // C++ default: Perspective Resample LERP
	imageResampleBlur = 1.0 // C++ slider default
	imageResampleNode = -1
	imageResampleQuad = [4][2]float64{
		{140, 140},
		{460, 140},
		{460, 460},
		{140, 460},
	}
)

func handleImageResampleMouseDown(x, y float64) bool {
	return handleQuadMouseDown(x, y, &imageResampleQuad, &imageResampleNode)
}

func handleImageResampleMouseMove(x, y float64) bool {
	return handleQuadMouseMove(x, y, &imageResampleQuad, &imageResampleNode)
}

func handleImageResampleMouseUp() {
	handleQuadMouseUp(&imageResampleNode)
}

func setImageResampleType(v int) {
	if v < 0 {
		v = 0
	}
	if v > 5 {
		v = 5
	}
	imageResampleType = v
}

func setImageResampleBlur(v float64) {
	if v < 0.5 {
		v = 0.5
	}
	if v > 5.0 {
		v = 5.0
	}
	imageResampleBlur = v
}

func setImageResampleQuad(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	imageResampleQuad[0][0], imageResampleQuad[0][1] = x0, y0
	imageResampleQuad[1][0], imageResampleQuad[1][1] = x1, y1
	imageResampleQuad[2][0], imageResampleQuad[2][1] = x2, y2
	imageResampleQuad[3][0], imageResampleQuad[3][1] = x3, y3
}

func drawImageResampleDemo() {
	imageresample.Draw(ctx, imageresample.Config{
		Mode: imageResampleType,
		Blur: imageResampleBlur,
		Quad: imageResampleQuad,
	})
}
