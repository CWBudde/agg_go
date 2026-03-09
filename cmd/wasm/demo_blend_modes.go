// Blend mode gallery demo (separate from compositing.cpp direct port).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
)

func drawBlendModesDemo() {
	ctx.Clear(agg.RGBA(0.9, 0.9, 0.9, 1.0))

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.FontGSV(10)
	agg2d.TextAlignment(agg.AlignCenter, agg.AlignTop)

	modes := []struct {
		name string
		mode agg.BlendMode
	}{
		{"Alpha", agg.BlendAlpha},
		{"Multiply", agg.BlendMultiply},
		{"Screen", agg.BlendScreen},
		{"Overlay", agg.BlendOverlay},
		{"Darken", agg.BlendDarken},
		{"Lighten", agg.BlendLighten},
		{"ColorDodge", agg.BlendColorDodge},
		{"ColorBurn", agg.BlendColorBurn},
		{"HardLight", agg.BlendHardLight},
		{"SoftLight", agg.BlendSoftLight},
		{"Difference", agg.BlendDifference},
		{"Exclusion", agg.BlendExclusion},
	}

	const cols = 4
	rows := (len(modes) + cols - 1) / cols
	canvasW := float64(ctx.GetImage().Width())
	canvasH := float64(ctx.GetImage().Height())
	const (
		gapX = 8.0
		gapY = 8.0
	)
	cellW := (canvasW - float64(cols-1)*gapX) / float64(cols)
	cellH := (canvasH - float64(rows-1)*gapY) / float64(rows)
	const (
		r   = 40.0
		c1x = 70.0
		c1y = 60.0
		c2x = 110.0
		c2y = 60.0
		c3x = 90.0
		c3y = 100.0
	)
	// Center the three-circle cluster within each grid cell.
	groupMinX := min3(c1x-r, c2x-r, c3x-r)
	groupMaxX := max3(c1x+r, c2x+r, c3x+r)
	groupMinY := min3(c1y-r, c2y-r, c3y-r)
	groupMaxY := max3(c1y+r, c2y+r, c3y+r)
	shiftX := (cellW-(groupMaxX-groupMinX))*0.5 - groupMinX
	const labelBandH = 20.0
	drawAreaH := cellH - labelBandH
	shiftY := (drawAreaH-(groupMaxY-groupMinY))*0.5 - groupMinY

	for i, m := range modes {
		x := float64(i%cols) * (cellW + gapX)
		y := float64(i/cols) * (cellH + gapY)

		agg2d.BlendMode(agg.BlendAlpha)
		agg2d.FillColor(agg.RGBA(1.0, 1.0, 1.0, 0.6))
		agg2d.NoLine()
		agg2d.ResetPath()
		agg2d.MoveTo(x+4, y+4)
		agg2d.LineTo(x+cellW-4, y+4)
		agg2d.LineTo(x+cellW-4, y+cellH-4)
		agg2d.LineTo(x+4, y+cellH-4)
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)

		agg2d.FillColor(agg.NewColor(255, 0, 0, 180))
		agg2d.FillCircle(x+shiftX+c1x, y+shiftY+c1y, r)
		agg2d.FillColor(agg.NewColor(0, 255, 0, 180))
		agg2d.FillCircle(x+shiftX+c2x, y+shiftY+c2y, r)

		agg2d.BlendMode(m.mode)
		agg2d.FillColor(agg.NewColor(0, 0, 255, 200))
		agg2d.FillCircle(x+shiftX+c3x, y+shiftY+c3y, r)

		agg2d.BlendMode(agg.BlendAlpha)
		agg2d.FillColor(agg.NewColor(20, 20, 20, 255))
		agg2d.Text(x+cellW*0.5, y+drawAreaH+4, m.name, false, 0, 0)
	}

	agg2d.BlendMode(agg.BlendAlpha)
}

func min3(a, b, c float64) float64 {
	if a > b {
		a = b
	}
	if a > c {
		a = c
	}
	return a
}

func max3(a, b, c float64) float64 {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}
