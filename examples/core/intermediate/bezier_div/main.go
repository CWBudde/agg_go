// Port of AGG C++ bezier_div.cpp – Bezier curve subdivision accuracy demo.
//
// Shows a cubic Bezier curve rendered as a wide stroked shape together with
// the subdivision points. Default values from the WASM demo are used as
// constants; interactive sliders belong in the platform (SDL2/X11) variant.
//
// Default: subdivision mode, control points (170,424)(13,87)(488,423)(26,333),
// angle tolerance=15°, approx scale=1.0, stroke width=50.
package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/examples/shared/renderutil"
	"agg_go/internal/basics"
	"agg_go/internal/conv"
	"agg_go/internal/curves"
	"agg_go/internal/path"
)

const (
	width  = 655
	height = 520

	cx1, cy1 = 170.0, 424.0
	cx2, cy2 = 13.0, 87.0
	cx3, cy3 = 488.0, 423.0
	cx4, cy4 = 26.0, 333.0

	defaultAngleTol    = 15.0 // degrees
	defaultApproxScale = 1.0
	defaultCuspLimit   = 0.0 // degrees
	defaultWidth       = 50.0
)

func main() {
	ctx := agg.NewContext(width, height)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Light cream background.
	a.FillColor(agg.NewColor(255, 255, 242, 255))
	a.NoLine()
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(float64(width), 0)
	a.LineTo(float64(width), float64(height))
	a.LineTo(0, float64(height))
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	angleTol := defaultAngleTol * math.Pi / 180.0
	cuspLimit := defaultCuspLimit * math.Pi / 180.0

	// Build the curve using subdivision.
	curve := curves.NewCurve4Div()
	curve.SetApproximationScale(defaultApproxScale)
	curve.SetAngleTolerance(angleTol)
	curve.SetCuspLimit(cuspLimit)
	curve.Init(cx1, cy1, cx2, cy2, cx3, cy3, cx4, cy4)

	// Collect subdivision points.
	curvePath := path.NewPathStorageStl()
	curve.Rewind(0)
	for {
		x, y, cmd := curve.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			curvePath.MoveTo(x, y)
		} else if basics.IsVertex(cmd) {
			curvePath.LineTo(x, y)
		}
	}

	// Count vertices.
	numPoints := 0
	curvePath.Rewind(0)
	for {
		_, _, cmd := curvePath.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsVertex(basics.PathCommand(cmd)) {
			numPoints++
		}
	}

	// Wide stroke from the curve.
	curveAdapter := path.NewPathStorageStlVertexSourceAdapter(curvePath)
	stroke := conv.NewConvStroke(curveAdapter)
	stroke.SetWidth(defaultWidth)
	stroke.SetLineJoin(basics.MiterJoin)
	stroke.SetLineCap(basics.ButtCap)

	// Fill the wide stroke (semi-transparent green).
	a.ResetPath()
	stroke.Rewind(0)
	for {
		x, y, cmd := stroke.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		switch {
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsLineTo(cmd):
			a.LineTo(x, y)
		case basics.IsEndPoly(cmd) && basics.IsClose(uint32(cmd)):
			a.ClosePolygon()
		}
	}
	a.FillColor(agg.RGBA(0, 0.5, 0, 0.5))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	// Outline of the wide stroke.
	stroke2 := conv.NewConvStroke(stroke)
	stroke2.SetWidth(1.5)
	a.ResetPath()
	stroke2.Rewind(0)
	for {
		x, y, cmd := stroke2.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		switch {
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsLineTo(cmd):
			a.LineTo(x, y)
		case basics.IsEndPoly(cmd) && basics.IsClose(uint32(cmd)):
			a.ClosePolygon()
		}
	}
	a.FillColor(agg.RGBA(0, 0, 0, 0.5))
	a.DrawPath(agg.FillOnly)

	// Subdivision points as small dots.
	a.FillColor(agg.RGBA(0, 0, 0, 0.6))
	a.NoLine()
	curvePath.Rewind(0)
	for {
		x, y, cmd := curvePath.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsVertex(basics.PathCommand(cmd)) {
			a.FillCircle(x, y, 1.5)
		}
	}

	// Control polygon.
	a.NoFill()
	a.LineColor(agg.NewColor(80, 80, 200, 180))
	a.LineWidth(1.0)
	a.ResetPath()
	a.MoveTo(cx1, cy1)
	a.LineTo(cx2, cy2)
	a.LineTo(cx3, cy3)
	a.LineTo(cx4, cy4)
	a.DrawPath(agg.StrokeOnly)

	// Control points.
	for _, pt := range [][2]float64{{cx1, cy1}, {cx2, cy2}, {cx3, cy3}, {cx4, cy4}} {
		a.FillColor(agg.NewColor(0, 0, 200, 200))
		a.NoLine()
		a.FillCircle(pt[0], pt[1], 5)
	}

	fmt.Printf("Num Points=%d\n", numPoints)

	const filename = "bezier_div.png"
	if err := renderutil.SavePNG(ctx.GetImage(), filename); err != nil {
		panic(err)
	}
	fmt.Println(filename)
}
