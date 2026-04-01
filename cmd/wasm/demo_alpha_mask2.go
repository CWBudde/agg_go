package main

import (
	"fmt"
	"math"

	alphamask2demo "github.com/cwbudde/agg_go/internal/demo/alphamask2"
)

const (
	am2RefWidth  = 512
	am2RefHeight = 400
)

var (
	am2NumEllipses = 10
	am2LionAngle   = 0.0
	am2LionScale   = 1.0
	am2LionSkewX   = 0.0
	am2LionSkewY   = 0.0
	am2SliderValue = 10.0
)

func drawAlphaMask2Demo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	if float64(am2NumEllipses) != am2SliderValue {
		am2NumEllipses = int(am2SliderValue)
	}

	// Scale the lion proportionally so it fills the canvas the same way
	// the C++ original fills its 512×400 window.
	canvasScale := math.Min(float64(w)/am2RefWidth, float64(h)/am2RefHeight)
	scale := am2LionScale * canvasScale

	// Render into a BGR24 work buffer like the original AGG example.
	workBuf := make([]uint8, w*h*3)
	alphamask2demo.RenderToBGR24(workBuf, w, h, alphamask2demo.Config{
		NumEllipses: am2NumEllipses,
		Angle:       am2LionAngle,
		Scale:       scale,
		SkewX:       am2LionSkewX,
		SkewY:       am2LionSkewY,
	})

	// Copy BGR24 work buffer to RGBA canvas with Y-flip. RenderToBGR24 renders
	// in AGG's Y-up coordinate system (row 0 = bottom of scene), so we must
	// flip vertically when copying to the top-down WASM/PNG canvas.
	srcStride := w * 3
	dstStride := img.Stride()
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * srcStride
		dstOff := y * dstStride
		for x := 0; x < w; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			img.Data[d+0] = workBuf[s+2] // R ← B in BGR
			img.Data[d+1] = workBuf[s+1] // G
			img.Data[d+2] = workBuf[s+0] // B ← R in BGR
			img.Data[d+3] = 255
		}
	}

	applyLinearToSRGB(img)

	logStatus(fmt.Sprintf("Alpha Mask 2 Demo: Ellipses=%d", am2NumEllipses))
}

func handleAlphaMask2MouseDown(x, y float64, flags int) bool {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	dx := x - float64(w)/2
	// Canvas Y is top-down; AGG rendering uses Y-up, so flip Y relative to centre
	// to match C++ platform_support flip_y=true behaviour.
	dy := float64(h)/2 - y
	am2LionAngle = math.Atan2(dy, dx)
	am2LionScale = math.Sqrt(dy*dy+dx*dx) / 100.0
	return true
}

func handleAlphaMask2RightMouseDown(x, y float64) bool {
	am2LionSkewX = x
	// Flip Y for the same reason as in handleAlphaMask2MouseDown.
	am2LionSkewY = float64(ctx.GetImage().Height()) - y
	return true
}

func setAlphaMask2NumEllipses(n float64) {
	am2SliderValue = n
}
