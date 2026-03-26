package main

import "github.com/MeKo-Christian/agg_go/internal/demo/linethickness"

// Port of AGG C++ line_thickness.cpp.
//
// Web variant keeps controls outside AGG widgets: parameters are controlled
// via JS/URL query params.
var (
	lineThicknessState = linethickness.DefaultState()
)

func setLineThicknessFactor(v float64) {
	lineThicknessState.Thickness = v
	lineThicknessState.Clamp()
}

func setLineThicknessBlur(v float64) {
	lineThicknessState.Blur = v
	lineThicknessState.Clamp()
}

func setLineThicknessMono(v bool) { lineThicknessState.Mono = v }

func setLineThicknessInvert(v bool) { lineThicknessState.Invert = v }

func getLineThicknessBlurTime() float64 { return linethickness.LastBlurMS() }

func drawLineThicknessDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	// Render into work buffer using C++ y-up coordinate frame (row 0 = bottom).
	workBuf := make([]uint8, w*h*4)
	linethickness.Draw(workBuf, w, h, lineThicknessState)

	// Copy work buffer to canvas with y-flip.
	stride := w * 4
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * stride
		dstOff := y * stride
		copy(img.Data[dstOff:dstOff+stride], workBuf[srcOff:srcOff+stride])
	}

	applyLinearToSRGB(img)
}
