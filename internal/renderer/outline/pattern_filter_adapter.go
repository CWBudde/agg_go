// Package outline provides pattern filter adapters for outline rendering.
package outline

import (
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
)

const rgba8ToFloat64 = 1.0 / 255.0

// PatternFilterRGBAAdapter adapts span pattern filters to work with outline renderer Filter interface.
// This bridges the gap between the span package's pattern filters and the outline package's requirements.
type PatternFilterRGBAAdapter struct{}

// NewPatternFilterRGBAAdapter creates a new pattern filter adapter.
func NewPatternFilterRGBAAdapter() *PatternFilterRGBAAdapter {
	return &PatternFilterRGBAAdapter{}
}

// Dilation returns the filter dilation.
func (pfa *PatternFilterRGBAAdapter) Dilation() int {
	return 1
}

func rgbaComponent(v float64) int {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return int(v*255.0 + 0.5)
}

// PixelHighRes performs high-resolution bilinear sampling directly on RGBA rows.
func (pfa *PatternFilterRGBAAdapter) PixelHighRes(rows [][]color.RGBA, p *color.RGBA, x, y int) {
	if len(rows) == 0 || p == nil {
		*p = color.NewRGBA(0, 0, 0, 0)
		return
	}

	xLr := x >> primitives.LineSubpixelShift
	yLr := y >> primitives.LineSubpixelShift

	x &= primitives.LineSubpixelMask
	y &= primitives.LineSubpixelMask

	if yLr < 0 || yLr >= len(rows) || xLr < 0 || xLr >= len(rows[yLr]) {
		*p = color.NewRGBA(0, 0, 0, 0)
		return
	}

	var r, g, b, a int

	weight := (primitives.LineSubpixelScale - x) * (primitives.LineSubpixelScale - y)
	ptr := rows[yLr][xLr]
	r += weight * rgbaComponent(ptr.R)
	g += weight * rgbaComponent(ptr.G)
	b += weight * rgbaComponent(ptr.B)
	a += weight * rgbaComponent(ptr.A)

	if xLr+1 < len(rows[yLr]) {
		weight = x * (primitives.LineSubpixelScale - y)
		ptr = rows[yLr][xLr+1]
		r += weight * rgbaComponent(ptr.R)
		g += weight * rgbaComponent(ptr.G)
		b += weight * rgbaComponent(ptr.B)
		a += weight * rgbaComponent(ptr.A)
	}

	if yLr+1 < len(rows) && xLr < len(rows[yLr+1]) {
		weight = (primitives.LineSubpixelScale - x) * y
		ptr = rows[yLr+1][xLr]
		r += weight * rgbaComponent(ptr.R)
		g += weight * rgbaComponent(ptr.G)
		b += weight * rgbaComponent(ptr.B)
		a += weight * rgbaComponent(ptr.A)
	}

	if yLr+1 < len(rows) && xLr+1 < len(rows[yLr+1]) {
		weight = x * y
		ptr = rows[yLr+1][xLr+1]
		r += weight * rgbaComponent(ptr.R)
		g += weight * rgbaComponent(ptr.G)
		b += weight * rgbaComponent(ptr.B)
		a += weight * rgbaComponent(ptr.A)
	}

	shift := primitives.LineSubpixelShift * 2
	*p = color.NewRGBA(
		float64(r>>shift)*rgba8ToFloat64,
		float64(g>>shift)*rgba8ToFloat64,
		float64(b>>shift)*rgba8ToFloat64,
		float64(a>>shift)*rgba8ToFloat64,
	)
}
