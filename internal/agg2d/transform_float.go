// Package agg2d float transform + fill-mode methods (L5/breadth). Color-agnostic
// delegations mirrored onto Agg2DFloat; bodies match transform.go / fill_rules.go
// since they only manipulate the shared world transform and fill state.
package agg2d

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/transform"
)

// Rotate applies a rotation (radians) to the world transform.
func (a *Agg2DFloat) Rotate(angle float64) {
	a.transform.Rotate(angle)
	a.updateApproximationScales()
}

// Scale applies a scaling transformation.
func (a *Agg2DFloat) Scale(sx, sy float64) {
	a.transform.ScaleXY(sx, sy)
	a.updateApproximationScales()
}

// UniformScale applies uniform scaling on both axes.
func (a *Agg2DFloat) UniformScale(s float64) { a.Scale(s, s) }

// Skew applies a skewing transformation (radians).
func (a *Agg2DFloat) Skew(sx, sy float64) {
	skew := transform.NewTransAffineFromValues(1.0, math.Tan(sy), math.Tan(sx), 1.0, 0.0, 0.0)
	a.Affine(skew)
}

// Translate applies a translation.
func (a *Agg2DFloat) Translate(x, y float64) {
	a.transform.Translate(x, y)
}

// Affine multiplies the world transform by tr.
func (a *Agg2DFloat) Affine(tr *transform.TransAffine) {
	a.transform.Multiply(tr)
	a.updateApproximationScales()
}

// Parallelogram applies a transform mapping the unit square to a parallelogram
// defined by three corners; the fourth is derived. Mirrors transform.go.
func (a *Agg2DFloat) Parallelogram(x1, y1, x2, y2, x3, y3 float64) {
	sx := x2 - x1
	shx := x3 - x1
	sy := y3 - y1
	shy := y2 - y1
	tx := x1
	ty := y1

	parallelogramTransform := transform.NewTransAffineFromValues(sx, shy, shx, sy, tx, ty)
	a.Affine(parallelogramTransform)
}

// ParallelogramFromRect applies a transform mapping the given source rectangle
// onto the parallelogram defined by three corners.
func (a *Agg2DFloat) ParallelogramFromRect(rectX1, rectY1, rectX2, rectY2,
	x1, y1, x2, y2, x3, y3 float64,
) {
	rectWidth := rectX2 - rectX1
	rectHeight := rectY2 - rectY1

	if rectWidth == 0 || rectHeight == 0 {
		return
	}

	a.Translate(-rectX1, -rectY1)
	a.Scale(1.0/rectWidth, 1.0/rectHeight)
	a.Parallelogram(x1, y1, x2, y2, x3, y3)
}

// NoFill disables fill by setting a transparent fill color.
func (a *Agg2DFloat) NoFill() {
	a.fillColor = Color{0, 0, 0, 0}
	a.fillGradientFlag = Solid
}

// NoLine disables stroke by setting a transparent line color.
func (a *Agg2DFloat) NoLine() {
	a.lineColor = Color{0, 0, 0, 0}
	a.lineGradientFlag = Solid
}

// FillEvenOdd selects the even-odd (true) or non-zero (false) fill rule.
func (a *Agg2DFloat) FillEvenOdd(evenOddFlag bool) {
	a.evenOddFlag = evenOddFlag
	if a.rasterizer != nil {
		a.rasterizer.FillingRule(a.getFillRule())
	}
}

// GetFillEvenOdd reports whether the even-odd fill rule is active.
func (a *Agg2DFloat) GetFillEvenOdd() bool { return a.evenOddFlag }

func (a *Agg2DFloat) getFillRule() basics.FillingRule {
	if a.evenOddFlag {
		return basics.FillEvenOdd
	}
	return basics.FillNonZero
}
