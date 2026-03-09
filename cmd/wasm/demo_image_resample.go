package main

import (
	agg "github.com/MeKo-Christian/agg_go"
)

// Port of AGG C++ image_resample.cpp.
//
// Note: current agg_go image transform path is affine (3-point parallelogram).
// Perspective mode entries are mapped to closest available affine+resample behavior.
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
	imageResampleImg *agg.Image
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

func setImageResampleQuad(
	x0, y0, x1, y1, x2, y2, x3, y3 float64,
) {
	imageResampleQuad[0][0], imageResampleQuad[0][1] = x0, y0
	imageResampleQuad[1][0], imageResampleQuad[1][1] = x1, y1
	imageResampleQuad[2][0], imageResampleQuad[2][1] = x2, y2
	imageResampleQuad[3][0], imageResampleQuad[3][1] = x3, y3
}

func imageResampleMode(t int) agg.ImageResample {
	switch t {
	case 1, 4, 5:
		return agg.ResampleBilinear
	default:
		return agg.ResampleNearest
	}
}

func drawImageResampleDemo() {
	if imageResampleImg == nil {
		imageResampleImg = createSpheresImage(320, 320)
	}
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	quad := imageResampleQuad
	if imageResampleType < 2 {
		// Match C++ affine modes: implicit 4th point of parallelogram.
		quad[3][0] = quad[0][0] + (quad[2][0] - quad[1][0])
		quad[3][1] = quad[0][1] + (quad[2][1] - quad[1][1])
	}

	// UI overlay quad fill.
	a.FillColor(agg.RGBA(0, 0.3, 0.5, 0.2))
	a.NoLine()
	a.ResetPath()
	a.MoveTo(quad[0][0], quad[0][1])
	a.LineTo(quad[1][0], quad[1][1])
	a.LineTo(quad[2][0], quad[2][1])
	a.LineTo(quad[3][0], quad[3][1])
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Build clipping path (destination polygon).
	a.ResetPath()
	a.MoveTo(quad[0][0], quad[0][1])
	a.LineTo(quad[1][0], quad[1][1])
	a.LineTo(quad[2][0], quad[2][1])
	a.LineTo(quad[3][0], quad[3][1])
	a.ClosePolygon()

	a.SetImageFilterRadius(agg.FilterBilinear, imageResampleBlur)
	a.ImageResample(imageResampleMode(imageResampleType))

	par := []float64{
		quad[0][0], quad[0][1],
		quad[1][0], quad[1][1],
		quad[2][0], quad[2][1],
	}
	_ = a.TransformImagePathParallelogramSimple(imageResampleImg, par)

	// Quad outline and handles.
	ctx.SetColor(agg.RGBA(0, 0.2, 0.3, 0.9))
	ctx.SetLineWidth(1.5)
	a.NoFill()
	a.ResetPath()
	a.MoveTo(quad[0][0], quad[0][1])
	a.LineTo(quad[1][0], quad[1][1])
	a.LineTo(quad[2][0], quad[2][1])
	a.LineTo(quad[3][0], quad[3][1])
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)

	ctx.SetColor(agg.RGBA(0.8, 0.1, 0.1, 0.75))
	for i := 0; i < 4; i++ {
		ctx.FillCircle(quad[i][0], quad[i][1], 4.0)
	}
}
