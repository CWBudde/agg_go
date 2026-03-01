// Based on the original AGG examples: image_filters.cpp.
//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"math"

	agg "agg_go"
)

var (
	imgFilterAngle  = 0.0
	imgFilterType   = agg.FilterBilinear
	imgFilterRadius = 4.0
	testImage       *agg.Image
)

func drawImageFiltersDemo() {
	if testImage == nil {
		testImage = createTestImage(200, 200)
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Update status with current settings
	logStatus(fmt.Sprintf("Filter: %d, Radius: %.2f, Angle: %.1f", imgFilterType, imgFilterRadius, imgFilterAngle))

	// Set the filter
	agg2d.SetImageFilterRadius(imgFilterType, imgFilterRadius)

	// Draw the image multiple times with rotation

	// Draw original image for reference
	agg2d.TransformImageSimple(testImage, 50, 50, 250, 250)

	// Draw rotated and scaled versions
	agg2d.ResetTransformations()

	// Center rotation
	imgW, imgH := float64(testImage.Width()), float64(testImage.Height())

	// We'll draw 3 versions at different scales
	scales := []float64{0.5, 1.0, 2.0}
	for i, s := range scales {
		x := 350.0 + float64(i)*150.0
		y := 150.0

		// Map source image to destination parallelogram
		// This is what original AGG demo effectively does via its transform_image

		// Simple rotation + scale around image center
		angleRad := imgFilterAngle * math.Pi / 180.0

		// Destination points
		// We'll use TransformImageParallelogram for maximum flexibility
		// although Agg2D has simpler methods, this showcases AGG's power.

		halfW := (imgW * s) / 2.0
		halfH := (imgH * s) / 2.0

		cosA := math.Cos(angleRad)
		sinA := math.Sin(angleRad)

		// Parallelogram: x1,y1, x2,y2, x3,y3 (3 corners)
		para := make([]float64, 6)

		// Corner 1: top-left
		para[0] = x - halfW*cosA + halfH*sinA
		para[1] = y - halfW*sinA - halfH*cosA

		// Corner 2: top-right
		para[2] = x + halfW*cosA + halfH*sinA
		para[3] = y + halfW*sinA - halfH*cosA

		// Corner 3: bottom-right
		para[4] = x + halfW*cosA - halfH*sinA
		para[5] = y + halfW*sinA + halfH*cosA

		agg2d.TransformImageParallelogram(testImage, 0, 0, int(imgW), int(imgH), para)
	}
}

func createTestImage(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)

	imgCtx.Clear(agg.White)

	// Draw a grid
	imgCtx.SetColor(agg.RGBA(0.8, 0.8, 0.8, 1.0))
	for i := 0; i < w; i += 20 {
		imgCtx.DrawLine(float64(i), 0, float64(i), float64(h))
	}
	for i := 0; i < h; i += 20 {
		imgCtx.DrawLine(0, float64(i), float64(w), float64(i))
	}

	// Draw some shapes
	imgCtx.SetColor(agg.Red)
	imgCtx.FillCircle(float64(w)/2, float64(h)/2, float64(w)/4)

	imgCtx.SetColor(agg.Blue)
	imgCtx.SetStrokeWidth(5.0)
	imgCtx.DrawRectangle(10, 10, float64(w-20), float64(h-20))

	// Draw high-frequency pattern (diagonal lines)
	imgCtx.SetColor(agg.Black)
	imgCtx.SetStrokeWidth(1.0)
	for i := -w; i < w; i += 4 {
		imgCtx.DrawLine(float64(i), 0, float64(i+w), float64(h))
	}

	return img
}
