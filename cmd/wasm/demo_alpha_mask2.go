package main

import (
	"fmt"
	"math"

	alphamask2demo "github.com/MeKo-Christian/agg_go/internal/demo/alphamask2"
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

	// Convert BGR24 work buffer to RGBA canvas image. No y-flip needed:
	// the work buffer uses a top-down layout matching the web canvas, and
	// the standalone example's FlipY is already accounted for by the π
	// rotation in RenderToBGR24.
	copyBGR24ToRGBA(workBuf, img.Data, w, h, img.Stride())

	applyLinearToSRGB(img)

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

// copyBGR24ToRGBA copies a BGR24 buffer into an RGBA buffer without y-flip.
func copyBGR24ToRGBA(src, dst []uint8, width, height, dstStride int) {
	srcStride := width * 3
	for y := 0; y < height; y++ {
		srcOff := y * srcStride
		dstOff := y * dstStride
		for x := 0; x < width; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			dst[d+0] = src[s+2] // R ← B in BGR
			dst[d+1] = src[s+1] // G
			dst[d+2] = src[s+0] // B ← R in BGR
			dst[d+3] = 255
		}
	}
}
