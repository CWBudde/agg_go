// Port of AGG C++ component_rendering.cpp – component (channel) rendering.
//
// Three large circles are each rendered into an individual color channel
// (Red, Green, Blue) by using Multiply blend mode with the complementary
// color (Cyan/Magenta/Yellow). The effect shows subtractive CMY mixing:
//   Red ∩ Green  → Blue   (Cyan × Magenta)
//   Red ∩ Blue   → Green  (Cyan × Yellow)
//   Green ∩ Blue → Red    (Magenta × Yellow)
//   All three    → Black
// An alpha slider controls how strongly each channel is darkened.
package main

import (
	agg "agg_go"
)

// --- State ---

var compAlpha = 255 // 0..255

// --- Drawing ---

func drawComponentRenderingDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// White background is already cleared by renderDemo, but ensure it here
	// in case the demo is rendered standalone (stub path).
	ctx.Clear(agg.White)

	w := float64(width)
	h := float64(height)
	cx := w / 2
	cy := h / 2

	// Circle layout: equilateral triangle, scaled to fill the canvas.
	// Matches the original offset=100, radius=200 for an 800×600 canvas.
	const (
		offset = 100.0
		radius = 200.0
	)

	alpha := uint8(compAlpha)

	// Use Multiply blend so drawing a CMY circle darkens only the
	// corresponding channel – mathematically equivalent to the C++
	// per-channel gray rendering.
	a.BlendMode(agg.BlendMultiply)

	// Red channel → Cyan (removes R).
	a.FillColor(agg.NewColor(0, 255, 255, alpha))
	a.FillCircle(cx-0.87*offset, cy-0.5*offset, radius)

	// Green channel → Magenta (removes G).
	a.FillColor(agg.NewColor(255, 0, 255, alpha))
	a.FillCircle(cx+0.87*offset, cy-0.5*offset, radius)

	// Blue channel → Yellow (removes B).
	a.FillColor(agg.NewColor(255, 255, 0, alpha))
	a.FillCircle(cx, cy+offset, radius)

	// Restore normal blending for subsequent demos.
	a.BlendMode(agg.BlendAlpha)
}
