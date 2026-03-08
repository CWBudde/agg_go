package main

import (
	"agg_go/internal/demo/patternperspective"
)

var (
	patternPerspectiveType = 2
	patternPerspectiveQuad = [4][2]float64{{200, 100}, {600, 100}, {600, 500}, {200, 500}}
)

func setPatternPerspectiveType(v int) {
	if v < 0 {
		v = 0
	}
	if v > 2 {
		v = 2
	}
	patternPerspectiveType = v
}

func setPatternPerspectiveQuad(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	patternPerspectiveQuad[0][0], patternPerspectiveQuad[0][1] = x0, y0
	patternPerspectiveQuad[1][0], patternPerspectiveQuad[1][1] = x1, y1
	patternPerspectiveQuad[2][0], patternPerspectiveQuad[2][1] = x2, y2
	patternPerspectiveQuad[3][0], patternPerspectiveQuad[3][1] = x3, y3
}

func drawPatternPerspectiveDemo() {
	patternperspective.Draw(ctx, patternperspective.Config{
		Mode: patternPerspectiveType,
		Quad: patternPerspectiveQuad,
	})
}
