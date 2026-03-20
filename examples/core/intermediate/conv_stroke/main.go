// Port of AGG C++ conv_stroke.cpp example.
//
// Original caption: "AGG Example. Line Join"
//
// Demonstrates conv_stroke features: line join styles (miter, round, bevel)
// and line cap styles (butt, square, round) applied to open and closed paths.
// Since this Go port produces a static image, all join/cap combinations are
// shown in a 3×3 grid instead of the original interactive controls.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
)

// drawStrokeDemo draws the conv_stroke demo for a single join/cap combination
// at the given offset (ox, oy) in a cell of size cellW × cellH.
func drawStrokeDemo(a *agg.Agg2D, ox, oy, cellW, cellH float64, join agg.LineJoin, capStyle agg.LineCap, strokeWidth, miterLimit float64) {
	// Original points from conv_stroke.cpp on a 500×330 canvas (offset +100 on x).
	// We scale them to fit within the cell.
	const origW, origH = 500.0, 330.0
	scaleX := cellW / origW
	scaleY := cellH / origH
	scale := math.Min(scaleX, scaleY)

	tp := func(px, py float64) (float64, float64) {
		return ox + px*scaleX, oy + py*scaleY
	}

	// Three control points (same as in C++ example, x offset 100).
	px := [3]float64{57 + 100, 369 + 100, 143 + 100}
	py := [3]float64{60, 170, 310}

	buildPaths := func() {
		// Open zigzag path with one extra midpoint for stability check (same as C++).
		x0, y0 := tp(px[0], py[0])
		xm01, ym01 := tp((px[0]+px[1])/2, (py[0]+py[1])/2)
		x1, y1 := tp(px[1], py[1])
		x2, y2 := tp(px[2], py[2])

		a.MoveTo(x0, y0)
		a.LineTo(xm01, ym01)
		a.LineTo(x1, y1)
		a.LineTo(x2, y2)
		a.LineTo(x2, y2) // duplicate – numerical stability check (same as C++)

		// Closed triangle from midpoints.
		xm12, ym12 := tp((px[1]+px[2])/2, (py[1]+py[2])/2)
		xm20, ym20 := tp((px[2]+px[0])/2, (py[2]+py[0])/2)
		a.MoveTo(xm01, ym01)
		a.LineTo(xm12, ym12)
		a.LineTo(xm20, ym20)
		a.ClosePolygon()
	}

	// (1) Wide stroked path with the selected join/cap style.
	a.ResetPath()
	buildPaths()
	a.LineJoin(join)
	a.LineCap(capStyle)
	a.MiterLimit(miterLimit)
	a.LineWidth(strokeWidth * scale)
	a.LineColor(agg.NewColor(204, 178, 153, 255))
	a.NoFill()
	a.DrawPath(agg.StrokeOnly)

	// (2) Thin outline of the original raw path in black.
	a.ResetPath()
	buildPaths()
	a.LineJoin(agg.JoinMiter)
	a.LineCap(agg.CapButt)
	a.LineWidth(1.5 * scale)
	a.LineColor(agg.Black)
	a.DrawPath(agg.StrokeOnly)

	// (3) Semi-transparent fill of the raw path.
	a.ResetPath()
	buildPaths()
	a.FillColor(agg.NewColor(0, 0, 0, 51)) // rgba(0,0,0,0.2)
	a.NoLine()
	a.DrawPath(agg.FillOnly)
}

func drawCellBorder(a *agg.Agg2D, x, y, w, h float64) {
	a.LineColor(agg.NewColor(180, 180, 180, 255))
	a.LineWidth(1)
	a.NoFill()
	a.ResetPath()
	a.MoveTo(x, y)
	a.LineTo(x+w, y)
	a.LineTo(x+w, y+h)
	a.LineTo(x, y+h)
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)
}

const (
	cols    = 3 // join styles
	rows    = 3 // cap styles
	cellW   = 160.0
	cellH   = 96.0
	margin  = 5.0
	headerH = 22.0 // space for column headers at the top

	totalW = float64(cols)*cellW + float64(cols+1)*margin
	totalH = float64(rows)*cellH + float64(rows+1)*margin + headerH
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	joins := []agg.LineJoin{agg.JoinMiter, agg.JoinRound, agg.JoinBevel}
	caps := []agg.LineCap{agg.CapButt, agg.CapSquare, agg.CapRound}

	strokeWidth := 12.0
	miterLimit := 4.0

	for row, cap := range caps {
		for col, join := range joins {
			ox := margin + float64(col)*(cellW+margin)
			oy := headerH + margin + float64(row)*(cellH+margin)
			drawCellBorder(a, ox, oy, cellW, cellH)
			drawStrokeDemo(a, ox, oy, cellW, cellH, join, cap, strokeWidth, miterLimit)
		}
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Conv Stroke",
		Width:  int(totalW),
		Height: int(totalH),
	}, &demo{})
}
