package main

import (
	"github.com/MeKo-Christian/agg_go/internal/demo/patternresample"
)

var (
	patternResampleType  = 4
	patternResampleGamma = 2.0
	patternResampleBlur  = 1.0
	patternResampleNode  = -1
	patternResampleQuad  = [4][2]float64{{200, 100}, {600, 100}, {600, 500}, {200, 500}}
)

func handlePatternResampleMouseDown(x, y float64) bool {
	return handleQuadMouseDown(x, y, &patternResampleQuad, &patternResampleNode)
}

func handlePatternResampleMouseMove(x, y float64) bool {
	return handleQuadMouseMove(x, y, &patternResampleQuad, &patternResampleNode)
}

func handlePatternResampleMouseUp() {
	handleQuadMouseUp(&patternResampleNode)
}

func setPatternResampleType(v int) {
	if v < 0 {
		v = 0
	}
	if v > 5 {
		v = 5
	}
	patternResampleType = v
}

func setPatternResampleGamma(v float64) {
	if v < 0.5 {
		v = 0.5
	}
	if v > 3.0 {
		v = 3.0
	}
	patternResampleGamma = v
}

func setPatternResampleBlur(v float64) {
	if v < 0.5 {
		v = 0.5
	}
	if v > 2.0 {
		v = 2.0
	}
	patternResampleBlur = v
}

func setPatternResampleQuad(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	patternResampleQuad[0][0], patternResampleQuad[0][1] = x0, y0
	patternResampleQuad[1][0], patternResampleQuad[1][1] = x1, y1
	patternResampleQuad[2][0], patternResampleQuad[2][1] = x2, y2
	patternResampleQuad[3][0], patternResampleQuad[3][1] = x3, y3
}

func drawPatternResampleDemo() {
	patternresample.Draw(ctx, patternresample.Config{
		Mode:  patternResampleType,
		Gamma: patternResampleGamma,
		Blur:  patternResampleBlur,
		Quad:  patternResampleQuad,
	})
}
