package main

import (
	"fmt"
	"math"

	alphamask2demo "github.com/MeKo-Christian/agg_go/internal/demo/alphamask2"
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
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()

	if float64(am2NumEllipses) != am2SliderValue {
		am2NumEllipses = int(am2SliderValue)
	}

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	// Render into a BGR24 work buffer like the original AGG example, then copy
	// back to the RGBA canvas image.
	workBuf := make([]uint8, w*h*3)
	alphamask2demo.RenderToBGR24(workBuf, w, h, alphamask2demo.Config{
		NumEllipses: am2NumEllipses,
		Angle:       am2LionAngle,
		Scale:       am2LionScale,
		SkewX:       am2LionSkewX,
		SkewY:       am2LionSkewY,
	})

	copyBGR24ToRGBA(workBuf, ctx.GetImage().Data, w, h)

	logStatus(fmt.Sprintf("Alpha Mask 2 Demo: Ellipses=%d", am2NumEllipses))
}

func handleAlphaMask2MouseDown(x, y float64, flags int) bool {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	dx := x - float64(w)/2
	dy := y - float64(h)/2
	am2LionAngle = math.Atan2(dy, dx)
	am2LionScale = math.Sqrt(dy*dy+dx*dx) / 100.0
	return true
}

func handleAlphaMask2RightMouseDown(x, y float64) bool {
	am2LionSkewX = x
	am2LionSkewY = y
	return true
}

func setAlphaMask2NumEllipses(n float64) {
	am2SliderValue = n
}

func copyBGR24ToRGBA(src, dst []uint8, width, height int) {
	for y := 0; y < height; y++ {
		srcOff := y * width * 3
		dstOff := y * width * 4
		for x := 0; x < width; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			dst[d+0] = src[s+2]
			dst[d+1] = src[s+1]
			dst[d+2] = src[s+0]
			dst[d+3] = 255
		}
	}
}
