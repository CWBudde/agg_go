package main

import "agg_go/internal/demo/linepatterns"

// Port of AGG C++ line_patterns.cpp (web variant).
var (
	linePatternScaleX = 1.0
	linePatternStartX = 0.0
)

func setLinePatternScaleX(v float64) {
	if v < 0.2 {
		v = 0.2
	}
	if v > 3.0 {
		v = 3.0
	}
	linePatternScaleX = v
}

func setLinePatternStartX(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 10 {
		v = 10
	}
	linePatternStartX = v
}

func drawLinePatternsDemo() {
	linepatterns.Draw(ctx.GetImage(), linePatternScaleX, linePatternStartX)
}
