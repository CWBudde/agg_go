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

func drawLineThicknessDemo() {
	linethickness.Draw(ctx, lineThicknessState)
}
