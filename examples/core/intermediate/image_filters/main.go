package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
)

func main() {
	ctx := agg.NewContext(1100, 820)
	ctx.Clear(agg.White)
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	testImage := createTestImage(180, 180)
	filters := []struct {
		name   string
		filter agg.ImageFilter
		radius float64
		angle  float64
		scale  float64
		x      float64
		y      float64
	}{
		{"Bilinear", agg.FilterBilinear, 4, 12, 0.8, 210, 160},
		{"Bicubic", agg.FilterBicubic, 4, 18, 1.0, 530, 160},
		{"Lanczos", agg.FilterLanczos, 4, 25, 1.2, 850, 160},
		{"Blackman", agg.FilterBlackman, 4, -18, 1.4, 370, 500},
		{"Spline36", agg.FilterSpline36, 4, -10, 0.9, 690, 500},
	}

	agg2d.TransformImageSimple(testImage, 40, 40, 180, 180)
	for _, item := range filters {
		agg2d.ResetTransformations()
		agg2d.SetImageFilterRadius(item.filter, item.radius)
		para := buildParallelogram(float64(testImage.Width()), float64(testImage.Height()), item.x, item.y, item.scale, item.angle)
		if err := agg2d.TransformImageParallelogram(testImage, 0, 0, testImage.Width(), testImage.Height(), para); err != nil {
			panic(err)
		}
	}

	const filename = "image_filters.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}

func buildParallelogram(imgW, imgH, x, y, scale, angleDeg float64) []float64 {
	halfW := (imgW * scale) / 2.0
	halfH := (imgH * scale) / 2.0
	angleRad := angleDeg * math.Pi / 180.0
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	return []float64{
		x - halfW*cosA + halfH*sinA,
		y - halfW*sinA - halfH*cosA,
		x + halfW*cosA + halfH*sinA,
		y + halfW*sinA - halfH*cosA,
		x + halfW*cosA - halfH*sinA,
		y + halfW*sinA + halfH*cosA,
	}
}

func createTestImage(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)
	imgCtx.Clear(agg.White)

	imgCtx.SetColor(agg.RGBA(0.8, 0.8, 0.8, 1.0))
	for i := 0; i < w; i += 20 {
		imgCtx.DrawLine(float64(i), 0, float64(i), float64(h))
	}
	for i := 0; i < h; i += 20 {
		imgCtx.DrawLine(0, float64(i), float64(w), float64(i))
	}

	imgCtx.SetColor(agg.Red)
	imgCtx.FillCircle(float64(w)/2, float64(h)/2, float64(w)/4)
	imgCtx.SetColor(agg.Blue)
	imgCtx.SetStrokeWidth(5.0)
	imgCtx.DrawRectangle(10, 10, float64(w-20), float64(h-20))

	imgCtx.SetColor(agg.Black)
	imgCtx.SetStrokeWidth(1.0)
	for i := -w; i < w; i += 4 {
		imgCtx.DrawLine(float64(i), 0, float64(i+w), float64(h))
	}

	return img
}
