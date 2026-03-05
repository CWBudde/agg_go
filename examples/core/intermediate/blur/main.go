// Based on the original AGG example: blur.cpp
// Renders an "a" glyph shape, draws a shadow, blurs it, then draws the shape on top.
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/color"
	"agg_go/internal/effects"
)

const (
	blurRadius = 15.0
	blurMethod = 0 // 0: Stack blur, 1: Recursive blur
)

func main() {
	const width, height = 800, 600
	ctx := agg.NewContext(width, height)
	ctx.Clear(agg.White)
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Define the "a" glyph path (from blur.cpp / WASM demo)
	agg2d.ResetPath()
	agg2d.MoveTo(28.47, 6.45)
	agg2d.QuadricCurveTo(21.58, 1.12, 19.82, 0.29)
	agg2d.QuadricCurveTo(17.19, -0.93, 14.21, -0.93)
	agg2d.QuadricCurveTo(9.57, -0.93, 6.57, 2.25)
	agg2d.QuadricCurveTo(3.56, 5.42, 3.56, 10.60)
	agg2d.QuadricCurveTo(3.56, 13.87, 5.03, 16.26)
	agg2d.QuadricCurveTo(7.03, 19.58, 11.99, 22.51)
	agg2d.QuadricCurveTo(16.94, 25.44, 28.47, 29.64)
	agg2d.LineTo(28.47, 31.40)
	agg2d.QuadricCurveTo(28.47, 38.09, 26.34, 40.58)
	agg2d.QuadricCurveTo(24.22, 43.07, 20.17, 43.07)
	agg2d.QuadricCurveTo(17.09, 43.07, 15.28, 41.41)
	agg2d.QuadricCurveTo(13.43, 39.75, 13.43, 37.60)
	agg2d.LineTo(13.53, 34.77)
	agg2d.QuadricCurveTo(13.53, 32.52, 12.38, 31.30)
	agg2d.QuadricCurveTo(11.23, 30.08, 9.38, 30.08)
	agg2d.QuadricCurveTo(7.57, 30.08, 6.42, 31.35)
	agg2d.QuadricCurveTo(5.27, 32.62, 5.27, 34.81)
	agg2d.QuadricCurveTo(5.27, 39.01, 9.57, 42.53)
	agg2d.QuadricCurveTo(13.87, 46.04, 21.63, 46.04)
	agg2d.QuadricCurveTo(27.59, 46.04, 31.40, 44.04)
	agg2d.QuadricCurveTo(34.28, 42.53, 35.64, 39.31)
	agg2d.QuadricCurveTo(36.52, 37.21, 36.52, 30.71)
	agg2d.LineTo(36.52, 15.53)
	agg2d.QuadricCurveTo(36.52, 9.13, 36.77, 7.69)
	agg2d.QuadricCurveTo(37.01, 6.25, 37.57, 5.76)
	agg2d.QuadricCurveTo(38.13, 5.27, 38.87, 5.27)
	agg2d.QuadricCurveTo(39.65, 5.27, 40.23, 5.62)
	agg2d.QuadricCurveTo(41.26, 6.25, 44.19, 9.18)
	agg2d.LineTo(44.19, 6.45)
	agg2d.QuadricCurveTo(38.72, -0.88, 33.74, -0.88)
	agg2d.QuadricCurveTo(31.35, -0.88, 29.93, 0.78)
	agg2d.QuadricCurveTo(28.52, 2.44, 28.47, 6.45)
	agg2d.ClosePolygon()

	agg2d.MoveTo(28.47, 9.62)
	agg2d.LineTo(28.47, 26.66)
	agg2d.QuadricCurveTo(21.09, 23.73, 18.95, 22.51)
	agg2d.QuadricCurveTo(15.09, 20.36, 13.43, 18.02)
	agg2d.QuadricCurveTo(11.77, 15.67, 11.77, 12.89)
	agg2d.QuadricCurveTo(11.77, 9.38, 13.87, 7.06)
	agg2d.QuadricCurveTo(15.97, 4.74, 18.70, 4.74)
	agg2d.QuadricCurveTo(22.41, 4.74, 28.47, 9.62)
	agg2d.ClosePolygon()

	agg2d.Scale(4.0, -4.0)
	agg2d.Translate(150, 400)

	// Draw shadow (dark fill)
	agg2d.FillColor(agg.NewColor(25, 25, 25, 255))
	agg2d.DrawPath(agg.FillOnly)

	// Apply blur to the current buffer
	blurImage(ctx.GetImage(), blurRadius, blurMethod)

	// Draw the shape itself on top
	agg2d.FillColor(agg.NewColor(153, 230, 179, 204)) // rgba(0.6, 0.9, 0.7, 0.8)
	agg2d.DrawPath(agg.FillOnly)

	const filename = "blur.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}

func blurImage(img *agg.Image, radius float64, method int) {
	if radius <= 0 {
		return
	}

	w, h := img.Width(), img.Height()
	data := img.Data
	stride := w * 4

	// Convert flat data to [][]color.RGBA8[color.Linear]
	pixels := make([][]color.RGBA8[color.Linear], h)
	for y := 0; y < h; y++ {
		pixels[y] = make([]color.RGBA8[color.Linear], w)
		for x := 0; x < w; x++ {
			idx := y*stride + x*4
			pixels[y][x] = color.RGBA8[color.Linear]{
				R: data[idx],
				G: data[idx+1],
				B: data[idx+2],
				A: data[idx+3],
			}
		}
	}

	if method == 0 {
		sb := effects.NewSimpleStackBlur()
		sb.Blur(pixels, int(radius))
	} else {
		rb := effects.NewSimpleRecursiveBlur()
		rb.BlurHorizontal(pixels, radius)
		pixels = transposePixels(pixels)
		rb.BlurHorizontal(pixels, radius)
		pixels = transposePixels(pixels)
	}

	// Copy back to flat data
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*stride + x*4
			pix := pixels[y][x]
			data[idx] = uint8(pix.R)
			data[idx+1] = uint8(pix.G)
			data[idx+2] = uint8(pix.B)
			data[idx+3] = uint8(pix.A)
		}
	}
}

func transposePixels(pixels [][]color.RGBA8[color.Linear]) [][]color.RGBA8[color.Linear] {
	if len(pixels) == 0 {
		return pixels
	}
	h := len(pixels)
	w := len(pixels[0])
	newPixels := make([][]color.RGBA8[color.Linear], w)
	for x := 0; x < w; x++ {
		newPixels[x] = make([]color.RGBA8[color.Linear], h)
		for y := 0; y < h; y++ {
			newPixels[x][y] = pixels[y][x]
		}
	}
	return newPixels
}
