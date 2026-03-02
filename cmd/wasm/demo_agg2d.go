// Port of the original AGG C++ example: agg2d_demo.cpp.
// Demonstrates the high-level Agg2D API: viewport mapping, rounded rects,
// aqua-style gradient buttons, ellipses, arc paths, blend modes,
// and radial gradients. (The font/text section is omitted in the web demo
// as no TTF files are available in the WASM environment.)
package main

import (
	"math"

	agg "agg_go"
)

// c8 is a shorthand for NewColor with uint8 components (0-255).
func c8(r, g, b, a uint8) agg.Color { return agg.NewColor(r, g, b, a) }

func drawAgg2DDemo() {
	g := ctx.GetAgg2D()

	g.ClearAll(agg.White)

	// Map a 600×600 logical coordinate space to the canvas, centred.
	g.Viewport(0, 0, 600, 600,
		0, 0, float64(width), float64(height),
		agg.XMidYMid)

	// --- Border ---
	g.LineColor(c8(0, 0, 0, 255))
	g.NoFill()
	g.LineWidth(1.0)
	g.RoundedRect(0.5, 0.5, 599.5, 599.5, 20.0)

	// --- Aqua Button (normal) ---
	xb1, yb1 := 400.0, 80.0
	xb2, yb2 := xb1+150, yb1+36

	g.FillColor(c8(0, 50, 180, 180))
	g.LineColor(c8(0, 0, 80, 255))
	g.LineWidth(1.0)
	g.RoundedRectXY(xb1, yb1, xb2, yb2, 12, 18)

	g.LineColor(c8(0, 0, 0, 0))
	g.FillLinearGradient(xb1, yb1, xb1, yb1+30,
		c8(100, 200, 255, 255), c8(255, 255, 255, 0), 1.0)
	g.RoundedRectVariableRadii(xb1+3, yb1+2.5, xb2-3, yb1+30, 1, 1, 9, 18)

	g.FillLinearGradient(xb1, yb2-20, xb1, yb2-3,
		c8(0, 0, 255, 0), c8(100, 255, 255, 255), 1.0)
	g.RoundedRectVariableRadii(xb1+3, yb2-20, xb2-3, yb2-2, 1, 1, 9, 18)

	// --- Aqua Button (pressed) ---
	xb1, yb1 = 400, 30
	xb2, yb2 = xb1+150, yb1+36

	g.FillColor(c8(0, 50, 180, 180))
	g.LineColor(c8(0, 0, 0, 255))
	g.LineWidth(2.0)
	g.RoundedRectXY(xb1, yb1, xb2, yb2, 12, 18)

	g.LineColor(c8(0, 0, 0, 0))
	g.FillLinearGradient(xb1, yb1+2, xb1, yb1+25,
		c8(60, 160, 255, 255), c8(100, 255, 255, 0), 1.0)
	g.RoundedRectVariableRadii(xb1+3, yb1+2.5, xb2-3, yb1+30, 9, 18, 1, 1)

	g.FillLinearGradient(xb1, yb2-25, xb1, yb2-5,
		c8(0, 180, 255, 0), c8(0, 200, 255, 255), 1.0)
	g.RoundedRectVariableRadii(xb1+3, yb2-25, xb2-3, yb2-2, 1, 1, 9, 18)

	// --- Ellipse ---
	g.LineWidth(3.5)
	g.LineColor(c8(20, 80, 80, 255))
	g.FillColor(c8(200, 255, 80, 200))
	g.Ellipse(450, 200, 50, 90)

	// --- Paths with arcs ---
	// Path 1: red semi-transparent filled arc
	g.ResetPath()
	g.FillColor(c8(255, 0, 0, 100))
	g.LineColor(c8(0, 0, 255, 100))
	g.LineWidth(2)
	g.MoveTo(300.0/2, 200.0/2)
	g.HorLineRel(-150.0 / 2)
	g.ArcRel(150.0/2, 150.0/2, 0, true, false, 150.0/2, -150.0/2)
	g.ClosePolygon()
	g.DrawPath(agg.FillAndStroke)

	// Path 2: yellow semi-transparent filled arc
	g.ResetPath()
	g.FillColor(c8(255, 255, 0, 100))
	g.LineColor(c8(0, 0, 255, 100))
	g.LineWidth(2)
	g.MoveTo(275.0/2, 175.0/2)
	g.VerLineRel(-150.0 / 2)
	g.ArcRel(150.0/2, 150.0/2, 0, false, false, -150.0/2, 150.0/2)
	g.ClosePolygon()
	g.DrawPath(agg.FillAndStroke)

	// Path 3: winding arc stroke
	deg2rad := func(d float64) float64 { return d * math.Pi / 180.0 }
	g.ResetPath()
	g.NoFill()
	g.LineColor(c8(127, 0, 0, 255))
	g.LineWidth(5)
	g.MoveTo(600.0/2, 350.0/2)
	g.LineRel(50.0/2, -25.0/2)
	g.ArcRel(25.0/2, 25.0/2, deg2rad(-30), false, true, 50.0/2, -25.0/2)
	g.LineRel(50.0/2, -25.0/2)
	g.ArcRel(25.0/2, 50.0/2, deg2rad(-30), false, true, 50.0/2, -25.0/2)
	g.LineRel(50.0/2, -25.0/2)
	g.ArcRel(25.0/2, 75.0/2, deg2rad(-30), false, true, 50.0/2, -25.0/2)
	g.LineRel(50.0, -25.0)
	g.ArcRel(25.0/2, 100.0/2, deg2rad(-30), false, true, 50.0/2, -25.0/2)
	g.LineRel(50.0/2, -25.0/2)
	g.DrawPath(agg.StrokeOnly)

	// --- Master alpha: everything below is slightly translucent ---
	g.MasterAlpha(0.85)

	// --- Blend mode ellipses ---
	g.NoLine()
	g.FillColor(c8(70, 70, 0, 255))
	g.BlendMode(agg.BlendAdd)
	g.Ellipse(500, 280, 20, 40)

	g.FillColor(c8(255, 255, 255, 255))
	g.BlendMode(agg.BlendOverlay)
	g.Ellipse(540, 280, 20, 40)

	// --- Radial gradient ellipse ---
	g.BlendMode(agg.BlendAlpha)
	g.FillRadialGradientMultiStop(400, 500, 40,
		c8(255, 255, 0, 0),
		c8(0, 0, 127, 255),
		c8(0, 255, 0, 0))
	g.NoLine()
	g.Ellipse(400, 500, 40, 40)

	// Restore state
	g.MasterAlpha(1.0)
	g.BlendMode(agg.BlendAlpha)
}
