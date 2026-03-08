// Port of AGG C++ component_rendering.cpp – component (channel) rendering.
//
// Three large circles are each rendered into an individual color channel
// (Red, Green, Blue) by using Multiply blend mode with the complementary
// color (Cyan/Magenta/Yellow). The effect shows subtractive CMY mixing:
//
//	Red ∩ Green  → Blue   (Cyan × Magenta)
//	Red ∩ Blue   → Green  (Cyan × Yellow)
//	Green ∩ Blue → Red    (Magenta × Yellow)
//	All three    → Black
package main

import (
	agg "agg_go"
	"agg_go/examples/shared/demorunner"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	w := float64(ctx.Width())
	h := float64(ctx.Height())
	cx := w / 2
	cy := h / 2

	const (
		offset = 100.0
		radius = 200.0
	)

	// Use Multiply blend so drawing a CMY circle darkens only the
	// corresponding channel – mathematically equivalent to the C++
	// per-channel gray rendering.
	a.BlendMode(agg.BlendMultiply)

	// Red channel → Cyan (removes R).
	a.FillColor(agg.NewColor(0, 255, 255, 255))
	a.FillCircle(cx-0.87*offset, cy-0.5*offset, radius)

	// Green channel → Magenta (removes G).
	a.FillColor(agg.NewColor(255, 0, 255, 255))
	a.FillCircle(cx+0.87*offset, cy-0.5*offset, radius)

	// Blue channel → Yellow (removes B).
	a.FillColor(agg.NewColor(255, 255, 0, 255))
	a.FillCircle(cx, cy+offset, radius)

	// Restore normal blending.
	a.BlendMode(agg.BlendAlpha)
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Component Rendering", Width: 800, Height: 600}, &demo{})
}
