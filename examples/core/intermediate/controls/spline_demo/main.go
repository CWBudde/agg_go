// Package main demonstrates the spline control implementation.
// This example shows how to create and interact with a spline curve editor.
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/color"
	"agg_go/internal/ctrl/spline"
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	// Create a spline control with RGBA colors
	ctrl := spline.NewSplineCtrlRGBA(10, 10, 300, 200, 6, false)

	// Set up some initial control points for a nice curve
	ctrl.SetPoint(1, 0.2, 0.8) // High peak near start
	ctrl.SetPoint(2, 0.4, 0.3) // Dip in middle
	ctrl.SetPoint(3, 0.6, 0.7) // Another peak
	ctrl.SetPoint(4, 0.8, 0.2) // Low end

	// Demonstrate color customization
	ctrl.SetCurveColor(color.NewRGBA(0.0, 0.8, 0.0, 1.0))       // Green curve
	ctrl.SetActivePointColor(color.NewRGBA(0.8, 0.4, 0.0, 1.0)) // Orange active point

	// Display info
	a := ctx.GetAgg2D()
	a.FillColor(agg.Black)
	a.TextAlignment(agg.AlignLeft, agg.AlignTop)

	lines := []string{
		"Spline Control Demo",
		fmt.Sprintf("Control bounds: (%.0f,%.0f) to (%.0f,%.0f)", ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2()),
		"Spline values:",
	}
	for i, line := range lines {
		a.Text(10, float64(220+i*18), line, false, 0, 0)
	}

	for i := 0; i <= 5; i++ {
		x := float64(i) / 5.0
		y := ctrl.Value(x)
		a.Text(10, float64(220+len(lines)*18+i*18), fmt.Sprintf("  x=%.1f -> y=%.3f", x, y), false, 0, 0)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Spline Demo",
		Width:  400,
		Height: 400,
	}, &demo{})
}
