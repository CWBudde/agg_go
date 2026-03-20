// Package main ports AGG's image_resample.cpp demo (closest affine equivalent).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
)

const (
	resampleMode = 4
	resampleBlur = 1.0
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	srcImg := createSpheresImage(320, 320)
	quad := [4][2]float64{
		{140, 140},
		{460, 140},
		{460, 460},
		{140, 460},
	}
	if resampleMode < 2 {
		quad[3][0] = quad[0][0] + (quad[2][0] - quad[1][0])
		quad[3][1] = quad[0][1] + (quad[2][1] - quad[1][1])
	}

	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	a.FillColor(agg.RGBA(0, 0.3, 0.5, 0.2))
	a.NoLine()
	a.ResetPath()
	a.MoveTo(quad[0][0], quad[0][1])
	a.LineTo(quad[1][0], quad[1][1])
	a.LineTo(quad[2][0], quad[2][1])
	a.LineTo(quad[3][0], quad[3][1])
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	a.ResetPath()
	a.MoveTo(quad[0][0], quad[0][1])
	a.LineTo(quad[1][0], quad[1][1])
	a.LineTo(quad[2][0], quad[2][1])
	a.LineTo(quad[3][0], quad[3][1])
	a.ClosePolygon()

	a.SetImageFilterRadius(agg.FilterBilinear, resampleBlur)
	if resampleMode == 1 || resampleMode == 4 || resampleMode == 5 {
		a.ImageResample(agg.ResampleBilinear)
	} else {
		a.ImageResample(agg.ResampleNearest)
	}

	par := []float64{
		quad[0][0], quad[0][1],
		quad[1][0], quad[1][1],
		quad[2][0], quad[2][1],
	}
	_ = a.TransformImagePathParallelogramSimple(srcImg, par)

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

func createSpheresImage(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)

	imgCtx.SetColor(agg.RGBA(0.05, 0.05, 0.12, 1.0))
	imgCtx.FillRectangle(0, 0, float64(w), float64(h))

	type sphere struct {
		x, y, r    float64
		r0, g0, b0 float64
	}
	spheres := []sphere{
		{float64(w) * 0.22, float64(h) * 0.30, float64(w) * 0.18, 0.9, 0.2, 0.1},
		{float64(w) * 0.65, float64(h) * 0.28, float64(w) * 0.15, 0.1, 0.4, 0.9},
		{float64(w) * 0.45, float64(h) * 0.68, float64(w) * 0.20, 0.1, 0.8, 0.3},
		{float64(w) * 0.78, float64(h) * 0.65, float64(w) * 0.12, 0.9, 0.7, 0.1},
		{float64(w) * 0.15, float64(h) * 0.72, float64(w) * 0.10, 0.7, 0.1, 0.8},
	}

	for _, sp := range spheres {
		imgCtx.SetColor(agg.RGBA(0, 0, 0, 0.35))
		imgCtx.FillCircle(sp.x+sp.r*0.15, sp.y+sp.r*0.15, sp.r)
		imgCtx.SetColor(agg.RGBA(sp.r0, sp.g0, sp.b0, 0.85))
		imgCtx.FillCircle(sp.x, sp.y, sp.r)
		imgCtx.SetColor(agg.RGBA(1.0, 1.0, 1.0, 0.6))
		imgCtx.FillCircle(sp.x-sp.r*0.30, sp.y-sp.r*0.30, sp.r*0.30)
	}
	return img
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Resample",
		Width:  600,
		Height: 600,
	}, &demo{})
}
