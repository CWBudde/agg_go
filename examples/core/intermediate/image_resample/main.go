// Package main ports AGG's image_resample.cpp demo (closest affine equivalent).
package main

import (
	"fmt"

	agg "agg_go"
)

func main() {
	const (
		w       = 600
		h       = 600
		mode    = 4
		blur    = 1.0
		outFile = "image_resample.png"
	)

	img := createSpheresImage(320, 320)
	quad := [4][2]float64{
		{140, 140},
		{460, 140},
		{460, 460},
		{140, 460},
	}
	if mode < 2 {
		quad[3][0] = quad[0][0] + (quad[2][0] - quad[1][0])
		quad[3][1] = quad[0][1] + (quad[2][1] - quad[1][1])
	}

	ctx := agg.NewContext(w, h)
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

	a.SetImageFilterRadius(agg.FilterBilinear, blur)
	if mode == 1 || mode == 4 || mode == 5 {
		a.ImageResample(agg.ResampleBilinear)
	} else {
		a.ImageResample(agg.ResampleNearest)
	}

	par := []float64{
		quad[0][0], quad[0][1],
		quad[1][0], quad[1][1],
		quad[2][0], quad[2][1],
	}
	_ = a.TransformImagePathParallelogramSimple(img, par)

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

	if err := ctx.GetImage().SaveToPNG(outFile); err != nil {
		fmt.Printf("error writing %s: %v\n", outFile, err)
		return
	}
	fmt.Printf("wrote %s (%dx%d)\n", outFile, w, h)
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
